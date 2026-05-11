# Plan profesional de 12 puntos

Fecha: 2026-05-11

Este documento deja los 12 frentes convertidos en piezas operativas del proyecto. La regla es simple: cada mejora debe poder validarse por script, CI o evidencia.

## 1. Ambiente staging

- Compose: `deploy/docker-compose.staging.yml`.
- Variables: `deploy/.env.staging.example`.
- Local: `.\scripts\staging_up.ps1 -ConfigOnly` o `.\scripts\staging_up.ps1 -Build`.
- VPS: `bash deploy/scripts/vps-staging-up.sh`.
- Nginx/SSL: `bash deploy/scripts/vps-configure-staging-nginx.sh`.
- Refresco de datos: `bash deploy/scripts/vps-refresh-staging-from-production.sh`.
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
- 2FA/TOTP para super administrador: `/super/seguridad_2fa.html` y `/super/api/administradores/2fa`.
- El login administrativo acepta `otp_code` y lo exige cuando el 2FA esta activo.
- Nginx staging instala headers seguros y CSP en modo Report-Only mediante `deploy/scripts/vps-configure-staging-nginx.sh`.
- Revisa cookies, SameSite, HttpOnly, helper Secure, rutas publicas, CORS, recaptcha, sesiones y alcance multiempresa.
- Puede ejecutarse en modo estricto: `node tools/security_audit.mjs --strict`.

## 5. Observabilidad

- Auditor: `tools/observability_report.mjs`.
- Monitoreo real: `deploy/monitoring/docker-compose.monitoring.yml` con Prometheus, Grafana, node-exporter, cAdvisor, dashboard y reglas de alerta.
- Auditor de capacidad/negocio: `tools/business_observability_audit.mjs`.
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

## Extension empresarial 2026-05-11

- Staging anonimizado por defecto: `deploy/scripts/vps-anonymize-staging.sh`.
- Verificacion staging anonimizado: `deploy/scripts/vps-verify-staging-anonymization.sh` y `tools/staging_anonymization_audit.mjs`.
- Monitoreo Prometheus/Grafana: `deploy/monitoring/docker-compose.monitoring.yml`.
- Dashboard Grafana: `deploy/monitoring/grafana/dashboards/pcs-operacion.json`.
- SLO/SLA operativo: `documentos/gobernanza_tecnica/slo_sla_operativo.md` y `tools/slo_sla_audit.mjs`.
- Hardening VPS: `deploy/scripts/vps-hardening-audit.sh` y `tools/vps_hardening_audit.mjs`.
- Backups externos: `deploy/scripts/vps-external-backup.sh`.
- QA por roles: `tools/qa_roles_matrix.mjs`.
- Matriz de pagos/comprobantes: `tools/payment_matrix_audit.mjs`.
- Contrato de pagos reales: `documentos/gobernanza_tecnica/contratos/contrato_matriz_pagos_reales.md` y `tools/payment_real_matrix_audit.mjs`.
- Prueba de carga smoke: `tools/load_smoke_test.mjs`.
- Plan de carga por etapas: `tools/load_capacity_plan.mjs`.
- Manifiesto de release: `tools/release_manifest.mjs`.
- Centro de soporte: `tools/support_center_audit.mjs`.
- Contrato del centro de soporte: `documentos/gobernanza_tecnica/contratos/contrato_centro_soporte.md`.
- Normalizacion documental: `tools/docs_normalization_audit.mjs`.
- Runbook completo: `documentos/gobernanza_tecnica/runbooks/runbook_madurez_empresarial_12_pasos.md`.

## Comando profesional recomendado

```powershell
.\scripts\profesional_preflight.ps1 -Full
```

## Comando de publicacion segura

```powershell
.\rs.ps1
```

## Cierre implementable 1 a 8 - 2026-05-11

- Pagos/licencias: backend ya conserva correos de pago rechazado, banderas anti-duplicado, idempotencia de activacion y disponibilidad por pais para Wompi/Epayco; el checkout publico ahora agrega selector manual de pais para reconsultar `/api/public/licencias/payment_methods?pais_codigo=XX`.
- DIAN: permanece como limite externo la integracion SOAP/WSDL oficial completa, empaquetado ZIP, TrackId y firma certificable final. La plataforma mantiene diagnostico, XML/firma base, set de pruebas y contrato tecnico para no declararlo como cumplimiento final.
- QA real: queda cubierto por release gate, staging, E2E visual y auditorias; la certificacion funcional completa todavia depende de credenciales sandbox/productivas, hardware real y datos controlados por empresa.
- Integraciones fisicas/externas: RustDesk, sensores, impresoras, GPS, OnlyOffice, Nextcloud, voz IA, bancos y pasarelas reales deben cerrarse con pruebas de proveedor o equipo fisico; no se simulan como completadas en documentacion.
- Rendimiento: las auditorias profesionales quedan disponibles como compuerta; la deuda restante recomendada es sacar validaciones de esquema de peticiones calientes hacia arranque/migracion controlada donde el modulo lo permita.
- Documentacion: se corrigen referencias activas a documentos historicos inexistentes y queda pendiente una normalizacion masiva de mojibake en `CHANGELOG.md` e `historial_de_cambios`, que debe hacerse con una migracion documental controlada para no perder trazabilidad.
