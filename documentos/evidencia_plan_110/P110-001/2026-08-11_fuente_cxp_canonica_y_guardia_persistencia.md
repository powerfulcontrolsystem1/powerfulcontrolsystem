# P110-001 - Fuente CxP canónica y guardia de persistencia

Fecha: 2026-08-11

Ambiente: rama limpia P110 local

Producción: no modificada

## Decisión verificada

`documentos/arquitectura/adr_106_cxp_fuente_canonica.md` declara
`empresa_cuentas_por_pagar` como fuente canónica de las nuevas obligaciones de
proveedores. Los pagos canónicos usan `empresa_cxp_pagos`, movimiento financiero
y outbox dentro de una transacción por `empresa_id` e idempotencia.

`empresa_contabilidad_cartera_cxp` conserva únicamente el historial para
consulta y conciliación. No se migraron ni eliminaron datos históricos.

## Guardias añadidas

- `CreateEmpresaCarteraCXP` rechaza `tipo=cxp` antes de tocar SQL.
- `AplicarEmpresaCarteraCXPAbono` rechaza un registro histórico CxP después de
  resolverlo por `empresa_id` e ID, antes de modificar saldo o crear efecto.
- La excepción no afecta CxC histórico.
- El handler ya bloqueaba esas acciones; ahora una llamada interna tampoco puede
  reabrir la segunda fuente de verdad.

## Validación ejecutada

Desde `backend`:

```text
go test ./handlers -run TestSoporteComprasIAClamAVCleanMalwareAndFailClosed|TestSupportAntivirusMetricsAreConcurrentAndTenantFree|TestGenerateDIANUBLBase|TestSignDIANXMLXAdESBase|TestBuildDIANGetStatusZipEnvelope -count=1
go test ./db -run Test.*ControlElectrico|Test.*Cartera|Test.*CXP -count=1
go test ./db -run TestLegacyAccountingCXPRejectsNewWritesBeforeDatabase|TestEmpresaCxPAtomicSchemaHasTenantIdempotencyAndLedger|TestRegistrarEmpresaCxPAbonoKeepsTenantScopedAtomicInvariants -count=1
```

Resultado: PASS. La nueva prueba comprueba el rechazo de CxP histórico antes de
conectar a la base y conserva CxC. La compuerta profesional completa posterior
aprobó auditorías, migraciones, Compose, `go test ./...`, `go vet ./...` y
`git diff --check`.

## Pendientes de aceptación

- CI completo, base vacía, upgrade y rollback del SHA final.
- Conciliación PCS de ambas fuentes y firma del contador.
- Prueba autenticada A/B, carrera y recuperación outbox sobre el digest final.

Estado: **parcial**. No se aplicaron cambios a staging ni producción.
