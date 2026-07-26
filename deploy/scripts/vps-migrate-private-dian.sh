#!/usr/bin/env bash
set -euo pipefail

EMPRESA_ID="${EMPRESA_ID:-}"
BACKEND_CONTAINER="${BACKEND_CONTAINER:-pcs-backend}"

if ! printf '%s' "$EMPRESA_ID" | grep -Eq '^[1-9][0-9]*$'; then
  echo "[dian-private] ERROR: EMPRESA_ID debe ser un entero positivo." >&2
  exit 2
fi
if ! docker inspect "$BACKEND_CONTAINER" >/dev/null 2>&1; then
  echo "[dian-private] ERROR: contenedor backend no disponible." >&2
  exit 3
fi

echo "[dian-private] Reparando acceso temporal solo para la firma heredada de la empresa indicada."
docker exec -i -u 0 "$BACKEND_CONTAINER" sh -eu -s -- "$EMPRESA_ID" <<'EOS'
empresa_id="$1"
uploads_root="/app/web/uploads/empresas"
matched=0

if [ ! -d "$uploads_root" ]; then
  exit 0
fi

for signature_dir in "$uploads_root"/empresa_"$empresa_id"_*/facturacion_electronica/firma_electronica; do
  [ -d "$signature_dir" ] || continue
  [ ! -L "$signature_dir" ] || {
    echo "[dian-private] ERROR: se rechazo una carpeta de firma simbolica." >&2
    exit 4
  }
  case "$signature_dir" in
    "$uploads_root"/empresa_"$empresa_id"_*/facturacion_electronica/firma_electronica) ;;
    *)
      echo "[dian-private] ERROR: ruta de firma fuera del alcance permitido." >&2
      exit 5
      ;;
  esac
  chown 10001:10001 "$signature_dir"
  chmod 0700 "$signature_dir"
  find "$signature_dir" -mindepth 1 -maxdepth 1 -type f -name '*.pem' \
    -exec chown 10001:10001 {} + \
    -exec chmod 0600 {} +
  matched=$((matched + 1))
done

printf '[dian-private] Carpetas heredadas preparadas: %s\n' "$matched"
EOS

echo "[dian-private] Simulando migracion DIAN al volumen privado."
docker exec "$BACKEND_CONTAINER" /app/backend/pcs-migrate-private-uploads \
  --web-root=/app/web \
  --empresa-id="$EMPRESA_ID" \
  --category=dian

echo "[dian-private] Aplicando migracion DIAN confirmada al volumen privado."
docker exec "$BACKEND_CONTAINER" /app/backend/pcs-migrate-private-uploads \
  --web-root=/app/web \
  --empresa-id="$EMPRESA_ID" \
  --category=dian \
  --apply \
  --confirm=MIGRATE_PRIVATE_UPLOADS

echo "[dian-private] Verificando que no queden referencias DIAN heredadas elegibles."
docker exec "$BACKEND_CONTAINER" /app/backend/pcs-migrate-private-uploads \
  --web-root=/app/web \
  --empresa-id="$EMPRESA_ID" \
  --category=dian

echo "[dian-private] Migracion privada DIAN completada."
