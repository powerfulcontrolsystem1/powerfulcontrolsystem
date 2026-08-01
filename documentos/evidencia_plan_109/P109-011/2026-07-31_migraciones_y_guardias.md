# P109-011 - Migraciones y guardias de runtime

Fecha: 2026-07-31
Entorno: suite local y ensayo previo del digest de staging.

## Ensayo y pruebas

El candidato desplegado en staging completó migración sobre base vacía y sobre
una copia efímera de staging. La segunda pasada fue idempotente y el ensayo de
checksum alterado falló de forma cerrada antes de restaurar el estado controlado.

Se repitió localmente:

```text
go test ./db ./internal/platform/runtimeconfig -run 'Test(MigrationChecksumIncludesImmutableBody|ValidateMigrationCatalogRejectsInvalidOrderingAndDuplicates|PlatformMigrationCatalogsAreOrderedAndChecksummed|LegacySchemaManifestV1KeepsReleasedChecksums|RuntimeSchemaGuard|LoadProductionDisablesCompatibilityBootstrapByDefault|LoadProductionRejectsAPIBootstrapOverride|LoadProductionRejectsWorkerBootstrapOverride|LoadMigrationRoleEnablesSchemaBootstrap)' -count=1
```

Resultado: `db` PASS y `runtimeconfig` PASS. La cobertura exige catálogos
ordenados/checksum, rechaza bootstrap en API/worker de producción y permite DDL
solo al rol de migración explícito.

## Límite de cierre

P109-011 permanece parcial. Aún debe ensayarse rollback coordinado de datos y
compatibilidad hacia atrás bajo una falla controlada, dentro de RPO/RTO y con
evidencia de que no quedan residuos.

## Rollback de datos posterior a migración - 2026-08-01

El candidato exacto se aplicó al snapshot efímero y, después de crear un soporte
por el flujo oficial, se congelaron las dos bases y el volumen privado. Con las
APIs detenidas se perdieron las copias temporales, se recrearon ambas bases y se
restauraron junto con los archivos. El runtime volvió a readiness, autenticó,
recuperó la fila y el SHA-256, y conservó las cinco tablas críticas. El rollback
coordinado tardó 23 segundos y dejó cero recursos efímeros.

P109-011 continúa **parcial**: queda pendiente simular fallos antes y durante la
migración y completar la matriz de compatibilidad hacia atrás. El ensayo nuevo
cierra únicamente el escenario posterior a migración y no fija por sí solo el
RPO/RTO contractual.
