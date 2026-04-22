#!/bin/bash
set -euo pipefail

DEPLOY_PATH="${deploy_path}"

# Create deploy directory
mkdir -p "$DEPLOY_PATH/etc"
mkdir -p "$DEPLOY_PATH/cert"
mkdir -p "$DEPLOY_PATH/log"
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
User=root
WorkingDirectory=/home/ec2-user/nhp-ac
ExecStart=/home/ec2-user/nhp-ac/nhp-acd run
Restart=on-failure
RestartSec=5
LimitNOFILE=65536
AmbientCapabilities=CAP_NET_ADMIN

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable nhp-acd
