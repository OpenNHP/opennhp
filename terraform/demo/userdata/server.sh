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
#
# nhp-serverd attaches an XDP ingress filter (nhp/ebpf/xdp/nhp_server_xdp.c)
# that narrows this host down to the NHP knock port plus SSH from the relay.
# That needs three capabilities and a writable bpffs:
#
#   CAP_BPF         load the program and create/pin its maps
#   CAP_NET_ADMIN   attach the program to the interface
#   CAP_PERFMON     open the perf ring the accept/drop event log reads
#
# Granted as ambient caps so the unprivileged user inherits them across exec,
# and bounded so the process cannot pick up more at runtime. No CAP_SYS_ADMIN
# and no CAP_DAC_OVERRIDE: as ambient capabilities on an internet-facing
# daemon both are effectively root, and NoNewPrivileges does not mitigate
# either. Instead the ExecStartPre below runs as root (systemd's "+" prefix)
# to make /sys/fs/bpf group-writable for just this user — systemd's
# sys-fs-bpf.mount leaves it 0700 root:root, which the daemon cannot pin into.
#
# The three that are granted are not held for long. nhp-serverd gives them
# back as the last step of attaching the filter (dropLoaderPrivileges in
# nhp/utils/ebpf/caps_linux.go): ambient cleared, permitted and effective
# emptied — or reduced to CAP_BPF alone where kernel.unprivileged_bpf_disabled
# makes bpf(2) privileged and the xdp.toml reload path still has a map to
# write. That matters because this process parses untrusted UDP from the whole
# internet and dlopens auth plugins into its own address space, and CAP_BPF
# with CAP_PERFMON loads tracing programs that can read arbitrary kernel
# memory. systemd grants what the load needs; the daemon decides what it keeps.
# The deploy job asserts the drop actually happened by reading
# /proc/<pid>/status.
#
# Missing any of this is not fatal: the daemon logs a warning and runs with no
# ingress filter (fail-open), which keeps a kernel or permission problem from
# taking the gateway off the air. The deploy-server job in
# .github/workflows/deploy-demo-v2.yml installs the same settings as a drop-in
# for hosts created before this file changed, and then fails the run if no XDP
# program ended up attached — keep the two in sync.
cat > /etc/systemd/system/nhp-serverd.service <<'EOF'
[Unit]
Description=NHP Server Daemon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/home/ec2-user/nhp-server
ExecStartPre=+/bin/sh -c 'mountpoint -q /sys/fs/bpf || mount -t bpf bpf /sys/fs/bpf; chgrp ec2-user /sys/fs/bpf; chmod 0770 /sys/fs/bpf'
ExecStart=/home/ec2-user/nhp-server/nhp-serverd run
Restart=on-failure
RestartSec=5
LimitNOFILE=65536
AmbientCapabilities=CAP_BPF CAP_NET_ADMIN CAP_PERFMON
CapabilityBoundingSet=CAP_BPF CAP_NET_ADMIN CAP_PERFMON
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable nhp-serverd
