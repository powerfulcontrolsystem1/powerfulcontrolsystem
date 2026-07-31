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
