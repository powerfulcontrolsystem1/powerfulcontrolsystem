# Plan de profesionalizacion de la plataforma

Fecha: 2026-05-11

Este plan convierte las siete recomendaciones operativas en controles ejecutables del proyecto. La meta es que cada cambio pueda validarse antes de publicarse y que la VPS conserve respaldo, limpieza y trazabilidad.

## 1. Estabilidad y QA continuo

- Script base: `scripts/profesional_preflight.ps1`.
- Auditoria tecnica: `tools/professional_audit.mjs`.
- Integracion: `scripts/rs.ps1` ejecuta preflight antes de actualizar repositorio y sincronizar VPS, salvo que se use `-SkipPreflight`.
- Modo rapido: sintaxis PowerShell, sintaxis JavaScript, auditoria de modulos/permisos/portal, Docker Compose si esta disponible y `git diff --check`.
- Modo completo: `.\scripts\profesional_preflight.ps1 -Full` agrega `go test ./...` en `backend`.

## 2. Gobierno Docker/VPS

- `scripts/sync_to_vps.ps1` incluye limpieza remota segura por defecto.
- Limpia temporales de sync, caches locales, contenedores detenidos, imagenes dangling y cache BuildKit.
- No ejecuta `docker volume prune` ni elimina bases de datos, uploads, descargas o backups persistentes.
- Respaldo operativo: `scripts/vps_backup_operacion.ps1` genera dump PostgreSQL y empaqueta volumenes persistentes.

## 3. Permisos y licencias

- `tools/professional_audit.mjs` inventaria modulos de permisos backend, wrappers `WithEmpresa*Permissions`, enlaces del menu empresarial y catalogo de 19 plantillas.
- El reporte queda en `documentos/reportes_profesionales/`.
- La auditoria marca advertencias sin bloquear por defecto; con `-Strict` en preflight puede usarse como compuerta dura.

## 4. Experiencia de usuario

- El preflight valida sintaxis de scripts inline HTML y archivos `web/js`.
- Las capturas y evidencias visuales siguen guardandose en `documentos/evidencias_qa/`.
- Las nuevas pantallas deben conservar submenus, estados vacios, mensajes de error y controles consistentes con `web/estilos.css`.

## 5. Datos y preconfiguraciones

- La auditoria exige que las 19 plantillas tengan `module`, `title`, `fullTitle`, descripcion larga y al menos cinco secciones.
- Las preconfiguraciones inteligentes se validan por pruebas Go existentes en `backend/db/tipo_empresa_preconfiguracion_test.go`.
- El modo completo del preflight ejecuta las pruebas de backend para detectar regresiones.

## 6. Pagos, facturacion e impresion

- El preflight cubre contratos estaticos y pruebas backend si se ejecuta en modo completo.
- Las pruebas visuales de facturas, tickets y comprobantes deben seguir dejando evidencia en `documentos/evidencias_qa/`.
- Los cambios en Wompi/Epayco o licencias deben actualizar `documentos/gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md`.

## 7. Documentacion comercial y operativa

- Documentos obligatorios: `manual_de_instalacion.md`, `docker_vps_operacion.md`, `deploy/README-compose-platform.md`, `matriz_roles_permisos_pos_multiempresa.md`, `descripcion_de_modulos`, `descripcion_del_proyecto` y `CHANGELOG.md`.
- `tools/professional_audit.mjs` verifica que existan y los reporta.
- Todo cambio transversal debe registrar resumen en `documentos/CHANGELOG.md` e `historial_de_cambios`.

## Comandos recomendados

```powershell
.\scripts\profesional_preflight.ps1
.\scripts\profesional_preflight.ps1 -Full
.\scripts\vps_backup_operacion.ps1
.\rs.ps1
```

## Comandos de emergencia

```powershell
.\rs.ps1 -SkipPreflight
.\scripts\sync_to_vps.ps1 -CleanupRemoteUnusedFiles:$false
.\scripts\vps_backup_operacion.ps1 -DryRun
```

## Madurez empresarial adicional

El bloque de 12 pasos profesionales queda materializado en `documentos/gobernanza_tecnica/runbooks/runbook_madurez_empresarial_12_pasos.md`.

Controles nuevos:

- Anonimizacion automatica de staging despues de refrescar desde produccion.
- Monitoreo Prometheus/Grafana desplegable en VPS.
- Backup externo por rclone o S3.
- Deploy automatico opcional a staging desde GitHub Actions.
- Auditorias de roles, pagos, soporte y normalizacion documental.
- Prueba de carga smoke incluida en release gate.
- Manifiesto formal de release por version.
