# Matriz de integracion de plantillas

Actualizacion: 2026-05-12

## Conexion con preconfiguraciones

Las preconfiguraciones de tipo de empresa quedan conectadas directamente con esta matriz extendida mediante el bloque JSON `integracion_vertical`. Ese bloque declara modulo, estado de integracion, decision comercial, prioridad de produccion masiva, modulos activados, tablas tocadas, permisos requeridos, flujo de venta y reportes producidos.

Para produccion masiva se publican exactamente 30 plantillas canonicos: 10 clasicos reales y 20 nuevos con ranking 1-20. Los alias historicos no inflan el conteo: `consultorio_odontologico` queda fusionado en `odontologia`, `taxi` queda fusionado en `taxi_system`, y `turnos_atencion`/`turnos` quedan como capacidad de soporte transversal para plantillas que requieran fila, agenda o puestos. El detalle de decision y plan esta en `documentos/plan_plantillas_produccion_masiva_2026-05-11.md`.

## Regla profesional

El sistema debe mantener un solo nucleo operativo. Ningun vertical puede crear un circuito paralelo para clientes, productos/servicios vendibles, ventas, pagos, finanzas, facturacion, permisos o reportes si ya existe un modulo central para esa responsabilidad.

Desde 2026-05-11, la activacion de licencia tambien gobierna el alcance vertical de la empresa. La licencia base debe coincidir con el tipo de empresa, aplica la preconfiguracion de ese tipo de forma idempotente y deja un `vertical_scope` efectivo para que el menu y los endpoints solo permitan el vertical elegido mas los modulos del nucleo universal. Una empresa gimnasio no ve odontologia; una empresa odontologia no ve gimnasio; ambas comparten ventas, productos, clientes, finanzas, pagos, facturacion y reportes centrales.

Desde 2026-05-12, los 30 plantillas canonicos tambien deben quedar atados a ingresos y egresos del nucleo. Cada vertical visible declara `financial_core_modules`, `income_flow`, `expense_flow`, `financial_tables` y `financial_reports`. Los ingresos pasan por venta/carrito, pago central y movimiento tipo ingreso en `empresa_finanzas_movimientos`; los egresos pasan por compras/gastos/soportes centrales y movimiento tipo egreso en la misma tabla financiera.

Nucleo obligatorio:

- `clientes`
- `inventario` / productos y servicios
- `ventas` / carrito / venta directa
- `pagos`
- `finanzas` / ingresos, egresos, periodos y conciliacion
- `facturacion`
- `reportes`
- `seguridad` / usuarios / roles / permisos

Los plantillas son plantillas o especializaciones. Pueden tener tablas propias solo para datos especificos del negocio: historia clinica, control de acceso, turnos, rutas, tickets, habitaciones, evidencias, agenda especializada o trazabilidad tecnica.

## Estados

- `plantilla_integrada_nucleo`: vertical visible; opera como plantilla sobre modulos base y no reemplaza ventas, clientes, productos ni pagos.
- `integrado_soporte`: capacidad transversal que puede ser activada por una plantilla, pero no cuenta como vertical de producto.
- `pendiente_integracion_nucleo`: oculto del menu operativo; existe en codigo, pero requiere integrarse al nucleo antes de mostrarse como solucion lista.
- `comercial_no_operativo`: no visible en administracion; puede existir solo como contenido comercial o backlog.
- `descartable`: candidato a fusionar o eliminar si no aporta especialidad distinta al nucleo.

## Matriz inicial

| Vertical | Estado | Visible en operacion | Usa nucleo requerido | Tablas/flujo propio permitido | Duplicado a resolver | Decision |
|---|---|---:|---|---|---|---|
| 20 nuevas plantillas (`agencia_viajes`, `salon_spa`, `taller_mecanico`, etc.) | `plantilla_integrada_nucleo` | Si | clientes, productos/servicios, ventas, pagos, facturacion, reportes, permisos | `empresa_modulos_colombia_*` para seguimiento, agenda, evidencias, aprobaciones y riesgo | Ninguno critico detectado en esta fase | Mantener visible como plantilla |
| Gimnasio | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Socios, planes, acceso, credenciales, clases, asistencias y dispositivos | Ninguno publico; opera como plantilla fitness | Mantener visible |
| Odontologia | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Historia clinica, odontograma, profesionales, consultorios y citas clinicas | Ninguno publico; opera como plantilla clinica | Mantener visible |
| Parqueadero | `plantilla_integrada_nucleo` | Si | servicios vendibles, ventas/carritos, pagos y reportes | Ticket QR, placa, entrada/salida, tiempos y reglas tarifarias | Ninguno publico; opera como plantilla de parking | Mantener visible |
| Taxi system | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Conductores, despacho, GPS, tracking, ofertas y rutas | Ninguno publico; opera como plantilla de transporte | Mantener visible |
| Domicilios | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Tracking, domiciliarios, restaurantes aliados, menu, ofertas y estados logisticos | Ninguno publico; opera como plantilla logistica | Mantener visible |
| Apartamentos turisticos | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Unidades, disponibilidad, tarifas, check-in, checkout y tareas | Ninguno publico; opera como plantilla de alojamiento | Mantener visible |
| Propiedad horizontal | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Unidades, asambleas, PQR, residentes, cartera y recaudos | Ninguno publico; opera como plantilla de copropiedad | Mantener visible |
| Alquileres | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Contratos, activos, garantias, mantenimientos, kilometraje y GPS | Ninguno publico; opera como plantilla de renta | Mantener visible |
| Drogueria / farmacia | `plantilla_integrada_nucleo` | Si | productos, inventario, compras, ventas, clientes, facturacion y reportes | Expediente sanitario, lotes, INVIMA, formulas, controlados y farmacovigilancia | Validado: no crea inventario/venta paralela; usa `empresa_modulos_colombia_*` | Mantener visible |
| Construccion / AIU | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, facturacion y reportes | Capitulos, AIU, presupuestos, retenciones, anticipo, garantia y auditoria tecnica | Ninguno publico; opera como plantilla de construccion | Mantener visible |

Fusiones canonicas:

- `odontologia` absorbe el alias operativo `consultorio_odontologico`.
- `taxi_system` absorbe el alias operativo `taxi`.
- `turnos_atencion` y `turnos` son soporte reutilizable, no plantillas comerciales independientes.
- Plantillas con logica parecida, como `gimnasio`/`club_deportivo`, `odontologia`/`clinica_consultorios` y `taxi_system`/`transporte_carga_tms`, permanecen separados solo cuando el flujo de negocio, permisos, reportes y datos especializados justifican una plantilla propia; comparten siempre clientes, productos/servicios, ventas, pagos, facturacion y reportes.

## Contrato tecnico obligatorio

1. Toda venta debe terminar en el modulo central de ventas.
2. Todo cobro debe pasar por pagos centrales o dejar una referencia reconciliable con pagos centrales.
2.1. Todo ingreso del vertical debe quedar conciliable con `empresa_finanzas_movimientos` como ingreso del nucleo.
2.2. Todo egreso del vertical debe quedar conciliable con `empresa_finanzas_movimientos` como egreso del nucleo.
3. Todo cliente o paciente facturable debe existir en clientes del nucleo.
4. Todo producto, servicio, plan, tarifa o procedimiento vendible debe existir como producto/servicio del nucleo.
5. Todo reporte vertical debe leer datos del nucleo o declarar una tabla especializada no duplicada.
6. Todo vertical visible debe validar licencia, rol, pagina y `empresa_id`.
7. Todo vertical pendiente o soporte transversal queda protegido por ruta y permisos, pero no se cuenta como vertical comercial independiente.
8. Todo vertical visible debe declarar `template_activates`, `tables_touched`, `required_permissions`, `sale_flow` y `reports_produced` para que la matriz sea auditable por negocio.
9. Todo vertical visible debe declarar `financial_core_modules`, `income_flow`, `expense_flow`, `financial_tables` y `financial_reports`; si falta alguno, la prueba de contrato debe marcar brecha.

## Oleadas de implementacion

1. Normalizacion visible: ocultar pendientes y publicar matriz.
2. Igualdad de contrato: 10 plantillas clasicos canonicos y 20 plantillas nuevas se declaran como plantillas reales sobre nucleo comun.
3. Alcance por licencia: cada tipo de empresa activa solo su vertical y conserva el nucleo universal compartido.
4. Auditoria: cada vertical visible declara modulos activados, tablas tocadas, permisos, flujo de venta, ingresos, egresos y reportes.

## Implementacion actual

- `web/js/plantillas_integracion_catalogo.js` define el respaldo local de los 10 plantillas clasicos canonicos y fusiona alias/soportes fuera del conteo de producto.
- `web/js/administrar_empresa.js` refresca ese catalogo desde `/api/empresa/plantillas_integracion/catalogo` cuando hay contexto de empresa y conserva el JS local como respaldo si la API no responde; luego aplica permisos/licencias.
- El shell de `web/administrar_empresa.html` muestra un resumen operativo compacto de la matriz activa para distinguir fuente API/local y conteo de plantillas visibles u ocultos.
- La matriz completa ahora expone por cada uno de los 30 plantillas canonicos los modulos/plantilla que activa, tablas tocadas, permisos requeridos, flujo de venta, ingresos, egresos y reportes producidos; esto convierte cada solucion en una plantilla gobernada y no en un circuito duplicado.
- `/api/public/plantillas_integracion/catalogo`, `/api/empresa/plantillas_integracion/catalogo` y `/super/api/plantillas_integracion/catalogo` exponen la misma matriz operativa para auditoria, super y empresa.
- `web/js/plantillas_nuevas_catalogo.js` marca los 20 plantillas nuevas como `plantilla_integrada_nucleo`.
- La prueba de contrato exige que `/api/*/plantillas_integracion/catalogo` devuelva exactamente 30 items, sin alias ni soportes publicados como plantillas, y con ingresos/egresos financieros del nucleo declarados.
- `/api/empresa/plantillas_nuevas/catalogo`, `/super/api/plantillas_nuevas/catalogo` y `/api/public/plantillas_nuevas/catalogo` exponen estado de integracion para los nuevas plantillas.
- `web/super/plantillas_produccion_masiva.html` usa `/super/api/plantillas_nuevas/catalogo` para auditar los 20 plantillas de produccion masiva, metadata extendida y exportacion CSV desde el panel super.
- Desde esa vista super, cada vertical abre sus pantallas relacionadas con filtro inicial: `Tipos de empresa`, `Preconfiguraciones` y `Licencias`.
- La senal `Listo venta` exige metadata extendida completa, preconfiguracion activa con `integracion_vertical` y licencia activa de catalogo que incluya el modulo.
- La accion `Asegurar 20` asegura tipos, preconfiguraciones y licencias recomendadas para los 20 plantillas nuevas.
- Drogueria/farmacia queda visible como plantilla integrada porque no crea tablas paralelas de producto, inventario, venta ni pago: usa `empresa_modulos_colombia_*` para expediente sanitario y exige los modulos centrales de productos, inventario, compras, ventas, clientes y facturacion en sus licencias/preconfiguracion.
