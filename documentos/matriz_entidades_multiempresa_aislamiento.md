# Matriz de entidades multiempresa y llaves de aislamiento por endpoint

Fecha de actualizacion: 2026-04-04
Alcance: punto 2 del plan maestro (aislamiento por empresa y llaves de alcance por endpoint)

## Regla general

- Regla primaria: toda operacion empresarial debe resolverse por `empresa_id` (query, body JSON o multipart form).
- Regla secundaria: cada modulo agrega llaves de recurso/contexto (`id`, `carrito_id`, `conversacion_id`, `sucursal_id`, `bodega_id`, etc.).
- Regla de middleware: las rutas registradas con `WithEmpresa*Permissions` validan autenticacion, modulo/accion y alcance por `empresa_id`.
- Excepciones controladas:
  - `/api/empresa/facturacion_electronica/paises_disponibles` es catalogo global.
  - `/api/empresa/usuarios/login` y `/api/empresa/usuarios/establecer_password` no usan wrapper de permisos, pero aplican validacion de alcance por usuario/empresa en handler.
  - rutas de Chat IA no usan wrapper de permisos tradicional; aplican validacion de alcance con cuenta Google autenticada.

## Matriz formal por endpoint

| Endpoint | Metodos | Llave primaria de aislamiento | Llaves secundarias de recurso/alcance | Control de alcance | Fuente tecnica |
|---|---|---|---|---|---|
| `/api/empresa/bodegas` | GET/POST/PUT/DELETE | `empresa_id` | `id` (PUT action, DELETE) | `WithEmpresaInventarioPermissions` | `EmpresaBodegasHandler` |
| `/api/empresa/categorias_productos` | GET/POST/PUT/DELETE | `empresa_id` | `id` (PUT action, DELETE) | `WithEmpresaInventarioPermissions` | `EmpresaCategoriasProductosHandler` |
| `/api/empresa/productos` | GET/POST/PUT/DELETE | `empresa_id` | `id`, `bodega_id`, `categoria_id` | `WithEmpresaInventarioPermissions` | `EmpresaProductosHandler` |
| `/api/empresa/productos/imagen` | POST (multipart) | `empresa_id` (form) | `producto_id` (form) | `WithEmpresaInventarioPermissions` | `EmpresaProductoImagenUploadHandler` |
| `/api/empresa/inventario/existencias` | GET | `empresa_id` | `producto_id`, `bodega_id` | `WithEmpresaInventarioPermissions` | `EmpresaInventarioExistenciasHandler` |
| `/api/empresa/inventario/movimientos` | GET | `empresa_id` | `producto_id` | `WithEmpresaInventarioPermissions` | `EmpresaInventarioMovimientosHandler` |
| `/api/empresa/inventario/transferir` | POST | `empresa_id` (body) | `producto_id`, `bodega_origen_id`, `bodega_destino_id` | `WithEmpresaInventarioPermissions` | `EmpresaInventarioTransferHandler` |
| `/api/empresa/inventario/ajustar` | POST | `empresa_id` (body) | `producto_id`, `bodega_id` | `WithEmpresaInventarioPermissions` | `EmpresaInventarioAjusteHandler` |
| `/api/empresa/inventario/cambiar_producto` | POST | `empresa_id` (body) | `producto_origen_id`, `producto_destino_id`, `bodega_id` | `WithEmpresaInventarioPermissions` | `EmpresaInventarioCambioProductoHandler` |
| `/api/empresa/productos/precios_historial` | GET | `empresa_id` | `producto_id` | `WithEmpresaInventarioPermissions` | `EmpresaProductoPrecioHistorialHandler` |
| `/api/empresa/proveedores` | GET/POST/PUT/DELETE | `empresa_id` | `id/proveedor_id`, `documento_codigo` (acciones transaccionales) | `WithEmpresaComprasPermissions` | `EmpresaProveedoresHandler` |
| `/api/empresa/servicios` | GET/POST/PUT/DELETE | `empresa_id` | `id` (PUT action, DELETE) | `WithEmpresaInventarioPermissions` | `EmpresaServiciosHandler` |
| `/api/empresa/usuarios/login` | POST | `empresa_id` (body o query opcional) | `email` | Validacion de usuario por alcance (`GetEmpresaUsuarioByEmailScoped`) | `EmpresaUsuarioLoginHandler` |
| `/api/empresa/usuarios/establecer_password` | POST | `empresa_id` (body o query opcional) | `email`, `documento_identidad` | Validacion de usuario por alcance (`GetEmpresaUsuarioByEmailScoped`) | `EmpresaUsuarioSetPasswordHandler` |
| `/api/empresa/usuarios` | GET/POST/PUT/DELETE | `empresa_id` | `id` (PUT/DELETE y acciones), `rol_usuario_id` (body) | `WithEmpresaSeguridadPermissions` | `EmpresaUsuariosHandler` |
| `/api/empresa/clientes` | GET/POST/PUT/DELETE | `empresa_id` | `id` (PUT/DELETE y acciones) | `WithEmpresaClientesPermissions` | `EmpresaClientesHandler` |
| `/api/empresa/configuracion_avanzada` | GET/POST/PUT | `empresa_id` | n/a (config por empresa) | `WithEmpresaSeguridadPermissions` | `EmpresaConfiguracionAvanzadaHandler` |
| `/api/empresa/roles_de_usuario` | GET | `empresa_id` | `include_inactive` (filtro) | `WithEmpresaSeguridadPermissions` | `EmpresaRolesDeUsuarioHandler` |
| `/api/empresa/auditoria/eventos` | GET/PUT/POST | `empresa_id` | `recurso_id`, `codigo_http`, `retencion_dias` | `WithEmpresaSeguridadPermissions` | `EmpresaAuditoriaEventosHandler` |
| `/api/empresa/carritos_compra` | GET/POST/PUT/DELETE | `empresa_id` | `id`, `action`, `reset_items` | `WithEmpresaVentasPermissions` | `EmpresaCarritosCompraHandler` |
| `/api/empresa/carritos_compra/items` | GET/POST/PUT/DELETE | `empresa_id` | `carrito_id`, `id`, `action` | `WithEmpresaVentasPermissions` | `EmpresaCarritoItemsHandler` |
| `/api/empresa/chat_tareas/conversaciones` | GET/POST/PUT/DELETE | `empresa_id` | `id`, `action` | `WithEmpresaVentasPermissions` | `EmpresaChatTareasConversacionesHandler` |
| `/api/empresa/chat_tareas/participantes` | GET/POST/PUT/DELETE | `empresa_id` | `conversacion_id`, `id`, `participante_ref_id` | `WithEmpresaVentasPermissions` | `EmpresaChatTareasParticipantesHandler` |
| `/api/empresa/chat_tareas/mensajes` | GET/POST/PUT/DELETE | `empresa_id` | `conversacion_id`, `id` | `WithEmpresaVentasPermissions` | `EmpresaChatTareasMensajesHandler` |
| `/api/empresa/chat_tareas/mensajes/adjunto` | POST (multipart) | `empresa_id` (form) | `conversacion_id` (form), `autor_ref_id` | `WithEmpresaVentasPermissions` | `EmpresaChatTareasAdjuntoUploadHandler` |
| `/api/empresa/chat_tareas/tareas` | GET/POST/PUT/DELETE | `empresa_id` | `id`, `conversacion_id`, `action` | `WithEmpresaVentasPermissions` | `EmpresaChatTareasTareasHandler` |
| `/api/empresa/ubicacion_gps/dispositivos` | GET/POST/PUT/DELETE | `empresa_id` | `id` | `WithEmpresaInventarioPermissions` | `EmpresaUbicacionGPSDispositivosHandler` |
| `/api/empresa/ubicacion_gps/recorridos` | GET/POST/PUT/DELETE | `empresa_id` | `dispositivo_id`, `id` | `WithEmpresaInventarioPermissions` | `EmpresaUbicacionGPSRecorridosHandler` |
| `/api/empresa/facturacion_electronica` | GET/POST/PUT | `empresa_id` | `pais_codigo`, `documento_codigo`, `entidad_id`, `action` | `WithEmpresaFacturacionPermissions` | `EmpresaFacturacionElectronicaHandler` |
| `/api/empresa/facturacion_electronica/pais_detectado` | GET | `empresa_id` (opcional) | `tz`, `lang` | `WithEmpresaFacturacionPermissions` | `EmpresaFacturacionElectronicaPaisDetectadoHandler` |
| `/api/empresa/facturacion_electronica/paises_disponibles` | GET | n/a (catalogo global) | n/a | sin wrapper (catalogo) | `EmpresaFacturacionElectronicaPaisesDisponiblesHandler` |
| `/api/empresa/finanzas/movimientos` | GET/POST/PUT/DELETE | `empresa_id` | `id`, `action`, `desde`, `hasta`, `periodo`, `tipo` | `WithEmpresaFinanzasPermissions` | `EmpresaFinanzasMovimientosHandler` |
| `/api/empresa/finanzas/configuracion` | GET/POST/PUT | `empresa_id` | n/a (config por empresa) | `WithEmpresaFinanzasPermissions` | `EmpresaFinanzasConfiguracionHandler` |
| `/api/empresa/finanzas/periodos` | GET/POST/PUT | `empresa_id` | `periodo`, `action` | `WithEmpresaFinanzasPermissions` | `EmpresaFinanzasPeriodosHandler` |
| `/api/empresa/finanzas/asientos_contables` | GET/POST/PUT | `empresa_id` | `action`, `periodo`, `limit`, `max_reintentos` | `WithEmpresaFinanzasPermissions` | `EmpresaFinanzasAsientosContablesHandler` |
| `/api/empresa/finanzas/cierres_caja` | GET/POST/PUT/DELETE | `empresa_id` | `id`, `sucursal_id`, `caja_codigo`, `estado_cierre`, `action` | `WithEmpresaFinanzasPermissions` | `EmpresaFinanzasCierresCajaHandler` |
| `/api/empresa/chat_con_inteligencia_artificial/modelos` | GET | `empresa_id` | cuenta Google autenticada | validacion interna `ensureEmpresaAccessByAccount` | `ModelosHandler` |
| `/api/empresa/chat_con_inteligencia_artificial/modelo_preferido` | GET/PUT | `empresa_id` | `model_id`, cuenta Google autenticada | validacion interna `ensureEmpresaAccessByAccount` | `ModeloPreferidoHandler` |
| `/api/empresa/chat_con_inteligencia_artificial/consultar` | POST | `empresa_id` | `model_id`, `pregunta`, cuenta Google autenticada | validacion interna `ensureEmpresaAccessByAccount` | `ConsultarHandler` |
| `/api/empresa/chat_con_inteligencia_artificial/historial` | GET | `empresa_id` | `limit`, cuenta Google autenticada | validacion interna `ensureEmpresaAccessByAccount` | `HistorialHandler` |

## Observaciones de consistencia de aislamiento

- Las rutas multipart (`productos/imagen`, `chat_tareas/mensajes/adjunto`) extraen `empresa_id` desde formulario y exponen `X-Empresa-ID` en respuesta para trazabilidad de middleware.
- En rutas con acciones (`action=...`), la combinacion `empresa_id + id` (o equivalente) es la llave operativa de mutacion.
- Los endpoints de login/set_password aceptan `empresa_id` opcional y limitan el acceso por matching de usuario en DB, evitando lectura transversal de empresas.
- Chat IA aplica aislamiento por `empresa_id` y por cuenta Google administradora con validacion explicita de alcance por empresa.
