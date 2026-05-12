#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"
INCLUDE_ENV_SECRETS="${INCLUDE_ENV_SECRETS:-0}"
BACKUP_ROOT="$PROJECT_DIR/backups/vps-snapshots"
STAMP="$(date +%Y%m%d_%H%M%S)"
BACKUP_DIR="$BACKUP_ROOT/$STAMP"

mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_ROOT" "$BACKUP_DIR" 2>/dev/null || true

echo "[INFO] Backup VPS: destino $BACKUP_DIR"

if docker ps --format '{{.Names}}' | grep -qx 'pcs-postgres'; then
  echo "[INFO] Backup VPS: generando pg_dumpall desde pcs-postgres."
  docker exec pcs-postgres sh -lc 'pg_dumpall -U "$POSTGRES_USER"' > "$BACKUP_DIR/postgres_all.sql"
  gzip -9 "$BACKUP_DIR/postgres_all.sql"
else
  echo "[WARN] Backup VPS: pcs-postgres no esta activo; se omite dump PostgreSQL Docker."
fi

volumes="
powerful-control-system_pcs_web_uploads
powerful-control-system_pcs_downloads
powerful-control-system_pcs_backend_logs
powerful-control-system_pcs_backups
powerful-control-system_pcs_postgres_data
powerful-control-system_pcs_letsencrypt
powerful-control-system_pcs_certbot_www
"

for volume in $volumes; do
  if docker volume inspect "$volume" >/dev/null 2>&1; then
    echo "[INFO] Backup VPS: empaquetando volumen $volume"
    docker run --rm -v "$volume:/volume:ro" -v "$BACKUP_DIR:/backup" alpine:3.20 sh -lc "cd /volume && tar -czf /backup/$volume.tar.gz ."
  else
    echo "[INFO] Backup VPS: volumen no encontrado, omitido: $volume"
  fi
done

if [ "$INCLUDE_ENV_SECRETS" = "1" ] && [ -f "$PROJECT_DIR/deploy/.env.platform" ]; then
  cp "$PROJECT_DIR/deploy/.env.platform" "$BACKUP_DIR/env.platform.backup"
  chmod 600 "$BACKUP_DIR/env.platform.backup" 2>/dev/null || true
fi

find "$BACKUP_ROOT" -mindepth 1 -maxdepth 1 -type d -mtime +"$RETENTION_DAYS" -print -exec rm -rf {} \; 2>/dev/null || true
du -sh "$BACKUP_DIR" 2>/dev/null || true
echo "[OK] Backup VPS completado: $BACKUP_DIR"
