# P110-003 — Corrección de acciones canónicas del Centro IA

Fecha: 2026-08-12 (America/Bogota)  
Alcance: corrección del candidato; la detección ocurrió durante una prueba
autenticada de PCS en staging.

## Hallazgo reproducible

El catálogo visible emite IDs canónicos como `compras_gastos`,
`cumplimiento_dian`, `cobranza_pagos` e `inventario_productos`. El normalizador
del backend aceptaba solo sinónimos parciales, por lo que un ID canónico no
reconocido caía silenciosamente en `diagnostico_erp`. La ejecución visible de
compras confirmó el síntoma: consumió una consulta IA pero mostró diagnóstico.

## Corrección

`normalizeEmpresaIAAccion` ahora preserva todos los IDs publicados por el
catálogo, además de sus sinónimos. Las entradas ajenas al catálogo continúan
fallando de forma segura hacia `diagnostico_erp`; no se habilitan acciones
arbitrarias ni mutaciones automáticas.

## Verificación

- Nueva prueba unitaria cubre los siete IDs canónicos y una acción desconocida.
- `go test ./handlers -run "TestNormalizeEmpresaIAAccion|TestCentroIAEmpresarial" -count=1`: PASS.
- `go vet ./handlers`: PASS.

La corrección queda pendiente de construir como nuevo candidato inmutable y de
repetir las funciones IA afectadas en staging. P110-003 sigue **parcial**.
