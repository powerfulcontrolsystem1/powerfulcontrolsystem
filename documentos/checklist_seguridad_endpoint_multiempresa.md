# Checklist de seguridad por endpoint multiempresa

Checklist obligatoria antes de crear, modificar o revisar endpoints que lean,
creen, actualicen, eliminen, exporten, importen o sincronicen datos de una
empresa. Aplica a rutas `/api/empresa/*`, acciones empresariales embebidas en
handlers compartidos, jobs que toquen datos por empresa, backups, borrados,
pagos, facturacion, caja, inventario, licencias y cualquier flujo donde un
usuario pueda enviar o alterar `empresa_id`.

## 1. Clasificacion del endpoint

- [ ] Identificar ruta, metodo HTTP, accion o parametro `action`.
- [ ] Identificar modulo funcional y pagina frontend que lo llama.
- [ ] Identificar tablas leidas o modificadas.
- [ ] Identificar si es lectura, creacion, actualizacion, eliminacion,
  exportacion, importacion, sincronizacion o accion financiera.
- [ ] Identificar si afecta dinero, inventario, facturacion, caja, licencias,
  usuarios, permisos, backups, archivos o datos personales.
- [ ] Documentar si el endpoint es publico, empresarial autenticado o exclusivo
  de super administrador.

## 2. Autenticacion y sesion

- [ ] El endpoint empresarial exige sesion valida.
- [ ] No depende solo de localStorage, cache, cookies no validadas o datos del
  frontend para autorizar.
- [ ] Si usa cookies/sesion, valida que la sesion corresponda al usuario actual.
- [ ] Si es endpoint publico, devuelve solo datos publicos saneados y no acepta
  mutaciones empresariales.
- [ ] Si es super administrador, valida rol `super_administrador` en backend.

## 3. Resolucion segura de empresa

- [ ] El `empresa_id` se obtiene y normaliza en backend.
- [ ] El usuario autenticado tiene relacion real con esa empresa o rol efectivo
  que permita operar en ella.
- [ ] Un `empresa_id` editado en URL, payload, cache o consola no permite acceder
  a otra empresa.
- [ ] Si el endpoint recibe IDs secundarios, se valida que pertenezcan al mismo
  `empresa_id`.
- [ ] Si el endpoint crea registros hijos, asigna `empresa_id` desde el contexto
  validado, no desde un campo libre del cliente.

## 4. Permisos, roles y licencias

- [ ] La ruta usa el wrapper de permisos correcto en `backend/main.go` o una
  validacion equivalente dentro del handler.
- [ ] La accion requerida coincide con el riesgo real: lectura, creacion,
  actualizacion, anulacion, eliminacion, exportacion o administracion.
- [ ] La licencia/modulo habilitado se valida en backend, no solo en el menu.
- [ ] Los roles operativos solo ven y ejecutan lo permitido.
- [ ] Los accesos de supervisor/admin no rompen el aislamiento por `empresa_id`.
- [ ] Si se agrega pagina o boton, se actualiza la matriz de permisos.

## 5. Consultas SQL y persistencia

- [ ] Toda consulta a tabla empresarial filtra por `empresa_id`.
- [ ] Todo `UPDATE`, `DELETE`, `SELECT`, `INSERT ... SELECT` o `UPSERT` incluye
  aislamiento por `empresa_id` cuando aplique.
- [ ] Los `JOIN` entre tablas empresariales unen tambien por `empresa_id`, no
  solo por `id`.
- [ ] Los `COUNT`, totales, reportes y exportaciones filtran por `empresa_id`.
- [ ] Las operaciones por ID verifican `id` y `empresa_id` juntos.
- [ ] Las transacciones mantienen el mismo `empresa_id` en todos los pasos.
- [ ] Despues de mutaciones criticas se revisa `RowsAffected` cuando aplique para
  detectar IDs inexistentes o de otra empresa.
- [ ] No se usan motores distintos a PostgreSQL ni sintaxis que rompa el runtime
  PostgreSQL vigente.

## 6. Validacion de entrada

- [ ] IDs numericos, fechas, montos, porcentajes y cantidades se parsean y
  normalizan en backend.
- [ ] Montos negativos, descuentos, devoluciones, abonos o cantidades solo se
  aceptan cuando la regla de negocio lo permite.
- [ ] Textos libres se limitan y sanean antes de guardar o imprimir.
- [ ] Archivos subidos validan tipo, tamano, ruta destino y pertenencia a empresa.
- [ ] URLs externas, plantillas, webhooks o APIs de proveedores se validan antes
  de usarse.
- [ ] No se aceptan campos sensibles que el cliente no deba controlar, como rol,
  estado fiscal final, usuario creador, empresa de destino o totales calculados.

## 7. Respuestas y manejo de errores

- [ ] Los errores no exponen SQL, DSN, tokens, certificados, claves, rutas
  privadas ni datos de otra empresa.
- [ ] Las respuestas incluyen estado claro para frontend sin filtrar informacion
  sensible.
- [ ] Los fallos de proveedor externo no dejan transacciones parciales sin
  estado auditable.
- [ ] Los endpoints idempotentes documentan su clave de idempotencia o criterio
  de reintento.

## 8. Auditoria y trazabilidad

- [ ] Registrar auditoria para caja, pagos, abonos, egresos, ingresos,
  anulaciones, facturacion, licencias, usuarios, permisos, backups, borrados,
  importaciones, exportaciones y cambios de configuracion.
- [ ] La auditoria incluye `empresa_id`, usuario, accion, fecha/hora y referencia
  operativa suficiente.
- [ ] La auditoria no guarda contrasenas, tokens, claves privadas, certificados
  completos ni secretos.
- [ ] Si el cambio afecta conectividad/offline, la cola local se sincroniza y
  registra sin duplicar ventas.

## 9. Operaciones destructivas o masivas

- [ ] Borrados, reinicios de datos, backups, importaciones y limpiezas exigen
  confirmacion fuerte o permiso de alto nivel.
- [ ] La previsualizacion de impacto filtra por `empresa_id`.
- [ ] Los backups previos no incluyen secretos innecesarios y respetan aislamiento
  de empresa.
- [ ] No se elimina configuracion, usuarios, permisos, licencias, impresoras,
  integraciones ni preferencias salvo instruccion explicita.
- [ ] Se documenta si el borrado es logico o fisico.

## 10. Pruebas minimas obligatorias

- [ ] Caso positivo con empresa valida y rol permitido.
- [ ] Caso sin sesion o sesion invalida.
- [ ] Caso con rol sin permiso suficiente.
- [ ] Caso alterando `empresa_id` en URL o payload hacia otra empresa.
- [ ] Caso usando ID secundario de otra empresa.
- [ ] Caso con empresa inexistente, inactiva o sin licencia/modulo si aplica.
- [ ] Caso de payload invalido: IDs, fechas, montos, textos o archivos.
- [ ] Caso de concurrencia o doble clic cuando pueda duplicar ventas, pagos,
  abonos, licencias, documentos o sincronizaciones.
- [ ] Para reportes/exportaciones: verificar que no aparezcan datos de otra
  empresa.
- [ ] Para frontend: comprobar que ocultar botones no sustituye la validacion del
  backend.

## 11. Evidencia de cierre

Al cerrar una tarea que toque endpoint multiempresa, el resumen debe indicar:

- Endpoint(s), handler(s) y tabla(s) afectadas.
- Como se valida `empresa_id`.
- Wrapper de permisos/licencia usado.
- Pruebas positivas y negativas ejecutadas.
- Riesgo residual si alguna prueba no se pudo ejecutar.

## 12. Excepciones controladas

- Endpoints publicos solo pueden devolver contenido publico saneado, catalogos
  no sensibles o informacion agregada sin datos personales.
- Endpoints super administrador pueden operar globalmente solo con rol
  `super_administrador` validado en backend y sin exponer secretos.
- Jobs internos deben registrar alcance, origen y filtros por empresa aunque no
  reciban `empresa_id` desde HTTP.

