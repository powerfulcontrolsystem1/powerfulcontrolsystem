# P109-002 - Revision y rechazo real del soporte CxP/IA

Fecha: 2026-08-01

Entorno: staging, PCS (`empresa_id=12`)

Candidato: `89d6e042e24d57cba920439521704acaacf7bd00`

La bandeja autenticada mostro simultaneamente los soportes `SCI-0001` y
`SCI-0002`, incluido el segundo en estado `extraido`. Esto confirma en runtime
la correccion de la lista sin filtro de estado.

Sobre `SCI-0002`, creado previamente con una imagen marcada como documento de
prueba no fiscal, se ejecuto el flujo oficial visible:

- se edito el numero a `P109-IA-EDITADO-8847`;
- se edito el total a `1191`;
- se guardo la revision humana y aparecio el evento `editar_revision`;
- se pulso `Rechazar` y el estado visible cambio a `Rechazado`;
- el evento y los datos editados quedaron persistidos por empresa.

La fila final conserva `convertido_tipo` vacio y `convertido_id=0`. La consulta
de cartera dio cero coincidencias para el documento y produccion tambien dio
cero coincidencias. No se creo CxP, pago ni asiento contable.

El modulo se reviso visualmente en escritorio y movil. La tabla mantuvo filas,
columnas, importes, estados y controles legibles, sin hallazgos automaticos.

Estado: **P109-002 parcial**. Quedan aprobacion/confirmacion controlada, doble
clic e idempotencia, degradacion del proveedor, evals, Centro IA y aislamiento
A/B con una identidad no global.
