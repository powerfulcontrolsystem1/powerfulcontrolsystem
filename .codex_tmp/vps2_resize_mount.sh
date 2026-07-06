#!/usr/bin/env bash
set -euo pipefail

echo "[1/6] Estado inicial"
vgs
lvs -a -o lv_name,vg_name,lv_size,lv_path
df -hT || true

echo "[2/6] Limpiando entradas invalidas de /etc/fstab para /srv/data"
cp /etc/fstab "/etc/fstab.bak.$(date +%Y%m%d%H%M%S)"
awk '$2 != "/srv/data" && $0 != "UUID=" { print }' /etc/fstab > /tmp/fstab.clean
cat /tmp/fstab.clean > /etc/fstab

echo "[3/6] Retirando volumen de datos temporal si existe"
if mountpoint -q /srv/data; then
  umount /srv/data
fi
if lvs /dev/ubuntu-vg/data-lv >/dev/null 2>&1; then
  lvremove -y /dev/ubuntu-vg/data-lv
fi

echo "[4/6] Extendiendo particion de sistema a 300 GB"
lvextend -L 300G /dev/ubuntu-vg/ubuntu-lv
resize2fs /dev/ubuntu-vg/ubuntu-lv

echo "[5/6] Creando volumen de datos con el resto principal del disco"
lvcreate -L 1450G -n data-lv ubuntu-vg
mkfs.ext4 -F /dev/ubuntu-vg/data-lv
mkdir -p /srv/data
uuid="$(blkid -s UUID -o value /dev/ubuntu-vg/data-lv)"
printf 'UUID=%s /srv/data ext4 defaults,noatime 0 2\n' "$uuid" >> /etc/fstab
mount /srv/data
mkdir -p /srv/data/backups /srv/data/nextcloud /srv/data/pcs
chown -R admin1:admin1 /srv/data

echo "[6/6] Estado final"
findmnt / /boot /srv/data
df -hT / /srv/data
vgs
lvs -a -o lv_name,vg_name,lv_size,lv_path
