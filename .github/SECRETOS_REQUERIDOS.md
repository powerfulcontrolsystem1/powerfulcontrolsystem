# Secretos requeridos en GitHub Actions

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Los nombres son un inventario de configuración, nunca valores. E2E requiere cuenta y empresa aisladas autorizadas, no una empresa comercial por defecto.
- Los secretos de producción enumerados para futuro despliegue no prueban que un workflow los consuma. Verificar el YAML y entorno protegido antes de configurarlos.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Configurar en `Settings > Secrets and variables > Actions`.

## E2E visual

- `PCS_QA_EMAIL`: usuario de pruebas de un entorno aislado y autorizado.
- `PCS_QA_PASSWORD`: clave del usuario de pruebas.

## Futuro deploy automatico

No se activa despliegue automatico por seguridad. Si se decide activarlo en el futuro, crear la variable `PCS_ENABLE_STAGING_DEPLOY=true` y secretos separados para staging y produccion:

- `PCS_STAGING_HOST`
- `PCS_STAGING_USER`
- `PCS_STAGING_SSH_KEY`
- `PCS_STAGING_PATH` opcional, por defecto `/root/powerfulcontrolsystem`.
- `PCS_PRODUCTION_HOST`
- `PCS_PRODUCTION_USER`
- `PCS_PRODUCTION_SSH_KEY`

Mantener produccion manual hasta que staging tenga E2E verde de forma repetida.

## Backups externos

En la VPS, no en GitHub, configurar variables para `deploy/scripts/vps-external-backup.sh`:

- `EXTERNAL_BACKUP_TARGET`: `none`, `rclone` o `s3`.
- `RCLONE_REMOTE`: destino tipo `remote:carpeta` cuando se usa rclone.
- `S3_URI`: destino tipo `s3://bucket/ruta` cuando se usa AWS CLI.

## Monitoreo

En la VPS, cambiar `deploy/monitoring/.env.monitoring` despues de ejecutar `bash deploy/scripts/vps-monitoring-up.sh`:

- `GRAFANA_ADMIN_PASSWORD`
- `PROMETHEUS_BIND`, `PROMETHEUS_PORT`
- `GRAFANA_BIND`, `GRAFANA_PORT`

## Fuentes y aceptación de la revisión

[professional-ci.yml](workflows/professional-ci.yml), [e2e-visual.yml](workflows/e2e-visual.yml).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../documentos/requisitos/especificacion_y_trazabilidad.md)).
