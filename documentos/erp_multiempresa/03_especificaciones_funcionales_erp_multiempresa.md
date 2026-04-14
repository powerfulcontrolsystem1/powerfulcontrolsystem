# Especificaciones funcionales ERP multiempresa

Fecha: 2026-04-14
Version: 1.0
Estado: listo para revision

## 1. Objetivo

Definir requisitos funcionales verificables del ERP multiempresa, organizados por proceso de negocio, reglas de negocio e integraciones.

## 2. Convenciones

- Tipo de requisito funcional: `RF-XXX`.
- Tipo de regla de negocio: `RB-XXX`.
- Tipo de integracion: `INT-XXX`.
- Todo requisito aplica aislamiento por `empresa_id` salvo catalogos globales explicitos.

## 3. Procesos empresariales y requisitos funcionales

## 3.1 Gobierno y seguridad de plataforma

- RF-001: El sistema debe permitir administrar empresas, tipos y licencias desde panel super.
- RF-002: El login administrativo debe usar OAuth Google con seleccion explicita de cuenta.
- RF-003: La primera sesion administrativa debe requerir aceptacion contractual.
- RF-004: El sistema debe exponer matriz de permisos efectivos por rol/modulo/accion.
- RF-005: Las rutas empresariales criticas deben validar permisos antes de ejecutar mutaciones.
- RF-006: El sistema debe registrar auditoria de acciones criticas y denegaciones.
- RF-007: El sistema debe permitir configurar y ejecutar backups por empresa.

## 3.2 Comercial (Lead-to-Cash)

- RF-101: Gestionar clientes con CRUD, perfil comercial e historial de compras.
- RF-102: Gestionar cotizaciones y pedidos con transiciones de estado.
- RF-103: Permitir conversion cotizacion -> pedido -> documento final.
- RF-104: Operar carritos de compra con items de producto, servicio y combo.
- RF-105: Validar metodos de pago habilitados por contexto operativo.
- RF-106: Aplicar descuentos por codigo segun vigencia, uso y reglas de canal.
- RF-107: Integrar propinas y comisiones segun configuracion empresarial.
- RF-108: Soportar venta publica por empresa con slug, catalogo y ordenes.

## 3.3 Operacion de estaciones y reservas

- RF-151: Configurar estaciones por empresa con estado operativo visible.
- RF-152: Permitir multiples carritos/sesiones concurrentes por estacion.
- RF-153: Calcular tarifas por minutos y/o por dia segun reglas vigentes.
- RF-154: Gestionar reservas por estacion/habitacion con anti-overbooking.
- RF-155: Permitir cierre y recuperacion de flujo de estacion con trazabilidad.

## 3.4 Abastecimiento (Procure-to-Pay)

- RF-201: Gestionar proveedores y condiciones comerciales.
- RF-202: Crear documentos de compra y transicionar por estados controlados.
- RF-203: Soportar recepcion parcial y total de compras.
- RF-204: Contabilizar compra con evidencia documental.
- RF-205: Validar consistencia proveedor-documento para control operativo.

## 3.5 Inventario y operaciones de bodega

- RF-251: Gestionar productos, servicios, categorias y bodegas.
- RF-252: Registrar movimientos de inventario (entrada, salida, traslado, ajuste).
- RF-253: Exponer kardex filtrable por bodega, tipo y rango de fechas.
- RF-254: Detectar quiebre de stock por producto/bodega.
- RF-255: Exponer tendencia y balance de inventario por bodega.
- RF-256: Generar proyeccion de quiebre y plan de reposicion por proveedor.

## 3.6 Finanzas y contabilidad (Record-to-Report)

- RF-301: Registrar ingresos y egresos por empresa con codificacion contable.
- RF-302: Gestionar periodos contables abiertos/cerrados.
- RF-303: Gestionar cierres de caja por sucursal/turno con aprobacion.
- RF-304: Generar eventos contables desde modulos transaccionales.
- RF-305: Procesar asientos contables por lotes con control de idempotencia.
- RF-306: Exponer conciliacion por periodo entre eventos y asientos.
- RF-307: Exportar tablero financiero-operativo en formatos estandar.

## 3.7 Credito y cartera

- RF-351: Crear creditos por cliente con validacion de limites.
- RF-352: Generar plan de cuotas y actualizar estado de cartera.
- RF-353: Registrar abonos totales/parciales con trazabilidad de canal de pago.
- RF-354: Exponer alertas de mora y ranking de morosidad.
- RF-355: Gestionar workflows de reverso y refinanciacion con aprobacion.

## 3.8 Talento humano

- RF-401: Registrar asistencia diaria por empresa.
- RF-402: Gestionar liquidacion de nomina por periodo.
- RF-403: Gestionar vacaciones/licencias por flujo empresarial.

## 3.9 Colaboracion, soporte y continuidad

- RF-451: Gestionar conversaciones, mensajes y tareas por empresa.
- RF-452: Permitir adjuntos de imagen, audio y documentos de oficina.
- RF-453: Gestionar soporte remoto por dispositivo y sesion.
- RF-454: Mantener bitacora de sesiones de soporte y exportacion.
- RF-455: Permitir exportacion de reportes en PDF, XLS, CSV, JSON y TXT.

## 4. Reglas de negocio (RB)

## 4.1 Reglas transversales

- RB-001: Ninguna mutacion empresarial puede ejecutarse sin `empresa_id` valido.
- RB-002: Toda accion critica debe registrar actor, fecha, recurso y resultado.
- RB-003: Las transiciones de estado invalidas deben responder conflicto de negocio.
- RB-004: Las rutas protegidas operan con politica deny-by-default.

## 4.2 Reglas comerciales y de cobro

- RB-101: Un codigo de descuento no puede exceder sus limites de uso y vigencia.
- RB-102: El cierre de carrito debe validar metodo de pago habilitado por contexto.
- RB-103: La propina solo aplica cuando la politica empresarial lo permite.
- RB-104: Las comisiones por servicio deben respetar escala y tope configurado.

## 4.3 Reglas de estaciones

- RB-151: Una estacion puede representar mesa, habitacion o punto de atencion equivalente.
- RB-152: La operacion debe soportar concurrencia de multiples carritos por estacion.
- RB-153: Toda operacion de estacion debe quedar trazada por carrito y cliente.

## 4.4 Reglas financieras y contables

- RB-201: Todo movimiento financiero debe estar asociado a periodo contable.
- RB-202: El cierre/aprobacion de caja requiere permiso de aprobacion.
- RB-203: Los asientos deben conservar idempotencia por evento origen.
- RB-204: La exportacion de reportes debe mantener estructura y totales entre formatos.

## 4.5 Reglas de interoperabilidad

- RB-251: La salida de reportes debe ser compatible con consumo contable externo.
- RB-252: Toda exportacion debe conservar empresa, documento y periodo.
- RB-253: Integraciones fiscales deben preservar evidencia de envio y acuse.

## 5. Integraciones funcionales (INT)

## 5.1 Identidad y seguridad

- INT-001: OAuth Google para login administrativo.
- INT-002: Tokens/sesiones con revocacion y control de expiracion.

## 5.2 Comunicaciones

- INT-011: SMTP para confirmacion de correo, alertas y notificaciones.

## 5.3 Pagos

- INT-021: Wompi/Mercado Pago para creacion y consulta de transacciones.
- INT-022: Persistencia local de estado de pago para conciliacion.

## 5.4 Facturacion electronica

- INT-031: DIAN para habilitacion, validacion documental y envio de lotes.
- INT-032: Soporte de software compartido SaaS y credenciales por empresa.

## 5.5 Mapa e IA

- INT-041: OpenStreetMap para geolocalizacion operativa.
- INT-042: Proveedor IA configurable para chat empresarial con trazabilidad.

## 6. Especificacion de reportes y exportaciones

## 6.1 Reportes minimos obligatorios

- Tablero financiero-operativo.
- Cartera y morosidad.
- Auditoria empresarial.
- Backups y restauraciones.
- Ventas por periodo y por estacion.

## 6.2 Formatos de exportacion obligatorios

- PDF
- XLS
- CSV
- JSON
- TXT

## 6.3 Reglas de consistencia de exportacion

- Misma estructura de columnas clave por reporte.
- Mismos totales para el mismo filtro temporal.
- Mismo criterio de empresa/documento/periodo entre formatos.

## 7. Criterios de aceptacion funcional (resumen)

1. Cada proceso empresarial tiene requisitos y reglas verificables.
2. La seguridad por rol y alcance multiempresa se mantiene en toda mutacion.
3. Los flujos comerciales, financieros y contables quedan trazables.
4. Las integraciones externas estan explicitadas por objetivo y control.
5. Los reportes cumplen regla de multiformato consistente.

## 8. Matriz breve de trazabilidad

| Proceso | Requisitos principales | Reglas principales |
|---|---|---|
| Gobierno y seguridad | RF-001..RF-007 | RB-001..RB-004 |
| Comercial | RF-101..RF-108 | RB-101..RB-104 |
| Estaciones | RF-151..RF-155 | RB-151..RB-153 |
| Abastecimiento | RF-201..RF-205 | RB-001, RB-003 |
| Inventario | RF-251..RF-256 | RB-001, RB-003 |
| Finanzas/contabilidad | RF-301..RF-307 | RB-201..RB-204 |
| Credito/cartera | RF-351..RF-355 | RB-201, RB-251..RB-253 |
| Talento humano | RF-401..RF-403 | RB-001, RB-002 |
| Colaboracion/soporte | RF-451..RF-455 | RB-002, RB-251..RB-253 |

## 9. Consideraciones de evolucion

- Todo modulo nuevo debe mapearse a RF y RB antes de desarrollo.
- Todo cambio funcional debe reflejar permisos, aislamiento y auditoria.
- Toda ampliacion de integracion debe agregar controles de falla y evidencia.
