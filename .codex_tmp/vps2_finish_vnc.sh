#!/usr/bin/env bash
set -euo pipefail

echo "[1/5] Configurando clave VNC valida para admin1"
install -d -m 700 -o admin1 -g admin1 /home/admin1/.vnc
su - admin1 -c "printf 'admin1\n' | vncpasswd -f > /home/admin1/.vnc/passwd"
chmod 600 /home/admin1/.vnc/passwd
chown admin1:admin1 /home/admin1/.vnc/passwd

echo "[2/5] Configurando xstartup XFCE"
cat > /home/admin1/.vnc/xstartup <<'EOF'
#!/bin/sh
unset SESSION_MANAGER
unset DBUS_SESSION_BUS_ADDRESS
exec startxfce4
EOF
chmod +x /home/admin1/.vnc/xstartup
chown admin1:admin1 /home/admin1/.vnc/xstartup

echo "[3/5] Creando servicio systemd"
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

echo "[4/5] Habilitando y arrancando VNC"
systemctl daemon-reload
systemctl enable --now vncserver@1.service

echo "[5/5] Estado"
systemctl is-enabled vncserver@1.service
systemctl is-active vncserver@1.service
ss -ltnp | grep -E ':22|:5901' || true
