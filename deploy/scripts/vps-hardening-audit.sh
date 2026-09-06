#!/usr/bin/env bash
set -euo pipefail

echo "== PCS VPS hardening audit =="
warnings=0

check_file() {
  local label="$1"
  local path="$2"
  if [ -f "$path" ]; then
    echo "[OK] $label: $path"
  else
    echo "[WARN] $label no encontrado: $path"
    warnings=$((warnings + 1))
  fi
}

check_cmd() {
  local label="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    echo "[OK] $label"
  else
    echo "[WARN] $label"
    warnings=$((warnings + 1))
  fi
}

check_file "sshd_config" /etc/ssh/sshd_config
if effective_ssh="$(sshd -T 2>/dev/null)"; then
  for rule in '^permitrootlogin (no|prohibit-password|without-password)$' '^passwordauthentication no$' '^x11forwarding no$' '^maxauthtries [1-3]$'; do
    if grep -Eq "$rule" <<< "$effective_ssh"; then
      echo "[OK] SSH efectivo: $rule"
    else
      echo "[WARN] SSH efectivo: $rule"
      warnings=$((warnings + 1))
    fi
  done
else
  echo "[WARN] no se pudo consultar sshd -T; cobertura incompleta"
  warnings=$((warnings + 1))
fi

# ufw status exits zero even when the firewall is inactive.
if ufw status 2>/dev/null | grep -qx 'Status: active'; then
  echo "[OK] firewall ufw activo"
else
  echo "[WARN] firewall ufw inactivo o no verificable"
  warnings=$((warnings + 1))
fi
check_cmd "fail2ban activo" systemctl is-active --quiet fail2ban
check_cmd "docker instalado" docker version
check_cmd "docker compose instalado" docker compose version

echo "[INFO] Uso de disco:"
df -h /

echo "[INFO] Contenedores PCS:"
docker ps --format 'table {{.Names}}\t{{.Status}}' | grep -E 'pcs-|NAMES' || true

if [ -f /var/run/reboot-required ]; then
  echo "[WARN] reinicio requerido; comprobar kernel activo y ventana operativa"
  warnings=$((warnings + 1))
fi
for certificate in /etc/letsencrypt/live/*/cert.pem; do
  [ -f "$certificate" ] || continue
  if ! openssl x509 -in "$certificate" -checkend 1209600 -noout >/dev/null 2>&1; then
    echo "[WARN] certificado vencido o vence en menos de 14 dias: $(basename "$(dirname "$certificate")")"
    warnings=$((warnings + 1))
  fi
done
echo "[INFO] warnings=$warnings; revision host parcial, no certificacion integral"
if [ "${1:-}" = "--strict" ] && [ "$warnings" -gt 0 ]; then exit 2; fi
