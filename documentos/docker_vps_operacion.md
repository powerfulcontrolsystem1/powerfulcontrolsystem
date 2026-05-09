# Docker en VPS - Operacion y migracion

Fecha de actualizacion: 2026-05-09

## Estado actual

La VPS actual (`2.24.197.58`) ya tiene Docker Engine y Docker Compose v2. El nucleo de la plataforma quedo ejecutandose en Docker y publicado por Nginx del host:

- Nginx publico del host: `80/443`.
- Frontend Docker interno: `127.0.0.1:8081`.
- Backend Docker interno: `pcs-backend:8080` dentro de la red Docker.
- PostgreSQL Docker: `pcs-postgres:5432` dentro de la red Docker.
- Red Docker: `pcs_internal`.

El archivo activo de Compose es:

```bash
/root/powerfulcontrolsystem/deploy/docker-compose.platform.yml
```

El archivo de entorno real de la VPS es:

```bash
/root/powerfulcontrolsystem/deploy/.env.platform
```

No debe subirse al repositorio porque contiene secretos.

## Servicios activos por Docker

Arranque base:

- `pcs-postgres`: PostgreSQL 16 con las bases `pcs_superadministrador` y `pcs_empresas`.
- `pcs-backend`: API Go de la plataforma.
- `pcs-frontend`: Nginx interno que sirve `web` y reenvia API al backend.

Comando de verificacion:

```bash
cd /root/powerfulcontrolsystem
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps
curl -I http://127.0.0.1:8081/
curl -I https://powerfulcontrolsystem.com
```

## Datos migrados a volumenes

Se migraron las dos bases actuales al volumen PostgreSQL Docker. Tambien se copiaron archivos persistentes:

- `web/uploads` -> `pcs_web_uploads`
- `descargas` -> `pcs_downloads`
- `backend/logs` -> `pcs_backend_logs`
- `backup` y `backups` -> `pcs_backups`

Los dumps de migracion quedan en:

```bash
/root/powerfulcontrolsystem/backups/docker-migration/
```

## Nginx y rollback

Nginx del host fue conmutado de `127.0.0.1:8080` a `127.0.0.1:8081`.

Backup creado durante la conmutacion:

```bash
/etc/nginx/sites-available/powerfulcontrolsystem.bak.20260509-193744
```

Rollback rapido:

```bash
sudo cp /etc/nginx/sites-available/powerfulcontrolsystem.bak.20260509-193744 /etc/nginx/sites-available/powerfulcontrolsystem
sudo nginx -t
sudo systemctl reload nginx
```

El servicio anterior `powerfulcontrolsystem.service` quedo activo para rollback operativo rapido mientras se estabiliza Docker. Cuando se confirme estabilidad por varios dias, puede evaluarse pausarlo o deshabilitarlo.

## Servicios definidos como perfiles

OnlyOffice, Nextcloud, voz IA y RustDesk estan definidos en el Compose, pero no se levantan por defecto para evitar colisiones con servicios ya existentes en la VPS:

- OnlyOffice ya existia como contenedor en `127.0.0.1:8088`.
- Nextcloud ya existia como contenedores en `127.0.0.1:8090`.
- RustDesk ya usaba puertos publicos `21115`, `21116` y `21117` en el host.

Perfiles disponibles:

```bash
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile office up -d
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile cloud up -d
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile voice up -d
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile rustdesk up -d
```

Activalos solo despues de detener o migrar los servicios antiguos correspondientes.

## Scripts operativos

Desde `/root/powerfulcontrolsystem`:

```bash
bash deploy/scripts/vps-docker-preflight.sh
bash deploy/scripts/vps-compose-sidecar-up.sh
CONFIRM_MIGRATE=YES bash deploy/scripts/vps-postgres-migrate-to-volume.sh
CONFIRM_FILE_MIGRATE=YES bash deploy/scripts/vps-migrate-files-to-volumes.sh
CONFIRM_CUTOVER=YES bash deploy/scripts/vps-cutover-docker-nginx.sh
```

Los scripts usan confirmaciones explicitas para operaciones sensibles y no imprimen secretos.

## Faltantes controlados

No falta nada para que el nucleo publico funcione por Docker. Quedan tareas recomendadas para cerrar el ciclo profesional:

- Definir si `powerfulcontrolsystem.service` se deja como rollback temporal o se deshabilita tras varios dias de estabilidad.
- Decidir si OnlyOffice, Nextcloud y RustDesk se migran al Compose unificado o se mantienen como servicios separados.
- Publicar imagenes `pcs-backend`, `pcs-frontend` y `pcs-voice-stream` en un registry privado si se quiere mover la VPS sin reconstruir.
- Programar backups periodicos de volumenes Docker y dumps PostgreSQL.
- Documentar el procedimiento exacto de restauracion en servidor nuevo con volumenes e imagenes.

## Migracion futura a nuevo servidor

Para mover a otra VPS:

1. Copiar el repositorio o llevar las imagenes publicadas.
2. Copiar `deploy/.env.platform` de forma segura.
3. Migrar volumenes Docker:
   - `pcs_postgres_data`
   - `pcs_web_uploads`
   - `pcs_downloads`
   - `pcs_backend_logs`
   - `pcs_backups`
   - `pcs_nextcloud_*`
   - `pcs_onlyoffice_*`
   - `pcs_voice_*`
   - `pcs_rustdesk_data`
4. Levantar:

```bash
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d
```

5. Configurar Nginx del nuevo host para apuntar a `127.0.0.1:8081`.
