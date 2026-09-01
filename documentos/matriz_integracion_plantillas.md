# Matriz de integracion de plantillas

Actualizacion: 2026-08-23

## Regla vigente

El sistema mantiene un solo nucleo operativo. Las plantillas especializan la operacion, pero no pueden duplicar clientes, productos o servicios, inventario, ventas, pagos, finanzas, facturacion, permisos ni reportes.

La licencia y el tipo de empresa activan solamente la plantilla elegida y los modulos centrales necesarios. Cada consulta y mutacion empresarial conserva aislamiento por `empresa_id` y permiso efectivo.

Nucleo obligatorio:

- `clientes`
- `inventario` / productos y servicios
- `ventas` / carrito / venta directa
- `pagos`
- `finanzas`
- `facturacion`
- `reportes`
- `seguridad`

## Catalogo vigente

El catalogo operativo publica 13 plantillas canonicas: cuatro clasicas y nueve plantillas nuevas. `turnos_atencion` es una capacidad transversal y no una plantilla comercial adicional.

| Plantilla | Estado | Especialidad permitida |
|---|---|---|
| Parqueadero | `plantilla_integrada_nucleo` | Tickets QR, placas, entrada/salida y reglas tarifarias |
| Domicilios | `plantilla_integrada_nucleo` | Pedidos, domiciliarios, tracking y estados logisticos |
| Alquileres | `plantilla_integrada_nucleo` | Contratos, activos, garantias y mantenimientos |
| Construccion / AIU | `plantilla_integrada_nucleo` | Capitulos, AIU, presupuestos, retenciones y avance de obra |
| Eventos y boleteria | `plantilla_integrada_nucleo` | Eventos, boletas QR, aforo y validacion |
| Salon, barberia y spa | `plantilla_integrada_nucleo` | Agenda, profesionales, insumos y comisiones |
| Veterinaria y pet shop | `plantilla_integrada_nucleo` | Pacientes animales, vacunas, servicios y productos |
| Lavanderia y tintoreria | `plantilla_integrada_nucleo` | Recepcion, prendas, etiquetas, proceso y entrega |
| Taller mecanico | `plantilla_integrada_nucleo` | Ordenes, diagnostico, repuestos y garantias |
| Transporte de carga / TMS | `plantilla_integrada_nucleo` | Fletes, manifiestos, conductores y tracking de carga |
| Servicios tecnicos | `plantilla_integrada_nucleo` | Ordenes, agenda, tecnicos, repuestos y firmas |
| Funeraria y servicios exequiales | `plantilla_integrada_nucleo` | Planes, afiliados, salas, documentos y cierre |
| Parque recreativo | `plantilla_integrada_nucleo` | Entradas, manillas QR, atracciones, aforo e incidentes |

## Modulos retirados

Desde 2026-08-23 se retiraron Gimnasio, Taxi System, Apartamentos turisticos, Propiedad horizontal y Odontologia. Tambien se retiro Drogueria/Farmacia como modulo independiente.

Las empresas de drogueria o farmacia pueden conservar un tipo de empresa y datos guia, pero su operacion usa `inventario`, productos, compras, ventas y facturacion centrales. Lotes, fechas de vencimiento, alertas y trazabilidad de producto pertenecen a Inventario; no existe permiso, licencia, pagina, endpoint ni expediente paralelo `drogueria_farmacia`.

## Contrato tecnico

1. Toda venta termina en Ventas.
2. Todo cobro pasa por Pagos y todo movimiento queda conciliable en Finanzas.
3. Todo tercero facturable existe en Clientes.
4. Todo producto, servicio, tarifa o procedimiento vendible existe en Inventario/productos/servicios.
5. Toda plantilla visible valida licencia, rol, pagina y `empresa_id`.
6. Cada plantilla declara modulos activados, tablas, permisos, flujo comercial y reportes.
7. Una capacidad transversal no se publica como modulo comercial separado.

## Implementacion

- El backend publica la matriz desde `/api/public/plantillas_integracion/catalogo`, `/api/empresa/plantillas_integracion/catalogo` y `/super/api/plantillas_integracion/catalogo`.
- `web/js/plantillas_integracion_catalogo.js` contiene solamente las cuatro plantillas clasicas activas; las nueve nuevas se completan desde `web/js/plantillas_nuevas_catalogo.js`.
- Las pruebas de contrato exigen exactamente 13 plantillas, sin alias ni modulos retirados.
- La portada `web/index.html` y las landings
  `web/descripcion_de_los_sistemas.html` / `.ht` agregan las cuatro plantillas
  clasicas y las nueve nuevas al catalogo visible. Si una tarjeta personalizada
  ya usa la misma clave `module`, se reemplaza por la ficha canonica para no
  duplicarla.
- La configuracion persistida de pagina principal se normaliza en backend y se
  vuelve a filtrar en frontend para impedir que anuncios historicos retirados
  reaparezcan en el portal.
