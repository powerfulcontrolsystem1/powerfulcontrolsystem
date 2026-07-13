#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
ENV_FILE="${ENV_FILE:-$PROJECT_DIR/deploy/.env.platform}"
SITE_AVAILABLE="${SITE_AVAILABLE:-/etc/nginx/sites-available/00-pcs-mailu}"
SITE_ENABLED="${SITE_ENABLED:-/etc/nginx/sites-enabled/00-pcs-mailu}"
ACME_ROOT="${ACME_ROOT:-/var/www/html}"

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
        gsub(/^["'\''"]|["'\''"]$/, "", line)
        sub(/\r$/, "", line)
        value=line
      }
    }
    END { if (value != "") print value }
  ' "$file"
}

cert_covers_domain() {
  local cert_file="$1"
  local domain="$2"
  [ -f "$cert_file" ] || return 1
  openssl x509 -in "$cert_file" -noout -ext subjectAltName 2>/dev/null | grep -Fq "DNS:$domain"
}

write_http_challenge_site() {
  local domain="$1"
  mkdir -p "$ACME_ROOT/.well-known/acme-challenge"
  cat > "$SITE_AVAILABLE" <<NGINX
server {
    listen 80;
    server_name $domain;
    client_max_body_size 100m;

    location ^~ /.well-known/acme-challenge/ {
        root $ACME_ROOT;
        default_type "text/plain";
        try_files \$uri =404;
    }

    location / {
        return 200 "Mailu proxy preparing TLS for $domain\\n";
        add_header Content-Type text/plain;
    }
}
NGINX
  ln -sf "$SITE_AVAILABLE" "$SITE_ENABLED"
  nginx -t >&2
  systemctl reload nginx >&2
}

ensure_mail_certificate() {
  local domain="$1"
  local fallback_name="$2"
  local preferred_dir="/etc/letsencrypt/live/$domain"
  local fallback_dir="/etc/letsencrypt/live/$fallback_name"

  if cert_covers_domain "$preferred_dir/fullchain.pem" "$domain"; then
    echo "$preferred_dir"
    return 0
  fi
  if cert_covers_domain "$fallback_dir/fullchain.pem" "$domain"; then
    echo "$fallback_dir"
    return 0
  fi
  if ! command -v certbot >/dev/null 2>&1; then
    echo "[mailu-nginx] omitido: certbot no esta instalado y no hay certificado valido para $domain" >&2
    return 1
  fi

  echo "[mailu-nginx] creando certificado Let's Encrypt para $domain" >&2
  write_http_challenge_site "$domain"
  if certbot certonly --webroot -w "$ACME_ROOT" -d "$domain" --agree-tos --non-interactive --register-unsafely-without-email --keep-until-expiring >&2; then
    if cert_covers_domain "$preferred_dir/fullchain.pem" "$domain"; then
      echo "$preferred_dir"
      return 0
    fi
  fi

  echo "[mailu-nginx] omitido: no se pudo emitir certificado valido para $domain" >&2
  return 1
}

test -f "$ENV_FILE" || { echo "[mailu-nginx] omitido: no existe $ENV_FILE"; exit 0; }
command -v nginx >/dev/null 2>&1 || { echo "[mailu-nginx] omitido: nginx del host no esta instalado"; exit 0; }

mail_domain="$(get_env_value "$ENV_FILE" EDGE_MAIL_DOMAIN)"
cert_name="$(get_env_value "$ENV_FILE" EDGE_CERT_NAME)"
http_port="$(get_env_value "$ENV_FILE" MAILU_HTTP_PORT)"
webmail_port="$(get_env_value "$ENV_FILE" MAILU_WEBMAIL_PORT)"
app_http_port="$(get_env_value "$ENV_FILE" HTTP_PORT)"
email_enabled="$(get_env_value "$ENV_FILE" EMAIL_CORPORATIVO_ENABLED)"

mail_domain="${mail_domain:-mail.powerfulcontrolsystem.com}"
cert_name="${cert_name:-powerfulcontrolsystem.com}"
http_port="${http_port:-8089}"
webmail_port="${webmail_port:-8091}"
app_http_port="${app_http_port:-8081}"
wait_seconds="${MAILU_PROXY_WAIT_SECONDS:-180}"

if [ "$email_enabled" != "1" ] && [ "$email_enabled" != "true" ]; then
  echo "[mailu-nginx] omitido: EMAIL_CORPORATIVO_ENABLED no esta activo"
  exit 0
fi

if ! docker ps --format '{{.Names}}' | grep -qx 'pcs-mailu-front'; then
  echo "[mailu-nginx] omitido: contenedor pcs-mailu-front no esta corriendo"
  exit 0
fi

ready=0
attempts=$(( wait_seconds / 5 ))
if [ "$attempts" -lt 6 ]; then
  attempts=6
fi
for attempt in $(seq 1 "$attempts"); do
  if curl -fsS "http://127.0.0.1:$http_port/" >/dev/null 2>&1; then
    ready=1
    break
  fi
  echo "[mailu-nginx] esperando Mailu HTTP 127.0.0.1:$http_port intento $attempt/$attempts"
  sleep 5
done
if [ "$ready" != "1" ]; then
  echo "[mailu-nginx] omitido: Mailu no responde en 127.0.0.1:$http_port"
  exit 0
fi

if ! cert_dir="$(ensure_mail_certificate "$mail_domain" "$cert_name")"; then
  exit 0
fi

cat > "$SITE_AVAILABLE" <<NGINX
server {
    listen 80;
    server_name $mail_domain;
    client_max_body_size 100m;

    location ^~ /.well-known/acme-challenge/ {
        root /var/www/html;
        default_type "text/plain";
        try_files \$uri =404;
    }

    location / {
        return 301 https://\$host\$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name $mail_domain;
    client_max_body_size 100m;

    ssl_certificate $cert_dir/fullchain.pem;
    ssl_certificate_key $cert_dir/privkey.pem;
    ssl_session_timeout 1d;
    ssl_session_cache shared:PCS_MAIL:10m;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers off;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    location = /pcs-mail-autologin {
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_read_timeout 60s;
        proxy_send_timeout 60s;
        proxy_pass http://127.0.0.1:$app_http_port/api/internal/email_corporativo/autologin;
    }

    location = /webmail {
        return 302 /webmail/;
    }

    location ^~ /api/ {
        return 404;
    }

    location ^~ /webmail/ {
        rewrite ^/webmail/?(.*)$ /\$1 break;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_set_header X-Auth-Email "";
        proxy_set_header X-Remote-User "";
        proxy_set_header X-Remote-User-Token "";
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
        proxy_hide_header X-Frame-Options;
        proxy_hide_header Content-Security-Policy;
        add_header Content-Security-Policy "frame-ancestors 'self' https://powerfulcontrolsystem.com" always;
        proxy_pass http://127.0.0.1:$webmail_port;
    }

    location / {
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_set_header X-Auth-Email "";
        proxy_set_header X-Remote-User "";
        proxy_set_header X-Remote-User-Token "";
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
        proxy_hide_header X-Frame-Options;
        proxy_hide_header Content-Security-Policy;
        add_header Content-Security-Policy "frame-ancestors 'self' https://powerfulcontrolsystem.com" always;
        proxy_pass http://127.0.0.1:$http_port;
    }
}
NGINX

ln -sf "$SITE_AVAILABLE" "$SITE_ENABLED"
rm -f /etc/nginx/sites-enabled/00-pcs-iredmail 2>/dev/null || true
nginx -t
systemctl reload nginx
echo "[mailu-nginx] OK: $mail_domain proxy -> 127.0.0.1:$http_port"
