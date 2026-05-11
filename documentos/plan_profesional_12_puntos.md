# Plan profesional de 12 puntos

Fecha: 2026-05-11

Este documento deja los 12 frentes convertidos en piezas operativas del proyecto. La regla es simple: cada mejora debe poder validarse por script, CI o evidencia.

## 1. Ambiente staging

- Compose: `deploy/docker-compose.staging.yml`.
- Variables: `deploy/.env.staging.example`.
- Local: `.\scripts\staging_up.ps1 -ConfigOnly` o `.\scripts\staging_up.ps1 -Build`.
- VPS: `bash deploy/scripts/vps-staging-up.sh`.
- Nginx/SSL: `bash deploy/scripts/vps-configure-staging-nginx.sh`.
- Puerto recomendado: `127.0.0.1:8082`, con dominio futuro `staging.powerfulcontrolsystem.com`.

## 2. CI/CD automatico

- Workflow: `.github/workflows/professional-ci.yml`.
- Valida preflight completo, auditorias, OpenAPI, Docker production/staging y reporte de observabilidad.
- No publica a VPS automaticamente para evitar despliegues no revisados.

## 3. Pruebas E2E visuales

- Workflow manual: `.github/workflows/e2e-visual.yml`.
- Tambien se ejecuta programado de lunes a viernes contra staging si existen secretos.
- Scripts existentes integrados: `tools/qa_e2e_buttons.cjs` y `tools/qa_print_formats.cjs`.
- Requiere secretos GitHub `PCS_QA_EMAIL` y `PCS_QA_PASSWORD`.
- Evidencias: `test_runs/` como artifact de GitHub Actions.

## 4. Seguridad fuerte

- Auditor: `tools/security_audit.mjs`.
- Nginx staging instala headers seguros y CSP en modo Report-Only mediante `deploy/scripts/vps-configure-staging-nginx.sh`.
- Revisa cookies, SameSite, HttpOnly, helper Secure, rutas publicas, CORS, recaptcha, sesiones y alcance multiempresa.
- Puede ejecutarse en modo estricto: `node tools/security_audit.mjs --strict`.

## 5. Observabilidad

- Auditor: `tools/observability_report.mjs`.
- Snapshot VPS: `bash deploy/scripts/vps-observability-snapshot.sh`.
- Revisa healthchecks Docker, logs persistentes, modulo de alertas, scanner VPS y reportes profesionales.
- Los reportes quedan en `documentos/reportes_profesionales/`.

## 6. Backups con prueba de restauracion

- Backup: `.\scripts\vps_backup_operacion.ps1`.
- Cron VPS: `bash deploy/scripts/vps-install-backup-cron.sh`.
- Validacion no destructiva: `.\scripts\vps_restore_validation.ps1`.
- Restauracion temporal real: `.\scripts\vps_restore_validation.ps1 -ExecuteDrill`.
- La prueba usa contenedor PostgreSQL temporal y no toca la base de produccion.

## 7. Migraciones versionadas

- Base existente: `backend/db/migrations.go`.
- Auditor: `tools/migration_audit.mjs`.
- El preflight completo ejecuta pruebas Go para proteger migraciones.
- El inventario profesional exige documentar migraciones nuevas en changelog e historial.

## 8. Matriz de permisos automatizada

- Auditor: `tools/permissions_license_audit.mjs`.
- Compara modulos backend, wrappers, licencias y frontend.
- Sirve como compuerta para detectar funciones nuevas sin rol/licencia asignada.

## 9. Logs centralizados

- Docker mantiene `pcs_backend_logs` y su variante staging.
- El reporte de observabilidad verifica persistencia de logs.
- El modulo de alertas del super administrador concentra eventos de infraestructura.

## 10. Documentacion API

- Generador: `tools/openapi_inventory.mjs`.
- Salida: `documentos/api/openapi.generated.yaml`.
- Inventaria automaticamente rutas de `backend/main.go`.

## 11. Recuperacion ante desastre

- Runbook: `documentos/gobernanza_tecnica/runbooks/runbook_recuperacion_desastre_docker_vps.md`.
- Release gate: `scripts/release_gate.ps1`.
- Cubre imagenes, volumenes, secretos, Nginx, SSL, restauracion y verificacion.

## 12. Pagos y facturacion

- Contrato base existente: `documentos/gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md`.
- QA visual de impresion: `tools/qa_print_formats.cjs`.
- Contrato QA de modulos criticos: `tools/qa_module_contracts.mjs`.
- E2E visual manual incluye capturas de botones, pagos y comprobantes cuando se configuran rutas/credenciales.

## Comando profesional recomendado

```powershell
.\scripts\profesional_preflight.ps1 -Full
```

## Comando de publicacion segura

```powershell
.\rs.ps1
```
