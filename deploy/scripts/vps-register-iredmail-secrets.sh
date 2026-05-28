#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
ENV_FILE="${ENV_FILE:-$PROJECT_DIR/deploy/.env.platform}"

get_env_value() {
  local file="$1"
  local key="$2"
  [ -f "$file" ] || return 0
  awk -v key="$key" '
    {
      line=$0
      sub(/^[ \t]+/, "", line)
      if (index(line, key "=") == 1) {
        sub(/^[^=]*=/, "", line)
        gsub(/^["'\'']|["'\'']$/, "", line)
        value=line
      }
    }
    END { if (value != "") print value }
  ' "$file"
}

require_secret() {
  local key="$1"
  local value
  value="$(get_env_value "$ENV_FILE" "$key")"
  if [ -z "$value" ] || [[ "$value" == change-me-* ]]; then
    echo "[iredmail] ERROR: falta $key en $ENV_FILE" >&2
    exit 1
  fi
  echo "[iredmail] $key: OK"
}

test -f "$ENV_FILE" || {
  echo "[iredmail] ERROR: no existe $ENV_FILE. Ejecuta deploy/scripts/vps-docker-preflight.sh primero." >&2
  exit 1
}

require_secret CONFIG_ENC_KEY
require_secret EMAIL_CORPORATIVO_IREDADMIN_PASSWORD
require_secret IREDMAIL_ADMIN_PASSWORD
require_secret IREDMAIL_MLMMJADMIN_API_TOKEN
require_secret IREDMAIL_ROUNDCUBE_DES_KEY
require_secret IREDMAIL_MYSQL_ROOT_PASSWORD

echo "[iredmail] Secretos presentes. El backend los registrara cifrados al arrancar."
echo "[iredmail] Registrando ahora en PostgreSQL..."
if [ -x "$PROJECT_DIR/backend/bin/register_iredmail_secrets" ]; then
  "$PROJECT_DIR/backend/bin/register_iredmail_secrets"
elif [ -x /tmp/register_iredmail_secrets_linux ]; then
  /tmp/register_iredmail_secrets_linux
elif command -v go >/dev/null 2>&1; then
  cd "$PROJECT_DIR/backend"
  go run ./tools/register_iredmail_secrets
else
  echo "[iredmail] ERROR: no hay binario backend/bin/register_iredmail_secrets ni Go instalado." >&2
  echo "[iredmail] Compila el binario en CI/local y subelo al VPS antes de ejecutar este script." >&2
  exit 1
fi
