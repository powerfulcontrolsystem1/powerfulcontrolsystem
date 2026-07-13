#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
COMPOSE_FILE="${COMPOSE_FILE:-$PROJECT_DIR/deploy/docker-compose.platform.yml}"
ENV_FILE="${ENV_FILE:-$PROJECT_DIR/deploy/.env.platform}"
HEALTH_TIMEOUT_SECONDS="${HEALTH_TIMEOUT_SECONDS:-180}"

get_env_value() {
  local file="$1"
  local key="$2"
  [ -f "$file" ] || return 0
  awk -v key="$key" '
    {
      line=$0
      sub(/\r$/, "", line)
      sub(/^[ \t]+/, "", line)
      if (index(line, key "=") == 1) {
        sub(/^[^=]*=/, "", line)
        gsub(/^["'\'']|["'\'']$/, "", line)
        sub(/\r$/, "", line)
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

test -f "$COMPOSE_FILE" || { echo "ERROR: no existe $COMPOSE_FILE"; exit 1; }
test -f "$ENV_FILE" || { echo "ERROR: no existe $ENV_FILE. Ejecuta primero vps-docker-preflight.sh"; exit 1; }

cd "$PROJECT_DIR"
# Keep the deployment environment compatible with the backend's strict
# encryption-key validation before Compose recreates application containers.
bash deploy/scripts/vps-docker-preflight.sh
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" config >/dev/null

email_enabled="$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_ENABLED)"
mailu_enabled="$(get_env_value "$ENV_FILE" MAILU_ENABLED)"
compose_profiles=()
if [ "$email_enabled" = "1" ] || [ "$email_enabled" = "true" ] || [ "$mailu_enabled" = "1" ] || [ "$mailu_enabled" = "true" ]; then
  compose_profiles+=(--profile mail)
  echo "[sidecar] Perfil mail activo: se levanta Mailu junto al stack."
  if docker ps -a --format '{{.Names}}' | grep -qx 'pcs-iredmail'; then
    echo "[sidecar] Eliminando contenedor antiguo de correo iRedMail."
    docker rm -f pcs-iredmail >/dev/null 2>&1 || true
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
  if [ -z "$provision_mode" ] || printf '%s' "$provision_mode" | grep -Eqi 'ired|mailu_direct|docker_direct|direct_sql'; then
    upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_PROVISION_MODE "mailu_api"
  fi
  provision_command="$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND)"
  if [ -z "$provision_command" ] || printf '%s' "$provision_command" | grep -qi 'ired'; then
    upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND "/app/project_export/deploy/scripts/vps-provision-mailu-mailbox.sh"
  fi
  webmail_url="$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_WEBMAIL_URL)"
  if [ -z "$webmail_url" ] || printf '%s' "$webmail_url" | grep -qi '/mail/\?$'; then
    upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_WEBMAIL_URL "https://mail.powerfulcontrolsystem.com/webmail/"
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
  if [ -z "$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_AUTOLOGIN_SECRET)" ]; then
    if command -v openssl >/dev/null 2>&1; then
      upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_AUTOLOGIN_SECRET "$(openssl rand -hex 32)"
    else
      upsert_env "$ENV_FILE" EMAIL_CORPORATIVO_AUTOLOGIN_SECRET "$(date +%s%N)-mail-autologin"
    fi
  fi
  if [ -z "$(get_env_value "$ENV_FILE" MAILU_SECRET_KEY)" ]; then
    if command -v openssl >/dev/null 2>&1; then
      upsert_env "$ENV_FILE" MAILU_SECRET_KEY "$(openssl rand -hex 32)"
    else
      upsert_env "$ENV_FILE" MAILU_SECRET_KEY "$(date +%s%N)-mailu"
    fi
  fi
  if [ -z "$(get_env_value "$ENV_FILE" MAILU_API_TOKEN)" ]; then
    if command -v openssl >/dev/null 2>&1; then
      upsert_env "$ENV_FILE" MAILU_API_TOKEN "$(openssl rand -hex 32)"
    else
      echo "ERROR: openssl es obligatorio para generar el token interno de Mailu" >&2
      exit 1
    fi
  fi
fi

docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "${compose_profiles[@]}" up -d --build

http_bind="$(get_env_value "$ENV_FILE" HTTP_BIND)"
http_port="$(get_env_value "$ENV_FILE" HTTP_PORT)"
http_bind="${http_bind:-127.0.0.1}"
http_port="${http_port:-8081}"

echo "[sidecar] Contenedores:"
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps

echo "[sidecar] Probando frontend Docker en http://127.0.0.1:$http_port"
ready=0
attempts=$(( HEALTH_TIMEOUT_SECONDS / 5 ))
if [ "$attempts" -lt 6 ]; then
  attempts=6
fi
for attempt in $(seq 1 "$attempts"); do
  backend_status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' pcs-backend 2>/dev/null || true)"
  frontend_status="$(docker inspect -f '{{.State.Status}}' pcs-frontend 2>/dev/null || true)"
  if [ "$backend_status" = "healthy" ] && [ "$frontend_status" = "running" ] && curl -fsS "http://127.0.0.1:$http_port/" >/dev/null 2>&1; then
    ready=1
    break
  fi
  echo "[sidecar] Esperando Docker... intento $attempt/$attempts backend=$backend_status frontend=$frontend_status"
  sleep 5
done

if [ "$ready" != "1" ]; then
  echo "[sidecar] ERROR: el frontend Docker no respondio a tiempo en http://127.0.0.1:$http_port" >&2
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps >&2 || true
  docker logs --tail 80 pcs-backend >&2 || true
  docker logs --tail 80 pcs-frontend >&2 || true
  exit 1
fi

echo "[sidecar] OK. Docker esta arriba en paralelo; Nginx publico aun puede seguir apuntando al servicio actual."

if [ "${#compose_profiles[@]}" -gt 0 ] && docker ps --format '{{.Names}}' | grep -qx 'pcs-mailu-webmail'; then
  docker exec pcs-mailu-webmail sh -lc '
set -eu
config="/data/_data_/_default_/configs/application.ini"
mkdir -p "$(dirname "$config")"
touch "$config"
set_ini_value() {
  section="$1"
  key="$2"
  value="$3"
  if grep -q "^[[:space:]]*$key[[:space:]]*=" "$config"; then
    sed -i "s|^[[:space:]]*$key[[:space:]]*=.*|$key = $value|" "$config"
    return
  fi
  if grep -q "^[[:space:]]*\\[$section\\][[:space:]]*$" "$config"; then
    awk -v section="$section" -v key="$key" -v value="$value" '"'"'
      $0 ~ "^[[:space:]]*\\[" section "\\][[:space:]]*$" { print; print key " = " value; next }
      { print }
    '"'"' "$config" > "$config.tmp" && mv "$config.tmp" "$config"
  else
    printf "\n[%s]\n%s = %s\n" "$section" "$key" "$value" >> "$config"
  fi
}
set_ini_value "webmail" "popup_identity" "Off"
set_ini_value "webmail" "theme" "\"PCSLight@custom\""
set_ini_value "defaults" "theme" "\"PCSLight@custom\""
set_ini_value "security" "secfetch_allow" "\"mode=navigate,dest=iframe,site=same-site\""
theme_source="/pcs-themes"
if [ -d "$theme_source" ]; then
  copied=0
  for candidate in \
    /app/themes \
    /app/snappymail/themes \
    /app/snappymail/v/0.0.0/themes \
    /var/www/html/themes \
    /var/www/html/snappymail/themes \
    /var/www/html/snappymail/v/0.0.0/themes \
    /snappymail/themes \
    /snappymail/v/0.0.0/themes
  do
    if [ -d "$candidate" ]; then
      cp -R "$theme_source"/PCS* "$candidate"/ 2>/dev/null || true
      chown -R mailu:mailu "$candidate"/PCS* 2>/dev/null || true
      copied=1
    fi
  done
  if [ "$copied" = "0" ]; then
    for candidate in $(find / -maxdepth 6 -type d \( -path "*/snappymail/v/0.0.0/themes" -o -path "*/snappymail/themes" -o -path "*/html/themes" \) 2>/dev/null | head -n 6); do
      cp -R "$theme_source"/PCS* "$candidate"/ 2>/dev/null || true
      chown -R mailu:mailu "$candidate"/PCS* 2>/dev/null || true
    done
  fi
fi
' >/dev/null 2>&1 || true
fi
