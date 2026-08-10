#!/usr/bin/env bash
# Ensaya el catálogo de migraciones sobre PostgreSQL vacío y una segunda pasada.
# Todos los recursos Docker son aislados, efímeros y se eliminan al terminar.
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "$0")/../.." && pwd)}"
SOURCE_ENV="${SOURCE_ENV:-$PROJECT_DIR/deploy/.env.staging}"
PCS_MIGRATE_IMAGE_DIGEST="${PCS_MIGRATE_IMAGE_DIGEST:-}"
PCS_PREVIOUS_API_IMAGE_DIGEST="${PCS_PREVIOUS_API_IMAGE_DIGEST:-}"
P108_DRILL_ID="${P108_DRILL_ID:-p108-empty-$(date +%s)}"
P108_EMPTY_SCHEMA_BOOTSTRAP="${P108_EMPTY_SCHEMA_BOOTSTRAP:-1}"
P109_VERIFY_MIGRATION_FAILURES="${P109_VERIFY_MIGRATION_FAILURES:-0}"
P109_BACKWARD_HOST_PORT="${P109_BACKWARD_HOST_PORT:-18089}"

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
[[ "$P109_VERIFY_MIGRATION_FAILURES" =~ ^[01]$ ]] ||
  fail "P109_VERIFY_MIGRATION_FAILURES debe ser 0 o 1."
pre_failure_checks=0
if [ "$P109_VERIFY_MIGRATION_FAILURES" = "1" ]; then
  [[ "$PCS_PREVIOUS_API_IMAGE_DIGEST" =~ ^[^[:space:]@]+@sha256:[a-f0-9]{64}$ ]] ||
    fail "PCS_PREVIOUS_API_IMAGE_DIGEST debe usar repositorio@sha256:<64 hex>."
  [[ "$P109_BACKWARD_HOST_PORT" =~ ^[0-9]+$ ]] ||
    fail "P109_BACKWARD_HOST_PORT debe ser numérico."
  ((P109_BACKWARD_HOST_PORT >= 1024 && P109_BACKWARD_HOST_PORT <= 65535)) ||
    fail "P109_BACKWARD_HOST_PORT debe estar entre 1024 y 65535."
fi
[ -f "$SOURCE_ENV" ] || fail "No existe SOURCE_ENV."

require_command docker
require_command openssl
require_command curl
docker image inspect "$PCS_MIGRATE_IMAGE_DIGEST" >/dev/null 2>&1 ||
  fail "La imagen de migración exacta no está disponible localmente."
if [ "$P109_VERIFY_MIGRATION_FAILURES" = "1" ]; then
  docker image inspect "$PCS_PREVIOUS_API_IMAGE_DIGEST" >/dev/null 2>&1 ||
    fail "La imagen API anterior no está disponible localmente."
fi

network="${P108_DRILL_ID}-net"
volume="${P108_DRILL_ID}-pgdata"
postgres="${P108_DRILL_ID}-postgres"
migrate_first="${P108_DRILL_ID}-migrate-1"
migrate_second="${P108_DRILL_ID}-migrate-2"
migrate_drift="${P108_DRILL_ID}-migrate-drift"
migrate_recovered="${P108_DRILL_ID}-migrate-recovered"
migrate_before="${P108_DRILL_ID}-migrate-before"
migrate_during="${P108_DRILL_ID}-migrate-during"
migrate_tx_recovered="${P108_DRILL_ID}-migrate-tx-recovered"
previous_api="${P108_DRILL_ID}-previous-api"
workdir="/tmp/${P108_DRILL_ID}"

for resource in "$postgres" "$migrate_first" "$migrate_second" "$migrate_drift" "$migrate_recovered" "$migrate_before" "$migrate_during" "$migrate_tx_recovered" "$previous_api"; do
  if docker container inspect "$resource" >/dev/null 2>&1; then
    fail "El contenedor aislado ya existe y no será sobrescrito: $resource"
  fi
done
docker volume inspect "$volume" >/dev/null 2>&1 &&
  fail "El volumen aislado ya existe y no será sobrescrito: $volume"
docker network inspect "$network" >/dev/null 2>&1 &&
  fail "La red aislada ya existe y no será sobrescrita: $network"
[ ! -e "$workdir" ] || fail "El directorio temporal ya existe: $workdir"
if [ "$P109_VERIFY_MIGRATION_FAILURES" = "1" ] && ss -H -ltn "sport = :$P109_BACKWARD_HOST_PORT" 2>/dev/null | grep -q .; then
  fail "El puerto local $P109_BACKWARD_HOST_PORT ya está ocupado."
fi

cleanup() {
  docker rm -f "$migrate_first" "$migrate_second" "$migrate_drift" "$migrate_recovered" "$migrate_before" "$migrate_during" "$migrate_tx_recovered" "$previous_api" "$postgres" >/dev/null 2>&1 || true
  docker volume rm "$volume" >/dev/null 2>&1 || true
  docker network rm "$network" >/dev/null 2>&1 || true
  if [[ "$workdir" == /tmp/p108-empty-* ]]; then
    rm -rf -- "$workdir"
  fi
}
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

test_password="$(openssl rand -hex 24)"
restricted_password="$(openssl rand -hex 24)"
runtime_password="$(openssl rand -hex 24)"
mkdir -p "$workdir/private_storage"
chown 10001:10001 "$workdir/private_storage"
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

escaped_runtime_password="${runtime_password//\'/\'\'}"
docker exec -i "$postgres" psql -v ON_ERROR_STOP=1 -U pcs -d postgres <<SQL >/dev/null
CREATE ROLE pcs_empty_runtime LOGIN PASSWORD '${escaped_runtime_password}'
  NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS;
GRANT CONNECT ON DATABASE pcs_empresas TO pcs_empty_runtime;
GRANT CONNECT ON DATABASE pcs_superadministrador TO pcs_empty_runtime;
SQL

run_migrate() {
  local name="$1"
  local db_user="${2:-pcs}"
  local db_password="${3:-$test_password}"
  local schema_bootstrap="${4:-$P108_EMPTY_SCHEMA_BOOTSTRAP}"
  local migrate_command="${5:-/app/backend/pcs-backend && /app/backend/pcs-migrate}"
  docker run \
    --name "$name" \
    --network "$network" \
    --env-file "$SOURCE_ENV" \
    -e PCS_ENV=production \
    -e PCS_RUNTIME_ROLE=migrate \
    -e "PCS_RUNTIME_SCHEMA_BOOTSTRAP=$schema_bootstrap" \
    -e PCS_RUNTIME_DB_USER=pcs_empty_runtime \
    -e PCS_RUNTIME_DB_PASSWORD="$runtime_password" \
    -e PCS_SKIP_CORPORATE_EMAIL_STARTUP_SYNC=1 \
    -e PCS_MIGRATION_SKIP_EXTERNAL_SYNC=1 \
    -e PCS_PRIVATE_STORAGE_DIR=/app/private_storage \
    -e DB_DIALECT=postgres \
    -e "DB_EMPRESAS_DSN=postgres://${db_user}:${db_password}@${postgres}:5432/pcs_empresas?sslmode=disable" \
    -e "DB_SUPERADMIN_DSN=postgres://${db_user}:${db_password}@${postgres}:5432/pcs_superadministrador?sslmode=disable" \
    -v "$workdir/private_storage:/app/private_storage:rw" \
    --entrypoint /bin/sh \
    "$PCS_MIGRATE_IMAGE_DIGEST" \
    -ec "$migrate_command"
}

if [ "$P109_VERIFY_MIGRATION_FAILURES" = "1" ]; then
  echo "[RUN] Fallo cerrado antes de migrar con rol sin DDL."
  escaped_restricted_password="${restricted_password//\'/\'\'}"
  docker exec -i "$postgres" psql -v ON_ERROR_STOP=1 -U pcs -d postgres <<SQL >/dev/null
CREATE ROLE pcs_migration_restricted LOGIN PASSWORD '${escaped_restricted_password}'
  NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS;
GRANT CONNECT ON DATABASE pcs_empresas TO pcs_migration_restricted;
GRANT CONNECT ON DATABASE pcs_superadministrador TO pcs_migration_restricted;
SQL
  for database in pcs_empresas pcs_superadministrador; do
    docker exec "$postgres" psql -v ON_ERROR_STOP=1 -U pcs -d "$database" -c \
      "GRANT USAGE ON SCHEMA public TO pcs_migration_restricted" >/dev/null
  done
  tables_before_restricted="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
    "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'"),$(docker exec "$postgres" psql -U pcs -d pcs_superadministrador -Atc \
    "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'")"
  if run_migrate "$migrate_before" pcs_migration_restricted "$restricted_password" 0 "/app/backend/pcs-migrate" >"$workdir/migrate-before.log" 2>&1; then
    fail "El migrador aceptó un rol sin permisos DDL."
  fi
  pre_failure_checks=$((pre_failure_checks + 1))
  tables_after_restricted="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
    "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'"),$(docker exec "$postgres" psql -U pcs -d pcs_superadministrador -Atc \
    "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'")"
  [ "$tables_before_restricted" = "$tables_after_restricted" ] ||
    fail "El fallo previo dejó cambios parciales en el esquema."
  pre_failure_checks=$((pre_failure_checks + 1))
  restricted_flags="$(docker exec "$postgres" psql -U pcs -d postgres -Atc \
    "SELECT rolsuper::int||','||rolcreatedb::int||','||rolcreaterole::int||','||rolbypassrls::int FROM pg_roles WHERE rolname='pcs_migration_restricted'")"
  [ "$restricted_flags" = "0,0,0,0" ] || fail "El rol restringido obtuvo privilegios DDL."
  pre_failure_checks=$((pre_failure_checks + 1))
fi

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

transaction_failure_checks=0
backward_checks=0
if [ "$P109_VERIFY_MIGRATION_FAILURES" = "1" ]; then
  echo "[RUN] Fallo durante migración después del DDL y antes del ledger."
  fault_version="20260731-001-ai-usage-unique-v1"
  fault_index="ux_empresa_ai_uso_diario_tenant_model_day"
  fault_ledger_before="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
    "SELECT count(*) FROM schema_migrations WHERE scope='platform' AND version='$fault_version'")"
  [ "$fault_ledger_before" = "1" ] || fail "La migración objetivo no está aplicada una sola vez."
  failed_runs_before="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
    "SELECT count(*) FROM schema_migration_runs WHERE scope='platform' AND version='$fault_version' AND state='failed'")"
  docker exec -i "$postgres" psql -v ON_ERROR_STOP=1 -U pcs -d pcs_empresas <<SQL >/dev/null
DROP INDEX IF EXISTS ${fault_index};
DELETE FROM schema_migrations WHERE scope='platform' AND version='${fault_version}';
CREATE OR REPLACE FUNCTION p109_fail_target_migration_ledger() RETURNS trigger
LANGUAGE plpgsql AS \$fn\$
BEGIN
  IF NEW.scope = 'platform' AND NEW.version = '${fault_version}' THEN
    RAISE EXCEPTION 'P109 controlled migration ledger failure';
  END IF;
  RETURN NEW;
END;
\$fn\$;
CREATE TRIGGER p109_fail_target_migration_ledger
BEFORE INSERT ON schema_migrations
FOR EACH ROW EXECUTE FUNCTION p109_fail_target_migration_ledger();
SQL
  if run_migrate "$migrate_during" pcs "$test_password" 0 "/app/backend/pcs-migrate" >"$workdir/migrate-during.log" 2>&1; then
    fail "El fallo transaccional controlado no detuvo el migrador."
  fi
  index_after_failure="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
    "SELECT to_regclass('public.${fault_index}') IS NOT NULL")"
  ledger_after_failure="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
    "SELECT count(*) FROM schema_migrations WHERE scope='platform' AND version='$fault_version'")"
  failed_runs_after="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
    "SELECT count(*) FROM schema_migration_runs WHERE scope='platform' AND version='$fault_version' AND state='failed'")"
  [ "$index_after_failure" = "f" ] || fail "El DDL sobrevivió al rollback transaccional."
  [ "$ledger_after_failure" = "0" ] || fail "El ledger registró una migración fallida."
  [ "$failed_runs_after" -gt "$failed_runs_before" ] || fail "El fallo durante migración no quedó auditado."
  transaction_failure_checks=3

  docker exec -i "$postgres" psql -v ON_ERROR_STOP=1 -U pcs -d pcs_empresas <<SQL >/dev/null
DROP TRIGGER p109_fail_target_migration_ledger ON schema_migrations;
DROP FUNCTION p109_fail_target_migration_ledger();
SQL
  run_migrate "$migrate_tx_recovered" pcs "$test_password" 0 "/app/backend/pcs-migrate"
  recovered_index="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
    "SELECT to_regclass('public.${fault_index}') IS NOT NULL")"
  recovered_ledger="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
    "SELECT count(*) FROM schema_migrations WHERE scope='platform' AND version='$fault_version'")"
  [ "$recovered_index" = "t" ] && [ "$recovered_ledger" = "1" ] ||
    fail "La recuperación no reaplicó índice y ledger atómicamente."
  transaction_failure_checks=$((transaction_failure_checks + 2))

  echo "[RUN] API del candidato anterior sobre el esquema actualizado."
  escaped_runtime_password="${runtime_password//\'/\'\'}"
  docker exec -i "$postgres" psql -v ON_ERROR_STOP=1 -U pcs -d postgres <<SQL >/dev/null
CREATE ROLE pcs_backward_runtime LOGIN PASSWORD '${escaped_runtime_password}'
  NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS;
GRANT CONNECT ON DATABASE pcs_empresas TO pcs_backward_runtime;
GRANT CONNECT ON DATABASE pcs_superadministrador TO pcs_backward_runtime;
SQL
  for database in pcs_empresas pcs_superadministrador; do
    docker exec -i "$postgres" psql -v ON_ERROR_STOP=1 -U pcs -d "$database" <<SQL >/dev/null
GRANT USAGE ON SCHEMA public TO pcs_backward_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO pcs_backward_runtime;
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO pcs_backward_runtime;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO pcs_backward_runtime;
SQL
  done
  tables_before_backward="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
    "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'")"
  docker run -d \
    --name "$previous_api" \
    --network "$network" \
    --env-file "$SOURCE_ENV" \
    --read-only \
    --cap-drop ALL \
    --security-opt no-new-privileges:true \
    --pids-limit 256 \
    --tmpfs /tmp:rw,noexec,nosuid,size=64m \
    --tmpfs /app/backend/logs:rw,noexec,nosuid,size=32m,uid=10001,gid=10001 \
    --tmpfs /app/web/uploads:rw,noexec,nosuid,size=32m,uid=10001,gid=10001 \
    --tmpfs /app/backup:rw,noexec,nosuid,size=16m,uid=10001,gid=10001 \
    --tmpfs /app/descargas:rw,noexec,nosuid,size=16m,uid=10001,gid=10001 \
    -v "$workdir/private_storage:/app/private_storage:rw" \
    -p "127.0.0.1:${P109_BACKWARD_HOST_PORT}:8080" \
    -e PCS_ENV=staging \
    -e PORT=8080 \
    -e PCS_RUNTIME_ROLE=api \
    -e PCS_RUNTIME_SCHEMA_BOOTSTRAP=0 \
    -e PCS_RUNTIME_DB_USER=pcs_backward_runtime \
    -e PCS_RUNTIME_DB_PASSWORD="$runtime_password" \
    -e DB_DIALECT=postgres \
    -e "DB_EMPRESAS_DSN=postgres://pcs_backward_runtime:${runtime_password}@${postgres}:5432/pcs_empresas?sslmode=disable" \
    -e "DB_SUPERADMIN_DSN=postgres://pcs_backward_runtime:${runtime_password}@${postgres}:5432/pcs_superadministrador?sslmode=disable" \
    -e PCS_SKIP_CORPORATE_EMAIL_STARTUP_SYNC=1 \
    -e PCS_SKIP_SERVER_STARTUP_EVENT=1 \
    -e NEXTCLOUD_ENABLED=0 \
    -e EMAIL_CORPORATIVO_ENABLED=0 \
    -e OPENAI_API_KEY= \
    -e GEMINI_API_KEY= \
    "$PCS_PREVIOUS_API_IMAGE_DIGEST" >/dev/null
  for attempt in $(seq 1 60); do
    if curl -fsS "http://127.0.0.1:${P109_BACKWARD_HOST_PORT}/ready" >/dev/null 2>&1; then
      break
    fi
    [ "$attempt" -lt 60 ] || fail "La API anterior no alcanzó readiness sobre el esquema nuevo."
    sleep 2
  done
  curl -fsS "http://127.0.0.1:${P109_BACKWARD_HOST_PORT}/health" >/dev/null
  backward_checks=2
  tables_after_backward="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
    "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'")"
  [ "$tables_before_backward" = "$tables_after_backward" ] ||
    fail "La API anterior modificó el esquema nuevo."
  backward_checks=$((backward_checks + 1))
  runtime_flags="$(docker exec "$postgres" psql -U pcs -d postgres -Atc \
    "SELECT rolsuper::int||','||rolcreatedb::int||','||rolcreaterole::int||','||rolbypassrls::int FROM pg_roles WHERE rolname='pcs_backward_runtime'")"
  [ "$runtime_flags" = "0,0,0,0" ] || fail "La API anterior obtuvo privilegios DDL."
  backward_checks=$((backward_checks + 1))
fi

echo "[CHECK] Catálogo y tablas resultantes."
docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
  "SELECT 'empresas_migrations=' || COUNT(*) FROM schema_migrations;
   SELECT 'empresas_tables=' || COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"
docker exec "$postgres" psql -U pcs -d pcs_superadministrador -Atc \
  "SELECT 'super_migrations=' || COUNT(*) FROM schema_migrations;
   SELECT 'super_tables=' || COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"

echo "[OK] Base vacía y segunda pasada completadas; fallos_previos=${pre_failure_checks} rollback_transaccional=${transaction_failure_checks} compatibilidad_anterior=${backward_checks}; se eliminarán los recursos aislados."
