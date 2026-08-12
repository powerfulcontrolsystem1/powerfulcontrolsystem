#!/usr/bin/env bash
# Replica limitada de DIAN desde principal hacia staging para una sola empresa.
# La replica siempre fuerza habilitacion y nunca activa emision local en staging.
set -euo pipefail

EMPRESA_ID="${EMPRESA_ID:-12}"
CONFIRM_STAGE_DIAN="${CONFIRM_STAGE_DIAN:-}"
DRY_RUN="${DRY_RUN:-1}"
PROD_POSTGRES_CONTAINER="${PROD_POSTGRES_CONTAINER:-pcs-postgres}"
PROD_BACKEND_CONTAINER="${PROD_BACKEND_CONTAINER:-pcs-backend}"
STAGING_POSTGRES_CONTAINER="${STAGING_POSTGRES_CONTAINER:-pcs-staging-postgres}"
STAGING_BACKEND_CONTAINER="${STAGING_BACKEND_CONTAINER:-pcs-staging-backend}"
PROD_DB_USER="${PROD_DB_USER:-pcs}"
STAGING_DB_USER="${STAGING_DB_USER:-pcs_staging}"
EMPRESAS_DB="${EMPRESAS_DB:-pcs_empresas}"
STAGING_DIAN_ENDPOINT="${STAGING_DIAN_ENDPOINT:-https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl}"

fail() { echo "[ERROR] $*" >&2; exit 1; }
cleanup() {
  local status=$?
  if [[ "${files_copied:-0}" == "1" && "$status" -ne 0 ]]; then
    docker exec "$STAGING_BACKEND_CONTAINER" sh -c "rm -f '$target_key' '$target_cert'" >/dev/null 2>&1 || true
  fi
  [[ -z "${workdir:-}" ]] || rm -rf -- "$workdir"
  exit "$status"
}
trap cleanup EXIT

[[ "$EMPRESA_ID" =~ ^[1-9][0-9]*$ ]] || fail "EMPRESA_ID debe ser positivo."
if [[ "$DRY_RUN" != "1" && "$CONFIRM_STAGE_DIAN" != "RESTORE_DIAN_EMPRESA_${EMPRESA_ID}" ]]; then
  fail "Defina CONFIRM_STAGE_DIAN=RESTORE_DIAN_EMPRESA_${EMPRESA_ID} para aplicar."
fi
for c in "$PROD_POSTGRES_CONTAINER" "$PROD_BACKEND_CONTAINER" "$STAGING_POSTGRES_CONTAINER" "$STAGING_BACKEND_CONTAINER"; do
  docker inspect "$c" >/dev/null 2>&1 || fail "Falta un contenedor requerido."
done

source_count="$(docker exec "$PROD_POSTGRES_CONTAINER" psql -v ON_ERROR_STOP=1 -U "$PROD_DB_USER" -d "$EMPRESAS_DB" -Atc "SELECT count(*) FROM empresa_dian_configuracion WHERE empresa_id=${EMPRESA_ID}")"
staging_count="$(docker exec "$STAGING_POSTGRES_CONTAINER" psql -v ON_ERROR_STOP=1 -U "$STAGING_DB_USER" -d "$EMPRESAS_DB" -Atc "SELECT count(*) FROM empresa_dian_configuracion WHERE empresa_id=${EMPRESA_ID}")"
[[ "$source_count" == "1" ]] || fail "Principal no tiene una configuracion DIAN unica para la empresa."
[[ "$staging_count" == "0" ]] || fail "Staging ya tiene configuracion DIAN; no se reemplaza automaticamente."

key_ref="$(docker exec "$PROD_POSTGRES_CONTAINER" psql -v ON_ERROR_STOP=1 -U "$PROD_DB_USER" -d "$EMPRESAS_DB" -Atc "SELECT COALESCE(certificado_clave_ref,'') FROM empresa_dian_configuracion WHERE empresa_id=${EMPRESA_ID}")"
cert_ref="$(docker exec "$PROD_POSTGRES_CONTAINER" psql -v ON_ERROR_STOP=1 -U "$PROD_DB_USER" -d "$EMPRESAS_DB" -Atc "SELECT COALESCE(certificado_url,'') FROM empresa_dian_configuracion WHERE empresa_id=${EMPRESA_ID}")"
[[ "$key_ref" == file:* && "$cert_ref" == file:* ]] || fail "Solo se admiten referencias file: legibles para la replica controlada."

source_key="${key_ref#file:}"
source_cert="${cert_ref#file:}"
docker exec -e P110_SOURCE_KEY="$source_key" -e P110_SOURCE_CERT="$source_cert" "$PROD_BACKEND_CONTAINER" sh -ceu '
  [ -r "$P110_SOURCE_KEY" ] && [ -s "$P110_SOURCE_KEY" ]
  [ -r "$P110_SOURCE_CERT" ] && [ -s "$P110_SOURCE_CERT" ]
  openssl pkey -in "$P110_SOURCE_KEY" -noout >/dev/null 2>&1
  openssl x509 -in "$P110_SOURCE_CERT" -noout >/dev/null 2>&1
' || fail "La firma fuente no es legible o valida desde el backend principal."

target_dir="/app/private_storage/dian/empresa_${EMPRESA_ID}"
target_key="${target_dir}/staging_dian_key.pem"
target_cert="${target_dir}/staging_dian_cert.pem"

echo "[INFO] Fuente DIAN y firma principal verificadas para empresa_id=${EMPRESA_ID}."
echo "[INFO] Destino staging quedara en habilitacion con emision local desactivada."
if [[ "$DRY_RUN" == "1" ]]; then
  echo "[OK] DRY_RUN: no se copiaron archivos ni se insertaron configuraciones."
  exit 0
fi

workdir="$(mktemp -d)"
chmod 700 "$workdir"
umask 077
docker exec -e P110_SOURCE_FILE="$source_key" "$PROD_BACKEND_CONTAINER" sh -ceu 'cat "$P110_SOURCE_FILE"' >"$workdir/key.pem"
docker exec -e P110_SOURCE_FILE="$source_cert" "$PROD_BACKEND_CONTAINER" sh -ceu 'cat "$P110_SOURCE_FILE"' >"$workdir/cert.pem"
docker exec "$PROD_POSTGRES_CONTAINER" psql -v ON_ERROR_STOP=1 -U "$PROD_DB_USER" -d "$EMPRESAS_DB" -Atc "SELECT encode(convert_to(row_to_json(d)::text, 'UTF8'), 'base64') FROM empresa_dian_configuracion d WHERE empresa_id=${EMPRESA_ID}" >"$workdir/config.json"

docker exec "$STAGING_BACKEND_CONTAINER" sh -ceu "mkdir -p '$target_dir'; chown pcs:pcs '$target_dir'; chmod 700 '$target_dir'; test ! -e '$target_key'; test ! -e '$target_cert'"
docker exec -i -e P110_TARGET_FILE="$target_key" "$STAGING_BACKEND_CONTAINER" sh -ceu 'umask 077; cat > "$P110_TARGET_FILE"' <"$workdir/key.pem"
docker exec -i -e P110_TARGET_FILE="$target_cert" "$STAGING_BACKEND_CONTAINER" sh -ceu 'umask 077; cat > "$P110_TARGET_FILE"' <"$workdir/cert.pem"
files_copied=1
docker exec "$STAGING_BACKEND_CONTAINER" sh -ceu "chmod 600 '$target_key' '$target_cert'; openssl pkey -in '$target_key' -noout >/dev/null 2>&1; openssl x509 -in '$target_cert' -noout >/dev/null 2>&1"

docker cp "$workdir/config.json" "${STAGING_POSTGRES_CONTAINER}:/tmp/p110_dian_config.json"
docker exec "$STAGING_POSTGRES_CONTAINER" sh -ceu 'chmod 644 /tmp/p110_dian_config.json'
docker exec -i "$STAGING_POSTGRES_CONTAINER" psql -v ON_ERROR_STOP=1 -U "$STAGING_DB_USER" -d "$EMPRESAS_DB" \
  -v p110_endpoint="$STAGING_DIAN_ENDPOINT" -v p110_key="file:${target_key}" -v p110_cert="file:${target_cert}" <<'P110SQL'
BEGIN;
SELECT set_config('p110.endpoint', :'p110_endpoint', true) \g /dev/null
SELECT set_config('p110.key', :'p110_key', true) \g /dev/null
SELECT set_config('p110.cert', :'p110_cert', true) \g /dev/null
DO $$
DECLARE payload jsonb; cols text;
BEGIN
  payload := convert_from(decode(trim(pg_read_file('/tmp/p110_dian_config.json')), 'base64'), 'UTF8')::jsonb;
  payload := jsonb_set(payload, '{certificado_clave_ref}', to_jsonb(current_setting('p110.key', true)));
  payload := jsonb_set(payload, '{certificado_url}', to_jsonb(current_setting('p110.cert', true)));
  payload := jsonb_set(payload, '{tipo_ambiente}', '"habilitacion"'::jsonb);
  payload := jsonb_set(payload, '{produccion_local_activa}', '0'::jsonb);
  payload := jsonb_set(payload, '{url_dian}', to_jsonb(current_setting('p110.endpoint', true)));
  payload := jsonb_set(payload, '{estado_dian}', '"pendiente"'::jsonb);
  payload := jsonb_set(payload, '{ultimo_envio}', '""'::jsonb);
  payload := jsonb_set(payload, '{consecutivo_actual}', '0'::jsonb);
  SELECT string_agg(quote_ident(column_name), ', ' ORDER BY ordinal_position) INTO cols
    FROM information_schema.columns
   WHERE table_schema='public' AND table_name='empresa_dian_configuracion' AND column_name <> 'id';
  EXECUTE format('INSERT INTO empresa_dian_configuracion (%1$s) SELECT %1$s FROM jsonb_populate_record(NULL::empresa_dian_configuracion, $1)', cols)
    USING payload;
END $$;
COMMIT;
P110SQL
docker exec "$STAGING_POSTGRES_CONTAINER" rm -f /tmp/p110_dian_config.json
docker exec -e P110_KEY="$target_key" -e P110_CERT="$target_cert" "$STAGING_BACKEND_CONTAINER" sh -ceu '
  openssl pkey -in "$P110_KEY" -pubout -outform DER >/tmp/p110-key.der 2>/dev/null
  openssl x509 -in "$P110_CERT" -pubkey -noout | openssl pkey -pubin -outform DER >/tmp/p110-cert.der 2>/dev/null
  cmp -s /tmp/p110-key.der /tmp/p110-cert.der
  rm -f /tmp/p110-key.der /tmp/p110-cert.der
'
files_copied=0
echo "[OK] Replica DIAN staging preparada en habilitacion; no se emitio documento fiscal."
