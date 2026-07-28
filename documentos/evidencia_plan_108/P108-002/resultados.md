# P108-002 - migraciones y runtime sin DDL

Fecha: 2026-07-28  
Ambiente: staging aislado  
SHA inicial: `e65f6dcddca0733f85d95cc0ae07ef33ef35e7c3`

SHA candidato por digest: `fce1655cedff6d3e9235424bfaf0029e80b2ff0c`

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

Estado: **parcial; no certifica P108-002**.

## Repetición sobre el candidato por digest

- `pcs-migrate` usó
  `ghcr.io/powerfulcontrolsystem1/pcs-migrate@sha256:67cb7efbd883eafc6ab5a00e9bcabcf59d7aab680e52cbea0910c34ea825d08d`.
- Terminó con código `0`, `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0`,
  `empresas applied=0 existing=8` y
  `superadministrador applied=0 existing=7`.
- API, worker y frontend arrancaron desde sus digests exactos y
  `/health`/`/ready` quedaron verdes.

La repetición por digest queda aprobada. P108-002 continúa parcial por las
pruebas desde base vacía, upgrade representativo, restricción DDL y rollback.

## Hallazgo de base vacía y corrección candidata

El ensayo aislado creó PostgreSQL 16 sin datos y eliminó automáticamente su
contenedor, red y volumen al terminar.

1. Con `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0`, el migrador falló cerrado porque no
   existía la raíz `empresas`.
2. Con `PCS_RUNTIME_SCHEMA_BOOTSTRAP=1` y rol `migrate`, el bootstrap encontró
   que `administradores` todavía no existía antes de extender su esquema.

La rama candidata agrega raíces mínimas, idempotentes y exclusivas del rol
migrador para `empresas`, `administradores` y `sesiones`, junto con
`deploy/scripts/vps-p108-empty-migration-drill.sh`. Falta construir el nuevo
digest y repetir el ensayo antes de aprobar este subcriterio.
