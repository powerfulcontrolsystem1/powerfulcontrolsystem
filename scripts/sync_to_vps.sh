#!/usr/bin/env bash
# sync_to_vps.sh
# Sincroniza una carpeta local hacia un VPS Linux usando rsync sobre SSH
# y reinicia el backend remoto mediante systemd para dejarlo persistente.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd -P)"
DEFAULT_LOCAL_DIR="$(cd "$SCRIPT_DIR/.." && pwd -P)"

LOCAL_DIR="${LOCAL_DIR:-$DEFAULT_LOCAL_DIR}"
REMOTE_USER="${REMOTE_USER:-root}"
REMOTE_HOST="${REMOTE_HOST:-2.24.197.58}"
REMOTE_PORT="${REMOTE_PORT:-22}"
REMOTE_DIR="${REMOTE_DIR:-/root/powerfulcontrolsystem}"
SSH_KEY="${SSH_KEY:-$HOME/.ssh/id_rsa}"
SSH_STRICT_HOSTKEY="${SSH_STRICT_HOSTKEY:-accept-new}"
SSH_CLIENT="${SSH_CLIENT:-ssh}"
PLINK_EXE="${PLINK_EXE:-}"
PLINK_KEY_WIN="${PLINK_KEY_WIN:-}"
RESTART_SERVER="${RESTART_SERVER:-1}"
SERVER_PORT="${SERVER_PORT:-8080}"
REMOTE_BINARY="${REMOTE_BINARY:-backend/bin/server_linux_amd64}"
REMOTE_STDOUT_LOG="${REMOTE_STDOUT_LOG:-backend/server.log}"
REMOTE_STDERR_LOG="${REMOTE_STDERR_LOG:-backend/server.err}"
HEALTH_TIMEOUT="${HEALTH_TIMEOUT:-45}"
BOOTSTRAP_SERVER="${BOOTSTRAP_SERVER:-1}"
GOOGLE_CLIENT_ID="${GOOGLE_CLIENT_ID:-}"
GOOGLE_CLIENT_SECRET="${GOOGLE_CLIENT_SECRET:-}"
GOOGLE_REDIRECT_URL="${GOOGLE_REDIRECT_URL:-}"
DB_DIALECT="${DB_DIALECT:-}"
DB_EMPRESAS_DSN="${DB_EMPRESAS_DSN:-}"
DB_SUPERADMIN_DSN="${DB_SUPERADMIN_DSN:-}"

DRY_RUN=0
EXCLUDE_FILE=""

EXCLUDES=(
  ".git"
  ".git/*"
  ".gitignore"
  ".codex"
  ".codex/*"
  ".codex-gocache"
  ".codex-gocache/*"
  ".codex-tmp-go"
  ".codex-tmp-go/*"
  ".agents"
  ".agents/*"
  ".cache"
  ".cache/*"
  ".gocache"
  ".gocache/*"
  ".gotmp"
  ".gotmp/*"
  "*/.codex-gocache"
  "*/.codex-gocache/*"
  "*/.codex-tmp-go"
  "*/.codex-tmp-go/*"
  "*/.gocache"
  "*/.gocache/*"
  "*/.gotmp"
  "*/.gotmp/*"
  ".cursor"
  ".cursor/*"
  ".github"
  ".github/*"
  ".vscode"
  ".vscode/*"
  "backup"
  "backup/*"
  "descargas"
  "descargas/*"
  "node_modules"
  "*/node_modules"
  "*/*/node_modules"
  "logs"
  "logs/*"
  "scripts/logs"
  "scripts/logs/*"
  "tmp"
  "tmp/*"
  "test_runs"
  "test_runs/*"
  "documentos/evidencias_qa"
  "documentos/evidencias_qa/*"
  "coverage"
  "coverage/*"
  "dist"
  "dist/*"
  "build"
  "build/*"
  ".pytest_cache"
  ".pytest_cache/*"
  "__pycache__"
  "*/__pycache__"
  "*/*/__pycache__"
  "*.db"
  "*.sqlite"
  "*.sqlite3"
  "*.log"
  "*.exe"
  "*.vsix"
  "*.tmp"
  "*.bak"
  "backend/.env.local"
  "backend/.env"
  "backend/server_linux_amd64"
  "backend/tools"
  "backend/tools/*"
  "backend/tmp"
  "backend/tmp/*"
  "backend/.codex-gocache"
  "backend/.codex-gocache/*"
  "backend/.codex-tmp-go"
  "backend/.codex-tmp-go/*"
  "backend/server.log"
  "backend/server.err"
  "herramientas"
  "herramientas/*"
  "*.ppk"
  "*.pem"
  "*.key"
)

CURRENT_STEP="inicio"

info() {
  echo "[INFO] $*"
}

ok() {
  echo "[OK] $*"
}

warn() {
  echo "[WARN] $*" >&2
}

error() {
  echo "[ERROR] $*" >&2
}

set_step() {
  CURRENT_STEP="$1"
  info "$1"
}

handle_err() {
  local exit_code=$?
  trap - ERR
  if (( exit_code == 0 )); then
    return
  fi

  error "Fallo el paso: $CURRENT_STEP (codigo $exit_code)."
  case "$CURRENT_STEP" in
    "verificando acceso SSH")
      error "Revise conectividad, puerto SSH, firewall y la clave privada configurada."
      ;;
    "sincronizando archivos con rsync")
      error "Revise permisos de escritura en REMOTE_DIR, exclusiones y espacio libre en el VPS."
      ;;
    "bootstrap remoto del VPS")
      error "Revise las lineas BOOTSTRAP_ERROR y BOOTSTRAP_HINT impresas arriba."
      ;;
    "reiniciando backend remoto con systemd")
      error "Revise las lineas DEPLOY_ERROR y DEPLOY_HINT, junto con systemctl/journalctl/logs impresos arriba."
      ;;
  esac
  exit "$exit_code"
}

trap 'handle_err' ERR

usage() {
  cat <<EOF
Uso: $(basename "$0") [opciones]

Opciones:
  --dry-run, -n            Simular sincronización sin aplicar cambios
  --local PATH             Ruta local a sincronizar (default: raíz del repositorio)
  --host HOST              Host remoto (default: ${REMOTE_HOST})
  --user USER              Usuario remoto (default: ${REMOTE_USER})
  --remote PATH            Directorio remoto destino (default: ${REMOTE_DIR})
  --port PORT              Puerto SSH (default: ${REMOTE_PORT})
  --identity FILE          Clave privada SSH (default: ${SSH_KEY})
  --exclude-file FILE      Archivo adicional de exclusiones (uno por línea)
  --restart-server         Reinicia backend remoto al finalizar sync (default)
  --no-restart-server      No reinicia backend remoto tras sync
  --server-port PORT       Puerto HTTP para healthcheck remoto (default: ${SERVER_PORT})
  --remote-binary PATH     Binario relativo dentro de REMOTE_DIR (default: ${REMOTE_BINARY})
  --stdout-log PATH        Log stdout relativo dentro de REMOTE_DIR (default: ${REMOTE_STDOUT_LOG})
  --stderr-log PATH        Log stderr relativo dentro de REMOTE_DIR (default: ${REMOTE_STDERR_LOG})
  --health-timeout SEG     Timeout healthcheck en segundos (default: ${HEALTH_TIMEOUT})
  -h, --help               Mostrar ayuda

Variables opcionales:
  SSH_STRICT_HOSTKEY       accept-new | yes | no (default: accept-new)
  SSH_CLIENT               ssh | plink (default: ssh)
  PLINK_EXE                Ruta a plink.exe (requerido si SSH_CLIENT=plink)
  PLINK_KEY_WIN            Ruta Windows a .ppk (requerido si SSH_CLIENT=plink)
  BOOTSTRAP_SERVER         1 ejecuta bootstrap remoto de .env.local (default: 1)
  DB_DIALECT               postgres (opcional; si no se envía, conserva el remoto)
  DB_EMPRESAS_DSN          DSN PostgreSQL para base pcs_empresas
  DB_SUPERADMIN_DSN        DSN PostgreSQL para base pcs_superadministrador
  GOOGLE_CLIENT_ID         Valor opcional para escribir en backend/.env.local
  GOOGLE_CLIENT_SECRET     Valor opcional para escribir en backend/.env.local
  GOOGLE_REDIRECT_URL      Valor opcional para escribir en backend/.env.local

Ejemplos:
  $(basename "$0") --dry-run
  $(basename "$0") --host 2.24.197.58 --user root --remote /root/powerfulcontrolsystem
EOF
}

require_arg() {
  if [[ $# -lt 2 || -z "${2:-}" ]]; then
    echo "Error: falta valor para $1" >&2
    exit 1
  fi
}

escape_sq() {
  printf "%s" "$1" | sed "s/'/'\\\\''/g"
}

is_uint() {
  [[ "$1" =~ ^[0-9]+$ ]]
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run|-n)
      DRY_RUN=1
      shift
      ;;
    --local)
      require_arg "$@"
      LOCAL_DIR="$2"
      shift 2
      ;;
    --host)
      require_arg "$@"
      REMOTE_HOST="$2"
      shift 2
      ;;
    --user)
      require_arg "$@"
      REMOTE_USER="$2"
      shift 2
      ;;
    --remote)
      require_arg "$@"
      REMOTE_DIR="$2"
      shift 2
      ;;
    --port)
      require_arg "$@"
      REMOTE_PORT="$2"
      shift 2
      ;;
    --identity)
      require_arg "$@"
      SSH_KEY="$2"
      shift 2
      ;;
    --exclude-file)
      require_arg "$@"
      EXCLUDE_FILE="$2"
      shift 2
      ;;
    --restart-server)
      RESTART_SERVER=1
      shift
      ;;
    --no-restart-server)
      RESTART_SERVER=0
      shift
      ;;
    --server-port)
      require_arg "$@"
      SERVER_PORT="$2"
      shift 2
      ;;
    --remote-binary)
      require_arg "$@"
      REMOTE_BINARY="$2"
      shift 2
      ;;
    --stdout-log)
      require_arg "$@"
      REMOTE_STDOUT_LOG="$2"
      shift 2
      ;;
    --stderr-log)
      require_arg "$@"
      REMOTE_STDERR_LOG="$2"
      shift 2
      ;;
    --health-timeout)
      require_arg "$@"
      HEALTH_TIMEOUT="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Opción no reconocida: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if ! command -v rsync >/dev/null 2>&1; then
  echo "Error: 'rsync' no está disponible en el sistema." >&2
  exit 2
fi

if [[ "$SSH_CLIENT" == "ssh" ]]; then
  if ! command -v ssh >/dev/null 2>&1; then
    echo "Error: 'ssh' no está disponible en el sistema." >&2
    exit 2
  fi
elif [[ "$SSH_CLIENT" == "plink" ]]; then
  if [[ -z "$PLINK_EXE" || ! -x "$PLINK_EXE" ]]; then
    echo "Error: SSH_CLIENT=plink pero PLINK_EXE no es ejecutable: '$PLINK_EXE'" >&2
    exit 2
  fi
  if [[ -z "$PLINK_KEY_WIN" ]]; then
    echo "Error: SSH_CLIENT=plink requiere PLINK_KEY_WIN con ruta Windows del .ppk." >&2
    exit 2
  fi
else
  echo "Error: SSH_CLIENT no soportado: '$SSH_CLIENT' (usa ssh o plink)." >&2
  exit 2
fi

if [[ ! -d "$LOCAL_DIR" ]]; then
  echo "Error: ruta local inexistente: $LOCAL_DIR" >&2
  exit 3
fi
LOCAL_DIR="$(cd "$LOCAL_DIR" && pwd -P)"

if ! is_uint "$REMOTE_PORT" || (( REMOTE_PORT < 1 || REMOTE_PORT > 65535 )); then
  echo "Error: REMOTE_PORT debe ser un entero entre 1 y 65535. Valor actual: $REMOTE_PORT" >&2
  exit 3
fi

if ! is_uint "$SERVER_PORT" || (( SERVER_PORT < 1 || SERVER_PORT > 65535 )); then
  echo "Error: SERVER_PORT debe ser un entero entre 1 y 65535. Valor actual: $SERVER_PORT" >&2
  exit 3
fi

if ! is_uint "$HEALTH_TIMEOUT" || (( HEALTH_TIMEOUT < 5 || HEALTH_TIMEOUT > 300 )); then
  echo "Error: HEALTH_TIMEOUT debe ser un entero entre 5 y 300 segundos. Valor actual: $HEALTH_TIMEOUT" >&2
  exit 3
fi

if [[ "$SSH_CLIENT" == "ssh" && -n "$SSH_KEY" && ! -r "$SSH_KEY" ]]; then
  echo "Error: no se puede leer la clave SSH: $SSH_KEY" >&2
  exit 4
fi

if [[ -n "$EXCLUDE_FILE" && ! -f "$EXCLUDE_FILE" ]]; then
  echo "Error: exclude-file no existe: $EXCLUDE_FILE" >&2
  exit 5
fi

REMOTE_TARGET="${REMOTE_USER}@${REMOTE_HOST}"

if [[ "$SSH_CLIENT" == "ssh" ]]; then
  SSH_ARGS=(-p "$REMOTE_PORT" -o BatchMode=yes -o ConnectTimeout=15 -o ServerAliveInterval=30 -o ServerAliveCountMax=4 -o "StrictHostKeyChecking=$SSH_STRICT_HOSTKEY" -o LogLevel=ERROR)
  if [[ -n "$SSH_KEY" ]]; then
    SSH_ARGS=(-i "$SSH_KEY" "${SSH_ARGS[@]}")
  fi
else
  PLINK_ARGS=(-batch -P "$REMOTE_PORT" -i "$PLINK_KEY_WIN")
fi

run_remote() {
  local cmd="$1"
  if [[ "$SSH_CLIENT" == "ssh" ]]; then
    ssh "${SSH_ARGS[@]}" "$REMOTE_TARGET" "$cmd"
  else
    "$PLINK_EXE" "${PLINK_ARGS[@]}" "$REMOTE_TARGET" "$cmd"
  fi
}

info "Local:  $LOCAL_DIR"
info "Remoto: $REMOTE_TARGET:$REMOTE_DIR"
info "Puerto SSH: $REMOTE_PORT"
info "DryRun: $DRY_RUN"
info "Cliente SSH: $SSH_CLIENT"
info "Bootstrap remoto: $BOOTSTRAP_SERVER"
info "Restart remoto: $RESTART_SERVER"

set_step "verificando acceso SSH"
run_remote "echo ok" >/dev/null

set_step "validando sistema operativo remoto"
REMOTE_OS="$(run_remote "uname -s 2>/dev/null || true" | tr -d '\r')"
if [[ "${REMOTE_OS,,}" != "linux" ]]; then
  echo "Error: el host remoto no reporta Linux (valor: '$REMOTE_OS')." >&2
  exit 6
fi

set_step "asegurando directorio remoto"
run_remote "mkdir -p '$REMOTE_DIR'"

RSYNC_ARGS=(-az --delete --partial --human-readable --progress --stats --timeout=180)
if (( DRY_RUN == 1 )); then
  RSYNC_ARGS+=(--dry-run --itemize-changes)
fi
for pattern in "${EXCLUDES[@]}"; do
  RSYNC_ARGS+=(--exclude "$pattern")
done
if [[ -n "$EXCLUDE_FILE" ]]; then
  RSYNC_ARGS+=(--exclude-from "$EXCLUDE_FILE")
fi

if [[ "$SSH_CLIENT" == "ssh" ]]; then
  RSYNC_SSH="ssh -p $REMOTE_PORT -o BatchMode=yes -o ConnectTimeout=15 -o ServerAliveInterval=30 -o ServerAliveCountMax=4 -o StrictHostKeyChecking=$SSH_STRICT_HOSTKEY -o LogLevel=ERROR"
  if [[ -n "$SSH_KEY" ]]; then
    RSYNC_SSH="ssh -i '$SSH_KEY' -p $REMOTE_PORT -o BatchMode=yes -o ConnectTimeout=15 -o ServerAliveInterval=30 -o ServerAliveCountMax=4 -o StrictHostKeyChecking=$SSH_STRICT_HOSTKEY -o LogLevel=ERROR"
  fi
else
  RSYNC_SSH="\"$PLINK_EXE\" -batch -P $REMOTE_PORT -i \"$PLINK_KEY_WIN\""
fi

set_step "sincronizando archivos con rsync"
rsync "${RSYNC_ARGS[@]}" -e "$RSYNC_SSH" "$LOCAL_DIR/" "$REMOTE_TARGET:$REMOTE_DIR/"

if (( DRY_RUN == 1 )); then
  info "DryRun activo: no se reinicia proceso remoto."
  ok "Sincronización completada correctamente."
  exit 0
fi

if [[ "$BOOTSTRAP_SERVER" == "1" ]]; then
  SAFE_REMOTE_DIR="$(escape_sq "$REMOTE_DIR")"
  SAFE_SERVER_PORT="$(escape_sq "$SERVER_PORT")"
  SAFE_DB_DIALECT="$(escape_sq "$DB_DIALECT")"
  SAFE_DB_EMPRESAS_DSN="$(escape_sq "$DB_EMPRESAS_DSN")"
  SAFE_DB_SUPERADMIN_DSN="$(escape_sq "$DB_SUPERADMIN_DSN")"
  SAFE_GOOGLE_CLIENT_ID="$(escape_sq "$GOOGLE_CLIENT_ID")"
  SAFE_GOOGLE_CLIENT_SECRET="$(escape_sq "$GOOGLE_CLIENT_SECRET")"
  SAFE_GOOGLE_REDIRECT_URL="$(escape_sq "$GOOGLE_REDIRECT_URL")"

  BOOTSTRAP_CMD="$(cat <<EOF_REMOTE_BOOTSTRAP
set -e;
log(){ echo \"BOOTSTRAP_STEP:\$1\"; };
warn(){ echo \"BOOTSTRAP_WARN:\$1\"; };
ok(){ echo \"BOOTSTRAP_OK:\$1\"; };
hint(){ echo \"BOOTSTRAP_HINT:\$1\"; };
fail(){ echo \"BOOTSTRAP_ERROR:\$1\"; exit 1; };
can_run_root(){
  if [ \"\$(id -u)\" -eq 0 ]; then
    return 0;
  fi;
  if command -v sudo >/dev/null 2>&1 && sudo -n true >/dev/null 2>&1; then
    return 0;
  fi;
  return 1;
};
run_root(){
  if [ \"\$(id -u)\" -eq 0 ]; then
    \"\$@\";
    return \$?;
  fi;
  sudo -n \"\$@\";
};
backend_dir='$SAFE_REMOTE_DIR/backend';
env_file=\"\$backend_dir/.env.local\";
server_port='$SAFE_SERVER_PORT';
case \"\$server_port\" in
  ''|*[!0-9]*) fail \"INVALID_SERVER_PORT SERVER_PORT debe ser numerico. Valor recibido: \$server_port\" ;;
esac;
if [ \"\$server_port\" -lt 1 ] || [ \"\$server_port\" -gt 65535 ]; then
  fail \"INVALID_SERVER_PORT SERVER_PORT fuera de rango (1-65535). Valor recibido: \$server_port\";
fi;
log \"preparando directorio remoto y archivo de entorno\";
mkdir -p \"\$backend_dir\" \"\$backend_dir/bin\" \"\$backend_dir/tmp\";
touch \"\$env_file\";
chmod 600 \"\$env_file\" || true;
ok \"ENV_FILE listo en \$env_file\";
log \"detectando sistema y dependencias del VPS\";
os_name=\"\$(uname -s 2>/dev/null || echo desconocido)\";
os_release=\"\$(uname -r 2>/dev/null || echo desconocido)\";
os_arch=\"\$(uname -m 2>/dev/null || echo desconocido)\";
ok \"SYSTEM_INFO host=\$os_name arch=\$os_arch kernel=\$os_release\";
if command -v apt-get >/dev/null 2>&1; then
  ok \"PKG_MANAGER detectado apt-get\";
  missing_base_deps=0;
  for cmd in curl wget lsof ps; do
    if ! command -v \"\$cmd\" >/dev/null 2>&1; then missing_base_deps=1; fi;
  done;
  if [ \"\$missing_base_deps\" -eq 0 ]; then
    ok \"PKG_INSTALL_SKIP paquetes base ya disponibles; se omite apt-get update/install\";
  elif can_run_root; then
    export DEBIAN_FRONTEND=noninteractive;
    run_root apt-get update -y >/dev/null 2>&1 || warn \"APT_UPDATE apt-get update reporto incidencias; se intentara instalar de todos modos\";
    if run_root apt-get install -y ca-certificates curl wget procps lsof >/dev/null 2>&1; then
      ok \"PKG_INSTALL paquetes base instalados con apt-get: ca-certificates curl wget procps lsof\";
    else
      fail \"PACKAGE_INSTALL_FAILED fallo apt-get install de paquetes base\";
    fi;
  else
    warn \"PRIVILEGE_REQUIRED sin root ni sudo -n; se omite instalacion automatica de paquetes\";
    hint \"Conectate como root o habilita sudo sin contraseña si quieres que el script prepare dependencias del VPS\";
  fi;
elif command -v dnf >/dev/null 2>&1; then
  ok \"PKG_MANAGER detectado dnf\";
  if can_run_root; then
    if run_root dnf install -y ca-certificates curl wget procps-ng lsof >/dev/null 2>&1; then
      ok \"PKG_INSTALL paquetes base instalados con dnf: ca-certificates curl wget procps-ng lsof\";
    else
      fail \"PACKAGE_INSTALL_FAILED fallo dnf install de paquetes base\";
    fi;
  else
    warn \"PRIVILEGE_REQUIRED sin root ni sudo -n; se omite instalacion automatica de paquetes\";
  fi;
elif command -v yum >/dev/null 2>&1; then
  ok \"PKG_MANAGER detectado yum\";
  if can_run_root; then
    if run_root yum install -y ca-certificates curl wget procps-ng lsof >/dev/null 2>&1; then
      ok \"PKG_INSTALL paquetes base instalados con yum: ca-certificates curl wget procps-ng lsof\";
    else
      fail \"PACKAGE_INSTALL_FAILED fallo yum install de paquetes base\";
    fi;
  else
    warn \"PRIVILEGE_REQUIRED sin root ni sudo -n; se omite instalacion automatica de paquetes\";
  fi;
elif command -v apk >/dev/null 2>&1; then
  ok \"PKG_MANAGER detectado apk\";
  if can_run_root; then
    if run_root apk add --no-cache ca-certificates curl wget procps lsof >/dev/null 2>&1; then
      ok \"PKG_INSTALL paquetes base instalados con apk: ca-certificates curl wget procps lsof\";
    else
      fail \"PACKAGE_INSTALL_FAILED fallo apk add de paquetes base\";
    fi;
  else
    warn \"PRIVILEGE_REQUIRED sin root ni sudo -n; se omite instalacion automatica de paquetes\";
  fi;
elif command -v zypper >/dev/null 2>&1; then
  ok \"PKG_MANAGER detectado zypper\";
  if can_run_root; then
    if run_root zypper --non-interactive install ca-certificates curl wget procps lsof >/dev/null 2>&1; then
      ok \"PKG_INSTALL paquetes base instalados con zypper: ca-certificates curl wget procps lsof\";
    else
      fail \"PACKAGE_INSTALL_FAILED fallo zypper install de paquetes base\";
    fi;
  else
    warn \"PRIVILEGE_REQUIRED sin root ni sudo -n; se omite instalacion automatica de paquetes\";
  fi;
else
  warn \"PKG_MANAGER_UNKNOWN no se detecto apt-get, dnf, yum, apk ni zypper\";
  hint \"Verifica manualmente que el VPS tenga ca-certificates, curl o wget, lsof y utilidades base antes del reinicio\";
fi;
if command -v systemctl >/dev/null 2>&1; then
  ok \"SYSTEMD_OK systemctl disponible\";
else
  warn \"SYSTEMD_MISSING systemctl no esta disponible; el backend persistente requiere systemd activo\";
  hint \"Usa una VM Linux con systemd o ajusta manualmente el metodo de arranque del VPS\";
fi;
db_dialect='$SAFE_DB_DIALECT';
db_empresas_dsn='$SAFE_DB_EMPRESAS_DSN';
db_superadmin_dsn='$SAFE_DB_SUPERADMIN_DSN';
google_client_id='$SAFE_GOOGLE_CLIENT_ID';
google_client_secret='$SAFE_GOOGLE_CLIENT_SECRET';
google_redirect_url='$SAFE_GOOGLE_REDIRECT_URL';
get_env_value(){ grep -E \"^\$1=\" \"\$env_file\" | tail -n1 | cut -d= -f2- || true; };
upsert_env(){
  key=\"\$1\";
  value=\"\$2\";
  grep -v \"^\$key=\" \"\$env_file\" > \"\$env_file.tmp\" 2>/dev/null || true;
  mv \"\$env_file.tmp\" \"\$env_file\" 2>/dev/null || true;
  printf '%s=%s\\n' \"\$key\" \"\$value\" >> \"\$env_file\";
};
current_dbdialect=\"\$(get_env_value DB_DIALECT)\";
current_dbemp=\"\$(get_env_value DB_EMPRESAS_DSN)\";
current_dbsuper=\"\$(get_env_value DB_SUPERADMIN_DSN)\";
current_gid=\"\$(get_env_value GOOGLE_CLIENT_ID)\";
current_gsec=\"\$(get_env_value GOOGLE_CLIENT_SECRET)\";
current_grurl=\"\$(get_env_value GOOGLE_REDIRECT_URL)\";
effective_dbdialect=\"\$db_dialect\";
effective_dbemp=\"\$db_empresas_dsn\";
effective_dbsuper=\"\$db_superadmin_dsn\";
effective_gid=\"\$google_client_id\";
effective_gsec=\"\$google_client_secret\";
effective_grurl=\"\$google_redirect_url\";
if [ -z \"\$effective_dbdialect\" ]; then effective_dbdialect=\"\$current_dbdialect\"; fi;
if [ -z \"\$effective_dbemp\" ]; then effective_dbemp=\"\$current_dbemp\"; fi;
if [ -z \"\$effective_dbsuper\" ]; then effective_dbsuper=\"\$current_dbsuper\"; fi;
if [ -z \"\$effective_gid\" ]; then effective_gid=\"\$current_gid\"; fi;
if [ -z \"\$effective_gsec\" ]; then effective_gsec=\"\$current_gsec\"; fi;
if [ -z \"\$effective_grurl\" ]; then effective_grurl=\"\$current_grurl\"; fi;
if [ -z \"\$effective_dbdialect\" ] && { [ -n \"\$effective_dbemp\" ] || [ -n \"\$effective_dbsuper\" ]; }; then
  effective_dbdialect=postgres;
fi;
if [ \"\$effective_dbdialect\" = \"postgres\" ] && { [ -z \"\$effective_dbemp\" ] || [ -z \"\$effective_dbsuper\" ]; }; then
  echo \"BOOTSTRAP_ERROR:POSTGRES_MISSING_DSN faltan DB_EMPRESAS_DSN y/o DB_SUPERADMIN_DSN para modo postgres\";
  echo \"BOOTSTRAP_HINT:Define DbEmpresasDsn y DbSuperadminDsn, o deja ambos DSN validos en backend/.env.local del VPS\";
  exit 1;
fi;
log \"sincronizando backend/.env.local remoto\";
upsert_env SERVER_PORT \"\$server_port\";
ok \"SERVER_PORT actualizado a \$server_port\";
if [ -n \"\$effective_dbdialect\" ]; then upsert_env DB_DIALECT \"\$effective_dbdialect\"; fi;
if [ -n \"\$effective_dbemp\" ]; then upsert_env DB_EMPRESAS_DSN \"\$effective_dbemp\"; fi;
if [ -n \"\$effective_dbsuper\" ]; then upsert_env DB_SUPERADMIN_DSN \"\$effective_dbsuper\"; fi;
if [ -n \"\$effective_gid\" ]; then upsert_env GOOGLE_CLIENT_ID \"\$effective_gid\"; fi;
if [ -n \"\$effective_gsec\" ]; then upsert_env GOOGLE_CLIENT_SECRET \"\$effective_gsec\"; fi;
if [ -n \"\$effective_grurl\" ]; then upsert_env GOOGLE_REDIRECT_URL \"\$effective_grurl\"; fi;
for k in DB_DIALECT DB_SUPERADMIN_DSN DB_EMPRESAS_DSN GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET GOOGLE_REDIRECT_URL SERVER_PORT CONFIG_ENC_KEY; do
  line=\"\$(grep -E \"^\$k=\" \"\$env_file\" | tail -n1 || true)\";
  if [ -z \"\$line\" ]; then
    case \"\$k\" in
      GOOGLE_CLIENT_ID|GOOGLE_CLIENT_SECRET|GOOGLE_REDIRECT_URL)
        echo \"BOOTSTRAP_WARN:\$k ausente (solo requerido para login Google)\";
        ;;
      CONFIG_ENC_KEY)
        echo \"BOOTSTRAP_WARN:CONFIG_ENC_KEY ausente (requerida para cifrado de secretos)\";
        echo \"BOOTSTRAP_HINT:Define CONFIG_ENC_KEY en backend/.env.local antes de guardar credenciales sensibles\";
        ;;
      *)
        echo \"BOOTSTRAP_WARN:\$k ausente en backend/.env.local\";
        ;;
    esac;
  else
    val=\"\${line#*=}\";
    if [ -z \"\$val\" ]; then
      case \"\$k\" in
        GOOGLE_CLIENT_ID|GOOGLE_CLIENT_SECRET|GOOGLE_REDIRECT_URL)
          echo \"BOOTSTRAP_WARN:\$k vacio (solo requerido para login Google)\";
          ;;
        CONFIG_ENC_KEY)
          echo \"BOOTSTRAP_WARN:CONFIG_ENC_KEY vacia (requerida para cifrado de secretos)\";
          echo \"BOOTSTRAP_HINT:Define CONFIG_ENC_KEY en backend/.env.local antes de guardar credenciales sensibles\";
          ;;
        *)
          echo \"BOOTSTRAP_WARN:\$k vacio en backend/.env.local\";
          ;;
      esac;
    else
      echo \"BOOTSTRAP_OK:\$k configurado\";
    fi;
  fi;
done;
ok \"BOOTSTRAP_COMPLETE entorno remoto preparado para el redeploy\";
EOF_REMOTE_BOOTSTRAP
)"

  set_step "bootstrap remoto del VPS"
  run_remote "$BOOTSTRAP_CMD"
fi

if (( RESTART_SERVER == 1 )); then
  SAFE_REMOTE_DIR="$(escape_sq "$REMOTE_DIR")"
  SAFE_REMOTE_BINARY="$(escape_sq "${REMOTE_BINARY#/}")"
  SAFE_STDOUT_LOG="$(escape_sq "${REMOTE_STDOUT_LOG#/}")"
  SAFE_STDERR_LOG="$(escape_sq "${REMOTE_STDERR_LOG#/}")"
  SAFE_SERVER_PORT="$(escape_sq "$SERVER_PORT")"
  SAFE_HEALTH_TIMEOUT="$(escape_sq "$HEALTH_TIMEOUT")"

  RESTART_CMD="$(cat <<EOF_REMOTE_RESTART
set -e;
log(){ echo \"DEPLOY_STEP:\$1\"; };
warn(){ echo \"DEPLOY_WARN:\$1\"; };
hint(){ echo \"DEPLOY_HINT:\$1\"; };
fail(){ echo \"DEPLOY_ERROR:\$1\"; exit 1; };
can_run_root(){
  if [ \"\$(id -u)\" -eq 0 ]; then
    return 0;
  fi;
  if command -v sudo >/dev/null 2>&1 && sudo -n true >/dev/null 2>&1; then
    return 0;
  fi;
  return 1;
};
run_root(){
  if [ \"\$(id -u)\" -eq 0 ]; then
    \"\$@\";
    return \$?;
  fi;
  sudo -n \"\$@\";
};
repo_dir='$SAFE_REMOTE_DIR';
backend_dir=\$repo_dir/backend;
bin_rel='$SAFE_REMOTE_BINARY';
stdout_rel='$SAFE_STDOUT_LOG';
stderr_rel='$SAFE_STDERR_LOG';
port='$SAFE_SERVER_PORT';
health_timeout='$SAFE_HEALTH_TIMEOUT';
bin_path=\$repo_dir/\$bin_rel;
stdout_log=\$repo_dir/\$stdout_rel;
stderr_log=\$repo_dir/\$stderr_rel;
env_file=\$backend_dir/.env.local;
pid_file=\$repo_dir/backend/server.pid;
service_base=\$(basename \"\$repo_dir\");
service_name=\$(printf '%s' \"\$service_base\" | tr -c 'A-Za-z0-9_.@-' '_');
service_unit=\$service_name.service;
service_file=/etc/systemd/system/\$service_unit;
dump_diagnostics(){
  echo \"DEPLOY_LOG:systemctl status \$service_unit\";
  run_root systemctl status \"\$service_unit\" --no-pager -l || true;
  if command -v journalctl >/dev/null 2>&1; then
    echo \"DEPLOY_LOG:journalctl -u \$service_unit -n 80\";
    run_root journalctl -u \"\$service_unit\" -n 80 --no-pager || true;
  fi;
  if command -v ss >/dev/null 2>&1; then
    echo \"DEPLOY_LOG:ss -ltnp (*:\$port)\";
    run_root ss -ltnp 2>/dev/null | awk -v port=\"\$port\" '\$4 ~ \":\" port \"\\$\" { print; }' || true;
  fi;
  if [ -f \"\$stderr_log\" ]; then
    echo \"DEPLOY_LOG:tail -n 80 \$stderr_log\";
    tail -n 80 \"\$stderr_log\" || true;
  fi;
  if [ -f \"\$stdout_log\" ]; then
    echo \"DEPLOY_LOG:tail -n 40 \$stdout_log\";
    tail -n 40 \"\$stdout_log\" || true;
  fi;
};
pids_listening_on_port(){
  if command -v ss >/dev/null 2>&1; then
    run_root ss -ltnp 2>/dev/null | awk -v port=\"\$port\" '
      \$4 ~ \":\" port \"\\$\" {
        line = \$0;
        while (match(line, /pid=[0-9]+/)) {
          pid = substr(line, RSTART + 4, RLENGTH - 4);
          print pid;
          line = substr(line, RSTART + RLENGTH);
        }
      }
    ' | sort -u;
    return 0;
  fi;
  if command -v lsof >/dev/null 2>&1; then
    run_root lsof -tiTCP:\"\$port\" -sTCP:LISTEN 2>/dev/null | sort -u;
    return 0;
  fi;
  return 1;
};
skip_cleanup_pid(){
  pid=\"\$1\";
  case \"\$pid\" in
    ''|*[!0-9]*|0|1)
      return 0;
      ;;
  esac;
  if [ \"\$pid\" = \"\$\$\" ] || [ \"\$pid\" = \"\$PPID\" ]; then
    return 0;
  fi;
  return 1;
};
cleanup_port_conflicts(){
  run_root systemctl stop \"\$service_unit\" >/dev/null 2>&1 || true;
  stray_pids=\$(pids_listening_on_port || true);
  if [ -z \"\$stray_pids\" ]; then
    return 0;
  fi;
  echo \"DEPLOY_WARN:PORT_IN_USE se detectaron procesos fuera de systemd ocupando el puerto \$port; se intentara liberarlo\";
  for pid in \$stray_pids; do
    if skip_cleanup_pid \"\$pid\"; then
      continue;
    fi;
    exe_path=\$(readlink -f \"/proc/\$pid/exe\" 2>/dev/null || true);
    cmdline=\$(tr '\\0' ' ' <\"/proc/\$pid/cmdline\" 2>/dev/null | sed 's/[[:space:]]*\$//');
    echo \"DEPLOY_WARN:PORT_PID pid=\$pid exe=\${exe_path:-unknown} cmd=\${cmdline:-unknown}\";
    run_root kill \"\$pid\" >/dev/null 2>&1 || true;
  done;
  sleep 2;
  stray_pids=\$(pids_listening_on_port || true);
  if [ -n \"\$stray_pids\" ]; then
    for pid in \$stray_pids; do
      if skip_cleanup_pid \"\$pid\"; then
        continue;
      fi;
      echo \"DEPLOY_WARN:PORT_PID_FORCE pid=\$pid\";
      run_root kill -9 \"\$pid\" >/dev/null 2>&1 || true;
    done;
    sleep 1;
  fi;
  stray_pids=\$(pids_listening_on_port || true);
  if [ -n \"\$stray_pids\" ]; then
    echo \"DEPLOY_ERROR:PORT_STILL_BUSY no se pudo liberar el puerto \$port (pids=\$(printf '%s' \"\$stray_pids\" | tr '\\n' ' ' | sed 's/[[:space:]]*\$//'))\";
    echo \"DEPLOY_HINT:Deten manualmente los procesos que aun escuchan en el puerto o cambia SERVER_PORT antes de reintentar\";
    dump_diagnostics;
    exit 1;
  fi;
};
log \"preparando servicio systemd persistente\";
if ! can_run_root; then
  fail \"PRIVILEGE_REQUIRED se requiere root o sudo -n para instalar o reiniciar la unidad systemd\";
fi;
if ! command -v systemctl >/dev/null 2>&1; then
  fail \"SYSTEMD_UNAVAILABLE systemctl no esta disponible en el VPS\";
fi;
if [ ! -f \"\$bin_path\" ]; then
  fail \"BIN_NOT_FOUND binario remoto no encontrado en \$bin_path\";
fi;
run_root mkdir -p \"\$(dirname \"\$stdout_log\")\" \"\$(dirname \"\$stderr_log\")\";
run_root touch \"\$stdout_log\" \"\$stderr_log\";
chmod +x \"\$bin_path\" || true;
rm -f \"\$pid_file\" 2>/dev/null || true;
tmp_service=\$(mktemp \"\${TMPDIR:-/tmp}/pcs_sync_service.XXXXXX\");
cat > \"\$tmp_service\" <<EOF_SERVICE_FILE
[Unit]
Description=Powerful Control System backend (\$service_name)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=\$backend_dir
EnvironmentFile=-\$env_file
ExecStart=\$bin_path
Restart=always
RestartSec=5
StartLimitIntervalSec=0
KillSignal=SIGTERM
TimeoutStopSec=30
StandardOutput=append:\$stdout_log
StandardError=append:\$stderr_log

[Install]
WantedBy=multi-user.target
EOF_SERVICE_FILE
run_root cp \"\$tmp_service\" \"\$service_file\";
rm -f \"\$tmp_service\";
echo \"DEPLOY_STEP:unidad systemd actualizada en \$service_file\";
run_root systemctl daemon-reload;
run_root systemctl enable \"\$service_unit\" >/dev/null 2>&1;
run_root systemctl reset-failed \"\$service_unit\" >/dev/null 2>&1 || true;
cleanup_port_conflicts;
log \"reiniciando \$service_unit y validando arranque\";
if ! run_root systemctl start \"\$service_unit\"; then
  echo \"DEPLOY_ERROR:SERVICE_RESTART_FAILED no fue posible reiniciar \$service_unit en el puerto \$port\";
  echo \"DEPLOY_HINT:Revisa backend/.env.local, los DSN PostgreSQL, CONFIG_ENC_KEY y el binario remoto antes de reintentar\";
  dump_diagnostics;
  exit 1;
fi;
healthy=0;
i=1;
while [ \$i -le \$health_timeout ]; do
  if ! run_root systemctl is-active --quiet \"\$service_unit\"; then
    echo \"DEPLOY_ERROR:SERVICE_NOT_RUNNING el servicio \$service_unit se detuvo durante el healthcheck\";
    echo \"DEPLOY_HINT:Revisa backend/.env.local, el puerto SERVER_PORT=\$port y los logs mostrados abajo\";
    dump_diagnostics;
    exit 1;
  fi;
  if command -v curl >/dev/null 2>&1; then
    http_code=\$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:\$port/ || true);
    if [ -n \"\$http_code\" ] && [ \"\$http_code\" != \"000\" ]; then healthy=1; break; fi;
  elif command -v wget >/dev/null 2>&1; then
    if wget -qO- http://127.0.0.1:\$port/ >/dev/null 2>&1; then healthy=1; break; fi;
  else
    healthy=1; break;
  fi;
  sleep 1;
  i=\$((i+1));
done;
main_pid=\$(run_root systemctl show -p MainPID --value \"\$service_unit\" 2>/dev/null || echo 0);
enabled_state=\$(run_root systemctl is-enabled \"\$service_unit\" 2>/dev/null || true);
if [ \$healthy -eq 1 ]; then
  echo \"DEPLOY_OK:SERVICE_READY servicio \$service_unit activo (pid=\$main_pid, puerto=\$port, enabled=\$enabled_state)\";
else
  echo \"DEPLOY_WARN:HEALTHCHECK_TIMEOUT el servicio \$service_unit quedo activo (pid=\$main_pid, enabled=\$enabled_state) pero no respondio al healthcheck en \$health_timeout s\";
  echo \"DEPLOY_HINT:Verifica que SERVER_PORT=\$port coincida con backend/.env.local y que GET / responda localmente en el VPS\";
fi;
EOF_REMOTE_RESTART
)"

  set_step "reiniciando backend remoto con systemd"
  run_remote "$RESTART_CMD"
fi

ok "Sincronización completada correctamente."
