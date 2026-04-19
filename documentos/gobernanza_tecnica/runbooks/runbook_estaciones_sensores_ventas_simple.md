# Runbook: estaciones, sensores y carrito unificado por estacion

Fecha: 2026-04-18
Estado: vigente

## Sintomas cubiertos

- una estacion no aparece o desaparece despues de guardar configuracion.
- una estacion abre el carrito incorrecto o no abre su carrito base.
- el indicador del sensor no se pone verde aunque exista actividad fisica.
- el carrito unificado de estacion no permite iniciar una nueva sesion o responde que la venta ya fue pagada.
- el cierre de `pagar_estacion` se completa pero no aparecen metricas o documento de venta.
- la correccion rapida post-cobro no deja trazabilidad esperada.

## Alcance

Aplica al flujo `configuracion_de_estaciones -> estaciones -> carrito_de_compras -> carritos_compra` y a la asociacion con sensores del modulo `sensor_puertas`. `ventas_simple.html` se considera solo una redireccion de compatibilidad hacia el carrito unificado.

## Fuentes de evidencia

- `empresa_estacion_prefs` con la clave `estaciones_config`
- `carritos_compras` de la empresa afectada
- `carrito_compra_items` del carrito enlazado
- `empresa_ventas_estacion_metricas`
- `empresa_sensor_puertas_devices` y, si hace falta, `empresa_sensor_puertas_messages`
- logs del backend para `/api/empresa/estacion_prefs`, `/api/empresa/sensor_puertas` y `/api/empresa/carritos_compra`
- URL abierta en `estaciones.html`, `carrito_de_compras.html` o `ventas_simple.html` con `empresa_id`, `estacion_id`, `carrito_codigo` y `carrito_id`

## Verificaciones iniciales

1. confirmar el `empresa_id` de la sesion y el `estacion_id` implicado.
2. verificar que exista `estaciones_config` para la empresa en `empresa_estacion_prefs` con `estacion_id=0`.
3. revisar si la estacion afectada existe dentro del JSON persistido y si su nombre o id coinciden con la expectativa operativa.
4. comprobar si existe el carrito base canonico de esa estacion con `codigo=EST-empresa-estacion` o `referencia_externa=ESTACION_<id>`.
5. validar el estado real del carrito: `estado`, `estado_carrito`, `pagado_en`, `activado_en`, `metodo_pago`.
6. si el problema es visual de sensor, revisar el dispositivo asignado a la misma estacion y su `last_state` junto a `last_seen`.
7. si el problema es de metricas o documento, confirmar si `pagar_estacion`, `anular_cierre_parcial` o `recuperar_interrumpido` devolvieron `200` o dejaron error en logs.

## Causas probables

- `estaciones_config` guardada con payload legacy o inconsistente.
- carrito base faltante, renombrado manualmente o no sincronizado tras guardar configuracion.
- intento de operar una venta pagada sin reactivar con `reset_items=1`.
- referencia o metodo de pago invalido en `pagar_estacion`.
- metodo de pago bloqueado por configuracion operativa del rol.
- sensor asociado a otra estacion, sin heartbeat reciente o con `last_seen` vencido.
- documento transaccional no emitido por esquema legacy de configuracion avanzada o por tabla transaccional sin secuencia/default valida en PostgreSQL.
- metrica no registrada por fallo posterior al cierre o por identificar mal la estacion desde el carrito.

## Acciones de recuperacion

1. volver a leer y, si hace falta, regrabar `estaciones_config` usando el endpoint `PUT /api/empresa/estacion_prefs` con `estacion_id=0` y `clave=estaciones_config` para forzar `SyncEmpresaEstacionCarritos`.
2. confirmar que el carrito base exista una sola vez por estacion y que conserve identidad canonica `EST-empresa-estacion` y `ESTACION_<id>`.
3. si la estacion abre un carrito incorrecto, comparar la URL abierta en `carrito_de_compras.html` contra el `carrito_codigo` esperado y el carrito real resuelto en backend.
4. si el carrito ya fue pagado, ejecutar `activar_estacion` con `reset_items=1` antes de intentar una nueva venta.
5. si `pagar_estacion` falla, revisar primero metodo de pago, referencia minima y restricciones de configuracion operativa del rol.
6. si el sensor no se refleja, validar que el dispositivo este ligado a la misma estacion, que `last_state` sea un estado activo reconocido y que `last_seen` no este desactualizado.
7. si hay actividad fisica pero no cambia `last_seen`, revisar el heartbeat en `/api/public/sensor_puertas?action=heartbeat` y el estado del dispositivo en `empresa_sensor_puertas_devices`.
8. si no se generaron metricas o documento, verificar que el cierre haya llegado a `pagar_estacion`, que el carrito haya quedado en `venta_pagada` y que no existan errores de autorreparacion PostgreSQL o configuracion avanzada legacy en logs.
9. si la correccion post-cobro no aparece, revisar `anular_cierre_parcial`, la auditoria y la tabla `empresa_ventas_estacion_metricas` para el mismo `carrito_id`.

## Validacion posterior

- la estacion vuelve a ser visible y abre el carrito correcto de su empresa.
- el carrito base queda unico por estacion y recupera su estado base `inactivo/cerrado` cuando corresponde.
- el sensor solo se pinta verde cuando el dispositivo correcto reporta actividad reciente.
- `carrito_de_compras.html` puede iniciar nueva venta, cobrar y volver a estaciones sin perder `empresa_id`; `ventas_simple.html` debe redirigir sin perder contexto.
- el cierre de venta deja documento de venta y fila de metrica coherentes con la estacion operada.
- la correccion post-cobro queda registrada con trazabilidad y sin romper el historial operativo.

## Contrato relacionado

- `documentos/gobernanza_tecnica/contratos/contrato_estaciones_sensores_ventas_simple.md`

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`