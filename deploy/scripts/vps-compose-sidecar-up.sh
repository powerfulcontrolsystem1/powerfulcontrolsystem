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
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" config >/dev/null

email_enabled="$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_ENABLED)"
iredmail_enabled="$(get_env_value "$ENV_FILE" IREDMAIL_ENABLED)"
compose_profiles=()
if [ "$email_enabled" = "1" ] || [ "$email_enabled" = "true" ] || [ "$iredmail_enabled" = "1" ] || [ "$iredmail_enabled" = "true" ]; then
  compose_profiles+=(--profile mail)
  echo "[sidecar] Perfil mail activo: se levanta iRedMail junto al stack."
  if [ -z "$(get_env_value "$ENV_FILE" IREDMAIL_BIND)" ] || [ "$(get_env_value "$ENV_FILE" IREDMAIL_BIND)" = "0.0.0.0" ]; then
    upsert_env "$ENV_FILE" IREDMAIL_BIND "127.0.0.1"
  fi
  for pair in IREDMAIL_HTTP_PORT=8089 IREDMAIL_HTTPS_PORT=8449 IREDMAIL_SMTP_PORT=2525 IREDMAIL_POP3_PORT=8110 IREDMAIL_IMAP_PORT=8143 IREDMAIL_SMTPS_PORT=8465 IREDMAIL_SUBMISSION_PORT=8587 IREDMAIL_IMAPS_PORT=8993 IREDMAIL_POP3S_PORT=8995 IREDMAIL_SIEVE_PORT=8419; do
    key="${pair%%=*}"
    value="${pair#*=}"
    if [ -z "$(get_env_value "$ENV_FILE" "$key")" ]; then
      upsert_env "$ENV_FILE" "$key" "$value"
    fi
  done
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
