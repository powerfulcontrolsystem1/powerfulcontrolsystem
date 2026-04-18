# Contrato tecnico: estaciones, sensores y ventas simples por estacion

Fecha: 2026-04-18
Estado: vigente

## Alcance

Este contrato cubre la configuracion de estaciones por empresa, la asociacion de sensores de puertas a estaciones, la sincronizacion del carrito base por estacion y la operacion de venta simple sobre ese carrito enlazado.

## Rutas implicadas

- `GET /api/empresa/estacion_prefs`
- `PUT /api/empresa/estacion_prefs`
- `GET /api/empresa/sensor_puertas`
- `POST /api/public/sensor_puertas?action=heartbeat`
- `GET /api/empresa/carritos_compra`
- `GET /api/empresa/carritos_compra?action=totales_pago`
- `GET /api/empresa/carritos_compra?action=metricas_estacion`
- `PUT /api/empresa/carritos_compra?action=activar_estacion`
- `PUT /api/empresa/carritos_compra?action=pagar_estacion`
- `PUT /api/empresa/carritos_compra?action=recuperar_interrumpido`
- `PUT /api/empresa/carritos_compra?action=anular_cierre_parcial`
- `GET /api/empresa/carritos_compra/items`
- `POST /api/empresa/carritos_compra/items`
- `PUT /api/empresa/carritos_compra/items`
- `DELETE /api/empresa/carritos_compra/items`
- `GET /api/empresa/productos`
- frontend de apoyo: `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/ventas_simple.html`, `web/js/ventas_simple.js`

## Entradas obligatorias

### Configuracion de estaciones

- `empresa_id`: obligatorio en toda ruta protegida del flujo.
- `estacion_id=0`: obligatorio para la preferencia global `estaciones_config`.
- `clave=estaciones_config`: obligatoria cuando se persiste la definicion completa de estaciones.
- `valor`: JSON serializado con `cantidad` y arreglo `estaciones`.

### Operacion de venta simple

- `empresa_id`: obligatorio.
- `id`: obligatorio en acciones sobre un carrito concreto.
- `action`: obligatorio para `activar_estacion`, `pagar_estacion`, `recuperar_interrumpido` o `anular_cierre_parcial`.

### Cierre de venta por estacion

- `metodo_pago`: obligatorio en `pagar_estacion`.
- `total_pagado`: obligatorio cuando el metodo no cubre el total por reglas automaticas.
- `referencia_pago`: obligatoria para `tarjeta_credito`, `tarjeta_debito` y `transferencia_bancaria` con minimo 4 caracteres.

## Entradas opcionales

- `reset_items=1`: reinicia la sesion al activar una estacion ya utilizada.
- `descuento_tipo`, `descuento_codigo`, `codigo_descuento`, `descuento_valor`.
- `devolucion_total`.
- `aplicar_propina`.
- `usuario_lavador`.
- `pagos_mixtos` o `pagos` cuando `metodo_pago=mixto`.
- `q`, `limit`, `days`, `include_inactive`, `estacion_id` en consultas auxiliares.

## Salidas y estados funcionales

Estados operativos del carrito:

- registro: `activo` o `inactivo`
- operacion: `abierto` o `cerrado`
- venta: `venta_abierta`, `venta_pagada` o recuperada segun accion aplicada

Estados funcionales del sensor en la vista:

- indicador negro: sin dispositivo util, sin actividad valida reciente o ultimo estado no activo
- indicador verde: dispositivo asociado a la estacion con `last_state` activo y `last_seen` fresco

Respuestas exitosas esperadas:

- `200` con `sync` al guardar `estaciones_config`.
- `200` con `estado`, `estado_carrito` y `estado_venta` en acciones de carrito.
- `200` con `documento_venta`, `propina`, `comision` y `configuracion_operativa` al cerrar `pagar_estacion`.
- `200` con resumen y filas en `metricas_estacion` y `totales_pago`.

## Invariantes

1. `empresa_id` es la frontera obligatoria del flujo; una estacion, un sensor, un carrito o una metrica no pueden resolverse fuera de su empresa.
2. La configuracion maestra de estaciones vive en `empresa_estacion_prefs` bajo `estacion_id=0` y `clave=estaciones_config`.
3. Guardar `estaciones_config` debe ejecutar sincronizacion backend de carritos; no se permite depender de una sincronizacion exclusiva del frontend.
4. Cada estacion configurada debe tener a lo sumo un carrito base canonico por empresa con identidad `codigo=EST-empresa-estacion` y `referencia_externa=ESTACION_<id>`.
5. El carrito base enlazado debe quedar en estado base `inactivo/cerrado` hasta que una accion operativa lo active.
6. El render de estaciones debe abrir la venta usando el carrito enlazado por identidad canonica, no por texto libre de la tarjeta.
7. La vista de estaciones solo debe mostrar el indicador cuadrado superior derecho como evidencia visual del sensor.
8. El estado del sensor debe resolverse por la lectura mas reciente de `last_seen`; lecturas antiguas no deben seguir pintando la estacion como activa.
9. `pagar_estacion` solo puede cerrar una sesion operativa valida y debe respetar metodos de pago habilitados para la empresa y el rol efectivo.
10. `pagar_estacion` debe registrar metrica operativa por estacion y generar automaticamente el documento de venta segun `modo_documento_venta`.
11. `anular_cierre_parcial` y `recuperar_interrumpido` deben dejar trazabilidad operativa y tambien registrar metrica por estacion.
12. El frontend de venta simple puede operar con cola local para items y activacion diferida, pero el cierre de cobro requiere conectividad y confirmacion backend.

## Side effects obligatorios

- upsert de `empresa_estacion_prefs`
- sincronizacion de carritos base enlazados por estacion
- lectura de sensores en `empresa_sensor_puertas_devices`
- altas, edicion o eliminacion de items del carrito
- cierre operativo del carrito con validaciones de metodo, descuento, propina, comision y devolucion
- escritura en `empresa_ventas_estacion_metricas`
- generacion de `factura_electronica` o `comprobante_pago` segun configuracion
- auditoria y evento contable del cierre o correccion

## Errores de contrato esperados

- `400` por `empresa_id` ausente, `id` ausente, JSON invalido, metodo de pago invalido, referencia insuficiente o suma invalida de pago mixto.
- `403` cuando el metodo de pago no esta habilitado para la empresa o rol efectivo.
- `404` cuando el carrito no existe en la empresa indicada.
- `409` cuando la transicion de estado no es valida, la venta ya fue pagada o la sesion ya estaba activa sin `reset_items=1`.
- `500` cuando falla la sincronizacion de carritos, el registro de venta, la generacion documental o una autorreparacion de esquema necesaria para la operacion.

## Reglas de compatibilidad

- el flujo debe tolerar configuraciones legacy de `estaciones_config` y normalizar `estado` vacio como `activo` en preferencias de estacion.
- la sincronizacion de carritos y el registro de metricas deben seguir siendo compatibles con PostgreSQL como runtime canonico.
- las inserciones del flujo no deben depender de `LastInsertId` cuando el runtime sea PostgreSQL.

## Evidencia tecnica minima

- pruebas de `empresa_estacion_prefs` para persistencia, aislamiento por `empresa_id` y sincronizacion de carritos base.
- pruebas del handler de carritos para `pagar_estacion`, `anular_cierre_parcial`, `recuperar_interrumpido` y validacion de metodos permitidos.
- diagnostico del editor limpio en documentos y vistas tocadas cuando se cambie el flujo.
- validacion manual de que `estaciones.html` abre la estacion correcta y `ventas_simple.html` puede volver a `estaciones.html` manteniendo `empresa_id`.

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`

## Runbook relacionado

- `documentos/gobernanza_tecnica/runbooks/runbook_estaciones_sensores_ventas_simple.md`