# Reporte de backup y restauracion

Estado: **PENDIENTE DE EJECUCION EFIMERA**.

No se creó, descargó ni restauró un backup real. Antes de produccion se exige
una restauracion completa de PostgreSQL, archivos privados, configuracion
cifrada y datos de integraciones en una red aislada. Deben medirse RPO, RTO,
integridad de conteos por empresa y capacidad de iniciar la aplicacion sin
credenciales productivas.

El procedimiento operativo se mantiene en
`documentos/gobernanza_tecnica/runbooks/runbook_recuperacion_desastre_docker_vps.md`.

## Evidencia operativa 2026-09-05

- Se midió el VPS antes de limpiar: 55 GB usados de 96 GB (58%) y 42 GB libres.
- Se creó fuera del VPS un dump lógico PostgreSQL comprimido de 103.474.315
  bytes. Se verificaron el SHA-256 remoto/local y la cabecera del dump; no se
  registran aquí rutas privadas ni el hash.
- Se eliminaron 17 árboles de candidatos staging sin referencias de
  contenedores. Se conservó el único candidato staging activo y todos los
  volúmenes conectados.
- Después de la limpieza quedaron 35 GB usados (36%) y 62 GB libres.
- Se instaló `rclone` en el host. No hay remoto OAuth configurado ni cron externo
  activo todavía; falta elegir y autorizar la cuenta de nube y ejecutar una
  restauración integral aislada. Por ello el estado global de restauración
  completa continúa pendiente.
- Se ejecutó además el alcance `vps`: 1.066.063.602 bytes en 24 archivos con
  dump PCS, dump Nextcloud, proyecto filtrado y volúmenes activos permitidos.
  `SHA256SUMS` aprobó en el VPS y nuevamente después de descargar el conjunto a
  un disco local externo al servidor. Esta verificación de integridad no
  sustituye el restore drill pendiente.
- La revisión de salud encontró al worker en `not_ready` por conflictos de
  idempotencia del scheduler: la clave era estable por intervalo pero el payload
  incluía la hora de cada intento. Se corrigió localmente haciendo estable
  también el payload por intervalo; la corrección requiere el flujo normal de
  revisión y release antes de observarse en producción.
