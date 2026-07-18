# Reporte de ejecucion de staging

Estado: **EJECUCION TECNICA PARCIAL COMPLETADA 2026-07-18**.

Se validó un stack aislado con el prefijo `pcs_staging_`, sin modificar el
stack productivo ni sus volumenes. La fuente de datos fue el volumen de
staging existente; su anonimización no fue auditada en esta ejecucion, por lo
que este reporte no habilita pruebas funcionales con datos empresariales.

## Evidencia tecnica 2026-07-18

- El migrador se ejecutó tres veces despues de normalizar el ledger heredado;
  las dos ejecuciones repetidas fueron idempotentes.
- API candidata: `/health=200` y `/ready=200` desde la red privada de staging.
- Worker candidato: `healthy`; el trabajo de metricas opera con un intervalo
  acotado y rechaza configuraciones menores de cinco segundos.
- Frontend Nginx: smoke privado `/health=200` y `/ready=200`, con filesystem
  de solo lectura, `no-new-privileges` y capacidades retiradas.
- Rollback de aplicacion al SHA anterior: API privada `/health=200` y
  `/ready=200`. Despues se restauró la candidata en staging.
- El puerto publicado del frontend del host conserva una reserva Docker
  fantasma. No se reinició Docker porque comparte host con produccion; el
  smoke se ejecutó en la red privada. Este incidente de host queda pendiente
  antes de una prueba externa por navegador.
- SHA candidata validada en staging: `fd1f60c`.

No se ejecutaron proveedores externos, DIAN, pagos, correo, WhatsApp,
Nextcloud, OnlyOffice, RustDesk ni flujos visuales de empresas durante esta
ejecucion.

## Ejecucion requerida

1. Crear una red y volumenes con el prefijo `pcs_staging_` definido en
   `deploy/docker-compose.staging.yml`.
2. Cargar solamente una copia anonima o una base vacia de prueba.
3. Ejecutar `docker compose -f deploy/docker-compose.staging.yml config` y
   levantar el perfil aislado.
4. Aplicar migraciones dos veces, comprobar healthchecks y repetir la prueba
   con dos empresas, dos usuarios y roles distintos.
5. Adjuntar aqui IDs de ejecucion de CI, hash de imagen y resultado de cada
   smoke test. No registrar secretos, URLs privadas ni datos empresariales.

El resultado requerido es una evidencia reproducible, no una declaracion de
que staging fue probado sin haberlo sido.
