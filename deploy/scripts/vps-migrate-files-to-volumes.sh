#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"

if [ "${CONFIRM_FILE_MIGRATE:-}" != "YES" ]; then
  echo "Este script copia archivos persistentes actuales a volumenes Docker."
  echo "Ejecuta con CONFIRM_FILE_MIGRATE=YES despues de tener Docker levantado."
  exit 2
fi

copy_dir_to_volume() {
  local source_dir="$1"
  local volume_name="$2"
  local target_subdir="${3:-}"

  if [ ! -d "$PROJECT_DIR/$source_dir" ]; then
    echo "[files] Omitido $source_dir: no existe"
    return
  fi

  echo "[files] Copiando $source_dir -> $volume_name${target_subdir:+/$target_subdir}"
  docker run --rm \
    -v "$PROJECT_DIR/$source_dir:/source:ro" \
    -v "$volume_name:/target" \
    alpine:3.20 sh -c "set -e; mkdir -p \"/target/$target_subdir\"; cp -a /source/. \"/target/$target_subdir/\""
}

copy_dir_to_volume "web/uploads" "powerful-control-system_pcs_web_uploads"
copy_dir_to_volume "descargas" "powerful-control-system_pcs_downloads"
copy_dir_to_volume "backend/logs" "powerful-control-system_pcs_backend_logs"
copy_dir_to_volume "backup" "powerful-control-system_pcs_backups" "backup"
copy_dir_to_volume "backups" "powerful-control-system_pcs_backups" "backups"

echo "[files] OK. Archivos persistentes copiados a volumenes Docker."
