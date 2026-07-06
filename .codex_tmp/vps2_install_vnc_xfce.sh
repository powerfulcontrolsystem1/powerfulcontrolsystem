#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

echo "[1/7] Instalando escritorio liviano XFCE y TigerVNC"
apt-get update
apt-get install -y \
  xfce4 \
  xfce4-goodies \
  dbus-x11 \
  tigervnc-standalone-server \
  tigervnc-common \
  xfonts-base

echo "[2/7] Configurando password VNC para admin1"
install -d -m 700 -o admin1 -g admin1 /home/admin1/.vnc
printf 'admin\nadmin\nn\n' | su - admin1 -c 'vncpasswd' >/tmp/vncpasswd.log 2>&1 || {
  cat /tmp/vncpasswd.log
  exit 1
}

echo "[3/7] Configurando arranque XFCE en VNC"
cat > /home/admin1/.vnc/xstartup <<'EOF'
#!/bin/sh
unset SESSION_MANAGER
unset DBUS_SESSION_BUS_ADDRESS
exec startxfce4
EOF
chmod +x /home/admin1/.vnc/xstartup
chown admin1:admin1 /home/admin1/.vnc/xstartup

echo "[4/7] Creando servicio systemd vncserver@:1"
cat > /etc/systemd/system/vncserver@.service <<'EOF'
[Unit]
Description=TigerVNC Server for display :%i
After=network.target

[Service]
Type=forking
User=admin1
Group=admin1
WorkingDirectory=/home/admin1
PIDFile=/home/admin1/.vnc/%H:%i.pid
ExecStartPre=-/usr/bin/vncserver -kill :%i
ExecStart=/usr/bin/vncserver :%i -localhost no -geometry 1280x800 -depth 24
ExecStop=/usr/bin/vncserver -kill :%i
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

echo "[5/7] Habilitando VNC al inicio"
systemctl daemon-reload
systemctl enable --now vncserver@1.service

echo "[6/7] Estado VNC"
systemctl status vncserver@1.service --no-pager || true
ss -ltnp | grep -E ':5901|:22' || true

echo "[7/7] Listo: VNC en puerto 5901"
