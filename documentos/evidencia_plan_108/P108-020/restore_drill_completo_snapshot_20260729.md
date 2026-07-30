# P108-020 - Restore drill aislado de snapshot operativo

Fecha: 2026-07-29  
Ambiente: VPS, contenedor temporal aislado  
Snapshot: `20260729_082234`  
Imagen de restauración: `postgres:16.14-alpine`

## Integridad del snapshot

El snapshot generado después de alinear el contrato de backup y restore validó
correctamente:

- dump `postgres_all.sql.gz`;
- uploads, descargas, logs y backups;
- almacenamiento privado;
- certificados Mailu;
- datos, librerías y logs de OnlyOffice.

Los volúmenes opcionales de Let's Encrypt/Certbot no están configurados en el
VPS y se registraron como opcionales, no como archivos faltantes.

## Restore drill

Se restauró `postgres_all.sql.gz` dentro de un contenedor PostgreSQL temporal.
La consulta de comprobación finalizó correctamente y el contenedor temporal fue
eliminado al terminar.

| Métrica | Resultado |
| --- | ---: |
| Tamaño del snapshot | 295 MB |
| RTO observado | 23 s |
| RPO observado | 79 s |
| Tarballs obligatorios válidos | 9 / 9 |
| Contenedor temporal residual | No |

## Límite

P108-020 permanece **parcial**. Este drill demuestra restaurabilidad de base y
archivos empaquetados, pero falta restaurar la aplicación completa en un entorno
aislado, verificar login, archivos privados, CxP, contabilidad e IA, y ensayar
rollback compatible de aplicación/base.

## Repetición y corrección del runner 2026-07-30

La validación detectó que el snapshot `20260730_031501` había sido generado por
una copia antigua del runner diario y no contenía almacenamiento privado,
Mailu ni OnlyOffice. Ese snapshot se conserva para trazabilidad, pero queda
**rechazado** para recuperación.

Se reinstaló `/usr/local/bin/pcs_vps_backup_daily.sh` desde el script versionado
y se comprobó igualdad SHA-256. El nuevo snapshot `20260730_194951` pesa 302 MB
y contiene los nueve artefactos obligatorios.

| Métrica | Resultado |
| --- | ---: |
| Validación de dump/tarballs | PASS |
| Restore PostgreSQL temporal | PASS |
| RTO observado | 24 s |
| RPO observado | 107 s |
| Contenedores residuales | 0 |
| Salud staging posterior | `ok` / `ready` |
| Disco VPS posterior | 54 % usado |

El runner diario queda corregido. P108-020 continúa parcial solamente respecto
a restauración funcional de aplicación, archivos privados, CxP, contabilidad,
IA y rollback coordinado de aplicación/base.
