#!/usr/bin/env bash
set -u

cd /srv/data/pcs/powerfulcontrolsystem || exit 1

echo "---PREVIOUS PROCESSES---"
pgrep -af 'docker compose.*profile mail|restore_to_new_vps' || true

pkill -f 'docker compose.*profile mail' || true
pkill -f restore_to_new_vps.sh || true
sleep 3

echo "---START MAIL PROFILE---"
nohup docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile mail up -d > /tmp/pcs_mail_profile.log 2>&1 &
echo $! > /tmp/pcs_mail_profile.pid
cat /tmp/pcs_mail_profile.pid
