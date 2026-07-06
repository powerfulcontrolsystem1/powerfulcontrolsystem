#!/usr/bin/env bash
set -euo pipefail

cd /srv/data/pcs/powerfulcontrolsystem
cp deploy/.env.platform "deploy/.env.platform.bak-httpbind-$(date +%Y%m%d%H%M%S)"
if grep -q '^HTTP_BIND=' deploy/.env.platform; then
  sed -i 's/^HTTP_BIND=.*/HTTP_BIND=0.0.0.0/' deploy/.env.platform
else
  printf '\nHTTP_BIND=0.0.0.0\n' >> deploy/.env.platform
fi
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d frontend
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps
curl -I --max-time 10 http://127.0.0.1:8081/ || true
