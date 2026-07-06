set -euo pipefail
cd /srv/data/pcs/powerfulcontrolsystem/deploy
COMPOSE="docker compose --env-file .env.platform -f docker-compose.platform.yml --profile mail"
$COMPOSE stop mailu-antispam || true
$COMPOSE rm -f mailu-antispam || true
docker run --rm -v powerful-control-system_mailu_overrides:/overrides alpine sh -lc 'rm -rf /overrides/rspamd; printf "%s\n" "disable_pcre_jit = true;" > /overrides/options.inc; find /overrides -maxdepth 2 -type f -print -exec sed -n "1,80p" {} \;'
$COMPOSE up -d mailu-antispam
sleep 35
docker ps --filter name=pcs-mailu-antispam --format '{{.Names}}\t{{.Status}}'
echo "== logs =="
docker logs --tail 120 pcs-mailu-antispam || true
