# Inventario de supply chain - Plan 105

Fecha de corte: 2026-07-21. Revision estatica de Compose, manifiesto y scripts;
no consulta registros de imagenes ni modifica VPS.

## Estado comprobado

El release de PCS ya exige referencias por digest para migrador, API y worker:
`PCS_MIGRATE_IMAGE_DIGEST`, `PCS_API_IMAGE_DIGEST` y
`PCS_WORKER_IMAGE_DIGEST`. La compuerta local se bloquea correctamente cuando
faltan. Los roles PCS del Compose de plataforma incluyen `read_only`,
`no-new-privileges` y `cap_drop`.

No es un cierre P105-016: el Compose contiene servicios auxiliares con tags
versionados o variables de tag (PostgreSQL, Redis, Mailu, Nginx, Certbot,
OnlyOffice, RustDesk, Prometheus, Grafana y cAdvisor). `cAdvisor` es
privilegiado. Aun con tag fijo, no se demostro digest, SBOM, escaneo ni
rollback reproducible para cada imagen habilitada.

## Secuencia ejecutable para Terra

1. Generar un inventario de servicios efectivamente habilitados en cada perfil
   (produccion, staging, monitoreo, OnlyOffice), con imagen, digest, origen,
   licencia, propietario y justificacion de privilegios/mounts.
2. Resolver cada tag a digest y moverlo a variables `*_IMAGE_DIGEST` sin
   sustituir silenciosamente servicios no PCS. El release debe fallar si falta
   cualquiera de los digests habilitados.
3. Construir SBOM y escanear los digests. Registrar CVE, severidad, mitigacion,
   fecha de revision y aceptacion explicita de toda excepcion alta/critica.
4. Revisar `user`, `read_only`, `no-new-privileges`, `cap_drop`, limites y
   healthcheck por servicio. Para cAdvisor mantener el minimo de mounts y red,
   justificar `privileged` o reemplazarlo por una configuracion menos amplia.
5. En staging arrancar desde los mismos digests, probar migrador una vez,
   healthchecks, rollback y restauracion. Adjuntar los manifiestos y hashes al
   release; no usar tags mutables para el piloto.

## Criterio de cierre

Todos los servicios habilitados usan digest validado, tienen procedencia y
escaneo trazables, privilegios justificados, y staging demuestra update y
rollback con el mismo manifiesto. La ausencia de los tres digests PCS bloquea
el release actual antes de acciones remotas.
