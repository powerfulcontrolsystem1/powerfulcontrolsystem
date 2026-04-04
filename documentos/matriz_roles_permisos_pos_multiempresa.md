# Matriz base de roles y permisos POS multiempresa

Fecha de actualizacion: 2026-04-04
Alcance: punto 3 del plan maestro (permisos y seguridad)

## Roles base

| Rol | Alcance | Descripcion |
|---|---|---|
| super_administrador | global | administra configuracion, empresas, licencias, auditoria y seguridad global |
| admin_empresa | empresa | administra configuracion, catalogos, usuarios y cierres de su empresa |
| supervisor_sucursal | sucursal | supervisa operacion, aprueba cierres y movimientos criticos |
| cajero | sucursal/caja | registra ventas, cobros, devoluciones permitidas y cierre de caja |
| inventario | sucursal/bodega | gestiona productos, existencias y movimientos de bodega |
| compras | empresa/sucursal | crea ordenes de compra, recepciones y ajustes de costo |
| contabilidad | empresa | valida asientos, periodos y reportes financieros |
| auditor | empresa/global | consulta reportes, logs y trazabilidad sin modificar datos |

## Permisos por modulo

Leyenda:
- C: crear
- R: leer
- U: actualizar
- D: eliminar/anular
- A: aprobar/cerrar

| Modulo | super_administrador | admin_empresa | supervisor_sucursal | cajero | inventario | compras | contabilidad | auditor |
|---|---|---|---|---|---|---|---|---|
| Ventas POS | CRUDA | CRUA | CRUA | CRU | R | R | R | R |
| Inventarios | CRUDA | CRUA | CRUA | R | CRUDA | R | R | R |
| Clientes | CRUDA | CRUA | CRUA | CRU | R | R | R | R |
| Proveedores | CRUDA | CRUA | R | R | R | CRUA | R | R |
| Compras | CRUDA | CRUA | CRUA | R | R | CRUDA | R | R |
| Facturacion electronica | CRUDA | CRUA | R | CRU | R | R | R | R |
| Contabilidad y periodos | CRUDA | CRUA | R | R | R | R | CRUDA | R |
| Reportes financieros | CRUA | CRUA | R | R | R | R | CRUA | R |
| Cierres de caja | CRUDA | CRUA | CRUA | CRUA | R | R | R | R |
| Seguridad y permisos | CRUDA | CRUA | R | R | R | R | R | R |

## Estado de implementacion tecnica inicial (2026-04-04)

- Se implementa middleware en `backend/handlers/empresa_permisos.go` para validar:
	- identidad administrativa activa,
	- alcance de `empresa_id`,
	- permisos por rol/accion (C/R/U/D/A) por modulo.
- Cobertura inicial aplicada en `backend/main.go` sobre rutas criticas:
	- Ventas: `/api/empresa/carritos_compra`, `/api/empresa/carritos_compra/items`.
	- Inventario: `/api/empresa/bodegas`, `/api/empresa/categorias_productos`, `/api/empresa/productos`, `/api/empresa/inventario/*`, `/api/empresa/productos/precios_historial`.
	- Finanzas: `/api/empresa/finanzas/movimientos`, `/api/empresa/finanzas/configuracion`, `/api/empresa/finanzas/periodos`, `/api/empresa/finanzas/asientos_contables`.
- Cobertura ampliada (2026-04-04):
	- Clientes: `/api/empresa/clientes`.
	- Compras/Proveedores: `/api/empresa/proveedores`.
	- Facturacion: `/api/empresa/facturacion_electronica`, `/api/empresa/facturacion_electronica/pais_detectado`.
	- Servicios de catalogo: `/api/empresa/servicios` bajo politica de inventario.
- Cobertura adicional (2026-04-04 - cierre de rutas pendientes):
	- Seguridad/usuarios:
		- `/api/empresa/usuarios`.
		- `/api/empresa/configuracion_avanzada`.
		- `/api/empresa/roles_de_usuario`.
		- `/api/empresa/auditoria/eventos`.
	- Inventario:
		- `/api/empresa/productos/imagen`.
		- `/api/empresa/ubicacion_gps/dispositivos`.
		- `/api/empresa/ubicacion_gps/recorridos`.
	- Colaboracion operativa (politica ventas):
		- `/api/empresa/chat_tareas/conversaciones`.
		- `/api/empresa/chat_tareas/participantes`.
		- `/api/empresa/chat_tareas/mensajes`.
		- `/api/empresa/chat_tareas/mensajes/adjunto`.
		- `/api/empresa/chat_tareas/tareas`.
- Cobertura automatizada inicial en `backend/handlers/empresa_permisos_test.go`:
	- denegacion de escritura sin permiso por rol,
	- aprobacion permitida para rol contabilidad en cierre de periodos,
	- bloqueo por fuera de alcance de empresa.
	- denegacion/escritura por rol en modulos `compras` y `facturacion`, y aprobacion de escritura en `clientes` para `cajero` segun matriz.
	- denegacion de escritura en modulo seguridad para `supervisor_sucursal`.
	- aprobacion permitida en modulo seguridad para `admin_empresa`.
	- denegacion para `cajero` al procesar asientos (`action=procesar_asientos`) en modulo finanzas.
	- aprobacion para `contabilidad` al procesar asientos (`action=procesar_asientos`) en modulo finanzas.
	- registro automatico de auditoria para acciones criticas autorizadas (`C/U/D/A`) en middleware de permisos empresariales.
	- cobertura de auditoria automatica por modulo con pruebas en `backend/handlers/auditoria_empresa_test.go` para:
		- `ventas` (`action=cerrar`),
		- `compras` (`action=emitir_orden`),
		- `facturacion` (`action=emitir`).

## Matriz UAT de cierres de caja (roles y transiciones)

Fecha de actualizacion: 2026-04-04

### Casos por rol en endpoint `/api/empresa/finanzas/cierres_caja`

| Caso | Rol | Metodo/accion | Resultado esperado |
|---|---|---|---|
| UAT-CC-R1 | cajero | `PUT action=aprobar` | `403 forbidden` |
| UAT-CC-R2 | supervisor_sucursal | `PUT action=aprobar` | `403 forbidden` |
| UAT-CC-R3 | admin_empresa | `PUT action=aprobar` | `200 ok` |

### Casos de transicion del estado de cierre

| Caso | Estado actual | Accion | Precondicion | Resultado esperado |
|---|---|---|---|---|
| UAT-CC-T1 | abierto | aprobar | ninguna | `409 conflict` (transicion invalida) |
| UAT-CC-T2 | abierto | cerrar | `caja_fisica` valida | `200 ok`, estado `cerrado` |
| UAT-CC-T3 | cerrado | aprobar | ninguna | `200 ok`, estado `aprobado` |
| UAT-CC-T4 | aprobado | reabrir | ninguna | `200 ok`, estado `abierto` |
| UAT-CC-T5 | aprobado | editar/eliminar | ninguna | bloqueo (`409`/error de negocio) |

## Matriz final endpoint/rol (implementacion vigente 2026-04-04)

Leyenda de roles:
- SA: super_administrador
- AE: admin_empresa
- SS: supervisor_sucursal
- CJ: cajero
- IN: inventario
- CO: compras
- CT: contabilidad
- AU: auditor

Regla de lectura comun (R):
- En rutas con wrapper de permisos, lectura queda habilitada para SA, AE, SS, CJ, IN, CO, CT y AU.

| Endpoint | Wrapper/modulo | C/U/A habilitado | D habilitado | Observaciones de accion |
|---|---|---|---|---|
| `/api/empresa/carritos_compra` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | `action=cerrar|reabrir|pagar_estacion|activar_estacion|pagar|suspender|reactivar` exige `A` |
| `/api/empresa/carritos_compra/items` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | mutaciones de items bajo politica de ventas |
| `/api/empresa/chat_tareas/conversaciones` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/chat_tareas/participantes` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/chat_tareas/mensajes` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/chat_tareas/mensajes/adjunto` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | multipart con `empresa_id` obligatorio |
| `/api/empresa/chat_tareas/tareas` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/bodegas` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | CRUD inventario |
| `/api/empresa/categorias_productos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | CRUD inventario |
| `/api/empresa/productos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | CRUD inventario |
| `/api/empresa/productos/imagen` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | upload multipart |
| `/api/empresa/servicios` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | catalogo operativo en politica inventario |
| `/api/empresa/inventario/existencias` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | lectura y mutaciones bajo modulo inventario |
| `/api/empresa/inventario/movimientos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | lectura y mutaciones bajo modulo inventario |
| `/api/empresa/inventario/transferir` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | transferencias de bodega |
| `/api/empresa/inventario/ajustar` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | ajustes de existencias |
| `/api/empresa/inventario/cambiar_producto` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | remapeo operativo producto/bodega |
| `/api/empresa/productos/precios_historial` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | historial de precios |
| `/api/empresa/ubicacion_gps/dispositivos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | geolocalizacion en politica inventario |
| `/api/empresa/ubicacion_gps/recorridos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | geolocalizacion en politica inventario |
| `/api/empresa/clientes` | `WithEmpresaClientesPermissions` | SA, AE, SS, CJ | - | modulo clientes sin `D` por politica actual |
| `/api/empresa/proveedores` | `WithEmpresaComprasPermissions` | SA, AE, SS, CO | - | `action=emitir_orden|recepcionar_compra|contabilizar_compra|aprobar` exige `A` |
| `/api/empresa/facturacion_electronica` | `WithEmpresaFacturacionPermissions` | SA, AE, CJ | - | `action=emitir|nota_credito|emitir_factura|emitir_documento` exige `A` |
| `/api/empresa/facturacion_electronica/pais_detectado` | `WithEmpresaFacturacionPermissions` | SA, AE, CJ | - | consulta/actualizacion bajo politica facturacion |
| `/api/empresa/finanzas/movimientos` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | `action=cerrar|reabrir|aprobar|procesar_asientos|procesar` exige `A` |
| `/api/empresa/finanzas/configuracion` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | configuracion financiera |
| `/api/empresa/finanzas/periodos` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | cierre/reapertura de periodos en `A` |
| `/api/empresa/finanzas/asientos_contables` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | `action=procesar_asientos` validado por rol |
| `/api/empresa/finanzas/cierres_caja` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | `action=aprobar` restringido por permiso `A` |
| `/api/empresa/usuarios` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | seguridad/usuarios solo administracion empresa |
| `/api/empresa/configuracion_avanzada` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | seguridad/configuracion sensible |
| `/api/empresa/roles_de_usuario` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | consulta catalogo de roles con control de alcance |
| `/api/empresa/auditoria/eventos` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | consulta y retencion (`action=retener|purgar`) |

### Endpoints fuera de wrapper (control alterno)

| Endpoint | Control aplicado | Nota |
|---|---|---|
| `/api/empresa/usuarios/login` | validacion de alcance por usuario/empresa en handler | sin middleware de modulo |
| `/api/empresa/usuarios/establecer_password` | validacion de alcance por usuario/empresa en handler | sin middleware de modulo |
| `/api/empresa/facturacion_electronica/paises_disponibles` | catalogo global | sin `empresa_id` obligatorio |
| `/api/empresa/chat_con_inteligencia_artificial/modelos` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |
| `/api/empresa/chat_con_inteligencia_artificial/modelo_preferido` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |
| `/api/empresa/chat_con_inteligencia_artificial/consultar` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |
| `/api/empresa/chat_con_inteligencia_artificial/historial` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |

## Checklist UAT de Punto 3 (permisos y seguridad)

| ID | Verificacion | Estado | Evidencia automatizada |
|---|---|---|---|
| P3-UAT-01 | Denegar escritura inventario a `cajero` | ok | `TestWithEmpresaInventarioPermissionsDeniesCajeroWrite` |
| P3-UAT-02 | Denegar escritura GPS a `cajero` | ok | `TestWithEmpresaInventarioPermissionsDeniesCajeroWriteGPS` |
| P3-UAT-03 | Permitir chat adjunto a `cajero` autenticado | ok | `TestWithEmpresaVentasPermissionsAllowsCajeroChatAdjuntoMultipart` |
| P3-UAT-04 | Rechazar chat adjunto sin autenticacion | ok | `TestWithEmpresaVentasPermissionsRejectsChatAdjuntoWithoutAuth` |
| P3-UAT-05 | Bloquear acceso fuera de alcance de empresa | ok | `TestWithEmpresaVentasPermissionsDeniesOutOfScopeEmpresa` |
| P3-UAT-06 | Denegar `procesar_asientos` a `cajero` | ok | `TestWithEmpresaFinanzasPermissionsDeniesCajeroProcesarAsientos` |
| P3-UAT-07 | Permitir `procesar_asientos` a `contabilidad` | ok | `TestWithEmpresaFinanzasPermissionsAllowsContabilidadProcesarAsientos` |
| P3-UAT-08 | Denegar escritura seguridad a `supervisor_sucursal` | ok | `TestWithEmpresaSeguridadPermissionsDeniesSupervisorWrite` |
| P3-UAT-09 | Permitir accion de seguridad a `admin_empresa` | ok | `TestWithEmpresaSeguridadPermissionsAllowsAdminApprove` |
| P3-UAT-10 | Registrar auditoria en acciones criticas ventas/compras/facturacion | ok | `TestWithEmpresaVentasPermissionsRegistraAuditoriaAccionCritica`, `TestWithEmpresaComprasPermissionsRegistraAuditoriaAccionCritica`, `TestWithEmpresaFacturacionPermissionsRegistraAuditoriaAccionCritica` |

Ejecucion de validacion actual (2026-04-04):
- `runTests` sobre:
	- `backend/handlers/empresa_permisos_test.go`.
	- `backend/handlers/auditoria_empresa_test.go`.
- Resultado: 25 pruebas aprobadas, 0 fallidas.

## Reglas de seguridad obligatorias

1. Todo endpoint debe validar empresa_id y, cuando aplique, sucursal_id antes de operar.
2. Ningun usuario puede actuar fuera de su alcance de empresa/sucursal.
3. Toda accion critica debe dejar auditoria con request_id, empresa_id, usuario, accion y timestamp.
4. Operaciones de cierre/aprobacion deben requerir rol con permiso A.
5. Eliminaciones funcionales deben implementarse como anulacion/inactivacion cuando aplique trazabilidad legal.

## Acciones tecnicas siguientes (cierre operativo punto 3)

1. Publicar catalogo de permisos en frontend para ocultar opciones no autorizadas por rol.
2. Incorporar pruebas UAT de regresion para endpoints sin wrapper de modulo (`usuarios/login`, `establecer_password`, chat IA por cuenta Google).
3. Definir politica de aprobacion para rutas de lectura sensible en seguridad (`auditoria/eventos`) segun perfil `auditor` vs `admin_empresa`.
