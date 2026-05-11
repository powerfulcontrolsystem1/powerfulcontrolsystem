# Diseno tecnico ERP multiempresa

Fecha: 2026-04-14
Version: 1.0
Estado: listo para revision

## 1. Objetivo tecnico

Definir la arquitectura tecnica objetivo del ERP multiempresa, su modelo de datos, seguridad, integraciones y lineamientos de evolucion sin romper la operacion existente.

## 2. Principios de arquitectura

1. Multiempresa por diseno: aislamiento obligatorio por `empresa_id`.
2. Seguridad por capas: autenticacion, autorizacion y trazabilidad.
3. Modularidad por dominio: handlers, DB y UI desacoplados por contexto.
4. Trazabilidad extremo a extremo: request, actor, documento y periodo.
5. Compatibilidad operativa: transicion controlada entre motores de datos.
6. Estandar de interoperabilidad: exportaciones consistentes y formatos comunes.

## 3. Vista de arquitectura logica

### 3.1 Capa de presentacion

- Frontend web segmentado por contexto:
  - panel super administrador
  - panel administrar empresa
  - vistas publicas operativas

### 3.2 Capa de aplicacion

- Servidor Go con registro central de rutas.
- Handlers por dominio (ventas, inventario, finanzas, seguridad, etc.).
- Middleware transversal para:
  - validacion de sesion
  - validacion de permisos
  - validacion de alcance por empresa
  - manejo uniforme de errores y logging

### 3.3 Capa de dominio y datos

- Capa DB con servicios por modulo.
- Tablas empresariales con `empresa_id` y campos estandar de trazabilidad.
- Procesamiento contable por eventos y asientos canonicos.

### 3.4 Capa de integracion externa

- OAuth Google, SMTP, pasarela de pagos, DIAN, mapas y proveedor IA.
- Conectores con validacion previa, manejo de errores y evidencia operativa.

## 4. Topologia de datos

## 4.1 Bases objetivo

- `pcs_superadministrador` (PostgreSQL VPS): gobierno global.
- `pcs_empresas` (PostgreSQL VPS): operacion transaccional por empresa.

## 4.2 Bases legado

- No hay archivos `.db` legacy versionados en el repositorio vigente.
- El legado motor legado retirado queda solo como referencia historica de migracion o como respaldo tecnico externo y no como artefacto operativo local.

## 4.3 Patrone de modelado

- Campos comunes: id, fechas, estado, usuario_creador, observaciones.
- Indices por llaves de negocio y busqueda operativa.
- Restricciones de unicidad por empresa para codigos/documentos criticos.
- Estados de flujo por modulo con transiciones validas.

## 5. Seguridad y control de acceso

## 5.1 Autenticacion

- Administradores: OAuth Google con flujo de aceptacion contractual.
- Usuarios empresa: email/password con confirmacion y primer ingreso.

## 5.2 Autorizacion

- Modelo RBAC por rol/modulo/accion.
- Wrappers de permisos sobre rutas empresariales criticas.
- Acciones sensibles (aprobar, cerrar, restaurar) con control reforzado.

## 5.3 Aislamiento multiempresa

- Todo endpoint empresarial exige `empresa_id`.
- Toda consulta/mutacion aplica filtro por empresa.
- Auditoria de denegaciones fuera de alcance.

## 5.4 Seguridad de datos sensibles

- No exponer secretos en texto plano.
- Uso de referencias seguras para credenciales de terceros.
- Registro de cambios criticos con metadata de aprobacion.

## 6. Procesos tecnicos transversales

## 6.1 Auditoria empresarial

- Registro no bloqueante de acciones criticas C/U/D/A.
- Filtros por modulo, recurso, actor, codigo http y request id.
- Exportacion operativa y forense.
- Retencion manual y automatica programada.

## 6.2 Contabilidad automatizada

- Emision de eventos contables por modulo.
- Worker de asientos por lotes con idempotencia y reintentos.
- Conciliacion periodo-evento-asiento.

## 6.3 Backups empresariales

- Snapshots por empresa con metadata.
- Restauracion con bitacora de restauraciones.
- Exportacion de evidencia en formatos comunes.

## 6.4 Reporteria unificada

- Motor de reportes con bloques operativos, financieros y contables.
- Regla de consistencia multiformato: PDF/XLS/CSV/JSON/TXT.

## 7. Integraciones tecnicas

| Integracion | Proposito | Entradas clave | Salidas clave | Control tecnico |
|---|---|---|---|---|
| Google OAuth | login admin | client_id, redirect_uri | sesion admin | validacion callback host/proto |
| SMTP | notificaciones | destinatario, plantilla | confirmacion envio | logging y manejo de fallo |
| Wompi/Mercado Pago | cobros | orden, monto, referencia | estado pago | trazabilidad por transaccion |
| DIAN | facturacion colombia | NIT, software id/pin, firma | acuse, estado documento | validacion y contingencia |
| OpenStreetMap | mapa operativo | coordenadas/rutas | visualizacion mapa | modo degradado en fallo |
| IA externa | asistencia negocio | pregunta, contexto empresa | respuesta y uso | cuota/limite y auditoria |

## 8. Requerimientos no funcionales

## 8.1 Disponibilidad y resiliencia

- Arranque controlado del backend.
- Recovery con backups y restauraciones trazables.
- Reintentos en procesos asincronos criticos.

## 8.2 Rendimiento

- Consultas con indices por empresa y rango temporal.
- Paginacion en consultas pesadas.
- Endpoints de resumen para tableros directivos.

## 8.3 Observabilidad

- Logs estructurados por modulo y severidad.
- Evidencia de errores en backend y mensajes claros en frontend.
- Historico de eventos para diagnostico operacional.

## 8.4 Cumplimiento y gobierno de datos

- Trazabilidad por empresa, documento y periodo.
- Controles de retencion y exportacion de evidencia.
- Compatibilidad de intercambio con software contable externo.

## 9. Mapa tecnico por dominio

| Dominio | Capa API | Capa DB | Entidades nucleares |
|---|---|---|---|
| Seguridad | handlers de auth/seguridad | db sesiones/usuarios/roles | administradores, users, sesiones |
| Ventas | handlers carritos/ventas | db carritos/items | carritos_compras, carrito_compra_items |
| Inventario | handlers inventario/productos | db productos/inventario | productos, bodegas, inventario_movimientos |
| Compras | handlers proveedores/compras | db compras/proveedores | proveedores, empresa_compras_documentos |
| Finanzas | handlers finanzas | db finanzas/eventos/asientos | empresa_finanzas_movimientos, empresa_eventos_contables |
| Facturacion | handlers facturacion/DIAN | db documentos FE | empresa_facturacion_documentos, empresa_dian_configuracion |
| Creditos | handlers creditos | db creditos/cartera | empresa_creditos, cuotas, movimientos |
| Auditoria | handlers auditoria | db auditoria | empresa_auditoria_eventos |
| Backups | handlers backups | db backups | empresa_backups, empresa_backups_restauraciones |

## 10. Estrategia de evolucion tecnica

1. Mantener `main.go` como orquestador liviano de arranque y rutas.
2. Extender dominios en archivos especializados por responsabilidad.
3. Aplicar cambios de schema con trazabilidad documental y pruebas.
4. Validar siempre impacto en permisos, aislamiento y reporteria.
5. Mantener actualizados diagramas y documentos canonicos relacionados.

## 11. Riesgos tecnicos y mitigacion

- Riesgo: divergencia entre documentacion y rutas reales.
  - Mitigacion: validacion cruzada por endpoint/handler en revisiones.
- Riesgo: regresion de permisos en cambios de modulo.
  - Mitigacion: pruebas de middleware de permisos por rol/accion.
- Riesgo: inconsistencia entre formatos de reporte.
  - Mitigacion: set de pruebas para comparacion de columnas/totales.
- Riesgo: dependencia de integracion externa no disponible.
  - Mitigacion: modo degradado, cola de reintentos y trazabilidad de error.

## 12. Criterios de aceptacion tecnica

- Arquitectura por capas y dominios claramente definida.
- Aislamiento por empresa y seguridad documentados en toda operacion.
- Integraciones externas con contrato tecnico y controles operativos.
- Reglas de datos y contabilidad reflejadas en diseno.
- Base apta para ejecucion de implementacion incremental.
