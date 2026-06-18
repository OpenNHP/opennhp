#!/bin/bash
set -euo pipefail

DEPLOY_PATH="${deploy_path}"

# Create deploy directory
mkdir -p "$DEPLOY_PATH/etc"
mkdir -p "$DEPLOY_PATH/plugins"
mkdir -p "$DEPLOY_PATH/cert"
# nhp-serverd writes to ExeDirPath/logs (plural); see endpoints/server/udpserver.go.
mkdir -p "$DEPLOY_PATH/logs"
chown -R ec2-user:ec2-user "$DEPLOY_PATH"

# Install certbot for TLS certificates
dnf install -y certbot

# Create systemd service
cat > /etc/systemd/system/nhp-serverd.service <<'EOF'
[Unit]
Description=NHP Server Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/home/ec2-user/nhp-server
ExecStart=/home/ec2-user/nhp-server/nhp-serverd run
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable nhp-serverd
