#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
CRON_FILE="${CRON_FILE:-/etc/cron.d/pcs-vps-external-backup}"
SCHEDULE="${SCHEDULE:-25 3 * * *}"
ENV_FILE="${ENV_FILE:-/etc/pcs-external-backup.env}"

if [ ! -s "$ENV_FILE" ]; then
  echo "[ERROR] Falta $ENV_FILE. Configure el destino externo antes de activar el cron." >&2
  exit 1
fi
mode="$(stat -c '%a' "$ENV_FILE" 2>/dev/null || true)"
if [ "$mode" != "600" ]; then
  echo "[ERROR] $ENV_FILE debe tener permisos 600; permisos actuales: ${mode:-desconocidos}." >&2
  exit 1
fi
if ! grep -Eq '^EXTERNAL_BACKUP_TARGET=(rclone|s3)$' "$ENV_FILE"; then
  echo "[ERROR] EXTERNAL_BACKUP_TARGET debe ser rclone o s3." >&2
  exit 1
fi

cat > "$CRON_FILE" <<EOF
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
PROJECT_DIR=$PROJECT_DIR
$SCHEDULE root set -a; . $ENV_FILE; set +a; cd $PROJECT_DIR && bash deploy/scripts/vps-external-backup.sh >> $PROJECT_DIR/backups/vps-external-backup.log 2>&1
EOF

chmod 644 "$CRON_FILE"
echo "[OK] Cron de backup externo instalado en $CRON_FILE"
