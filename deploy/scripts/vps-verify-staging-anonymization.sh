#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
STAGING_ENV="${STAGING_ENV:-$PROJECT_DIR/deploy/.env.staging}"

cd "$PROJECT_DIR"

if [ ! -f "$STAGING_ENV" ]; then
  echo "[ERROR] No existe $STAGING_ENV"
  exit 1
fi

set -a
# shellcheck disable=SC1090
. "$STAGING_ENV"
set +a

DB_URL="${DB_SUPERADMIN_DSN:-${DATABASE_SUPERADMIN_URL:-}}"
if [ -z "$DB_URL" ]; then
  echo "[ERROR] No se encontro DB_SUPERADMIN_DSN/DATABASE_SUPERADMIN_URL en staging."
  exit 1
fi

psql "$DB_URL" -v ON_ERROR_STOP=1 <<'SQL'
SELECT 'administradores_emails_reales' AS check_name, COUNT(*) AS findings
FROM administradores
WHERE COALESCE(email, '') NOT LIKE '%@staging.local'
  AND COALESCE(email, '') NOT LIKE '%@example.test'
  AND COALESCE(email, '') <> '';
SQL

echo "[OK] Verificacion de anonimizacion ejecutada. El conteo debe ser 0 antes de abrir staging a terceros."
