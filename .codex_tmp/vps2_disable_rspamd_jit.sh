set -euo pipefail
cd /srv/data/pcs/powerfulcontrolsystem/deploy
COMPOSE="docker compose --env-file .env.platform -f docker-compose.platform.yml --profile mail"
echo "== inspect mounts =="
docker inspect pcs-mailu-antispam --format '{{range .Mounts}}{{println .Name .Destination}}{{end}}'
echo "== create compatibility override =="
docker exec pcs-mailu-antispam sh -lc 'mkdir -p /overrides/rspamd/local.d /overrides/rspamd/override.d; printf "%s\n" "disable_pcre_jit = true;" > /overrides/rspamd/local.d/options.inc; ls -l /overrides/rspamd/local.d/options.inc; cat /overrides/rspamd/local.d/options.inc' || true
echo "== restart =="
$COMPOSE restart mailu-antispam
sleep 25
docker ps --filter name=pcs-mailu-antispam --format '{{.Names}}\t{{.Status}}'
echo "== logs =="
docker logs --tail 100 pcs-mailu-antispam || true
