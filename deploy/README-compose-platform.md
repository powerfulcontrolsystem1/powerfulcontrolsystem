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
- `deploy/scripts/vps-docker-edge-up.sh`: mueve el frente publico `80/443` al contenedor `pcs-edge` con Nginx y certificados Let's Encrypt.
- `deploy/scripts/vps-docker-edge-renew.sh`: renueva certificados desde Docker/Certbot y recarga `pcs-edge`.

## Flujo seguro en la VPS actual

Desde `/root/powerfulcontrolsystem`:

```bash
bash deploy/scripts/vps-docker-preflight.sh
bash deploy/scripts/vps-compose-sidecar-up.sh
```

Ese arranque levanta el nucleo `postgres`, `backend` y `frontend`. En la VPS actual, OnlyOffice puede operar como contenedor separado y RustDesk ya usa los puertos publicos del host; por eso esos servicios quedan definidos en perfiles para no causar colisiones durante la migracion. Nextcloud empresarial se levanta aparte con `deploy/scripts/vps-nextcloud-up.sh`.

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

## Modo 100% Docker en la VPS

Cuando el stack interno ya este validado, puedes mover tambien el Nginx publico y TLS a Docker. Este modo deja la aplicacion, PostgreSQL, frontend, certificados, OnlyOffice opcional, voz IA opcional y RustDesk opcional dentro de contenedores con volumenes nombrados.

Requisitos antes de ejecutar:

- DNS de `EDGE_DOMAIN` y dominios extra apuntando a la VPS.
- `EDGE_CERT_EMAIL` configurado con un correo real en `deploy/.env.platform`.
- Puertos `80` y `443` libres o autorizacion para detener el Nginx del host.
- Stack interno validado con `bash deploy/scripts/vps-compose-sidecar-up.sh`.

Activacion:

```bash
CONFIRM_DOCKER_EDGE=YES bash deploy/scripts/vps-docker-edge-up.sh
```

Renovacion de certificados:

```bash
bash deploy/scripts/vps-docker-edge-renew.sh
```

Para cron:

```bash
0 4 * * * cd /root/powerfulcontrolsystem && bash deploy/scripts/vps-docker-edge-renew.sh >/var/log/pcs-edge-renew.log 2>&1
```

En este modo, el host conserva solo Docker, firewall/SSH, cron de renovacion si se usa y herramientas basicas de recuperacion. La operacion de la plataforma queda en Compose.

## Servicios incluidos

- `postgres`: PostgreSQL 16 para `pcs_superadministrador` y `pcs_empresas`.
- `backend`: API Go de la plataforma.
- `frontend`: Nginx con archivos de `web` y proxy interno al backend.
- `edge`: Nginx publico Docker para `80/443`, proxy HTTPS y ACME webroot.
- `certbot`: emision/renovacion de certificados Let's Encrypt usando volumen Docker.
- `onlyoffice-documentserver`: OnlyOffice con JWT.
- `nextcloud`, `nextcloud-cron`, `nextcloud-db` y `nextcloud-redis`: stack empresarial independiente definido en `deploy/nextcloud/docker-compose.yml`.
- `voice-stream`: servicio FastAPI/Piper para voz IA.
- `rustdesk-hbbs`, `rustdesk-hbbr`: relay/ID server RustDesk.
- `mailu-*`: perfil opcional `mail` para correo empresarial portable con Mailu
  y webmail SnappyMail. El backend registra `EMAIL_CORPORATIVO_*`/`MAILU_*`
  desde el entorno, cifra secretos con `CONFIG_ENC_KEY` y usa autologin HMAC
  para abrir la bandeja sin pedir clave al usuario.

Todos los servicios comparten la red Docker `pcs_internal` y usan volumenes nombrados para persistencia. La plataforma y sus perfiles oficiales usan PostgreSQL como unico motor relacional; no se debe introducir MariaDB/MySQL en compose, scripts o runbooks operativos del proyecto.

Perfiles opcionales:

```bash
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile office up -d
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile edge up -d
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile mail up -d
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
- Migra los volumenes `pcs_postgres_data`, `pcs_web_uploads`, `pcs_downloads`, `pcs_letsencrypt`, `pcs_certbot_www`, `pcs_onlyoffice_*`, `mailu_*`, `pcs_voice_*`, `pcs_rustdesk_data` y `pcs_nextcloud_*`, junto con la ruta configurada en `NEXTCLOUD_DATA_PATH`.
- Conserva `deploy/.env.platform` con los mismos secretos.
- En el nuevo servidor, restaura volumenes, copia el proyecto o las imagenes y ejecuta `docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile edge up -d`.

No subas `deploy/.env.platform` al repositorio: contiene secretos operativos.

## Nextcloud empresarial

Nextcloud usa un Compose independiente en `deploy/nextcloud/docker-compose.yml`.
Sus secretos se suministran como archivos Docker y los datos empresariales se
montan desde `NEXTCLOUD_DATA_PATH`; no forman parte del arbol publico ni del
repositorio. El arranque validado se realiza con:

```bash
bash deploy/scripts/vps-nextcloud-up.sh
```

No elimine contenedores, volumenes o la ruta de datos sin un backup verificado y
una prueba de restauracion. Consulte `documentos/nextcloud_empresarial.md`.
