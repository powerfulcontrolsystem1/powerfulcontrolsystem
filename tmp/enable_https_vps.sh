set -e
export DEBIAN_FRONTEND=noninteractive
apt-get update -y >/dev/null 2>&1 || true
apt-get install -y certbot python3-certbot-nginx >/dev/null 2>&1
certbot --nginx -d powerfulcontrolsystem.com --non-interactive --agree-tos --register-unsafely-without-email --redirect
nginx -t
systemctl reload nginx
ss -ltnp | grep -E ':443 ' || true
