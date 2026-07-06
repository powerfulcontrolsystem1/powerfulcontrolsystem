#!/usr/bin/env bash
set -euo pipefail

sed -i 's|^PIDFile=.*|PIDFile=/home/admin1/.config/tigervnc/%H:%i.pid|' /etc/systemd/system/vncserver@.service
systemctl daemon-reload
systemctl restart vncserver@1.service
sleep 3
systemctl is-enabled vncserver@1.service
systemctl is-active vncserver@1.service
ss -ltnp | grep 5901 || true
