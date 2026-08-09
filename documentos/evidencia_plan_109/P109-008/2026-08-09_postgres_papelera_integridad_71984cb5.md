# P109-008 - Integración PostgreSQL de papelera e integridad CxP/IA

Fecha: 2026-08-09

Commit de prueba: `71984cb5` más el ajuste local documentado

Motor: PostgreSQL `16.14-alpine`, contenedor efímero aislado

## Alcance

El fixture ahora instala los prerrequisitos reales de `pcs-migrate` para CxP,
proveedores y outbox antes de crear las tablas temporales de soportes. Esto evita
que una tabla mínima oculte diferencias del esquema productivo.

La ejecución enfocada aprobó:

- eliminar y restaurar dentro de `empresa_id=12`;
- rechazo de restauración cruzada desde otra empresa;
- vista previa por retención;
- bloqueo de restauración cuando existe duplicado activo;
- bloqueo de eliminación de un soporte contabilizado;
- auditoría de transiciones;
- invalidación segura de aprobación ante incidente de integridad;
- preservación del terminal contabilizado;
- purga en dos fases, reintentos y replay idempotentes;
- imposibilidad de restaurar un tombstone purgado;
- precisión monetaria de cartera;
- indicador local DIAN separado por empresa.

Comando ejecutado con `PCS_TEST_POSTGRES_DSN` efímero:

```text
go test ./db -count=1 -run 'TestSoporteComprasIAPapeleraPostgres|TestSoporteComprasIARetentionEligibility|TestEmpresaCarteraMoneyPrecisionPostgres|TestEmpresaDIANLocalProductionFlagPostgres'
ok github.com/you/pos-backend/db 97.807s
```

Regresión posterior:

```text
go test ./db -count=1
go vet ./db
go test ./handlers -count=1 -run 'Test.*(Soporte|Cartera|DIAN|Usuario|Permission|Empresa)'
```

Los tres comandos aprobaron.

La compuerta `scripts/profesional_preflight.ps1` aprobó sus 20 controles:
PowerShell, JavaScript, módulos, seguridad, permisos, OpenAPI, observabilidad,
migraciones, QA funcional/roles/pagos, anonimización, SLO, hardening, UX,
documentación, Docker Compose y `git diff --check`.

## Limpieza

Se comprobó el nombre exacto antes de eliminar
`p109-pg-integration-71984cb5`. Después quedaron cero contenedores con ese
nombre, cero listener local en `15432` y el túnel SSH detenido. No se usaron ni
alteraron las bases operativas.

## Resultado

**PASS** para el contrato PostgreSQL aislado. P109-008 permanece parcial porque
todavía faltan clamd real/EICAR, purga real vencida con recuperación ante caída
y cierre integral de backup/restore del piloto.
