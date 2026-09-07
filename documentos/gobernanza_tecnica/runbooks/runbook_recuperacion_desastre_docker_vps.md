# Runbook de recuperacion ante desastre Docker/VPS

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Existen dos formatos: vps-backup-operacion.sh produce postgres_all.sql.gz; el paquete del panel produce postgres/pg_dumpall.sql mediante pg_dump de ambas bases. Elegir el procedimiento según el manifiesto del backup, no por nombres parecidos.
- Inventariar contenedores por servicio/proyecto: los nombres fijos son ejemplos históricos y las réplicas pueden usar nombres Compose. El paquete sin dump no acredita recuperación de BD.
- Restaurar en aislamiento y migrar al esquema del candidato antes de abrir API/worker. El cambio DNS y la retirada del host anterior requieren validación funcional y autoridad de operación.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Objetivo

Levantar Powerful Control System en un VPS nuevo usando imagenes Docker, volumenes persistentes, variables de entorno y backups PostgreSQL.

## Insumos obligatorios

- Repositorio actualizado.
- Paquete `.tar.gz` descargado desde `Super Administrador > Plataforma > Docker VPS`, si no se va a clonar el repositorio directamente.
- `deploy/.env.platform` privado.
- Ultimo snapshot de `/root/powerfulcontrolsystem/backups/vps-snapshots`.
- Replica verificada fuera del VPS (medio local independiente o remoto cifrado
  de nube) y su `SHA256SUMS`.
- Imagenes Docker publicadas o capacidad de construirlas desde el repo.
- Acceso DNS del dominio principal y subdominios.

## Preparacion del nuevo VPS

```bash
apt update
apt install -y docker.io docker-compose-plugin curl
systemctl enable --now docker
```

## Restauracion

1. Copiar el proyecto a `/root/powerfulcontrolsystem` desde repositorio o desde el paquete portable Docker descargado en el panel super.
2. Copiar `deploy/.env.platform` con permisos `600`.
3. Restaurar volumenes Docker desde los tarballs del snapshot, incluyendo
   almacenamiento privado, certificados de Mailu y datos persistentes de
   OnlyOffice cuando correspondan.
4. Levantar PostgreSQL y restaurar `postgres_all.sql.gz`. No restaurar un
   tarball fisico de PostgreSQL tomado en caliente: el dump logico es la fuente
   consistente.
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

Antes de retirar el servidor anterior, ejecutar `sha256sum -c SHA256SUMS` sobre
la copia recuperada. Un registro de subida sin `rclone check --checksum` o una
copia que solo existe en el mismo VPS no satisface el requisito de continuidad.

## Prueba de restauracion periodica

Desde Windows/local:

```powershell
.\scripts\vps_restore_validation.ps1
.\scripts\vps_restore_validation.ps1 -ExecuteDrill
```

La primera validacion no escribe datos. La segunda restaura en un contenedor PostgreSQL temporal y lo elimina al finalizar; ambas registran RPO y RTO en segundos. Conservar esa evidencia junto con SHA y fecha del snapshot.

## Rollback

Si el nuevo VPS no queda funcional, mantener DNS apuntando al VPS anterior. No eliminar backups ni volumenes del servidor anterior hasta completar una prueba funcional de login, licencias, facturacion, archivos subidos y panel super.

## Fuentes y aceptación de la revisión

[vps-backup-operacion.sh](../../../deploy/scripts/vps-backup-operacion.sh), [super_vps_snapshots.go](../../../backend/handlers/super_vps_snapshots.go), [vps_restore_validation.ps1](../../../scripts/vps_restore_validation.ps1), [incidentes_y_continuidad.md](../../operacion/incidentes_y_continuidad.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../requisitos/especificacion_y_trazabilidad.md)).
