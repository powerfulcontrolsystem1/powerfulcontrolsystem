# P109-001 - Recuperación CxP controlada en staging

Fecha: 2026-07-31
Entorno: staging, empresa Powerful Control System (`empresa_id=12`).

## Recuperación ya verificada

La vista de superadministración mostró únicamente eventos `dead` del topic de
pagos CxP y no expuso el payload crudo. Se reactivaron de manera explícita los
dos eventos autorizados del ensayo PCS, conservando un tercer evento ajeno al
alcance en estado `dead`.

Tras el worker programado, los dos eventos reactivados quedaron publicados. La
conciliación de solo lectura comprobó para ambos: pago y movimiento existentes,
evento contable único, asiento balanceado, saldo pendiente en cero y cuenta CxP
en estado pagada. La auditoría registró una entrada por evento y no se generó un
pago duplicado.

## Pruebas de contrato repetidas

Se ejecutó desde `backend`:

```text
go test ./db ./handlers -run 'Test(NormalizeOutboxRecoveryInput|OutboxRecoveryIDClauseUsesBoundParameters|BuildSuperOutboxRecoveryItemsDoesNotExposeRawPayload|ValidateSuperOutboxRecoveryRequestRequiresUniqueBoundedIDsAndReason|SuperConfigHandlersRequireSuperAdmin)' -count=1
```

Resultado: `db` PASS y `handlers` PASS. La suite cubre empresa/topic/IDs
obligatorios, duplicados, límite, parámetros enlazados, ocultamiento del payload
y acceso exclusivo de superadministración.

## Límite de cierre

P109-001 sigue parcial: falta ejecutar en runtime autenticado los rechazos para
ID publicado/inexistente/repetido y una matriz A/B con una identidad operativa
de otra empresa. No se reprocesarán pagos ni se modificarán datos fuera del flujo
autorizado.
