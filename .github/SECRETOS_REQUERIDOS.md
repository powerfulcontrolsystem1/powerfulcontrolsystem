# Secretos requeridos en GitHub Actions

Configurar en `Settings > Secrets and variables > Actions`.

## E2E visual

- `PCS_QA_EMAIL`: usuario de pruebas, recomendado Motel Calipso.
- `PCS_QA_PASSWORD`: clave del usuario de pruebas.

## Futuro deploy automatico

No se activa despliegue automatico por seguridad. Si se decide activarlo en el futuro, crear secretos separados para staging y produccion:

- `PCS_STAGING_HOST`
- `PCS_STAGING_USER`
- `PCS_STAGING_SSH_KEY`
- `PCS_PRODUCTION_HOST`
- `PCS_PRODUCTION_USER`
- `PCS_PRODUCTION_SSH_KEY`

Mantener produccion manual hasta que staging tenga E2E verde de forma repetida.
