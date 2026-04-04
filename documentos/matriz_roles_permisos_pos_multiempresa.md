# Matriz base de roles y permisos POS multiempresa

Fecha de actualizacion: 2026-04-03
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
	- Finanzas: `/api/empresa/finanzas/movimientos`, `/api/empresa/finanzas/configuracion`, `/api/empresa/finanzas/periodos`.
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

## Reglas de seguridad obligatorias

1. Todo endpoint debe validar empresa_id y, cuando aplique, sucursal_id antes de operar.
2. Ningun usuario puede actuar fuera de su alcance de empresa/sucursal.
3. Toda accion critica debe dejar auditoria con request_id, empresa_id, usuario, accion y timestamp.
4. Operaciones de cierre/aprobacion deben requerir rol con permiso A.
5. Eliminaciones funcionales deben implementarse como anulacion/inactivacion cuando aplique trazabilidad legal.

## Acciones tecnicas siguientes (continuacion punto 3)

1. Traducir esta matriz a permisos concretos por endpoint API.
2. Definir middleware de autorizacion por recurso (empresa/sucursal) y por accion (C/R/U/D/A).
3. Agregar pruebas automatizadas de permisos denegados y permitidos por rol.
4. Publicar catalogo de permisos en frontend para ocultar opciones no autorizadas por rol.
