#!/usr/bin/env bash
set -euo pipefail

if docker buildx version >/dev/null 2>&1; then
  echo "[OK] Docker Buildx ya esta instalado."
  docker buildx version
  exit 0
fi

if command -v apt-get >/dev/null 2>&1; then
  apt-get update
  apt-get install -y docker-buildx-plugin
elif command -v apk >/dev/null 2>&1; then
  apk add --no-cache docker-cli-buildx
else
  echo "[ERROR] Gestor de paquetes no soportado para instalar Buildx." >&2
  exit 1
fi

docker buildx version
echo "[OK] Docker Buildx instalado."
