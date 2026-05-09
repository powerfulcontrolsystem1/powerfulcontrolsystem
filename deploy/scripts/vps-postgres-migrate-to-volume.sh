#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
COMPOSE_FILE="${COMPOSE_FILE:-$PROJECT_DIR/deploy/docker-compose.platform.yml}"
ENV_FILE="${ENV_FILE:-$PROJECT_DIR/deploy/.env.platform}"
BACKEND_ENV="${BACKEND_ENV:-$PROJECT_DIR/backend/.env.local}"
BACKUP_DIR="${BACKUP_DIR:-$PROJECT_DIR/backups/docker-migration/$(date +%Y%m%d-%H%M%S)}"

if [ "${CONFIRM_MIGRATE:-}" != "YES" ]; then
  echo "Este script copia las bases actuales hacia el volumen PostgreSQL de Docker."
  echo "Ejecuta con CONFIRM_MIGRATE=YES cuando ya tengas respaldo y el stack docker arriba."
  exit 2
fi

get_env_value() {
  local file="$1"
  local key="$2"
  [ -f "$file" ] || return 0
  awk -v key="$key" '
    {
      line=$0
      sub(/^[ \t]+/, "", line)
      if (index(line, key "=") == 1) {
        sub(/^[^=]*=/, "", line)
        gsub(/^["'\'']|["'\'']$/, "", line)
        value=line
      }
    }
    END { if (value != "") print value }
  ' "$file"
}

test -f "$BACKEND_ENV" || { echo "ERROR: no existe $BACKEND_ENV"; exit 1; }
test -f "$ENV_FILE" || { echo "ERROR: no existe $ENV_FILE"; exit 1; }

DB_SUPERADMIN_DSN="$(get_env_value "$BACKEND_ENV" DB_SUPERADMIN_DSN)"
DB_EMPRESAS_DSN="$(get_env_value "$BACKEND_ENV" DB_EMPRESAS_DSN)"
POSTGRES_USER="$(get_env_value "$ENV_FILE" POSTGRES_USER)"
POSTGRES_USER="${POSTGRES_USER:-pcs}"

[ -n "$DB_SUPERADMIN_DSN" ] || { echo "ERROR: DB_SUPERADMIN_DSN vacio en $BACKEND_ENV"; exit 1; }
[ -n "$DB_EMPRESAS_DSN" ] || { echo "ERROR: DB_EMPRESAS_DSN vacio en $BACKEND_ENV"; exit 1; }

if ! command -v pg_dump >/dev/null 2>&1; then
  echo "ERROR: pg_dump no esta instalado. Instala postgresql-client en la VPS."
  exit 1
fi

mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"

echo "[migrate] Generando dumps en $BACKUP_DIR"
pg_dump "$DB_SUPERADMIN_DSN" -Fc -f "$BACKUP_DIR/pcs_superadministrador.dump"
pg_dump "$DB_EMPRESAS_DSN" -Fc -f "$BACKUP_DIR/pcs_empresas.dump"

cd "$PROJECT_DIR"
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d postgres

echo "[migrate] Esperando PostgreSQL Docker"
for attempt in $(seq 1 60); do
  if docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" exec -T postgres pg_isready -h 127.0.0.1 -U "$POSTGRES_USER" -d pcs_superadministrador >/dev/null 2>&1; then
    break
  fi
  if [ "$attempt" -eq 60 ]; then
    echo "ERROR: PostgreSQL Docker no estuvo listo a tiempo."
    exit 1
  fi
  sleep 2
done

echo "[migrate] Restaurando pcs_superadministrador en volumen Docker"
cat "$BACKUP_DIR/pcs_superadministrador.dump" | docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" exec -T postgres pg_restore --clean --if-exists --no-owner --no-acl -h 127.0.0.1 -U "$POSTGRES_USER" -d pcs_superadministrador

echo "[migrate] Restaurando pcs_empresas en volumen Docker"
cat "$BACKUP_DIR/pcs_empresas.dump" | docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" exec -T postgres pg_restore --clean --if-exists --no-owner --no-acl -h 127.0.0.1 -U "$POSTGRES_USER" -d pcs_empresas

echo "[migrate] OK. Dumps conservados en $BACKUP_DIR"
