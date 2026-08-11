# P109-010 - candidato ClamAV exclusivo de staging

Fecha: 2026-08-09

Estado: **parcial en staging; no aprobado para producción**.

## Cambio

Se añadió `deploy/docker-compose.staging-antivirus.yml`, incluido por los
arranques de staging y por las promociones por digest. Declara un servicio
interno `clamav`, sin `ports` publicados, con volumen exclusivo de firmas,
`no-new-privileges`, `tmpfs` restringidos y healthcheck de `clamdscan --ping
1`. Parte de `cap_drop: ALL` y añade solamente `CHOWN`, `DAC_OVERRIDE`,
`FOWNER`, `SETGID` y `SETUID`, necesarias para que el inicializador oficial
prepare `/run/clamav`, acceda al volumen de firmas y cambie al usuario de
servicio.

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

## Verificación real en staging (2026-08-11)

- Se construyó y recreó solamente `pcs-staging-backend` desde el candidato
  `e70a9406`; `pcs-staging-clamav` quedó interno y saludable. `/ready` devolvió
  `200` y ninguno de estos cambios se aplicó a producción.
- Desde la interfaz autenticada de PCS (`empresa_id=12`), un XML limpio fue
  aceptado y quedó visible en la bandeja. La sonda EICAR oficial fue rechazada
  visualmente con el mensaje de antivirus; no creó una nueva fila de soporte.
- Se detuvo únicamente `pcs-staging-clamav`: el mismo flujo de archivo limpio
  respondió indisponible y no radicó el soporte. Tras iniciar el servicio y
  recuperar su healthcheck, una carga limpia volvió a completarse.
- Las métricas internas, sin dimensiones de empresa ni archivo, registraron
  `clean=1`, `malware=1`, `unavailable=1` y `bypassed=0` después del reinicio.
- Para que la sonda oficial sea bloqueable antes de cualquier parser, el backend
  analiza los bytes multipart sin confiar en su formato y solo después valida
  PNG/JPEG/WebP/PDF/XML; así no persiste ni entrega contenido hostil a parsers.
- La revisión posterior detectó que Prometheus conservaba el inode de reglas
  anterior (67 líneas). Se recreó únicamente `pcs-prometheus` usando el
  archivo de entorno de monitoreo ya existente; `promtool` confirmó 17 reglas
  y el contenedor pasó a montar las 132 líneas, incluidas las cuatro alertas
  de antivirus. Backend, base de datos, ClamAV y producción no se tocaron.
- Con las reglas ya cargadas, una caída controlada de solo ClamAV generó un
  rechazo visual fail-closed del archivo limpio. Prometheus evaluó
  `PCSAntivirusSoportesNoDisponible` como `firing` y Alertmanager interno la
  recibió. ClamAV se recuperó saludable; la resolución automática queda sujeta
  a la ventana de 10 minutos de la regla y no se presentó como confirmada.

## Pendiente obligatorio de cierre

1. Confirmar en el receptor externo la entrega, deduplicación y resolución de
   las alertas de antivirus; los contadores se verificaron, pero esa entrega no.
2. Completar las demás señales P0 de P109-010 (outbox/lease/disco) con sus
   responsables, escalamiento, SLO y postmortem.
3. Repetir la matriz sobre el digest final de liberación antes de cualquier
   decisión GO/NO-GO.

P109-010 permanece parcial hasta completar los cierres operativos anteriores.
Producción permanece sin modificar.
