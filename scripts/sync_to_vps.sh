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

DRY_RUN=0
EXCLUDE_FILE=""

EXCLUDES=(
  ".git"
  ".gitignore"
  "node_modules"
  "logs"
  "test_runs"
  "*.db"
  "*.sqlite"
  "*.exe"
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
  -h, --help               Mostrar ayuda

Variables opcionales:
  SSH_STRICT_HOSTKEY       accept-new | yes | no (default: accept-new)
  SSH_CLIENT               ssh | plink (default: ssh)
  PLINK_EXE                Ruta a plink.exe (requerido si SSH_CLIENT=plink)
  PLINK_KEY_WIN            Ruta Windows a .ppk (requerido si SSH_CLIENT=plink)

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

echo "[OK] Sincronización completada correctamente."
