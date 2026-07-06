set -euo pipefail
cd /srv/data/pcs/powerfulcontrolsystem/deploy
COMPOSE="docker compose --env-file .env.platform -f docker-compose.platform.yml --profile mail"
echo "== volumes =="
docker volume ls --format '{{.Name}}' | grep -E 'mailu|rspamd|filter|powerful' || true
echo "== stop antispam =="
$COMPOSE stop mailu-antispam || true
STAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR=/srv/data/backups/pcs_restore/mail_repair_$STAMP
mkdir -p "$BACKUP_DIR"
for v in $(docker volume ls --format '{{.Name}}' | grep -E 'mailu.*filter|filter|rspamd' || true); do
  echo "Backing up volume $v"
  docker run --rm -v "$v:/volume:ro" -v "$BACKUP_DIR:/backup" alpine sh -c "cd /volume && tar czf /backup/${v}.tar.gz ." || true
  echo "Removing volume $v"
  docker volume rm "$v" || true
done
echo "== up antispam =="
$COMPOSE up -d mailu-antispam
echo "== wait =="
sleep 20
docker ps --filter name=pcs-mailu-antispam --format '{{.Names}}\t{{.Status}}'
echo "== logs =="
docker logs --tail 80 pcs-mailu-antispam || true
