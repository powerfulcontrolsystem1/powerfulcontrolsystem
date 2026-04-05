
## agente go

Reglas de documentación técnica:

- Antes de modificar arquitectura, rutas, modelos o flujos críticos, revisar `documentos/diagramas/estructura_del_codigo.md` como referencia de estructura vigente.
- Antes de implementar cambios que afecten tablas, consultas, migraciones o datos operativos, revisar `documentos/estructura_bd.md` como fuente canonica del esquema de base de datos.
- Mantener y actualizar los diagramas en `documentos/diagramas/` cuando haya cambios de backend, base de datos o frontend que afecten el flujo funcional.
- Toda actualización en `documentos/diagramas/` debe quedar registrada en `documentos/descripcion_de_archivos` y `documentos/historial_de_cambios`.

Regla funcional de estaciones:

- En este proyecto, una estacion puede representar distintos puntos operativos segun el tipo de negocio: mesas de restaurante, habitaciones de hotel, habitaciones de motel, puntos de caja u otros puntos de atencion equivalentes.
- Las estaciones deben soportar operacion concurrente de multiples carritos/sesiones para multiples clientes en simultaneo, manteniendo aislamiento por `empresa_id` y trazabilidad por carrito y cliente.

Regla de reportes e interoperabilidad contable:

- Todo reporte del sistema (nuevo o existente) debe poder exportarse como minimo a PDF y Excel, y tambien a formatos de uso comun como CSV, JSON y TXT.
- Las exportaciones de un mismo reporte deben conservar estructura, columnas clave y totales para evitar discrepancias entre formatos.
- El sistema debe mantener compatibilidad con software contable externo mediante formatos estandar de intercambio y datos contables trazables por `empresa_id`, documento y periodo.