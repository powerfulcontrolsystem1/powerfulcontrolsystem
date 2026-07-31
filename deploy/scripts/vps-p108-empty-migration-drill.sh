#!/usr/bin/env bash
# Ensaya el catálogo de migraciones sobre PostgreSQL vacío y una segunda pasada.
# Todos los recursos Docker son aislados, efímeros y se eliminan al terminar.
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "$0")/../.." && pwd)}"
SOURCE_ENV="${SOURCE_ENV:-$PROJECT_DIR/deploy/.env.staging}"
PCS_MIGRATE_IMAGE_DIGEST="${PCS_MIGRATE_IMAGE_DIGEST:-}"
P108_DRILL_ID="${P108_DRILL_ID:-p108-empty-$(date +%s)}"
P108_EMPTY_SCHEMA_BOOTSTRAP="${P108_EMPTY_SCHEMA_BOOTSTRAP:-1}"

fail() {
  echo "[ERROR] $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "Falta el comando requerido: $1"
}

[[ "$P108_DRILL_ID" =~ ^p108-empty-[a-z0-9-]+$ ]] ||
  fail "P108_DRILL_ID debe iniciar con p108-empty- y usar solo minúsculas, números o guiones."
[[ "$PCS_MIGRATE_IMAGE_DIGEST" =~ ^[^[:space:]@]+@sha256:[a-f0-9]{64}$ ]] ||
  fail "PCS_MIGRATE_IMAGE_DIGEST debe usar repositorio@sha256:<64 hex>."
[ "$P108_EMPTY_SCHEMA_BOOTSTRAP" = "1" ] ||
  fail "Una base vacía exige P108_EMPTY_SCHEMA_BOOTSTRAP=1 y solo el rol migrate puede usarlo."
[ -f "$SOURCE_ENV" ] || fail "No existe SOURCE_ENV."

require_command docker
require_command openssl

network="${P108_DRILL_ID}-net"
volume="${P108_DRILL_ID}-pgdata"
postgres="${P108_DRILL_ID}-postgres"
migrate_first="${P108_DRILL_ID}-migrate-1"
migrate_second="${P108_DRILL_ID}-migrate-2"
migrate_drift="${P108_DRILL_ID}-migrate-drift"
migrate_recovered="${P108_DRILL_ID}-migrate-recovered"

for resource in "$postgres" "$migrate_first" "$migrate_second" "$migrate_drift" "$migrate_recovered"; do
  if docker container inspect "$resource" >/dev/null 2>&1; then
    fail "El contenedor aislado ya existe y no será sobrescrito: $resource"
  fi
done
docker volume inspect "$volume" >/dev/null 2>&1 &&
  fail "El volumen aislado ya existe y no será sobrescrito: $volume"
docker network inspect "$network" >/dev/null 2>&1 &&
  fail "La red aislada ya existe y no será sobrescrita: $network"

cleanup() {
  docker rm -f "$migrate_first" "$migrate_second" "$migrate_drift" "$migrate_recovered" "$postgres" >/dev/null 2>&1 || true
  docker volume rm "$volume" >/dev/null 2>&1 || true
  docker network rm "$network" >/dev/null 2>&1 || true
}
trap cleanup EXIT

test_password="$(openssl rand -hex 24)"
docker network create "$network" >/dev/null
docker volume create "$volume" >/dev/null
docker run -d \
  --name "$postgres" \
  --network "$network" \
  -e POSTGRES_USER=pcs \
  -e POSTGRES_PASSWORD="$test_password" \
  -e POSTGRES_DB=pcs_superadministrador \
  -v "$volume:/var/lib/postgresql/data" \
  -v "$PROJECT_DIR/deploy/postgres/init:/docker-entrypoint-initdb.d:ro" \
  postgres:16.14-alpine >/dev/null

for attempt in $(seq 1 30); do
  if docker exec "$postgres" pg_isready -U pcs -d pcs_superadministrador >/dev/null 2>&1; then
    break
  fi
  [ "$attempt" -lt 30 ] || fail "PostgreSQL aislado no quedó listo."
  sleep 2
done

run_migrate() {
  local name="$1"
  docker run \
    --name "$name" \
    --network "$network" \
    --env-file "$SOURCE_ENV" \
    -e PCS_ENV=production \
    -e PCS_RUNTIME_ROLE=migrate \
    -e "PCS_RUNTIME_SCHEMA_BOOTSTRAP=$P108_EMPTY_SCHEMA_BOOTSTRAP" \
    -e PCS_SKIP_CORPORATE_EMAIL_STARTUP_SYNC=1 \
    -e PCS_MIGRATION_SKIP_EXTERNAL_SYNC=1 \
    -e DB_DIALECT=postgres \
    -e "DB_EMPRESAS_DSN=postgres://pcs:${test_password}@${postgres}:5432/pcs_empresas?sslmode=disable" \
    -e "DB_SUPERADMIN_DSN=postgres://pcs:${test_password}@${postgres}:5432/pcs_superadministrador?sslmode=disable" \
    --entrypoint /bin/sh \
    "$PCS_MIGRATE_IMAGE_DIGEST" \
    -ec '/app/backend/pcs-backend && /app/backend/pcs-migrate'
}

echo "[RUN] Migración sobre base vacía."
run_migrate "$migrate_first"
echo "[RUN] Segunda pasada idempotente."
run_migrate "$migrate_second"

echo "[RUN] Fallo cerrado por drift de checksum y recuperación controlada."
migration_row="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -AtF '|' -c \
  "SELECT version, checksum
     FROM schema_migrations
    WHERE scope = 'platform' AND btrim(COALESCE(checksum, '')) <> ''
    ORDER BY version DESC
    LIMIT 1;")"
IFS='|' read -r drift_version original_checksum <<< "$migration_row"
[[ "$drift_version" =~ ^[0-9a-z-]+$ ]] || fail "No se pudo seleccionar una migración controlada para el ensayo de drift."
[[ "$original_checksum" =~ ^[a-f0-9]{64}$ ]] || fail "El checksum original no tiene el formato esperado."

tables_before_drift="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
  "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")"
migrations_before_drift="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
  "SELECT COUNT(*) FROM schema_migrations;")"

docker exec "$postgres" psql -v ON_ERROR_STOP=1 -U pcs -d pcs_empresas -c \
  "UPDATE schema_migrations
      SET checksum = repeat('0', 64)
    WHERE scope = 'platform' AND version = '$drift_version';" >/dev/null

if run_migrate "$migrate_drift"; then
  fail "El migrador aceptó un checksum alterado."
fi

tables_after_drift="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
  "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")"
migrations_after_drift="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
  "SELECT COUNT(*) FROM schema_migrations;")"
[ "$tables_before_drift" = "$tables_after_drift" ] ||
  fail "El fallo por drift modificó la cantidad de tablas."
[ "$migrations_before_drift" = "$migrations_after_drift" ] ||
  fail "El fallo por drift modificó el ledger aplicado."

failed_runs="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
  "SELECT COUNT(*)
     FROM schema_migration_runs
    WHERE scope = 'platform'
      AND version = '$drift_version'
      AND state = 'failed'
      AND error_code = 'migration_failed';")"
[ "$failed_runs" -ge 1 ] || fail "El fallo de migración no quedó auditado."

docker exec "$postgres" psql -v ON_ERROR_STOP=1 -U pcs -d pcs_empresas -c \
  "UPDATE schema_migrations
      SET checksum = '$original_checksum'
    WHERE scope = 'platform' AND version = '$drift_version';" >/dev/null
run_migrate "$migrate_recovered"

echo "drift_failure=closed"
echo "drift_schema_unchanged=${tables_before_drift}"
echo "drift_ledger_unchanged=${migrations_before_drift}"
echo "drift_recovery=pass"

echo "[CHECK] Catálogo y tablas resultantes."
docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
  "SELECT 'empresas_migrations=' || COUNT(*) FROM schema_migrations;
   SELECT 'empresas_tables=' || COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"
docker exec "$postgres" psql -U pcs -d pcs_superadministrador -Atc \
  "SELECT 'super_migrations=' || COUNT(*) FROM schema_migrations;
   SELECT 'super_tables=' || COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"

echo "[OK] Base vacía y segunda pasada completadas; se eliminarán los recursos aislados."
