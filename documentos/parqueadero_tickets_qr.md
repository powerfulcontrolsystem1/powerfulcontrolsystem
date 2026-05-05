# Modulo Parqueadero: tickets QR y cobro automatico

El modulo de parqueadero permite operar entradas y salidas de vehiculos por empresa, con aislamiento por `empresa_id`.

## Funciones principales

- Configuracion de tarifa base, minutos base, tolerancia, fracciones, tope diario, moneda e IVA.
- Emision de ticket de ingreso por placa y tipo de vehiculo.
- Generacion de token unico y QR de salida en el recibo del cliente.
- Calculo automatico del cobro segun tiempo real de permanencia.
- Cierre de salida con metodo de pago y recibo imprimible.
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
- `POST action=cobrar_salida`: cobra y cierra la salida.
- `POST action=anular`: anula un ticket abierto.

## Permisos y licencia

El modulo usa la llave `parqueadero`, integrada a permisos finos, roles y configuracion de licencias. Los roles operativos autorizados pueden registrar entradas, calcular cobros y cerrar salidas segun las reglas del plan activo.
