#!/usr/bin/env bash
# Promueve a staging las mismas imagenes inmutables construidas y escaneadas
# por CI. No recompila codigo ni modifica el checkout que sirve produccion.
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "$0")/../.." && pwd)}"
STAGING_ENV="${STAGING_ENV:-$PROJECT_DIR/deploy/.env.staging}"
HTTP_PORT="${HTTP_PORT:-8082}"
COMPOSE_PROJECT_DIRECTORY="${COMPOSE_PROJECT_DIRECTORY:-$PROJECT_DIR/deploy}"
PLATFORM_COMPOSE_FILE="${PLATFORM_COMPOSE_FILE:-$PROJECT_DIR/deploy/docker-compose.platform.yml}"
STAGING_COMPOSE_FILE="${STAGING_COMPOSE_FILE:-$PROJECT_DIR/deploy/docker-compose.staging.yml}"
STAGING_ANTIVIRUS_COMPOSE_FILE="${STAGING_ANTIVIRUS_COMPOSE_FILE:-$PROJECT_DIR/deploy/docker-compose.staging-antivirus.yml}"
RELEASE_COMPOSE_FILE="${RELEASE_COMPOSE_FILE:-$PROJECT_DIR/deploy/docker-compose.release.yml}"

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
for compose_file in "$PLATFORM_COMPOSE_FILE" "$STAGING_COMPOSE_FILE" "$STAGING_ANTIVIRUS_COMPOSE_FILE" "$RELEASE_COMPOSE_FILE"; do
  [ -s "$compose_file" ] || fail "No existe el archivo Compose requerido: $compose_file"
done

require_digest PCS_API_IMAGE_DIGEST
require_digest PCS_MIGRATE_IMAGE_DIGEST
require_digest PCS_WORKER_IMAGE_DIGEST
require_digest PCS_FRONTEND_IMAGE_DIGEST
require_digest PCS_CLAMAV_IMAGE_DIGEST

compose=(
  docker compose
  --project-directory "$COMPOSE_PROJECT_DIRECTORY"
  --env-file "$STAGING_ENV"
  -f "$PLATFORM_COMPOSE_FILE"
  -f "$STAGING_COMPOSE_FILE"
  -f "$STAGING_ANTIVIRUS_COMPOSE_FILE"
  -f "$RELEASE_COMPOSE_FILE"
)

"${compose[@]}" config --quiet
configured_images="$("${compose[@]}" config --images)"
for image in \
  "$PCS_API_IMAGE_DIGEST" \
  "$PCS_MIGRATE_IMAGE_DIGEST" \
  "$PCS_WORKER_IMAGE_DIGEST" \
  "$PCS_FRONTEND_IMAGE_DIGEST" \
  "$PCS_CLAMAV_IMAGE_DIGEST"; do
  grep -Fqx "$image" <<<"$configured_images" ||
    fail "Compose no resolvio el digest exacto requerido: $image"
done

for image in \
  "$PCS_API_IMAGE_DIGEST" \
  "$PCS_MIGRATE_IMAGE_DIGEST" \
  "$PCS_WORKER_IMAGE_DIGEST" \
  "$PCS_FRONTEND_IMAGE_DIGEST" \
  "$PCS_CLAMAV_IMAGE_DIGEST"; do
  docker pull --quiet "$image" >/dev/null
done

"${compose[@]}" up -d --no-build postgres migrate clamav backend worker frontend

wait_for_endpoint "/health"
wait_for_endpoint "/ready"

echo "[OK] Candidato por digest activo en staging; no se reconstruyeron imagenes."
