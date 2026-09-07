# Nginx y terminación TLS

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se separan instrucciones actuales y migraciones históricas; cada intervención remota requiere entorno identificado y evidencia propia.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Alcance

El proxy debe corresponder a la topología efectiva: Nginx del host hacia frontend
privado o edge Docker. Identificar propietario de 80/443, certificado, upstream,
red y servicios antes de intervenir. No reemplazar un vhost por un ejemplo genérico.

## Procedimiento

1. Registrar candidato, hostname y estado; respaldar la configuración afectada sin exponer llaves.
2. Comprobar DNS/TLS desde red externa y salud del upstream desde red administrativa.
3. Revisar [Nginx interno](../deploy/nginx/pcs.conf), Compose y configuración efectiva;
   preservar CSP, HSTS, orígenes y límites. Los headers reenviados solo son fiables
   desde proxies conocidos por el backend.
4. Aplicar el cambio autorizado, ejecutar `nginx -t` y recargar solo el proxy afectado.
5. Validar TLS sin omitir certificados, health/ready, login, API y rutas del tenant.
6. Ante error restaurar configuración respaldada y validar antes de recargar;
   no restaurar un certificado vencido ni cambiar DNS para ocultar una falla.

El [runbook de incidentes](operacion/incidentes_y_continuidad.md) define escalamiento.
El inventario privado del entorno es la única fuente para rutas, certificados y
vhosts reales; ejemplos de ejecuciones anteriores no describen el servidor actual.

## Fuentes y aceptación de la revisión

[pcs.conf](../deploy/nginx/pcs.conf), [docker-compose.platform.yml](../deploy/docker-compose.platform.yml).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
