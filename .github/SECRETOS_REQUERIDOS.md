# Secretos requeridos en GitHub Actions

Configurar en `Settings > Secrets and variables > Actions`.

## E2E visual

- `PCS_QA_EMAIL`: usuario de pruebas, recomendado Motel Calipso.
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
El override de staging usa contenedores, worker y almacenamiento privado
separados; `scripts/staging_up.ps1` y CI validan ese contrato antes de levantarlo.

## Firma Android de distribucion

El workflow `Mobile release candidates` usa el entorno protegido
`mobile-production`. Configura alli, nunca en el repositorio, estos secretos:

- `ANDROID_KEYSTORE_BASE64`: keystore de subida codificado en Base64.
- `ANDROID_KEYSTORE_PASSWORD`
- `ANDROID_KEY_ALIAS`
- `ANDROID_KEY_PASSWORD`

La validacion normal de la aplicacion compila solo APK de depuracion e iOS sin
firma. El AAB firmado se genera exclusivamente en el workflow manual con los
cuatro secretos y elimina el material transitorio al terminar.

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
