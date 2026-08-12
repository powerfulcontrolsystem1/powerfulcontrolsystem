# P110-008 - Candidato final y restore efímero

Fecha: 2026-08-11

Ambiente: staging aislado.

Producción: no modificada.

## Promoción

El SHA `fd6a4a8a18b44100d423cdae07db4d957e69da97` generó en CI imágenes API,
migrador, worker y frontend con referencias exactas por digest. El script de
promoción verificó el Compose, descargó solo esas referencias y recreó
`migrate`, `clamav`, `backend`, `worker` y `frontend` sin compilación remota.

La inspección posterior confirmó que backend y worker comparten únicamente los
volúmenes de staging para uploads, almacenamiento privado, backups y logs;
frontend e init-container usan el mismo volumen de uploads aislado. El
migrador usa el almacenamiento privado de staging. `/health` y `/ready`
respondieron `200`.

## Restore

El runner recibió explícitamente el entorno privado y el snapshot de staging
fuera del checkout desechable. Creó PostgreSQL, red, API, archivos temporales y
puerto efímeros; todos se eliminaron al terminar.

Resultado observado:

```text
health=200 ready=200 bases=2 tablas=5 filas_empresa_12=32
archivos_privados=6 referencias_privadas=6 huerfanos_privados=0
referencias_heredadas=0 RTO=86s RPO=53909s runtime_privilegios=0
```

## Límite

Faltan la réplica autenticada A/B, pérdida de réplica, rollback coordinado de
datos/aplicación y la aceptación operativa del RPO/RTO. P110-008 permanece
**parcial**; no autoriza promoción a producción.

## Repetición A/B y rollback del candidato `9ee2fe9a`

El candidato posterior incorporó ClamAV por digest dentro de la red efímera;
no se redujo `PCS_SUPPORTS_CLAMAV_REQUIRED` para el ensayo. La sesión oficial
de PCS se utilizó únicamente en la copia restaurada y se eliminó al finalizar.

```text
bases=2 tablas=5 filas_empresa_12=32 endpoints_protegidos=4
dominios_autenticados=5 replica_checks=2 archivos_hostiles=5
archivos_privados=6 referencias_privadas=6 huerfanos_privados=0
referencias_heredadas=0 rollback_checks=7 rollback_dominios=5
rollback_RTO=98s RTO=226s RPO=59739s runtime_privilegios=0
```

El ensayo creó el documento de prueba en réplica A, lo descargó con checksum
igual en B, retiró A y repitió la lectura en B. Después congeló bases y
almacenamiento privado, simuló su pérdida total y restauró ambos de manera
coordinada. La inspección posterior no encontró contenedores efímeros; staging
continuó con health, readiness y ClamAV saludables.

La fase sigue **parcial** hasta que los objetivos RPO/RTO sean aceptados y
firmados y se completen las demás compuertas de P110.
