# P109-007 - Persistencia segura de acuse síncrono DIAN

Fecha: 2026-08-09  
Alcance: candidato aislado del Plan 109; sin despliegue ni cambio de datos de
producción.

## Causa confirmada

La emisión fiscal de PCS usa `SendBillSync` en producción. Ese contrato entrega
el acuse final dentro de la respuesta síncrona y no requiere ni devuelve
`TrackId`/`ZipKey`. El historial existente descartaba la respuesta por exigir un
TrackId, por lo que la auditoría visual parecía incompleta pese a existir acuse
de DIAN.

## Corrección implementada

- Cuando `SendBillSync` responde aceptado y no trae TrackId, el backend guarda
  el acuse usando una clave interna `sync:<codigo-documento>` exclusivamente
  para continuidad de auditoría.
- La interfaz lo etiqueta como `Acuse sincrónico (sin TrackId)` y deshabilita
  `Reconsultar`.
- El backend rechaza explícitamente una clave sintética en `GetStatusZip`; no
  la presenta como identificador entregado por DIAN ni la transmite al servicio
  externo.
- Para `SendBillAsync` y `SendTestSetAsync` se conserva el TrackId real y la
  reconsulta `GetStatusZip` existente.

## Verificación

- Suite focalizada de handlers DIAN/SOAP: PASS.
- `git diff --check`: PASS.

## Límite

La corrección debe promoverse al mismo digest del candidato y probarse con una
nueva operación autorizada antes de cerrar P109-007. Una operación asíncrona
sigue requiriendo `GetStatusZip StatusCode=00`; una síncrona queda evidenciada
por su acuse oficial persistido, no por un TrackId inventado.
