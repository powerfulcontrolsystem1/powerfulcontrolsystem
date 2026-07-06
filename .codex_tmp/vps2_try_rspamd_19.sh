set -euo pipefail
cd /srv/data/pcs/powerfulcontrolsystem/deploy
cp -a docker-compose.platform.yml docker-compose.platform.yml.before-rspamd-compat-$(date +%Y%m%d_%H%M%S)
python3 - <<'PY'
from pathlib import Path
p = Path('docker-compose.platform.yml')
s = p.read_text()
s = s.replace('image: ghcr.io/mailu/rspamd:${MAILU_VERSION:-2024.06}', 'image: ghcr.io/mailu/rspamd:${MAILU_RSPAMD_VERSION:-1.9}')
p.write_text(s)
PY
COMPOSE="docker compose --env-file .env.platform -f docker-compose.platform.yml --profile mail"
$COMPOSE stop mailu-antispam || true
$COMPOSE rm -f mailu-antispam || true
$COMPOSE pull mailu-antispam
$COMPOSE up -d mailu-antispam
sleep 45
docker ps --filter name=pcs-mailu-antispam --format '{{.Names}}\t{{.Image}}\t{{.Status}}'
echo "== logs =="
docker logs --tail 120 pcs-mailu-antispam || true
