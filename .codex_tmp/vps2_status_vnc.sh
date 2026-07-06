#!/usr/bin/env bash
set -euo pipefail

systemctl status vncserver@1.service --no-pager || true
echo "---PORT---"
ss -ltnp | grep 5901 || true
echo "---PROC---"
ps aux | egrep 'Xtigervnc|vncserver' | grep -v grep || true
echo "---JOURNAL---"
journalctl -u vncserver@1.service --no-pager -n 100 || true
