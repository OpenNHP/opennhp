#!/bin/bash
set -euo pipefail

# Contain the kernel flaw before starting any demo listener.
dnf install -y iptables-nft
base64 -d > /tmp/layerv-kernel-mitigation.sh <<'KERNEL_MITIGATION'
${kernel_mitigation_base64}
KERNEL_MITIGATION
bash /tmp/layerv-kernel-mitigation.sh
rm /tmp/layerv-kernel-mitigation.sh

DEPLOY_PATH="${deploy_path}"

# Create deploy directory
mkdir -p "$DEPLOY_PATH/etc"
mkdir -p "$DEPLOY_PATH/cert"
mkdir -p "$DEPLOY_PATH/log"
chown -R ec2-user:ec2-user "$DEPLOY_PATH"

# Install certbot for TLS certificates
dnf install -y certbot

# Create systemd service
cat > /etc/systemd/system/nhp-relayd.service <<'EOF'
[Unit]
Description=NHP Relay Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/home/ec2-user/nhp-relay
ExecStart=/home/ec2-user/nhp-relay/nhp-relayd run
Restart=on-failure
RestartSec=5
LimitNOFILE=65536
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable nhp-relayd
