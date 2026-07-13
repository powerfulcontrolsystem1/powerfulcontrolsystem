#!/usr/bin/env bash
set -euo pipefail

base="/srv/data/nextcloud"
mkdir -p "$base"/{db,html,data,redis}
chown -R 33:33 "$base/html" "$base/data" || true

cat > "$base/docker-compose.yml" <<'EOF'
services:
  db:
    image: mariadb:lts
    container_name: nextcloud-db
    restart: unless-stopped
    command: --transaction-isolation=READ-COMMITTED --binlog-format=ROW
    environment:
      MYSQL_ROOT_PASSWORD: nextcloud_root_admin1
      MYSQL_PASSWORD: nextcloud_db_admin1
      MYSQL_DATABASE: nextcloud
      MYSQL_USER: nextcloud
    volumes:
      - /srv/data/nextcloud/db:/var/lib/mysql

  redis:
    image: redis:alpine
    container_name: nextcloud-redis
    restart: unless-stopped
    volumes:
      - /srv/data/nextcloud/redis:/data

  app:
    image: nextcloud:stable-apache
    container_name: nextcloud-app
    restart: unless-stopped
    ports:
      - "8081:80"
    depends_on:
      - db
      - redis
    environment:
      MYSQL_HOST: db
      MYSQL_PASSWORD: nextcloud_db_admin1
      MYSQL_DATABASE: nextcloud
      MYSQL_USER: nextcloud
      NEXTCLOUD_ADMIN_USER: admin1
      NEXTCLOUD_ADMIN_PASSWORD: admin1next
      NEXTCLOUD_TRUSTED_DOMAINS: "192.168.1.188 localhost vps2"
      REDIS_HOST: redis
      PHP_MEMORY_LIMIT: 512M
      PHP_UPLOAD_LIMIT: 10G
    volumes:
      - /srv/data/nextcloud/html:/var/www/html
      - /srv/data/nextcloud/data:/var/www/html/data
EOF

cd "$base"
docker compose pull
docker compose up -d

echo "[NEXTCLOUD] Esperando puerto 8081"
for i in $(seq 1 60); do
  if curl -fsS -o /dev/null http://127.0.0.1:8081/status.php; then
    break
  fi
  sleep 3
done

docker compose ps
curl -fsS http://127.0.0.1:8081/status.php || true
