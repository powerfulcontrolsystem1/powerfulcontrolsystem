# Estado final de preproduccion

Fecha de corte: 2026-07-13. Este documento separa evidencia ejecutada de
validaciones que requieren un entorno efimero. No autoriza despliegues ni
sustituye los gates de CI, staging y aprobacion humana.

## Decision

**NO APROBADO PARA PRODUCCION TODAVIA.** El codigo queda preparado para la
siguiente validacion controlada, pero faltan evidencias obligatorias fuera de
esta estacion: Compose/Linux, restauracion, carga, proveedores sandbox y
firmas de responsable.

## Correcciones verificadas localmente

- El panel empresarial envia el token CSRF en mutaciones same-origin. El estado
  `no_mostrar_mas` de Configuracion guiada se conserva por empresa en
  PostgreSQL y ya no depende de cookies del navegador.
- Nextcloud resuelve la empresa desde el shell autenticado, no desde un valor
  de URL como autoridad. Las cuentas inactivas o no aprovisionadas no reciben
  URL de acceso ni WebDAV.
- La cuota inicial de Nextcloud se toma de la configuracion global; ya no se
  sobrescribe con 1024 MB en cada arranque.
- Se retiró el servidor directo del backend para `/descargas/`. Nginx solo
  expone una lista cerrada de instaladores RustDesk; los documentos de empresa
  deben continuar por handlers autenticados.

## Evidencia ejecutada en esta estacion

| Prueba | Resultado |
| --- | --- |
| `go mod verify` | Correcta en `backend/`. |
| `go test ./...` | Correcta antes de estas correcciones. |
| Prueba estatica de CSRF, Nextcloud y descargas | Correcta. |
| Pruebas de Configuracion guiada y Nextcloud | Correctas. |
| Pruebas CSRF del middleware | Correctas. |
| `git diff --check` | Correcta al cierre de la edicion. |

## Bloqueantes de aprobacion

1. Ejecutar CI/Linux con `go test -race`, Docker Compose, Trivy, SBOM y
   escaneo de secretos. Esta estacion no tiene Docker ni compilador C; el
   detector de carrera no puede ser sustituido por pruebas normales.
2. Levantar `deploy/docker-compose.staging.yml` con datos anonimos, ejecutar
   migraciones dos veces y validar rollback contra PostgreSQL efimero.
3. Restaurar una copia anonima en entorno efimero y medir RPO/RTO.
4. Probar ePayco, Wompi, Bre-B, DIAN, WhatsApp, correo, Nextcloud y OnlyOffice
   exclusivamente con sandbox o cuentas de prueba autorizadas.
5. Ejecutar aislamiento multiempresa de extremo a extremo en PostgreSQL y
   pruebas de archivo/symlink en Linux.
6. Ejecutar carga, WebRTC y soporte remoto en staging; registrar limites y
   alertas antes de habilitar trafico real.

Los reportes relacionados documentan los comandos y la evidencia requerida.
