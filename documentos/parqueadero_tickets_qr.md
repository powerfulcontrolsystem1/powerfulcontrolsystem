# Modulo Parqueadero: tickets QR y cobro automatico

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- La API pública de QR es de consulta. El cierre/cobro pertenece al endpoint empresarial y sus permisos.
- El cálculo de tarifa y el registro local de pago no acreditan confirmación de pasarela ni apertura física de una barrera. Probar reintento de salida y consistencia del enlace al carrito.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

El modulo de parqueadero permite operar entradas y salidas de vehiculos por empresa, con aislamiento por `empresa_id`.

## Funciones principales

- Configuracion de tarifa base, minutos base, tolerancia, fracciones, tope diario, moneda e IVA.
- Emision de ticket de ingreso por placa y tipo de vehiculo.
- Generacion de token unico y QR de salida en el recibo del cliente.
- Calculo automatico del cobro segun tiempo real de permanencia.
- Cierre de salida con metodo de pago, recibo imprimible y venta/pago central en `carritos_compras`.
- Validacion de salida por token QR desde endpoint publico de solo consulta.
- Control de tickets abiertos, salidas del dia, anulaciones e ingresos diarios.

## Rutas

- Administracion: `/administrar_empresa/parqueadero.html`
- API empresa: `/api/empresa/parqueadero`
- Validacion publica QR: `/api/public/parqueadero?empresa_id={id}&action=validar_salida&token={token}`

## Acciones API empresa

- `GET action=dashboard`: resumen operativo, configuracion, tickets abiertos y salidas recientes.
- `GET action=config`: configuracion de tarifas.
- `GET action=tickets`: lista de tickets, con filtro opcional `estado`.
- `GET action=validar_salida&token=...`: consulta y calcula el cobro del QR.
- `POST action=config`: guarda tarifas.
- `POST action=entrada`: emite ticket de ingreso.
- `POST action=calcular`: calcula el valor de salida sin cerrar.
- `POST action=cobrar_salida`: cobra, cierra la salida, genera carrito central, item de servicio y pago reconciliable.
- `POST action=anular`: anula un ticket abierto.

## Integracion con nucleo

Parqueadero no debe duplicar ventas ni pagos. La tabla `empresa_parqueadero_tickets` conserva la especialidad operativa: placa, QR, entrada, salida, minutos, tarifa y anulacion. Cuando el ticket se cobra, se enlaza con:

- `cliente_id`: opcional, solo si el ticket trae cliente o documento.
- `servicio_id`: servicio vendible central por tipo de vehiculo.
- `carrito_id` y `carrito_item_id`: venta central y item de servicio creados al cobrar la salida.

## Permisos y licencia

El modulo usa la llave `parqueadero`, integrada a permisos finos, roles y configuracion de licencias. Los roles operativos autorizados pueden registrar entradas, calcular cobros y cerrar salidas segun las reglas del plan activo.

## Fuentes y aceptación de la revisión

[parqueadero.go](../backend/handlers/parqueadero.go), [parqueadero.go](../backend/db/parqueadero.go), [parqueadero.html](../web/administrar_empresa/parqueadero.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go).

Requisitos aplicables: PCS-REQ-001 a PCS-REQ-005, PCS-REQ-009, PCS-REQ-013, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
