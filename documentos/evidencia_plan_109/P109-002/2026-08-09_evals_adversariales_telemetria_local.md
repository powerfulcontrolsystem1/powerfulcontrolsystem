# P109-002 - Evals adversariales y telemetria de extracción IA

Fecha: 2026-08-09
Entorno: candidato local, sin proveedor ni datos PCS
Rama: `codex/p109-batch-no-pr`

## Contrato endurecido

- La respuesta se limita a 128 KiB y a las claves declaradas en el prompt.
- Campos escalares no aceptan objetos/arreglos; `lineas_detectadas` debe ser una
  lista acotada a 500 elementos.
- Textos e importes tienen límites; se rechazan `NaN`, infinito, negativos,
  confianza fuera de 0..1 y valores monetarios mayores a 1e15.
- Las fechas inválidas ya no se recortan como si fueran válidas.
- Falta de fecha, tipo documental, moneda COP, proveedor/documento, baja
  confianza o descuadre contable obliga revisión humana.
- El resultado del modelo nunca puede desactivar una revisión que ya solicitó.

## Operación observable

Se agregaron contadores atomicos para `consistent`, `human_review`,
`provider_error`, `invalid_response`, `canceled` y `persistence_error`. Son
agregados por proceso y solo usan la etiqueta acotada `result`; no contienen
empresa, usuario, documento, archivo, prompt ni respuesta.

Prometheus alerta fallos de proveedor, contrato y persistencia. Grafana muestra
los resultados recientes por job. Cancelar sigue siendo un resultado esperado y
no genera alerta.

## Pruebas

- JSON válido cercado y extracción consistente: PASS.
- JSON inexistente, campo inesperado, tipo compuesto, lista inválida, `NaN`,
  negativo, confianza >1, fecha inválida y respuesta sobredimensionada: PASS
  como rechazo o revisión humana.
- 64 registros simultáneos del contador consistente: PASS sin pérdida.
- Go completo, vet y preflight Full/Strict: PASS.

## Límite

No se llamó al proveedor, no se cargó archivo y no se mutó PCS. Falta repetir
estos casos en el candidato desplegado, degradación real y aislamiento A/B.
P109-002/P109-010 siguen parciales; Plan 109 permanece 53,3 % y NO-GO.
