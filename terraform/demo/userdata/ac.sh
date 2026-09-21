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
#                    disables audit logging.
#   CAP_DAC_OVERRIDE pin programs/maps under /sys/fs/bpf, which systemd mounts
#                    0700 root:root. CAP_SYS_ADMIN does not bypass DAC file
#                    permission checks, so this one is required regardless.
# CAP_SYS_ADMIN and CAP_SYS_RESOURCE are carried as a safety margin for the
# initial cutover (BPF/perf superset; RemoveMemlock on older kernels) and can
# be dropped once the deployment is verified. Granted as ambient caps so the
# unprivileged user inherits them across exec; bounded so the process cannot
# acquire additional capabilities at runtime. Keep this in sync with the
# live-unit sed patch in .github/workflows/deploy-demo-v2.yml (userdata only
# runs at instance creation).
AmbientCapabilities=CAP_BPF CAP_NET_ADMIN CAP_PERFMON CAP_SYS_ADMIN CAP_SYS_RESOURCE CAP_DAC_OVERRIDE
CapabilityBoundingSet=CAP_BPF CAP_NET_ADMIN CAP_PERFMON CAP_SYS_ADMIN CAP_SYS_RESOURCE CAP_DAC_OVERRIDE
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable nhp-acd
