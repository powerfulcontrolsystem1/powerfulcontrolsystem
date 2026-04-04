# CHANGELOG

## 2026-04-04
- Se implementa el módulo financiero multiempresa con enfoque unificado de ingresos y egresos en `web/administrar_empresa/finanzas.html`.
- Se crea `backend/db/finanzas.go` con esquema, validaciones y CRUD de:
	- `empresa_finanzas_movimientos`
	- `empresa_finanzas_configuracion`
- Se crea `backend/handlers/finanzas.go` y se publican rutas:
	- `GET/POST/PUT/DELETE /api/empresa/finanzas/movimientos`
	- `GET/POST/PUT /api/empresa/finanzas/configuracion`
- Se actualiza `backend/main.go` para asegurar el esquema financiero y registrar la migración `2026-04-03-003-finanzas`.
- Se integra el acceso al módulo en `web/administrar_empresa.html` y `web/js/administrar_empresa.js`.
- Se agrega `backend/db/finanzas_test.go` con pruebas de configuración y flujo CRUD de movimientos financieros.
- Se amplía `backend/tools/seed_motel_malibu/main.go` para sembrar configuración financiera y movimientos demo de ingreso/egreso.
- Se separa visualmente el libro financiero en dos pestañas operativas dentro del módulo: `Ingresos` y `Egresos`.
- Se agrega la pestaña `Todos` para consolidar ingresos y egresos en una sola vista del libro financiero.
- Se agrega exportación del libro financiero filtrado por fechas a:
	- Excel (CSV compatible con Excel).
	- PDF (vista de impresión).
	- JSON contable para integración externa (incluye resumen, detalle y asientos recomendados).
- Se amplía la configuración financiera por empresa para contabilidad externa con parametrización de:
	- destino de integración (`generico`, `siigo`, `world_office`, `alegra`),
	- cuentas base (caja/bancos, ingresos, IVA generado, gastos, IVA descontable),
	- cuentas por categoría para ingresos y egresos.
- La exportación `JSON contable` deja de usar cuentas fijas y ahora construye asientos con la parametrización real guardada por empresa.
- El JSON exportado incorpora `accounting_profile` y `erp_projection` por movimiento para facilitar mapeo hacia software contable externo.
- Se actualiza `backend/db/finanzas_test.go` para validar persistencia de la nueva parametrización contable.
- Se amplía `web/administrar_empresa/finanzas.html` con salidas contables adicionales:
	- Plantilla dedicada SIIGO (CSV) para importación de asientos.
	- Balance de prueba (CSV).
	- Estado de resultados (CSV).
- Se crea `documentos/plantillas/siigo_plantilla_importacion_asientos.csv` como plantilla de referencia ERP.
- Se crea `documentos/informe_contable_directivo_2026-04-03.md` con revisión de cumplimiento contable/directivo, brechas y plan recomendado.
- Validación técnica ejecutada:
	- `go test ./... -count=1` (ok).
	- `go run ./tools/seed_motel_malibu` (ok, incluye creación de 4 movimientos financieros demo).
	- `runTests` global (ok: 3/3).

## 2026-04-03
- Se implementa control de inventario en carrito: al agregar items de producto se descuenta stock y al desactivar/eliminar items abiertos se revierte automáticamente.
- Se asegura que, al cerrar una venta, el descuento de inventario permanezca aplicado y no se revierta en el pago.
- Se mejoran respuestas de API para stock insuficiente en operaciones de items de carrito.
- Se agrega `backend/db/carritos_inventario_test.go` con pruebas de descuento de inventario y caso de stock insuficiente.
- Se amplía `backend/tools/seed_motel_malibu/main.go` para registrar 10 clientes y 10 usuarios de empresa.
- La semilla valida automáticamente el flujo comercial completo: venta cerrada, descuento de inventario al agregar y persistencia tras pagar.
- Se confirma en seed la validación de impresión con vista previa POS y Carta.
- Se amplía `web/administrar_empresa/reportes.html` con reporte de ventas, reporte de productos y reporte de compra de productos, todos con búsqueda por rango de fechas.
- Validación técnica ejecutada: `go test ./auth ./db ./handlers ./metrics ./utils` (ok) y `go run ./tools/seed_motel_malibu` (ok).
- Se agrega el vínculo `Ayuda` en el menú flotante global (`web/menu.js`) y se reestructura `web/ayuda/ayuda.html` como centro de ayuda con menú interno y sección de APIs.
- Se adapta `web/administrar_empresa/carrito_de_compras.html` para operación con lector de código de barras (escaneo por código/SKU, Enter para agregar y acumulación opcional de cantidad).
- Se extiende `web/administrar_empresa/configuracion.html` con configuración por empresa para el lector: habilitar, autofoco y acumulación.
- Se amplía `web/administrar_empresa/reportes.html` con KPI de productos bajo mínimo y reporte de inventario actual por bodega.
- Validación técnica ejecutada para flujo carrito/inventario multiempresa: `go test ./db -run Carrito -count=1` (ok) y `go test ./handlers -run Carritos -count=1` (ok).

## 2026-04-02
- Se crea la herramienta `backend/tools/seed_motel_malibu/main.go` para cargar datos demo comerciales en la empresa Motel Malibu.
- La semilla inserta 10 productos con precios COP, 5 clientes y crea una venta de prueba cerrada para validar el flujo comercial.
- Se valida la configuracion de impresion con vista previa de formatos POS y Carta desde la herramienta de seed.
- Se implementa la seccion `web/administrar_empresa/reportes.html` con KPIs, ventas cerradas, top productos, top clientes y resumen de impresion.
- Se reestructura `backend/tools` en subcarpetas por herramienta para eliminar conflictos de compilación por múltiples `main`.
- Se valida backend completo con `go test ./...` (ok).
- Se valida el módulo GPS con pruebas específicas:
	- `go test ./db -run TestEmpresaGPSDispositivosYRecorridosCRUD -count=1` (ok).
	- `go test ./handlers -run TestEmpresaUbicacionGPSHandlersCRUDFlow -count=1` (ok).
- Se implementa el modulo de ubicacion GPS por empresa con soporte de multiples dispositivos.
- Se agregan tablas `empresa_gps_dispositivos` y `empresa_gps_recorridos` en `empresas.db`.
- Se crean endpoints CRUD para dispositivos y recorridos GPS en `/api/empresa/ubicacion_gps/*`.
- Se agrega la pagina `web/administrar_empresa/ubicacion_gps.html` con mapa OpenStreetMap (Leaflet).
- Se habilita tracking automatico de recorridos cada 10 segundos por dispositivo.
- Se agregan pruebas en `backend/db/ubicacion_gps_test.go` y `backend/handlers/ubicacion_gps_test.go`.
