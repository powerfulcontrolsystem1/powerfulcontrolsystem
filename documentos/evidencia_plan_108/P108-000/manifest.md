# P108-000 - manifiesto del candidato inicial

Fecha local: 2026-07-25
Rama candidata: `codex/plan-108-candidate`
Base: `3dd48d243527bf51581d66c9b396029a0dbd43bb`
Estado: **parcial; commit candidato creado, digest de imagen pendiente**

Commit candidato: `f6f9546aeb926d01f76d67e250f2c78962aa0ced`

## Alcance clasificado

| Bloque | Archivos principales | Estado de validación |
|---|---|---|
| CxP y soportes IA | `backend/db/cuentas_por_pagar.go`, `backend/db/soportes_compras_ia.go`, `backend/handlers/modulos_faltantes.go`, `backend/handlers/soportes_compras_ia.go`, `web/administrar_empresa/finanzas.html` | pruebas dirigidas aprobadas; falta staging PostgreSQL y flujo autenticado |
| IA empresarial | `backend/ai/enterprise.go`, `backend/db/ai_enterprise.go`, `backend/db/chat_inteligencia_artificial.go`, handlers y `web/js/ai_chat_drawer.js` | pruebas dirigidas aprobadas; falta publicación y evaluación E2E |
| Contador y ReportSpec | `backend/handlers/reportes*.go`, `backend/db/reportes_programacion.go`, `web/administrar_empresa/reportes_ejecutivos.html` | suite Go aprobada; falta staging/UAT contable |
| Migraciones y esquema | `backend/db/platform_migrations.go`, catálogo e inventarios `Ensure*` | auditoría estricta aprobada; falta aplicación y segunda pasada en staging |
| UX, impresión y QA | `tools/qa_e2e_buttons.cjs`, `tools/qa_print_formats.cjs`, inventarios y páginas empresariales | sintaxis y auditoría estática aprobadas; falta recorrido visual autenticado |
| Documentación | planes 106, 107, IA y 108, arquitectura, mapas, historial y descripciones | revisada; falta actualizar evidencias de fases posteriores |

## Validaciones de este corte

- `go test ./... -count=1`: aprobado.
- `go vet ./...`: aprobado.
- `go build ./cmd/pcs-migrate` y `go build ./cmd/pcs-worker`: aprobados.
- Pruebas dirigidas de `db`, `handlers` e `ai`: aprobadas.
- Sintaxis JavaScript de CxP/IA/QA: aprobada.
- `tools/ensure_bootstrap_inventory.mjs --check`: 154 funciones `Ensure*` y 122 pasos de catálogo legado.
- `tools/migration_audit.mjs --strict`: aprobado.
- `scripts/profesional_preflight.ps1 -Full`: aprobado; reporte
  `documentos/reportes_profesionales/preflight_20260725_232918.md`.
- `git diff --check`: aprobado antes de crear este manifiesto.

## Exclusiones y bloqueos que permanecen

- No hay digest de imagen ni despliegue del candidato; por tanto no hay
  certificación de staging o producción.
- No se aplicaron migraciones ni se tocaron datos reales.
- No se enviaron documentos DIAN, cobros, correos, WhatsApp ni mutaciones IA.
- Las pruebas reales de Powerful Control System comienzan solamente después del
  commit candidato, su despliegue aislado y la autorización operacional de cada
  efecto externo.

## Siguiente paso

Construir la imagen por digest en un runner con Docker, ejecutar la matriz CI
Linux incluida `-race` y promover exclusivamente ese digest a staging. No se
promociona a producción hasta superar las compuertas de staging.
