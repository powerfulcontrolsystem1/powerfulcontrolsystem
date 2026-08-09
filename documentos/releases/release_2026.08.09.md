# Release 2026.08.09

Fecha: 2026-08-09T10:34:45.412Z
Rama: codex/p109-batch-no-pr
Commit: 8662b6ba24faf3ef329c8d41744fec1175d11ad4
Working tree limpio: si
Base: origin/main (5bf245889e784892a038197ae2e11d2b142ff005)
Base ancestro del candidato: si
Upstream: origin/codex/p109-batch-no-pr
Bloqueos de release: PCS_API_IMAGE_DIGEST_missing, PCS_MIGRATE_IMAGE_DIGEST_missing, PCS_WORKER_IMAGE_DIGEST_missing, PCS_FRONTEND_IMAGE_DIGEST_missing

## Imagenes inmutables

- PCS_API_IMAGE_DIGEST: pendiente
- PCS_MIGRATE_IMAGE_DIGEST: pendiente
- PCS_WORKER_IMAGE_DIGEST: pendiente
- PCS_FRONTEND_IMAGE_DIGEST: pendiente

## Checks requeridos

- scripts/profesional_preflight.ps1 -Full
- scripts/vps_backup_operacion.ps1
- scripts/vps_restore_validation.ps1 -ExecuteDrill
- tools/qa_e2e_buttons.cjs against staging
- tools/qa_print_formats.cjs
- tools/load_smoke_test.mjs against staging

## Commits recientes

- 8662b6ba fix: hacer reproducible la compuerta de release
- bb285968 fix: presentar respuestas IA seguras en Centro IA
- 488c65be test: cerrar aislamiento AB e integridad CxP P109
- 71984cb5 docs: registrar drills y retencion P109
- ab4bb3fd docs: verificar candidato CxP IA ampliado
- 47a8ef9e docs: registrar papelera visual en staging
- 34cfd852 fix: hacer verificables acciones de soportes IA
- d020e4c1 docs: registrar flujo CxP IA del candidato
- 3783de09 docs: certificar candidato DIAN en staging
- 17c55dd8 fix: migrar indicador local DIAN por catálogo
