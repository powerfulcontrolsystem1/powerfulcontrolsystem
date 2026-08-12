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

## Candidato y repetición en staging

El workflow de candidato inmutable aprobó para `d3d2141498a540a3d5db8a06ab37855a6ee70757`:
construcción única, Trivy, SBOM, publicación por digest y validación de Compose.
Se promovieron solamente esos digests a staging junto con el digest ClamAV ya
certificado. Backend, worker, frontend y ClamAV quedaron saludables; `/health`
y `/ready` respondieron correctamente.

La repetición autenticada confirmó que **Compras y gastos IA** entregó controles
de soportes, duplicados, causación e impuestos, y que **Cumplimiento DIAN**
entregó recomendaciones fiscales. Ninguna acción emitió, pagó, contabilizó ni
modificó documentos. P110-003 sigue **parcial**: aún faltan adjuntos, timeout,
cancelación, doble envío y cobertura efectiva de todos los roles.

La misma sesión cubrió las cuatro funciones restantes (`borrador_factura`,
`cobranza_pagos`, `inventario_productos` y `conciliacion_bancaria`). Las cuatro
completaron y mostraron contenido correspondiente a su dominio; junto con el
diagnóstico inicial, compras y fiscal, los **siete** IDs del catálogo fueron
ejercitados contra el candidato corregido. Todas quedaron en modo de
recomendación/borrador, sin confirmación ni mutación de negocio.
