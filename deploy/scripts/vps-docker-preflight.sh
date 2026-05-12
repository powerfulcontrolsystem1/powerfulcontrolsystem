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
  upsert_env "$ENV_FILE" CONFIG_ENC_KEY "$(rand_secret)"
  upsert_env "$ENV_FILE" ONLYOFFICE_JWT_SECRET "$(rand_secret)"
  upsert_env "$ENV_FILE" NEXTCLOUD_DB_PASSWORD "$(rand_secret)"
  upsert_env "$ENV_FILE" NEXTCLOUD_ADMIN_PASSWORD "$(rand_secret)"
  echo "[preflight] Creado $ENV_FILE con secretos nuevos y permisos 600."
fi

for key in GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET GOOGLE_REDIRECT_URL PUBLIC_BASE_URL GEMINI_API_KEY OPENAI_API_KEY GOOGLE_RECAPTCHA_SITE_KEY GOOGLE_RECAPTCHA_SECRET_KEY RECAPTCHA_PROVIDER RECAPTCHA_DEV_BYPASS; do
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

echo "[preflight] Variables requeridas presentes:"
for key in POSTGRES_PASSWORD CONFIG_ENC_KEY ONLYOFFICE_JWT_SECRET NEXTCLOUD_DB_PASSWORD NEXTCLOUD_ADMIN_PASSWORD; do
  value="$(get_env_value "$ENV_FILE" "$key")"
  if [ -z "$value" ] || [[ "$value" == change-me-* ]]; then
    echo "  - $key: FALTA"
  else
    echo "  - $key: OK"
  fi
done

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
echo "[preflight] Nota: OnlyOffice, Nextcloud, voz IA y RustDesk estan definidos como perfiles para evitar colisiones con servicios existentes."
