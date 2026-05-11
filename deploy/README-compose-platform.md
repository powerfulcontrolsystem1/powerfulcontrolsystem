# Docker Compose de Powerful Control System

Este despliegue prepara la plataforma para ejecutarse en Docker dentro de la VPS actual sin cortar el servicio que ya esta en produccion por `systemd` y Nginx. El flujo recomendado levanta Docker en paralelo por `127.0.0.1:8081`, valida, migra datos y solo despues conmuta Nginx.

## Operacion empresarial adicional

Staging seguro:

```bash
bash deploy/scripts/vps-refresh-staging-from-production.sh
```

El refresco ejecuta `deploy/scripts/vps-anonymize-staging.sh` por defecto. Para una copia exacta temporal, usar `STAGING_ANONYMIZE=0` de forma consciente.

Monitoreo:

```bash
bash deploy/scripts/vps-monitoring-up.sh
```

Prometheus y Grafana quedan ligados a `127.0.0.1` por defecto.
El script crea las redes internas faltantes, genera una clave fuerte de Grafana si solo existe el placeholder y carga el dashboard `pcs-operacion`.

Verificacion de anonimizacion staging:

```bash
bash deploy/scripts/vps-verify-staging-anonymization.sh
```

Hardening VPS:

```bash
bash deploy/scripts/vps-hardening-audit.sh
```

Backup externo:

```bash
EXTERNAL_BACKUP_TARGET=rclone RCLONE_REMOTE=remote:pcs-backups bash deploy/scripts/vps-external-backup.sh
```

o:

```bash
EXTERNAL_BACKUP_TARGET=s3 S3_URI=s3://bucket/powerfulcontrolsystem bash deploy/scripts/vps-external-backup.sh
```

## Archivos principales

- `deploy/docker-compose.platform.yml`: stack principal.
- `deploy/.env.platform.example`: plantilla de variables y secretos.
- `deploy/docker/*.Dockerfile`: imagenes propias de backend, frontend y voz IA.
- `deploy/nginx/pcs.conf`: Nginx interno del contenedor frontend.
- `deploy/postgres/init/01-create-databases.sql`: crea `pcs_superadministrador` y `pcs_empresas` en el primer arranque del volumen PostgreSQL.
- `deploy/scripts/vps-docker-preflight.sh`: valida Docker, crea `deploy/.env.platform` si falta y comprueba el Compose.
- `deploy/scripts/vps-compose-sidecar-up.sh`: levanta el stack en paralelo sin cambiar el Nginx publico.
- `deploy/scripts/vps-postgres-migrate-to-volume.sh`: migra las bases actuales al volumen Docker PostgreSQL, con confirmacion explicita.
- `deploy/scripts/vps-migrate-files-to-volumes.sh`: copia uploads, descargas, logs y respaldos actuales a volumenes Docker.
- `deploy/scripts/vps-cutover-docker-nginx.sh`: cambia el upstream de Nginx de `127.0.0.1:8080` a Docker, con backup y confirmacion explicita.

## Flujo seguro en la VPS actual

Desde `/root/powerfulcontrolsystem`:

```bash
bash deploy/scripts/vps-docker-preflight.sh
bash deploy/scripts/vps-compose-sidecar-up.sh
```

Ese arranque levanta el nucleo `postgres`, `backend` y `frontend`. En la VPS actual, OnlyOffice y Nextcloud ya existen como contenedores y RustDesk ya usa los puertos publicos del host; por eso esos servicios quedan definidos en perfiles para no causar colisiones durante la migracion.

El wrapper local `scripts/sync_to_vps.ps1` reconstruye este stack y, por defecto, limpia temporales antiguos de sincronizacion y cache Docker no usado al final del despliegue. La limpieza no ejecuta `docker volume prune` ni elimina datos persistentes. Para desactivarla temporalmente:

```powershell
.\scripts\sync_to_vps.ps1 -CleanupRemoteUnusedFiles:$false
```

Antes de conmutar el dominio publico, migra las bases actuales al PostgreSQL de Docker:

```bash
CONFIRM_MIGRATE=YES bash deploy/scripts/vps-postgres-migrate-to-volume.sh
CONFIRM_FILE_MIGRATE=YES bash deploy/scripts/vps-migrate-files-to-volumes.sh
bash deploy/scripts/vps-compose-sidecar-up.sh
curl -I http://127.0.0.1:8081/
```

Cuando `127.0.0.1:8081` responda bien y los datos se vean correctos, conmuta Nginx:

```bash
CONFIRM_CUTOVER=YES bash deploy/scripts/vps-cutover-docker-nginx.sh
curl -I https://powerfulcontrolsystem.com
```

El script de conmutacion deja `powerfulcontrolsystem.service` activo para rollback rapido. Si necesitas volver al backend antiguo, restaura el backup indicado por el script o cambia el upstream de Nginx de nuevo a `127.0.0.1:8080` y recarga Nginx.

## Servicios incluidos

- `postgres`: PostgreSQL 16 para `pcs_superadministrador` y `pcs_empresas`.
- `backend`: API Go de la plataforma.
- `frontend`: Nginx con archivos de `web` y proxy interno al backend.
- `onlyoffice-documentserver`: OnlyOffice con JWT.
- `nextcloud-db`, `nextcloud-redis`, `nextcloud`: stack Nextcloud.
- `voice-stream`: servicio FastAPI/Piper para voz IA.
- `rustdesk-hbbs`, `rustdesk-hbbr`: relay/ID server RustDesk.

Todos los servicios comparten la red Docker `pcs_internal` y usan volumenes nombrados para persistencia.

Perfiles opcionales:

```bash
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile office up -d
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile cloud up -d
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile voice up -d
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile rustdesk up -d
```

Activa esos perfiles solo despues de apagar o migrar los servicios antiguos que ya ocupan los mismos nombres o puertos.

## Staging

Para probar cambios antes de produccion:

```powershell
.\scripts\staging_up.ps1 -ConfigOnly
.\scripts\staging_up.ps1 -Build
```

El override `deploy/docker-compose.staging.yml` cambia nombre de proyecto, contenedores, puerto `8082` y volumenes persistentes para no mezclar datos con produccion. En VPS se puede usar:

```bash
bash deploy/scripts/vps-staging-up.sh
```

## Migracion futura a otro servidor

Para mover la plataforma despues:

- Exporta o publica las imagenes `pcs-backend`, `pcs-frontend` y `pcs-voice-stream`.
- Migra los volumenes `pcs_postgres_data`, `pcs_web_uploads`, `pcs_downloads`, `pcs_nextcloud_*`, `pcs_onlyoffice_*`, `pcs_voice_*`, `pcs_rustdesk_data`.
- Conserva `deploy/.env.platform` con los mismos secretos.
- En el nuevo servidor, restaura volumenes, copia el proyecto o las imagenes y ejecuta `docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d`.

No subas `deploy/.env.platform` al repositorio: contiene secretos operativos.
