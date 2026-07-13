#!/usr/bin/env bash
set -euo pipefail

cd /srv/data/nextcloud
sed -i 's/"8081:80"/"8082:80"/' docker-compose.yml
docker compose up -d

cd /srv/data/pcs/powerfulcontrolsystem
docker rm -f pcs-frontend 2>/dev/null || true
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d

echo "---NEXTCLOUD---"
cd /srv/data/nextcloud
docker compose ps
curl -fsS http://127.0.0.1:8082/status.php || true

echo "---PCS---"
cd /srv/data/pcs/powerfulcontrolsystem
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps
curl -I --max-time 10 http://127.0.0.1:8081/ || true
