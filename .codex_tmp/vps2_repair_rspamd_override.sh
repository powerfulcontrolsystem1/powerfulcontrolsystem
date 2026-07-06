set -euo pipefail
cd /srv/data/pcs/powerfulcontrolsystem/deploy
COMPOSE="docker compose --env-file .env.platform -f docker-compose.platform.yml --profile mail"
echo "== stop/remove antispam container =="
$COMPOSE stop mailu-antispam || true
$COMPOSE rm -f mailu-antispam || true
echo "== write overrides volume =="
docker run --rm -v powerful-control-system_mailu_overrides:/overrides alpine sh -lc 'mkdir -p /overrides/rspamd/local.d /overrides/rspamd/override.d; printf "%s\n" "disable_pcre_jit = true;" > /overrides/rspamd/local.d/options.inc; find /overrides -maxdepth 4 -type f -print -exec sed -n "1,80p" {} \;'
echo "== remove filter volume after container removal =="
docker volume rm powerful-control-system_mailu_filter || true
echo "== up antispam =="
$COMPOSE up -d mailu-antispam
sleep 35
docker ps --filter name=pcs-mailu-antispam --format '{{.Names}}\t{{.Status}}'
echo "== recent logs =="
docker logs --tail 120 pcs-mailu-antispam || true
