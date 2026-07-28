#!/usr/bin/env bash
# Clona lógicamente el PostgreSQL de staging en un entorno efímero y ensaya el
# migrador exacto sin modificar el origen.
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "$0")/../.." && pwd)}"
SOURCE_ENV="${SOURCE_ENV:-$PROJECT_DIR/deploy/.env.staging}"
SOURCE_POSTGRES_CONTAINER="${SOURCE_POSTGRES_CONTAINER:-pcs-staging-postgres}"
PCS_MIGRATE_IMAGE_DIGEST="${PCS_MIGRATE_IMAGE_DIGEST:-}"
P108_DRILL_ID="${P108_DRILL_ID:-p108-upgrade-$(date +%s)}"

fail() {
  echo "[ERROR] $*" >&2
  exit 1
}

[[ "$SOURCE_POSTGRES_CONTAINER" =~ ^pcs-staging-[a-z0-9-]+$ ]] ||
  fail "El origen debe ser un contenedor aislado de staging con prefijo pcs-staging-."
[[ "$P108_DRILL_ID" =~ ^p108-upgrade-[a-z0-9-]+$ ]] ||
  fail "P108_DRILL_ID debe iniciar con p108-upgrade-."
[[ "$PCS_MIGRATE_IMAGE_DIGEST" =~ ^[^[:space:]@]+@sha256:[a-f0-9]{64}$ ]] ||
  fail "PCS_MIGRATE_IMAGE_DIGEST debe usar repositorio@sha256:<64 hex>."
[ -f "$SOURCE_ENV" ] || fail "No existe SOURCE_ENV."
docker container inspect "$SOURCE_POSTGRES_CONTAINER" >/dev/null 2>&1 ||
  fail "No existe el PostgreSQL de staging autorizado."

network="${P108_DRILL_ID}-net"
volume="${P108_DRILL_ID}-pgdata"
postgres="${P108_DRILL_ID}-postgres"
migrate_first="${P108_DRILL_ID}-migrate-1"
migrate_second="${P108_DRILL_ID}-migrate-2"

for resource in "$postgres" "$migrate_first" "$migrate_second"; do
  docker container inspect "$resource" >/dev/null 2>&1 &&
    fail "El contenedor aislado ya existe y no será sobrescrito: $resource"
done
docker volume inspect "$volume" >/dev/null 2>&1 &&
  fail "El volumen aislado ya existe y no será sobrescrito: $volume"
docker network inspect "$network" >/dev/null 2>&1 &&
  fail "La red aislada ya existe y no será sobrescrita: $network"

cleanup() {
  docker rm -f "$migrate_first" "$migrate_second" "$postgres" >/dev/null 2>&1 || true
  docker volume rm "$volume" >/dev/null 2>&1 || true
  docker network rm "$network" >/dev/null 2>&1 || true
}
trap cleanup EXIT

test_password="$(openssl rand -hex 24)"
source_user="$(docker exec "$SOURCE_POSTGRES_CONTAINER" sh -lc 'printf %s "$POSTGRES_USER"')"
[ -n "$source_user" ] || fail "No se pudo resolver el usuario PostgreSQL de staging."

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

echo "[RUN] Copia lógica consistente de staging."
for database in pcs_empresas pcs_superadministrador; do
  docker exec "$SOURCE_POSTGRES_CONTAINER" \
    pg_dump -U "$source_user" --format=custom --no-owner --no-privileges "$database" |
    docker exec -i "$postgres" \
      pg_restore -U pcs --dbname="$database" --clean --if-exists --no-owner --no-privileges
done

before_empresas_tables="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
  "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'")"
before_super_tables="$(docker exec "$postgres" psql -U pcs -d pcs_superadministrador -Atc \
  "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'")"

run_migrate() {
  local name="$1"
  docker run \
    --name "$name" \
    --network "$network" \
    --env-file "$SOURCE_ENV" \
    -e PCS_ENV=production \
    -e PCS_RUNTIME_ROLE=migrate \
    -e PCS_RUNTIME_SCHEMA_BOOTSTRAP=0 \
    -e PCS_SKIP_CORPORATE_EMAIL_STARTUP_SYNC=1 \
    -e PCS_MIGRATION_SKIP_EXTERNAL_SYNC=1 \
    -e DB_DIALECT=postgres \
    -e "DB_EMPRESAS_DSN=postgres://pcs:${test_password}@${postgres}:5432/pcs_empresas?sslmode=disable" \
    -e "DB_SUPERADMIN_DSN=postgres://pcs:${test_password}@${postgres}:5432/pcs_superadministrador?sslmode=disable" \
    --entrypoint /bin/sh \
    "$PCS_MIGRATE_IMAGE_DIGEST" \
    -ec '/app/backend/pcs-backend && /app/backend/pcs-migrate'
}

echo "[RUN] Upgrade del snapshot representativo."
run_migrate "$migrate_first"
echo "[RUN] Segunda pasada idempotente."
run_migrate "$migrate_second"

after_empresas_tables="$(docker exec "$postgres" psql -U pcs -d pcs_empresas -Atc \
  "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'")"
after_super_tables="$(docker exec "$postgres" psql -U pcs -d pcs_superadministrador -Atc \
  "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'")"

[ "$after_empresas_tables" -ge "$before_empresas_tables" ] ||
  fail "El upgrade perdió tablas empresariales."
[ "$after_super_tables" -ge "$before_super_tables" ] ||
  fail "El upgrade perdió tablas administrativas."

echo "empresas_tables_before=$before_empresas_tables after=$after_empresas_tables"
echo "super_tables_before=$before_super_tables after=$after_super_tables"
echo "[OK] Upgrade y segunda pasada completados; se eliminarán los recursos aislados."
