#!/usr/bin/env bash
set -euo pipefail

container="${MAILU_ADMIN_CONTAINER_NAME:-pcs-mailu-admin}"
email="$(printf '%s' "${PCS_MAILU_EMAIL:-}" | tr '[:upper:]' '[:lower:]' | tr -d '\r\n\t ')"
password="${PCS_MAILU_PASSWORD:-}"
name="$(printf '%s' "${PCS_MAILU_NAME:-}" | tr '\r\n\t' '   ' | cut -c1-255)"
domain="$(printf '%s' "${PCS_MAILU_DOMAIN:-}" | tr '[:upper:]' '[:lower:]' | tr -d '\r\n\t ')"
theme_mode="$(printf '%s' "${PCS_MAILU_THEME_MODE:-light}" | tr '[:upper:]' '[:lower:]' | tr -d '\r\n\t ')"
theme_name="$(printf '%s' "${PCS_MAILU_THEME:-}" | tr -d '\r\n\t ')"

if [ -z "$email" ] || [ -z "$password" ]; then
  echo "email y clave son obligatorios para provisionar Mailu" >&2
  exit 2
fi

local_part="${email%@*}"
email_domain="${email#*@}"
if [ -z "$domain" ]; then
  domain="$email_domain"
fi
if [ "$domain" != "$email_domain" ] || [ -z "$local_part" ] || [ -z "$email_domain" ]; then
  echo "email corporativo invalido para el dominio configurado" >&2
  exit 2
fi
case "$theme_mode" in
  dark|oscuro) theme_mode="dark"; theme_name="${theme_name:-PCSDark}" ;;
  *) theme_mode="light"; theme_name="${theme_name:-PCSLight}" ;;
esac

if ! docker ps --format '{{.Names}}' | grep -qx "$container"; then
  echo "contenedor Mailu admin no esta activo" >&2
  exit 3
fi

docker exec \
  -e PCS_MAILU_EMAIL="$email" \
  -e PCS_MAILU_LOCAL="$local_part" \
  -e PCS_MAILU_DOMAIN="$domain" \
  -e PCS_MAILU_PASSWORD="$password" \
  -e PCS_MAILU_NAME="$name" \
  "$container" sh -lc '
set -eu
run_flask() {
  if command -v flask >/dev/null 2>&1; then
    flask "$@"
    return $?
  fi
  if [ -x /app/venv/bin/flask ]; then
    /app/venv/bin/flask "$@"
    return $?
  fi
  if command -v python >/dev/null 2>&1; then
    python -m flask "$@"
    return $?
  fi
  if command -v python3 >/dev/null 2>&1; then
    python3 -m flask "$@"
    return $?
  fi
  echo "cli Mailu/Flask no disponible dentro del contenedor admin" >&2
  return 127
}
local_part="$PCS_MAILU_LOCAL"
domain="$PCS_MAILU_DOMAIN"
password="$PCS_MAILU_PASSWORD"

run_flask mailu domain "$domain" >/dev/null 2>&1 || true
run_flask mailu user "$local_part" "$domain" "$password" >/dev/null 2>&1 || true
run_flask mailu password "$local_part" "$domain" "$password" >/dev/null

echo "mailu-user-ok"
'

imap_container="${MAILU_IMAP_CONTAINER_NAME:-pcs-mailu-imap}"
if docker ps --format '{{.Names}}' | grep -qx "$imap_container"; then
  docker exec "$imap_container" sh -lc 'command -v doveadm >/dev/null 2>&1 && doveadm auth cache flush >/dev/null 2>&1 || true' >/dev/null 2>&1 || true
fi

webmail_container="${MAILU_WEBMAIL_CONTAINER_NAME:-pcs-mailu-webmail}"
if docker ps --format '{{.Names}}' | grep -qx "$webmail_container"; then
  docker exec \
    -e PCS_MAILU_EMAIL="$email" \
    -e PCS_MAILU_NAME="$name" \
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
case "$email" in
  *@*) ;;
  *) exit 0 ;;
esac
safe_local="$(printf "%s" "$local_part" | sed "s/[^A-Za-z0-9_.@-]/_/g; s/^[._-]*//; s/[._-]*$//")"
safe_domain="$(printf "%s" "$domain" | sed "s/[^A-Za-z0-9_.@-]/_/g; s/^[._-]*//; s/[._-]*$//")"
[ -n "$safe_local" ] || safe_local="unknown"
[ -n "$safe_domain" ] || safe_domain="unknown"
display_name="$(printf "%s" "${PCS_MAILU_NAME:-}" | tr "\r\n\t" "   " | tr "\\\"" "  " | cut -c1-120)"
[ -n "$display_name" ] || display_name="$email"
path="/data/_data_/_default_/storage/$safe_domain/$safe_local"
mkdir -p "$path"
printf "[{\"Id\":\"\",\"Label\":\"\",\"Email\":\"%s\",\"Name\":\"%s\",\"ReplyTo\":\"\",\"Bcc\":\"\",\"Signature\":\"\",\"SignatureInsertBefore\":false,\"sentFolder\":\"\",\"pgpEncrypt\":false,\"pgpSign\":false,\"smimeKey\":\"\",\"smimeCertificate\":\"\"}]\n" "$email" "$display_name" > "$path/identities"
settings_dir="$path/settings"
mkdir -p "$settings_dir"
printf "[webmail]\ntheme = \"%s@custom\"\n\n[defaults]\ntheme = \"%s@custom\"\n" "$theme_name" "$theme_name" > "$settings_dir/settings_local"
printf "{\"theme\":\"%s@custom\",\"mode\":\"%s\"}\n" "$theme_name" "$theme_mode" > "$settings_dir/pcs-theme.json"
chown -R mailu:mailu "$path" 2>/dev/null || true
chmod 700 "$path" 2>/dev/null || true
chmod 600 "$path/identities" 2>/dev/null || true
chmod 700 "$settings_dir" 2>/dev/null || true
chmod 600 "$settings_dir/settings_local" "$settings_dir/pcs-theme.json" 2>/dev/null || true
' >/dev/null 2>&1 || true
fi
