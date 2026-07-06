#!/usr/bin/env bash
set -u

cd /srv/data/pcs/powerfulcontrolsystem

echo "---MAIL ENV---"
grep -nE 'MAILU|SMTP|IMAP|WEBMAIL|BIND|PORT|DOMAIN|HOSTNAME' deploy/.env.platform 2>/dev/null || true

echo "---MAIL SERVICES---"
grep -n 'profiles: \["mail"\]\|MAILU_\|mailu_\|ports:' deploy/docker-compose.platform.yml | head -220 || true

echo "---COMPOSE CONFIG MAIL PORTS---"
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile mail config 2>/tmp/compose-mail-config.err | grep -A8 -B2 'ports:' | head -220 || {
  cat /tmp/compose-mail-config.err
  exit 0
}
