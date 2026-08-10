# P109-000 - Candidato final 89d6e042 en staging

Fecha: 2026-08-01

Entorno: staging aislado

Revision: `89d6e042e24d57cba920439521704acaacf7bd00`

La PR 119 quedo aprobada y fusionada. El workflow inmutable `30700866694`
construyo una sola vez el SHA fusionado y publico cuatro referencias por
digest:

- API: `sha256:1fe91f58ba7aba2089b26553b08c4936c2ee0256ee56e4e0f2693995626442ef`.
- migrador: `sha256:38048c4cebf4ee260167846d41af7f11b4e0fb3eb5c510a5283424be3a43a916`.
- worker: `sha256:25259d8ea878c02c5267fefc8541377813efb0b8244864c9aa80de5cb440b509`.
- frontend: `sha256:b37b6e7eaaa2f084fb504d3e93f876b2a04a45dbc3fc507958dcc5f487b055bf`.

Los cuatro reportes Trivy quedaron en cero HIGH/CRITICAL y se generaron
cuatro SBOM CycloneDX. La promocion uso `vps-staging-digest-up.sh`, sin
recompilar. Backend, worker, frontend y PostgreSQL quedaron saludables, sin
reinicios; `/health` y `/ready` respondieron 200.

Produccion conservo exactamente sus imagenes anteriores: backend
`0e1e77ba...`, worker `2b07cee3...` y frontend `bbfac963...`.

El migrador exacto aprobo:

- base vacia: 337 tablas empresariales, 49 administrativas y segunda pasada
  idempotente;
- drift de checksum: fallo cerrado, esquema/ledger invariantes y recuperacion;
- copia logica de staging: 350/59 tablas antes y despues de dos pasadas.

El rollback de aplicacion se ejecuto por digest: staging volvio al candidato
`8847288b...`, aprobo `/health` y `/ready`, y luego recupero los cuatro digests
de `89d6e042...`. El soporte CxP/IA rechazado conservo estado, documento, valor
y `convertido_id=0`; todos los servicios finalizaron saludables sin reinicios.
Produccion mantuvo sus tres imagenes durante todo el ciclo.

Estado: **P109-000 aprobada** para `89d6e042...`. Esto no autoriza produccion
ni cambia el veredicto general NO-GO.
