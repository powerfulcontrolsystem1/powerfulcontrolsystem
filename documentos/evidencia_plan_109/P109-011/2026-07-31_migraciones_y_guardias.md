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

## Fallos antes/durante y compatibilidad hacia atrás - 2026-08-01

El modo `P109_VERIFY_MIGRATION_FAILURES=1` se ejecutó en la VPS únicamente
sobre PostgreSQL, volumen, red y puerto efímeros. Usó el migrador exacto del
candidato `89d6e042...` y la API anterior `8847288b...`, ambos por digest.

- Antes: un rol con `CONNECT/USAGE` pero sin DDL intentó ejecutar el migrador y
  falló cerrado; el número de tablas permaneció idéntico.
- Durante: en la copia efímera se retiraron el índice y ledger de
  `20260731-001-ai-usage-unique-v1`; un trigger temporal provocó error al
  insertar el ledger después de ejecutar el `CREATE UNIQUE INDEX` real.
- El fallo dejó índice ausente y ledger ausente, demostrando rollback de la
  transacción completa, y agregó una corrida `failed` a la auditoría.
- Tras retirar el trigger, el migrador reaplicó índice y ledger una sola vez.
- La API del candidato anterior alcanzó `/health` y `/ready` sobre el esquema
  nuevo, conservó el número de tablas y utilizó un rol sin privilegios DDL.

Resultado:

```text
empresas_migrations=18
empresas_tables=337
super_migrations=10
super_tables=49
[OK] Base vacía y segunda pasada completadas; fallos_previos=3 rollback_transaccional=5 compatibilidad_anterior=4
```

La limpieza confirmó cero contenedores, volúmenes, redes y temporales. Staging
y producción conservaron sus APIs activas y sus imágenes originales. Con esto
quedan demostrados técnicamente los escenarios antes, durante y después, la
segunda pasada, drift, rollback transaccional, rollback coordinado de datos y
compatibilidad con el candidato anterior. P109-011 conserva estado **parcial**
hasta aprobar formalmente el RPO/RTO y vincular la aceptación al digest final
que se elija para el piloto.

También se interrumpió una ejecución aislada con `TERM` mientras PostgreSQL
iniciaba. `timeout` devolvió `124`, como corresponde a la interrupción externa,
y la trampa de salida dejó `0` contenedores, `0` volúmenes, `0` redes y ningún
directorio temporal. La prueba no reinició ni modificó los servicios activos.

## Aprobacion sobre el candidato edd2bdac - 2026-08-01

El API y migrador exactos de
`edd2bdac64f760d18289249d6c2f4720aba6f925` repitieron ambos ensayos en
recursos efimeros del VPS.

Base vacia y guardias:

```text
empresas_migrations=19
empresas_tables=337
super_migrations=10
super_tables=49
fallos_previos=3
rollback_transaccional=5
compatibilidad_anterior=4
```

La segunda pasada fue idempotente, el drift de checksum fallo cerrado, el DDL
y ledger se revirtieron juntos ante el fallo inyectado y la API anterior
alcanzo health/readiness sobre el esquema nuevo con cero privilegios DDL.

Restore y rollback coordinado:

```text
bases=2 tablas_criticas=5 filas_empresa_12=28
endpoints_anonimos_bloqueados=4 dominios_autenticados=5
replicas=2 archivos_hostiles=5
archivos_privados=2 referencias_privadas=2 huerfanos=0 heredadas=0
rollback_checks=7 rollback_dominios=5
rollback_RTO=24s RTO_total=48s RPO=5466s runtime_privilegios=0
```

El SLO vigente fija RTO de produccion en 2 horas y RPO diario en 24 horas. Los
48 segundos y 5.466 segundos observados quedan dentro de ambos objetivos. La
limpieza final verifico cero contenedores, redes, volumenes y temporales; staging
y produccion conservaron readiness.

Estado: **P109-011 aprobada** para el candidato `edd2bdac...`. Quedan cubiertos
base vacia, upgrade, idempotencia, fallos antes/durante/despues, compatibilidad
hacia atras, checksum, auditoria, rollback de aplicacion/datos y ausencia de
residuos dentro del RPO/RTO publicado.
