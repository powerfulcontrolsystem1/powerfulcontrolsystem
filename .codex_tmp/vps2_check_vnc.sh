#!/usr/bin/env bash
set -euo pipefail

echo "---APTLOCKS---"
pgrep -a apt apt-get dpkg || true
echo "---VNC-PKGS---"
dpkg -l | egrep 'xfce4|tigervnc|xfonts-base|dbus-x11' || true
echo "---VNC-SERVICE---"
systemctl status vncserver@1.service --no-pager 2>/dev/null || true
echo "---PORTS---"
ss -ltnp | grep -E ':22|:5901' || true
echo "---FAILED---"
systemctl --failed --no-pager || true
