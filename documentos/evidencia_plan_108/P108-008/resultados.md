# P108-008 - preflight y contrato P107 contador

Fecha: 2026-07-28  
Ambiente: staging aislado, empresa de prueba Powerful Control System (`empresa_id=12`)  
Alcance: solo lectura; no se crearon registros.

## Resultado

El preflight P107 contra `http://127.0.0.1:8082/health` respondió `200 OK` y
declaró `ready_for_fixture_data=true`. Se verificó además que el host de staging
dispone de Docker. El contrato preserva el alcance: staging únicamente, sin
DIAN, banca, correos ni datos productivos.

El manifiesto determinista `P107-QA-CONTADOR` fue generado para la empresa 12:

- versión: `p107-fixtures-v1`;
- escenarios: apertura, tres ventas de menta, abono CxC, compra a crédito,
  abono CxP, impuestos/retenciones y reverso auditable;
- SHA-256: `e29e0db2e37c559e78cecec411d1119713af6d0f1a9e75906de2d37088fe108b`.

El manifiesto incluye claves idempotentes y pasos de reversión mediante flujos
auditables, no SQL directo. No ejecuta ninguna de esas operaciones por sí solo.

## Estado

P108-008 queda **parcial**. Falta un ejecutor transaccional de fixtures que
pueda realizar y revertir la matriz por el flujo oficial, la conciliación de
inventario/caja/banco/CxC/CxP/impuestos/asientos, reportes finales y UAT
independiente de contador.
