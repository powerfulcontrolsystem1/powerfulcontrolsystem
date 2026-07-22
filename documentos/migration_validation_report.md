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

## Intento local Plan 105 - 2026-07-21

- Se verificaron los inventarios de bootstrap y runtime despues de cuatro lotes
  de extraccion de DDL HTTP: 153 funciones/122 pasos de catalogo y 110 llamadas
  runtime, de las cuales 37 siguen alcanzables por HTTP.
- El segundo lote reemplaza en solicitudes HTTP la inicializacion DDL de Energia
  Solar, Hoja de Vida Operativa y Reservas Hotel por verificaciones de esquema
  solo lectura. Sus pruebas de guard nulo y `go vet ./db ./handlers` pasan.
- El tercer lote retira la inicializacion DDL de programacion de reportes de los
  reportes por empresa y globales; verifica las tres tablas y sus columnas de
  contrato en modo lectura. Sus pruebas de guard nulo y `go vet ./db ./handlers`
  pasan.
- El cuarto lote retira DDL de las rutas HTTP de Chat/Tareas. El canal general
  conserva su creacion de dato empresarial, pero exige el esquema ya migrado en
  modo lectura. Sus pruebas de guard nulo y compilacion de handlers pasan.
- Se intento localizar un camino de PostgreSQL efimero reproducible. Este equipo
  no tiene el ejecutable Docker disponible, por lo que no se levanto ningun
  contenedor ni se simulo una migracion.
- El bloqueo no cambia el estado: sigue siendo obligatorio ejecutar instalacion
  vacia, upgrade, segunda corrida y rollback en Docker/CI/staging con PostgreSQL
  real antes de usar `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0` en una instalacion
  existente.
