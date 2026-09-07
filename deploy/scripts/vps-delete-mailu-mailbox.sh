#!/usr/bin/env bash
set -euo pipefail

container="${MAILU_ADMIN_CONTAINER_NAME:-pcs-mailu-admin}"
email="$(printf '%s' "${PCS_MAILU_EMAIL:-}" | tr '[:upper:]' '[:lower:]' | tr -d '\r\n\t ')"
domain="$(printf '%s' "${PCS_MAILU_DOMAIN:-}" | tr '[:upper:]' '[:lower:]' | tr -d '\r\n\t ')"
local_part="${email%@*}"

if [ -z "$email" ] || [ -z "$local_part" ] || [ -z "$domain" ] || [ "${email#*@}" != "$domain" ]; then
  echo "email corporativo invalido" >&2
  exit 2
fi
case "$local_part" in *[!a-z0-9._+-]*|'') echo "email corporativo invalido" >&2; exit 2;; esac
case "$domain" in *[!a-z0-9.-]*|'') echo "email corporativo invalido" >&2; exit 2;; esac
if ! docker ps --format '{{.Names}}' | grep -qx "$container"; then
  echo "contenedor Mailu admin no esta activo" >&2
  exit 3
fi

docker exec \
  -e PCS_MAILU_LOCAL="$local_part" \
  -e PCS_MAILU_DOMAIN="$domain" \
  "$container" sh -lc '
set -eu
run_flask() {
  if command -v flask >/dev/null 2>&1; then flask "$@"; return $?; fi
  if [ -x /app/venv/bin/flask ]; then /app/venv/bin/flask "$@"; return $?; fi
  if command -v python >/dev/null 2>&1; then python -m flask "$@"; return $?; fi
  python3 -m flask "$@"
}
run_flask mailu user-delete "$PCS_MAILU_LOCAL@$PCS_MAILU_DOMAIN" || true
' >/dev/null

imap_container="${MAILU_IMAP_CONTAINER_NAME:-pcs-mailu-imap}"
if docker ps --format '{{.Names}}' | grep -qx "$imap_container"; then
  docker exec -e PCS_MAILU_LOCAL="$local_part" -e PCS_MAILU_DOMAIN="$domain" "$imap_container" sh -lc '
set -eu
if command -v doveadm >/dev/null 2>&1; then
  doveadm purge -u "$PCS_MAILU_LOCAL@$PCS_MAILU_DOMAIN" >/dev/null 2>&1 || true
fi
find /mail -type d -path "*/$PCS_MAILU_DOMAIN/$PCS_MAILU_LOCAL" -prune -exec rm -rf {} + 2>/dev/null || true
'
fi

webmail_container="${MAILU_WEBMAIL_CONTAINER_NAME:-pcs-mailu-webmail}"
if docker ps --format '{{.Names}}' | grep -qx "$webmail_container"; then
  docker exec -e PCS_MAILU_LOCAL="$local_part" -e PCS_MAILU_DOMAIN="$domain" "$webmail_container" sh -lc '
set -eu
rm -rf "/data/_data_/_default_/storage/$PCS_MAILU_DOMAIN/$PCS_MAILU_LOCAL"
'
fi

docker exec -e PCS_MAILU_EMAIL="$email" "$container" sh -lc '
set -eu
run_flask() {
  if command -v flask >/dev/null 2>&1; then flask "$@"; return $?; fi
  if [ -x /app/venv/bin/flask ]; then /app/venv/bin/flask "$@"; return $?; fi
  if command -v python >/dev/null 2>&1; then python -m flask "$@"; return $?; fi
  python3 -m flask "$@"
}
run_flask mailu user-delete "$PCS_MAILU_EMAIL" -r
'

echo "mailu-user-deleted"
