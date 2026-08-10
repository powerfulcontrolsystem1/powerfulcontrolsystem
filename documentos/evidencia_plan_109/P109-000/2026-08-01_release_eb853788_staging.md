# P109-000 - Release inmutable y promoción controlada

Fecha: 2026-08-01

Candidato: `eb8537883f88ba654f23b0b87b675da7191eef2a`

Entorno modificado: staging aislado. Producción no fue desplegada.

## Integración y cadena de suministro

- La PR 114 fue aprobada, superó los dos checks obligatorios y se fusionó.
- El workflow de release `30686058740` finalizó correctamente sobre el SHA
  completo y publicó API, migrador, worker y frontend una sola vez.
- El artefacto `immutable-release-candidate-eb853788...` contiene cuatro SBOM
  CycloneDX y los cuatro informes Trivy. Todos reportaron cero vulnerabilidades.
- Digests promovidos:
  - API: `sha256:d871064450283d85653a233070ba8083924a534daab81d59f835cd44b028999d`.
  - migrador: `sha256:1b26dffbde640559924c8fe415b62e1571d3e4101feb852d5801cdb038c9a0e7`.
  - worker: `sha256:b8e0ca57e088584d63870eea3484108d26d9594025c92dcdab96d7cc66c1e37e`.
  - frontend: `sha256:d12d86dee75f4c4e928295779f2f1a5566ba2fe60b4cd75a9919a82b436455e3`.

## Promoción y salud

`deploy/scripts/vps-staging-digest-up.sh` resolvió los cuatro digests exactos,
ejecutó `up --no-build` y dejó backend, worker y frontend de staging en estado
`healthy`. Las comprobaciones externas devolvieron `200` con `status=ok` en
`/health` y `200` con `status=ready` en `/ready`.

Antes y después de la promoción, los contenedores de producción conservaron sus
imágenes previas. Producción respondió `200` en `/health` y `/ready`; no se
ejecutó `rs`, migración, reinicio ni cambio de configuración en producción.

## Migraciones del mismo digest

El migrador exacto se ensayó en recursos Docker efímeros:

- base vacía: 337 tablas empresariales, 49 administrativas, 18 y 10 entradas
  de migración respectivamente;
- segunda pasada: cero migraciones duplicadas;
- checksum alterado: fallo cerrado, esquema en 337 tablas y ledger en 18 sin
  cambios, auditoría de fallo presente y recuperación posterior correcta;
- copia lógica de staging: 350 tablas empresariales y 59 administrativas antes
  y después del upgrade y de la segunda pasada;
- limpieza: cero contenedores, volúmenes o redes residuales de los ensayos.

## Estado

P109-000: **aprobada** para el candidato `eb853788...`. La aprobación no implica
GO de producción: las demás compuertas P0 del Plan 109 continúan vigentes.
