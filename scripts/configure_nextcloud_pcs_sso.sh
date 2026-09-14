#!/usr/bin/env bash
set -euo pipefail

repo_path="${1:-/root/powerfulcontrolsystem}"
platform_env="$repo_path/deploy/.env.platform"
app_source="$repo_path/deploy/nextcloud/pcs_sso"
container_name="${NEXTCLOUD_CONTAINER:-pcs-nextcloud}"
allowed_origin="${PCS_NEXTCLOUD_EMBED_ORIGIN:-https://powerfulcontrolsystem.com}"
backup_dir="$repo_path/backups/nextcloud-pcs-sso"

test -f "$platform_env"
test -f "$app_source/appinfo/info.xml"
docker inspect "$container_name" >/dev/null
mkdir -p "$backup_dir"

current_secret="$(awk -F= '$1 == "NEXTCLOUD_SSO_SECRET" {sub(/^[^=]*=/, ""); print; exit}' "$platform_env")"
if [ "${#current_secret}" -lt 32 ]; then
  current_secret="$(openssl rand -hex 48)"
fi

tmp_env="$(mktemp "$platform_env.pcs-sso.XXXXXX")"
trap 'rm -f "$tmp_env"' EXIT
awk -v value="$current_secret" '
  BEGIN { found=0 }
  $0 ~ /^NEXTCLOUD_SSO_SECRET=/ { print "NEXTCLOUD_SSO_SECRET=" value; found=1; next }
  { print }
  END { if (!found) print "NEXTCLOUD_SSO_SECRET=" value }
' "$platform_env" > "$tmp_env"
chmod --reference="$platform_env" "$tmp_env"
chown --reference="$platform_env" "$tmp_env"
mv "$tmp_env" "$platform_env"

if docker exec "$container_name" test -d /var/www/html/custom_apps/pcs_sso; then
  docker exec "$container_name" tar -C /var/www/html/custom_apps -czf - pcs_sso > "$backup_dir/pcs_sso-$(date +%Y%m%d%H%M%S).tar.gz"
fi
docker exec -u root "$container_name" rm -rf /var/www/html/custom_apps/pcs_sso
docker exec -u root "$container_name" mkdir -p /var/www/html/custom_apps/pcs_sso
docker cp "$app_source/." "$container_name:/var/www/html/custom_apps/pcs_sso/"
docker exec -u root "$container_name" chown -R www-data:www-data /var/www/html/custom_apps/pcs_sso

while IFS= read -r php_file; do
  docker exec "$container_name" php -l "$php_file" >/dev/null
done < <(docker exec "$container_name" find /var/www/html/custom_apps/pcs_sso -type f -name '*.php' -print)

docker exec -u www-data "$container_name" php occ config:system:set pcs_sso_secret --value="$current_secret" >/dev/null
docker exec -u www-data "$container_name" php occ config:system:set pcs_sso_allowed_origin --value="$allowed_origin" >/dev/null
docker exec -u www-data "$container_name" php occ app:enable pcs_sso >/dev/null

printf 'nextcloud_pcs_sso=ready\n'
printf 'nextcloud_pcs_sso_origin=%s\n' "$allowed_origin"
