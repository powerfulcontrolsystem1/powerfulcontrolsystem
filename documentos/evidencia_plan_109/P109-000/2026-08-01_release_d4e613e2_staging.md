# P109-000 - Candidato final d4e613e2 promovido por digest

Fecha: 2026-08-01  
Candidato: `d4e613e2d2c6b34451838bcc18a5db5f7afcf2d0`  
Entorno modificado: staging aislado. Produccion no fue desplegada.

## Integracion y cadena de suministro

La PR 117 fue aprobada, supero Preflight, dependencias/SBOM y Trivy, y se
fusiono mediante el flujo protegido. El release inmutable `30690869278`
construyo una sola vez el SHA completo, genero cuatro SBOM CycloneDX y obtuvo
cero vulnerabilidades `HIGH` o `CRITICAL` en las cuatro imagenes.

Digests promovidos:

- API: `sha256:4f9621142631f481f322a3552ba94d932663044d66fb9900781f754ac390096e`.
- migrador: `sha256:657d637ec9f5c3f1ac730c960abaaf0483fa03f093ed3d4dede09476bd9921fa`.
- worker: `sha256:ecb3af978de4b7a8cc627d33ad69a45d6060521f94ee5e5bfdaaac7dd9935846`.
- frontend: `sha256:51cc3ed17f4099615923d65441a1d261db7b60b79aa743b5b35dac321e408113`.

## Migraciones y runtime

El migrador exacto aprobo en recursos Docker efimeros:

- base vacia: 337 tablas empresariales, 49 administrativas y segunda pasada
  idempotente;
- drift de checksum: fallo cerrado, tablas y ledger invariables, intento
  auditado y recuperacion aprobada;
- copia logica de staging: 350 tablas empresariales y 59 administrativas antes
  y despues de dos pasadas.

Los recursos aislados se eliminaron. El mismo arbol aprobo en Linux
`go test -race ./handlers ./db`; la primera preparacion incompleta se descarto
porque no incluia los fixtures `web/`, y la repeticion completa paso ambos
paquetes.

## Promocion y aislamiento

Staging resolvio las cuatro referencias completas por digest. PostgreSQL,
backend, worker y frontend quedaron saludables; `/health` y `/ready`
respondieron 200. Los IDs productivos de backend, worker y frontend fueron
identicos antes y despues de la promocion. No se ejecuto `rs`, migracion ni
reinicio productivo.

Estado: **P109-000 aprobada** para `d4e613e2...`.
