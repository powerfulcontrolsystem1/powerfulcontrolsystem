#!/usr/bin/env bash
# sync_to_vps.sh
# Sincroniza una carpeta local hacia un VPS Linux usando rsync sobre SSH.

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
  ".gitignore"
  "node_modules"
  "logs"
  "test_runs"
  "*.db"
  "*.exe"
  "backend/.env.local"
  "backend/server.err"
)

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
  SSH_ARGS=(-p "$REMOTE_PORT" -o BatchMode=yes -o "StrictHostKeyChecking=$SSH_STRICT_HOSTKEY" -o LogLevel=ERROR)
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

echo "[INFO] Local:  $LOCAL_DIR"
echo "[INFO] Remoto: $REMOTE_TARGET:$REMOTE_DIR"
echo "[INFO] Puerto: $REMOTE_PORT"
echo "[INFO] DryRun: $DRY_RUN"
echo "[INFO] Cliente SSH: $SSH_CLIENT"
echo "[INFO] Restart remoto: $RESTART_SERVER"

echo "[INFO] Verificando acceso SSH..."
run_remote "echo ok" >/dev/null

echo "[INFO] Validando SO remoto..."
REMOTE_OS="$(run_remote "uname -s 2>/dev/null || true" | tr -d '\r')"
if [[ "${REMOTE_OS,,}" != "linux" ]]; then
  echo "Error: el host remoto no reporta Linux (valor: '$REMOTE_OS')." >&2
  exit 6
fi

echo "[INFO] Asegurando directorio remoto..."
run_remote "mkdir -p '$REMOTE_DIR'"

RSYNC_ARGS=(-az --delete --partial --human-readable --progress)
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
  RSYNC_SSH="ssh -p $REMOTE_PORT -o BatchMode=yes -o StrictHostKeyChecking=$SSH_STRICT_HOSTKEY -o LogLevel=ERROR"
  if [[ -n "$SSH_KEY" ]]; then
    RSYNC_SSH="ssh -i '$SSH_KEY' -p $REMOTE_PORT -o BatchMode=yes -o StrictHostKeyChecking=$SSH_STRICT_HOSTKEY -o LogLevel=ERROR"
  fi
else
  RSYNC_SSH="\"$PLINK_EXE\" -batch -P $REMOTE_PORT -i \"$PLINK_KEY_WIN\""
fi

echo "[INFO] Ejecutando rsync..."
rsync "${RSYNC_ARGS[@]}" -e "$RSYNC_SSH" "$LOCAL_DIR/" "$REMOTE_TARGET:$REMOTE_DIR/"

if (( DRY_RUN == 1 )); then
  echo "[INFO] DryRun activo: no se reinicia proceso remoto."
  echo "[OK] Sincronización completada correctamente."
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

  BOOTSTRAP_CMD="set -e;
backend_dir='$SAFE_REMOTE_DIR/backend';
env_file=\"\$backend_dir/.env.local\";
mkdir -p \"\$backend_dir\";
touch \"\$env_file\";
chmod 600 \"\$env_file\" || true;
if ! grep -q '^SERVER_PORT=' \"\$env_file\" 2>/dev/null; then echo SERVER_PORT=$SAFE_SERVER_PORT >> \"\$env_file\"; fi;
db_dialect='$SAFE_DB_DIALECT';
db_empresas_dsn='$SAFE_DB_EMPRESAS_DSN';
db_superadmin_dsn='$SAFE_DB_SUPERADMIN_DSN';
google_client_id='$SAFE_GOOGLE_CLIENT_ID';
google_client_secret='$SAFE_GOOGLE_CLIENT_SECRET';
google_redirect_url='$SAFE_GOOGLE_REDIRECT_URL';
current_dbdialect=\"\$(grep -E '^DB_DIALECT=' \"\$env_file\" | tail -n1 | cut -d= -f2- || true)\";
current_dbemp=\"\$(grep -E '^DB_EMPRESAS_DSN=' \"\$env_file\" | tail -n1 | cut -d= -f2- || true)\";
current_dbsuper=\"\$(grep -E '^DB_SUPERADMIN_DSN=' \"\$env_file\" | tail -n1 | cut -d= -f2- || true)\";
effective_dbdialect=\"\$db_dialect\";
effective_dbemp=\"\$db_empresas_dsn\";
effective_dbsuper=\"\$db_superadmin_dsn\";
if [ -z \"\$effective_dbdialect\" ]; then effective_dbdialect=\"\$current_dbdialect\"; fi;
if [ -z \"\$effective_dbemp\" ]; then effective_dbemp=\"\$current_dbemp\"; fi;
if [ -z \"\$effective_dbsuper\" ]; then effective_dbsuper=\"\$current_dbsuper\"; fi;
if [ -z \"\$effective_dbdialect\" ] && { [ -n \"\$effective_dbemp\" ] || [ -n \"\$effective_dbsuper\" ]; }; then
  effective_dbdialect=postgres;
fi;
if [ \"\$effective_dbdialect\" = \"postgres\" ] && { [ -z \"\$effective_dbemp\" ] || [ -z \"\$effective_dbsuper\" ]; }; then
  echo \"BOOTSTRAP_ERROR:POSTGRES_MISSING_DSN\";
  echo \"BOOTSTRAP_HINT:Define DB_EMPRESAS_DSN and DB_SUPERADMIN_DSN\";
  exit 1;
fi;
for key in DB_DIALECT DB_EMPRESAS_DSN DB_SUPERADMIN_DSN; do
  grep -v \"^\$key=\" \"\$env_file\" > \"\$env_file.tmp\" 2>/dev/null || true;
  mv \"\$env_file.tmp\" \"\$env_file\" 2>/dev/null || true;
done;
if [ -n \"\$effective_dbdialect\" ]; then echo \"DB_DIALECT=\$effective_dbdialect\" >> \"\$env_file\"; fi;
if [ -n \"\$effective_dbemp\" ]; then echo \"DB_EMPRESAS_DSN=\$effective_dbemp\" >> \"\$env_file\"; fi;
if [ -n \"\$effective_dbsuper\" ]; then echo \"DB_SUPERADMIN_DSN=\$effective_dbsuper\" >> \"\$env_file\"; fi;
if [ -n \"\$google_client_id\" ]; then
  grep -v '^GOOGLE_CLIENT_ID=' \"\$env_file\" > \"\$env_file.tmp\" 2>/dev/null || true;
  mv \"\$env_file.tmp\" \"\$env_file\" 2>/dev/null || true;
  echo \"GOOGLE_CLIENT_ID=\$google_client_id\" >> \"\$env_file\";
fi;
if [ -n \"\$google_client_secret\" ]; then
  grep -v '^GOOGLE_CLIENT_SECRET=' \"\$env_file\" > \"\$env_file.tmp\" 2>/dev/null || true;
  mv \"\$env_file.tmp\" \"\$env_file\" 2>/dev/null || true;
  echo \"GOOGLE_CLIENT_SECRET=\$google_client_secret\" >> \"\$env_file\";
fi;
if [ -n "\$google_redirect_url" ]; then
  grep -v '^GOOGLE_REDIRECT_URL=' "\$env_file" > "\$env_file.tmp" 2>/dev/null || true;
  mv "\$env_file.tmp" "\$env_file" 2>/dev/null || true;
  echo "GOOGLE_REDIRECT_URL=\$google_redirect_url" >> "\$env_file";
fi;
for k in DB_DIALECT DB_SUPERADMIN_DSN DB_EMPRESAS_DSN GOOGLE_CLIENT_ID GOOGLE_CLIENT_SECRET GOOGLE_REDIRECT_URL SERVER_PORT CONFIG_ENC_KEY; do
  line=\"\$(grep -E \"^\$k=\" \"\$env_file\" | tail -n1 || true)\";
  if [ -z \"\$line\" ]; then
    echo \"BOOTSTRAP_WARN:\$k=MISSING\";
  else
    val=\"\${line#*=}\";
    if [ -z \"\$val\" ]; then
      echo \"BOOTSTRAP_WARN:\$k=EMPTY\";
    else
      echo \"BOOTSTRAP_OK:\$k=SET\";
    fi;
  fi;
done"

  echo "[INFO] Ejecutando bootstrap remoto de variables..."
  run_remote "$BOOTSTRAP_CMD"
fi

if (( RESTART_SERVER == 1 )); then
  SAFE_REMOTE_DIR="$(escape_sq "$REMOTE_DIR")"
  SAFE_REMOTE_BINARY="$(escape_sq "${REMOTE_BINARY#/}")"
  SAFE_STDOUT_LOG="$(escape_sq "${REMOTE_STDOUT_LOG#/}")"
  SAFE_STDERR_LOG="$(escape_sq "${REMOTE_STDERR_LOG#/}")"
  SAFE_SERVER_PORT="$(escape_sq "$SERVER_PORT")"
  SAFE_HEALTH_TIMEOUT="$(escape_sq "$HEALTH_TIMEOUT")"

  RESTART_CMD="set -e;
repo_dir='$SAFE_REMOTE_DIR';
bin_rel='$SAFE_REMOTE_BINARY';
bin_name=\$(basename \$bin_rel);
stdout_rel='$SAFE_STDOUT_LOG';
stderr_rel='$SAFE_STDERR_LOG';
port='$SAFE_SERVER_PORT';
health_timeout='$SAFE_HEALTH_TIMEOUT';
bin_path=\$repo_dir/\$bin_rel;
stdout_log=\$repo_dir/\$stdout_rel;
stderr_log=\$repo_dir/\$stderr_rel;
pid_file=\$repo_dir/backend/server.pid;
mkdir -p \$(dirname \$stdout_log) \$(dirname \$stderr_log);
if [ ! -f \$bin_path ]; then echo DEPLOY_ERROR:bin_not_found path=\$bin_path; exit 1; fi;
chmod +x \$bin_path || true;
old_pid=0;
if [ -f \$pid_file ]; then old_pid=\$(cat \$pid_file 2>/dev/null || echo 0); fi;
if [ \${old_pid:-0} -gt 0 ] 2>/dev/null && kill -0 \$old_pid 2>/dev/null; then
  kill \$old_pid 2>/dev/null || true;
  for i in \$(seq 1 15); do kill -0 \$old_pid 2>/dev/null || break; sleep 1; done;
  if kill -0 \$old_pid 2>/dev/null; then kill -9 \$old_pid 2>/dev/null || true; fi;
fi;
for pid in \$(pgrep -f \$bin_name 2>/dev/null || true); do
  if [ \${pid:-0} -le 0 ] 2>/dev/null; then continue; fi;
  if [ \$pid -eq \$\$ ] 2>/dev/null || [ \$pid -eq \$PPID ] 2>/dev/null; then continue; fi;
  kill \$pid 2>/dev/null || true;
done;
sleep 1;
for pid in \$(pgrep -f \$bin_name 2>/dev/null || true); do
  if [ \${pid:-0} -le 0 ] 2>/dev/null; then continue; fi;
  if [ \$pid -eq \$\$ ] 2>/dev/null || [ \$pid -eq \$PPID ] 2>/dev/null; then continue; fi;
  kill -9 \$pid 2>/dev/null || true;
done;
nohup \$bin_path >> \$stdout_log 2>> \$stderr_log < /dev/null &
new_pid=\$!;
echo \$new_pid > \$pid_file;
healthy=0;
i=1;
while [ \$i -le \$health_timeout ]; do
  if ! kill -0 \$new_pid 2>/dev/null; then
    echo DEPLOY_ERROR:process_not_running pid=\$new_pid port=\$port;
    exit 1;
  fi;
  if command -v curl >/dev/null 2>&1; then
    http_code=\$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:\$port/ || true);
    if [ -n "\$http_code" ] && [ "\$http_code" != "000" ]; then healthy=1; break; fi;
  elif command -v wget >/dev/null 2>&1; then
    if wget -qO- http://127.0.0.1:\$port/ >/dev/null 2>&1; then healthy=1; break; fi;
  else
    if kill -0 \$new_pid 2>/dev/null; then healthy=1; break; fi;
  fi;
  sleep 1;
  i=\$((i+1));
done;
if [ \$healthy -eq 1 ]; then echo DEPLOY_OK:pid=\$new_pid port=\$port; else echo DEPLOY_WARN:healthcheck_timeout pid=\$new_pid port=\$port; fi"

  echo "[INFO] Reiniciando backend remoto con nueva versión..."
  run_remote "$RESTART_CMD"
fi

echo "[OK] Sincronización completada correctamente."
