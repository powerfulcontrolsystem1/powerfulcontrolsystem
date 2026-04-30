#!/usr/bin/env sh
set -eu

# Descarga los archivos oficiales estables de EmulatorJS para servirlos localmente.
# Uso: sh deploy/fetch_emulatorjs.sh

TARGET_DIR="${1:-./emulator}"
TMP_DIR="$(mktemp -d)"

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

echo "EmulatorJS instalado en $TARGET_DIR/data"
