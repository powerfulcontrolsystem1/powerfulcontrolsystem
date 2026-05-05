# Módulo Contabilidad Colombia

## Objetivo

El módulo `contabilidad_colombia` agrega la capa contable/legal colombiana por empresa. Complementa `finanzas`: finanzas registra operación diaria, mientras contabilidad organiza PUC, terceros, impuestos, comprobantes de doble partida, asientos, libros base y cierres.

## Alcance funcional

- Configuración contable por empresa: moneda, periodo actual, versión PUC, base NIIF y bloqueo de periodos cerrados.
- PUC colombiano base con cuentas de caja, bancos, clientes, proveedores, IVA, retenciones, ingresos, costos y gastos.
- Terceros contables con documento, régimen, responsabilidades y contacto.
- Impuestos y retenciones Colombia: IVA, retefuente, reteICA y extensibles.
- Comprobantes contables: nota contable, ingreso, egreso, causación y ajuste.
- Validación de doble partida: débito y crédito deben cuadrar antes de contabilizar.
- Consulta de comprobantes y detalle de asientos.
- Cierre y reapertura controlada de periodos.
- Dashboard con cuentas, terceros, comprobantes del mes, totales y diferencia.

## Seguridad y aislamiento

- Todas las tablas incluyen `empresa_id`.
- Endpoint protegido: `/api/empresa/contabilidad_colombia`.
- Wrapper de permisos: `WithEmpresaContabilidadColombiaPermissions`.
- Módulo de licencia: `contabilidad_colombia`.
- Página de menú: `linkContabilidadColombia`.

## Archivos principales

- Base de datos: `backend/db/contabilidad_colombia.go`
- Handler: `backend/handlers/contabilidad_colombia.go`
- Ruta: `backend/main.go`
- Permisos: `backend/handlers/empresa_permisos.go`
- Interfaz: `web/administrar_empresa/contabilidad_colombia.html`
- Licencias: `web/super/licencias.html`

## Flujo operativo

1. Cargar el PUC base.
2. Crear o completar terceros.
3. Configurar impuestos y retenciones.
4. Registrar comprobantes con mínimo dos líneas.
5. Verificar que débito y crédito estén balanceados.
6. Cerrar el periodo cuando no existan diferencias.

## Pendientes naturales de evolución

- Integración automática desde ventas, compras, nómina, inventario y bancos hacia comprobantes contables.
- Exportes formales de libro diario, libro mayor, balance de prueba e información exógena.
- Integración directa con nómina electrónica, documento soporte y RADIAN cuando se active el proveedor DIAN definitivo.
