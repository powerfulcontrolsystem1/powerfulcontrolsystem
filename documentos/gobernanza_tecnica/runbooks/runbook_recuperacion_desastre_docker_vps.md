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
apt install -y docker.io docker-compose-plugin nginx certbot python3-certbot-nginx
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

## Verificacion

```bash
docker ps
docker compose --env-file deploy/.env.platform -f deploy/docker-compose.platform.yml ps
curl -I http://127.0.0.1:8081/
curl -I https://powerfulcontrolsystem.com
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
