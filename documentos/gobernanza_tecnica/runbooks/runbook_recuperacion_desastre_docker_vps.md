# Runbook de recuperacion ante desastre Docker/VPS

## Objetivo

Levantar Powerful Control System en un VPS nuevo usando imagenes Docker, volumenes persistentes, variables de entorno y backups PostgreSQL.

## Insumos obligatorios

- Repositorio actualizado.
- `deploy/.env.platform` privado.
- Ultimo snapshot de `/root/powerfulcontrolsystem/backups/vps-snapshots`.
- Imagenes Docker publicadas o capacidad de construirlas desde el repo.
- Acceso DNS del dominio principal y subdominios.

## Preparacion del nuevo VPS

```bash
apt update
apt install -y docker.io docker-compose-plugin curl
systemctl enable --now docker
```

## Restauracion

1. Copiar el proyecto a `/root/powerfulcontrolsystem`.
2. Copiar `deploy/.env.platform` con permisos `600`.
3. Restaurar volumenes Docker desde los tarballs del snapshot.
4. Levantar PostgreSQL y restaurar `postgres_all.sql.gz`.
5. Ejecutar:

```bash
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml up -d --build
```

Para publicar tambien `80/443` desde Docker:

```bash
CONFIRM_DOCKER_EDGE=YES bash deploy/scripts/vps-docker-edge-up.sh
```

## Verificacion

```bash
docker ps
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps
curl -I http://127.0.0.1:8081/
curl -I https://powerfulcontrolsystem.com
```

Verificar que `pcs-edge` este activo cuando el VPS nuevo opere sin Nginx del host:

```bash
docker inspect -f '{{.State.Status}}' pcs-edge
```

## Prueba de restauracion periodica

Desde Windows/local:

```powershell
.\scripts\vps_restore_validation.ps1
.\scripts\vps_restore_validation.ps1 -ExecuteDrill
```

La primera validacion no escribe datos. La segunda restaura en un contenedor PostgreSQL temporal y lo elimina al finalizar.

## Rollback

Si el nuevo VPS no queda funcional, mantener DNS apuntando al VPS anterior. No eliminar backups ni volumenes del servidor anterior hasta completar una prueba funcional de login, licencias, facturacion, archivos subidos y panel super.
