# P110-001 — migración vacía y upgrade aislado de `e308ca4b`

Fecha: 2026-08-12  
Ambiente: VPS con Docker, PostgreSQL, red, volumen y directorios temporales.
Staging fue solo el origen de una copia lógica; no recibió escrituras.

## Base vacía

El migrador exacto creó el catálogo en una base PostgreSQL vacía y la segunda
pasada no agregó migraciones. Se alteró temporalmente un checksum del ledger:
el proceso falló cerrado, el esquema y ledger quedaron sin cambio, se restauró
el checksum controlado y la pasada posterior aprobó.

```text
empresas_migrations=24 empresas_tables=339
super_migrations=10 super_tables=49
drift_failure=closed drift_schema_unchanged=339
drift_ledger_unchanged=24 drift_recovery=pass
```

## Upgrade desde staging

El primer intento descubrió que el runner no montaba el directorio privado que
exige el candidato en modo producción. Se corrigió exclusivamente el runner
aislado: crea, monta y elimina un directorio privado temporal. La repetición
aplicó el migrador y una segunda pasada sobre la copia lógica de staging.

```text
empresas_tables_before=352 after=352
super_tables_before=59 after=59
```

Al finalizar se verificó que no quedaban contenedores ni volúmenes del prefijo
del drill y que `health`/`ready` de staging seguían correctos.

## Límite

P110-001 sigue parcial hasta la revisión/aceptación del ADR y una estrategia
formal de rollback compatible con toda migración futura. Este ensayo no firma
RPO/RTO ni autoriza promoción.
