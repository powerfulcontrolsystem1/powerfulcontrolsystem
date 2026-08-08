# P109-011 - Migracion y restore aislados del candidato `c8094f5b`

Fecha: 2026-08-08 01:42 America/Bogota  
Entorno: VPS de staging, recursos Docker efimeros y aislados  
Alcance: candidato exacto de staging; no se modificaron PCS, staging activo ni produccion.

## Artefacto y origen

- Codigo candidato: `c8094f5be638bbd6262e12e191d365793ed92f6b`.
- Release inmutable: workflow `31243743197`; los digests se tomaron de su artefacto local `release-images.env`.
- Snapshot restaurado: `20260808_031501`.
- Los scripts se copiaron a `/tmp` con permisos `0700`, se contrastaron con SHA-256 antes de ejecutarse y se eliminaron al terminar.

## Restore de aplicacion sobre snapshot

`vps-p109-restored-app-drill.sh` inicio el migrador y API exactos contra la copia restaurada. El resultado fue:

- health y readiness HTTP: `200`;
- dos bases restauradas, cinco tablas criticas y 28 filas criticas de PCS;
- cuatro endpoints anonimos protegidos rechazados;
- inventario privado: 2 archivos y 2 referencias, sin huerfanos ni referencias heredadas;
- rol runtime sin DDL;
- RTO: 23 s; RPO observado: 12.287 s (3 h 24 min 47 s), dentro de los objetivos publicados de 2 h y 24 h.

## Migracion, fallos y compatibilidad

`vps-p108-empty-migration-drill.sh` se ejecuto con `P109_VERIFY_MIGRATION_FAILURES=1` sobre una base vacia separada. Aprobo:

1. tres rechazos previos con rol sin privilegios DDL;
2. primera migracion y segunda pasada idempotente;
3. drift de checksum fail-closed, sin cambio de esquema ni ledger (`337` tablas y `19` entradas); recuperacion posterior aprobada;
4. cinco controles de fallo durante la transaccion, con rollback de DDL y ledger y recuperacion atomica;
5. cuatro controles de compatibilidad: API del candidato anterior sobre el esquema actualizado, sin mutar tablas ni recibir DDL;
6. catalogo final: 337 tablas/19 migraciones empresariales y 49 tablas/10 migraciones de superadministrador.

La limpieza posterior verifico cero contenedores, redes, volumenes o scripts efimeros con el identificador del ensayo.

## Conclusion y limites

P109-011 queda **aprobada para el candidato exacto `c8094f5b`**: migracion, rollback de datos/aplicacion y restore aislados fueron ensayados dentro de RPO/RTO publicados. Esto no autoriza produccion: siguen abiertas las demas compuertas P0/P1 y la aprobacion final humana del piloto.
