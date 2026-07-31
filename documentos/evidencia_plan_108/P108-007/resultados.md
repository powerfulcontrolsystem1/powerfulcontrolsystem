# P108-007 - consistencia financiera y contable

Fecha: 2026-07-30  
Ambiente: staging aislado  
Empresa autorizada: Powerful Control System (`empresa_id=12`)  
Candidato observado: `f7214329ed70b15085f300d823244617b9cb998f`

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

El candidato final demostró en PCS un pago CxP controlado de `0,01`, outbox
`published`, job `completed`, evento procesado y asiento con débito/crédito
`0,01` y diferencia `0`. Pago, evento y asiento conservan una sola fila por su
identidad natural. La interfaz final muestra también la unidad menor sin
redondearla visualmente a cero.

P108-007 permanece **parcial / NO-GO**: el caso CxP queda aprobado, pero aún
faltan la matriz completa de ventas, devoluciones, impuestos, retenciones,
anulaciones y monedas, además de la recuperación controlada de los dos eventos
históricos `dead`. No se hará una reactivación global sin inventario y
previsualización por empresa.
