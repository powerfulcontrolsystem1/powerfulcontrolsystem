#!/usr/bin/env bash
# Restaura el ultimo snapshot en PostgreSQL efimero y arranca la API exacta del
# candidato contra esa copia. No conecta ni escribe en staging o produccion.
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-$(cd "$(dirname "$0")/../.." && pwd)}"
SOURCE_ENV="${SOURCE_ENV:-$PROJECT_DIR/deploy/.env.platform}"
BACKUP_DIR="${BACKUP_DIR:-}"
PCS_API_IMAGE_DIGEST="${PCS_API_IMAGE_DIGEST:-}"
PCS_MIGRATE_IMAGE_DIGEST="${PCS_MIGRATE_IMAGE_DIGEST:-}"
P109_DRILL_ID="${P109_DRILL_ID:-p109-restore-app-$(date +%s)}"
P109_HOST_PORT="${P109_HOST_PORT:-18083}"
RESTORE_IMAGE="${RESTORE_IMAGE:-postgres:16.14-alpine}"
P109_QA_EMAIL="${P109_QA_EMAIL:-}"
P109_QA_PASSWORD="${P109_QA_PASSWORD:-}"
P109_HOLD_SECONDS="${P109_HOLD_SECONDS:-0}"
P109_VERIFY_REPLICA="${P109_VERIFY_REPLICA:-0}"
P109_REPLICA_HOST_PORT="${P109_REPLICA_HOST_PORT:-$((P109_HOST_PORT + 1))}"
started_epoch="$(date +%s)"

fail() {
  echo "[ERROR] $*" >&2
  exit 1
}

json_escape() {
  local value="$1"
  value="${value//\\/\\\\}"
  value="${value//\"/\\\"}"
  value="${value//$'\n'/\\n}"
  value="${value//$'\r'/\\r}"
  printf '%s' "$value"
}

[[ "$P109_DRILL_ID" =~ ^p109-restore-app-[a-z0-9-]+$ ]] ||
  fail "P109_DRILL_ID debe iniciar con p109-restore-app-."
[[ "$PCS_API_IMAGE_DIGEST" =~ ^[^[:space:]@]+@sha256:[a-f0-9]{64}$ ]] ||
  fail "PCS_API_IMAGE_DIGEST debe usar repositorio@sha256:<64 hex>."
[[ "$PCS_MIGRATE_IMAGE_DIGEST" =~ ^[^[:space:]@]+@sha256:[a-f0-9]{64}$ ]] ||
  fail "PCS_MIGRATE_IMAGE_DIGEST debe usar repositorio@sha256:<64 hex>."
[[ "$P109_HOST_PORT" =~ ^[0-9]+$ ]] || fail "P109_HOST_PORT debe ser numerico."
((P109_HOST_PORT >= 1024 && P109_HOST_PORT <= 65535)) ||
  fail "P109_HOST_PORT debe estar entre 1024 y 65535."
[[ "$P109_HOLD_SECONDS" =~ ^[0-9]+$ ]] || fail "P109_HOLD_SECONDS debe ser numerico."
((P109_HOLD_SECONDS <= 900)) || fail "P109_HOLD_SECONDS no puede superar 900."
[[ "$P109_VERIFY_REPLICA" =~ ^[01]$ ]] || fail "P109_VERIFY_REPLICA debe ser 0 o 1."
[[ "$P109_REPLICA_HOST_PORT" =~ ^[0-9]+$ ]] || fail "P109_REPLICA_HOST_PORT debe ser numerico."
((P109_REPLICA_HOST_PORT >= 1024 && P109_REPLICA_HOST_PORT <= 65535)) ||
  fail "P109_REPLICA_HOST_PORT debe estar entre 1024 y 65535."
if [ "$P109_VERIFY_REPLICA" = "1" ]; then
  [ -n "$P109_QA_EMAIL" ] && [ -n "$P109_QA_PASSWORD" ] ||
    fail "La prueba de replicas requiere las credenciales QA por variables de sesion."
  [ "$P109_REPLICA_HOST_PORT" != "$P109_HOST_PORT" ] ||
    fail "Los puertos de las replicas deben ser distintos."
fi
[ -f "$SOURCE_ENV" ] || fail "No existe SOURCE_ENV."
docker image inspect "$PCS_API_IMAGE_DIGEST" >/dev/null 2>&1 ||
  fail "La imagen API exacta no esta disponible localmente."
docker image inspect "$PCS_MIGRATE_IMAGE_DIGEST" >/dev/null 2>&1 ||
  fail "La imagen de migracion exacta no esta disponible localmente."

backup_root="$PROJECT_DIR/backups/vps-snapshots"
if [ -z "$BACKUP_DIR" ]; then
  BACKUP_DIR="$(find "$backup_root" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort | tail -n 1)"
fi
[ -d "$BACKUP_DIR" ] || fail "No existe el snapshot seleccionado."
[ -s "$BACKUP_DIR/postgres_all.sql.gz" ] || fail "Falta postgres_all.sql.gz."
gzip -t "$BACKUP_DIR/postgres_all.sql.gz"

private_tar="$BACKUP_DIR/powerful-control-system_pcs_private_storage.tar.gz"
[ -s "$private_tar" ] || fail "Falta el volumen privado del snapshot."
tar -tzf "$private_tar" >/dev/null

network="${P109_DRILL_ID}-net"
postgres="${P109_DRILL_ID}-postgres"
api="${P109_DRILL_ID}-api"
api_replica="${P109_DRILL_ID}-api-replica"
migrate="${P109_DRILL_ID}-migrate"
workdir="/tmp/${P109_DRILL_ID}"
runtime_user="pcs_restore_runtime"
restore_password="$(openssl rand -hex 24)"
runtime_password="$(openssl rand -hex 24)"

for container in "$postgres" "$migrate" "$api" "$api_replica"; do
  docker container inspect "$container" >/dev/null 2>&1 &&
    fail "El contenedor aislado ya existe y no sera sobrescrito: $container"
done
docker network inspect "$network" >/dev/null 2>&1 &&
  fail "La red aislada ya existe y no sera sobrescrita: $network"
if ss -H -ltn "sport = :$P109_HOST_PORT" 2>/dev/null | grep -q .; then
  fail "El puerto local $P109_HOST_PORT ya esta ocupado."
fi
if [ "$P109_VERIFY_REPLICA" = "1" ] && ss -H -ltn "sport = :$P109_REPLICA_HOST_PORT" 2>/dev/null | grep -q .; then
  fail "El puerto local $P109_REPLICA_HOST_PORT ya esta ocupado."
fi
[ ! -e "$workdir" ] || fail "El directorio temporal ya existe: $workdir"

cleanup() {
  docker rm -f "$api" "$api_replica" "$migrate" "$postgres" >/dev/null 2>&1 || true
  docker network rm "$network" >/dev/null 2>&1 || true
  if [[ "$workdir" == /tmp/p109-restore-app-* ]]; then
    rm -rf -- "$workdir"
  fi
}
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

mkdir -p "$workdir/private_storage"
tar -xzf "$private_tar" -C "$workdir/private_storage"
private_root="$workdir/private_storage"
if [ -d "$private_root/_data" ]; then
  private_root="$private_root/_data"
elif [ -d "$private_root/data" ]; then
  private_root="$private_root/data"
fi
chown -R 10001:10001 "$private_root"

docker network create "$network" >/dev/null
docker run -d \
  --name "$postgres" \
  --network "$network" \
  --network-alias postgres \
  -e POSTGRES_PASSWORD="$restore_password" \
  "$RESTORE_IMAGE" >/dev/null

for attempt in $(seq 1 45); do
  if docker exec "$postgres" pg_isready -U postgres >/dev/null 2>&1; then
    break
  fi
  [ "$attempt" -lt 45 ] || fail "PostgreSQL aislado no quedo listo."
  sleep 2
done

echo "[RUN] Restaurando el snapshot en PostgreSQL efimero."
gunzip -c "$BACKUP_DIR/postgres_all.sql.gz" |
  docker exec -i "$postgres" psql -v ON_ERROR_STOP=1 -U postgres \
    >/tmp/${P109_DRILL_ID}-restore.log

database_count="$(docker exec "$postgres" psql -U postgres -Atc \
  "SELECT count(*) FROM pg_database WHERE datname IN ('pcs_empresas','pcs_superadministrador')")"
[ "$database_count" = "2" ] || fail "La restauracion no contiene las dos bases PCS."

escaped_runtime_password="${runtime_password//\'/\'\'}"
docker exec -i "$postgres" psql -v ON_ERROR_STOP=1 -U postgres -d postgres <<SQL >/dev/null
DROP ROLE IF EXISTS ${runtime_user};
CREATE ROLE ${runtime_user} LOGIN PASSWORD '${escaped_runtime_password}'
  NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOBYPASSRLS;
GRANT CONNECT ON DATABASE pcs_empresas TO ${runtime_user};
GRANT CONNECT ON DATABASE pcs_superadministrador TO ${runtime_user};
SQL

for database in pcs_empresas pcs_superadministrador; do
  docker exec -i "$postgres" psql -v ON_ERROR_STOP=1 -U postgres -d "$database" <<SQL >/dev/null
GRANT USAGE ON SCHEMA public TO ${runtime_user};
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ${runtime_user};
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO ${runtime_user};
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO ${runtime_user};
SQL
done

# El dump puede contener el hash historico del propietario. Solo en esta copia
# efimera se restablece una clave aleatoria para que el migrador use TCP.
escaped_restore_password="${restore_password//\'/\'\'}"
docker exec "$postgres" psql -v ON_ERROR_STOP=1 -U postgres -d postgres -c \
  "ALTER ROLE postgres PASSWORD '${escaped_restore_password}'" >/dev/null

echo "[RUN] Aplicando el migrador exacto sobre la copia restaurada."
docker run --rm \
  --name "$migrate" \
  --network "$network" \
  --env-file "$SOURCE_ENV" \
  --read-only \
  --cap-drop ALL \
  --security-opt no-new-privileges:true \
  --pids-limit 128 \
  --tmpfs /tmp:rw,noexec,nosuid,size=32m \
  -v "$private_root:/app/private_storage:rw" \
  -e PCS_ENV=staging \
  -e PCS_RUNTIME_ROLE=migrate \
  -e PCS_RUNTIME_SCHEMA_BOOTSTRAP=0 \
  -e PCS_RUNTIME_DB_USER="$runtime_user" \
  -e PCS_RUNTIME_DB_PASSWORD="$runtime_password" \
  -e DB_DIALECT=postgres \
  -e "DB_EMPRESAS_DSN=postgres://postgres:${restore_password}@postgres:5432/pcs_empresas?sslmode=disable" \
  -e "DB_SUPERADMIN_DSN=postgres://postgres:${restore_password}@postgres:5432/pcs_superadministrador?sslmode=disable" \
  -e PCS_SKIP_CORPORATE_EMAIL_STARTUP_SYNC=1 \
  -e PCS_MIGRATION_SKIP_EXTERNAL_SYNC=1 \
  "$PCS_MIGRATE_IMAGE_DIGEST" >/dev/null

critical_tables="empresa_cuentas_por_pagar empresa_asientos_contables empresa_ai_memoria empresa_dian_configuracion empresa_documentos_gestion"
checked_tables=0
tenant_rows=0
for table in $critical_tables; do
  exists="$(docker exec "$postgres" psql -U postgres -d pcs_empresas -Atc \
    "SELECT to_regclass('public.$table') IS NOT NULL")"
  [ "$exists" = "t" ] || fail "Tabla critica ausente: $table"
  rows="$(docker exec "$postgres" psql -U postgres -d pcs_empresas -Atc \
    "SELECT count(*) FROM $table WHERE empresa_id=12")"
  tenant_rows=$((tenant_rows + rows))
  checked_tables=$((checked_tables + 1))
done

start_api() {
  local container_name="$1"
  local host_port="$2"
  docker run -d \
    --name "$container_name" \
    --network "$network" \
    --env-file "$SOURCE_ENV" \
    --read-only \
    --cap-drop ALL \
    --security-opt no-new-privileges:true \
    --pids-limit 256 \
    --tmpfs /tmp:rw,noexec,nosuid,size=64m \
    --tmpfs /app/backend/logs:rw,noexec,nosuid,size=32m,uid=10001,gid=10001 \
    --tmpfs /app/web/uploads:rw,noexec,nosuid,size=32m,uid=10001,gid=10001 \
    --tmpfs /app/backup:rw,noexec,nosuid,size=16m,uid=10001,gid=10001 \
    --tmpfs /app/descargas:rw,noexec,nosuid,size=16m,uid=10001,gid=10001 \
    -v "$private_root:/app/private_storage:rw" \
    -p "127.0.0.1:${host_port}:8080" \
    -e PCS_ENV=staging \
    -e PORT=8080 \
    -e PCS_RUNTIME_ROLE=api \
    -e PCS_RUNTIME_SCHEMA_BOOTSTRAP=0 \
    -e PCS_RUNTIME_DB_USER="$runtime_user" \
    -e PCS_RUNTIME_DB_PASSWORD="$runtime_password" \
    -e DB_DIALECT=postgres \
    -e "DB_EMPRESAS_DSN=postgres://${runtime_user}:${runtime_password}@postgres:5432/pcs_empresas?sslmode=disable" \
    -e "DB_SUPERADMIN_DSN=postgres://${runtime_user}:${runtime_password}@postgres:5432/pcs_superadministrador?sslmode=disable" \
    -e PCS_SKIP_CORPORATE_EMAIL_STARTUP_SYNC=1 \
    -e PCS_SKIP_SERVER_STARTUP_EVENT=1 \
    -e NEXTCLOUD_ENABLED=0 \
    -e EMAIL_CORPORATIVO_ENABLED=0 \
    -e OPENAI_API_KEY= \
    -e GEMINI_API_KEY= \
    "$PCS_API_IMAGE_DIGEST" >/dev/null
}

wait_api() {
  local container_name="$1"
  local host_port="$2"
  local ready=0
  for attempt in $(seq 1 60); do
    if curl -fsS "http://127.0.0.1:${host_port}/ready" >/dev/null 2>&1; then
      ready=1
      break
    fi
    if [ "$(docker inspect -f '{{.State.Running}}' "$container_name" 2>/dev/null || true)" != "true" ]; then
      docker logs --tail 80 "$container_name" 2>&1 | sed -E \
        's#(postgres(ql)?://)[^/@[:space:]]+@#\1[REDACTED]@#g' >&2
      fail "La API $container_name termino antes de quedar lista."
    fi
    sleep 2
  done
  if [ "$ready" != "1" ]; then
    docker logs --tail 80 "$container_name" 2>&1 | sed -E \
      's#(postgres(ql)?://)[^/@[:space:]]+@#\1[REDACTED]@#g' >&2
    fail "La API $container_name no alcanzo readiness en 120 segundos."
  fi
}

echo "[RUN] Arrancando la API exacta contra la restauracion."
start_api "$api" "$P109_HOST_PORT"
wait_api "$api" "$P109_HOST_PORT"
if [ "$P109_VERIFY_REPLICA" = "1" ]; then
  echo "[RUN] Arrancando una segunda replica exacta sobre el mismo volumen restaurado."
  start_api "$api_replica" "$P109_REPLICA_HOST_PORT"
  wait_api "$api_replica" "$P109_REPLICA_HOST_PORT"
fi

health_status="$(curl -sS -o /dev/null -w '%{http_code}' \
  "http://127.0.0.1:${P109_HOST_PORT}/health")"
ready_status="$(curl -sS -o /dev/null -w '%{http_code}' \
  "http://127.0.0.1:${P109_HOST_PORT}/ready")"
[ "$health_status" = "200" ] || fail "/health devolvio HTTP $health_status."
[ "$ready_status" = "200" ] || fail "/ready devolvio HTTP $ready_status."

protected_paths=(
  "cxp|/api/empresa/finanzas/cuentas_pagar?empresa_id=12"
  "ia_compras|/api/empresa/soportes_compras_ia?empresa_id=12"
  "dian|/api/empresa/facturacion_electronica?empresa_id=12"
  "documentos|/api/empresa/documentos/gestion?empresa_id=12"
)
protected_checks=0
for check in "${protected_paths[@]}"; do
  name="${check%%|*}"
  path="${check#*|}"
  status="$(curl -sS -o /dev/null -w '%{http_code}' \
    "http://127.0.0.1:${P109_HOST_PORT}${path}")"
  case "$status" in
    401|403) ;;
    *) fail "El endpoint $name acepto acceso anonimo (HTTP $status)." ;;
  esac
  protected_checks=$((protected_checks + 1))
done

authenticated_checks=0
replica_checks=0
hostile_file_checks=0
if [ -n "$P109_QA_EMAIL" ] || [ -n "$P109_QA_PASSWORD" ]; then
  [ -n "$P109_QA_EMAIL" ] && [ -n "$P109_QA_PASSWORD" ] ||
    fail "La prueba autenticada requiere correo y clave juntos."
  cookie_jar="$workdir/session.cookies"
  login_response="$workdir/login.response"
  login_payload="{\"email\":\"$(json_escape "$P109_QA_EMAIL")\",\"password\":\"$(json_escape "$P109_QA_PASSWORD")\",\"otp_code\":\"\",\"recaptcha_token\":\"\"}"
  login_status="$(curl -sS -o "$login_response" -w '%{http_code}' \
    -c "$cookie_jar" \
    -H 'Content-Type: application/json' \
    --data-binary "$login_payload" \
    "http://127.0.0.1:${P109_HOST_PORT}/super/api/administradores/login")"
  [ "$login_status" = "200" ] || fail "El login oficial devolvio HTTP $login_status."
  grep -Eq '"ok"[[:space:]]*:[[:space:]]*true' "$login_response" ||
    fail "El login oficial no creo una sesion valida."

  authenticated_paths=(
    "cxp|/api/empresa/finanzas/cuentas_pagar?empresa_id=12"
    "contabilidad|/api/empresa/contabilidad_colombia_avanzada?empresa_id=12&action=dashboard"
    "ia_compras|/api/empresa/soportes_compras_ia?empresa_id=12"
    "dian|/api/empresa/facturacion_electronica/dian?action=diagnostico_oficial&empresa_id=12"
    "documentos|/api/empresa/documentos/gestion?empresa_id=12"
  )
  for check in "${authenticated_paths[@]}"; do
    name="${check%%|*}"
    path="${check#*|}"
    status="$(curl -sS -o /dev/null -w '%{http_code}' \
      -b "$cookie_jar" \
      "http://127.0.0.1:${P109_HOST_PORT}${path}")"
    [ "$status" = "200" ] ||
      fail "El dominio $name no respondio por el flujo autenticado (HTTP $status)."
    authenticated_checks=$((authenticated_checks + 1))
  done

  if [ "$P109_VERIFY_REPLICA" = "1" ]; then
    cookie_replica="$workdir/session-replica.cookies"
    login_replica_response="$workdir/login-replica.response"
    login_replica_status="$(curl -sS -o "$login_replica_response" -w '%{http_code}' \
      -c "$cookie_replica" \
      -H 'Content-Type: application/json' \
      --data-binary "$login_payload" \
      "http://127.0.0.1:${P109_REPLICA_HOST_PORT}/super/api/administradores/login")"
    [ "$login_replica_status" = "200" ] ||
      fail "El login oficial de la segunda replica devolvio HTTP $login_replica_status."
    grep -Eq '"ok"[[:space:]]*:[[:space:]]*true' "$login_replica_response" ||
      fail "La segunda replica no creo una sesion valida."

    sample_file="$PROJECT_DIR/web/img/logo.png"
    [ -s "$sample_file" ] || fail "No existe el archivo controlado para la prueba de replicas."
    source_hash="$(sha256sum "$sample_file" | awk '{print $1}')"
    csrf_token="$(awk '$6 == "pcs_csrf" {print $7}' "$cookie_jar" | tail -n 1)"
    [ -n "$csrf_token" ] || fail "El login oficial no emitio token CSRF."
    upload_response="$workdir/replica-upload.response"
    document_number="P109-REPLICA-${P109_DRILL_ID##*-}-$(date +%s)"
    upload_status="$(curl -sS -o "$upload_response" -w '%{http_code}' \
      -b "$cookie_jar" \
      -H "X-CSRF-Token: ${csrf_token}" \
      -H "Origin: http://127.0.0.1:${P109_HOST_PORT}" \
      -F "archivo=@${sample_file};type=image/png" \
      -F 'tipo_soporte=gasto' \
      -F 'documento_tipo=otro' \
      -F "documento_numero=${document_number}" \
      -F 'proveedor_nombre=P109 replica aislada' \
      -F 'subtotal=1' \
      -F 'total=1' \
      "http://127.0.0.1:${P109_HOST_PORT}/api/empresa/soportes_compras_ia?empresa_id=12&action=radicar")"
    [ "$upload_status" = "200" ] || fail "La carga por replica A devolvio HTTP $upload_status."
    grep -Eq '"ok"[[:space:]]*:[[:space:]]*true' "$upload_response" ||
      fail "La carga por replica A no fue confirmada."
    soporte_id="$(sed -n 's/.*"id"[[:space:]]*:[[:space:]]*\([0-9][0-9]*\).*/\1/p' "$upload_response" | head -n 1)"
    [ -n "$soporte_id" ] || fail "No se pudo resolver el soporte creado en la replica A."
    uploaded_name="$(sed -n 's/.*"archivo_nombre"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$upload_response" | head -n 1)"
    [[ "$uploaded_name" =~ ^[a-f0-9]{64}\.png$ ]] || fail "El nombre privado no cumple el contrato aleatorio."

    wrong_tenant_status="$(curl -sS -o /dev/null -w '%{http_code}' \
      -b "$cookie_replica" \
      "http://127.0.0.1:${P109_REPLICA_HOST_PORT}/api/empresa/soportes_compras_ia?empresa_id=7&action=descargar&soporte_id=${soporte_id}")"
    case "$wrong_tenant_status" in
      403|404) hostile_file_checks=$((hostile_file_checks + 1)) ;;
      *) fail "La descarga cruzada de empresa devolvio HTTP $wrong_tenant_status." ;;
    esac

    rows_before_invalid="$(docker exec "$postgres" psql -U postgres -d pcs_empresas -Atc \
      "SELECT count(*) FROM empresa_soportes_compras_ia WHERE empresa_id=12")"
    hostile_html="$workdir/p109-hostile.html"
    printf '%s' '<script>alert(1)</script>' >"$hostile_html"
    hostile_html_status="$(curl -sS -o /dev/null -w '%{http_code}' \
      -b "$cookie_jar" \
      -H "X-CSRF-Token: ${csrf_token}" \
      -H "Origin: http://127.0.0.1:${P109_HOST_PORT}" \
      -F "archivo=@${hostile_html};type=text/html" \
      -F 'tipo_soporte=gasto' \
      "http://127.0.0.1:${P109_HOST_PORT}/api/empresa/soportes_compras_ia?empresa_id=12&action=radicar")"
    [ "$hostile_html_status" = "400" ] || fail "El contenido activo devolvio HTTP $hostile_html_status."
    hostile_file_checks=$((hostile_file_checks + 1))

    oversized_file="$workdir/p109-oversized.png"
    truncate -s 16M "$oversized_file"
    oversized_status="$(curl -sS -o /dev/null -w '%{http_code}' \
      -b "$cookie_jar" \
      -H "X-CSRF-Token: ${csrf_token}" \
      -H "Origin: http://127.0.0.1:${P109_HOST_PORT}" \
      -F "archivo=@${oversized_file};type=image/png" \
      -F 'tipo_soporte=gasto' \
      "http://127.0.0.1:${P109_HOST_PORT}/api/empresa/soportes_compras_ia?empresa_id=12&action=radicar")"
    case "$oversized_status" in
      400|413) hostile_file_checks=$((hostile_file_checks + 1)) ;;
      *) fail "El archivo sobredimensionado devolvio HTTP $oversized_status." ;;
    esac
    rows_after_invalid="$(docker exec "$postgres" psql -U postgres -d pcs_empresas -Atc \
      "SELECT count(*) FROM empresa_soportes_compras_ia WHERE empresa_id=12")"
    [ "$rows_after_invalid" = "$rows_before_invalid" ] ||
      fail "Una carga hostil creo una fila empresarial."
    hostile_file_checks=$((hostile_file_checks + 1))

    replica_download="$workdir/replica-download.png"
    download_status="$(curl -sS -o "$replica_download" -w '%{http_code}' \
      -b "$cookie_replica" \
      "http://127.0.0.1:${P109_REPLICA_HOST_PORT}/api/empresa/soportes_compras_ia?empresa_id=12&action=descargar&soporte_id=${soporte_id}")"
    [ "$download_status" = "200" ] || fail "La descarga por replica B devolvio HTTP $download_status."
    [ "$(sha256sum "$replica_download" | awk '{print $1}')" = "$source_hash" ] ||
      fail "La replica B devolvio contenido con checksum distinto."
    replica_checks=$((replica_checks + 1))

    docker rm -f "$api" >/dev/null
    curl -fsS "http://127.0.0.1:${P109_REPLICA_HOST_PORT}/ready" >/dev/null ||
      fail "La replica B perdio readiness al retirar la replica A."
    download_after_loss="$workdir/replica-after-loss.png"
    after_loss_status="$(curl -sS -o "$download_after_loss" -w '%{http_code}' \
      -b "$cookie_replica" \
      "http://127.0.0.1:${P109_REPLICA_HOST_PORT}/api/empresa/soportes_compras_ia?empresa_id=12&action=descargar&soporte_id=${soporte_id}")"
    [ "$after_loss_status" = "200" ] || fail "La descarga tras retirar A devolvio HTTP $after_loss_status."
    [ "$(sha256sum "$download_after_loss" | awk '{print $1}')" = "$source_hash" ] ||
      fail "La descarga tras retirar A no conservo el checksum."
    replica_checks=$((replica_checks + 1))

    uploaded_path="$private_root/soportes_compras_ia/empresa_12/$uploaded_name"
    resolved_upload="$(realpath "$uploaded_path")"
    [[ "$resolved_upload" == "$private_root/soportes_compras_ia/empresa_12/"* ]] ||
      fail "El archivo controlado quedo fuera del directorio empresarial."
    mv -- "$uploaded_path" "$workdir/original-upload.png"
    ln -s "$sample_file" "$uploaded_path"
    symlink_status="$(curl -sS -o /dev/null -w '%{http_code}' \
      -b "$cookie_replica" \
      "http://127.0.0.1:${P109_REPLICA_HOST_PORT}/api/empresa/soportes_compras_ia?empresa_id=12&action=descargar&soporte_id=${soporte_id}")"
    [ "$symlink_status" = "404" ] || fail "El escape por symlink devolvio HTTP $symlink_status."
    hostile_file_checks=$((hostile_file_checks + 1))
  fi
  unset login_payload
  rm -f -- "$cookie_jar" "$login_response" "${cookie_replica:-}" "${login_replica_response:-}"
fi

runtime_flags="$(docker exec "$postgres" psql -U postgres -d postgres -Atc \
  "SELECT rolsuper::int||','||rolcreatedb::int||','||rolcreaterole::int||','||rolbypassrls::int FROM pg_roles WHERE rolname='${runtime_user}'")"
[ "$runtime_flags" = "0,0,0,0" ] || fail "El rol efimero obtuvo privilegios de plataforma."

snapshot_epoch="$(stat -c %Y "$BACKUP_DIR/postgres_all.sql.gz")"
finished_epoch="$(date +%s)"
echo "[OK] app_restore_smoke health=200 ready=200 bases=2 tablas=${checked_tables} filas_empresa_12=${tenant_rows} endpoints_protegidos=${protected_checks} dominios_autenticados=${authenticated_checks} replica_checks=${replica_checks} archivos_hostiles=${hostile_file_checks} runtime_privilegios=0 RTO=$((finished_epoch-started_epoch))s RPO=$((finished_epoch-snapshot_epoch))s"
if ((P109_HOLD_SECONDS > 0)); then
  echo "[INFO] Ventana visual local activa durante ${P109_HOLD_SECONDS}s."
  sleep "$P109_HOLD_SECONDS"
fi
echo "[OK] Los contenedores, la red y los archivos efimeros se eliminaran automaticamente."
