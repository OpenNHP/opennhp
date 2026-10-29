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
# reboot into it, so a fresh instance is ready for FilterMode = 1 from first
# boot.
#
# Installing the package is not enough: it does not reliably become the
# default boot entry, so the host would reboot straight back into 6.1. AWS's
# documented 6.12 procedure pins it with grubby, which is what the
# --set-default below does. The 6.1 `kernel` package stays installed here
# (dnf will not remove the running kernel), so a later `dnf upgrade` pulling a
# newer 6.1 build could take the default back via UPDATEDEFAULT=yes; the
# deploy-ac job removes that package on its first run and re-checks
# `grubby --default-kernel` on every eBPF-mode deploy.
#
# This covers fresh instances only: userdata runs once per instance, and
# aws_instance.ac both leaves user_data_replace_on_change at false and ignores
# user_data changes in its lifecycle block (see ec2.tf), so editing this file
# neither recreates nor restarts a running host. Long-lived hosts are
# upgraded by the deploy-ac job in .github/workflows/deploy-demo-v2.yml, which
# runs the same dnf install and grubby pin, reboots and waits for the host to
# come back - behind its fail-closed backstop. Keep the two in step.
#
# Avoid $${...} shell syntax in this file: it is rendered through Terraform
# templatefile(), which would read it as an interpolation. Same for rpm's
# %%{...} query tags below - templatefile() reads a bare %%{ as the start of a
# directive and fails to render, so it has to be doubled.
NEED_REBOOT=0
KVER_MAJOR=$(uname -r | cut -d. -f1)
KVER_MINOR=$(uname -r | cut -d. -f2)
if [ "$KVER_MAJOR" -lt 6 ] || { [ "$KVER_MAJOR" -eq 6 ] && [ "$KVER_MINOR" -lt 6 ]; }; then
  if dnf install -y kernel6.12; then
    KVER_612=$(rpm -q --qf '%%{version}-%%{release}.%%{arch}\n' kernel6.12 2>/dev/null | sort -V | tail -1) || KVER_612=""
    if [ -n "$KVER_612" ] && grubby --set-default "/boot/vmlinuz-$KVER_612"; then
      NEED_REBOOT=1
    else
      echo "WARNING: could not make kernel6.12 the default boot entry; not rebooting into it" >&2
    fi
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
#                    the daemon's log for "Start listening for eBPF events".
# This is the whole set. No CAP_SYS_ADMIN: as an ambient capability on an
# internet-facing daemon it is root-equivalent, and NoNewPrivileges= does not
# mitigate it. No CAP_DAC_OVERRIDE either, for the same reason by a shorter
# path - it bypasses every DAC file permission check, so a compromised nhp-acd
# could write /etc/cron.d/*, ~root/.ssh/authorized_keys or this unit file.
# It used to be here so the daemon could pin programs/maps under /sys/fs/bpf,
# which systemd's sys-fs-bpf.mount leaves 0700 root:root; the backstop drop-in
# described further down now runs "nhp-ac-backstop.sh bpffs-prep" from a root
# ExecStartPre instead, which chgrp/chmods that one directory to 0770 ec2-user. And no CAP_SYS_RESOURCE (BPF memory has been
# memcg-accounted since 5.11, so rlimit.RemoveMemlock() is a no-op here and
# ebpfegine.go only logs if it fails). Granted as ambient caps so the
# unprivileged user inherits them across exec; bounded so the process cannot
# acquire additional capabilities at runtime. Keep in sync with the live-unit
# sed patch in .github/workflows/deploy-demo-v2.yml (userdata only runs at
# instance creation). That job also drops in
# nhp-acd.service.d/10-ebpf-backstop.conf, which adds the ExecStartPre/
# ExecStartPost/ExecStopPost hooks for deploy/scripts/nhp-ac-backstop.sh - the
# netfilter default-deny that keeps the protected port closed while nhp-acd is
# not enforcing (the XDP links die with the process) - plus the /sys/fs/bpf
# preparation above. Those hooks use systemd's "+" prefix, so they run as root
# and are exempt from the sets below. Starting this unit in FilterMode = 1
# *without* that drop-in therefore fails at the pin step: the first deploy-ac
# run installs both, and the binary is not on the host before it anyway.
AmbientCapabilities=CAP_BPF CAP_NET_ADMIN CAP_PERFMON
CapabilityBoundingSet=CAP_BPF CAP_NET_ADMIN CAP_PERFMON
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
