# Runbook de madurez empresarial - 12 pasos

Fecha: 2026-05-11

Este runbook deja implementados los controles que elevan la plataforma a operacion empresarial. Cada punto tiene script, auditoria o workflow asociado.

## 1. Staging con datos anonimizados

- Script: `deploy/scripts/vps-anonymize-staging.sh`.
- Integracion: `deploy/scripts/vps-refresh-staging-from-production.sh` lo ejecuta por defecto.
- Variable de escape controlado: `STAGING_ANONYMIZE=0`.
- Emails preservados por defecto: `powerfulcontrolsystem@gmail.com`.

## 2. Monitoreo profesional

- Compose: `deploy/monitoring/docker-compose.monitoring.yml`.
- Config: `deploy/monitoring/prometheus.yml`.
- Script VPS: `deploy/scripts/vps-monitoring-up.sh`.
- Servicios: Prometheus, Grafana, node-exporter y cAdvisor.
- Binding por defecto: `127.0.0.1`, para publicarlo solo detras de Nginx/SSL y autenticacion.

## 3. E2E por roles

- Auditor: `tools/qa_roles_matrix.mjs`.
- Workflow: `.github/workflows/e2e-visual.yml`.
- Roles cubiertos por contrato: super administrador, administrador empresa, cajero, vendedor, asesor comercial y soporte.

## 4. Control formal de releases

- Compuerta: `scripts/release_gate.ps1`.
- Manifiesto: `tools/release_manifest.mjs`.
- Salida: `documentos/releases/release_<version>.md` y `.json`.

## 5. Backups externos

- Script: `deploy/scripts/vps-external-backup.sh`.
- Cron: `deploy/scripts/vps-install-external-backup-cron.sh`.
- Destinos: `rclone` o `s3`.
- El backup local existente sigue siendo la fuente primaria.

## 6. Seguridad avanzada

- Auditor base: `tools/security_audit.mjs`.
- Refuerzo operativo: recaptcha, cookies seguras, sesiones revocables, rate limit y rutas publicas controladas.
- Siguiente capa activable: 2FA para cuentas super antes de exponer paneles criticos.

## 7. Observabilidad de negocio

- Auditor: `tools/observability_report.mjs`.
- Snapshot VPS: `deploy/scripts/vps-observability-snapshot.sh`.
- Paneles recomendados: empresas activas, licencias, pagos fallidos, errores por modulo, uso de IA y capacidad VPS.

## 8. Normalizacion documental

- Auditor: `tools/docs_normalization_audit.mjs`.
- Objetivo: detectar documentos con codificacion historica danada sin modificar trazabilidad automaticamente.

## 9. Deploy automatico a staging

- Workflow: `.github/workflows/professional-ci.yml`.
- Activacion: variable GitHub `PCS_ENABLE_STAGING_DEPLOY=true`.
- Secretos: `PCS_STAGING_HOST`, `PCS_STAGING_USER`, `PCS_STAGING_SSH_KEY`, opcional `PCS_STAGING_PATH`.
- Produccion sigue manual por seguridad.

## 10. Matriz de pagos y comprobantes

- Auditor: `tools/payment_matrix_audit.mjs`.
- Cubre Wompi, Epayco, webhooks, retorno de pasarela y evidencia visual de comprobantes.

## 11. Pruebas de carga

- Smoke load: `tools/load_smoke_test.mjs`.
- Variables: `PCS_LOAD_BASE_URL`, `PCS_LOAD_CONCURRENCY`, `PCS_LOAD_REQUESTS`, `PCS_LOAD_P95_THRESHOLD_MS`.
- Integrado en `scripts/release_gate.ps1`.

## 12. Centro de soporte interno

- Auditor: `tools/support_center_audit.mjs`.
- Cubre soporte remoto, errores de sistema, alertas, auditoria de empresa y runbooks.

## Comandos

```powershell
.\scripts\profesional_preflight.ps1 -Full
.\scripts\release_gate.ps1 -SkipE2E
node tools\load_smoke_test.mjs
node tools\release_manifest.mjs --version=1.0.0
```

```bash
bash deploy/scripts/vps-refresh-staging-from-production.sh
bash deploy/scripts/vps-monitoring-up.sh
bash deploy/scripts/vps-external-backup.sh
```
