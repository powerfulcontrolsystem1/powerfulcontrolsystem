# P108-003 - outbox, worker e idempotencia

Fecha: 2026-07-28  
Ambiente: PostgreSQL efímero aislado en VPS  
SHA probado: `4a5eb92b7dad07e1aab8498305ed3665ca0030da`

## Ensayo duradero aislado

La prueba `TestP108OutboxWorkerDurabilityIntegration` se ejecutó contra una
base PostgreSQL 16 temporal cuyo nombre inicia por `p108_`. La prueba exige de
forma explícita `P108_OUTBOX_TEST_DATABASE=isolated`; no acepta un DSN que no
contenga ese prefijo y no recibió datos, credenciales ni conexiones de PCS.

Resultado: **PASS** (`go test ./db -run
^TestP108OutboxWorkerDurabilityIntegration$ -count=1`).

Cobertura verificada:

- inserción repetida del mismo evento: una sola fila persistente;
- despacho repetido: un único job durable;
- dos workers concurrentes: un único claim mediante `FOR UPDATE SKIP LOCKED`;
- caída tras el claim: el lease vence y otro worker recupera y completa el job;
- error definitivo: el job conserva estado `dead` (dead-letter).

## Limpieza y aislamiento

Al terminar se verificó que no quedaron el contenedor, red ni checkout con el
prefijo `p108-outbox-4a5eb92b`. Staging continuó saludable en frontend, API,
worker y PostgreSQL.

## Estado

P108-003 permanece **parcial**: el mecanismo central y sus fallos de entrega
están cubiertos, pero faltan inventario completo de jobs históricos, pruebas de
proveedores reales (DIAN, pagos, correo), métricas operativas de cola y ensayo
con réplicas worker del candidato desplegado.

## Estado de cola del candidato 2026-07-30

Después de 500 lecturas autenticadas y un reinicio controlado del backend, el
PostgreSQL de staging mostró un evento `published`, un `dead` histórico, cero
eventos `pending/processing` listos y cero leases vencidos. Los logs del worker
de los últimos 15 minutos no registraron errores, pánicos ni dead-letter nuevos.

El subcriterio de ausencia de acumulación inmediata queda **PASS**. P108-003
continúa parcial por proveedores reales, métricas/alertas de cola y réplicas
concurrentes del worker desplegado.

## Hallazgo CxP del candidato `5566a213` - 2026-07-30

La conciliación SQL de la CxP QA `P108-CONC-5EC1-941964` confirmó que los dos
pagos suman `2`, apuntan a dos movimientos distintos que también suman `2`, y
la cuenta terminó pagada con saldo `0`. Sin embargo, sus dos eventos
`cuentas_por_pagar.pago_registrado` terminaron `dead` después de 5 intentos con
el error `outbox topic has no enabled worker handler`.

La corrección registra el topic en el worker y lo convierte en un evento
contable `finanzas/abono_proveedor_registrado`. El procesador:

- resuelve pago, CxP y movimiento con `empresa_id` en cada `JOIN`;
- bloquea la fila del pago antes de revisar el evento natural;
- rechaza diferencias de monto entre payload, pago y movimiento;
- usa `empresa_cxp_pagos/pago_id` como identidad natural para que un retry no
  duplique el evento ni el asiento posterior.

Se añadieron contratos unitarios y una integración PostgreSQL descartable para
primera ejecución, replay y rechazo entre tenants. Falta publicar el nuevo
digest, probar un pago PCS controlado y definir la recuperación explícita de
los dos eventos históricos `dead`; no se reactivan masivamente de forma
automática.

## Cierre del handler CxP en staging - candidato `f7214329`

El código del handler quedó incluido en el candidato final
`f7214329ed70b15085f300d823244617b9cb998f`. CI profesional
`30595918626` y release inmutable `30595920016` terminaron `success`. Los
cuatro digests exactos se promovieron solo a staging; migrador `exit 0`,
salud/readiness verdes y huella de producción sin cambios.

Se registró desde la interfaz oficial de Finanzas un abono interno de `0,01`
sobre `P108-CXP-397E42A7`. La trazabilidad de solo lectura por
`empresa_id=12` confirmó:

| Etapa | Identidad | Resultado |
| --- | ---: | --- |
| Pago CxP | `pago_id=4` | monto `0,01`, movimiento `35` |
| Outbox | `id=5` | `published`, un intento, sin error |
| Job durable | `id=6065388` | `completed`, un intento, sin error |
| Evento contable | `id=97` | único, monto `0,01`, procesado |
| Asiento | `id=86` | débito `0,01`, crédito `0,01`, diferencia `0` |

Las líneas del asiento debitan `220505` y acreditan `110505`. Los conteos
naturales son un evento y un asiento para el pago, por lo que no hubo replay
duplicado.

El flujo nuevo queda **PASS**. P108-003 permanece **parcial / NO-GO** porque
los dos eventos CxP históricos `dead` requieren inventario, previsualización y
recuperación explícita por empresa; no se reactivaron automáticamente.
