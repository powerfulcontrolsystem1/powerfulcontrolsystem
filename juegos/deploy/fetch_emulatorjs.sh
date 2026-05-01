#!/usr/bin/env sh
set -eu

# Descarga los archivos oficiales estables de EmulatorJS para servirlos localmente.
# Uso: sh deploy/fetch_emulatorjs.sh

TARGET_DIR="${1:-./emulator}"
TMP_DIR="$(mktemp -d)"
CORES="${PCS_EMULATORJS_CORES:-snes9x fceumm gambatte mgba genesis_plus_gx mupen64plus_next}"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

mkdir -p "$TARGET_DIR"

echo "Descargando EmulatorJS estable desde GitHub..."
curl -L "https://github.com/EmulatorJS/EmulatorJS/archive/refs/heads/main.zip" -o "$TMP_DIR/emulatorjs.zip"
unzip -q "$TMP_DIR/emulatorjs.zip" -d "$TMP_DIR"

SRC_DATA="$(find "$TMP_DIR" -type d -path '*/data' | head -n 1)"
if [ -z "$SRC_DATA" ]; then
  echo "No se encontro carpeta data en el paquete descargado." >&2
  exit 1
fi

rm -rf "$TARGET_DIR/data"
mkdir -p "$TARGET_DIR"
cp -R "$SRC_DATA" "$TARGET_DIR/data"

echo "Alineando loader y metadatos con la version estable..."
curl -fL "https://cdn.emulatorjs.org/stable/data/loader.js" -o "$TARGET_DIR/data/loader.js"
curl -fL "https://cdn.emulatorjs.org/stable/data/version.json" -o "$TARGET_DIR/data/version.json"
curl -fL "https://cdn.emulatorjs.org/stable/data/emulator.css" -o "$TARGET_DIR/data/emulator.css"

echo "Descargando build minificado estable..."
curl -L "https://cdn.emulatorjs.org/stable/data/emulator.min.zip" -o "$TMP_DIR/emulator.min.zip"
unzip -oq "$TMP_DIR/emulator.min.zip" -d "$TMP_DIR/emulator_min"
find "$TMP_DIR/emulator_min" -type f \( -name 'emulator.min.js' -o -name 'emulator.min.css' \) -exec cp {} "$TARGET_DIR/data/" \;

mkdir -p "$TARGET_DIR/data/cores/reports"
for core in $CORES; do
  echo "Descargando core EmulatorJS: $core"
  curl -fL "https://cdn.emulatorjs.org/stable/data/cores/reports/$core.json" -o "$TARGET_DIR/data/cores/reports/$core.json" || true
  curl -fL "https://cdn.emulatorjs.org/stable/data/cores/$core-wasm.data" -o "$TARGET_DIR/data/cores/$core-wasm.data"
  curl -fL "https://cdn.emulatorjs.org/stable/data/cores/$core-legacy-wasm.data" -o "$TARGET_DIR/data/cores/$core-legacy-wasm.data" || true
done

echo "EmulatorJS instalado en $TARGET_DIR/data"
