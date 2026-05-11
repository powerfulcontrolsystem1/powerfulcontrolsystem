#!/usr/bin/env bash
set -euo pipefail

echo "== PCS VPS hardening audit =="

check_file() {
  local label="$1"
  local path="$2"
  if [ -f "$path" ]; then
    echo "[OK] $label: $path"
  else
    echo "[WARN] $label no encontrado: $path"
  fi
}

check_cmd() {
  local label="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    echo "[OK] $label"
  else
    echo "[WARN] $label"
  fi
}

check_file "sshd_config" /etc/ssh/sshd_config
if [ -f /etc/ssh/sshd_config ]; then
  grep -Eiq '^\s*PermitRootLogin\s+(no|prohibit-password)' /etc/ssh/sshd_config && echo "[OK] root login SSH restringido" || echo "[WARN] revisa PermitRootLogin"
  grep -Eiq '^\s*PasswordAuthentication\s+no' /etc/ssh/sshd_config && echo "[OK] autenticacion por password SSH desactivada" || echo "[WARN] considera PasswordAuthentication no"
fi

check_cmd "firewall ufw activo" ufw status
check_cmd "fail2ban disponible" systemctl is-enabled fail2ban
check_cmd "docker instalado" docker version
check_cmd "docker compose instalado" docker compose version

echo "[INFO] Uso de disco:"
df -h /

echo "[INFO] Contenedores PCS:"
docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' | grep -E 'pcs-|NAMES' || true
