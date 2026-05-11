#!/usr/bin/env bash
set -euo pipefail

DOMAIN="${DOMAIN:-staging.powerfulcontrolsystem.com}"
UPSTREAM="${UPSTREAM:-http://127.0.0.1:8082}"
SITE="/etc/nginx/sites-available/pcs-staging"
ENABLED="/etc/nginx/sites-enabled/pcs-staging"
SNIPPET="/etc/nginx/snippets/pcs-security-headers.conf"

mkdir -p /etc/nginx/snippets
cat > "$SNIPPET" <<'EOF'
add_header X-Content-Type-Options "nosniff" always;
add_header X-Frame-Options "SAMEORIGIN" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;
add_header Strict-Transport-Security "max-age=15552000; includeSubDomains" always;
add_header Content-Security-Policy-Report-Only "default-src 'self' https: data: blob: 'unsafe-inline' 'unsafe-eval'; report-uri /api/public/security/csp-report" always;
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

    ssl_certificate /etc/letsencrypt/live/powerfulcontrolsystem.com-0001/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/powerfulcontrolsystem.com-0001/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;
    include $SNIPPET;

    location / {
        proxy_pass $UPSTREAM;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
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
echo "[OK] Nginx staging activo para $DOMAIN -> $UPSTREAM"
