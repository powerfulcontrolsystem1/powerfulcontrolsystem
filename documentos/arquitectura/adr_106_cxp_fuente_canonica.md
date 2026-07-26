# ADR 106 - fuente canonica de cuentas por pagar de proveedores

Estado: aceptada para implementacion P0; `NO-GO` para produccion hasta cerrar
la conciliacion, staging y pruebas indicadas en el Plan 106.

## Decision

`empresa_cuentas_por_pagar` y la ruta
`/api/empresa/finanzas/cuentas_pagar` son la fuente canonica futura de CxP por
empresa. Los nuevos abonos se registran mediante `empresa_cxp_pagos`, que
guarda la asignacion explicita, el movimiento financiero y un evento outbox en
una transaccion, con idempotencia por hash de clave y aislamiento por
`empresa_id`.

`empresa_contabilidad_cartera_cxp` queda como superficie historica temporal
para sus registros previos. Desde P106 no acepta nuevas CxP ni abonos CxP: no
se migra, fusiona ni borra automaticamente ningun registro historico.

## Motivo

Los reportes y la captura de soportes IA ya consultan o escriben
`empresa_cuentas_por_pagar`; conservar nuevas escrituras CxP en la tabla
contable avanzada deja saldos fuera de reportes. La anterior secuencia de
abono creaba el egreso financiero y luego actualizaba la CxP sin una
transaccion comun.

## Invariantes P0 implementadas

- cada pago nuevo requiere `Idempotency-Key`; solo se persiste su hash;
- el pago se bloquea por cuenta y empresa, rechaza un monto mayor al saldo y
  no lo reduce silenciosamente;
- un reintento concurrente con la misma clave vuelve a comprobar la asignación
  después de adquirir el bloqueo de la cuenta y devuelve el resultado original;
- el registro de asignacion, el egreso financiero, el saldo CxP y el evento
  outbox se confirman o revierten juntos;
- la tabla de asignaciones usa dinero `NUMERIC(18,2)` y una unicidad por
  empresa y clave de idempotencia;
- API y worker no ejecutan DDL: la tabla aparece exclusivamente en la
  migracion `20260724-001-cxp-atomic-payments-v1` de `pcs-migrate`.

## Pendiente antes de habilitar produccion

1. Inventario y conciliacion por empresa de ambas tablas, documento,
   proveedor, fecha y saldo; aprobacion de contador para cada diferencia.
2. Migracion de importes historicos `REAL` y fechas `TEXT` tras un ensayo
   PostgreSQL en staging, backup y rollback probado.
3. Servicio transaccional para crear obligaciones desde compras y soportes IA;
   hoy no se debe interpretar este cambio como autorizacion de esas rutas.
4. Pruebas PostgreSQL A/B, carrera, reintento y recuperacion de transaccion;
   prueba visual de los flujos y validacion real controlada segun Plan 106.
5. Permisos separados de proponer, aprobar, pagar, conciliar, ajustar y
   reversar; ningun pago bancario real queda habilitado por este ADR.
