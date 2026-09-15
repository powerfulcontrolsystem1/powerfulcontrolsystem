#!/usr/bin/env bash
set -euo pipefail

webmail_container="${MAILU_WEBMAIL_CONTAINER_NAME:-pcs-mailu-webmail}"
email="$(printf '%s' "${PCS_MAILU_EMAIL:-}" | tr '[:upper:]' '[:lower:]' | tr -d '\r\n\t ')"
theme_mode="$(printf '%s' "${PCS_MAILU_THEME_MODE:-light}" | tr '[:upper:]' '[:lower:]' | tr -d '\r\n\t ')"
theme_name="$(printf '%s' "${PCS_MAILU_THEME:-}" | tr -d '\r\n\t ')"

case "$theme_mode" in
  dark|oscuro) theme_mode="dark"; theme_name="${theme_name:-PCSDark}" ;;
  *) theme_mode="light"; theme_name="${theme_name:-PCSLight}" ;;
esac
case "$theme_name" in
  PCSLight|PCSDark) ;;
  *) theme_name="PCSLight" ;;
esac
case "$email" in
  *@*) ;;
  *) echo "email corporativo invalido" >&2; exit 2 ;;
esac

if ! docker ps --format '{{.Names}}' | grep -qx "$webmail_container"; then
  echo "contenedor SnappyMail no esta activo" >&2
  exit 3
fi

docker exec \
  -e PCS_MAILU_EMAIL="$email" \
  -e PCS_MAILU_THEME_MODE="$theme_mode" \
  -e PCS_MAILU_THEME="$theme_name" \
  "$webmail_container" sh -lc '
set -eu
email="$(printf "%s" "$PCS_MAILU_EMAIL" | tr "[:upper:]" "[:lower:]" | tr -d "\r\n\t ")"
local_part="${email%@*}"
domain="${email#*@}"
theme_mode="$(printf "%s" "${PCS_MAILU_THEME_MODE:-light}" | tr "[:upper:]" "[:lower:]" | tr -d "\r\n\t ")"
theme_name="$(printf "%s" "${PCS_MAILU_THEME:-}" | tr -d "\r\n\t ")"
case "$theme_mode" in
  dark|oscuro) theme_mode="dark"; theme_name="${theme_name:-PCSDark}" ;;
  *) theme_mode="light"; theme_name="${theme_name:-PCSLight}" ;;
esac
case "$theme_name" in
  PCSLight|PCSDark) ;;
  *) theme_name="PCSLight" ;;
esac
safe_local="$(printf "%s" "$local_part" | sed "s/[^A-Za-z0-9_.@-]/_/g; s/^[._-]*//; s/[._-]*$//")"
safe_domain="$(printf "%s" "$domain" | sed "s/[^A-Za-z0-9_.@-]/_/g; s/^[._-]*//; s/[._-]*$//")"
[ -n "$safe_local" ] || exit 2
[ -n "$safe_domain" ] || exit 2
path="/data/_data_/_default_/storage/$safe_domain/$safe_local"
mkdir -p "$path"
settings_file="$path/settings"
if [ -d "$settings_file" ]; then
  rm -f "$settings_file/settings_local" "$settings_file/pcs-theme.json"
  rmdir "$settings_file"
fi
update_theme_setting() {
  settings_target="$1"
  settings_tmp="$settings_target.pcs-theme.tmp"
  if [ ! -f "$settings_target" ] || ! grep -q "^[[:space:]]*{" "$settings_target"; then
    printf "{\"Theme\":\"%s@custom\"}" "$theme_name" > "$settings_tmp"
  elif grep -q "\"Theme\"[[:space:]]*:" "$settings_target"; then
    sed -E "s/\"Theme\"[[:space:]]*:[[:space:]]*\"[^\"]*\"/\"Theme\":\"$theme_name@custom\"/" "$settings_target" > "$settings_tmp"
  elif grep -Eq "^[[:space:]]*\\{[[:space:]]*\\}[[:space:]]*$" "$settings_target"; then
    printf "{\"Theme\":\"%s@custom\"}" "$theme_name" > "$settings_tmp"
  else
    sed "s/^[[:space:]]*{/{\"Theme\":\"$theme_name@custom\",/" "$settings_target" > "$settings_tmp"
  fi
  mv "$settings_tmp" "$settings_target"
}
update_theme_setting "$path/settings"
update_theme_setting "$path/settings_local"
printf "{\"theme\":\"%s@custom\",\"mode\":\"%s\"}\n" "$theme_name" "$theme_mode" > "$path/pcs-theme.json"
chown -R mailu:mailu "$path" 2>/dev/null || true
chmod 700 "$path" 2>/dev/null || true
chmod 600 "$path/settings" "$path/settings_local" "$path/pcs-theme.json" 2>/dev/null || true
'

printf 'snappymail-theme=%s@custom\n' "$theme_name"
