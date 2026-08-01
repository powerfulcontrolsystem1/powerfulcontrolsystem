# P109-002 - Confirmación, idempotencia y evals CxP/IA

Fecha: 2026-08-01

Alcance: implementación y pruebas locales del candidato posterior a
`e7ac38ee`; todavía no desplegado.

## Controles implementados

- confirmación visible antes de aprobar, rechazar o contabilizar;
- guardia inmediata del cliente contra doble acción;
- advisory lock PostgreSQL por soporte y namespace propio, conservado en la
  misma conexión durante la llamada IA;
- el intento concurrente falla antes de reservar cuota o llamar al proveedor;
- aprobación/rechazo con `FOR UPDATE`, evento en la misma transacción e
  idempotencia cuando se repite el mismo estado;
- estados rechazado, duplicado y contabilizado no pueden revivirse;
- aprobación exige proveedor activo del mismo `empresa_id`, documento y total
  positivo;
- la extracción no sobrescribe una decisión humana concurrente;
- respuesta pública segura ante proveedor caído, limitado o inválido;
- evaluación cerrada de JSON inválido, campos faltantes, confianza baja y
  descuadre entre subtotal, impuestos, retenciones y total.

## Pruebas

```text
go test ./db ./handlers -run "SoporteComprasIA|SoportesComprasIA" -count=1
go test ./...
node --check web/js/soportes_compras_ia.js
node tools/security_audit.mjs
```

Resultado: PASS. La carrera local no se ejecutó porque el runtime Windows tiene
`CGO_ENABLED=0`; el CI Linux obligatorio conserva esa compuerta.

## Límite

P109-002 permanece **parcial**. El siguiente candidato debe demostrar en
staging doble solicitud concurrente, una sola cuota/consulta, aprobación y
cancelación visibles, contabilización única, proveedor degradado y Centro IA.
El aislamiento A/B requiere una segunda identidad no global autorizada.
