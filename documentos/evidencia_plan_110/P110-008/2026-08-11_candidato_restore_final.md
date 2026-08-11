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
