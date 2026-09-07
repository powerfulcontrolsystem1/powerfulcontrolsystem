#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
RETENTION_DAYS="${RETENTION_DAYS:-2}"
INCLUDE_ENV_SECRETS="${INCLUDE_ENV_SECRETS:-0}"
BACKUP_SCOPE="${BACKUP_SCOPE:-vps}"
BACKUP_ROOT="$PROJECT_DIR/backups/vps-snapshots"
STAMP="${BACKUP_STAMP:-$(date +%Y%m%d_%H%M%S)}"
BACKUP_DIR="$BACKUP_ROOT/$STAMP"

case "$BACKUP_SCOPE" in
  database|system|vps) ;;
  *) echo "[ERROR] BACKUP_SCOPE invalido: $BACKUP_SCOPE. Use database, system o vps." >&2; exit 1 ;;
esac
if ! printf '%s' "$STAMP" | grep -Eq '^[0-9]{8}_[0-9]{6}$'; then
  echo "[ERROR] BACKUP_STAMP invalido." >&2
  exit 1
fi
if [ -e "$BACKUP_DIR" ]; then
  echo "[ERROR] El destino del backup ya existe; no se sobrescribe: $BACKUP_DIR" >&2
  exit 1
fi

mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_ROOT" "$BACKUP_DIR" 2>/dev/null || true

echo "[INFO] Backup VPS: alcance $BACKUP_SCOPE; destino $BACKUP_DIR"

{
  printf 'created_at=%s\n' "$(date -Is)"
  printf 'scope=%s\n' "$BACKUP_SCOPE"
  printf 'project_dir=%s\n' "$PROJECT_DIR"
  uname -a 2>/dev/null || true
  docker version --format 'docker_server={{.Server.Version}}' 2>/dev/null || true
  docker ps --format 'container={{.Names}} image={{.Image}} status={{.Status}}' 2>/dev/null || true
  df -hT / 2>/dev/null || true
} > "$BACKUP_DIR/MANIFEST.txt"

if docker ps --format '{{.Names}}' | grep -qx 'pcs-postgres'; then
  echo "[INFO] Backup VPS: generando pg_dumpall desde pcs-postgres."
  docker exec pcs-postgres sh -lc 'pg_dumpall -U "$POSTGRES_USER"' > "$BACKUP_DIR/postgres_all.sql"
  gzip -9 "$BACKUP_DIR/postgres_all.sql"
else
  echo "[ERROR] Backup VPS: pcs-postgres no esta activo; no se puede crear una copia recuperable." >&2
  exit 1
fi

if [ "$BACKUP_SCOPE" = "vps" ] && docker ps --format '{{.Names}}' | grep -qx 'pcs-nextcloud-db'; then
  echo "[INFO] Backup VPS: generando dump logico de Nextcloud/MariaDB."
  docker exec pcs-nextcloud-db sh -lc 'exec mariadb-dump -u root -p"$MYSQL_ROOT_PASSWORD" --all-databases --single-transaction --routines --events' \
    | gzip -9 > "$BACKUP_DIR/nextcloud_all.sql.gz"
  test -s "$BACKUP_DIR/nextcloud_all.sql.gz"
fi

if [ "$BACKUP_SCOPE" != "database" ]; then
  echo "[INFO] Backup VPS: empaquetando sistema PCS sin secretos ni copias anidadas."
  parent="$(dirname "$PROJECT_DIR")"
  base="$(basename "$PROJECT_DIR")"
  tar -czf "$BACKUP_DIR/pcs_system.tar.gz" \
    --exclude="$base/.git" \
    --exclude="$base/deploy/.env.platform" \
    --exclude="$base/backups" \
    --exclude="$base/backup/vps_snapshots" \
    --exclude="$base/.gotmp" \
    --exclude="$base/.gocache" \
    --exclude="$base/node_modules" \
    -C "$parent" "$base"
fi

volumes=""
if [ "$BACKUP_SCOPE" = "system" ]; then
  volumes="
powerful-control-system_pcs_web_uploads
powerful-control-system_pcs_downloads
powerful-control-system_pcs_backend_logs
powerful-control-system_pcs_backups
powerful-control-system_pcs_private_storage
"
elif [ "$BACKUP_SCOPE" = "vps" ]; then
  required_vps_volumes="
powerful-control-system_pcs_web_uploads
powerful-control-system_pcs_downloads
powerful-control-system_pcs_private_storage
powerful-control-system_mailu_certs
powerful-control-system_pcs_onlyoffice_data
powerful-control-system_pcs_onlyoffice_lib
powerful-control-system_pcs_onlyoffice_logs
"
  volumes="$({ printf '%s\n' "$required_vps_volumes"; docker inspect $(docker ps -q) --format '{{range .Mounts}}{{if eq .Type "volume"}}{{println .Name}}{{end}}{{end}}' 2>/dev/null || true; } \
    | sort -u \
    | grep -E '^(powerful-control-system_|nextcloud_|onlyoffice_|pcs_|pcs-)' \
    | grep -Ev '(postgres_data|nextcloud_db)$' || true)"
fi

for volume in $volumes; do
  if docker volume inspect "$volume" >/dev/null 2>&1; then
    echo "[INFO] Backup VPS: empaquetando volumen $volume"
    docker run --rm -v "$volume:/volume:ro" -v "$BACKUP_DIR:/backup" alpine:3.20 sh -lc "cd /volume && tar -czf /backup/$volume.tar.gz ."
  else
    echo "[INFO] Backup VPS: volumen no encontrado, omitido: $volume"
  fi
done

# Un backup debe fallar de forma visible si faltan sus artefactos minimos.
required_files="postgres_all.sql.gz"
if [ "$BACKUP_SCOPE" != "database" ]; then
  required_files="$required_files pcs_system.tar.gz powerful-control-system_pcs_web_uploads.tar.gz powerful-control-system_pcs_downloads.tar.gz powerful-control-system_pcs_private_storage.tar.gz"
fi
if [ "$BACKUP_SCOPE" = "vps" ]; then
  required_files="$required_files powerful-control-system_mailu_certs.tar.gz powerful-control-system_pcs_onlyoffice_data.tar.gz powerful-control-system_pcs_onlyoffice_lib.tar.gz powerful-control-system_pcs_onlyoffice_logs.tar.gz"
fi
for required in $required_files; do
  if [ ! -s "$BACKUP_DIR/$required" ]; then
    echo "[ERROR] Backup VPS incompleto: falta $required"
    exit 1
  fi
done

if [ "$INCLUDE_ENV_SECRETS" = "1" ] && [ -f "$PROJECT_DIR/deploy/.env.platform" ]; then
  cp "$PROJECT_DIR/deploy/.env.platform" "$BACKUP_DIR/env.platform.backup"
  chmod 600 "$BACKUP_DIR/env.platform.backup" 2>/dev/null || true
fi

(
  cd "$BACKUP_DIR"
  find . -maxdepth 1 -type f ! -name SHA256SUMS -printf '%P\0' \
    | sort -z \
    | xargs -0 -r sha256sum > SHA256SUMS
)

find "$BACKUP_ROOT" -mindepth 1 -maxdepth 1 -type d -mtime +"$RETENTION_DAYS" -print -exec rm -rf {} \; 2>/dev/null || true
du -sh "$BACKUP_DIR" 2>/dev/null || true
echo "[OK] Backup VPS completado: $BACKUP_DIR"
