#!/bin/bash
# Install certbot + nginx, obtain Let's Encrypt certificate via Cloudflare DNS-01,
# and install an nginx vhost. Idempotent: safe to re-run.
#
# Expected env vars (set by caller):
#   COMPONENT        = "server" | "ac" | "relay"
#   PRIMARY_DOMAIN   = e.g. "demologin.opennhp.org"
#   EXTRA_DOMAINS    = space-separated additional SANs (may be empty)
#   CF_API_TOKEN     = Cloudflare API token with Zone:DNS:Edit + Zone:Zone:Read
#   ACME_EMAIL       = contact email for Let's Encrypt
#   NGINX_CONF       = path to rendered nginx vhost file on this host

set -euo pipefail

: "${COMPONENT:?}"
: "${PRIMARY_DOMAIN:?}"
: "${CF_API_TOKEN:?}"
: "${ACME_EMAIL:?}"
: "${NGINX_CONF:?}"
EXTRA_DOMAINS="${EXTRA_DOMAINS:-}"

echo "[tls] component=$COMPONENT primary=$PRIMARY_DOMAIN extras='$EXTRA_DOMAINS'"

# --- Install prerequisites (idempotent) ---
# nginx comes from dnf; certbot + cloudflare plugin via venv because
# Amazon Linux 2023 doesn't ship python3-certbot-dns-cloudflare.
if ! rpm -q nginx >/dev/null 2>&1; then
    echo "[tls] installing nginx"
    sudo dnf install -y nginx
fi

CERTBOT_VENV=/opt/certbot
CERTBOT_BIN=$CERTBOT_VENV/bin/certbot
if [ ! -x "$CERTBOT_BIN" ]; then
    echo "[tls] installing certbot + dns-cloudflare plugin into $CERTBOT_VENV"
    sudo dnf install -y python3 python3-pip
    sudo python3 -m venv "$CERTBOT_VENV"
    sudo "$CERTBOT_VENV/bin/pip" install --upgrade pip
    sudo "$CERTBOT_VENV/bin/pip" install certbot certbot-dns-cloudflare
fi
# Expose as /usr/local/bin/certbot so systemd timer and other callers find it
if [ ! -L /usr/local/bin/certbot ]; then
    sudo ln -sf "$CERTBOT_BIN" /usr/local/bin/certbot
fi

# --- Write Cloudflare credentials (0600) ---
sudo install -d -m 700 /etc/letsencrypt
CF_INI=/etc/letsencrypt/cloudflare.ini
sudo tee "$CF_INI" >/dev/null <<EOF
dns_cloudflare_api_token = $CF_API_TOKEN
EOF
sudo chmod 600 "$CF_INI"

# --- Obtain certificate (skip if current cert is still valid for >30 days) ---
CERT_DIR="/etc/letsencrypt/live/$PRIMARY_DOMAIN"
NEED_ISSUE=1
if [ -f "$CERT_DIR/fullchain.pem" ]; then
    if sudo openssl x509 -checkend $((30 * 86400)) -noout -in "$CERT_DIR/fullchain.pem" >/dev/null 2>&1; then
        echo "[tls] existing certificate for $PRIMARY_DOMAIN still valid >30d, skipping issue"
        NEED_ISSUE=0
    fi
fi

if [ "$NEED_ISSUE" = "1" ]; then
    D_ARGS="-d $PRIMARY_DOMAIN"
    for d in $EXTRA_DOMAINS; do
        D_ARGS="$D_ARGS -d $d"
    done

    echo "[tls] requesting certificate: $D_ARGS"
    sudo "$CERTBOT_BIN" certonly \
        --non-interactive --agree-tos \
        --email "$ACME_EMAIL" \
        --dns-cloudflare --dns-cloudflare-credentials "$CF_INI" \
        --dns-cloudflare-propagation-seconds 30 \
        --keep-until-expiring \
        $D_ARGS
fi

# --- Install nginx vhost ---
VHOST_DST="/etc/nginx/conf.d/${COMPONENT}.conf"
sudo install -m 644 "$NGINX_CONF" "$VHOST_DST"

# Disable distro's default server block if present (it also listens on :80)
if [ -f /etc/nginx/nginx.conf ]; then
    if grep -q "default_server" /etc/nginx/nginx.conf; then
        sudo sed -i 's/listen\s*80\s*default_server;/# & # disabled by bootstrap-tls.sh/g; s/listen\s*\[::\]:80\s*default_server;/# & # disabled by bootstrap-tls.sh/g' /etc/nginx/nginx.conf
    fi
fi

sudo nginx -t
sudo systemctl enable nginx
sudo systemctl reload nginx 2>/dev/null || sudo systemctl restart nginx

# --- Deploy hook: re-install and reload on renewal ---
HOOK=/etc/letsencrypt/renewal-hooks/deploy/reload-nginx.sh
sudo install -d -m 755 "$(dirname "$HOOK")"
sudo tee "$HOOK" >/dev/null <<'EOF'
#!/bin/bash
systemctl reload nginx
EOF
sudo chmod +x "$HOOK"

# --- Renewal timer ---
# The venv certbot is the source of truth (it has the cloudflare plugin).
# Install a dedicated systemd timer that invokes it daily.
sudo tee /etc/systemd/system/certbot-venv.service >/dev/null <<EOF
[Unit]
Description=Let's Encrypt renewal via certbot venv
After=network-online.target

[Service]
Type=oneshot
ExecStart=$CERTBOT_BIN renew --quiet
EOF

sudo tee /etc/systemd/system/certbot-venv.timer >/dev/null <<'EOF'
[Unit]
Description=Run certbot renewal twice daily

[Timer]
OnCalendar=*-*-* 03,15:00:00
RandomizedDelaySec=3600
Persistent=true

[Install]
WantedBy=timers.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now certbot-venv.timer

# Disable the dnf-provided timer if present (it would call system certbot
# which lacks the cloudflare plugin)
sudo systemctl disable --now certbot-renew.timer 2>/dev/null || true

echo "[tls] done"
