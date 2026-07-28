# P108-002 - migraciones y runtime sin DDL

Fecha: 2026-07-28  
Ambiente: staging aislado  
SHA: `e65f6dcddca0733f85d95cc0ae07ef33ef35e7c3`

## Evidencia aprobada parcial

- El candidato se desplegó en el Compose aislado de staging.
- PostgreSQL, API y worker quedaron saludables.
- `/health` respondió `ok` y `/ready` respondió `ready`.
- `pcs-migrate` ejecutó el catálogo inmutable con
  `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0`.
- Segunda pasada: `applied=0`, con 8 migraciones existentes de empresas y 7 de
  superadministrador; no intentó reactivar el bootstrap histórico.
- El worker respondió su `/ready` interno.

## Pendiente para aprobar

- repetir desde una base vacía;
- actualizar desde un snapshot representativo y comparar el esquema;
- demostrar mediante roles PostgreSQL que API y worker carecen de DDL;
- ensayar migración fallida y rollback compatible;
- repetir sobre las imágenes exactas promovidas por digest.

Estado: **parcial; no certifica P108-002**.
