#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
CRON_FILE="/etc/cron.d/pcs-vps-backup"
RUNNER="/usr/local/bin/pcs_vps_backup_daily.sh"
LOG_DIR="$PROJECT_DIR/backups/logs"

mkdir -p "$LOG_DIR"
install -m 0750 "$PROJECT_DIR/deploy/scripts/vps-backup-operacion.sh" "$RUNNER"

cat > "$CRON_FILE" <<EOF
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
15 3 * * * root PROJECT_DIR=$PROJECT_DIR RETENTION_DAYS=14 $RUNNER >> $LOG_DIR/vps-backup-\$(date +\%Y\%m\%d).log 2>&1
EOF

chmod 0644 "$CRON_FILE"
systemctl reload cron 2>/dev/null || systemctl reload crond 2>/dev/null || true
echo "[OK] Cron de backup instalado: $CRON_FILE"
