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

## Aplicación y evidencia del 2026-09-06

[Revisión del VPS](../../seguridad/revision_vps_2026-09-06.md) e
[inventario de imágenes](../../seguridad/revision_vps_2026-09-06_imagenes.json).
`vps-hardening-audit.sh --strict` devuelve error con advertencias y consulta
`sshd -T`; encontrar un archivo no demuestra que sus directivas estén activas.
El verificador Node es documental y nunca certifica el host.

El [aplicador HTTP](../../../deploy/scripts/vps-http-hardening.py) exige
`--main-site` explícito; sin `--apply` solo calcula archivos afectados. Antes de
usarlo confirmar que el host es el borde directo, sin CDN/LB cuyos encabezados
necesiten tratamiento propio. Con `--apply` respalda, valida y recarga o revierte.
Solo el sitio PCS seleccionado recibe límites; los demás sitios saneamiento de
cabeceras. La CSP Webmail del proveedor se conserva.

El [aplicador SSH](../../../deploy/scripts/vps-ssh-hardening.sh) exige acceso
existente por llave y conserva puerto/credenciales. Mantener una sesión abierta,
validar la configuración efectiva y comprobar una segunda conexión tras recargar.
Los respaldos privados quedan bajo `/var/backups/pcs-security/`.

Cambiar cookies no reinicia el presupuesto del chat público: 30 peticiones por IP
cada cinco minutos y 10 por conversación. Se limita a 16.384 contadores por proceso;
para varias réplicas se requiere un contador compartido o protección equivalente.
