#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
RUN_SCRIPT="$REPO_ROOT/scripts/run_vps_security_scan.sh"
CRON_SCHEDULE="${1:-0 2 * * *}"
CRON_LOG="$REPO_ROOT/backend/logs/vps_security/cron.log"
CRON_CMD="cd $REPO_ROOT && /bin/bash $RUN_SCRIPT --trigger cron >> $CRON_LOG 2>&1"

mkdir -p "$(dirname "$CRON_LOG")"

if [[ "${EUID}" -eq 0 ]]; then
  CRON_FILE="/etc/cron.d/powerfulcontrolsystem_vps_security"
  printf '%s root %s\n' "$CRON_SCHEDULE" "$CRON_CMD" > "$CRON_FILE"
  chmod 644 "$CRON_FILE"
  echo "[OK] Cron instalado en $CRON_FILE"
else
  TMP_FILE=$(mktemp)
  crontab -l 2>/dev/null | grep -v 'run_vps_security_scan.sh' > "$TMP_FILE" || true
  printf '%s %s\n' "$CRON_SCHEDULE" "$CRON_CMD" >> "$TMP_FILE"
  crontab "$TMP_FILE"
  rm -f "$TMP_FILE"
  echo "[OK] Cron instalado en el crontab del usuario actual"
fi

echo "[INFO] Programación activa: $CRON_SCHEDULE"