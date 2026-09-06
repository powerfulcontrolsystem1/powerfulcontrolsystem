# Operación Docker, release y recuperación

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se separan instrucciones actuales y migraciones históricas; cada intervención remota requiere entorno identificado y evidencia propia.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Topología e inventario

El [Compose base](../deploy/docker-compose.platform.yml) declara PostgreSQL,
migrador, API, worker, frontend e inicialización de permisos de almacenamiento.
Los perfiles agregan edge/certificados, correo, oficina, voz o soporte remoto.
El Nginx del host y edge Docker son alternativas de terminación que deben
inventariarse; no asumir cuál atiende el dominio a partir de una nota histórica.

Consultar servicios por proyecto Compose. API/worker/frontend pueden usar nombres
generados y réplicas; `pcs-backend` no es identidad universal. Los servicios
opcionales y Nextcloud externo requieren inventario propio.

## Secuencia de release

1. Identificar candidato limpio, commit y configuración privada del destino.
2. Completar CI, migraciones y aceptación de staging con evidencias asociadas.
3. Usar el [override de release](../deploy/docker-compose.release.yml) con digests
   inmutables de API, migrador, worker y frontend. El init de uploads usa el digest API.
4. Validar configuración sin imprimir valores mediante `config --quiet`, con
   base+override y archivo privado efectivo. Ejecutar migrador antes de API/worker.
5. Promover con autoridad explícita de despliegue y validar runtime/negocio.

`rs` puede integrar cambios, actualizar Git y sincronizar el VPS. Sus modos y
compuertas están en [comandos](comandos_codex.md), el
[checklist release](release_checklist.md) y el
[runbook](gobernanza_tecnica/runbooks/runbook_release_profesional.md).
No usar legacy/hybrid, SkipPreflight o AllowNonMainDeployment como bypass de aprobación.

## Escala y archivos privados

Más de una API requiere `PCS_PRIVATE_STORAGE_MODE=shared` y volumen realmente
compartido; `object` se rechaza porque no tiene adaptador operativo. Dimensionar
pools por proceso/réplica y probar leases/idempotencia antes de aumentar workers.
No inferir capacidad solo porque Compose acepte una cifra de réplicas.

## Paquetes y backup

El paquete portable del panel contiene código, no datos/secretos. El snapshot
del panel puede incluir `postgres/pg_dumpall.sql`, nombre histórico para dumps
lógicos de ambas bases mediante pg_dump y rol de backup. El script operacional
`vps-backup-operacion.sh` produce otro formato: `postgres_all.sql.gz`.
Revisar el manifiesto, presencia/integridad del dump y archivos privados; un
paquete descargable no demuestra recuperación.

[Continuidad](operacion/incidentes_y_continuidad.md) y
[restore](gobernanza_tecnica/runbooks/runbook_recuperacion_desastre_docker_vps.md)
exigen ensayo aislado, claves recuperables y copia externa verificada. No
restaurar un tar de PostgreSQL tomado en caliente ni borrar volúmenes por nombre.

El inventario y las conmutaciones antiguas se conservan en la
[referencia histórica](historico/2026-09-05/docker_vps_operacion_referencia_acumulada.md).

## Fuentes y aceptación de la revisión

[docker-compose.platform.yml](../deploy/docker-compose.platform.yml), [docker-compose.release.yml](../deploy/docker-compose.release.yml), [super_vps_snapshots.go](../backend/handlers/super_vps_snapshots.go), [vps-backup-operacion.sh](../deploy/scripts/vps-backup-operacion.sh).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
