# P109-010 - Durabilidad de outbox y jobs en PostgreSQL real aislado

Fecha: 2026-08-09

## Ensayo

Se levantó PostgreSQL 16.14 efímero con la base explícitamente marcada
`p108_outbox`. No se conectó ni alteró ninguna base de staging o producción.

La ejecución aprobó en 27.759 segundos:

```text
P108_OUTBOX_TEST_DATABASE=isolated
go test ./db -count=1 -run 'TestP108OutboxWorkerDurabilityIntegration|TestP108CxPPaymentAccountingIntegration'
ok github.com/you/pos-backend/db 27.759s
```

## Controles aprobados

- Dos inserciones lógicas del mismo evento conservan una sola fila outbox.
- Dos dispatches conservan un único job durable.
- Dos workers concurrentes reclaman exactamente un único job mediante lease y
  `SKIP LOCKED`.
- Un worker abandonado deja un lease recuperable; el job se reclama y completa
  después del vencimiento.
- Un fallo deliberado deja el job en `dead` con trazabilidad.
- El pago CxP genera un único evento contable ante reintento y rechaza el mismo
  identificador desde otra empresa.

## Limpieza

El túnel SSH y el contenedor exacto `p109-outbox-integration-488c65be` se
detuvieron y eliminaron. El puerto local temporal `15433` quedó sin listener.

## Resultado

**PASS** para lease, deduplicación, dead-letter, recuperación e idempotencia en
PostgreSQL real aislado. P109-010 sigue parcial hasta operar el canal externo,
responsables y antivirus real en el candidato final.
