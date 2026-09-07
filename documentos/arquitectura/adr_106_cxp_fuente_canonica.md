# ADR 106 - fuente canonica de cuentas por pagar de proveedores

Estado: Vigente. Responsable: Ingeniería backend y datos. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- La fuente canónica de nuevas CxP ya es Finanzas; la tabla contable antigua queda para lectura y conciliación.
- La decisión persiste aunque Plan 106 sea histórico; los pendientes de conciliación y aceptación siguen sin quedar acreditados por esta revisión.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

la conciliacion, staging y pruebas indicadas en el Plan 106.

## Decision

`empresa_cuentas_por_pagar` y la ruta
`/api/empresa/finanzas/cuentas_pagar` son la fuente canónica de nuevas CxP por
empresa. Los nuevos abonos se registran mediante `empresa_cxp_pagos`, que
guarda la asignacion explicita, el movimiento financiero y un evento outbox en
una transaccion, con idempotencia por hash de clave y aislamiento por
`empresa_id`.

`empresa_contabilidad_cartera_cxp` queda como superficie historica temporal
para sus registros previos. Desde P106 no acepta nuevas CxP ni abonos CxP: no
se migra, fusiona ni borra automaticamente ningun registro historico.

El catálogo operativo de proveedores para nuevas CxP es `proveedores`, que es
el mismo que administra Compras mediante `/api/empresa/proveedores`. La tabla
`empresa_proveedores` no se usa para CxP: pertenece a una variante ERP no
publicada por esa ruta y consultarla habría dejado fuera los proveedores
creados en la interfaz operativa.

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
- P110 refuerza la decisión en persistencia: `CreateEmpresaCarteraCXP` y
  `AplicarEmpresaCarteraCXPAbono` rechazan `tipo=cxp` aun si se invocan fuera
  del handler HTTP. La tabla histórica permanece disponible solo para lectura
  y conciliación; CxC histórico no cambia.

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

## Fuentes y aceptación de la revisión

[contabilidad_colombia_avanzada.go](../../backend/db/contabilidad_colombia_avanzada.go), [finanzas.go](../../backend/handlers/finanzas.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../requisitos/especificacion_y_trazabilidad.md)).
