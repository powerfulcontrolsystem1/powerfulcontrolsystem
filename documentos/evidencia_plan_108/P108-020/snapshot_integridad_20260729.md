# P108-020 - Integridad de snapshot VPS

Fecha: 2026-07-29  
VPS: entorno autorizado  
Snapshot: `20260729_070539`

## Resultado

El script operativo actualizado generó un snapshot de 296 MB e incluyó los
artefactos obligatorios:

- `postgres_all.sql.gz`;
- `pcs_web_uploads`;
- `pcs_downloads`;
- `pcs_backups`;
- `pcs_private_storage`.

Se validó `gzip -t` para el dump de PostgreSQL y `tar -tzf` para cada volumen.
La corrección elimina la omisión previa de almacenamiento privado.

## Límite

Esto no sustituye un restore drill completo con aplicación, autenticación,
documentos privados, CxP y rollback. P108-020 permanece **parcial**.

## Control del runner diario 2026-07-30

El ejecutable instalado del cron estaba desactualizado respecto al script del
repositorio y aceptó como completo un snapshot sin artefactos críticos. Se
reinstaló el cron/runner, se verificó igualdad SHA-256 y el snapshot siguiente
incluyó PostgreSQL, uploads, descargas, backups, almacenamiento privado, Mailu
y los tres volúmenes obligatorios de OnlyOffice.
