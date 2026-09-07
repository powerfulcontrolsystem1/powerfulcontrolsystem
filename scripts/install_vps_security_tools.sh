#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
BACKEND_DIR="$REPO_ROOT/backend"
CONFIG_TEMPLATE="$BACKEND_DIR/vpssecurity/config/default_vps_security_config.json"
CONFIG_TARGET="$BACKEND_DIR/secure/vps_security_config.json"

if [[ "${EUID}" -ne 0 ]]; then
  SUDO="sudo"
else
  SUDO=""
fi

echo "[INFO] Instalando herramientas base de seguridad VPS..."
$SUDO apt-get update
$SUDO apt-get install -y lynis nmap curl ca-certificates gnupg lsb-release cron apt-transport-https

if ! command -v trivy >/dev/null 2>&1; then
  echo "[INFO] Instalando Trivy (alternativa ligera a OpenVAS)..."
  curl -fsSL https://aquasecurity.github.io/trivy-repo/deb/public.key | $SUDO gpg --dearmor -o /usr/share/keyrings/trivy.gpg
  echo "deb [signed-by=/usr/share/keyrings/trivy.gpg] https://aquasecurity.github.io/trivy-repo/deb generic main" | $SUDO tee /etc/apt/sources.list.d/trivy.list >/dev/null
  $SUDO apt-get update
  $SUDO apt-get install -y trivy
fi

mkdir -p "$BACKEND_DIR/logs/vps_security/runs"
mkdir -p "$BACKEND_DIR/secure"

if [[ -f "$CONFIG_TEMPLATE" && ! -f "$CONFIG_TARGET" ]]; then
  cp "$CONFIG_TEMPLATE" "$CONFIG_TARGET"
  echo "[OK] Configuración inicial copiada a $CONFIG_TARGET"
fi

echo "[OK] Herramientas instaladas."
echo "[INFO] Siguiente paso sugerido: $REPO_ROOT/scripts/install_vps_security_cron.sh"