#!/usr/bin/env bash
# Compuerta de solo lectura para el candidato P110 en staging.
# No despliega, migra, emite documentos, copia secretos ni modifica alertas.
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "$0")/../.." && pwd)}"
HTTP_PORT="${HTTP_PORT:-8082}"
EMPRESA_ID="${EMPRESA_ID:-12}"

blocked=0

run_gate() {
  local name="$1"
  shift
  echo "[INFO] Ejecutando ${name}"
  set +e
  "$@"
  local code=$?
  set -e
  if [[ "$code" -ne 0 ]]; then
    echo "[BLOCKED] ${name} devolvio ${code}."
    blocked=1
  fi
}

for endpoint in /health /ready; do
  if curl -fsS --max-time 10 "http://127.0.0.1:${HTTP_PORT}${endpoint}" >/dev/null; then
    echo "[OK] staging${endpoint}"
  else
    echo "[BLOCKED] staging${endpoint} no responde."
    blocked=1
  fi
done

run_gate "paridad DIAN" env EMPRESA_ID="$EMPRESA_ID" REQUIRE_STAGING_COMPLETE=1 \
  bash "$PROJECT_DIR/deploy/scripts/vps-dian-parity-audit.sh"
run_gate "entrega externa Alertmanager" \
  bash "$PROJECT_DIR/deploy/scripts/vps-alertmanager-delivery-audit.sh"

if [[ "$blocked" -ne 0 ]]; then
  echo "[NO-GO] El candidato no cumple las compuertas P110 automaticas."
  exit 2
fi

echo "[OK] Compuertas automaticas P110 aprobadas."
echo "[INFO] Aun se requieren pruebas DIAN oficiales, carga, roles, UAT e impresion fisica."
