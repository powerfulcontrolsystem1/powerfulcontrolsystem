#!/usr/bin/env bash
set -euo pipefail

EMPRESA_ID="${EMPRESA_ID:-}"
BACKEND_CONTAINER="${BACKEND_CONTAINER:-$(docker ps --filter label=com.docker.compose.project=powerful-control-system --filter label=com.docker.compose.service=backend --format '{{.ID}}' | head -n 1)}"

if ! printf '%s' "$EMPRESA_ID" | grep -Eq '^[1-9][0-9]*$'; then
  echo "[dian-private] ERROR: EMPRESA_ID debe ser un entero positivo." >&2
  exit 2
fi
if ! docker inspect "$BACKEND_CONTAINER" >/dev/null 2>&1; then
  echo "[dian-private] ERROR: contenedor backend no disponible." >&2
  exit 3
fi

echo "[dian-private] Reparando acceso temporal solo para la firma heredada de la empresa indicada."
uploads_source="$(
  docker inspect "$BACKEND_CONTAINER" \
    --format '{{range .Mounts}}{{if eq .Destination "/app/web/uploads"}}{{.Source}}{{end}}{{end}}'
)"
if [ -z "$uploads_source" ] || [ ! -d "$uploads_source" ]; then
  echo "[dian-private] ERROR: no se encontro el volumen de uploads del backend." >&2
  exit 4
fi
case "$uploads_source" in
  /*) ;;
  *)
    echo "[dian-private] ERROR: el origen del volumen de uploads no es absoluto." >&2
    exit 5
    ;;
esac

uploads_root="$uploads_source/empresas"
uploads_source_real="$(readlink -f "$uploads_source")"
uploads_root_real="$(readlink -f "$uploads_root")"
case "$uploads_root_real" in
  "$uploads_source_real"/empresas) ;;
  *)
    echo "[dian-private] ERROR: la carpeta empresas escapa del volumen de uploads." >&2
    exit 6
    ;;
esac

matched=0

if [ ! -d "$uploads_root" ]; then
  echo "[dian-private] Carpetas heredadas preparadas: 0"
else
  for signature_dir in "$uploads_root"/empresa_"$EMPRESA_ID"_*/facturacion_electronica/firma_electronica; do
    [ -d "$signature_dir" ] || continue
    [ ! -L "$signature_dir" ] || {
      echo "[dian-private] ERROR: se rechazo una carpeta de firma simbolica." >&2
      exit 7
    }
    signature_real="$(readlink -f "$signature_dir")"
    case "$signature_real" in
      "$uploads_root_real"/empresa_"$EMPRESA_ID"_*/facturacion_electronica/firma_electronica) ;;
      *)
        echo "[dian-private] ERROR: ruta de firma fuera del alcance permitido." >&2
        exit 8
        ;;
    esac
    facturacion_dir="$(dirname "$signature_dir")"
    [ ! -L "$facturacion_dir" ] || {
      echo "[dian-private] ERROR: se rechazo una carpeta de facturacion simbolica." >&2
      exit 9
    }
    facturacion_real="$(readlink -f "$facturacion_dir")"
    case "$facturacion_real" in
      "$uploads_root_real"/empresa_"$EMPRESA_ID"_*/facturacion_electronica) ;;
      *)
        echo "[dian-private] ERROR: carpeta de facturacion fuera del alcance permitido." >&2
        exit 10
        ;;
    esac
    chown 10001:10001 "$facturacion_dir"
    chmod 0700 "$facturacion_dir"
    chown 10001:10001 "$signature_dir"
    chmod 0700 "$signature_dir"
    find "$signature_dir" -mindepth 1 -maxdepth 1 -type f -name '*.pem' \
      -exec chown 10001:10001 {} + \
      -exec chmod 0600 {} +
    matched=$((matched + 1))
  done

  printf '[dian-private] Carpetas heredadas preparadas: %s\n' "$matched"
fi

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
