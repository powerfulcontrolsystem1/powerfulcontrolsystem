# P109-000 - Candidato 6da6c134 promovido por digest

Fecha: 2026-08-01
Candidato: `6da6c13453a40b2d84e23285fa83f255f34788da`
Entorno modificado: staging aislado. Producción no fue desplegada.

## Integración y cadena de suministro

- La PR 115 fue aprobada, superó Preflight, SBOM/dependencias y Trivy, y se
  fusionó mediante el flujo protegido.
- El release inmutable `30687430898` construyó una sola vez el SHA completo.
- Los cuatro informes Trivy reportaron cero vulnerabilidades y el artefacto
  conserva cuatro SBOM CycloneDX.
- Digests publicados y promovidos:
  - API: `sha256:cd37ca066e38862acc33d9252e719b52c069f7df279727bca03f3bb681a869b7`.
  - migrador: `sha256:c50abd065fd914dbb0d5a3d4bf46dc9bb0ccf1908f0f7de73163d4bcd89e63c6`.
  - worker: `sha256:ff0f27acb045b56f1c9f4c2d11b05c78a8e535cb4a4f5cafa713a5bce858cfeb`.
  - frontend: `sha256:6f43de3eeac015a742710eb83a17875515ce1952ae368b1c0d9d62c7a1775534`.

## Migración y compatibilidad

El digest migrador exacto aprobó en recursos Docker efímeros:

- base vacía: 337 tablas empresariales, 49 administrativas y ledgers de 18 y
  10 filas;
- segunda pasada idempotente;
- checksum alterado: fallo cerrado, esquema y ledger sin cambios, intento
  auditado y recuperación correcta;
- copia lógica de staging: 350 tablas empresariales y 59 administrativas antes
  y después de dos pasadas;
- cero contenedores, volúmenes o redes residuales.

También se levantaron temporalmente los cuatro digests anteriores contra el
esquema actualizado. API, worker y frontend quedaron saludables y `/health` y
`/ready` aprobaron. Después se restauraron los cuatro digests actuales y se
repitió salud. Esto demuestra rollback de aplicación compatible; no sustituye
el rollback de datos pendiente en P109-011.

## Salud y aislamiento de producción

Staging resolvió las referencias completas `repositorio@sha256`, dejó API,
worker y frontend saludables y respondió `200` en salud y readiness. Los IDs de
imagen de producción permanecieron iguales antes y después; producción también
conservó salud y readiness. No se ejecutó `rs`, migración ni reinicio productivo.

Estado: **P109-000 aprobada** para `6da6c134...`.
