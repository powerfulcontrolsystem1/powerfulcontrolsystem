# P108-007 - consistencia financiera y contable

Fecha: 2026-07-30  
Ambiente: staging aislado  
Empresa autorizada: Powerful Control System (`empresa_id=12`)  
Candidato observado: `5566a213c919a8ed1c1c0d2b6a823600c760afb2`

## Conciliación real CxP

La CxP QA `P108-CONC-5EC1-941964` se creó y pagó por el flujo oficial. La
consulta de solo lectura, siempre filtrada por `empresa_id=12`, produjo:

| Invariante | Resultado |
| --- | ---: |
| Valor original / pagado / saldo | `2 / 2 / 0` |
| Estado de cartera | `pagada` |
| Filas de asignación de pago | `2` |
| Suma de asignaciones | `2` |
| Movimientos financieros distintos | `2` |
| Suma de movimientos | `2` |
| Eventos outbox | `2` |
| Estado outbox | `dead, dead` |

La transacción financiera central no perdió ni duplicó dinero. El fallo se
encontró después de la confirmación: el worker no tenía habilitado el topic
`cuentas_por_pagar.pago_registrado`, por lo que ambos eventos agotaron cinco
intentos sin generar el evento contable.

## Corrección implementada

El worker registra el topic y construye un único evento
`finanzas/abono_proveedor_registrado`, que produce débito a proveedores y
crédito a caja/bancos. La función valida que:

- `empresa_id` coincida en pago, cuenta, movimiento y evento;
- los IDs secundarios pertenezcan al mismo tenant;
- payload, pago y movimiento tengan el mismo monto redondeado;
- el pago se bloquee con `FOR UPDATE`;
- un replay encuentre el evento natural por
  `empresa_cxp_pagos/pago_id` y no lo inserte nuevamente.

## Estado

P108-007 permanece **parcial / NO-GO** hasta publicar el nuevo worker,
demostrar en PCS un outbox `published`, job `completed`, evento procesado y
asiento balanceado, y aprobar la recuperación controlada de los dos eventos
históricos `dead`. No se hará una reactivación global sin inventario y
previsualización por empresa.
