#!/usr/bin/env bash
# Auditoria de paridad DIAN entre principal y staging.
# No copia, imprime ni persiste certificados, claves, NIT ni valores DIAN.
set -euo pipefail

EMPRESA_ID="${EMPRESA_ID:-12}"
PROD_POSTGRES_CONTAINER="${PROD_POSTGRES_CONTAINER:-pcs-postgres}"
PROD_BACKEND_CONTAINER="${PROD_BACKEND_CONTAINER:-pcs-backend}"
STAGING_POSTGRES_CONTAINER="${STAGING_POSTGRES_CONTAINER:-pcs-staging-postgres}"
STAGING_BACKEND_CONTAINER="${STAGING_BACKEND_CONTAINER:-pcs-staging-backend}"
PROD_DB_USER="${PROD_DB_USER:-pcs}"
STAGING_DB_USER="${STAGING_DB_USER:-pcs_staging}"
EMPRESAS_DB="${EMPRESAS_DB:-pcs_empresas}"

fail() {
  echo "[ERROR] $*" >&2
  exit 1
}

for required in "$PROD_POSTGRES_CONTAINER" "$PROD_BACKEND_CONTAINER" "$STAGING_POSTGRES_CONTAINER" "$STAGING_BACKEND_CONTAINER"; do
  docker inspect "$required" >/dev/null 2>&1 || fail "No existe el contenedor requerido para la auditoria."
done

[[ "$EMPRESA_ID" =~ ^[1-9][0-9]*$ ]] || fail "EMPRESA_ID debe ser un entero positivo."

config_summary() {
  local postgres_container="$1"
  local db_user="$2"
  docker exec "$postgres_container" psql -v ON_ERROR_STOP=1 -U "$db_user" -d "$EMPRESAS_DB" -Atc "
    SELECT
      'row=' || (count(*) = 1)::int ||
      ',identity=' || count(*) FILTER (WHERE NULLIF(nit, '') IS NOT NULL) ||
      ',ambient=' || count(*) FILTER (WHERE NULLIF(tipo_ambiente, '') IS NOT NULL) ||
      ',range=' || count(*) FILTER (WHERE NULLIF(prefijo, '') IS NOT NULL AND rango_desde > 0 AND rango_hasta >= rango_desde) ||
      ',testset=' || count(*) FILTER (WHERE NULLIF(test_set_id, '') IS NOT NULL) ||
      ',cert_ref=' || count(*) FILTER (WHERE NULLIF(certificado_url, '') IS NOT NULL) ||
      ',key_ref=' || count(*) FILTER (WHERE NULLIF(certificado_clave_ref, '') IS NOT NULL) ||
      ',enabled=' || count(*) FILTER (WHERE produccion_local_activa = 1)
    FROM empresa_dian_configuracion
    WHERE empresa_id = ${EMPRESA_ID};"
}

reference_validation() {
  local postgres_container="$1"
  local backend_container="$2"
  local db_user="$3"
  local key_ref cert_ref
  key_ref="$(docker exec "$postgres_container" psql -v ON_ERROR_STOP=1 -U "$db_user" -d "$EMPRESAS_DB" -Atc "SELECT COALESCE(certificado_clave_ref, '') FROM empresa_dian_configuracion WHERE empresa_id = ${EMPRESA_ID}")"
  cert_ref="$(docker exec "$postgres_container" psql -v ON_ERROR_STOP=1 -U "$db_user" -d "$EMPRESAS_DB" -Atc "SELECT COALESCE(certificado_url, '') FROM empresa_dian_configuracion WHERE empresa_id = ${EMPRESA_ID}")"

  docker exec -e P110_DIAN_KEY_REF="$key_ref" -e P110_DIAN_CERT_REF="$cert_ref" "$backend_container" sh -ceu '
    check_ref() {
      label="$1"
      ref="$2"
      case "$ref" in
        file:*) path=${ref#file:}; [ -r "$path" ] && [ -s "$path" ] && echo "${label}_ref=readable" || echo "${label}_ref=unreadable" ;;
        env:*) name=${ref#env:}; eval "value=\${$name-}"; [ -n "$value" ] && echo "${label}_ref=resolved" || echo "${label}_ref=unresolved" ;;
        base64:*) [ -n "${ref#base64:}" ] && echo "${label}_ref=present" || echo "${label}_ref=empty" ;;
        "") echo "${label}_ref=missing" ;;
        *) echo "${label}_ref=inline" ;;
      esac
    }
    check_ref key "$P110_DIAN_KEY_REF"
    check_ref cert "$P110_DIAN_CERT_REF"
  '
}

echo "[INFO] Auditoria DIAN saneada empresa_id=${EMPRESA_ID}"
echo "[INFO] principal $(config_summary "$PROD_POSTGRES_CONTAINER" "$PROD_DB_USER")"
echo "[INFO] principal $(reference_validation "$PROD_POSTGRES_CONTAINER" "$PROD_BACKEND_CONTAINER" "$PROD_DB_USER" | tr '\n' ',')"
echo "[INFO] staging $(config_summary "$STAGING_POSTGRES_CONTAINER" "$STAGING_DB_USER")"
echo "[INFO] staging $(reference_validation "$STAGING_POSTGRES_CONTAINER" "$STAGING_BACKEND_CONTAINER" "$STAGING_DB_USER" | tr '\n' ',')"
echo "[OK] Auditoria terminada. No se copio ni mostro material fiscal sensible."
