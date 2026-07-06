#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

echo "[1/6] Reparando paquetes pendientes"
dpkg --configure -a
apt-get install -f -y

echo "[2/6] Instalando dependencias base y Docker disponible en Ubuntu"
apt-get update || true
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
  docker-compose-v2 \
  docker-buildx

echo "[3/6] Drivers recomendados disponibles"
ubuntu-drivers devices || true
ubuntu-drivers autoinstall || true

echo "[4/6] Habilitando SSH y Docker al inicio"
systemctl enable --now ssh
systemctl enable --now docker
usermod -aG docker admin1

echo "[5/6] Evitando suspension/hibernacion"
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

echo "[6/6] Resumen"
systemctl is-enabled ssh
systemctl is-active ssh
systemctl is-enabled docker
systemctl is-active docker
systemctl is-enabled sleep.target suspend.target hibernate.target hybrid-sleep.target || true
docker --version
docker compose version || true
df -hT / /srv/data
