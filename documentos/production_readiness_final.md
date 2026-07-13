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
| `go mod verify` | Correcta el 2026-07-13. |
| `go test ./...` y `go vet ./...` | Correctas el 2026-07-13. |
| Preflight profesional completo | Correcto el 2026-07-13. |
| `go test -race ./...` | No ejecutable en esta estacion: CGO esta deshabilitado y no hay compilador C. Gate obligatorio de CI/Linux. |
| `govulncheck -show verbose ./...` | Sin vulnerabilidades alcanzables; permanece un aviso informativo de `openpgp` no alcanzable y sin correccion publicada. |
| `gosec ./...` | Inconcluso localmente: se detuvo sin resultado final tras superar el tiempo operativo y 400 MB de memoria. Gate obligatorio de CI/Linux. |

## Controles operativos corregidos

- Los scripts de backup, restauracion y release gate ya no contactan una VPS
  remota por defecto. Requieren `-AllowRemoteTarget` luego de validar que el
  destino es staging aislado o una operacion expresamente autorizada.
- Cuando `main` esta protegida, `rs` crea una PR, solicita auto-merge y espera
  las verificaciones y aprobacion independiente. No sincroniza la VPS antes de
  que GitHub confirme la fusion.

## Evidencia publicada adicional

- En la empresa interna PCS, el asistente IA respondio una consulta de contexto
  empresarial con el nombre de la empresa, sin exponer su identificador ni
  proponer acciones no solicitadas.
- La cuenta empresarial de Nextcloud de PCS quedo aprovisionada visualmente con
  cuota de 1024 MB y la interfaz habilito su apertura sin entregar credenciales
  persistentes en la pagina.
- Se corrigio la ruta de configuracion de Seguridad VPS para el usuario no root
  del backend. La confirmacion visual publicada de esta pantalla se debe hacer
  despues del siguiente `rs`.

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
