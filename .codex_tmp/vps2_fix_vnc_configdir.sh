#!/usr/bin/env bash
set -euo pipefail

systemctl stop vncserver@1.service || true
install -d -m 700 -o admin1 -g admin1 /home/admin1/.config/tigervnc
cp /home/admin1/.vnc/passwd /home/admin1/.config/tigervnc/passwd
cp /home/admin1/.vnc/xstartup /home/admin1/.config/tigervnc/xstartup
chown -R admin1:admin1 /home/admin1/.config
chmod 600 /home/admin1/.config/tigervnc/passwd
chmod +x /home/admin1/.config/tigervnc/xstartup
systemctl reset-failed vncserver@1.service || true
systemctl start vncserver@1.service
sleep 2
systemctl status vncserver@1.service --no-pager || true
ss -ltnp | grep -E ':5901|:22' || true
