#!/usr/bin/env bash
set -euo pipefail

pkill -f 'docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile mail up -d' || true
pkill -f restore_to_new_vps.sh || true
sleep 3
echo "---RUNNING-AFTER-KILL---"
pgrep -af 'restore_to_new_vps|docker compose' || true

cd /srv/data/pcs/powerfulcontrolsystem
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps
