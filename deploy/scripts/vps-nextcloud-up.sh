#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE="$ROOT/deploy/nextcloud/docker-compose.yml"
ENV_FILE="$ROOT/deploy/nextcloud/.env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Falta $ENV_FILE; cree el archivo desde .env.example." >&2
  exit 2
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

required=(NEXTCLOUD_DATA_PATH NEXTCLOUD_DB_PASSWORD_FILE NEXTCLOUD_ADMIN_PASSWORD_FILE NEXTCLOUD_REDIS_PASSWORD_FILE)
for name in "${required[@]}"; do
  value="${!name:-}"
  if [[ -z "$value" ]]; then
    echo "Falta $name en $ENV_FILE." >&2
    exit 2
  fi
done

install -d -m 0750 "$NEXTCLOUD_DATA_PATH"
for secret_file in "$NEXTCLOUD_DB_PASSWORD_FILE" "$NEXTCLOUD_ADMIN_PASSWORD_FILE" "$NEXTCLOUD_REDIS_PASSWORD_FILE"; do
  if [[ ! -s "$secret_file" ]]; then
    echo "Falta el secreto legible $secret_file." >&2
    exit 2
  fi
  mode="$(stat -c '%a' "$secret_file")"
  if (( 10#$mode % 100 != 0 )); then
    echo "Permisos inseguros en $secret_file; use chmod 600." >&2
    exit 2
  fi
done

docker compose --env-file "$ENV_FILE" -f "$COMPOSE" config --quiet
docker compose --env-file "$ENV_FILE" -f "$COMPOSE" pull
docker compose --env-file "$ENV_FILE" -f "$COMPOSE" up -d --remove-orphans
docker compose --env-file "$ENV_FILE" -f "$COMPOSE" ps
