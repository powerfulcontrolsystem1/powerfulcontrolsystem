# Runbook de hardening VPS

## Controles base

- SSH: `PermitRootLogin no` o `prohibit-password` y `PasswordAuthentication no` cuando las llaves esten instaladas.
- Firewall: permitir solo 22, 80, 443 y puertos internos necesarios ligados a `127.0.0.1`.
- Fail2ban: habilitado para SSH y Nginx si aplica.
- Docker: revisar `docker ps`, redes internas y volumenes antes de limpiar.
- Secretos: no versionar `.env.platform`, `.env.staging`, claves de Grafana ni credenciales de backup externo.

## Auditoria rapida

Ejecutar en el VPS:

```bash
bash deploy/scripts/vps-hardening-audit.sh
```

El script no cambia configuraciones; solo informa hallazgos para actuar con seguridad.

## Cadencia

- Antes de abrir servicios nuevos.
- Despues de mover la plataforma a un servidor nuevo.
- Cuando una alerta indique exceso de conexiones, trafico anormal o errores repetidos.
