#!/usr/bin/env bash
set -euo pipefail

# Levanta exclusivamente el Nextcloud empresarial del VPS principal. VPS2 no se
# consulta ni se modifica. El archivo .env es privado y nunca debe versionarse.
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE="$ROOT/deploy/nextcloud/docker-compose.yml"
ENV_FILE="$ROOT/deploy/nextcloud/.env"
if [[ ! -f "$ENV_FILE" ]]; then
  echo "Falta $ENV_FILE. Copie .env.example y complete los secretos solo en el VPS principal." >&2
  exit 2
fi
docker compose --env-file "$ENV_FILE" -f "$COMPOSE" config --quiet
docker compose --env-file "$ENV_FILE" -f "$COMPOSE" up -d
docker compose --env-file "$ENV_FILE" -f "$COMPOSE" ps
