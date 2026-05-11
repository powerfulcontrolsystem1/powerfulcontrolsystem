#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
CRON_FILE="${CRON_FILE:-/etc/cron.d/pcs-vps-external-backup}"
SCHEDULE="${SCHEDULE:-25 3 * * *}"

cat > "$CRON_FILE" <<EOF
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
PROJECT_DIR=$PROJECT_DIR
$SCHEDULE root set -a; [ -f /etc/pcs-external-backup.env ] && . /etc/pcs-external-backup.env; set +a; cd $PROJECT_DIR && bash deploy/scripts/vps-external-backup.sh >> $PROJECT_DIR/backups/vps-external-backup.log 2>&1
EOF

chmod 644 "$CRON_FILE"
echo "[OK] Cron de backup externo instalado en $CRON_FILE"
