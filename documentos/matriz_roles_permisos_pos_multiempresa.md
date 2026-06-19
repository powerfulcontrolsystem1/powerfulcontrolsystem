2026-06-19: Nota de Empresas en super administrador
- `/super/api/empresas_estado` es exclusivo del panel super y queda envuelto en
  `WithSuperAuditoria` como `super_empresas_estado`; no concede permisos a
  usuarios empresariales ni acepta mutaciones.
- `web/super/empresas.html` es una vista de solo lectura para
  `super_administrador`, con filtros por nombre/NIT y estado de licencia.
- El boton `Ver` del Centro de mando abre la misma vista filtrada por licencia
  activa; no permite editar, crear, borrar ni reasignar licencias.

2026-06-18: Nota de configuracion interactiva e IA documental
- El asistente inicial de `panel.html` aplica configuracion solo desde contexto
  empresarial autenticado y endpoint `/api/empresa/configuracion_guiada` con
  `empresa_id`; no crea rol nuevo ni salta wrappers existentes.
- `agente_configuracion_de_empresa` se selecciona en el chat IA, pero sus
  acciones reales siguen limitadas por permisos del usuario, licencia,
  endpoints permitidos y confirmacion humana. `UI_CLICK` solo opera controles
  visibles/permitidos del frontend.
- Botones IA en Ingresos, Egresos, Compras y Productos reutilizan
  `soportes_compras_ia` y permisos existentes de cada modulo; la IA precarga
  datos para revision y no guarda movimientos/productos sin accion del usuario.

2026-06-18: Nota de agentes de mantenimiento IA super
- `/super/api/agentes_mantenimiento` queda protegido por `paginaPrincipalRequireSuperAdmin` y auditoria `super_agentes_mantenimiento`; solo `super_administrador` puede leer, configurar, ejecutar manualmente o activar el agente.
- No aplica `empresa_id` ni roles empresariales porque es una funcion global del panel super. El agente no concede permisos a empresas ni usuarios operativos.
- El correo de notificacion se valida como direccion de email; el envio usa la configuracion Gmail SMTP existente sin exponer contrasenas ni claves.
- La clasificacion con OpenAI registra consumo en las tablas IA existentes y no habilita SQL libre ni acciones automaticas sobre facturacion electronica.

2026-06-18: Nota de auditoria DIAN produccion PCS
- `facturacion_electronica` conserva sus wrappers existentes (`WithEmpresaFacturacionPermissions`) y no agrega permisos nuevos.
- Para declarar DIAN produccion estable se exige evidencia de acuse oficial o portal DIAN; en PCS quedaron `1PCS2` y `1PCS3` como `Aprobado con notificacion`.
- `Regla 90` no concede aceptacion automatica. Debe revisarse CUFE/TrackId, historial DIAN, cola de reintentos o portal antes de cerrar el documento.
- La auditoria operativa minima de un incidente DIAN debe incluir usuario/empresa, accion ejecutada, prefijo/folio, resultado DIAN, estado de cola, contador siguiente y si hubo cambio de configuracion, firma, rango o certificado.

2026-06-11: Nota de configuracion dedicada del rol cajero
- `linkConfiguracionRolCajero` queda en el menu de Configuracion bajo modulo
  `seguridad` y accion `U`; solo roles con administracion de configuracion
  empresarial deben verlo.
- La pagina no concede permisos nuevos al cajero. Es una consola para que el
  administrador edite, por `empresa_id`, la regla operativa del rol `cajero`,
  checks del carrito POS, control de estaciones y un perfil personalizado
  basado en el rol global.
- El rol global `cajero` no se modifica. Si se cambia nombre/descripcion, PCS
  crea o actualiza un rol personalizado de la empresa con `rol_base_id=cajero`.
- Los permisos efectivos del cajero siguen saliendo de la matriz existente,
  wrappers backend, licencia, estacion asignada y configuracion operativa.

2026-06-11: Nota de roles personalizados por empresa
- `/api/empresa/roles_de_usuario` mantiene `WithEmpresaSeguridadPermissions`, pero ahora `GET` lista roles globales y roles propios de la empresa, y `POST/PUT/DELETE` administra solo roles con `empresa_id` igual a la empresa activa.
- `admin_empresa` puede crear roles personalizados desde Administrar usuarios siempre que tenga permiso de seguridad efectivo; el cambio pasa por la evidencia trazable ya exigida para cambios de roles.
- Un rol personalizado no concede permisos nuevos por nombre libre: hereda el rol base global guardado en `rol_base_id` y el snapshot de permisos resuelve ese rol base para autorizar modulos/paginas.
- Impacto de matriz: no se agregan roles base globales nuevos; se agrega capacidad empresarial de crear alias/roles propios aislados por `empresa_id` bajo el modulo `seguridad`.

2026-06-11: Nota de lectura administrativa IA por empresa
- `linkChatIA` puede seguir disponible como ayuda operativa, pero la lectura
  amplia de base de datos y las respuestas administrativas directas solo se
  habilitan para `super_administrador`, `administrador_total` y `admin_empresa`.
- Para esos roles, el backend entrega datos reales filtrados por `empresa_id`,
  como conteos de `users`, sin exponer secretos ni permitir SQL libre del modelo.
- Roles como `cajero`, `vendedor`, `contador`, `inventario` o
  `responsable_bodega` conservan el chat operativo, pero no reciben contexto
  total de tablas por IA.
- Cualquier modificacion solicitada desde IA debe pasar por endpoints PCS,
  wrappers de permisos, validacion multiempresa y confirmacion cuando aplique.

2026-06-10: Nota de IA operativa por rol
- `linkChatIA` sigue bajo modulo `ventas` con accion de lectura (`R`), y las
  consultas del chat empresarial pasan por wrappers de ventas para permitir uso
  operativo sin abrir paginas administrativas.
- `cajero` puede usar el chat IA como API auxiliar del carrito: consultar el
  asistente, ejecutar `/api/empresa/ia_pedidos_estacion/ejecutar` para pedidos
  de estacion/mesa/habitacion y `/api/empresa/ia_radio/activar` para la emisora.
- `cajero` no recibe paginas de Productos, Nomina, Tarifas ni Seguridad por este
  cambio. Si intenta crear productos, modificar nomina o configurar tarifas, los
  wrappers de inventario, nomina o ventas y la matriz efectiva deciden el acceso.
- La IA solo propone acciones confirmables y no puede ejecutar `DELETE`, SQL
  libre ni endpoints fuera de la lista permitida del drawer.

2026-06-10: Nota de E-mail Corporativo
- Se agrega pagina `linkEmailCorporativo` bajo modulo `seguridad` con accion de lectura (`R`) para abrir el buzon corporativo por empresa.
- La configuracion del buzon sigue usando `Configuracion > Email corporativo` y requiere permisos de seguridad segun wrapper existente.
- El panel de empresa muestra notificaciones de email solo a roles administrativos/super administrativos detectados por contexto efectivo; usuarios operativos sin ese rol no ven el correo corporativo en el panel.
- El conteo de no leidos se consulta con `check_unread=1` desde `/api/empresa/email_corporativo`, siempre con `empresa_id` validado por backend y sin devolver credenciales.

2026-06-10: Nota de snapshot completo VPS
- `/super/api/vps_snapshots` es exclusivo de `super_administrador` y queda
  envuelto en `WithSuperAuditoria` como `super_vps_snapshots`.
- No recibe ni concede permisos por `empresa_id`: opera sobre infraestructura
  global de la VPS y por eso no aparece en la matriz empresarial.
- Las descargas solo sirven archivos dentro de `backup/vps_snapshots`; la ruta
  fisica no se expone en la lista del frontend.
- La configuracion de nube guarda ruta `rclone`, no credenciales OAuth, tokens,
  certificados ni claves privadas.

2026-06-10: Nota de retencion de empresas vencidas
- `/super/api/licencias/vencimiento_alertas` mantiene acceso exclusivo de
  `super_administrador` y ahora tambien administra `retencion_empresas`.
- La accion manual `retencion_run_now` y el worker solo procesan empresas no
  operativas, sin licencia base vigente y con preaviso registrado antes de
  cualquier eliminacion.
- El reporte de preavisos, errores y empresas eliminadas queda en
  `licencia_empresa_retencion_log` con `empresa_ref_id`, sin conceder permisos
  a administradores empresariales.

2026-06-09: Nota de IA empresarial activa sin avatar
- `linkChatIA` y `linkCentroIAEmpresarial` ya no se fuerzan ocultos por defecto;
  se muestran cuando rol, licencia y wrapper efectivo lo permiten.
- `linkRentaIA`, `linkSoportesComprasIA` y `linkSoportesComprasIAMenu` siguen
  ocultos hasta regla fina explicita por su impacto tributario/contable.
- El chat flotante queda activo por defecto en preproduccion, siempre como
  recuadro normal. Robot y secretaria se retiran de la experiencia visual.
- Las acciones propuestas por IA siguen requiriendo confirmacion y endpoints
  permitidos; la IA no concede permisos ni ejecuta SQL libre.

2026-06-18: Centro de mando super
- `/super/api/panel_control/reset` queda reservado a `super_administrador` y
  protegido con auditoria super.
- Acciones permitidas: `metricas` limpia `metrics`; `errores` limpia
  `super_errores_sistema`. No hay alcance empresarial ni permisos delegados.

2026-06-18: Retiro OCR y preparacion operativa
- Se retira del catalogo activo el modulo `ocr`, la pagina `linkOCR`, el wrapper
  `WithEmpresaOCRPermissions` y las rutas `/api/empresa/ocr` y
  `/super/api/config/ocr`.
- La captura de soportes de compras/gastos se conserva en
  `soportes_compras_ia` y debe operar con IA GPT-5.5 bajo los limites definidos
  por Super Administrador.
- Nomina e Impuestos no agregan permisos nuevos para la barra 0-100; usan los
  permisos existentes del modulo y solo muestran faltantes calculados con datos
  de la empresa activa.
- Facturacion electronica mantiene aislamiento por `empresa_id`; la anulacion
  nueva exige documento emitido y motivo operativo antes de crear nota credito
  total.

2026-06-08: Nota de OCR documental sin IA
- Se agrega modulo `ocr`, pagina `linkOCR` y wrapper `WithEmpresaOCRPermissions`.
- Matriz efectiva: lectura para roles con consulta general; procesar documentos requiere permiso de creacion/actualizacion segun la matriz empresarial y licencia activa.
- Licencias: fallback a `reportes` o `seguridad` para no bloquear empresas que ya tienen analitica/administracion, pero planes nuevos pueden habilitar `ocr` como modulo independiente.
- Seguridad: `/api/empresa/ocr` valida `empresa_id`, guarda archivos en carpeta empresarial y no aplica cambios automaticos sobre DIAN, inventario ni usuarios; las sugerencias OCR deben ser revisadas por el administrador.

2026-06-07: Nota de navegacion financiera sin duplicados
- El grupo `Finanzas y cumplimiento` del menu principal queda como entrada resumida: `Centro financiero y contable`, `Facturacion electronica` y `Reportes ejecutivos`.
- `Suite contador`, `NIIF`, `Creditos y cartera`, `Gestion de cobranza` e `Impuestos` se mantienen dentro del Centro financiero para no duplicar funciones visibles.
- No cambian permisos efectivos, wrappers ni paginas del catalogo; las rutas directas siguen protegidas para enlaces historicos y submenus internos.
- Actualizacion 2026-06-11: `linkCreditosTarjeta` queda como tarjeta interna del Centro financiero y usa `finanzas:C`; las subpaginas de creditos se publican bajo el grupo financiero universal para mantener la matriz sin grupos heredados.

2026-06-07: Nota de auditoria integral de modulos nuevos
- No se agregan permisos, endpoints ni wrappers nuevos.
- `operativo_modulos_resumen` queda dentro de `/api/empresa/reportes` y conserva `WithEmpresaReportesPermissions`, por lo que requiere `reportes:R` y `empresa_id` valido.
- `web/administrar_empresa/auditoria.html` solo amplia opciones visibles de filtro; consultar eventos sigue dependiendo del permiso/wrapper de auditoria existente.
- Los hubs sin tabla propia se muestran como `sin_tabla`; esto documenta alcance operativo y no concede acceso a datos, mutaciones, emision DIAN, pagos, IA ni documentos.

2026-06-07: Nota historica de IA oculta por defecto
- Regla reemplazada el 2026-06-09 para `linkChatIA` y
  `linkCentroIAEmpresarial`: ahora se muestran con permisos/licencia efectivos.
- La restriccion fina sigue vigente para `linkRentaIA`,
  `linkSoportesComprasIA` y `linkSoportesComprasIAMenu`.

2026-06-07: Nota de Suite contador
- Se agrega pagina `linkSuiteContador` bajo el modulo `finanzas` con accion de lectura (`R`).
- El hub no tiene endpoint nuevo ni tablas; consulta `/api/empresa/permisos_contexto` y enlaza modulos existentes conservando `empresa_id`.
- El rol `contador` puede ver la suite y accesos contables/fiscales principales. La escritura, aprobacion, emision DIAN, pagos, inventario y cambios de configuracion siguen sujetos a cada wrapper y matriz de permisos.
- Seguridad: la pagina solo navega y muestra estados de disponibilidad; no ejecuta mutaciones ni expone secretos.

2026-06-07: Nota de Modulo NIIF
- Se agrega pagina `linkNIIF` bajo el modulo `finanzas` con accion de lectura
  (`R`).
- La pagina no tiene endpoint propio ni tablas; lee el dashboard contable
  existente cuando el usuario tiene permiso y mantiene diagnostico local por
  navegador/empresa.
- El rol `contador` puede consultar NIIF, exportar su diagnostico local y abrir
  enlaces contables, pero no recibe escritura, aprobacion, DIAN, ventas, caja ni
  configuracion por este cambio.

2026-06-07: Nota de Centro IA empresarial
- Se agrega pagina `linkCentroIAEmpresarial` bajo el modulo `reportes` con accion de lectura (`R`).
- El endpoint `/api/empresa/ia_empresarial` queda detras de `WithEmpresaReportesPermissions` y solo consulta datos reales filtrados por `empresa_id`.
- Roles `contador` y `empresario` pueden ver el Centro IA empresarial como apoyo de analisis; no reciben permisos de escritura, aprobacion, emision de documentos, registro de pagos ni cambios de inventario.
- Las funciones IA registran consumo diario por empresa y devuelven borradores/recomendaciones revisables. Cualquier accion operativa final debe ejecutarse en su modulo autorizado correspondiente.

2026-06-06: Nota de Renta IA financiera
- Se agrega pagina `linkRentaIA` bajo modulo `finanzas` con accion de lectura (`R`).
- Roles con lectura financiera, incluido `contador`, pueden consultar el calculo de renta estimada.
- La accion IA registra consumo diario por empresa, pero no concede permisos de creacion, aprobacion, declaracion oficial ni edicion contable.
- Seguridad: `/api/empresa/finanzas/renta_ia` queda detras de `WithEmpresaFinanzasPermissions` y filtra todas las fuentes por `empresa_id`.

2026-06-06: Nota de modulo Bolsa empresarial
- Se agrega modulo `bolsa`, pagina `linkBolsa` y wrapper `WithEmpresaBolsaPermissions`.
- Matriz efectiva: lectura (`R`) para roles con acceso general de consulta; administracion no agrega acciones de escritura porque el endpoint es solo informativo.
- Licencias: fallback a `reportes` o `finanzas` para no bloquear empresas con planes actuales que ya tienen analitica financiera.
- Seguridad: `/api/empresa/bolsa` valida `empresa_id` mediante el wrapper empresarial, no guarda datos, no crea tablas y solo devuelve indicadores de mercado saneados.

2026-06-01: Nota de preconfiguracion solar y licencias por modulo
- 2026-06-04: se agrega modulo independiente `camaras`, pagina
  `linkCamaras`, wrapper `WithEmpresaCamarasPermissions` y fallback de licencia
  a `control_electrico` o `seguridad`. Los roles con lectura general pueden
  consultar; administracion/supervision pueden crear/actualizar; la eliminacion
  logica queda restringida a `admin_empresa`.
- Las preconfiguraciones de tipos de empresa incluyen `modulos.energia_solar` como modulo opcional apagado por defecto.
- El catalogo solar base contiene proveedores Victron VRM, SMA Sunny Portal, SolarEdge Monitoring y gateway local, baterias comunes y alertas recomendadas.
- El rol `tecnico_solar` se mantiene con `energia_solar:R` y pagina `linkEnergiaSolar`; no obtiene permisos de configuracion, ventas, caja, inventario ni facturacion.
- Licencias nuevas deben habilitar `energia_solar` como clave independiente cuando el plan incluya monitoreo solar; licencias antiguas pueden conservar fallback desde `control_electrico` o `seguridad` para no cortar empresas ya configuradas.

2026-05-31: Nota de catalogo base de roles empresariales comunes
- 2026-06-01: La asignacion de usuarios empresariales ya no limita el selector
  de roles al `tipo_empresa_id` de la empresa. `/api/empresa/roles_de_usuario`
  devuelve un catalogo global deduplicado por nombre/alias para todos los tipos
  de empresa, y `/api/empresa/usuarios` acepta cualquier rol activo del catalogo.
  La autorizacion efectiva sigue dependiendo de `rol_nombre` normalizado, matriz
  de permisos, licencia y `empresa_id`; no se concede acceso por editar la URL.
  La UI muestra cada rol con descripcion y enlace `Saber mas` a la ayuda.
- Las preconfiguraciones de tipos de empresa incluyen roles comunes para asignacion directa desde usuarios empresariales: `supervisor_sucursal`, `vendedor`, `recepcion`, `jefe_bodega`, `responsable_bodega`, `recursos_humanos` y `tecnico_solar`.
- `tecnico_solar`: `energia_solar:R`, pagina `linkEnergiaSolar`, sin permisos de configuracion, domotica, inventario, ventas ni reportes.
- `jefe_bodega`: `inventario:R/C/U/A` y `compras:R`; puede ver paginas de inventario, bodegas, categorias, recetas, historial y codigos de barras, pero no ventas, caja ni configuracion. No recibe `D` para eliminar inventario.
- `responsable_bodega`: `inventario:R/C/U/A` y `compras:R`; pensado para el usuario responsable de una bodega especifica. Puede operar productos, existencias, traslados y bodegas sin ventas, caja, configuracion ni `inventario:D`.
- `recursos_humanos`: `horarios_trabajadores:R/C/U`, `asistencia_empleados:R/C/U` y `nomina_sueldos:R/C/U`; sin ventas, caja ni permisos generales de seguridad.
- `vendedor` y `recepcion`: orientados a ventas/clientes con consulta de inventario, sin administracion global.
- Defensa backend: las restricciones de roles especializados se reaplican despues de licencia, vertical, empresa y acceso compartido.
- `cajero`: las variantes historicas `Caja`, `Caja principal` y `Caja turno`
  se normalizan como `cajero`. En `login_usuario.html` el menu queda limitado a
  `Venta directa`, `Estaciones`, `Corte de Caja` y `Buscar ventas y facturas`,
  aunque el rol conserve permisos operativos internos para cobrar, facturar,
  cerrar turno, consultar catalogo de inventario, crear/actualizar clientes
  desde el carrito y reimprimir o reenviar documentos ya generados. Las APIs
  auxiliares de carrito para productos, servicios, recetas, clientes, codigos de
  descuento, propinas y comisiones pueden ejecutarse sin mostrar las paginas
  administrativas de Productos o Clientes en el menu.

2026-06-11: Nota de busqueda de ventas y facturas para cajero
- `linkVentas` pasa a ser acceso operativo permitido para `cajero` con
  `ventas:R`, usando `web/administrar_empresa/ventas.html`.
- Alcance del cajero: consultar ventas/facturas, previsualizar, reimprimir,
  abrir facturas electronicas relacionadas y reenviar correo al cliente cuando
  el documento lo permita.
- Para evitar 403 en la consulta documental, `linkFacturasElectronicas` y
  `linkFacturacionElectronica` quedan permitidos como soporte interno del flujo,
  pero el frontend de cajero no los expone como botones de menu.
- No abre paginas administrativas de Productos, Clientes, Finanzas,
  Configuracion ni Reportes; los endpoints mantienen aislamiento por
  `empresa_id` y wrappers de ventas/facturacion.

2026-06-11: Nota de caja detectada por computador
- `web/login_usuario.html` identifica el computador con `localStorage.pcs_dispositivo_id`
  y entra automaticamente con la caja asignada a ese navegador/equipo cuando
  la asociacion existe para la empresa.
- La asignacion no concede permisos: solo selecciona `caja_codigo`,
  `caja_nombre` y `caja_descripcion` para el flujo de cajero ya autorizado.
- Ya no existe seleccion manual de caja en el login del cajero. Si un equipo
  necesita cambiar de caja, el administrador actualiza la asociacion desde
  Configuracion > Impresoras y caja.

2026-06-11: Nota de ingresos/egresos manuales para cajero
- El rol `cajero` no recibe finanzas completas por defecto. Puede registrar
  ingresos o egresos manuales solo si el administrador activa los checks por rol
  `permitir_ingresos_manuales` y/o `permitir_egresos_manuales` en Configuracion
  operativa de cobro.
- La excepcion de permisos del middleware aplica solo a
  `/api/empresa/finanzas/movimientos` con `POST` o `PUT` de movimientos
  manuales. Importacion bancaria, Bre-B, conciliacion, configuracion financiera,
  periodos y otros endpoints de finanzas siguen bloqueados por matriz/wrapper.
- Cuando el check esta activo, el menu puede mostrar `Ingresos`, `Egresos` y el
  acceso resumido de finanzas necesario para llegar a esas paginas, siempre con
  `empresa_id` validado y sin conceder `DELETE` financiero general.

2026-05-31: Nota de rol Servicio de limpieza para estaciones
- Se agrega el rol `servicio_limpieza` al catalogo base de roles empresariales y a las preconfiguraciones de tipos de empresa.
- Matriz efectiva: `ventas:R` solo para cargar el tablero de estaciones; no tiene `ventas:A`, `C/U/D`, caja, carrito, inventario, reportes ni configuracion.
- Visibilidad: `permisos_contexto` restringe el rol a `linkEstaciones`; el menu empresarial oculta el resto de paginas.
- Defensa backend: `carritos_compra` permite al rol solo `GET` del tablero y `carritos_compra/items` responde 403. El cambio de sucia a limpia se realiza por `/api/empresa/estacion_aseo?action=finalizar`, con registro de duracion y usuario.
- Reportes de aseo: el rol puede reportar aseo terminado, pero no obtiene permiso gerencial para consultar reportes historicos salvo que un rol superior lo autorice.

2026-05-31: Nota de rol empresario para resultados y reportes
- Se agrega el rol `empresario` al catalogo base de roles empresariales y a las preconfiguraciones de tipos de empresa.
- Matriz efectiva: `reportes:R` para consultar resultados y reportes ejecutivos; no tiene `C/U/D/A` ni acceso a ventas, estaciones, carritos, caja, creditos, inventario, finanzas, impuestos, usuarios o configuracion.
- Visibilidad: `permisos_contexto` restringe el rol a `linkReportes` y `linkReportesEjecutivos`; el submenu de reportes oculta `linkReportesTurnos` para no mezclar vista ejecutiva con caja/turnos.
- Defensa backend: las restricciones por rol se reaplican despues de overrides de empresa/compartidos para que un check de configuracion no amplie accidentalmente el alcance de `empresario`.

2026-05-31: Nota de rol operativo contador
- Se agrega el rol `contador` al catalogo base de roles empresariales y a las preconfiguraciones de tipos de empresa.
- Matriz efectiva: `finanzas:R` y `facturacion:R` para consultar centro financiero e impuestos; no tiene `C/U/D/A` ni acceso a ventas, estaciones, carritos, caja, creditos, inventario, usuarios o configuracion.
- Visibilidad: `permisos_contexto` restringe el rol a `linkFinanzas`, `linkFinanzasMain` y `linkImpuestos`; el submenu financiero tambien oculta accesos rapidos no autorizados.
- Defensa backend: las restricciones por rol se reaplican despues de overrides de empresa/compartidos para que un check de configuracion no amplie accidentalmente el alcance de `contador`.

2026-05-31: Nota de rol operativo portero
- Se agrega el rol `portero` al catalogo base de roles empresariales y a las preconfiguraciones de tipos de empresa.
- Matriz efectiva: `ventas:R` para consultar el tablero de estaciones y `ventas:A` exclusivamente para `action=activar_estacion`; no tiene `C/U/D`, inventario, finanzas, clientes ni facturacion.
- Visibilidad: `permisos_contexto` restringe el rol a `linkEstaciones`; el menu empresarial oculta panel, caja, corte, venta directa, carritos y configuracion.
- Defensa backend: `carritos_compra` permite al portero solo `GET` operativo y `PUT action=activar_estacion`; `carritos_compra/items` devuelve 403 aunque se intente desde consola o URL.

2026-05-28: Nota de auditoria especial super administrador
- `super_administrador.html > Acceso > Auditoria super` queda reservado para roles super y consulta `/super/api/auditoria?scope=super_panel`.
- El backend rechaza `scope=super_panel` para administradores que no sean roles de super panel, aunque intenten modificar la URL o llamar la API manualmente.
- Las APIs sensibles de configuracion super quedan auditadas con `WithSuperAuditoria`; contrasenas, claves, tokens y secretos se redactan antes de persistir metadata.

2026-05-28: Nota de auditoria global del selector
- `seleccionar_empresa.html > Auditoria` queda disponible para administrador principal y super administrador.
- `/super/api/auditoria` valida sesion administrativa y aplica alcance por `principal_email`; un administrador normal no consulta movimientos de otros principales.
- El super administrador puede ver la auditoria global cuando entra desde el panel super; si usa `scope=principal`, ve solo el alcance solicitado.
- `WithSuperAuditoria` no concede permisos nuevos: solo registra trazabilidad de endpoints super ya autorizados.

2026-05-27: Nota de filtrado de administradores en selector y super por invitacion
- `seleccionar_empresa.html > Administradores` usa `scope=principal`; debe listar solo administradores invitados por la cuenta autenticada, no la lista global del sistema.
- La vista global de administradores queda reservada al panel super cuando se accede sin `scope=principal`.
- Los nuevos roles `super_administrador` se crean por invitacion con token; no deben quedar activos sin aceptar correo y completar registro.

2026-05-27: Nota de delegacion de portafolio entre administradores
- Un administrador principal puede compartir su portafolio de empresas con un administrador ya confirmado sin cambiar `administradores.usuario_creador` ni la propiedad de empresas.
- La relacion vive en `admin_principal_delegaciones`; el acceso se evalua en backend junto al acceso directo por creador y al acceso compartido por empresa.
- El administrador invitado conserva sus empresas propias y ve adicionalmente las empresas compartidas como acceso delegado; no puede compartirlas como propietario.
- Revocar desde el listado del principal elimina solo la delegacion cuando la cuenta pertenece a otro administrador, evitando borrar o desactivar una cuenta ajena.

2026-05-27: Nota de administradores por administrador principal
- El enlace `Administradores` en `seleccionar_empresa.html` queda visible para el administrador principal normal y para roles super; no queda visible para administradores delegados.
- `/super/api/administradores` valida backend: un principal normal solo puede listar, crear, activar/desactivar o eliminar administradores dentro de su propio alcance `usuario_creador`.
- Crear administrador desde este modulo envia invitacion por correo; el invitado debe registrarse con `invitation_token` antes de iniciar sesion.
- Los administradores creados por el principal heredan acceso a las empresas del principal como administracion delegada. Esto no los vuelve propietarios: no pueden compartir la empresa ni ver administradores de otros principales.
- El super administrador conserva vista y gestion global; el control super mantiene su alcance definido por rutas super permitidas.

2026-05-27: Nota obligatoria de seguridad por endpoint multiempresa
- Antes de crear, modificar o revisar endpoints empresariales se debe aplicar `documentos/checklist_seguridad_endpoint_multiempresa.md`.
- La checklist exige validar sesion, `empresa_id`, permisos, licencia, SQL con aislamiento, IDs secundarios, auditoria, errores saneados y pruebas negativas de intento de cruce entre empresas.
- Ocultar botones o paginas en frontend no reemplaza la validacion backend ni el wrapper de permisos correspondiente.

2026-05-28: Nota de Energia solar
- Se agrega clave de modulo independiente `energia_solar`, pagina `linkEnergiaSolar` y wrapper `WithEmpresaEnergiaSolarPermissions`.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa`, `supervisor_sucursal`, `contabilidad` o `auditor` segun la matriz de analisis/control; eliminar queda restringido a `admin_empresa`.
- Compatibilidad de licencias: si una licencia antigua habilita `control_electrico` o `seguridad`, el modulo solar queda disponible como fallback para no cortar operaciones existentes; las licencias nuevas pueden habilitar explicitamente `energia_solar`.
- El endpoint `/api/empresa/energia_solar` exige `empresa_id`, valida alcance por empresa y las tablas `empresa_energia_solar_*` siempre filtran por empresa. Las API keys reales deben quedar como referencias `env:*`.

2026-05-27: Nota de informacion editable de modulos del index
- `web/super/informacion_de_modulos.html` queda disponible solo dentro del panel de Super Administrador.
- `/super/api/informacion_de_modulos` exige sesion y rol `super_administrador` mediante la misma validacion usada por pagina principal.
- `/api/public/informacion_de_modulos` es lectura publica del portal y solo devuelve contenido editorial saneado; no concede permisos empresariales ni acceso a datos privados.

2026-05-25: Nota operativa para `creditos` y `finanzas`
- `Creditos y cartera` se divide en subpaginas internas: panel, nuevo credito, cartera, morosidad, riesgo/limites, operaciones, aprobaciones y estado de cuenta.
- Los nuevos links `linkCreditosPanelMenu`, `linkCreditosCrearMenu`, `linkCreditosCarteraMenu`, `linkCreditosMorosidadMenu`, `linkCreditosLimitesMenu`, `linkCreditosOperacionesMenu`, `linkCreditosAprobacionesMenu` y `linkCreditosEstadoMenu` conservan modulo `finanzas` con accion `C`, igual que `linkCreditos`.
- No cambian wrappers, tablas, datos ni aislamiento por `empresa_id`; el backend sigue resolviendo `/api/empresa/creditos` bajo permisos financieros.

2026-05-25: Nota de licencia del sistema descargable
- `linkLicenciaSistema` se agrega al grupo `Administracion y configuracion` con modulo `seguridad` y accion `R`.
- La pagina `web/administrar_empresa/licencia_sistema.html` solo consulta el contexto efectivo de permisos/licencia de la empresa y genera una descarga local en el navegador; no crea endpoint nuevo ni concede acceso adicional a datos operativos.
- El documento descargable se genera por `empresa_id` y conserva el aislamiento multiempresa.

2026-05-25: Nota de orden de menu financiero
- `Finanzas y cumplimiento` se reubica debajo de `Inventario y compras` en el menu principal de `Administrar empresa`.
- El cambio es solo de navegacion visual; conserva los mismos `PaginaClave`, modulos, acciones y wrappers existentes.

2026-05-25: Nota de navegacion financiera y paginas huerfanas
- `linkCreditos` y `linkCobranza` se conservan como paginas protegidas y accesos internos del Centro financiero; ya no se duplican como botones directos del grupo principal `Finanzas y cumplimiento`.
- `linkChatIA`, `linkConfiguracionGuiada` y `linkConfiguracionIntegraciones` se conectan desde menus visibles para paginas empresariales ya existentes. No se crean wrappers nuevos ni se relaja el aislamiento por `empresa_id`.
- Las paginas legacy o generadas por accion, como `frecuencia_fp.html` y `soporte_remoto_view.html`, no se promocionan a menu principal porque son alias/visores de flujo, no modulos de entrada operativa.

2026-05-20: Nota de nombres configurables de estaciones
- La edicion de singular/plural del recurso operativo vive en `web/administrar_empresa/configuracion_de_estaciones.html` y reutiliza el permiso existente de configuracion de estaciones.
- No se agregan permisos, wrappers ni modulos de licencia: la preferencia queda por `empresa_id` en `empresa_estacion_prefs.estaciones_config`.
- El helper `web/js/estaciones_labels.js` solo adapta textos visibles; no concede acceso adicional a estaciones, carritos ni configuracion.

2026-05-20: Nota de datáfonos POS multiempresa
- `/api/empresa/datafonos` queda protegido con `WithEmpresaVentasPermissions`, igual que carritos, porque puede iniciar cobros y aplicar pagos al POS.
- No se crea permiso independiente en esta fase: la configuracion tecnica del terminal exige acceso empresarial autenticado y el mismo alcance de ventas/caja del usuario.
- La aplicacion al carrito valida `empresa_id`, caja abierta del usuario y referencia/monto confirmados por el proveedor antes de cerrar la venta.
- Las credenciales de proveedor no son visibles para roles operativos; se guardan solo referencias `env:*` y no secretos reales.

2026-05-19: Nota de Docker VPS portable
- `web/super/docker_portabilidad.html` queda disponible solo para `super_administrador` dentro del grupo Plataforma del panel super.
- `/super/api/docker_portabilidad` exige sesion y rol `super_administrador` mediante `paginaPrincipalRequireSuperAdmin`; no se agrega permiso empresarial ni acceso para `control_super_administrador`.
- La descarga es operativa y tecnica: no concede acceso a datos empresariales, no expone secretos y no sustituye los permisos de backups ni de PostgreSQL.

2026-05-20: Nota de catalogo DIAN Colombia
- `web/administrar_empresa/facturacion_electronica.html` agrega la tarjeta de documentos electronicos DIAN Colombia dentro de la pagina existente del modulo `facturacion_electronica`.
- El endpoint usa la ruta ya protegida `/api/empresa/facturacion_electronica?action=documentos_dian_colombia`, por lo que hereda `WithEmpresaFacturacionPermissions` y no crea permiso, rol ni licencia nuevos.
- La lista de obligaciones del contador es informativa/configurable por empresa; no concede acceso a declaraciones o exogena fuera de los modulos contables que ya correspondan.

2026-05-19: Nota de Facturacion electronica Ecuador
- La pagina `web/administrar_empresa/facturacion_electronica_ecuador.html` y el endpoint `/api/empresa/facturacion_electronica/ecuador` quedan bajo el modulo independiente `facturacion_ecuador`.
- La licencia puede habilitar o deshabilitar Ecuador/SRI sin activar DIAN Colombia ni Panama/DGI; el permiso de pagina es `linkFacturacionEcuador` con accion crear/configurar.
- El submenu `Facturacion electronica` sigue como contenedor. Sus paginas internas se muestran por pais detectado automaticamente: Ecuador habilita `Ecuador / SRI` si el rol y la licencia lo permiten.
- El endpoint de deteccion de pais no concede emision ni configuracion: solo permite a usuarios autenticados de la empresa decidir que paginas del submenu deben aparecer.

2026-05-19: Nota de clientes desde carrito
- El boton `Clientes` del carrito no agrega permisos nuevos: registra clientes usando el endpoint existente `/api/empresa/clientes` y asigna el resultado al carrito activo.
- La regla `cliente_obligatorio_pago` es configuracion de carrito, no permiso nuevo. Solo restringe el cierre de pago cuando el carrito no tiene `cliente_id`.
- La validacion conserva aislamiento por `empresa_id`; un usuario solo puede crear/asignar clientes dentro de la empresa efectiva y bajo los permisos existentes de carrito/clientes.

2026-05-19: Nota de Facturacion electronica Panama
- La pagina `web/administrar_empresa/facturacion_electronica_panama.html` y el endpoint `/api/empresa/facturacion_electronica/panama` quedan bajo el modulo independiente `facturacion_panama`.
- La licencia puede habilitar o deshabilitar Panama/DGI sin activar DIAN Colombia; el permiso de pagina es `linkFacturacionPanama` con accion crear/configurar.
- El submenu `Facturacion electronica` sigue como contenedor. Sus paginas internas se muestran por pais detectado automaticamente: Colombia habilita DIAN, pruebas DIAN y proveedores de firma; Panama habilita DGI/SFEP si la licencia y el rol lo permiten.
- El endpoint de deteccion de pais no concede emision ni configuracion: solo permite a usuarios autenticados de la empresa decidir que paginas del submenu deben aparecer.

2026-05-19: Nota de caja/turno independiente por usuario
- La independencia de cajas simultaneas se resuelve por datos (`empresa_id`, `usuario_creador`, `cierre_caja_id`) y no requiere nuevos permisos.
- Los usuarios creados por un administrador de empresa mantienen los mismos permisos de `finanzas`, `corte_caja`, `carritos` y `ventas`, pero sus cajas abiertas, abonos, pagos y movimientos se validan contra el usuario autenticado.
- Los roles administrativos conservan sus permisos de supervision y reportes historicos; la operacion diaria del carrito lista solo cajas abiertas del usuario actual.
- El estado de estaciones no se separa por usuario: sigue siendo comun por empresa para que varias cajas vean la misma ocupacion/disponibilidad.

2026-05-18: Nota de Reportes de turnos
- Se agrega `linkReportesTurnos` al catalogo de paginas empresariales bajo modulo `reportes` y accion lectura.
- La pagina `web/administrar_empresa/reportes_turnos.html` permite consultar, imprimir, compartir, exportar y enviar por email reportes historicos de turnos ya cerrados.
- `/api/empresa/corte_caja?action=turnos|turno_reporte|turno_export|turno_email` se mapea a `linkReportesTurnos`; el cierre operativo, corte automatico y guardado del turno siguen bajo `linkCorteCaja`.
- El aislamiento por `empresa_id` se conserva y la reconstruccion del reporte usa `cierre_caja_id` de la misma empresa.

2026-05-18: Nota de configuracion empresarial por paginas
- Las nuevas claves `linkConfiguracionIdentidadVisual`, `linkConfiguracionCobroOperativo`, `linkConfiguracionReporteCorte`, `linkConfiguracionBackupsPasarelas` y `linkConfiguracionPasarelasPago` quedan en el catalogo de paginas empresariales bajo modulo `seguridad` y accion `actualizar`.
- `linkConfiguracionMain` representa Productos y pedidos; `linkConfiguracionAvanzada` representa Formato monetario.
- La separacion es de navegacion y UX: los formularios siguen usando los mismos endpoints protegidos por `WithEmpresaSeguridadPermissions` y mantienen aislamiento por `empresa_id`.

2026-05-18: Nota de guardado de formato monetario
- `/api/empresa/configuracion_avanzada` mantiene el wrapper `WithEmpresaSeguridadPermissions` y no cambia permisos efectivos.
- La correccion solo normaliza persistencia de `empresa_configuracion_avanzada` y etiquetas universales del catalogo de permisos.
- El aislamiento por `empresa_id` se conserva en lectura y guardado de moneda, sistema numerico, decimales y cantidad de decimales.

2026-05-18: Nota de boton Corte de Caja en Operacion y ventas
- `web/administrar_empresa.html` muestra el enlace `linkCorteCaja` bajo `Operacion y ventas`, inmediatamente despues de `linkEstaciones`, con texto `Corte de Caja`.
- No se crea ruta ni endpoint nuevo: `web/js/administrar_empresa.js` lo conserva visible como acceso operativo directo y el backend sigue aplicando las validaciones de sesion/empresa del flujo de corte.
- La ruta abierta es la misma pantalla de caja usada desde estaciones.

2026-05-18: Nota de reporte de turno en Caja
- El boton `Ver reporte de mi turno` vive dentro de `web/administrar_empresa/corte_de_caja.html`, bajo los permisos existentes de finanzas/corte de caja.
- No crea pagina, permiso ni accion nueva de matriz; el endpoint reutiliza `/api/empresa/corte_caja` y filtra por `empresa_id` y usuario autenticado.
- La accion solo consulta/imprime; no guarda ni cierra turno.

2026-05-17: Nota de comunicaciones super unificadas
- El modulo `Comunicaciones` del Super Administrador agrupa mantenimiento, correos masivos, alertas y configuraciones de mensajeria.
- La agrupacion no agrega permisos ni wrappers nuevos; conserva las mismas paginas y endpoints super ya permitidos.
- Las empresas no reciben permisos nuevos por este cambio de navegacion.

2026-05-17: Nota de venta directa con carrito de estaciones
- `linkVentaDirecta` mantiene el mismo permiso base de ventas (`crear`) y el mismo wrapper de carritos.
- El cambio solo hace que `modo=venta_directa` use la vista enfocada del carrito de estaciones; no concede acceso a estaciones ni a configuracion.
- El carrito canonico `VENTA-DIRECTA-{empresa_id}-0` conserva aislamiento por `empresa_id`.

2026-05-17: Nota de mantenimiento super principal
- `Mantenimiento sistema` queda como modulo del menu principal super, no como configuracion empresarial.
- `action=limpiar_viejos` solo opera sobre configuracion global super de avisos y no concede permisos empresariales nuevos.
- La limpieza borra avisos desactivados o vencidos y mantiene separado `mantenimiento_activo`.

2026-05-17: Nota de configuracion super por paginas
- Las paginas bajo `web/super/configuracion/` pertenecen solo al shell de Super Administrador.
- No agregan permisos empresariales ni wrappers nuevos; reutilizan los endpoints super existentes y la validacion de sesion/rol del panel super.
- El rol `control_super_administrador` mantiene su lista reducida; estas paginas no se agregan a ese alcance limitado.

2026-05-17: Nota de avisos de mantenimiento super
- La tabla de avisos programados vive en `web/super/mantenimiento_sistema.html`, accesible desde el submenu principal super, y solo se opera desde el panel super.
- `/super/api/config/mantenimiento?action=desactivar|eliminar` no concede permisos empresariales nuevos y no cambia `mantenimiento_activo` salvo que el super edite explicitamente el switch de bloqueo real.
- `/api/empresa/mantenimiento_programado` sigue siendo de lectura para empresas autenticadas y solo expone el aviso activo visible.

2026-05-17: Nota de Panel primero en Administrar empresa
- Se retira `linkInicio` del catalogo de paginas porque el grupo lateral `Inicio` ya no se usa.
- `linkPanelEmpresa` queda como pagina siempre visible y primer boton directo del menu empresarial.
- No se agregan acciones, wrappers, rutas, endpoints ni permisos efectivos.

2026-05-17: Nota de default de carrito por tipo de empresa
- El preset simplificado del carrito no crea permisos nuevos ni cambia wrappers.
- Solo los roles con acceso vigente a configuracion del carrito/estaciones pueden modificar `estaciones_config.carrito_ui_global`.
- Las tarjetas avanzadas apagadas por defecto (`Cobro y estados`, descuentos, propina, comision, desglose y lavador) pueden reactivarse por empresa si el rol/licencia ya permite administrar esa configuracion.

2026-05-17: Nota de inicio de Administrar empresa
- El shell `web/administrar_empresa.html` carga `linkPanelEmpresa` como pagina inicial del iframe al entrar.
- Este comportamiento no concede permisos nuevos: `linkPanelEmpresa` sigue siendo una pagina siempre visible del contexto empresarial y las demas paginas mantienen sus controles por rol/licencia/modulo.

2026-05-17: Nota de navegacion para Inventario y compras
- `web/administrar_empresa.html` ubica el grupo `Inventario y compras` inmediatamente debajo de `Operacion y ventas`.
- No cambian claves de permisos, modulos, acciones ni rutas; solo el orden visual del submenu empresarial.

2026-05-17: Nota de navegacion para Operacion y ventas
- El grupo `Operacion y ventas` de `web/administrar_empresa.html` queda reservado a los accesos de operacion diaria directa: `linkVentaDirecta` y `linkEstaciones`; desde 2026-06-11 tambien incluye `linkVentas` como consulta operativa de ventas/facturas para reimpresion y reenvio.
- `linkVentas` se conserva en el catalogo bajo `Permisos base de ventas` y se muestra en el menu operativo cuando el rol tiene alcance de caja.
- `linkVentaPublica`, `linkRedSocialComercial`, `linkCodigosDescuento` y `linkChatTareas` se agrupan administrativamente como `Canales digitales y colaboracion`.
- `linkReservasHotel` se agrupa como `Plantillas de negocio` y se muestra en `Soluciones por negocio`.
- Impacto de matriz: no cambian acciones, wrappers, rutas, endpoints ni permisos efectivos; solo cambia la ubicacion visual y el grupo de catalogo.

2026-05-15: Nota de navegacion para menu empresarial
- `web/administrar_empresa.html` mueve el grupo `Operacion y ventas` al inicio del menu para que `Venta directa` sea el primer acceso real de Administrar Empresa y `Estaciones` el segundo.
- `Carritos` no aparece en el menu principal operativo y permanece disponible en `web/administrar_empresa/configuracion_menu.html` dentro de `Ventas y cobro`.
- Impacto de matriz: no cambia wrappers, acciones ni endpoints; `linkCarritoCompras` se agrupa administrativamente como `Configuracion - Ventas y cobro` en el catalogo de permisos.

2026-05-15: Empresas compartidas con alcance por rol y modulos
- Al compartir empresa, el propietario define alcance `Solo ver`, `Acceso total` o `Solo ciertos modulos`; el alcance no sustituye rol/licencia, solo lo restringe.
- `Solo ver` conserva lectura y desactiva crear, actualizar, eliminar y aprobar para el administrador invitado.
- `Solo ciertos modulos` apaga acciones y paginas de modulos no seleccionados; los modulos seleccionados siguen sujetos a licencia, vertical, rol efectivo y politicas finas de empresa.
- `Acceso total` mantiene el comportamiento anterior dentro de los permisos efectivos, sin convertir al invitado en propietario ni habilitar eliminacion estructural de la empresa.
- La revocacion o "dejar de compartir" invalida cache de acceso/permisos para que el retiro sea inmediato.

2026-05-15: Licencia de prueba unica por empresa
- La prueba de 15 dias y las activaciones gratis/sin pago quedan limitadas a una sola activacion activa por `empresa_id`; no depende del rol del usuario sino de trazabilidad en `licencias_activaciones_gratis`.
- El codigo de asesor puede enviarse como `asesor_id` o `codigo_asesor` en la prueba de 15 dias y en activaciones sin pago; debe existir, estar aceptado y no estar inactivo.
- Registrar un asesor en la licencia de prueba no concede permisos, roles ni acceso a empresas ajenas; solo deja trazabilidad comercial y posible asociacion de comision.
- El checkout y los endpoints conservan el aislamiento por `empresa_id` y no agregan permisos empresariales nuevos.

2026-05-15: Reportes unificados
- El submenu principal del modulo `reportes` queda en un acceso: `Centro de reportes` (`linkReportesEjecutivos`). La pagina dedicada `reportes_ia_chat.html` y el permiso de pagina `linkReportesIAChat` se retiran.
- Se retiran del catalogo de permisos empresarial las entradas antiguas `linkReportesMain`, `linkReportesVentas`, `linkReportesInventario`, `linkReportesFinanzas`, `linkReportesImpuestos` y `linkGraficosEstadisticas`.
- Las vistas antiguas `reportes.html`, `reportes_inventario.html`, `reportes_finanzas.html` y `graficos_estadisticas.html` dejan de existir; sus consultas quedan consolidadas en el catalogo con vista previa de `reportes_ejecutivos.html`.
- Los datasets exportables continuan bajo `/api/empresa/reportes` con `WithEmpresaReportesPermissions` y aislamiento por `empresa_id`.

2026-05-13: Reportes ejecutivos profesionales
- `web/administrar_empresa/reportes_ejecutivos.html` y `reportes_menu.html` permanecen dentro del modulo `reportes`; las vistas separadas de inventario/finanzas fueron absorbidas por el centro unico el 2026-05-14.
- No agrega permisos ni wrappers nuevos: los datos y exportaciones siguen usando `/api/empresa/reportes` bajo `WithEmpresaReportesPermissions`.
- La simplificacion es de UX: menos botones en el submenu principal y datasets especializados dentro de la suite, conservando aislamiento por `empresa_id`.

2026-05-13: Facturacion electronica separa pruebas DIAN
- `web/administrar_empresa/facturacion_electronica_pruebas_dian.html` queda dentro del mismo submenu y alcance del modulo `facturacion_electronica`.
- No agrega permisos ni wrappers nuevos: la pagina usa los endpoints existentes de facturacion electronica bajo `WithEmpresaFacturacionPermissions`.
- La separacion es de UX y operacion: configuracion empresarial en la pagina principal; diagnostico, pruebas DIAN y emision documental manual en subpagina dedicada.

2026-05-13: Control de aseo en estaciones
- `users.control_aseo_estaciones` habilita o deshabilita por usuario operativo la accion de reportar aseo terminado sobre una estacion sucia.
- `/api/empresa/estacion_aseo?action=finalizar` usa `WithEmpresaSelfServicePermissions` y ademas exige usuario con control de aseo activo o rol administrativo/supervision; el reporte gerencial queda restringido a roles de administracion, supervision, auditoria o reportes.
- El flujo no cambia permisos de carritos ni de configuracion de estaciones; conserva aislamiento por `empresa_id` y atribuye cada cierre al usuario autenticado.

2026-05-13: Proveedores de firma digital en facturacion electronica
- `web/administrar_empresa/proveedores_firma_digital.html` queda dentro del submenu empresarial de `Facturacion electronica`.
- No agrega permisos, roles ni licencias: hereda el acceso vigente al modulo `facturacion_electronica` y solo publica informacion/enlaces externos para adquirir certificados.
- La compra externa no modifica credenciales DIAN, secretos ni configuracion fiscal dentro del sistema.

2026-05-13: Juegos y records publicos
- `/Juegos/menu_juegos.html`, `/Juegos/*` y `/api/public/juegos/records` se mantienen como superficie publica del portal. El endpoint de records solo guarda/lee `juego`, `nombre_jugador`, `empresa_id`, `puntaje` y `nivel`; no otorga acceso a modulos privados ni modifica permisos de `super_administrador`, `administrador` o usuarios de empresa.
- Cuando existe `empresa_id` en sesion/localStorage/URL, el record se etiqueta para trazabilidad, pero el juego sigue sin habilitar acciones empresariales protegidas.

2026-05-13: Shell empresarial exige sesion administrativa valida
- `web/js/administrar_empresa.js` valida `/me` antes de montar el iframe del panel empresarial. Si `/me` responde `401/403`, redirige a `login.html` y no deja visible un shell operativo sin sesion.
- `web/administrar_empresa/carrito_de_compras.html`, `administrar_clientes.html` y `bodega.html` diferencian `401` y `403`: el primero reenvia al login administrativo por sesion expirada; el segundo informa que el rol no tiene permiso para usar ese modulo.
- No se agregan permisos, wrappers ni roles nuevos; solo se alinea la UX del frontend con el control de sesion que ya aplicaba el backend.

2026-05-13: Auditoria y compartido documental del nucleo
- `auditoria` no agrega permisos nuevos, pero sus filtros visibles y contexto IA ahora cubren modulos recientes: `carritos`, `venta_publica`, `crm_unificado`, `reportes`, `backups`, `documentos_onlyoffice`, `tickets_ayuda`, `mantenimiento_programado`, `propinas`, `comisiones`, plantillas operativos y `control_electrico`.
- Las opciones de compartir por WhatsApp/correo en ventas, facturas, reportes, ingresos, egresos y menu flotante no conceden acceso nuevo: comparten texto/enlace de la pantalla o documento activo y el receptor sigue necesitando permisos/sesion cuando el recurso no es publico.
- Los backups automaticos locales siguen bajo `linkBackups`/modulo `backups`; la descarga al equipo depende de sesion activa y `empresa_id`, y no sustituye los permisos de snapshot/restauracion.

2026-05-13: Vinculo de propinas y comisiones con usuarios creados
- `propinas` y `comisiones por servicio` no agregan permisos nuevos; siguen bajo sus paginas/API empresariales existentes y el `empresa_id` efectivo.
- Los movimientos ahora guardan `users.id` de la misma empresa (`usuario_origen_id`, `usuario_asignado_id`, `usuario_lavador_id`) ademas de la etiqueta texto historica.
- Los reportes y filtros pueden resolver por email, nombre, documento o id de usuario, pero no conceden acceso a usuarios de otra empresa.

2026-05-13: Reubicacion de Backup profesional
- El acceso `linkBackups` se muestra junto a `Configuracion` dentro del grupo `Administracion` de `administrar_empresa.html`.
- No cambia la matriz efectiva: `/api/empresa/backups` sigue bajo `WithEmpresaSeguridadPermissions` y las acciones de restauracion/depuracion mantienen aprobacion.

2026-05-13: Login global de usuarios operativos
- `login_usuario.html` es la entrada unica para usuarios de todas las empresas; no requiere subdominio empresarial ni `empresa_id` visible para iniciar sesion.
- El backend resuelve la cuenta por email y clave; si un correo esta en varias empresas y no puede resolverse de forma unica, no concede acceso a una empresa al azar.
- El primer ingreso sigue cerrado a invitacion: `token_invitacion`, email, documento y contrato vigente se validan antes de crear sesion.
- No se agregan roles ni privilegios nuevos; tras autenticar, el panel empresarial sigue filtrando por `/api/empresa/permisos_contexto` y `empresa_id`.

2026-05-12: Retiro de Nextcloud
- `nextcloud` deja de ser modulo licenciado o pagina de empresa.
- Se retiran `linkNextcloud`, `WithEmpresaNextcloudPermissions`, `/api/empresa/nextcloud` y `/super/api/config/nextcloud`.
- Las licencias nuevas no deben incluir `nextcloud`; `documentos_onlyoffice`, `gestion_documental`, `backups` y soporte remoto conservan sus permisos independientes.
- La cuota de GB que antes se asociaba a almacenamiento se interpreta ahora como `db_max_gb`, limite administrativo de tamano maximo de base de datos por empresa.

2026-05-12: Recuperacion publica de invitacion operativa
- `POST /api/empresa/usuarios/recuperar_invitacion` es publico solo para solicitar reenvio de invitacion; no inicia sesion, no devuelve datos del usuario y no revela si el correo existe.
- El reenvio se ejecuta unicamente para usuarios activos sin contrasena configurada; usuarios inexistentes, inactivos o ya registrados reciben la misma respuesta enmascarada.
- La accion rota `email_confirm_token`/`email_confirm_expira`; el acceso real sigue exigiendo completar invitacion y luego cargar permisos por `/api/empresa/permisos_contexto`.

2026-05-12: Registro operativo por invitacion y permisos al primer ingreso
- `login_usuario.html` no concede alta publica: el usuario operativo solo crea contrasena desde una invitacion enviada por un administrador autorizado de la empresa.
- El token de invitacion vive en `users.email_confirm_token`, vence con `email_confirm_expira` y se consume al guardar el primer password; el documento y el contrato vigente siguen siendo controles adicionales.
- Al completar invitacion, `createEmpresaUsuarioSessionAndRespond` registra/acopla el correo al rol asignado, crea sesion y devuelve `redirect_url=/administrar_empresa.html?id={empresa_id}`.
- La visibilidad final en el panel empresarial no depende de texto del login: `web/js/administrar_empresa.js` consulta `/api/empresa/permisos_contexto` y aplica rol efectivo, licencia, paginas y reglas finas por `empresa_id`.
- No se agregan roles ni privilegios nuevos; solo se endurece el canal de incorporacion de usuarios operativos.

2026-05-12: Nota operativa para tema en login de usuarios
- `web/login_usuario.html` y `web/menu.js` priorizan la cookie visible `pcs_theme` sobre `localStorage` para pintar el modo claro/oscuro antes de autenticar.
- La cookie `pcs_theme` solo guarda preferencia visual; no reemplaza ni expone la cookie de sesion `HttpOnly`, no concede acceso y no altera wrappers ni permisos.
- El comportamiento esperado es que el login operativo cargue con la apariencia del ultimo usuario que inicio sesion o cambio tema en ese navegador.

2026-05-12: Nota operativa para tickets de ayuda empresariales profesionalizados
- `GET /api/empresa/tickets_ayuda?id=...` y comentarios empresariales validan que el ticket pertenezca al `empresa_id` activo; si no pertenece, responde como no encontrado.
- Las notas internas (`interno=1`) del super administrador no se devuelven al usuario empresarial; solo la bandeja super puede verlas.
- El menu flotante puede enviar contexto tecnico seguro de la pantalla activa, pero no cookies, localStorage, tokens, claves ni secretos.
- La administracion global de estado, prioridad, asignacion y cierre sigue reservada a `super_administrador`.

2026-05-12: Nota operativa para mesa central de tickets de ayuda
- `POST/GET /api/empresa/tickets_ayuda` queda disponible para usuarios autenticados con alcance validado sobre la empresa solicitada mediante `WithEmpresaSelfServicePermissions`; no concede permisos de administracion central ni acceso a tickets de otras empresas.
- `web/super/tickets_ayuda.html` y `/super/api/tickets_ayuda` quedan reservados a `super_administrador` por `paginaPrincipalRequireSuperAdmin`; `control_super_administrador` no recibe este modulo en su navegacion limitada.
- El menu flotante solo crea tickets para la empresa activa detectada; la bandeja de respuesta, cierre, prioridad y asignacion se administra exclusivamente desde super administrador.

2026-05-12: Nota operativa para ayuda de CRM unificado
- El boton `Ayuda` agregado al menu de `crm_unificado` abre una pestana interna de `web/administrar_empresa/crm_comercial.html`; no crea pagina publica, endpoint nuevo ni permiso adicional.
- La ayuda esta dentro del alcance ya protegido por `linkCRMComercial` y `crm_unificado`, y primero define CRM antes de explicar el tablero, leads, seguimientos, cotizaciones, forecast, metas y embudo.
- La regla operativa queda explicita: CRM complementa clientes, ventas y facturacion, pero no duplica esos nucleos ni cambia sus permisos.

2026-05-12: Nota operativa para visibilidad de licencias
- 2026-05-12: Explorador de Archivos super
- `web/super/explorador_archivos.html` y `GET /super/api/explorador_archivos` quedan reservados a `super_administrador`; no hay acceso para `control_super_administrador`, roles empresariales ni usuarios operativos.
- La operacion es de solo lectura sobre metadata de archivos/carpetas y no agrega permisos empresariales, wrappers `/api/empresa/*`, tablas, acciones de descarga, edicion, subida o borrado.

- La pagina `web/super/licencias.html` permite a `super_administrador` mostrar u ocultar licencias para clientes usando `licencias.activo`.
- Ocultar una licencia no borra historico ni permisos ya documentados; la retira del catalogo comercial y bloquea checkout publico nuevo.
- Los flujos publicos de pago o activacion (`checkout_summary`, Wompi, Nequi, Epayco y activacion sin pago) rechazan licencias ocultas antes de iniciar la compra.

2026-05-12: Nota operativa para adaptacion profesional por tipo de empresa
- `linkPlantillasIntegracion` se muestra como `Adaptacion por tipo` en Configuracion empresarial y conserva el permiso `seguridad:R`; ver la matriz no concede operacion sobre plantillas.
- `/api/empresa/permisos_contexto` sigue siendo la fuente efectiva de rol, licencia, reglas finas y `vertical_scope`; la pantalla solo lo usa para mostrar el perfil activo y no para saltarse wrappers de API.
- `/api/*/plantillas_integracion/catalogo` ahora cubre exactamente 30 plantillas canonicos como contrato unico y publica `professional_ready`/`readiness_score` para auditar que cada vertical visible tenga nucleo, permisos, flujo y reportes declarados.
- Los alias `consultorio_odontologico` y `taxi` se fusionan en `odontologia` y `taxi_system`; `turnos_atencion`/`turnos` son soporte transversal y no amplian permisos como plantillas comerciales independientes.
- La misma matriz exige `financial_core_modules`, `income_flow`, `expense_flow`, `financial_tables` y `financial_reports`: ver o configurar un vertical no concede permisos financieros, pero todo ingreso/egreso operativo debe terminar en los modulos centrales `ventas`, `pagos`, `finanzas`, `bancos_pagos`, `tesoreria_presupuesto` y `reportes` segun licencia/rol.
- Un vertical puede aparecer como profesional solo si no duplica clientes, productos, ventas ni pagos y declara alcance de configuracion, tablas, permisos, flujo y reportes; la prueba backend bloquea regresiones.

2026-05-12: Nota operativa para nucleo configurable por plantilla
- Los roles como `cajero`, `profesor`, `entrenador`, `paciente`, `odontologo`, `tecnico`, `estilista` o `asesor` se declaran en la plantilla, pero se crean/gobiernan desde usuarios y roles del nucleo.
- Los productos y servicios cobrables de cada vertical se declaran como datos guia y se administran desde el nucleo de productos/servicios, no desde tablas paralelas del vertical.
- Las estaciones son recursos configurables del negocio. La plantilla decide su nombre visible entre estaciones, apartamentos, puestos, carros/vehiculos, bahias, aulas, consultorios, oficinas, canchas, salas o similares.
- `adaptacion_nucleo` es metadata de configuracion y no concede permisos por si sola; la operacion sigue dependiendo de licencia, rol, permisos de pagina/modulo y `empresa_id`.

2026-05-12: Nota operativa para Probar Gratis del portal
- `/descripcion_de_los_sistemas.html` es la landing publica oficial para mostrar la informacion del sistema o vertical elegido desde `web/index.html`.
- `/descripcion_de_los_sistemas.ht` queda como compatibilidad publica legacy servida como HTML por backend; no concede sesion, rol ni acceso a datos empresariales.
- El flujo `index.html` -> ficha descriptiva -> `Probar Gratis` conserva el contrato publico: solo muestra informacion comercial y deriva al registro/licencia correspondiente.

2026-05-12: Nota operativa para Centro de mando super
- `web/super/licencias_resumen.html` sigue siendo la pagina de Centro de mando del panel `super_administrador`.
- La reconstruccion visual no agrega permisos, roles, endpoints ni acceso a empresas; reutiliza las APIs super ya autenticadas para metricas, PostgreSQL, alertas, errores, servidores, licencias, empresas y consumos.
- El rol `control_super_administrador` conserva su navegacion limitada definida por el shell super; la pagina nueva no amplía ese alcance por si sola.

2026-05-11: Nota operativa para 2FA en login de administradores
- `security.admin_2fa.enabled` vive en configuracion avanzada y solo lo gobierna `super_administrador` mediante `/super/api/config/admin_2fa`.
- Cuando el switch global esta apagado, `login.html` oculta el campo de codigo 2FA y `AdminLoginHandler` no exige OTP aunque la cuenta conserve secreto TOTP guardado.
- Cuando el switch global esta activo, solo las cuentas con `administradores.totp_enabled=1` y secreto confirmado deben enviar `otp_code`; las demas mantienen el flujo normal de correo/contrasena.
- La vista `web/super/seguridad_2fa.html` sigue siendo el control por cuenta para generar, confirmar o desactivar el secreto; no concede permisos adicionales ni cambia roles de empresa.

2026-05-11: Nota operativa para alcance vertical por licencia
- El permiso efectivo de una empresa combina rol, licencia activa, reglas finas de empresa y `vertical_scope`.
- `vertical_scope` se resuelve desde el tipo/preconfiguracion de empresa o, si no hay preconfiguracion, desde la licencia cuando declara un solo vertical. Solo restringe plantillas de negocio; el nucleo universal de ventas, estaciones, productos, clientes, finanzas, pagos, facturacion, reportes y seguridad sigue gobernado por licencia y rol.
- Ejemplo: una empresa tipo `gimnasio` puede ver y operar `gimnasio` si su rol/licencia lo permiten, pero `odontologia`, `parqueadero`, `domicilios`, `taxi_system`, `apartamentos_turisticos`, `propiedad_horizontal`, `alquileres`, `drogueria_farmacia`, `aiu_construccion`, `reservas_hotel` y los 20 plantillas nuevas quedan ocultos/bloqueados salvo que correspondan al tipo activo.
- Los endpoints plantillas tambien quedan protegidos por esta regla; no basta ocultar el enlace del menu.
- El checkout de licencias valida que la licencia base coincida con el tipo de empresa antes de crear pago, activar sin pago o procesar confirmaciones.

2026-05-11: Nota operativa para integracion de plantillas
- `/api/public/plantillas_nuevas/catalogo` y `/api/public/plantillas_integracion/catalogo` son rutas publicas de solo lectura para portada, tarjetas y matriz comercial. No conceden operacion sobre plantillas, no exponen datos de empresas y no reemplazan los endpoints autenticados de `/api/empresa/...` o `/super/api/...`.
- Los permisos/licencias siguen siendo necesarios, pero no son suficientes para mostrar un vertical: ademas debe estar marcado como visible operativo en la matriz de integracion.
- Los 20 plantillas nuevas permanecen visibles como `plantilla_integrada_nucleo` porque usan el motor comun y no duplican clientes, productos/servicios, ventas ni pagos.
- Para version masiva, la preconfiguracion marca los 20 plantillas nuevas como `produccion_masiva` con prioridad 1-20.
- Los 20 plantillas nuevas quedan como plantillas reales sobre el nucleo unico, con decision de produccion masiva y prioridad 1-20.
- `web/super/plantillas_produccion_masiva.html` queda como vista de gobierno para `super_administrador`; consume `/super/api/plantillas_nuevas/catalogo` y no crea permisos nuevos. El rol `control_super_administrador` no recibe este acceso en su navegacion limitada.
- La accion `Asegurar 20` usa el mismo endpoint super con POST y requiere alcance de super administrador; no concede permisos empresariales ni activa modulos por rol, solo asegura catalogo comercial/preconfiguracion/licencias.
- Gimnasio, odontologia, parqueadero, taxi system, domicilios, apartamentos turisticos, propiedad horizontal, alquileres, drogueria/farmacia y AIU construccion quedan visibles como `plantilla_integrada_nucleo` al enlazar sus personas/servicios/tickets/viajes/pedidos/reservas/cargos/recaudos/contratos/facturas o expedientes sanitarios con clientes, servicios, inventario, ventas, pagos y facturacion centrales.
- La ocultacion por matriz sigue aplicando a cualquier vertical futuro que no tenga integracion real; no concede ni retira permisos de API, solo evita prometer al usuario un modulo operativo antes de su migracion al nucleo.
- `linkPlantillasIntegracion` queda protegido por `seguridad:R`; permite consultar la matriz operativa desde Administrar empresa > Configuracion sin conceder permisos de operacion sobre cada vertical.
- Los botones `Sincronizar` dentro de esa matriz tambien consultan `/api/empresa/permisos_contexto`; se habilitan solo cuando la pagina del vertical esta permitida o el modulo tiene accion de creacion efectiva. Cada endpoint vertical mantiene su propio wrapper de autorizacion y `empresa_id`.
- El catalogo de la matriz declara `required_permissions` por vertical para auditoria; esa metadata no concede permisos por si sola, solo documenta el rol/modulo/accion que debe existir para operar la plantilla.

2026-05-11: Nota operativa para panel super profesional
- El panel principal de `super_administrador` muestra solo accesos de gobierno diario: centro de mando, seleccion de empresa, tipos de empresa, licencias, administradores, roles, permisos, 2FA, integracion IA, reglas de chat, contexto de negocio, alertas, seguridad VPS, PostgreSQL, configuracion y ayuda.
- No se agregan permisos ni se elimina autorizacion de rutas existentes; los modulos secundarios retirados del menu principal conservan proteccion por rol y pueden mantenerse como rutas directas o historicas.
- El rol `control_super_administrador` conserva navegacion limitada y no recibe ayuda privada ni configuracion ampliada.

2026-05-10: Nota operativa para preconfiguraciones y configuracion guiada
- Las preconfiguraciones por tipo de empresa se aseguran por `tipo_empresa_id`, no por conteo global; si un tipo activo no tiene plantilla, el arranque o `seed_defaults` crea una plantilla inicial sin sobrescribir personalizaciones.
- Los 20 plantillas nuevas conservan plantilla inteligente propia para estaciones, productos/servicios, usuarios guia, roles y tareas iniciales.
- La pagina/API interna de configuracion guiada conserva control por modulo `seguridad`, pero el boton visible del submenu empresarial fue retirado y el robot no se inicia automaticamente al entrar a una empresa nueva.
- La preconfiguracion aplicada al crear empresa sigue respetando `empresa_id` y solo crea datos guia de esa empresa; el usuario decide conservar o eliminar esos datos guia.

2026-05-10: Nota operativa para modo tactil en carritos
- `carritos` mantiene los mismos permisos existentes para configurar y operar el carrito.
- La bandera `modo_pantalla_tactil` no crea permiso nuevo: se guarda dentro de `estaciones_config` y solo la pueden modificar los roles que ya tienen acceso a configuracion de estaciones/carrito.
- El efecto visible aplica en `carrito_de_compras.html` y `buscar_producto_botones.html`, respetando `empresa_id`, permisos de estacion/venta directa y visibilidad de botones configurada.

2026-05-10: Nota operativa para `alertas_sistema` super
- El modulo `Alertas del sistema` vive solo en Super administrador > Infraestructura y comunicaciones.
- Pagina: `web/super/alertas_sistema.html`; API: `/super/api/alertas_sistema`.
- Roles: acceso exclusivo para `super_administrador`; el rol `control_super_administrador` conserva su lista limitada y no recibe este modulo.
- Acciones: leer configuracion/estado/historial, actualizar umbrales, probar correo y ejecutar evaluacion manual.
- Alcance: opera sobre `pcs_superadministrador`, metricas del VPS y configuracion Gmail SMTP global; no concede acceso a datos de empresas ni modifica licencias.
- Side effects: puede enviar correo a `powerfulcontrolsystem@gmail.com` o al destino configurado y registra cada intento en `super_alertas_eventos`.

2026-05-10: Nota operativa para roles finos, paginas nuevas y ayuda super
- Se agregan claves independientes al catalogo de permisos/licencias: `crm_unificado`, `reservas_hotel`, `chat_tareas`, `horarios_trabajadores`, `asistencia_empleados`, `vehiculos_registro`, `hoja_vida_operativa`, `ubicacion_gps`, `nomina_sueldos`, `reportes`, `auditoria`, `backups`, `documentos_onlyoffice` y `nextcloud`.
- Las paginas del menu empresarial y submenus quedan cubiertas por `permissionPagesCatalogOrdered` en backend y por `menuPermissionCatalog` en frontend; no debe existir un boton `link...` visible sin regla de modulo/accion.
- Los endpoints que antes dependian de modulos amplios ahora usan wrappers especificos cuando corresponde: CRM unificado, reservas hoteleras, chat/tareas, personal/activos, nomina, reportes, auditoria, backups, OnlyOffice y Nextcloud.
- Compatibilidad de licencias: si una licencia antigua habilita `ventas`, `seguridad`, `finanzas`, `inventario` o `clientes`, esas claves siguen habilitando las funciones que fueron separadas en modulos finos para evitar cortes operativos durante la migracion.
- La ayuda administrativa completa `/ayuda/ayuda.html` es privada para `super_administrador` y se accede desde el boton `Ayuda super administrador` del panel super; el rol `control_super_administrador` no recibe este acceso.
- Validacion: `go test ./...`, `node --check web/js/administrar_empresa.js`, `node --check web/js/super_administrador.js` y auditoria de IDs `link...` contra catalogos backend/frontend.

2026-05-06: Nota operativa para modulos empresariales Colombia
- Se mantienen claves independientes: `bancos_pagos`, `gestion_documental`, `cumplimiento_kyc`, `contratos_obligaciones` y `calidad_procesos`, activables por licencia mediante `licencias.modulos_habilitados`.
- Las paginas `linkBancosPagos`, `linkGestionDocumental`, `linkCumplimientoKYC`, `linkContratosObligaciones` y `linkCalidadProcesos` quedan registradas en el catalogo de paginas y en el menu de Administrar empresa. Los tickets de ayuda usan flujo propio desde menu flotante y bandeja de super administrador.
- Cada endpoint usa wrapper propio `WithEmpresa*Permissions` y mantiene alcance estricto por `empresa_id`.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa`, `supervisor_sucursal`, `contabilidad` y `auditor`; eliminar para `admin_empresa`.
- Se comparte un nucleo tecnico para evitar duplicar formularios, dashboards, bitacoras o tablas por modulo.

2026-05-06: Nota operativa para `logistica_wms`
- Se agrega clave independiente `logistica_wms`, activable por licencia mediante `licencias.modulos_habilitados`.
- La pagina `linkLogisticaWMS` queda registrada en el catalogo de paginas y se muestra en Administrar empresa > Inventario y compras.
- El endpoint `/api/empresa/logistica_wms` usa `WithEmpresaWMSPermissions`; no abre rutas publicas.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa`, `supervisor_sucursal`, `inventario` y `compras`.
- Todas las ubicaciones, ordenes, items, despachos y eventos incluyen `empresa_id`; el modulo se integra con inventario sin duplicar productos ni existencias.

2026-05-21: Nota de navegacion para `linkImpuestos`
- `linkImpuestos` conserva el modulo efectivo `facturacion` con accion `U`.
- El acceso visible cambia de Configuracion a `Finanzas y cumplimiento` para alinear impuestos con el centro financiero, contable y fiscal; el catalogo interno lo mantiene dentro del grupo financiero-contable universal.

2026-05-06: Nota operativa para `declaraciones_tributarias`
- Se agrega clave independiente `declaraciones_tributarias`, activable por licencia mediante `licencias.modulos_habilitados`.
- Las paginas `linkDeclaracionesTributarias` y `linkDeclaracionesTributariasMenu` quedan registradas en el catalogo de paginas y se muestran en Finanzas y cumplimiento / Centro financiero y contable.
- El endpoint `/api/empresa/declaraciones_tributarias` usa `WithEmpresaDeclaracionesTributariasPermissions`; no abre rutas publicas.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa` y `contabilidad`, alineado con la matriz financiera.
- Todas las declaraciones, movimientos y vencimientos incluyen `empresa_id`; el calendario editable no concede permisos cruzados ni mezcla obligaciones entre empresas.

2026-05-06: Nota operativa para `portal_terceros_certificados`
- Se agrega clave independiente `portal_terceros_certificados`, activable por licencia mediante `licencias.modulos_habilitados`.
- Las paginas `linkPortalTercerosCertificados` y `linkPortalTercerosCertificadosMenu` quedan registradas en el catalogo de paginas y se muestran en Finanzas y cumplimiento / Centro financiero y contable.
- El endpoint administrativo `/api/empresa/portal_terceros_certificados` usa `WithEmpresaPortalTercerosPermissions`.
- La consulta externa `/api/public/certificados_tributarios` solo permite ver certificados emitidos/enviados mediante token publico y no lista informacion de otros terceros.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa` y `contabilidad`.
- Todas las tablas incluyen `empresa_id`; los tokens no reemplazan el aislamiento interno por empresa.

2026-05-06: Nota operativa para `activos_fijos_niif_fiscal`
- Se agrega clave independiente `activos_fijos_niif_fiscal`, activable por licencia mediante `licencias.modulos_habilitados`.
- Las paginas `linkActivosFijosNIIF` y `linkActivosFijosNIIFMenu` quedan registradas en el catalogo de paginas y se muestran en Finanzas y cumplimiento / Centro financiero y contable.
- El endpoint `/api/empresa/activos_fijos_niif_fiscal` usa `WithEmpresaActivosFijosNIIFPermissions`; no abre rutas publicas.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa` y `contabilidad`, alineado con la matriz financiera.
- Reutiliza `empresa_contabilidad_activos_fijos`, `empresa_contabilidad_activos_depreciacion` y `empresa_contabilidad_activos_eventos`, todas filtradas por `empresa_id`.

2026-05-06: Nota operativa para `propiedad_horizontal`
- Se agrega clave independiente `propiedad_horizontal`, activable por licencia mediante `licencias.modulos_habilitados`.
- La pagina `linkPropiedadHorizontal` queda registrada en el catalogo de paginas y se muestra en Administrar empresa.
- El endpoint `/api/empresa/propiedad_horizontal` usa `WithEmpresaPropiedadHorizontalPermissions`; no abre rutas publicas.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa`, `supervisor_sucursal` y roles operativos autorizados por plantillas.
- Todas las tablas nuevas incluyen `empresa_id`; unidades, residentes, cargos, recaudos, PQR y asambleas no se mezclan entre empresas.

2026-05-06: Nota operativa para promocion de licencias por asesor
- La promocion se controla desde Super administrador > Asesor comercial, con check de activacion y porcentaje.
- El codigo de asesor debe existir, estar activo y tener invitacion aceptada para aplicar descuento.
- La promocion no reemplaza la comision del asesor: conserva `asesor_id` en pagos Wompi/Epayco/manual y sigue registrando comisiones.

2026-05-06: Nota operativa para `cierre_fiscal`
- Se agrega clave independiente `cierre_fiscal`, activable por licencia mediante `licencias.modulos_habilitados`.
- Las paginas `linkCierreFiscal` y `linkCierreFiscalMenu` quedan registradas en el catalogo de paginas y se muestran en Finanzas y cumplimiento / Centro financiero y contable.
- El endpoint `/api/empresa/cierre_fiscal` usa `WithEmpresaCierreFiscalPermissions`; no abre rutas publicas.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa` y `contabilidad`, alineado con los controles de cierre financiero.
- Las tablas nuevas incluyen `empresa_id`; las excepciones y eventos se registran por periodo, modulo, accion y usuario sin conceder permisos cruzados ni mezclar empresas.
- El cierre/reapertura desde `contabilidad_colombia` sincroniza el periodo fiscal para evitar duplicar fuentes de verdad.

2026-05-06: Nota operativa para `centros_costo`
- Se agrega clave independiente `centros_costo`, activable por licencia mediante `licencias.modulos_habilitados`.
- Las paginas `linkCentrosCosto` y `linkCentrosCostoMenu` quedan registradas en el catalogo de paginas y se muestran en Finanzas y cumplimiento / Centro financiero y contable.
- El endpoint `/api/empresa/centros_costo` usa `WithEmpresaCentrosCostoPermissions`; no abre rutas publicas.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa` y `contabilidad`, alineado con los modulos financieros.
- Todas las tablas nuevas incluyen `empresa_id`; los movimientos del dashboard se consultan desde modulos existentes sin conceder permisos cruzados ni mezclar empresas.

2026-05-06: Nota operativa para `crm_ventas_avanzadas`
- Se formaliza bajo el modulo/licencia `crm_unificado`, pagina `linkCRMComercial`; no se duplican clientes ni ventas.
- Los endpoints `/api/empresa/crm_avanzado` y `/api/empresa/crm/*` usan `WithEmpresaCRMUnificadoPermissions`.
- Roles base: lectura para roles comerciales autorizados por CRM; crear/actualizar para `admin_empresa`, `supervisor_sucursal` y roles con permiso de CRM unificado.
- Las metas, forecast, scoring, agenda, salud comercial, acciones priorizadas y conversiones se calculan por `empresa_id`.

2026-05-12: Nota operativa para CRM empresarial
- El tablero CRM agrega salud comercial, valor en riesgo, leads sin contacto, oportunidades estancadas, responsables y canales sin crear rutas publicas nuevas.
- El permiso efectivo queda ligado a `crm_unificado`; `clientes` conserva su CRUD propio en `/api/empresa/clientes`.
- No se amplian privilegios: usuarios sin acceso a `crm_unificado` no pueden usar leads, interacciones, campanas ni dashboard avanzado.

2026-05-06: Nota operativa para `inventario_avanzado`
- Se agrega la pagina `linkInventarioAvanzado` bajo el modulo/licencia existente `inventario`; no se crea un inventario paralelo.
- El endpoint `/api/empresa/inventario_avanzado` reutiliza `WithEmpresaInventarioPermissions`.
- Roles base: lectura para roles operativos autorizados por inventario; crear/actualizar para `admin_empresa`, `supervisor_sucursal` e `inventario`.
- Los lotes, seriales, reservas y valorizaciones se calculan siempre por `empresa_id`.

2026-05-06: Nota operativa para `compras_avanzadas`
- Se agrega la pagina `linkComprasAvanzadas` bajo el modulo/licencia existente `compras`; no se crea un modulo paralelo.
- El endpoint `/api/empresa/compras_avanzadas` reutiliza `WithEmpresaComprasPermissions`.
- Roles base: lectura para roles operativos autorizados por compras; crear/actualizar/aprobar para `admin_empresa`, `supervisor_sucursal` y `compras`.
- Las requisiciones, items, cotizaciones, aprobaciones y recepciones guardan `empresa_id`, manteniendo aislamiento multiempresa.

2026-05-06: Nota operativa para QA transversal de modulos
- La evaluacion frontend de permisos en `web/js/administrar_empresa.js` reconoce `administrador_total` como rol con acceso total, igual que `super_administrador`, para evitar que el menu oculte modulos que el backend permite.
- La prueba autenticada sobre Motel Calipso confirma `super_administrador` efectivo en `/api/empresa/permisos_contextoempresa_id=7&include_matrix=1`.
- No se agregan rutas publicas nuevas ni se amplian permisos por fuera de las licencias/modulos existentes; el cambio solo corrige coherencia de menu y documenta QA.

2026-05-06: Nota operativa para `portal_contador`
- Se agrega clave independiente `portal_contador`, activable por licencia mediante `licencias.modulos_habilitados`.
- Las paginas `linkPortalContador` y `linkPortalContadorMenu` quedan registradas en el catalogo de paginas y se muestran dentro del Centro financiero y contable.
- El endpoint `/api/empresa/portal_contador` usa `WithEmpresaPortalContadorPermissions`; no abre rutas publicas.
- Roles base: lectura para roles operativos; crear/actualizar para `admin_empresa` y `contabilidad` segun la matriz financiera.
- Todas las tablas incluyen `empresa_id`. El campo `cliente_empresa_id` es una referencia opcional a otra empresa del sistema, no reemplaza el control de alcance ni concede permisos cruzados.

2026-05-06: Nota operativa para `cobranza`
- Se agrega clave independiente `cobranza`, activable por licencia mediante `licencias.modulos_habilitados`.
- Las paginas `linkCobranza` y `linkCobranzaMenu` quedan registradas en el catalogo de paginas y se muestran dentro del Centro financiero y contable.
- El endpoint `/api/empresa/cobranza` usa `WithEmpresaCobranzaPermissions`; no abre rutas publicas ni proveedores externos directos.
- Roles base: lectura para roles operativos; crear/actualizar para `admin_empresa`, `supervisor_sucursal` y `contabilidad`; marcar promesas como cumplidas/incumplidas usa accion de aprobacion.
- Todas las plantillas, campanas, gestiones y promesas incluyen `empresa_id`; las gestiones referencian `empresa_cuentas_por_cobrar` por `cuenta_id` para no duplicar cartera.

2026-05-06: Nota operativa para `importaciones_costeo`
- Se agrega clave independiente `importaciones_costeo`, activable por licencia mediante `licencias.modulos_habilitados`.
- La pagina `linkImportacionesCosteo` queda registrada en el catalogo de paginas y se muestra en Administrar empresa > Inventario y compras.
- El endpoint `/api/empresa/importaciones_costeo` usa `WithEmpresaImportacionesCosteoPermissions`; no abre rutas publicas.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa`, `supervisor_sucursal`, `compras` e `inventario`, porque el costo aterrizado cruza compra internacional e inventario.
- Todas las importaciones, items y costos incluyen `empresa_id`; no se mezclan embarques ni costos entre empresas.

2026-05-06: Nota operativa para compatibilidad de `activos_fijos` avanzado
- Las acciones antiguas de activos dentro de `/api/empresa/contabilidad_colombia_avanzada` se conservan por compatibilidad y heredan el control financiero/contable.
- El modulo formal nuevo es `activos_fijos_niif_fiscal`, con wrapper, licencia y pantalla propia; ambos caminos usan las mismas tablas para no duplicar activos, eventos ni depreciaciones.
- Las tablas `empresa_contabilidad_activos_depreciacion` y `empresa_contabilidad_activos_eventos` usan `empresa_id` y no comparten eventos, depreciaciones ni mantenimientos entre empresas.

2026-05-06: Nota operativa para `nomina_sueldos` y Nomina Colombia avanzada
- La capa Colombia de nomina se implementa dentro del modulo existente `nomina_sueldos`; no crea licencia, wrapper ni pagina duplicada para evitar fragmentar el control financiero.
- Las acciones `conceptos_colombia`, `novedades_colombia`, `pila_colombia`, `dashboard_colombia`, `concepto_colombia`, `novedad_colombia`, `generar_pila`, `seed_colombia`, `aprobar_novedad_colombia` y `seed_motel_calipso` siguen bajo `/api/empresa/nomina` y heredan el mismo alcance por `empresa_id` de la nomina actual.
- Las nuevas tablas `empresa_nomina_colombia_conceptos`, `empresa_nomina_colombia_novedades` y `empresa_nomina_colombia_pila_resumen` son multiempresa y se consultan solo desde el contexto empresarial autorizado.

2026-05-06: Nota operativa para `tesoreria_presupuesto`
- Se agrega clave independiente `tesoreria_presupuesto`, activable por licencia mediante `licencias.modulos_habilitados`.
- La pagina `linkTesoreriaPresupuesto` queda registrada en el catalogo de paginas y se muestra dentro de Centro financiero y contable.
- El endpoint `/api/empresa/tesoreria_presupuesto` usa `WithEmpresaTesoreriaPresupuestoPermissions`; no abre rutas publicas.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa` y `contabilidad`; eliminacion queda alineada con la politica financiera existente.
- Todas las cuentas, presupuestos, partidas y flujos tienen `empresa_id`; no se mezclan bancos ni presupuestos entre empresas.

2026-05-06: Nota operativa para `produccion_mrp`
- Se agrega clave de modulo independiente `produccion_mrp`, activable por licencia mediante `licencias.modulos_habilitados`.
- La pagina `linkProduccionMRP` queda registrada en el catalogo de paginas y se muestra en Administrar empresa > Inventario y compras.
- El endpoint `/api/empresa/produccion_mrp` usa `WithEmpresaProduccionMRPPermissions`; no comparte wrappers genericos ni expone rutas publicas.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar/eliminar para `admin_empresa`, `supervisor_sucursal`, `inventario` y `compras`, porque el flujo cruza produccion, componentes y planeacion de abastecimiento.
- Todas las tablas del modulo incluyen `empresa_id`; los consumos y planes MRP de una empresa no se mezclan con recetas u ordenes de otra.

2026-05-05: Auditoria de enlaces por empresa y no duplicidad de modulos
- El menu principal de `web/administrar_empresa.html` queda alineado con `permissionPagesCatalogOrdered` y `web/js/administrar_empresa.js`: no hay enlaces visibles sin regla de pagina ni sin regla frontend.
- Las claves de modulo del backend (`empresa_permisos.go`), licencias (`web/super/licencias.html`) y menu empresa (`web/js/administrar_empresa.js`) coinciden para modulos base y plantillas.
- No se duplican wrappers ni funciones para resolver permisos: los nuevas plantillas reutilizan `withEmpresaRolePermissions` y solo declaran wrappers finos cuando necesitan una clave de modulo independiente.
- Auditoria estatica ejecutada: sin rutas `/api/empresa` duplicadas, sin claves de modulo duplicadas y todas las rutas empresariales registradas con `WithEmpresa*` o `WithEmpresaPublicScope`.

2026-05-05: Modulo `carnets` para empleados y usuarios
- Se agrega clave de licencia y permisos `carnets`.
- La pagina `linkCarnets` queda bajo modulo `carnets` con accion `C`.
- El endpoint `/api/empresa/carnets` usa `WithEmpresaCarnetsPermissions` y opera siempre dentro del `empresa_id` validado.
- Roles base: lectura para roles operativos; crear/actualizar/aprobar para `admin_empresa`, `supervisor_sucursal` y `cajero`; eliminar/revocar queda restringido a `admin_empresa` y `supervisor_sucursal`.
- La pantalla de licencias (`web/super/licencias.html`) incluye `carnets` como modulo activable y dentro del preset enterprise.

2026-05-05: Rectificacion de aislamiento por empresa en todos los modulos
- Todas las rutas privadas bajo `/api/empresa/...` deben pasar por un wrapper `WithEmpresa*` o `WithEmpresaPublicScope`.
- El wrapper central valida que el `empresa_id` usado para permisos sea consistente entre query string, cabecera `X-Empresa-ID`, formulario/multipart y cuerpo JSON.
- Si una peticion envia `empresa_id` contradictorios, el backend responde `400` con `empresa_id no coincide con el contexto de empresa` antes de llegar al handler del modulo.
- El contexto `empresaID` queda inyectado en `request.Context()` y las funciones comunes `parseEmpresaIDQuery` / `parseInt64QueryOptional(..., "empresa_id")` priorizan ese contexto.
- La validacion aplica transversalmente a ventas/POS, estaciones, inventario, compras, clientes, finanzas, facturacion, usuarios, roles, hotel/motel, gimnasio, odontologia, domicilios, Taxi System, turnos, control electrico, hoja de vida, reportes, IA empresarial, backups, soporte remoto, documentos y modulos ERP adicionales.
- Las rutas publicas (`/api/public/...`) siguen siendo de solo lectura o flujos publicos controlados, y deben resolver empresa por `empresa_id`, slug o token publico segun el contrato de cada modulo.

2026-05-05: Nota operativa para `portal_publico`, carta publica y Motel Calipso
- `web/index.html` es una superficie publica comercial; sus descripciones de modulos y tarjetas fallback no requieren sesion ni permisos empresariales porque no ejecutan operaciones.
- La pagina `visualizar_productos_y_precios_publico.html` y la ruta `/{empresa_slug}/visualizar_productos_y_precios_publico.html` son publicas y de solo lectura. Deben pasar por `AuthMiddleware` sin sesion, igual que `venta_publica.html`, porque el alcance de datos se resuelve por slug/empresa y API publica.
- La administracion de la carta sigue protegida: `administrar_empresa/carta_productos_publica.html` usa modulo `venta_publica`, `linkCartaProductosPublica`, licencia activa y permisos por rol.
- La publicacion operativa de Motel Calipso (`empresa_id=7`, slug `motel-calipso`) no cambia permisos globales; solo crea/actualiza datos de venta publica, paginas, items y publicaciones de red social dentro de esa empresa.

2026-05-05: Nota operativa para `modulos plantillas`, roles y licencias
- Se agregan claves de modulo independientes para `venta_publica`, `gimnasio`, `taxi_system`, `domicilios`, `alquileres`, `odontologia`, `turnos_atencion` y `control_electrico`.
- Cada clave entra al catalogo central de permisos por rol (R/C/U/D/A), al catalogo de paginas `link*` del panel empresa y a `licencias.modulos_habilitados`, permitiendo activar o desactivar el acceso tanto por licencia como por usuario/rol.
- Los endpoints protegidos principales dejan de depender del modulo generico `ventas` o `seguridad` cuando existe modulo vertical propio: `WithEmpresaDomiciliosPermissions`, `WithEmpresaTaxiSystemPermissions`, `WithEmpresaGimnasioPermissions`, `WithEmpresaAlquileresPermissions`, `WithEmpresaOdontologiaPermissions`, `WithEmpresaTurnosAtencionPermissions`, `WithEmpresaVentaPublicaPermissions` y `WithEmpresaControlElectricoPermissions`.
- Las rutas publicas asociadas a taxi, domicilios, turnos y venta publica conservan alcance publico sanitario/operativo; la administracion interna sigue gobernada por licencia + rol + pagina.

2026-05-05: Nota operativa para `carta publica de productos`
- El acceso `linkCartaProductosPublica` queda registrado bajo modulo `venta_publica` con accion `C`, igual que el endpoint administrativo `/api/empresa/venta_publica`.
- La pantalla administrativa tambien consulta `/api/empresa/productos` para leer inventario activo; por tanto los roles operativos deben combinar permiso de `venta_publica` para publicar y permiso de inventario para consultar productos.
- La pagina publica `visualizar_productos_y_precios_publico.html` usa `/api/public/venta_publicaaction=catalogo` y solo expone catalogo de lectura: no crea pedidos, pagos ni carrito.

2026-04-30: Nota operativa para `chat IA`, `documentos dinamicos`, `empresas compartidas` y `pagos`
- La configuracion del chat IA flotante, voz y emisora online queda en Administrar empresa > Configuracion > Configurar chat IA; es una configuracion empresarial por `empresa_id` y no una concesion de permisos super.
- `/generate` y `/download` para documentos dinamicos requieren sesion y deben asociarse al contexto empresarial cuando se usen desde empresa; no entregan credenciales ni SQL libre a la IA.
- La consulta/revocacion de administradores compartidos se controla por pertenencia administrativa a la empresa compartida; quien compartio y quien recibio pueden retirar acceso, registrando actor y fecha.
- El fallback clasico de Epayco no cambia permisos de checkout: solo modifica el transporte hacia la pasarela usando POST firmado cuando Smart Checkout no entrega token.
2026-04-25: Nota de gobernanza (permisos por rol, licencia y menÒº empresa)
- El panel `web/super/permisos_rol.html` configura la **matriz por rol** (mÃƒÂ³dulo Ãƒâ€” R/C/U/D/A) y **anulaciones por funciÃƒÂ³n** del menÃƒÂº `administrar_empresa` (claves `link*`). El backend expone en `GET /super/api/roles_de_usuario/permisos` etiquetas legibles y agrupaciÃƒÂ³n para auditorÃƒÂ­a y UI.
- La **licencia** (`licencias.modulos_habilitados` en `web/super/licencias.html`) define el **techo** de mÒ³dulos contratados: lista vacÒ­a = sin restricciÒ³n de mÒ³dulo; lista con valores = solo esos mÒ³dulos para la empresa, aplicada antes de la matriz de rol.
- No se agrega un tercer sistema Ã¢â‚¬Å“universalÃ¢â‚¬Â paralelo: la combinaciÃƒÂ³n licencia + rol + reglas de pÃƒÂ¡gina del catÃƒÂ¡logo `permissionPagesCatalogOrdered` en `empresa_permisos.go` es el modelo soportado.

2026-04-26: Nota operativa para `inventario`, `finanzas`, `asistencia` y `usuarios`
- El nuevo acceso `linkGeneradorCodigosBarras` queda bajo modulo `inventario` con accion `U` porque actualiza `productos.codigo_barras`; no crea wrapper nuevo ni ruta publica.
- La vinculacion de asistencia con usuarios internos reutiliza `/api/empresa/usuarios` y `/api/empresa/asistencia_empleados`, manteniendo `empresa_id` y permisos ya existentes de seguridad.
- Las nuevas opciones de integracion contable en finanzas son valores de configuracion, no conexiones externas automaticas; no agregan permisos ni secretos.

2026-04-26: Nota operativa para `reportes`, `finanzas` y `chat IA`
- `POST /api/empresa/reportes_ia_chat` queda como soporte tecnico del asistente global en modo reportes; usa el permiso general `linkReportes` y limita el consumo por empresa: 10 preguntas texto con GPT-5.4 mini y 2 reportes/exportes con GPT-5.5 al dia.
- La exportacion generada sigue pasando por `/api/empresa/reportesaction=export`, por lo que conserva los filtros, formatos y controles existentes del modulo de reportes.

2026-04-26: Nota operativa para `finanzas` ERP MVP
- Plan de cuentas, CxC, CxP, abonos/pagos de cartera y conciliacion bancaria por extractos quedan dentro del mismo modulo `finanzas`; reutilizan `WithEmpresaFinanzasPermissions` y no agregan nuevos wrappers publicos.
- Las acciones de abono/pago crean movimientos financieros y actualizan saldos de cartera dentro del mismo `empresa_id`; siguen bloqueadas por periodo contable cerrado cuando aplica.
- Los nuevos datasets contables (`contable_plan_cuentas`, `contable_cuentas_por_cobrar`, `contable_cuentas_por_pagar`, `contable_conciliacion_bancaria`) se consultan desde reportes con permisos financieros existentes.

2026-04-24: Nota operativa para `asesor comercial`, `licencias`, `pagos`, `super` y `seleccionar_empresa`
- `super_administrador` administra `web/super/asesor_comercial.html`: invita administradores registrados, configura porcentaje/plazo, desactiva asesores y marca comisiones pagadas mediante `/super/api/asesor_comercial`.
- `administrador` invitado solo obtiene la vista `Mis clientes` tras aceptar la invitacion por correo; no recibe permisos super ni acceso a empresas ajenas. La vista consume `/api/asesor_comercial/mis_clientes` y filtra por el email de la sesion.
- El checkout publico de licencias acepta `asesor_id` como codigo de asesor comercial. La asociacion comercial no cambia permisos de empresa: solo registra comisiones sobre pagos aprobados y renovaciones dentro del plazo configurado.

2026-05-04: Nota operativa para `asesor comercial`, `licencias`, `pagos` y `super`
- `super_administrador` mantiene control exclusivo sobre `/super/api/asesor_comercial` y ahora configura informacion de transferencia/comision por asesor: metodo, entidad, cuenta, titular, contacto, periodicidad, minimo y soporte requerido.
- La gestion de comisiones permite registrar estado de liquidacion, referencia de transferencia, fecha programada, soporte y observaciones. Esto no ejecuta pagos bancarios externos ni abre permisos a asesores; solo deja trazabilidad interna de liquidacion.
- El asesor comercial sigue viendo solo sus clientes/comisiones mediante `/api/asesor_comercial/mis_clientes`; no puede editar datos bancarios globales ni marcar pagos.

2026-04-20.3: Nota operativa para `soporte remoto` y `super` sobre activacion por defecto de RustDesk
- `super_administrador` sigue siendo el unico rol que configura y opera RustDesk desde la vista super; el cambio solo fija que la primera lectura de configuracion llegue activa por defecto con portal publico habilitado y modo local preseleccionado.
- No se crean permisos nuevos ni se amplian privilegios de `administrador`; se corrige un default funcional de la vista para que coincida con la operacion simplificada del modulo.

2026-04-21: Nota operativa para `compras`, `finanzas` y `administrar_empresa` sobre comprobantes adjuntos
- `administrador` y demÒ¡s perfiles empresariales ya autorizados en los mÒ³dulos de compras y finanzas pueden adjuntar y consultar comprobantes solo dentro del mismo `empresa_id`; el cambio no crea roles nuevos ni expone rutas pÒºblicas adicionales.
- `POST /api/empresa/compras/documentos/comprobante` reutiliza `WithEmpresaComprasPermissions` y `POST /api/empresa/finanzas/movimientos/comprobante` reutiliza `WithEmpresaFinanzasPermissions`; por tanto el alcance efectivo queda igual que en el CRUD principal de cada mÒ³dulo.
- La visualizaciÒ³n posterior del comprobante en listados usa una URL servida desde el mismo Ò¡rbol web del sistema, pero la referencia solo se genera para registros ya permitidos por el contexto autenticado de empresa.

2026-04-21: Nota operativa para `soporte remoto`, `super`, `administrar_empresa` y `portal_publico` sobre RustDesk simplificado
- `super_administrador` mantiene el control exclusivo del servicio base RustDesk desde `web/super/servidores.html`, incluyendo acciones de encendido, apagado, reinicio y prueba. El cambio no crea un rol nuevo ni delega esas acciones al panel empresarial.
- `administrador` de empresa sigue limitado a configurar los datos visibles del acceso remoto de su empresa en `web/administrar_empresa/soporte_remoto.html`: host, clave, instrucciones y enlaces de descarga. No obtiene permisos para operar el servicio del VPS.
- La pÒ¡gina pÒºblica `web/soporte_remoto_acceso.html` continÒºa siendo solo de consulta del acceso compartido por sesiÒ³n; expone descargas y datos de conexiÒ³n ya autorizados, sin ampliar privilegios a visitantes o usuarios sin sesiÒ³n.

2026-05-12: Nota operativa para `documentos_onlyoffice` y `backups`
- Los modos locales agregados para documentos y backups no abren permisos nuevos: siguen protegidos por `WithEmpresaDocumentosOnlyOfficePermissions` y `WithEmpresaBackupsPermissions`.
- `create_local`, `exportar_local` y `exportar_configuracion_local` descargan al navegador del usuario autenticado y no crean artefactos permanentes en el VPS.
- Los backups automaticos locales dependen de una sesion activa en la pagina de backups y de las reglas de descarga del navegador del dispositivo.

2026-05-13: Nota operativa para `documentos_onlyoffice`
- `create_edit_local` y `download&delete=1` siguen bajo `WithEmpresaDocumentosOnlyOfficePermissions`; no habilitan rutas publicas ni permisos nuevos.
- El editor OnlyOffice necesita una copia temporal por `empresa_id` para la sesion de edicion, pero el flujo oficial descarga el resultado al dispositivo y elimina el temporal al guardar.
- La correccion de apertura del editor solo ajusta ruta temporal y URL publica del Document Server para el navegador; no cambia roles, permisos, endpoints protegidos ni tokens temporales.
- La autorizacion efectiva sigue dependiendo del acceso al modulo `documentos_onlyoffice` y del aislamiento por empresa.

2026-05-13: Nota operativa para `mantenimiento programado`
- La configuracion del aviso vive en super administrador y sigue restringida al panel super mediante `/super/api/config/mantenimiento`.
- La consulta empresarial `/api/empresa/mantenimiento_programado` usa `WithEmpresaSelfServicePermissions`: exige usuario autenticado y alcance sobre `empresa_id`, pero no abre permisos administrativos nuevos.
- El check de aviso programado solo muestra mensaje en `administrar_empresa/panel.html`; no activa el bloqueo real `mantenimiento_activo`.

2026-04-20: Nota operativa para `backups empresariales`, `administrar_empresa` y `configuracion`
- La exportacion/importacion de configuracion por empresa reutiliza el modulo `backups empresariales` y no abre permisos nuevos: sigue limitada al acceso ya existente al enlace `Backups empresariales` del panel de empresa.
- El flujo importa solo tablas de configuracion asociadas al `empresa_id` destino y no restaura datos transaccionales, usuarios ni historiales operativos; por eso el alcance funcional se mantiene en configuracion/aprobacion y no altera wrappers de venta, inventario o finanzas.

2026-04-24: Nota operativa para `estaciones`, `ventas` y chat IA sobre pedidos por voz/texto
- La tarjeta **Pedidos con IA** embebida en `estaciones.html` llama a `POST /api/empresa/ia_pedidos_estacion/ejecutar`, registrada con el mismo wrapper de permisos de ventas que los carritos (`WithEmpresaVentasPermissions`). Quien pueda operar carritos/estaciones y tenga IA habilitada puede usar el asistente; no se agrega clave nueva de menu ni rol distinto.

2026-04-20: Nota operativa para `estaciones`, `carritos` y `administrar_empresa` sobre estaciones especiales reordenables
- `Caja`, `YouTube` y `Notas` siguen siendo recursos visibles del modulo de estaciones y no crean permisos nuevos ni cambian wrappers backend; la autorizacion sigue determinada por el acceso existente a la vista empresarial de estaciones.
- La nueva tarjeta `Notas` funciona como apoyo operativo local de la empresa con recordatorio temporizado; no altera roles, CRUD/A transaccional ni los carritos base enlazados a estaciones numeradas.

2026-04-20: Nota operativa para `estaciones` sobre multiples notas y repeticion local
- La evolucion de `Notas` a multiples recordatorios, countdown persistente y repeticion automatica no introduce endpoints nuevos ni modifica wrappers backend. Sigue dependiendo del mismo permiso de acceso a `administrar_empresa/estaciones.html`.
- El runtime multiple queda aislado en navegador por `empresa_id`; por eso no amplifica alcance entre roles o empresas, pero tampoco constituye respaldo compartido ni evidencia multiusuario persistida en servidor.

2026-04-20: Nota operativa para `autenticacion y sesiones` sobre correccion del ojito de contraseÒ±a
- La correccion del toggle visual en el login administrativo no cambia permisos, wrappers ni alcance publico de `/login.html`; solo restaura el comportamiento del control cliente sobre el campo de contraseÒ±a.
- El acceso sigue siendo publico para administradores y la autorizacion efectiva continua igual bajo el backend existente.

2026-04-20: Nota operativa para `autenticacion y sesiones` sobre recuperacion administrativa por enlace directo
- El restablecimiento de contraseÒ±a administrativa sigue siendo publico solo para la cuenta que recibe el correo de recuperaciÒ³n; el cambio no abre rutas nuevas ni amplÒ­a roles, solo elimina la necesidad de copiar un token manual al formulario.
- La validaciÒ³n efectiva continÒºa en backend con el mismo cÒ³digo de recuperaciÒ³n y expiraciÒ³n existentes; la diferencia es de UX y no de privilegios.

2026-04-20: Nota operativa para `portal_publico`, `autenticacion`, `super`, `administrar_empresa` y vistas embebidas sobre contraste de apariencias
- La correcciÒ³n de contraste y color para los seis temas no crea rutas nuevas ni modifica permisos por rol; solo garantiza que textos, tarjetas, estados vacÒ­os y componentes mantengan legibilidad coherente segÒºn la apariencia elegida.
- Los mÒ³dulos afectados conservan exactamente el mismo alcance pÒºblico o autenticado que ya tenÒ­an; el cambio es exclusivamente de presentaciÒ³n y consistencia visual.

2026-04-20: Nota operativa para `super`, `portal_publico` y `pagina principal publica` sobre WhatsApp configurable
- La nueva tarjeta de WhatsApp en configuraciÒ³n avanzada solo permite cambiar el nÒºmero del CTA flotante pÒºblico del `index.html`; no introduce rutas nuevas, permisos adicionales ni cambios de wrapper.
- El botÒ³n flotante del portal sigue siendo pÒºblico y de solo lectura/uso para todos los roles; la ediciÒ³n del nÒºmero continÒºa limitada al panel super ya existente.

2026-04-20: Nota operativa para `portal_publico`, `pagina principal publica` y `autenticacion y sesiones` sobre CTA superior de registro
- El nuevo botÒ³n `Crear cuenta` en `web/index.html` reutiliza el estilo visual del header de `/descripcion_de_los_sistemas.ht`, pero mantiene el mismo alcance pÒºblico del portal y solo deriva al registro administrativo ya existente en `/registrar_nuevo_usuario_administrador.html`.
- No se crean wrappers, permisos ni rutas nuevas: el cambio solo expone mejor un flujo pÒºblico ya permitido, mantiene `Iniciar sesiÒ³n` con la misma funciÒ³n anterior y compacta el header mÒ³vil sin alterar el alcance del portal.

2026-04-20: Nota operativa para `portal_publico` y `pagina principal publica` sobre simplificacion visual
- Retirar el hero y la tarjeta de accesos rapidos de `/descripcion_de_los_sistemas.ht` no cambia permisos, wrappers ni el destino del flujo publico; solo hace que la landing llegue directo al contenido detallado.
- La navegacion por hash y el CTA `Probar Gratis` mantienen exactamente el mismo alcance publico para todos los roles y para visitantes sin sesion.

2026-04-20: Nota operativa para `portal_publico` y `pagina principal publica`
- La renovacion visual de `/descripcion_de_los_sistemas.ht` no cambia roles, wrappers ni destinos protegidos: solo reorganiza los accesos internos como un menu profesional y sincroniza la apariencia con el tema activo del portal.
- El CTA `Probar Gratis` y la navegacion por hash mantienen el mismo alcance publico sin ampliar privilegios para `super_administrador`, `administrador` o usuarios de empresa.

2026-04-20: Nota operativa para `inventario`, `compras`, `finanzas` y `reportes`
- La correccion PostgreSQL de tendencia, proyeccion, reposicion preventiva, tablero financiero y salida PEPS no cambia roles, wrappers ni alcance por `empresa_id`. Solo sustituye SQL no portable y corrige la transaccion de lotes para que las mismas rutas protegidas respondan correctamente en PostgreSQL real.
- El flujo de `compras` para emitir, recepcionar y contabilizar ordenes de reposicion, asi como `finanzas/cierres_caja` y el dataset `empresarial_tablero`, mantienen exactamente el mismo modelo de permisos. La validacion runtime confirma operacion real sin ampliar privilegios para `inventario`, `compras`, `contabilidad`, `admin_empresa` o `super_administrador`.

2026-04-20: Nota operativa para `creditos`, `chat_y_tareas` y `administrar_empresa`
- La correccion final de PostgreSQL para abonos y citas no cambia roles, wrappers ni alcance por `empresa_id`. Solo elimina fallos de persistencia y de autorreparacion de esquema en runtime.
- `administrador` y usuarios autorizados de empresa mantienen exactamente las mismas acciones sobre creditos y agenda compartida; el cambio solo evita errores `400/500` falsos al ejecutar abonos o gestionar citas sobre PostgreSQL.

2026-04-20: Nota operativa para `creditos`, `finanzas` y `administrar_empresa`
- La correccion PostgreSQL de cartera y resumen de creditos no cambia permisos ni wrappers: solo reemplaza el mecanismo de insercion y endurece la consulta agregada para que las rutas existentes respondan coherentemente en PostgreSQL.
- Los roles empresariales siguen sujetos al mismo contexto `empresa_id`; el cambio evita respuestas falsas `400/500` en altas y resumenes sin ampliar alcance funcional.

2026-04-20: Nota operativa para `red_social_comercial`, `administrar_empresa` y portal publico
- La correccion PostgreSQL del modulo de publicaciones comerciales no agrega permisos nuevos ni modifica wrappers: la escritura sigue exigiendo el contexto autenticado de empresa y la lectura publica conserva el alcance ya previsto para vitrinas comerciales.
- El cambio solo estabiliza persistencia y lectura por `empresa_id` en PostgreSQL para que las publicaciones creadas por la empresa aparezcan tanto en su panel como en el feed publico sin error `500`.

2026-04-24: Nota operativa para `red_social_empresarial` y `venta_publica`
- La red social empresarial conserva escritura bajo `/api/empresa/publicaciones` con `WithEmpresaVentasPermissions`; la lectura pÒºblica `/api/public/publicaciones` no expone datos sensibles y ahora muestra nombre de empresa para el feed.
- Venta publica se administra desde el modulo independiente `venta_publica`; `/api/empresa/venta_publicaaction=paginas|config|catalogo` usa `WithEmpresaVentaPublicaPermissions`, mientras `/api/public/venta_publica` permanece publico solo para catalogo, creacion/estado de pago y datos sanitizados.

2026-04-20: Nota operativa para apariencia global, autenticacion y acceso publico a Juegos
- La reparacion de `menu.js`, `login.js`, `login_usuario.js` y los endpoints de login solo sincroniza una preferencia visual por usuario (`apariencia`) y no agrega permisos nuevos ni altera wrappers de acceso.
- La entrada `Juegos` vuelve al menÒº flotante como acceso pÒºblico controlado a `/Juegos/menu_juegos.html` y `/Juegos/n64/index.html`, sin exponer mÒ³dulos privados ni modificar el alcance de `super_administrador`, `administrador` o usuarios de empresa.

2026-04-19: Nota operativa para mensajeria multiusuario en `administrar_empresa` y `chat_y_tareas`
- La administradora puede buscar y marcar varios usuarios activos de su misma empresa para crear o ampliar conversaciones, pero el cambio no amplÒ­a roles ni wrappers: la operaciÒ³n sigue limitada al contexto autenticado y al `empresa_id` ya permitido.
- El backend ahora rechaza participantes tipo `usuario` cuyo `participante_ref_id` o correo pertenezcan a otra empresa, cerrando un cruce de datos sin conceder privilegios nuevos a administradores o usuarios empresa.

2026-04-19: Nota operativa de robustez para `administrar_empresa` y `chat_y_tareas`
- El dashboard principal de `Chat y tareas` agrega tarjetas resumen, acciones rapidas y refresco parcial, pero conserva el mismo alcance por rol y por `empresa_id` del modulo colaborativo.
- Las nuevas validaciones de `conversacion_id` y `tarea_id` solo bloquean referencias invalidas o cruzadas; no agregan permisos nuevos ni cambian wrappers de acceso para administradores o usuarios empresa autorizados.

2026-04-19: Nota operativa para `administrar_empresa` y `chat_y_tareas`
- El panel empresa ahora prioriza visualmente `Chat y tareas` como punto de entrada, pero solo cuando ese enlace sigue visible para el rol autenticado dentro del contexto real de permisos.
- La agenda de reuniones continÒºa compartida por `empresa_id`; mover el mÒ³dulo al inicio del shell y hacer protagonista el calendario no crea rutas nuevas ni amplÒ­a privilegios. Solo reduce fricciÒ³n para los mismos usuarios autorizados a consultar o registrar citas de la empresa.

2026-04-19: Nota operativa de UX para `estaciones` y `carritos`
- La mejora del estado visible de error en `carrito_de_compras.html` no cambia roles, wrappers ni acciones permitidas. Solo traduce y presenta mejor los fallos iniciales del flujo de apertura para que el mismo usuario autorizado pueda reintentar o volver a `estaciones.html` con menor ambigÒ¼edad operativa.

2026-04-19: Nota operativa para `estaciones` y `carritos` en PostgreSQL
- La correccion del listado y de los totales de `carritos_compra` no altera wrappers, roles ni alcance por `empresa_id`. El ajuste solo cambia la forma de consultar y redondear datos para que `WithEmpresaVentasPermissions` siga permitiendo la misma operacion de estaciones/carritos cuando el runtime usa PostgreSQL.

2026-04-19: Nota operativa para `estaciones` y `carritos`
- La apertura de una estaciÒ³n debe seguir funcionando para los mismos roles que ya tienen acceso operativo al mÒ³dulo, incluso si el carrito enlazado proviene de datos legado. La resoluciÒ³n del carrito por referencia o nombre no cambia permisos ni alcance por `empresa_id`; solo evita recreaciones fallidas del carrito base.

# Matriz base de roles y permisos POS multiempresa

Fecha de actualizacion: 2026-04-17
Alcance: punto 3 del plan maestro (permisos y seguridad)

- Actualizacion 2026-04-19 (super: plantillas de email y guardado unificado de configuracion avanzada):
	- `super_administrador` agrega la capacidad operativa de editar plantillas reales desde `/super/formato_para_emviar_email.html` y persistirlas por `GET/PUT /super/api/config/email_templates`.
	- Las plantillas controlan correos administrativos, usuarios de empresa, pagos de licencia, recuperaciÒ³n de contraseÒ±a y alertas de reinicio, pero no cambian el modelo de permisos ni abren acceso a otros roles.
	- `web/super/configuracion_avanzada.html` deja de guardar por tarjetas separadas y usa un botÒ³n global arriba y abajo de la vista para persistir Wompi, Epayco, Gmail e IA dentro del mismo mÒ³dulo ya protegido.
	- Impacto de matriz: sin rutas abiertas al pÒºblico ni nuevos wrappers; se fortalece la operabilidad del panel `super` bajo el mismo rol exclusivo existente.

- Actualizacion 2026-04-19 (portal publico y selector: visibilidad por alcance real):
	- El CTA `Probar Gratis` de la landing descriptiva ya no debe conducir a pantallas administrativas como `administrar_empresa`, `super_administrador` o `seleccionar_empresa`; cuando la tarjeta original apunta a un flujo protegido, el destino pÒºblico correcto pasa a ser el registro de administrador.
	- En `seleccionar_empresa`, `Licencias` puede seguir visible para cuentas con gestiÒ³n propia de empresas/licencias, pero `Administradores` y `Reportes globales` quedan reservados a `super_administrador` principal.
	- Los administradores delegados dejan de ver navegaciÒ³n global aunque mantengan rol heredado `super_administrador`, alineando la UI con el alcance efectivo ya impuesto por backend.
	- Impacto de matriz: no cambian endpoints ni wrappers; se endurece la visibilidad operativa del panel y la coherencia del portal pÒºblico.

- Actualizacion 2026-04-19 (super/licencias: cierre de alcance delegado y publicaciÒ³n pÒºblica de Epayco):
	- `super_administrador` delegado ya no obtiene acceso global implÒ­cito por el nombre del rol; si fue creado por otro administrador, su alcance efectivo queda limitado al conjunto de empresas del administrador principal resuelto por cadena de creaciÒ³n.
	- `super_administrador` principal conserva visibilidad global, mientras el delegado solo administra empresas dentro de su portafolio autorizado y recibe `403` al consultar empresas externas.
	- `GET /api/public/licencias/payment_methods` puede anunciar `epayco` cuando existe `public_key`, sin convertir ese anuncio en permiso para ejecutar cobros sin credenciales privadas en los pasos internos del checkout.
	- Impacto de matriz: no se abren rutas ni roles nuevos; se endurece el control real del mÒ³dulo crÒ­tico de empresas y se aclara la diferencia entre visibilidad pÒºblica de mÒ©todo y autorizaciÒ³n de cobro.

## Regla de mantenimiento por modulo

- Cuando se cree un modulo nuevo o se modifique uno existente, esta matriz debe actualizarse en la misma iteracion para reflejar permisos por rol/modulo/accion y el impacto en paginas del panel.
- Esta actualizacion debe quedar sincronizada con `documentos/descripcion_de_modulos`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios` y `CHANGELOG.md`.

- Actualizacion 2026-04-18 (carritos/estaciones: carrito unificado configurable):
	- `web/administrar_empresa/carrito_de_compras.html` pasa a ser la unica UI operativa del carrito para empresa y estaciones; los bloques visibles del formulario, cliente, impuestos, lector y cobro se controlan por configuracion persistida en `estaciones_config`.
	- `web/administrar_empresa/configuracion_de_estaciones.html` y `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html` administran la configuracion del carrito unificado por estacion y por empresa, sin introducir endpoints nuevos ni cambiar wrappers existentes.
	- `web/administrar_empresa/estaciones.html` abre siempre `carrito_de_compras.html`; `ventas_simple.html` queda solo como compatibilidad de redireccion para URLs legacy.
	- Impacto de matriz: sin cambios de roles ni wrappers; el ajuste es de operacion y UX dentro del mismo alcance autenticado del modulo critico `carritos`.

- Actualizacion 2026-04-18 (chat con IA: simplificacion visual) Ã¢â‚¬â€ **obsoleta** (ver 2026-04-24 layout Gemini): en su momento se oculto selector e historial; la revision posterior los restituye en sidebar y topbar sin tocar permisos.

- Actualizacion 2026-04-24 (super pagina principal: tema en iframe y mp-card en modo claro):
	- `web/super/pagina_principal.html` sincroniza tema al cargar; `web/estilos.css` ajusta `body.super-page`, titulo de cabecera y el contenedor del editor (`pp-main-card`, evitando gradientes rosados de `mp-card`) para contraste en tema claro y oscuro.
	- Impacto de matriz: sin cambios en rutas `/super/api/pagina_principal` ni roles; solo UX.

- Actualizacion 2026-04-24 (chat IA: tema y sugerencias):
	- Las pantallas de chat IA sincronizan tema con el panel (`pcs_theme` / `localStorage`) y dejan de mostrar sugerencias tipo pill; los estilos compartidos en `web/estilos.css` mejoran legibilidad en modo claro tambien dentro de `chat_y_tareas`.
	- Impacto de matriz: sin cambios de permisos, rutas protegidas ni wrappers; solo UX y apariencia.

- Actualizacion 2026-04-24 (chat IA: layout tipo Gemini y mensajes de limite):
	- `web/administrar_empresa/chat_con_inteligencia_artificial.html` y `web/super/chat_con_ia_global.html` muestran sidebar (conversaciones locales + historial API), selector y resumen de modelo/uso, compartir respuesta y banner explicativo ante 429 o bloqueo; sin cambiar wrappers ni rutas.
	- Impacto de matriz: sin cambios de permisos, rutas protegidas ni wrappers; solo UX, persistencia local opcional y mensajes al usuario.

- Actualizacion 2026-04-19 (chat IA Gemini-only):
	- `super_administrador` mantiene la facultad exclusiva de activar o desactivar el servicio IA y el proveedor `google` desde `ConfiguraciÒ³n avanzada`.
	- `administrador` de empresa y `super_administrador` consumidor del chat ya no eligen entre proveedores; ambos usan exclusivamente `google:gemini-2.0-flash` cuando el servicio estÒ¡ habilitado.
	- Impacto de matriz: sin cambios en wrappers o alcance por `empresa_id`; se reduce la superficie funcional del mÒ³dulo IA al proveedor Gemini.

- Actualizacion 2026-04-18 (inventario/productos: compras con vista dedicada dentro del submodulo):
	- `web/administrar_empresa/administrar_productos.html` agrega `view=compras` para aislar compras preventivas, consolidado por proveedor y ciclo de orden del frente de inventario puro.
	- Desde 2026-05-14 `Compras` del nucleo de productos se abre directo con `web/administrar_empresa/administrar_productos.html?view=compras`; se elimina el wrapper de navegacion sin cambiar permisos ni endpoints.
	- Impacto de matriz: sin cambios en permisos, roles o wrappers; la segmentacion es solo de UX dentro del mismo alcance autenticado del modulo inventario/productos.

- Actualizacion 2026-04-18 (inventario/productos: proveedores y precios con vistas dedicadas):
	- `web/administrar_empresa/administrar_productos.html` separa el CRUD de proveedores y el historial de cambios de precio de la vista principal de productos usando `view=proveedores` y `view=precios`.
	- Desde 2026-05-14 proveedores y precios se abren directo con `web/administrar_empresa/administrar_productos.html?view=proveedores|precios`; se eliminan wrappers sin crear rutas backend ni duplicar logica CRUD.
	- Impacto de matriz: sin cambios en permisos, roles o wrappers; la segmentacion es solo de UX dentro del mismo alcance autenticado del modulo inventario/productos.

- Actualizacion 2026-04-18 (chat IA super/empresa: resiliencia PostgreSQL legacy y timeout operativo de Ambis):
	- `backend/db/chat_inteligencia_artificial.go` autorrepara el esquema `empresa_ai_*` y `super_ai_*` cuando una instalacion heredada llega con tablas o columnas faltantes, sin abrir endpoints nuevos ni alterar wrappers.
	- `backend/handlers/chat_con_inteligencia_artificial_controller.go` amplÒ­a el timeout usado solo por `ollama:ambis` para soportar respuestas lentas del modelo local/VPS, sin cambiar permisos de acceso ni catÒ¡logo por rol.
	- Impacto de matriz: sin cambios en roles, permisos, wrappers o visibilidad; el ajuste solo mejora estabilidad operativa del mismo modulo IA ya autorizado.

- Actualizacion 2026-04-18 (portal publico: solo queda el emulador N64):
	- `web/menu.js` cambia la entrada flotante de `Juegos` a `Emulador N64` apuntando directo a `/Juegos/n64/index.html`.
	- `web/Juegos/menu_juegos.html` deja de listar juegos y queda como puerta de entrada secundaria con un unico CTA al emulador.
	- `backend/handlers/auth_users_carritos_test.go` actualiza la verificacion publica para `/Juegos/n64/index.html` en lugar de rutas de juegos retirados.
	- Impacto de matriz: sin cambios en roles ni wrappers; la superficie publica `/Juegos/*` queda reducida al emulador N64.

- Actualizacion 2026-04-18 (inventario/productos: vistas separadas para bodegas y categorias):
	- `web/administrar_empresa/administrar_productos.html` concentra la experiencia del modulo y expone tres vistas por query string: `productos`, `bodegas` y `categorias`.
	- Desde 2026-05-14 productos, bodegas y categorias se consumen desde `web/administrar_empresa/administrar_productos.html?view=productos|bodegas|categorias`; se retiran wrappers sin crear rutas nuevas ni duplicar logica CRUD.
	- Impacto de matriz: sin cambios en permisos, roles o wrappers; la segmentacion es solo de UX dentro del mismo alcance autenticado del modulo inventario/productos.

- Actualizacion 2026-04-18 (licencias: activacion gratis unica por empresa y checkout con total cero):
	- `backend/main.go` expone `GET /api/public/licencias/checkout_summary` como apoyo al checkout publico de licencias, sin abrir privilegios nuevos ni exigir sesion.
	- `backend/handlers/payments_handlers.go` mantiene `POST /licencias/activar_sin_pago`, pero ahora solo lo permite si el total final es cero y si la empresa no habia usado ya esa licencia gratis.
	- `web/elegir_licencia.html` y `web/pagar_licencia.html` cambian el CTA visible a `Activar licencia` cuando el valor final queda en cero, sin alterar la matriz de roles; la restriccion efectiva sigue siendo comercial y multiempresa, no de autenticacion.
	- Impacto de matriz: sin cambios de roles o wrappers; se refuerza una regla funcional del checkout publico para impedir reutilizacion gratuita de la misma licencia por la misma empresa.

- Actualizacion 2026-04-18 (chat IA super/empresa: interruptor global de servicio):
	- `web/super/configuracion_avanzada.html` y `backend/handlers/ai_config_handlers.go` agregan el interruptor global `ai.global.enabled` sin abrir rutas nuevas ni ampliar privilegios.
	- `/api/empresa/chat_con_inteligencia_artificial/*` y `/super/api/chat_con_ia_global/*` conservan los mismos wrappers y controles de acceso, pero ahora rechazan el uso cuando la IA global estÒ¡ desactivada desde super.

- Actualizacion 2026-04-18 (chat IA super/empresa: control por proveedor):
	- `super_administrador` puede habilitar o deshabilitar por separado `DeepSeek Chat` y `Ambis Local` desde `ConfiguraciÒ³n avanzada`.
	- `administrador` de empresa solo puede usar los proveedores que el panel super mantenga habilitados; no puede reactivar proveedores desde empresa.
	- Impacto de matriz: sin rutas nuevas ni ampliaciÒ³n de roles; se endurece el control operativo del catÒ¡logo IA desde super.
	- Impacto de matriz: sin cambios en roles, wrappers ni visibilidad base; solo se aÒ±ade una compuerta operativa global administrada por `super_administrador`.

- Actualizacion 2026-04-18 (chat IA super/empresa: aviso visual y prueba operativa):
	- `web/administrar_empresa/chat_con_inteligencia_artificial.html` y `web/super/chat_con_ia_global.html` comunican explÒ­citamente el estado `IA desactivada` sin cambiar permisos efectivos.
	- `web/super/configuracion_avanzada.html` expone el botÒ³n `Probar IA contra VPS` como acciÒ³n operativa de diagnÒ³stico para `super_administrador`.
	- Impacto de matriz: sin cambios en roles ni wrappers; mejora solo la claridad operativa del panel y de los chats.

- Actualizacion 2026-04-18 (gobernanza tecnica documental: integraciones externas y reconciliacion documental):
	- `documentos/gobernanza_tecnica/contratos/contrato_integraciones_bancarias_y_conectores_externos.md` formaliza `estado`, `health_check`, `sync_manual`, `rotar_credencial` y `monitoreo` sin alterar wrappers ni permisos efectivos.
	- `documentos/gobernanza_tecnica/runbooks/runbook_reconciliacion_documental_fiscal_y_contable_externa.md` fija el diagnostico operativo entre compras, facturacion, reintentos fiscales y repositorio/versionado, sin introducir cambios en roles.
	- Impacto de matriz: sin cambios en permisos o visibilidad; el alcance es documental y operativo transversal.

- Actualizacion 2026-04-18 (gobernanza tecnica documental: repositorio documental y firmas externas):
	- `documentos/gobernanza_tecnica/contratos/contrato_repositorio_documental_y_firmas_externas.md` formaliza el comportamiento de `/api/empresa/documentos/gestion` y `/api/empresa/documentos/firmas`, incluyendo `acceso`, `repositorio`, `versiones`, `versionar` y herencia de permisos desde el modulo documental.
	- `documentos/gobernanza_tecnica/runbooks/runbook_versionado_documental_y_firmas_externas.md` fija el procedimiento reproducible para diagnosticar accesos denegados, historial incompleto, firmas huÒ©rfanas y versiones no marcadas como historicas.
	- Impacto de matriz: sin cambios en permisos, wrappers o visibilidad; el alcance es documental y operativo transversal sobre reglas ya existentes.

- Actualizacion 2026-04-18 (gobernanza tecnica documental: reconciliacion y evidencia regulatoria endurecida):
	- Los contratos de repositorio documental, interoperabilidad y reportes ahora exigen reconciliar exportes regulatorios con la versiÒ³n documental vigente y la firma asociada cuando exista.
	- Los runbooks documentales dejan explÒ­cito que `include_denegados=1` solo sirve para diagnÒ³stico y que un exporte no otorga acceso ni sustituye evidencia firmada.
	- Impacto de matriz: sin cambios en permisos, wrappers o visibilidad; el ajuste solo endurece reglas de trazabilidad y uso operativo de la evidencia.

- Actualizacion 2026-04-18 (gobernanza tecnica documental: checklist rapida para QA/soporte):
	- `documentos/gobernanza_tecnica/runbooks/checklist_evidencia_documental_para_qa_y_soporte.md` resume el orden de validacion para `empresa_id`, rol, version vigente, firma y exporte documental antes de escalar incidentes.
	- Impacto de matriz: sin cambios en permisos, wrappers o visibilidad; la checklist solo acelera diagnostico operativo sobre reglas ya vigentes.

- Actualizacion 2026-04-18 (gobernanza tecnica documental: interoperabilidad documental e integraciones externas):
	- `documentos/gobernanza_tecnica/contratos/contrato_interoperabilidad_documental_contable_y_fiscal_externa.md` formaliza el comportamiento de compras documentales, facturacion documental, versionado/acceso del repositorio y reconciliacion fiscal, sin alterar wrappers ni roles vigentes.
	- `documentos/gobernanza_tecnica/runbooks/runbook_contingencias_integraciones_bancarias_y_conectores.md` fija el procedimiento operativo para `/api/empresa/integraciones/apis` y `/api/empresa/integraciones/bancos`, sin introducir cambios en permisos funcionales.
	- Impacto de matriz: sin cambios en permisos, wrappers o visibilidad de paginas; el alcance es documental y operativo transversal.

- Actualizacion 2026-04-18 (gobernanza tecnica documental: cierre de periodo y conciliacion bancaria):
	- `documentos/gobernanza_tecnica/contratos/contrato_conciliacion_bancaria_y_cierre_periodo_contable.md` formaliza el comportamiento de `/api/empresa/finanzas/periodos`, `/api/empresa/finanzas/movimientos` y la conciliacion bancaria asociada, sin alterar wrappers ni roles vigentes.
	- `documentos/gobernanza_tecnica/runbooks/runbook_cierre_periodo_y_conciliacion_bancaria.md` fija un procedimiento operativo reproducible para bloqueos por periodo cerrado, importacion de extractos y estados de conciliacion.
	- Impacto de matriz: sin cambios en permisos, wrappers o visibilidad de paginas; el alcance es documental y operativo transversal.

- Actualizacion 2026-04-18 (gobernanza tecnica documental: soporte remoto y contingencia de reportes):
	- `documentos/gobernanza_tecnica/contratos/contrato_soporte_remoto_por_empresa_y_mesa_tecnica_central.md` formaliza el comportamiento de `/api/empresa/soporte_remoto`, `/api/public/soporte_remoto` y `/super/api/soporte_remoto` sin cambiar wrappers ni roles vigentes.
	- `documentos/gobernanza_tecnica/runbooks/runbook_reportes_programados_y_exportaciones_contables.md` y `documentos/gobernanza_tecnica/runbooks/runbook_soporte_remoto_sesiones_y_dispositivos.md` fijan diagnÒ³stico operativo reproducible para exportes/reportes y para sesiones/dispositivos remotos.
	- Impacto de matriz: sin cambios en permisos, wrappers o visibilidad de pÒ¡ginas; el alcance es documental y operativo transversal.

- Actualizacion 2026-04-18 (estaciones: tarjeta especial YouTube):
	- `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/youtube_station_browser.html` y `web/estilos.css` agregan una estacion especial `YouTube` dentro del mismo modulo autenticado de estaciones.
	- La estacion especial reproduce solo videos o playlists embebibles vÒ¡lidos; cuando la referencia configurada es texto libre, el fallback sigue siendo abrir YouTube fuera del sistema y no cambia permisos ni alcance del modulo.
	- La operadora tambiÒ©n puede pegar y guardar desde la propia tarjeta la URL o ID del video, playlist o `Shorts` sin salir de `estaciones.html`; el cambio sigue usando el mismo permiso autenticado del modulo y no abre capacidades nuevas por rol.
	- `web/descripcion_de_los_sistemas.ht` adopta el mismo lenguaje visual de las tarjetas del index, pero el modulo sigue siendo publico y sin cambios en permisos, CRUD/A ni wrappers.
	- Impacto de matriz: sin cambios en permisos, roles o wrappers; la nueva tarjeta reutiliza la misma autorizacion empresarial y no abre endpoints ni acciones administrativas adicionales.

- Actualizacion 2026-04-18 (gobernanza tecnica documental: DIAN, alertas de reinicio y reportes):
	- `documentos/gobernanza_tecnica/runbooks/runbook_dian_set_pruebas_y_diagnostico_oficial.md` documenta el alcance real del soporte DIAN de empresa y evita tratar la base operativa actual como integracion oficial completa.
	- `documentos/gobernanza_tecnica/runbooks/runbook_alertas_reinicio_y_monitoreo_gmail_smtp.md` fija la operacion de `POST /super/api/config/gmailaction=test`, `super_servidor_eventos`, `gmail.restart_alert_to` y `gmail.smtp_test_mode`.
	- `documentos/gobernanza_tecnica/contratos/contrato_reportes_contables_financieros_y_exportacion_multiformato.md` formaliza el comportamiento del modulo de reportes empresariales y globales super sin alterar wrappers ni roles.
	- Impacto de matriz: sin cambios en permisos, wrappers o visibilidad de paginas; el alcance es documental y operativo transversal.

- Actualizacion 2026-04-18 (gobernanza tecnica documental: paquete base):
	- `documentos/README.md` y `documentos/gobernanza_tecnica/*` agregan la capa de ADRs, contratos tecnicos, runbooks y cambio seguro para el repositorio.
	- Impacto de matriz: sin cambios en permisos, wrappers o visibilidad de paginas; el alcance es transversal y documental.

- Actualizacion 2026-04-18 (gobernanza interna: orquestacion del equipo de agentes):
	- `.github/agents/agente_go.agent.md` pasa a definir a `agente_go` como director del equipo y entrada por defecto del repositorio, coordinando a `agente_backend_db`, `agente_frontend_ux` y `agente_qa_operacion`.
	- `.github/agents/agente_backend_db.agent.md`, `.github/agents/agente_frontend_ux.agent.md`, `.github/agents/agente_qa_operacion.agent.md` y `.github/agents/README.md` formalizan responsabilidades internas de desarrollo, UX y validacion operativa.
	- Impacto de matriz: sin cambios en permisos funcionales, wrappers ni visibilidad de paginas para usuarios del sistema; el alcance es interno al equipo tecnico del repositorio.

- Actualizacion 2026-04-18 (gobernanza interna: protocolo de delegacion y plantilla de ejecucion):
	- `.github/agents/protocolo_delegacion.md` y `.github/agents/plantilla_trabajo_por_modulo.md` agregan reglas internas para que `agente_go` active especialistas segun modulo e impacto tecnico.
	- Los archivos de agentes quedan especializados por modulos como `pagos`, `licencias`, `venta_publica`, `facturacion electronica`, `DIAN`, `estaciones`, `carritos`, `autenticacion`, `reportes` y `paneles`.
	- Impacto de matriz: sin cambios en permisos funcionales, wrappers ni visibilidad de paginas; el alcance sigue siendo interno a la disciplina del equipo tecnico.

- Actualizacion 2026-04-18 (gobernanza interna: tabla rapida y cierre obligatorio en modulos criticos):
	- `.github/agents/protocolo_delegacion.md` agrega una tabla rapida por modulo y ejemplos reales para reducir ambiguedad al activar especialistas.
	- `.github/agents/agente_go.agent.md` exige participacion obligatoria de varios especialistas en modulos criticos antes de cerrar.
	- Impacto de matriz: sin cambios en permisos funcionales, wrappers ni visibilidad de paginas; la mejora sigue siendo interna al equipo tecnico.

- Actualizacion 2026-04-18 (gobernanza interna: semaforo y evidencia minima de cierre):
	- `.github/agents/protocolo_delegacion.md`, `.github/agents/plantilla_trabajo_por_modulo.md` y los archivos de agentes especialistas endurecen la clasificacion rapida y rechazan cierres sin evidencia minima.
	- Impacto de matriz: sin cambios en permisos funcionales, wrappers ni visibilidad de paginas; la mejora sigue siendo interna a la disciplina del equipo tecnico.

- Actualizacion 2026-04-18 (checkout Epayco: correo de activacion recuperable e idempotente):
	- `backend/handlers/payments_handlers.go` reintenta el correo de activacion en aprobados posteriores cuando la licencia ya quedÒ³ activa pero la notificacion aun no se habia confirmado, y marca el envio dentro del `raw_payload` para no duplicarlo.
	- Impacto de matriz: sin cambios en roles ni wrappers; `/epayco/*` sigue siendo un flujo publico de checkout y la mejora solo fortalece la entrega del correo transaccional.

- Actualizacion 2026-04-18 (checkout de licencias: validacion de contexto esperado por empresa/licencia):
	- `backend/handlers/payments_handlers.go`, `backend/db/db.go`, `backend/handlers/payments_handlers_test.go` y `web/pagar_licencia.html` endurecen la conciliacion publica para que `/epayco/transaction_status` y `/wompi/transaction_status` comparen el pago resuelto contra el `empresa_id` y `licencia_id` de la pagina abierta.
	- El correo de activacion resuelve ahora la empresa por `id` fisico o por `empresa_id` logico, manteniendo el aislamiento funcional por empresa incluso cuando la tabla `empresas` evoluciono con ids distintos del alcance operativo.
	- Impacto de matriz: sin cambios en permisos ni wrappers; la mejora solo refuerza el aislamiento multiempresa del checkout publico.

- Actualizacion 2026-04-18 (estaciones: retiro del circulo inferior de estado):
	- `web/administrar_empresa/estaciones.html` y `web/estilos.css` simplifican la tarjeta de estaciones para dejar un solo indicador visual: el cuadrito superior derecho que luego reflejara el estado del sensor.
	- Impacto de matriz: sin cambios en permisos, roles o wrappers; la vista de estaciones mantiene el mismo acceso autenticado y las mismas acciones.

- Actualizacion 2026-04-18 (ventas simples por estacion: boton de regreso a estaciones):
	- `web/administrar_empresa/ventas_simple.html` y `web/js/ventas_simple.js` agregan un retorno explicito hacia `administrar_empresa/estaciones.html` para la operacion por estacion.
	- Impacto de matriz: sin cambios en permisos, roles o wrappers; el submodulo conserva el mismo acceso autenticado y solo mejora la navegacion interna.

- Actualizacion 2026-04-18 (ventas simples por estacion: variante `carrito_compacto`):
	- `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/ventas_simple.html` y `web/js/ventas_simple.js` agregan una presentacion compacta del mismo carrito por estacion, activada por `variant=compacto` y compatible con un flag opcional de configuracion remota si el backend lo expone.
	- Impacto de matriz: sin cambios en permisos, roles o wrappers; la variante reutiliza la misma autorizacion empresarial, los mismos endpoints de carrito/items y el mismo aislamiento por `empresa_id`.

- Actualizacion 2026-04-18 (selector de empresas: tarjetas compactas con botonera al pie):
	- `web/estilos.css` reduce el tamano visual de las tarjetas de `seleccionar_empresa.html` y fija la botonera inferior centrada al pie de cada bloque, sin tocar rutas, wrappers ni acciones disponibles.
	- Impacto de matriz: sin cambios en permisos, roles o alcance; el selector mantiene la misma operacion autenticada.

- Actualizacion 2026-04-18 (arcade publico: N64 vertical mobile para ROM legal del usuario):
	- `web/Juegos/n64/index.html`, `web/Juegos/n64/styles.css` y `web/Juegos/n64/n64-wrapper.js` agregan una pÒ¡gina pÒºblica especÒ­fica para mÒ³vil con controles tÒ¡ctiles, ROM legal persistida en IndexedDB y respaldo local de la memoria del cartucho.
	- `web/Juegos/menu_juegos.html` publica la entrada del nuevo juego en el lobby general del arcade.
	- Impacto de matriz: sin cambios en roles, wrappers ni permisos; `Juegos` continÒºa siendo una superficie pÒºblica sin autenticaciÒ³n bajo `/Juegos/*`.

- Actualizacion 2026-04-18 (super configuracion avanzada: prueba real de Gmail):
	- `backend/handlers/usuarios_empresa.go` agrega `POST /super/api/config/gmailaction=test` para enviar un correo de prueba real con la configuracion SMTP ya guardada, y `web/super/configuracion_avanzada.html` lo invoca desde el boton `Probar Gmail`.
	- Impacto de matriz: sin cambios en roles ni wrappers; la accion sigue reservada al mismo modulo super protegido de configuracion avanzada.

- Actualizacion 2026-04-17 (arcade publico: Brigada burbujas 3D plus mejora su UX movil):
	- `web/Juegos/brigada_burbujas_3d_plus.html` incorpora apuntado tactil sobre el canvas, ayudas visuales en pantalla y un layout mas comodo en vertical para completar el flujo movil del juego.
	- Impacto de matriz: sin cambios en roles, wrappers ni permisos; `Juegos` sigue siendo una superficie publica sin autenticacion bajo `/Juegos/*`.

- Actualizacion 2026-04-17 (arcade publico: Brigada burbujas 3D plus suma campaÒ±a y rivales de pasarela caricaturesca):
	- `web/Juegos/brigada_burbujas_3d_plus.html` amplÒ­a el juego con transformaciones especiales y rivales adultas estilizadas tipo pasarela, sin desnudez ni contenido explicito, manteniendo la experiencia dentro de un tono arcade apto para portal publico.
	- Impacto de matriz: sin cambios en roles, wrappers ni permisos; `Juegos` sigue siendo una superficie publica sin autenticacion bajo `/Juegos/*`.

- Actualizacion 2026-04-18 (arcade publico: Brigada burbujas 3D plus agrega arsenal y sectores mixtos):
	- `web/Juegos/brigada_burbujas_3d_plus.html` incorpora arsenal con tres armas, pickups de salud/municion, HUD de arma/sector y una IA que patrulla, busca y convoca refuerzos, manteniendo el mismo acceso publico del arcade.
	- Impacto de matriz: sin cambios en roles, wrappers ni permisos; `Juegos` sigue siendo una superficie publica sin autenticacion bajo `/Juegos/*`.

- Actualizacion 2026-04-18 (arcade publico: Brigada burbujas 3D plus asegura combate movil en escenario):
	- `web/Juegos/brigada_burbujas_3d_plus.html` mueve a una barra tactica interna del escenario el cambio rapido de arma y la pausa para que el flujo movil no dependa de scroll ni de tarjetas externas.
	- Impacto de matriz: sin cambios en roles, wrappers ni permisos; `Juegos` sigue siendo una superficie publica sin autenticacion bajo `/Juegos/*`.

- Actualizacion 2026-04-18 (arcade publico: Brigada burbujas 3D plus agrega joystick y fullscreen movil):
	- `web/Juegos/brigada_burbujas_3d_plus.html` incorpora joystick tactil de movimiento libre, pantalla completa guiada y HUD interno reordenado para una mano, manteniendo intacto el acceso publico del juego.
	- Impacto de matriz: sin cambios en roles, wrappers ni permisos; `Juegos` sigue siendo una superficie publica sin autenticacion bajo `/Juegos/*`.

- Actualizacion 2026-04-18 (arcade publico: Brigada burbujas 3D plus agrega ajustes tactiles persistentes):
	- `web/Juegos/brigada_burbujas_3d_plus.html` aÒ±ade auto-disparo opcional, vibracion y ajustes de sensibilidad persistidos en `localStorage`, sin alterar el modelo de acceso del arcade.
	- Impacto de matriz: sin cambios en roles, wrappers ni permisos; `Juegos` sigue siendo una superficie publica sin autenticacion bajo `/Juegos/*`.

- Actualizacion 2026-04-18 (arcade publico: Brigada burbujas 3D plus refina HUD y ayuda de mira movil):
	- `web/Juegos/brigada_burbujas_3d_plus.html` agrega un boton visible de auto-disparo en el HUD y una asistencia suave de apuntado configurable solo para la experiencia tactil.
	- Impacto de matriz: sin cambios en roles, wrappers ni permisos; `Juegos` sigue siendo una superficie publica sin autenticacion bajo `/Juegos/*`.

- Actualizacion 2026-04-18 (arcade publico: Brigada burbujas 3D plus activa preset facil por defecto):
	- `web/Juegos/brigada_burbujas_3d_plus.html` migra configuraciones tactiles antiguas para arrancar con auto-disparo activo y ayuda de mira reforzada en celular.
	- Impacto de matriz: sin cambios en roles, wrappers ni permisos; `Juegos` sigue siendo una superficie publica sin autenticacion bajo `/Juegos/*`.

- Actualizacion 2026-04-17 (super seguridad: vista en modo oscuro):
	- `web/super/seguridad.html` pasa a usar una paleta oscura dentro del mismo modulo de seguridad VPS Linux del panel super.
	- Impacto de matriz: sin cambios en roles ni permisos; `Seguridad VPS Linux (super)` mantiene el mismo alcance exclusivo de `super_administrador`.

- Actualizacion 2026-04-17 (menus administrativos: ocultar/mostrar en movil):
	- `web/administrar_empresa.html`, `web/super_administrador.html` y `web/seleccionar_empresa.html` agregan un boton final para contraer o expandir el menu lateral solo en movil.
	- `web/menu.js` y `web/estilos.css` mantienen el sidebar completo en escritorio y, en moviles, dejan visible solo el boton de recuperacion cuando el usuario colapsa el menu.
	- Impacto de matriz: sin cambios en roles, wrappers ni permisos; solo se ajusta la experiencia responsive sobre paginas ya autorizadas.

- Actualizacion 2026-04-18 (submenu configuracion: ocultar/mostrar en movil):
	- `web/administrar_empresa/configuracion_menu.html` adopta el mismo wrapper `admin-sidebar-mobile-collapsible` y el mismo boton final de `Ocultar menÒº` / `Mostrar menÒº` del shell administrativo principal.
	- `web/menu.js` reaprovecha la misma logica de colapso sin abrir nuevas rutas ni modificar accesos del submodulo de configuracion.
	- Impacto de matriz: sin cambios en roles, wrappers ni permisos; solo se amplÒ­a el mismo patron responsive a otra vista ya autenticada.

- Actualizacion 2026-04-18 (submenu configuracion: permisos reales y guardado real de integraciones):
	- `web/administrar_empresa/configuracion_permisos.html` deja de fingir alta/guardado de roles y pasa a consumir `GET /api/empresa/permisos_contextoempresa_id=...&include_matrix=1`, mostrando solo informacion permitida por el wrapper de seguridad existente.
	- `web/administrar_empresa/configuracion_integraciones.html` queda informativa; Wompi/Epayco y la publicacion comercial por paginas se administran desde `web/administrar_empresa/venta_publica.html`, dentro del mismo alcance autorizado del modulo de venta publica por empresa.
	- Impacto de matriz: sin cambios en roles ni wrappers; `Permisos` queda como consulta de solo lectura y `Venta publica` reaprovecha el permiso autenticado ya existente sobre `venta_publica`.

- Actualizacion 2026-04-18 (submenu configuracion: persistencia real del bloque general):
	- `web/administrar_empresa/configuracion.html` reemplaza el guardado local del bloque `Productos y pedidos` por `GET/PUT /api/empresa/configuracion_generalempresa_id=...`.
	- `backend/handlers/empresa_configuracion_general.go` publica la ruta bajo `WithEmpresaSeguridadPermissions`, por lo que mantiene el mismo alcance autenticado de configuracion empresarial sin abrir permisos adicionales.
	- Impacto de matriz: sin cambios en roles; solo se sustituye persistencia local por backend real dentro del wrapper de seguridad existente.

- Actualizacion 2026-04-17 (ventas: selector de documento por empresa):
	- `web/administrar_empresa/configuracion.html` agrega el selector `Documento al vender` y `backend/handlers/carritos_compras.go` lo aplica al cierre de `pagar_estacion` para emitir `factura_electronica` o `comprobante_pago`.
	- `web/administrar_empresa/facturas_electronicas.html` amplÒ­a la consulta del historial para incluir comprobantes dentro del mismo modulo documental empresarial.
	- Impacto de matriz: sin ampliacion de privilegios; el cambio reutiliza permisos existentes de configuracion de empresa, carritos y consulta documental bajo el mismo `empresa_id` autenticado.

- Actualizacion 2026-04-17 (editar empresa: comprar licencia si esta vencida):
	- `web/editar_empresa.html` agrega un CTA comercial `Comprar licencia`, pero solo se expone cuando la empresa tiene licencias vencidas y ya no cuenta con una licencia activa vigente.
	- `web/js/editar_empresa.js` resuelve esa condicion usando la lectura existente de `/super/api/licenciasscope=mine&con_empresa=1`, sin requerir nuevos permisos ni nuevas rutas super.
	- Impacto de matriz: sin ampliacion de privilegios; el rol mantiene las mismas lecturas y el CTA solo redirige al flujo comercial permitido de `elegir_licencia.html`.

- Actualizacion 2026-04-17 (editar empresa y eliminacion total confirmada):
	- `web/editar_empresa.html` y `web/js/editar_empresa.js` habilitan una vista autenticada para editar `nombre` y `descripcion` de la empresa seleccionada y para ejecutar su cierre total solo si el usuario escribe el nombre exacto como confirmacion.
	- `backend/utils/utils.go` amplÒ­a la excepcion del rol `administrador` sobre `/super/api/empresas` para permitir `PUT` y `DELETE`, manteniendo `GET` como unica accion disponible en `/super/api/tipos_empresas` y `/super/api/licencias`.
	- Impacto de matriz: `administrador` pasa a tener `R/U/D` solo sobre sus empresas en el flujo del selector; `POST /super/api/empresas` y el resto de `/super/*` continÒºan reservados a `super_administrador`.

- Actualizacion 2026-04-17 (selector de empresas: editar sale de la tarjeta y pasa al menu):
	- `web/seleccionar_empresa.html` agrega la entrada lateral `Editar empresa` y `web/js/seleccionar_empresa.js` usa la empresa activa del contexto para abrir `editar_empresa.html` sin necesitar un boton extra dentro de cada tarjeta.
	- Las tarjetas conservan la accion principal de entrar a la empresa o de comprar licencia, y solo mantienen el boton cuadrado de descarga como accion secundaria del bloque.
	- Impacto de matriz: sin cambios en permisos efectivos; el cambio es de UX y no abre nuevas rutas ni nuevas mutaciones.

- Actualizacion 2026-04-17 (registro de contrasena Google: CTA unico):
	- `web/registrar_contrasena_usuario_de_google.html` elimina el enlace `Continuar` y deja solo el boton `Guardar contrasena` centrado en el formulario.
	- Impacto de matriz: sin cambios en roles ni wrappers; el flujo sigue siendo parte de autenticacion administrativa y solo ajusta la UX del paso obligatorio posterior a Google.

- Actualizacion 2026-04-17 (descarga de informacion empresarial: misma ruta funcional y dark mode):
	- `backend/main.go` agrega el alias `/descargar_informacion_de_la_empresa` hacia la vista HTML real, y `web/js/descargar_informacion_de_la_empresa.js` mantiene la descarga en la misma pantalla con `fetch + blob` para PDF, XLS, CSV, JSON y TXT.
	- Impacto de matriz: sin cambios de rol; sigue siendo una capacidad autenticada del flujo super/administrador dentro del alcance permitido sobre la empresa seleccionada.

- Actualizacion 2026-04-17 (selector de empresas: operacion minima permitida para administrador):
	- `backend/utils/utils.go` habilita para el rol `administrador` `GET/POST` sobre `/super/api/empresas` y mantiene `GET` sobre `/super/api/tipos_empresas` y `/super/api/licencias`, porque son los endpoints minimos que consume `seleccionar_empresa.html` para listar y crear empresas propias.
	- Impacto de matriz: el rol `administrador` no obtiene acceso global al panel `/super/*`; solo recupera la operacion minima del selector. `super_administrador` mantiene acceso total y los demas roles siguen en `403`.

- Actualizacion 2026-04-17 (checkout Epayco: pantalla de exito y retorno con referencia real):
	- `web/epayco/respuesta.html`, `web/pagar_licencia.html` y la nueva `web/epayco/pago_exitoso.html` reorganizan solo la UX del retorno aprobado para que el usuario salga del checkout hacia una confirmacion profesional y no quede atrapado en el formulario con estados ambiguos.
	- Impacto de matriz: sin cambios en roles ni permisos; `/epayco/*` sigue siendo superficie publica del checkout comercial y `seleccionar_empresa.html` conserva el mismo acceso autenticado posterior.

- Actualizacion 2026-04-17 (seleccionar empresa: correccion de render del listado):
	- `web/js/seleccionar_empresa.js` recupera el helper `escapeHtml` que usan las tarjetas del panel para mostrar nombre y observaciones sin romper el render inicial con `escapeHtml is not defined`.
	- Impacto de matriz: sin cambios en roles, permisos ni wrappers; la vista `seleccionar_empresa.html` conserva el mismo alcance del panel super y solo se corrige un fallo de frontend.

- Actualizacion 2026-04-17 (pagina principal publica: CTA inferior fijo y texto contrastado):
	- `web/estilos.css` fija visualmente el CTA `Explorar oferta` al pie de cada tarjeta del home y da mayor contraste al titulo y descripcion usando tipografia clara con iluminado exterior negro suave, sin paneles de fondo sobre el contenido textual.
	- Impacto de matriz: sin cambios en permisos, roles ni wrappers; `index.html` sigue siendo una pagina publica de solo consulta con los mismos accesos visibles.

- Actualizacion 2026-04-17 (venta publica por empresa: pasarelas propias Wompi/Epayco):
	- `web/administrar_empresa/configuracion.html` aÒ±ade una seccion de `Pasarelas de pago` y `web/administrar_empresa/venta_publica.html` conserva la ediciÒ³n detallada de la misma configuraciÒ³n empresarial.
	- `backend/handlers/venta_publica.go` y `backend/handlers/payments_handlers.go` mantienen el alcance por `empresa_id`: la empresa autenticada solo administra sus propias credenciales, mientras el checkout pÒºblico solo recibe flags sanitizados y nunca secretos.
	- Impacto de matriz: sin nuevos roles; el permiso sigue dentro del mÒ³dulo de administraciÒ³n de empresa, y la tienda pÒºblica conserva acceso abierto solo a catÒ¡logo, creaciÒ³n de pago y consulta de estado.

- Actualizacion 2026-04-17 (autenticacion administrativa: super restringido al correo reservado):
	- `backend/handlers/auth_admin_handlers.go` hace que el registro administrativo publico, el login por correo y el callback Google dejen las cuentas nuevas en rol `administrador` por defecto.
	- `backend/utils/utils.go` deja de promocionar cuentas legacy/autoregistradas a `super_administrador` y aplica la misma politica al validar acceso a `/super/*`.
	- `powerfulcontrolsystem@gmail.com` queda como unico correo reservado para conservar `super_administrador` dentro del flujo publico de autenticacion administrativa.
	- Impacto de matriz: el login/registro administrativo sigue siendo publico, pero la visibilidad y las acciones `CRUA` del panel super quedan restringidas al correo reservado; las cuentas nuevas ya no ganan permisos globales por autoregistro ni por OAuth.

- Actualizacion 2026-04-17 (seleccionar empresa: descarga empresarial en pagina dedicada):
	- `web/js/seleccionar_empresa.js` deja de abrir el modal de backup y navega a `descargar_informacion_de_la_empresa.html`, mientras `backend/handlers/system_empresas_handlers.go` atiende `resumen_descarga` y `exportar_informacion` sobre la misma ruta protegida `/super/api/empresas`.
	- `web/descargar_informacion_de_la_empresa.html` y `web/js/descargar_informacion_de_la_empresa.js` habilitan exportacion ejecutiva y operativa en `PDF`, `XLS`, `CSV`, `JSON` y `TXT` para la empresa seleccionada.
	- Impacto de matriz: sin cambios en roles; la lectura y exportacion siguen siendo alcance `R` del panel super para el administrador autenticado con filtro de empresa ya validado por `/super/api/empresas`.

- Actualizacion 2026-04-17 (checkout publico de licencias: tarjetas resumen y panel Epayco compacto):
	- `web/pagar_licencia.html` reorganiza la experiencia publica en tarjetas visuales para `licencia` y `codigos de descuento`, elimina el catalogo visual de medios Epayco y deja el panel de pago en una tarjeta mas compacta sin campo de correo visible.
	- `web/estilos.css` ajusta solo la presentacion del flujo, incluyendo una tarjeta blanca centrada para Epayco y conservando intacto el alcance publico del checkout.
	- Impacto de matriz: sin cambios en permisos, roles, wrappers ni visibilidad; las mismas rutas publicas de checkout conservan su alcance actual y el panel super mantiene la administracion exclusiva de configuracion/licencias.

- Actualizacion 2026-04-17 (navegacion general: misma pestaÒ±a por defecto):
	- Se retiran aperturas automÒ¡ticas en nueva ventana para navegaciÒ³n normal entre mÒ³dulos, portales pÒºblicos, ayudas y exportes comunes del sistema.
	- Los reportes/exportes de `Clientes`, `Asistencia`, `Backups`, `Tarifas por dÒ­a` y `Soporte remoto` descargan el archivo sin sacar al usuario del mÒ³dulo actual.
	- Se conservan como excepciÒ³n los documentos legales (`contrato`, tÒ©rminos de pasarela) y los popups tÒ©cnicos de impresiÒ³n o vista previa documental.
	- Impacto de matriz: sin cambios en roles ni permisos; solo cambia el comportamiento de navegaciÒ³n de rutas ya permitidas.

- Actualizacion 2026-04-17 (checkout publico de licencias: Epayco activa licencia y limpia rechazos finales):
	- `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go` y `web/pagar_licencia.html` refuerzan el post-pago de Epayco para activar la licencia con fallback por `invoice`, enviar correo al cliente al quedar activa y detener el estado `pending` cuando el retorno ya es final rechazado o fallido.
	- Impacto de matriz: sin cambios en permisos; `/epayco/create_transaction`, `/epayco/transaction_status`, `/epayco/webhook` y `/epayco/respuesta.html` conservan alcance publico del checkout comercial, mientras la gestion de licencias y configuraciones sigue restringida a `super_administrador`.

- Actualizacion 2026-04-17 (licencias publicas: tarjetas compactas sin estado visible):
	- `web/elegir_licencia.html` deja las tarjetas de licencias con una presentacion mas corta y elimina textos operativos de estado/asignacion antes del pago.
	- Impacto de matriz: sin cambios en permisos; el modulo mantiene las mismas rutas y acciones disponibles para el flujo de compra de licencias.

- Actualizacion 2026-04-17 (licencias publicas: orden ascendente por valor):
	- `web/elegir_licencia.html` ordena las licencias visibles desde el menor valor hasta el mayor antes de renderizar el grid de compra.
	- Impacto de matriz: sin cambios en permisos; solo se modifica el criterio visual de orden del mismo listado disponible para el flujo de compra.

- Actualizacion 2026-04-17 (seleccionar empresa: boton de descarga solo con icono):
	- `web/js/seleccionar_empresa.js` y `web/estilos.css` cambian la presentacion del boton de descarga de la tarjeta para dejarlo como icono blanco con tooltip explicativo.
	- Impacto de matriz: sin cambios en permisos; el boton conserva la misma accion disponible dentro del flujo del panel super.

- Actualizacion 2026-04-17 (licencias super: valor 0 visible y editable):
	- `web/super/licencias.html` conserva el valor `0` en listado y formulario de edicion para el CRUD de licencias.
	- Impacto de matriz: sin cambios en permisos; solo corrige el comportamiento visual del modulo `Licencias` para `super_administrador`.

- Actualizacion 2026-04-17 (licencias del selector: historial y estado con vencimiento):
	- `backend/db/db.go` y `web/super/licencias.html` convierten la vista `scope=mine&con_empresa=1` en un historial de licencias pagadas/asignadas, con fecha de vencimiento, estados `activa/por vencer/vencida` y CTA de renovacion sin opciones de eliminar en esa pantalla.
	- `backend/handlers/payments_handlers_test.go` valida que el endpoint protegido siga filtrando por creador y entregue empresa + fechas al frontend.
	- Impacto de matriz: sin cambios en roles ni permisos; el modulo `Licencias` mantiene acceso de lectura/gestion para `super_administrador`, pero en el flujo del selector la misma ruta se presenta como historial restringido al alcance del administrador autenticado.

- Actualizacion 2026-04-17 (checkout publico de licencias: Epayco usa Smart Checkout v2):
	- `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html` y `web/super/configuracion_avanzada.html` migran Epayco desde el flujo manual `checkout.php` al Smart Checkout oficial con `sessionId` y `checkout-v2.js`.
	- Impacto de matriz: sin cambios en permisos; `/epayco/create_transaction`, `/epayco/transaction_status`, `/epayco/webhook` y `/epayco/respuesta.html` mantienen alcance publico dentro del checkout comercial, mientras la configuracion avanzada sigue siendo exclusiva de `super_administrador`.

- Actualizacion 2026-04-17 (crear clave por correo: visibilidad de contrasena):
	- `web/registrar_contrasena_usuario_de_google.html`, `web/js/registrar_contrasena_usuario_de_google.js` y `web/estilos.css` agregan el control visual para mostrar u ocultar la contrasena antes de guardarla.
	- Impacto de matriz: sin cambios en permisos; sigue siendo el mismo flujo autenticado de autogestion posterior al login con Google.

- Actualizacion 2026-04-17 (licencias publicas: tarjetas con estilo del home):
	- `web/elegir_licencia.html` reutiliza la estructura visual de `index.html` para las tarjetas de licencias, sin cambiar rutas, protecciones ni acciones disponibles.
	- Impacto de matriz: sin cambios en permisos; la vista y el flujo de compra conservan el mismo alcance previo.

- Actualizacion 2026-04-17 (reportes globales super: una empresa o varias):
	- `web/super/reportes_globales.html` y `web/js/super_reportes_globales.js` agregan un selector de alcance para alternar entre analisis de una sola empresa o de varias empresas del mismo administrador.
	- `backend/handlers/reportes_globales_test.go` cubre el uso de `empresa_id` en la misma API protegida.
	- Impacto de matriz: sin cambios en permisos; `Reportes globales (super)` se mantiene como permiso `R` exclusivo de `super_administrador`.

- Actualizacion 2026-04-17 (autenticacion administrativa: login en una sola tarjeta visual):
	- `web/login.html` y `web/estilos.css` integran el formulario por correo dentro de la misma tarjeta del acceso con Google, retirando el recuadro secundario del flujo publico de administradores.
	- Impacto de matriz: sin cambios en permisos, roles, wrappers o visibilidad; el login administrativo sigue siendo publico y el panel super sigue reservado a `super_administrador`.

- Actualizacion 2026-04-19 (autenticacion administrativa: recordar usuario por correo):
	- `web/login.html` agrega nuevamente la casilla `Recordar usuario` y `web/js/login.js` guarda solo el correo administrativo en persistencia local cuando el propio usuario lo solicita.
	- Impacto de matriz: sin cambios en permisos, roles, wrappers ni visibilidad; el login administrativo sigue siendo publico y el panel super mantiene el mismo alcance exclusivo para `super_administrador`.

- Actualizacion 2026-04-16 (reportes globales super por administrador creador):
	- `backend/handlers/reportes_globales.go` expone `/super/api/reportes_globales` filtrando empresas por `usuario_creador = admin autenticado`.
	- `web/super/reportes_globales.html` y `web/js/super_reportes_globales.js` permiten ver datasets consolidados o separados por empresa solo dentro del panel super.
	- Impacto de matriz: el modulo `Reportes globales (super)` queda con permiso `R` exclusivo de `super_administrador`.

- Actualizacion 2026-04-17 (portal publico: arcade movil con runtime comun de poderes y premios):
	- `web/Juegos/arcade_shared.js` y `web/Juegos/arcade_window.css` pasan a ser la base comun del arcade publico para countdown, sonido, records, poderes y premios en todos los juegos activos.
	- Los nueve juegos `*_plus.html` del lobby reutilizan el mismo runtime sin ampliar rutas privadas ni introducir permisos nuevos.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` mantiene acceso publico y de solo uso.

- Actualizacion 2026-04-18 (portal publico: Brigada burbujas 3D plus):
	- `web/Juegos/brigada_burbujas_3d_plus.html` agrega una nueva ruta publica del arcade con shooter retro 3D simulado, controles tactiles y record local por slug `brigada_burbujas_3d`.
	- `web/Juegos/menu_juegos.html` publica el decimo juego activo del lobby y `web/img/juegos/brigada_burbujas_3d.svg` incorpora la portada dedicada.
	- Impacto de matriz: sin cambios en permisos; `Portal publico - Juegos` se mantiene publico, sin autenticacion y sin CRUD administrativo.

- Actualizacion 2026-04-17 (reportes globales super: graficos y lectura ejecutiva):
	- `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js` y `web/estilos.css` agregan visualizaciones ejecutivas sobre el mismo modulo protegido de lectura.
	- Impacto de matriz: sin cambios en permisos; `Reportes globales (super)` se mantiene como `R` exclusivo de `super_administrador`.

- Actualizacion 2026-04-16 (facturacion electronica: estabilidad de pruebas automatizadas):
	- `backend/db/finanzas_test.go` fuerza el dialecto `motor_legado_retirado` en `openFinanzasTestDB` para evitar que la suite del modulo herede configuracion `postgres` del entorno local y falle por compatibilidad SQL durante pruebas de esquema y documentos transaccionales.
	- Impacto de matriz: sin cambios en permisos, roles, wrappers o visibilidad del modulo `facturacion electronica`; el ajuste solo endurece la validacion automatizada del backend.

- Actualizacion 2026-04-16 (portal publico: menu flotante navegable en movil):
	- `web/menu.js` conserva la navegacion tactil de las rutas publicas al cerrar el panel solo en `click`.
	- `web/estilos.css` mejora la respuesta tactil del boton y de los items del menu.
	- Impacto de matriz: sin cambios en permisos o roles; solo se recupera la usabilidad movil de enlaces publicos ya visibles.

- Actualizacion 2026-04-16 (menu flotante publico: separacion frente a botones de cabecera):
	- `web/menu.js` marca las paginas que reciben el menu flotante con `has-floating-menu` y `web/estilos.css` reserva espacio en encabezados y barras de acciones compartidas para que el toggle no quede encima de botones cercanos.
	- Impacto de matriz: sin cambios en permisos, CRUD/A, wrappers o visibilidad por rol; la correccion es solo de layout compartido.

- Actualizacion 2026-04-16 (seleccionar empresa: formato clasico restaurado):
	- `web/js/seleccionar_empresa.js` recupera el render clasico de tarjetas del panel super sin alterar rutas ni acciones del modulo.
	- Impacto de matriz: sin cambios en permisos, roles o wrappers; solo se restituye una presentacion visual previa del selector de empresas.

- Actualizacion 2026-04-17 (portal publico de usuarios de empresa con contrato y subdominio propio):
	- `backend/handlers/usuarios_empresa.go` exige aceptacion del contrato vigente antes de login, primer password, reset o cambio de contrasena; la restriccion aplica al mismo usuario autenticandose, no a un rol administrativo nuevo.
	- `web/login_usuario.html` y `web/js/login_usuario.js` pasan a ser una puerta publica de acceso por invitacion, mientras `web/administrar_empresa.html` y `web/js/administrar_empresa.js` resuelven el enlace correcto por empresa usando `empresa_slug` o `dominio_publico` de la configuracion publica.
	- Impacto de matriz: sin cambios en CRUD/A ni en wrappers empresariales; el portal publico no amplÒ­a privilegios y la visibilidad final del panel sigue determinada por rol/permisos_contexto ya existentes.

- Actualizacion 2026-04-16 (estaciones: carrito base sincronizado desde backend):
	- `backend/handlers/empresa_estacion_prefs.go` sincroniza automaticamente los carritos enlazados al guardar `estaciones_config`, y `backend/db/empresa_estacion_prefs.go` asegura la regla `una estacion -> un carrito base` por `empresa_id`.
	- `backend/handlers/empresa_estacion_prefs_test.go` y `backend/db/empresa_estacion_prefs_test.go` cubren la creacion y actualizacion del enlace sin depender de la pagina de configuracion.
	- Impacto de matriz: sin cambios en roles ni permisos; la capacidad sigue dentro del alcance actual de configuracion/ventas empresariales y no abre nuevas rutas ni acciones a otros roles.

- Actualizacion 2026-04-16 (portal publico: contacto debajo del grid del home):
	- `web/index.html` mueve `Informacion de contacto` debajo de `portalCardsGrid` y `web/estilos.css` lo centra sin alterar accesos publicos ni rutas protegidas.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; el ajuste es visual dentro del portal publico.

- Actualizacion 2026-04-17 (soporte remoto: limites por plan y mesa tecnica super):
	- `backend/handlers/soporte_remoto.go` mantiene el alcance empresarial en modulo `seguridad`, pero ahora devuelve `uso` y bloquea sesiones/dispositivos cuando la empresa supera los topes configurados.
	- `backend/handlers/super_soporte_remoto.go` expone `/super/api/soporte_remoto` y `web/super/soporte_remoto.html` agrega una mesa tecnica central solo para `super_administrador`.
	- Impacto de matriz: `linkSoporteRemoto` sigue requiriendo accion `A` sobre `seguridad` en panel empresa; el nuevo panel super de soporte remoto es exclusivo de `super_administrador` y no amplÒ­a permisos de roles empresariales.

- Actualizacion 2026-04-20 (soporte remoto: portal pÒºblico RustDesk y configuraciÒ³n desde super):
	- `backend/handlers/soporte_remoto.go` y `backend/handlers/super_soporte_remoto.go` mantienen el mismo wrapper empresarial/super, pero ahora la mesa tÒ©cnica super tambiÒ©n puede editar la configuraciÒ³n pÒºblica por empresa vÒ­a `action=config`.
	- `web/administrar_empresa/soporte_remoto.html` y `web/super/soporte_remoto.html` muestran descargas de cliente/servidor y portal pÒºblico sin ampliar los permisos del resto de mÒ³dulos.
	- Impacto de matriz: `linkSoporteRemoto` sigue exigiendo acciÒ³n `A` sobre `seguridad` en empresa; el portal pÒºblico `soporte_remoto_acceso.html` es libre por token de sesiÒ³n y no agrega privilegios permanentes; la ediciÒ³n central sigue reservada a `super_administrador`.

- Actualizacion 2026-04-20.2 (soporte remoto: tope diario RustDesk):
	- La ediciÒ³n del nuevo campo `max_minutos_dia_rustdesk` se concentra en `/super/api/soporte_remotoaction=config` y su vista super asociada.
	- Impacto de matriz: no se amplÒ­an permisos empresariales; el control del tope diario sigue reservado a `super_administrador` y solo afecta la creaciÒ³n/aprobaciÒ³n de sesiones RustDesk para la empresa configurada.

- Actualizacion 2026-04-16 (deploy VPS: limpieza de procesos previos del backend):
	- `scripts/sync_to_vps.ps1` limpia procesos previos asociados a `backend/bin/server_linux_amd64` antes del arranque y genera una unidad `systemd` sin el warning de clave invalida en `Service`.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad; el ajuste es operativo de infraestructura.

- Actualizacion 2026-04-16 (checkout publico de licencias: alias `sambox` en Epayco):
	- `backend/handlers/payments_handlers.go` normaliza `epayco.mode=sambox` como `sandbox` antes de construir el checkout publico.
	- Impacto de matriz: sin cambios en roles ni permisos; `/epayco/create_transaction` conserva el mismo alcance publico.

- Actualizacion 2026-04-16 (portal publico: arcade activo con ocho juegos):
	- `web/Juegos/menu_juegos.html` publica ocho juegos activos y fija popup uniforme `700x700` sin barras en escritorio, manteniendo apertura directa en movil.
	- `web/Juegos/arcade_window.css` y los ocho juegos `*_plus.html` mantienen una experiencia publica homogÒ©nea con pausa real, records locales y nombre de jugador compartido.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` sigue siendo de lectura/uso para todos los roles y tambien sin autenticacion.

- Actualizacion 2026-04-17 (portal publico: nuevo Ajedrez 3D plus):
	- `web/Juegos/ajedrez_3d_plus.html` agrega una nueva ruta publica del arcade con tablero en perspectiva 3D simulada y selector de cinco dificultades.
	- `web/Juegos/menu_juegos.html` publica la nueva tarjeta del lobby y `web/img/juegos/ajedrez_3d.svg` suma la portada visual del titulo.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` mantiene acceso publico y de solo uso.

- Actualizacion 2026-04-16 (checkout publico de licencias: metodo unico y compatibilidad Epayco legacy):
	- `web/pagar_licencia.html` omite el selector de forma de pago cuando solo hay una pasarela disponible y entra directo al panel correspondiente.
	- `backend/handlers/payments_handlers.go` aÒ±ade `p_key` al checkout de Epayco cuando existe `epayco.private_key`, manteniendo el mismo alcance publico de `/epayco/*` y `/api/public/licencias/payment_methods`.
	- Impacto de matriz: sin cambios en roles ni permisos; el ajuste es funcional en checkout publico.

- Actualizacion 2026-04-16 (checkout publico de licencias: Epayco sin popup intermedio):
	- `web/pagar_licencia.html` ya no deja el pago en una pestaÒ±a emergente; ahora redirige la misma pestaÒ±a al checkout de Epayco y reutiliza `/epayco/respuesta.html` para volver con contexto a la pantalla de licencia.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `/epayco/*` sigue siendo publico y de solo consumo para el flujo comercial.

- Actualizacion 2026-04-16 (home publico: ajuste visual de accesos superiores):
	- `index.html` mantiene exactamente la misma visibilidad publica; el cambio solo compacta y centra los botones superiores en movil.
	- Impacto de matriz: sin cambios en permisos o acceso por rol.

- Actualizacion 2026-04-16 (licencias super: persistencia del valor):
	- `super/licencias.html` y `/super/api/licencias` mantienen los mismos permisos y alcance, pero ahora exponen correctamente los errores de guardado al administrador.
	- Impacto de matriz: sin cambios en permisos, roles ni visibilidad; se corrige comportamiento del CRUD super existente.

- Actualizacion 2026-04-16 (seleccionar empresa: mejora visual de tarjetas):
	- `seleccionar_empresa.html` mantiene el mismo acceso por rol y licencia, pero las tarjetas internas ahora se adaptan mejor al contenido variable.
	- Impacto de matriz: sin cambios en permisos, CRUD/A ni visibilidad de modulo; la correccion es exclusivamente visual y de lectura.

- Actualizacion 2026-04-16 (pagina principal dinamica: correccion de render en editor super):
	- `web/super/pagina_principal.html` respeta la cantidad persistida de tarjetas al recargar el editor.
	- Impacto de matriz: sin cambios en permisos, roles ni visibilidad de modulo; la correccion solo evita un recorte incorrecto en el panel super.

- Actualizacion 2026-04-16 (pagina principal dinamica: sin parpadeo inicial del campo cantidad):
	- `web/super/pagina_principal.html` deja el campo `Cantidad de tarjetas` en estado de carga hasta recibir la configuracion real, evitando mostrar de forma transitoria el valor `5` que no corresponde a la configuracion activa.
	- El mismo editor sincroniza la cantidad con el numero real de tarjetas persistidas, manteniendo alineacion con `index.html` y `/descripcion_de_los_sistemas.ht`.
	- Impacto de matriz: sin cambios en permisos, CRUD/A, wrappers o visibilidad por rol; `Pagina principal (tarjetas index)` sigue siendo CRUA exclusivo de `super_administrador`.

- Actualizacion 2026-04-16 (infraestructura publica: wildcard HTTPS y subdominio venta digital):
	- `venta-digital.powerfulcontrolsystem.com` queda publicado como acceso publico HTTPS hacia la pagina global `venta_digital.html`.
	- La raiz generica de subdominios por empresa sigue en `venta_publica.html`; no cambia la matriz de permisos ni la visibilidad de modulos internos.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni permisos; la modificacion solo amplÒ­a una entrada publica servida por infraestructura.

- Actualizacion 2026-04-16 (autenticacion administrativa: registro con pais y ciudad):
	- `web/registrar_nuevo_usuario_administrador.html` y `web/js/registrar_nuevo_usuario_administrador.js` amplian el formulario de alta administrativa para capturar `pais` y `ciudad`.
	- `backend/handlers/auth_admin_handlers.go` mantiene la regla de confirmacion obligatoria del correo antes del ingreso, y `backend/db/db.go` agrega persistencia evolutiva de esos datos en `administradores`.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad de modulos; el ajuste afecta solo los datos del registro y la completitud del perfil administrativo.

- Actualizacion 2026-04-16 (autenticacion administrativa: esquema `administradores` compatible con PostgreSQL):
	- `backend/db/db.go` agrega una regularizacion reusable de columnas de seguridad de `administradores` y `backend/main.go` la invoca en el arranque del backend.
	- `backend/db/administradores_auth_schema_test.go` cubre la reparacion del esquema y la escritura de password inicial sobre tablas incompletas.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad; el ajuste es de compatibilidad de backend para el mismo flujo administrativo autenticado.

- Actualizacion 2026-04-16 (super: seguridad VPS Linux):
	- `backend/handlers/security_vps_handlers.go` expone la API protegida `/super/api/security/vps/config|run|status|history|report|compare` y `backend/main.go` registra el servicio central del modulo.
	- `web/super/seguridad.html` y `web/js/super_seguridad.js` agregan la vista operativa para configurar, ejecutar, revisar historial, comparar y exportar reportes del VPS.
	- `backend/tools/vps_security_scan/main.go` y los scripts Linux asociados permiten ejecutar el mismo modulo desde consola y programarlo por cron sin abrir acceso a otros roles.
	- Impacto de matriz: se agrega `Seguridad VPS Linux (super)` como modulo `CRUA` exclusivo de `super_administrador`; no hay acceso para roles empresariales.

- Actualizacion 2026-04-16 (portal publico: orden visual del header):
	- `web/index.html` y `web/estilos.css` reordenan el header del home para dejar `Informacion de contacto` al extremo derecho de la misma fila.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; el ajuste es visual y no altera autorizacion.

- Actualizacion 2026-04-16 (autenticacion estable multi-host sin recordar usuario/cuenta):
	- `web/login.html` y `web/login_usuario.html` eliminan los checkboxes de `Recordar cuenta` y `Recordar usuario`, reduciendo divergencias por almacenamiento local entre `localhost`, dominio raÒ­z y `www`.
	- `backend/handlers/auth_admin_handlers.go` deja de propagar `login_hint` en el inicio OAuth; el login Google arranca limpio y consistente.
	- `web/menu.js`, `web/js/super_administrador.js`, `web/js/seleccionar_empresa.js`, `web/super/licencias.html` y `web/super/tipos_empresas.html` retiran lÒ³gica `remember*` y conservan solo seÒ±al de sesiÒ³n para navegaciÒ³n/autenticaciÒ³n visible.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; la modificaciÒ³n es operativa/UX y no amplÒ­a privilegios.

- Actualizacion 2026-04-16 (autenticacion administrativa: registro separado y recuperacion guiada):
	- `web/login.html` mantiene acceso publico por Google o correo/clave, pero mueve el registro administrativo a `/registrar_nuevo_usuario_administrador.html` y deja la recuperaciÒ³n en formularios propios dentro del login.
	- `backend/handlers/auth_admin_handlers.go` endurece el alta y la recuperaciÒ³n de administradores, mientras `backend/utils/utils.go` libera `/registrar_nuevo_usuario_administrador.html` y `/auth/confirmar_admin` como rutas pÒºblicas reales.
	- `backend/handlers/auth_admin_handlers_test.go` y `backend/handlers/auth_users_carritos_test.go` cubren el alta/login/reset administrativo y la nueva superficie pÒºblica del middleware.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; el login/registro/confirmaciÒ³n administrativa sigue siendo pÒºblico y la administraciÒ³n global continÒºa bajo `super_administrador`.

- Actualizacion 2026-04-16 (autenticacion administrativa: creaciÒ³n de clave local tras Google):
	- `backend/handlers/auth_admin_handlers.go` y `backend/handlers/accept_handlers.go` redirigen a `/registrar_contrasena_usuario_de_google.html` cuando la cuenta autenticada por Google todavÒ­a no tiene `password_set`.
	- `backend/handlers/account_handlers.go` expone `/api/account/set_google_password` como endpoint autenticado de solo autoservicio para el administrador en sesiÒ³n.
	- `web/registrar_contrasena_usuario_de_google.html` completa el alta de contraseÒ±a local sin ampliar permisos ni abrir una nueva superficie pÒºblica.
	- Impacto de matriz: sin cambios en roles, CRUD/A o wrappers; la capacidad sigue restringida al mismo administrador autenticado sobre su propia cuenta.

- Actualizacion 2026-04-16 (Epayco: respuesta publica fija):
	- `web/epayco/respuesta.html` queda disponible como pagina publica para retorno desde la pasarela.
	- `backend/handlers/payments_handlers.go` usa `/epayco/respuesta.html` como `response` y `/epayco/webhook` como `confirmation` para licencias.
	- Impacto de matriz: sin cambios en roles ni permisos internos; ambas rutas siguen siendo publicas por integracion de pasarela.

- Actualizacion 2026-04-16 (checkout visual y seleccion de empresa):
	- `web/pagar_licencia.html` preselecciona de forma visible la unica pasarela disponible y muestra el logo de Epayco en tarjeta y panel cuando corresponde.
	- `web/js/seleccionar_empresa.js` vuelve al formato compacto previo para tarjetas de empresa, sin alterar accesos ni permisos.
	- Impacto de matriz: sin cambios en roles, CRUD/A o visibilidad funcional; las pantallas siguen con el mismo alcance de acceso que antes.

- Actualizacion 2026-04-16 (super: tamano estimado por empresa en administracion PostgreSQL):
	- `web/super/administrar_base_de_datos.html` agrega una tarjeta operativa con el boton `Cargar Empresas` para consultar consumo estimado por empresa dentro de `pcs_empresas`.
	- `backend/handlers/postgres_performance.go` extiende `/super/api/postgres/performance` con `action=empresas_storage`, manteniendo el endpoint protegido y de solo lectura.
	- `backend/handlers/postgres_performance_test.go` cubre la accion invalida del panel y utilidades asociadas.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad; `Administracion DB PostgreSQL (super)` sigue siendo lectura exclusiva de `super_administrador`.

## Roles base

| Rol | Alcance | Descripcion |
|---|---|---|
| super_administrador | global | administra configuracion, empresas, licencias, auditoria y seguridad global |
| admin_empresa | empresa | administra configuracion, catalogos, usuarios y cierres de su empresa |
| supervisor_sucursal | sucursal | supervisa operacion, aprueba cierres y movimientos criticos |
| cajero | sucursal/caja | registra ventas, cobros, devoluciones permitidas y cierre de caja |
| inventario | sucursal/bodega | gestiona productos, existencias y movimientos de bodega |
| compras | empresa/sucursal | crea ordenes de compra, recepciones y ajustes de costo |
| contabilidad | empresa | valida asientos, periodos y reportes financieros |
| auditor | empresa/global | consulta reportes, logs y trazabilidad sin modificar datos |

## Permisos por modulo

Leyenda:
- C: crear
- R: leer
- U: actualizar
- D: eliminar/anular
- A: aprobar/cerrar

| Modulo | super_administrador | admin_empresa | supervisor_sucursal | cajero | inventario | compras | contabilidad | auditor |
|---|---|---|---|---|---|---|---|---|
| Ventas POS | CRUDA | CRUA | CRUA | CRU | R | R | R | R |
| Inventarios | CRUDA | CRUA | CRUA | R | CRUDA | R | R | R |
| Clientes | CRUDA | CRUA | CRUA | CRU | R | R | R | R |
| Proveedores | CRUDA | CRUA | R | R | R | CRUA | R | R |
| Compras | CRUDA | CRUA | CRUA | R | R | CRUDA | R | R |
| Facturacion electronica | CRUDA | CRUA | R | CRU | R | R | R | R |
| Contabilidad y periodos | CRUDA | CRUA | R | R | R | R | CRUDA | R |
| Reportes financieros | CRUA | CRUA | R | R | R | R | CRUA | R |
| Cierres de caja | CRUDA | CRUA | CRUA | CRUA | R | R | R | R |
| Seguridad y permisos | CRUDA | CRUA | R | R | R | R | R | R |
| Impresoras operativas | CRUDA | CRUA | R | R | R | R | R | R |
| Seguridad VPS Linux (super) | CRUA | - | - | - | - | - | - | - |
| Explorador de Archivos (super) | R | - | - | - | - | - | - | - |
| Administracion DB PostgreSQL (super) | R | - | - | - | - | - | - | - |
| Reportes globales (super) | R | - | - | - | - | - | - | - |
| Pagina principal (tarjetas index) | CRUA | - | - | - | - | - | - | - |
| Portal publico - Emulador N64 | R | R | R | R | R | R | R | R |
| Contrato administrativo (super) | CRUA | - | - | - | - | - | - | - |
| Monitor de errores del sistema (super) | R | - | - | - | - | - | - | - |
| Pasarelas de licencias (Wompi/Epayco) | CRUA | - | - | - | - | - | - | - |

## Estado de implementacion tecnica inicial (2026-04-04)

- Actualizacion 2026-04-16 (super: seguridad VPS Linux):
	- `web/super/seguridad.html` amplÒ­a el monitor de seguridad del panel super para cubrir configuracion, ejecucion de escaneo, hallazgos, historial, comparacion y exportes del VPS.
	- `backend/handlers/security_vps_handlers.go` y `backend/vpssecurity/*` mantienen el modulo encapsulado y protegido solo para `super_administrador`.
	- `backend/tools/vps_security_scan/main.go` junto a los scripts Linux permiten operacion manual y por cron sin ampliar privilegios a otros roles.
	- Impacto de matriz: nuevo modulo `Seguridad VPS Linux (super)` con `CRUA` exclusivo de `super_administrador`; sin cambios para roles de empresa.

- Actualizacion 2026-04-16 (portal publico: boton de contacto al extremo derecho del home):
	- `web/index.html` y `web/estilos.css` ajustan solo la composicion visual del header comercial.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol.

- Actualizacion 2026-04-16 (autenticacion administrativa: registro separado y recuperacion guiada):
	- `web/login.html` centra el acceso por correo, deja debajo `Registrarse` y `Ã‚Â¿OlvidÃƒÂ³ su contraseÃƒÂ±a`, y sustituye los `prompt()` por formularios reales para recuperaciÃƒÂ³n y restablecimiento.
	- `web/registrar_nuevo_usuario_administrador.html` agrega una superficie pÒºblica especÒ­fica para alta administrativa y `backend/utils/utils.go` la libera junto con `/auth/confirmar_admin`.
	- `backend/handlers/auth_admin_handlers.go` evita sobrescribir cuentas confirmadas y exige `nombre`, `telefono` y contraseÒ±a mÒ­nima para el registro administrativo.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; el ajuste corrige el flujo pÒºblico de autenticaciÒ³n administrativa sin ampliar permisos.

- Actualizacion 2026-04-16 (super: tamano estimado por empresa en administracion PostgreSQL):
	- `web/super/administrar_base_de_datos.html` suma una lectura puntual del peso estimado por empresa en la base operativa compartida y la presenta ordenada de mayor a menor.
	- `backend/handlers/postgres_performance.go` mantiene el mismo modulo y endpoint protegido, agregando solo una accion de solo lectura (`action=empresas_storage`).
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; la consulta sigue reservada a `super_administrador`.

- Actualizacion 2026-04-16 (portal publico: arcade movil reforzado y countdown en Patito volando):
	- `web/Juegos/arcade_shared.js` suma sonidos de cuenta regresiva reutilizables por el arcade publico.
	- `web/Juegos/patito_volando.html` agrega una cuenta regresiva previa de 5 segundos y `web/Juegos/pollitos_cataplum.html`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html` y `web/Juegos/rebote_bloques.html` ajustan su layout para celular.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` sigue siendo de lectura/uso para todos los roles y tambien sin autenticacion.

- Actualizacion 2026-04-16 (frontend compartido: ajustes base para celular y menu flotante):
	- `web/menu.js` cierra el panel del menu flotante al seleccionar una opcion para descubrir la pagina destino de inmediato en movil.
	- `web/estilos.css` mejora el responsive compartido de tablas, navegacion administrativa, panel del menu flotante y CTA fijo de WhatsApp en `index.html`.
	- `web/login.html` vuelve a cargar la hoja compartida correcta, recuperando estilos y comportamiento responsive del login administrativo.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; el ajuste es transversal de frontend.

- Actualizacion 2026-04-16 (portal publico: botones superiores alineados al CTA de ofertas):
	- `web/estilos.css` hace que `Registrarse o iniciar sesiÒ³n` e `Informacion de contacto` reutilicen el mismo tratamiento visual del boton `Explorar oferta` de las tarjetas del home.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; el ajuste es visual dentro del portal publico.

- Actualizacion 2026-04-16 (checkout de licencias: Epayco sandbox estable en PostgreSQL):
	- `backend/db/db.go` agrega `EnsurePaymentGatewaySchema(...)` para asegurar `pagos_epayco` y `pagos_wompi` en `pcs_superadministrador` cuando el backend corre sobre PostgreSQL.
	- `backend/main.go` ejecuta ese bootstrap al arrancar y `backend/handlers/payments_handlers.go` evita degradar a `ERROR` una referencia Epayco que sigue pendiente pero aun no aparece en la validacion externa.
	- `backend/handlers/payments_handlers_test.go` cubre el nuevo criterio de polling pendiente para checkout publico de licencias.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pasarelas de licencias (Wompi/Epayco)` sigue siendo CRUA exclusivo de `super_administrador` y las rutas publicas mantienen uso de solo consumo.

- Actualizacion 2026-04-16 (portal publico: arcade con perfil compartido y cinco juegos):
	- `web/Juegos/arcade_shared.js` centraliza nombre de jugador, top local y control de sonido para todos los juegos publicos del portal sin introducir autenticacion obligatoria.
	- `web/Juegos/menu_juegos.html` pasa a ser un lobby cuadrado con portadas SVG, resumen de records y tarjetas por juego, y ahora enlaza cinco titulos jugables.
	- `web/Juegos/patito_volando.html` y `web/Juegos/pollitos_cataplum.html` se alinean al nuevo perfil compartido; `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html` y `web/Juegos/rebote_bloques.html` amplian el modulo con nuevas mecanicas.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` sigue siendo de lectura/uso para todos los roles y tambien sin autenticacion.

- Actualizacion 2026-04-15 (alertas de reinicio del servidor: activacion configurable):
	- `web/super/configuracion_avanzada.html` agrega el switch `restart_alert_enabled` para activar o desactivar la notificacion automatica de inicio/reinicio del backend sin borrar `restart_alert_to`.
	- `backend/handlers/server_runtime_notifications.go` mantiene la bitacora y el log local aun cuando la alerta esta apagada; solo omite el envio de correo.
	- `backend/handlers/usuarios_empresa.go` y `backend/handlers/super_config_backup_handlers.go` incorporan el nuevo flag en la API y en el respaldo de configuracion critica.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; la configuracion sigue siendo exclusiva de `super_administrador`.

- Actualizacion 2026-04-15 (selector de empresas: iconografia por tipo y tarjetas mas profesionales):
	- `web/js/seleccionar_empresa.js` asigna icono, tono visual y texto de apoyo a cada tarjeta usando `tipo_nombre`, mejorando la lectura rapida del negocio antes de abrir su panel o elegir licencia.
	- `web/estilos.css` renueva el bloque visual de `seleccionar_empresa.html` con tarjetas mas coloridas, chips de estado, metadata de acceso y CTA mas claro, sin alterar rutas ni wrappers.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni privilegios; la seleccion y administracion de empresas sigue dentro del alcance de `super_administrador`.

- Actualizacion 2026-04-15 (pagina_principal super: cantidad integrada al guardado):
	- `web/super/pagina_principal.html` elimina el paso manual `Aplicar cantidad`; la cantidad ahora se refleja en el editor al cambiar el campo y se persiste con el mismo guardado de configuracion.
	- `backend/handlers/pagina_principal_handlers_test.go` agrega cobertura para una configuracion ampliada de tarjetas y confirma que el backend conserva la cantidad solicitada al volver a cargar.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pagina principal (tarjetas index)` sigue siendo CRUA exclusivo de `super_administrador`.

- Actualizacion 2026-04-15 (portal publico: segundo juego `Pollitos al cataplum`):
	- `web/Juegos/menu_juegos.html` agrega una segunda tarjeta jugable y soporta popup por `slug`, ancho y alto por juego.
	- `web/Juegos/pollitos_cataplum.html` aÒ±ade un juego publico de resortera con control arrastrar/soltar y niveles cortos.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; `Portal publico - Juegos` sigue siendo de lectura/uso para todos los roles y tambien sin autenticacion.

- Actualizacion 2026-04-15 (checkout de licencias: fallback canonico para Epayco/Wompi):
	- `backend/handlers/payments_handlers.go` ya no aborta el checkout solo porque la solicitud llega desde `localhost`; resuelve una base publica valida y cae al dominio canonico del sistema cuando no existe otra opcion publica.
	- `backend/handlers/payments_handlers_test.go` cubre el nuevo fallback canonico y la exclusion de `gmail.confirm_base_url` local.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni wrappers; el ajuste afecta solo la operacion del checkout publico de licencias.

- Actualizacion 2026-04-15 (portal publico: menu de juegos y Patito volando):
	- `web/menu.js` expone la entrada publica `Juegos` dentro del menu flotante compartido y `web/Juegos/menu_juegos.html` sirve como catalogo de tarjetas por juego.
	- `backend/utils/utils.go` deja publico el prefijo `/Juegos/` y `backend/handlers/auth_users_carritos_test.go` valida que el middleware no exija sesion para `menu_juegos.html` ni `patito_volando.html`.
	- `web/Juegos/patito_volando.html` no agrega acciones CRUD/A ni superficies administrativas; es una experiencia publica de solo lectura/uso recreativo.
	- Impacto de matriz: se agrega `Portal publico - Juegos` como lectura/uso accesible para todos los roles, con nota operativa de que tambien queda disponible sin autenticacion por formar parte del portal comercial.

- Actualizacion 2026-04-15 (contrato versionado y editor super):
	- `backend/handlers/super_contrato_handlers.go` incorpora `GET/PUT /super/api/contrato` como superficie exclusiva de `super_administrador` para administrar el contrato vigente y publicar nuevas versiones con trazabilidad.
	- `web/super/contrato.html` agrega una pagina visible solo dentro del panel super para editar el contrato, revisar historial y reutilizar una version previa como base de una nueva publicacion.
	- `backend/handlers/auth_admin_handlers.go` y `backend/handlers/accept_handlers.go` hacen que el acceso administrativo dependa de la `contrato_version_aceptada` frente a la version vigente publicada por super.
	- Impacto de matriz: se agrega el control global `Contrato administrativo (super)` como CRUA exclusivo de `super_administrador`; `/api/public/contrato` y `/contrato.html` quedan de lectura publica para sostener el flujo de aceptacion previa al panel.

- Actualizacion 2026-04-15 (monitor centralizado de errores y recovery global):
	- `backend/utils/system_errors.go` y `backend/utils/utils.go` incorporan captura transversal de errores HTTP, panicos recuperados y procesos internos protegidos, con respuesta amigable al usuario final y detalle tecnico reservado al monitor super.
	- `backend/handlers/super_error_handlers.go` publica `GET /super/api/errores` y `web/super/errores.html` entrega la vista filtrable del sistema completo.
	- Impacto de matriz: se agrega `Monitor de errores del sistema (super)` como lectura exclusiva de `super_administrador`; no se expone a administradores de empresa ni a otros roles.

- Actualizacion 2026-04-15 (checkout de licencias: Epayco visible con Public Key y rutas publicas reales):
	- `backend/handlers/payments_handlers.go` ajusta la disponibilidad de Epayco para el checkout de licencias: con `epayco.enabled=1`, basta `epayco.public_key` para publicar la pasarela y construir el checkout actual.
	- `backend/utils/utils.go` deja realmente publicas `/api/public/licencias/payment_methods`, `/wompi/*` y `/epayco/*`, alineando el middleware con la matriz y con la documentacion del portal de pagos.
	- `web/pagar_licencia.html` mejora el mensaje de indisponibilidad para distinguir entre pasarela desactivada y configuracion incompleta.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad del panel; `Pasarelas de licencias (Wompi/Epayco)` sigue como CRUA exclusivo de `super_administrador`, mientras las rutas publicas de checkout permanecen de solo consumo para el flujo comercial.

- Actualizacion 2026-04-15 (login admin sin hint visible y Gmail SMTP editable en directo):
	- `web/login.html` elimina el bloque visible del correo recordado sin cambiar el flujo de autenticacion Google ni el alcance publico del login.
	- `web/super/configuracion_avanzada.html` habilita edicion directa de la configuracion Gmail existente sobre la misma ruta protegida `/super/api/config/gmail`.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; la configuracion global de correo sigue bajo `super_administrador` y el login administrativo permanece publico.

- Actualizacion 2026-04-15 (portal publico: tamanos configurables en home y landing desde pagina_principal):
	- `backend/handlers/pagina_principal_handlers.go` y `backend/handlers/pagina_principal_handlers_test.go` amplian el contrato de `pagina_principal` para publicar ajustes visuales globales de tamano de tarjeta y tamano de texto.
	- `web/super/pagina_principal.html` permite editar esos ajustes y `web/index.html` junto con `web/descripcion_de_los_sistemas.ht` los consumen desde la API publica para escalar el portal sin rutas adicionales.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pagina principal (tarjetas index)` se mantiene como CRUA exclusivo de `super_administrador` y el portal sigue siendo publico de solo lectura.

- Actualizacion 2026-04-15 (portal publico: CTA superior de WhatsApp y botones tipo mini-tarjeta):
	- `web/index.html` mantiene las mismas rutas publicas del home, pero actualiza la presentacion del header para que los accesos comerciales principales se muestren como mini-tarjetas compactas.
	- `web/estilos.css` reposiciona el CTA flotante `Contactenos` hacia la esquina superior derecha y preserva su comportamiento responsive sin introducir acciones nuevas.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pagina principal (tarjetas index)` se mantiene como CRUA exclusivo de `super_administrador` y el portal sigue siendo publico de solo lectura.

- Actualizacion 2026-04-15 (portal publico: landing descriptiva configurable desde pagina_principal):
	- `backend/handlers/pagina_principal_handlers.go` amplÒ­a la configuracion de tarjetas del portal para incluir el contenido extendido consumido por `/descripcion_de_los_sistemas.ht`.
	- `web/super/pagina_principal.html` agrega campos de edicion para etiqueta, titular ampliado, parrafos y capacidades clave; `web/descripcion_de_los_sistemas.ht` renderiza ese contenido desde la API publica y deja de depender de textos fijos por nombre de tarjeta.
	- `backend/handlers/pagina_principal_handlers_test.go` cubre la normalizacion y exposicion de esos campos ampliados.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pagina principal (tarjetas index)` mantiene CRUA exclusivo de `super_administrador` y la landing descriptiva sigue siendo publica de solo lectura.

- Actualizacion 2026-04-15 (checkout de licencias: retorno recuperable tras Epayco/Wompi):
	- `backend/handlers/payments_handlers.go` devuelve a `web/pagar_licencia.html` con contexto operativo del cobro y permite lookup Wompi por `reference` para reconsultar el estado real despues del redirect.
	- `web/pagar_licencia.html` solo endurece el flujo publico de licencias: guarda el pago pendiente, reanuda polling al volver y muestra feedback claro sin crear pantallas administrativas ni acciones nuevas.
	- `backend/handlers/payments_handlers_test.go` cubre la recuperacion por referencia y la URL de retorno enriquecida del checkout.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad por rol; `Pasarelas de licencias (Wompi/Epayco)` sigue siendo CRUA exclusivo de `super_administrador` y el checkout continua siendo publico de solo consumo.

- Actualizacion 2026-04-15 (fix Epayco: llave pÒºblica correcta y callbacks con dominio pÒºblico):
	- `backend/handlers/payments_handlers.go` corrige el contrato de Epayco para separar `public_key`, `private_key` y `customer_id`, ademÒ¡s de reutilizar una base pÒºblica vÒ¡lida en los callbacks de Epayco/Wompi para licencias.
	- `web/super/configuracion_avanzada.html` ajusta Òºnicamente la semÒ¡ntica y persistencia de la configuraciÒ³n global de pasarelas; no crea nuevas acciones empresariales ni altera wrappers de autorizaciÒ³n.
	- `backend/handlers/payments_handlers_test.go` cubre el escenario de checkout pÒºblico con dominio canÒ³nico y credenciales coherentes.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad por rol; `Pasarelas de licencias (Wompi/Epayco)` permanece como CRUA exclusivo de `super_administrador`.

- Actualizacion 2026-04-15 (host canÒ³nico para login Google y carga visible en estaciones):
	- `backend/utils/utils.go` incorpora un middleware de host canÒ³nico que redirige `www.powerfulcontrolsystem.com` al dominio raÒ­z antes de autenticaciÒ³n, evitando mezclar cookies y `redirect_uri` entre dos hosts pÒºblicos.
	- `backend/main.go` integra ese middleware sin crear rutas nuevas ni ampliar privilegios; el acceso administrativo conserva el mismo modelo de sesiÒ³n y rol existente.
	- `web/administrar_empresa/estaciones.html` aÒ±ade un estado visual `Cargando estaciones...` y mensaje de error de carga, sin modificar endpoints ni permisos del mÒ³dulo estaciones.
	- Impacto de matriz: sin cambios en roles, CRUD/A, wrappers o visibilidad administrativa por rol; solo se estabiliza el acceso y la UX operativa.

- Actualizacion 2026-04-15 (portal publico: contacto visible y pagina de informacion):
	- `web/index.html` incorpora un enlace superior a `/Informacion_de_contacto.html` y un CTA flotante `Contactenos` que abre WhatsApp con el numero publico comercial.
	- El acceso principal del header se renombra a `Registrarse o iniciar sesiÒ³n` y queda agrupado junto al enlace de contacto, sin alterar rutas protegidas ni permisos.
	- `AuthMiddleware` trata `index.html` y `/Informacion_de_contacto.html` como rutas publicas exactas, por lo que el portal comercial no requiere sesion.
	- `web/Informacion_de_contacto.html` expone descripcion general del sistema y datos de contacto (`powerfulcontrolsystem@hmail.com`, `3043306506`) sin requerir autenticacion.
	- `web/estilos.css` solo agrega soporte visual para el CTA flotante y la nueva landing de contacto; no se agregan endpoints protegidos ni cambios de wrappers.
	- Impacto de matriz: sin cambios en roles, CRUD/A ni visibilidad de paginas administrativas; la nueva experiencia es completamente publica y de solo lectura.

- Actualizacion 2026-04-15 (portal publico: landing descriptiva unica por tarjetas):
	- `web/index.html` sustituye el destino directo de `Explorar oferta` por la landing publica `/descripcion_de_los_sistemas.ht#<seccion>`, conservando el catalogo en una sola pagina y el salto a la descripcion correcta.
	- `AuthMiddleware` incluye `/descripcion_de_los_sistemas.ht` en el whitelist publico para que la navegacion desde las tarjetas no pida login.
	- `web/descripcion_de_los_sistemas.ht` consume `/api/public/pagina_principal` para renderizar todas las secciones del catalogo y reutiliza el `enlace` configurado en super como CTA `Probar Gratis` por tarjeta.
	- `web/super/pagina_principal.html` solo ajusta la semantica del campo `enlace`; no hay nuevos privilegios CRUD/A ni cambios de wrappers.
	- Impacto de matriz: `Pagina principal (tarjetas index)` sigue siendo CRUA exclusivo de `super_administrador`; la nueva landing publica es de lectura y no altera permisos empresariales.

- Actualizacion 2026-04-15 (checkout de licencias: Epayco primero y Wompi gobernado por configuracion avanzada):
	- `backend/handlers/payments_handlers.go` agrega `GET /api/public/licencias/payment_methods` para exponer al checkout publico solo las pasarelas realmente disponibles y en orden operativo (`epayco`, `wompi`).
	- `web/pagar_licencia.html` consume ese endpoint para mostrar solo Epayco y Wompi, con Epayco primero y Wompi debajo; `web/super/configuracion_avanzada.html` ya permite activar o desactivar ambas pasarelas desde super sin alterar privilegios empresariales.
	- `WompiTermsHandler` y `WompiCreateNequiTransactionHandler` bloquean acceso cuando Wompi esta desactivado o no tiene llaves completas.
	- Impacto de matriz: no se agregan roles nuevos ni se amplian permisos CRUD/A; `super_administrador` conserva CRUA sobre `Pasarelas de licencias (Wompi/Epayco)` y la ruta publica nueva es exclusivamente de lectura para el portal de licencias.

- Actualizacion 2026-04-15 (responsive transversal portal/admin):
	- Se ajustan `web/index.html` y `web/estilos.css` para mejorar portabilidad entre movil y escritorio en el portal publico, panel super y panel empresa.
	- No se agregan ni retiran permisos CRUD/A; el cambio no altera rutas protegidas, wrappers ni visibilidad por rol.
	- Impacto de matriz: sin cambios funcionales de autorizacion; mejora exclusiva de presentacion y navegabilidad.

- Actualizacion 2026-04-15 (hardening login Google y recordar cuenta):
	- `backend/handlers/auth_admin_handlers.go` omite `login_hint` invalidos y la capa cliente limpia `rememberedEmail` corruptos en login/menu/paneles administrativos antes de construir `/auth/google/login`.
	- No se agregan rutas nuevas ni se altera el alcance por rol para autenticacion, super o empresa.
	- Impacto de matriz: sin cambios CRUD/A ni wrappers; mejora exclusiva de integridad del acceso administrativo.

- Actualizacion 2026-04-14 (fix login Google en VPS/local y recordar cuenta):
	- Se habilitan como rutas publicas `GET /js/login.js` y `GET /api/public/pagina_principal` dentro de `AuthMiddleware`, evitando respuestas `401` en pagina de login/portal antes de autenticacion.
	- Se robustece el callback OAuth para adaptar `redirect_uri` al host real de la solicitud (`localhost` o dominio VPS) y forzar `https` en dominio publico, sin cambiar privilegios por rol.
	- Se emite una cookie auxiliar visible `browser_session_active` junto a `session_token` para que login/menu detecten sesion activa sin leer la cookie `HttpOnly` real.
	- Se evita fetch de perfil sin sesion en login/menu para reducir errores `401` visibles en cliente antes de autenticar.
	- Impacto de matriz: no hay ampliacion de permisos CRUD/A; se mantiene el mismo modelo de acceso por sesion para rutas protegidas y contexto global para login administrativo.

- Actualizacion 2026-04-14 (checkout de licencias con Epayco):
	- Se completa `POST /epayco/create_transaction`, `GET /epayco/transaction_status` y `POST/GET /epayco/webhook` para flujo de pago/confirmacion de licencias.
	- Se mantiene gestion global de pasarelas en contexto super (`/super/api/config/epayco`), sin cambios de privilegios en wrappers `/api/empresa/*`.
	- Impacto de matriz: solo `super_administrador` conserva permisos CRUD/A en modulo de pasarelas de licencias.

- Actualizacion 2026-04-14 (pagina principal dinamica administrada por super):
	- Se agrega `backend/handlers/pagina_principal_handlers.go` con `GET/PUT /super/api/pagina_principal` y `GET /api/public/pagina_principal`.
	- `web/super/pagina_principal.html` permite configurar tarjetas del home (cantidad, imagen, titulo, descripcion, enlace) y `web/index.html` las renderiza dinamicamente con CTA `Explorar oferta`.
	- Alcance de seguridad: gestion exclusiva para `super_administrador`; sin cambios en wrappers ni privilegios de `/api/empresa/*`.

- Actualizacion 2026-04-14 (venta publica por subdominio empresarial):
	- Se habilita resolucion automatica de `empresa_slug` por `Host`/`X-Forwarded-Host` en `backend/handlers/venta_publica.go` para subdominios tipo `{slug}.powerfulcontrolsystem.com`.
	- La raiz de subdominio empresarial (`/`) redirige internamente a `venta_publica.html` en `backend/main.go` para consumo publico sin query manual.
	- No hay cambios en privilegios CRUD/A por rol: se mantiene endpoint publico `/api/public/venta_publica` y wrappers existentes de seguridad empresarial.

- Actualizacion 2026-04-14 (impresoras operativas por empresa):
	- Se incorpora `backend/handlers/empresa_impresoras.go` y `backend/db/empresa_impresoras.go` para administrar impresoras empresariales y resolver destino de impresion.
	- Se registran rutas `GET/POST/PUT/DELETE /api/empresa/impresoras` (wrapper de seguridad) y `GET /api/empresa/impresoras/resolver` (wrapper de ventas).
	- Sin cambios de privilegios globales: se mantiene politica de lectura comun para rutas con wrapper y mutacion restringida por modulo de seguridad/ventas segun rol.

- Actualizacion 2026-04-14 (super: administracion de base de datos PostgreSQL):
	- Se incorpora endpoint global `/super/api/postgres/performance` para lectura de salud y rendimiento del motor PostgreSQL.
	- Se agrega la vista `web/super/administrar_base_de_datos.html` en el panel de superadministrador para monitoreo y accion operativa.
	- No hay cambios en permisos empresariales ni wrappers de `/api/empresa/*`; el acceso permanece exclusivo para `super_administrador`.

- Actualizacion 2026-04-14 (fase 4 PostgreSQL - estabilizacion worker contable):
	- Se ajusta tecnicamente el procesamiento de eventos/asientos en backend (`backend/db/eventos_contables.go`) para compatibilidad SQL portable en PostgreSQL durante salida controlada.
	- Se restablece runtime VPS con DSN PostgreSQL activos y validacion de salud operativa.
	- No hay cambios en permisos por rol, matrices CRUD/A ni wrappers de autorizacion por modulo.

- Actualizacion 2026-04-14 (migracion PostgreSQL runtime en VPS):
	- Se completa la conmutacion de runtime backend para operar con `DB_DIALECT=postgres` y DSN por base (`DB_SUPERADMIN_DSN`, `DB_EMPRESAS_DSN`).
	- Se agregan capas de compatibilidad SQL para transicion motor legado retirado/PostgreSQL en modulos core sin ampliar privilegios por rol.
	- No hay cambios en la matriz CRUD/A ni en wrappers de autorizacion: se preserva el mismo control por modulo y aislamiento por `empresa_id`.

- Actualizacion 2026-04-13 (estaciones, sensores y facturacion visual por estacion):
	- No hay ampliacion de privilegios ni cambios en matriz CRUD/A; se mantiene el mismo control para `/api/empresa/estacion_prefs`, `/api/empresa/configuracion_avanzada`, `/api/empresa/carritos_compra` y endpoints de sensores empresariales.
	- La reubicacion de colores de carrito a `configuracion_de_estaciones` es un cambio de UX/flujo, no de autorizacion.
	- Se valida aislamiento por `empresa_id` con prueba de handler en `empresa_estacion_prefs`, reforzando separacion de datos entre empresas.

- Actualizacion 2026-04-18 (gobernanza tecnica de estaciones y venta simple):
	- No hay ampliacion de privilegios ni cambios en la matriz CRUD/A; el alcance vigente para `/api/empresa/estacion_prefs`, `/api/empresa/sensor_puertas`, `/api/empresa/carritos_compra` y `/api/empresa/carritos_compra/items` se mantiene intacto.
	- Se agregan artefactos documentales de gobernanza para fijar contratos, invariantes multiempresa y recuperacion operativa del flujo sin modificar wrappers ni roles efectivos.

- Actualizacion 2026-04-13 (fix persistencia `empresa_estacion_prefs`):
	- Se corrige normalizacion de estado en capa DB (`estado` vacio => `activo`) sin alterar permisos ni wrappers de autorizacion.
	- El alcance de seguridad permanece igual: controles por `empresa_id` y permisos vigentes en rutas `/api/empresa/estacion_prefs`.

- Actualizacion 2026-04-13 (login empresa, seleccion y estaciones):
	- Se mantiene el mismo esquema de permisos por rol/modulo para endpoints empresariales (`/api/empresa/usuarios/*`, `/api/empresa/estacion_prefs`, `/api/empresa/carritos_compra`).
	- Los cambios son de robustez de flujo y contexto (`empresa_id`) en frontend, sin ampliacion de privilegios ni cambio de matriz CRUD/A.
	- Se preserva aislamiento por `empresa_id` para operacion concurrente de multiples estaciones y carritos por empresa.

- Actualizacion 2026-04-12 (login admin: contrato + reCAPTCHA real):
	- Se consolida la ruta administrativa `login.html -> /auth/google/* -> /accept.html -> /accept/complete` con persistencia de aceptaciÒ³n por cuenta en `administradores.acepta_contrato`.
	- No cambia la matriz CRUD por rol/modulo para rutas empresariales; el ajuste aplica al acceso administrativo global y al endurecimiento de autenticaciÒ³n.
	- Se mantiene aislamiento por `empresa_id` en acceso posterior, ya dentro de wrappers `/api/empresa/*` existentes.

- Actualizacion 2026-04-18 (gobernanza tecnica de autenticacion y tunel PostgreSQL):
	- No hay ampliacion de privilegios ni cambios en la matriz CRUD/A; los documentos nuevos formalizan el comportamiento de rutas ya existentes y del arranque local del backend.
	- Se explicita que las rutas publicas de autenticacion de usuarios de empresa siguen sujetas a validacion de alcance por `empresa_id` y que el acceso administrativo super permanece restringido al rol gestionado por backend.

- Actualizacion 2026-04-18 (gobernanza tecnica de venta publica empresarial):
	- No cambia la matriz de permisos existente: la configuracion y administracion del modulo sigue bajo permisos empresariales de ventas sobre `/api/empresa/venta_publica`.
	- Se formaliza que `/api/public/venta_publica` permanece publico solo para catalogo, creacion de pago y consulta de estado de orden dentro del alcance de la empresa resuelta por `empresa_id` o `empresa_slug`.

- Actualizacion 2026-04-18 (gobernanza tecnica de permisos_contexto y wrappers):
	- No se amplian privilegios; se documenta formalmente la politica vigente de wrappers por modulo para `/api/empresa/*` y la excepcion controlada de `WithEmpresaPublicScope` para autenticacion empresarial.
	- Se explicita que `permisos_contexto` responde con rol efectivo, matriz por modulo y visibilidad de paginas, aplicando overrides dinamicos y restricciones por licencia sin romper el aislamiento por `empresa_id`.

- Actualizacion 2026-04-18 (gobernanza tecnica de facturacion y documentos):
	- No cambia la matriz de permisos existente; el modulo de facturacion sigue bajo `WithEmpresaFacturacionPermissions` y las acciones `emitir`, `nota_credito`, `emitir_nota_credito`, `anular`, `procesar_reintentos`, `reconciliar_estados`, `firmar_xml_real`, `enviar_documento_real`, `reconexion_dian` y `consultar_acuse_real` permanecen en el dominio de permisos de facturacion.
	- Se documenta que los documentos de venta generados por carrito mantienen el mismo aislamiento por `empresa_id` aunque el tipo documental final varie entre `factura_electronica` y `comprobante_pago`.

- Actualizacion 2026-04-08 (super: alertas de inicio/reinicio de servidor):
	- Se agrega en configuracion avanzada super la clave `gmail.restart_alert_to` para correo destino de alertas operativas de arranque/reinicio.
	- El cambio no altera permisos de roles empresariales en wrappers `/api/empresa/*`; aplica al ambito global de `super_administrador`.
	- Se mantiene aislamiento multiempresa en operacion: los eventos runtime se registran en contexto super y no exponen datos de empresas fuera de su alcance.

- Actualizacion 2026-04-08 (chat/tareas usuario-admin con adjuntos documentales):
	- Se mantiene el control de acceso del modulo ventas para `/api/empresa/chat_tareas/*` (sin cambios de rol/accion respecto a la matriz vigente).
	- El backend de chat/tareas deriva actor desde sesion autenticada para distinguir `usuario` y `admin`, evitando suplantacion de autor en mensajes/adjuntos.
	- Se habilita colaboracion directa usuario-admin al autoagregar admin propietario de la empresa cuando una conversacion es creada por usuario.
	- Se amplian adjuntos permitidos para colaboracion operativa con documentos de oficina (`doc/docx/xls/xlsx/ppt/pptx/rtf/odt/ods/odp`).

- Actualizacion 2026-04-08 (configuracion monetaria y numerica empresarial):
	- Se agrega en `administrar_empresa/configuracion.html` la seccion de formato monetario/numerico por empresa (`moneda_codigo`, `sistema_numerico`, `usar_decimales`, `cantidad_decimales`).
	- Se mantiene el mismo control de acceso del modulo seguridad para `/api/empresa/configuracion_avanzada` (sin cambios de rol/accion respecto a la matriz actual).

- Actualizacion 2026-04-08 (chat IA empresarial):
	- Se alinea la configuracion avanzada de super para gestionar credencial `deepseek:deepseek-chat`.
	- La pagina `chat_con_inteligencia_artificial` de empresa se actualiza a mensajes/modelo IA generico con ejecucion operativa en DeepSeek, manteniendo control de alcance por `empresa_id`.

- Actualizacion 2026-04-18 (chat IA empresarial: selector DeepSeek/Ambis):
	- La pagina `chat_con_inteligencia_artificial` permite elegir entre `deepseek:deepseek-chat` y `ollama:ambis` sin relajar el control de acceso ni el alcance por `empresa_id`.
	- `Ambis Local` se consume solo desde el backend por loopback del VPS (`127.0.0.1:11434`), sin acceso directo desde navegadores empresariales.
	- Se mantienen los mismos wrappers y validaciones sobre `/api/empresa/chat_con_inteligencia_artificial/modelos`, `/modelo_preferido`, `/consultar` y `/historial`.

- Se implementa middleware en `backend/handlers/empresa_permisos.go` para validar:
	- identidad administrativa activa,
	- alcance de `empresa_id`,
	- permisos por rol/accion (C/R/U/D/A) por modulo.
- Cobertura inicial aplicada en `backend/main.go` sobre rutas criticas:
	- Ventas: `/api/empresa/carritos_compra`, `/api/empresa/carritos_compra/items`.
	- Inventario: `/api/empresa/bodegas`, `/api/empresa/categorias_productos`, `/api/empresa/productos`, `/api/empresa/inventario/*`, `/api/empresa/productos/precios_historial`.
	- Finanzas: `/api/empresa/finanzas/movimientos`, `/api/empresa/finanzas/configuracion`, `/api/empresa/finanzas/periodos`, `/api/empresa/finanzas/asientos_contables`.
- Cobertura ampliada (2026-04-04):
	- Clientes: `/api/empresa/clientes`.
	- Compras/Proveedores: `/api/empresa/proveedores`.
	- Facturacion: `/api/empresa/facturacion_electronica`, `/api/empresa/facturacion_electronica/pais_detectado`, `/api/empresa/facturacion_electronica/dian`.
	- Servicios de catalogo: `/api/empresa/servicios` bajo politica de inventario.
- Cobertura adicional (2026-04-04 - cierre de rutas pendientes):
	- Seguridad/usuarios:
		- `/api/empresa/usuarios`.
		- `/api/empresa/configuracion_avanzada`.
		- `/api/empresa/roles_de_usuario`.
		- `/api/empresa/auditoria/eventos`.
	- Inventario:
		- `/api/empresa/productos/imagen`.
		- `/api/empresa/ubicacion_gps/dispositivos`.
		- `/api/empresa/ubicacion_gps/recorridos`.
	- Colaboracion operativa (politica ventas):
		- `/api/empresa/chat_tareas/conversaciones`.
		- `/api/empresa/chat_tareas/participantes`.
		- `/api/empresa/chat_tareas/mensajes`.
		- `/api/empresa/chat_tareas/mensajes/adjunto`.
		- `/api/empresa/chat_tareas/tareas`.
		- `/api/empresa/chat_tareas/citas`.
- Cobertura adicional (2026-04-05 - contexto de permisos por rol):
	- Seguridad:
		- `/api/empresa/permisos_contexto` con soporte de matriz expandida (`include_matrix=1`) para consulta de permisos efectivos por modulo/accion.
- Cobertura adicional (2026-04-05 - control visual de menu por permisos efectivos):
	- Frontend empresa:
		- `web/js/administrar_empresa.js` consume `/api/empresa/permisos_contexto` para ocultar enlaces no autorizados por rol/modulo.
		- `web/administrar_empresa.html` muestra evidencia visual (`menuPermsEvidence`) con rol y fuente de permisos activa para UAT.
- Cobertura automatizada inicial en `backend/handlers/empresa_permisos_test.go`:
	- denegacion de escritura sin permiso por rol,
	- aprobacion permitida para rol contabilidad en cierre de periodos,
	- bloqueo por fuera de alcance de empresa.
	- denegacion/escritura por rol en modulos `compras` y `facturacion`, y aprobacion de escritura en `clientes` para `cajero` segun matriz.
	- denegacion de escritura en modulo seguridad para `supervisor_sucursal`.
	- aprobacion permitida en modulo seguridad para `admin_empresa`.
	- denegacion para `cajero` al procesar asientos (`action=procesar_asientos`) en modulo finanzas.
	- aprobacion para `contabilidad` al procesar asientos (`action=procesar_asientos`) en modulo finanzas.
	- registro automatico de auditoria para acciones criticas autorizadas (`C/U/D/A`) en middleware de permisos empresariales.
	- cobertura de auditoria automatica por modulo con pruebas en `backend/handlers/auditoria_empresa_test.go` para:
		- `ventas` (`action=cerrar`),
		- `compras` (`action=emitir_orden`),
		- `facturacion` (`action=emitir`).

## Matriz UAT de cierres de caja (roles y transiciones)

Fecha de actualizacion: 2026-04-04

### Casos por rol en endpoint `/api/empresa/finanzas/cierres_caja`

| Caso | Rol | Metodo/accion | Resultado esperado |
|---|---|---|---|
| UAT-CC-R1 | cajero | `PUT action=aprobar` | `403 forbidden` |
| UAT-CC-R2 | supervisor_sucursal | `PUT action=aprobar` | `403 forbidden` |
| UAT-CC-R3 | admin_empresa | `PUT action=aprobar` | `200 ok` |

### Casos de transicion del estado de cierre

| Caso | Estado actual | Accion | Precondicion | Resultado esperado |
|---|---|---|---|---|
| UAT-CC-T1 | abierto | aprobar | ninguna | `409 conflict` (transicion invalida) |
| UAT-CC-T2 | abierto | cerrar | `caja_fisica` valida | `200 ok`, estado `cerrado` |
| UAT-CC-T3 | cerrado | aprobar | ninguna | `200 ok`, estado `aprobado` |
| UAT-CC-T4 | aprobado | reabrir | ninguna | `200 ok`, estado `abierto` |
| UAT-CC-T5 | aprobado | editar/eliminar | ninguna | bloqueo (`409`/error de negocio) |

## Matriz final endpoint/rol (implementacion vigente 2026-04-04)

Leyenda de roles:
- SA: super_administrador
- AE: admin_empresa
- SS: supervisor_sucursal
- CJ: cajero
- IN: inventario
- CO: compras
- CT: contabilidad
- AU: auditor

Regla de lectura comun (R):
- En rutas con wrapper de permisos, lectura queda habilitada para SA, AE, SS, CJ, IN, CO, CT y AU.

| Endpoint | Wrapper/modulo | C/U/A habilitado | D habilitado | Observaciones de accion |
|---|---|---|---|---|
| `/api/empresa/carritos_compra` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | `action=cerrar|reabrir|pagar_estacion|activar_estacion|pagar|suspender|reactivar` exige `A` |
| `/api/empresa/carritos_compra/items` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | mutaciones de items bajo politica de ventas |
| `/api/empresa/chat_tareas/conversaciones` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/chat_tareas/participantes` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/chat_tareas/mensajes` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/chat_tareas/mensajes/adjunto` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | multipart con `empresa_id` obligatorio |
| `/api/empresa/chat_tareas/tareas` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | colaboracion operativa bajo modulo ventas |
| `/api/empresa/chat_tareas/citas` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | agenda de citas compartida por empresa con recordatorios y estado operativo |
| `/api/empresa/publicaciones` | `WithEmpresaVentasPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | posts de red social empresarial por `empresa_id`; lectura pÒºblica separada en `/api/public/publicaciones` |
| `/api/empresa/venta_publica` | `WithEmpresaVentaPublicaPermissions` | SA, AE, SS, CJ | SA, AE, SS, CJ | configura tienda/pasarelas propias, paginas publicas (`action=paginas`), productos publicados y ordenes por empresa |
| `/api/empresa/bodegas` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | CRUD inventario |
| `/api/empresa/categorias_productos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | CRUD inventario |
| `/api/empresa/productos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | CRUD inventario |
| `/api/empresa/productos/imagen` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | upload multipart |
| `/api/empresa/servicios` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | catalogo operativo en politica inventario |
| `/api/empresa/inventario/existencias` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | lectura y mutaciones bajo modulo inventario |
| `/api/empresa/inventario/movimientos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | lectura y mutaciones bajo modulo inventario |
| `/api/empresa/inventario/transferir` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | transferencias de bodega |
| `/api/empresa/inventario/ajustar` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | ajustes de existencias |
| `/api/empresa/inventario/cambiar_producto` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | remapeo operativo producto/bodega |
| `/api/empresa/productos/precios_historial` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | historial de precios |
| `/api/empresa/ubicacion_gps/dispositivos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | geolocalizacion en politica inventario; `POST` respeta tope super `empresa.limitaciones.gps.max_dispositivos` (default 2) y devuelve `409` si la empresa ya alcanzo el cupo |
| `/api/empresa/ubicacion_gps/recorridos` | `WithEmpresaInventarioPermissions` | SA, AE, SS, IN | SA, AE, SS, IN | geolocalizacion en politica inventario |
| `/api/empresa/clientes` | `WithEmpresaClientesPermissions` | SA, AE, SS, CJ | - | modulo clientes sin `D` por politica actual |
| `/api/empresa/proveedores` | `WithEmpresaComprasPermissions` | SA, AE, SS, CO | - | `action=emitir_orden|recepcionar_compra|contabilizar_compra|aprobar` exige `A` |
| `/api/empresa/soportes_compras_ia` | `WithEmpresaSoportesComprasIAPermissions` | SA, AE, SS, CO, CT | - | captura foto/PDF/XML de compras y gastos; `action=extraer_ia` usa `U`, `aprobar|rechazar|contabilizar` exige `A` |
| `/api/empresa/facturacion_electronica` | `WithEmpresaFacturacionPermissions` | SA, AE, CJ | - | `action=emitir|nota_credito|emitir_factura|emitir_documento` exige `A` |
| `/api/empresa/facturacion_electronica/pais_detectado` | `WithEmpresaFacturacionPermissions` | SA, AE, CJ | - | consulta/actualizacion bajo politica facturacion |
| `/api/empresa/facturacion_electronica/dian` | `WithEmpresaFacturacionPermissions` | SA, AE, CJ | - | incluye `action=guia_onboarding|validar_credenciales|subir_firma|checklist|validar|generar_cufe_demo|generar_xml_demo|generar_xml_ubl_base|firmar_xml_real|firmar_xml_xades_base|validar_documento_dian|diagnostico_oficial|enviar_documento_real|consultar_acuse_real|reconexion_dian|pruebas_dian|enviar_set_pruebas|activar_produccion_local`; las acciones reales de firma/envio/set/activacion requieren aprobacion; opera por `empresa_id` con `NIT/token/certificado` por empresa y software compartido opcional |
| `/api/empresa/finanzas/movimientos` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | `action=cerrar|reabrir|aprobar|procesar_asientos|procesar` exige `A` |
| `/api/empresa/finanzas/configuracion` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | configuracion financiera |
| `/api/empresa/finanzas/periodos` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | cierre/reapertura de periodos en `A` |
| `/api/empresa/finanzas/asientos_contables` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | `action=procesar_asientos` validado por rol |
| `/api/empresa/finanzas/cierres_caja` | `WithEmpresaFinanzasPermissions` | SA, AE, CT | SA, CT | `action=aprobar` restringido por permiso `A` |
| `/api/empresa/usuarios` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | seguridad/usuarios solo administracion empresa |
| `/api/empresa/usuarios?action=foto` | `WithEmpresaSeguridadPermissions` | - | SA, AE | upload multipart de foto de usuario; valida `empresa_id` + `usuario_id` y guarda en carpeta empresarial `imagenes/usuarios` |
| `/api/empresa/configuracion_avanzada` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | seguridad/configuracion sensible |
| `/api/empresa/configuracion_avanzada/logo` | `WithEmpresaSeguridadPermissions` | - | SA, AE | carga multipart del logo empresarial; persiste `empresa_configuracion_avanzada.logo_url` y `mostrar_logo` por `empresa_id` |
| `/api/empresa/impresoras` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | CRUD impresoras y acciones `predeterminada|activar|desactivar|funcionalidad|producto|producto_regla|receta`, con catalogo de productos/categorias por empresa |
| `/api/empresa/impresoras/resolver` | `WithEmpresaVentasPermissions` | - | - | endpoint operativo de solo lectura para resolver impresora objetivo por `funcionalidad`, `producto_id` o `receta_id`, aplicando prioridad producto/categoria/todos |
| `/api/empresa/control_electrico` | `WithEmpresaControlElectricoPermissions` | SA, AE, SS | SA, AE | configuracion Domotica por empresa: controladores, aparatos, fotos, lecturas, reglas de sensores, alarmas, reportes y sincronizacion; cambios y comandos requieren `A` |
| `/super/api/domotica_storage` | `WithSuperAuditoria` + sesion super | SA | SA | limites de imagenes y revision de carpetas empresariales de Domotica; no expone secretos |
| `/api/empresa/roles_de_usuario` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | consulta roles globales + roles propios por `empresa_id`; crea/edita/desactiva solo roles personalizados de la empresa con evidencia trazable |
| `/api/empresa/permisos_contexto` | `WithEmpresaSeguridadPermissions` | - | - | endpoint `GET` para visualizar permisos efectivos por modulo/accion; `include_matrix=1` retorna matriz comparativa por rol |
| `/api/empresa/auditoria/eventos` | `WithEmpresaAuditoriaPermissions` | SA, AE | SA, AE | consulta y retencion (`action=retener|purgar`); `action=conexion` registra perdida/restauracion de internet como accion de lectura operativa para usuarios con acceso a la empresa |
| `/api/empresa/backups` | `WithEmpresaSeguridadPermissions` | SA, AE | SA, AE | snapshots/restauracion y depuracion por fecha (`action=restaurar|depurar_fecha` requiere `A`) |

### Endpoints fuera de wrapper (control alterno)

| Endpoint | Control aplicado | Nota |
|---|---|---|
| `/api/empresa/usuarios/login` | validacion de alcance por usuario/empresa en handler | sin middleware de modulo |
| `/api/empresa/usuarios/establecer_password` | validacion de alcance por usuario/empresa en handler | sin middleware de modulo |
| `/api/empresa/facturacion_electronica/paises_disponibles` | catalogo global | sin `empresa_id` obligatorio |
| `/api/empresa/chat_con_inteligencia_artificial/modelos` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |
| `/api/empresa/chat_con_inteligencia_artificial/modelo_preferido` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |
| `/api/empresa/chat_con_inteligencia_artificial/consultar` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |
| `/api/empresa/chat_con_inteligencia_artificial/historial` | `ensureEmpresaAccessByAccount` | validacion por cuenta Google + `empresa_id` |
| `/super/api/chat_con_ia_global/modelos` | `paginaPrincipalRequireSuperAdmin` | requiere sesion super valida y rol `super_administrador` |
| `/super/api/chat_con_ia_global/modelo_preferido` | `paginaPrincipalRequireSuperAdmin` | requiere sesion super valida y rol `super_administrador` |
| `/super/api/chat_con_ia_global/consultar` | `paginaPrincipalRequireSuperAdmin` | requiere sesion super valida y rol `super_administrador` |
| `/super/api/chat_con_ia_global/historial` | `paginaPrincipalRequireSuperAdmin` | requiere sesion super valida y rol `super_administrador` |
| `/super/api/licencias/vencimiento_alertas` | `paginaPrincipalRequireSuperAdmin` | configuracion, vista previa y ejecucion manual de alertas de licencia proximas a vencer; acceso exclusivo de `super_administrador` |
| `/super/api/correos_masivos` | `paginaPrincipalRequireSuperAdmin` | previsualizacion y envio de comunicados globales a administradores y usuarios de empresa; acceso exclusivo de `super_administrador` y confirmacion obligatoria para enviar |

## Checklist UAT de Punto 3 (permisos y seguridad)

| ID | Verificacion | Estado | Evidencia automatizada |
|---|---|---|---|
| P3-UAT-01 | Denegar escritura inventario a `cajero` | ok | `TestWithEmpresaInventarioPermissionsDeniesCajeroWrite` |
| P3-UAT-02 | Denegar escritura GPS a `cajero` | ok | `TestWithEmpresaInventarioPermissionsDeniesCajeroWriteGPS` |
| P3-UAT-03 | Permitir chat adjunto a `cajero` autenticado | ok | `TestWithEmpresaVentasPermissionsAllowsCajeroChatAdjuntoMultipart` |
| P3-UAT-04 | Rechazar chat adjunto sin autenticacion | ok | `TestWithEmpresaVentasPermissionsRejectsChatAdjuntoWithoutAuth` |
| P3-UAT-05 | Bloquear acceso fuera de alcance de empresa | ok | `TestWithEmpresaVentasPermissionsDeniesOutOfScopeEmpresa` |
| P3-UAT-06 | Denegar `procesar_asientos` a `cajero` | ok | `TestWithEmpresaFinanzasPermissionsDeniesCajeroProcesarAsientos` |
| P3-UAT-07 | Permitir `procesar_asientos` a `contabilidad` | ok | `TestWithEmpresaFinanzasPermissionsAllowsContabilidadProcesarAsientos` |
| P3-UAT-08 | Denegar escritura seguridad a `supervisor_sucursal` | ok | `TestWithEmpresaSeguridadPermissionsDeniesSupervisorWrite` |
| P3-UAT-09 | Permitir accion de seguridad a `admin_empresa` | ok | `TestWithEmpresaSeguridadPermissionsAllowsAdminApprove` |
| P3-UAT-10 | Registrar auditoria en acciones criticas ventas/compras/facturacion | ok | `TestWithEmpresaVentasPermissionsRegistraAuditoriaAccionCritica`, `TestWithEmpresaComprasPermissionsRegistraAuditoriaAccionCritica`, `TestWithEmpresaFacturacionPermissionsRegistraAuditoriaAccionCritica` |
| P3-UAT-11 | Exponer contexto de permisos por rol/modulo en endpoint de seguridad | ok | `TestEmpresaPermisosContextoHandlerRetornaPermisosPorRol`, `TestEmpresaPermisosContextoHandlerIncluyeMatrizRoles` |
| P3-UAT-12 | Ocultar menu por permisos efectivos y mostrar evidencia visual por rol en panel empresa | ok | evidencia visual `menuPermsEvidence` + consumo `GET /api/empresa/permisos_contexto` en `web/js/administrar_empresa.js` |

Ejecucion de validacion actual (2026-04-05):
- `go test ./handlers -run "PermisosContexto|WithEmpresa.*Permissions" -count=1`.
- Resultado: validacion del bloque de permisos y endpoint de contexto (ok).

## Reglas de seguridad obligatorias

1. Todo endpoint debe validar empresa_id y, cuando aplique, sucursal_id antes de operar.
2. Ningun usuario puede actuar fuera de su alcance de empresa/sucursal.
3. Toda accion critica debe dejar auditoria con request_id, empresa_id, usuario, accion y timestamp.
4. Operaciones de cierre/aprobacion deben requerir rol con permiso A.
5. Eliminaciones funcionales deben implementarse como anulacion/inactivacion cuando aplique trazabilidad legal.
6. La apertura o reapertura de cajas debe respetar el limite `max_cajas_simultaneas` de la licencia activa de la empresa; si no hay licencia vigente se aplica el default conservador de 2 cajas. La configuracion empresarial puede desactivar cajas simultaneas o fijar un limite interno menor, pero nunca ampliar el cupo de la licencia.

## Acciones tecnicas siguientes (cierre operativo punto 3)

1. Incorporar pruebas UAT de regresion para endpoints sin wrapper de modulo (`usuarios/login`, `establecer_password`, chat IA por cuenta Google).
2. Definir politica de aprobacion para rutas de lectura sensible en seguridad (`auditoria/eventos`) segun perfil `auditor` vs `admin_empresa`.
3. Evaluar prueba automatizada E2E del menu dinamico para evitar regresiones de visibilidad por rol.

## Actualizacion 2026-05-03 - Criterio operativo por rol

Para declarar un modulo listo en produccion se debe validar por rol: acceso a la empresa correcta, visibilidad de menu, permiso de lectura, permiso de escritura, accion principal, reporte asociado y auditoria. Los modulos con hardware o proveedor externo requieren prueba adicional con el dispositivo o servicio real: impresoras, cajon monedero, RFID/NFC, GPS, pasarela de pago, facturacion electronica, SMTP, Nextcloud, OnlyOffice e IA.
