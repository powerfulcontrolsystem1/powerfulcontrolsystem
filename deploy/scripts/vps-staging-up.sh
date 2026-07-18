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

secret_base64_32() {
  openssl rand 32 2>/dev/null | base64 | tr -d '\n'
}

valid_config_enc_key() {
  value="$1"
  [ -n "$value" ] || return 1
  decoded_size=$(printf '%s' "$value" | base64 -d 2>/dev/null | wc -c | tr -d ' ')
  [ "$decoded_size" = "32" ] || return 1
  canonical=$(printf '%s' "$value" | base64 -d 2>/dev/null | base64 | tr -d '\n')
  [ "$canonical" = "$value" ]
}

replace_if_placeholder() {
  key="$1"
  value="$2"
  if grep -q "^${key}=CAMBIAR" "$ENV_FILE" 2>/dev/null \
    || grep -q "^${key}=$" "$ENV_FILE" 2>/dev/null \
    || grep -Eq "^${key}=<[^>]+>$" "$ENV_FILE" 2>/dev/null \
    || { [ "$key" = "CONFIG_ENC_KEY" ] && ! valid_config_enc_key "$(sed -n "s/^${key}=//p" "$ENV_FILE" | head -n 1)"; }; then
    sed -i "s|^${key}=.*|${key}=${value}|" "$ENV_FILE"
  fi
}

replace_if_placeholder "POSTGRES_PASSWORD" "$(secret_hex 24)"
replace_if_placeholder "CONFIG_ENC_KEY" "$(secret_base64_32)"
replace_if_placeholder "ONLYOFFICE_JWT_SECRET" "$(secret_hex 24)"
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
