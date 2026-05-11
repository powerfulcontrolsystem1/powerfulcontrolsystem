#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/root/powerfulcontrolsystem}"
STAGING_POSTGRES_CONTAINER="${STAGING_POSTGRES_CONTAINER:-pcs-staging-postgres}"
STAGING_KEEP_EMAIL_REGEX="${STAGING_KEEP_EMAIL_REGEX:-^(powerfulcontrolsystem@gmail\.com)$}"
DATABASES="${STAGING_DATABASES:-pcs_superadministrador pcs_empresas}"

if ! docker ps --format '{{.Names}}' | grep -qx "$STAGING_POSTGRES_CONTAINER"; then
  echo "[ERROR] Contenedor staging PostgreSQL no activo: $STAGING_POSTGRES_CONTAINER" >&2
  exit 1
fi

echo "[INFO] Anonimizando datos sensibles en staging."
echo "[INFO] Emails preservados por regex: $STAGING_KEEP_EMAIL_REGEX"

for db in $DATABASES; do
  echo "[INFO] Anonimizando base staging: $db"
  docker exec -i "$STAGING_POSTGRES_CONTAINER" sh -lc "psql -v ON_ERROR_STOP=1 -U \"\$POSTGRES_USER\" '$db'" <<SQL
DO \$\$
DECLARE
  keep_email_regex text := '${STAGING_KEEP_EMAIL_REGEX}';
  rec record;
  id_expr text;
  set_expr text;
BEGIN
  FOR rec IN
    SELECT table_schema, table_name, column_name
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND lower(column_name) IN (
        'email', 'correo', 'correo_electronico', 'telefono', 'celular', 'whatsapp',
        'documento', 'numero_documento', 'identificacion', 'nit',
        'nombre', 'name', 'cliente_nombre', 'contacto_nombre', 'direccion'
      )
    ORDER BY table_name, column_name
  LOOP
    SELECT CASE
      WHEN EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = rec.table_schema AND table_name = rec.table_name AND column_name = 'id'
      )
      THEN 'COALESCE(id::text, md5(ctid::text))'
      ELSE 'md5(ctid::text)'
    END INTO id_expr;

    IF lower(rec.column_name) IN ('email', 'correo', 'correo_electronico') THEN
      set_expr := format(
        '%I = CASE WHEN COALESCE(%I, '''') ~* %L THEN %I ELSE concat(''anon+'', %s, ''@staging.local'') END',
        rec.column_name, rec.column_name, keep_email_regex, rec.column_name, id_expr
      );
    ELSIF lower(rec.column_name) IN ('telefono', 'celular', 'whatsapp') THEN
      set_expr := format('%I = CASE WHEN COALESCE(%I, '''') = '''' THEN %I ELSE ''3000000000'' END', rec.column_name, rec.column_name, rec.column_name);
    ELSIF lower(rec.column_name) IN ('documento', 'numero_documento', 'identificacion', 'nit') THEN
      set_expr := format('%I = CASE WHEN COALESCE(%I, '''') = '''' THEN %I ELSE concat(''900'', right(md5(%s), 7)) END', rec.column_name, rec.column_name, rec.column_name, id_expr);
    ELSIF lower(rec.column_name) IN ('direccion') THEN
      set_expr := format('%I = CASE WHEN COALESCE(%I, '''') = '''' THEN %I ELSE ''Direccion staging anonima'' END', rec.column_name, rec.column_name, rec.column_name);
    ELSE
      set_expr := format('%I = CASE WHEN COALESCE(%I, '''') = '''' THEN %I ELSE concat(''Dato staging '', left(md5(%s), 8)) END', rec.column_name, rec.column_name, rec.column_name, id_expr);
    END IF;

    EXECUTE format('UPDATE %I.%I SET %s WHERE %I IS NOT NULL', rec.table_schema, rec.table_name, set_expr, rec.column_name);
  END LOOP;
END
\$\$;
SQL
done

echo "[OK] Staging anonimizado."
