#!/bin/bash
# Install demo.nhp certificate, key, and nginx config with idempotent checks.
# Only reloads nginx when any artifact actually changes.
#
# Usage: install-demo-nhp-cert.sh
#
# Expected input files (must exist before calling):
#   /tmp/demo.nhp.pem.new      - Certificate PEM
#   /tmp/demo.nhp-key.pem.new  - Private key PEM
#   /tmp/ac-demo-nhp.conf      - Nginx config (optional)
#
# The script cleans up temp files on exit.

set -euo pipefail

CERT_TMP="/tmp/demo.nhp.pem.new"
KEY_TMP="/tmp/demo.nhp-key.pem.new"
CONF_TMP="/tmp/ac-demo-nhp.conf"

CERT_DST="/etc/nginx/certs/demo.nhp.pem"
KEY_DST="/etc/nginx/certs/demo.nhp-key.pem"
CONF_DST="/etc/nginx/conf.d/ac-demo-nhp.conf"

# Cleanup temp files on exit
cleanup() {
    rm -f "$CERT_TMP" "$KEY_TMP" "$CONF_TMP"
}
trap cleanup EXIT

# Ensure certs directory exists with restrictive permissions.
# nginx master reads cert/key as root before dropping privileges to workers,
# so 750 root:root is sufficient and limits exposure if the host is compromised.
sudo mkdir -p /etc/nginx/certs
sudo chmod 750 /etc/nginx/certs

reload_nginx=0

# Install certificate if changed
if [ -f "$CERT_TMP" ]; then
    if ! sudo test -f "$CERT_DST" || \
       ! sudo cmp -s "$CERT_TMP" "$CERT_DST"; then
        sudo install -m 644 "$CERT_TMP" "$CERT_DST"
        echo "demo.nhp certificate installed"
        reload_nginx=1
    else
        echo "demo.nhp certificate unchanged"
    fi
fi

# Install private key if changed
if [ -f "$KEY_TMP" ]; then
    if ! sudo test -f "$KEY_DST" || \
       ! sudo cmp -s "$KEY_TMP" "$KEY_DST"; then
        sudo install -m 600 "$KEY_TMP" "$KEY_DST"
        echo "demo.nhp private key installed"
        reload_nginx=1
    else
        echo "demo.nhp private key unchanged"
    fi
fi

# Install nginx config if changed
if [ -f "$CONF_TMP" ]; then
    if ! sudo test -f "$CONF_DST" || \
       ! sudo cmp -s "$CONF_TMP" "$CONF_DST"; then
        sudo install -m 644 "$CONF_TMP" "$CONF_DST"
        echo "ac-demo-nhp.conf installed"
        reload_nginx=1
    else
        echo "ac-demo-nhp.conf unchanged"
    fi
fi

# Reload nginx only if something changed
if [ "$reload_nginx" -eq 1 ]; then
    sudo nginx -t
    sudo systemctl reload nginx
    echo "nginx reloaded"
else
    echo "demo.nhp artifacts unchanged; skipping nginx reload"
fi
