#!/usr/bin/env bash
# Preserve the existing key-only access and port. Run from a verified SSH session.
set -euo pipefail
umask 077
sshd -t
effective="$(sshd -T)"
grep -qx 'passwordauthentication no' <<< "$effective" || {
  echo '[ERROR] primero verificar acceso por llave; no se cambia autenticacion'; exit 1;
}
backup="/var/backups/pcs-security/ssh-$(date -u +%Y%m%dT%H%M%S)"
mkdir -p "$backup"
cp -a /etc/ssh/sshd_config "$backup/sshd_config"
python3 - <<'PY'
from pathlib import Path
p = Path('/etc/ssh/sshd_config')
s = p.read_text()
marker = '# PCS security baseline 2026-09-06\n'
policy = ('X11Forwarding no\nAllowAgentForwarding no\nMaxAuthTries 3\n'
          'LoginGraceTime 30\nMaxStartups 10:30:60\nPermitEmptyPasswords no\n')
# sshd uses the first value; a disconnected or late drop-in is not effective.
if marker not in s:
    p.write_text(marker + policy + s)
PY
if ! sshd -t || ! systemctl reload ssh; then
  cp -a "$backup/sshd_config" /etc/ssh/sshd_config
  sshd -t
  systemctl reload ssh
  echo '[ERROR] cambio revertido'
  exit 1
fi
sshd -T | grep -E '^(port|permitrootlogin|passwordauthentication|x11forwarding|allowagentforwarding|maxauthtries|logingracetime|maxstartups) '
echo "[OK] SSH recargado; respaldo=$backup; comprobar una segunda conexion"
