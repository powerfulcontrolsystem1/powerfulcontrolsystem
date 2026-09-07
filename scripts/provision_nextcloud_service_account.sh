#!/usr/bin/env bash
set -euo pipefail

# Runs on the main VPS. It creates or rotates a dedicated OCS account without
# printing credentials, then writes only root-readable runtime settings.
repo_path="${1:?repository path required}"
service_user="pcs_ocs_service"
nextcloud_env="$repo_path/deploy/nextcloud/.env"
platform_env="$repo_path/deploy/.env.platform"

test -f "$nextcloud_env"
test -f "$platform_env"
set -a
. "$nextcloud_env"
set +a

secret="$(openssl rand -base64 36 | tr -d '\n')"
if docker exec -u www-data pcs-nextcloud php occ user:info "$service_user" >/dev/null 2>&1; then
  docker exec -e OC_PASS="$secret" -u www-data pcs-nextcloud php occ user:resetpassword --password-from-env "$service_user" >/dev/null
else
  docker exec -e OC_PASS="$secret" -u www-data pcs-nextcloud php occ user:add --password-from-env --display-name="PCS OCS Service" "$service_user" >/dev/null
fi
docker exec -u www-data pcs-nextcloud php occ group:adduser admin "$service_user" >/dev/null 2>&1 || true

backup_dir="/root/pcs-config-backups"
install -d -m 700 "$backup_dir"
cp -p "$platform_env" "$backup_dir/platform.env.before-nextcloud-$(date +%Y%m%d%H%M%S)"
tmp_file="$(mktemp)"
grep -Ev '^(NEXTCLOUD_ENABLED|NEXTCLOUD_BASE_URL|NEXTCLOUD_ADMIN_USER|NEXTCLOUD_ADMIN_SECRET|NEXTCLOUD_DEFAULT_QUOTA_MB)=' "$platform_env" > "$tmp_file" || true
{
  cat "$tmp_file"
  printf '\nNEXTCLOUD_ENABLED=1\n'
  printf 'NEXTCLOUD_BASE_URL=%s\n' 'https://nextcloud.powerfulcontrolsystem.com'
  printf 'NEXTCLOUD_ADMIN_USER=%s\n' "$service_user"
  printf 'NEXTCLOUD_ADMIN_SECRET=%s\n' "$secret"
  printf 'NEXTCLOUD_DEFAULT_QUOTA_MB=1024\n'
} > "$platform_env"
rm -f "$tmp_file"
chmod 600 "$platform_env"
printf 'nextcloud_service_account=ready\n'
