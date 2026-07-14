# Desarrollo paralelo web y movil

Actualizacion: 2026-07-14.

Web y movil comparten reglas de negocio, permisos, auditoria y persistencia en
Go/PostgreSQL. No comparten componentes de interfaz ni duplican transacciones.

1. La web puede conservar su ruta historica mientras se crea una fachada
   `/api/v1/` equivalente, paginada y con sobre JSON estable.
2. Cada endpoint movil valida empresa en servidor y reutiliza los servicios
   existentes de venta, pago, inventario, facturacion y notificaciones.
3. Toda mutacion movil usa `Idempotency-Key`; reintentos de red no duplican una
   venta, pago, movimiento de inventario ni documento.
4. El navegador sigue usando CSRF para cookies. La aplicacion usa Bearer sobre
   TLS y no depende de cookies web.
5. Un modulo no se publica en movil hasta tener contrato OpenAPI, pruebas de
   permiso/tenant y estados visibles de carga, vacio, error y sin conexion.

El primer modulo de lectura es Productos. Las siguientes entregas priorizan POS:
ventas, carrito, medios de pago, emision, sincronizacion offline y notificaciones
push, sin retirar rutas web antes de telemetria y ventana de deprecacion.
