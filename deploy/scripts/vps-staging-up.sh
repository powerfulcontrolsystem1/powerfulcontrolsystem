#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/../.."

ENV_FILE="${ENV_FILE:-deploy/.env.staging}"
if [ ! -f "$ENV_FILE" ]; then
  cp deploy/.env.staging.example "$ENV_FILE"
  chmod 600 "$ENV_FILE" || true
  echo "[WARN] Se creo $ENV_FILE desde la plantilla. Ajuste secretos antes de publicar staging."
fi

docker compose \
  --env-file "$ENV_FILE" \
  -f deploy/docker-compose.platform.yml \
  -f deploy/docker-compose.staging.yml \
  config --quiet

docker compose \
  --env-file "$ENV_FILE" \
  -f deploy/docker-compose.platform.yml \
  -f deploy/docker-compose.staging.yml \
  up -d --build

echo "[OK] Staging levantado en 127.0.0.1:${HTTP_PORT:-8082}"
