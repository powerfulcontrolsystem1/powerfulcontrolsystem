#!/usr/bin/env bash
set -euo pipefail

cd /srv/data/pcs/powerfulcontrolsystem
echo "---MATCHES-BEFORE---"
grep -R '127\.0\.0\.1:8081\|8081:80\|FRONTEND\|PUBLIC' -n deploy .env* 2>/dev/null | head -80 || true

grep -RIl '127.0.0.1:8081' deploy .env* 2>/dev/null | while read -r file; do
  cp "$file" "$file.bak-expose-$(date +%Y%m%d%H%M%S)"
  sed -i 's/127\.0\.0\.1:8081/0.0.0.0:8081/g' "$file"
done

docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d frontend

echo "---MATCHES-AFTER---"
grep -R '127\.0\.0\.1:8081\|0\.0\.0\.0:8081\|8081:80' -n deploy .env* 2>/dev/null | head -80 || true
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps
curl -I --max-time 10 http://127.0.0.1:8081/ || true
