set -e
echo '--- ports ---'
ss -ltnp | grep -E ':80 |:443 ' || true
echo '--- nginx ---'
systemctl is-active nginx || true
nginx -t 2>&1 || true
echo '--- letsencrypt ---'
ls -la /etc/letsencrypt/live 2>/dev/null || true
echo '--- site conf ---'
ls -la /etc/nginx/sites-enabled 2>/dev/null || true
