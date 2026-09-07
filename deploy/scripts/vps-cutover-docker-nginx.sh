#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
ENV_FILE="${ENV_FILE:-$PROJECT_DIR/deploy/.env.platform}"
NGINX_SITE="${NGINX_SITE:-/etc/nginx/sites-available/powerfulcontrolsystem}"

if [ "${CONFIRM_CUTOVER:-}" != "YES" ]; then
  echo "Este script conmuta Nginx del host desde 127.0.0.1:8080 hacia Docker."
  echo "Ejecuta con CONFIRM_CUTOVER=YES despues de validar http://127.0.0.1:8081 y la migracion de datos."
  exit 2
fi

get_env_value() {
  local file="$1"
  local key="$2"
  [ -f "$file" ] || return 0
  awk -v key="$key" '
    {
      line=$0
      sub(/^[ \t]+/, "", line)
      if (index(line, key "=") == 1) {
        sub(/^[^=]*=/, "", line)
        gsub(/^["'\'']|["'\'']$/, "", line)
        value=line
      }
    }
    END { if (value != "") print value }
  ' "$file"
}

http_port="$(get_env_value "$ENV_FILE" HTTP_PORT)"
http_port="${http_port:-8081}"

curl -fsSI "http://127.0.0.1:$http_port/" >/dev/null

test -f "$NGINX_SITE" || { echo "ERROR: no existe $NGINX_SITE"; exit 1; }
backup="$NGINX_SITE.bak.$(date +%Y%m%d-%H%M%S)"
cp "$NGINX_SITE" "$backup"
echo "[cutover] Backup Nginx: $backup"

sed -i "s|127\.0\.0\.1:8080|127.0.0.1:$http_port|g; s|localhost:8080|127.0.0.1:$http_port|g" "$NGINX_SITE"

if ! nginx -t; then
  cp "$backup" "$NGINX_SITE"
  nginx -t
  echo "[cutover] nginx -t fallo; se restauro el backup."
  exit 1
fi

systemctl reload nginx
echo "[cutover] OK. Nginx ahora apunta a Docker en 127.0.0.1:$http_port."
echo "[cutover] El servicio systemd anterior queda activo para rollback rapido."
