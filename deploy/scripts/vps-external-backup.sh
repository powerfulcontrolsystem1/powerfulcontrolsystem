#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
BACKUP_ROOT="$PROJECT_DIR/backups/vps-snapshots"
TARGET="${EXTERNAL_BACKUP_TARGET:-none}"
RCLONE_REMOTE="${RCLONE_REMOTE:-}"
S3_URI="${S3_URI:-}"
KEEP_LOCAL_DAYS="${KEEP_LOCAL_DAYS:-14}"

cd "$PROJECT_DIR"

bash deploy/scripts/vps-backup-operacion.sh

latest="$(find "$BACKUP_ROOT" -mindepth 1 -maxdepth 1 -type d -printf '%T@ %p\n' | sort -nr | awk 'NR==1 {print $2}')"
if [ -z "${latest:-}" ] || [ ! -d "$latest" ]; then
  echo "[ERROR] No se encontro backup local reciente en $BACKUP_ROOT" >&2
  exit 1
fi

case "$TARGET" in
  none|"")
    echo "[INFO] EXTERNAL_BACKUP_TARGET=none. Backup externo omitido."
    ;;
  rclone)
    if [ -z "$RCLONE_REMOTE" ]; then
      echo "[ERROR] Defina RCLONE_REMOTE, por ejemplo: myremote:pcs-backups" >&2
      exit 1
    fi
    command -v rclone >/dev/null 2>&1 || { echo "[ERROR] rclone no esta instalado." >&2; exit 1; }
    rclone copy "$latest" "$RCLONE_REMOTE/$(basename "$latest")" --fast-list --transfers 4 --checkers 8
    echo "[OK] Backup externo enviado con rclone a $RCLONE_REMOTE/$(basename "$latest")"
    ;;
  s3)
    if [ -z "$S3_URI" ]; then
      echo "[ERROR] Defina S3_URI, por ejemplo: s3://bucket/powerfulcontrolsystem" >&2
      exit 1
    fi
    command -v aws >/dev/null 2>&1 || { echo "[ERROR] aws cli no esta instalado." >&2; exit 1; }
    aws s3 sync "$latest" "$S3_URI/$(basename "$latest")" --only-show-errors
    echo "[OK] Backup externo enviado a $S3_URI/$(basename "$latest")"
    ;;
  *)
    echo "[ERROR] EXTERNAL_BACKUP_TARGET invalido: $TARGET. Use none, rclone o s3." >&2
    exit 1
    ;;
esac

find "$BACKUP_ROOT" -mindepth 1 -maxdepth 1 -type d -mtime +"$KEEP_LOCAL_DAYS" -print -exec rm -rf {} \; 2>/dev/null || true
