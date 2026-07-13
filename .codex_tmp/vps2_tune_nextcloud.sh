#!/usr/bin/env bash
set -euo pipefail

docker exec -u www-data nextcloud-app php occ background:cron
docker exec -u www-data nextcloud-app php occ app:install memories || true
docker exec -u www-data nextcloud-app php occ app:enable memories || true

(crontab -l 2>/dev/null | grep -v 'nextcloud-app php -f /var/www/html/cron.php' || true; \
  echo '*/5 * * * * docker exec -u www-data nextcloud-app php -f /var/www/html/cron.php >/dev/null 2>&1') | crontab -

echo "---MEMORIES---"
docker exec -u www-data nextcloud-app php occ app:list | grep -i memories || true
echo "---CRON---"
crontab -l
