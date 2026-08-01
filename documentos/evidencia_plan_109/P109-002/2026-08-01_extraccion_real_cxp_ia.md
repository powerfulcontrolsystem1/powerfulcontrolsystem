# P109-002 - Extracción real y edición CxP/IA

Fecha: 2026-08-01

Entorno: staging, PCS (`empresa_id=12`)

Candidato: `8847288bedb82abba7ab505391c11395c9efbc02`

Se generó una factura visual no fiscal, rotulada como prueba P109, y se cargó
desde el botón `Cargar factura o recibo con IA` de Finanzas. El flujo oficial
respondió 200 al radicar y 200 al extraer con el proveedor IA configurado.

La extracción recuperó correctamente:

- documento `P109-IA-8847288B`;
- valor `1190`;
- emisión `2026-08-01` y vencimiento `2026-08-31`;
- proveedor sugerido y borrador de CxP sujeto a revisión humana.

En la pantalla real se editaron el documento a `P109-IA-EDITADO-8847` y el
valor a `1191`; ambos valores permanecieron editables y visibles. No se envió
el formulario: no se creó CxP, pago ni asiento contable. El soporte de prueba
quedó radicado con ID 2 para trazabilidad.

La revisión posterior encontró que `action=soportes` sin parámetro `estado`
normalizaba el vacío como `radicado`; por eso el soporte `extraido` no aparecía
en la lista aunque sí existía en la misma base de staging. Se corrigió la
semántica para que vacío signifique todos los estados activos y se añadió una
regresión focalizada. Esta corrección requiere promoción y repetición antes de
darla por certificada.

Estado: **P109-002 parcial**. Se cerró la extracción externa y la edición del
borrador. Faltan confirmación/cancelación oficial completa, doble clic,
degradación del proveedor, evals, Centro IA y aislamiento con una identidad A/B
no global.
