# Release 2026.07.22

Fecha: 2026-07-22T03:16:20.017Z
Rama: codex/rs-20260718-143849
Commit: 72fa7007e1d068e7c732f280e0f226fd2858b427
Working tree limpio: no
Base: origin/main (3d39ba372befc74fab336279f9a070f32b290d62)
Base ancestro del candidato: si
Upstream: origin/codex/rs-20260718-143849
Bloqueos de release: working_tree_dirty, PCS_API_IMAGE_DIGEST_missing, PCS_MIGRATE_IMAGE_DIGEST_missing, PCS_WORKER_IMAGE_DIGEST_missing

## Imagenes inmutables

- PCS_API_IMAGE_DIGEST: pendiente
- PCS_MIGRATE_IMAGE_DIGEST: pendiente
- PCS_WORKER_IMAGE_DIGEST: pendiente

## Checks requeridos

- scripts/profesional_preflight.ps1 -Full
- scripts/vps_backup_operacion.ps1
- scripts/vps_restore_validation.ps1 -ExecuteDrill
- tools/qa_e2e_buttons.cjs against staging
- tools/qa_print_formats.cjs
- tools/load_smoke_test.mjs against staging

## Commits recientes

- 72fa7007 Actualizar registros automaticos de repositorio
- e9c58602 rs: corregir contrato CI del bootstrap operativo
- 74d878e9 rs: corregir bootstrap operativo Docker en VPS
- 3d39ba37 fix: provision migration mode on VPS (#40)
- d9648512 security: require explicit production migration bootstrap (#39)
- ef5816a1 rs: actualizar repositorio y sincronizar VPS (#37)
- cf002d46 rs: actualizar repositorio y sincronizar VPS (#36)
- 057be6c2 fix: estabilizar cierre de carrito en PostgreSQL (#35)
- da09eea5 fix: normalize generated inventory line endings (#34)
- 7720f973 fix: asegurar CSRF en subpaginas empresariales (#33)
