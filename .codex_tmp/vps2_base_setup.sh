#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

echo "[1/7] Recargando montaje y servicios"
systemctl daemon-reload
mount -a

echo "[2/7] Actualizando Ubuntu y paquetes base"
apt-get update
apt-get -y full-upgrade
apt-get install -y \
  openssh-server \
  ca-certificates \
  curl \
  gnupg \
  tar \
  gzip \
  unzip \
  rsync \
  linux-firmware \
  ubuntu-drivers-common \
  intel-microcode \
  docker.io \
  docker-compose-plugin

echo "[3/7] Aplicando drivers recomendados disponibles"
ubuntu-drivers devices || true
ubuntu-drivers autoinstall || true

echo "[4/7] Habilitando SSH y Docker al inicio"
systemctl enable --now ssh
systemctl enable --now docker
usermod -aG docker admin1

echo "[5/7] Deshabilitando suspension, hibernacion y ahorro agresivo"
systemctl mask sleep.target suspend.target hibernate.target hybrid-sleep.target
mkdir -p /etc/systemd/logind.conf.d
cat > /etc/systemd/logind.conf.d/99-no-sleep.conf <<'EOF'
[Login]
HandleLidSwitch=ignore
HandleLidSwitchExternalPower=ignore
HandleLidSwitchDocked=ignore
IdleAction=ignore
EOF
systemctl restart systemd-logind || true

echo "[6/7] Preparando carpetas de datos"
mkdir -p /srv/data/backups /srv/data/nextcloud /srv/data/pcs
chown -R admin1:admin1 /srv/data

echo "[7/7] Resumen"
lsb_release -a || cat /etc/os-release
uname -a
systemctl is-enabled ssh
systemctl is-active ssh
systemctl is-enabled docker
systemctl is-active docker
df -hT / /srv/data
docker --version
docker compose version || true
