#!/usr/bin/env bash
# Promueve a staging las mismas imagenes inmutables construidas y escaneadas
# por CI. No recompila codigo ni modifica el checkout que sirve produccion.
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "$0")/../.." && pwd)}"
STAGING_ENV="${STAGING_ENV:-$PROJECT_DIR/deploy/.env.staging}"
HTTP_PORT="${HTTP_PORT:-8082}"

fail() {
  echo "[ERROR] $*" >&2
  exit 1
}

require_digest() {
  local name="$1"
  local value="${!name:-}"
  [[ "$value" =~ ^[^[:space:]@]+@sha256:[a-f0-9]{64}$ ]] ||
    fail "$name debe usar repositorio@sha256:<64 hex>."
}

wait_for_endpoint() {
  local endpoint="$1"
  local attempt
  for attempt in {1..12}; do
    if curl -fsS --max-time 10 "http://127.0.0.1:${HTTP_PORT}${endpoint}" >/dev/null; then
      return 0
    fi
    sleep 5
  done
  fail "Staging no respondio ${endpoint} despues de promover el digest."
}

command -v docker >/dev/null 2>&1 || fail "Docker no esta disponible."
command -v curl >/dev/null 2>&1 || fail "curl no esta disponible."
[ -f "$STAGING_ENV" ] || fail "No existe el entorno privado de staging."

require_digest PCS_API_IMAGE_DIGEST
require_digest PCS_MIGRATE_IMAGE_DIGEST
require_digest PCS_WORKER_IMAGE_DIGEST
require_digest PCS_FRONTEND_IMAGE_DIGEST

for image in \
  "$PCS_API_IMAGE_DIGEST" \
  "$PCS_MIGRATE_IMAGE_DIGEST" \
  "$PCS_WORKER_IMAGE_DIGEST" \
  "$PCS_FRONTEND_IMAGE_DIGEST"; do
  docker pull --quiet "$image" >/dev/null
done

compose=(
  docker compose
  --env-file "$STAGING_ENV"
  -f "$PROJECT_DIR/deploy/docker-compose.platform.yml"
  -f "$PROJECT_DIR/deploy/docker-compose.staging.yml"
  -f "$PROJECT_DIR/deploy/docker-compose.release.yml"
)

"${compose[@]}" config --quiet
"${compose[@]}" up -d --no-build

wait_for_endpoint "/health"
wait_for_endpoint "/ready"

echo "[OK] Candidato por digest activo en staging; no se reconstruyeron imagenes."
