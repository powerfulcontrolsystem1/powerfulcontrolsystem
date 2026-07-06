#!/usr/bin/env bash
set -euo pipefail

root_dev="/dev/disk/by-id/dm-uuid-LVM-lPvYUGNNve3ZVC1c1eS3cm9TBInuwIUtUxPfOuKPKUMaG8DUFiT9ETslJxEXENhD"
boot_uuid="b063dcd3-d3ea-4d6c-8807-fbc0dcef8b40"
data_uuid="$(blkid -s UUID -o value /dev/ubuntu-vg/data-lv)"

cp /etc/fstab "/etc/fstab.repair.$(date +%Y%m%d%H%M%S)"
cat > /etc/fstab <<EOF
# /etc/fstab: static file system information.
#
# <file system> <mount point> <type> <options> <dump> <pass>
$root_dev / ext4 defaults 0 1
/dev/disk/by-uuid/$boot_uuid /boot ext4 defaults 0 1
/swap.img none swap sw 0 0
UUID=$data_uuid /srv/data ext4 defaults,noatime 0 2
EOF

mkdir -p /srv/data
mountpoint -q /srv/data || mount /srv/data
findmnt --verify --verbose
df -hT / /boot /srv/data
nl -ba /etc/fstab
