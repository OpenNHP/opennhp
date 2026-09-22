#!/bin/bash
set -euo pipefail

DEPLOY_PATH="${deploy_path}"

# Create deploy directory
mkdir -p "$DEPLOY_PATH/etc"
mkdir -p "$DEPLOY_PATH/cert"
# nhp-acd writes to ExeDirPath/logs (plural); see endpoints/ac/udpac.go.
mkdir -p "$DEPLOY_PATH/logs"
chown -R ec2-user:ec2-user "$DEPLOY_PATH"

# Install certbot for TLS certificates
dnf install -y certbot nginx

# nhp-acd runs in eBPF/XDP mode, and endpoints/ac/ebpf/ebpfegine.go attaches the
# egress program with link.AttachTCX - the kernel TCX / bpf_mprog API added in
# Linux 6.6, with no fallback path. AL2023 AMIs still boot 6.1 by default, where
# the attach fails and nhp-acd refuses to start, so pull in the 6.12 kernel and
# reboot into it. The deploy-ac job refuses to deploy FilterMode = 1 onto
# anything older, so without this a fresh instance could only run in iptables
# mode. Avoid $${...} shell syntax in this file: it is rendered through
# Terraform templatefile(), which would read it as an interpolation.
NEED_REBOOT=0
KVER_MAJOR=$(uname -r | cut -d. -f1)
KVER_MINOR=$(uname -r | cut -d. -f2)
if [ "$KVER_MAJOR" -lt 6 ] || { [ "$KVER_MAJOR" -eq 6 ] && [ "$KVER_MINOR" -lt 6 ]; }; then
  if dnf install -y kernel6.12; then
    NEED_REBOOT=1
  else
    echo "WARNING: kernel6.12 unavailable; staying on $(uname -r), eBPF mode will not deploy" >&2
  fi
fi

# Create systemd service
cat > /etc/systemd/system/nhp-acd.service <<'EOF'
[Unit]
Description=NHP Access Controller Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=ec2-user
Group=ec2-user
WorkingDirectory=/home/ec2-user/nhp-ac
ExecStart=/home/ec2-user/nhp-ac/nhp-acd run
Restart=on-failure
RestartSec=5
LimitNOFILE=65536
# nhp-acd runs in eBPF/XDP mode (FilterMode = 1 in etc/config.toml), so the
# unprivileged user needs the BPF capability set rather than the iptables one:
#   CAP_BPF          load the XDP + TCX programs and create their maps
#   CAP_NET_ADMIN    attach XDP (generic mode) and the TCX egress hook
#   CAP_PERFMON      open the "events" perf ring buffer used for the
#                    accept/deny audit logs. Missing this does NOT stop the
#                    daemon - the perf reader only logs - it just silently
#                    disables audit logging, which is why the deploy job greps
#                    for the "Start listening for eBPF events" line.
#   CAP_DAC_OVERRIDE pin programs/maps under /sys/fs/bpf, which systemd mounts
#                    0700 root:root.
# This is the whole set: no CAP_SYS_ADMIN (as an ambient capability on an
# internet-facing daemon it is root-equivalent, and NoNewPrivileges= does not
# mitigate it) and no CAP_SYS_RESOURCE (BPF memory has been memcg-accounted
# since 5.11, so rlimit.RemoveMemlock() is a no-op here and ebpfegine.go only
# logs if it fails). Granted as ambient caps so the unprivileged user inherits
# them across exec; bounded so the process cannot acquire additional
# capabilities at runtime. Keep in sync with the live-unit sed patch in
# .github/workflows/deploy-demo-v2.yml (userdata only runs at instance
# creation). That job also drops in
# nhp-acd.service.d/10-ebpf-backstop.conf, which adds the ExecStartPre/
# ExecStartPost/ExecStopPost hooks for deploy/scripts/nhp-ac-backstop.sh - the
# netfilter default-deny that keeps the protected port closed while nhp-acd is
# not enforcing (the XDP links die with the process). Those hooks use systemd's
# "+" prefix, so they run as root and are exempt from the sets below.
AmbientCapabilities=CAP_BPF CAP_NET_ADMIN CAP_PERFMON CAP_DAC_OVERRIDE
CapabilityBoundingSet=CAP_BPF CAP_NET_ADMIN CAP_PERFMON CAP_DAC_OVERRIDE
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
# Enabled, not started: the binary, configs, eBPF objects and the backstop
# drop-in all arrive with the first deploy-ac run.
systemctl enable nhp-acd

# Reboot last, once the unit is in place, to pick up the 6.12 kernel installed
# above. cloud-init userdata runs once per instance, so this does not loop.
if [ "$NEED_REBOOT" -eq 1 ]; then
  echo "rebooting into the newly installed kernel for eBPF/TCX support"
  reboot
fi
