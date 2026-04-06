# Modulos del proyecto

Fecha de corte: 2026-04-06

Total de modulos funcionales identificados: 34

## Resumen por categoria

- Nucleo y administracion global: 2
- Operacion por empresa (negocio): 31
- Utilidad transversal: 1

## Inventario de modulos y descripcion

| # | Modulo | Categoria | Descripcion breve |
|---|---|---|---|
| 1 | Autenticacion y sesiones | Nucleo | Login, sesion, proteccion de rutas y control de acceso. |
| 2 | Administracion global (super) | Nucleo | Gestion de empresas, tipos, licencias, administradores y configuraciones globales. |
| 3 | Usuarios de empresa | Operacion | Alta, confirmacion de correo, estado, primer ingreso y login empresarial. |
| 4 | Asistencia de empleados | Operacion | Registro de entrada/salida, horas trabajadas y novedades por empresa. |
| 5 | Nomina de sueldos | Operacion | Configuracion legal, empleados, festivos y liquidaciones integradas con asistencia. |
| 6 | Registro de vehiculos | Operacion | Control de ingreso/salida vehicular con conductor, propietario y motivo. |
| 7 | Reservas por estacion/habitacion | Operacion | Disponibilidad por rango, creacion de reservas y gestion de estados de pago. |
| 8 | Tarifas por minutos | Operacion | Cobro base + bloque extra por tiempo consumido y dia de semana. |
| 9 | Tarifas por dia | Operacion | Cobro diario por servicio con ventana check-in/check-out configurable. |
| 10 | Clientes | Operacion | CRUD de clientes y soporte comercial por empresa. |
| 11 | Inventario | Operacion | Productos, bodegas, proveedores, kardex, KPI y reposicion operativa. |
| 12 | Combos de productos | Operacion | Combos con receta de ingredientes y precio unico de venta. |
| 13 | Codigos de descuento | Operacion | Codigos promocionales por empresa con vigencia, tipo y limite de uso. |
| 14 | Propinas | Operacion | Configuracion de porcentaje y distribucion, con movimientos y reporte. |
| 15 | Comisiones por servicio | Operacion | Configuracion de comision y acumulados por usuario/lavador. |
| 16 | Compras | Operacion | Documentos de compra y transiciones de ciclo (emitir, recibir, contabilizar). |
| 17 | Facturacion electronica | Operacion | Configuracion fiscal, emision/anulacion y trazabilidad legal documental. |
| 18 | Facturacion electronica DIAN | Operacion | Checklist, validacion y utilidades DIAN (CUFE/XML demo) por empresa. |
| 19 | Gestion comercial extendida | Operacion | Cotizaciones, pedidos y devoluciones con maquina de estados documental. |
| 20 | Contabilidad operativa extendida | Operacion | Plan de cuentas y cartera (cuentas por cobrar/pagar). |
| 21 | Inventario extendido | Operacion | Lotes, series y devoluciones a proveedor por empresa. |
| 22 | RRHH extendido | Operacion | Vacaciones y licencias empresariales. |
| 23 | CRM/Produccion/Logistica | Operacion | Base CRUD para leads, ordenes, rutas y envios. |
| 24 | Documental e integraciones | Operacion | Documentos/firma e integraciones API/bancos con acciones ejecutables. |
| 25 | Panel ERP extendido | Operacion | Hub y submodulos para operar dominios ERP desde frontend empresarial. |
| 26 | Carritos de compra e items | Operacion | Flujo de venta, mezcla de items y cierre de pago validado por backend. |
| 27 | Ventas simples por estacion | Operacion | Carrito rapido tipo supermercado por estacion. |
| 28 | Finanzas y contabilidad | Operacion | Movimientos, periodos, cierres de caja, eventos y asientos. |
| 29 | Auditoria empresarial | Operacion | Traza de acciones criticas por modulo, usuario y resultado. |
| 30 | Seguridad y permisos | Operacion | Contexto efectivo por rol/modulo y control de visibilidad de menu. |
| 31 | Reportes | Operacion | Tablero KPI y exportaciones PDF/XLS/CSV/JSON/TXT por empresa. |
| 32 | Graficos y estadisticas | Operacion | Visualizaciones de ventas, finanzas y operacion por rango. |
| 33 | Configuracion operativa de cobro | Operacion | Reglas de metodos de pago, propinas y comisiones por rol. |
| 34 | Calculadora por empresa | Utilidad | Calculadora operativa con memoria/historial aislados por empresa. |

## Matriz base modulo -> reportes recomendados

Esta matriz sirve para continuar el punto de "generar reportes en reportes" a partir de cada modulo.

| Modulo | Reporte recomendado | Estado en Reportes |
|---|---|---|
| Carritos de compra e items | Ventas por franja, ticket promedio, metodos de pago | Activo (tablero KPI) |
| Inventario | Rotacion, quiebres, valorizacion por bodega | Activo (`operativo_inventario_bodega`) |
| Compras | Costo por proveedor, recepcion vs orden | Activo (`operativo_compras_movimientos`) |
| Finanzas y contabilidad | Flujo de caja diario, resultado, balance | Activo |
| Nomina de sueldos | Liquidaciones por periodo y empleado | Activo |
| Propinas | Acumulado por usuario y periodo | Activo (`operativo_propinas_acumulado`) |
| Comisiones por servicio | Acumulado por lavador y periodo | Activo (`operativo_comisiones_lavador`) |
| Reservas por estacion/habitacion | Ocupacion y cumplimiento de reserva | Activo (`operativo_reservas_ocupacion`) |
| Tarifas por minutos/dia | Ingreso por modelo de tarifa | Activo (`operativo_tarifas_ingresos`) |
| Facturacion electronica | Documentos emitidos/anulados y trazabilidad | Activo (`operativo_facturacion_trazabilidad`) |
| CRM/Produccion/Logistica | Conversion comercial y cumplimiento logistico | Activo (`operativo_cadena_cumplimiento`) |
| Auditoria empresarial | Acciones criticas por modulo/usuario | Activo (`operativo_auditoria_acciones`) |

## Criterio de conteo

- Se cuenta cada modulo funcional operable por separado en backend/frontend.
- Subcapacidades internas no se cuentan como modulo independiente si pertenecen al mismo flujo principal.
- El modulo "Calculadora por empresa" se considera utilidad transversal por su uso en cualquier empresa sin mezclar historial entre empresas.
