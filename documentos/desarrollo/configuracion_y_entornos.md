# Configuración y separación de entornos

Estado: Vigente. Responsable: QA/operación e ingeniería backend. Revisión documental: 2026-09-05.

## Fuentes

La configuración efectiva se obtiene de loaders del código y Compose/scripts de cada entorno. Esta guía registra categorías críticas y fuentes, sin volcar valores. El catálogo completo por variable sigue pendiente de validación; no modificar defaults a partir de esta tabla.

| Configuración | Sensibilidad y finalidad | Fuente de consulta |
| --- | --- | --- |
| `DB_EMPRESAS_DSN`, `DB_SUPERADMIN_DSN` | Secreto: conexiones de los dos contextos PostgreSQL | [Migrador](../../backend/cmd/pcs-migrate/main.go), [worker](../../backend/cmd/pcs-worker/main.go), Compose |
| `PCS_RUNTIME_DB_USER`, `PCS_RUNTIME_DB_PASSWORD` | Identidad y secreto del rol runtime distinto del propietario | Migrador y [decisiones](../decisiones_tecnicas.md) |
| `CONFIG_ENC_KEY` | Secreto de cifrado; debe conservar recuperación/rotación controlada | [Decisiones](../decisiones_tecnicas.md), contrato del módulo |
| `PCS_RUNTIME_SCHEMA_BOOTSTRAP` | Control operativo sensible de compatibilidad de esquema | Migrador, [comandos](../comandos_codex.md) y Compose; API/worker productivos en 0 |
| `PCS_ENV`, `PCS_ENVIRONMENT` | Identidad del entorno según consumidor; comprobar ambas cuando apliquen | Compose y loaders del código |
| `PUBLIC_BASE_URL`, `GOOGLE_REDIRECT_URL` | URL del entorno y callback; no copiar producción a staging | [Manual](../manual_de_instalacion.md), Compose |
| `PCS_TEST_POSTGRES_DSN` | Secreto de conexión de integración aislada | Tests PostgreSQL y estrategia QA |
| `PCS_WORKER_ID`, `PCS_WORKER_HEALTH_ADDR` | Identidad del worker y salud administrativa | Código del worker; no publicar health del worker por proxy |
| Credenciales de pago, DIAN, IA y correo | Secretos por empresa o plataforma según contrato | Configuración privada del módulo; nunca en ejemplos ni catálogo |

## Entornos

| Entorno | Uso | Aislamiento exigido |
| --- | --- | --- |
| Desarrollo | Compilación y pruebas enfocadas | Configuración propia; revisar efectos antes de arrancar API/worker |
| Integración efímera | Migraciones, SQL, concurrencia y restore de prueba | Bases/volúmenes/red exclusivos, datos sintéticos o anonimizados y destinos externos controlados |
| Staging | Aceptación del candidato | [Override staging](../../deploy/docker-compose.staging.yml) junto a plataforma; revisar archivos efectivos y secretos separados |
| Producción | Operación autorizada | [Release por digests](../../deploy/docker-compose.release.yml), privilegios mínimos y recuperación aprobada |

No usar un DSN productivo en `PCS_TEST_POSTGRES_DSN`. Los nombres de variables no prueban el destino real. Registrar identificación segura del entorno, versión de esquema y artefactos antes de pruebas con datos.

## Cambio y recuperación de secretos

Registrar finalidad, propietario, consumidor, versión/referencia segura, procedimiento de rotación y recuperación, acceso mínimo y evidencia de validación sin valor secreto. No reemplazar claves de cifrado sin plan para descifrar datos existentes y restaurar backups. Revisar revocación de sesiones, webhooks y certificados según el incidente.

Los comandos de Compose que interpolan configuración pueden mostrar secretos: no adjuntar su salida sin redacción. El catálogo documental excluye archivos privados ignorados por Git; no sustituye un escáner de secretos ni una revisión de seguridad.
