#!/usr/bin/env bash
set -euo pipefail

pkill -f 'docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml' || true
sleep 3

docker rm -f pcs-frontend pcs-backend pcs-postgres 2>/dev/null || true

cd /srv/data/pcs/powerfulcontrolsystem
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps
