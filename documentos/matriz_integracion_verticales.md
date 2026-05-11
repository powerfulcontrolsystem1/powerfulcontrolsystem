# Matriz de integracion de verticales

Actualizacion: 2026-05-11

## Conexion con preconfiguraciones

Las preconfiguraciones de tipo de empresa quedan conectadas directamente con esta matriz extendida mediante el bloque JSON `integracion_vertical`. Ese bloque declara modulo, estado de integracion, decision comercial, prioridad de produccion masiva, modulos activados, tablas tocadas, permisos requeridos, flujo de venta y reportes producidos.

Para produccion masiva se priorizan los 20 verticales nuevos con ranking 1-20. Los verticales `operador_turistico`, `colegio_academia`, `guarderia_infantil`, `inmobiliaria_comercial`, `seguridad_privada`, `club_deportivo`, `funeraria_exequial`, `parque_recreativo`, `cooperativa_fondo` y `capacitacion_empresarial` quedan incluidos como plantillas reales sobre el mismo nucleo unico. El detalle de decision y plan esta en `documentos/plan_verticales_produccion_masiva_2026-05-11.md`.

## Regla profesional

El sistema debe mantener un solo nucleo operativo. Ningun vertical puede crear un circuito paralelo para clientes, productos/servicios vendibles, ventas, pagos, facturacion, permisos o reportes si ya existe un modulo central para esa responsabilidad.

Nucleo obligatorio:

- `clientes`
- `inventario` / productos y servicios
- `ventas` / carrito / venta directa
- `pagos`
- `facturacion`
- `reportes`
- `seguridad` / usuarios / roles / permisos

Los verticales son plantillas o especializaciones. Pueden tener tablas propias solo para datos especificos del negocio: historia clinica, control de acceso, turnos, rutas, tickets, habitaciones, evidencias, agenda especializada o trazabilidad tecnica.

## Estados

- `plantilla_integrada_nucleo`: vertical visible; opera como plantilla sobre modulos base y no reemplaza ventas, clientes, productos ni pagos.
- `integrado_soporte`: visible; es una capacidad transversal que no sustituye el nucleo comercial.
- `pendiente_integracion_nucleo`: oculto del menu operativo; existe en codigo, pero requiere migrar duplicados reales al nucleo antes de mostrarse como solucion lista.
- `comercial_no_operativo`: no visible en administracion; puede existir solo como contenido comercial o backlog.
- `descartable`: candidato a fusionar o eliminar si no aporta especialidad distinta al nucleo.

## Matriz inicial

| Vertical | Estado | Visible en operacion | Usa nucleo requerido | Tablas/flujo propio permitido | Duplicado a resolver | Decision |
|---|---|---:|---|---|---|---|
| 20 nuevos verticales (`agencia_viajes`, `salon_spa`, `taller_mecanico`, etc.) | `plantilla_integrada_nucleo` | Si | clientes, productos/servicios, ventas, pagos, facturacion, reportes, permisos | `empresa_modulos_colombia_*` para seguimiento, agenda, evidencias, aprobaciones y riesgo | Ninguno critico detectado en esta fase | Mantener visible como plantilla |
| Turnos de atencion | `integrado_soporte` | Si | seguridad, reportes y operacion empresarial | Turnos, puestos, pantalla publica y seguimiento de fila | No reemplaza ventas/pagos/clientes | Mantener visible |
| Gimnasio | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Socios como especialidad, acceso, credenciales, clases, asistencias y dispositivos | Migracion historica disponible por accion `sincronizar_nucleo` | Mantener visible |
| Odontologia | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Historia clinica, odontograma, profesionales, consultorios y citas clinicas | Migracion historica disponible por accion `sincronizar_nucleo` | Mantener visible |
| Parqueadero | `plantilla_integrada_nucleo` | Si | ventas/carritos, pagos, servicios vendibles y reportes | Ticket QR, placa, entrada/salida y reglas tarifarias | Migracion historica disponible por accion `sincronizar_nucleo` | Mantener visible |
| Taxi system | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Conductores, despacho, GPS, tracking y rutas | Migracion historica disponible por accion `sincronizar_nucleo` | Mantener visible |
| Domicilios | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Tracking, domiciliarios, restaurantes aliados, menu, ofertas y estados logisticos | Migracion historica disponible por accion `sincronizar_nucleo` | Mantener visible |
| Apartamentos turisticos | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Unidades, disponibilidad, tareas, tarifas, check-in y checkout | Migracion historica disponible por accion `sincronizar_nucleo` | Mantener visible |
| Propiedad horizontal | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Unidades, asambleas, PQR, residentes, cartera y recaudos | Migracion historica disponible por accion `sincronizar_nucleo` | Mantener visible |
| Alquileres | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, pagos y reportes | Contratos, activos, garantias, mantenimientos, kilometraje y mapa GPS | Migracion historica disponible por accion `sincronizar_nucleo` | Mantener visible |
| Drogueria / farmacia | `plantilla_integrada_nucleo` | Si | productos, inventario, compras, ventas, clientes, facturacion y reportes | Expediente sanitario, lotes, INVIMA, formulas, controlados y farmacovigilancia | Validado: no crea inventario/venta paralela; usa `empresa_modulos_colombia_*` | Mantener visible |
| Construccion / AIU | `plantilla_integrada_nucleo` | Si | clientes, servicios vendibles, ventas/carritos, facturacion y reportes | Capitulos, AIU, presupuestos de obra, retenciones, anticipo, garantia y auditoria tecnica | Migracion historica disponible por accion `sincronizar_nucleo` | Mantener visible |

## Contrato tecnico obligatorio

1. Toda venta debe terminar en el modulo central de ventas.
2. Todo cobro debe pasar por pagos centrales o dejar una referencia reconciliable con pagos centrales.
3. Todo cliente o paciente facturable debe existir en clientes del nucleo.
4. Todo producto, servicio, plan, tarifa o procedimiento vendible debe existir como producto/servicio del nucleo.
5. Todo reporte vertical debe leer datos del nucleo o declarar una tabla especializada no duplicada.
6. Todo vertical visible debe validar licencia, rol, pagina y `empresa_id`.
7. Todo vertical pendiente queda protegido por ruta y permisos, pero no se muestra como solucion operativa.
8. Todo vertical visible debe declarar `template_activates`, `tables_touched`, `required_permissions`, `sale_flow` y `reports_produced` para que la matriz sea auditable por negocio.

## Oleadas de implementacion

1. Normalizacion visible: ocultar pendientes y publicar matriz.
2. Gimnasio: migrar socios a clientes, planes a servicios/productos, pagos a ventas/pagos. Estado: implementado para flujos nuevos y con accion de sincronizacion historica.
3. Odontologia: migrar pacientes facturables a clientes, tratamientos a servicios y pagos a ventas/pagos; conservar historia clinica. Estado: implementado para flujos nuevos y con accion de sincronizacion historica.
4. Parqueadero: al cerrar ticket, crear venta/pago central y vincular ticket con la transaccion. Estado: implementado para flujos nuevos y con accion de sincronizacion historica.
5. Domicilios y taxi: conectar pedido/viaje con venta/pago central. Taxi system: implementado para viajes completados y con accion de sincronizacion historica. Domicilios: implementado para pedidos entregados y con accion de sincronizacion historica.
6. Apartamentos, alquileres, propiedad horizontal y drogueria: validar duplicados, fusionar o dejar como plantilla. Apartamentos turisticos: implementado para huespedes, servicios de alojamiento y reservas cerradas. Propiedad horizontal: implementado para propietarios/residentes, servicios de cargos y recaudos cobrados. Alquileres: implementado para clientes de contrato, activos/tarifas como servicios y contratos con venta central. Drogueria/farmacia: validado como expediente sanitario sobre productos, inventario, ventas y facturacion centrales.
7. Construccion / AIU: implementado para clientes de obra, contratos/conceptos como servicios y facturas enlazadas a carritos centrales sin recalcular impuestos en el carrito.

## Implementacion actual

- `web/js/verticales_integracion_catalogo.js` define el estado visible/oculto de verticales clasicos.
- `web/js/administrar_empresa.js` refresca ese catalogo desde `/api/empresa/verticales_integracion/catalogo` cuando hay contexto de empresa y conserva el JS local como respaldo si la API no responde; luego aplica permisos/licencias.
- El shell de `web/administrar_empresa.html` muestra un resumen operativo compacto de la matriz activa para distinguir fuente API/local y conteo de verticales visibles u ocultos.
- `web/administrar_empresa/verticales_integracion.html` permite consultar la matriz completa desde el panel empresa y ejecutar sincronizaciones historicas cuando el catalogo declara `sync_path` + `sync_action_name`. La pantalla consulta `/api/empresa/permisos_contexto` antes de habilitar botones, confirma la accion con el usuario y conserva la autorizacion efectiva de cada endpoint vertical.
- La matriz completa ahora expone por vertical los modulos/plantilla que activa, tablas tocadas, permisos requeridos, flujo de venta y reportes producidos; esto convierte cada solucion en una plantilla gobernada y no en un circuito duplicado.
- `/api/public/verticales_integracion/catalogo`, `/api/empresa/verticales_integracion/catalogo` y `/super/api/verticales_integracion/catalogo` exponen la misma matriz operativa para auditoria, super y empresa.
- `web/js/nuevos_verticales_catalogo.js` marca los 20 verticales nuevos como `plantilla_integrada_nucleo`.
- `/api/empresa/verticales_nuevos/catalogo`, `/super/api/verticales_nuevos/catalogo` y `/api/public/verticales_nuevos/catalogo` exponen estado de integracion para los nuevos verticales.
- `web/super/verticales_produccion_masiva.html` usa `/super/api/verticales_nuevos/catalogo` para auditar los 20 verticales de produccion masiva, metadata extendida y exportacion CSV desde el panel super.
- Desde esa vista super, cada vertical abre sus pantallas relacionadas con filtro inicial: `Tipos de empresa`, `Preconfiguraciones` y `Licencias`.
- La senal `Listo venta` exige metadata extendida completa, preconfiguracion activa con `integracion_vertical` y licencia activa de catalogo que incluya el modulo.
- La accion `Asegurar 20` asegura tipos, preconfiguraciones y licencias recomendadas para los 20 verticales nuevos.
- Gimnasio crea/sincroniza `cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id` contra el nucleo. La accion `POST /api/empresa/gimnasio?action=sincronizar_nucleo&empresa_id=...` migra referencias historicas sin borrar las tablas especializadas del vertical y es repetible porque los pagos usan `referencia_externa` estable en carritos.
- Odontologia crea/sincroniza `cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id` contra el nucleo. La accion `POST /api/empresa/odontologia?action=sincronizar_nucleo&empresa_id=...` migra pacientes, tratamientos y pagos historicos sin borrar historia clinica, odontogramas, agenda ni presupuesto especializado y es repetible porque los pagos usan `referencia_externa` estable en carritos.
- Parqueadero crea/sincroniza `cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id` contra el nucleo. La accion `POST /api/empresa/parqueadero?action=sincronizar_nucleo&empresa_id=...` migra tickets cerrados historicos sin borrar placas, QR, tiempos, tarifas ni trazabilidad de entrada/salida.
- Taxi system crea/sincroniza `cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id` contra el nucleo. La accion `POST /api/empresa/taxi_system?action=sincronizar_nucleo&empresa_id=...` migra viajes completados historicos sin borrar conductores, GPS, despacho, ofertas ni trazabilidad de ruta.
- Domicilios crea/sincroniza `cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id` contra el nucleo. La accion `POST /api/empresa/domicilios?action=sincronizar_nucleo&empresa_id=...` migra pedidos entregados historicos sin borrar restaurantes, domiciliarios, ofertas, tracking GPS, menu ni estados logisticos.
- Apartamentos turisticos crea/sincroniza `cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id` contra el nucleo. La accion `POST /api/empresa/apartamentos_turisticos?action=sincronizar_nucleo&empresa_id=...` migra reservas en check-in/checkout historicas sin borrar unidades, tarifas, disponibilidad, codigos de acceso ni tareas operativas.
- Propiedad horizontal crea/sincroniza `cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id` contra el nucleo. La accion `POST /api/empresa/propiedad_horizontal?action=sincronizar_nucleo&empresa_id=...` migra propietarios/residentes, unidades, cargos y recaudos historicos sin borrar PQR, asambleas, coeficientes, cartera ni trazabilidad de copropiedad.
- Alquileres crea/sincroniza `cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id` contra el nucleo. La accion `POST /api/empresa/alquileres?action=sincronizar_nucleo&empresa_id=...` migra clientes de contratos, activos, tarifas y contratos historicos sin borrar garantias, kilometraje, GPS, mantenimientos ni devoluciones.
- Drogueria/farmacia queda visible como plantilla integrada porque no crea tablas paralelas de producto, inventario, venta ni pago: usa `empresa_modulos_colombia_*` para expediente sanitario y exige los modulos centrales de productos, inventario, compras, ventas, clientes y facturacion en sus licencias/preconfiguracion.
- Construccion / AIU crea/sincroniza `cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id` contra el nucleo. La accion `POST /api/empresa/aiu_construccion?action=sincronizar_nucleo&empresa_id=...` migra contratos, conceptos y facturas historicas sin borrar capitulos, calculo AIU, retenciones, anticipo, garantia ni auditoria tecnica.
