# Flujos operativos

Guia corta de los procesos que mas se prueban y modifican. Cada flujo debe
mantener aislamiento por `empresa_id`, permisos por rol y trazabilidad cuando
afecte dinero, documentos, licencias o seguridad.

## Roles personalizados por empresa

1. Abrir `Administrar empresa > Administrar usuarios`.
2. En `Roles personalizados de esta empresa`, escribir el nombre del rol,
   descripcion y elegir un `Rol base` global.
3. Guardar. PCS crea el rol en `roles_de_usuario` con `empresa_id` de la
   empresa activa, `origen=empresa` y `rol_base_id` del rol global elegido.
4. El rol aparece inmediatamente en el selector de usuarios con la etiqueta
   `(personalizado)` y se puede asignar desde `Editar rol de usuario` o desde
   el formulario de creacion/edicion.
5. Para desactivar un rol propio, usar `Desactivar` en la misma tarjeta. El rol
   no desaparece del historial, pero deja de estar disponible para nuevas
   asignaciones activas.
6. La autorizacion efectiva usa el rol base global. Esto evita que un nombre
   libre sin matriz conocida abra permisos por accidente y mantiene el alcance
   por `empresa_id`.
7. Un rol personalizado de otra empresa no puede asignarse ni resolverse en la
   empresa actual; el backend lo rechaza aunque se manipule la URL o el payload.

## Configuracion dedicada del rol cajero

1. Abrir `Administrar empresa > Configuracion > Ventas y cobro > Rol cajero`.
2. Revisar el resumen: rol base, usuarios cajeros, pago mixto y control por
   estaciones.
3. En `Perfil del rol`, crear o actualizar el nombre/descripcion visible del
   cajero. PCS conserva el rol global `cajero` y guarda un rol personalizado
   de la empresa cuando se requiere modificar el perfil.
4. En `Cobro y caja`, activar o desactivar medios de pago, propinas,
   comisiones, ingresos manuales y egresos manuales. El backend guarda la regla
   en `empresa_configuracion_operativa_roles` para `rol=cajero`.
5. En `Carrito POS`, `Botones visibles` y `Medios de pago POS`, ajustar lo que
   ve el cajero en venta directa y estaciones. La pagina reutiliza
   `estaciones_config.carrito_ui_global`, por lo que no duplica la configuracion
   completa del carrito.
6. En `Estaciones y caja fisica`, activar el control por estaciones cuando cada
   cajero debe operar solo estaciones asignadas. Las asignaciones finas por
   usuario siguen en `Administrar usuarios`.
7. Guardar por seccion o usar `Guardar configuracion del cajero` para guardar
   operativa, carrito/estaciones y perfil. Todas las llamadas llevan
   `empresa_id`; no se debe editar el rol global ni mezclar empresas.

## Modo POS tactil por empresa

1. Abrir `Administrar empresa > Configuracion > Configuracion carrito`.
2. Activar `Modo POS tactil para carrito, estaciones y corte de caja`.
3. El valor se guarda por empresa en
   `empresa_estacion_prefs.estaciones_config.carrito_ui_global.modo_pantalla_tactil`
   usando `/api/empresa/estacion_prefs`.
4. Al abrir `Venta directa` o un carrito por estacion, el carrito aplica
   `pos-touch-mode`/`cart-touch-mode`: botones, campos, descuentos, taxi,
   pagos mixtos, cliente y acciones quedan con objetivos tactiles grandes.
5. El catalogo por botones recibe `touch=1` desde el carrito y aplica la misma
   ergonomia para agregar productos con el dedo.
6. `Estaciones` lee la misma configuracion global y agranda tarjetas, especiales
   como Caja/Notas/YouTube y acciones operativas sin cambiar el flujo normal.
7. `Corte de caja` lee la preferencia al cargar y agranda los botones de
   generar, reporte, cierre e impresion; si la preferencia no carga, queda en
   modo normal.
8. La opcion no crea tablas nuevas, no cambia permisos ni mezcla empresas; cada
   pantalla sigue filtrando por su `empresa_id` y sus endpoints existentes.

## Atajos POS por empresa

1. Abrir `Administrar empresa > Configuracion > Configuracion carrito`.
2. En `Atajos de teclado POS`, activar o desactivar la captura de teclado y
   asignar cada accion a F1-F12, ESC, ENTER, CTRL+B, CTRL+D, CTRL+P o ALT+F4.
3. PCS guarda el mapa en
   `empresa_estacion_prefs.estaciones_config.carrito_ui_global.atajos_pos`
   usando `/api/empresa/estacion_prefs`, siempre filtrado por `empresa_id`.
4. En `Venta directa` o carritos por estacion, el listener del carrito ejecuta
   acciones existentes: ayuda, buscar producto/cliente, descuento, cantidad,
   precio, suspender/recuperar, inventario, cobrar, cajon, imprimir, cancelar,
   agregar producto y salir.
5. ENTER no interfiere con formularios generales; solo agrega producto cuando el
   foco esta en el lector/cantidad del escaner o no se esta escribiendo en otro
   campo. Las acciones finales siguen pasando por permisos y validaciones del
   backend.

## Creditos y cartera por empresa

1. Abrir `Administrar empresa > Finanzas y cumplimiento > Centro financiero y
   contable`.
2. Usar la tarjeta `Creditos y cartera` o el enlace lateral del centro
   financiero.
3. En `Riesgo y limites`, abrir o ajustar el cupo del cliente; el limite queda
   por `empresa_id` y cliente.
4. En `Nuevo credito`, crear el credito con cliente, monto, plazo y condiciones
   aprobadas.
5. En `Abonos y operaciones`, registrar abonos, reversos o refinanciaciones
   permitidas.
6. En `Cartera`, `Morosidad` y `Estado de cuenta`, consultar saldos, cuotas,
   vencimientos y movimientos.
7. El carrito usa el mismo modulo cuando el medio de pago es `Credito cliente`:
   valida que el cliente tenga cupo activo y crea la cartera ligada a la venta,
   sin sumar ese valor como efectivo de caja.

## Menu visible por empresa

1. Abrir `Administrar empresa > Configuracion > Menu visible`.
2. Marcar los modulos que deben mostrarse y desmarcar los que se quieren ocultar
   del menu para todos los usuarios de esa empresa.
3. PCS guarda la preferencia en
   `empresa_estacion_prefs.menu_visual_config` con `estacion_id=0` y lista
   `hidden_links`, siempre por `empresa_id`.
4. Al cargar `Administrar empresa`, primero se aplican permisos, licencia y rol;
   luego se aplica este filtro visual. Por eso un modulo oculto no aparece en el
   menu, pero sus endpoints siguen protegidos por backend.
5. `Panel`, `Configuracion`, `Menu visible` y `Volver a empresas` no se pueden
   ocultar para evitar que el administrador pierda la ruta de recuperacion.

## Nomina y nomina electronica DIAN

1. Abrir `Administrar empresa > Finanzas y cumplimiento > Nomina`.
2. Usar el submenu interno para configurar parametros legales, crear empleados
   de nomina, registrar festivos y revisar liquidaciones.
3. En `Liquidaciones`, elegir periodo, calcular nomina y conciliar asistencia
   antes de continuar.
4. En `Pagos y PILA`, validar control contable, consultar provisiones, generar
   pagos y generar PILA cuando aplique.
5. En `Nomina electronica DIAN`, usar `Preparar lote DIAN`; PCS arma el resumen
   por empleado y muestra requisitos pendientes.
6. Si no hay pendientes, usar `Enviar nomina electronica a DIAN`; la pantalla
   vuelve a preparar el lote y llama el flujo fiscal existente de facturacion
   electronica con `tipo_documento=nomina_electronica`.
7. El envio conserva `empresa_id`, permisos existentes y configuracion DIAN de
   la empresa; si DIAN/proveedor rechaza o no hay conectividad, la respuesta se
   muestra en el panel de nomina y queda disponible para reintentos del modulo
   fiscal.
8. El tutorial vive en el mismo submenu y explica preparacion, envio, acuse,
   ajustes y archivo documental.

## Snapshot completo VPS desde super administrador

1. Abrir `Super Administrador > Plataforma > Docker VPS`.
2. En `Snapshot completo VPS`, activar snapshots y definir retencion local,
   intervalo automatico y si se incluyen imagenes Docker.
3. Para nube, configurar previamente `rclone` en la VPS y guardar solo la ruta
   remota en PCS, por ejemplo `gdrive:PCS/backups` o `mega:PCS/backups`.
4. Usar `Crear y descargar snapshot` para generar un `.tar.gz` restaurable en
   `backup/vps_snapshots` y descargarlo desde el historial.
5. Usar `Crear y subir a nube` cuando el remoto `rclone` ya este probado; PCS no
   guarda tokens OAuth ni claves de Google Drive, Mega, OneDrive o S3.
6. Si `Ejecutar automatico` queda activo, el worker horario verifica el ultimo
   respaldo y crea uno nuevo cuando se cumpla el intervalo configurado.
7. La limpieza local/remota solo elimina archivos con patron
   `pcs-vps-snapshot-*.tar.gz` mayores a los dias de retencion configurados.
8. El paquete incluye proyecto portable, `pg_dumpall` cuando este disponible,
   volumenes Docker PCS, `MANIFEST.json` y `RESTAURAR_VPS.md`; no incluye
   `.env.platform` ni secretos por defecto.

## Transferir cuenta entre mesas, habitaciones o estaciones

1. Abrir `Administrar empresa > Configuracion > Configuracion carrito`.
2. Activar `Visualizar boton Transferir cuenta`; por defecto queda apagado para
   no exponer el flujo a empresas que no lo usan.
3. Abrir una cuenta activa desde `Estaciones`, agregar productos/servicios,
   abonos o tarifa segun aplique.
4. En el carrito, usar `Transferir cuenta`, elegir el destino y registrar un
   motivo operativo.
5. PCS valida en backend que el origen este abierto/no pagado, que el destino
   sea diferente y este disponible, y que ambos pertenezcan al mismo
   `empresa_id`.
6. En motel/hotel, si el origen tiene tarifa por minutos o diaria, el destino
   debe tener una tarifa activa equivalente; si no la tiene, la transferencia se
   rechaza para evitar cobros incompatibles.
7. Al aprobarse, se mueven items, abonos, cliente, descuentos, caja y estado de
   sesion al destino; el origen queda cerrado/en cero y el usuario es llevado al
   carrito destino.
8. No se duplica inventario, no se crea una venta nueva y los roles restringidos
   de tablero de estaciones siguen sin poder ejecutar la transferencia.

## Impresoras por empresa y agente local

1. Abrir `Administrar empresa > Configuracion > Configuracion de impresora`.
2. Registrar impresoras activas por empresa con codigo, nombre, tipo de conexion,
   direccion o cola, area operativa, formato POS/carta y predeterminada.
3. Asignar impresora por funcionalidad (`ticket_cobro`, `factura_caja`,
   `orden_servicio`, `cocina`, `barra`, `reporte_caja`) o por producto/categoria.
4. En `Impresora por computador`, PCS detecta el equipo con
   `localStorage.pcs_dispositivo_id` y permite asociar ese computador/caja a una
   impresora por funcionalidad o como regla general del equipo.
5. Cuando una venta, factura, comanda o cierre necesite imprimir, el backend
   resuelve por producto/receta/categoria, luego por computador detectado,
   luego por funcionalidad y finalmente por impresora predeterminada.
6. `/api/empresa/impresoras/resolver` acepta `dispositivo_id` o `agente_id`;
   al encolar trabajos, `cola_trabajo` usa el `agente_id` del computador para
   apuntar automaticamente a la impresora asociada cuando no se envio una
   impresora directa.
7. El agente local de cada caja consulta
   `/api/empresa/impresoras/agente?action=tomar` con `empresa_id`, `agente_id` y
   `estacion_id`; solo recibe trabajos pendientes de su empresa, agente o estacion.
8. Tras imprimir en la impresora fisica del equipo, el agente marca el trabajo
   como `impreso` o `error` en `/api/empresa/impresoras/agente?action=estado`.
9. El administrador puede ver la cola reciente y reintentar trabajos fallidos
   desde la pagina de configuracion.
10. La ruta de agente usa permisos de ventas y no permite editar impresoras,
   reglas ni predeterminadas; esas acciones siguen bajo permisos de seguridad.

## Factura electronica desde una venta existente

1. Abrir `Administrar empresa > Ventas` y ubicar un comprobante de pago emitido.
2. La misma bandeja lista ventas internas y facturas electronicas, mostrando
   `Solo venta`, `Venta con factura electronica` o `Factura electronica` para
   evitar consultar paginas separadas.
3. Usar `Hacer factura electronica` sobre una fila `Solo venta`.
4. Seleccionar un cliente existente de la empresa o crear cliente rapido con
   tipo/documento, nombre y correo si no existe.
5. PCS asocia el cliente al comprobante origen por `empresa_id` antes de generar
   la factura electronica.
6. La venta origen conserva valor, pagos, inventario y total; solo se crea el
   documento fiscal derivado.
7. La fecha y hora fiscal de la factura electronica se generan al preparar el
   documento legal, por lo que pueden diferir de la fecha/hora de la venta.
8. Si se activa el check de correo, la factura se envia al correo del cliente
   cuando no haya sido enviada ya por la configuracion automatica.
9. Los campos de la representacion impresa se configuran por empresa en
   `Configuracion > Impresoras y caja > Documento de venta e impresion`.
   `Total en letras` es opcional. Si el documento es factura electronica, nota,
   soporte, nomina o equivalente electronico, PCS conserva impresos los campos
   legales obligatorios aunque el check este apagado; esos checks solo pueden
   ocultar campos en ventas/comprobantes sin factura electronica.

## Modulo NIIF

1. Abrir `Administrar empresa > Finanzas y cumplimiento > NIIF` o el acceso
   `NIIF` dentro del Centro financiero y Suite contador.
2. La pagina conserva `empresa_id` y consulta
   `/api/empresa/contabilidad_colombia?action=dashboard&empresa_id={id}` cuando
   el usuario tiene permisos contables/financieros suficientes.
3. El modulo no guarda datos en servidor: las marcas de diagnostico, politicas
   y cierre se conservan localmente por navegador y empresa para no crear
   mutaciones sin endpoint dedicado.
4. El tablero muestra base NIIF, periodo, diferencia debito/credito y avance de
   adopcion calculado por estandares, politicas y checklist.
5. Medicion permite calcular deterioro, depreciacion, valor razonable neto y
   conciliacion contable-fiscal de apoyo gerencial.
6. Notas genera un borrador exportable JSON/TXT; no reemplaza estados
   financieros oficiales ni revision del contador.

## Bolsa empresarial

1. Abrir `Administrar empresa > Analisis y control > Bolsa`.
2. La pagina conserva `empresa_id` del panel y consulta
   `/api/empresa/bolsa?empresa_id={id}` con zona horaria e idioma del navegador.
3. El backend valida sesion, empresa y permiso `bolsa:R` mediante
   `WithEmpresaBolsaPermissions`.
4. El pais se detecta desde la configuracion empresarial/facturacion; si no hay
   dato usable, se usa Colombia como fallback operativo.
5. El handler consulta indicadores internacionales y locales desde el servidor,
   calcula precio, variacion y porcentaje, y cachea por 60 segundos.
6. La pantalla muestra resumen, tablas y errores por indicador; no guarda datos,
   no emite ordenes y no constituye recomendacion de inversion.

## Suite contador

1. Abrir `Administrar empresa > Finanzas y cumplimiento > Suite contador` o
   desde el Centro financiero y contable. En el Centro financiero, `Suite
   contador` es la unica entrada contable tipo hub; Portal contador y la suite
   contable Colombia se abren desde dentro de esta pagina para evitar duplicados.
2. La pagina conserva `empresa_id` y consulta
   `/api/empresa/permisos_contexto?empresa_id={id}` para pintar cada modulo como
   disponible o bloqueado.
3. El hub no crea endpoints, tablas ni documentos; solo coordina modulos reales:
   Portal contador, PUC/NIIF, suite contable avanzada, impuestos, DIAN,
   declaraciones, certificados, cierres, activos, nomina, bancos, reportes,
   y, solo si la empresa lo habilita, compras IA y Renta IA.
4. Al abrir un modulo, el enlace reenvia `empresa_id`; la operacion final queda
   protegida por el wrapper y permiso propio de ese modulo.
5. El rol `contador` puede ver la suite y accesos contables clave, pero no recibe
   permisos de escritura, aprobacion, pagos, inventario ni emision DIAN por este
   cambio.

## Motor contable automatico

1. Los modulos operativos emiten eventos en `empresa_eventos_contables` con
   `empresa_id`, modulo, evento, documento, periodo, monto y `payload_json`.
2. El procesador de asientos toma pendientes por empresa desde
   `/api/empresa/finanzas/asientos_contables?action=procesar_asientos` o desde
   el worker automatico.
3. Cada evento se transforma en `empresa_asientos_contables` con debitos y
   creditos en `lineas_json`, totales y hash de idempotencia.
4. Si el evento no genera minimo dos lineas reales, o si la diferencia
   debito/credito no queda en cero, PCS no guarda el asiento y registra el error
   saneado en el evento para revision contable.
5. Cobertura actual del motor: ventas/facturas/notas, recibos de caja, compras,
   documento soporte, pagos/egresos, abonos, CxC/CxP, anticipos, inventario,
   costo de venta, nomina, nomina electronica, activos fijos, depreciacion y
   deterioro.
6. Las cuentas salen de la configuracion financiera de la empresa o de cuentas
   PUC base de respaldo; el payload puede enviar cuentas especificas por modulo
   cuando el origen tenga parametrizacion mas fina.
7. Pruebas: procesar un evento de factura con IVA/retencion, confirmar que
   debito y credito cuadran; procesar evento sin base contable y confirmar que
   queda fallido sin asiento.

## Renta IA en finanzas

1. Habilitar primero la pagina `linkRentaIA` en permisos finos de la empresa;
   por defecto no se muestra en ninguna empresa.
2. Abrir `Administrar empresa > Finanzas y cumplimiento > Renta IA` dentro del
   centro financiero.
3. La pagina conserva `empresa_id`, define periodo y permite ajustar tarifa,
   sobretasa, ingresos no constitutivos, rentas exentas, deducciones,
   descuentos, retenciones y anticipo.
4. El frontend llama `/api/empresa/finanzas/renta_ia?action=renta_ia` con
   `credentials: same-origin`.
5. El backend valida sesion, empresa, licencia y permiso `finanzas:R` mediante
   `WithEmpresaFinanzasPermissions`.
6. El calculo consolida datos reales filtrados por `empresa_id`: ventas
   cerradas, ingresos/egresos financieros, compras de inventario y nomina
   liquidada.
7. Para reducir doble conteo, si hay POS e ingresos financieros simultaneos se
   usa el mayor y se muestra alerta; lo mismo para egresos financieros frente a
   compras + nomina.
8. Si el usuario pide `Analizar con IA`, GPT-5.4 mini recibe solo el JSON del
   calculo, registra uso IA diario por empresa y devuelve diagnostico
   orientativo.
9. El resultado es estimacion gerencial; no reemplaza formulario oficial,
   declaracion tributaria ni revision del contador.

## Centro IA empresarial

0. En contexto empresarial, el icono circular flotante del asistente IA se
   mantiene activo para todas las empresas; no depende de permisos del Centro IA
   ni de `localStorage` antiguo. Robot/secretaria no se muestran.
1. Confirmar que el rol/licencia tenga permiso `reportes:R`; el Centro IA ya no
   se oculta por defecto, pero no concede permisos de escritura ni aprobacion.
2. Abrir `Administrar empresa > Canales digitales y colaboracion > Centro IA
   empresarial` o el acceso rapido `IA empresarial` del Centro financiero.
3. La pagina conserva `empresa_id`, periodo desde/hasta y consulta
   `/api/empresa/ia_empresarial?empresa_id={id}` con `credentials:
   same-origin`.
4. El backend valida sesion, empresa, licencia y permiso `reportes:R` mediante
   `WithEmpresaReportesPermissions`.
5. El snapshot se calcula solo con datos reales filtrados por `empresa_id`:
   ventas cerradas, ingresos, egresos, clientes, productos, servicios, bajo
   stock y diferencia ventas versus ingresos financieros.
6. Las funciones IA disponibles son diagnostico ERP, borrador de
   factura/cotizacion, cobranza/pagos, inventario/productos, conciliacion
   bancaria, compras/gastos y cumplimiento DIAN.
7. La IA no emite documentos, no registra pagos, no crea clientes, no cambia
   inventario y no activa DIAN; devuelve borradores revisables, riesgos, datos
   faltantes y siguiente accion sugerida.
8. Cada ejecucion IA usa el modelo mini configurado, registra consumo diario por empresa y
   muestra error saneado si la IA global, la IA empresarial o la credencial del
   proveedor no estan disponibles.
9. Los botones que ejecutan funciones IA deben mostrar el icono GPT y badge
   `IA` mediante `web/js/ai_button_icons.js`.

## Chat IA flotante operativo

1. El icono circular flotante esta activo por empresa y abre el chat normal con
   voz conservada; no usa robot ni secretaria.
2. En VPS/Docker, `OPENAI_API_KEY` debe existir en `deploy/.env.platform` remoto
   para que Compose la entregue al backend; `rs/sync_to_vps` la sincroniza sin
   imprimir el secreto cuando existe en los archivos locales ignorados.
3. El boton `Stop` del toolbar corta audio, aborta la respuesta activa cuando el
   navegador soporta `AbortController` y evita que respuestas tardias se pinten
   despues de cancelar.
4. Toda accion operativa generada por IA debe salir como `PCS_ACTION`, mostrarse
   al usuario y ejecutarse solo despues de confirmar.
5. El frontend solo permite `OPEN`, `POST` y `PUT` sobre endpoints cerrados; no
   permite `DELETE` ni rutas genericas.
6. Para `cajero`, las acciones permitidas son pedir/agregar productos a una
   estacion, mesa, habitacion o venta directa mediante
   `/api/empresa/ia_pedidos_estacion/ejecutar`, y activar/desactivar la emisora
   mediante `/api/empresa/ia_radio/activar`.
7. Productos manuales usan `/api/empresa/productos` y dependen de permisos de
   inventario; nomina usa `/api/empresa/nomina` y depende de permisos de nomina;
   tarifas de motel/hotel/tiempo usan los endpoints de tarifas bajo permisos de
   ventas.
8. El backend vuelve a validar sesion, `empresa_id`, licencia, pagina y permisos
   efectivos en cada ruta; la IA no concede permisos ni ejecuta SQL libre.
9. Para `super_administrador`, `administrador_total` y `admin_empresa`, el backend
   puede entregar contexto real de lectura de la base de datos de la empresa
   activa y respuestas directas como conteo de usuarios. Esa lectura siempre usa
   consultas controladas, omite secretos y filtra por `empresa_id`; cualquier
   cambio de datos debe ejecutarse por una funcion PCS existente con permisos y
   confirmacion, no por SQL libre del modelo.

## Auditoria integral de modulos

1. Abrir `Administrar empresa > Reportes > Centro de reportes`.
2. Seleccionar el area `Operacion y auditoria` y abrir el reporte
   `Auditoria de modulos`.
3. El frontend consulta `/api/empresa/reportes?dataset=operativo_modulos_resumen`
   conservando `empresa_id`.
4. El backend valida sesion, licencia y permiso `reportes:R` mediante
   `WithEmpresaReportesPermissions`.
5. El dataset recorre el inventario de modulos nuevos y existentes, verifica si
   la tabla existe, si tiene `empresa_id`, totales, activos, anulados, rango de
   fechas y ultimo registro.
6. Los hubs o calculos sin tabla propia, como Suite contador, Renta IA y Bolsa,
   se reportan como `tabla=sin_tabla` para dejar evidencia sin crear datos
   ficticios.
7. En `Administrar empresa > Auditoria`, el filtro `Modulo` incluye DIAN,
   Bolsa, Renta IA, Suite contador, Centro IA, compras IA, OnlyOffice,
   contabilidad, nomina, Bre-B QR, buzon, tareas, chat empresarial,
   impresoras, menu visible, atajos POS, productos import/export,
   bodegas/traslados, verticales y analisis/control para revisar eventos
   reales y exportar CSV/JSON.

## GRAFOLOGIX grafologia OCR

1. Abrir `Administrar empresa > Analisis y control > GRAFOLOGIX`.
2. Cargar imagen desde PC, arrastrar archivo o tomar fotografia con camara.
3. Ajustar brillo, contraste, recorte central o recorte automatico por tinta.
4. Presionar `Analizar manuscrito`.
5. El frontend envia `multipart/form-data` a
   `/api/empresa/grafologia?empresa_id={id}&action=analizar`.
6. El backend valida empresa, permisos y tipo `image/*`, guarda la imagen en la
   carpeta de la empresa y ejecuta el motor Go puro.
7. Si la VPS tiene `GRAFOLOGIA_TESSERACT_ENABLED=1`, se intenta OCR por
   Tesseract CLI; si falla, el analisis geometrico continua.
8. El sistema guarda metricas, interpretacion y reporte HTML en
   `empresa_grafologia_analisis`.
9. El usuario puede abrir HTML imprimible, JSON o vista `PDF / imprimir`.
10. El resultado es orientativo; no debe tratarse como diagnostico ni decision
    automatizada.

## Camaras y DVR

1. Abrir `Administrar empresa > Analisis y control > Camaras`.
2. Registrar nombre, ubicacion, DVR/NVR, host, canal, fabricante, tecnologia
   origen y tipo de visor.
3. Si la fuente es RTSP u ONVIF, configurar un gateway HLS, WebRTC o MJPEG para
   que el navegador pueda mostrar video en tiempo real.
4. Asociar opcionalmente la camara a una estacion y marcar `Mostrar en
   estaciones`.
5. Guardar; el backend valida `empresa_id`, URL segura y registra en
   `empresa_camaras`.
6. En `Configuracion de estaciones`, activar `Mostrar Camaras` y elegir si
   cargan antes o despues de las estaciones.
7. Para convertir una estacion en visor, elegir `Tipo = Camara` y seleccionar
   `camara_id`.
8. En `Estaciones`, la tarjeta de camara abre el visor/modulo sin entrar al
   carrito; estaciones normales conservan su flujo operativo.
9. Pruebas negativas: intentar editar una camara de otra empresa debe devolver
   404/403; URL `javascript:` o `data:` no debe guardarse.

## Registro administrador

1. Usuario abre `web/login.html` y entra a registro de administrador.
2. Frontend envia datos al handler de autenticacion administrativa.
3. Backend crea usuario, prepara confirmacion segun configuracion y nunca expone
   clave ni token en consola o documentacion.
4. Si `Alertas sistema` tiene activo el check de registro, se envia aviso al
   correo configurado y se registra evento en `super_alertas_eventos`.
5. Pruebas: registro en PC y celular, login posterior, OAuth Google si aplica.

## Crear empresa

1. Administrador entra a `web/seleccionar_empresa.html`.
2. Presiona agregar empresa, elige tipo y completa datos.
3. Backend crea empresa en `pcs_empresas`, aplica preconfiguracion por tipo,
   permisos y modulos.
4. Ademas aplica la preconfiguracion Colombia `CO-2026-06`: impuestos base,
   configuracion legal de nomina, conceptos de nomina Colombia y marcador
   `preconfiguracion_colombia_fiscal_nomina` por `empresa_id`.
5. La preconfiguracion crea o reactiva una bodega base llamada `Bodega 1` por
   `empresa_id`, sin productos, existencias, movimientos ni stock simulado.
6. La preconfiguracion no crea empleados, ventas, liquidaciones ni documentos
   electronicos; solo deja parametros y ubicacion base reales para operacion
   posterior.
7. La creacion debe ser idempotente: doble clic, reintento o solicitud
   concurrente con el mismo administrador, tipo, nombre y NIT debe devolver la
   empresa ya creada sin insertar otra ni repetir avisos.
8. El backend prepara la carpeta empresarial
   `web/uploads/empresas/empresa_{id}_{slug}/` con subcarpeta `imagenes` y la
   carpeta privada `facturacion_electronica/firma_electronica`.
9. Si esta activo el aviso de empresa nueva, se notifica al super administrador
   solo cuando realmente se inserta una empresa nueva.
10. Pruebas: empresa creada, aparece en selector, entra a panel, conserva
   `empresa_id` correcto y lista `Bodega 1` como bodega activa.
11. Pruebas negativas: doble submit del mismo formulario y dos POST iguales no
   deben crear empresas duplicadas.

## Ayuda contextual en formularios

1. En `Administrar empresa > Productos > Nuevo producto`, cada campo del
   formulario muestra un boton `?` junto a su etiqueta.
2. Al presionarlo se abre una ventana pequena con texto operativo normal que
   explica que dato debe ir en el campo.
3. En `Administrar empresa > Facturacion electronica`, el mismo patron cubre
   firma DIAN, configuracion DIAN Colombia, configuracion por pais y
   configuracion avanzada de facturacion/impresion.
4. Los textos de DIAN no deben incluir PIN, claves, tokens, certificados ni
   datos privados reales; solo explican el tipo de dato y el cuidado operativo.
5. La ayuda es frontend, no sustituye validaciones backend ni permisos por
   `empresa_id`.

## Importar y exportar productos

1. En `Administrar empresa > Inventario (Productos) > Productos`, la parte alta
   de la seccion muestra la tarjeta `Exportar productos` y `Importar lista`.
2. La exportacion llama `GET /api/empresa/productos?action=exportar` con el
   `empresa_id` validado y conserva filtros visibles por busqueda/categoria.
3. Formatos soportados: CSV para Excel, JSON para integraciones y HTML
   imprimible en tamano carta o POS 80 mm.
4. La plantilla CSV se descarga desde
   `GET /api/empresa/productos?action=plantilla_importacion`.
5. La importacion sube un CSV multipart a
   `POST /api/empresa/productos?action=importar`; el backend valida permisos,
   campos obligatorios configurados, duplicados, categoria, bodega y stock
   inicial por `empresa_id`.
6. Si la fila trae stock inicial sin bodega, se usa la primera bodega activa de
   la empresa; si no existe bodega activa, la fila queda rechazada.
7. El resultado informa productos creados, omitidos y errores por fila para que
   el administrador corrija la lista sin repetir datos ya existentes.

## Logos de empresa y factura

1. El logo corporativo se configura en `Configuracion > Identidad visual` y se
   usa en panel, reportes y documentos generales.
2. El logo de factura se configura en la seccion de documento de venta o en
   facturacion electronica y puede ser diferente al corporativo.
3. Los uploads usan `/api/empresa/configuracion_avanzada/logo` con
   `tipo_logo=empresa` o `tipo_logo=factura`.
4. Los archivos quedan dentro de la carpeta de la empresa:
   `web/uploads/empresas/empresa_{id}_{slug}/imagenes/logos/empresa/` o
   `web/uploads/empresas/empresa_{id}_{slug}/imagenes/logos/factura/`.
5. Si no hay logo de factura dedicado, la impresion de factura puede usar el
   logo corporativo como respaldo. Esta configuracion no cambia XML DIAN,
   numeracion, CUFE/CUDE ni reglas fiscales.

## Ordenar empresas en el selector

1. Administrador entra a `web/seleccionar_empresa.html`.
2. La pantalla carga solo empresas visibles para esa cuenta: propias,
   delegadas o compartidas.
3. El usuario mantiene presionada una tarjeta en PC, o el asa de mover en
   celular, y la arrastra dentro del grupo de empresas con licencia activa o
   sin licencia activa.
4. El frontend guarda el orden de IDs visibles en
   `/api/user/configuracion` como preferencia del usuario autenticado y mantiene
   respaldo local del navegador si la red falla.
5. Al recargar, las empresas nuevas o no ordenadas se agregan despues de las ya
   guardadas, conservando el orden alfabetico base.
6. La tarjeta informativa de orden y el boton visible de restablecer fueron
   retirados de la pantalla; el arrastre conserva su soporte interno.
7. Seguridad: el orden no concede acceso a empresas; solo reordena tarjetas que
   `/super/api/empresas` ya autorizo para la sesion actual.
8. Pruebas: mover tarjetas en activas e inactivas, recargar y confirmar
   persistencia.

## Eliminar empresa

1. Solo el administrador propietario puede iniciar la eliminacion total desde el
   selector de empresas o desde `editar_empresa.html`.
2. Antes del borrado debe validar impacto, escribir el nombre exacto de la
   empresa, escribir `ELIMINAR` y aceptar el riesgo irreversible.
3. Justo antes de enviar el DELETE, el frontend pregunta si desea descargar toda
   la informacion de la empresa. Si acepta, abre
   `descargar_informacion_de_la_empresa.html` en una nueva pestana y luego
   continua; si cancela, continua sin descarga.
4. El endpoint destructivo recibe `descarga_ofrecida` para auditoria y mantiene
   las validaciones backend de propietario, confirmacion y aislamiento por
   empresa.
5. El backend elimina en transaccion los registros con `empresa_id` en la base
   operativa y en la base super, incluyendo licencias, pagos, invitaciones,
   accesos compartidos y datos de modulo cuando existan esas columnas.
6. Ademas limpia `usuario_configuracion.selector_empresas_orden_json` de todos
   los usuarios para quitar la empresa del orden personalizado del selector,
   invalida caches de licencia, resolucion de empresa y accesos compartidos, y
   borra carpetas empresariales asociadas.
7. Pruebas: no permitir borrado sin validaciones, ofrecer descarga previa,
   eliminar solo la empresa indicada y volver al selector sin filtrar datos de
   otra empresa; confirmar con otro administrador invitado o delegado que la
   empresa eliminada ya no aparece.

## Administradores delegados

1. El administrador principal entra a `seleccionar_empresa.html` y abre
   `Administradores`; ese enlace usa `scope=principal`, por lo que la lista solo
   muestra administradores invitados por la cuenta autenticada.
2. Invita administradores con rol forzado `administrador`; el backend guarda
   `administradores.usuario_creador` con el correo del principal y envia correo
   con enlace de invitacion.
3. Si el correo no existe o no esta confirmado, el invitado abre el enlace, acepta la invitacion, completa datos y crea su
   contrasena; sin token vigente no puede completar el registro.
4. Si el correo ya pertenece a un administrador confirmado, no se crea otra cuenta: se activa `admin_principal_delegaciones` y se envia solo un aviso por correo.
5. Al iniciar sesion, el delegado ve sus empresas propias y las empresas creadas por los principales que le compartieron portafolio como
   administracion delegada y entra con permisos empresariales efectivos.
6. El delegado no puede compartir esas empresas ni administrar otros
   administradores; el propietario sigue siendo el principal.
7. El super administrador si puede compartir, reenviar o revocar accesos de una
   empresa aunque no sea su propietario, por gobierno global del sistema; esta
   excepcion se valida en backend por rol super.
8. Eliminar desde el listado revoca la delegacion si la cuenta ya era de otro administrador; no borra su cuenta.
9. Pruebas: principal invita delegado, correo/enlace funciona, delegado ve
   empresas del principal, no ve empresas de otro principal y no puede compartir
   por URL ni boton; super administrador comparte una empresa ajena y queda
   auditado.

## Super administradores por invitacion

1. El super administrador entra al panel super y abre `Administradores` sin
   `scope=principal`.
2. Invita un correo con rol `super_administrador`.
3. Backend crea una cuenta pendiente, genera token y envia correo; no queda
   acceso activo hasta que el invitado complete registro con ese token.
4. Al aceptar, la cuenta conserva rol `super_administrador` y el login redirige
   al modulo de super administrador.
5. Pruebas: invitar super, intentar registro sin token, aceptar con token, login
   y entrada al panel super.

## Licencia gratis 15 dias

1. El catalogo base de licencias es global para todos los tipos de empresa:
   prueba gratis 15 dias, planes mensuales COP 60000, COP 110000 y COP 200000,
   y planes anuales COP 600000, COP 1100000 y COP 2200000.
   Las licencias base antiguas por tipo y addons de catalogo se eliminan del
   catalogo sin empresa asignada.
   El super administrador configura cuantas compras adelantadas de la misma
   licencia se permiten; el valor por defecto es 2.
2. Desde el checkout de licencia se obtiene resumen publico.
3. Si el total es cero o prueba permitida, `POST /licencias/activar_sin_pago`
   activa la licencia.
4. El backend valida que esa empresa no haya usado antes la licencia gratis,
   mirando historial completo de activaciones y licencias gratis antiguas,
   aunque la licencia anterior ya este vencida o inactiva.
5. La activacion debe ser idempotente si el primer intento ya dejo la licencia
   vigente.
6. Al quedar activa, el sistema puede enviar un correo de bienvenida al
   administrador/cliente comprador con el mensaje `licencia_activation_payment`,
   configurable en `web/super/formato_para_emviar_email.html`.
   La licencia de software ya no se adjunta por correo; solo se descarga desde
   Administrar empresa > Licencia > Licencia del sistema usando la plantilla
   `licencia_software_pdf`.
   Si el pago comercial queda aprobado con valor mayor a cero, y la regla esta
   activa en Super administrador > Licencias, ademas se emite automaticamente una
   factura electronica desde la empresa interna `Powerful Control System`
   (tambien reconoce el nombre existente `Powerful Control Systen`) y el PDF
   resumen de esa factura se adjunta al mismo correo de bienvenida. El documento se guarda en `empresa_facturacion_documentos` de la
   empresa emisora y el pago queda marcado con
   `licencia_factura_electronica_emitida` en `pagos_epayco` o `pagos_wompi`
   para idempotencia.
   La empresa emisora interna se resuelve por
   `configuraciones.licencias.facturacion_empresa_sistema_id`. La licencia
   tecnica heredada `PCS_SYSTEM_INTERNAL_PERPETUAL` se desactiva si existe, de
   modo que la empresa interna debe comprar, renovar o adquirir una licencia
   comercial vigente como cualquier otra empresa. Las activaciones con total
   pagado cero por prueba o descuento total pueden enviar bienvenida si esta
   activa, pero no emiten factura electronica en el flujo final.
7. La empresa puede descargar el mismo documento desde Administrar empresa >
   Licencia > Licencia del sistema; el endpoint
   `/api/empresa/licencia_sistema/pdf` debe quedar protegido por permisos de
   empresa y aislamiento `empresa_id`.
8. Una licencia de prueba de 15 dias con valor cero no se renueva desde el
   historial; si el administrador necesita continuar, debe escoger una licencia
   comercial desde el cambio de plan.
9. En licencias pagadas, `pagar_licencia.html` consulta
   `/api/public/licencias/payment_methods`; Epayco y Wompi deben aparecer si
   estan configurados, salvo que el super administrador los apague de forma
   explicita con `epayco.enabled=0` o `wompi.enabled=0`.
10. Si una licencia comercial se paga antes del vencimiento actual, la nueva
   vigencia se agenda desde la fecha fin acumulada mas lejana de la empresa.
   Ejemplo: si vence el 10 de junio y paga 30 dias el 1 de junio, la licencia
   pagada inicia el 10 de junio y vence el 10 de julio; un segundo pago
   anticipado inicia desde el 10 de julio. Los webhooks/consultas repetidos de
   la misma referencia quedan idempotentes con `licencia_activation_status`.
11. Los codigos de descuento para licencias se crean desde Super administrador >
   Comercial y licencias > Codigos descuento. El formato tecnico es
   `CODIGO=10%`, `CODIGO=50000` o `CODIGO=gratis`; el checkout los calcula en
   `/api/public/licencias/checkout_summary` y los conserva al pagar por Epayco,
   Wompi o activacion sin pago. Cada codigo queda limitado a un uso por empresa
   mediante `pagos_epayco`, `pagos_wompi` y `licencias_activaciones_gratis`.
12. Los asesores de venta se configuran desde Super administrador > Comercial y
   licencias > Asesor en ventas. Cada asesor puede tener porcentaje del primer
   ano, porcentaje anual desde el segundo ano y meses de renovacion propios; el
   backend calcula si el pago corresponde a `primer_anio`, `renovacion_anual` o
   queda fuera del plazo antes de registrar la comision.
13. Pruebas: activar una vez, reintentar sin duplicar mientras sigue vigente,
   bloquear segundo uso real despues del vencimiento y comprobar que el
   historial muestra otras licencias cuando la prueba no es renovable, ademas
   de validar que el correo capturado o enviado no incluya PDF de licencia y,
   cuando el pago sea mayor a cero y la configuracion lo permita, incluya el PDF
   de factura electronica en el mismo mensaje; confirmar tambien que descuento total o valor cero no genera factura
   electronica. La descarga empresarial debe devolver `application/pdf`; para
   pago, seleccionar Epayco y Wompi, aceptar terminos y comprobar que cada
   proveedor pasa a verificacion con referencia propia. Para renovaciones comerciales, simular
   pago anticipado con licencia vigente y validar que `fecha_inicio` queda en el
   vencimiento anterior, que `fecha_fin` suma la duracion comprada y que repetir
   la misma referencia no vuelve a extender.

## Configurar empresa

1. El menu `Configuracion` abre paginas independientes por seccion.
2. Cada pagina carga datos por `empresa_id`, guarda solo su seccion y conserva el
   resto de configuracion.
3. Las preferencias flexibles de estaciones/carrito usan
   `empresa_estacion_prefs.estaciones_config`.
4. La configuracion del carrito permite activar pitido y vibracion de botones por
   separado para PC y celular desde `carrito_ui_global`.
5. La configuracion de impresora permite activar `Mostrar deducido del impuesto
   en la impresion`; el carrito usa `base_gravable` y `valor_impuesto` ya
   calculados para mostrar base e impuesto en el papel, sin cambiar el XML ni la
   validacion legal DIAN de la factura electronica.
6. La misma pagina permite definir por checks los campos imprimibles del recibo
   operativo de venta (`impresion_recibo_items_json`) y de reportes de corte o
   cierre de turno (`impresion_corte_items_json`). Esta configuracion no aplica
   a factura electronica, porque sus campos legales quedan definidos por la
   normativa DIAN y el XML/documento electronico vigente.
7. Pruebas: guardar seccion, recargar pagina, validar que otra seccion no cambio
   e imprimir una venta con impuesto para confirmar el bloque.

## Venta por peso con bascula

1. Abrir `Administrar empresa > Ventas > Carrito y venta directa` con
   `empresa_id`.
2. El navegador debe ser Chrome o Edge en HTTPS/local y la bascula debe exponer
   un puerto serial/USB compatible con Web Serial.
3. En Configuracion carrito, activar la tarjeta independiente `Bascula
   electronica`. Por defecto esta apagada para todas las empresas.
4. En el carrito, seleccionar baudios y unidad leida (`kg`, `g` o `lb`),
   presionar `Conectar bascula` y aceptar el permiso del navegador.
5. Si el plato o recipiente ya esta sobre la bascula, usar `Tara local`; PCS
   descuenta ese peso en pantalla sin enviar comandos al equipo.
6. El producto debe estar configurado con unidad de peso (`kg`, `g`, `lb`, `oz`
   o alias). Si esta en `unidad`, el sistema bloquea aplicar peso para evitar
   ventas decimales incorrectas.
7. Escanear o digitar el SKU/codigo de barras. Si `aplicar peso` esta activo y
   hay lectura mayor a cero, PCS manda la cantidad real al item del carrito.
8. El backend acepta cantidades decimales solo para unidades de peso; en
   unidades normales sigue exigiendo cantidades naturales positivas.
9. Pruebas: producto `kg` con lectura real, producto `unidad` con lectura activa
   debe bloquearse, desconectar/reconectar la bascula, validar que otro
   `empresa_id` no accede al carrito ni a items ajenos mediante los endpoints
   existentes.

## Firma electronica DIAN

1. La empresa configura primero los datos base de facturacion electronica
   Colombia.
2. Al cargar la firma, el endpoint multipart valida `empresa_id`, configuracion
   DIAN existente, archivo no vacio, tamano maximo 10 MB y contenido con llave
   privada RSA.
3. El backend guarda la llave privada y el certificado publico extraido en
   `web/uploads/empresas/empresa_{id}_{slug}/facturacion_electronica/firma_electronica/`.
   Para P12/PFX con multiples bolsas o cadenas de certificados, el backend
   convierte internamente a PEM con la dependencia existente antes de extraer la
   llave RSA; para P12/PFX modernos no soportados por Go, el contenedor backend
   usa OpenSSL con la clave en una variable de entorno temporal del proceso.
4. Las rutas guardadas en la configuracion DIAN son referencias internas `file:`
   a archivos con permiso `0600`; no deben convertirse en enlaces publicos.
5. El backend extrae del X.509 la fecha real de vencimiento y la guarda en
   `certificado_vencimiento` / `certificado_vencimiento_en`.
6. La clave del P12/PFX se usa solo para decodificar el archivo; no se guarda ni
   se muestra en claro. La pantalla muestra un resumen seguro de ultima carga:
   fecha/hora, archivo, formato, titular, serial y estado de clave.
7. La accion
   `/api/empresa/facturacion_electronica/dian?action=vencimiento_certificado`
   muestra estado, dias restantes y ventana de alerta. Si el certificado esta
   vencido o a 30 dias de vencer, envia correo al administrador de la empresa
   como maximo una vez cada 24 horas.
8. Despues de cargar firma, el siguiente paso operativo es
   `action=validar_credenciales` y luego `action=pruebas_dian`.
9. La pagina `Facturacion electronica > Pasar test DIAN` muestra estado de
   ambiente, rango, TestSetId y credenciales; desde alli se guarda el objetivo
   del set que aparece en el portal DIAN, incluyendo totales requeridos y
   minimos aceptados por facturas, notas debito y notas credito. La barra
   `Avance de validacion DIAN` muestra un porcentaje operativo 0-100% por
   hitos, pero no reemplaza el acuse final de DIAN.
10. `Ejecutar set automatico` usa los valores guardados para generar el lote
   completo; los botones `Enviar factura`, `Enviar nota debito` y `Enviar nota
   credito` permiten probar un documento a la vez y ver si fue recibido,
   aceptado, rechazado o queda pendiente.
   Para la empresa interna Powerful Control System, el set registrado desde el
   portal de habilitacion es: 50 documentos totales, 30 facturas electronicas,
   10 notas debito, 10 notas credito, minimo aceptado total 1 y minimo aceptado
   de facturas 1.
11. El resultado visual debe mostrar resumen por estado, aceptados por tipo,
   mensaje de recepcion y si el minimo configurado ya se cumple. No se debe
   declarar produccion local hasta tener acuse suficiente de DIAN/proveedor.
12. Para endpoint oficial SOAP/WCF DIAN no se exige `token_emisor_ref`; ese
   token solo aplica a proveedor/API con bearer token. En habilitacion real si
   es obligatorio `test_set_id`, porque DIAN lo usa para `SendTestSetAsync`.
13. El sobre SOAP oficial vigente para DIAN firma el header `wsa:To`, referencia
   el `BinarySecurityToken` con `wsse:Reference URI="#X509-..."` e incluye
   `InclusiveNamespaces`. Esta forma es la que debe conservarse para evitar
   errores de seguridad del transporte WCF.
14. Las pruebas DIAN automaticas no aceptan simulacion: deben enviar al ambiente
   de habilitacion, recibir `ZipKey` cuando aplique y consultar `GetStatusZip`
   hasta un acuse final. Solo una ejecucion real con acuse aceptado puede
   cambiar la empresa a habilitada/produccion local.
15. La pagina `Facturacion electronica > Tutorial DIAN` resume el flujo
   operativo para conectar DIAN: datos del portal, configuracion Colombia,
   carga de firma, set completo, acuse final y activacion local
   de produccion. Debe mantenerse sin secretos reales.
16. Estado operativo actual: el set real configurado por empresa debe quedar
   como envio real con HTTP 200, TrackId/ZipKey y respuesta inicial `Batch en
   proceso de validacion`. Ese estado todavia no equivale a aceptacion final.
17. Lo que falta en el modulo DIAN/documentos electronicos es consultar y
   persistir el acuse final por TrackId hasta aceptado/rechazado, reconciliar
   `facturacion_electronica_reintentos` y `empresa_facturacion_documentos`,
   mostrar un resumen claro en la pantalla y habilitar produccion local solo
   cuando los minimos aceptados del set esten cumplidos.
18. Pruebas: subir PEM/P12 valido, verificar carpeta empresarial, validar que el
   archivo no se guarda en `/uploads/dian`, y confirmar que otro `empresa_id` no
   puede consultar ni modificar la configuracion. Luego usar `Verificar
   vencimiento` en la pantalla para confirmar que se ve fecha, dias restantes y
   estado de alerta. En `Centro de habilitacion DIAN`, guardar objetivo, validar
   credenciales, ejecutar al menos un envio manual de factura de prueba y correr
   el set automatico real; despues consultar `GetStatusZip` hasta cierre real
   del acuse.

## Login usuarios operativos

1. El administrador de una empresa crea el usuario desde `Administrar usuarios`;
   el sistema envia invitacion por correo y guarda token temporal en `users`.
   Este alta operativa usa permisos `seguridad:C/U/D` y auditoria por
   `empresa_id`; no debe pedir `aprobado_por` ni `codigo_aprobacion`. Esos
   campos se reservan para cambios de roles o matriz fina de permisos.
   Si el correo no se puede entregar, el usuario queda pendiente y la pantalla
   debe permitir reintentar o reenviar confirmacion sin crear duplicados.
   Si se intenta crear otra vez el mismo correo en la misma empresa, el endpoint
   responde `409` con `usuario_existente` sin exponer tokens; la interfaz debe
   recargar con `include_inactive=1`, resaltar el usuario y dejar disponible
   `Reenviar confirmacion`.
   El mismo correo puede existir como administrador del panel principal y como
   usuario operativo, porque son credenciales y pantallas distintas; en `users`
   la unicidad vigente es por `lower(email), empresa_id`.
2. El usuario abre `login_usuario.html` desde la invitacion para completar
   registro o iniciar con Google. Sin invitacion o usuario empresarial confirmado
   no hay alta publica. En este primer ingreso, un usuario pendiente puede tener
   `estado=inactivo`; si el token es valido, el sistema permite crear la
   contrasena, confirma el correo y cambia el estado a `activo`. Un usuario ya
   confirmado e inactivo sigue bloqueado hasta que el administrador lo active.
   Los errores del primer ingreso muestran mensaje especifico, detalle y
   `request_id` cuando el backend los marca como publicos seguros; los errores
   internos no marcados siguen protegidos para no exponer SQL, rutas, claves o
   datos sensibles.
3. `Iniciar sesion con Google` usa `/auth/google/usuario/login`, conserva
   `empresa_id` y token de invitacion en cookies tecnicas de corta vida y vuelve
   por `/auth/google/callback`.
4. El callback solo abre sesion si Google confirma el correo y este coincide con
   una invitacion vigente o con un usuario de empresa ya confirmado. Si falta
   contrato vigente, vuelve al formulario para aceptarlo.
5. La sesion redirige siempre a `administrar_empresa.html?id={empresa_id}`; el
   panel carga rol, permisos y estaciones asignadas de esa empresa.
6. Para `cajero`, PCS ya no muestra seleccion manual de caja en el login.
   El navegador crea un `pcs_dispositivo_id` local por computador y usa la caja
   asociada a ese equipo en la empresa, por ejemplo `CAJA-1`. La asociacion se
   registra desde Configuracion > Impresoras y caja > Impresora por computador
   y se guarda por `empresa_id`/navegador. Si la caja sigue activa, PCS entra
   automaticamente y propaga `caja_codigo`, `caja_nombre` y `caja_descripcion`
   a estaciones, carrito y corte de caja. Si hay varias cajas y el computador
   aun no esta asociado, el login no abre selector manual: deja marcada la
   sesion como pendiente de asignacion para que un administrador configure ese
   equipo.
7. Para `cajero`, el menu queda limitado a `Venta directa`, `Estaciones`,
   `Corte de Caja` y `Buscar ventas y facturas`. El carrito debe cargar
   completo: catalogo de productos, servicios, recetas, clientes, descuentos,
   propinas/comisiones y valores por medio de pago. La busqueda de ventas y
   facturas permite reimprimir, consultar y reenviar documentos del POS sin
   mostrar paginas administrativas de Productos o Clientes; internamente el
   rol conserva acceso documental a facturas electronicas para que la consulta,
   vista previa y reenvio no fallen por permisos.
8. Pruebas: Google sin invitacion debe rechazar, Google con invitacion debe
   consumir token y entrar, correo ambiguo exige enlace de empresa, tema claro u
   oscuro se conserva, el boton `Instalar app` permanece visible y el cajero
   no ve selector manual de caja: debe entrar con caja detectada por computador
   o quedar pendiente de asociacion si el equipo no esta registrado.

## Abrir, usar y cerrar caja

1. El usuario entra a caja desde `Corte de Caja` o desde la estacion Caja.
2. La empresa puede configurar varias cajas fisicas en
   `estaciones_config.cajas_config`, cada una con codigo, nombre, descripcion y
   estado activo. La estacion Caja muestra esos nombres, por ejemplo
   `CAJA-1 - FRUTERA`.
3. El login operativo de cajeros usa la caja asignada al computador; la
   asignacion se administra desde Configuracion > Impresoras y caja, sin pedir
   al cajero que elija caja al iniciar sesion.
4. Al hacer clic en una caja configurada, `corte_de_caja.html` recibe
   `caja_codigo`, `caja_nombre` y descripcion para abrir el corte de esa caja.
5. La caja puede abrirse manual o automaticamente segun flujo vigente.
6. Cada usuario/caja mantiene turno, pagos, ingresos, egresos y reporte
   independiente.
   Para el rol `cajero`, registrar ingresos o egresos manuales es opcional por
   empresa y rol: el administrador debe activarlo en Configuracion de impresora
   y caja > Configuracion operativa de cobro > Override por rol. El check no
   modifica los cobros normales del carrito.
7. `Corte automatico` calcula desde apertura hasta el momento actual sin pedir
   fechas.
7. `Cerrar turno e imprimir reporte` imprime y luego cierra sesion.
8. Pruebas: abrir caja con dos usuarios, registrar movimientos, cerrar una caja
   sin afectar la otra.

## Venta directa

1. `Venta directa` abre `carrito_de_compras.html` en modo venta directa.
2. Debe usar el mismo carrito unificado de estaciones, con configuracion global
   del carrito.
3. El cajero agrega productos/servicios/recetas, cliente opcional u obligatorio,
   abonos y pagos mixtos.
   La fila superior del lector separa `Codigo de barras o SKU` y `Busqueda por
   nombre`: el cajero puede escanear/digitar codigo, o escribir el nombre del
   producto en el campo de la derecha y agregarlo desde la misma seccion. El
   buscador de catalogo avanzado conserva codigo, SKU, codigo de barras o
   nombre del producto, y conserva codigo del servicio/receta. Las coincidencias
   exactas por codigo/SKU deben aparecer primero. El formulario rapido de
   clientes permite registrar persona natural
   o empresa con NIT/DV, regimen IVA, responsabilidad tributaria, correo,
   telefono y direccion fiscal para facturacion electronica.
   Las cantidades de items se pueden cambiar desde la tabla del carrito. Deben
   ser numeros naturales positivos (`1, 2, 3...`) salvo unidades de peso
   (`kg/g/lb/oz`), donde se permiten decimales positivos. El backend recalcula
   totales y reservas de inventario en cada `PUT`.
   `Devolver producto` pide confirmacion integrada de dos pasos, elimina la
   linea del carrito con `DELETE`, libera el inventario reservado y recalcula la
   cuenta; las notas credito o devoluciones fiscales deben manejarse por el
   flujo documental correspondiente.
   Si la empresa detectada es Colombia o el carrito usa `COP`, precios y pagos
   del carrito se capturan, muestran y sincronizan como pesos enteros positivos,
   sin centavos ni sufijo `.00`.
4. La venta directa usa el carrito canonico `VENTA-DIRECTA-{empresa_id}-0` y
   puede abrirse en pantalla completa desde una tarjeta compacta en el encabezado
   del cliente; el icono dentro de esa tarjeta cambia a `Salir` y vuelve a la
   vista normal.
5. Visualmente debe conservar el modo plano del carrito: tarjetas sin sombras ni
   apariencia 3D, pero con el fondo estructural mas oscuro que las tarjetas para
   diferenciar zonas en cualquier apariencia.
6. La vista enfocada del carrito debe mantener visibles, en este orden, las
   tarjetas base: cliente, productos agregados, detalle del pago y acciones del
   carrito. Las preferencias antiguas no deben ocultar esa estructura minima.
7. Si el carrito ya fue seleccionado y una sincronizacion secundaria falla, la
   pantalla no debe desmontarse; debe conservar la venta visible y mostrar una
   advertencia operativa.
8. Pruebas: entrar con rol `cajero` por `login_usuario`, crear/asignar cliente
   natural y empresa, agregar item por nombre y por codigo, cambiar cantidad,
   devolver producto, pagar, imprimir, validar inventario y caja, abrir/salir
   de pantalla completa y revisar contraste fondo/tarjetas.

## Estaciones y carrito

1. La pagina de estaciones carga configuracion y carritos por `empresa_id`.
2. Al activar una estacion cambia estado y otros usuarios deben ver el cambio.
3. Si el check de primer clic esta activo, el primer clic solo activa; el segundo
   entra al carrito.
4. El carrito de estacion comparte UI y reglas con venta directa.
5. La apariencia del carrito de estacion tambien depende de `carrito-flat-page`:
   fondo mas oscuro que tarjetas, sin sombras, con botones de accion visibles
   como botones.
6. Si en Configuracion > Impresora esta activo el deducido de impuesto, la
   impresion del recibo o factura muestra base gravable e impuesto deducido por
   los impuestos del carrito.
7. Pruebas: dos sesiones/usuarios, estado compartido, abrir carrito correcto,
   contraste visual y ausencia de relieves.

## Pagar e imprimir

1. El usuario presiona pagar en el carrito.
2. Backend valida caja, items, totales, abonos, cliente obligatorio si aplica,
   descuentos, inventario y permisos.
3. El inventario de productos y recetas ya debe estar reservado/descontado desde
   que se agrego el item al carrito. Esto permite multiples cajas simultaneas
   sin sobrevender stock.
4. El backend cierra el carrito con una transicion atomica: solo una solicitud
   puede cambiarlo de abierto a pagado. Reintentos, doble clic o concurrencia
   reciben respuesta idempotente y no duplican caja, documento, metricas ni
   movimientos de inventario.
5. Se registra venta/pago y se genera documento.
6. En venta directa o estaciones, despues del pago exitoso PCS reabre la cuenta
   operativa con `reset_items=1` y la pantalla debe quedar sin items, abonos ni
   cliente para el siguiente cliente, sin recarga manual.
7. La impresion debe salir en blanco y negro como papel real, POS 80mm por
   defecto, sin tema claro/oscuro. El recibo operativo debe respetar los checks
   por empresa de fecha, cajero, cliente, metodo, impresora, copias, carrito,
   codigo y empresa.
8. Si hay QR DIAN activo y documento con CUFE/CUDE/codigo, se imprime al final.
9. Venta a credito: si el medio es `credito_cliente` o un tramo mixto de
   credito, el carrito exige cliente registrado con cupo activo; consulta
   `empresa_creditos_clientes_limites`, valida disponible, cierra la venta y
   crea `empresa_creditos` ligado a `carritos_compras.id` en `venta_origen_id`.
   El valor a credito no incrementa efectivo de caja, pero queda visible en
   recibo/factura y en cartera.
10. Pruebas: efectivo, debito, credito, credito cliente, otro, pago mixto,
   vuelto, abono, descuento, dos cajeros simultaneos y doble solicitud de pago
   sobre el mismo carrito.

## Facturacion electronica

1. El submenu de facturacion permanece visible, pero las paginas internas se
   muestran segun pais detectado y licencia.
2. Colombia usa configuracion DIAN, firma, resolucion, documentos electronicos,
   tutorial operativo, pruebas y cola documental.
3. Panama y Ecuador tienen paginas propias con configuracion de DGI/SRI.
4. Credenciales, firma, NIT/RUC y trazabilidad son por empresa.
5. En Colombia el envio real de habilitacion puede quedar primero como `Batch en
   proceso de validacion`; el sistema debe tratarlo como pendiente hasta que
   `GetStatusZip` entregue acuse final.
6. Pruebas: guardar configuracion por pais, validar checklist, generar documento,
   abrir `Tutorial DIAN`, enviar correo si aplica, revisar cola/reintentos y
   reconciliar estados DIAN finales por TrackId.

## Modo offline

1. La empresa activa modo offline y marca de documento offline si corresponde.
2. Cada cajero debe haber iniciado sesion y tener una caja abierta/cargada antes
   de perder internet. La venta offline queda ligada a `empresa_id`, usuario,
   codigo de caja, estacion/carrito y `sync_key` unico.
3. Si se pierde internet en caja/carrito con offline activo, aparece aviso
   persistente y se permite vender e imprimir provisionalmente solo para la caja
   abierta de ese cajero.
4. Si se pierde internet en modulo sin soporte offline, el aviso debe pedir
   esperar reconexion.
5. Al volver internet, se muestra aviso, se registra auditoria y se sincroniza la
   cola por `/api/empresa/offline_ventas`. El backend rechaza ventas de otro
   cajero o sin caja explicita y trata reintentos sobre carritos ya pagados como
   idempotentes para no duplicar caja, inventario ni documentos.
6. Pruebas: cortar red, vender, imprimir, restaurar red, sincronizar una sola
   vez, y repetir con dos cajeros/cajas para validar colas separadas.

## Energia solar

1. La empresa entra por Administrar empresa > Energia solar si la licencia y el
   rol permiten `energia_solar`.
2. Las preconfiguraciones incluyen el modulo apagado por defecto; al activarlo,
   la empresa registra proveedor, modelo, instalacion, bateria, BMS y correo de
   alertas.
3. El tecnico solar solo consulta dashboard, lecturas, eventos y alertas.
4. Administrador o supervisor configura sistemas, alertas y lecturas manuales o
   recibidas desde gateway/API.
5. Las lecturas disparan eventos por umbral o estado y pueden enviar correo si
   el sistema lo tiene activo.
6. Pruebas: crear sistema Victron/SMA/SolarEdge/gateway, registrar lectura con
   SOC bajo, validar evento/correo, intentar leer con `tecnico_solar` y guardar
   con `tecnico_solar` esperando bloqueo.

## Reportes de turno

1. `Ver reporte de mi turno` calcula el turno del usuario autenticado y caja
   actual.
2. El reporte muestra datos empresa, fecha/hora, usuario, consecutivo, detalle de
   ventas ordenado por fecha/hora y resumenes configurables.
3. Debe incluir ventas, descuentos, ingresos, egresos, productos, servicios,
   medios de pago y efectivo esperado segun checks.
4. Los checks de `Configuracion > Impresora` para corte/cierre controlan campos
   del encabezado y columnas del detalle impreso en `corte_de_caja.html` y en
   `reportes_turnos.html`.
5. Vista e impresion deben adaptarse a POS 80mm y carta; POS 80mm es default.
6. Pruebas: turno con ventas, descuento, tarjeta, ingreso, egreso, anulacion,
   exportar/imprimir, ocultar cajero o fecha desde la configuracion y confirmar
   que no aparece en el reporte impreso.

## Cajeros simultaneos y estaciones asignadas

1. El administrador crea los usuarios de la empresa en
   `Administrar usuarios`.
2. En la seccion `Acceso a estaciones por cajero` activa el control y elige el
   usuario cajero.
3. Marca por check las estaciones que ese usuario puede ver y operar. Si el
   check `Ver estacion Caja y corte de turno` queda apagado, la tarjeta Caja no
   se muestra para ese usuario.
4. El tablero de estaciones filtra la vista por usuario autenticado. Los
   endpoints de carritos e items validan la misma regla en backend, por lo que
   editar URL, cache o consola no permite operar estaciones no asignadas.
5. La estacion Caja, los totales de caja, caja abierta y reporte de turno se
   mantienen independientes por `usuario_creador`; varios cajeros pueden operar
   la misma empresa al mismo tiempo con reportes separados.
6. Pruebas: dos usuarios cajeros en la misma empresa, estaciones diferentes,
   estados visibles compartidos, bloqueo 403 al intentar abrir/agregar/pagar una
   estacion no asignada y corte de turno independiente.

## Rol portero

1. El rol `portero` se crea como rol base de cada tipo de empresa.
2. En el menu empresarial solo debe quedar visible `Estaciones`.
3. En `Estaciones`, el portero puede ver el estado de las estaciones y activar
   una estacion disponible, pero no debe abrir carrito, Caja, corte, items,
   venta directa, pagos, abonos ni configuracion.
4. Backend mantiene la restriccion aunque el usuario intente llamar la API:
   `carritos_compra` permite al portero solo `GET` de estado y `PUT
   action=activar_estacion`; `carritos_compra/items` queda bloqueado.
5. Pruebas: usuario operativo con rol `portero`, entrar por `login_usuario`,
   confirmar menu con solo Estaciones, activar una estacion, intentar abrir
   carrito/items/pagar por URL o consola y recibir 403.

## Rol Servicio de limpieza

1. El rol `servicio_limpieza` se crea como rol base de cada tipo de empresa.
2. En el menu empresarial solo debe quedar visible `Estaciones`.
3. En `Estaciones`, el usuario puede ver el estado de cada estacion. Si la
   estacion esta marcada como sucia, al hacer clic reporta aseo terminado y el
   sistema cambia la estacion a limpia/disponible.
4. Si la estacion no esta sucia, el clic solo muestra un aviso; no abre carrito,
   no activa estacion, no entra a Caja ni ejecuta ventas.
5. Backend mantiene la restriccion aunque el usuario intente llamar la API:
   `carritos_compra` permite solo `GET` del tablero, `carritos_compra/items`
   queda bloqueado y el cambio sucia->limpia pasa por
   `/api/empresa/estacion_aseo?action=finalizar`.
6. Pruebas: usuario operativo con rol `servicio_limpieza`, entrar por
   `login_usuario`, confirmar menu con solo Estaciones, marcar una estacion
   sucia como limpia, intentar activar estacion limpia, abrir carrito/items/caja
   por URL o consola y recibir bloqueo.

## Roles empresariales comunes

1. Las empresas cuentan con roles base para asignar usuarios sin crear permisos
   desde cero: `supervisor_sucursal`, `vendedor`, `recepcion`, `jefe_bodega`,
   `responsable_bodega`, `recursos_humanos`, `tecnico_solar`, `cajero`, `portero`,
   `servicio_limpieza`, `contador`, `empresario`, `compras`, `inventario`,
   `contabilidad` y `auditor`.
2. `tecnico_solar` solo consulta el estado de energia solar: dashboard,
   lecturas, eventos y alertas. No puede modificar sistemas ni configuracion.
3. `jefe_bodega` administra inventario y bodegas: existencias, traslados,
   categorias, recetas y codigos; no puede operar ventas, caja ni eliminar
   inventario.
4. `responsable_bodega` administra la bodega asignada con inventario,
   existencias y traslados; no opera ventas, caja, configuracion ni elimina
   inventario.
5. `recursos_humanos` gestiona horarios, asistencia y nomina operativa; no abre
   ventas, caja ni configuracion general.
6. Pruebas: crear usuarios con esos roles, iniciar por `login_usuario`, validar
   menu visible y probar llamadas directas a endpoints fuera del alcance con
   respuesta 403.

## Rol contador

1. El rol `contador` se crea como rol base de cada tipo de empresa.
2. En el menu empresarial solo debe quedar visible `Centro financiero y
   contable` e `Impuestos`.
3. Dentro del centro financiero, los accesos rapidos y el submenu deben ocultar
   contabilidad avanzada, creditos, caja, cobranza, tesoreria, portal contador y
   demas paginas no permitidas.
4. Backend conserva el control efectivo: `contador` solo tiene `R` en
   `finanzas` y `facturacion` para consultar finanzas e impuestos. Cualquier
   `POST`, `PUT`, `PATCH`, `DELETE` o accion de aprobacion debe devolver 403.
5. Pruebas: usuario operativo con rol `contador`, entrar por `login_usuario`,
   confirmar menu limitado, abrir finanzas e impuestos, intentar guardar un
   impuesto o crear movimiento financiero y recibir 403.

## Rol empresario

1. El rol `empresario` se crea como rol base de cada tipo de empresa.
2. En el menu empresarial solo debe quedar visible `Reportes ejecutivos`.
3. Dentro del centro de reportes, el usuario debe abrir la vista ejecutiva de
   resultados y no debe ver reportes de turnos/caja.
4. Backend conserva el control efectivo: `empresario` solo tiene `R` en
   `reportes`. Cualquier intento de operar ventas, caja, inventario, finanzas,
   usuarios, configuracion o acciones `C/U/D/A` debe devolver 403.
5. Pruebas: usuario operativo con rol `empresario`, entrar por `login_usuario`,
   confirmar menu limitado, abrir centro de reportes, exportar o previsualizar
   un reporte permitido e intentar abrir turnos/caja/ventas por URL recibiendo
   bloqueo.

## Alertas super administrador

1. `web/super/alertas_sistema.html` concentra alertas y notificaciones.
2. Las opciones de correo para registros y empresas nuevas se guardan en
   `super_alertas_config`.
3. Los envios quedan auditados en `super_alertas_eventos`.
4. Un fallo SMTP no debe bloquear el flujo de negocio que disparo la alerta.
5. Pruebas: guardar checks, enviar prueba, crear admin/empresa y revisar evento.

## Email corporativo Mailu

1. El super administrador configura `web/super/email_corporativo.html`.
2. Si `auto_create` esta activo, cada empresa nueva recibe un correo unico basado
   en su nombre y dominio configurado.
3. Si el modulo global esta desactivado, el correo queda generado pero pendiente.
4. Si `mailu_direct` esta activo, se intenta crear el buzon en Mailu mediante
   `deploy/scripts/vps-provision-mailu-mailbox.sh`.
5. La creacion de empresa no debe fallar por errores del servidor de correo.
6. En el panel empresarial se muestra la tarjeta de webmail solo si el modulo esta
   activo y la empresa tiene cuenta.
7. La tarjeta detecta la apariencia activa y envia `theme=light|dark` para usar
   `PCSLight@custom` o `PCSDark@custom` en SnappyMail.
8. En `Configuracion > Email corporativo`, la empresa puede desactivar la
   apertura automatica del buzon; por defecto queda activa.
9. Desde la misma pagina se puede cambiar la contrasena interna del buzon. El
   backend cifra la clave, no la devuelve al navegador y actualiza Mailu si
   `mailu_direct` esta disponible.
10. Si aparecen estados de error, usar `Probar Mailu` en super administrador para
   validar el contenedor `pcs-mailu-admin` y el comando directo antes de
   reintentar provision.
11. Pruebas: guardar configuracion, sincronizar empresas existentes, crear empresa
   duplicada de nombre similar, comprobar sufijo unico, abrir webmail, desactivar
   autoapertura y cambiar clave sin exponerla.

## Nomina multi-sede y documentos DIAN Colombia

1. Al abrir nomina, la configuracion de empresa debe traer salario minimo
   mensual y auxilio de transporte legal desde `empresa_nomina_configuracion`;
   las empresas nuevas/existentes de preproduccion reciben esos valores por
   preconfiguracion Colombia.
2. La tarjeta `Parametros legales` consulta la version aplicada y disponible.
   Si hay version pendiente, el administrador puede aplicar manualmente o dejar
   activa la autoactualizacion para el worker diario.
3. La aplicacion legal actualiza parametros reales de impuestos/nomina y guarda
   version en `empresa_parametros_legales_aplicados`; no crea empleados,
   liquidaciones ni documentos.
4. La ficha de nomina del empleado guarda `sede_codigo`, `sede_nombre` y
   `centro_costo`; estos valores se copian a cada liquidacion para conservar la
   trazabilidad historica aunque luego cambie la ficha.
5. Al crear empleado nuevo, el formulario sugiere salario minimo y auxilio
   legal, pero el usuario puede ajustar segun contrato real.
6. La liquidacion se genera desde asistencia, novedades aprobadas, recargos,
   comisiones, provisiones y deducciones; el dashboard resume empleados,
   liquidaciones, pagos, costo empresa y sedes activas.
7. Para Colombia, la seccion avanzada consulta conceptos, novedades, PILA y el
   resumen de documentos electronicos de nomina.
8. `GET /api/empresa/nomina?action=documentos_electronicos_colombia` valida el
   periodo y muestra liquidaciones listas por empleado para documento soporte de
   pago de nomina electronica.
9. `POST /api/empresa/nomina?action=preparar_nomina_electronica` prepara el lote
   por empleado con devengados, deducciones, neto, IBC, sede y centro de costo.
   El envio real a DIAN sigue dependiendo de firma, CUNE, numeracion,
   credenciales y transporte documental configurados por empresa en facturacion
   electronica.
10. La pagina `web/administrar_empresa/nomina_tutorial.html` debe abrirse desde
   el boton `Tutorial` de nomina y conservar `empresa_id`; explica el orden
   recomendado: parametros legales, configuracion, empleados, novedades,
   liquidacion, pagos/PILA, preparacion del lote DIAN y revision de acuses.
11. Pruebas: usar `Crear nomina demo Motel Calipso`, verificar empleados en varias
   sedes, liquidaciones por sede, PILA, pagos, desprendible y botones `Ver estado
   DIAN` / `Preparar lote DIAN`; abrir `Tutorial` y `Ayuda` desde nomina.

## Carrito: medios de pago y pagos combinados

1. La empresa habilita/deshabilita medios desde
   `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`.
2. Los flags se guardan en `estaciones_config.carrito_ui_global` para efectivo,
   tarjeta credito, tarjeta debito, transferencia Bre-B, Nequi y otra
   transferencia. `permitir_pago_mixto` gobierna el pago combinado.
3. El carrito muestra los mismos medios en abonos, selector principal, pago
   mixto y tarjeta `Detalle del pago`.
4. Si el cajero escribe valores en mas de un medio, el frontend arma
   `metodo_pago=mixto` y envia `pagos_mixtos` al endpoint de pago.
5. Backend normaliza los metodos y exige referencia minima de 4 caracteres para
   tarjetas y transferencias. Tambien valida que el metodo este habilitado por
   configuracion empresarial y por rol operativo.
6. Bre-B se registra como `transferencia_bre_b`. La confirmacion automatica real
   requiere webhook/API bancaria o proveedor de conciliacion que notifique la
   referencia recibida; mientras no exista esa integracion, el cajero debe
   verificar y registrar la referencia manualmente.

## Pagos Bre-B QR en Finanzas

1. Abrir `Administrar empresa > Finanzas y cumplimiento > Pagos Bre-B QR`.
2. Activar `Mostrar pago por QR en carrito` y `Habilitar medio Transferencia
   Bre-B`; ambos valores se guardan por `empresa_id` dentro de
   `empresa_estacion_prefs.estaciones_config.carrito_ui_global`.
3. Registrar una o varias cuentas receptoras con proveedor, tipo de llave,
   llave/cuenta, caja asignada y, cuando el banco lo entregue, payload oficial o
   plantilla QR con `{valor}`, `{referencia}`, `{empresa_id}`, `{carrito_id}` y
   `{carrito_codigo}`.
4. Para varias cajas simultaneas, usar una cuenta por caja o una referencia
   dinamica unica por venta. El cajero debe validar que la referencia recibida en
   el banco coincide con la referencia del carrito antes de cerrar la venta.
5. La pantalla lista ventas cerradas con `transferencia_bre_b`, abonos Bre-B y
   registros bancarios manuales de `empresa_finanzas_bancos_movimientos`.
6. `Registrar pago recibido` crea un movimiento bancario de ingreso con origen
   `breb_qr_manual`, hash idempotente y estado de conciliacion pendiente o
   conciliado; no crea venta ni reemplaza el cierre del carrito.
7. La conciliacion automatica solo debe activarse cuando exista webhook/API real
   de banco o proveedor; hasta entonces PCS conserva trazabilidad y conciliacion
   manual sin simular confirmaciones.

## Buzon de usuario, tareas y almacenamiento empresarial

1. El panel de `Administrar empresa` carga `/api/empresa/buzon?action=resumen` con el `empresa_id` activo y muestra campana, contador de no leidos, ultimos mensajes, chat y estado de almacenamiento.
2. El menu flotante global replica el contador de no leidos en su boton principal. Al abrirlo, la campana `Notificaciones` aparece como primera accion y despliega dentro del mismo menu el resumen del buzon; al hacer clic en una notificacion la marca como leida y navega a la pagina relacionada.
3. Un usuario puede enviar `Mensaje` o `Asignar tarea` a otro usuario registrado de la misma empresa. Las tareas quedan como mensajes tipo `tarea` con prioridad, fecha limite opcional y estado `pendiente`.
4. Si el mensaje incluye foto, archivo o audio grabado desde microfono, el backend guarda el binario en la carpeta de la empresa y registra metadata en `empresa_buzon_adjuntos`.
5. El destinatario finaliza una tarea desde su propio buzon; debe escribir descripcion de cierre y puede adjuntar evidencia. El backend valida que el mensaje pertenezca a su buzon antes de adjuntar o cerrar.
6. En traslados de bodega, el endpoint de inventario crea una notificacion de buzon para el responsable de la bodega destino o usuarios de inventario/administracion.
7. Super administrador controla la cuota global por empresa desde Configuracion avanzada > Almacenamiento: limite MB, porcentaje de alerta, maximo por archivo, bloqueo al superar cuota y limpieza de archivos antiguos del buzon.
