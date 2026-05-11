#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
ENV_FILE="${ENV_FILE:-$PROJECT_DIR/deploy/.env.staging}"

if [ ! -f "$ENV_FILE" ]; then
  echo "[ERROR] No existe $ENV_FILE. Ejecuta primero deploy/scripts/vps-staging-up.sh." >&2
  exit 1
fi

cd "$PROJECT_DIR"

for container in pcs-postgres pcs-staging-postgres; do
  if ! docker ps --format '{{.Names}}' | grep -qx "$container"; then
    echo "[ERROR] Contenedor requerido no activo: $container" >&2
    exit 1
  fi
done

echo "[INFO] Refrescando bases staging desde produccion. Produccion queda solo lectura por pg_dump."

for db in pcs_superadministrador pcs_empresas; do
  echo "[INFO] Reiniciando base staging: $db"
  docker exec pcs-staging-postgres sh -lc "dropdb -U \"\$POSTGRES_USER\" --if-exists '$db'"
  docker exec pcs-staging-postgres sh -lc "createdb -U \"\$POSTGRES_USER\" '$db'"
  echo "[INFO] Copiando $db hacia staging"
  docker exec pcs-postgres sh -lc "pg_dump -U \"\$POSTGRES_USER\" --no-owner --no-privileges '$db'" \
    | docker exec -i pcs-staging-postgres sh -lc "psql -v ON_ERROR_STOP=1 -U \"\$POSTGRES_USER\" '$db' >/dev/null"
done

docker compose --env-file "$ENV_FILE" -f deploy/docker-compose.platform.yml -f deploy/docker-compose.staging.yml restart backend frontend

echo "[OK] Staging refrescado desde produccion."
