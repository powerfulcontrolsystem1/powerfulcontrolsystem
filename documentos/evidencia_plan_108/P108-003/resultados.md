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
