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
