# Validacion de migraciones

Estado: **PENDIENTE DE POSTGRESQL EFIMERO**.

Las migraciones y utilidades de archivos se deben ejecutar primero en modo
simulacion. La validacion debe cubrir: esquema vacio, esquema actualizado,
segunda ejecucion idempotente, datos incompletos, rollback y aislamiento por
`empresa_id`.

Para adjuntos heredados, `backend/tools/migrate_private_uploads` se usa sin
aplicar cambios hasta revisar inventario. La migracion real solo se autoriza en
staging anonimizado, dentro de una ventana con backup verificado.

No se ejecutaron migraciones contra bases reales en este trabajo.
