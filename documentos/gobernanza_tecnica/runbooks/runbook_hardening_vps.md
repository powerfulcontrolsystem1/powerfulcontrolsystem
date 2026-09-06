# Runbook de hardening VPS

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- La regla efectiva usa el puerto SSH del inventario autorizado. Abrir 22 no es una reparación por defecto; verificar consola y mantener acceso alternativo durante cambios.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Controles base

- SSH: `PermitRootLogin no` o `prohibit-password` y `PasswordAuthentication no` cuando las llaves esten instaladas.
- SSH productivo: el VPS usa el puerto `49222`; el puerto `22` debe permanecer cerrado salvo recuperacion temporal desde consola del proveedor.
- Firewall: permitir el puerto SSH autorizado y 80/443 donde corresponda; servicios internos ligados a loopback o red privada.
- Fail2ban: habilitado para SSH y Nginx si aplica.
- Docker: revisar `docker ps`, redes internas y volumenes antes de limpiar.
- Secretos: no versionar `.env.platform`, `.env.staging`, claves de Grafana ni credenciales de backup externo.

## Verificacion de acceso SSH

Si rs falla por SSH, comprobar primero que usa el puerto del inventario autorizado. Desde consola del proveedor verificar listener, servicio ssh y firewall del mismo puerto. No abrir 22 por herencia de un comando antiguo.

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

## Fuentes y aceptación de la revisión

[vps-hardening-audit.sh](../../../deploy/scripts/vps-hardening-audit.sh), [manual_vps_seguridad.md](../../manual_vps_seguridad.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../requisitos/especificacion_y_trazabilidad.md)).
