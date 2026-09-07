#!/usr/bin/env bash
set -euo pipefail

DOMAIN="${DOMAIN:-staging.powerfulcontrolsystem.com}"
UPSTREAM="${UPSTREAM:-http://127.0.0.1:8082}"
CERT_NAME="${CERT_NAME:-pcs-staging}"
SITE="/etc/nginx/sites-available/pcs-staging"
ENABLED="/etc/nginx/sites-enabled/pcs-staging"
SNIPPET="/etc/nginx/snippets/pcs-security-headers.conf"
CERT_DIR="/etc/letsencrypt/live/$CERT_NAME"
CERT_FILE="$CERT_DIR/fullchain.pem"
CERT_KEY="$CERT_DIR/privkey.pem"

case "$CERT_NAME" in
  *[!A-Za-z0-9._-]*|"")
    echo "[ERROR] CERT_NAME invalido." >&2
    exit 1
    ;;
esac

if [ ! -s "$CERT_FILE" ] || [ ! -s "$CERT_KEY" ]; then
  echo "[ERROR] No existe certificado o llave de staging para CERT_NAME=$CERT_NAME." >&2
  exit 1
fi
if ! command -v openssl >/dev/null 2>&1 || ! openssl x509 -in "$CERT_FILE" -noout -checkend 0 >/dev/null; then
  echo "[ERROR] El certificado de staging esta vencido o no puede validarse." >&2
  exit 1
fi

mkdir -p /etc/nginx/snippets
cat > "$SNIPPET" <<'EOF'
add_header X-Frame-Options "SAMEORIGIN" always;
add_header Strict-Transport-Security "max-age=15552000; includeSubDomains" always;
EOF

if [ -f "$SITE" ]; then
  cp "$SITE" "$SITE.bak.$(date +%Y%m%d-%H%M%S)"
fi

cat > "$SITE" <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name $DOMAIN;

    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    location / {
        return 301 https://\$host\$request_uri;
    }
}

server {
    listen 443 ssl;
    listen [::]:443 ssl;
    server_name $DOMAIN;

    ssl_certificate $CERT_FILE;
    ssl_certificate_key $CERT_KEY;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
    include $SNIPPET;

    location / {
        proxy_pass $UPSTREAM;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$remote_addr;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 300;
        proxy_send_timeout 300;
    }
}
EOF

ln -sfn "$SITE" "$ENABLED"
nginx -t
systemctl reload nginx
echo "[OK] Nginx staging activo para $DOMAIN -> $UPSTREAM usando CERT_NAME=$CERT_NAME"
