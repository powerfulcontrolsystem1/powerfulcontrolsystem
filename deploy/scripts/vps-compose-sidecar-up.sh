#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
COMPOSE_FILE="${COMPOSE_FILE:-$PROJECT_DIR/deploy/docker-compose.platform.yml}"
ENV_FILE="${ENV_FILE:-$PROJECT_DIR/deploy/.env.platform}"

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

test -f "$COMPOSE_FILE" || { echo "ERROR: no existe $COMPOSE_FILE"; exit 1; }
test -f "$ENV_FILE" || { echo "ERROR: no existe $ENV_FILE. Ejecuta primero vps-docker-preflight.sh"; exit 1; }

cd "$PROJECT_DIR"
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" config >/dev/null
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --build

http_bind="$(get_env_value "$ENV_FILE" HTTP_BIND)"
http_port="$(get_env_value "$ENV_FILE" HTTP_PORT)"
http_bind="${http_bind:-127.0.0.1}"
http_port="${http_port:-8081}"

echo "[sidecar] Contenedores:"
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps

echo "[sidecar] Probando frontend Docker en http://127.0.0.1:$http_port"
curl -fsSI "http://127.0.0.1:$http_port/" >/dev/null
echo "[sidecar] OK. Docker esta arriba en paralelo; Nginx publico aun puede seguir apuntando al servicio actual."
