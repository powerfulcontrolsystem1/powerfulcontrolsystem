#!/usr/bin/env bash
set -euo pipefail

if docker buildx version >/dev/null 2>&1; then
  echo "[OK] Docker Buildx ya esta instalado."
  docker buildx version
  exit 0
fi

if command -v apt-get >/dev/null 2>&1; then
  apt-get update
  if ! apt-get install -y docker-buildx-plugin; then
    echo "[WARN] docker-buildx-plugin no esta disponible por apt; se instalara desde release oficial."
  fi
elif command -v apk >/dev/null 2>&1; then
  apk add --no-cache docker-cli-buildx || true
else
  echo "[WARN] Gestor de paquetes no soportado; se instalara desde release oficial."
fi

if ! docker buildx version >/dev/null 2>&1; then
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) buildx_arch="amd64" ;;
    aarch64|arm64) buildx_arch="arm64" ;;
    *) echo "[ERROR] Arquitectura no soportada para Buildx: $arch" >&2; exit 1 ;;
  esac
  version="$(python3 - <<'PY'
import json, urllib.request
with urllib.request.urlopen("https://api.github.com/repos/docker/buildx/releases/latest", timeout=20) as r:
    print(json.load(r)["tag_name"])
PY
)"
  plugin_dir="/usr/lib/docker/cli-plugins"
  mkdir -p "$plugin_dir"
  curl -fsSL "https://github.com/docker/buildx/releases/download/${version}/buildx-${version}.linux-${buildx_arch}" -o "$plugin_dir/docker-buildx"
  chmod +x "$plugin_dir/docker-buildx"
fi

docker buildx version
echo "[OK] Docker Buildx instalado."
