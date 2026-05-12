#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
COMPOSE_FILE="${COMPOSE_FILE:-$PROJECT_DIR/deploy/docker-compose.platform.yml}"
ENV_FILE="${ENV_FILE:-$PROJECT_DIR/deploy/.env.platform}"
HTTP_ONLY_TEMPLATE="${HTTP_ONLY_TEMPLATE:-./nginx/edge-http-only.conf.template}"
TLS_TEMPLATE="${TLS_TEMPLATE:-./nginx/edge.conf.template}"

if [ "${CONFIRM_DOCKER_EDGE:-}" != "YES" ]; then
  echo "Este script mueve el frente publico 80/443 al stack Docker."
  echo "Detiene Nginx del host si esta activo, emite/renueva certificado Let's Encrypt y levanta pcs-edge."
  echo "Ejecuta con CONFIRM_DOCKER_EDGE=YES solo despues de validar el stack Docker y DNS."
  exit 2
fi

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

compose_base() {
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

test -f "$COMPOSE_FILE" || { echo "ERROR: no existe $COMPOSE_FILE"; exit 1; }
test -f "$ENV_FILE" || { echo "ERROR: no existe $ENV_FILE. Ejecuta deploy/scripts/vps-docker-preflight.sh"; exit 1; }

domain="$(get_env_value "$ENV_FILE" EDGE_DOMAIN)"
www_domain="$(get_env_value "$ENV_FILE" EDGE_WWW_DOMAIN)"
cert_name="$(get_env_value "$ENV_FILE" EDGE_CERT_NAME)"
cert_email="$(get_env_value "$ENV_FILE" EDGE_CERT_EMAIL)"
extra_domains="$(get_env_value "$ENV_FILE" EDGE_CERT_EXTRA_DOMAINS)"

domain="${domain:-powerfulcontrolsystem.com}"
cert_name="${cert_name:-$domain}"

if [ -z "$cert_email" ] || [ "$cert_email" = "admin@powerfulcontrolsystem.com" ]; then
  echo "ERROR: define EDGE_CERT_EMAIL en $ENV_FILE con un correo real para Let's Encrypt." >&2
  exit 1
fi

domain_args=(-d "$domain")
if [ -n "${www_domain:-}" ]; then
  domain_args+=(-d "$www_domain")
fi
if [ -n "${extra_domains:-}" ]; then
  IFS=',' read -r -a extra_array <<< "$extra_domains"
  for item in "${extra_array[@]}"; do
    item="$(printf '%s' "$item" | xargs)"
    if [ -n "$item" ]; then
      domain_args+=(-d "$item")
    fi
  done
fi

cd "$PROJECT_DIR"
compose_base config >/dev/null

nginx_was_active=0
edge_done=0
rollback_host_nginx() {
  if [ "$edge_done" != "1" ] && [ "$nginx_was_active" = "1" ]; then
    echo "[edge] Fallo durante la conmutacion; intentando reactivar Nginx del host..." >&2
    systemctl start nginx >/dev/null 2>&1 || true
  fi
}
trap rollback_host_nginx EXIT

echo "[edge] Levantando nucleo Docker interno..."
compose_base up -d --build postgres backend frontend

if systemctl is-active nginx >/dev/null 2>&1; then
  nginx_was_active=1
  echo "[edge] Deteniendo Nginx del host para liberar puertos 80/443..."
  systemctl stop nginx
fi

echo "[edge] Levantando edge HTTP temporal para ACME..."
EDGE_NGINX_TEMPLATE="$HTTP_ONLY_TEMPLATE" docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" --profile edge up -d edge

echo "[edge] Solicitando certificado Let's Encrypt para: ${domain_args[*]}"
compose_base --profile certbot run --rm certbot certonly \
  --webroot \
  --webroot-path /var/www/certbot \
  --non-interactive \
  --agree-tos \
  --email "$cert_email" \
  --cert-name "$cert_name" \
  "${domain_args[@]}"

echo "[edge] Cambiando edge a HTTPS dentro de Docker..."
EDGE_NGINX_TEMPLATE="$TLS_TEMPLATE" docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" --profile edge up -d --force-recreate edge

echo "[edge] Estado Docker:"
compose_base --profile edge ps

echo "[edge] Verificando HTTPS local contra $domain..."
curl -kfsSI "https://127.0.0.1/" -H "Host: $domain" >/dev/null

edge_done=1
echo "[edge] OK. El trafico publico 80/443 queda dentro de Docker en pcs-edge."
echo "[edge] Programa renovacion con: deploy/scripts/vps-docker-edge-renew.sh"
