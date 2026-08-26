# Inventario avanzado

Ampliacion del modulo `inventario` para trazabilidad empresarial sin crear un inventario paralelo.

## Alcance

- Lotes por empresa, producto y bodega con codigo, fabricacion, vencimiento, cantidad, costo, calidad, proveedor, documento y ubicacion interna.
- Seriales asociados a producto, bodega y lote, con estado operativo, estado de inventario, garantia y cliente reservado.
- Reservas de inventario para ventas, pedidos, servicios u otros modulos, con control de stock libre y confirmacion de salida.
- Valorizacion por producto/bodega con cantidad disponible, reservada, libre, costo promedio y valor disponible.
- Dashboard con lotes activos, seriales, reservas, vencimientos y valor de inventario trazable.

## Rutas

- `GET /api/empresa/inventario_avanzado?action=dashboard&empresa_id=ID`
- `GET /api/empresa/inventario_avanzado?action=lotes&empresa_id=ID`
- `GET /api/empresa/inventario_avanzado?action=seriales&empresa_id=ID`
- `GET /api/empresa/inventario_avanzado?action=reservas&empresa_id=ID`
- `GET /api/empresa/inventario_avanzado?action=valorizacion&empresa_id=ID`
- `POST /api/empresa/inventario_avanzado?action=lote`
- `POST /api/empresa/inventario_avanzado?action=serial`
- `POST /api/empresa/inventario_avanzado?action=reserva`
- `POST /api/empresa/inventario_avanzado?action=confirmar_reserva`

## Integracion

`CreateEmpresaInventarioLoteAvanzado` registra entrada en `inventario_existencias` y en `inventario_movimientos`, por lo que el kardex y las existencias existentes siguen siendo la fuente operativa central.

La pantalla empresarial carga productos, bodegas, lotes, seriales y reservas
activas como selectores descriptivos. Ninguna operacion exige que el usuario
conozca IDs internos y la superficie productiva no expone creacion de datos
demo.

## Seguridad

La pagina `linkInventarioAvanzado` y la API usan el modulo/licencia existente `inventario`, con `WithEmpresaInventarioPermissions`. Todas las tablas incluyen `empresa_id`.

Las mutaciones exigen `Idempotency-Key` durable por empresa y operacion. Las
referencias a producto, servicio, bodega, categoria o proveedor de otra empresa se
rechazan como recurso no disponible, sin filtrar IDs ni convertir el rechazo
esperado en un error 500.

Los valores economicos, impuestos, umbrales y stock inicial de productos se
validan en servidor. Las escrituras propagan fallos de `RowsAffected` y los
listados no silencian errores tardios del cursor PostgreSQL.

## QA

La prueba operativa en la empresa PCS crea producto, lote, serial y reserva,
confirma la salida y valida existencias, Kardex, valorizacion y dashboard.
