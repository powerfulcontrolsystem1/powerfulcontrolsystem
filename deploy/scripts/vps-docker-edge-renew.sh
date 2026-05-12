#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
COMPOSE_FILE="${COMPOSE_FILE:-$PROJECT_DIR/deploy/docker-compose.platform.yml}"
ENV_FILE="${ENV_FILE:-$PROJECT_DIR/deploy/.env.platform}"

test -f "$COMPOSE_FILE" || { echo "ERROR: no existe $COMPOSE_FILE"; exit 1; }
test -f "$ENV_FILE" || { echo "ERROR: no existe $ENV_FILE"; exit 1; }

cd "$PROJECT_DIR"

docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" --profile certbot run --rm certbot renew \
  --webroot \
  --webroot-path /var/www/certbot \
  --quiet

if docker ps --format '{{.Names}}' | grep -qx 'pcs-edge'; then
  docker exec pcs-edge nginx -s reload
fi

echo "[edge-renew] Renovacion revisada y Nginx Docker recargado."
