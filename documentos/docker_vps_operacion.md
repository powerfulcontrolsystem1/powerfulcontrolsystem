# Docker en VPS - Operacion y migracion

Fecha de actualizacion: 2026-05-19

## Paquete portable desde Super Administrador

El panel `Super Administrador > Plataforma > Docker VPS` permite revisar el estado del paquete Docker portable y descargar un `.tar.gz` del proyecto base.

La descarga usa:

```text
GET /super/api/docker_portabilidad?action=status
GET /super/api/docker_portabilidad?action=download
```

El endpoint exige rol `super_administrador`. El paquete incluye codigo, `deploy/docker-compose.platform.yml`, Dockerfiles, scripts y documentacion operativa, pero no incluye secretos ni datos runtime:

- No incluye `deploy/.env.platform`, `backend/.env*`, llaves privadas ni `.env` reales.
- No incluye `web/uploads`, `descargas`, backups, logs, caches, binarios ni evidencias QA.
- No incluye dumps de PostgreSQL ni volumenes Docker.

Para una migracion real se debe descargar este paquete y combinarlo con el snapshot operativo de PostgreSQL/volumenes generado por `scripts/vps_backup_operacion.ps1`.

## Estado actual

La VPS actual (`2.24.197.58`) ya tiene Docker Engine y Docker Compose v2. El nucleo de la plataforma quedo ejecutandose en Docker y publicado por Nginx del host:

- Nginx publico del host: `80/443`.
- Frontend Docker interno: `127.0.0.1:8081`.
- Backend Docker interno: `pcs-backend:8080` dentro de la red Docker.
- PostgreSQL Docker: `pcs-postgres:5432` dentro de la red Docker.
- Red Docker: `pcs_internal`.

Actualizacion 2026-05-12: el despliegue ya incluye modo `edge` para que tambien el frente publico `80/443`, TLS y certificados Let's Encrypt queden bajo Docker con `pcs-edge` y `pcs-certbot`. El Nginx del host queda solo como fase de transicion o rollback.

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

## Limpieza automatica durante sync

`scripts/sync_to_vps.ps1` ejecuta limpieza remota segura despues de reconstruir Docker:

- Borra `/tmp/pcs_sync_*.tar.gz` antiguos.
- Borra caches no persistentes del proyecto (`.gotmp`, `.gocache`, `tmp`, caches de pruebas Go).
- Ejecuta `docker container prune` solo para contenedores detenidos antiguos.
- Ejecuta `docker image prune` solo para imagenes dangling.
- Ejecuta `docker builder prune` para cache BuildKit no usado.

No toca volumenes Docker ni bases de datos. Para omitirla en un despliegue puntual:

```powershell
.\scripts\sync_to_vps.ps1 -CleanupRemoteUnusedFiles:$false
```

## Backups operativos

Para crear un snapshot operativo de PostgreSQL y volumenes persistentes:

```powershell
.\scripts\vps_backup_operacion.ps1
```

El respaldo queda en:

```bash
/root/powerfulcontrolsystem/backups/vps-snapshots/<fecha>
```

Incluye `pg_dumpall` comprimido y tarballs de volumenes de uploads, descargas, logs, backups y datos PostgreSQL. Por defecto no copia `deploy/.env.platform`; si se necesita respaldar secretos en la VPS, usar `-IncludeEnvSecrets` y conservar permisos privados.

Para validar que el backup se puede restaurar:

```powershell
.\scripts\vps_restore_validation.ps1
```

Para ejecutar una restauracion real en un contenedor temporal de PostgreSQL:

```powershell
.\scripts\vps_restore_validation.ps1 -ExecuteDrill
```

## Staging Docker

Staging usa el mismo Compose base con override aislado:

```powershell
.\scripts\staging_up.ps1 -ConfigOnly
.\scripts\staging_up.ps1 -Build
```

En VPS:

```bash
bash deploy/scripts/vps-staging-up.sh
```

Usa puerto `8082`, volumenes `pcs_staging_*` y variables `deploy/.env.staging`.

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

## Edge Docker 80/443

Para dejar la plataforma completa bajo Docker:

```bash
cd /root/powerfulcontrolsystem
bash deploy/scripts/vps-compose-sidecar-up.sh
CONFIRM_DOCKER_EDGE=YES bash deploy/scripts/vps-docker-edge-up.sh
```

El script:

- Levanta `postgres`, `backend` y `frontend`.
- Detiene Nginx del host para liberar `80/443`.
- Inicia `pcs-edge` temporal en HTTP para ACME.
- Emite certificado Let's Encrypt con `pcs-certbot`.
- Recrea `pcs-edge` con HTTPS y proxy hacia `pcs-frontend`.

Renovacion:

```bash
bash deploy/scripts/vps-docker-edge-renew.sh
```

Cron sugerido:

```bash
0 4 * * * cd /root/powerfulcontrolsystem && bash deploy/scripts/vps-docker-edge-renew.sh >/var/log/pcs-edge-renew.log 2>&1
```

Rollback del edge:

```bash
docker stop pcs-edge
sudo systemctl start nginx
sudo nginx -t
```

## Servicios definidos como perfiles

OnlyOffice, edge publico, voz IA y RustDesk estan definidos en el Compose por perfiles. Nextcloud queda retirado del producto y del Compose oficial:

- OnlyOffice ya existia como contenedor en `127.0.0.1:8088`.
- RustDesk ya usaba puertos publicos `21115`, `21116` y `21117` en el host.
- Si quedan contenedores Nextcloud legacy, retirar con `deploy/scripts/vps-remove-nextcloud.sh`; usar `--purge-data` solo despues de confirmar que no se requiere recuperacion.

Perfiles disponibles:

```bash
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile office up -d
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile edge up -d
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

## Sincronizacion desde Windows/local

El script principal de despliegue quedo adaptado al runtime Docker de produccion:

```powershell
.\scripts\sync_to_vps.ps1
```

Comportamiento por defecto:

- Usa `DeploymentMode=docker`.
- Sincroniza el proyecto al VPS.
- Omite la compilacion local del binario Go, porque Docker construye el backend dentro del contenedor.
- No reinicia `powerfulcontrolsystem.service` por `systemd`.
- Reconstruye y levanta `pcs-backend` y `pcs-frontend` con Docker Compose.
- Espera a que `pcs-backend` quede `healthy`, `pcs-frontend` quede `running` y `127.0.0.1:8081` responda.
- Excluye evidencias de QA, temporales, logs, backups, llaves y secretos del paquete de produccion.

Modos disponibles:

```powershell
.\scripts\sync_to_vps.ps1 -DeploymentMode docker
.\scripts\sync_to_vps.ps1 -DeploymentMode hybrid
.\scripts\sync_to_vps.ps1 -DeploymentMode legacy
```

- `docker`: modo recomendado actual. Solo Docker Compose.
- `hybrid`: mantiene compatibilidad temporal, actualiza `systemd` y luego Docker Compose.
- `legacy`: solo binario + `systemd`, sin Docker Compose.

Para probar sin tocar el VPS:

```powershell
.\scripts\sync_to_vps.ps1 -DryRun
.\scripts\sync_to_vps.ps1 -PreviewOnly
```

Si se necesita incluir evidencias de QA en una sincronizacion puntual:

```powershell
.\scripts\sync_to_vps.ps1 -ExcludeEvidenceFromPackage:$false
```

El contexto Docker tambien excluye `documentos/evidencias_qa`, `test_runs`, llaves y archivos temporales mediante `.dockerignore`, para evitar builds pesados o filtracion accidental de artefactos locales.

## Faltantes controlados

No falta nada para que el nucleo publico funcione por Docker. Quedan tareas recomendadas para cerrar el ciclo profesional:

- Definir si `powerfulcontrolsystem.service` se deja como rollback temporal o se deshabilita tras varios dias de estabilidad.
- Decidir si OnlyOffice y RustDesk se migran al Compose unificado o se mantienen como servicios separados.
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
- `pcs_onlyoffice_*`
- `pcs_voice_*`
- datos iRedMail si se activo perfil `mail`: `IREDMAIL_DATA_DIR` y
  `IREDMAIL_BACKUP_DIR`, por defecto `/opt/powerfulcontrolsystem/iredmail/*`
- `pcs_rustdesk_data`
4. Levantar:

```bash
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d
```

5. Configurar Nginx del nuevo host para apuntar a `127.0.0.1:8081`.

## Email corporativo iRedMail portable

El correo empresarial se integra al stack portable mediante variables en
`deploy/.env.platform` y perfil Docker opcional `mail`.

- El backend recibe `EMAIL_CORPORATIVO_*` e `IREDADMIN_*` desde el compose.
- Al arrancar, registra esas variables en `pcs_superadministrador.configuraciones`.
- La clave `EMAIL_CORPORATIVO_IREDADMIN_PASSWORD` se guarda cifrada usando
  `CONFIG_ENC_KEY`; no se escribe en logs ni documentacion.
- El perfil `mail` publica SMTP/IMAP/POP3 y deja el webmail disponible para el
  edge Docker en `EDGE_MAIL_DOMAIN`.
- `deploy/scripts/vps-configure-iredmail-host-nginx.sh` valida el certificado
  del subdominio `mail`; si el certificado existente no lo cubre, usa certbot
  con webroot en `/var/www/html` y luego recarga Nginx con el proxy HTTPS hacia
  iRedMail.
- Antes de activar `--profile mail`, validar DNS A, MX, SPF, DKIM, DMARC, PTR,
  TLS y que la imagen definida en `IREDMAIL_IMAGE` sea la aprobada para la VPS.

Arranque futuro cuando el correo este listo:

```bash
deploy/scripts/vps-register-iredmail-secrets.sh
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile mail up -d iredmail
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml --profile edge up -d edge
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d --build backend
```
