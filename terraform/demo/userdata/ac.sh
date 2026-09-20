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

# eBPF/XDP dependencies for FilterMode = 1.
# - bpftool/kernel-tools: ops + troubleshooting (nhp-acd does not shell
#   out to it, but its absence usually means kernel-headers tooling is
#   also missing, which we want surfaced loudly).
# - iptables/ipset are NO LONGER required: the demo pipeline runs the
#   daemon in FilterMode = 1 (eBPF/XDP). They may still be installed
#   by the AMI image; we deliberately do not remove them because
#   other host tooling may rely on them.
if ! command -v bpftool >/dev/null 2>&1; then
  dnf install -y bpftool kernel-tools || true
fi

# Mount the bpffs so nhp-acd can pin programs/maps at /sys/fs/bpf.
# Pinning is required by endpoints/ac/ebpf/ebpfegine.go: every map and
# program is pinned under /sys/fs/bpf so a daemon reload reattaches
# them deterministically.
if ! mountpoint -q /sys/fs/bpf; then
  mkdir -p /sys/fs/bpf
  mount -t bpf bpf /sys/fs/bpf || true
fi
# Persist bpffs across reboots so future userdata runs don't need to
# re-mount (idempotent: fstab entry is added only when absent).
if ! grep -Eq '^[^#]*\s+/sys/fs/bpf\s+bpf\b' /etc/fstab; then
  echo "bpf /sys/fs/bpf bpf defaults 0 0" >> /etc/fstab
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
# FilterMode = 1 (eBPF/XDP) needs:
#   - CAP_NET_ADMIN: attach/detach XDP and TC programs
#   - CAP_BPF:      kernel >= 5.8 to load BPF programs/maps without
#                   falling back to CAP_SYS_ADMIN
# CAP_NECESSARY_NOT_USED_HERE (CAP_NET_RAW / CAP_DAC_OVERRIDE) were
# retained historically for the iptables/ipset path and are no longer
# required now that the demo deploys in eBPF mode.
AmbientCapabilities=CAP_NET_ADMIN CAP_BPF
CapabilityBoundingSet=CAP_NET_ADMIN CAP_BPF
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable nhp-acd