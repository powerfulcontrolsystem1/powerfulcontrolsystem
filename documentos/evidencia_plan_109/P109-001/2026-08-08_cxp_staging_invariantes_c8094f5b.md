# P109-001 - invariantes CxP canónica en staging

Fecha: 2026-08-08  
Ambiente: `staging`; empresa PCS (`empresa_id=12`); produccion excluida.

## Fuente y esquema verificados

- El ledger registra aplicada la migración
  `20260724-001-cxp-atomic-payments-v1` bajo el scope `platform`.
- `empresa_cxp_pagos` existe.
- `empresa_cxp_pagos.monto` y `empresa_cuentas_por_pagar.saldo` son
  `NUMERIC(18,2)`.
- PCS conserva 3 obligaciones canónicas y 0 registros en la superficie CxP
  histórica; no se modificaron esas filas.

## Invariantes de datos actuales

- 5 pagos CxP para PCS.
- 0 pagos sin obligación canónica asociada.
- 0 pagos sin movimiento financiero asociado.
- 0 pagos no positivos.
- 0 obligaciones canónicas con saldo negativo.

## Pruebas de código

- `go test ./db ./handlers -run 'CxP|CXP|Cartera' -count=1`: PASS.
- Las pruebas de esquema, hash de idempotencia, bloqueo por cuenta, reintento y
  conciliación unitaria aprobaron.
- La integración PostgreSQL local `TestEmpresaCarteraMoneyPrecisionPostgres`
  quedó `SKIP` porque `PCS_TEST_POSTGRES_DSN` no está configurado; no se
  sustituye por esta lectura de staging.

## Estado

Las invariantes y el esquema de la fuente canónica están presentes en staging,
pero P109-001 sigue parcial: faltan una segunda identidad/empresa para A/B,
concurrencia transaccional real sobre fixture reversible y aprobación de
conciliación contable.
