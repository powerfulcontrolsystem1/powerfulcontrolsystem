# Gobierno de datos y evolución del esquema

Estado: Vigente. Responsable: Ingeniería backend y datos. Revisión documental: 2026-09-05.

## Fuentes y ownership

[Estructura BD](../estructura_bd.md) contiene tablas y relaciones; [platform_migrations.go](../../backend/db/platform_migrations.go) registra las migraciones; [pcs-migrate](../../backend/cmd/pcs-migrate/main.go) aplica el esquema. Los diagramas son derivados, no autoridad DDL. Un inventario de código no demuestra el esquema instalado.

| Conjunto | Frontera y autoridad | Controles exigidos |
| --- | --- | --- |
| Operación empresarial | Base empresarial, módulo dueño y `empresa_id` | IDs relacionados del mismo tenant, joins y agregados filtrados, transacciones |
| Administración global | Base super y rol autorizado | Alcance global explícito, auditoría y exportaciones minimizadas |
| Vida y datos personales de usuario | `empresa_id + usuario_id` | No heredar acceso por compartir empresa; adjuntos con igual aislamiento |
| Configuración y secretos | Configuración del módulo/entorno y cifrado aplicable | No plaintext en repositorio/log; recuperar claves por canal privado |
| Fuentes fiscales y documentos | Empresa, país, familia y operación original | Inmutabilidad, numeración y referencias; acuse por adaptador |
| Jobs, outbox e idempotencia | Migrador para esquema; productor/consumidor para operación | Identidad estable, reclamo, lease, reintento y reconciliación |
| Archivos | Módulo y almacenamiento privado por tenant | Autorización antes de servir, integridad, límites y recuperación consistente |

## Contrato mínimo de cada entidad nueva

Registrar nombre, propósito y módulo dueño; clasificación y titular; PK/FK; tenant/usuario; nulabilidad y defaults; claves únicas/idempotencia; unidades y precisión monetaria; zona horaria y semántica de fechas; índices y patrón de consultas; creación/actualización/borrado; retención; migración; consumidores y pruebas. No sustituir este contrato por un `CREATE TABLE` sin explicación.

## Migraciones

1. Añadir una migración nueva con identificador, target y checksum; no reescribir cuerpos aplicados.
2. Describir precondiciones, locking, volumen estimado, compatibilidad con el binario previo y tratamiento de datos inconsistentes.
3. Probar esquema vacío, reejecución y actualización de un clon anonimizado del esquema vigente en PostgreSQL aislado.
4. Revisar lectura/escritura anterior y nueva, concurrencia e índices. Detener ante pérdida de precisión o inconsistencias que requieren conciliación.
5. API y worker verifican disponibilidad; no asumen que pueden crear tablas. Revisar roles reales del entorno, no solo variables declaradas.
6. Definir recuperación: rollback de binario si esquema compatible, o corrección hacia adelante; restaurar datos requiere plan específico por pérdida potencial de escrituras posteriores.

Usar los comandos de [comandos Codex](../comandos_codex.md). No ejecutar scripts ni DSN productivas por copiar ejemplos. El migrador conecta ambas bases y requiere credenciales privadas; el arranque completo puede crear datos o despachar trabajos.

## Ciclo de vida y privacidad

No hay un plazo universal de retención. Para cada categoría establecer finalidad, base contractual/legal aplicable, duración aprobada, suspensión por litigio/auditoría, procedimiento de solicitud, anonimización/borrado y cobertura de backups. Las duraciones legales y contratos con terceros requieren validación de la organización; esta guía no las inventa.

Clasificar al menos: identidad y contacto; operación comercial; datos financieros/fiscales; nómina; Vida personal; adjuntos; telemetría; auditoría; secretos. Mantener datos sensibles fuera de ejemplos, fixtures versionados e informes públicos. Ver [amenazas y privacidad](../seguridad/modelo_amenazas_y_privacidad.md).

## Recuperación y coherencia

El respaldo debe permitir recuperar las dos bases, archivos privados, manifiestos y configuración necesaria de forma coherente. Las claves de cifrado deben ser recuperables por separado con control de acceso. Un dump sin archivos o sin claves utilizables no demuestra recuperabilidad del negocio.

Aplicar [runbook de recuperación](../gobernanza_tecnica/runbooks/runbook_recuperacion_desastre_docker_vps.md), medir RTO/RPO y reconciliar saldos, referencias, documentos y aislamiento antes de reabrir escrituras. No restaurar un volumen físico PostgreSQL copiado en caliente como si fuese un backup consistente.
