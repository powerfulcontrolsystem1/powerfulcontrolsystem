# Arquitectura movil v1

Actualizacion: 2026-07-15.

La aplicacion Flutter y el portal web consumen la fachada `/api/v1/` sin
duplicar reglas POS. La API usa JSON uniforme, autenticacion de dispositivo,
permisos empresariales validados en servidor, paginacion/filtros y
`Idempotency-Key` persistente para escrituras que puedan reintentarse.

Las mutaciones de carrito, cobro, venta, sincronizacion offline, documentos y
notificaciones deben incluir una clave idempotente por empresa. Una repeticion
legitima devuelve el resultado compatible y no crea una segunda venta, pago,
movimiento ni mensaje. La empresa se obtiene del token/sesion validada; nunca
desde un encabezado o cuerpo impuesto por el cliente.

El contrato vivo se mantiene en `documentos/api/openapi.mobile.v1.yaml` y
`documentos/api/mobile_api_v1.md`. Toda ampliacion movil se agrega de forma
aditiva a v1, con cursor para listados grandes, seleccion de campos cuando
aplique, limites de archivo, errores JSON consistentes y prueba de aislamiento
entre empresas.
