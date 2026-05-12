#!/usr/bin/env bash
set -euo pipefail

PURGE_DATA="${1:-}"

echo "[pcs] Stopping obsolete Nextcloud containers if they exist..."
docker rm -f pcs-nextcloud pcs-nextcloud-db pcs-nextcloud-redis 2>/dev/null || true

echo "[pcs] Removing obsolete Nextcloud compose stack if the old folder still exists..."
if [ -f "deploy/nextcloud/docker-compose.yml" ]; then
  docker compose -f deploy/nextcloud/docker-compose.yml down --remove-orphans 2>/dev/null || true
fi

echo "[pcs] Removing obsolete platform cloud profile containers if present..."
docker compose -f deploy/docker-compose.platform.yml --profile cloud down --remove-orphans 2>/dev/null || true

if [ "$PURGE_DATA" = "--purge-data" ]; then
  echo "[pcs] Purging obsolete Nextcloud Docker volumes..."
  docker volume rm \
    pcs_nextcloud_db \
    pcs_nextcloud_redis \
    pcs_nextcloud_html \
    pcs_nextcloud_data \
    pcs_nextcloud_apps \
    pcs_nextcloud_config \
    2>/dev/null || true
else
  echo "[pcs] Data volumes were preserved. Re-run with --purge-data after confirming no recovery is needed."
fi

echo "[pcs] Nextcloud decommission completed."
