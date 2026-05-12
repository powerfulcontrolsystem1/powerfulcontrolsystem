#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

ENV_FILE="${ENV_FILE:-deploy/.env.staging}"
if [ ! -f "$ENV_FILE" ]; then
  cp deploy/.env.staging.example "$ENV_FILE"
  chmod 600 "$ENV_FILE" || true
  echo "[INFO] Se creo $ENV_FILE desde la plantilla."
fi

secret_hex() {
  openssl rand -hex "${1:-24}" 2>/dev/null || date +%s%N | sha256sum | awk '{print $1}'
}

replace_if_placeholder() {
  key="$1"
  value="$2"
  if grep -q "^${key}=CAMBIAR" "$ENV_FILE" 2>/dev/null || grep -q "^${key}=$" "$ENV_FILE" 2>/dev/null; then
    sed -i "s|^${key}=.*|${key}=${value}|" "$ENV_FILE"
  fi
}

replace_if_placeholder "POSTGRES_PASSWORD" "$(secret_hex 24)"
replace_if_placeholder "CONFIG_ENC_KEY" "$(secret_hex 32)"
replace_if_placeholder "ONLYOFFICE_JWT_SECRET" "$(secret_hex 24)"
replace_if_placeholder "NEXTCLOUD_DB_PASSWORD" "$(secret_hex 24)"
replace_if_placeholder "NEXTCLOUD_ADMIN_PASSWORD" "$(secret_hex 24)"
chmod 600 "$ENV_FILE" || true

docker compose \
  --env-file "$ENV_FILE" \
  -f deploy/docker-compose.platform.yml \
  -f deploy/docker-compose.staging.yml \
  config --quiet

if [ "${RESET_STAGING:-0}" = "1" ]; then
  echo "[WARN] RESET_STAGING=1: se eliminaran contenedores y volumenes de staging."
  docker compose \
    --env-file "$ENV_FILE" \
    -f deploy/docker-compose.platform.yml \
    -f deploy/docker-compose.staging.yml \
    down -v
fi

docker compose \
  --env-file "$ENV_FILE" \
  -f deploy/docker-compose.platform.yml \
  -f deploy/docker-compose.staging.yml \
  up -d --build

echo "[OK] Staging levantado en 127.0.0.1:${HTTP_PORT:-8082}"
