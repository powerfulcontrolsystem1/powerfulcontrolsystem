#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
BACKEND_DIR="$REPO_ROOT/backend"
CONFIG_PATH="$BACKEND_DIR/secure/vps_security_config.json"

ARGS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --config)
      CONFIG_PATH="$2"
      shift 2
      ;;
    *)
      ARGS+=("$1")
      shift
      ;;
  esac
done

BINARY="$BACKEND_DIR/bin/vps_security_scan_linux_amd64"

cd "$BACKEND_DIR"

if [[ -x "$BINARY" ]]; then
  exec "$BINARY" --config "$CONFIG_PATH" "${ARGS[@]}"
fi

if command -v go >/dev/null 2>&1; then
  exec go run ./tools/vps_security_scan --config "$CONFIG_PATH" "${ARGS[@]}"
fi

echo "[ERROR] No se encontró $BINARY ni Go en el VPS para ejecutar el escaneo." >&2
exit 1