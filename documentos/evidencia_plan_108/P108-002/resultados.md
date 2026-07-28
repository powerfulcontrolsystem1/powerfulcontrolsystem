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
3. El primer digest corregido avanzó hasta el cifrado de configuración y
   detectó que faltaban las raíces `configuraciones` y `tipos_de_empresas`.

La rama candidata agrega raíces mínimas, idempotentes y exclusivas del rol
migrador para `empresas`, `administradores`, `sesiones`, `configuraciones` y
`tipos_de_empresas`, junto con
`deploy/scripts/vps-p108-empty-migration-drill.sh`.

## Resultado final del ensayo vacío

Commit: `64d89b6f02ffb0d63d600e44ba320bcaffd09d22`

Migrador:
`ghcr.io/powerfulcontrolsystem1/pcs-migrate@sha256:bcc26035d207c6112e2e543f46fe638b2141cf3ff4660a59aea6b3daea93cb37`

- Workflow `30338397992`: construcción, Trivy, SBOM y publicación aprobadas.
- Primera pasada: empresas `applied=8`, superadministrador `applied=7`.
- Segunda pasada: empresas `applied=0 existing=8`,
  superadministrador `applied=0 existing=7`.
- Catálogo heredado: 91 pasos empresariales y 31 administrativos.
- Resultado: 336 tablas empresariales y 48 administrativas.
- Limpieza: no quedaron contenedores, redes ni volúmenes con el ID del ensayo.

Subcriterios base vacía y segunda pasada: **PASS**.

P108-002 continúa parcial solamente por upgrade desde snapshot representativo,
restricción DDL mediante roles PostgreSQL y simulación de fallo/rollback.

## Upgrade desde snapshot representativo

`deploy/scripts/vps-p108-upgrade-migration-drill.sh` tomó dos copias lógicas
consistentes del PostgreSQL de staging y las restauró en un PostgreSQL 16
efímero. El origen se validó mediante el prefijo `pcs-staging-` y nunca recibió
escrituras.

- Digest migrador: `bcc26035d207...`.
- Primera pasada: empresas `applied=0 existing=8`; superadministrador
  `applied=0 existing=7`.
- Segunda pasada: los mismos resultados idempotentes.
- Tablas empresariales: `349` antes y `349` después.
- Tablas administrativas: `59` antes y `59` después.
- Limpieza: cero recursos efímeros residuales.

Subcriterio upgrade representativo: **PASS**.

P108-002 continúa parcial solamente por restricción DDL mediante roles
PostgreSQL y simulación de fallo/rollback.

## Hallazgo de privilegios runtime

La consulta de solo lectura ejecutada desde `pcs-staging-worker` mostró el mismo
resultado en ambas bases:

```text
current_user=pcs_staging
rolsuper=true
rolcreatedb=true
rolcreaterole=true
public.CREATE=true
```

Estado: **P0 reproducido**. API y worker no deben usar al propietario migrador.

La corrección candidata incorpora `PCS_RUNTIME_DB_USER` y
`PCS_RUNTIME_DB_PASSWORD`, provisiona/rota ese login desde `pcs-migrate`, le
concede únicamente DML, secuencias y ejecución de funciones, revoca `CREATE`
en `public` y verifica que no conserve superusuario, creación de bases/roles o
`BYPASSRLS`. Backend y worker forman sus DSN exclusivamente con ese usuario.

Falta publicar el digest, configurar el secreto privado de staging y demostrar
los privilegios negativos antes de aprobar este subcriterio.

### Primer despliegue y rollback

El primer digest del rol separado falló cerrado en `pcs-migrate` antes de
arrancar API/worker porque PostgreSQL no podía inferir el tipo de `$2` dentro
de `format()`. Se restauró el digest anterior sin reconstrucción y staging
volvió a `/health=ok` y `/ready=ready`.

La corrección tipa explícitamente `$1::text` y `$2::text`; un contrato impide
retirar esos casts. Este ejercicio aporta evidencia real de rollback técnico,
pero falta repetir el digest corregido.

## Repetición con runtime restringido y corrección de solicitudes

SHA candidato: `7ca5fb1be10d1f02fe3e0a7c5009f559c9d6f853`.

- Workflow inmutable `30378926932`: build, Trivy, SBOM, publicación de los
  cuatro digests y validación de Compose: **PASS**.
- Staging aislado se actualizó exclusivamente por digest. API y worker quedaron
  saludables; `/health` respondió `ok` y `/ready` respondió `ready`.
- El migrador aplicó una migración administrativa adicional y terminó con
  `empresas applied=0 existing=8`, `superadministrador applied=1 existing=7` y
  la verificación explícita de que `pcs_staging_runtime` no posee privilegios
  DDL.
- La tabla `portal_visitas_paises` existe bajo el propietario migrador. El
  handler ya no intenta crearla durante una solicitud HTTP.
- En una sesión autenticada de la empresa de prueba se validó visualmente
  `configuracion.html?empresa_id=12` y
  `carrito_de_compras.html?empresa_id=12`: la configuración cargó sin errores
  de consola y la tabla de carritos mostró tres filas con acciones operativas.
- Los logs posteriores no contienen `permission denied` ni `SQLSTATE 42501`.

La ruta pública de visitas devolvió `401` al probarla directamente en staging;
esa respuesta ocurre antes del handler por la barrera de autenticación del
entorno y queda pendiente de una prueba de portal público separada. No afecta la
migración ni se usa como evidencia de aprobación del endpoint público.

Estado: **P108-002 parcial**. Se aprobaron la base vacía, upgrade,
idempotencia, runtime sin DDL y rollback técnico; falta completar la matriz de
rollback de datos y las demás puertas del Plan 108.
