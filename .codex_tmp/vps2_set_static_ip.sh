#!/usr/bin/env bash
set -euo pipefail

cp -a /etc/netplan "/etc/netplan.bak.$(date +%Y%m%d%H%M%S)"
cat > /etc/netplan/00-installer-config.yaml <<'EOF'
network:
  version: 2
  ethernets:
    enp0s25:
      match:
        macaddress: 00:21:70:6c:b0:f3
      set-name: enp0s25
      dhcp4: false
      dhcp6: false
      addresses:
        - 192.168.1.188/24
      routes:
        - to: default
          via: 192.168.1.1
      nameservers:
        addresses:
          - 200.21.200.10
          - 200.21.200.80
          - 1.1.1.1
EOF

chmod 600 /etc/netplan/00-installer-config.yaml
netplan generate
netplan apply
