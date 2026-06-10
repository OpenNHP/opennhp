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
# nhp-acd shells out to iptables/ipset. On Amazon Linux 2023 (iptables-nft
# backend) the unprivileged user needs more than CAP_NET_ADMIN: CAP_NET_RAW
# for the raw sockets iptables opens, and CAP_DAC_OVERRIDE so it can take the
# /run/xtables.lock file lock. Without these, `iptables -L` at startup fails
# with "exit status 1" and the daemon never starts. Granted as ambient caps so
# the unprivileged user inherits them across exec; bounded so the process
# cannot acquire additional capabilities at runtime.
AmbientCapabilities=CAP_NET_ADMIN CAP_NET_RAW CAP_DAC_OVERRIDE
CapabilityBoundingSet=CAP_NET_ADMIN CAP_NET_RAW CAP_DAC_OVERRIDE
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable nhp-acd
