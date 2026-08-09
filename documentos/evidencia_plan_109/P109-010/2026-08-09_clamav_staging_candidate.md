# P109-010 - candidato ClamAV exclusivo de staging

Fecha: 2026-08-09

Estado: **preparado para despliegue controlado; no aprobado**.

## Cambio

Se añadió `deploy/docker-compose.staging-antivirus.yml`, incluido por los
arranques de staging y por las promociones por digest. Declara un servicio
interno `clamav`, sin `ports` publicados, con volumen exclusivo de firmas,
`no-new-privileges`, todas las capabilities retiradas, `tmpfs` restringidos y
healthcheck de `clamdscan --ping 1`.

El backend de staging queda dependiente de la salud del servicio y recibe
`PCS_SUPPORTS_CLAMAV_ADDR=clamav:3310` y
`PCS_SUPPORTS_CLAMAV_REQUIRED=1`. Producción no incluye este overlay ni recibe
esa dependencia.

La imagen externa autorizada es la oficial `clamav/clamav`, fijada en la
plantilla de staging por digest de la versión 1.5.4 Debian slim. El script de
promoción exige que también sea un digest, lo verifica en Compose, lo descarga
explícitamente y levanta `clamav` junto con los servicios de staging.

## Verificación local

- Pruebas Go de ClamAV simulado: PASS para archivo limpio, firma EICAR/malware,
  endpoint ausente con modo obligatorio y métricas concurrentes.
- La estación de trabajo no tiene Docker ni una distribución WSL; por ello
  `docker compose config` y el daemon real no se ejecutaron localmente.

## Prueba pendiente obligatoria en staging

1. Promover el candidato por digest y esperar healthcheck de `clamav`.
2. Subir un archivo limpio por el endpoint oficial y confirmar aceptación.
3. Subir la cadena EICAR controlada y confirmar rechazo HTTP 422, sin archivo
   persistido.
4. Detener únicamente `pcs-staging-clamav`, reintentar el archivo limpio y
   confirmar HTTP 503 fail-closed; reiniciar el servicio y confirmar
   recuperación.
5. Verificar contadores/alertas Prometheus y que no aparezcan nombres, rutas,
   empresas o contenido de archivos en métricas.

No se modificó staging ni producción en este bloque. P109-010 y P109-008
permanecen parciales hasta completar estas cinco verificaciones sobre el mismo
digest.
