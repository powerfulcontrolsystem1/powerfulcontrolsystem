#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
COMPOSE_FILE="${COMPOSE_FILE:-$PROJECT_DIR/deploy/docker-compose.platform.yml}"
ENV_FILE="${ENV_FILE:-$PROJECT_DIR/deploy/.env.platform}"
ENV_EXAMPLE="${ENV_EXAMPLE:-$PROJECT_DIR/deploy/.env.platform.example}"
BACKEND_ENV="${BACKEND_ENV:-$PROJECT_DIR/backend/.env.local}"

rand_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32 | tr -d '\n'
    return
  fi
  date +%s%N | sha256sum | awk '{print $1}'
}

rand_config_enc_key() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand 32 | base64 | tr -d '\n'
    return
  fi
  return 1
}

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

upsert_env() {
  local file="$1"
  local key="$2"
  local value="$3"
  if grep -qE "^[[:space:]]*$key=" "$file"; then
    sed -i "s|^[[:space:]]*$key=.*|$key=$value|" "$file"
  else
    printf '%s=%s\n' "$key" "$value" >> "$file"
  fi
}

# Converts the legacy raw/hex effective 32-byte key to canonical Base64 while
# preserving the exact AES key that encrypted existing values. It never prints
# the key and is safe to run repeatedly.
normalize_config_enc_key() {
  local file="$1"
  local value decoded_len normalized
  value="$(get_env_value "$file" CONFIG_ENC_KEY)"
  [ -n "$value" ] || return 0
  decoded_len="$(printf '%s' "$value" | base64 -d 2>/dev/null | wc -c | tr -d ' ')"
  if [ "$decoded_len" = "32" ] && [ "$(printf '%s' "$value" | base64 -d 2>/dev/null | base64 | tr -d '\n')" = "$value" ]; then
    return 0
  fi
  normalized="$(printf '%s' "$value" | head -c 32 | base64 | tr -d '\n')"
  [ "${#normalized}" -gt 0 ] || { echo "[preflight] ERROR: no se pudo normalizar CONFIG_ENC_KEY"; exit 1; }
  upsert_env "$file" CONFIG_ENC_KEY "$normalized"
  chmod 600 "$file" || true
  echo "[preflight] CONFIG_ENC_KEY migrada a Base64 de 32 bytes sin alterar datos cifrados."
}

echo "[preflight] Proyecto: $PROJECT_DIR"
test -f "$COMPOSE_FILE" || { echo "[preflight] ERROR: no existe $COMPOSE_FILE"; exit 1; }
test -f "$ENV_EXAMPLE" || { echo "[preflight] ERROR: no existe $ENV_EXAMPLE"; exit 1; }

if ! command -v docker >/dev/null 2>&1; then
  echo "[preflight] ERROR: Docker no esta instalado en la VPS."
  echo "Instala con: curl -fsSL https://get.docker.com | sh"
  exit 1
fi

docker compose version >/dev/null
echo "[preflight] Docker: $(docker --version)"
echo "[preflight] Compose: $(docker compose version --short)"

if [ ! -f "$ENV_FILE" ]; then
  cp "$ENV_EXAMPLE" "$ENV_FILE"
  chmod 600 "$ENV_FILE"
  upsert_env "$ENV_FILE" POSTGRES_PASSWORD "$(rand_secret)"
  upsert_env "$ENV_FILE" CONFIG_ENC_KEY "$(rand_config_enc_key)"
  upsert_env "$ENV_FILE" ONLYOFFICE_JWT_SECRET "$(rand_secret)"
  mail_admin_password="$(rand_secret)"
  upsert_env "$ENV_FILE" MAILU_ADMIN_PASSWORD "$mail_admin_password"
  upsert_env "$ENV_FILE" MAILU_SECRET_KEY "$(rand_secret)"
  upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_MAILU_API_TOKEN "$(rand_secret)"
  echo "[preflight] Creado $ENV_FILE con secretos nuevos y permisos 600."
fi

for key in GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET GOOGLE_REDIRECT_URL PUBLIC_BASE_URL GEMINI_API_KEY OPENAI_API_KEY GOOGLE_RECAPTCHA_SITE_KEY GOOGLE_RECAPTCHA_SECRET_KEY RECAPTCHA_PROVIDER RECAPTCHA_DEV_BYPASS EMAIL_CORPORATIVO_MAILU_API_TOKEN MAILU_ADMIN_PASSWORD MAILU_SECRET_KEY; do
  current="$(get_env_value "$ENV_FILE" "$key")"
  legacy="$(get_env_value "$BACKEND_ENV" "$key")"
  if [ -z "$current" ] && [ -n "$legacy" ]; then
    upsert_env "$ENV_FILE" "$key" "$legacy"
  fi
done

if [ -f "$BACKEND_ENV" ]; then
  legacy_key="$(get_env_value "$BACKEND_ENV" CONFIG_ENC_KEY)"
  current_key="$(get_env_value "$ENV_FILE" CONFIG_ENC_KEY)"
  if [ -n "$legacy_key" ] && [ "$current_key" != "$legacy_key" ]; then
    upsert_env "$ENV_FILE" CONFIG_ENC_KEY "$legacy_key"
    echo "[preflight] CONFIG_ENC_KEY reutilizada desde backend/.env.local para conservar secretos cifrados."
  fi
fi

normalize_config_enc_key "$ENV_FILE"
if [ -f "$BACKEND_ENV" ]; then
  normalize_config_enc_key "$BACKEND_ENV"
fi

if [ -z "$(get_env_value "$ENV_FILE" MAILU_ADMIN_PASSWORD)" ]; then
  upsert_env "$ENV_FILE" MAILU_ADMIN_PASSWORD "$(rand_secret)"
fi
if [ -z "$(get_env_value "$ENV_FILE" MAILU_SECRET_KEY)" ]; then
  upsert_env "$ENV_FILE" MAILU_SECRET_KEY "$(rand_secret)"
fi
if [ -z "$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_MAILU_API_TOKEN)" ]; then
  upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_MAILU_API_TOKEN "$(rand_secret)"
fi
if [ -z "$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_AUTOLOGIN_SECRET)" ]; then
  upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_AUTOLOGIN_SECRET "$(rand_secret)"
fi
if [ -z "$(get_env_value "$ENV_FILE" MAILU_BIND)" ] || [ "$(get_env_value "$ENV_FILE" MAILU_BIND)" = "0.0.0.0" ]; then
  upsert_env "$ENV_FILE" MAILU_BIND "127.0.0.1"
fi
for pair in MAILU_HTTP_PORT=8089 MAILU_WEBMAIL_PORT=8091 MAILU_HTTPS_PORT=8449 MAILU_SMTP_PORT=2525 MAILU_IMAP_PORT=8143 MAILU_SMTPS_PORT=8465 MAILU_SUBMISSION_PORT=8587 MAILU_IMAPS_PORT=8993; do
  key="${pair%%=*}"
  value="${pair#*=}"
  if [ -z "$(get_env_value "$ENV_FILE" "$key")" ]; then
    upsert_env "$ENV_FILE" "$key" "$value"
  fi
done
if [ -z "$(get_env_value "$ENV_FILE" MAILU_RESOLVER_IP)" ]; then
  upsert_env "$ENV_FILE" MAILU_RESOLVER_IP "192.168.203.254"
fi
for pair in MAILU_REDIS_IP=192.168.203.2 MAILU_SMTP_IP=192.168.203.3 MAILU_ANTISPAM_IP=192.168.203.4 MAILU_WEBMAIL_IP=192.168.203.5 MAILU_IMAP_IP=192.168.203.6 MAILU_ADMIN_IP=192.168.203.7 MAILU_FRONT_IP=192.168.203.8; do
  key="${pair%%=*}"
  value="${pair#*=}"
  if [ -z "$(get_env_value "$ENV_FILE" "$key")" ]; then
    upsert_env "$ENV_FILE" "$key" "$value"
  fi
done
if [ -z "$(get_env_value "$ENV_FILE" MAILU_MESSAGE_SIZE_LIMIT)" ]; then
  upsert_env "$ENV_FILE" MAILU_MESSAGE_SIZE_LIMIT "50000000"
fi
mailu_webmail="$(get_env_value "$ENV_FILE" MAILU_WEBMAIL)"
if [ -z "$mailu_webmail" ] || [ "$mailu_webmail" = "roundcube" ]; then
  upsert_env "$ENV_FILE" MAILU_WEBMAIL "snappymail"
fi
provision_mode="$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_PROVISION_MODE)"
if [ -z "$provision_mode" ] || printf '%s' "$provision_mode" | grep -qi 'ired'; then
  upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_PROVISION_MODE "mailu_direct"
fi
provision_command="$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND)"
if [ -z "$provision_command" ] || printf '%s' "$provision_command" | grep -qi 'ired'; then
  upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND "/app/project_export/deploy/scripts/vps-provision-mailu-mailbox.sh"
fi
if [ -z "$(get_env_value "$ENV_FILE" MAILU_WEB_WEBMAIL)" ]; then
  upsert_env "$ENV_FILE" MAILU_WEB_WEBMAIL "/webmail"
fi
if [ -z "$(get_env_value "$ENV_FILE" MAILU_WEB_ADMIN)" ]; then
  upsert_env "$ENV_FILE" MAILU_WEB_ADMIN "/admin"
fi
if [ -z "$(get_env_value "$ENV_FILE" MAILU_WEBROOT_REDIRECT)" ]; then
  upsert_env "$ENV_FILE" MAILU_WEBROOT_REDIRECT "/webmail"
fi
if [ -z "$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_INTERNAL_SNAPPYMAIL_URL)" ]; then
  upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_INTERNAL_SNAPPYMAIL_URL "http://mailu-webmail/"
fi
if [ "$(get_env_value "$ENV_FILE" MAILU_WEBMAIL_PORT)" = "8090" ]; then
  upsert_env "$ENV_FILE" MAILU_WEBMAIL_PORT "8091"
fi
if [ -z "$(get_env_value "$ENV_FILE" MAILU_PROXY_AUTH_WHITELIST)" ]; then
  upsert_env "$ENV_FILE" MAILU_PROXY_AUTH_WHITELIST "10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
fi
if [ -z "$(get_env_value "$ENV_FILE" MAILU_PROXY_AUTH_HEADER)" ]; then
  upsert_env "$ENV_FILE" MAILU_PROXY_AUTH_HEADER "X-Auth-Email"
fi
if [ -z "$(get_env_value "$ENV_FILE" MAILU_PROXY_AUTH_CREATE)" ]; then
  upsert_env "$ENV_FILE" MAILU_PROXY_AUTH_CREATE "false"
fi
if [ -z "$(get_env_value "$ENV_FILE" MAILU_REAL_IP_FROM)" ]; then
  upsert_env "$ENV_FILE" MAILU_REAL_IP_FROM "10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
fi
if [ -z "$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_INTERNAL_MAILU_API_BASE_URL)" ]; then
  upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_INTERNAL_MAILU_API_BASE_URL "http://mailu-front/api"
fi
if [ -z "$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_INTERNAL_WEBMAIL_URL)" ]; then
  upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_INTERNAL_WEBMAIL_URL "http://mailu-front/webmail/"
fi
if [ -z "$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND)" ]; then
  upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND "/app/project_export/deploy/scripts/vps-provision-mailu-mailbox.sh"
fi
webmail_url="$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_WEBMAIL_URL)"
if [ -z "$webmail_url" ] || printf '%s' "$webmail_url" | grep -qi '/mail/\?$'; then
  upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_WEBMAIL_URL "https://mail.powerfulcontrolsystem.com/webmail/"
fi

echo "[preflight] Variables requeridas presentes:"
for key in POSTGRES_PASSWORD CONFIG_ENC_KEY ONLYOFFICE_JWT_SECRET; do
  value="$(get_env_value "$ENV_FILE" "$key")"
  if [ -z "$value" ] || [[ "$value" == change-me-* ]]; then
    echo "  - $key: FALTA"
  else
    echo "  - $key: OK"
  fi
done

echo "[preflight] Email corporativo/Mailu:"
email_enabled="$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_ENABLED)"
mailu_version="$(get_env_value "$ENV_FILE" MAILU_VERSION)"
mailu_secret="$(get_env_value "$ENV_FILE" MAILU_SECRET_KEY)"
if [ "$email_enabled" = "1" ] || [ "$email_enabled" = "true" ]; then
  [ -n "$mailu_secret" ] && echo "  - secreto Mailu: OK" || echo "  - secreto Mailu: FALTA"
  if [ -n "$mailu_version" ]; then
    echo "  - version Mailu: OK"
  else
    echo "  - version Mailu: FALTA"
  fi
else
  echo "  - modulo: desactivado; los secretos quedan listos para activacion posterior"
fi

echo "[preflight] Servicio actual:"
systemctl is-active powerfulcontrolsystem.service >/dev/null 2>&1 && echo "  - powerfulcontrolsystem.service: activo" || echo "  - powerfulcontrolsystem.service: no activo o no existe"
systemctl is-active nginx >/dev/null 2>&1 && echo "  - nginx: activo" || echo "  - nginx: no activo o no existe"

if command -v ss >/dev/null 2>&1; then
  echo "[preflight] Puertos relevantes:"
  ss -ltnp | awk 'NR==1 || /:80 |:443 |:8080 |:8081 |:8088 |:8090 |:8011 |:2111[5-7] /'
fi

docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" config >/dev/null
echo "[preflight] docker compose config: OK"

echo "[preflight] Listo. Siguiente paso seguro:"
echo "  PROJECT_DIR=$PROJECT_DIR bash $PROJECT_DIR/deploy/scripts/vps-compose-sidecar-up.sh"
echo "[preflight] Para mover tambien 80/443 a Docker, valida DNS y ejecuta despues:"
echo "  CONFIRM_DOCKER_EDGE=YES PROJECT_DIR=$PROJECT_DIR bash $PROJECT_DIR/deploy/scripts/vps-docker-edge-up.sh"
echo "[preflight] Nota: OnlyOffice, voz IA, RustDesk y edge publico estan definidos como perfiles para evitar colisiones durante la migracion."
