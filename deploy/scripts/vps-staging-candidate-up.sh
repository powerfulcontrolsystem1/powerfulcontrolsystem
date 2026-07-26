#!/usr/bin/env bash
# Deploy a reviewed commit to the isolated staging Compose project without
# changing the checkout that serves production. It intentionally requires both
# a branch/ref and an immutable SHA.
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "$0")/../.." && pwd)}"
CANDIDATE_REF="${CANDIDATE_REF:-}"
CANDIDATE_SHA="${CANDIDATE_SHA:-}"
WORKTREE_DIR="${WORKTREE_DIR:-/root/powerfulcontrolsystem-staging-candidate}"
STAGING_ENV="${STAGING_ENV:-$PROJECT_DIR/deploy/.env.staging}"
SOURCE_REPOSITORY="${SOURCE_REPOSITORY:-https://github.com/powerfulcontrolsystem1/powerfulcontrolsystem.git}"
HTTP_PORT="${HTTP_PORT:-8082}"

fail() {
  echo "[ERROR] $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "Falta el comando requerido: $1"
}

[ -n "$CANDIDATE_REF" ] || fail "Defina CANDIDATE_REF con la rama o referencia candidata."
[ -n "$CANDIDATE_SHA" ] || fail "Defina CANDIDATE_SHA con el commit inmutable aprobado."
[ "${RESET_STAGING:-0}" = "0" ] || fail "RESET_STAGING no está permitido para un candidato P108."

require_command git
require_command docker
require_command curl

PROJECT_DIR="$(cd "$PROJECT_DIR" && pwd)"
if [ "$WORKTREE_DIR" = "$PROJECT_DIR" ]; then
  fail "WORKTREE_DIR debe ser distinto al checkout de producción."
fi

[ -f "$STAGING_ENV" ] || fail "No existe el archivo de entorno staging indicado."

if [ -d "$PROJECT_DIR/.git" ]; then
  cd "$PROJECT_DIR"
  git fetch --quiet origin "$CANDIDATE_REF"
  resolved_sha="$(git rev-parse FETCH_HEAD)"
  if [ "$resolved_sha" != "$CANDIDATE_SHA" ]; then
    fail "La referencia candidata no coincide con el SHA aprobado."
  fi
  if [ -e "$WORKTREE_DIR" ]; then
    [ -d "$WORKTREE_DIR" ] || fail "WORKTREE_DIR existe y no es un directorio."
    existing_sha="$(git -C "$WORKTREE_DIR" rev-parse HEAD 2>/dev/null || true)"
    [ "$existing_sha" = "$CANDIDATE_SHA" ] || fail "WORKTREE_DIR ya contiene otro commit; revíselo o elimínelo manualmente."
  else
    git worktree add --detach "$WORKTREE_DIR" "$CANDIDATE_SHA"
  fi
else
  # VPS deployments commonly receive a source package without .git. Clone only
  # the reviewed candidate into the isolated staging directory; never convert
  # or modify the production package.
  if [ -e "$WORKTREE_DIR" ]; then
    [ -d "$WORKTREE_DIR" ] || fail "WORKTREE_DIR existe y no es un directorio."
    existing_sha="$(git -C "$WORKTREE_DIR" rev-parse HEAD 2>/dev/null || true)"
    [ "$existing_sha" = "$CANDIDATE_SHA" ] || fail "WORKTREE_DIR ya contiene otro commit; revíselo o elimínelo manualmente."
  else
    git clone --no-checkout "$SOURCE_REPOSITORY" "$WORKTREE_DIR"
    git -C "$WORKTREE_DIR" fetch --quiet origin "$CANDIDATE_REF"
    resolved_sha="$(git -C "$WORKTREE_DIR" rev-parse FETCH_HEAD)"
    [ "$resolved_sha" = "$CANDIDATE_SHA" ] || fail "La referencia candidata clonada no coincide con el SHA aprobado."
    git -C "$WORKTREE_DIR" checkout --detach "$CANDIDATE_SHA"
  fi
fi

compose=(docker compose --env-file "$STAGING_ENV"
  -f "$WORKTREE_DIR/deploy/docker-compose.platform.yml"
  -f "$WORKTREE_DIR/deploy/docker-compose.staging.yml")

"${compose[@]}" config --quiet
"${compose[@]}" up -d --build

if ! curl -fsS --max-time 45 "http://127.0.0.1:${HTTP_PORT}/health" >/dev/null; then
  fail "Staging no respondió health después de construir el candidato."
fi
if ! curl -fsS --max-time 45 "http://127.0.0.1:${HTTP_PORT}/ready" >/dev/null; then
  fail "Staging no respondió ready después de construir el candidato."
fi

echo "[OK] Candidato staging activo. SHA=${CANDIDATE_SHA} worktree=${WORKTREE_DIR} puerto=${HTTP_PORT}"
