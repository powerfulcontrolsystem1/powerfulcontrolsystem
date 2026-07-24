#!/usr/bin/env bash
set -euo pipefail

SITE_AVAILABLE="${SITE_AVAILABLE:-/etc/nginx/sites-available/nextcloud-powerfulcontrolsystem}"
SITE_ENABLED="${SITE_ENABLED:-/etc/nginx/sites-enabled/nextcloud-powerfulcontrolsystem}"
NEXTCLOUD_DOMAIN="${NEXTCLOUD_DOMAIN:-nextcloud.powerfulcontrolsystem.com}"
NEXTCLOUD_UPSTREAM="${NEXTCLOUD_UPSTREAM:-127.0.0.1:8090}"
EMBED_ORIGIN="${EMBED_ORIGIN:-https://powerfulcontrolsystem.com}"
MARKER_BEGIN="# PCS_EMBED_POLICY_BEGIN"
MARKER_END="# PCS_EMBED_POLICY_END"

fail() {
  echo "[nextcloud-nginx] ERROR: $*" >&2
  exit 1
}

[ "${EUID}" -eq 0 ] || fail "debe ejecutarse como root"
command -v nginx >/dev/null 2>&1 || fail "nginx no esta instalado en el host"
[ -f "$SITE_AVAILABLE" ] || fail "no existe $SITE_AVAILABLE"
[[ "$NEXTCLOUD_DOMAIN" =~ ^[a-z0-9.-]+$ ]] || fail "NEXTCLOUD_DOMAIN invalido"
[[ "$NEXTCLOUD_UPSTREAM" =~ ^127\.0\.0\.1:[0-9]{2,5}$ ]] || fail "NEXTCLOUD_UPSTREAM debe usar loopback"
[[ "$EMBED_ORIGIN" =~ ^https://[a-z0-9.-]+$ ]] || fail "EMBED_ORIGIN invalido"

grep -Eq "^[[:space:]]*server_name[[:space:]]+${NEXTCLOUD_DOMAIN//./\\.}([[:space:]]|;)" "$SITE_AVAILABLE" \
  || fail "el sitio no pertenece a $NEXTCLOUD_DOMAIN"
grep -Fq "proxy_pass http://$NEXTCLOUD_UPSTREAM;" "$SITE_AVAILABLE" \
  || fail "el sitio no apunta al upstream esperado $NEXTCLOUD_UPSTREAM"

# Nextcloud emite una CSP dinamica con nonce. Nginx estandar no puede sustituir
# solo frame-ancestors sin ocultar esa CSP (lo que rompería sus nonces). Si el
# upstream ya declara la directiva, agregar otro encabezado la intersecta y no
# habilita el iframe. Fallar antes de escribir conserva la configuracion segura.
if curl -kfsSI --max-time 15 "https://$NEXTCLOUD_DOMAIN/" | tr -d '\r' | grep -qi '^content-security-policy:.*frame-ancestors'; then
  fail "Nextcloud ya emite frame-ancestors; no se puede ampliar de forma segura con Nginx estandar. Configure una politica soportada por Nextcloud o un filtro de encabezados que preserve nonces antes de reintentar"
fi

if grep -Fq "$MARKER_BEGIN" "$SITE_AVAILABLE"; then
  grep -Fq "frame-ancestors 'self' $EMBED_ORIGIN" "$SITE_AVAILABLE" \
    || fail "marcador existente con origen de iframe distinto"
  nginx -t
  systemctl reload nginx
  echo "[nextcloud-nginx] OK: politica integrada ya estaba aplicada"
  exit 0
fi

backup="${SITE_AVAILABLE}.bak.$(date -u +%Y%m%dT%H%M%SZ)"
tmp_file="$(mktemp "${SITE_AVAILABLE}.tmp.XXXXXX")"
trap 'rm -f "$tmp_file"' EXIT
cp -a "$SITE_AVAILABLE" "$backup"

awk -v marker_begin="$MARKER_BEGIN" -v marker_end="$MARKER_END" -v origin="$EMBED_ORIGIN" '
  {
    print
    if ($0 ~ /^[[:space:]]*listen[[:space:]]+443([[:space:];]|$)/ && $0 ~ /ssl/) {
      indent = $0
      sub(/[^[:space:]].*$/, "", indent)
      print indent "    " marker_begin
      print indent "    add_header Content-Security-Policy \"frame-ancestors '\''self'\'' " origin "\" always;"
      print indent "    " marker_end
      inserted += 1
    }
  }
  END {
    if (inserted != 1) exit 42
  }
' "$SITE_AVAILABLE" > "$tmp_file" || fail "se esperaba exactamente un bloque HTTPS en el sitio"

install -m 0644 "$tmp_file" "$SITE_AVAILABLE"
ln -sfn "$SITE_AVAILABLE" "$SITE_ENABLED"

if ! nginx -t; then
  cp -a "$backup" "$SITE_AVAILABLE"
  nginx -t || true
  fail "nginx rechazo la configuracion; se restauro $backup"
fi

systemctl reload nginx
echo "[nextcloud-nginx] OK: $NEXTCLOUD_DOMAIN permite iframe solo desde $EMBED_ORIGIN"
echo "[nextcloud-nginx] backup: $backup"
