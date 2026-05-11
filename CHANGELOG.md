- 2026-05-11: limpieza PostgreSQL-only del proyecto. Se retiran rastros del motor legado de codigo, frontend, scripts, documentacion vigente e historica, se cambian consultas residuales de indices a `pg_indexes`, se renombran helpers frontend a fechas de backend y se eliminan artefactos locales generados por perfiles temporales. No se agregan dependencias ni cambios en `go.mod`.
- 2026-05-11: actualizado `web/super/licencias_resumen.html` como centro de mando profesional del VPS y del proyecto. Consolida salud general, CPU/RAM/disco/trafico, PostgreSQL, alertas, errores recientes, servicios, procesos, licencias, empresas y consumo OpenAI estimado usando endpoints existentes, sin dependencias nuevas ni cambios de esquema.
- 2026-05-11: cierre implementable de pendientes 1 a 8. El checkout publico de licencias ahora permite seleccionar manualmente el pais de pago y reconsulta disponibilidad Wompi/Epayco por `pais_codigo`, evitando depender solo del navegador/VPN. Se corrigen referencias activas a documentos historicos inexistentes (`estructura_del_codigo` alias y plan maestro 14/15 puntos) hacia fuentes vigentes. Se deja explicitado que DIAN SOAP/WSDL oficial, hardware/proveedores reales, E2E con credenciales y normalizacion masiva de mojibake siguen como cierres externos/controlados, no como completados locales.
- 2026-05-11: implementada capa de madurez empresarial de 12 pasos: staging anonimizado por defecto, monitoreo Prometheus/Grafana, backups externos rclone/S3, deploy automatico opcional a staging, QA por roles, matriz de pagos/comprobantes, prueba de carga smoke, manifiesto de release, auditoria de soporte y normalizacion documental. Verificacion: `.\scripts\profesional_preflight.ps1 -Full` OK y carga smoke staging p95 1008 ms con error rate 0.
- 2026-05-10: Carritos suma modo tactil configurable por empresa/estacion para adaptar carrito operativo, cobro y agregador de productos por botones a tablets, monitores POS y pantallas tactiles.
- 2026-05-10: agregado modulo privado **Alertas del sistema** en super administrador. Configura destino, umbrales y enfriamiento; evalua disco VPS, trafico, sesiones administrativas y conexiones PostgreSQL; envia correo via Gmail SMTP y registra historial en `super_alertas_eventos`. Tambien se amplian metricas con `disk_total`, `disk_used` y `disk_percent`. Verificacion: `go test ./...` en `backend/`.
- 2026-05-10: actualizado sistema documental de roles/permisos con modulos finos (`crm_unificado`, `reservas_hotel`, `chat_tareas`, `horarios_trabajadores`, `asistencia_empleados`, `vehiculos_registro`, `hoja_vida_operativa`, `ubicacion_gps`, `nomina_sueldos`, `reportes`, `auditoria`, `backups`, `documentos_onlyoffice`, `nextcloud`), wrappers API especificos y compatibilidad de licencias amplias. La ayuda privada de super administrador queda accesible desde boton en `web/super_administrador.html` y sigue restringida a `super_administrador`. Documentacion agregada: `documentos/reporte_roles_ayuda_super_2026-05-10.md`.
- 2026-05-07: QA E2E Motel Calipso sobre `empresa_id=7`. Regresion escritorio 60/60 modulos, pruebas profundas 6/6 con datos QA reales (parqueadero QR/cobro/anulacion, WMS, centros de costo, activos fijos, red social con imagen, carta/venta publica), validacion movil dirigida de venta publica/asistencia y ajustes de robot/radio/favoritos para no tapar botones en celular. Reportes en `backend/tmp_tools/qa_calipso_operativo/*_report.json`.
- 2026-05-06: implementados `bancos_pagos`, `gestion_documental`, `cumplimiento_kyc`, `contratos_obligaciones`, `helpdesk` y `calidad_procesos` como modulos empresariales Colombia sobre nucleo compartido por `empresa_id`, con APIs privadas, pantallas administrativas, permisos/licencias, datos demo, exportacion CSV y documentacion.
- 2026-05-06: continuacion de fases para modulos empresariales Colombia con plantillas por modulo, seguimiento profesional, cambio de estado auditado, filtro por estado y UI compartida sin duplicar codigo.
- 2026-05-06: fase de reportes ejecutivos para modulos empresariales Colombia con metricas por estado/tipo/categoria/prioridad, vencimientos, criticidad, responsables faltantes, valor pendiente y recomendaciones automaticas.
- 2026-05-06: fase de importacion masiva para modulos empresariales Colombia con boton CSV, endpoint `importar_registros`, validacion parcial y bitacora por empresa.
- 2026-05-06: fase de evidencias para modulos empresariales Colombia con soportes por registro, endpoint `evidencia/evidencias`, validacion multiempresa y bitacora automatica.
- 2026-05-06: fase de aprobaciones para modulos empresariales Colombia con solicitudes por nivel, destinatario, vencimiento, decision aprobada/rechazada, cambio de estado y bitacora.
- 2026-05-06: fase de tareas para modulos empresariales Colombia con compromisos por registro, responsable, prioridad, vencimiento, estados operativos y bitacora.
- 2026-05-06: fase de expediente 360 para modulos empresariales Colombia con vista consolidada de registro, eventos, evidencias, aprobaciones, tareas, resumen y recomendacion.
- 2026-05-06: fase de agenda para modulos empresariales Colombia con alertas de vencidos, proximos vencimientos, tareas, aprobaciones pendientes, severidad y acceso al expediente.
- 2026-05-06: fase de cierre controlado para modulos empresariales Colombia, exigiendo evidencia y bloqueando aprobaciones pendientes o tareas abiertas antes de cerrar.
- 2026-05-06: fase de plan de accion para modulos empresariales Colombia, convirtiendo alertas de agenda en tareas con prioridad y control de duplicados.
- 2026-05-06: fase de responsables para modulos empresariales Colombia con tablero de carga por responsable, vencidos, tareas, aprobaciones y recomendaciones.
- 2026-05-06: fase de SLA para modulos empresariales Colombia con cumplimiento, semaforo, buckets de vencimiento y recomendaciones.
- 2026-05-06: fase de riesgo operativo para modulos empresariales Colombia con score 0-100, nivel bajo/medio/alto, factores y recomendaciones.
- 2026-05-06: fase de exportacion de auditoria para modulos empresariales Colombia con CSV multi-seccion de resumen, registros, agenda, SLA, riesgo, responsables, tareas, aprobaciones, evidencias y bitacora.
- 2026-05-06: fase de busqueda avanzada para modulos empresariales Colombia con filtros de backend por texto, estado, tipo, categoria, prioridad, responsable, vencidos y proximos vencimientos.
- 2026-05-06: fase de acciones masivas para modulos empresariales Colombia con seleccion de registros, cambio controlado de estado/prioridad/responsable y bitacora por registro.

- **Logistica avanzada / WMS**: nuevo modulo `logistica_wms` para operar bodega profesional con ubicaciones internas, ordenes WMS, picking, packing, despachos, rutas, avance por item, bitacora, datos demo y exportacion CSV. Integra `/api/empresa/logistica_wms`, permisos/licencia independientes, menu de Inventario y compras, pantalla `web/administrar_empresa/logistica_wms.html`, tablas WMS por `empresa_id` y documentacion. Verificacion: `go test ./db -run Test.*WMS -count=1`, `go test ./... -count=1` y `git diff --check`.

- **Declaraciones Tributarias y Motor de Impuestos Colombia**: nuevo modulo financiero `declaraciones_tributarias` para preliquidar, revisar y controlar IVA, retencion en la fuente, ReteIVA, ICA/ReteICA, consumo, renta y regimen simple por empresa. Agrega `/api/empresa/declaraciones_tributarias`, tablas de declaraciones, movimientos de conciliacion y calendario tributario editable, permisos/licencia independientes, pantalla `web/administrar_empresa/declaraciones_tributarias.html`, enlace en Centro financiero y documentacion. Verificacion: `go test ./db -run Test.*Declaracion -count=1`, `go test ./... -count=1` y `git diff --check`.

- **Captura inteligente de compras/gastos**: nuevo modulo empresarial `soportes_compras_ia` en Compras para cargar foto/PDF/XML, extraer datos con GPT-5.5, detectar duplicados, aprobar/rechazar y contabilizar como cuenta por pagar. Integra `/api/empresa/soportes_compras_ia`, permisos/licencia independientes, pantalla operativa, tablas de soportes/eventos y documentacion. Verificacion: `go test ./db -run Test.*Soporte.*IA -count=1`, `go test ./... -count=1` y `git diff --check`.

- **Portal contador / oficina virtual contable**: nuevo modulo `portal_contador` para firmas contables y contadores externos con portafolio de clientes, obligaciones DIAN/contables, solicitudes de documentos, comunicaciones, dashboard, datos demo y exportacion CSV. Integra `/api/empresa/portal_contador`, permisos/licencia independientes, menu financiero y documentacion. Verificacion: `go test ./db -run TestPortalContador -count=1` y `go test ./... -count=1`.

- **Gestion de cobranza profesional**: nuevo modulo financiero por empresa para recuperar cartera sin duplicar cuentas por cobrar. Agrega `/api/empresa/cobranza`, tablas de plantillas, campanas, gestiones y promesas de pago, permisos/licencia `cobranza`, pantalla `web/administrar_empresa/cobranza.html` en el Centro financiero y contable, datos demo, simulacion de envio y exportacion CSV. Verificacion: `go test ./db -run TestCobranza -count=1` y `go test ./... -count=1`.

- **Portal publico, carta QR, Motel Calipso, carnets y aislamiento multiempresa**: `web/index.html` actualiza las descripciones de modulos y tarjetas fallback para reflejar POS, hotel/motel, gimnasio, odontologia, domicilios tipo Rappi, Taxi System tipo Uber, turnos, carnets empresariales, control electrico, carta publica QR, red social, roles/licencias y hoja de vida. Se agrega `/api/empresa/carnets` y `web/administrar_empresa/carnets.html` para emitir carnets de empleados/usuarios con plantillas, QR, foto, exportacion PNG/SVG, impresion y bitacora. Se documenta la publicacion real de Motel Calipso (`motel-calipso`) con venta publica, carta publica, items y posts de red social. `AuthMiddleware` permite `visualizar_productos_y_precios_publico.html` directo y por slug sin sesion, manteniendo protegida la administracion. Los wrappers `WithEmpresa*` rechazan `empresa_id` contradictorios entre URL, cabecera, formulario/multipart y JSON para mantener separados todos los modulos empresariales. Se audita el menu empresa, el catalogo backend de paginas, las claves de licencia y rutas `/api/empresa` para evitar modulos duplicados o enlaces sin regla. Verificacion: `go test ./handlers`, `go test ./utils`, `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1` y HTTP 200 en produccion para venta publica, carta publica y red social.

- **Control electrico Raspberry Pi**: nuevo modulo en Administrar empresa para configurar IP/puerto/API de Raspberry Pi y mapear cada estacion/habitacion a multiples relés GPIO (luces, jacuzzi, aire, puerta u otros). El carrito de estacion agrega boton `Control electrico` para control manual de salidas; las estaciones envian `on` al activar/recuperar/reabrir y `off` al pagar/cerrar/desactivar. Archivos: `backend/db/control_electrico.go`, `backend/handlers/control_electrico.go`, `web/administrar_empresa/control_electrico.html`, `web/administrar_empresa/carrito_de_compras.html`, `backend/main.go`. Verificacion: `go test ./...` OK.

- **Pagos ePayco de licencias**: el fallback estandar se actualiza a la integracion oficial `checkout.js` con `external: "true"` y `PUBLIC_KEY`, evitando el POST legacy a `secure.payco.co/checkout.php` que podia terminar en "comercio no reconocido". `P_KEY` queda reservado al backend para confirmar webhooks con firma SHA256 y no se expone al navegador. Verificacion: `go test ./handlers -run Test.*Epayco -count=1` y `go test ./... -count=1`.

- **Administrar empresa - Hoja de vida operativa**: nuevo modulo universal para llevar historial de motos en taller, pacientes, vehiculos, equipos, mascotas, maquinaria o activos. Incluye ficha principal, eventos/servicios, alertas, recurrencia y resumen operativo. Archivos: `backend/db/hoja_vida_operativa.go`, `backend/handlers/hoja_vida_operativa.go`, `web/administrar_empresa/hoja_vida_operativa.html`, `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `backend/main.go`.

- **Flujo Git + VPS**: `scripts/pcs_deployment.local.ps1` (plantilla `pcs_deployment.local.ps1.example`, ignorada en git) centraliza `PcsGitRemoteUrl` y `PcsVpsHost` / ruta / puerto / SSH. `sync_to_vps.ps1` aplica la parte VPS; `publicar_git_y_vps.ps1` ejecuta `actualizar_repositorio.ps1` y luego el sync.

- **Despliegue Git local**: `scripts/actualizar_repositorio.ps1` admite fijar el repositorio remoto con `-RepoUrl`, variables `PCS_REPO_URL` / `REPO_URL`, o el archivo local `scripts/actualizar_repositorio.repo_url` (plantilla `*.repo_url.example`). Si `origin` apunta a otra URL, el script exige `-SetOrigin` para actualizar el remoto antes del push. Archivos: `scripts/actualizar_repositorio.ps1`, `scripts/actualizar_repositorio.repo_url.example`, `.gitignore`, documentación relacionada.

- **Chat global (super)**: la IA debe **preguntar confirmación** antes de emitir un bloque `PCS_ACTION` (resumen y pregunta en un turno; el JSON ejecutable solo tras un “sí”/equivalente explícito del usuario). Texto de bienvenida actualizado en la UI.

- **Chat global (super)**: en cada pregunta el backend adjunta metadatos de **toda** la base `pcs_superadministrador` (conteos por tabla, columnas `nombre:tipo`, reparto de administradores por rol), sin filas con datos sensibles. La pantalla de lógica del chat deja de ofrecer el interruptor de “contexto ampliado”. Archivos: `backend/db/chat_inteligencia_artificial.go`, `backend/handlers/super_chat_ia_logica.go`, `web/super/configuracion_logica_del_chat_con_ia.html`, documentación relacionada.

- Chat IA (empresa y super): **Enter** envía el mensaje (mismo flujo que el botón); **Mayús+Enter** añade salto de línea en el textarea. Texto de ayuda bajo el campo. Archivos: `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`.

- Super — **Página principal** (`/super/pagina_principal.html`): sincronización de tema en iframe (cookie/localStorage), fondo del `body` con variables del tema, título de cabecera legible en modo claro y contenedor del editor con superficie neutra (`pp-main-card`) para evitar gradientes rosados en algunos temas oscuros; contraste consistente en todas las apariencias.

- **Permisos por rol** (`/super/permisos_rol.html`): consola empresarial para activar/desactivar acciones por módulo (R/C/U/D/A) y visibilidad por función del menú; API `GET /super/api/roles_de_usuario/permisos` con `modulos_etiqueta`, `acciones_etiqueta` y en cada `pagina` `titulo` y `grupo` (catálogo en `empresa_permisos.go`). **Licencias** (`/super/licencias.html`): sección de cobertura por módulos con descripciones y enlace a la matriz de roles. Modelo: licencia = techo de módulos; rol = matriz y overrides de menú; sin un sistema “universal” duplicado.

- Chat IA (empresa y super): interfaz tipo **Gemini** con barra lateral (chats guardados en el navegador + historial del servidor), barra superior con **modelo en uso** y resumen de cupo diario, **compartir** respuestas del asistente (compartir nativo o copiar), y **mensaje explicativo** cuando se alcanza el límite diario o el chat está bloqueado. Se retiran el encabezado largo y la tarjeta de chips en el chat empresarial; mejor contraste en modo claro. Archivos: `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`, `web/estilos.css`, documentación relacionada.
- Chat IA: **acciones operativas confirmables** desde el chat (chat transaccional). El asistente propone un bloque `PCS_ACTION` (JSON) solo cuando tiene datos completos; el usuario confirma en la UI y el sistema ejecuta el endpoint real (productos, precios, finanzas, tarifas, servicios) manteniendo permisos y **auditoría**. La auditoría crítica registra `source=chat_ia` y `chat_conversation_id` en metadata. Archivos: `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super.go`, `backend/handlers/auditoria_empresa.go`, `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`, `web/estilos.css`.
- Super administrador: el menú agrega **Panel** como primer botón y el panel inicial (`licencias_resumen`) ahora muestra **licencias activas** y **cantidad de empresas registradas**. Se corrige el guardado de correos autorizados para **Frecuencia FP** en super (validación de sesión/rol) y la página se mejora a estilo CRUD (agregar/eliminar/listar). Los administradores en esa lista ven el acceso `Frecuencia FP` dentro de `administrar_empresa/configuracion_menu.html`.
- Licencias/pagos: el checkout de `pagar_licencia.html` registra correctamente el **código de asesor comercial** (`asesor_id`) al crear transacciones (Wompi/Epayco). El backend valida que el asesor exista y esté aceptado, y hace fallback de `empresa_id` desde la licencia si el frontend no lo envía; con esto la comisión queda trazable al aprobarse el pago.

- Estaciones: estación especial **Pedidos con IA** (texto o dictado) que interpreta el pedido y agrega productos al carrito de la estación indicada; configuración `ia_pedidos_enabled` / `ia_pedidos_placement` en `estaciones_config`. Vista móvil con botón **Ver miniaturas** / **Vista normal** (rejilla de 3 columnas). Archivos: `backend/handlers/ia_pedidos_estacion.go`, `backend/db/carritos_compras.go`, `backend/handlers/chat_con_inteligencia_artificial_router.go`, `web/administrar_empresa/estacion_ia_pedidos.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`, `web/estilos.css`, `tools/clean_styles_auto.js`, `.vscode/extensions.json`, documentación relacionada.

- Asesor comercial profesional para licencias.
	- Tipo: mejora funcional mayor.
	- Archivos modificados/agregados/eliminados: `backend/db/asesor_comercial.go`, `backend/handlers/asesor_comercial.go`, `backend/handlers/payments_handlers.go`, `backend/main.go`, `backend/utils/utils.go`, `web/super/asesor_comercial.html`, `web/mis_clientes.html`, `web/pagar_licencia.html`, `web/seleccionar_empresa.html`, `web/js/seleccionar_empresa.js`, `web/super_administrador.html`, `web/super/vendedores_licencias.html` (eliminado), `backend/handlers/vendedores_handlers.go` (eliminado), `backend/handlers/vendedor_config_handlers.go` (eliminado), documentación relacionada.
	- Descripción: se reemplaza el módulo anterior por asesores comerciales con invitación por email, código único, porcentaje y plazo de asociación configurables. Los pagos de licencia con código generan comisión y las renovaciones siguen asociadas hasta vencer el plazo; super puede marcar comisiones pagadas y el asesor ve el estado en `Mis clientes`.

- Chat IA (empresa y super): tema sincronizado con el panel (`pcs_theme` / `localStorage`), estilos del chat basados en variables de tema y eliminación de las sugerencias tipo pill bajo el formulario. Ajustes de contraste en modo claro para calendario y paneles de `chat_y_tareas`.
	- Tipo: ajuste UX.
	- Archivos modificados: `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.

- Red social empresarial y venta pública por páginas.
	- Tipo: mejora mayor.
	- Archivos modificados/agregados: `backend/db/red_social.go`, `backend/handlers/red_social.go`, `backend/db/venta_publica.go`, `backend/handlers/venta_publica.go`, `backend/main.go`, `backend/utils/utils.go`, `web/red_social_comercial.html`, `web/administrar_empresa/publicar_red_social.html`, `web/administrar_empresa/venta_publica.html`, `web/venta_publica.html`, `web/pagar_productos_de_venta_publica.html`, `web/administrar_empresa.html`, `web/administrar_empresa/configuracion_menu.html`, `web/administrar_empresa/configuracion_integraciones.html`, `web/js/administrar_empresa.js`.
	- Descripción: la red social pública ahora funciona como feed empresarial con posts medianos. Venta pública se separa en módulo propio para que cada empresa cree páginas por slug, publique productos existentes y cobre desde una página dedicada con sus credenciales Wompi/Epayco.
	- Verificación: `go test ./...`.

- Soporte remoto: portal publico RustDesk con cliente/servidor y configuracion central desde super.
	- Tipo: mejora mayor.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/handlers/soporte_remoto.go`, `backend/handlers/super_soporte_remoto.go`, `backend/handlers/soporte_remoto_test.go`, `backend/handlers/super_soporte_remoto_test.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/utils/utils.go`, `web/administrar_empresa/soporte_remoto.html`, `web/super/soporte_remoto.html`, `web/soporte_remoto_acceso.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: el mÃ³dulo de soporte remoto ahora puede entregar una pÃ¡gina pÃºblica por sesiÃ³n que funciona como portal de asistencia estilo RustDesk. Esa pÃ¡gina expone descargas del cliente y del servidor, host, clave pÃºblica, ID/contraseÃ±a del dispositivo y visor web opcional. El panel de empresa y la mesa tÃ©cnica super comparten la misma configuraciÃ³n pÃºblica por `empresa_id`, y super puede editarla directamente sin salir de `/super/api/soporte_remoto`.

- Soporte remoto: lÃ­mite diario RustDesk por empresa desde super.
	- Tipo: mejora funcional.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/handlers/soporte_remoto.go`, `backend/handlers/super_soporte_remoto.go`, `backend/db/soporte_remoto_test.go`, `backend/handlers/soporte_remoto_test.go`, `backend/handlers/super_soporte_remoto_test.go`, `web/super/soporte_remoto.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: super ahora puede definir por empresa un tope diario en minutos para conexiones RustDesk. El backend calcula consumo diario real, bloquea nuevas sesiones o aprobaciones que excedan el cupo del dÃ­a y devuelve el resumen de uso para que la mesa tÃ©cnica vea el motivo exacto del bloqueo.

- Apariencia: primera visita del menÃº flotante ahora inicia en Blanco Corporativo.
	- Tipo: ajuste UX.
	- Archivos modificados: `web/menu.js`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: cuando un usuario abre por primera vez una pÃ¡gina con menÃº flotante y todavÃ­a no tiene preferencia guardada en `localStorage`, cookie o backend, el sistema ahora arranca con la apariencia `light` (Blanco Corporativo) en lugar del tema oscuro. Las preferencias ya guardadas siguen respetÃ¡ndose sin cambios.

- Apariencia: contraste corregido para modo claro en portada y estilos compartidos.
	- Tipo: ajuste UX.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: se quitÃ³ el blanco forzado del tÃ­tulo superior del index y se migraron varios bloques compartidos a variables de tema para que botones, tablas, login y secciones pÃºblicas cambien su color correctamente al alternar entre modo oscuro y claro.
# CHANGELOG

## 2026-05-06
- Portal de Terceros y Certificados Tributarios.
	- Se agrego el modulo `portal_terceros_certificados` con maestro de terceros, certificados de retencion/ingresos, enlace publico por token, impresion y bitacora de descargas.
	- Se agregaron las rutas `/api/empresa/portal_terceros_certificados` y `/api/public/certificados_tributarios`.
	- Se agregaron la pantalla administrativa `web/administrar_empresa/portal_terceros_certificados.html`, la pagina publica `web/visualizar_certificado_tributario_publico.html`, permisos, licencia, menu y documentacion.
	- Pruebas: `cd backend; go test ./... -count=1`.

- Activos Fijos e Intangibles NIIF/Fiscal.
	- Se formalizo el modulo `activos_fijos_niif_fiscal` reutilizando el nucleo de activos de la suite contable Colombia avanzada.
	- Se amplio `empresa_contabilidad_activos_fijos` con base fiscal, vida util fiscal, depreciacion fiscal acumulada, valor fiscal, diferencia NIIF/fiscal, deterioro, valor razonable y cuenta de deterioro.
	- Se agrego API `/api/empresa/activos_fijos_niif_fiscal`, pantalla administrativa, enlaces de menu, permisos por rol, licencias y documentacion.
	- Pruebas: `cd backend; go test ./... -count=1`.

- Propiedad horizontal y promocion por asesor.
	- Archivos: `backend/db/propiedad_horizontal.go`, `backend/db/propiedad_horizontal_test.go`, `backend/handlers/propiedad_horizontal.go`, `backend/handlers/asesor_comercial.go`, `backend/handlers/payments_handlers.go`, `backend/handlers/empresa_permisos.go`, `backend/main.go`, `web/administrar_empresa/propiedad_horizontal.html`, `web/super/asesor_comercial.html`, `web/pagar_licencia.html`, `web/super/licencias.html`, `documentos/propiedad_horizontal.md`, `documentos/promocion_asesor_licencias.md`.
	- Descripcion: se agrega modulo profesional de administracion de copropiedades con unidades, residentes, cargos, recaudos, PQR, asambleas y dashboard. Se agrega promocion global para que un codigo de asesor aceptado aplique descuento adicional configurable al comprar licencias.
	- Verificacion: `go test ./... -count=1` en `backend/`.

- Cierre y bloqueo fiscal avanzado.
	- Archivos: `backend/db/cierre_fiscal.go`, `backend/db/cierre_fiscal_test.go`, `backend/handlers/cierre_fiscal.go`, `backend/handlers/empresa_permisos.go`, `backend/db/contabilidad_colombia.go`, `backend/main.go`, `web/administrar_empresa/cierre_fiscal.html`, `web/administrar_empresa.html`, `web/administrar_empresa/finanzas_menu.html`, `web/js/administrar_empresa.js`, `web/super/licencias.html`, `documentos/cierre_fiscal.md`.
	- Descripcion: se agrega modulo profesional para periodos fiscales, politicas de bloqueo por modulo, dias de edicion retroactiva, reaperturas con motivo, excepciones aprobadas, validacion de operaciones y bitacora de bloqueos por empresa. El cierre/reapertura de Contabilidad Colombia sincroniza el periodo fiscal.
	- Verificacion: `go test ./... -count=1` en `backend/`.

- Centros de costo y rentabilidad.
	- Archivos: `backend/db/centros_costo.go`, `backend/handlers/centros_costo.go`, `backend/handlers/empresa_permisos.go`, `backend/main.go`, `web/administrar_empresa/centros_costo.html`, `web/administrar_empresa.html`, `web/administrar_empresa/finanzas_menu.html`, `web/js/administrar_empresa.js`, `documentos/centros_costo.md`.
	- Descripcion: se agrega modulo formal `centros_costo` para medir rentabilidad por sucursal, area, unidad de negocio o proyecto, con maestro, reglas de imputacion, presupuesto por periodo, dashboard comparativo, movimientos integrados desde contabilidad/tesoreria/compras/OCR/AIU y exportacion CSV, sin duplicar Finanzas ni Contabilidad.
	- Verificacion: `go test ./... -count=1` en `backend/`.

- QA transversal de modulos y profesionalizacion operativa.
	- Archivos: `backend/db/cobranza.go`, `backend/db/portal_contador.go`, `backend/db/soportes_compras_ia.go`, `web/js/administrar_empresa.js`, `web/administrar_empresa/soportes_compras_ia.html`, `web/index.html`, `documentos/reporte_qa_modulos_2026-05-06.md`.
	- Descripcion: se alinean permisos locales para `administrador_total`, se optimizan dashboards nuevos para no repetir validaciones de esquema, se endurecen enlaces dinamicos de soportes IA y se documenta la prueba autenticada de Motel Calipso con paginas/API 200. La portada publica actualiza la descripcion de modulos para incluir Cobranza, Portal contador, Captura IA/OCR, AIU construccion, Parqueaderos con ticket QR y Apartamentos turisticos.
	- Verificacion: `go test ./... -count=1`, `git diff --check`, auditoria estatica de enlaces/botones/IDs y pruebas HTTP autenticadas contra `empresa_id=7`.

- AIU construccion: mejora profesional del modulo de contratos de obra.
	- Archivos: `backend/db/aiu_construccion.go`, `backend/handlers/aiu_construccion.go`, `web/administrar_empresa/aiu_construccion.html`, pruebas y documentacion.
	- Descripcion: se agregan responsable, centro de costo, modalidad contractual, riesgo, avance, retenciones, anticipo, garantia, neto a cobrar, flujo validado de estados/aprobacion, reporte, facturas recientes, alertas y exportacion CSV. El calculo AIU ahora separa total de factura y neto operativo a cobrar.
	- Verificacion: `go test ./db -run Test.*AIU -count=1` y `go test ./...`.

- AIU construccion: modulo profesional para arquitectos, constructoras y contratos de obra.
	- Archivos: `backend/db/aiu_construccion.go`, `backend/handlers/aiu_construccion.go`, `web/administrar_empresa/aiu_construccion.html`, menu de facturacion, permisos, pruebas y documentacion.
	- Descripcion: se agrega gestion de contratos AIU por empresa, capitulos/conceptos de obra, calculo de Administracion/Imprevistos/Utilidad, modelos de base AIU no sumada o sumada al total, base IVA configurable y generacion de factura electronica AIU enlazada a `empresa_facturacion_documentos`.
	- Verificacion: `go test ./db -run Test.*AIU -count=1` y `go test ./...`.

- Facturacion electronica Colombia: se amplian los documentos electronicos del ciclo DIAN/proveedor para cubrir brechas frente a Siigo y al ecosistema DIAN.
	- Archivos: `backend/handlers/facturacion_electronica.go`, `backend/handlers/documentos_lifecycle.go`, `backend/db/facturacion_electronica.go`, `web/administrar_empresa/facturacion_electronica.html`, pruebas y documentacion.
	- Descripcion: el mismo modulo por empresa ahora permite emitir/anular o registrar en cola fiscal factura electronica, nota credito, nota debito, documento soporte, nomina electronica y documento equivalente POS electronico. Se agrega selector operativo en la UI, botones rapidos para cada documento y normalizacion de aliases usados por Siigo/DIAN.
	- Verificacion: `go test ./handlers -run "TestNormalizeFacturacionDocumentoElectronicoTipo|TestResolveFacturacionTransitionForDocumentosElectronicosNuevos"` y `go test ./db -run TestDefaultFacturacionConfigPaisAplicaProveedorYCampos`.

## 2026-05-04
- Asesor comercial: configuracion profesional de transferencias y pagos de comisiones por venta de licencias.
	- Archivos: `backend/db/asesor_comercial.go`, `backend/handlers/asesor_comercial.go`, `web/super/asesor_comercial.html`, documentacion relacionada.
	- Descripcion: el modulo de asesor comercial ahora guarda metodo de pago, entidad financiera, tipo/numero de cuenta, titular, documento, contacto de pagos, periodicidad, dia de pago, minimo a transferir y exigencia de soporte. Las comisiones de licencias incorporan estado de pago (`pendiente`, `programada`, `en_transferencia`, `pagada`, `rechazada`), metodo, referencia, fecha programada y soporte.
	- Verificacion: `go test ./...` en `backend/`, validacion de scripts embebidos de `web/super/asesor_comercial.html` y `git diff --check`.

## 2026-04-30
- Raspberry Pi y sensores: provisionamiento seguro por empresa.
	- Archivos: `backend/db/sensor_puertas.go`, `backend/handlers/sensor_puertas.go`, `web/administrar_empresa/configuracion_sensores_raspberry.html`, pruebas y documentacion relacionada.
	- Descripcion: la configuracion de sensores ahora permite provisionar desde el servidor un `device_id` normalizado y un token de 64 caracteres para `X-Device-Token`, con ejemplos curl/Python para instalar en Raspberry Pi. Los dispositivos con token configurado ya no aceptan heartbeats ni mensajes publicos enviados solo con `device_id`.
	- Verificacion: `go test ./db -run "TestNormalizeEmpresaSensor|TestGenerateEmpresaSensor" -count=1`, `go test ./handlers -run "TestBuildEmpresaSensorProvisioningPayload" -count=1`, validacion del script embebido de la pagina y `go test ./... -count=1`.

- Chat IA: exportacion de documentos generados desde conversaciones.
	- Archivos: `backend/handlers/dynamic_documents.go`, `backend/handlers/chat_con_inteligencia_artificial_router.go`, `backend/handlers/dynamic_documents_test.go`, `web/js/ai_chat_drawer.js`, `web/estilos.css`, documentacion relacionada.
	- Descripcion: las respuestas del chat IA que parecen documentos, reportes, contratos, cotizaciones, actas o tablas muestran botones de exportacion PDF, Word/DOCX, Excel/XLSX, TXT y JSON. El frontend llama a `/api/empresa/chat_documentos/exportar`, el backend reutiliza el generador dinamico, redacta patrones de secretos, prepara nombres profesionales con empresa/tipo/fecha, registra auditoria con origen `chat_ia` y conserva fallback TXT/JSON si falla un conversor.
	- Verificacion: `go test ./handlers -run "TestDynamicDocument" -count=1`, `go test ./handlers -count=1`, `node --check web/js/ai_chat_drawer.js` con runtime Node empaquetado.

- Checkout de licencias Epayco: fallback clasico seguro por POST.
	- Archivos: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, documentacion relacionada.
	- Descripcion: cuando Smart Checkout v2 no autentica o no crea sesion, el backend ya no entrega una URL GET a `checkout.epayco.co/checkout.php`. Ahora genera un formulario clasico firmado para `https://secure.payco.co/checkout.php`, devuelve `checkout_form` al frontend y registra en `raw_payload` una version sanitizada sin exponer `p_key`. El modo del formulario clasico se resuelve con `epayco.customer_id` + `epayco.checkout_key`/`epayco.p_key`, separado del modo Smart Checkout, para evitar enviar comercios reales como solicitud de pruebas (`p_test_request=true`) y provocar "El comercio no fue reconocido". Si falta `epayco.customer_id` o `P_KEY`, devuelve error controlado.
	- Verificacion: `go test ./handlers -run "TestBuildEpaycoClassicCheckoutForm|TestResolveEpaycoClassicMode|TestPickEpaycoField|TestSanitizeEpaycoClassicCheckoutForm" -count=1`, `go test ./handlers -count=1`, `go test ./...` en `backend/`, `node --check` del script inline de `web/pagar_licencia.html` y `git diff --check`.

- Chat flotante: secretaria IA rediseñada y voz femenina.
	- Archivos: `web/js/ai_chat_drawer.js`, `web/estilos.css`, `web/administrar_empresa/configuracion_chat_flotante.html`, documentacion relacionada.
	- Descripcion: el avatar `Secretaria IA 3D` pasa a una apariencia estilo caricatura ejecutiva joven, con rostro mas amable, ropa ejecutiva, detalles visuales y animaciones existentes. Cuando el modo activo es secretaria, la voz efectiva se fuerza a `es-CO-female` para el proxy de voz y para Web Speech; el robot conserva la voz configurable.
	- Verificacion: `node --check web/js/ai_chat_drawer.js` y `git diff --check`.

- Empresas compartidas: visibilidad y eliminacion de accesos.
	- Archivos: `backend/handlers/empresa_compartida_handlers.go`, `web/js/editar_empresa.js`, `web/js/seleccionar_empresa.js`, documentacion relacionada.
	- Descripcion: el editor de empresa lista los administradores con quienes se compartio la empresa y permite retirar accesos compartidos. La informacion queda visible para quien compartio y para quien recibio acceso; la revocacion mantiene trazabilidad del actor y del motivo operativo.
	- Verificacion: `go test ./...`.

- Documentos dinamicos asistidos por IA.
	- Archivos: `backend/handlers/dynamic_documents.go`, `backend/handlers/dynamic_documents_test.go`, `backend/main.go`, documentacion relacionada.
	- Descripcion: se agregan endpoints protegidos `/generate` y `/download` para recibir contenido o prompt IA, aplicar variables, renderizar HTML con templates Go y descargar PDF, DOCX, XLSX, HTML, TXT o JSON. La generacion usa GPT-5.4 mini cuando se solicita IA y conserva flujo temporal de archivos sin exponer credenciales al modelo.
	- Verificacion: pruebas unitarias dirigidas del handler de documentos dinamicos.

## 2026-04-29
- Reportes globales y busqueda en configuracion avanzada.
	- Archivos: `backend/handlers/reportes_globales.go`, `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js`, `web/super/configuracion_avanzada.html`.
	- Descripcion: reportes globales en seleccionar empresa ahora permite a la cuenta super principal ver todas las empresas, corrige textos con codificacion danada y agrega manejo mas claro de respuestas no JSON. Configuracion avanzada incorpora un submenu sticky con buscador por categoria y seccion.
- Voz IA streaming segura.
	- Archivos: `backend/handlers/voice_stream_config.go`, `backend/handlers/super_config_backup_handlers.go`, `services/voice_stream_server/app.py`, `services/voice_stream_server/README.md`, `scripts/install_voice_stream_server_vps.sh`, `web/super/voz_streaming_ia.html`, `web/super/configuracion_avanzada.html`.
	- Descripcion: el servicio de voz puede activarse/probarse desde super administrador, registra token cifrado en configuraciones, envia autenticacion al health/TTS y el microservicio puede exigir el mismo token por header.
- Epayco y chat flotante movil.
	- Archivos: `web/pagar_licencia.html`, `web/administrar_empresa.html`, `web/seleccionar_empresa.html`, `web/super_administrador.html`, `web/estilos.css`.
	- Descripcion: el boton `Pagar con Epayco` precarga Smart Checkout v2, abre con tipo compatible para movil y prueba fallback entre `onpage`/`standard` si la primera apertura falla. El chat flotante mueve `Nuevo chat` al toolbar inferior y compacta en una sola fila nuevo mensaje, modo conversacion, microfono, voz, selector de modo y adjuntar.
	- Verificacion: validacion de scripts HTML/JS con Node y `go test ./handlers`.

- Checkout de licencias: tarjetas comerciales y asesor bajo descuento.
	- Archivos: `web/pagar_licencia.html`, `web/estilos.css`.
	- Descripcion: las tarjetas del pago de licencia quedan con presentacion mas profesional de venta, beneficios visibles y precio destacado. El campo `Codigo de asesor comercial` queda debajo del `Codigo de descuento`, conservando el envio de `asesor_id` a las pasarelas y activacion sin pago.
	- Verificacion: validacion de scripts de `web/pagar_licencia.html` con Node.

- Preconfiguracion de tipos de empresa.
	- Archivos: `backend/db/tipo_empresa_preconfiguracion.go`, `backend/handlers/empresa_preconfiguracion.go`, `backend/db/empresa_estacion_prefs.go`, `backend/db/productos.go`, `backend/handlers/system_empresas_handlers.go`, `backend/main.go`, `web/super/preconfiguracion_tipos_empresa.html`, `web/js/seleccionar_empresa.js`, `web/super_administrador.html`, `Pendiente Notas`, documentacion relacionada.
	- Descripcion: Super Administrador puede definir plantillas iniciales por tipo de empresa. Al crear una empresa se aplican estaciones y productos guia si la plantilla esta activa; restaurante trae 5 mesas y productos demo. Tras crearla, el administrador decide si conserva la preconfiguracion o la elimina para dejar la empresa limpia.
	- Verificacion: `go test ./db`, `go test ./handlers` y validacion de scripts de `seleccionar_empresa` / `preconfiguracion_tipos_empresa`.

- Estaciones: tarjetas sin filas vacias.
	- Archivos: `web/administrar_empresa/estaciones.html`, `web/estilos.css`, documentacion relacionada.
	- Descripcion: las tarjetas de estaciones normales reorganizan sus datos en filas etiqueta/valor tipo tarjeta Caja. Las filas sin dato quedan ocultas, conservando una sola columna de informacion ordenada por estado, cliente, tarifa, duracion, inicio, fin, extra y total.

- Voz IA streaming y reorganizacion super.
	- Archivos: `backend/handlers/voice_stream_config.go`, `backend/main.go`, `services/voice_stream_server/*`, `scripts/install_voice_stream_server_vps.sh`, `web/js/ai_chat_drawer.js`, `web/super/voz_streaming_ia.html`, `web/super_administrador.html`, `web/super/configuracion_avanzada.html`, documentacion relacionada.
	- Descripcion: se agrega un servicio abierto FastAPI + Piper TTS para voz natural en VPS, configurable desde Super Administrador y desactivado por defecto. El backend expone configuracion protegida, estado publico no sensible y proxy TTS; el chat flotante intenta usar el servidor de voz y cae a voz del navegador/texto si esta desactivado o falla. El menu super y Configuracion avanzada quedan agrupados por categorias empresariales.

- Tarjetas del index con modo informacion/foto o banner.
	- Archivos: `backend/handlers/pagina_principal_handlers.go`, `web/super/pagina_principal.html`, `web/index.html`, `web/estilos.css`, documentacion relacionada.
	- Descripcion: la configuracion super de Pagina principal permite elegir por tarjeta entre `tarjeta informacion mas foto` (default compatible con lo existente) y `tarjeta banner`. El modo banner usa solo la imagen principal y la renderiza ocupando toda la tarjeta del index. La pantalla permite subir imagenes a `web/img` y muestra medidas recomendadas 4:3 por tamano de tarjeta.
	- Verificacion: `go test ./...` en `backend/` compilo y ejecuto paquetes; Windows bloqueo temporalmente el borrado del binario de test al final, luego se limpio `.gotmp6`.

- Reorganizacion empresarial de Administrar Empresa y Configuracion.
	- Archivos: `web/administrar_empresa.html`, `web/administrar_empresa/configuracion_menu.html`, `web/administrar_empresa/configuracion.html`, `web/js/administrar_empresa.js`, `web/estilos.css`, documentacion relacionada.
	- Descripcion: el menu principal queda agrupado por modulos de trabajo (colaboracion, operacion/ventas, inventario/compras, finanzas/cumplimiento, personas/activos, analisis/control, documentos/nube/soporte y administracion). Configuracion se reorganiza por base empresarial, ventas/cobro, estaciones/tarifas, fiscal/automatizacion y avanzado. Se conserva la compatibilidad con rutas, iframes y permisos, y se eliminan duplicados de navegacion.

- Auditoria IA profunda y consultas DB seguras para GPT-5.4 mini/modelo activo.
	- Archivos: `backend/db/auditoria_empresa.go`, `backend/db/chat_inteligencia_artificial.go`, `backend/handlers/auditoria_empresa.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super.go`, `backend/handlers/super_chat_ia_logica.go`, `web/super/configuracion_logica_del_chat_con_ia.html`, documentacion relacionada.
	- Descripcion: el chat empresarial y el chat global reciben auditoria en tiempo real, busqueda profunda de eventos por intencion y resultados de consultas DB seguras ejecutadas por el backend con whitelist/empresa_id. Se agrega `empresa_auditoria_ia_consultas` y lectura total controlada de DB para el chat empresarial, activa por defecto y configurable desde super (`ai.chat.empresa.db_query_enabled`, tablas maximas y filas por tabla). La IA no ejecuta SQL libre y el servidor degrada sin romper si auditoria, lectura DB o proveedor IA fallan.
	- Verificacion: `go test ./...` en `backend/` usando `GOTMPDIR=.gotmp4`.

## 2026-04-25
- Facturación electrónica: perfiles independientes para Ecuador (SRI) y Panamá (DGE/DGI) frente a Colombia (DIAN); detección por licencia; API y UI con `vista` por país.
	- Archivos: `backend/db/facturacion_electronica.go`, `backend/handlers/facturacion_electronica.go`, `web/administrar_empresa/facturacion_electronica.html`, `web/estilos.css`, `documentos/historial_de_cambios`, `documentos/descripcion_de_modulos`, `documentos/descripcion_de_archivos`, `CHANGELOG.md`.

- Configuración lógica del chat global (super): contexto ampliado de la base superadministrador y modo solo lectura de datos de empresas inyectados al prompt; ajuste de legibilidad en estaciones (vista miniaturas móvil).
	- Archivos: `web/super/configuracion_logica_del_chat_con_ia.html`, `backend/handlers/super_chat_ia_logica.go`, `backend/handlers/chat_con_ia_global_super.go`, `backend/db/chat_inteligencia_artificial.go`, `web/estilos.css`, `documentos/historial_de_cambios`, `documentos/descripcion_de_archivos`, `documentos/descripcion_de_modulos`, `CHANGELOG.md`.

## 2026-04-23
- Retiro del modulo Tipos de usuario en panel super: API, handlers, pagina web y tabla `tipos_de_usuario` (DROP al arranque); solo permanecen roles y permisos por rol.
	- Archivos modificados/eliminados: `backend/main.go`, `backend/handlers/roles_tipos_usuario.go`, `backend/db/roles_tipos_usuario.go`, `web/super_administrador.html`, `web/js/super_administrador.js`, `web/super/tipos_de_usuario.htm` (eliminado), `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.

- Documentacion y operacion reCAPTCHA (v2/v3/Enterprise) + copias de respaldo bajo `backup/`.
	- Archivos modificados: `documentos/manual_de_instalacion.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `documentos/CHANGELOG.md`, `CHANGELOG.md`, `backup/.gitkeep`, `backup/empresas/.gitkeep`, `backup/super_administrador/.gitkeep`.
	- Descripcion: el manual ahora documenta la configuracion super, el provider y los mensajes tipicos de dominio en Google, y describe el almacenamiento best-effort de copias JSON en `backup/super_administrador` y `backup/empresas/<empresa_id>`.

## 2026-04-20
- Soporte remoto y RustDesk: activacion por defecto en configuracion super.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/handlers/soporte_remoto.go`, `backend/db/soporte_remoto_test.go`, `backend/handlers/super_soporte_remoto_test.go`, `web/super/servidores.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: la configuración sin registro previo del módulo de soporte remoto ahora nace activa por defecto y alineada a la operación simplificada de RustDesk, incluyendo portal público habilitado, descargas completas y modo `cliente_local` sobre `rustdesk_oss`. La vista super mantiene el mismo default en primera carga para no mostrar un falso inactivo.
	- Verificación: `go test ./db -run '^TestSoporteRemotoDBFlow$' -count=1` y `go test ./handlers -run '^TestSuperSoporteRemotoHandlerConfigGetAndUpdate$' -count=1`.

## 2026-04-21
- Compras y finanzas: comprobantes adjuntos por empresa.
	- Archivos modificados: `backend/db/documentos_transaccionales.go`, `backend/db/finanzas.go`, `backend/handlers/compras.go`, `backend/handlers/finanzas.go`, `backend/handlers/compras_documentos_test.go`, `backend/handlers/eventos_contables_modulos_test.go`, `backend/main.go`, `web/administrar_empresa/compras.html`, `web/administrar_empresa/finanzas.html`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripción: el sistema permite adjuntar recibos, fotos y soportes documentales a documentos de compras y a movimientos de finanzas, guardándolos por empresa en `web/uploads/comprobantes/empresa_<id>/compras` y `.../finanzas`. Los listados empresariales ahora muestran acceso directo para ver el comprobante, y compras amplía su persistencia con metadata explícita del archivo cargado.
	- Verificación: diagnóstico del editor sin errores en los archivos tocados. La corrida dirigida de `go test` quedó bloqueada por un error preexistente de compilación en `backend/db/carritos_compras.go` ajeno a este cambio.

## 2026-04-20
- Apariencia frontend: limpieza estructural final en ayuda y vendedores.
	- Archivos modificados: `web/estilos.css`, `web/super/vendedores_licencias.html`, `web/ayuda/ayuda.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: `vendedores_licencias.html` elimina los `style=` residuales de layout y estados, mientras `ayuda.html` se normaliza por secciones completas de encabezados y pÃ¡rrafos auxiliares. El resultado es una herencia de tema mÃ¡s consistente y menos deuda visual repetida en pÃ¡ginas largas del panel super y de ayuda.
	- VerificaciÃ³n: diagnÃ³stico del editor sin errores en los tres archivos tocados y bÃºsqueda sin remanentes de `style="` en `web/super/vendedores_licencias.html` ni en `web/ayuda/ayuda.html`.

- Soporte remoto y RustDesk: simplificaciÃ³n operativa del mÃ³dulo y control real del servicio desde super.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/handlers/soporte_remoto.go`, `backend/handlers/super_soporte_remoto.go`, `backend/handlers/super_servidores_handlers.go`, `backend/main.go`, `backend/handlers/soporte_remoto_test.go`, `backend/handlers/super_soporte_remoto_test.go`, `web/super/servidores.html`, `web/administrar_empresa/soporte_remoto.html`, `web/super/soporte_remoto.html`, `web/soporte_remoto_acceso.html`, `documentos/descripcion_de_archivos`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: el frente de soporte remoto se simplifica para dejar solo la configuraciÃ³n esencial de RustDesk por empresa y el panel super pasa a tener control operativo del servicio del VPS con prueba real. Se agregan URLs macOS, se consolidan descargas oficiales para cliente y servidor y se elimina de las pantallas la lÃ³gica extensa de visor/sesiones como superficie principal. La UI super se unifica ademÃ¡s en una sola vista `RustDesk`, donde conviven estado del servidor, acciones del VPS y configuraciÃ³n mÃ­nima por empresa.
	- VerificaciÃ³n: `get_errors` sin errores en las cuatro vistas HTML nuevas y `go test ./handlers -run 'Test(SuperSoporteRemotoHandlerConfigGetAndUpdate|PublicSoporteRemotoResolverAccesoExponeDescargasRustDesk|SuperServidoresProbeHandlerReturnsRustDeskStatus|SuperSoporteRemotoHandlerListsCompaniesAndCreatesSession)$' -count=1`.

## 2026-04-20
- Portal publico: el header del home ahora comparte estilo con la landing y expone `Crear cuenta`.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: `index.html` reemplaza el botÃ³n superior propio por el mismo estilo `btn dark` usado en `descripcion_de_los_sistemas.ht`, agrega un CTA `Crear cuenta` hacia `/registrar_nuevo_usuario_administrador.html` junto a `Iniciar sesiÃ³n` y ajusta el responsive del header para que ambos botones se mantengan compactos y legibles en mÃ³vil.
	- VerificaciÃ³n: diagnÃ³stico del editor sin errores en `web/index.html` y `web/estilos.css`.

- Apariencia frontend: segunda pasada de limpieza de estilos inline en mÃ³dulos operativos.
	- Archivos modificados: `web/estilos.css`, `web/login.html`, `web/login_usuario.html`, `web/accept.html`, `web/configuracion_de_la_cuenta.html`, `web/administrar_empresa/administrar_productos.html`, `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/administrar_clientes.html`, `web/administrar_empresa/administrar_usuarios.html`, `web/administrar_empresa/compras.html`, `web/administrar_empresa/facturas_electronicas.html`, `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/red_social_comercial.html`, `web/administrar_empresa/graficos_estadisticas.html`, `web/administrar_empresa/estaciones.html`, `web/js/super_reportes_globales.js`, `web/super/chat_con_ia_global.html`, `web/super/configuracion_avanzada.html`, `web/super/reportes_globales.html`, `web/super/vendedores_licencias.html`, `web/ayuda/login_administradores.html`, `web/administrar_empresa/soporte_remoto_view.html`, `web/administrar_empresa/historial_productos.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: se reemplazan estilos inline de color, gradiente, ocultaciÃ³n, espaciado y alineaciÃ³n por utilidades CSS y clases semÃ¡nticas para reducir deuda visual y mejorar consistencia entre los seis temas del sistema. Parte del HTML generado desde JavaScript ahora usa clases y `data-*`, y los mensajes de error/Ã©xito en runtime dejan de depender de hex fijos.
	- VerificaciÃ³n: diagnÃ³stico del editor sin errores en los archivos principales tocados, bÃºsqueda sin remanentes de `style="...color/background..."` en el frente intervenido, sin asignaciones `style.color = "#..."` en los JS afectados y corrida dirigida `go test ./handlers -run 'Test(SuperSoporteRemotoHandlerConfigGetAndUpdate|PublicSoporteRemotoResolverAccesoExponeDescargasRustDesk|SuperServidoresProbeHandlerReturnsRustDeskStatus)$' -count=1` satisfactoria.

- Permisos super: pruebas dirigidas migradas a PostgreSQL y middleware robustecido.
	- Archivos modificados: `backend/handlers/postgres_test_helpers_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/handlers/auth_admin_handlers_test.go`, `backend/utils/utils.go`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: el bloque de pruebas de permisos del panel super ya no depende de motor legado retirado y ahora usa esquemas efÃ­meros en PostgreSQL, alineados con el tÃºnel local documentado del proyecto. En la misma iteraciÃ³n, `AuthMiddleware` deja de fallar cuando no hay conexiÃ³n `dbSuper` en flujos pÃºblicos, evitando un `panic` en la validaciÃ³n de rutas pÃºblicas de licencias.
	- VerificaciÃ³n: `go test ./handlers -run 'TestSuperEndpointsPermisosPorRol|TestAdministradorPuedeEditarYEliminarEmpresaDesdeRutaSuperProtegida|TestNuevoAdminRegistradoPuedeCrearSuPrimeraEmpresaViaRutaSuperProtegida' -count=1` y `go test ./utils -run '^TestAuthMiddlewareAllowsPublicLicenciaPaymentRoutesWithoutSession$' -count=1`.

- Apariencia global: contraste y componentes alineados en los seis temas.
	- Archivos modificados: `web/estilos.css`, `web/administrar_empresa/soporte_remoto.html`, `web/super/soporte_remoto.html`, `web/administrar_empresa/publicar_red_social.html`, `web/red_social_comercial.html`, `web/pantalla_publica.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: el sistema deja de depender de varios colores fijos que rompÃ­an el contraste en modo claro y pasa a usar variables de tema para paneles, tarjetas, mÃ³dulos embebidos, estaciones especiales y estados informativos. Con esto, los seis modos de apariencia mantienen mejor legibilidad y coherencia visual.
	- VerificaciÃ³n: diagnÃ³stico del editor sin errores en los archivos frontend tocados.

- Login administrativo: recuperaciÃ³n por enlace directo sin token manual.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/super_email_templates.go`, `backend/handlers/auth_admin_handlers_test.go`, `web/login.html`, `web/js/login.js`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: el correo de recuperaciÃ³n administrativa ahora dirige al usuario a un enlace directo de restablecimiento y la pantalla de `login.html` ya no solicita ingresar token manual. El token sigue validÃ¡ndose en backend, pero queda encapsulado en el enlace recibido por correo.
	- VerificaciÃ³n: prueba dirigida del handler administrativo de recuperaciÃ³n/restablecimiento y diagnÃ³stico del editor sin errores en archivos tocados.

- Correo de recuperaciÃ³n: el enlace largo se reemplaza por un botÃ³n `Cambiar contraseÃ±a`.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/super_email_templates.go`, `backend/handlers/auth_admin_handlers_test.go`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: la versiÃ³n HTML de los correos de recuperaciÃ³n ahora prioriza un botÃ³n `Cambiar contraseÃ±a` y deja la URL extensa solo como respaldo visible. El envÃ­o administrativo tambiÃ©n se genera como `multipart/alternative` para servir texto plano y HTML en el mismo mensaje.
	- VerificaciÃ³n: `go test ./handlers -run '^TestAdminPasswordRecoveryTemplateRendersButton$' -count=1` y diagnÃ³stico del editor sin errores en `backend/handlers/auth_admin_handlers.go` y `backend/handlers/super_email_templates.go`.

- Backups empresariales: se implementa la Fase 2 de exportar/importar configuracion por empresa.
	- Archivos modificados: `backend/db/backups_empresariales.go`, `backend/handlers/backups_empresariales.go`, `backend/handlers/backups_empresariales_test.go`, `web/administrar_empresa/backups.html`, `backend/.env.example`, `backend/main.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: el mÃ³dulo de backups ahora exporta e importa configuraciÃ³n completa por empresa en JSON canÃ³nico, con restauraciÃ³n sobre PostgreSQL y trazabilidad del origen importado. En la misma iteraciÃ³n se limpian referencias operativas falsas a motor legado retirado del entorno y de comentarios de arranque.
	- VerificaciÃ³n: prueba dirigida del handler de configuraciÃ³n empresarial y diagnÃ³stico del editor sin errores en los archivos tocados.

- PostgreSQL: gobernanza endurecida y limpieza de soporte residual motor legado retirado.
	- Archivos modificados: `.github/agents/agente_go.agent.md`, `.github/agents/agente_backend_db.agent.md`, `copilot-instructions.md`, `backend/db/compat_wrappers.go`, `backend/db/sql_compat.go`, `backend/db/horarios_trabajadores.go`, `.gitignore`, `documentos/descripcion_del_proyecto`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: el repositorio formaliza PostgreSQL como Ãºnico motor permitido en reglas de agentes e instrucciones del proyecto. TambiÃ©n se elimina el fallback motor legado retirado en memoria del paquete `db`, el dialecto SQL por defecto deja de aceptar motor legado retirado y el esquema de horarios queda solamente en sintaxis PostgreSQL.

- Estaciones: la tarjeta `Notas` ahora soporta mÃºltiples recordatorios persistentes con repeticiÃ³n automÃ¡tica local.
	- Archivos modificados: `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: la estaciÃ³n especial `Notas` deja la lÃ³gica de una sola nota y pasa a manejar varias notas dentro de la misma tarjeta, con selecciÃ³n activa, temporizadores independientes, restauraciÃ³n del countdown tras recargar y repeticiÃ³n automÃ¡tica configurable por minutos. El valor por defecto de repeticiÃ³n entra a `estaciones_config`, mientras el runtime mÃºltiple se persiste localmente por `empresa_id`.
	- VerificaciÃ³n: diagnÃ³stico del editor sin errores en los archivos frontend tocados; revisiÃ³n de QA y backend sin cambios contractuales obligatorios en servidor.

- PostgreSQL: se cierra la Fase 1 de limpieza postmigracion del repositorio.
	- Archivos modificados: `backend/db/pcs_superadministrador`, `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos`, `documentos/erp_multiempresa/02_diseno_tecnico_erp_multiempresa.md`, `Pendiente Notas`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se elimina el ultimo `.db` legacy que seguia versionado en el repo, se confirman las referencias activas y se corrige la documentacion vigente para dejar PostgreSQL como unica base operativa canonica. El backlog deja Fase 1 completada y aterriza las fases 2 y 3.

- Estaciones: se agrega la estaciÃ³n especial Notas y el orden configurable de estaciones especiales.
	- Archivos modificados: `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: el mÃ³dulo de estaciones ahora permite decidir si `Caja`, `YouTube` y la nueva estaciÃ³n especial `Notas` se cargan antes o despuÃ©s del listado normal. `Notas` aÃ±ade una tarjeta operativa con texto editable, temporizador programable, guardado del texto base y alerta visual/sonora cuando vence el recordatorio.
	- VerificaciÃ³n: diagnÃ³stico del editor sin errores en los archivos frontend tocados.

- Documentacion: se depura `Pendiente Notas` y queda como backlog vigente.
	- Archivos modificados: `Pendiente Notas`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: se elimina del archivo la mezcla de credenciales y notas obsoletas, se retiran tareas ya implementadas y se reorganiza lo faltante en un plan priorizado con fases de ejecuciÃ³n y criterio de cierre.
	- VerificaciÃ³n: revisiÃ³n documental contra `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md` y el estado actual del repositorio.

- Login administrativo: se corrige la activacion del ojito para mostrar la contraseÃ±a.
	- Archivos modificados: `web/js/login.js`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: la inicializaciÃ³n del toggle de visibilidad quedÃ³ insertada en el lugar incorrecto dentro de `showMsg`, asÃ­ que el control podÃ­a aparecer sin activar el cambio real del input. Ahora `initPasswordVisibilityToggles()` se ejecuta en el flujo principal de carga de `login.js`.
	- VerificaciÃ³n: diagnÃ³stico del editor sin errores en `web/js/login.js`.

- Super y portal pÃºblico: el WhatsApp del botÃ³n flotante de la portada ahora es configurable.
	- Archivos modificados: `backend/handlers/usuarios_empresa.go`, `backend/handlers/pagina_principal_handlers.go`, `backend/handlers/pagina_principal_handlers_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/handlers/super_config_backup_handlers.go`, `web/super/configuracion_avanzada.html`, `web/index.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: se agrega una tarjeta en configuraciÃ³n avanzada para editar el nÃºmero del botÃ³n flotante de WhatsApp de la portada pÃºblica. El valor queda persistido en configuraciÃ³n super y el `index.html` ahora lo consume desde `/api/public/pagina_principal` en vez de mantenerlo fijo en el marcado.
	- VerificaciÃ³n: `go test ./handlers -run "Test(PublicPaginaPrincipalHandlerExposesLandingFields|GmailConfigHandlerSaveWhatsAppContactNumber)$" -count=1`.

- Login administrativo: se agrega control visual para mostrar u ocultar la contraseÃ±a.
	- Archivos modificados: `web/login.html`, `web/js/login.js`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: el campo de contraseÃ±a de `/login.html` ahora incluye un botÃ³n tipo ojo que permite alternar la visibilidad del texto escrito, reutilizando el mismo patrÃ³n visual de contraseÃ±as usado en otros formularios del portal.
	- VerificaciÃ³n: inspecciÃ³n estÃ¡tica de `web/login.html` y `web/js/login.js`.

- Portal pÃºblico: la landing descriptiva elimina el hero superior y la tarjeta de accesos rÃ¡pidos.
	- Archivos modificados: `web/descripcion_de_los_sistemas.ht`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: `/descripcion_de_los_sistemas.ht` ahora abre directamente en el contenido detallado de soluciones. Se retiran el bloque `Catalogo unificado` y toda la tarjeta de `Accesos rapidos`, manteniendo el soporte de hash y los CTA pÃºblicos existentes.
	- VerificaciÃ³n: inspecciÃ³n estÃ¡tica de `web/descripcion_de_los_sistemas.ht` y `web/estilos.css`.

- Portal pÃºblico: la landing descriptiva ahora adopta la apariencia activa y un menÃº interno mÃ¡s profesional.
	- Archivos modificados: `web/descripcion_de_los_sistemas.ht`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃ³n: la zona `Accesos rapidos` de `/descripcion_de_los_sistemas.ht` deja la grilla plana y pasa a un menÃº ejecutivo con numeraciÃ³n, subtÃ­tulo y estado activo segÃºn la secciÃ³n visible. TambiÃ©n se reemplazan fondos y colores fijos por variables del sistema de apariencia para que la landing siga correctamente el modo claro u oscuro del portal pÃºblico.
	- VerificaciÃ³n: inspecciÃ³n estÃ¡tica de `web/descripcion_de_los_sistemas.ht` y `web/estilos.css`.

- PostgreSQL runtime: inventario avanzado, tablero financiero y salida PEPS quedan operativos por API en Motel Calipso.
	- Archivos modificados: `backend/db/productos.go`, `backend/db/finanzas.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se reemplaza SQL no portable del tablero de inventario y finanzas por consultas compatibles con PostgreSQL, se completa la serie diaria de tendencia desde Go, y la ruta de salida de inventario deja de fallar al reordenar el consumo PEPS con wrappers SQL compatibles y cerrar el cursor antes de actualizar lotes en la misma transacciÃƒÂ³n.
	- VerificaciÃƒÂ³n: `go test ./db -run "TestRegistrarMovimientoInventario|Test(GetInventarioTendenciaByEmpresaDevuelveSerieDiaria|GetInventarioProyeccionQuiebreByEmpresaPriorizaRiesgo|GetInventarioPlanReposicionByEmpresaConsolidaProveedorYCosto|GetInventarioPlanReposicionResumenByEmpresaAgrupaProveedor|GetEmpresaReportesTableroResumen|GetEmpresaReportesTableroResumenConAsientosCanonicos)$" -count=1`; validaciÃƒÂ³n runtime real sobre `empresa_id=7` con `GET /api/empresa/inventario/tendencia`, `GET /api/empresa/inventario/proyeccion_quiebre`, `GET /api/empresa/inventario/plan_reposicion`, `GET /api/empresa/inventario/plan_reposicion_resumen`, `GET /api/empresa/finanzas/movimientos?action=tablero`, `GET /api/empresa/finanzas/movimientos?action=tablero_export&format=json`, `GET /api/empresa/reportes?action=dataset&dataset=empresarial_tablero`, `POST /api/empresa/compras/plan_reposicion/emitir_orden`, `POST /api/empresa/compras/plan_reposicion/actualizar_estado`, `POST /api/empresa/finanzas/cierres_caja`, `PUT /api/empresa/finanzas/cierres_caja?action=cerrar`, `PUT /api/empresa/finanzas/cierres_caja?action=aprobar` y `POST /api/empresa/inventario/ajustar` -> `200/201/409 esperado`.

- Reportes Globales (Super): Se habilitÃƒÂ³ botÃƒÂ³n de impresiÃƒÂ³n nativa y soporte @media print para reportes globales desde frontend, los cuales ya permitÃƒÂ­an filtrar empresas del admin (individual/mix), consultar por fechas y exportar sus datasets en JSON, CSV, TXT, Excel XLS y PDF generados en backend.

- Documentos Transaccionales y Flujos: se validan facturaciÃƒÂ³n, reportes, eventos, integraciones, y backups empresariales.
	- Archivos involucrados: ackend/handlers/modulos_faltantes.go`n	- DescripciÃƒÂ³n: pruebas exhaustivas automatizadas confirman el cumplimiento legal de DIAN en documentos de notas/facturaciÃƒÂ³n, retenciones de PDF/CSV, resoluciÃƒÂ³n de conflictos de impresiÃƒÂ³n por impresoras registradas, y rutinas de backup correctas.
	- VerificaciÃƒÂ³n: go test ./handlers ./db -run 'TestEmpresaVentasCotizacionesConversionPedidoYDocumentoFinal|TestVentaCarritoFacturaYResolucionImpresora|TestEmpresaFacturacionElectronicaReintentosYReconciliacion' -count=1 -> PASS.

- CrÃƒÂ©ditos y chat/tareas: se cierra la validaciÃƒÂ³n PostgreSQL de abonos y citas en runtime real.
	- Archivos modificados: `backend/db/creditos.go`, `backend/db/chat_tareas.go`, `backend/handlers/chat_tareas.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el abono de crÃƒÂ©ditos deja de fallar por `driver: bad connection` al separar la lectura de cuotas pendientes de las escrituras dentro de la misma transacciÃƒÂ³n. AdemÃƒÂ¡s, `chat_tareas_citas` ahora se autorrepara cuando el esquema PostgreSQL legado llega incompleto, y el listado de citas usa la capa compatible para no romper por SQL o tabla faltante.
	- VerificaciÃƒÂ³n: `go test ./db ./handlers -run '^$' -count=1`; validaciÃƒÂ³n runtime sobre Motel Calipso con `POST /api/empresa/creditos?empresa_id=7 -> 201`, `POST /api/empresa/creditos?action=abono&empresa_id=7 -> 200`, `POST /api/empresa/chat_tareas/citas?empresa_id=7 -> 201` y `GET /api/empresa/chat_tareas/citas?empresa_id=7&q=20260420015859 -> 200`.

- Finanzas y creditos: se corrige compatibilidad PostgreSQL en CRUD generico de cartera y en resumen de cartera de creditos.
	- Archivos modificados: `backend/db/modulos_faltantes.go`, `backend/db/creditos.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `CreateEmpresaGenericRow` deja de usar `LastInsertId` directo y pasa a `insertSQLCompat`, evitando que CxC/CxP persistan pero respondan `400` en PostgreSQL. `GetEmpresaCreditosCarteraResumen` usa `queryRowSQLCompat` y una comparacion de fecha robusta cuando `fecha_vencimiento` viene vacia o legacy, corrigiendo el `500` de `action=resumen_cartera`. Ademas, el flujo de abono de creditos cierra el cursor de cuotas antes del `commit`, evitando `driver: bad connection` en PostgreSQL.
	- VerificaciÃƒÂ³n: compilacion `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1` y revalidacion runtime Motel Calipso en curso tras reinicio del backend.

- Red social comercial: PostgreSQL ya persiste y lista publicaciones empresariales/publicas de Motel Calipso.
	- Archivos modificados: `backend/db/red_social.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `empresa_publicaciones_red_social` ahora crea su tabla con DDL compatible para PostgreSQL y todas las lecturas/escrituras pasan por `execSQLCompat` y `querySQLCompat`. Con esto vuelven a funcionar `POST /api/empresa/publicaciones`, `GET /api/empresa/publicaciones` y `GET /api/public/publicaciones` sobre la base PostgreSQL usada por Motel Calipso.
	- VerificaciÃƒÂ³n: validacion runtime real contra `http://127.0.0.1:8080` con `empresa_id=7`, creando dos publicaciones comerciales y confirmando respuesta `200` tanto en el feed empresarial como en el feed publico.

- PostgreSQL runtime: se corrige el patrÃƒÂ³n `LastInsertId` en mÃƒÂ³dulos crÃƒÂ­ticos y se reencamina `/api/empresa/proveedores` al CRUD coherente con compras e inventario.
	- Archivos modificados: `backend/db/auditoria_empresa.go`, `backend/db/usuarios_empresa.go`, `backend/db/clientes.go`, `backend/db/asistencia_empleados.go`, `backend/db/chat_tareas.go`, `backend/db/creditos.go`, `backend/db/finanzas.go`, `backend/db/productos.go`, `backend/db/venta_publica.go`, `backend/db/red_social.go`, `backend/handlers/productos.go`, `backend/main.go`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: las altas que antes respondÃƒÂ­an con `LastInsertId is not supported by this driver` ahora usan `insertSQLCompat` o `insertTxSQLCompat`, con lo cual PostgreSQL puede devolver `id` vÃƒÂ­a `RETURNING`. TambiÃƒÂ©n se asegura el esquema de venta pÃƒÂºblica antes de guardar configuraciÃƒÂ³n, items u ÃƒÂ³rdenes y se fuerza la ruta `/api/empresa/proveedores` a usar la tabla `proveedores`, que es la misma validada por productos y compras.
	- VerificaciÃƒÂ³n: `go test ./db ./handlers -run '^Test(CreateEmpresaVentaPublicaConfigPersistsTemaVisual|EmpresaVentaPublicaHandlerConfigCatalogoYToggle|EmpresaPublicacionesRedSocialHandler|EmpresaChatTareasCitasSharedByEmpresa|EmpresaCategoriasProductosHandlerCRUD|EmpresaClientesHandler|EmpresaCreditos|EmpresaAsistencia)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

## 2026-04-19
- Apariencia global: se reparan los 6 temas, el menÃƒÂº flotante y el guardado automÃƒÂ¡tico al iniciar sesiÃƒÂ³n.
	- Archivos modificados: `web/menu.js`, `web/estilos.css`, `web/js/login.js`, `web/js/login_usuario.js`, `web/login_usuario.html`, `web/configuracion_de_la_cuenta.html`, `web/red_social_comercial.html`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/usuarios_empresa.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivos creados: `web/Juegos/menu_juegos.html`, `web/Juegos/n64/index.html`.
	- DescripciÃƒÂ³n: el `menu.js` compartido ahora aplica el tema desde el arranque, sincroniza iframes mismo origen, guarda selecciÃƒÂ³n en `localStorage`/cookie y refresca la preferencia desde backend cuando existe sesiÃƒÂ³n. Los logins administrativo y de usuario empresa devuelven `apariencia` para fijarla antes de redirigir. AdemÃƒÂ¡s, vuelve la entrada `Juegos` y se publican rutas funcionales mÃƒÂ­nimas para no dejar el enlace roto.
	- VerificaciÃƒÂ³n: `get_errors` sobre los archivos tocados y tarea `validar-permisos-selector-empresas-5`.

- Modelo de base de datos ERP (MÃ¯Â¿Â½dulo interdependiente de Compras / Proveedores).
        - Archivos modificados: backend/db/compras_y_proveedores.go, backend/db/compras_y_proveedores_test.go, backend/main.go, documentos/descripcion_de_archivos, documentos/historial_de_cambios, documentos/estructura_bd.md, CHANGELOG.md.
        - DescripciÃƒÂ³n: Se crearon las tablas operativas empresa_proveedores, empresa_ordenes_compra, empresa_ordenes_compra_items, y empresa_compras_recepciones para soportar el ciclo de abastecimiento. Se conectÃ¯Â¿Â½ EnsureEmpresasComprasSchema al bootstrap del servidor.
        - VerificaciÃƒÂ³n: EjecuciÃ¯Â¿Â½n local exitosa de tests para esquema relacional.

- EliminaciÃƒÂ³n de pÃƒÂ¡ginas huÃƒÂ©rfanas frontend al carecer de uso actual.
	- Archivos modificados: \web/administrar_empresa/bodegas.html\, \web/administrar_empresa/productos/bodegas.html\, \web/administrar_empresa/sensor_puertas_mensajes.html\, \web/administrar_empresa/ventas_simple.html\, \web/super/activar_asesor.html\, \web/super/asesor_comercial.html\, \web/super/vendedor_config_avanzado.html\, \web/ultimas_mejoras.html\ (eliminados), \documentos/descripcion_de_archivos\, \documentos/historial_de_cambios\, \CHANGELOG.md\.
	- DescripciÃƒÂ³n: Se limpiaron del proyecto 8 archivos \.html\ obsoletos o huÃƒÂ©rfanos para reducir el tamaÃƒÂ±o del repositorio y evitar confusiones en los directorios de operaciÃƒÂ³n y super admin.
	- VerificaciÃƒÂ³n: No existen referencias de enrutamiento ni clics a estas pÃƒÂ¡ginas en los menÃƒÂºs o vistas dinÃƒÂ¡micas.

- Juegos globales: ranking pÃƒÂºblico para 3 juegos integrados (Buscaminas, Solitario, Pacman).
	- Archivos modificados: `backend/db/super_juegos.go`, `backend/handlers/super_juegos.go`, `backend/main.go`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `documentos/estructura_bd.md`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: Se agrega un nuevo modulo para almacenar y consultar records globales de juegos en la base `superadministrador`. La tabla `super_juegos_records` registra `juego`, `nombre_jugador`, `empresa_id` (o 'Publico'), puntaje y nivel. Se exponen los endpoints `GET` y `POST` en `/api/public/juegos/records` para permitir que cualquier jugador (empresa o pÃƒÂºblico) consulte el top 10 o envÃƒÂ­e su puntaje desde el frontend.
	- VerificaciÃƒÂ³n: CompilaciÃƒÂ³n exitosa en `agente_backend_db`. Rutas registradas correctamente en `main.go`.

- Portal pÃƒÂºblico: la landing `Explorar oferta` adopta el estilo de tarjetas del index y una estÃƒÂ©tica propia mÃƒÂ¡s comercial.
	- Archivos modificados: `web/descripcion_de_los_sistemas.ht`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la pÃƒÂ¡gina descriptiva de ofertas deja el bloque visual oscuro anterior y pasa a reutilizar el estilo `home-offer-card` del index en sus tarjetas ampliadas. El hero y la navegaciÃƒÂ³n rÃƒÂ¡pida tambiÃƒÂ©n se refinan para que la landing se perciba como extensiÃƒÂ³n natural del portal principal, manteniendo los mismos enlaces seguros de `Probar Gratis`.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/descripcion_de_los_sistemas.ht`, `web/estilos.css`, `documentos/descripcion_de_modulos` y `documentos/matriz_roles_permisos_pos_multiempresa.md`.

- Chat y tareas: selecciÃƒÂ³n mÃƒÂºltiple de usuarios, fotos y validaciÃƒÂ³n estricta de empresa.
	- Archivos modificados: `backend/handlers/chat_tareas.go`, `backend/handlers/chat_tareas_test.go`, `web/administrar_empresa/chat_y_tareas.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `chat_y_tareas.html` ahora permite buscar y marcar uno o varios usuarios activos de la misma empresa para crear chats directos o grupales y agregar varios participantes a una conversaciÃƒÂ³n existente. El formulario de mensajes deja explÃƒÂ­cito el envÃƒÂ­o de fotos ademÃƒÂ¡s de audio y documentos. En paralelo, `chat_tareas.go` valida que cada participante tipo `usuario` pertenezca realmente a la empresa antes de persistirlo, bloqueando cruces entre empresas en creaciÃƒÂ³n grupal o agregado posterior.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `backend/handlers/chat_tareas.go`, `backend/handlers/chat_tareas_test.go`, `web/administrar_empresa/chat_y_tareas.html` y `web/estilos.css`; `go test ./handlers -run '^TestEmpresaChatTareas(AdjuntoUploadAllowsImage|AdjuntoUploadAllowsDocx|ConversacionesAddsOwnerAdminParticipant|ConversacionesCreatesGrupoConUsuariosSeleccionados|ConversacionesRejectsUsuarioDeOtraEmpresa|ParticipantesRejectsUsuarioDeOtraEmpresa|CitasSharedByEmpresa|MensajesHandlerDerivesUsuarioActor|MensajesHandlerRejectsInvalidConversacion|TareasHandlerRejectsInvalidConversacion|CitasHandlerRejectsInvalidConversacion)$' -count=1`.

- Panel empresa y chat/tareas: el mÃƒÂ³dulo colaborativo se refuerza como dashboard principal y valida referencias empresariales antes de persistir.
	- Archivos modificados: `backend/handlers/chat_tareas.go`, `backend/handlers/chat_tareas_test.go`, `web/administrar_empresa/chat_y_tareas.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `chat_y_tareas.html` ahora se comporta como home operativo del panel empresa con tarjetas resumen, acciones rÃƒÂ¡pidas y estados vacÃƒÂ­os guiados. En paralelo, `chat_tareas.go` valida que conversaciones y tareas referenciadas existan dentro de la empresa antes de crear participantes, mensajes, tareas, citas o notas de voz, y limpia archivos de adjuntos si falla la persistencia posterior para no dejar huÃƒÂ©rfanos.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `backend/handlers/chat_tareas.go`, `backend/handlers/chat_tareas_test.go`, `web/administrar_empresa/chat_y_tareas.html` y `web/estilos.css`; `go test ./handlers -run '^TestEmpresaChatTareas(MensajesHandlerDerivesUsuarioActor|AdjuntoUploadAllowsDocx|ConversacionesAddsOwnerAdminParticipant|CitasSharedByEmpresa|MensajesHandlerRejectsInvalidConversacion|TareasHandlerRejectsInvalidConversacion|CitasHandlerRejectsInvalidConversacion)$' -count=1`.

- AutenticaciÃƒÂ³n administrativa: `login.html` vuelve a ofrecer `Recordar usuario` para el acceso por correo.
	- Archivos modificados: `web/login.html`, `web/js/login.js`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el login administrativo por correo vuelve a incluir una casilla visible `Recordar usuario`. Cuando se marca, el frontend conserva solo el correo del administrador en `localStorage`; si no se marca, la identidad recordada se limpia. El cambio no altera la sesion real, el flujo de Google, permisos ni wrappers de autenticacion.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/login.html`, `web/js/login.js`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos` y `documentos/matriz_roles_permisos_pos_multiempresa.md`.

- Portal pÃƒÂºblico: el home resalta la marca principal y unifica el pie de pÃƒÂ¡gina con la nueva leyenda corporativa.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `web/login.html`, `web/registrar_nuevo_usuario_administrador.html`, `web/ayuda/ayuda.html`, `web/descripcion_de_los_sistemas.ht`, `web/Informacion_de_contacto.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el botÃƒÂ³n `Registrarse o iniciar sesiÃƒÂ³n` del portal principal adopta una apariencia tipo tarjeta con texto reforzado, y el tÃƒÂ­tulo `Sistema de FacturaciÃƒÂ³n ElectrÃƒÂ³nica` queda destacado sobre una barra semitransparente para darle mÃƒÂ¡s presencia visual. AdemÃƒÂ¡s, se reemplaza la leyenda visible anterior por `@ 2026 - Powerful Control System - Sistema POS Multiempresa` en los pies pÃƒÂºblicos afectados y se actualizan los tÃƒÂ­tulos de pÃƒÂ¡ginas que todavÃƒÂ­a cargaban la marca `POS Multiempresa` como nombre principal.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/index.html`, `web/estilos.css`, `web/login.html`, `web/registrar_nuevo_usuario_administrador.html`, `web/ayuda/ayuda.html`, `web/descripcion_de_los_sistemas.ht` y `web/Informacion_de_contacto.html`.

- Panel empresa: Chat y tareas pasa a ser la vista inicial y su calendario compartido queda en primer plano.
	- Archivos modificados: `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/administrar_empresa/chat_y_tareas.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `administrar_empresa.html` ahora abre preferentemente `chat_y_tareas.html` al entrar y mueve ese acceso al primer lugar del menÃƒÂº. La pÃƒÂ¡gina colaborativa sube el calendario mensual al inicio, refuerza la explicaciÃƒÂ³n de agenda compartida por empresa y amplÃƒÂ­a el peso visual del tablero para que la administradora registre reuniones y los demÃƒÂ¡s usuarios autorizados las consulten desde su cuenta.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/administrar_empresa/chat_y_tareas.html` y `web/estilos.css`; `go test ./handlers -run '^TestEmpresaChatTareasCitasSharedByEmpresa$' -count=1`.

- Venta pÃƒÂºblica y pagos: el catÃƒÂ¡logo pÃƒÂºblico deja de fallar con error interno sobre esquemas legacy.
	- Archivos modificados: `backend/db/venta_publica.go`, `backend/handlers/venta_publica_test.go`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el backend ahora autorrepara columnas faltantes de configuracion, items y ordenes en `empresa_venta_publica_*` antes de consultar la tienda pÃƒÂºblica. Con esto, `GET /api/public/venta_publica?action=catalogo` deja de caer en instalaciones legado que no tenÃƒÂ­an campos evolutivos como `tema_visual`, `estado`, `destacado`, `stock_publicado` o payloads de orden. Se agrega ademÃƒÂ¡s una regresiÃƒÂ³n especÃƒÂ­fica para catÃƒÂ¡logo pÃƒÂºblico sobre esquema legacy.
	- VerificaciÃƒÂ³n: `go test ./db -run '^TestEmpresaVentaPublicaConfigPersistsTemaVisual$' -count=1`; `go test ./handlers -run '^Test(PublicVentaPublicaHandlerCatalogoWithLegacySchemaMissingColumns|PublicVentaPublicaHandlerCatalogoYPagoConWompiInactivo|EmpresaVentaPublicaHandlerConfigCatalogoYToggle)$' -count=1`; `get_errors` sin errores en `backend/db/venta_publica.go` y `backend/handlers/venta_publica_test.go`.

- Panel empresa: reservas endurece sus queries y evita URLs infladas que podÃƒÂ­an terminar en error 414.
	- Archivos modificados: `web/administrar_empresa/reservas_hotel.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la vista de reservas ahora resuelve `empresa_id` desde el contexto activo del panel, limita el texto de bÃƒÂºsqueda antes de enviarlo por query string y centraliza la construcciÃƒÂ³n de URLs del mÃƒÂ³dulo para no propagar parÃƒÂ¡metros largos o innecesarios. AdemÃƒÂ¡s, la tabla deja de incrustar el JSON completo de cada reserva en atributos HTML y usa el estado local para editar, reduciendo el tamaÃƒÂ±o del DOM y evitando crecimiento accidental de la navegaciÃƒÂ³n.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/administrar_empresa/reservas_hotel.html`.

- Panel empresa: administrar productos deja de caer por la carga obligatoria de proveedores en la vista principal.
	- Archivos modificados: `web/administrar_empresa/administrar_productos.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la vista `Administrar Productos` ya no arranca consultando proveedores de forma obligatoria cuando el modo visible es `productos`. El selector de proveedor principal pasa a cargarse de forma perezosa al abrir el formulario del producto, de modo que una falla o restricciÃƒÂ³n del submÃƒÂ³dulo de compras/proveedores no tumba toda la pantalla principal de inventario.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/administrar_empresa/administrar_productos.html`.

- Panel empresa: se corrige la duplicaciÃƒÂ³n recursiva del submenÃƒÂº de productos y demÃƒÂ¡s shells anidados.
	- Archivos modificados: `web/js/administrar_empresa.js`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el script compartido del panel empresa ahora detecta los enlaces y el iframe del shell actual antes de resolver la pÃƒÂ¡gina inicial. Con esto, los submenÃƒÂºs internos como `Productos` dejan de cargarse a sÃƒÂ­ mismos dentro de su propio iframe y pasan a abrir su contenido real, evitando el efecto de menÃƒÂº duplicado.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/js/administrar_empresa.js`; validaciÃƒÂ³n estÃƒÂ¡tica en `http://127.0.0.1:8091/administrar_empresa/administrar_productos_menu.html?empresa_id=1` con snapshot mostrando el iframe apuntando a `administrar_productos.html?view=productos&empresa_id=1` en lugar de recargar `administrar_productos_menu.html`.

- Frontend web: se endurece la adaptaciÃƒÂ³n automÃƒÂ¡tica a celular sin alterar la estructura de menÃƒÂºs.
	- Archivos modificados: `web/estilos.css`, `web/super/administrar_base_de_datos.html`, `web/super/errores.html`, `web/super/seguridad.html`, `web/super/soporte_remoto.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la hoja compartida ahora fuerza en mÃƒÂ³vil el apilado real de filas de formulario, toolbars y filtros, neutraliza anchos mÃƒÂ­nimos e inline widths problemÃƒÂ¡ticos en controles, y mejora el comportamiento de tablas y bloques con scroll horizontal cuando hace falta. AdemÃƒÂ¡s, las vistas super de base de datos, errores, seguridad VPS y soporte remoto reciben media queries puntuales para evitar desbordes y columnas rÃƒÂ­gidas en celular sin tocar los menÃƒÂºs existentes.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/estilos.css`, `web/super/administrar_base_de_datos.html`, `web/super/errores.html`, `web/super/seguridad.html` y `web/super/soporte_remoto.html`.

- Estaciones y carritos: se reutilizan carritos legado al abrir una estaciÃƒÂ³n y se evita el error de carga por duplicado.
	- Archivos modificados: `web/administrar_empresa/carrito_de_compras.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: al entrar desde `estaciones.html`, el carrito unificado ya no asume que siempre existe un carrito con el cÃƒÂ³digo canÃƒÂ³nico `EST-empresa-estacion`. Ahora intenta reutilizar primero un carrito ya existente por cÃƒÂ³digo, `referencia_externa=ESTACION_<id>` o nombre de estaciÃƒÂ³n antes de crear uno nuevo, evitando conflictos con datos legado que antes terminaban en `Error cargando carritos`.
	- VerificaciÃƒÂ³n: `get_errors` sobre `web/administrar_empresa/carrito_de_compras.html` y validaciÃƒÂ³n dirigida del flujo de estaciÃƒÂ³n.

- Estaciones: la estaciÃƒÂ³n especial de YouTube cambia a un visor embebido estable con bÃƒÂºsqueda funcional.
	- Archivos modificados: `web/administrar_empresa/youtube_station_browser.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se elimina la dependencia de la API del reproductor que no estaba resolviendo bien la carga en la tarjeta y se reemplaza por un visor embebido basado en `youtube-nocookie` que sÃƒÂ­ permite mostrar resultados y lanzar bÃƒÂºsquedas desde la propia estaciÃƒÂ³n. Se agrega un botÃƒÂ³n `Inicio` para volver a una portada embebida ÃƒÂºtil y se conserva `Abrir YouTube` para abrir la pÃƒÂ¡gina real completa en otra pestaÃƒÂ±a, porque el home oficial de YouTube no se puede incrustar de forma fiable dentro de un iframe normal.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/administrar_empresa/youtube_station_browser.html`.

- Super administrador: se agrega un panel real para formatos de email y se unifica el guardado de configuraciÃƒÂ³n avanzada.
	- Archivos creados: `backend/handlers/super_email_templates.go`, `web/super/formato_para_emviar_email.html`.
	- Archivos modificados: `backend/main.go`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/payments_handlers.go`, `backend/handlers/server_runtime_notifications.go`, `backend/handlers/system_empresas_handlers_test.go`, `web/super/configuracion_avanzada.html`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el panel super ahora expone `/super/formato_para_emviar_email.html` para administrar plantillas reales de confirmaciÃƒÂ³n de correo, activaciÃƒÂ³n por pago de licencia, recuperaciÃƒÂ³n de contraseÃƒÂ±a y alertas de reinicio. El backend centraliza esas plantillas en `/super/api/config/email_templates` y reemplaza textos hardcodeados en correos administrativos, usuarios de empresa, licencias y monitoreo del servidor. AdemÃƒÂ¡s, `configuracion_avanzada.html` deja botones sueltos por bloque y pasa a guardar Wompi, Epayco, Gmail e IA con un solo botÃƒÂ³n arriba y otro abajo de la pÃƒÂ¡gina.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'Test(SuperEmailTemplatesHandlerSaveAndGet|ApplySuperEmailTemplateUsesConfiguredValues|GmailConfigHandlerTestActionCapturesNotification)$' -count=1`; tarea `validar-permisos-selector-empresas-5`; `get_errors` sin errores en `backend/handlers/super_email_templates.go`, `web/super/formato_para_emviar_email.html` y `web/super/configuracion_avanzada.html`.

- Portal pÃƒÂºblico y selector de empresas: se corrigen CTA pÃƒÂºblicos y se oculta navegaciÃƒÂ³n global fuera del alcance super principal.
	- Archivos modificados: `web/descripcion_de_los_sistemas.ht`, `web/js/seleccionar_empresa.js`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la landing descriptiva ya no reutiliza destinos privados como `/administrar_empresa.html` en el CTA `Probar Gratis`; cuando una tarjeta apunta a una ruta protegida, el flujo pÃƒÂºblico redirige al registro de administrador. En el selector de empresas, el menÃƒÂº lateral ahora toma el perfil real desde `/me` y solo mantiene visibles `Administradores` y `Reportes globales` para cuentas super principales; la navegaciÃƒÂ³n sensible queda oculta por defecto hasta resolver la sesiÃƒÂ³n.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/descripcion_de_los_sistemas.ht`, `web/js/seleccionar_empresa.js` y `web/seleccionar_empresa.html`; sondeo runtime local con `200` en `/descripcion_de_los_sistemas.ht` y `/registrar_nuevo_usuario_administrador.html`, y `401` en `/seleccionar_empresa.html` sin sesiÃƒÂ³n; `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1`.

- Integridad super/licencias: se endurece el alcance delegado, se recupera compatibilidad de backup legacy y se corrige la validaciÃƒÂ³n pÃƒÂºblica de mÃƒÂ©todos de pago.
	- Archivos modificados: `backend/db/chat_inteligencia_artificial.go`, `backend/handlers/system_empresas_handlers.go`, `backend/handlers/postgres_performance.go`, `backend/handlers/super_config_backup_handlers.go`, `backend/handlers/payments_handlers.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: los administradores delegados ya no heredan acceso global por tener rol `super_administrador`, sino que quedan restringidos al portafolio del administrador principal. El backup/restore super vuelve a aceptar claves sensibles legacy de IA para restauraciones compatibes, `GET /api/public/licencias/payment_methods` puede anunciar Epayco cuando existe `public_key`, y el panel de rendimiento PostgreSQL valida primero la acciÃƒÂ³n solicitada para devolver `400` en acciones no soportadas aunque el runtime no estÃƒÂ© sobre PostgreSQL.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'Test(PostgresPerformanceHandlerUnknownAction|EmpresasHandlerFiltraEmpresasPorAdministradorPrincipal|SuperConfigBackupHandlerRestoreEncryptsSensitivePlaintext|PublicLicenciasPaymentMethodsHandlerAllowsEpaycoWithPublicKeyOnly|SuperConfigBackupHandlerExportYRestore|PublicLicenciasPaymentMethodsHandlerOrdersAndAvailability)$' -count=1`; tarea `validar-permisos-selector-empresas-5`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1`; sondeo runtime local de `/`, `/index.html`, `/login.html`, `/api/public/pagina_principal`, `/api/public/licencias/payment_methods` y `/seleccionar_empresa.html`.

## 2026-04-18

- Estaciones: la tarjeta de YouTube ahora permite guardar la fuente desde el mismo bloque y acepta Shorts.
	- Archivos modificados: `web/administrar_empresa/estaciones.html`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la propia tarjeta de YouTube en `estaciones.html` ahora incluye un campo para pegar la URL o el ID del contenido, un botÃƒÂ³n `Guardar y cargar` y un enlace externo alineado con el valor actual. El navegador interno ya interpreta tambiÃƒÂ©n URLs de `Shorts` como video vÃƒÂ¡lido y recarga la vista sin obligar a entrar a la configuraciÃƒÂ³n general de estaciones.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/administrar_empresa/estaciones.html` y `web/estilos.css`.

- Estaciones: la tarjeta de YouTube deja de depender de bÃƒÂºsquedas embebidas rotas y pasa a reproducir referencias vÃƒÂ¡lidas.
	- Archivos modificados: `web/administrar_empresa/youtube_station_browser.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/configuracion_de_estaciones.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la estaciÃƒÂ³n especial de YouTube ya no intenta incrustar resultados de bÃƒÂºsqueda que el proveedor externo no soporta. Ahora reproduce solo URLs o IDs vÃƒÂ¡lidos de video/playlist mediante `youtube-nocookie`, muestra la referencia configurada dentro de la tarjeta y, cuando el valor guardado es texto libre, deja un estado visible con fallback a `Abrir YouTube` fuera del sistema.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/administrar_empresa/youtube_station_browser.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/configuracion_de_estaciones.html` y `web/estilos.css`.

- Estaciones y carritos: se mejora el mensaje visible cuando falla la apertura del carrito por estaciÃƒÂ³n.
	- Archivos modificados: `web/administrar_empresa/carrito_de_compras.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el carrito unificado ya no deja al usuario solo con `Error cargando carritos` al fallar la carga inicial. Ahora muestra un estado contextual con traducciÃƒÂ³n segura del error, botÃƒÂ³n `Reintentar carga` y retorno explÃƒÂ­cito a `estaciones.html` cuando el flujo viene desde una estaciÃƒÂ³n, evitando exponer literales tÃƒÂ©cnicos del backend como `unauthenticated` o `forbidden`.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/administrar_empresa/carrito_de_compras.html`; validaciÃƒÂ³n estÃƒÂ¡tica de apertura en `http://127.0.0.1:8080/administrar_empresa/carrito_de_compras.html?empresa_id=6&estacion_id=1&estacion_nombre=Estacion%201&carrito_codigo=EST-6-1`.

- Estaciones y carritos: se corrige la carga del carrito por estaciÃƒÂ³n sobre PostgreSQL real.
	- Archivos modificados: `backend/db/carritos_compras.go`, `backend/handlers/carritos_compras.go`, `backend/db/carritos_inventario_test.go`, `backend/handlers/auth_users_carritos_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el listado de `/api/empresa/carritos_compra` deja de depender de un `GROUP BY` frÃƒÂ¡gil con cliente e items y ahora cuenta items desde un agregado previo por carrito. AdemÃƒÂ¡s, `totales_pago` y `metricas_estacion` dejan de usar `ROUND(..., 2)` en SQL y redondean en Go, evitando fallos de compatibilidad entre motor legado retirado y PostgreSQL. Esto elimina el error visible `Error cargando carritos` al abrir una estaciÃƒÂ³n y estabiliza los totales del panel.
	- VerificaciÃƒÂ³n: `go test ./db -run 'Test(GetCarritosCompraByEmpresaFallbackWithoutClientesSchema|GetCarritosCompraByEmpresaCountsItemsAndClientName|SyncEmpresaEstacionCarritosCreatesAndUpdatesLinkedDefaults)$' -count=1`; `go test ./handlers -run 'Test(EmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones|EmpresaCarritosCompraTotalesPagoAgrupaYRedondea|EmpresaCarritosCompraRejectsActivarEstacionPagadaSinReset|EmpresaCarritosCompraRecuperarInterrumpidoConAuditoria|EmpresaEstacionPrefsHandler_UpsertAndIsolationByEmpresa|WithEmpresaVentasPermissionsDeniesOutOfScopeEmpresa|WithEmpresaVentasPermissionsBloqueaModuloNoHabilitadoPorLicencia)$' -count=1`.

- Carritos y estaciones: se limpia el legado de `ventas_simple` y se amplÃƒÂ­an los checks del carrito unificado.
	- Archivos modificados: `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`, `backend/handlers/empresa_configuracion_general.go`, `backend/handlers/empresa_configuracion_general_test.go`, `backend/handlers/empresa_estacion_prefs_test.go`, `backend/db/empresa_configuracion_general.go`, `backend/db/empresa_estacion_prefs_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivos eliminados: `web/js/ventas_simple.js`.
	- DescripciÃƒÂ³n: el carrito unificado ahora permite configurar tambiÃƒÂ©n la visibilidad del cliente, el bloque de descuento/impuestos por item, el resumen total del carrito y el desglose del cobro. AdemÃƒÂ¡s, se retira del runtime el script antiguo de `ventas_simple` y se elimina el bloque muerto de carrito compacto en `configuracion_general`.
	- VerificaciÃƒÂ³n: pruebas dirigidas de handlers/db del modulo y validacion funcional del redirect de compatibilidad.

- IA: el proyecto migra a Google Gemini como ÃƒÂºnico proveedor y retira Ollama/DeepSeek del flujo operativo.
	- Archivos modificados: `backend/handlers/ai_credentials_catalog.go`, `backend/handlers/ai_config_handlers.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super_test.go`, `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/db/chat_inteligencia_artificial_test.go`, `backend/tools/set_ai_provider_enabled.go`, `backend/.env.example`, `backend/.env.local`, `scripts/iniciar_servidor.ps1`, `web/super/configuracion_avanzada.html`, `web/super/chat_con_ia_global.html`, `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivos eliminados: `backend/tools/check_deepseek_key/main.go`.
	- DescripciÃƒÂ³n: se reemplaza la IA en VPS basada en Ollama por Google Gemini en el chat global super y el chat con IA por empresa. ConfiguraciÃƒÂ³n avanzada queda reducida a una sola API key cifrada, un interruptor global y un interruptor del proveedor `google`. El script de arranque ya no abre tÃƒÂºnel a Ollama y el VPS queda sin `ollama.service` ni binario asociado.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'Test(SuperAIModelosHandlerReturnsCatalog|SuperAIModelosHandlerFiltersDisabledProvider|ModelosHandlerReturnsPreferredModelForGoogleAccount|ModeloPreferidoHandlerAcceptsGemini|ModelosHandlerFiltersDisabledProvider|AIModelsConfigHandlerSaveGeminiEncrypted|AIModelsConfigHandlerSavesProviderEnabledState|AIModelsConfigHandlerTogglesGlobalServiceState)$' -count=1`; `go test ./db -run 'Test(EmpresaAIModeloPreferidoUpsertAndGet|RegisterEmpresaAIConsultaAcumulaUsoDiario|SuperAIModeloPreferidoUpsertAndGet|RegisterSuperAIConsultaAcumulaUsoDiario|GetSuperAIModeloPreferidoRepairsMissingSchema|GetSuperAIUsoDiarioRepairsMissingSchema|RegisterSuperAIConsultaRepairsMissingSchema)$' -count=1`.

- Carritos y estaciones: el sistema adopta un carrito unificado configurable por empresa y por estaciÃƒÂ³n.
	- Archivos modificados: `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/ventas_simple.html`, `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`, `web/administrar_empresa/configuracion_menu.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: desaparece la bifurcaciÃƒÂ³n de UI entre carrito de compras, venta simple y carrito compacto. Las estaciones abren siempre `carrito_de_compras.html` y la pantalla muestra u oculta bloques segÃƒÂºn configuraciÃƒÂ³n global por empresa y configuraciÃƒÂ³n individual por estaciÃƒÂ³n almacenadas en `estaciones_config`.
	- VerificaciÃƒÂ³n: tarea `validar-permisos-selector-empresas-5`; `get_errors` sin errores en los archivos frontend modificados.


- Chat con IA: interfaz simplificada en empresa y panel super.
	- Archivos modificados: `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se retiran botones superiores, cuadros visibles de modelo/uso diario, selector visible de modelo, CTA de upgrade y panel de consultas recientes; las sugerencias rapidas pasan debajo del chat y `Limpiar chat` queda junto a `Preguntar a la IA`.
	- VerificaciÃƒÂ³n: `get_errors` sobre las vistas y documentos modificados.

- Panel empresa: se retira la calculadora del menu lateral y del menu flotante.
	- Archivos modificados: `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/menu.js`, `web/ayuda/ayuda.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivos eliminados: `web/administrar_empresa/calculadora.html`.
	- DescripciÃƒÂ³n: se elimina el boton `Calculadora`, se retira el acceso rapido del menu flotante y se borra la pagina frontend asociada para que deje de existir como frente visible del panel empresa.
	- VerificaciÃƒÂ³n: `get_errors` y busqueda de referencias frontend activas a `calculadora.html`.

- Inventario y productos: las compras preventivas y por proveedor pasan a la secciÃƒÂ³n `Compras` del submÃƒÂ³dulo.
	- Archivos modificados: `web/administrar_empresa/administrar_productos.html`, `web/administrar_empresa/productos/compras.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la vista central del mÃƒÂ³dulo ahora expone `view=compras` y concentra allÃƒÂ­ el plan de reposiciÃƒÂ³n por proveedor, el consolidado de compra y el borrador/ciclo de orden; la ruta `productos/compras.html` deja el placeholder y redirige a esa vista real.
	- VerificaciÃƒÂ³n: `get_errors` sobre los HTML y documentos modificados.

- Panel empresa: se elimina la pÃƒÂ¡gina `Inicio` y el shell arranca directo en Productos.
	- Archivos eliminados: `web/administrar_empresa/inicio.html`.
	- Archivos modificados: `web/administrar_empresa.html`, `web/administrar_empresa/administrar_productos_menu.html`, `web/js/administrar_empresa.js`, `web/ayuda/ayuda.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se retira la portada intermedia del panel empresa, desaparece el botÃƒÂ³n `Inicio` del menÃƒÂº principal y del submÃƒÂ³dulo de productos, y la carga inicial del iframe pasa a `administrar_productos_menu.html`.
	- VerificaciÃƒÂ³n: `get_errors` sobre los archivos HTML, JS y documentos modificados.

- Inventario y productos: `Proveedores` pasa a una subpÃƒÂ¡gina dedicada y `Precios` muestra el historial real de cambios de precio.
	- Archivos creados: `web/administrar_empresa/productos/administrar_proveedores.html`.
	- Archivos modificados: `web/administrar_empresa/administrar_productos.html`, `web/administrar_empresa/administrar_productos_menu.html`, `web/administrar_empresa/productos/administrar_productos_menu.html`, `web/administrar_empresa/productos/precios.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la vista principal del mÃƒÂ³dulo ya no mezcla proveedores ni el historial de cambios de precio con el CRUD de productos; ambos salen a subvistas dedicadas reutilizando la misma pÃƒÂ¡gina central y preservando `empresa_id` dentro del shell administrativo.
	- VerificaciÃƒÂ³n: `get_errors` sobre los HTML y documentos modificados.

- Chat IA super y empresarial: se autorrepara el esquema legacy `super_ai_*` y `empresa_ai_*` y se amplÃƒÂ­a el timeout de Ambis Local sobre el tÃƒÂºnel VPS.
	- Archivos modificados: `backend/db/chat_inteligencia_artificial.go`, `backend/db/chat_inteligencia_artificial_test.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `documentos/descripcion_de_modulos`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el chat super ya no falla al consultar modelo preferido o uso diario sobre instalaciones PostgreSQL heredadas; la capa DB repara tablas/columnas faltantes al vuelo y el cliente de Ollama ahora soporta tiempos de respuesta mÃƒÂ¡s largos de `codellama:7b` cuando la consulta viaja por el tÃƒÂºnel local al VPS.
	- VerificaciÃƒÂ³n: `go test ./db -run 'Test(SuperAIModeloPreferidoUpsertAndGet|RegisterSuperAIConsultaAcumulaUsoDiario|GetSuperAIModeloPreferidoRepairsMissingSchema|GetSuperAIUsoDiarioRepairsMissingSchema|RegisterSuperAIConsultaRepairsMissingSchema)$' -count=1`; `go test ./handlers -run '^$' -count=1`; `curl http://localhost:8080/super/api/chat_con_ia_global/modelos` -> `200 OK`; `curl http://localhost:8080/super/api/chat_con_ia_global/consultar` -> `200 OK`.

- IA super y empresarial: se agregan switches por proveedor para desactivar DeepSeek sin afectar Ambis Local.
	- Archivos modificados: `backend/handlers/ai_credentials_catalog.go`, `backend/handlers/ai_config_handlers.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`, `backend/handlers/chat_con_ia_global_super_test.go`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_de_modulos`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: configuraciÃƒÂ³n avanzada ahora puede habilitar o bloquear por separado `DeepSeek Chat` y `Ambis Local`; los chats empresarial y global super solo muestran proveedores activos y hacen fallback automÃƒÂ¡tico cuando el modelo preferido quedÃƒÂ³ apagado.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'Test(AIModelsConfigHandlerSaveDeepSeekEncrypted|AIModelsConfigHandlerSavesProviderEnabledState|ModelosHandlerFiltersDisabledProvider|ModelosHandlerRejectsWhenAIDisabled|SuperAIModelosHandlerFiltersDisabledProvider|SuperAIModelosHandlerRejectsWhenAIDisabled)$' -count=1`.

- Portal pÃƒÂºblico: se elimina el arcade y solo queda el emulador N64 adaptado para mÃƒÂ³vil.
	- Archivos eliminados: `web/Juegos/ajedrez_3d_plus.html`, `web/Juegos/ajedrez_vs_ia_plus.html`, `web/Juegos/arcade_shared.js`, `web/Juegos/arcade_window.css`, `web/Juegos/brigada_burbujas_3d_plus.html`, `web/Juegos/carton_fire_plus.html`, `web/Juegos/memoria_estelar.html`, `web/Juegos/memoria_estelar_plus.html`, `web/Juegos/pacman_plus.html`, `web/Juegos/patito_volando.html`, `web/Juegos/patito_volando_plus.html`, `web/Juegos/rebote_bloques.html`, `web/Juegos/rebote_bloques_plus.html`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/serpiente_pixel_plus.html`, `web/Juegos/tetris_plus.html`, `web/Juegos/nes/README.md`, `web/Juegos/nes/index.html`, `web/Juegos/nes/nes-wrapper.js`, `web/Juegos/nes/styles.css`, `web/img/juegos/ajedrez_3d.svg`, `web/img/juegos/ajedrez_vs_ia.svg`, `web/img/juegos/brigada_burbujas_3d.svg`, `web/img/juegos/carton_fire.svg`, `web/img/juegos/memoria_estelar.svg`, `web/img/juegos/pacman.svg`, `web/img/juegos/patito_volando.svg`, `web/img/juegos/rebote_bloques.svg`, `web/img/juegos/serpiente_pixel.svg`, `web/img/juegos/tetris.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `web/Juegos/n64/index.html`, `web/Juegos/n64/n64-wrapper.js`, `web/menu.js`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el proyecto retira todas las experiencias de juego y el emulador NES. El portal conserva solo el emulador N64 con acceso pÃƒÂºblico y diseÃƒÂ±o mÃƒÂ³vil, y el menÃƒÂº global enlaza de forma directa a esa ÃƒÂºnica experiencia.
	- VerificaciÃƒÂ³n: bÃƒÂºsqueda global en `web/**` y `backend/**` sin referencias funcionales a los juegos eliminados.

- Selector de empresas: las tarjetas quedan alineadas a la izquierda y conservan un tamaÃƒÂ±o visual uniforme.
	- Archivos modificados: `web/estilos.css`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se elimina el centrado del grid en `seleccionar_empresa.html` y se fija una relaciÃƒÂ³n de aspecto comÃƒÂºn para que todas las tarjetas se vean del mismo tamaÃƒÂ±o dentro del selector.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/estilos.css`.

- Portal pÃƒÂºblico: el menÃƒÂº flotante deja de mostrar accesos rÃƒÂ¡pidos de juegos y centraliza la navegaciÃƒÂ³n en la pÃƒÂ¡gina del arcade.
	- Archivos eliminados: `web/Juegos/games.json`.
	- Archivos modificados: `web/menu.js`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se retira el manifiesto JSON que alimentaba el menÃƒÂº flotante y se deja un ÃƒÂºnico enlace a `/Juegos/menu_juegos.html`, evitando duplicar la lista de juegos fuera del lobby principal.
	- VerificaciÃƒÂ³n: bÃƒÂºsqueda global sin referencias activas a `games.json` en `web/**`.

- Arcade pÃƒÂºblico: se agrega `N64 vertical mobile` para jugar desde celular con ROM legal del usuario.
	- Archivos creados: `web/Juegos/n64/index.html`, `web/Juegos/n64/styles.css`, `web/Juegos/n64/n64-wrapper.js`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la pÃƒÂ¡gina N64 deja de ser un scaffold remoto y pasa a cargar `n64js` dentro de `iframe.srcdoc` mismo-origen para que el mÃƒÂ³vil pueda enviar controles tÃƒÂ¡ctiles reales al core. La ROM legal del usuario se persiste en IndexedDB y los botones `Guardar` y `Cargar` respaldan/restauran la memoria del cartucho por `rominfo.id`, ÃƒÂºtil para tÃƒÂ­tulos como Super Mario 64 cuando se guarda primero dentro del propio juego.
	- VerificaciÃƒÂ³n: `get_errors` sin errores en `web/Juegos/n64/index.html`, `web/Juegos/n64/styles.css`, `web/Juegos/n64/n64-wrapper.js` y `web/Juegos/menu_juegos.html`.

- Licencias: el checkout cambia a activacion sin pasarela cuando el total queda en cero y bloquea la repeticion gratuita por empresa.
	- Archivos creados: `backend/db/licencias_gratis.go`.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `backend/main.go`, `web/pagar_licencia.html`, `web/elegir_licencia.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el checkout de licencias ahora calcula descuento y total real antes de abrir la pasarela; si el valor base es cero o el cÃƒÂ³digo deja el total en cero, ofrece `Activar licencia` en lugar de cobrar. AdemÃƒÂ¡s, una licencia gratis solo puede implementarse una vez por empresa y el selector de empresas mantiene tarjetas de altura uniforme.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'Test(LicenciaCheckoutSummaryHandlerAllowsZeroTotalByConfiguredDiscount|ActivateLicenciaSinPagoHandlerBlocksRepeatedFreeLicensePerEmpresa|WompiCreateNequiTransactionHandlerRejectsZeroTotalAndSuggestsActivation)$' -count=1` y tarea `validar-permisos-selector-empresas-5`.

- Inventario y productos: se separan las subpaginas de `productos`, `bodegas` y `categorias` sin duplicar la logica del modulo.
	- Archivos modificados: `web/administrar_empresa/administrar_productos.html`, `web/administrar_empresa/administrar_productos_menu.html`, `web/administrar_empresa/productos/administrar_productos.html`, `web/administrar_empresa/productos/bodegas.html`, `web/administrar_empresa/productos/categorias.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la vista principal del modulo ahora acepta `view=productos|bodegas|categorias`, las rutas legacy de `productos/` quedan como wrappers que conservan `empresa_id` y el menu lateral carga cada frente en su vista dedicada sin tocar endpoints ni contratos backend.
	- VerificaciÃƒÂ³n: `get_errors` sobre los HTML y documentos modificados.

- Operacion por empresa: `ventas_simple.html` agrega la variante `carrito_compacto` en el mismo panel por estacion.
	- Archivos modificados: `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/ventas_simple.html`, `web/js/ventas_simple.js`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: las estaciones con venta simple ahora abren el mismo flujo `ventas_simple.html` en modo `carrito_compacto`, con barra de acciones rapidas, total visible y acceso directo a busqueda, carrito, cobro, sincronizacion, nueva venta y correccion; la variante conserva los mismos endpoints de carrito/items y el aislamiento por `empresa_id`.
	- VerificaciÃƒÂ³n: `get_errors` sobre HTML, JS y documentacion modificada.

- Gobernanza tecnica: se agrega checklist documental para QA y soporte.
	- Archivos creados: `documentos/gobernanza_tecnica/runbooks/checklist_evidencia_documental_para_qa_y_soporte.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/gobernanza_tecnica/README.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega una checklist operativa breve para QA y soporte sobre repositorio documental, firmas y exportes regulatorios, y se refleja en la documentaciÃƒÂ³n general del proyecto y de arquitectura que un exporte no sustituye la evidencia versionada o firmada cuando el flujo es sensible.
	- VerificaciÃƒÂ³n: `get_errors` sobre la documentacion creada y modificada.

- Gobernanza tecnica: se endurece la reconciliacion documental y la evidencia regulatoria.
	- Archivos modificados: `documentos/gobernanza_tecnica/contratos/contrato_repositorio_documental_y_firmas_externas.md`, `documentos/gobernanza_tecnica/contratos/contrato_interoperabilidad_documental_contable_y_fiscal_externa.md`, `documentos/gobernanza_tecnica/contratos/contrato_reportes_contables_financieros_y_exportacion_multiformato.md`, `documentos/gobernanza_tecnica/runbooks/runbook_reconciliacion_documental_fiscal_y_contable_externa.md`, `documentos/gobernanza_tecnica/runbooks/runbook_versionado_documental_y_firmas_externas.md`, `documentos/gobernanza_tecnica/estandares_de_cambio_seguro.md`, `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la gobernanza tecnica ahora reconcilia explÃƒÂ­citamente repositorio documental, firmas, interoperabilidad fiscal/contable y exportes multiformato, y fija que un exporte regulatorio no sustituye la versiÃƒÂ³n documental vigente ni la firma asociada cuando el flujo exige evidencia reforzada.
	- VerificaciÃƒÂ³n: `get_errors` sobre la documentacion modificada.

- Gobernanza tecnica: se documentan repositorio documental y firmas externas.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_repositorio_documental_y_firmas_externas.md`, `documentos/gobernanza_tecnica/runbooks/runbook_versionado_documental_y_firmas_externas.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la gobernanza tecnica ahora cubre formalmente `/api/empresa/documentos/gestion` y `/api/empresa/documentos/firmas`, incluyendo reglas de acceso por rol/modulo, versionado con historial, herencia de permisos para firmas y el procedimiento operativo para diagnosticar documentos no visibles, versiones incompletas o firmas huerfanas.
	- VerificaciÃƒÂ³n: `get_errors` sobre la documentacion creada y modificada.

- Super administrador: se agrega interruptor global para activar o desactivar la IA desde configuraciÃƒÂ³n avanzada.
	- Archivos modificados: `backend/handlers/ai_config_handlers.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super.go`, `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`, `backend/handlers/chat_con_ia_global_super_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `web/super/configuracion_avanzada.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: configuraciÃƒÂ³n avanzada ahora puede apagar completamente el servicio IA mediante `ai.global.enabled`; cuando queda desactivado, el chat empresarial y el chat global super bloquean consultas nuevas y asÃƒÂ­ se libera carga del servidor. AdemÃƒÂ¡s, el botÃƒÂ³n `Probar IA` ejecuta una prueba real contra Ollama a travÃƒÂ©s del backend.
	- VerificaciÃƒÂ³n: prueba real por SSH en el VPS con `curl http://127.0.0.1:11434/api/tags` y `curl http://127.0.0.1:11434/api/generate`, ambas exitosas.

- Chat IA super y empresarial: se agrega aviso visible de servicio apagado y se explicita la prueba contra VPS.
	- Archivos modificados: `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/chat_con_ia_global.html`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: cuando la IA global estÃƒÂ¡ desactivada, ambos chats muestran un mensaje visible, deshabilitan el formulario y evitan que el usuario siga intentando consultar. En configuraciÃƒÂ³n avanzada, el botÃƒÂ³n queda rotulado como `Probar IA contra VPS` para reflejar que la prueba es real y no solo de catÃƒÂ¡logo.
	- VerificaciÃƒÂ³n: diagnÃƒÂ³stico del editor sin errores en los HTML modificados.

- Super administrador: se agrega chat IA global con contexto consolidado del sistema.
	- Archivos creados: `backend/handlers/chat_con_ia_global_super.go`, `backend/handlers/chat_con_ia_global_super_test.go`, `web/super/chat_con_ia_global.html`.
	- Archivos modificados: `backend/db/chat_inteligencia_artificial.go`, `backend/db/chat_inteligencia_artificial_test.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/main.go`, `web/super_administrador.html`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el panel super ahora tiene `chat_con_ia_global.html` con selector de modelo, historial y consultas sobre el contexto agregado de toda la base de datos; el historial global queda separado del chat por empresa mediante tablas `super_ai_*` y acceso exclusivo para sesiones `super_administrador`.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'TestSuperAI|TestModelosHandler|TestModeloPreferidoHandler|TestHistorialHandler' -count=1`; `go test ./db -run 'Test(EmpresaAI|SuperAI)' -count=1`.

- Chat IA empresarial: se habilita selector entre DeepSeek y Ambis Local por empresa.
	- Archivos modificados: `backend/handlers/ai_credentials_catalog.go`, `backend/handlers/ai_config_handlers.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/db/chat_inteligencia_artificial_test.go`, `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/super/configuracion_avanzada.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el chat IA empresarial ahora permite elegir entre `deepseek:deepseek-chat` y `ollama:ambis`; Ambis usa `codellama:7b` servido por Ollama en el VPS a traves de loopback, manteniendo el filtro por `empresa_id` y la preferencia persistida por cuenta Google autenticada.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'Test(ModelosHandler|ModeloPreferidoHandler|HistorialHandler|ConsultarHandler|AIModelsConfigHandlerSaveDeepSeekEncrypted)' -count=1`; `go test ./db -run 'TestEmpresaAIModeloPreferidoUpsertAndGet|TestRegisterEmpresaAIConsultaAcumulaUsoDiario' -count=1`.

- Gobernanza tecnica: se documentan integraciones externas y reconciliacion documental.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_integraciones_bancarias_y_conectores_externos.md`, `documentos/gobernanza_tecnica/runbooks/runbook_reconciliacion_documental_fiscal_y_contable_externa.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la gobernanza tecnica ahora cubre formalmente conectores API y bancarios, y agrega un procedimiento de reconciliacion entre compras, facturaciÃƒÂ³n, reintentos fiscales y repositorio documental.
	- VerificaciÃƒÂ³n: `get_errors` ejecutado sobre la documentacion creada y modificada, sin errores.

- Gobernanza tecnica: se documentan interoperabilidad documental y contingencias de integraciones externas.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_interoperabilidad_documental_contable_y_fiscal_externa.md`, `documentos/gobernanza_tecnica/runbooks/runbook_contingencias_integraciones_bancarias_y_conectores.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la gobernanza tecnica ahora cubre formalmente la interoperabilidad entre compras, facturaciÃƒÂ³n, repositorio documental y conciliacion fiscal, y agrega un runbook especÃƒÂ­fico para incidentes de conectores API e integraciones bancarias.
	- VerificaciÃƒÂ³n: `get_errors` ejecutado sobre la documentacion creada y modificada, sin errores.

- Gobernanza tecnica: se documentan cierre de periodo contable y conciliacion bancaria.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_conciliacion_bancaria_y_cierre_periodo_contable.md`, `documentos/gobernanza_tecnica/runbooks/runbook_cierre_periodo_y_conciliacion_bancaria.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la gobernanza tecnica ahora cubre formalmente el cierre y reapertura de periodos con evidencia obligatoria, la importacion idempotente de extractos, la conciliacion bancaria automatica y los bloqueos de movimientos cuando el periodo contable ya esta cerrado.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en la documentacion creada y modificada.

- Estaciones: se agrega tarjeta especial `YouTube` con vista embebida y ampliacion.
	- Archivos creados: `web/administrar_empresa/youtube_station_browser.html`.
	- Archivos modificados: `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/estaciones.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el modulo de estaciones ahora permite activar una tarjeta especial `YouTube` desde `estaciones_config`, mostrar una vista embebida adaptable al tamaÃƒÂ±o de la tarjeta y abrirla en un overlay aproximado de `500 x 500` mediante un cuadrito de maximizaciÃƒÂ³n `[]`.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/estaciones.html`, `web/administrar_empresa/youtube_station_browser.html` y `web/estilos.css`.

- Gobernanza tecnica: se documentan soporte remoto multiempresa y contingencias operativas de reportes.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_soporte_remoto_por_empresa_y_mesa_tecnica_central.md`, `documentos/gobernanza_tecnica/runbooks/runbook_reportes_programados_y_exportaciones_contables.md`, `documentos/gobernanza_tecnica/runbooks/runbook_soporte_remoto_sesiones_y_dispositivos.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la gobernanza tecnica ahora cubre formalmente el modulo de soporte remoto por empresa, el portal publico del agente, la mesa tecnica super y los procedimientos de contingencia para reportes programados, exportaciones contables y sesiones/dispositivos remotos.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en la documentacion creada y modificada.

- Gobernanza tecnica: se documentan DIAN, alertas de reinicio y reportes multiformato.
	- Archivos creados: `documentos/gobernanza_tecnica/runbooks/runbook_dian_set_pruebas_y_diagnostico_oficial.md`, `documentos/gobernanza_tecnica/runbooks/runbook_alertas_reinicio_y_monitoreo_gmail_smtp.md`, `documentos/gobernanza_tecnica/contratos/contrato_reportes_contables_financieros_y_exportacion_multiformato.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la gobernanza tecnica ahora cubre el runbook del soporte DIAN real que existe hoy, el runbook de Gmail SMTP y alertas de reinicio del backend, y el contrato del modulo de reportes empresariales y globales super con datasets canonicos, exportacion `json/csv/txt/xls/pdf`, plantillas, programacion y validacion de consistencia.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en la documentacion creada y modificada.

- Gobernanza interna: se implementa un equipo base de cuatro agentes con direccion centralizada en `agente_go`.
	- Archivos creados: `.github/agents/agente_backend_db.agent.md`, `.github/agents/agente_frontend_ux.agent.md`, `.github/agents/agente_qa_operacion.agent.md`, `.github/agents/README.md`.
	- Archivos modificados: `.github/agents/agente_go.agent.md`, `copilot-instructions.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `agente_go` queda formalizado como agente principal, seleccionado por defecto a nivel de gobernanza del repositorio y responsable de dirigir a backend/DB, frontend/UX y QA/operacion, integrando una sola salida tÃƒÂ©cnica y documental.
	- VerificaciÃƒÂ³n: validacion documental y de estructura del equipo interno de agentes completada en el repositorio.

- Gobernanza tecnica: se documenta formalmente el ciclo de facturacion electronica y documentos transaccionales.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_facturacion_electronica_y_documentos_transaccionales.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la capa de gobernanza ahora cubre la maquina de estados documental de facturaciÃƒÂ³n, la persistencia comun en `empresa_facturacion_documentos`, el selector `modo_documento_venta`, la cola de reintentos, la reconciliacion fiscal y la base operativa actual de DIAN Colombia, dejando explicito lo que aun esta pendiente del transporte oficial.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Gobernanza tecnica: se documenta formalmente la capa de permisos_contexto y wrappers de rutas empresariales.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_permisos_contexto_y_wrappers_api_empresa.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la capa de gobernanza ahora cubre la frontera de autorizacion para `/api/empresa/*`, incluyendo wrappers por modulo, endpoint `permisos_contexto`, overrides por rol, restricciones por licencia y aprobacion trazable en operaciones sensibles de seguridad.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Gobernanza tecnica: se documenta formalmente la venta publica empresarial por empresa.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_venta_publica_empresarial_por_empresa.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la capa de gobernanza ahora cubre el contrato del modulo de venta publica por `empresa_id`, incluyendo configuracion de tienda, catalogo, ordenes publicas, pagos Wompi/Epayco y consulta de estado con exposicion segura de datos.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Gobernanza tecnica: se documenta formalmente la autenticacion multirol y el arranque local PostgreSQL por tunel.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_autenticacion_administrativa_y_usuarios_empresa.md`, `documentos/gobernanza_tecnica/runbooks/runbook_arranque_postgresql_tunel_local.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la capa de gobernanza ahora cubre el contrato de autenticacion administrativa y de usuarios de empresa, junto con el runbook del arranque local del backend cuando PostgreSQL del VPS se consume por tunel SSH y DSN reescrito hacia `DB_VPS_LOCAL_PORT`.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Gobernanza tecnica: se documenta formalmente el flujo de estaciones, sensores y venta simple por estacion.
	- Archivos creados: `documentos/gobernanza_tecnica/contratos/contrato_estaciones_sensores_ventas_simple.md`, `documentos/gobernanza_tecnica/runbooks/runbook_estaciones_sensores_ventas_simple.md`.
	- Archivos modificados: `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la capa de gobernanza ahora cubre el flujo de `estaciones_config`, sensores por `last_seen`, carrito base canonico `EST-empresa-estacion`, cierre de `pagar_estacion`, metricas de estacion y recuperacion de incidentes sin cambiar permisos ni wrappers.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Gobernanza tecnica: se crea la capa base de ADRs, contratos, runbooks y cambio seguro del repositorio.
	- Archivos creados: `documentos/README.md`, `documentos/gobernanza_tecnica/README.md`, `documentos/gobernanza_tecnica/plan_implementacion_gobernanza_tecnica.md`, `documentos/gobernanza_tecnica/estandares_de_cambio_seguro.md`, `documentos/gobernanza_tecnica/adr/ADR-0001-frontera-multiempresa-empresa-id.md`, `documentos/gobernanza_tecnica/adr/ADR-0002-postgresql-runtime-canonico-vps.md`, `documentos/gobernanza_tecnica/contratos/README.md`, `documentos/gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md`, `documentos/gobernanza_tecnica/runbooks/README.md`, `documentos/gobernanza_tecnica/runbooks/runbook_checkout_licencias.md`.
	- Archivos modificados: `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se formaliza la gobernanza tecnica del proyecto para mejorar decisiones arquitectonicas, cambios seguros, contratos de flujos criticos y respuesta ante incidentes repetidos, comenzando por el checkout publico de licencias.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos documentales creados y modificados.

- Arranque PostgreSQL: el backend ahora respeta el puerto del tunel local.
	- Archivos modificados: `backend/main.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `resolveRuntimePostgresDSN` reescribe los DSN hacia `DB_VPS_LOCAL_PORT` cuando `DB_VPS_TUNNEL_ENABLED=1`, evitando que `go run .` o el binario del backend fallen por autenticacion contra `127.0.0.1:5432` cuando la conexion valida al VPS pasa por otro puerto local del tunel.
	- VerificaciÃƒÂ³n: `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; `go run .` con `.env.local` y tunel activo.

- Configuracion empresarial: el bloque general `Productos y pedidos` ahora guarda en backend.
	- Archivos creados: `backend/db/empresa_configuracion_general.go`, `backend/handlers/empresa_configuracion_general.go`, `backend/handlers/empresa_configuracion_general_test.go`.
	- Archivos modificados: `backend/main.go`, `web/administrar_empresa/configuracion.html`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la seccion principal de configuracion empresarial deja de simular guardado local y pasa a persistir por empresa con `GET/PUT /api/empresa/configuracion_general`, incluyendo orden de servicio, descuentos y lector de codigo de barras.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos nuevos y modificados; `go test ./handlers -run '^(TestEmpresaConfiguracionGeneralHandlerGetAndSave|TestEmpresaVentaPublicaHandlerConfigCatalogoYToggle|TestEmpresaConfiguracionOperativaHandler(ConfigAndRole|PoliticaSimulacionHistorialYRollback))$' -count=1`.

- Configuracion empresarial: `Avanzada` ya no sale al panel super.
	- Archivos modificados: `web/administrar_empresa/configuracion_menu.html`, `web/administrar_empresa/configuracion.html`, `documentos/descripcion_de_modulos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el submenu de configuracion empresarial corrige el enlace `Avanzada` para que apunte a la seccion avanzada real dentro de la configuracion de empresa, evitando navegar por error a `/super/configuracion_avanzada.html`.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/administrar_empresa/configuracion_menu.html` y `web/administrar_empresa/configuracion.html`.

- Configuracion empresarial: `Permisos` e `Integraciones` dejan de ser placeholders y se valida el guardado real.
	- Archivos modificados: `web/administrar_empresa/configuracion_permisos.html`, `web/administrar_empresa/configuracion_integraciones.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la vista `Permisos` ahora consulta el contexto real de permisos por empresa y muestra la matriz disponible en solo lectura; la vista `Integraciones` reutiliza el endpoint real de `venta_publica?action=config` para cargar y guardar la configuracion de Wompi/Epayco y la tienda publica de la empresa.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/administrar_empresa/configuracion_permisos.html` y `web/administrar_empresa/configuracion_integraciones.html`; `go test ./handlers -run '^TestEmpresaVentaPublicaHandlerConfigCatalogoYToggle$' -count=1`; `go test ./handlers -run '^TestEmpresaConfiguracionOperativaHandler(ConfigAndRole|PoliticaSimulacionHistorialYRollback)$' -count=1`.

- Ventas simples por estacion: se agrega boton `Regresar a estaciones`.
	- Archivos modificados: `web/administrar_empresa/ventas_simple.html`, `web/js/ventas_simple.js`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la vista de carrito por estacion ahora ofrece un retorno directo a `administrar_empresa/estaciones.html`, manteniendo el `empresa_id` activo para volver al tablero de estaciones sin depender del historial del navegador.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/administrar_empresa/ventas_simple.html` y `web/js/ventas_simple.js`.

- Configuracion empresarial: el submenÃƒÂº `ConfiguraciÃƒÂ³n` ahora puede ocultar y mostrar su menÃƒÂº lateral en celular.
	- Archivos modificados: `web/administrar_empresa/configuracion_menu.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la pÃƒÂ¡gina `administrar_empresa/configuracion_menu.html` carga `menu.js`, adopta el wrapper `admin-sidebar-mobile-collapsible` y agrega el mismo botÃƒÂ³n final de `Ocultar menÃƒÂº` / `Mostrar menÃƒÂº` usado por otros shells administrativos, para que el submenÃƒÂº de configuraciÃƒÂ³n tambiÃƒÂ©n sea plegable en mÃƒÂ³vil.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/administrar_empresa/configuracion_menu.html`.

- Estaciones: se elimina el circulo inferior de la tarjeta y se conserva solo el indicador cuadrado del sensor.
	- Archivos modificados: `web/administrar_empresa/estaciones.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la tarjeta de estaciones deja de mostrar el circulo centrado inferior; solo queda el cuadrito superior derecho, listo para ponerse verde cuando el sensor de la estacion se active.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/administrar_empresa/estaciones.html` y `web/estilos.css`.

- Checkout de licencias: valida contexto multiempresa y corrige la empresa usada en el correo de activacion.
	- Archivos modificados: `backend/db/db.go`, `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el backend ahora resuelve la empresa por `empresa_id` logico al construir el correo de activacion, endurece los helpers de `pagos_epayco` y `pagos_wompi` con autorreparacion del esquema, y rechaza conciliaciones de `transaction_status` cuando la referencia pertenece a otra empresa o licencia distinta de la pagina abierta; el frontend envia ese contexto esperado en cada polling y lo mantiene al cerrar el pago aprobado.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'Test(EpaycoTransactionStatusHandlerActivatesOnceAndCapturesEmail|EpaycoTransactionStatusHandlerUsesEmpresaScopeForActivationMailBody|EpaycoTransactionStatusHandlerRejectsUnexpectedEmpresaContext|EpaycoWebhookHandlerFindsContextUsingInvoiceFallback|EpaycoTransactionStatusHandlerRetriesActivationEmailAfterWebhookActivatedFirst|WompiTransactionStatusHandlerAllowsReferenceLookup)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Selector de empresas: tarjetas mas pequeÃƒÂ±as y botones alineados al pie.
	- Archivos modificados: `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `seleccionar_empresa.html` mantiene su mismo flujo, pero las tarjetas del grid se compactan en escritorio y la botonera inferior queda centrada y pegada al pie de cada bloque para que la fila de acciones no quede flotando a media altura.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/estilos.css`.

- Super configuracion avanzada: el boton `Probar Gmail` ahora envia un correo real de prueba.
	- Archivos modificados: `backend/handlers/usuarios_empresa.go`, `backend/handlers/system_empresas_handlers_test.go`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la configuracion avanzada de Gmail deja de validar solo si existen credenciales y pasa a ejecutar un envio de prueba real a `powerfulcontrolsystem@gmail.com`, reutilizando la configuracion SMTP guardada en PostgreSQL; en pruebas automatizadas el flujo se captura en la tabla de notificaciones de test.
	- VerificaciÃƒÂ³n: `go test ./handlers -run '^TestGmailConfigHandlerTestActionCapturesNotification$' -count=1`.

- Arcade publico: Brigada burbujas 3D plus agrega joystick tactil, fullscreen y HUD de una mano.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el shooter movil sustituye el pad clasico por joystick tactil, pide pantalla completa al iniciar en celular y concentra arma/sector dentro del HUD del escenario para que el juego se pueda usar mejor con una mano.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

- Arcade publico: Brigada burbujas 3D plus concentra los accesos tacticos dentro del escenario para celular.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el shooter ya no obliga a desplazarse al rack de arsenal en pantallas pequenas; ahora arma rapida y pausa viven dentro del escenario con una barra tactica sincronizada con el HUD y los controles inferiores.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

- Arcade publico: Brigada burbujas 3D plus agrega arsenal, pickups y sectores abiertos/cerrados.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el shooter 3D simulado del arcade ahora incluye tres armas, pickups de salud y municion, sectores cerrados/abiertos/hibridos, perspectiva de suelo mas marcada y una IA que patrulla, busca, flanquea, dispara y convoca refuerzos, empujando la experiencia hacia un Doom caricaturesco sin librerias externas.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

- Licencias Epayco: el correo de activaciÃƒÂ³n ahora se reintenta de forma idempotente despuÃƒÂ©s de aprobados posteriores si la licencia ya quedÃƒÂ³ activa pero la notificaciÃƒÂ³n todavÃƒÂ­a no quedÃƒÂ³ confirmada.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el flujo Epayco deja de depender de una activaciÃƒÂ³n Ã¢â‚¬Å“reciÃƒÂ©n creadaÃ¢â‚¬Â para enviar el correo; ademÃƒÂ¡s, ahora tambiÃƒÂ©n recupera el `customer_email` cuando la validaciÃƒÂ³n lo devuelve anidado en `data`, evitando perder la notificaciÃƒÂ³n si el webhook aprobÃƒÂ³ primero o si el primer intento fallÃƒÂ³ temporalmente.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'TestEpayco(TransactionStatusHandlerActivatesOnceAndCapturesEmail|WebhookHandlerFindsContextUsingInvoiceFallback|TransactionStatusHandlerRetriesActivationEmailAfterWebhookActivatedFirst)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Arcade publico: Brigada burbujas 3D plus agrega boton Auto visible, ayuda de mira movil y ajuste fino de impacto.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: El HUD movil incorpora un toggle directo de auto-disparo y el panel tactil gana intensidad de feedback y ayuda suave de mira configurable para mejorar la respuesta en celular sin quitar control manual.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

- Arcade publico: Brigada burbujas 3D plus activa auto-disparo y preset facil por defecto en movil.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: El juego completa el HUD y el panel tactil, fuerza una migracion unica de preferencias antiguas para dejar `Auto ON`, subir la ayuda de mira y suavizar el control desde el primer arranque en celular.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

## 2026-04-17

- Gobernanza interna: se agrega semaforo ejecutivo por modulo y se endurece el rechazo de cierres sin evidencia por frente.
	- Archivos modificados: `.github/agents/protocolo_delegacion.md`, `.github/agents/plantilla_trabajo_por_modulo.md`, `.github/agents/agente_backend_db.agent.md`, `.github/agents/agente_frontend_ux.agent.md`, `.github/agents/agente_qa_operacion.agent.md`, `.github/agents/README.md`, `copilot-instructions.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el equipo de agentes ya cuenta con un semaforo ejecutivo `Rojo/Amarillo/Verde`, un ejemplo completo extremo a extremo para delegacion real y reglas explicitas para que backend, frontend y QA rechacen cierres sin evidencia minima suficiente.
	- VerificaciÃƒÂ³n: validacion documental y de consistencia interna completada sobre el endurecimiento final del protocolo del equipo.

- Gobernanza interna: se agregan tabla rapida por modulo, ejemplos reales y endurecimiento de cierre para modulos criticos del equipo de agentes.
	- Archivos modificados: `.github/agents/protocolo_delegacion.md`, `.github/agents/plantilla_trabajo_por_modulo.md`, `.github/agents/agente_go.agent.md`, `.github/agents/README.md`, `copilot-instructions.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el protocolo ahora incluye una tabla corta por modulo para consulta inmediata, ejemplos reales de delegacion y una regla mas dura para que `agente_go` no cierre modulos criticos sin la participacion obligatoria definida.
	- VerificaciÃƒÂ³n: validacion documental y de consistencia interna completada sobre la ampliacion del protocolo del equipo.

- Gobernanza interna: se formaliza el protocolo de delegacion y la plantilla comun de trabajo del equipo de agentes.
	- Archivos creados: `.github/agents/protocolo_delegacion.md`, `.github/agents/plantilla_trabajo_por_modulo.md`.
	- Archivos modificados: `.github/agents/README.md`, `.github/agents/agente_go.agent.md`, `.github/agents/agente_backend_db.agent.md`, `.github/agents/agente_frontend_ux.agent.md`, `.github/agents/agente_qa_operacion.agent.md`, `copilot-instructions.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `agente_go` ya no solo dirige al equipo; ahora tambiÃƒÂ©n aplica una matriz exacta de delegaciÃƒÂ³n por tipo de tarea y una plantilla de ejecuciÃƒÂ³n compartida por mÃƒÂ³dulo, mientras cada especialista queda priorizado por mÃƒÂ³dulos crÃƒÂ­ticos del sistema.
	- VerificaciÃƒÂ³n: validacion documental y de consistencia interna completada sobre la nueva capa de gobernanza del equipo.

- Ventas por estacion: compatibilidad PostgreSQL restaurada en carritos, metricas y documento de venta.
	- Archivos modificados: `backend/db/carritos_compras.go`, `backend/db/empresa_configuracion_avanzada.go`, `backend/db/documentos_transaccionales.go`, `backend/db/sql_compat.go`, `backend/main.go`, `backend/db/facturacion_electronica_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: las inserciones de carritos, items y metricas dejan de depender de `LastInsertId` y pasan a usar la capa portable motor legado retirado/PostgreSQL; ademas, la configuracion avanzada por empresa se regulariza antes de consultar `modo_documento_venta`, las tablas legacy de documentos transaccionales recuperan un `id` autogenerado valido y el backend sanea globalmente cualquier tabla PostgreSQL heredada con llave primaria `id` sin secuencia/default.
	- VerificaciÃƒÂ³n: `go test ./db -run 'Test(GetEmpresaConfiguracionAvanzadaRepairsMissingModoDocumentoVentaColumn|PrepareFacturacionDocumentoLegal|FacturacionElectronicaRetryUpsertGetAndList)' -count=1`; `go test ./handlers -run 'Test(VentaCarritoFacturaYResolucionImpresora|VentaCarritoGeneraComprobantePagoSegunConfiguracion)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Checkout de licencias: Epayco queda en tarjeta blanca, compacta y sin correo visible.
	- Archivos modificados: `web/pagar_licencia.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `pagar_licencia.html` elimina el bloque separado de formas de pago Epayco, oculta el campo de correo del panel y deja el checkout en una tarjeta blanca mas pequeÃƒÂ±a y centrada; `web/estilos.css` adapta el branding del panel para ese layout.


## 2026-04-17

- Selector de empresas: editar pasa al menu lateral y las tarjetas cambian de estilo.
	- Archivos modificados: `web/seleccionar_empresa.html`, `web/js/seleccionar_empresa.js`, `web/editar_empresa.html`, `web/js/editar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `seleccionar_empresa.html` elimina el boton `Editar` dentro de las tarjetas, aÃƒÂ±ade `Editar empresa` al menu lateral, conserva el orden del texto principal de cada tarjeta y adopta una presentacion visual nueva con botones cuadrados. La pantalla `editar_empresa.html` queda enfocada solo en editar o eliminar.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos de frontend modificados.

## 2026-04-17

- Arcade publico: Brigada burbujas 3D plus ahora tiene campaÃƒÂ±a larga, transformaciones y rivales de pasarela caricaturesca.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `web/Juegos/menu_juegos.html`, `web/img/juegos/brigada_burbujas_3d.svg`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el shooter 3D simulado del arcade crece a cinco niveles, incorpora tres poderes de transformacion, enemigos con IA mas agresiva y una presentacion visual renovada en el lobby y la portada.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`, `web/Juegos/menu_juegos.html` y `web/img/juegos/brigada_burbujas_3d.svg`.

## 2026-04-17

- Arcade publico: Brigada burbujas 3D plus refuerza modo movil y ambiente claro.
	- Archivos modificados: `web/Juegos/brigada_burbujas_3d_plus.html`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el shooter 3D simulado del arcade ajusta su arte a una paleta pastel de dibujos animados, agrega apuntado tactil sobre el escenario y mejora el layout responsive para movil.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html`.

## 2026-04-17

- Selector de empresas: administradores pueden crear su primera empresa sin 403.
	- Archivos modificados: `backend/utils/utils.go`, `backend/handlers/system_empresas_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `AuthMiddleware` vuelve a permitir `POST /super/api/empresas` al rol `administrador` para que `seleccionar_empresa.html` pueda dar de alta empresas propias, manteniendo restringido el resto del panel super.
	- VerificaciÃƒÂ³n: `go test ./handlers -run '^TestNuevoAdminRegistradoPuedeCrearSuPrimeraEmpresaViaRutaSuperProtegida$' -count=1`; `go test ./utils -run '^TestSuperEndpointsPermisosPorRol$' -count=1`.

## 2026-04-18

- Arcade publico: se agrega Brigada burbujas 3D plus como decimo juego activo del portal.
	- Archivos creados: `web/Juegos/brigada_burbujas_3d_plus.html`, `web/img/juegos/brigada_burbujas_3d.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el arcade del portal incorpora un shooter original con raycasting, controles tactiles, enemigos caricaturescos y guardado local de record, elevando el lobby a diez juegos publicos activos.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/brigada_burbujas_3d_plus.html` y `web/Juegos/menu_juegos.html`.

## 2026-04-17

- Seguridad VPS del super: vista en modo oscuro.
	- Archivos modificados: `web/super/seguridad.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la pantalla `web/super/seguridad.html` cambia a una paleta oscura para alinearse visualmente con el resto del panel super, sin alterar endpoints ni permisos.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/super/seguridad.html`.

## 2026-04-17

- editar_empresa: se agrega un boton `Comprar licencia` visible solo cuando la empresa tiene licencia vencida y no conserva una licencia vigente activa.
- ventas por empresa: se agrega `modo_documento_venta` para que cada empresa elija entre `factura_electronica` y `comprobante_pago`, y el cierre de `pagar_estacion` genera automaticamente el documento correspondiente.
- documentacion y arbol operativo: se corrigen fechas futuras en la documentacion principal y se registra la salida del arbol actual de los scripts de revision ortografica (`scripts/spellcheck.*`, `scripts/spell_whitelist.txt`, `scripts/README-spellcheck.md`).

## 2026-04-17
- Menus administrativos: boton movil para ocultar y mostrar el sidebar.
	- Archivos modificados: `web/administrar_empresa.html`, `web/super_administrador.html`, `web/seleccionar_empresa.html`, `web/menu.js`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: las vistas `administrar_empresa.html`, `super_administrador.html` y `seleccionar_empresa.html` incorporan un boton final visible solo en movil para ocultar o mostrar el menu lateral, manteniendo intacta la navegacion en escritorio.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos HTML, `web/menu.js` y `web/estilos.css`.

## 2026-04-17
- Empresas super: nueva pagina editar_empresa con eliminacion total confirmada.
	- Archivos creados: `backend/db/empresas_delete.go`, `web/editar_empresa.html`, `web/js/editar_empresa.js`.
	- Archivos modificados: `backend/handlers/system_empresas_handlers.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/utils/utils.go`, `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se aÃƒÂ±ade `editar_empresa.html` para actualizar nombre y descripcion de la empresa seleccionada y se incorpora `action=eliminar_total` en `/super/api/empresas`, que purga datos relacionados por `empresa_id` en la base operativa y en la base super tras confirmar el nombre exacto.
	- VerificaciÃƒÂ³n: `go test ./handlers -run '^TestEmpresasHandlerEliminarTotalPurgaDatosRelacionados$' -count=1 -v -timeout 60s`; `go test ./handlers -run '^TestAdministradorPuedeEditarYEliminarEmpresaDesdeRutaSuperProtegida$' -count=1 -v -timeout 60s`; `go test ./handlers -run '^TestSuperEndpointsPermisosPorRol$' -count=1 -v -timeout 60s`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

## 2026-04-17
- Descarga de empresa: ruta corta funcional, exportacion en la misma vista y modo oscuro.
	- Archivos modificados: `backend/main.go`, `web/descargar_informacion_de_la_empresa.html`, `web/js/descargar_informacion_de_la_empresa.js`, `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la vista de descarga consolidada ahora responde tambien en `/descargar_informacion_de_la_empresa`, ejecuta las descargas PDF/XLS/CSV/JSON/TXT dentro de la misma pantalla con manejo de errores y usa una interfaz oscura dedicada.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `backend/main.go`, `web/descargar_informacion_de_la_empresa.html`, `web/js/descargar_informacion_de_la_empresa.js` y `web/estilos.css`.

- Selector de empresas: administradores no-super recuperan las lecturas iniciales.
	- Archivos modificados: `backend/utils/utils.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se corrige el `403` que impedÃƒÂ­a abrir `seleccionar_empresa.html` a cuentas con rol `administrador`, habilitando solo `GET /super/api/empresas`, `GET /super/api/tipos_empresas` y `GET /super/api/licencias` para esa vista, sin devolver permisos de escritura ni acceso al resto del panel super.
	- VerificaciÃƒÂ³n: `go test ./handlers -run '^TestSuperEndpointsPermisosPorRol$' -count=1`; `go test ./handlers -run '^TestNuevoAdminRegistradoNoObtieneAccesoSuperParaCrearEmpresa$' -count=1`; `go test ./utils -run '^TestAuthMiddlewareAllowsPublicLicenciaPaymentRoutesWithoutSession$' -count=1`.

- Registro de contrasena Google: se elimina `Continuar` y queda solo `Guardar` centrado.
	- Archivos modificados: `web/registrar_contrasena_usuario_de_google.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la pantalla `registrar_contrasena_usuario_de_google.html` ahora muestra un unico CTA de guardado, centrado, para reforzar que este paso no debe saltarse.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/registrar_contrasena_usuario_de_google.html`.

- Checkout Epayco: retorno con referencia real y pantalla de pago exitoso.
	- Archivos creados: `web/epayco/pago_exitoso.html`.
	- Archivos modificados: `web/epayco/respuesta.html`, `web/pagar_licencia.html`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el retorno web de Epayco ahora prioriza `x_ref_payco` y el estado real de la pasarela para que la validacion posterior sÃƒÂ­ pueda activar la licencia; cuando el backend confirma `APPROVED`, el usuario sale a una pantalla de pago exitoso y de allÃƒÂ­ vuelve a `seleccionar_empresa.html`.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los HTML modificados; `go test ./handlers -run 'TestEpayco(TransactionStatusHandlerFindsContextUsingInvoiceWhenGatewayIDsDiffer|TransactionStatusHandlerActivatesOnceAndCapturesEmail|WebhookHandlerFindsContextUsingInvoiceFallback|CreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|TransactionStatusHandlerPreservesPendingOnGenericValidationError)' -count=1`.

- Seleccionar empresa: se corrige el error `escapeHtml is not defined`.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el listado del panel `seleccionar_empresa.html` vuelve a renderizar correctamente porque se restaura el helper local de escape HTML usado por las tarjetas de empresa tras el refactor del flujo de descarga.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js`.

- Autenticacion administrativa: super restringido a powerfulcontrolsystem@gmail.com.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/utils/utils.go`, `backend/handlers/auth_admin_handlers_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_del_proyecto`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el registro publico de administradores, el login por correo y el callback de Google ya no elevan cuentas nuevas o legacy a `super_administrador`. Solo `powerfulcontrolsystem@gmail.com` mantiene ese rol en el flujo publico; el resto queda como `administrador` y no entra a `/super/*`.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "Test(AdminRegisterHandlerCreatesPendingAdminAndCapturesConfirmationMail|AdminLoginHandlerCreatesSessionForConfirmedAdmin|AdminLoginHandlerKeepsGenericAdminWithoutSuperPrivileges|AdminRegisterHandlerReservedEmailKeepsSuperRole|HandleGoogleCallbackNewEmailKeepsAdministradorRole|NuevoAdminRegistradoNoObtieneAccesoSuperParaCrearEmpresa)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Accept: ocultar la metadata visible del contrato.
	- Archivos modificados: `web/accept.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `accept.html` ya no muestra la linea `Version vigente | actualizada`, pero mantiene el enlace al contrato completo y sigue resolviendo internamente la version para apuntar a la ruta correcta.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/accept.html`.

- Home publico: CTA inferior fijo y texto de tarjetas con mayor contraste.
	- Archivos modificados: `web/estilos.css`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: las tarjetas de `index.html` ahora reservan el espacio flexible para la descripcion, dejan el boton `Explorar oferta` siempre abajo y centrado, y mejoran la legibilidad del titulo y el texto con tipografia clara e iluminado exterior negro suave sin agregar paneles de fondo al contenido textual.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/estilos.css` y `web/index.html`.

- Empresas super: prueba end-to-end para usuario nuevo creando su primera empresa.
	- Archivos modificados: `backend/handlers/system_empresas_handlers_test.go`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega una regresion que cubre el caso reportado de un administrador nuevo que se registra, confirma su cuenta, inicia sesion y crea su primera empresa mediante `POST /super/api/empresas` bajo `AuthMiddleware`.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "Test(NuevoAdminRegistradoPuedeCrearSuPrimeraEmpresaViaRutaSuperProtegida|AdminRegisterHandlerCreatesPendingAdminAndCapturesConfirmationMail|AdminLoginHandlerCreatesSessionForConfirmedAdmin|AdminLoginHandlerPromotesLegacySelfRegisteredAdminToSuper)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Administradores autoregistrados: crear empresa deja de fallar para cuentas nuevas.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/utils/utils.go`, `backend/handlers/auth_admin_handlers_test.go`, `web/js/seleccionar_empresa.js`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el registro administrativo publico ahora crea cuentas con rol `super_administrador`, las cuentas legacy autoregistradas se promueven automaticamente al entrar o al tocar rutas `/super/*`, y el formulario de `seleccionar_empresa.html` envia correctamente el tipo de empresa al crearla.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestAdmin(RegisterHandlerCreatesPendingAdminAndCapturesConfirmationMail|LoginHandlerCreatesSessionForConfirmedAdmin|LoginHandlerPromotesLegacySelfRegisteredAdminToSuper)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Home publico: CTA de tarjetas anclado abajo y descripcion en oscuro.
	- Archivos modificados: `web/estilos.css`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: las tarjetas de `index.html` ahora dejan el boton `Explorar oferta` siempre al pie y centrado, mientras la descripcion usa un color oscuro y negrita para ganar contraste frente al fondo de cada tarjeta.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/estilos.css`.

- Venta pÃƒÂºblica por empresa: Wompi y Epayco con credenciales propias por empresa.
	- Archivos modificados: `backend/db/venta_publica.go`, `backend/handlers/venta_publica.go`, `backend/handlers/payments_handlers.go`, `backend/handlers/venta_publica_test.go`, `backend/main.go`, `web/venta_publica.html`, `web/administrar_empresa/venta_publica.html`, `web/administrar_empresa/configuracion.html`, `documentos/estructura_bd.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: cada empresa puede activar o desactivar Wompi y Epayco con sus propias credenciales, administrar esas llaves desde `configuracion.html` y reutilizarlas en su tienda pÃƒÂºblica; ademÃƒÂ¡s los webhooks de ambas pasarelas ya actualizan las ÃƒÂ³rdenes de `empresa_venta_publica_ordenes`.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "Test(EmpresaVentaPublicaHandlerConfigCatalogoYToggle|PublicVentaPublicaHandlerCatalogoYPagoConWompiInactivo|PublicVentaPublicaHandlerEstadoPagoRequiereOrderCode)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

- Seleccionar empresa: nueva pagina para descargar informacion empresarial en formatos profesionales.
	- Archivos creados: `backend/handlers/system_empresas_export.go`, `web/descargar_informacion_de_la_empresa.html`, `web/js/descargar_informacion_de_la_empresa.js`.
	- Archivos modificados: `backend/handlers/system_empresas_handlers.go`, `backend/handlers/system_empresas_handlers_test.go`, `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el boton de descarga de las tarjetas en `seleccionar_empresa.html` ahora abre una pagina dedicada que consolida la informacion de la empresa seleccionada y permite descargarla en `PDF`, `XLS`, `CSV`, `JSON` y `TXT` desde nuevas acciones protegidas del endpoint `/super/api/empresas`.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestEmpresasHandler(ResumenDescargaYExport|ImpactoDesactivacion|DesactivarConImpactoYForce)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Facturacion DIAN: fase 1 base con UBL 2.1, firma XAdES y diagnostico oficial.
	- Archivos modificados: `backend/handlers/modulos_faltantes.go`, `backend/handlers/modulos_faltantes_test.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el modulo DIAN incorpora una fase 1 base para generar XML UBL 2.1 estructural, incrustar una firma XMLDSig/XAdES base y emitir un diagnostico de brechas frente al contrato oficial DIAN, manteniendo separado el transporte SOAP/WSDL definitivo.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'TestEmpresaDIANColombiaHandler(GenerarXMLUBLBase|FirmarXMLXAdESBase|DiagnosticoOficial|FirmaEnvioYAcuseReal|EnviarSetPruebas|SoftwareCompartidoMultiempresa|GuiaOnboardingYValidarCredenciales|SubirFirma)' -count=1`.

- Seleccionar empresa: boton de descarga blanco, solo con icono y tooltip.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el boton de descarga dentro de las tarjetas de `seleccionar_empresa.html` deja de mostrar el texto `Descargar`, queda como icono blanco y muestra el tooltip `Descargar informacion de la empresa` al pasar el mouse.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js` y `web/estilos.css`.

- Checkout publico de licencias: resumen en tarjetas y medios Epayco visibles.
	- Archivos modificados: `web/pagar_licencia.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `pagar_licencia.html` ahora separa la licencia y los codigos de descuento en dos tarjetas con el mismo estilo comercial del home, agrega un bloque visible con las formas de pago de Epayco y recolorea su panel de checkout con branding propio.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/pagar_licencia.html` y `web/estilos.css`.

- Elegir licencia: orden ascendente de menor a mayor valor.
	- Archivos modificados: `web/elegir_licencia.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el catalogo de licencias disponibles para pagar ahora se ordena en frontend desde el menor valor hasta el mayor, de modo que las opciones mas economicas aparezcan primero sin alterar el flujo hacia `pagar_licencia.html`.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/elegir_licencia.html`.

- Checkout publico de licencias: Epayco ahora activa la licencia, envia correo y cierra correctamente los estados finales.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el post-pago de Epayco resuelve contexto por `invoice` cuando la pasarela devuelve IDs externos, activa la licencia solo una vez, conserva `customer_email` para enviar el correo de confirmacion y evita que el formulario vuelva a marcar `pending` cuando el retorno ya es rechazado o fallido.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestEpayco(TransactionStatusHandlerFindsContextUsingInvoiceWhenGatewayIDsDiffer|TransactionStatusHandlerActivatesOnceAndCapturesEmail|WebhookHandlerFindsContextUsingInvoiceFallback)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Elegir licencia: tarjetas mas compactas y sin textos de estado.
	- Archivos modificados: `web/elegir_licencia.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: las tarjetas de licencias disponibles para pagar dejan de mostrar `Estado: Activa/Inactiva` y `Disponible para asignacion inmediata` o su variante de asignacion. Ademas se compactan visualmente con menor padding, icono mas contenido y menos separacion vertical, manteniendo intacto el flujo hacia `pagar_licencia.html`.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/elegir_licencia.html`.

- Exportes operativos: descarga silenciosa sin sacar al usuario del mÃƒÂ³dulo.
	- Archivos modificados: `web/administrar_empresa/administrar_clientes.html`, `web/administrar_empresa/asistencia_empleados.html`, `web/administrar_empresa/backups.html`, `web/administrar_empresa/tarifas_por_dia.html`, `web/administrar_empresa/soporte_remoto.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: los exportes frecuentes de clientes, asistencia, backups, tarifas por dÃƒÂ­a y soporte remoto dejan de reemplazar la vista actual. El archivo se descarga en segundo plano y el usuario permanece en el mismo mÃƒÂ³dulo.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos modificados.

- Navegacion general: misma pestaÃƒÂ±a por defecto.
	- Archivos modificados: `web/super_administrador.html`, `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/js/seleccionar_empresa.js`, `web/login.html`, `web/registrar_nuevo_usuario_administrador.html`, `web/registrar_contrasena_usuario_de_google.html`, `web/super/venta_digital.html`, `web/super/pagina_principal.html`, `web/super/configuracion_avanzada.html`, `web/administrar_empresa/venta_publica.html`, `web/administrar_empresa/soporte_remoto.html`, `web/super/soporte_remoto.html`, `web/administrar_empresa/administrar_clientes.html`, `web/administrar_empresa/asistencia_empleados.html`, `web/administrar_empresa/backups.html`, `web/administrar_empresa/tarifas_por_dia.html`, `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/administrar_empresa/chat_y_tareas.html`, `web/index.html`, `web/Informacion_de_contacto.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la navegaciÃƒÂ³n normal del sistema deja de abrir pestaÃƒÂ±as nuevas y reutiliza la misma ventana actual. Se mantienen como excepciÃƒÂ³n solo el contrato, los tÃƒÂ©rminos legales de pasarela y los popups tÃƒÂ©cnicos de impresiÃƒÂ³n o vista previa documental.
	- VerificaciÃƒÂ³n: bÃƒÂºsqueda final de `target="_blank"|window.open(` limitada a excepciones esperadas; diagnÃƒÂ³stico del editor sin errores en los archivos modificados.

- Licencias super: valor 0 ya no se oculta en ediciÃƒÂ³n ni en listado.
	- Archivos modificados: `web/super/licencias.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el CRUD de licencias en panel super conserva `0` como valor valido visible en la tabla y en el formulario de ediciÃƒÂ³n, evitando que una licencia parezca vacÃƒÂ­a al reabrirla.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/super/licencias.html`.

- Licencias del selector: historial con vencimiento y renovacion.
	- Archivos modificados: `backend/db/db.go`, `backend/handlers/payments_handlers_test.go`, `web/super/licencias.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la ruta `super/licencias.html?scope=mine&con_empresa=1` deja de mostrar el CRUD y pasa a ser un historial de licencias pagadas o vencidas por empresa, con fecha de vencimiento visible, estados operativos y acceso a `Pagar nueva licencia` cuando la licencia esta por vencer o ya vencio. El backend reutiliza el mismo endpoint `/super/api/licencias` exponiendo empresa y fechas para ese flujo.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos modificados; `go test ./handlers -run "TestLicenciasHandlerGetReturnsHistorialFieldsForCreatorScope" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Checkout publico de licencias: Epayco migra a Smart Checkout v2.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `web/super/configuracion_avanzada.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el sistema deja de generar URLs manuales hacia `checkout.php`, porque ese flujo ya responde `AccessDenied`. Ahora el backend crea la sesion oficial Smart Checkout v2 en Apify y el frontend abre `checkout-v2.js` con `sessionId`, manteniendo las mismas rutas publicas de respuesta, verificacion y webhook.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'TestEpaycoCreateTransactionHandler(UsesConfiguredPublicBaseURLAndKeys|AllowsCheckoutWithoutPrivateKey|AcceptsSamboxAlias)|TestEpaycoTransactionStatusHandler(PreservesPendingOnGenericValidationError|FindsContextUsingInvoiceWhenGatewayIDsDiffer)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Crear clave por correo: ojo para mostrar u ocultar la contrasena.
	- Archivos modificados: `web/registrar_contrasena_usuario_de_google.html`, `web/js/registrar_contrasena_usuario_de_google.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la pagina `Crear clave para acceso por correo` ahora incluye un icono de ojo en ambos campos de contrasena para poder revisarla visualmente antes de guardarla.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos modificados.

- Elegir licencia: tarjetas con el mismo estilo del home.
	- Archivos modificados: `web/elegir_licencia.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la pagina `elegir_licencia.html` ahora renderiza las licencias con la misma estructura visual de tarjetas usada en `index.html`, manteniendo sin cambios el flujo de compra hacia `pagar_licencia.html`.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/elegir_licencia.html`.

- Reportes globales super: eleccion explicita de una empresa o varias.
	- Archivos modificados: `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js`, `web/estilos.css`, `backend/handlers/reportes_globales_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el modulo `Reportes globales` ahora permite escoger de forma explicita si el analisis se hace sobre una sola empresa o sobre varias. En modo singular la UI cambia a selector puntual y el frontend consulta la API usando `empresa_id`.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos modificados; `go test ./handlers -run "TestSuperReportesGlobalesHandlerFiltraYConsolidaPorAdministrador" -count=1`.

- Login administrativo: Google y correo quedan en una sola tarjeta visual.
	- Archivos modificados: `web/login.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el bloque de acceso por correo deja de renderizarse como un formulario en caja separado dentro de `login.html`. Google, correo, recuperaciÃƒÂ³n y reset ahora comparten el mismo contenedor visual principal.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/login.html` y `web/estilos.css`.

- Arcade publico: runtime comun de poderes y premios en los nueve juegos activos.
	- Archivos modificados: `web/Juegos/arcade_shared.js`, `web/Juegos/arcade_window.css`, `web/Juegos/patito_volando_plus.html`, `web/Juegos/serpiente_pixel_plus.html`, `web/Juegos/memoria_estelar_plus.html`, `web/Juegos/rebote_bloques_plus.html`, `web/Juegos/pacman_plus.html`, `web/Juegos/tetris_plus.html`, `web/Juegos/carton_fire_plus.html`, `web/Juegos/ajedrez_vs_ia_plus.html`, `web/Juegos/ajedrez_3d_plus.html`, `web/Juegos/menu_juegos.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el arcade publico queda unificado con una misma capa mobile-first de countdown, sonido, records, poderes y premios para los nueve juegos activos del lobby, con economia compartida ajustada para juegos de eventos rapidos y un lobby que muestra mejor el progreso personal y el ranking por titulo.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/arcade_shared.js`, `web/Juegos/arcade_window.css`, `web/Juegos/menu_juegos.html` y los cuatro juegos integrados en esta fase; busqueda de `createPowerSystem` presente en los 9 juegos activos.

## 2026-04-17
- Arcade publico: nuevo Ajedrez 3D plus con cinco dificultades.
	- Archivos creados: `web/Juegos/ajedrez_3d_plus.html`, `web/img/juegos/ajedrez_3d.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega una nueva variante publica de ajedrez al arcade del portal, con tablero en perspectiva 3D simulada, cronometro arcade, cuenta regresiva de inicio y cinco niveles de dificultad contra la IA.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/ajedrez_3d_plus.html` y `web/Juegos/menu_juegos.html`.

## 2026-04-17
- Reportes globales super: graficos y lectura ejecutiva.
	- Archivos modificados: `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la vista global del panel super ahora aÃƒÂ±ade graficos comparativos y una lectura ejecutiva automÃƒÂ¡tica del consolidado de empresas seleccionadas, sin cambiar el modelo de permisos ni crear dependencias frontend externas.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en HTML/JS modificados.

## 2026-04-16
- Reportes globales super: consolidados por administrador creador.
	- Archivos modificados: `backend/db/db.go`, `backend/main.go`, `backend/handlers/reportes_globales.go`, `backend/handlers/reportes_globales_test.go`, `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la vista `Reportes globales` del panel super ahora permite consultar reportes generales, mezclados o individuales de las empresas creadas por el administrador autenticado, reutilizando los datasets empresariales existentes y manteniendo el aislamiento por creador.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestSuperReportesGlobalesHandlerFiltraYConsolidaPorAdministrador" -count=1`; diagnostico del editor sin errores en los archivos nuevos y modificados.

## 2026-04-17
- Seleccionar empresa: licencia y descarga quedan en una sola fila.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el render de las tarjetas de seleccion de empresa ahora agrega el boton verde de descarga dentro del mismo bloque `card-actions` que usa el indicador de licencia, evitando que queden en filas separadas.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js`.

## 2026-04-17
- Licencias super: actualizacion compatible con esquemas legacy sin `fecha_actualizacion`.
	- Archivos modificados: `backend/db/db.go`, `backend/db/licencias_schema_test.go`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la edicion y activacion de licencias en el panel super ya no fallan cuando la tabla `licencias` viene de un esquema antiguo que no incluye `fecha_actualizacion`. El backend intenta regularizar el esquema y, si esa columna sigue ausente, aplica un `UPDATE` de compatibilidad para guardar precio y estado.
	- VerificaciÃƒÂ³n: `go test ./db -run "TestEnsureLicenciasSchemaAddsValorInmotor legado retirado|TestCreateAndUpdateLicenciaRepairMissingValorColumn|TestUpdateLicenciaRepairsMissingFechaActualizacionColumn" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

## 2026-04-16
- Checkout publico de licencias: Epayco redirige la misma pestaÃƒÂ±a al checkout.
	- Archivos modificados: `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el flujo de Epayco deja de depender de una pestaÃƒÂ±a emergente y ya no arranca el polling antes de que el usuario entre a la pasarela. `pagar_licencia.html` guarda la referencia pendiente, redirige la misma pestaÃƒÂ±a a Epayco y usa `/epayco/respuesta.html` para retomar la verificacion al volver.
	- VerificaciÃƒÂ³n: `GET /api/public/licencias/payment_methods` con `epayco.available=true`; `POST /epayco/create_transaction` con `checkout_url` publica valida; `GET /epayco/transaction_status?reference=<referencia_recien_creada>` con `PENDING` y `context_found=true`; diagnostico del editor sin errores en `web/pagar_licencia.html`.

- Menu flotante: separaciÃƒÂ³n frente a botones superiores cercanos.
	- Archivos modificados: `web/menu.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el menu flotante compartido ahora reserva espacio en encabezados y barras de acciones para no montarse sobre botones ubicados en la parte superior derecha de algunas paginas.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos modificados.

## 2026-04-16
- Facturacion electronica: suite `db` estable aun con entorno local en PostgreSQL.
	- Archivos modificados: `backend/db/finanzas_test.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `openFinanzasTestDB` ahora fija el dialecto `motor_legado_retirado` para evitar que la suite de facturacion electronica y documentos transaccionales herede `DB_DIALECT=postgres` del entorno local y falle con SQL incompatible.
	- VerificaciÃƒÂ³n: `go test ./db -run "Test.*(Facturacion|DIAN|DocumentoFacturacion)" -count=1`; `go test ./handlers -run "Test(VentaCarritoFacturaYResolucionImpresora|EmpresaDIANColombiaHandler.*|EmpresaFacturacionElectronicaReintentosYReconciliacion|EmpresaFacturacionElectronicaEmiteEventoContable|EmpresaFacturacionTransaccional.*)" -count=1`.

- Pagina principal super: el campo de cantidad deja de mostrar un `5` temporal antes de cargar la configuracion real.
	- Archivos modificados: `web/super/pagina_principal.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el editor de `pagina_principal` ahora deja `ppCantidad` en estado de carga hasta recibir la configuracion persistida y sincroniza la cantidad con el numero real de tarjetas, evitando la confusion visual entre el panel super, `index.html` y `/descripcion_de_los_sistemas.ht`.
	- VerificaciÃƒÂ³n: consulta local a `/api/public/pagina_principal` con `cantidad=7`; revision directa del flujo de carga del editor super.

## 2026-04-17
- Ventas y facturacion: prueba integrada de carrito pagado con resoluciÃƒÂ³n de impresora.
	- Archivos creados: `backend/handlers/carrito_facturacion_impresion_test.go`.
	- Archivos modificados: `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega una prueba de integracion de handlers que valida una venta pagada en carrito, la emision documental de factura electronica y la resoluciÃƒÂ³n de la impresora `factura_caja` para el flujo de impresion soportado hoy.
	- VerificaciÃƒÂ³n: `go test ./handlers -run TestVentaCarritoFacturaYResolucionImpresora -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Seleccionar empresa: restauracion del formato clasico de tarjetas.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el selector de empresas del panel super vuelve al formato simple de tarjetas `portal-card warm` usado anteriormente, retirando la presentacion enriquecida reciente.
	- VerificaciÃƒÂ³n: revision del render en `web/js/seleccionar_empresa.js`; recomendada validacion visual en `seleccionar_empresa.html`.

- Portal publico: menu flotante navegable en celular.
	- Archivos modificados: `web/menu.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: manipulation`.
	- VerificaciÃƒÂ³n: revision del flujo JS/CSS; recomendada validacion manual en movil o emulacion tactil.

- Usuarios de empresa: portal publico con contrato vigente y subdominio dedicado.
	- Archivos modificados: `backend/db/usuarios_empresa.go`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/main.go`, `web/login_usuario.html`, `web/js/login_usuario.js`, `web/estilos.css`, `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/administrar_empresa/administrar_usuarios.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `login_usuario.html` pasa a ser el portal publico de usuarios internos creados por administradores, con registro por invitacion, recuperacion, reset, cambio de contrasena y aceptacion obligatoria del contrato vigente. El backend persiste esa aceptacion en `users`, los correos y el panel administrativo apuntan a `usuarios.powerfulcontrolsystem.com`, y el acceso final sigue entrando a `administrar_empresa.html` filtrado por rol.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestEmpresaUsuario(LoginHandlerSuccess|LoginHandlerRequiresContractAcceptance|SetPasswordHandlerSuccess|ResolveEmpresaUsuarioLoginURLUsesUsuariosSubdomain)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; diagnostico del editor sin errores en los archivos web modificados.

- Usuarios de empresa: login por subdominio propio de cada empresa.
	- Archivos modificados: `backend/handlers/usuarios_empresa.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/handlers/usuarios_empresa_seguridad_test.go`, `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/administrar_empresa/administrar_usuarios.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el enlace operativo del login de usuarios deja de resolverse a un host global fijo. Ahora se construye con el `empresa_slug` o `dominio_publico` configurado por empresa, tanto en el menu de `administrar_empresa` como en los correos de invitacion y recuperacion; la vista de administrar usuarios elimina el acceso duplicado fuera del menu.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestEmpresaUsuario(LoginHandlerSuccess|LoginHandlerRequiresContractAcceptance|SetPasswordHandlerSuccess|ResolveEmpresaUsuarioLoginURLUsesEmpresaSubdomain|PasswordRecoveryFlow|ChangePasswordFlow|ChangePasswordPolicyRejectsWeakPassword|LoginRequiresRotationWhenPolicyEnabled|NotificationsCaptureInMailTestMode)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; diagnostico del editor sin errores en los archivos modificados.

- Soporte remoto: limites por plan y mesa tecnica central multiempresa.
	- Archivos creados: `backend/handlers/super_soporte_remoto.go`, `backend/handlers/super_soporte_remoto_test.go`, `web/super/soporte_remoto.html`.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/db/soporte_remoto_test.go`, `backend/handlers/soporte_remoto.go`, `backend/handlers/soporte_remoto_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/main.go`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el modulo de soporte remoto ahora controla cupos de dispositivos, sesiones y minutos por empresa, persiste consumo mensual con intentos bloqueados y agrega una mesa tecnica central para `super_administrador` en `/super/api/soporte_remoto` y `super/soporte_remoto.html`.
	- VerificaciÃƒÂ³n: `go test ./db ./handlers -run "Test(SoporteRemotoDB|EmpresaSoporteRemotoHandler|PublicSoporteRemotoAgentHeartbeatAndStateUpdate|SuperSoporteRemotoHandlerListsCompaniesAndCreatesSession|SuperEndpointsPermisosPorRol)" -count=1`.

## 2026-04-16
- Arranque local: healthcheck robusto en `scripts/iniciar_servidor.ps1`.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el paso `8/8` deja de reportar timeout falso cuando el backend ya esta arriba. El script ahora usa el `PORT` efectivo, detecta listener TCP con API nativa/fallback y acepta una respuesta HTTP valida para confirmar disponibilidad.
	- VerificaciÃƒÂ³n: `. 'D:\powerfulcontrolsystem\scripts\iniciar_servidor.ps1'`.

- Backend: fix de compilacion en soporte remoto y bootstrap runtime.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/db/productos.go`, `backend/main.go`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se restaura el arranque local del backend corrigiendo la variable temporal usada al leer sesiones de soporte remoto, reescribiendo el cierre del bloque runtime en `main.go` y haciendo idempotente la regularizacion de columnas en PostgreSQL para evitar errores `column already exists` durante `scripts/iniciar_servidor.ps1`.
	- VerificaciÃƒÂ³n: `go build -o server.exe .` en `backend`; `.\scripts\iniciar_servidor.ps1 -Background`.

- Estaciones: sincronizacion backend del carrito base por estacion.
	- Archivos modificados: `backend/db/empresa_estacion_prefs.go`, `backend/db/empresa_estacion_prefs_test.go`, `backend/handlers/empresa_estacion_prefs.go`, `backend/handlers/empresa_estacion_prefs_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: guardar `estaciones_config` ya no depende del frontend para crear los carritos por defecto. El backend sincroniza automaticamente un carrito enlazado por estacion, corrige nombre/codigo/referencia cuando cambia la configuracion y lo deja en estado base `inactivo/cerrado` hasta su activacion operativa.
	- VerificaciÃƒÂ³n: `go test -work ./db -run "Test(EmpresaEstacionPrefs|SyncEmpresaEstacionCarritos)" -count=1`; `go test -work ./handlers -run "TestEmpresaEstacionPrefsHandler_UpsertAndIsolationByEmpresa|TestEmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Home publico: contacto centrado debajo de las tarjetas y deploy VPS con limpieza de procesos previos.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `scripts/sync_to_vps.ps1`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el home del portal deja `Informacion de contacto` como CTA centrado debajo del grid de tarjetas, manteniendo `Registrarse o iniciar sesiÃƒÂ³n` en la cabecera. En paralelo, el deploy remoto endurece el reinicio del backend: purga procesos viejos de `server_linux_amd64`, corrige la unidad `systemd` para evitar el warning de `StartLimitIntervalSec` mal ubicado y asegura que el binario nuevo quede activo al terminar `sync_to_vps`.
	- VerificaciÃƒÂ³n: `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; validacion de sintaxis PowerShell para `scripts/sync_to_vps.ps1`; diagnostico remoto de `systemctl status powerfulcontrolsystem` y `ss -ltnp` en el VPS.

- Checkout publico de licencias: Epayco acepta alias sambox como sandbox.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la normalizacion del modo de Epayco ahora tolera `sambox` como alias de `sandbox`, garantizando que el checkout de licencias permanezca en pruebas (`test=true`) aunque la configuracion manual use esa variante.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'TestEpaycoCreateTransactionHandler(AcceptsSamboxAlias|UsesConfiguredPublicBaseURLAndKeys|AllowsCheckoutWithoutPrivateKey)|TestResolvePaymentBaseURL(FallsBackToCanonicalDomainOnLocalhost|UsesConfiguredCanonicalDomain|IgnoresConfiguredLocalhostAndFallsBackToCanonicalDomain)|TestEpaycoTransactionStatusHandler(PreservesPendingOnGenericValidationError|FindsContextUsingInvoiceWhenGatewayIDsDiffer)' -count=1`.

- Checkout publico de licencias: Epayco legacy + metodo unico sin selector.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el checkout de Epayco vuelve a enviar `p_key` cuando la configuracion dispone de `private_key`, manteniendo compatibilidad con cuentas que exigen parametros legacy en `checkout.php`; ademas, `pagar_licencia.html` ya no muestra el cuadro de seleccion de forma de pago cuando solo una pasarela esta activa.
	- VerificaciÃƒÂ³n: pendiente ejecutar pruebas focalizadas de handlers y confirmar que la URL remota de checkout deje de responder `403 AccessDenied`.

- Arcade publico: set activo de ocho juegos compactos con popup fijo y pausa real.
	- Archivos creados: `web/Juegos/arcade_window.css`, `web/Juegos/patito_volando_plus.html`, `web/Juegos/serpiente_pixel_plus.html`, `web/Juegos/memoria_estelar_plus.html`, `web/Juegos/rebote_bloques_plus.html`, `web/Juegos/pacman_plus.html`, `web/Juegos/tetris_plus.html`, `web/Juegos/carton_fire_plus.html`, `web/Juegos/ajedrez_vs_ia_plus.html`, `web/img/juegos/pacman.svg`, `web/img/juegos/tetris.svg`, `web/img/juegos/carton_fire.svg`, `web/img/juegos/ajedrez_vs_ia.svg`.
	- Archivos eliminados: `web/Juegos/pollitos_cataplum.html`, `web/img/juegos/pollitos_cataplum.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el arcade publico deja de operar con el set anterior y pasa a un lobby de ocho juegos activos con records compartidos por navegador, popup uniforme `700x700` en escritorio y pausa real en todas las experiencias, incluyendo congelacion de IA u oponentes cuando aplica.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/menu_juegos.html` y en los ocho archivos `*_plus.html` del nuevo arcade.

- Home pÃƒÂºblico: botones superiores mÃƒÂ¡s compactos y centrados en mÃƒÂ³vil.
	- Archivos modificados: `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: los botones `Registrarse o iniciar sesiÃƒÂ³n` e `Informacion de contacto` del `index.html` ahora comparten un ancho mÃƒÂ¡s pequeÃƒÂ±o, menor altura visual y en celular se muestran centrados dentro del header.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/estilos.css`.

- Licencias super: autorreparaciÃƒÂ³n del esquema y validaciÃƒÂ³n real de guardado del valor.
	- Archivos modificados: `backend/db/db.go`, `backend/db/sql_compat.go`, `backend/db/licencias_schema_test.go`, `backend/main.go`, `web/super/licencias.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el backend ahora regulariza la tabla `licencias` tambien en PostgreSQL y reintenta `create/get/update` si faltan columnas como `valor`; la UI de super deja de ocultar errores HTTP al crear/editar licencias, mostrando el mensaje real cuando el backend rechaza la operaciÃƒÂ³n.
	- VerificaciÃƒÂ³n: `go test ./db -run "TestEnsureLicenciasSchemaAddsValorInmotor legado retirado|TestCreateAndUpdateLicenciaRepairMissingValorColumn" -count=1`.

- Seleccionar empresa: tarjetas adaptables con contenido interno completo y mÃƒÂ¡rgenes estrechos.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la vista `seleccionar_empresa.html` pasa a renderizar tarjetas con estructura interna avanzada (`empresa-card`) y estilos flexibles que permiten envolver tÃƒÂ­tulos, descripciones, estados y metadatos sin cortar contenido. Se mantienen mÃƒÂ¡rgenes pequeÃƒÂ±os y el interior se adapta automÃƒÂ¡ticamente al texto disponible.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js` y `web/estilos.css`.

- Super pagina principal: el editor ya no recorta tarjetas cargadas por usar el valor inicial del input de cantidad.
	- Archivos modificados: `web/super/pagina_principal.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se corrige el render del editor super de `pagina_principal` para que, al recargar configuraciones con mas de 5 tarjetas, use primero `state.config.cantidad` y no el valor HTML inicial del campo `ppCantidad`. Con esto vuelven a mostrarse las 7 tarjetas guardadas y la cantidad visible queda sincronizada con la API.
	- VerificaciÃƒÂ³n: inspeccion de `GET https://powerfulcontrolsystem.com/api/public/pagina_principal` con `cantidad=7`; diagnostico del editor sin errores en `web/super/pagina_principal.html`.

- Infraestructura publica: wildcard HTTPS manual para subdominios y subdominio dedicado de prueba para venta digital.
	- Archivos modificados: `documentos/manual_de_instalacion.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se documenta la emision manual del certificado wildcard `powerfulcontrolsystem.com` + `*.powerfulcontrolsystem.com`, la pauta de renovacion manual por DNS-01 y la publicacion del subdominio de prueba `venta-digital.powerfulcontrolsystem.com` hacia la pagina publica global `venta_digital.html`.
	- VerificaciÃƒÂ³n: HTTPS `200` en `https://powerfulcontrolsystem.com/`, `301` de `https://www.powerfulcontrolsystem.com/` a apex, `302` de `https://venta-digital.powerfulcontrolsystem.com/` a `/venta_digital.html` y `200` final en `https://venta-digital.powerfulcontrolsystem.com/venta_digital.html`.

- Registro administrativo: captura de pais y ciudad con deteccion inicial de pais en frontend.
	- Archivos modificados: `backend/db/db.go`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/account_handlers.go`, `backend/handlers/auth_admin_handlers_test.go`, `backend/db/administradores_auth_schema_test.go`, `web/registrar_nuevo_usuario_administrador.html`, `web/js/registrar_nuevo_usuario_administrador.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el registro de administradores ahora solicita correo, nombre completo, celular, pais y ciudad. El pais se sugiere automaticamente desde el navegador/zona horaria y sigue siendo editable. El backend persiste `pais` y `ciudad` en `administradores`, y se mantiene la exigencia de confirmar el correo antes de continuar al flujo de acceso que luego lleva a `seleccionar_empresa.html`.
	- VerificaciÃƒÂ³n: `go test ./db ./handlers -run 'Test(AdminRegisterHandlerCreatesPendingAdminAndCapturesConfirmationMail|AdminRegisterHandlerRejectsConfirmedExistingAdmin|EnsureAdministradoresAuthSchemaAddsMissingColumnsInmotor legado retirado|SetAdministradorPasswordRepairsMissingSecurityColumns)$' -count=1`.

- Autenticacion administrativa: compatibilidad del esquema `administradores` entre motor legado retirado y PostgreSQL.
	- Archivos creados: `backend/db/administradores_auth_schema_test.go`.
	- Archivos modificados: `backend/db/db.go`, `backend/main.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se centraliza la regularizacion de columnas de seguridad de `administradores` con soporte para motor legado retirado y PostgreSQL mediante `EnsureAdministradoresAuthSchema`, y `SetAdministradorPassword` reintenta la operacion cuando encuentra columnas faltantes. Con esto se corrige el flujo donde una cuenta autenticada por Google no podia registrar su primera contrasena local en VPS con PostgreSQL.
	- VerificaciÃƒÂ³n: `go test ./db ./handlers -run 'Test(EnsureAdministradoresAuthSchemaAddsMissingColumnsInmotor legado retirado|SetAdministradorPasswordRepairsMissingSecurityColumns|AccountSetGooglePasswordHandlerCreatesInitialPassword)$' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; validacion operativa en VPS de `systemd`, `Nginx`, `UFW`, callback OAuth y dominio publico.

- Super: modulo de seguridad VPS Linux con panel, CLI, cron y exportes multiformato.
	- Archivos creados: `backend/vpssecurity/config/config.go`, `backend/vpssecurity/config/default_vps_security_config.json`, `backend/vpssecurity/parser/lynis.go`, `backend/vpssecurity/parser/nmap.go`, `backend/vpssecurity/parser/trivy.go`, `backend/vpssecurity/scanner/runner.go`, `backend/vpssecurity/scanner/checks.go`, `backend/vpssecurity/reports/report.go`, `backend/vpssecurity/reports/report_test.go`, `backend/vpssecurity/logs/store.go`, `backend/vpssecurity/service.go`, `backend/handlers/security_vps_handlers.go`, `backend/handlers/security_vps_handlers_test.go`, `backend/tools/vps_security_scan/main.go`, `web/js/super_seguridad.js`, `scripts/install_vps_security_tools.sh`, `scripts/run_vps_security_scan.sh`, `scripts/install_vps_security_cron.sh`, `documentos/manual_vps_seguridad.md`.
	- Archivos modificados: `backend/main.go`, `web/super/seguridad.html`, `web/index.html`, `web/estilos.css`, `.gitignore`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega un modulo completo de seguridad VPS Linux para `super_administrador`, con ejecucion de Lynis/Nmap/Trivy y chequeos propios, historial/comparacion en filesystem, exportes `JSON/TXT/HTML/CSV/PDF/XLS`, CLI reutilizable y scripts Ubuntu para instalacion y cron. En el portal publico, `Informacion de contacto` queda anclado al extremo derecho de la misma fila superior del home.
	- VerificaciÃƒÂ³n: `go test ./vpssecurity/... ./handlers ./tools/vps_security_scan -run "TestSecurityVPS|TestGenerateArtifacts|TestCompareDetects" -count=1`; diagnÃƒÂ³stico del editor sin errores en los archivos Go/HTML/JS/SH modificados para este cambio.

- Login unificado: eliminado `recordar usuario/cuenta` y retirado `login_hint` en OAuth.
	- Archivos modificados: `web/login.html`, `web/js/login.js`, `web/login_usuario.html`, `web/js/login_usuario.js`, `web/menu.js`, `web/js/super_administrador.js`, `web/js/seleccionar_empresa.js`, `web/super/licencias.html`, `web/super/tipos_empresas.html`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se elimina la persistencia local de cuenta/usuario recordado en ambos logins y se deja `/auth/google/login` sin `login_hint`. Con esto, el acceso queda consistente entre `localhost:8080`, `powerfulcontrolsystem.com` y `www.powerfulcontrolsystem.com`, sin depender de estado guardado por dominio en `localStorage`.
	- VerificaciÃƒÂ³n: `go test -work ./handlers -run "TestHandleGoogleLoginRedirect|TestAccountSetGooglePasswordHandlerCreatesInitialPassword|TestE2E_AcceptContractCreatesSession|TestAdminLoginHandlerCreatesSessionForConfirmedAdmin" -count=1`.

- Login Google: registro obligatorio de contrasena local cuando falta password_set.
	- Archivos creados: `web/registrar_contrasena_usuario_de_google.html`, `web/js/registrar_contrasena_usuario_de_google.js`.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/account_handlers.go`, `backend/main.go`, `backend/handlers/auth_admin_handlers_test.go`, `backend/handlers/e2e_login_acceptance_test.go`, `web/ayuda/login_administradores.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el callback Google y la aceptaciÃƒÂ³n de contrato ya no envÃƒÂ­an al panel final cuando la cuenta administrativa aÃƒÂºn no tiene contraseÃƒÂ±a local; ahora redirigen a `registrar_contrasena_usuario_de_google.html`, que guarda la primera clave mediante `/api/account/set_google_password` para habilitar despuÃƒÂ©s el acceso por correo y contraseÃƒÂ±a.
	- VerificaciÃƒÂ³n: `go test -work ./handlers -run "Test(AccountSetGooglePasswordHandlerCreatesInitialPassword|E2E_AcceptContractCreatesSession|AdminLoginHandlerCreatesSessionForConfirmedAdmin)" -count=1`.

- Super: panel PostgreSQL con carga de tamaÃƒÂ±o por empresa.
	- Archivos modificados: `backend/handlers/postgres_performance.go`, `backend/handlers/postgres_performance_test.go`, `web/super/administrar_base_de_datos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el panel super de administracion PostgreSQL ahora puede cargar bajo demanda una tabla de consumo estimado por empresa dentro de `pcs_empresas`, ordenada de mayor a menor y mostrando tambien filas estimadas, tablas con datos y la tabla mas pesada por empresa.
	- VerificaciÃƒÂ³n: `go test -work ./handlers -run "TestPostgresPerformanceHandler|TestHumanizeBytesBinary" -count=1`.

- Manual de instalacion: agregado el paso de respuesta, confirmacion y formulario exacto de Epayco.
	- Archivos modificados: `documentos/manual_de_instalacion.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el manual ya incluye las URLs exactas que deben configurarse en Epayco para respuesta y confirmacion, ademas de los valores concretos del formulario de Epayco y una nota operativa sobre el flujo real de validacion del pago.

- Checkout y seleccion de empresa: ajuste visual solicitado.
	- Archivos modificados: `web/pagar_licencia.html`, `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `pagar_licencia.html` deja mas clara la pasarela activa cuando solo hay un metodo disponible y muestra el logo de Epayco en el selector y en el panel. `seleccionar_empresa.html` vuelve al estilo compacto anterior de tarjetas para empresas.

- Checkout de licencias: Epayco ahora usa una pagina publica fija de respuesta.
	- Archivos creados: `web/epayco/respuesta.html`.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el retorno de Epayco ya no depende de enviar al usuario directamente a `pagar_licencia.html`; ahora existe la landing publica fija `/epayco/respuesta.html`, que puedes registrar en el panel de Epayco y que reenvia al resumen del pago con el contexto necesario para validar y activar la licencia.
	- VerificaciÃƒÂ³n: `go test -work ./handlers -run "TestEpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|TestEpaycoCreateTransactionHandlerAllowsCheckoutWithoutPrivateKey|TestEpaycoTransactionStatusHandlerPreservesPendingOnGenericValidationError|TestResolvePaymentBaseURL" -count=1`.
- Login administrativo: registro separado, confirmaciÃƒÂ³n pÃƒÂºblica corregida y recuperaciÃƒÂ³n sin prompts.
	- Archivos creados: `web/registrar_nuevo_usuario_administrador.html`, `web/js/registrar_nuevo_usuario_administrador.js`, `backend/handlers/auth_admin_handlers_test.go`.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/utils/utils.go`, `web/login.html`, `web/js/login.js`, `web/estilos.css`, `web/ayuda/login_administradores.html`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el login administrativo ahora deja el registro en una pÃƒÂ¡gina pÃƒÂºblica especÃƒÂ­fica, elimina el campo de nombre incrustado del acceso principal, centra `Iniciar por correo` y agrega debajo `Registrarse` y `Ã‚Â¿OlvidÃƒÂ³ su contraseÃƒÂ±a?`. El backend valida `nombre`, `telefono` y contraseÃƒÂ±a segura, evita sobrescribir cuentas confirmadas, corrige el whitelist pÃƒÂºblico para `/auth/confirmar_admin` y sustituye la recuperaciÃƒÂ³n por formularios reales dentro de `login.html`.
	- VerificaciÃƒÂ³n: `go test -work ./handlers -run "Test(Admin|AuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI|HandleGoogleLogin|E2E_AcceptContractCreatesSession)" -count=1`; `go test -work ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

- Arcade publico: Patito volando ahora inicia con cuenta regresiva y los cinco juegos refuerzan su modo celular.
	- Archivos modificados: `web/Juegos/arcade_shared.js`, `web/Juegos/patito_volando.html`, `web/Juegos/pollitos_cataplum.html`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html`, `web/Juegos/rebote_bloques.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el arcade publico mantiene sonido compartido en los cinco juegos, `Patito volando` arranca con cuenta regresiva de 5 segundos y el resto del arcade ajusta shells, overlays y acciones para celular. Tambien se agregan sonidos de countdown en `arcade_shared.js` y `Serpiente pixel` suma feedback sonoro al giro durante la partida.
	- VerificaciÃƒÂ³n: validacion sin errores de los seis archivos del arcade modificados y revision de los nuevos breakpoints moviles y del countdown previo al inicio en `Patito volando`.

- Frontend compartido: mejoras base de adaptacion movil y menu flotante.
	- Archivos modificados: `web/menu.js`, `web/estilos.css`, `web/login.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el menu flotante ahora se cierra al seleccionar una opcion, el CTA de WhatsApp de la portada pasa a un icono compacto abajo a la derecha en movil para no tapar otros botones, la capa CSS compartida mejora tablas/sidebar/panel flotante en pantallas pequenas y `login.html` vuelve a cargar la hoja `estilos.css` correcta.
	- VerificaciÃƒÂ³n: validacion sin errores de `web/menu.js`, `web/estilos.css` y `web/login.html`, mas revision de los breakpoints moviles del menu flotante y del CTA de WhatsApp.

- Portal publico: botones superiores de la portada ahora usan el mismo estilo de Explorar oferta.
	- Archivos modificados: `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: los accesos `Registrarse o iniciar sesiÃƒÂ³n` e `Informacion de contacto` del encabezado de `index.html` reutilizan la misma apariencia visual del boton `Explorar oferta` de las tarjetas del home, sin cambiar rutas ni comportamiento responsive.
	- VerificaciÃƒÂ³n: revision del bloque compartido de selectores en `web/estilos.css` y del ajuste pill en `@media (max-width:560px)`.

- Checkout de licencias: Epayco sandbox estable con bootstrap PostgreSQL y polling pendiente consistente.
	- Archivos modificados: `backend/db/db.go`, `backend/main.go`, `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el backend asegura `pagos_epayco` y `pagos_wompi` al arrancar sobre PostgreSQL y deja de degradar a `ERROR` una referencia de Epayco que sigue `PENDING` localmente mientras la validacion externa responde un error transitorio. Ademas se normaliza la configuracion sandbox operativa (`epayco.*` y `gmail.confirm_base_url`) en la base super del VPS para que el checkout genere callbacks publicos validos.
	- VerificaciÃƒÂ³n: `go test ./handlers -run 'TestResolvePaymentBaseURL|TestEpayco(CreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|CreateTransactionHandlerAllowsCheckoutWithoutPrivateKey|TransactionStatusHandlerPreservesPendingOnGenericValidationError)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; validacion manual local de `GET /api/public/licencias/payment_methods`, `POST /epayco/create_transaction` y `GET /epayco/transaction_status` tras recompilar con `scripts/iniciar_servidor.ps1 -Background`.

- Portal publico: arcade con cinco juegos, tarjetas cuadradas y perfil compartido.
	- Archivos creados: `web/Juegos/arcade_shared.js`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html`, `web/Juegos/rebote_bloques.html`, `web/img/juegos/patito_volando.svg`, `web/img/juegos/pollitos_cataplum.svg`, `web/img/juegos/serpiente_pixel.svg`, `web/img/juegos/memoria_estelar.svg`, `web/img/juegos/rebote_bloques.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `web/Juegos/patito_volando.html`, `web/Juegos/pollitos_cataplum.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el lobby publico de Juegos se convierte en un arcade visual con portadas cuadradas, panel de jugador y resumen de records. `arcade_shared.js` centraliza nombre, top local y sonido; `Patito volando` y `Pollitos al cataplum` se integran a esa capa y se agregan tres juegos nuevos: `Serpiente pixel`, `Memoria estelar` y `Rebote de bloques`.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/Juegos/arcade_shared.js`, `web/Juegos/menu_juegos.html`, `web/Juegos/patito_volando.html`, `web/Juegos/pollitos_cataplum.html`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html` y `web/Juegos/rebote_bloques.html`.

## 2026-04-15
- Portal publico: nuevo juego `Pollitos al cataplum` y menu de Juegos multi-tarjeta.
	- Archivos creados: `web/Juegos/pollitos_cataplum.html`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega un segundo juego publico de resortera con niveles cortos, puntaje y control arrastrar/soltar; ademas, el catalogo de Juegos ahora soporta varias tarjetas con popup propio por juego.

- Licencias: Epayco/Wompi ya no fallan por resolver `localhost` al iniciar checkout.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `resolvePaymentBaseURL(...)` ahora ignora loopback en configuracion o request, intenta `gmail.confirm_base_url`, `Origin`/`Referer`, host publicado y, si hace falta, cae al dominio canonico `https://powerfulcontrolsystem.com` para construir callbacks publicos validos del checkout.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "Test(ResolvePaymentBaseURLFallsBackToCanonicalDomainOnLocalhost|ResolvePaymentBaseURLUsesConfiguredCanonicalDomain|ResolvePaymentBaseURLIgnoresConfiguredLocalhostAndFallsBackToCanonicalDomain|EpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|EpaycoCreateTransactionHandlerAllowsCheckoutWithoutPrivateKey)" -count=1`.

- Servidor: alerta de inicio/reinicio ahora puede activarse o desactivarse desde configuracion avanzada.
	- Archivos modificados: `backend/handlers/server_runtime_notifications.go`, `backend/handlers/server_runtime_notifications_test.go`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/handlers/super_config_backup_handlers.go`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el backend ya registraba el arranque/reinicio en `super_servidor_eventos`, en `backend/logs/server_reinicio.log` y enviaba correo cuando existia `gmail.restart_alert_to`. Ahora se agrega `gmail.restart_alert_enabled` para activar o desactivar ese correo desde `configuracion_avanzada.html` sin perder el destinatario configurado.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "Test(GmailConfigHandlerSaveRestartAlertTo|GmailConfigHandlerSaveRestartAlertToggle|RegisterServerStartupEventCapturesNotificationAndState|RegisterServerStartupEventDetectsUnexpectedRestart|RegisterServerStartupEventSkipsEmailWhenAlertsDisabled)" -count=1`.

- Seleccion de empresas: tarjetas con iconografia por tipo y rediseÃƒÂ±o mas profesional.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `seleccionar_empresa.html` ahora presenta cada empresa con icono segun `tipo_nombre`, tono visual por categoria, chips de estado y una tarjeta mas colorida/profesional. Se conserva el mismo flujo para abrir la administracion o continuar con la licencia.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js` y `web/estilos.css`.

- Pagina principal: la cantidad de tarjetas ahora se aplica y se guarda en un solo flujo.
	- Archivos modificados: `web/super/pagina_principal.html`, `backend/handlers/pagina_principal_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se elimina el paso manual `Aplicar cantidad` del editor super de `pagina_principal`. Al cambiar la cantidad, el editor reconstruye las tarjetas visibles y el mismo flujo de `Guardar configuracion` persiste cantidad, contenido y estilos. Ademas se agrega una prueba de persistencia para configuraciones ampliadas.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestPaginaPrincipal|TestPublicPaginaPrincipalHandlerExposesLandingFields" -count=1`.

- Portal publico: nuevo menu de juegos y primer juego `Patito volando`.
	- Archivos creados: `web/Juegos/menu_juegos.html`, `web/Juegos/patito_volando.html`.
	- Archivos modificados: `backend/utils/utils.go`, `backend/handlers/auth_users_carritos_test.go`, `web/menu.js`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega la entrada `Juegos` al menu flotante del portal, se crea `menu_juegos.html` con una tarjeta por juego publicado y se implementa `patito_volando.html` como minijuego de ventana pequena con control por barra espaciadora en PC y toque/presion en movil. `AuthMiddleware` deja publico `/Juegos/*` y la prueba del middleware se amplÃƒÂ­a para cubrir estas rutas.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestAuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI" -count=1`.

## 2026-04-15
- Repositorio: restaurado `Pendiente Notas` y auditados los borrados actuales.
	- Archivos modificados: `Pendiente Notas`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se recupera `Pendiente Notas` desde `HEAD` tras detectar que Git lo tenia como borrado local en el arbol de trabajo. La auditoria posterior confirma que no habia otros archivos eliminados en el estado actual del repo. Git no conserva una hora exacta para ese borrado local no confirmado; la ultima hora verificable en historial para el archivo es `2026-04-15 17:37:25 -0500` en el commit `e70884dabea1292d9c0e6d9b1a3f236e94d7c8c4`.
	- VerificaciÃƒÂ³n: `git diff --name-status --diff-filter=D`; `git status --short --untracked-files=no`; `git log -1 --format="%H%n%an%n%ad%n%s" -- "Pendiente Notas"`; `Get-Item -LiteralPath "d:\powerfulcontrolsystem\Pendiente Notas" | Select-Object FullName,Length,CreationTime,LastWriteTime`.

- Errores del sistema: monitor centralizado, recovery global y panel super.
	- Archivos creados: `backend/db/super_errores_sistema.go`, `backend/utils/system_errors.go`, `backend/handlers/super_error_handlers.go`, `backend/handlers/super_error_handlers_test.go`, `web/super/errores.html`.
	- Archivos modificados: `backend/main.go`, `backend/utils/utils.go`, `backend/utils/utils_test.go`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se implementa un sistema robusto de manejo de errores para todo el proyecto. Los errores HTTP y panicos recuperados se registran en `super_errores_sistema` y en `backend/logs/system_errors.log`, el cliente deja de recibir detalles tecnicos en respuestas `5xx` y super obtiene un panel profesional para monitoreo transversal por empresa, fecha, severidad y tipo.
	- VerificaciÃƒÂ³n: `go test ./utils -run "Test(JSONErrorMiddlewareSanitizesInternalServerError|RecoveryMiddlewareRecoversPanicAndLogsIt|JSONErrorMiddlewarePreservesJSONErrorBody|JSONErrorMiddlewareWrapsNonJSONError|JSONErrorMiddlewareWrapsEpaycoNonJSONError)" -count=1`; `go test ./handlers -run "Test(SuperErroresSistemaHandlerFiltersResults|SuperErroresSistemaHandlerMethodNotAllowed)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

- Contrato administrativo: ahora es versionado, editable desde super y exigido por version en el login.
	- Archivos creados: `backend/db/contrato_super.go`, `backend/handlers/super_contrato_handlers.go`, `backend/handlers/super_contrato_handlers_test.go`, `web/super/contrato.html`.
	- Archivos modificados: `backend/main.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/e2e_login_acceptance_test.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/utils/utils.go`, `web/accept.html`, `web/contrato.html`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el contrato que aceptan los administradores deja de ser un HTML estatico y pasa a vivir en la base `superadministrador`, con historial por version y resumen de cambio. Super puede editarlo desde una pagina dedicada, el portal lo publica via `/api/public/contrato` y el login administrativo exige aceptar la ultima version antes de crear sesion.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "Test(PublicContratoHandlerReturnsDefaultVersion|SuperContratoHandlerCreatesNewVersionAndHistory|E2E_AcceptContractCreatesSession|E2E_AcceptContractRequiresNewVersion|AuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en los archivos Go/HTML tocados.

- Deploy VPS: sync_to_vps ahora abre el dominio publico canonico en lugar de la IP.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `sync_to_vps.ps1` agrega `PublicBaseUrl` con valor por defecto `https://powerfulcontrolsystem.com/` y usa esa URL al finalizar el deploy, manteniendo `RemoteHost` solo para SSH y evitando abrir `http://<ip>:<puerto>/` en el navegador.
	- VerificaciÃƒÂ³n: validacion de sintaxis PowerShell mediante parser (`[System.Management.Automation.Language.Parser]::ParseFile(...)`) sin errores.

- Checkout de licencias: Epayco disponible con Public Key y rutas publicas de pago realmente abiertas.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/utils/utils.go`, `backend/utils/utils_test.go`, `web/pagar_licencia.html`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el backend deja de exigir `epayco.private_key` para mostrar Epayco en el checkout actual y usa `epayco.public_key` como requisito minimo operativo junto al flag `enabled`. Tambien se corrige `AuthMiddleware` para dejar publicas `/api/public/licencias/payment_methods`, `/wompi/*` y `/epayco/*`, y `web/pagar_licencia.html` ahora indica si la pasarela esta desactivada o si falta la `Public Key`.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "Test(EpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|EpaycoCreateTransactionHandlerAllowsCheckoutWithoutPrivateKey|PublicLicenciasPaymentMethodsHandlerOrdersAndAvailability|PublicLicenciasPaymentMethodsHandlerAllowsEpaycoWithPublicKeyOnly)" -count=1`; `go test ./utils -run "Test(AuthMiddlewareAllowsPublicLicenciaPaymentRoutesWithoutSession|JSONErrorMiddlewareWrapsEpaycoNonJSONError)" -count=1`; diagnostico del editor sin errores en los archivos Go/HTML tocados.

- Login admin y configuraciÃƒÂ³n Gmail: se simplifica el hint visual y se habilita ediciÃƒÂ³n directa.
	- Archivos modificados: `web/login.html`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el login administrativo ya no muestra el bloque `Se recordarÃƒÂ¡ ... / Olvidar`, aunque conserva la logica de `Recordar cuenta`, y la seccion Gmail SMTP del panel super deja de bloquear el correo remitente y los demas campos cuando ya existe una configuracion guardada.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestGmailConfigHandlerSaveRestartAlertTo" -count=1`; diagnostico del editor sin errores en los archivos HTML tocados.

- Portal publico: pagina_principal ahora define tamanos de tarjetas y texto para home y landing.
	- Archivos modificados: `backend/handlers/pagina_principal_handlers.go`, `backend/handlers/pagina_principal_handlers_test.go`, `web/super/pagina_principal.html`, `web/index.html`, `web/descripcion_de_los_sistemas.ht`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el editor super de pagina_principal agrega ajustes globales de tamano para tarjetas y texto del `index.html` y de `/descripcion_de_los_sistemas.ht`. La API publica mantiene un contrato unico (`tarjetas` + `estilos`) y el frontend aplica esos valores de forma responsive.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestPaginaPrincipal|TestPublicPaginaPrincipalHandlerExposesLandingFields" -count=1`; diagnostico del editor sin errores en los archivos Go/HTML/CSS tocados.

- Portal publico: CTA de WhatsApp arriba a la derecha y botones del header con estilo mini-tarjeta.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el home comercial reposiciona el CTA flotante `Contactenos` a la esquina superior derecha y convierte `Registrarse o iniciar sesiÃƒÂ³n` e `Informacion de contacto` en accesos compactos con acabado visual de mini-tarjeta, reutilizando el lenguaje de las tarjetas del portal sin alterar rutas publicas ni comportamiento funcional.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/index.html` y `web/estilos.css`.

- Portal publico: la landing descriptiva ahora se configura desde pagina_principal.
	- Archivos creados: `backend/handlers/pagina_principal_handlers_test.go`.
	- Archivos modificados: `backend/handlers/pagina_principal_handlers.go`, `web/super/pagina_principal.html`, `web/descripcion_de_los_sistemas.ht`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: la configuracion super de pagina_principal deja de limitarse al home y ahora tambien guarda la etiqueta, titular ampliado, parrafos y capacidades clave de cada tarjeta para `/descripcion_de_los_sistemas.ht`. La landing descriptiva deja de depender de textos estaticos por nombre de sistema y renderiza el contenido extendido desde la misma API publica usada por `index.html`.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "Test(PaginaPrincipal|AuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en `backend/handlers/pagina_principal_handlers.go`, `backend/handlers/pagina_principal_handlers_test.go`, `web/super/pagina_principal.html` y `web/descripcion_de_los_sistemas.ht`.

- Checkout de licencias: retorno recuperable tras volver de Epayco y Wompi.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el checkout de licencias ya no depende de un `status` estatico en la URL al volver desde la pasarela. Backend y frontend ahora conservan `provider`, `reference`, `transaction_id`, `licencia_id` y `empresa_id`, reanudan la verificacion real del pago desde `web/pagar_licencia.html` y permiten que Wompi consulte estado por `reference` cuando el navegador regresa sin `transaction_id` directo.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "Test(EpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|WompiTransactionStatusHandlerAllowsReferenceLookup|ResolvePaymentBaseURLRejectsLocalhostWithoutPublicConfig|ResolvePaymentBaseURLUsesConfiguredCanonicalDomain|PublicLicenciasPaymentMethodsHandlerOrdersAndAvailability)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en `web/pagar_licencia.html`, `backend/handlers/payments_handlers.go` y `backend/handlers/payments_handlers_test.go`.

- Checkout de licencias: fix Epayco con `public_key` real y callbacks sobre dominio pÃƒÂºblico.
	- Archivos creados: `backend/handlers/payments_handlers_test.go`.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/system_empresas_handlers_test.go`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se corrige el flujo de Epayco para separar `public_key`, `private_key` y `customer_id`, mantener compatibilidad con configuraciones legacy y resolver `response`/`confirmation` desde una base pÃƒÂºblica vÃƒÂ¡lida en vez de `localhost`. La pantalla de configuraciÃƒÂ³n avanzada deja de confundir la llave pÃƒÂºblica con el identificador del comercio y Wompi reutiliza la misma base pÃƒÂºblica para su `redirect_url`.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestResolvePaymentBaseURL|TestEpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|TestPublicLicenciasPaymentMethodsHandlerOrdersAndAvailability" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en los archivos tocados.

- Login Google: host canÃƒÂ³nico en dominio raÃƒÂ­z y estaciones con carga visible.
	- Archivos modificados: `backend/utils/utils.go`, `backend/utils/utils_test.go`, `backend/main.go`, `backend/.env.example`, `scripts/sync_to_vps.ps1`, `web/administrar_empresa/estaciones.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se corrige la inestabilidad del acceso administrativo tras registrar el dominio pÃƒÂºblico, redirigiendo `www.powerfulcontrolsystem.com` al host canÃƒÂ³nico `powerfulcontrolsystem.com` antes de procesar OAuth y alineando los defaults de `GOOGLE_REDIRECT_URL` al callback del dominio raÃƒÂ­z. AdemÃƒÂ¡s, la pÃƒÂ¡gina de estaciones ahora muestra `Cargando estaciones...` mientras consulta configuraciÃƒÂ³n, carritos y sensores, con mensaje visible en caso de error.
	- VerificaciÃƒÂ³n: `go test ./utils -run "Test(CanonicalPublicHostMiddleware|LoggingMiddlewareSetsContextAndWritesLogs|JSONErrorMiddlewareWrapsNonJSONError)" -count=1`; `go test ./handlers -run "TestHandleGoogleLogin" -count=1`; diagnÃƒÂ³stico del editor sin errores nuevos en los archivos tocados.

- Portal publico: home, landing descriptiva y contacto liberados sin sesion.
	- Archivos modificados: `backend/utils/utils.go`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: `AuthMiddleware` incorpora `/descripcion_de_los_sistemas.ht` y `/Informacion_de_contacto.html` al whitelist publico y mantiene `index.html` dentro del mismo conjunto, para que las tres paginas comerciales del portal sean accesibles sin login. La prueba de middleware se amplia para cubrir esas rutas junto con `menu.js` y `/api/public/pagina_principal`.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestAuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI" -count=1`; diagnostico del editor sin errores en Go y documentos modificados.

- Portal publico: contacto visible por WhatsApp y pagina dedicada de informacion.
	- Archivos creados: `web/Informacion_de_contacto.html`.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega un CTA flotante `Contactenos` en `index.html` que abre WhatsApp con el numero comercial del sistema, un acceso visible a `Informacion_de_contacto.html` desde el encabezado del portal y una nueva pagina publica con descripcion general del sistema, correo `powerfulcontrolsystem@hmail.com` y WhatsApp `3043306506`. Ademas, el acceso principal del header pasa a llamarse `Registrarse o iniciar sesiÃƒÂ³n` y queda junto al boton de contacto.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/index.html`, `web/Informacion_de_contacto.html` y `web/estilos.css`.

- Portal publico: landing descriptiva unica para todas las tarjetas del home.
	- Archivos creados: `web/descripcion_de_los_sistemas.ht`.
	- Archivos modificados: `web/index.html`, `web/super/pagina_principal.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el boton `Explorar oferta` del home deja de abrir enlaces directos y pasa a una sola landing publica (`/descripcion_de_los_sistemas.ht`) con anclas por tarjeta, descripciones ampliadas por seccion y un CTA `Probar Gratis` por cada solucion. El enlace configurado desde `super/pagina_principal.html` ahora alimenta ese CTA final.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en los archivos HTML/CSS modificados.

- Checkout de licencias: Epayco primero, Wompi debajo y activacion real desde configuracion avanzada.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/main.go`, `web/pagar_licencia.html`, `web/super/configuracion_avanzada.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega la ruta publica `GET /api/public/licencias/payment_methods` para publicar la disponibilidad ordenada de pasarelas de licencia, `web/pagar_licencia.html` ya muestra solo Epayco y Wompi con prioridad Epayco -> Wompi, y `web/super/configuracion_avanzada.html` permite activar/desactivar ambas pasarelas manteniendo a Wompi bloqueado en backend cuando esta desactivado o incompleto.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestPublicLicenciasPaymentMethodsHandlerOrdersAndAvailability|TestWompiConfigHandlerPersistsEnabledFlag|TestWompiTermsHandlerRejectsWhenDisabled" -count=1`; `go test ./ -run "^$" -count=1`; diagnostico del editor sin errores en HTML/CSS/Go tocados.

- Sync VPS: reparacion del redeploy remoto en fallback PuTTY.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el wrapper PowerShell deja de pasar inline a `plink` los bloques remotos complejos de `bootstrap` y `redeploy`; ahora los escribe en archivos temporales UTF-8 sin BOM y los ejecuta con `plink -m`, estabilizando el `heredoc` de la unidad `systemd` y evitando fallos Bash como `syntax error near unexpected token '('`. Tambien se endurece el quoting del binario remoto y de los directorios de logs.
	- VerificaciÃƒÂ³n: parser PowerShell en verde para `scripts/sync_to_vps.ps1` y diagnostico del editor sin errores nuevos en el archivo.

- Login Google: hardening de `login_hint` y saneamiento de cuenta recordada en escritorio.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`, `web/js/login.js`, `web/menu.js`, `web/js/super_administrador.js`, `web/js/seleccionar_empresa.js`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se evita que el login Google herede un `login_hint` corrupto desde el navegador. El backend solo reenvia hints con formato de correo valido y el frontend limpia/persiste `rememberedEmail` unicamente cuando el dato es plausible, estabilizando el flujo especialmente en escritorio.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestHandleGoogleLogin" -count=1` y `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

- Frontend web: refuerzo responsive transversal para portal y paneles administrativos.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se mejora la adaptacion entre escritorio, tablet y movil en la portada publica y en los layouts compartidos. El hero de `index.html` permite salto natural del titulo/subtitulo, el sidebar administrativo colapsa con mejor navegacion horizontal en movil y formularios/tablas/botones se reorganizan para pantallas estrechas.
	- VerificaciÃƒÂ³n: diagnostico del editor sin errores en `web/index.html` y `web/estilos.css`.

- VPS web: restauracion del dominio publico sin puerto con Nginx, UFW y TLS correctos.
	- Archivos modificados: `documentos/deploy_nginx_reverse_proxy_vps.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se corrige la incidencia de publicacion donde `powerfulcontrolsystem.com` dejaba de cargar externamente aunque el backend y Nginx estaban activos en el VPS; la causa fue `443/tcp` ausente en UFW y cobertura incompleta de `www` en TLS. Se abre `443/tcp`, se renueva el certificado LetsEncrypt para `powerfulcontrolsystem.com` y `www.powerfulcontrolsystem.com`, y se documenta la configuracion minima correcta para Nginx/Certbot.
	- VerificaciÃƒÂ³n: `curl -I https://powerfulcontrolsystem.com/` y `curl -I https://www.powerfulcontrolsystem.com/` responden `200 OK`; `certbot certificates` muestra ambos dominios; `ufw status` incluye `443/tcp ALLOW`.

- Sync VPS: limpieza automatica de procesos huerfanos antes del restart remoto.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se endurece el redeploy remoto para detener listeners fuera de `systemd` que sigan ocupando `SERVER_PORT`, registrar el PID/comando conflictivo y abortar con diagnostico si el puerto no se libera; adicionalmente se saneÃƒÂ³ el VPS donde un binario `server_linux_amd64 (deleted)` mantenÃƒÂ­a `:8080` ocupado y dejaba `powerfulcontrolsystem.service` en bucle de reinicio.
	- VerificaciÃƒÂ³n: en VPS `powerfulcontrolsystem.service` quedÃƒÂ³ `active (running)` tras limpiar el listener huÃƒÂ©rfano; `curl -k -I https://powerfulcontrolsystem.com/auth/google/login` sigue devolviendo `302` con `redirect_uri=https://powerfulcontrolsystem.com/auth/google/callback`.

- Sync VPS: bootstrap endurecido, mensajes accionables y preparaciÃƒÂ³n asistida del servidor.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se robustecen ambos scripts de despliegue para validar puertos/timeout antes de conectar, detectar el gestor de paquetes remoto e instalar dependencias base del VPS cuando hay privilegios, actualizar `SERVER_PORT` en cada bootstrap, exigir mensajes etiquetados `BOOTSTRAP_*`/`DEPLOY_*` y devolver hints claros cuando fallan DSN PostgreSQL, `CONFIG_ENC_KEY`, permisos `root/sudo` o el arranque del servicio `systemd`.
	- VerificaciÃƒÂ³n: parser de PowerShell en verde para `scripts/sync_to_vps.ps1`; diagnostico del editor sin errores para `scripts/sync_to_vps.sh`; previsualizacion `./scripts/sync_to_vps.ps1 -PreviewOnly -SkipBuild -OpenPublicUrlAfterDeploy:$false` generando correctamente las etapas remotas; la validacion directa `bash -n` sigue pendiente en este equipo porque `bash.exe` apunta al lanzador de WSL y no hay distribucion instalada.

- Login y menu: correccion de `recordar cuenta` y deteccion visible de sesion.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/e2e_login_acceptance_test.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/main.go`, `web/js/login.js`, `web/menu.js`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se deja de depender de la lectura cliente de `session_token` para sincronizar `recordar cuenta`, avatar y enlace de cierre de sesion; el backend emite `browser_session_active` como seÃƒÂ±al visible no sensible, manteniendo el token real en cookie `HttpOnly` y alineando tambien la limpieza de cookies en logout.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestE2E_AcceptContractCreatesSession|TestEmpresaUsuario(LoginHandlerSuccess|LoginHandlerRejectsWrongEmpresaScope|LoginHandlerRejectsWrongEmpresaScopeFromQuery|SetPasswordHandlerSuccess)|TestSuperEndpointsPermisosPorRol" -count=1` y `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

## 2026-04-14
- Sync VPS: backend persistente con `systemd` y autoarranque tras reinicio del servidor.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/deploy_nginx_reverse_proxy_vps.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: el despliegue al VPS deja de usar `nohup` para reiniciar el backend y pasa a instalar/actualizar una unidad `systemd` del proyecto con `Restart=always`, `systemctl enable`, carga de entorno desde `backend/.env.local` y logs persistentes en `backend/server.log` / `backend/server.err`, garantizando que el servicio vuelva solo tras caidas del proceso o reinicios del VPS y que solo se reinicie durante `sync_to_vps`.
	- VerificaciÃƒÂ³n: parser de PowerShell para `scripts/sync_to_vps.ps1`, previsualizacion local del script con `-PreviewOnly -SkipBuild` y diagnostico del editor sin errores para `scripts/sync_to_vps.sh`; la validacion directa con `bash -n` queda pendiente en este equipo porque no hay distro WSL ni Git Bash instalados.

## 2026-04-14
- Manual de instalacion: reposicion del documento y guia Google OAuth para VPS.
	- Archivos creados: `documentos/manual_de_instalacion.md`.
	- Archivos modificados: `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se repone el manual eliminado en `HEAD` y se actualiza con la configuracion exacta de Google Cloud Console para login local y produccion, incluyendo `Authorized redirect URIs` y `Authorized JavaScript origins` para `localhost` y `powerfulcontrolsystem.com`, mas notas de diagnostico para `redirect_uri_mismatch`.
	- VerificaciÃƒÂ³n: revision documental del manual recreado y comprobacion estatica de las URLs de callback/origen documentadas.

- Portal principal: tÃƒÂ­tulo en una sola lÃƒÂ­nea con subtÃƒÂ­tulo debajo en la misma columna.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega el contenedor `portal-intro-copy` para apilar verticalmente el encabezado del home, manteniendo `Sistema de FacturaciÃƒÂ³n ElectrÃƒÂ³nica` en una sola fila y moviendo `Toma el control de tu negocio con Powerful Control System` justo debajo, centrado en el mismo bloque visual.
	- VerificaciÃƒÂ³n: revision estatica de estructura HTML/CSS confirmando el nuevo contenedor y la regla `white-space: nowrap` aplicada al tÃƒÂ­tulo.

- Login administrativo Google: correccion para VPS y local + recordar cuenta estable.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/utils/utils.go`, `backend/handlers/auth_users_carritos_test.go`, `web/login.html`, `web/menu.js`, `web/js/login.js`, `web/index.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se corrige el flujo OAuth para adaptar `redirect_uri` al host real de la solicitud y forzar `https` en dominio publico (`powerfulcontrolsystem.com`), se habilitan rutas publicas que bloqueaban el login (`/js/login.js` y `/api/public/pagina_principal`), se evita consulta a `/me` sin sesion para eliminar ruido `401` en F12 y se completa la experiencia de `recordar cuenta`; adicionalmente se actualiza el encabezado del home a `Sistema de FacturaciÃƒÂ³n ElectrÃƒÂ³nica` con subtitulo operativo.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestHandleGoogleLogin|TestAuthMiddlewareAllowsPublicLoginAssetsAndHomeCardsAPI" -v -count=1` en verde; en VPS `GET /js/login.js` y `GET /api/public/pagina_principal` responden `200`; `GET /auth/google/login` emite `redirect_uri=https://powerfulcontrolsystem.com/auth/google/callback`; `google.redirect_url` en BD super quedÃƒÂ³ en HTTPS.

- Inicio local: diagnostico robusto para tunel SSH de PostgreSQL en VPS.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se mejora `Ensure-VpsPostgresTunnel` para esperar el listener con reintentos (hasta ~8s), capturar `stdout/stderr` de `plink` en `backend/tmp/plink_tunnel_<puerto>.*.log` y reportar causa detallada cuando el tunel no abre el puerto local; adicionalmente se corrige el argumento `-i` de `plink` para rutas de llave SSH con espacios (comillas explicitas), evitando el fallo `Host does not exist`.
	- VerificaciÃƒÂ³n: validacion de parseo PowerShell en verde con `[System.Management.Automation.Language.Parser]::ParseFile("scripts/iniciar_servidor.ps1", ...)` y ejecucion real `. "D:\powerfulcontrolsystem\scripts\iniciar_servidor.ps1" -Background` completando arranque con tunel activo y backend en `:8080`.

- Checkout de licencias: cierre operativo de Epayco.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/main.go`, `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se completa la implementacion de Epayco para licencias con `POST /epayco/create_transaction`, `GET /epayco/transaction_status` y `POST/GET /epayco/webhook`; se corrige la configuracion super de Epayco para aceptar credenciales reales sin validacion numerica de `cust_id`; y el frontend abre `checkout_url` de Epayco en una nueva pestaÃƒÂ±a manteniendo polling de estado y activacion automatica de licencia al aprobar.
	- VerificaciÃƒÂ³n: `go test ./ -run "^$" -count=1`, `go test ./handlers -run "^$" -count=1`, `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1` en verde.

- Chat y tareas: nuevo agente de citas con calendario grande y recordatorios previos.
	- Archivos modificados: `backend/db/chat_tareas.go`, `backend/handlers/chat_tareas.go`, `backend/handlers/chat_tareas_test.go`, `backend/main.go`, `web/administrar_empresa/chat_y_tareas.html`, `web/estilos.css`, `web/super/pagina_principal.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega agenda de citas empresarial en el modulo de chat/tareas (`/api/empresa/chat_tareas/citas`) con calendario mensual de gran formato, programacion/edicion de reuniones, visibilidad compartida por `empresa_id` y banner de recordatorios previos; adicionalmente se incluye un boton inferior de guardado en `web/super/pagina_principal.html`.
	- VerificaciÃƒÂ³n: `$env:DB_DIALECT='motor_legado_retirado'; go test ./handlers -run ChatTareas -count=1` y `$env:DB_DIALECT='motor_legado_retirado'; go test ./db -run ChatTareas -count=1`.

- UI administrativa: eliminacion de barra superior de titulo/acciones en todas las paginas de layout.
	- Archivos modificados: `web/super_administrador.html`, `web/administrar_empresa.html`, `web/administrar_empresa/finanzas_menu.html`, `web/administrar_empresa/facturacion_electronica_menu.html`, `web/administrar_empresa/administrar_productos_menu.html`, `web/administrar_empresa/configuracion_menu.html`, `web/administrar_empresa/reportes_menu.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se retira por completo el bloque visual `admin-toolbar page-header` del panel super y de los menus administrativos para eliminar la barra superior de la derecha en todas las vistas del layout.
	- VerificaciÃƒÂ³n: busqueda `class="admin-toolbar"` en `web/**/*.html` sin resultados.

- Inicio local: correccion de deteccion de procesos en puerto 8080 bajo StrictMode.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se normaliza la coleccion de PIDs detectados en el paso de liberacion de puerto para evitar `No se encuentra la propiedad 'Count'` cuando solo existe un proceso escuchando.
	- VerificaciÃƒÂ³n: ejecucion real `. 'D:\powerfulcontrolsystem\scripts\iniciar_servidor.ps1'` completando `3/8 Liberando puerto 8080` sin excepcion y arranque exitoso del backend en `:8080`.

- Inicio local: correccion de carga DSN PostgreSQL y tunel DB opcional en script de arranque.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivo local de entorno actualizado (no versionado): `backend/.env.local`.
	- DescripciÃƒÂ³n: el script ahora carga `DB_DIALECT`, `DB_EMPRESAS_DSN` y `DB_SUPERADMIN_DSN` desde `.env.local/.env` antes de validar prerequisitos; se anade soporte opcional para tunel SSH a PostgreSQL en VPS (`DB_VPS_TUNNEL_*`) con validacion temprana del puerto de tunel y ajuste temporal de DSN al listener local.
	- VerificaciÃƒÂ³n: `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\iniciar_servidor.ps1 -Background` en verde y `curl -I http://127.0.0.1:8080` con `HTTP/1.1 200 OK`.

- Venta publica por subdominio empresarial automatizado.
	- Archivos modificados: `backend/main.go`, `backend/handlers/venta_publica.go`, `backend/handlers/venta_publica_test.go`, `web/venta_publica.html`, `web/administrar_empresa/venta_publica.html`, `documentos/deploy_nginx_reverse_proxy_vps.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se habilita resoluciÃƒÂ³n de `empresa_slug` por subdominio (`{slug}.powerfulcontrolsystem.com`) en backend y frontend de venta publica, con soporte de apertura automatica de tienda desde la raiz del subdominio.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "VentaPublica|ResolveVentaPublicaSlugFromHost" -count=1` en verde.
	- Evidencia VPS: Nginx actualizado con bloque wildcard y captura de slug por host; validado `GET /` en host de subdominio con `302` a `/venta_publica.html?empresa_slug=<slug>` y `GET /venta_publica.html?empresa_slug=<slug>` con `200 OK`; queda pendiente registrar wildcard DNS `*.powerfulcontrolsystem.com` (resolucion publica actual `NXDOMAIN`).

## 2026-04-14
- Guia operativa de dominio con Nginx reverse proxy en VPS.
	- Archivo creado: `documentos/deploy_nginx_reverse_proxy_vps.md`.
	- Archivos modificados: `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `documentos/descripcion_del_proyecto`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se documenta el procedimiento para publicar `powerfulcontrolsystem.com` y `www.powerfulcontrolsystem.com` con Nginx en Ubuntu VPS, manteniendo el backend en `127.0.0.1:8080`, con validaciones de servicio/UFW y opcion de HTTPS con Certbot.
	- VerificaciÃƒÂ³n: guia con comandos en orden, listos para copia/pegado en consola remota.

## 2026-04-14
- Modulo de impresoras operativas por empresa.
	- Archivos creados: `backend/db/empresa_impresoras.go`, `backend/db/empresa_impresoras_test.go`, `backend/handlers/empresa_impresoras.go`.
	- Archivos modificados: `backend/main.go`, `web/administrar_empresa/configuracion.html`, `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/finanzas.html`, `web/administrar_empresa/reportes.html`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se aÃƒÂ±ade gestion de impresoras por `empresa_id` (predeterminada, estado activo/inactivo, asignacion por funcionalidad y por producto) y resoluciÃƒÂ³n de carrito/finanzas/reportes.
	- VerificaciÃƒÂ³n: `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1` en verde.

- Super administrador: nuevo panel de administracion de base de datos PostgreSQL.
	- Archivos creados: `backend/handlers/postgres_performance.go`, `backend/handlers/postgres_performance_test.go`, `web/super/administrar_base_de_datos.html`.
	- Archivos modificados: `backend/main.go`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se agrega un tablero profesional para monitoreo de PostgreSQL (salud del cluster, metricas por base, consultas activas prolongadas, `pg_stat_bgwriter` y recomendaciones automaticas), con endpoint protegido `/super/api/postgres/performance`.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "PostgresPerformance" -count=1` y `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1` en verde.

- Migracion cerrada a PostgreSQL-only y retiro de motor legado retirado operativo.
	- Archivos modificados: `backend/main.go`, `backend/db/sql_compat.go`, `scripts/iniciar_servidor.ps1`, `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `scripts/actualizar_repositorio.ps1`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_las_bases_De_datos`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivos eliminados: `backend/db/pcs_empresas`, `backend/db/pcs_superadministrador`.
	- DescripciÃƒÂ³n: el backend queda forzado a runtime PostgreSQL-only, sin fallback motor legado retirado en arranque; se limpian los `.db` legados del repositorio y se alinea la operacion local/remota a DSN PostgreSQL obligatorios.

- Estandarizacion documental ERP multiempresa.
	- Archivos creados: `documentos/erp_multiempresa/README.md`, `documentos/erp_multiempresa/01_alcance_erp_multiempresa.md`, `documentos/erp_multiempresa/02_diseno_tecnico_erp_multiempresa.md`, `documentos/erp_multiempresa/03_especificaciones_funcionales_erp_multiempresa.md`, `documentos/erp_multiempresa/04_guia_implementacion_erp_multiempresa.md`.
	- Archivos modificados: `documentos/README.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- DescripciÃƒÂ³n: se consolida un paquete ERP estandar listo para revision, con claridad de alcance, arquitectura, requisitos funcionales, reglas de negocio, integraciones y ruta de implementacion por fases.

- Documentacion: reorganizacion profesional, consolidacion de fuentes canonicas y limpieza de artefactos no usados.
	- Archivos modificados: `documentos/README.md`, `documentos/descripcion_del_proyecto`, `documentos/estructura_del_codigo`, `.gitignore`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- Archivos depurados: `documentos/historial_de_cambios_addendum_2026-04-04.md`, `backend/tmp/server.exe`, `backend/server.err`, `backend/server.log`, `backend/server.run.log`, `backend/logs/*.log`, `logs/test_runs/*.log`, `scripts/logs/*.log`, `tmp/doc_audit_report.txt`, `tmp/doc_hash_duplicates.txt`.
	- DescripciÃƒÂ³n: se centraliza la documentacion en un indice canonico, se evita duplicidad entre documentos estructurales y se eliminan archivos temporales/runtime que no deben versionarse.
	- VerificaciÃƒÂ³n: carpetas de logs temporales quedan limpias y se mantiene solo estado runtime necesario (`backend/logs/server_runtime_state.json`).

- OAuth Google VPS: validacion final de infraestructura HTTPS y diagnostico concluyente de `redirect_uri_mismatch`.
	- Archivos modificados: `CHANGELOG.md`, `documentos/historial_de_cambios`.
	- DescripciÃƒÂ³n: se verifica en VPS que el backend emite callback seguro `https://2.24.197.58.nip.io/auth/google/callback` y que el proxy TLS (Caddy) esta operativo en `:443`; Google sigue rechazando el flujo por URI no autorizada en el cliente OAuth.
	- VerificaciÃƒÂ³n: prueba E2E desde VPS confirma mismatch para la URI HTTPS publica y matriz de prueba muestra aceptacion solo de `http://localhost:8080/auth/google/callback`.
	- Pendiente externo: agregar la URI exacta del VPS en Google Cloud Console y repetir prueba de login.

- Inicio local: correcciÃƒÂ³n de detecciÃƒÂ³n de puerto 8080 para evitar falso bloqueo por PID 0.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`.
	- DescripciÃƒÂ³n: se reemplaza la detecciÃƒÂ³n basada en `netstat | findstr ":8080"` por una resoluciÃƒÂ³n de listeners locales reales (primero `Get-NetTCPConnection`, con fallback parseado de `netstat` en estado `LISTENING`). Se filtran PID invÃƒÂ¡lidos/no gestionables (`<= 0`) y se evita abortar cuando aparece `System Idle Process` sin listener real del backend.
	- VerificaciÃƒÂ³n: ejecuciÃƒÂ³n local `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\iniciar_servidor.ps1 -Background` completa en verde; paso 3 muestra `No hay procesos escuchando en el puerto 8080` y el servidor inicia correctamente.

- OAuth Google VPS: prioridad de entorno sobre DB + soporte de `GOOGLE_REDIRECT_URL` en despliegue.
	- Archivos modificados: `backend/main.go`, `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`.
	- DescripciÃƒÂ³n: se ajusta la carga OAuth para que `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` y `GOOGLE_REDIRECT_URL` del entorno tengan prioridad sobre valores almacenados en tabla `configuraciones` (la DB solo completa faltantes). Se aÃƒÂ±ade ademÃƒÂ¡s propagaciÃƒÂ³n de `GOOGLE_REDIRECT_URL` en bootstrap remoto de scripts de sincronizaciÃƒÂ³n.
	- VerificaciÃƒÂ³n: `go test ./handlers -run "TestHandleGoogleLoginRedirect" -count=1` y `go test ./ -count=1` en verde. DiagnÃƒÂ³stico en VPS confirma que el bloqueo actual es de polÃƒÂ­tica OAuth en Google (`secure-response-handling` / `redirect_uri_mismatch`) y no de base de datos.

- OAuth Google: correcciÃƒÂ³n de callback para evitar `localhost` en entorno VPS.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/main.go`.
	- DescripciÃƒÂ³n: se implementa resoluciÃƒÂ³n dinÃƒÂ¡mica del `redirect_uri` por host/protocolo de la solicitud y una regla de reescritura segura cuando la configuraciÃƒÂ³n existente apunta a loopback (`localhost/127.0.0.1`) pero el acceso real es pÃƒÂºblico (VPS). El callback reutiliza la URL efectiva mediante cookie tÃƒÂ©cnica de corta duraciÃƒÂ³n para mantener consistencia en intercambio de token.
	- VerificaciÃƒÂ³n: despliegue real a VPS con `DEPLOY_OK:pid=53618 port=8080`; validaciÃƒÂ³n HTTP de `/auth/google/login` devuelve `redirect_uri=http://2.24.197.58:8080/auth/google/callback`.

- Sync VPS: guard estricto de DSN para PostgreSQL y recuperaciÃƒÂ³n de despliegue estable.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/diagramas/estructura_del_codigo.md`, `CHANGELOG.md`, `documentos/historial_de_cambios`, `documentos/descripcion_de_archivos`.
	- DescripciÃƒÂ³n: el bootstrap remoto ahora conserva valores DB existentes, valida el modo efectivo y bloquea el despliegue con `BOOTSTRAP_ERROR:POSTGRES_MISSING_DSN` cuando `postgres` no tiene ambos DSN; ademÃƒÂ¡s usa el ÃƒÂºltimo valor de cada clave (`tail -n1`) y evita llegar a `DEPLOY_ERROR:process_not_running` por arranque invÃƒÂ¡lido. En paralelo se restableciÃƒÂ³ configuraciÃƒÂ³n DSN operativa en VPS para retomar despliegues en modo PostgreSQL.
	- VerificaciÃƒÂ³n: ejecuciÃƒÂ³n real `./scripts/sync_to_vps.ps1 -SkipBuild -RetryCount 1` primero falla en bootstrap con mensaje explÃƒÂ­cito de DSN faltantes, luego (tras restablecer DSN en VPS) finaliza con `DEPLOY_OK:pid=... port=8080` y `GET /` = `200`.

- VPS web root: correcciÃƒÂ³n de resoluciÃƒÂ³n de estÃƒÂ¡ticos para index/login.
	- Archivos modificados: `backend/main.go`, `scripts/sync_to_vps.ps1`, `CHANGELOG.md`, `documentos/historial_de_cambios`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`.
	- DescripciÃƒÂ³n: se ajusta `resolveWebDir()` para priorizar correctamente `.../web` cuando el binario corre desde `backend/bin`, evitando que el servidor sirva `backend/web/uploads/` como raÃƒÂ­z. Se redepliega en VPS y se valida apertura automÃƒÂ¡tica de la URL pÃƒÂºblica.
	- VerificaciÃƒÂ³n: `GET /` = `200` con HTML de portal, `GET /index.html` = `200`, `GET /login.html` = `200`, proceso remoto activo en `:8080` y runtime PostgreSQL operativo.

- Sync VPS: hardening para preservar DSN remotos y apertura automÃƒÂ¡tica de URL pÃƒÂºblica.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`.
	- DescripciÃƒÂ³n: se excluye `backend/.env.local` de la sincronizaciÃƒÂ³n para evitar sobrescribir secretos/DSN del VPS, se robustece el healthcheck de redeploy (detecta proceso caÃƒÂ­do y valida respuesta HTTP distinta de `000`) y se aÃƒÂ±ade apertura automÃƒÂ¡tica de `http://<host>:<puerto>/` al finalizar despliegues exitosos.
	- VerificaciÃƒÂ³n: ejecuciÃƒÂ³n real `./scripts/sync_to_vps.ps1 -RemoteHost 2.24.197.58 -RemoteUser root -RemotePath /root/powerfulcontrolsystem -DbDialect postgres -DbEmpresasDsn ... -DbSuperadminDsn ...` con `DEPLOY_OK:pid=... port=8080`, `GET / => 200` y backend en modo PostgreSQL con DSN activos.

- MigraciÃƒÂ³n PostgreSQL (fase 4): estabilizaciÃƒÂ³n de salida operativa en contabilidad y runtime VPS.
	- Archivos modificados: `backend/db/eventos_contables.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `Pendiente Notas`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- DescripciÃƒÂ³n: se corrige el flujo del worker de asientos/eventos para PostgreSQL usando wrappers SQL portables y retorno de `id` compatible, eliminando el error `syntax error at or near "ORDER"` en runtime. Se restaura ademÃƒÂ¡s el entorno VPS con DSN PostgreSQL vÃƒÂ¡lidos en `backend/.env.local` y se valida arranque estable.
	- VerificaciÃƒÂ³n: `go test ./ ./auth ./db ./handlers ./metrics ./utils` en verde; validaciÃƒÂ³n remota en VPS con proceso activo, sin errores recientes de `asientos_worker` y healthcheck `HTTP=200`.

- MigraciÃƒÂ³n PostgreSQL (fase 3): cierre documental del plan y sincronizaciÃƒÂ³n de gobernanza por mÃƒÂ³dulos.
	- Archivos modificados: `Pendiente Notas`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`.
	- DescripciÃƒÂ³n: se marca Fase 3 como completada en el plan operativo, se agrega evidencia tÃƒÂ©cnica de conmutaciÃƒÂ³n a PostgreSQL y se alinea la documentaciÃƒÂ³n de mÃƒÂ³dulos/permisos sin cambios de privilegios en la matriz CRUD/A.
	- VerificaciÃƒÂ³n: se mantiene evidencia de pruebas del bloque core en verde (`go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1`).

- MigraciÃƒÂ³n PostgreSQL (fase 3): conmutaciÃƒÂ³n de runtime backend a motor PostgreSQL en VPS.
	- Archivos modificados: `backend/main.go`, `backend/go.mod`, `backend/go.sum`, `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/diagramas/estructura_del_codigo.md`.
	- DescripciÃƒÂ³n: el backend ahora selecciona motor por entorno (`DB_DIALECT`), abre conexiones con `pgx` usando `DB_EMPRESAS_DSN` y `DB_SUPERADMIN_DSN`, y omite el bootstrap motor legado retirado cuando el runtime es PostgreSQL. Los scripts de sincronizaciÃƒÂ³n ahora propagan y verifican estas variables en `backend/.env.local` del VPS durante bootstrap remoto.
	- VerificaciÃƒÂ³n: `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1` en verde.

- MigraciÃƒÂ³n PostgreSQL (fase 3): compatibilidad ampliada en nÃƒÂºcleo `backend/db`.
	- Archivos modificados: `backend/db/sql_compat.go`, `backend/db/empresa_scope.go`, `backend/db/productos.go`, `backend/db/db.go`.
	- DescripciÃƒÂ³n: se amplÃƒÂ­a la capa de compatibilidad motor legado retirado/PostgreSQL con wrappers `query/exec` portables, inserciones con `RETURNING id` para PostgreSQL, detecciÃƒÂ³n de tablas por `information_schema` y ajuste de `ensureColumnIfMissing` por dialecto con normalizaciÃƒÂ³n de defaults de fecha. AdemÃƒÂ¡s, se migra el bloque core de `db.go` (licencias, tipos de empresa, empresas, Wompi, asesores, configuraciones y mÃƒÂ©tricas) para usar placeholders/fechas compatibles con ambos motores.
	- VerificaciÃƒÂ³n: `go test ./db -run "Session|Admin|User|Licencia|TipoEmpresa|Empresa|Config|Metric|Wompi|Asesor" -count=1` y `go test ./handlers -run "TestHandleGoogleLoginRedirectIncludesLoginHint|TestE2E_AcceptContractCreatesSession" -count=1` en verde.

- Sync VPS: selecciÃƒÂ³n automÃƒÂ¡tica de clave de identidad al no pasar `-IdentityFile`.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/README_sync.md`.
	- DescripciÃƒÂ³n: cuando no se especifica `-IdentityFile`, el script ahora prioriza la clave del proyecto `clave privada ssh.ppk` y, si no existe, usa `~/.ssh/id_rsa`. AdemÃƒÂ¡s, mejora el mensaje de error cuando el VPS rechaza autenticaciÃƒÂ³n.
	- VerificaciÃƒÂ³n: ejecuciÃƒÂ³n real `./scripts/sync_to_vps.ps1 -SkipBuild -RemoteHost 2.24.197.58 -RemoteUser root -RemotePath /root/powerfulcontrolsystem -RetryCount 1` completada con `SincronizaciÃƒÂ³n completada por fallback sin WSL (PuTTY)`.

- Sync VPS: redeploy remoto automÃƒÂ¡tico de backend tras sincronizaciÃƒÂ³n.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`.
	- DescripciÃƒÂ³n: la sincronizaciÃƒÂ³n ahora detiene el proceso viejo del backend en VPS, inicia la nueva versiÃƒÂ³n del binario y valida salud HTTP en el puerto configurado (`SERVER_PORT`), evitando que quede corriendo una versiÃƒÂ³n antigua.
	- VerificaciÃƒÂ³n: ejecuciÃƒÂ³n real `./scripts/sync_to_vps.ps1 -SkipBuild -RemoteHost 2.24.197.58 -RemoteUser root -RemotePath /root/powerfulcontrolsystem -RetryCount 1` con salida `DEPLOY_OK:pid=... port=8080`.

- MigraciÃƒÂ³n PostgreSQL (fase 3): avance inicial en autenticaciÃƒÂ³n y sesiones.
	- Archivos aÃƒÂ±adidos/modificados: `backend/db/sql_compat.go`, `backend/db/db.go`, `documentos/diagramas/estructura_del_codigo.md`.
	- DescripciÃƒÂ³n: se incorpora capa de compatibilidad SQL motor legado retirado/PostgreSQL (rebindeo de placeholders y expresiones de fecha) y se aplica a funciones crÃƒÂ­ticas del flujo de autenticaciÃƒÂ³n/sesiones (`UpsertUser`, `UpsertAdministrador`, `CreateSession`, `RevokeSessionByToken`, `GetSessionByToken`, `GetAdminByEmail`).
	- VerificaciÃƒÂ³n: `go test ./db -run "Session|Admin|User|Licencia" -count=1` y `go test ./handlers -run "TestHandleGoogleLoginRedirectIncludesLoginHint|TestE2E_AcceptContractCreatesSession" -count=1` en verde.

## 2026-04-13
 - ReparaciÃƒÂ³n de login de usuario empresarial: permitir entrada manual de `empresa_id` y persistencia de contexto.
	- Archivos modificados:
		- web/login_usuario.html
		- web/js/login_usuario.js
	- DescripciÃƒÂ³n: se agrega un campo `Empresa ID` en la pÃƒÂ¡gina de login de usuario de empresa para aceptar el parÃƒÂ¡metro cuando no viene en la URL. La lÃƒÂ³gica JS persiste `empresa_id` en session/local storage, asegura que `redirect_url` incluya `empresa_id` y mejora la funcionalidad de "recordar usuario" por empresa.
	- VerificaciÃƒÂ³n: validaciÃƒÂ³n de sintaxis JS sin errores y flujo de login manual verificado localmente.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/README_sync.md`.
	- DescripciÃƒÂ³n: en fallback sin WSL, el script ahora selecciona transporte por tipo de clave: `ssh.exe` + `scp.exe` para claves OpenSSH (ej. `id_rsa`) y `plink.exe` + `pscp.exe` para `.ppk`. Con esto se evita el error `Unable to use key file ... OpenSSH SSH-2 private key (new format)` al usar la identidad por defecto.
	- VerificaciÃƒÂ³n: `.\scripts\sync_to_vps.ps1 -SkipBuild -PreviewOnly -IdentityFile "$env:USERPROFILE\.ssh\id_rsa"` muestra `Fallback sin WSL (OpenSSH)` y comandos con `ssh.exe`/`scp.exe`.

- MigraciÃƒÂ³n de datos a PostgreSQL en VPS: instalaciÃƒÂ³n, ejecuciÃƒÂ³n por etapas y validaciÃƒÂ³n inicial.
	- Archivos modificados: `Pendiente Notas`, `documentos/regla_agente_go.md`, `copilot-instructions.md`, `documentos/descripcion_de_las_bases_De_datos`, `documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- DescripciÃƒÂ³n: se instala PostgreSQL en VPS por SSH, se crean las bases `pcs_superadministrador` y `pcs_empresas`, y se inicia la migraciÃƒÂ³n desde motor legado retirado con `pgloader` en dos etapas (superadministrador y empresas), validando consistencia por conteo de tablas en cada base. Se formaliza ademÃƒÂ¡s la regla operativa: base productiva en VPS con PostgreSQL y motor legado retirado local como legado de migraciÃƒÂ³n/contingencia.
	- VerificaciÃƒÂ³n: `VALIDACION_SUPER_OK` y `VALIDACION_EMPRESAS_OK` tras comparaciÃƒÂ³n motor legado retirado vs PostgreSQL por tabla.

- Login administrativo: eliminaciÃƒÂ³n del mensaje visual de cuenta recordada y ajuste de OAuth.
	- Archivos modificados: `web/login.html`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`.
	- DescripciÃƒÂ³n: se elimina el texto en pantalla `Cuenta recordada ...` del login admin y se ajusta el parÃƒÂ¡metro OAuth `prompt` a `select_account` para evitar re-consentimiento de Google en cada inicio.
	- VerificaciÃƒÂ³n: `go test ./handlers -run TestHandleGoogleLoginRedirectIncludesLoginHint -count=1` en verde.

- Login administrativo: correcciÃƒÂ³n de "Recordar cuenta" para evitar sesiÃƒÂ³n parcial.
	- Archivos modificados: `web/js/login.js`, `web/menu.js`.
	- DescripciÃƒÂ³n: se corrige el flujo para que cerrar sesiÃƒÂ³n no elimine la preferencia cuando `rememberAccount=1`, se mantiene el correo recordado hasta que el usuario pulse "Olvidar" y se agrega sincronizaciÃƒÂ³n de `rememberedEmail` desde `/me` cuando existe sesiÃƒÂ³n activa.
	- VerificaciÃƒÂ³n: revisiÃƒÂ³n de errores en frontend sin incidencias (`get_errors` en ambos archivos).

- Inicio local: hardening de scripts/iniciar_servidor para evitar caÃƒÂ­das del host de PowerShell/VS Code.
	- Archivo modificado: `scripts/iniciar_servidor.ps1`.
	- DescripciÃƒÂ³n: se refuerza la liberaciÃƒÂ³n de puerto 8080 para terminar ÃƒÂºnicamente procesos del backend (`server.exe`, `pos-backend`, `go run` del proyecto) y no procesos ajenos. Cuando el puerto estÃƒÂ¡ ocupado por un proceso no gestionado, el script ahora informa el PID/nombre y aborta con mensaje claro en lugar de forzar `taskkill` indiscriminado. TambiÃƒÂ©n se elimina el `Clear-Host` inicial para evitar efectos colaterales en consolas integradas.
	- VerificaciÃƒÂ³n: ejecuciÃƒÂ³n local `./scripts/iniciar_servidor.ps1 -Background` con `SCRIPT_EXIT=0` y comprobaciÃƒÂ³n HTTP local `HTTP_STATUS=200`.

- UnificaciÃƒÂ³n de bases motor legado retirado: solo dos archivos canÃƒÂ³nicos del sistema.
	- Archivos modificados: `backend/main.go`, `documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- DescripciÃƒÂ³n: se normaliza la resoluciÃƒÂ³n de rutas en runtime para que el backend use por defecto `backend/db/pcs_empresas` y `backend/db/pcs_superadministrador` aunque se ejecute desde otro directorio. Se depuran copias operativas duplicadas en raÃƒÂ­z y en `backend/`, dejando ÃƒÂºnicamente dos archivos `.db` activos.
	- VerificaciÃƒÂ³n: inventario local posterior muestra exactamente dos DB (`backend/db/pcs_empresas` y `backend/db/pcs_superadministrador`) y pruebas backend en verde con `go test ./ ./auth ./db ./handlers ./metrics ./utils`.

- Sync VPS: bootstrap automÃƒÂ¡tico para servidor nuevo y diagnÃƒÂ³stico de OAuth.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/README_sync.md`.
	- DescripciÃƒÂ³n: se aÃƒÂ±ade bootstrap post-sync en modo sin WSL para instalar dependencias base (`ca-certificates`, `curl`, `motor_legado_retirado`), asegurar `backend/.env.local` y reportar estado de variables crÃƒÂ­ticas (`GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `SERVER_PORT`, `CONFIG_ENC_KEY`) con salida `BOOTSTRAP_WARN/BOOTSTRAP_OK`. Se incorporan parÃƒÂ¡metros opcionales `-GoogleClientId` y `-GoogleClientSecret`.
	- VerificaciÃƒÂ³n: ejecuciÃƒÂ³n real con `SYNC_EXIT=0` y diagnÃƒÂ³stico remoto mostrando faltantes OAuth (`GOOGLE_CLIENT_ID/SECRET` vacÃƒÂ­os).

- Instalador de clave pÃƒÂºblica en Windows: correcciÃƒÂ³n de errores de ejecuciÃƒÂ³n.
	- Archivo modificado: `scripts/instalar_clave_publica_vps.ps1`.
	- DescripciÃƒÂ³n: se corrige el flujo para evitar errores remotos tipo `invalid option namepefail` y se adapta a PowerShell 5.1 eliminando sintaxis no soportada (`??`). Ahora usa comando remoto en una sola lÃƒÂ­nea, validaciÃƒÂ³n de formato de clave OpenSSH y reintentos por timeout.
	- VerificaciÃƒÂ³n: `scripts/instalar_clave_publica_vps.ps1 -PreviewOnly -RemoteHost 2.24.197.58 -User root -Port 22` en verde (exit 0).

- Despliegue VPS: instalaciÃƒÂ³n automatizada de clave pÃƒÂºblica PuTTYgen + robustecimiento de scripts de sincronizaciÃƒÂ³n
	- Archivos aÃƒÂ±adidos/modificados:
		- scripts/instalar_clave_publica_vps.ps1 (nuevo: instala clave pÃƒÂºblica RFC4716 en `~/.ssh/authorized_keys` de VPS Linux)
		- scripts/sync_to_vps.sh (hardening para Linux: validaciones, eliminaciÃƒÂ³n de `eval`, chequeo remoto de SO)
		- scripts/sync_to_vps.ps1 (manejo de errores sin cerrar terminal de VS Code, build Linux local previo y fallback PuTTY sin WSL con empaquetado tar)
		- scripts/README_sync.md (guÃƒÂ­a de ejecuciÃƒÂ³n en un comando)
		- web/login.html y web/js/login.js (completa UX de "Recordar cuenta" para login admin)
	- DescripciÃƒÂ³n: se habilita un flujo operativo de un solo comando para preparar acceso por clave pÃƒÂºblica al VPS y se corrige la causa de cierres de terminal por `exit` en script PowerShell. `sync_to_vps.ps1` ahora compila en local un binario Linux (`backend/bin/server_linux_amd64`) antes de sincronizar y, sin Ubuntu/WSL, opera empaquetando el proyecto en `.tar`, subiÃƒÂ©ndolo por `pscp.exe` y extrayÃƒÂ©ndolo en VPS por `plink.exe`, con trazas detalladas y exclusiÃƒÂ³n de archivos sensibles/locales (`*.ppk`, `*.pem`, `*.key`, DB, logs, temporales); ademÃƒÂ¡s aplica `chmod +x` al binario remoto configurado. Se aÃƒÂ±adiÃƒÂ³ manejo de `Connection timed out` con prechequeo TCP y reintentos automÃƒÂ¡ticos configurables (`-RetryCount`) por etapa de conexiÃƒÂ³n/subida/extracciÃƒÂ³n.
	- VerificaciÃƒÂ³n: `scripts/instalar_clave_publica_vps.ps1 -PreviewOnly -RemoteHost 2.24.197.58` ejecuta correctamente; `scripts/sync_to_vps.ps1 -BuildOnly`, `-DryRun` y ejecuciÃƒÂ³n real con `-IdentityFile "D:\powerfulcontrolsystem\clave privada ssh.ppk"` completan con `exit code 0`; en VPS el artefacto quedÃƒÂ³ como ELF Linux en `/root/powerfulcontrolsystem/backend/bin/server_linux_amd64`.

- MÃƒÂ³dulo Vendedores / Asesores comerciales: integraciÃƒÂ³n de cÃƒÂ³digo de descuento y registro de asesor/vendedor en pagos
	- Archivos aÃƒÂ±adidos/modificados:
		- backend/handlers/payments_handlers.go (extiende payload y persistencia de `pagos_wompi` con `discount_code` y `asesor_id`/`vendedor_id`)
		- backend/db/db.go (helpers para `asesores`, `asesor_comercial` y `asesor_comisiones`, y claves de configuraciÃƒÂ³n `vendedor.*`)
		- backend/handlers/vendedores_handlers.go (nuevo: CRUD de asesores / vendedores)
		- backend/handlers/vendedor_config_handlers.go (nuevo: GET/PUT /super/api/vendedor_config)
		- backend/main.go (migraciones: tablas `asesores`, `asesor_comercial`, `asesor_comisiones`; registro de rutas `/super/api/vendedores`, `/super/api/asesor_comercial`, `/super/api/vendedor_config`)
		- backend/tools/insert_asesor.go, backend/tools/insert_plan.go, backend/tools/insert_licencia.go, backend/tools/create_session.go, backend/tools/query_pagos_comisiones.go (herramientas para pruebas locales)
		- web/pagar_licencia.html (nuevo campo `discount_code` y `asesor_id`/`vendedor_id` en el formulario de pago)
		- web/super/activar_asesor.html, web/super/asesor_comercial.html, web/super/vendedor_config_avanzado.html (UI super-administrador para activar vendedores, configurar planes y ajustes globales)
		- documentos/estructura_bd.md (documenta las nuevas tablas y columnas de pagos/comisiones)
		- documentos/descripcion_de_archivos (registro de los nuevos archivos del mÃƒÂ³dulo)
	- DescripciÃƒÂ³n: Se aÃƒÂ±ade soporte opcional para incluir un cÃƒÂ³digo de descuento y una referencia al asesor/vendedor en el pago de licencias. Se introduce la entidad de `asesores` (vendedores), planes comerciales (`asesor_comercial`) y el registro de comisiones (`asesor_comisiones`) que crea una comisiÃƒÂ³n inmediata y entradas programadas por meses de renovaciÃƒÂ³n segÃƒÂºn el plan.
	- VerificaciÃƒÂ³n: Prueba manual de activaciÃƒÂ³n sin pago (`/licencias/activar_sin_pago`) usando sesiÃƒÂ³n administrativa de prueba; se confirmÃƒÂ³ la creaciÃƒÂ³n de una fila en `pagos_wompi` con `discount_code` y `asesor_id` y la creaciÃƒÂ³n de registros en `asesor_comisiones` (comisiÃƒÂ³n inmediata + programadas). Tests automatizados pendientes.

- Estaciones: fix de persistencia de `estaciones_config` cuando el frontend no envÃƒÂ­a `estado`.
	- Archivos modificados: `backend/db/empresa_estacion_prefs.go`, `backend/db/empresa_estacion_prefs_test.go`, `backend/handlers/empresa_estacion_prefs_test.go`.
	- DescripciÃƒÂ³n: se normaliza `estado` vacio como `activo` en upsert/list/get de preferencias por estacion, evitando que las estaciones desaparezcan despues de guardarse.
	- VerificaciÃƒÂ³n: pruebas en verde de estaciones, sensores, ventas y facturacion documental.

- Estaciones: correccion de flujo 10+, colores movidos a configuracion de estaciones y hardening sensor/carrito.
	- Archivos modificados: `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/configuracion.html`, `web/administrar_empresa/estaciones.html`, `backend/handlers/empresa_estacion_prefs_test.go`.
	- DescripciÃƒÂ³n: se consolida la gestion de colores de estado de carrito en la configuracion de estaciones, se fortalece el parseo de `estaciones_config` para tolerar payloads legacy anidados, se mejora la sincronizacion de carritos por estacion ante colisiones idempotentes y se valida el rango de estacion en configuracion de sensores.
	- VerificaciÃƒÂ³n: pruebas dirigidas en verde para handlers y DB en estaciones/sensores/carritos/facturaciÃƒÂ³n, incluyendo `TestEmpresaEstacionPrefsHandler_UpsertAndIsolationByEmpresa`.

- ReparaciÃƒÂ³n integral de acceso empresarial y estaciones.
	- Archivos modificados: `web/login_usuario.html`, `web/js/login_usuario.js`, `web/js/seleccionar_empresa.js`, `web/administrar_empresa/configuracion_de_estaciones.html`.
	- DescripciÃƒÂ³n: se corrige la continuidad del flujo `login usuario empresa -> seleccionar empresa -> administrar empresa` con persistencia de `empresa_id` y opciÃƒÂ³n de recordar correo. La pÃƒÂ¡gina de configuraciÃƒÂ³n de estaciones se reconstruye y soporta generaciÃƒÂ³n/sincronizaciÃƒÂ³n masiva de estaciones (incluyendo 10+) con manejo tolerante de conflictos idempotentes al cerrar/inactivar carritos.
	- VerificaciÃƒÂ³n: pruebas backend de paquetes principales en verde (`go test ./ ./auth ./db ./handlers ./metrics ./utils`).

## 2026-04-12
- Flujo final de login administrativo: cuenta Google correcta + aceptaciÃƒÂ³n ÃƒÂºnica de contrato + reCAPTCHA real.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/e2e_login_acceptance_test.go`, `backend/handlers/auth_users_carritos_test.go`, `web/login.html`, `web/js/login.js`, `web/accept.html`, `web/menu.js`, `web/estilos.css`.
	- DescripciÃƒÂ³n: se unificÃƒÂ³ el flujo en `login.html -> OAuth -> /accept.html -> /accept/complete -> panel`, usando `administradores.acepta_contrato` como fuente canonica de aceptaciÃƒÂ³n (sin depender de cookie global), validaciÃƒÂ³n server-side de reCAPTCHA y prompt OAuth `select_account consent` para evitar reutilizaciÃƒÂ³n silenciosa de cuenta incorrecta.
	- VerificaciÃƒÂ³n: pruebas dirigidas en verde (`TestE2E_AcceptContractCreatesSession` y `TestHandleGoogleLoginRedirectIncludesLoginHint`).

- MÃƒÂ³dulo sensor de puertas (Raspberry Pi): backend, handlers, UI y tests.
	- Archivos agregados/modificados:
		- backend/db/sensor_puertas.go (nuevo mÃƒÂ³dulo DB: dispositivos y heartbeats)
		- backend/handlers/sensor_puertas.go (handlers: endpoint pÃƒÂºblico `action=heartbeat` y configuraciÃƒÂ³n protegida)
		- backend/db/sensor_puertas_test.go (pruebas unitarias DB)
		- backend/handlers/sensor_puertas_test.go (pruebas handlers: heartbeat y configuraciÃƒÂ³n)
		- web/administrar_empresa/configuracion_de_estaciones.html (UI: registrar device Ã¢â€ â€™ estaciÃƒÂ³n)
		- web/administrar_empresa/estaciones.html (indicador visual sensor aÃƒÂ±adido)
		- web/estilos.css (estilos del indicador)
	- DescripciÃƒÂ³n: Se implementÃƒÂ³ un mÃƒÂ³dulo ligero para registrar dispositivos Raspberry Pi por empresa y estaciÃƒÂ³n, recibir heartbeats pÃƒÂºblicos y reflejar el estado (negro/verde) en las tarjetas de estaciones. Incluye pruebas unitarias para DB y handlers.
	- VerificaciÃƒÂ³n: `go test ./...` ejecutado y tests verdes.

## 2026-04-11
- Generador automÃƒÂ¡tico de cÃƒÂ³digos de descuento: formato moderno `PREFIJO-XXXX-XXXX` (`DSCT-AB12-CD34`).
	- Archivos modificados: `backend/db/codigos_descuento.go`, `web/administrar_empresa/codigos_de_descuento.html`.
	- Se mantiene ÃƒÂ­ndice ÃƒÂºnico por `(empresa_id, codigo)` y se implementÃƒÂ³ reintentos en inserciÃƒÂ³n para manejar colisiones raras.
	- Se actualizaron `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos` y `documentos/historial_de_cambios`.
	- Pruebas unitarias de DB asociadas: todas en verde.

## 2026-04-09
- Gobernanza de agente: se oficializa flujo DIAN SaaS multiempresa en instrucciones del repositorio.
	- `copilot-instructions.md` incorpora regla oficial: software DIAN compartido (un `Software ID`/`Software PIN` para la plataforma) con credenciales tributarias obligatorias por empresa (`nit`, `token_emisor_ref`, `certificado_clave_ref`).
	- Se explicita trazabilidad por `empresa_id` en cada envio real y prohibicion de reutilizar token/firma entre empresas.
	- Trazabilidad sincronizada en `documentos/historial_de_cambios`.
- Facturacion electronica DIAN (Colombia): modo SaaS multiempresa con software compartido y credenciales por empresa.
	- `backend/db/modulos_faltantes.go` amplia `empresa_dian_configuracion` con `usar_software_compartido`, `software_id_compartido_ref`, `software_pin_compartido_ref` e indice `ix_dian_empresa_shared_mode`.
	- `backend/handlers/modulos_faltantes.go` agrega resoluciÃƒÂ³n de software efectivo (`resolveDIANSoftwareCredentials`) con fallback global `DIAN_SHARED_SOFTWARE_ID/DIAN_SHARED_SOFTWARE_PIN`.
	- `sendDIANDocumentoReal` y `runDIANSetPruebasEnvio` reportan `software_modo` y `software_id` efectivo, manteniendo `NIT/token/certificado` por empresa.
	- `backend/handlers/modulos_faltantes_test.go` agrega `TestEmpresaDIANColombiaHandlerSoftwareCompartidoMultiempresa` (validado).
	- Documentacion sincronizada en `documentos/informacion_para_pruebas_plataforma_DIAN`, `documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos` y `documentos/historial_de_cambios`.
- Facturacion electronica DIAN (Colombia): se completa flujo de envio automatizado del set de habilitacion.
	- `backend/handlers/modulos_faltantes.go` agrega `action=enviar_set_pruebas` en `/api/empresa/facturacion_electronica/dian` para distribuir y enviar lotes de facturas/notas con resumen operacional por estado.
	- `sendDIANDocumentoReal` incorpora `documento_tipo` y override de `test_set_id` para interoperabilidad en el envio del lote.
	- `backend/handlers/modulos_faltantes_test.go` agrega `TestEmpresaDIANColombiaHandlerEnviarSetPruebas` y se valida junto con pruebas DIAN existentes (3 passed, 0 failed).
	- Se actualiza `documentos/informacion_para_pruebas_plataforma_DIAN` con aclaracion de configuracion de URL WSDL, resultados de pruebas y payload recomendado para los 50 documentos requeridos por DIAN.
	- Se sincroniza gobernanza documental en `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md` y `documentos/historial_de_cambios`.
- Documentacion DIAN: se crea `documentos/informacion_para_pruebas_plataforma_DIAN` para organizar en una sola referencia los datos de pruebas extraidos de `documentos/DATOS DIAN.mhtml`.
- Facturacion electronica DIAN (Colombia): carga de datos de prueba para Motel malibu con licencia activa.
	- Se registra configuracion DIAN en `backend/db/pcs_empresas` para `empresa_id=6` usando los datos de `documentos/DATOS DIAN.mhtml` (Software ID/PIN, TestSetId, prefijo, resolucion y rango).
	- Se valida tecnicamente el modulo DIAN con `handlers.EmpresaDIANColombiaHandler` ejecutando:
		- `checklist` y `validar` sin faltantes (`ok=true`).
		- `generar_cufe_demo` y `generar_xml_demo` con resultados correctos para documento de prueba.
	- En esta iteracion no fue necesario crear credenciales API adicionales en configuracion avanzada de super administrador para las pruebas internas de habilitacion.

## 2026-04-08
- Panel `administrar_empresa`: menu izquierdo desacoplado del contenido derecho.
	- `web/administrar_empresa.html` activa shell dedicado (`admin-empresa-shell`).
	- `web/estilos.css` separa scroll de sidebar e iframe para que navegar/desplazar el menu no afecte movimiento ni visibilidad de la subpagina cargada.
	- Incluye ajuste responsive para mantener usabilidad en pantallas moviles.

- Configuracion avanzada (DeepSeek/Gmail/Wompi): cifrado obligatorio robustecido y correcciÃƒÂ³n de guardado cuando faltaba `CONFIG_ENC_KEY`.
	- `backend/main.go` ahora carga `.env.local/.env`, asegura `CONFIG_ENC_KEY` (autogenera y persiste en desarrollo cuando no existe) y normaliza secretos legacy para dejarlos cifrados.
	- `backend/handlers/super_config_backup_handlers.go` fuerza cifrado en restore de secretos y rechaza restore plano sin clave de cifrado.
	- `scripts/iniciar_servidor.ps1` valida/carga/autogenera `CONFIG_ENC_KEY` antes del arranque para evitar errores `400` en `/super/api/config/ai`.
	- `web/super/configuracion_avanzada.html` muestra en DeepSeek solo fecha/hora de ultima actualizaciÃƒÂ³n, sin exponer fragmentos de credencial.
	- Enmascarado de secretos reforzado en backend (`ai_config_handlers.go`, `usuarios_empresa.go`, `payments_handlers.go`) con `********`.
	- Pruebas nuevas para restore cifrado y guardado cifrado DeepSeek en `backend/handlers/system_empresas_handlers_test.go`.

- Monitoreo operativo de arranque/reinicio de servidor con alerta por correo configurable.
	- `backend/db/super_servidor_eventos.go` agrega tabla de auditoria `super_servidor_eventos` para registrar inicio/reinicio, motivo, estado previo y resultado de notificacion.
	- `backend/handlers/server_runtime_notifications.go` implementa registro de arranque, deteccion de reinicio inesperado, escritura de estado runtime y bitacora local (`backend/logs/server_runtime_state.json`, `backend/logs/server_reinicio.log`).
	- `backend/main.go` integra registro de evento al arrancar, cierre controlado por seniales (`SIGINT/SIGTERM`) y trazabilidad de motivo de apagado.
	- `backend/handlers/usuarios_empresa.go` y `web/super/configuracion_avanzada.html` incorporan `gmail.restart_alert_to` en configuracion avanzada para correo destino de alertas de reinicio.
	- `backend/handlers/super_config_backup_handlers.go` incluye `gmail.restart_alert_to` dentro de claves criticas de backup/restore.
	- `scripts/iniciar_servidor.ps1` propaga `PCS_SERVER_START_REASON=inicio_script_iniciar_servidor` para enriquecer el motivo de arranque.
	- Pruebas nuevas/actualizadas en `backend/handlers/server_runtime_notifications_test.go` y `backend/handlers/system_empresas_handlers_test.go`.

- Script de despliegue local a GitHub mejorado en `scripts/actualizar_repositorio.ps1`.
	- Se corrige el armado de argumentos de `git push` para evitar enviar parametros vacios en PowerShell y reducir fallos intermitentes al subir cambios.
	- Se incorporan mensajes por etapas (`1/8` a `8/8`) con resumen final de commit, rama, remoto y estado de push.
	- Se centraliza el manejo de reintentos con `-ForcePush` y confirmacion explicita (`SI`) para mantener seguridad operacional.
	- Se refuerza el flujo de bitacoras automaticas para reportar mejor cuando falla el push documental.

- Arranque local y estÃƒÂ¡ticos web: mejoras en `scripts/iniciar_servidor.ps1` y correcciÃƒÂ³n de raÃƒÂ­z `/`.
	- `scripts/iniciar_servidor.ps1` ahora muestra progreso por etapas (`1/8` a `8/8`), mensajes `[INFO]/[OK]/[AVISO]/[ERROR]` y salida explÃƒÂ­cita para `-Background` sin abrir navegador.
	- `backend/main.go` corrige la resoluciÃƒÂ³n de carpeta web para priorizar candidatos con `index.html`, evitando servir accidentalmente `backend/web` (solo `uploads/`).
	- `backend/main.go` agrega manejo de `/favicon.ico` con fallback a `web/img/punto_venta.png` para evitar 404 en consola.
	- `web/index.html` declara favicon explÃƒÂ­cito con `link rel="icon"`.
	- Validaciones: compilaciÃƒÂ³n de `backend/main.go` (`go test . -run "^$"`) y parseo de PowerShell de `scripts/iniciar_servidor.ps1` OK.

- Backups empresariales: nueva opciÃƒÂ³n para eliminar informaciÃƒÂ³n por fecha de corte.
	- `backend/handlers/backups_empresariales.go` agrega `action=depurar_fecha` en `/api/empresa/backups`, con validaciÃƒÂ³n de `fecha_corte` y filtros opcionales `include_tables`/`exclude_tables`.
	- `backend/db/backups_empresariales.go` incorpora `PurgeEmpresaDataByDateCorte` para eliminar registros por `empresa_id` con fecha <= corte (inclusive), con detalle de eliminaciones por tabla.
	- La depuraciÃƒÂ³n permite generar backup previo automÃƒÂ¡tico antes de ejecutar borrado para trazabilidad y recuperaciÃƒÂ³n.
	- `backend/handlers/empresa_permisos.go` clasifica esta acciÃƒÂ³n como `permActionApprove` en mÃƒÂ³dulo seguridad.
	- `web/administrar_empresa/backups.html` incorpora UI de depuraciÃƒÂ³n por fecha con confirmaciÃƒÂ³n explÃƒÂ­cita y resumen de resultados.
	- Se agregan pruebas: `TestEmpresaBackupsPurgeByDateCorte` (DB) y `TestEmpresaBackupsHandlerPurgeByDate` (handler).
	- Validaciones: pruebas de backups en verde y compilaciÃƒÂ³n dirigida de paquetes backend crÃƒÂ­ticos OK.

- Chat y tareas: documentos/fotos entre usuarios de empresa y administrador.
	- `backend/handlers/chat_tareas.go` deriva autor desde sesion autenticada (usuario/admin), evita suplantacion de `autor_*` y auto-registra participantes emisores en conversaciones.
	- Al crear conversacion desde usuario, se agrega automaticamente el admin propietario de la empresa como participante para habilitar intercambio usuario-admin.
	- Se amplian extensiones permitidas de adjuntos en backend y UI: `doc/docx/xls/xlsx/ppt/pptx/rtf/odt/ods/odp` (ademas de imagen/audio/pdf/txt/csv/json).
	- `web/administrar_empresa/chat_y_tareas.html` ahora envÃƒÂ­a metadata de actor efectiva (`autor_tipo`, `autor_ref_id`, `autor_nombre`, `autor_email`) segun sesion.
	- Se agrega `backend/handlers/chat_tareas_test.go` con pruebas para actor usuario derivado, upload `.docx` y auto-participacion usuario/admin.
	- Validaciones: pruebas dirigidas de handlers chat/tareas y compilacion de paquetes backend criticos en verde.

- Chat y tareas: higiene de pruebas y limpieza de artefactos temporales.
	- `backend/handlers/chat_tareas_test.go` incorpora limpieza automatica (`t.Cleanup`) de uploads temporales por empresa para mantener el workspace limpio tras las pruebas.
	- Se retiran artefactos locales residuales de validacion (`.docx` y binarios `.test.exe`) para evitar ruido en el arbol de cambios.
	- Validaciones: `go test ./handlers -run "TestEmpresaChatTareas" -count=1` y compilacion dirigida de paquetes backend (`./auth ./db ./handlers ./metrics ./utils`) en verde.

- ConfiguraciÃƒÂ³n monetaria y numÃƒÂ©rica por empresa en panel de configuraciÃƒÂ³n.
	- `backend/db/empresa_configuracion_avanzada.go` amplÃƒÂ­a `empresa_configuracion_avanzada` con `moneda_codigo`, `sistema_numerico`, `usar_decimales` y `cantidad_decimales`.
	- `web/administrar_empresa/configuracion.html` agrega tarjeta para configurar moneda operativa, sistema numÃƒÂ©rico y precisiÃƒÂ³n decimal por empresa.
	- `backend/db/carritos_compras.go` aplica la moneda configurada por empresa como fallback al crear carritos sin moneda explÃƒÂ­cita.
	- `backend/main.go` registra la migraciÃƒÂ³n `2026-04-08-030-configuracion-monetaria-numerica`.
	- Validaciones: compilaciÃƒÂ³n de `db`, `handlers` y `main` en backend OK.

- ConfiguraciÃƒÂ³n IA migrada de Gemini a DeepSeek en super administrador y chat empresarial corregido.
	- `web/super/configuracion_avanzada.html` ahora gestiona credencial `deepseek:deepseek-chat` y corrige flujo de guardado de credenciales IA.
	- `backend/handlers/ai_credentials_catalog.go` registra `DEEPSEEK_API_KEY` como credencial IA activa en panel super.
	- `backend/handlers/chat_con_inteligencia_artificial_controller.go` usa DeepSeek como proveedor del chat IA por empresa.
	- `web/administrar_empresa/chat_con_inteligencia_artificial.html` actualiza etiquetas/mensajes para modelo IA genÃƒÂ©rico (sin acoplamiento a Gemini).
	- Validaciones: compilaciÃƒÂ³n de `handlers` y `main` en backend OK.

- Gobernanza documental reforzada para Agente Go y limpieza de documentos obsoletos.
	- Se actualiza `copilot-instructions.md` con regla obligatoria: si un modulo se crea o modifica, deben actualizarse `documentos/descripcion_de_modulos` y `documentos/matriz_roles_permisos_pos_multiempresa.md` en la misma iteracion.
	- Se refuerza la regla de sincronizacion de documentacion tecnica relacionada y trazabilidad en `documentos/descripcion_de_archivos` y `documentos/historial_de_cambios`.
	- Se elimina `documentos/modulos del proyecto.md` por duplicidad frente al documento canonico `documentos/descripcion_de_modulos`.
	- Se actualizan `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos` y `documentos/matriz_roles_permisos_pos_multiempresa.md` con la politica nueva.

- Cierre de pasos operativos 1, 2 y 3 solicitados en pendientes.
	- Paso 1: revisiÃƒÂ³n/ajuste de accesos directos de mÃƒÂ³dulos.
		- validaciÃƒÂ³n de consistencia de enlaces del panel empresa.
		- se agrega panel de accesos directos dinÃƒÂ¡mico en `web/administrar_empresa/inicio.html` con visibilidad por permisos/licencia.
	- Paso 2: notas de voz en chat y tareas.
		- backend: `chat_tareas` incorpora campos `nota_voz_*` y endpoint `POST /api/empresa/chat_tareas/tareas/nota_voz`.
		- frontend: `chat_y_tareas.html` incorpora grabaciÃƒÂ³n con MediaRecorder para mensajes/tareas, envÃƒÂ­o y reproducciÃƒÂ³n de audio.
	- Paso 3: super rol/permisos por licencia.
		- `licencias` incorpora `modulos_habilitados` y `super_rol_habilitado`.
		- middleware de permisos aplica restricciones por licencia y rol efectivo por empresa.
		- UI super de licencias permite configurar mÃƒÂ³dulos habilitados y super rol por plan.
	- Validaciones:
		- `go test ./handlers -run "Test(EmpresaPermisosContextoHandlerRestringeModulosPorLicencia|WithEmpresaFinanzasPermissionsSupervisorConSuperRolLicencia|WithEmpresaVentasPermissionsBloqueaModuloNoHabilitadoPorLicencia|EmpresaPermisosContextoHandlerRetornaPermisosPorRol)$" -count=1` -> OK.
		- `go test ./db ./handlers -run "^$" -count=1` -> OK.
		- `go test . -run "^$" -count=1` -> OK.

- Sincronizacion documental de pendientes operativos en `Pendiente Notas`.
	- Se actualiza fecha de corte a 2026-04-08 y se consolida estado real de avance.
	- Se registran como completados en pendientes: soporte remoto empresarial y venta digital global.
	- Se deja explicito el bloque pendiente de siguiente iteracion: notas de voz en chat/tareas, super rol por licencia y sensor de puertas para motel.

- Cierre de validacion final del modulo de soporte remoto empresarial.
	- Validaciones ejecutadas al cierre:
		- `go test ./db -run "TestSoporteRemoto" -count=1` -> OK.
		- `go test ./handlers -run "Test(EmpresaSoporteRemotoHandlerFlow|PublicSoporteRemotoAgentHeartbeatAndStateUpdate)$" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.
	- Estado: modulo soporte remoto listo para cierre tecnico con pruebas dirigidas y validacion global en verde.

- Implementacion del modulo de soporte remoto empresarial (estilo AnyDesk/TeamViewer simplificado) con aislamiento por `empresa_id`.
	- Backend DB: se crea `backend/db/soporte_remoto.go` con tablas `empresa_soporte_remoto_configuracion`, `empresa_soporte_remoto_dispositivos` y `empresa_soporte_remoto_sesiones`, incluyendo validacion de PIN hash, heartbeat de agente y token temporal de visualizacion.
	- Backend handlers: se crea `backend/handlers/soporte_remoto.go` con endpoints:
		- empresarial: `GET/POST/PUT/PATCH /api/empresa/soporte_remoto` (configuracion, CRUD dispositivos, sesiones, aprobacion/finalizacion, resolver visualizacion y exportes en `pdf/xls/csv/json/txt`).
		- publico agente/plugin: `POST /api/public/soporte_remoto` (heartbeat, aprobar/finalizar sesion desde agente).
	- Integracion backend: `backend/main.go` registra `EnsureEmpresaSoporteRemotoSchema`, migracion `2026-04-08-029-soporte-remoto-empresa`, rutas protegida/publica; `backend/utils/utils.go` habilita la ruta publica del agente.
	- Seguridad y menu: `backend/handlers/empresa_permisos.go`, `web/administrar_empresa.html` y `web/js/administrar_empresa.js` agregan `linkSoporteRemoto` con control de permisos por modulo seguridad (`A`).
	- Frontend: se crean `web/administrar_empresa/soporte_remoto.html` y `web/administrar_empresa/soporte_remoto_view.html` para gestion de dispositivos/sesiones y visor embebido por token.
	- QA: se crean `backend/db/soporte_remoto_test.go` y `backend/handlers/soporte_remoto_test.go`.
	- Validaciones ejecutadas: `go test ./db -run "TestSoporteRemoto" -count=1` y `go test ./handlers -run "Test(EmpresaSoporteRemotoHandlerFlow|PublicSoporteRemotoAgentHeartbeatAndStateUpdate)$" -count=1` en verde (con incidencia local de lock temporal de `handlers.test.exe` al limpiar artefacto en Windows).

- Implementacion del modulo de venta digital global (super administrador + compra publica Wompi).
	- Backend DB: se crea `backend/db/venta_digital.go` con tablas `super_venta_digital_configuracion`, `super_venta_digital_items` y `super_venta_digital_ordenes`, incluyendo indices y flujo de orden/pago/entrega.
	- Backend handlers: se crea `backend/handlers/venta_digital.go` con endpoints:
		- super: `GET/POST/PUT/PATCH/DELETE /super/api/venta_digital` (configuracion, CRUD catalogo, uploads y ordenes).
		- publico: `GET/POST /api/public/venta_digital` (catalogo, crear pago Nequi por Wompi, consulta de estado y entrega por correo).
	- Integracion backend: `backend/main.go`, `backend/utils/utils.go` y `backend/handlers/payments_handlers.go` integran schema/rutas/middleware publico y sincronizacion por webhook Wompi para entrega de licencia.
	- Frontend: se crean `web/super/venta_digital.html` y `web/venta_digital.html`, y se agregan accesos en `web/menu.js`, `web/super_administrador.html` y `web/super/configuracion_avanzada.html`.
	- Validacion tecnica: `go test ./... -run "^$" -count=1` en `backend` -> compilacion global OK.

- Refuerzo de QA para permisos dinamicos por rol.
	- Testing DB: se crea `backend/db/roles_permisos_usuario_test.go` para validar replace/list/lookup y fallback sin tablas en el esquema de permisos de rol.
	- Testing Handler: se crea `backend/handlers/roles_tipos_usuario_permisos_test.go` con cobertura de `GET/PUT /super/api/roles_de_usuario/permisos` y caso `rol_id` inexistente.
	- Ajuste de fiabilidad en pruebas: `backend/handlers/empresa_permisos_test.go` alinea el helper de schema con columna `observaciones` requerida por consultas reales.
	- Validacion ejecutada: `go test ./db -run "TestRolesPermisos" -count=1` y `go test ./handlers -run "TestRolesDeUsuarioPermisosHandler|TestEmpresaPermisosContextoHandler|TestWithEmpresaSeguridadPermissionsRequiereAprobacionParaCambioPermisos|TestSuperEndpointsPermisosPorRol" -count=1` en verde.

- Inicio de implementacion del modulo de permisos dinamicos por rol.
	- Backend: nuevas tablas y capa DB para overrides por `rol` en modulo/accion y pagina (`roles_de_usuario_permisos`, `roles_de_usuario_paginas_permisos`).
	- Backend: middleware de permisos empresariales ahora aplica overrides dinamicos y `/api/empresa/permisos_contexto` incluye mapa `paginas`.
	- Backend: nuevo endpoint super `GET/PUT /super/api/roles_de_usuario/permisos` para gestionar matriz de permisos por rol.
	- Frontend: nueva pantalla `/super/permisos_rol.html`, acceso desde menu super y boton directo en listado de roles.
	- Frontend: menu empresa aplica visibilidad por pagina desde contexto de permisos.
	- Validacion tecnica: pruebas dirigidas de permisos en handlers y compilacion de `main` en verde.

- Plan profesional agregado al tablero de pendientes para completar e integrar el modulo de roles y permisos por usuario de empresa.
	- Se actualiza `Pendiente Notas` al final del documento con fases de implementacion (1..10), cronograma sugerido y criterios de aceptacion.
	- Se mantiene enfoque de integracion multiempresa con aislamiento por `empresa_id`, trazabilidad y UAT por rol.

## 2026-04-07
- Normalizacion documental del tablero de pendientes (modulo 35).
	- Se ajusta `Pendiente Notas` para reemplazar la etiqueta `COMPLETADO PARCIAL` por contexto historico de fases.
	- Se mantiene el estado oficial de cierre total del modulo 35 sin cambios funcionales.

- Continuacion de cierre de pendientes: tablero operativo actualizado con evidencia de pruebas del modulo 37.
	- Se actualiza `Pendiente Notas` para aÃƒÂ±adir ejecucion dirigida de pruebas de `venta_publica` en handlers.
	- Se explicita estado general sin pendientes de modulos (`1..38` y bloque "Ultimo" en `COMPLETADO`).

- Pruebas dirigidas para el modulo 37 (Venta publica por empresa + Wompi).
	- Se agrega `backend/handlers/venta_publica_test.go` con cobertura de:
		- flujo empresarial (`config`, `crear`, `detalle`, `catalogo`, `activar/desactivar`).
		- flujo publico de catalogo y creacion de orden con Wompi inactivo (respuesta controlada `412`).
		- validacion de `estado_pago` cuando no se envia `order_code`.
	- Validacion: `runTests` en `backend/handlers/venta_publica_test.go` -> 3 passed, 0 failed.

- Cierre del modulo 37 (Venta publica por empresa + Wompi) y cierre de pendientes documentales 38/Ultimo.
	- Backend `db`:
		- `backend/db/venta_publica.go` (nuevo) agrega tablas `empresa_venta_publica_configuracion`, `empresa_venta_publica_items` y `empresa_venta_publica_ordenes`, con operaciones CRUD/listado/ordenes y resolucion por slug.
	- Backend `handlers`:
		- `backend/handlers/venta_publica.go` (nuevo) agrega `/api/empresa/venta_publica` y `/api/public/venta_publica`.
		- soporta configuracion Wompi por empresa, carga de imagen de catalogo, creacion de pago Nequi y consulta de estado de orden.
	- Integracion:
		- `backend/main.go` integra schema, migracion `2026-04-07-028-venta-publica-wompi`, rutas API y rewrite de `/{slug}/venta_publica.html`.
		- `backend/utils/utils.go` habilita acceso publico a rutas/API de venta publica y recursos en `/uploads/`.
	- Frontend:
		- `web/administrar_empresa/venta_publica.html` (nuevo) para administracion del canal online por empresa.
		- `web/venta_publica.html` (nuevo) para clientes finales (catalogo, carrito, pago y estado).
		- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` agregan `linkVentaPublica` con permisos de ventas.
	- Documentacion:
		- se crea `documentos/descripcion_de_modulos` (modulo 38).
		- se amplia `web/ayuda/ayuda.html` con tutoriales de ventas, productos/impuestos, venta publica y configuracion por empresa (cierre del "Ultimo").
		- se sincronizan `documentos/estructura_bd.md`, `estructura_bd.md` y diagramas (`estructura_del_codigo`, `diagrama_flujo_procesos`, `diagrama_entidad_relacion`).
	- Validaciones ejecutadas:
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.
		- `get_errors` en archivos modificados -> sin errores.

- Cierre del modulo 36 (Backups empresariales): snapshots por empresa, restauracion trazable, exportacion multiformato y UI dedicada.
	- Backend `db`:
		- `backend/db/backups_empresariales.go` agrega tablas `empresa_backups` y `empresa_backups_restauraciones` con `EnsureEmpresaBackupsSchema`.
		- implementa construccion de payload (`BuildEmpresaBackupPayload`), alta de snapshot (`CreateEmpresaBackupSnapshot`), historial/detalle (`List/Get`) y restauracion controlada (`RestoreEmpresaBackupByID`).
		- incorpora trazabilidad de integridad (`hash_contenido`) y metadata/version de snapshot.
		- `backend/db/backups_empresariales_test.go` agrega pruebas de flujo snapshot/restauracion y listado/payload.
	- Backend `handlers`:
		- `backend/handlers/backups_empresariales.go` agrega endpoint `/api/empresa/backups` con acciones `listar|crear|detalle|export|restaurar|activar|desactivar`.
		- `backend/handlers/backups_empresariales_test.go` agrega cobertura de create/list/detail/export/restore/toggle y not-found en restore.
		- `backend/handlers/empresa_permisos.go` clasifica `restaurar|restore` como accion de aprobacion (`permActionApprove`).
	- Integracion y frontend:
		- `backend/main.go` registra `EnsureEmpresaBackupsSchema`, migracion `2026-04-07-027-backups-empresariales` y ruta protegida `/api/empresa/backups`.
		- `web/administrar_empresa/backups.html` (nuevo) implementa flujo profesional de backups por empresa.
		- `web/administrar_empresa.html`, `web/js/administrar_empresa.js` y `web/estilos.css` integran acceso `linkBackups` y estilos del modulo.
	- Validaciones ejecutadas:
		- `go test ./db -run "^TestEmpresaBackups" -count=1` -> OK.
		- `go test ./handlers -run "^TestEmpresaBackupsHandler" -count=1` -> OK.
		- `go test . -run "^$" -count=1` -> compilacion de `main` OK.

- Cierre del modulo 35 (Creditos y cartera): reglas de limites por cliente, permisos finos por rol en workflow y auditoria ampliada.
	- Backend `db`:
		- `backend/db/creditos.go` agrega tabla `empresa_creditos_clientes_limites` con indice unico `(empresa_id, cliente_id)` y funciones `Get/List/Upsert/SetEstado` para administrar limites por cliente.
		- se incorpora validacion de limites por cliente en `CreateEmpresaCredito` y `UpdateEmpresaCredito` (saldo total maximo y maximo de creditos activos).
	- Backend `handlers`:
		- `backend/handlers/creditos.go` agrega acciones `limites_cliente`, `limite_cliente`, `upsert_limite_cliente` y `eliminar_limite_cliente` en `/api/empresa/creditos`.
		- se incorpora validacion de permiso fino por tipo de workflow: `contabilidad` puede decidir `reverso_abono` y `refinanciacion` queda restringida a `administrador`.
		- se amplian eventos de auditoria no bloqueante para solicitud/aprobacion/rechazo de workflow, cambios de limites y denegaciones por permiso fino.
	- Pruebas:
		- `backend/db/creditos_test.go` agrega `TestEmpresaCreditosClienteLimitesBloqueaExceso` y `TestEmpresaCreditosClienteLimitesCRUD`.
		- `backend/handlers/creditos_test.go` agrega `TestEmpresaCreditosHandlerLimitesClienteYBloqueo` y `TestEmpresaCreditosHandlerWorkflowPermisoFinoPorTipo`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestEmpresaCreditosClienteLimites(BloqueaExceso|CRUD)$" -count=1` -> OK.
		- `go test ./handlers -run "TestEmpresaCreditosHandler(LimitesClienteYBloqueo|WorkflowPermisoFinoPorTipo)$" -count=1` -> OK.
		- `runTests` sobre `backend/db/creditos_test.go` y `backend/handlers/creditos_test.go` (casos nuevos) -> 6 passed, 0 failed.

- Avance del modulo 35 (Creditos y cartera): workflow avanzado de reversos/anulaciones y refinanciacion con aprobacion multinivel.
	- Backend `db`:
		- `backend/db/creditos.go` agrega tabla `empresa_creditos_workflow`, filtros/listado por estado/tipo y funciones de negocio para `solicitud`, `aprobacion`, `rechazo` y ejecucion automatica.
		- se implementa ejecucion de `reverso_abono` y `refinanciacion` con trazabilidad de `movimiento_resultado_id`, `resultado_json` e `historial_aprobaciones_json`.
		- se corrige colision de `numero_cuota` en refinanciacion generando nuevas cuotas con secuencia incremental despues del ultimo numero historico.
		- `backend/db/creditos_test.go` agrega `TestEmpresaCreditosWorkflowReversoAprobadoEjecutaReversion` y `TestEmpresaCreditosWorkflowRefinanciacionAprobadaRegeneraCuotas`.
	- Backend `handlers`:
		- `backend/handlers/creditos.go` agrega acciones `workflows`, `solicitar_reverso`, `solicitar_refinanciacion`, `aprobar_workflow`, `rechazar_workflow`.
		- `backend/handlers/empresa_permisos.go` clasifica acciones de aprobacion/rechazo de workflow como `permActionApprove` en modulo finanzas.
		- `backend/handlers/creditos_test.go` agrega `TestEmpresaCreditosHandlerWorkflowReversoSolicitudYAprobacion`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestEmpresaCreditosWorkflow(ReversoAprobadoEjecutaReversion|RefinanciacionAprobadaRegeneraCuotas)$" -count=1` -> OK.
		- `go test ./handlers -run "TestEmpresaCreditosHandlerWorkflowReversoSolicitudYAprobacion$" -count=1` -> OK.
		- `go test ./... -count=1` -> OK.

- Avance del modulo 35 (Creditos y cartera): integracion contable/caja-bancos/pasarelas en abonos con asientos automaticos por politica.
	- Backend `db`:
		- `backend/db/eventos_contables.go` extiende contrato contable con `creditos.credito_abono_registrado`.
		- agrega plantilla de lineas contables para abonos de credito (caja/bancos, cartera de creditos, intereses y mora).
		- `backend/db/eventos_contables_test.go` agrega `TestProcessEmpresaEventosContablesPendientesCreditoAbonoGeneraLineasCartera`.
	- Backend `handlers`:
		- `backend/handlers/creditos.go` integra registro de evento contable al `action=abono` y procesamiento automatico de asientos por politica (`procesar_asientos`, `asientos_limit`, `max_reintentos`).
		- se incorpora metrica de canal de pago (`caja`, `bancos`, `pasarela`) para trazabilidad operativa por abono.
		- `backend/handlers/creditos_test.go` agrega `TestEmpresaCreditosHandlerAbonoIntegraContabilidadYAsientos`.
	- Trazabilidad funcional:
		- `Pendiente Notas` marca completada la integracion contable de modulo 35 y mantiene pendientes de reversos/refinanciacion y limites/permisos.
	- Validaciones ejecutadas:
		- `go test ./db -run "Test(EmpresaEventosContablesCreateAndList|ProcessEmpresaEventosContablesPendientesGeneraAsientosIdempotentes|ProcessEmpresaEventosContablesPendientesCreditoAbonoGeneraLineasCartera|EmpresaCreditosFlowCrearCuotasAbonoYResumen|EmpresaCreditosMoraDashboard)$" -count=1` -> OK.
		- `go test ./handlers -run "TestEmpresaCreditosHandler(FlujoBasico|AlertasMoraYReporte|AbonoIntegraContabilidadYAsientos)$" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

- Avance del modulo 35 (Creditos y cartera): alertas proactivas de vencimiento y ranking avanzado de morosidad.
	- Backend `db`:
		- `backend/db/creditos.go` agrega dashboard de morosidad (`GetEmpresaCreditosMoraDashboard`) con bloques de proximos a vencer, vencidos y ranking.
		- `backend/db/creditos_test.go` agrega `TestEmpresaCreditosMoraDashboard`.
	- Backend `handlers`:
		- `backend/handlers/creditos.go` agrega acciones `alertas|alertas_mora|morosidad|ranking_morosidad`.
		- `action=reporte` soporta `tipo=morosidad` para exportacion en `json/csv/txt/xls/pdf`.
		- `backend/handlers/creditos_test.go` agrega `TestEmpresaCreditosHandlerAlertasMoraYReporte`.
	- Frontend:
		- `web/administrar_empresa/creditos.html` incorpora panel operativo de alertas/ranking con filtros (`dias_proximos`, `top`, `include_inactive`) y exportacion dedicada.
		- `web/estilos.css` incorpora estilos `creditos-alertas-*` para toolbar y grilla responsive.
	- Diagramas y trazabilidad:
		- `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md` se actualizan con el nuevo flujo de morosidad.
	- Validaciones ejecutadas:
		- `go test ./db -run TestEmpresaCreditosMoraDashboard -count=1` -> OK.
		- `go test ./handlers -run TestEmpresaCreditosHandlerAlertasMoraYReporte -count=1` -> OK.
		- `go test ./... -run "^TestEmpresaCreditos" -count=1` -> OK.

- Avance del modulo 35 (Creditos y cartera): publicadas guias operativas en centro de ayuda y manual por rol.
	- Documentacion funcional:
		- `web/ayuda/ayuda.html` integra acceso rapido a creditos, bloque tutorial `30) Creditos y cartera`, guia operativa dedicada por flujo y manual por rol (administrador, caja/cobranza y auditoria).
		- Se documentan endpoints clave de `/api/empresa/creditos` en la seccion de APIs para operacion y soporte.
	- Trazabilidad:
		- `Pendiente Notas` retira del listado pendiente del modulo 35 la tarea de guias operativas y la marca dentro de completado parcial.
		- `documentos/descripcion_del_proyecto` sincroniza el alcance del modulo 35 incluyendo referencia a la guia operativa por rol.
	- Validaciones ejecutadas:
		- validacion de editor (`get_errors`) sobre `web/ayuda/ayuda.html`, `Pendiente Notas`, `CHANGELOG.md`, `documentos/historial_de_cambios`, `documentos/descripcion_de_archivos` y `documentos/descripcion_del_proyecto` -> sin errores.

- Avance del modulo 35 (Creditos y cartera): fase 2 frontend base implementada con pantalla dedicada e integracion de menu/permisos.
	- Frontend:
		- `web/administrar_empresa/creditos.html` (nuevo) incorpora formulario de creacion de credito, filtros de cartera, resumen, tabla de creditos, panel de abonos y estado de cuenta (cuotas/movimientos).
		- integra exportacion de cartera en `json/csv/txt/xls/pdf` usando `action=reporte` del backend.
		- incluye acciones de operacion diaria: prellenado de abono, cambio de estado de credito y activar/desactivar fila.
	- Navegacion y permisos:
		- `web/administrar_empresa.html` agrega enlace lateral `linkCreditos`.
		- `web/js/administrar_empresa.js` agrega `linkCreditos` al catalogo de permisos como modulo `finanzas` accion `C`.
	- Estilos:
		- `web/estilos.css` agrega componentes `creditos-*` para grids de filtros/resumen, acciones de tabla y detalle responsive.
	- Validaciones ejecutadas:
		- validacion de editor (`get_errors`) sobre archivos frontend modificados -> sin errores.

- Avance del modulo 35 (Creditos y cartera): fase 1 backend implementada con esquema, API y pruebas base.
	- Backend `db`:
		- `backend/db/creditos.go` agrega tablas `empresa_creditos`, `empresa_creditos_cuotas` y `empresa_creditos_movimientos`.
		- implementa creacion de creditos, generacion automatica de cuotas, registro de abonos y resumen de cartera.
		- `backend/db/creditos_test.go` agrega `TestEmpresaCreditosFlowCrearCuotasAbonoYResumen`.
	- Backend `handlers`:
		- `backend/handlers/creditos.go` expone `GET/POST/PUT/PATCH/DELETE /api/empresa/creditos`.
		- incorpora acciones `estado_cuenta`, `resumen_cartera`, `movimientos`, `cuotas`, `abono` y `reporte`.
		- soporta exportacion de reporte en `json/csv/txt/xls/pdf`.
		- `backend/handlers/creditos_test.go` agrega `TestEmpresaCreditosHandlerFlujoBasico`.
	- Integracion:
		- `backend/main.go` registra `EnsureEmpresaCreditosSchema`, migracion `2026-04-07-026-creditos-cartera` y ruta protegida `/api/empresa/creditos`.
	- Validaciones ejecutadas:
		- `go test ./db -run TestEmpresaCreditosFlowCrearCuotasAbonoYResumen -count=1` -> OK.
		- `go test ./handlers -run TestEmpresaCreditosHandlerFlujoBasico -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

- Cierre del modulo 34 (Calculadora por empresa): historial operativo persistente por empresa, asociaciones trazables y exportacion multiformato por rango/usuario.
	- Backend `db`:
		- `backend/db/calculadora_operativa.go` agrega tablas `empresa_calculadora_configuracion` y `empresa_calculadora_operaciones` con filtros por fecha/usuario/cliente/etiqueta.
		- se incorporan operaciones de configuracion (`integrar_carritos`, `integrar_cotizaciones`), registro de operaciones etiquetadas y limpieza logica por filtros.
		- `backend/db/calculadora_operativa_test.go` agrega `TestEmpresaCalculadoraConfiguracionYHistorialFlow`.
	- Backend `handlers`:
		- `backend/handlers/calculadora_operativa.go` expone `GET/POST/PUT/DELETE /api/empresa/calculadora` con acciones `config`, `referencias`, `export`, `limpiar`, `activar/desactivar`.
		- valida referencias opcionales de `carrito_id`/`cotizacion_id` segun configuracion y conserva trazabilidad por `empresa_id`, cliente/documento, carrito/cotizacion y usuario.
		- `backend/handlers/calculadora_operativa_test.go` agrega `TestEmpresaCalculadoraHandlerConfigOperacionesFiltrosYExport`.
	- Frontend:
		- `web/administrar_empresa/calculadora.html` migra de historial local a flujo API, con filtros por rango/usuario, etiquetas, asociaciones a cliente/documento y exportacion backend.
		- `web/estilos.css` agrega estilos `calc-config-row`, `calc-meta-grid` y `calc-filter-row` para controles de configuracion/metadata/filtros en desktop y mobile.
	- Integracion:
		- `backend/main.go` registra `EnsureEmpresaCalculadoraSchema`, migracion `2026-04-07-025-calculadora-operativa` y ruta protegida `/api/empresa/calculadora` bajo permisos de finanzas.
	- Validaciones ejecutadas:
		- `go test ./db -run TestEmpresaCalculadoraConfiguracionYHistorialFlow -count=1` -> OK.
		- `go test ./handlers -run TestEmpresaCalculadoraHandlerConfigOperacionesFiltrosYExport -count=1` -> OK.

- Cierre del modulo 33 (Configuracion operativa de cobro): politicas contextuales, simulador de reglas e historial con rollback operativo.
	- Backend `db`:
		- `backend/db/configuracion_operativa.go` agrega tablas y modelos `empresa_configuracion_operativa_politicas` y `empresa_configuracion_operativa_historial`.
		- se incorpora resoluciÃƒÂ³n de rollback.
		- `backend/db/configuracion_operativa_test.go` agrega `TestEmpresaConfiguracionOperativaPoliticaContextoYRollback`.
	- Backend `handlers`:
		- `backend/handlers/configuracion_operativa.go` amplÃƒÂ­a acciones HTTP con `action=politica`, `action=simular`, `action=historial` y `action=rollback`.
		- se agrega snapshot de trazabilidad no bloqueante tras publicaciones y simulaciones guardadas.
		- `backend/handlers/configuracion_operativa_test.go` agrega `TestEmpresaConfiguracionOperativaHandlerPoliticaSimulacionHistorialYRollback`.
	- Frontend:
		- `web/administrar_empresa/configuracion.html` agrega UI de politica contextual, simulador por contexto y panel de historial/rollback.
		- `web/estilos.css` incorpora estilos para el bloque operativo extendido de simulacion e historial.
	- Validaciones ejecutadas:
		- `go test ./db -run TestEmpresaConfiguracionOperativa -count=1` -> OK.
		- `go test ./handlers -run TestEmpresaConfiguracionOperativaHandler -count=1` -> OK.

- Cierre del modulo 32 (Graficos y estadisticas): cache por panel, comparativos entre periodos, filtros avanzados y optimizacion de series largas.
	- Backend `handlers`:
		- `backend/handlers/graficos_estadisticas.go` incorpora cache en memoria con `cache.hit`, soporte de comparativo (`comparar`, `comparar_desde`, `comparar_hasta`) y filtros avanzados (`sucursal_id`, `estacion_id`, `segmento`).
		- Se agrega cobertura de filtros en respuesta (`filtros.cobertura`) y aplicacion de snapshots para mantener KPI del tablero alineados con filtros aplicados.
		- Se reemplaza truncamiento de cola por compactacion por buckets en series de ventas/finanzas/compras/asistencia para rangos extensos.
	- Frontend:
		- `web/administrar_empresa/graficos_estadisticas.html` agrega controles avanzados de filtros, comparativo, refresco sin cache y tarjetas de variacion por metrica.
		- `web/estilos.css` agrega estilos de comparativo, tendencia y comportamiento responsive para la nueva capa de filtros.
	- Pruebas:
		- `backend/handlers/graficos_estadisticas_test.go` agrega `TestEmpresaGraficosEstadisticasHandlerFiltrosComparativoYCache` para validar filtros, comparativo y cache hit/miss.
	- Validaciones ejecutadas:
		- `go test ./handlers -run TestEmpresaGraficosEstadisticasHandler -v` -> OK.
		- `go test ./handlers -count=1` -> OK.

## 2026-04-07
- Hotfix de arranque en migraciones ERP legacy (modulos faltantes): correccion de orden de creacion de indices dependientes de columnas nuevas.
	- Backend `db`:
		- `backend/db/modulos_faltantes.go` evita crear en el bloque inicial los indices que dependen de columnas agregadas por migracion (`periodo_contable`, `bloqueado_venta`, campos de aprobacion/nomina RRHH).
		- Esos indices se mantienen en la fase final posterior a `ensureColumnIfMissing`, garantizando compatibilidad con bases antiguas.
	- Validaciones ejecutadas:
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.
		- `go run .` -> backend inicia correctamente y queda en `LISTENING` en puerto `8080` (sin error `no such column: periodo_contable`).

## 2026-04-07
- Cierre del modulo 31 (Reportes): programacion automatica, versionado de plantillas y validacion de consistencia multiformato.
	- Backend `handlers`:
		- `backend/handlers/reportes_programacion.go` consolida agenda y ejecucion de reportes (`action=programacion`, `action=ejecutar_programacion`, `action=ejecuciones`, `action=validar_consistencia`).
		- Se corrige robustez en listado de ejecuciones para manejar campos `NULL` de motor legado retirado (`error_detalle`, `programacion_id`, metadatos opcionales) sin error `500`.
	- Pruebas:
		- `backend/handlers/reportes_programacion_test.go` valida versionado de plantillas y ciclo completo de programacion/ejecucion/consistencia.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaReportesHandler(PlantillasVersionado|ProgramacionEjecucionYConsistencia)$" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

## 2026-04-07
- Cierre del modulo 30 (Seguridad y permisos): deny-by-default por endpoint, matriz automatizada rol/modulo y trazabilidad de aprobacion.
	- Backend `handlers`:
		- `backend/handlers/empresa_permisos.go` exige evidencia de aprobacion trazable (`aprobado_por`, `codigo_aprobacion`) para cambios criticos de permisos en `/api/empresa/usuarios` bajo modulo seguridad.
		- Se propagan cabeceras de aprobacion para trazabilidad (`X-Permission-Approved-By`, `X-Permission-Approval-Code`, `X-Permission-Approval-Reason`) y bandera `X-Permission-Approval-Required`.
		- `backend/handlers/auditoria_empresa.go` registra metadata de aprobacion en `empresa_auditoria_eventos` (`permission_approval_required`, `permission_approved_by`, `permission_approval_code`, `permission_approval_reason`).
	- Pruebas:
		- `backend/handlers/empresa_permisos_test.go` agrega cobertura de matriz completa rol/modulo/accion y de aprobacion trazable para cambios de permisos.
		- `backend/main_empresa_routes_security_test.go` agrega barrido automatizado deny-by-default para todas las rutas `/api/empresa/*` registradas por `http.HandleFunc`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Test(EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaPermisosContextoHandlerIncluyeMatrizRoles|EmpresaPermisosContextoHandlerMatrizRolesCumplePoliticaPorModuloAccion|WithEmpresaSeguridadPermissionsRequiereAprobacionParaCambioPermisos|WithEmpresaSeguridadPermissionsAceptaAprobacionTrazableYRegistraMetadata)$" -count=1` -> OK.
		- `go test . -run "TestEmpresaRoutesUsePermissionWrappers$" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

## 2026-04-07
- Cierre del modulo 29 (Auditoria empresarial): busqueda full-text con filtros y exportacion forense con cadena de custodia basica.
	- Backend `db`:
		- `backend/db/auditoria_empresa.go` agrega soporte de busqueda `search` con FTS (cuando esta disponible) y fallback por `LIKE`, manteniendo compatibilidad con filtros avanzados existentes.
		- Se incorpora inicializacion de esquema FTS de auditoria (tabla virtual, triggers y backfill) para indexacion de `modulo`, `accion`, `recurso`, `endpoint`, `metadata_json` y `observaciones`.
	- Backend `handlers`:
		- `backend/handlers/auditoria_empresa.go` extiende `GET /api/empresa/auditoria/eventos` con `action=export_forense|forense_export|cadena_custodia` y `format=json|csv`.
		- La exportacion forense genera `hash_registro`, `hash_cadena` y `hash_global` para trazabilidad basica de cadena de custodia.
	- Pruebas:
		- `backend/db/auditoria_empresa_test.go`: `TestListEmpresaAuditoriaEventosSearchFullTextConFiltros`.
		- `backend/handlers/auditoria_empresa_test.go`: `TestEmpresaAuditoriaEventosHandlerExportForenseJSONYCSV`.
	- Validaciones ejecutadas:
		- `go test ./db -run "Test(CreateAndListEmpresaAuditoriaEventos|PurgeEmpresaAuditoriaEventos|PurgeExpiredEmpresaAuditoriaEventos|CountAndListEmpresaAuditoriaEventosWithPaginationAndSearch|CreateEmpresaAuditoriaEventoAplicaPoliticaRetencionPorModuloYSeveridad|CreateEmpresaAuditoriaEventoMantieneRetencionExplicita|ListEmpresaAuditoriaEventosSearchFullTextConFiltros)$" -count=1` -> OK.
		- `go test ./handlers -run "TestEmpresaAuditoriaEventosHandler(ConsultaYPurga|FiltrosAvanzados|ExportForenseJSONYCSV)$" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

## 2026-04-07
- Avance del modulo 29 (Auditoria empresarial): politicas de retencion por modulo y severidad.
	- Backend `db`:
		- `backend/db/auditoria_empresa.go` ahora resuelve `retencion_dias` automaticamente por combinacion de `modulo` + `severidad` inferida (resultado/codigo HTTP/metadatos), manteniendo prioridad para `retencion_dias` explicita.
		- Se enriquece `metadata_json` con trazabilidad de politica aplicada (`retencion_politica_modulo`, `retencion_politica_severidad`, `retencion_dias_resuelto`).
	- Pruebas:
		- `backend/db/auditoria_empresa_test.go` agrega:
			- `TestCreateEmpresaAuditoriaEventoAplicaPoliticaRetencionPorModuloYSeveridad`.
			- `TestCreateEmpresaAuditoriaEventoMantieneRetencionExplicita`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestCreateEmpresaAuditoriaEventoAplicaPoliticaRetencionPorModuloYSeveridad|TestCreateEmpresaAuditoriaEventoMantieneRetencionExplicita" -count=1 -v` -> OK.
		- `go test ./db -run "Test(CreateAndListEmpresaAuditoriaEventos|PurgeEmpresaAuditoriaEventos|PurgeExpiredEmpresaAuditoriaEventos|CountAndListEmpresaAuditoriaEventosWithPaginationAndSearch|CreateEmpresaAuditoriaEventoAplicaPoliticaRetencionPorModuloYSeveridad|CreateEmpresaAuditoriaEventoMantieneRetencionExplicita)$" -count=1` -> OK.
		- `go test ./handlers -run "TestEmpresaAuditoriaEventosHandlerConsultaYPurga|TestEmpresaAuditoriaEventosHandlerFiltrosAvanzados" -count=1` -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

## 2026-04-07
- Cierre del pendiente de modulo 28 (Finanzas y contabilidad): politicas de cierre/reapertura de periodos con evidencia de autorizacion.
	- Backend `handlers`:
		- `backend/handlers/finanzas.go` exige en `PUT /api/empresa/finanzas/periodos?action=cerrar|reabrir` los campos `autorizado_por`, `motivo_autorizacion` y `evidencia_autorizacion`.
		- Se incorpora trazabilidad explicita en observaciones y payload de evento contable (`policy_autorizacion`, `autorizado_por`, `motivo_autorizacion`, `evidencia_autorizacion`, `codigo_autorizacion`, `ejecutado_por`).
		- La respuesta HTTP del cierre/reapertura retorna bloque `autorizacion` para auditoria operativa.
	- Pruebas:
		- `backend/handlers/eventos_contables_modulos_test.go` actualiza `TestEmpresaFinanzasEmiteEventosContables` para validar evidencia en payload.
		- Se agrega `TestEmpresaFinanzasPeriodosRequiereEvidenciaAutorizacion` para rechazar cierre sin evidencia.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaFinanzasEmiteEventosContables|TestEmpresaFinanzasPeriodosRequiereEvidenciaAutorizacion" -count=1 -v` -> OK.
		- `go test ./handlers -run "TestEmpresaFinanzas" -count=1` -> OK.

## 2026-04-07
- Avance del modulo 28 (Finanzas y contabilidad): conciliacion bancaria automatica y tablero de desviaciones por periodo.
	- Backend `db`:
		- Se agrega tabla `empresa_finanzas_bancos_movimientos` (extractos bancarios por `empresa_id`) en `EnsureEmpresaFinanzasSchema`.
		- Se agrega `backend/db/finanzas_conciliacion_bancaria.go` con:
			- importacion idempotente de extractos por `hash_movimiento`.
			- conciliacion bancaria automatica contra `empresa_finanzas_movimientos` con tolerancia de monto/dias.
			- resumen de conciliacion/desviaciones por periodo.
	- Backend `handlers`:
		- Se amplia `EmpresaFinanzasMovimientosHandler` con acciones:
			- `GET action=conciliacion_bancaria` y `GET action=conciliacion_bancaria_export`.
			- `GET action=extractos_bancarios`.
			- `POST action=importar_extractos_bancarios` (opcional `auto_conciliar`).
			- `PUT action=conciliar_bancaria_auto`.
		- Se actualiza `resolveFinanzasPermissionAction` para clasificar conciliacion bancaria automatica como `permActionApprove`.
	- Pruebas:
		- `backend/db/finanzas_test.go`: `TestEmpresaFinanzasConciliacionBancariaAutomatica`.
		- `backend/handlers/eventos_contables_modulos_test.go`: `TestEmpresaFinanzasMovimientosHandlerConciliacionBancariaAutomatica`.
	- Validaciones ejecutadas:
		- `runTests` sobre pruebas nuevas de db/handlers -> OK.
		- `go test ./... -run "^$" -count=1` -> compilacion global backend OK.

## 2026-04-07
- Refuerzo de cobertura en capas `auth`, `metrics` y `utils`:
	- Se agregan y amplian pruebas unitarias:
		- `backend/auth/auth_test.go`
		- `backend/metrics/collector_test.go`
		- `backend/utils/utils_test.go` (incluye pruebas de middleware, contexto y manejo de errores JSON)
	- Cobertura actualizada por paquete (corte de ejecucion):
		- `auth`: `85.3%`
		- `db`: `51.4%`
		- `handlers`: `50.4%`
		- `metrics`: `78.0%`
		- `utils`: `71.1%`
	- Se actualiza evidencia en `Pendiente Notas`, `documentos/punto_13_calidad_uat_despliegue.md` y `documentos/punto_13_validacion_integral_resultado.md`.
	- Validaciones ejecutadas:
		- `runTests` sobre `backend/utils/utils_test.go` -> 16 pruebas OK.
		- `go test ./auth ./db ./handlers ./metrics ./utils -cover -count=1` (OK).
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).

## 2026-04-07
- Cierre transversal de calidad y salida controlada:
	- Se actualiza `documentos/punto_13_calidad_uat_despliegue.md` con:
		- objetivo minimo de cobertura por capa,
		- acta UAT formal por rol (`super_admin`, `admin_empresa`, `usuario_empresa`),
		- matriz UAT por modulo en estado aprobado.
	- Se amplÃƒÂ­a `documentos/release_checklist.md` con checklist estandar "listo para produccion" por modulo (seguridad, rendimiento, trazabilidad, exportacion y pruebas).
	- Se amplÃƒÂ­a `documentos/punto_13_validacion_integral_resultado.md` con evidencia complementaria de cobertura y UAT por rol.
	- Se actualiza `Pendiente Notas` para marcar completados los 3 pendientes transversales.
	- Validaciones ejecutadas:
		- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\validar_punto_13.ps1` (OK).
		- `go test ./auth ./db ./handlers ./metrics ./utils -cover -count=1` (db `51.4%`, handlers `50.4%`).
		- `go test ./handlers -run "Test(SuperEndpointsPermisosPorRol|EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaPermisosContextoHandlerIncluyeMatrizRoles|EmpresaCarritosCompraBloqueaMetodoPagoSegunRol|EmpresaCarritosCompraRespetaBloqueoPropinaYComisionPorRol|EmpresaConfiguracionOperativaHandlerConfigAndRole|EmpresaDocumentosGestionHandlerVersionadoYControlAcceso)$" -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 27 (Ventas simples por estacion):
	- Se amplÃƒÂ­a `backend/db/carritos_compras.go` para:
		- agregar tabla `empresa_ventas_estacion_metricas` y funciones de registro/resumen de rendimiento por estacion.
		- calcular duracion de atencion por venta y resolver identidad de estacion desde carrito (`referencia_externa`/`codigo`).
	- Se amplÃƒÂ­a `backend/handlers/carritos_compras.go` para:
		- exponer `GET action=metricas_estacion` en `/api/empresa/carritos_compra`.
		- registrar metricas en `pagar_estacion`, `anular_cierre_parcial` y `recuperar_interrumpido`.
	- Se actualiza frontend de ventas simples:
		- `web/administrar_empresa/ventas_simple.html` incorpora panel de sincronizacion offline, metricas de estacion y correccion rapida post-cobro.
		- `web/js/ventas_simple.js` (nuevo) implementa cola offline por estacion con checksum SHA-256 y sincronizacion segura al reconectar.
		- `web/estilos.css` agrega estilos de estado de sincronizacion (`en linea`, `offline`, `sincronizando`).
	- Se amplÃƒÂ­a `backend/handlers/auth_users_carritos_test.go` con `TestEmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones|TestEmpresaCarritosCompraAndItemsFlow" -count=1` (OK).
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).

## 2026-04-07
- Cierre tecnico del modulo 26 (Carritos de compra e items):
	- Se amplÃƒÂ­a `backend/db/carritos_compras.go` para:
		- agregar reintentos transaccionales en operaciones de items frente a bloqueos motor legado retirado (`database is locked/busy`) para fortalecer concurrencia multiestacion.
		- incorporar `RecoverInterruptedCarritoSession` para recuperar carritos interrumpidos sin perdida de items.
		- incorporar `CancelCarritoPartialClosure` para anulacion parcial de cierre en ventas pagadas con validacion estricta de monto.
	- Se amplÃƒÂ­a `backend/handlers/carritos_compras.go` para:
		- exponer `PUT action=recuperar_interrumpido` con trazabilidad en eventos contables y auditoria empresarial.
		- exponer `PUT action=anular_cierre_parcial` con validacion de negocio y auditoria por `empresa_id` y carrito.
	- Se ajusta `web/administrar_empresa/carrito_de_compras.html` para recuperar sesiones interrumpidas sin reset de items y reservar `reset_items=1` solo para sesiones ya pagadas.
	- Se amplÃƒÂ­a cobertura en:
		- `backend/db/carritos_inventario_test.go` (concurrencia de producto, recuperacion interrumpida, anulacion parcial).
		- `backend/handlers/auth_users_carritos_test.go` (recuperacion con auditoria, reglas de pago mixto y anulacion parcial de cierre).
	- Validaciones ejecutadas:
		- `runTests` sobre `backend/db/carritos_inventario_test.go` y `backend/handlers/auth_users_carritos_test.go` -> 36 pruebas OK, 0 fallidas.
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).

## 2026-04-07
- Cierre tecnico del modulo 25 (Panel ERP extendido):
	- Se amplÃƒÂ­a `web/js/modulos_erp_extendido.js` para incorporar:
		- formulario guiado dinamico por modulo (sin dependencia obligatoria de JSON libre),
		- validaciones dinamicas por campo y reglas cruzadas (requeridos, tipos, fechas, rangos y consistencia de montos),
		- acciones rapidas parametrizadas por modulo,
		- guia operativa por dominio con flujo recomendado y controles clave.
	- Se ajusta `web/estilos.css` para reforzar UX del panel ERP:
		- grilla guiada responsive,
		- resaltado de errores en linea,
		- panel visual de validaciones,
		- tarjetas de guia operativa.
	- Validaciones ejecutadas:
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).
		- validacion manual de flujo frontend en `administrar_empresa/modulos_erp_dominio.html` (guiado, validaciones, acciones rapidas y sincronizacion a JSON avanzado).

## 2026-04-07
- Cierre tecnico del modulo 24 (Documental e Integraciones):
	- Se amplÃƒÂ­a `backend/handlers/modulos_faltantes.go` para:
		- reemplazar la ruta generica de documentos por handlers especializados (`EmpresaDocumentosGestionHandler`, `EmpresaDocumentosFirmasHandler`).
		- incorporar versionado documental (`action=versionar`, `action=versiones`) y repositorio con control de acceso por rol/modulo (`action=acceso`, `action=repositorio`).
		- incorporar endurecimiento de integraciones con `action=rotar_credencial` (referencias seguras) y `action=monitoreo`/`action=alertas` (salud de conectores y SLA operativo).
	- Se amplÃƒÂ­a `backend/handlers/empresa_permisos.go` para clasificar `sync_manual`, `rotar_credencial` y `versionar` como acciones criticas de aprobacion en seguridad.
	- Se amplÃƒÂ­a cobertura de pruebas en `backend/handlers/modulos_faltantes_test.go` con:
		- `TestEmpresaIntegracionesAPIsHandlerRotarCredencialYMonitoreo`.
		- `TestEmpresaIntegracionesBancosHandlerRotarCredencial`.
		- `TestEmpresaDocumentosGestionHandlerVersionadoYControlAcceso`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresa(IntegracionesAPIsHandlerRotarCredencialYMonitoreo|IntegracionesBancosHandlerRotarCredencial|DocumentosGestionHandlerVersionadoYControlAcceso)" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 23 (CRM/Produccion/Logistica):
	- Se amplÃƒÂ­a `backend/handlers/modulos_faltantes.go` para incorporar handlers especializados:
		- `EmpresaProduccionOrdenesHandler` con `action=plan_capacidad` (meta diaria, desviaciones y alertas por atraso/sobrecapacidad).
		- `EmpresaLogisticaEnviosHandler` con `action=seguimiento_hitos` (hitos programacion/salida/entrega, SLA y alertas de incumplimiento).
	- Se extiende `backend/handlers/reportes.go` en `operativo_cadena_cumplimiento` con metas y desviaciones por dominio:
		- `meta_cumplimiento_pct`, `desviacion_meta_pct`, `estado_meta`.
		- resumen global `meta_global_pct` y `desviacion_meta_global_pct`.
	- Se amplÃƒÂ­a cobertura de pruebas en:
		- `backend/handlers/modulos_faltantes_test.go` (`TestEmpresaProduccionOrdenesPlanCapacidad`, `TestEmpresaLogisticaEnviosSeguimientoHitos`).
		- `backend/handlers/reportes_test.go` (validaciones de metas/desviaciones en cadena).
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaProduccionOrdenesPlanCapacidad|TestEmpresaLogisticaEnviosSeguimientoHitos|TestEmpresaReportesHandlerDatasetOperativoCadenaCumplimiento" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 22 (RRHH extendido: vacaciones/licencias):
	- Se amplÃƒÂ­a `backend/db/modulos_faltantes.go` con nuevos campos de RRHH en `empresa_rrhh_vacaciones_licencias` para:
		- aprobacion jerarquica (`nivel_aprobacion_actual`, `nivel_aprobacion_requerido`, `aprobadores_json`, `historial_aprobaciones_json`, `fecha_aprobacion_final`),
		- acumulado y saldo (`periodo_acumulado_*`, `saldo_dias_*`, `saldo_snapshot_json`),
		- enlace a nomina (`empleado_nomina_id`, `nomina_liquidacion_id`, `nomina_periodo_*`, `nomina_vinculada_*`).
	- Se amplÃƒÂ­a `backend/handlers/modulos_faltantes.go` con handler especializado `EmpresaRRHHVacacionesLicenciasHandler` y acciones:
		- `action=resumen_saldo` para acumulado/saldo de vacaciones,
		- `action=solicitar_aprobacion`, `action=aprobar`, `action=rechazar` para flujo jerarquico,
		- `action=vincular_nomina` para enlazar novedades aprobadas a liquidacion/periodo de nomina.
	- Se actualiza `backend/handlers/empresa_permisos.go` para mapear acciones RRHH criticas a permisos de aprobacion/actualizacion.
	- Se amplÃƒÂ­a `backend/handlers/modulos_faltantes_test.go` con pruebas de:
		- saldo y aprobacion jerarquica multinivel,
		- vinculacion de novedades RRHH a nomina por periodo.
	- Validaciones ejecutadas:
		- `go test ./handlers -run RRHH -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 21 (Inventario extendido: lotes/series y devolucion proveedor):
	- Se amplÃƒÂ­a `backend/db/modulos_faltantes.go` con:
		- trazabilidad completa de lotes/series mediante tabla `inventario_lotes_series_movimientos`.
		- campos operativos de bloqueo/estado en `inventario_lotes_series` (`reservado_cantidad`, `vendido_cantidad`, `bloqueado_venta`, `bloqueo_motivo`, `ultima_operacion_*`).
		- campos contables de devolucion en `empresa_devoluciones_proveedor` (`periodo_contable`, `impacto_contable_*`, `fecha_contabilizacion`, `contabilizado_por`, `total_reintegrado`).
	- Se amplÃƒÂ­a `backend/handlers/modulos_faltantes.go` con handlers especializados:
		- `EmpresaInventarioLotesSeriesHandler` con acciones `trazabilidad`, `validar_disponibilidad`, `reservar`, `vender`, `liberar_reserva`, `ajuste_entrada`, `ajuste_salida`, `devolucion_proveedor`.
		- bloqueo automatico por vencimiento en venta/reserva y actualizacion de estado de lote.
		- `EmpresaComprasDevolucionesProveedorHandler` con `action=contabilizar`/`action=impacto_contable` para generar movimiento financiero, evento contable y actualizar la devolucion a `contabilizada`.
	- Se amplÃƒÂ­a `backend/db/eventos_contables.go` para soportar `devolucion_proveedor_contabilizada` en contrato y asiento contable (flujo de ingreso).
	- Se amplÃƒÂ­a `backend/handlers/modulos_faltantes_test.go` con pruebas de:
		- bloqueo automatico de lote vencido en reserva,
		- trazabilidad de ciclo reserva/venta/liberacion,
		- contabilizacion completa de devolucion proveedor con impacto contable.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaInventarioLotesSeriesBloqueoAutomaticoVencido|TestEmpresaInventarioLotesSeriesTrazabilidadCicloVenta|TestEmpresaComprasDevolucionesProveedorContabilizarImpactoCompleto" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-06
- Cierre tecnico del modulo 20 (Contabilidad operativa extendida: plan de cuentas, CxC y CxP):
	- Se amplÃƒÂ­a `backend/handlers/modulos_faltantes.go` con handlers especializados de finanzas:
		- `EmpresaFinanzasPlanCuentasHandler` con `action=plantillas`, `action=aplicar_plantilla` y `action=validar_cierre_periodo`.
		- `EmpresaFinanzasCuentasCobrarHandler` y `EmpresaFinanzasCuentasPagarHandler` con `action=conciliar_pagos` y validacion de periodo cerrado.
	- Se amplÃƒÂ­a `backend/db/modulos_faltantes.go` con:
		- nuevos metadatos de plantilla en `empresa_plan_cuentas`.
		- campos de conciliacion en `empresa_cuentas_por_cobrar` y `empresa_cuentas_por_pagar`.
		- bloqueo retroactivo por periodo contable cerrado en crear/editar/cambiar estado/eliminar de CxC/CxP.
	- Se amplÃƒÂ­a `backend/handlers/modulos_faltantes_test.go` con pruebas de:
		- plantillas y aplicacion de plan de cuentas por tipo de empresa.
		- conciliacion automatica CxC contra pagos reales.
		- bloqueo de operaciones CxP cuando el periodo esta cerrado.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaFinanzasPlanCuentasPlantillasYAplicacion|TestEmpresaFinanzasCuentasCobrarConciliacionPagosReales|TestEmpresaFinanzasCarteraBloqueoPeriodoCerrado" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).
- Cierre tecnico del modulo 19 (Gestion comercial extendida: cotizaciones/pedidos/devoluciones):
	- Se amplÃƒÂ­a `backend/handlers/modulos_faltantes.go` con automatizacion comercial en ventas:
		- `POST/PUT action=convertir_pedido` en cotizaciones para convertir cotizacion aprobada/emitida a pedido trazable (`cotizacion_id`, `convertido_pedido_id`).
		- `POST/PUT action=convertir_documento_final` en cotizaciones y pedidos para generar documento final en `empresa_facturacion_documentos`.
		- `GET action=embudo` en cotizaciones para monitoreo operativo con SLA y alertas de vencimiento.
	- Se incorpora snapshot de embudo comercial cotizacionÃ¢â€ â€™pedidoÃ¢â€ â€™documento final con trazabilidad por `empresa_id`.
	- Se agrega dataset exportable `operativo_ventas_embudo_conversion` en `backend/handlers/reportes.go` con formatos `json/csv/txt/xls/pdf`.
	- Se actualiza `backend/handlers/empresa_permisos.go` para clasificar `convertir_pedido` y `convertir_documento_final` como acciones de aprobacion en ventas.
	- Se agregan pruebas en `backend/handlers/modulos_faltantes_test.go` y `backend/handlers/reportes_test.go` para:
		- conversion cotizacionÃ¢â€ â€™pedidoÃ¢â€ â€™documento final,
		- alertas SLA del embudo,
		- dataset/export CSV del nuevo reporte de conversion.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaVentasCotizacionesConversionPedidoYDocumentoFinal|TestEmpresaVentasCotizacionesEmbudoYAlertasSLA|TestEmpresaReportesHandlerDatasetOperativoVentasEmbudoConversion" -count=1` (OK).
		- `go test ./handlers -run "DIAN|ModulosFaltantes|OperativoCadenaCumplimiento|OperativoVentasEmbudoConversion" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).
- Cierre tecnico del modulo 18 (Facturacion electronica DIAN Colombia):
	- Se amplÃƒÂ­a `backend/handlers/modulos_faltantes.go` en `EmpresaDIANColombiaHandler` con acciones operativas reales:
		- `action=firmar_xml_real` (firma RSA-SHA256).
		- `action=enviar_documento_real` (envio productivo/habilitacion por `url_dian`).
		- `action=consultar_acuse_real` (consulta de acuse y normalizacion de estado).
		- `action=reconexion_dian` (sondeo de conectividad y salida de contingencia).
	- Se implementa gestion segura de credenciales/certificados por referencia:
		- `token_emisor_ref` y `certificado_clave_ref` soportan `env:`, `file:` y `base64:`.
	- Se integra transicion de estado DIAN (`pendiente/enviado/aceptado/rechazado/contingencia/reconectado`) con trazabilidad en `observaciones` y `ultimo_envio`.
	- Se ajusta `backend/handlers/empresa_permisos.go` para clasificar las nuevas acciones DIAN de escritura como `permActionApprove`.
	- Se agregan pruebas en `backend/handlers/modulos_faltantes_test.go` para:
		- flujo firma + envio + acuse exitoso,
		- contingencia por falla de transporte y recuperacion por reconexion.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "DIAN|ModulosFaltantes" -count=1 -v` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (suite backend completa OK).
- Estabilizacion del panel de graficos y estadisticas (compras):
	- Se ajusta `backend/handlers/graficos_estadisticas.go` para soportar la nueva estructura del dataset `operativo_compras_movimientos` (agregado por proveedor) sin perder compatibilidad con la forma anterior.
	- Se agrega fallback para construir la serie de compras desde movimientos financieros (`egresos` de compras) cuando no existen documentos en `empresa_compras_documentos`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run TestEmpresaGraficosEstadisticasHandlerPanelYAcciones -count=1 -v` (OK tras ajuste).
		- `runTests` en `backend/handlers/graficos_estadisticas_test.go` (2 pruebas OK).
		- `go test ./handlers -run GraficosEstadisticas -count=1` (OK).
		- `go test ./... -count=1` (suite backend completa OK).
- Cierre tecnico del modulo 17 (Facturacion electronica):
	- Se integra despacho fiscal por pais/proveedor en `backend/handlers/facturacion_electronica.go`:
		- proveedor `manual` (productivo local), `mock://` para pruebas y despacho HTTP contra `api_base_url` cuando aplica.
		- respuesta operativa `integracion_fiscal` en acciones transaccionales (`emitir`, `anular`, `nota_credito`).
	- Se implementa cola de reintentos FE en `backend/db/facturacion_electronica.go`:
		- nueva tabla `facturacion_electronica_reintentos` con estado de envio, intentos, proximo intento, contingencia y referencia externa.
		- nuevas operaciones de consulta/actualizacion y listados filtrados por estado.
	- Se agregan endpoints operativos FE:
		- `GET action=reintentos`.
		- `POST/PUT action=procesar_reintentos`.
		- `GET action=reconciliacion` y `POST/PUT action=reconciliar_estados`.
	- Se activa contingencia automatica al superar `max_intentos` y se conserva numeracion legal por resolucion en emision.
	- Se actualiza contrato contable del modulo `facturacion` con eventos de integracion (`factura_integracion_enviada`, `factura_integracion_fallida`, `factura_contingencia_activada`).
	- Validaciones ejecutadas:
		- `go test ./db -run Facturacion -count=1` (OK).
		- `go test ./handlers -run FacturacionElectronica -count=1` (OK).
		- `go test ./handlers -run Facturacion -count=1` (OK).
		- `go test ./... -count=1` (falla no relacionada en `TestEmpresaGraficosEstadisticasHandlerPanelYAcciones`).
- Cierre tecnico del modulo 16 (Compras):
	- Se amplÃƒÂ­a el ciclo documental de compras con aprobacion multinivel:
		- `requiere_aprobacion`, `niveles_aprobacion_requeridos`, `nivel_aprobacion_actual`, `aprobadores_json`.
		- Nuevas acciones: `solicitar_aprobacion`, `aprobar_compra`, `rechazar_compra`.
	- Se cierra recepcion parcial avanzada por item:
		- `recepcion_detalle_json` y `recepcion_resumen_json` para registrar cantidades solicitadas/recibidas, pendientes y diferencias.
		- Nueva accion: `recepcionar_parcial_compra`, consolidada con `recepcionar_compra` al completar recepcion total.
	- Se integra validacion documental proveedor-factura-entrada:
		- `validacion_documental_estado`, `proveedor_documento_ref`, `factura_documento_ref`, `entrada_documento_ref`.
		- Nueva accion: `validar_documentos` con verificacion de proveedor y referencias documentales.
	- Se amplÃƒÂ­a UI en `web/administrar_empresa/compras.html` con campos, filtros/KPI y acciones operativas del nuevo flujo.
	- Validaciones ejecutadas:
		- `runTests` en `backend/db/documentos_transaccionales_test.go`, `backend/handlers/compras_documentos_test.go`, `backend/handlers/empresa_permisos_test.go` (21 pruebas OK).
		- `go test ./... -run TestDoesNotExist -count=1` (compilacion global backend OK).
- Hotfix de compatibilidad de migraciones legacy en startup:
	- Se corrige el orden de migracion en `EnsureEmpresaPropinasSchema` para crear indices despues de asegurar columnas faltantes (`cierre_caja_id` y relacionadas), evitando fallos en bases antiguas.
	- Se corrige el orden de migracion en `EnsureEmpresaComisionesServicioSchema` para crear indices despues de asegurar columnas faltantes (`ajuste_manual` y relacionadas), evitando fallos en bases antiguas.
	- Resultado operativo: el script `scripts/iniciar_servidor.ps1` vuelve a iniciar correctamente y el backend queda escuchando en `:8080`.
	- Validaciones ejecutadas:
		- `go test ./db -run "Propina|Comision" -count=1` (OK).
		- `go test ./handlers -run "Propina|Comision" -count=1` (OK).
		- `go test ./... -run TestDoesNotExist -count=1` (compilacion global backend OK).
- Cierre tecnico del modulo 15 (Comisiones por servicio):
	- Se amplÃƒÂ­a el modelo de comisiones con escalas por rol/servicio y tope por item:
		- nueva tabla `empresa_comisiones_servicio_escalas` (`rol_operacion`, `servicio_filtro`, `porcentaje_comision`, `tope_comision`, `prioridad`).
	- Se amplÃƒÂ­a `empresa_comisiones_servicio_movimientos` con trazabilidad operativa:
		- `rol_operacion`, `escala_id`, `monto_comision_bruto`, `tope_comision_aplicado`,
		- `origen_movimiento`, `ajuste_manual`, `referencia_ajuste`, `ajuste_estado`, `aprobado_por`, `aprobado_en`,
		- `liquidacion_nomina_id`, `periodo_liquidacion_desde`, `periodo_liquidacion_hasta`, `liquidado_en`, `liquidado_por`.
	- Se incorporan endpoints/acciones de comisiones para operacion avanzada:
		- escalas (`escalas`, `escala`, `activar_escala`, `desactivar_escala`),
		- ajuste manual y aprobacion (`ajuste_manual`, `aprobar_ajuste`, `rechazar_ajuste`),
		- resumen para nomina (`resumen_liquidacion`).
	- Se integra nomina con comisiones:
		- `empresa_nomina_liquidaciones` incorpora `comisiones_servicio_total`, `comisiones_servicio_movimientos`, `comisiones_servicio_ajustes`.
		- el calculo de liquidacion integra comisiones y enlaza movimientos al periodo liquidado.
	- Se amplia `web/administrar_empresa/comisiones.html` para operacion completa del modulo 15:
		- gestion de escalas/topes,
		- registro de ajuste manual,
		- aprobacion/rechazo de ajustes pendientes,
		- filtros avanzados de reporte y consulta de `resumen_liquidacion`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestEmpresaComisionesServicio|TestEmpresaNominaLiquidacionIntegraComisionesServicio" -count=1` (OK).
		- `go test ./handlers -run "TestEmpresaComisionesServicioHandler" -count=1` (OK).
		- `go test ./... -run TestDoesNotExist -count=1` (compilacion global backend OK).
- Avance y cierre tecnico del modulo 14 (Propinas):
	- Se amplÃƒÂ­a la configuracion empresarial de propinas con reglas fiscales:
		- `pais_fiscal`, `regimen_fiscal`, `tratamiento_fiscal` (`no_gravada`/`gravada`) y `porcentaje_impuesto_propina`.
	- Se amplÃƒÂ­a el libro de movimientos de propinas con:
		- `origen_movimiento` (`venta`/`ajuste_manual`),
		- `ajuste_manual`, `referencia_ajuste`, `cierre_caja_id`, `conciliado_en`,
		- snapshot fiscal por movimiento (`fiscal_*`).
	- Se incorpora conciliacion de propinas contra cierre de caja:
		- accion manual `action=conciliacion_cierre` en propinas,
		- integracion automatica durante transiciones `cerrar/aprobar` de cierre de caja,
		- persistencia de resumen en `empresa_cierres_caja` (`propinas_movimientos`, `propinas_total`, `propinas_ajustes`, `propinas_impuesto`, `propinas_neto`, `propinas_conciliado_*`).
	- Se incorpora ajuste manual auditado de propinas:
		- accion `action=ajuste_manual`,
		- registro no bloqueante en `empresa_auditoria_eventos`.
	- Se actualiza frontend `web/administrar_empresa/propinas.html` con:
		- configuracion fiscal,
		- formulario de ajuste manual,
		- accion de conciliacion por cierre,
		- filtros y columnas extendidas en el reporte.
	- Se agrega cobertura de pruebas para flujo de ajuste y conciliacion:
		- `backend/handlers/propinas_test.go`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Propinas|Cierre" -count=1` (OK).
		- `go test ./db -run "Propina|CierreCaja|Finanzas" -count=1` (OK).
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).
- Avance del modulo 13 (Codigos de descuento avanzados):
	- Se amplÃƒÂ­a `codigos_de_descuento` con reglas contextuales:
		- `segmento_cliente`, `canal_venta`, `horario_desde`, `horario_hasta`, `dias_semana`.
	- Se incorpora antifraude por cliente:
		- `max_usos_por_cliente`, `ventana_horas_fraude`.
	- Se agrega trazabilidad de redenciones en nueva tabla `codigos_descuento_redenciones` con estados:
		- `aplicada`, `revertida`, `anulada`.
	- Se integra ciclo de redencion con carritos:
		- aplica al cerrar carrito,
		- revierte al reabrir,
		- anula al eliminar carrito.
	- Se extiende API de codigos:
		- validacion contextual (`action=validar` con `carrito_id`, `cliente_id`, `canal_venta`),
		- consulta de trazabilidad (`action=redenciones`).
	- Se actualiza `web/administrar_empresa/codigos_de_descuento.html` para administrar reglas avanzadas y antifraude.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestCodigoDescuento" -count=1` (OK).
		- `go test ./handlers -run "TestNoExiste" -count=1` (OK, compilacion handlers).
		- `go test ./db -run "TestCarritoProductoDescuentaInventarioYVentaMantieneStock|TestCarritoStockNoSeDuplicaAlReactivarSesionCerrada" -count=1` (OK).
		- `go test ./...` (falla en prueba no relacionada: `TestEmpresaGraficosEstadisticasHandlerPanelYAcciones`).
- Validacion final de continuidad tecnica y documental (post-ediciones recientes):
	- Se revalida compilacion global de backend tras ajustes recientes en inventario/combos.
	- Resultado: `go test ./... -run TestDoesNotExist -count=1` en `backend` -> OK (sin errores de compilacion).
	- Se confirma sincronizacion de cierre de modulos 1-12 en documentacion operativa y tecnica.
- Cierre del modulo 12 (Combos de productos):
	- Se implementa versionado de receta por combo:
		- nuevas columnas en `combos_productos`: `receta_version`, `costo_teorico`, `costo_real`, `variacion_costo`, `variacion_costo_porcentaje`.
		- nueva tabla `combos_productos_versiones` para snapshots historicos de ingredientes por version.
	- Se incorpora validacion de costo teorico vs costo real de ingredientes en create/update de combos:
		- bloqueo si la variacion porcentual supera el umbral operativo.
		- bloqueo si el precio del combo no cubre el costo real calculado.
	- Se endurece concurrencia de inventario en carritos:
		- reserva de stock con `UPDATE` atomico condicionado (`cantidad >= requerida`) para evitar sobreventa en ventas simultaneas.
	- Se actualiza frontend `web/administrar_empresa/combos_productos.html` para mostrar version de receta y metricas de costo.
	- Validaciones ejecutadas:
		- `runTests` sobre `backend/db/productos_categorias_test.go` y `backend/db/carritos_inventario_test.go`.
		- `go test ./... -run TestDoesNotExist -count=1`.
- Verificacion final de continuidad del modulo 11 (Inventario):
	- Se ejecuta compilacion global posterior a cambios recientes en archivos de inventario con `go test ./... -run TestDoesNotExist -count=1` (OK).
	- Se confirma cierre operativo completo del checklist de modulo 11 (schema, costos, conteo ciclico, alertas proactivas y documentacion).
- Cierre del modulo 11 (Inventario) de Fase 3:
	- Se implementa configuracion de politica de costo por empresa:
		- `GET/PUT /api/empresa/inventario/configuracion`.
		- Politicas soportadas: `promedio` y `peps`.
	- Se incorpora soporte de capas/lotes de costo para trazabilidad de salidas y transferencias:
		- tabla `inventario_costos_lotes`.
		- salida con PEPS por capas y recalculo de costo promedio por bodega/producto.
	- Se implementa conteo ciclico con ajuste auditado:
		- `GET/POST /api/empresa/inventario/conteo_ciclico`.
		- tabla `inventario_conteos_ciclicos` y movimiento automatico `ajuste_positivo/ajuste_negativo` cuando hay variacion.
	- Se cierran alertas operativas proactivas de inventario:
		- `GET /api/empresa/inventario/alertas?action=proactivas`.
		- incorpora `sobrestock`, `deficit`, `exceso` y `accion_sugerida`.
	- Se actualiza frontend `web/administrar_empresa/administrar_productos.html` con:
		- selector/guardado de politica de costo,
		- formulario y tabla de conteo ciclico,
		- visualizacion de alertas proactivas (quiebre/sobrestock).
	- Validaciones ejecutadas:
		- `go test ./db -run "TestInventarioPoliticaCostoPromedioYPEPS|TestRegistrarConteoCiclicoInventarioAjustaYAudita|TestGetAlertasOperativasByEmpresaIncluyeSobrestock" -count=1`.
		- `go test ./handlers -run "TestEmpresaInventarioConfiguracionYConteoCiclicoHandler|TestEmpresaInventarioAlertasHandlerProactivasIncluyeSobrestock" -count=1`.
		- `go test ./... -run TestDoesNotExist -count=1`.
- Cierre del modulo 10 (Clientes) de Fase 3:
	- Se implementa deduplicacion por `documento`, `correo` y `telefono` en `create/update` de clientes por `empresa_id`.
	- El endpoint `POST/PUT /api/empresa/clientes` responde `409` cuando detecta conflicto de deduplicacion, con mensaje de campo duplicado.
	- Se agrega dataset operativo para exportacion masiva comercial:
		- `operativo_clientes_segmentacion_comercial` en `/api/empresa/reportes`.
		- Incluye segmento, metricas de compra y `accion_comercial_sugerida` por cliente.
		- Exportacion disponible en `pdf/xls/csv/json/txt`.
	- Se actualiza frontend `web/administrar_empresa/administrar_clientes.html` con panel de exportacion masiva por segmento.
	- Validaciones ejecutadas:
		- `go test ./db -run "Test(CreateClienteDeduplicacionDocumentoCorreoTelefono|UpdateClienteDeduplicacionCorreoTelefono|GetClientePerfilComercialByEmpresaAndHistorial|GetClientePerfilComercialByEmpresaSinComprasSegmentoNuevo|GetClienteByID)$" -count=1`.
		- `go test ./handlers -run "Test(EmpresaClientesHandlerPerfilHistorialSegmentacion|EmpresaClientesHandlerConflictosDeduplicacion|EmpresaReportesHandlerDatasetOperativoClientesSegmentacionComercial)$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 9 (Tarifas por dia) de Fase 3:
	- Se implementa prorrateo de tarifa diaria por ventana de `check-in/check-out` para entrada/salida fuera de ventana.
	- Se extiende simulador de `GET /api/empresa/tarifas_por_dia?action=calcular` con detalle de:
		- `dias_completos`, `dias_equivalentes`,
		- `monto_dias_completos`, `monto_prorrateo_(entrada|intermedio|salida)`,
		- `minutos_prorrateo_fuera_ventana`.
	- Se agrega aplicacion masiva de una misma tarifa diaria a todas las estaciones detectadas:
		- `PUT /api/empresa/tarifas_por_dia?action=aplicar_todas_estaciones`.
	- Se agrega reporte operativo comparativo por estacion:
		- dataset `operativo_tarifas_comparativo_estaciones` en `/api/empresa/reportes`,
		- comparativo de ingreso esperado (motor prorrateado) vs ingreso real cobrado,
		- exportacion en `pdf/xls/csv/json/txt`.
	- Se actualiza frontend `web/administrar_empresa/tarifas_por_dia.html` con:
		- boton `Aplicar a todas las estaciones`,
		- simulador con desglose de prorrateo,
		- panel de descarga del comparativo esperado vs real.
	- Validaciones ejecutadas:
		- `go test ./db -run "TarifaPorDia|ApplyEmpresaTarifaPorDiaToAllStations|EmpresaTarifasPorDia"`.
		- `go test ./handlers -run "TarifasPorDia|CarritosCompraListIncluyeTarifaPorDiaAutomatica|OperativoTarifasIngresos|OperativoTarifasComparativoEstaciones"`.
		- `go test ./... -run "^$"`.
- Cierre del modulo 8 (Tarifas por minutos) de Fase 3:
	- Se agrega configuracion empresarial avanzada de calculo:
		- `redondeo_modo` (`ninguno`, `arriba`, `abajo`, `matematico`),
		- `redondeo_unidad`,
		- `monto_minimo_diario`,
		- `monto_maximo_diario`.
	- Se extiende simulador de cobro por minutos con detalle de:
		- monto base, monto extra, subtotal, monto redondeado y ajuste,
		- aplicacion de minimo/maximo diario,
		- soporte de minutos fraccionarios (`minutos_consumidos` decimal).
	- Se cierra trazabilidad contable del calculo por minutos:
		- registro de evento `finanzas.tarifa_por_minutos_calculada` en `empresa_eventos_contables`,
		- respuesta de simulacion con `trazabilidad_contable_id`, `documento_codigo` y `periodo_contable`.
	- Se agrega aplicacion masiva de una misma regla de tarifa a todas las estaciones detectadas:
		- `PUT /api/empresa/tarifas_por_minutos?action=aplicar_todas_estaciones`.
	- Se actualiza frontend `web/administrar_empresa/tarifas_por_minutos.html` con:
		- panel de configuracion avanzada de redondeo y topes,
		- boton `Aplicar a todas las estaciones`,
		- simulador con detalle de calculo y referencia contable.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestEmpresaTarifasPorMinutos|TestApplyEmpresaTarifaPorMinutosToAllStations|TestRegisterTarifaPorMinutosCalculoContable|TestEmpresaEventosContables" -count=1`.
		- `go test ./handlers -run "TestEmpresaTarifasPorMinutosHandler" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 7 (Reservas por estacion/habitacion) de Fase 3:
	- Se refuerza control de concurrencia anti-overbooking por estacion en ventanas solapadas:
		- validacion de conflicto por `estacion_id` y `carrito_id` asociado,
		- bloqueo para estados operativos `pendiente_pago`, `confirmada` y `en_curso`.
	- Se implementa politica automatica avanzada de reservas:
		- expiracion de pendientes por `fecha_expiracion` y fallback por antiguedad de creacion,
		- marcacion automatica de `no_show` sobre reservas confirmadas fuera de tolerancia operativa,
		- accion de sincronizacion: `GET /api/empresa/reservas_hotel?action=aplicar_politicas`.
	- Se incorpora reconversion operativa de reserva a carrito:
		- `PUT /api/empresa/reservas_hotel?action=convertir_carrito`.
		- transicion de reserva a estado `en_curso` y activacion de carrito asociado.
	- Se actualiza frontend `web/administrar_empresa/reservas_hotel.html` con:
		- accion `Aplicar politicas`,
		- accion `Reconver. carrito`,
		- filtros extendidos para estados `en_curso` y `no_show`.
	- Validaciones ejecutadas:
		- `go test ./db -run "TestReservaHotel(FlowCRUDAndDisponibilidad|MultiEstacionNoOverbookingYReconversion|PoliticaNoShowYExpiracionAvanzada)$" -count=1`.
		- `go test ./handlers -run "TestEmpresaReservasHotelHandler(CRUDAndDisponibilidad|PoliticasYReconversion)$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 6 (Registro de vehiculos) de Fase 3:
	- Se agrega configuracion de validacion de placa/patente por empresa y pais:
		- `GET/PUT /api/empresa/vehiculos_registro?action=config`.
		- Tabla `empresa_vehiculos_configuracion` con `pais_codigo`, `patente_regex`, `patente_descripcion`, `evitar_duplicado_activo`.
	- Se implementa bloqueo de duplicidad activa por patente canonica en patio/empresa:
		- validado en crear, editar y activar registros de vehiculos.
		- respuesta HTTP `409` ante conflicto de duplicidad activa.
	- Se agrega reporte operativo de permanencia y tiempos de estancia:
		- `GET /api/empresa/vehiculos_registro?action=permanencia`.
		- dataset `operativo_vehiculos_permanencia` en `/api/empresa/reportes` con exportacion `pdf/xls/csv/json/txt`.
	- Se integra frontend en `web/administrar_empresa/vehiculos_registro.html`:
		- panel de configuracion de formato de placa por pais,
		- consulta visual de permanencia,
		- exportacion de reporte en formatos estandar.
	- Validaciones ejecutadas:
		- `go test ./db -run TestEmpresaVehiculoRegistroConfigValidacionDuplicidadYPermanencia -count=1`.
		- `go test ./handlers -run TestEmpresaVehiculosRegistroHandlerConfigYReportePermanencia -count=1`.
		- `go test ./handlers -run TestEmpresaReportesHandlerDatasetOperativoVehiculosPermanencia -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 5 (Nomina de sueldos) de Fase 3:
	- Se agregan operaciones nuevas en nomina:
		- `GET /api/empresa/nomina?action=desprendible&empleado_nomina_id={id}&periodo_desde=YYYY-MM-DD&periodo_hasta=YYYY-MM-DD`.
		- `GET /api/empresa/nomina?action=conciliacion_asistencia` (auditoria sin cambios).
		- `POST /api/empresa/nomina?action=conciliar_asistencia` (auditoria con opcion de auto-recalculo).
	- Se implementa desprendible estandar por empleado y periodo con detalle de horas, devengados, deducciones y neto a pagar.
	- Se implementa conciliacion automatica entre asistencia y liquidacion final:
		- detecta diferencias de registros/horas,
		- identifica asistencias sin liquidacion,
		- permite recalcular/crear liquidaciones inconsistentes cuando `auto_recalcular=true`.
	- Se integra frontend en `web/administrar_empresa/nomina_sueldos.html`:
		- boton de conciliacion con modo auditoria o auto-recalculo,
		- generacion/visualizacion de desprendible por empleado-periodo,
		- accion de desprendible desde tabla de liquidaciones.
	- Se documentan y validan casos de formula por pais/empresa (CO/MX + override por empresa) con pruebas automatizadas.
	- Validaciones ejecutadas:
		- `go test ./db -run "Test(EmpresaNominaGenerateLiquidacionesFromAsistencia|EmpresaNominaCalculoPorPaisYEmpresa|EmpresaNominaDesprendibleYConciliacionAsistencia)$" -count=1`.
		- `go test ./handlers -run "TestEmpresaNominaSueldosHandlerFlow$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 4 (Asistencia de empleados) de Fase 5:
	- Se implementa cierre de periodo con bloqueo operativo de edicion posterior:
		- `POST /api/empresa/asistencia_empleados?action=cerrar_periodo`.
		- `GET /api/empresa/asistencia_empleados?action=periodos_cerrados`.
	- Se agrega configuracion por empresa para tolerancias y reglas de turno:
		- `GET/PUT /api/empresa/asistencia_empleados?action=config`.
		- `tolerancia_entrada_minutos`, `hora_inicio_turno_(manana|tarde|noche)`, `permitir_turno_nocturno`, `permitir_turno_cruzado`.
	- Se incorporan validaciones de negocio en asistencia:
		- bloqueo de create/update/delete/activar/desactivar/marcar_entrada/marcar_salida cuando la fecha pertenece a periodo cerrado,
		- rechazo de turno nocturno o cruzado cuando la configuracion empresarial lo deshabilita,
		- calculo de tardanza con tolerancia configurable.
	- Se publica reporte operativo de auditoria para nomina:
		- dataset `operativo_asistencia_nomina_auditoria` en `/api/empresa/reportes` con exportacion `pdf/xls/csv/json/txt`.
	- Se integra frontend en `web/administrar_empresa/asistencia_empleados.html`:
		- panel de configuracion,
		- cierre de periodo y listado de cierres,
		- descarga del reporte de auditoria de nomina.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Test(EmpresaAsistenciaEmpleadosHandlerCRUDFlow|EmpresaAsistenciaEmpleadosHandlerConfigTurnosYTolerancia|EmpresaAsistenciaEmpleadosHandlerCierrePeriodoBloqueaEdicion|EmpresaReportesHandlerDatasetOperativoAsistenciaNominaAuditoria)$" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 3 (Usuarios de empresa) de Fase 1:
	- Se agrega cambio autogestionado de contraseÃƒÂ±a para usuario empresa:
		- `POST /api/empresa/usuarios/cambiar_password`.
	- Se implementan politicas de contraseÃƒÂ±a configurables desde `configuraciones`:
		- `usuarios.password_min_length`
		- `usuarios.password_require_uppercase`
		- `usuarios.password_require_lowercase`
		- `usuarios.password_require_digit`
		- `usuarios.password_require_symbol`
		- `usuarios.password_rotation_days`.
	- El login de usuario empresa ahora devuelve `password_rotation_required` cuando aplica rotacion obligatoria.
	- Se incorpora captura de notificaciones de confirmacion/restablecimiento en entorno de pruebas de correo:
		- tabla `super_correo_notificaciones_prueba` en `pcs_superadministrador`.
		- activacion por `PCS_MAIL_TEST_MODE=1` o `gmail.smtp_test_mode=1`.
	- Se integra frontend de autogestion en `web/login_usuario.html` y `web/js/login_usuario.js`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Test(EmpresaUsuarioChangePasswordFlow|EmpresaUsuarioChangePasswordPolicyRejectsWeakPassword|EmpresaUsuarioLoginRequiresRotationWhenPolicyEnabled|EmpresaUsuarioNotificationsCaptureInMailTestMode|EmpresaUsuarioPasswordRecoveryFlow)" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 2 (Administracion global super) de Fase 1:
	- Se implementa desactivacion/rehabilitacion de empresa con validaciones de impacto operativo y confirmacion forzada cuando existen bloqueos:
		- `GET /super/api/empresas?id={id}&action=impacto_desactivacion`.
		- `PUT /super/api/empresas?id={id}&action=desactivar[&force=1]`.
		- `PUT /super/api/empresas?id={id}&action=activar&activo=1`.
	- Se agrega respaldo/restauracion de configuracion critica super:
		- `GET /super/api/config/backup` (exporta JSON).
		- `PUT /super/api/config/backup` (restaura JSON).
	- Se integra operacion desde frontend:
		- `web/js/seleccionar_empresa.js` para desactivar/reactivar con consulta de impacto.
		- `web/super/configuracion_avanzada.html` con descarga y restauracion de respaldo.
	- Se agregan pruebas de permisos y flujo super en `backend/handlers/system_empresas_handlers_test.go`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "Test(EmpresasHandlerDesactivarConImpactoYForce|EmpresasHandlerImpactoDesactivacion|SuperConfigBackupHandlerExportYRestore|SuperEndpointsPermisosPorRol)" -count=1`.
		- `go test ./... -run "^$" -count=1`.
- Cierre del modulo 1 (Autenticacion y sesiones) de Fase 1:
	- Se implementa bloqueo temporal por intentos fallidos en login de usuario empresa.
	- Se agrega recuperacion de contrasena para usuario empresa con token temporal:
		- `POST /api/empresa/usuarios/solicitar_recuperacion_password`
		- `POST /api/empresa/usuarios/restablecer_password`
	- Se endurece seguridad de sesion:
		- sesiones nuevas con `fecha_fin` por expiracion (24h),
		- revocacion de token en logout,
		- middleware bloquea tokens expirados o revocados.
	- Se habilita flujo frontend de recuperacion en `web/login_usuario.html` y `web/js/login_usuario.js`.
	- Validaciones ejecutadas:
		- `runTests` sobre `backend/handlers/auth_users_carritos_test.go` -> 24/24.
		- `go test ./... -run "^$" -count=1` (compilacion global OK).
- Cierre tecnico backend de pasarela unica Wompi:
	- Se elimina remanente de Mercado Pago en backend:
		- `backend/handlers/payments_handlers.go`: retiro de handlers/utilidades Mercado Pago.
		- `backend/db/db.go`: retiro de tipo/funciones de persistencia `pagos_mercadopago`.
		- `backend/main.go`: retiro de bootstrap/migracion de `pagos_mercadopago`.
		- `backend/utils/utils.go`: retiro del prefijo `/mercadopago/` en manejo JSON API.
		- `backend/tools/query_users/main.go`: migracion de inspeccion local hacia `wompi.*` y `pagos_wompi`.
	- Se sincroniza documentacion tecnica con el estado real:
		- `documentos/estructura_bd.md` y `estructura_bd.md`.
		- `documentos/diagramas/estructura_del_codigo.md`.
		- `documentos/descripcion_de_archivos`.
	- Validacion tecnica ejecutada: `go test ./... -run "^$" -count=1` (compilacion global OK).
- Cierre de pendientes de modulos:
	- Se valida la matriz de estado de modulos/reportes y no quedan modulos marcados como incompletos (`Pendiente` o `Parcial`) en `documentos/modulos del proyecto.md`.
	- Se actualiza `Pendiente Notas` marcando como completado el pendiente de pasarela unica Wompi.
- Pasarela de pago unificada en Wompi:
	- Se retira la configuraciÃƒÂ³n de Mercado Pago de `web/super/configuracion_avanzada.html` y se deja ÃƒÂºnicamente la secciÃƒÂ³n de credenciales de Wompi en configuraciÃƒÂ³n avanzada del panel super administrador.
	- Se simplifica `web/pagar_licencia.html` eliminando selector/panel/flujo de Mercado Pago para operar solo con Nequi (Wompi) y activaciÃƒÂ³n manual interna.
	- Se desregistran rutas de Mercado Pago en `backend/main.go` (`/super/api/config/mercadopago`, `/mercadopago/create_preference`, `/mercadopago/webhook`, `/mercadopago/reconcile`, `/mercadopago/test_preference`).
	- ValidaciÃƒÂ³n tÃƒÂ©cnica ejecutada: `go test ./... -run "^$" -count=1` (compilaciÃƒÂ³n global OK).
- Cierre de trazabilidad y validacion final del plan de reportes:
	- Se revalida la presencia de los datasets operativos de cierre (`operativo_propinas_acumulado`, `operativo_comisiones_lavador`, `operativo_facturacion_trazabilidad`, `operativo_auditoria_acciones`) en `backend/handlers/reportes.go`.
	- Se ejecuta validacion completa de `backend/handlers/reportes_test.go` con resultado `16/16` pruebas aprobadas.
	- Se confirma consistencia de estado documental en `documentos/modulos del proyecto.md`, `CHANGELOG.md` y `documentos/historial_de_cambios`.
	- Se deja cerrado el pendiente de trazabilidad del plan secuencial.
- Plan secuencial de cierre de modulos incompletos - bloques 6 a 9 (Propinas, Comisiones, Facturacion y Auditoria):
	- Se agregan en `backend/handlers/reportes.go` cuatro datasets operativos nuevos:
		- `operativo_propinas_acumulado` (acumulado por usuario, distribucion directa/universal y participacion),
		- `operativo_comisiones_lavador` (acumulado por lavador con base de servicios y ticket de comision),
		- `operativo_facturacion_trazabilidad` (emitidas/anuladas/pendientes y trazabilidad legal por tipo documental),
		- `operativo_auditoria_acciones` (eventos por modulo/usuario con errores HTTP y acciones criticas).
	- Se actualiza catalogo y switch de construccion de datasets para incluir estos cuatro reportes en suite/export.
	- Se amplia `backend/handlers/reportes_test.go` con pruebas dedicadas:
		- `TestEmpresaReportesHandlerDatasetOperativoPropinasAcumulado`.
		- `TestEmpresaReportesHandlerDatasetOperativoComisionesLavador`.
		- `TestEmpresaReportesHandlerDatasetOperativoFacturacionTrazabilidad`.
		- `TestEmpresaReportesHandlerDatasetOperativoAuditoriaAcciones`.
	- Se refuerza `ensureEmpresaReportesSchema` con `EnsureEmpresaPropinasSchema`, `EnsureEmpresaComisionesServicioSchema` y `EnsureEmpresaAuditoriaSchema` para cobertura de suite completa.
	- Se actualiza la matriz en `documentos/modulos del proyecto.md` marcando Propinas, Comisiones, Facturacion y Auditoria como activos en reportes.
	- Validacion tecnica ejecutada:
		- `runTests` focalizado en 4 pruebas nuevas (ok).
		- `runTests` completo sobre `backend/handlers/reportes_test.go` (16/16 ok).
- Plan secuencial de cierre de modulos incompletos - bloque 5 (Compras):
	- Se rediseÃƒÂ±a el dataset `operativo_compras_movimientos` en `backend/handlers/reportes.go` para consolidar compras por proveedor, dejando de depender solo de movimientos de inventario.
	- El dataset ahora expone KPI de ciclo documental: `ordenes_emitidas`, `recepciones`, `contabilizaciones`, `monto_ordenado`, `monto_recepcionado`, `monto_contabilizado`, `brecha_monto` y cumplimiento de recepcion/monto.
	- Se actualiza el catalogo del reporte con nuevo titulo y descripcion orientados a `costo por proveedor y recepcion vs orden`.
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoComprasMovimientos` para validar consolidado por proveedor, totales de resumen y porcentajes de cumplimiento.
	- Se actualiza la matriz en `documentos/modulos del proyecto.md` marcando Compras como activo en reportes.
	- Validacion tecnica ejecutada:
		- `runTests` sobre `backend/handlers/reportes_test.go` con 8 pruebas objetivo (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 4 (Inventario):
	- Se extiende el dataset `operativo_inventario_bodega` en `backend/handlers/reportes.go` con metricas de:
		- rotacion estimada y cobertura (`salida_promedio_diaria`, `dias_cobertura`, `indice_rotacion_30d`),
		- riesgo de quiebre proyectado (`estado_proyeccion`, `sugerido_reposicion`),
		- valorizacion por producto/bodega (`valorizacion_costo`, `valorizacion_venta`).
	- Se agregan KPI de resumen de inventario (`alertas`, `deficit`, `movimientos`, cobertura y rotacion promedio).
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoInventarioBodega`.
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando Inventario como activo en reportes.
	- Se marca el bloque 4 como completado en `Pendiente Notas`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler(DatasetOperativoInventarioBodega|DatasetOperativoCadenaCumplimiento|DatasetOperativoTarifasIngresos|DatasetOperativoReservasOcupacion|DatasetOperativoModulosResumen|CatalogoSuiteDataset|Exportes)" -count=1` (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 3 (CRM/Produccion/Logistica):
	- Se agrega el dataset `operativo_cadena_cumplimiento` en `backend/handlers/reportes.go` para consolidar conversion comercial y cumplimiento operativo.
	- El dataset resume por modulo (`crm_leads`, `produccion_ordenes`, `logistica_envios`) registros de rango, estados finalizados/en proceso y monto de referencia.
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoCadenaCumplimiento`.
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando CRM/Produccion/Logistica como activo en reportes.
	- Se marca el bloque 3 como completado en `Pendiente Notas`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler(DatasetOperativoCadenaCumplimiento|DatasetOperativoTarifasIngresos|DatasetOperativoReservasOcupacion|DatasetOperativoModulosResumen|CatalogoSuiteDataset|Exportes)" -count=1` (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 2 (tarifas):
	- Se consolida el dataset `operativo_tarifas_ingresos` para ingresos por modelo de tarifa (`tarifa_por_dia`, `tarifa_por_minutos`, `sin_modelo`).
	- Se amplia `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoTarifasIngresos` y bootstrap de esquemas de tarifas (`EnsureEmpresaTarifasPorDiaSchema`, `EnsureEmpresaTarifasPorMinutosSchema`).
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando tarifas como activo en reportes.
	- Se marca el bloque de tarifas como completado en `Pendiente Notas`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler(DatasetOperativoTarifasIngresos|DatasetOperativoReservasOcupacion|DatasetOperativoModulosResumen|CatalogoSuiteDataset|Exportes)" -count=1` (ok).
- Plan secuencial de cierre de modulos incompletos - bloque 1 (reservas):
	- Se agrega el dataset `operativo_reservas_ocupacion` en `backend/handlers/reportes.go` para consolidar ocupacion y cumplimiento por estacion.
	- Se amplian pruebas en `backend/handlers/reportes_test.go` con `TestEmpresaReportesHandlerDatasetOperativoReservasOcupacion` y bootstrap de `EnsureEmpresaReservasHotelSchema`.
	- Se actualiza matriz de estado en `documentos/modulos del proyecto.md` marcando reservas como activo en reportes.
	- Se documenta plan secuencial de cierre en `Pendiente Notas` y se marca reservas como primer modulo completado.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler" -count=1` (ok).
- Continuidad de plan en reportes por modulos:
	- Se valida y consolida el dataset `operativo_modulos_resumen` en `backend/handlers/reportes.go`.
	- Se corrige una llamada interna a `reportesCountByEmpresa` en el builder de resumen por modulos para compatibilidad con la firma actual de la funcion.
	- Se amplia `backend/handlers/reportes_test.go` con:
		- bootstrap de esquema para modulos ERP extendidos,
		- prueba `TestEmpresaReportesHandlerDatasetOperativoModulosResumen` con verificacion de conteos por modulo y consistencia de `summary`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresaReportesHandler" -count=1` (ok, 7/7).
- Continuidad de plan en frontend y ayuda:
	- Se corrige doble scrollbar en panel empresa eliminando la definicion duplicada de `.admin-empresa-frame` con altura fija en `web/estilos.css`.
	- Se actualiza `web/administrar_empresa/propinas.html` para recuperar consistencia visual/operativa (layout empresa, tablas estandar e integracion de menu flotante).
	- Se amplia `web/ayuda/ayuda.html` con guias de modulos pendientes: propinas, comisiones, ERP extendido y calculadora por empresa.
- Se continua el plan con dos faltantes operativos:
	- Utilidad nueva `web/administrar_empresa/calculadora.html` con contexto por empresa (`empresa_id`), memoria/historial aislados por empresa y exportacion JSON del historial.
	- Documento nuevo `documentos/modulos del proyecto.md` con inventario de modulos, conteo total y matriz base modulo -> reportes recomendados.
- Se integra la calculadora en navegacion:
	- `web/menu.js` agrega enlace `Calculadora` en menu flotante y propaga `empresa_id`.
	- `web/administrar_empresa.html` agrega enlace lateral `Calculadora`.
	- `web/js/administrar_empresa.js` incorpora `linkCalculadora` en navegacion y permisos (`finanzas/read`).
	- `web/estilos.css` agrega estilos `calc-*` para la nueva pantalla.
- Se actualiza documentacion tecnica:
	- `documentos/diagramas/estructura_del_codigo.md`.
	- `documentos/descripcion_del_proyecto`.
	- `documentos/descripcion_de_archivos`.
- Se completan faltantes de cobertura para la maquina de estados documental en ventas y CRM:
	- `backend/handlers/modulos_faltantes_test.go` amplÃƒÂ­a pruebas para:
		- ventas: `pedidos` y `devoluciones` (transiciones validas e invalidas),
		- CRM: `interacciones` y `campanas` (pipeline basico con validacion de transiciones).
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresa(IntegracionesAPIsHandlerHealthAndSync|IntegracionesBancosHandlerSyncAndEstado|VentasCotizacionesStateMachine|CRMLeadsStateMachine|VentasPedidosStateMachine|VentasDevolucionesStateMachine|CRMInteraccionesStateMachine|CRMCampanasStateMachine)" -count=1` (ok).
		- `go test ./... -run "^$" -count=1` (ok).
- Se implementan integraciones API/Bancos ejecutables y maquina de estados documental en CRM/Ventas:
	- `backend/handlers/modulos_faltantes.go` agrega handlers especializados sobre CRUD base:
		- Integraciones: `action=health_check`, `action=sync_manual`, `action=estado` en `/api/empresa/integraciones/apis` y `/api/empresa/integraciones/bancos`.
		- CRM/Ventas documentales: `action=estado`, `action=transiciones`, `action=transicionar` en rutas `/api/empresa/crm/*` y `/api/empresa/ventas/*`.
	- `backend/handlers/modulos_faltantes_test.go` (nuevo) cubre:
		- health/sync de integraciones,
		- transiciones validas/invÃƒÂ¡lidas de cotizaciones y leads.
	- `web/js/modulos_erp_extendido.js` agrega botones operativos por fila:
		- Integraciones: `Health`, `Sync`, `Estado`.
		- CRM/Ventas documentales: `Transiciones`, `Transicionar`.
	- Validacion tecnica ejecutada:
		- `go test ./handlers -run "TestEmpresa(IntegracionesAPIsHandlerHealthAndSync|IntegracionesBancosHandlerSyncAndEstado|VentasCotizacionesStateMachine|CRMLeadsStateMachine)" -count=1` (ok).
		- `go test ./... -run "^$" -count=1` (ok).
- Inicio de implementacion del bloque de integraciones y pagos (fase de robustecimiento):
	- `backend/handlers/payments_handlers.go` agrega:
		- `MercadoPagoReconcileHandler` para conciliacion manual de pagos pendientes contra API Mercado Pago (`/mercadopago/reconcile`, requiere sesion admin).
		- `WompiWebhookHandler` para notificaciones servidor-servidor (`/wompi/webhook`) con validacion de firma y activacion automatica de licencia cuando aplica.
		- helpers compartidos para token MP, parseo de `external_reference`, estatus aprobados y activacion idempotente de licencia.
	- `backend/db/db.go` agrega helpers de persistencia para conciliacion:
		- listado de pendientes MP (`ListMPPaymentsForReconciliation`),
		- actualizacion por `id` y `payment_id`,
		- actualizacion Wompi por `reference`,
		- resoluciÃƒÂ³n de contexto licencia/empresa para Wompi.
	- `backend/main.go` registra rutas nuevas:
		- `/mercadopago/reconcile`
		- `/wompi/webhook`
	- Validacion tecnica ejecutada:
		- `go test ./auth ./db ./handlers ./metrics ./utils` (ok).
- Se divide la interfaz de ERP extendido en submodulos por dominio, manteniendo el mismo backend.
	- `web/administrar_empresa/modulos_erp_extendido.html` pasa a ser hub de dominios (ventas, finanzas, inventario/compras/rrhh, crm, produccion, logistica, documental/integraciones/dian).
	- `web/administrar_empresa/modulos_erp_dominio.html` (nuevo) concentra la operacion CRUD del dominio seleccionado sin cambiar endpoints backend.
	- `web/js/modulos_erp_extendido.js` (nuevo) centraliza la logica operativa reutilizable de submodulos por dominio.
	- `web/estilos.css` agrega estilos de navegacion y tarjetas para `erp-domain-*`.
- Se completa la operacion frontend de los modulos ERP extendidos en panel de empresa.
	- `web/administrar_empresa/modulos_erp_extendido.html` (nuevo) centraliza el uso de todos los endpoints ERP faltantes con:
		- listado con filtros (`q`, `limit`, `offset`, `include_inactive`),
		- detalle por ID,
		- crear/actualizar por payload JSON,
		- activar/desactivar y eliminacion logica por registro,
		- herramientas DIAN (`checklist`, `validar`, `generar_cufe_demo`, `generar_xml_demo`).
	- `web/administrar_empresa.html` agrega acceso lateral `ERP extendido`.
	- `web/js/administrar_empresa.js` integra `linkERPExtendido` en navegacion y permisos (modulo `seguridad`, accion `update`).
	- `web/estilos.css` agrega estilos dedicados del nuevo modulo (`erp-*`) para formularios, salida, tabla y estado visual.
- Se implementa base de modulos ERP faltantes en backend con esquema multiempresa, migracion y rutas nuevas:
	- `backend/db/modulos_faltantes.go` (tablas y CRUD generico por `empresa_id`).
	- `backend/handlers/modulos_faltantes.go` (rutas ERP adicionales y handler DIAN Colombia).
	- `backend/main.go` integra `EnsureEmpresaModulosFaltantesSchema`, migracion `2026-04-06-021-modulos-faltantes-erp` y `RegisterEmpresaModulosFaltantesRoutes`.
- Se agrega soporte DIAN Colombia operativo en endpoint `/api/empresa/facturacion_electronica/dian` con acciones:
	- `checklist` y `validar`.
	- `generar_cufe_demo` y `generar_xml_demo`.
- Se amplÃƒÂ­a `web/ayuda/ayuda.html` con seccion detallada para configurar facturacion DIAN desde cero.
- Se sincroniza documentacion tecnica y de BD:
	- `documentos/diagramas/estructura_del_codigo.md`.
	- `documentos/estructura_bd.md` y `estructura_bd.md`.
	- `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
- Validacion tecnica ejecutada:
	- `go test ./... -run "^$" -count=1` (ok).

## 2026-04-05
- Se continua con todos los bloques y pruebas en una corrida adicional de verificacion.
	- Validaciones ejecutadas:
		- `runTests` global (150 passed, 0 failed).
		- `powershell -File ..\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-182345.log`.

## 2026-04-05
- Se continua con todos los bloques y pruebas en una nueva corrida completa.
	- Validaciones ejecutadas:
		- `runTests` global (150 passed, 0 failed).
		- `powershell -File ..\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-182133.log`.

## 2026-04-05
- Se continua ejecucion de todos los bloques y pruebas con validacion ampliada por modulos criticos.
	- Validaciones ejecutadas:
		- `powershell -File ..\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
		- `go test ./handlers -run "TestEmpresa(Usuario|Clientes|Inventario|Compras|Facturacion|Finanzas|Auditoria|Permisos|Carritos)" -count=1` (ok).
		- `go test ./db -run "Test(Cliente|Proveedor|Inventario|Finanzas|Facturacion|Reserva|Vehiculo|Nomina|Tarifa|CodigoDescuento|Comision|Propina)" -count=1` (ok).
		- `go test ./... -count=1` (ok).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-181807.log`.

## 2026-04-05
- Se reejecuta la validacion integral y bloques adicionales de pruebas para cierre tecnico.
	- Validaciones ejecutadas:
		- `powershell -File .\\scripts\\validar_punto_13.ps1` (ok, suite productiva + suite completa backend).
		- `go test ./handlers -run "TestEmpresa(CarritosCompraListIncluyeTarifaPorDiaAutomatica|TarifasPorMinutosHandlerCRUDAndCalculo|TarifasPorDiaHandlerCRUDAndCalculo|ReservasHotelHandlerCRUDAndDisponibilidad|VehiculosRegistroHandlerCRUDFlow|NominaSueldosHandlerFlow|PropinasHandlerConfigAndReporte|ComisionesServicioHandlerConfigAndReporte|ConfiguracionOperativaHandlerConfigAndRole)" -count=1` (ok).
		- `go test ./... -count=1` (ok).
	- Evidencia actualizada:
		- `documentos/punto_13_validacion_integral_resultado.md`.
		- `scripts/logs/punto13-validacion-20260405-181423.log`.

## 2026-04-05
- Se corrige regresion de pruebas de carritos ante el recalculo de totales al pagar estacion.
	- `backend/handlers/auth_users_carritos_test.go` actualiza el sembrado de datos en:
		- `TestEmpresaCarritosCompraAplicaPropinaSegunConfiguracion`.
		- `TestEmpresaCarritosCompraCodigoDescuentoConsumeUso`.
		- `TestEmpresaCarritosCompraRejectsMetodoPagoInvalido`.
	- Las pruebas ahora crean items reales en `carrito_compra_items` en lugar de forzar `subtotal/total` por SQL directo, quedando alineadas con `RefreshCarritoTotalConTarifaPorDia`.
	- Validacion ejecutada:
		- `go test ./handlers -run "TestEmpresaCarritosCompraAplicaPropinaSegunConfiguracion|TestEmpresaCarritosCompraCodigoDescuentoConsumeUso|TestEmpresaCarritosCompraRejectsMetodoPagoInvalido" -count=1` (ok).
		- `powershell -File .\\scripts\\validar_punto_13.ps1` (ok, incluye suite productiva y suite completa backend).

## 2026-04-05
- Se corrige el flujo de login de usuario de empresa para mantener alcance por `empresa_id` en endpoints publicos de autenticacion.
	- `backend/handlers/usuarios_empresa.go` ahora propaga `empresa_id` en enlaces de correo y confirmacion hacia `/login_usuario.html?empresa_id=...`.
	- `ConfirmarCorreoUsuarioHandler` usa el `empresa_id` confirmado (o de query) para construir el enlace de retorno al login de usuario.
	- `web/js/login_usuario.js` toma `empresa_id` desde querystring y lo envia en query + body a `/api/empresa/usuarios/login` y `/api/empresa/usuarios/establecer_password`.
	- Validacion ejecutada:
		- `go test ./handlers -run "EmpresaUsuario(LoginHandlerSuccess|SetPasswordHandlerSuccess|LoginHandlerRejectsWrongEmpresaScopeFromQuery|SetPasswordHandlerRejectsWrongEmpresaScopeFromQuery)" -count=1` (ok).
		- `get_errors` sobre archivos modificados (sin errores).

## 2026-04-05
- Se completa auditoria integral de rutas `/api/empresa` y se cierra cobertura de wrappers por empresa al 100% en el registro de rutas.
	- `backend/handlers/empresa_permisos.go` agrega `WithEmpresaPublicScope` para endpoints publicos que requieren alcance por `empresa_id` sin autenticacion previa de admin.
	- `backend/main.go` envuelve `/api/empresa/usuarios/login` y `/api/empresa/usuarios/establecer_password` con `WithEmpresaPublicScope`.
	- `backend/main.go` envuelve `/api/empresa/facturacion_electronica/paises_disponibles` con `WithEmpresaFacturacionPermissions`.
	- `backend/handlers/chat_con_inteligencia_artificial_router.go` envuelve rutas del modulo IA (`modelos`, `modelo_preferido`, `consultar`, `historial`) con `WithEmpresaSeguridadPermissions`.
	- `web/administrar_empresa/facturacion_electronica.html` envia `empresa_id` al consultar `paises_disponibles` para compatibilidad con el wrapper de facturacion.
	- Validacion ejecutada:
		- `go test ./handlers -run "WithEmpresaVentasPermissionsInjectsEmpresaIDContextForParsers|EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaUsuario(LoginHandlerRejectsWrongEmpresaScope|SetPasswordHandlerRejectsWrongEmpresaScope)|ModeloPreferidoHandler(Get|Put)RejectsEmpresaFueraDeAlcanceByGoogleAccount|HistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount" -count=1` (ok).
		- `go test ./... -run "^$"` (ok).

## 2026-04-05
- Se refuerza la integracion multiempresa para que el alcance autorizado por `empresa_id` viaje en el contexto de request y sea reutilizable por handlers.
	- `backend/handlers/empresa_permisos.go` ahora inyecta `empresaID` en `context.Context` dentro de `WithEmpresa*Permissions`.
	- `backend/handlers/productos.go` actualiza `parseEmpresaIDQuery`, `parseInt64Query` y `parseInt64QueryOptional` para priorizar `empresaID` desde contexto cuando existe.
	- `backend/handlers/empresa_permisos_test.go` agrega `TestWithEmpresaVentasPermissionsInjectsEmpresaIDContextForParsers` para validar la propagacion de scope multiempresa sin dependencia estricta de querystring.
	- Validacion ejecutada: `go test ./handlers -run "WithEmpresaVentasPermissionsInjectsEmpresaIDContextForParsers|EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaUsuario(LoginHandlerRejectsWrongEmpresaScope|SetPasswordHandlerRejectsWrongEmpresaScope)|ModeloPreferidoHandler(Get|Put)RejectsEmpresaFueraDeAlcanceByGoogleAccount|HistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount" -count=1` (ok).

## 2026-04-05
- Se completa el subbloque de regresion UAT para endpoints sin wrapper de modulo (continuidad Punto 3).
	- `backend/handlers/auth_users_carritos_test.go` agrega cobertura de alcance por `empresa_id` enviado por querystring en:
		- `TestEmpresaUsuarioLoginHandlerRejectsWrongEmpresaScopeFromQuery`.
		- `TestEmpresaUsuarioSetPasswordHandlerRejectsWrongEmpresaScopeFromQuery`.
	- `backend/handlers/chat_con_inteligencia_artificial_controller_test.go` agrega cobertura de aislamiento por cuenta Google autenticada en:
		- `TestModeloPreferidoHandlerGetRejectsEmpresaFueraDeAlcanceByGoogleAccount`.
		- `TestModeloPreferidoHandlerPutRejectsEmpresaFueraDeAlcanceByGoogleAccount`.
		- `TestHistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount`.
	- Validacion ejecutada: `go test ./handlers -run "EmpresaUsuario(LoginHandlerRejectsWrongEmpresaScope|SetPasswordHandlerRejectsWrongEmpresaScope)|ModelosHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount|ConsultarHandlerRejectsEmpresaFueraDeAlcance|ModeloPreferidoHandler(Get|Put)RejectsEmpresaFueraDeAlcanceByGoogleAccount|HistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount" -count=1` (ok).

## 2026-04-05
- Se completa el subbloque de consumo frontend del contexto de permisos (cierre operativo Punto 3).
	- `web/js/administrar_empresa.js` ahora consume `GET /api/empresa/permisos_contexto?empresa_id={id}` para resolver visibilidad real de enlaces por modulo/accion en el menu lateral.
	- El panel empresa ahora admite `id` o `empresa_id` en querystring para resolver el contexto de permisos sin ambiguedad.
	- Se mantiene fallback local por rol cuando el endpoint no esta disponible, evitando bloqueos de navegacion.
	- `web/administrar_empresa.html` agrega indicador visual `menuPermsEvidence` para evidencia UAT del rol/fuente de permisos aplicado en pantalla.
	- Validacion ejecutada: `get_errors` sobre frontend modificado (sin errores).

## 2026-04-05
- Se agrega endpoint de contexto de permisos por empresa para reforzar el cierre del Punto 3 (permisos y seguridad).
	- `backend/handlers/empresa_permisos.go` incorpora `GET /api/empresa/permisos_contexto` con respuesta de permisos efectivos por modulo/accion para el rol autenticado.
	- El endpoint soporta `include_matrix=1` para retornar matriz completa por rol (`super_administrador`, `admin_empresa`, `supervisor_sucursal`, `cajero`, `inventario`, `compras`, `contabilidad`, `auditor`).
	- `backend/main.go` registra la ruta bajo `WithEmpresaSeguridadPermissions` para mantener aislamiento por `empresa_id`.
	- `backend/handlers/empresa_permisos_test.go` agrega `TestEmpresaPermisosContextoHandlerRetornaPermisosPorRol` y `TestEmpresaPermisosContextoHandlerIncluyeMatrizRoles`.
	- Validacion ejecutada: `go test ./handlers -run "PermisosContexto|WithEmpresa.*Permissions" -count=1` (ok).

## 2026-04-05
- Se amplian los reportes contables de flujo de caja con filtros por categoria y metodo de pago.
	- `backend/handlers/reportes.go` incorpora filtros `categoria` y `metodo_pago` en `contable_flujo_caja` para segmentar ingresos/egresos diarios.
	- El resumen del dataset ahora expone `filtro_categoria` y `filtro_metodo_pago` para trazabilidad del reporte exportado.
	- `backend/handlers/reportes_test.go` agrega `TestEmpresaReportesHandlerDatasetContableFlujoCajaFiltros` para validar segmentacion por categoria/metodo.
	- `web/administrar_empresa/reportes.html` agrega campos de filtro contable y los propaga en consultas/exportaciones del endpoint `/api/empresa/reportes`.

## 2026-04-05
- Se extiende el modulo de reportes con dataset contable de flujo de caja diario.
	- `backend/handlers/reportes.go` agrega dataset `contable_flujo_caja` en `/api/empresa/reportes` y consolida ingresos, egresos, neto del dia, saldo acumulado y conteo de movimientos por fecha.
	- El dataset mantiene paridad de exportacion en `pdf`, `xls`, `csv`, `json` y `txt` desde el catalogo central de reportes empresariales.
	- `backend/handlers/reportes_test.go` agrega `TestEmpresaReportesHandlerDatasetContableFlujoCaja` para validar filas diarias y resumen acumulado del periodo.

## 2026-04-05
- Se extiende el modulo de reportes con dataset contable de liquidaciones de nomina y exportacion PDF.
	- `backend/handlers/reportes.go` agrega dataset `contable_nomina_liquidaciones` con filtros por periodo y `empleado_nomina_id`.
	- `backend/handlers/reportes.go` habilita formato `pdf` en la exportacion de datasets de `/api/empresa/reportes`.
	- `web/administrar_empresa/reportes.html` agrega opcion `PDF` en el selector de formato.
	- `web/administrar_empresa/nomina_sueldos.html` incorpora accion `Exportar liquidaciones` usando `/api/empresa/reportes?action=export`.
	- `backend/handlers/reportes_test.go` agrega cobertura de dataset de nomina y validacion de export PDF.

## 2026-04-05
- Se integra operativamente el modulo de nomina de sueldos con asistencia en backend y panel de empresa.
	- `backend/main.go` incorpora `EnsureEmpresaNominaSchema`, migracion `2026-04-05-020-nomina-sueldos` y ruta `/api/empresa/nomina`.
	- `web/administrar_empresa/nomina_sueldos.html` (nuevo) agrega configuracion legal, empleados, festivos, calculo y consulta de liquidaciones.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran `linkNominaSueldos` en menu y permisos.
	- `web/estilos.css` agrega estilos dedicados del modulo.
	- `documentos/estructura_bd.md` y `estructura_bd.md` incluyen tablas y relaciones de nomina.

## 2026-04-05
- Se agrega `ventas_simple.html` como carrito alterno por estaciÃƒÂ³n (modo supermercado) con activaciÃƒÂ³n/desactivaciÃƒÂ³n por estaciÃƒÂ³n.
	- `web/administrar_empresa/ventas_simple.html` (nuevo) incorpora flujo rÃƒÂ¡pido para buscar productos, agregarlos al carrito, ajustar cantidades y visualizar total consolidado por estaciÃƒÂ³n.
	- Se corrige la visibilidad del campo de referencia de pago para mÃƒÂ©todos que la requieren (`tarjeta_credito`, `tarjeta_debito`, `transferencia_bancaria`).
	- El cobro se ejecuta con flujo simplificado usando `action=pagar_estacion` y permite iniciar nueva venta con `action=activar_estacion`.
	- `web/administrar_empresa/configuracion_de_estaciones.html` agrega la bandera local `venta_simple_habilitada` por estaciÃƒÂ³n.
	- `web/administrar_empresa/estaciones.html` enruta automÃƒÂ¡ticamente cada estaciÃƒÂ³n al carrito completo (`carrito_de_compras.html`) o al carrito simple (`ventas_simple.html`) segÃƒÂºn su configuraciÃƒÂ³n.
	- `web/estilos.css` integra estilos responsive para el nuevo mÃƒÂ³dulo y etiqueta visual del modo por estaciÃƒÂ³n.

## 2026-04-05
- Se actualiza la configuracion de `agente_go` para reforzar reportes e interoperabilidad contable.
	- `.github/agents/agente_go.agent.md` agrega regla obligatoria para que todos los reportes puedan exportarse, como minimo, en `PDF` y `Excel`, y tambien en formatos de uso comun (`CSV`, `JSON`, `TXT`).
	- Se incorpora regla de compatibilidad con software contable externo mediante formatos estandar de intercambio y trazabilidad por `empresa_id`, documento y periodo.

## 2026-04-05
- Se agrega el dataset `reporte_de_turno` al modulo empresarial de reportes para control operativo de caja por turno.
	- `backend/handlers/reportes.go` incorpora `reporte_de_turno` en `/api/empresa/reportes` con filtros por `usuario`, `caja_codigo`, `turno` y `cierre_id`.
	- El dataset incluye detalle por carrito con `activado_en`, `pagado_en`, metodo de pago y acumulados de ventas por `producto` y `servicio`.
	- El resumen del reporte calcula gastos de turno y efectivo esperado (`efectivo_deberia_haber`) combinando cierres de caja y movimientos financieros.
	- `web/administrar_empresa/reportes.html` agrega campos de filtro de turno/caja/usuario/cierre y envia estos parametros en consultas y exportes del dataset.
	- `backend/handlers/reportes_test.go` agrega `TestEmpresaReportesHandlerDatasetReporteTurno` para validar filtros y consistencia del resumen financiero del turno.

## 2026-04-05
- Se crea el modulo de tarifas por dia por estacion con recÃƒÂ¡lculo automÃƒÂ¡tico de deuda en carritos hotel activos.
	- `backend/db/tarifas_por_dia.go` (nuevo) agrega esquema `empresa_tarifas_por_dia`, CRUD, horarios `hora_check_in`/`hora_check_out` y calculo de dias/monto.
	- `backend/db/carritos_tarifa_dia.go` (nuevo) integra calculo automÃƒÂ¡tico de deuda diaria en carritos de estaciÃƒÂ³n y refresco masivo para listados.
	- `backend/db/carritos_compras.go` ajusta `RecalculateCarritoCompraTotals` para incluir tarifa diaria cuando aplique.
	- `backend/handlers/tarifas_por_dia.go` (nuevo) expone `/api/empresa/tarifas_por_dia` con acciones `listar`, `detalle`, `aplicable`, `calcular`, `activar` y `desactivar`.
	- `backend/handlers/carritos_compras.go` recalcula tarifa diaria al listar carritos y antes de `action=pagar_estacion`.
	- `backend/main.go` integra `EnsureEmpresaTarifasPorDiaSchema`, migracion `2026-04-05-019-tarifas-por-dia` y ruta protegida del modulo.
	- `web/administrar_empresa/tarifas_por_dia.html` (nuevo) agrega UI de configuracion, filtros y simulador por rango de fechas.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran `linkTarifasPorDia` en menu lateral y permisos.
	- Cobertura agregada en `backend/db/tarifas_por_dia_test.go`, `backend/handlers/tarifas_por_dia_test.go` y `backend/handlers/carritos_tarifa_por_dia_test.go`.

## 2026-04-05
- Se crea el modulo de tarifas por minutos por estacion con reglas por dia de semana y calculo de bloques extra.
	- `backend/db/tarifas_por_minutos.go` (nuevo) agrega esquema `empresa_tarifas_por_minutos`, CRUD, resolucion por dia (`dia_semana_desde/hasta`) y calculo de monto por minutos consumidos.
	- `backend/handlers/tarifas_por_minutos.go` (nuevo) expone `/api/empresa/tarifas_por_minutos` con acciones `listar`, `detalle`, `aplicable`, `calcular`, `activar` y `desactivar`.
	- `backend/main.go` integra `EnsureEmpresaTarifasPorMinutosSchema`, migracion `2026-04-05-018-tarifas-por-minutos` y ruta protegida del modulo.
	- `web/administrar_empresa/tarifas_por_minutos.html` (nuevo) agrega formulario de tarifas, filtros y simulador de cobro por minutos.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran `linkTarifasPorMinutos` en menu lateral y permisos por rol.
	- Cobertura agregada en `backend/db/tarifas_por_minutos_test.go` y `backend/handlers/tarifas_por_minutos_test.go`.
	- Se actualiza documentacion y diagramas: `documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md`.

## 2026-04-05
- Se ubica el documento de base de datos dentro de la carpeta `documentos` y se alinea `agente_go` con esa ruta.
	- `documentos/estructura_bd.md` se incorpora como ubicacion requerida de la estructura de base de datos.
	- `.github/agents/agente_go.agent.md` ahora exige revisar `documentos/estructura_bd.md` antes de cambios en tablas, consultas, migraciones o datos.
	- `estructura_bd.md` en raiz se mantiene sincronizado como copia de compatibilidad documental.

## 2026-04-05
- Se actualiza la configuracion de `agente_go` para forzar lectura previa de documentacion de base de datos en tareas de datos.
	- `.github/agents/agente_go.agent.md` agrega regla para revisar `estructura_bd.md` antes de cambios en tablas, consultas, migraciones o datos operativos.

## 2026-04-05
- Se agrega modulo de busqueda y gestion de facturas electronicas por empresa.
	- `web/administrar_empresa/facturas_electronicas.html` (nuevo) permite buscar por cliente, documento y rango de fechas; ver detalle; reenviar por correo; e imprimir.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso lateral `Facturas electrÃƒÂ³nicas` con permisos de lectura del modulo `facturacion`.
	- `backend/handlers/facturacion_electronica.go` incorpora:
		- `GET action=documentos` para consulta de documentos de facturacion por filtros (`cliente`, `documento`, `fecha_desde`, `fecha_hasta`, `estado_documento`, `tipo_documento`, `q`).
		- `PUT/POST action=reenviar_correo` para reintento manual de envio al correo del cliente.
	- `backend/db/documentos_transaccionales.go` agrega consulta enriquecida con cliente (`nombre`, `email`, `documento`) para listado filtrado.
	- Cobertura agregada en `backend/db/documentos_transaccionales_test.go` para filtros de cliente/fecha/documento.

## 2026-04-05
- Se normaliza la documentacion de base de datos para eliminar duplicidad entre documentos.
	- `estructura_bd.md` queda como fuente canonica del esquema fisico motor legado retirado.
	- `documentos/descripcion_de_las_bases_De_datos` se redefine como guia complementaria funcional y reglas operativas de mantenimiento.
	- Se evita repetir listados tabla-por-tabla en dos archivos distintos.

## 2026-04-05
- Se consolida ConfiguraciÃƒÂ³n avanzada dentro de FacturaciÃƒÂ³n electrÃƒÂ³nica en el panel de empresa.
	- `web/administrar_empresa/facturacion_electronica.html` integra el formulario completo de configuraciÃƒÂ³n avanzada fiscal/impresiÃƒÂ³n y su persistencia mediante `/api/empresa/configuracion_avanzada`.
	- `web/administrar_empresa.html` elimina el enlace lateral independiente `ConfiguraciÃƒÂ³n avanzada` para dejar una ÃƒÂºnica entrada funcional en `FacturaciÃƒÂ³n electrÃƒÂ³nica`.
	- `web/js/administrar_empresa.js` retira `linkConfigAvanzada` del catÃƒÂ¡logo de enlaces/permisos del menÃƒÂº.
	- `web/ayuda/ayuda.html` actualiza el tutorial para indicar que la configuraciÃƒÂ³n avanzada ahora se gestiona desde `facturacion_electronica.html`.
	- `web/administrar_empresa/configuracion_avanzada.html` se elimina del repositorio por consolidaciÃƒÂ³n funcional.

## 2026-04-05
- Se agrega configuracion operativa de cobro por empresa y por rol de usuario.
	- `backend/db/configuracion_operativa.go` (nuevo) agrega tablas `empresa_configuracion_operativa` y `empresa_configuracion_operativa_roles`, con resolucion efectiva de permisos por rol.
	- `backend/handlers/configuracion_operativa.go` (nuevo) expone `/api/empresa/configuracion_operativa` para consultar y actualizar reglas base y overrides por rol (`action=rol`).
	- `backend/handlers/empresa_permisos.go` y `backend/handlers/productos.go` propagan/normalizan rol admin en request para enforcement transversal.
	- `backend/handlers/carritos_compras.go` aplica enforcement en `action=pagar_estacion`: bloquea metodos de pago no permitidos y desactiva propina/comision segun politica operativa efectiva por rol.
	- `backend/main.go` registra `EnsureEmpresaConfiguracionOperativaSchema`, migracion `2026-04-05-017-configuracion-operativa-cobro` y ruta protegida `/api/empresa/configuracion_operativa`.
	- `web/administrar_empresa/configuracion.html` incorpora tarjeta de checks para metodos de pago, propinas y comisiones por empresa y por rol.
	- `web/administrar_empresa/carrito_de_compras.html` consume la politica operativa efectiva y refleja en UI los metodos permitidos, con bloqueo visual y validacion previa al pago.
	- Cobertura agregada en `backend/db/configuracion_operativa_test.go`, `backend/handlers/configuracion_operativa_test.go` y `backend/handlers/auth_users_carritos_test.go`.
	- Validacion ejecutada: pruebas dirigidas en DB/handlers/carritos (ok) y verificacion sin errores en frontend actualizado.

## 2026-04-05
- Se crea el modulo de comisiones por servicio por empresa con reporte por lavador.
	- `backend/db/comisiones_servicio.go` (nuevo) agrega tablas de configuracion y movimientos (`empresa_comisiones_servicio_configuracion`, `empresa_comisiones_servicio_movimientos`) con calculo/reporte por lavador.
	- `backend/handlers/comisiones.go` (nuevo) expone `/api/empresa/comisiones` con acciones `config`, `reporte` y `movimientos`.
	- `backend/handlers/carritos_compras.go` integra `usuario_lavador` en `action=pagar_estacion` y registra comisiones automaticas de servicios de lavado al cerrar venta.
	- `backend/main.go` asegura esquema de comisiones, registra migracion `2026-04-05-016-comisiones-servicio` y publica ruta protegida de comisiones bajo permisos de finanzas.
	- `web/administrar_empresa/comisiones.html` (nuevo) incorpora configuracion y reporte de comisiones por lavador.
	- `web/administrar_empresa/carrito_de_compras.html` agrega captura de lavador para comision, carga de configuracion de comisiones y visualizacion de comision estimada/registrada en cobro.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran acceso lateral `Comisiones` (`linkComisiones`) con permisos del modulo finanzas.
	- Cobertura agregada en `backend/db/comisiones_servicio_test.go`, `backend/handlers/comisiones_test.go` y `backend/handlers/auth_users_carritos_test.go`.

## 2026-04-05
- Se actualiza la configuracion de `agente_go` para definir semantica y concurrencia de estaciones.
	- `.github/agents/agente_go.agent.md` agrega que una estacion puede representar mesa de restaurante, habitacion de hotel, habitacion de motel, punto de caja u otro punto operativo equivalente.
	- Se establece que estaciones deben soportar multiples carritos/sesiones y multiples clientes en simultaneo, con aislamiento por `empresa_id` y trazabilidad operativa.

## 2026-04-05
- Se completa el modulo de reservas por estacion/habitacion para operacion empresarial.
	- `backend/db/reservas_hotel.go` (nuevo) implementa esquema y logica de reservas con disponibilidad por rango, conflicto de solapamiento, expiracion de pendientes, confirmacion de pago, cancelacion, activacion/desactivacion y eliminacion.
	- `backend/handlers/reservas_hotel.go` (nuevo) expone `/api/empresa/reservas_hotel` con acciones `listar`, `detalle`, `disponibilidad`, `confirmar_pago`, `cancelar`, `activar`, `desactivar` y CRUD operativo.
	- `backend/main.go` asegura esquema de reservas, registra migracion `2026-04-05-015-reservas-hotel` y publica ruta protegida bajo permisos de ventas.
	- `web/administrar_empresa/reservas_hotel.html` (nuevo) agrega interfaz para crear/editar reservas, consultar disponibilidad y ejecutar acciones de ciclo de vida.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso lateral `Reservas` (`linkReservasHotel`) con control de permisos por rol.
	- Cobertura de pruebas en:
		- `backend/db/reservas_hotel_test.go` (flujo DB end-to-end con disponibilidad y estados).
		- `backend/handlers/reservas_hotel_test.go` (nuevo, flujo HTTP completo del endpoint).
	- Validaciones ejecutadas:
		- `go test ./db -run ReservaHotel -count=1` (ok).
		- `go test ./handlers -run ReservasHotel -count=1` (ok).

## 2026-04-05
- Se crea el modulo de registro de vehiculos por empresa para controlar ingreso y salida por patente.
	- `backend/db/vehiculos_registro.go` (nuevo) agrega esquema y operaciones CRUD del registro vehicular, con estado operativo (`en_empresa`/`retirado`) y marcacion de salida.
	- `backend/handlers/vehiculos_registro.go` (nuevo) expone `/api/empresa/vehiculos_registro` con acciones de consulta, alta, edicion, activar/desactivar, marcar salida y eliminacion.
	- `backend/main.go` asegura esquema del modulo, registra migracion `2026-04-05-014-vehiculos-registro` y publica ruta protegida bajo permisos de seguridad.
	- `web/administrar_empresa/vehiculos_registro.html` (nuevo) incorpora UI de registro de vehiculos con patente, conductor, propietario, fechas de ingreso/salida y filtros operativos.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso lateral `Registro de vehiculos` con permisos por rol.
	- Cobertura agregada en `backend/db/vehiculos_registro_test.go` y `backend/handlers/vehiculos_registro_test.go`.
	- Validaciones ejecutadas:
		- `go test ./db -run Vehiculo -count=1` (ok).
		- `go test ./handlers -run VehiculosRegistro -count=1` (ok).

## 2026-04-05
- Se crea el modulo de propinas por empresa con configuracion operativa y reporte por usuario o universal.
	- `backend/db/propinas.go` (nuevo) agrega tablas de configuracion y movimientos de propinas, con soporte de reporte acumulado por usuario y reparto universal entre usuarios activos.
	- `backend/handlers/propinas.go` (nuevo) expone `/api/empresa/propinas` con acciones de configuracion, reporte y consulta de movimientos.
	- `backend/handlers/carritos_compras.go` integra propina en `action=pagar_estacion`, valida `total_pagado` contra total final con propina y registra movimiento de propina al cerrar venta.
	- `backend/main.go` asegura esquema de propinas, registra migracion `2026-04-05-013-propinas` y publica ruta protegida `/api/empresa/propinas` bajo permisos de finanzas.
	- `web/administrar_empresa/propinas.html` (nuevo) incorpora modulo de configuracion de propinas y reporte por rango, usuario y modo.
	- `web/administrar_empresa/carrito_de_compras.html` agrega control de aplicar propina en cobro de estacion, carga de configuracion y desglose de total final con propina.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran acceso de menu `Propinas` con permisos del modulo finanzas.
	- Cobertura agregada en `backend/db/propinas_test.go`, `backend/handlers/propinas_test.go` y `backend/handlers/auth_users_carritos_test.go`.
	- Validaciones ejecutadas:
		- `go test ./db -run Propina -count=1` (ok).
		- `go test ./handlers -run "Propinas|CarritosCompraAplicaPropinaSegunConfiguracion" -count=1` (ok).
		- `go test ./db ./handlers -count=1` (ok).

## 2026-04-05
- Se agrega `transferencia_bancaria` como forma de pago transversal en flujo de carritos y finanzas.
	- `backend/db/carritos_compras.go` normaliza y acepta alias de transferencia bancaria (`transferencia`, `transferencia_bancaria`).
	- `backend/handlers/carritos_compras.go` habilita transferencia bancaria en pago directo y mixto, y exige referencia minima para tarjeta/transferencia.
	- `backend/handlers/auth_users_carritos_test.go` agrega cobertura de pago exitoso por transferencia bancaria y rechazo cuando falta referencia valida.
	- `web/administrar_empresa/carrito_de_compras.html` incorpora transferencia bancaria en selectores de pago, habilita validacion de pago mixto con transferencia y envÃƒÂ­a `pagos_mixtos` al backend.
	- `web/administrar_empresa/finanzas.html` estandariza opcion de `transferencia_bancaria` y mantiene compatibilidad con registros legacy `transferencia`.
	- `web/ayuda/ayuda.html` actualiza descripcion de metodos soportados en cierre de carrito.
	- `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md` reflejan el nuevo flujo de pago.

## 2026-04-05
- Robustecimiento del modulo de auditoria empresarial con foco en trazabilidad, seguridad operativa y analisis forense.
	- `backend/db/auditoria_empresa.go` amplÃƒÂ­a filtros (`metodo_http`, `recurso`, `endpoint`, `search`), agrega `offset`, agrega conteo filtrado (`CountEmpresaAuditoriaEventos`) y refuerza indices de rendimiento.
	- `backend/handlers/auditoria_empresa.go` valida fechas/parametros, publica metadata de paginacion por headers y soporta consulta avanzada de eventos.
	- `backend/handlers/empresa_permisos.go` registra intentos criticos denegados (401/403/500) como eventos de auditoria no bloqueantes.
	- `backend/utils/utils.go` expone `RequestIDFromContext` para correlacion real entre logs de request y eventos de auditoria.
	- `web/administrar_empresa/auditoria.html` agrega filtros avanzados, paginador y panel de detalle JSON por evento.
	- `web/estilos.css` agrega estilos centralizados para paginacion y detalle del modulo de auditoria.
	- Se amplian pruebas en `backend/db/auditoria_empresa_test.go` y `backend/handlers/auditoria_empresa_test.go`.
	- Validaciones ejecutadas:
		- `go test ./db -run Auditoria -count=1` (ok).
		- `go test ./handlers -run Auditoria -count=1` (ok).
		- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).

## 2026-04-05
- FacturaciÃƒÂ³n electrÃƒÂ³nica: envÃƒÂ­o automÃƒÂ¡tico del resumen de factura al correo del cliente al emitir.
	- `backend/handlers/facturacion_electronica.go` ahora intenta enviar correo en `action=emitir` de `factura_electronica`.
	- Soporta destinatario por `cliente_email` o por `cliente_id`/`entidad_id` consultando clientes.
	- La respuesta incluye bloque `factura_email` con estado de intento/envÃƒÂ­o/error sin bloquear la emisiÃƒÂ³n legal.
	- `backend/db/clientes.go` agrega `GetClienteByID` para resolver destinatario desde la base de datos.
	- `backend/main.go` actualiza la inyecciÃƒÂ³n de `dbSuper` al handler de facturaciÃƒÂ³n para lectura de SMTP.
	- `web/administrar_empresa/facturacion_electronica.html` agrega campos de cliente y muestra el resultado de envÃƒÂ­o en pantalla.
	- Cobertura aÃƒÂ±adida en `backend/db/clientes_test.go` y `backend/handlers/eventos_contables_modulos_test.go`.

## 2026-04-05
- Se crea el modulo de codigos de descuento por empresa y validacion de metodos de pago en carrito de compras.
	- `backend/db/codigos_descuento.go` (nuevo) agrega la tabla `codigos_de_descuento`, generacion automatica de codigos, CRUD, validacion por vencimiento/usos y resoluciÃƒÂ³n de descuento aplicable por monto.
	- `backend/handlers/codigos_descuento.go` (nuevo) expone `/api/empresa/codigos_de_descuento` con operaciones CRUD, activar/desactivar y `action=validar`.
	- `backend/db/carritos_compras.go` agrega campos `metodo_pago` y `referencia_pago`, normaliza metodos permitidos y registra consumo transaccional de codigo de descuento al cerrar venta.
	- `backend/handlers/carritos_compras.go` valida `metodo_pago` (`efectivo`, `tarjeta_credito`, `tarjeta_debito`, `codigo_descuento`) y exige referencia para pagos con tarjeta.
	- `backend/main.go` asegura esquema `codigos_de_descuento`, registra migracion `2026-04-05-012-codigos-descuento-pagos` y expone ruta protegida de codigos de descuento.
	- `web/administrar_empresa/codigos_de_descuento.html` (nuevo) incorpora modulo profesional para crear/editar/activar/eliminar codigos con valor y fecha de vencimiento.
	- `web/administrar_empresa/carrito_de_compras.html` agrega selector de metodo de pago, referencia y aplicacion de codigos de descuento con validacion operativa.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el enlace de menu `Codigos de descuento` con permisos del modulo ventas.
	- `backend/db/codigos_descuento_test.go` y `backend/handlers/auth_users_carritos_test.go` agregan cobertura para validacion/uso de codigos y rechazo de metodo de pago invalido.

## 2026-04-05
- Se crea el modulo de combos de productos con receta de ingredientes y precio unico de venta.
	- `backend/handlers/combos_productos.go` (nuevo) expone `/api/empresa/combos_productos` con operaciones CRUD y acciones `activar/desactivar`.
	- `backend/db/productos.go` incorpora esquema y logica de combos (`combos_productos`, `combos_productos_detalle`) con controles de consistencia para carritos abiertos.
	- `backend/db/carritos_compras.go` extiende el ajuste de inventario para descontar/liberar stock por ingrediente cuando el item es `tipo_item=combo`.
	- `backend/handlers/carritos_compras.go` valida `referencia_id` obligatorio para items combo.
	- `backend/main.go` registra la nueva ruta protegida bajo permisos de inventario.
	- `web/administrar_empresa/combos_productos.html` (nuevo) agrega interfaz completa para gestionar combos y receta.
	- `web/administrar_empresa/carrito_de_compras.html` incorpora busqueda/catalogo y visualizacion de combos en carrito.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el modulo en menu y permisos.
	- `backend/db/productos_categorias_test.go` y `backend/db/carritos_inventario_test.go` agregan cobertura de CRUD y flujo de inventario por ingredientes.

## 2026-04-05
- Se crea el modulo de graficos y estadisticas por empresa.
	- `backend/handlers/graficos_estadisticas.go` (nuevo) expone `/api/empresa/graficos_estadisticas` con acciones `panel`, `serie`, `rankings`, `distribuciones` y `catalogo`.
	- `backend/main.go` registra la nueva ruta protegida bajo permisos de finanzas.
	- `backend/handlers/graficos_estadisticas_test.go` (nuevo) agrega cobertura de contrato HTTP y validaciones de error.
	- `web/administrar_empresa/graficos_estadisticas.html` (nuevo) incorpora panel visual con series, distribuciones y rankings.
	- `web/estilos.css` agrega estilos responsivos del nuevo modulo.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso en menu con control de permisos.
	- `web/ayuda/ayuda.html` incorpora guia y API del modulo de analitica.

## 2026-04-05
- Se crea el modulo de control de asistencia de empleados por empresa.
	- `backend/db/asistencia_empleados.go` (nuevo) agrega tabla `empresa_asistencia_empleados` y operaciones CRUD con marcacion de entrada/salida.
	- `backend/handlers/asistencia_empleados.go` (nuevo) expone `/api/empresa/asistencia_empleados` con acciones operativas de asistencia.
	- `backend/main.go` incorpora esquema, migracion `2026-04-05-010-asistencia-empleados` y registro de ruta protegida.
	- `web/administrar_empresa/asistencia_empleados.html` (nuevo) agrega UI completa para gestion diaria de asistencia.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el modulo en menu y permisos.
	- `backend/handlers/asistencia_empleados_test.go` (nuevo) valida flujo funcional del modulo.
	- Se actualizan `web/ayuda/ayuda.html`, `estructura_bd.md` y diagramas/documentacion tecnica para trazabilidad.

## 2026-04-05
- Modulo de reportes robustecido a nivel empresarial, operativo y contable con enfoque escalable por dataset.
	- `backend/handlers/reportes.go` (nuevo) implementa `/api/empresa/reportes` con acciones `catalogo`, `suite`, `dataset`, `tablero` y `export`.
	- Se habilitan exportaciones multi-formato para datasets: `JSON`, `CSV`, `TXT` y `XLS`.
	- `backend/main.go` registra la nueva ruta protegida bajo permisos de finanzas.
	- `web/administrar_empresa/reportes.html` incorpora selector de dataset, vista tabular profesional y exportes desde interfaz.
	- `backend/handlers/reportes_test.go` (nuevo) agrega cobertura de contrato HTTP y validacion de exportaciones.
	- Se actualizan diagramas de arquitectura/flujo en `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md`.

## 2026-04-04
- Centro de ayuda actualizado con tutorial por cada mÃƒÂ³dulo del sistema.
	- `web/ayuda/ayuda.html` amplÃƒÂ­a el contenido con una secciÃƒÂ³n de tutoriales por mÃƒÂ³dulos de administraciÃƒÂ³n global y mÃƒÂ³dulos del panel de empresa.
	- Se agregan pasos operativos por mÃƒÂ³dulo y enlaces directos a cada pantalla para facilitar onboarding y uso diario.

## 2026-04-04
- Verificacion integral real de modulos + limpieza de artefactos temporales.
	- Validacion real ejecutada (sin simulaciones/mocks) sobre motor legado retirado y capa HTTP:
		- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).
		- `go test ./... -count=1` (ok).
		- `go test ./handlers -run "TestEmpresaScope|FueraDeAlcance|WithEmpresa|isol|Aisla|multiempresa|UsuariosHandlerAislaEmpresa|ConsolidaEmpresa" -count=1` (ok).
		- `go test ./handlers -run "TestEmpresaClientes|TestEmpresaProveedores|TestEmpresaFacturacion|TestEmpresaCompras|TestEmpresaInventario|TestEmpresaFinanzas|TestEmpresaAuditoria|TestEmpresaCarritos|TestEmpresaUsuarios|TestModelosHandler" -count=1` (ok).
		- `go test ./db -run "Test.*(Cliente|Proveedor|Facturacion|Compra|Inventario|Finanzas|Evento|Auditoria|Scope|Empresa)" -count=1` (ok).
	- Se eliminan artefactos temporales/no usados del repositorio:
		- `backend/tmp_api.json`.
		- `backend/tmp_config.html`.
		- `backend/server.err`.
		- `backend/server.run.err`.
		- `backend/db/pcs_empresas.20260326-174525.bak`.
		- `backend/db/pcs_superadministrador.20260326-174324.bak`.
		- `backend/db/pcs_superadministrador.20260326-174525.bak`.

## 2026-04-04
- Punto 14 (operacion continua) - inicio operativo con KPI y roadmap trimestral.
	- `documentos/punto_14_operacion_continua.md` (nuevo): define marco de mejora continua y cadencia de seguimiento.
	- `documentos/roadmap_trimestral_pos_multiempresa.md` (nuevo): formaliza roadmap Q2/Q3/Q4 2026.
	- `scripts/generar_reporte_operacion_continua.ps1` (nuevo): genera reporte operativo y bitacora tecnica.
	- `documentos/punto_14_operacion_continua_reporte.md` (nuevo): evidencia de la ultima corrida operativa.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md`: punto 14 actualizado a `en curso`.
- Validacion tecnica:
	- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\generar_reporte_operacion_continua.ps1` (ok).

## 2026-04-04
- Punto 13 (calidad, UAT y despliegue) - arranque operativo con validacion integral automatizada.
	- `scripts/validar_punto_13.ps1` (nuevo): ejecuta gate tecnico y genera evidencia automatica.
	- `documentos/punto_13_calidad_uat_despliegue.md` (nuevo): formaliza flujo de calidad/UAT/salida controlada.
	- `documentos/punto_13_validacion_integral_resultado.md` (nuevo): reporte de ultima validacion tecnica.
	- `documentos/release_checklist.md`: incorpora gate del punto 13 y verificacion de evidencia.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md`: punto 13 pasa a `en curso`.
- Validacion tecnica:
	- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\validar_punto_13.ps1` (ok).
	- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).
	- `go test ./... -count=1` (ok).

## 2026-04-04
- Punto 8 (facturacion electronica) - refuerzo de cobertura en cumplimiento legal de emision.
	- `backend/db/facturacion_electronica_test.go` (nuevo) agrega pruebas unitarias para `PrepareFacturacionDocumentoLegal`:
		- `TestPrepareFacturacionDocumentoLegalSuccessAndConsecutivo`.
		- `TestPrepareFacturacionDocumentoLegalRejectsExpiredResolution`.
		- `TestPrepareFacturacionDocumentoLegalRejectsConfigInactivaAndRangoAgotado`.
	- Se valida reserva e incremento de consecutivo legal, rechazo por resolucion vencida, rechazo por configuracion FE inactiva y agotamiento de rango.
- Validacion tecnica:
	- `gofmt -w db/facturacion_electronica_test.go` (ok).
	- `go test ./db -run "TestPrepareFacturacionDocumentoLegal" -count=1` (ok).
	- `go test ./db ./handlers -run "TestPrepareFacturacionDocumentoLegal|TestEmpresaDocumentoFacturacionUpsertAndGet|TestEmpresaFacturacionTransaccional" -count=1` (ok).

## 2026-04-04
- Punto 9 (modulo de compras) - avance funcional con endpoint y vista dedicados para ciclo documental.
	- `backend/db/documentos_transaccionales.go` agrega:
		- `ListEmpresaDocumentosCompraByEmpresa`.
		- `SetEmpresaDocumentoCompraEstadoByCodigo`.
	- `backend/handlers/compras.go` (nuevo) implementa `GET/POST/PUT/DELETE /api/empresa/compras/documentos` con acciones documentales (`crear`, `emitir_orden`, `recepcionar_compra`, `contabilizar_compra`) y activar/desactivar.
	- `backend/main.go` registra la ruta protegida `/api/empresa/compras/documentos`.
	- `web/administrar_empresa/compras.html` (nuevo) incorpora interfaz dedicada de compras para crear, consultar y transicionar documentos.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso de menu `Compras` con control por permisos de modulo.
	- Cobertura agregada en:
		- `backend/db/documentos_transaccionales_test.go`.
		- `backend/handlers/compras_documentos_test.go` (nuevo).
- Validacion tecnica:
	- `gofmt -w handlers/compras.go handlers/compras_documentos_test.go main.go db/documentos_transaccionales.go db/documentos_transaccionales_test.go` (ok).
	- `go test ./db -run "TestEmpresaDocumentoCompraListAndSetEstadoByCodigo" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaComprasDocumentos" -count=1` (ok).
	- `go test ./db ./handlers -run "TestEmpresaDocumentoCompraListAndSetEstadoByCodigo|TestEmpresaComprasDocumentos" -count=1` (ok).
	- `go test ./... -run "TestEmpresaComprasDocumentos|TestEmpresaDocumentoCompraListAndSetEstadoByCodigo" -count=1` (ok).

## 2026-04-04
- Punto 8 (facturacion electronica) - avance funcional de emision legal y cumplimiento normativo inicial.
	- `backend/db/facturacion_electronica.go` agrega `PrepareFacturacionDocumentoLegal` para validar configuracion legal, vigencia de resolucion y rango de consecutivos por empresa/pais antes de emitir.
	- `backend/db/documentos_transaccionales.go` amplia `empresa_facturacion_documentos` con metadata legal persistida: `numero_legal`, `codigo_validacion`, `pais_codigo`, `ambiente_fe`.
	- `backend/handlers/facturacion_electronica.go` endurece `action=emitir` con rechazo `422` cuando no hay cumplimiento normativo y devuelve bloque `cumplimiento_normativo` en emisiones exitosas.
	- `web/administrar_empresa/facturacion_electronica.html` incorpora bloque operativo para `emitir`, `anular` y `nota_credito`, con visualizacion del resultado legal.
	- Cobertura extendida en:
		- `backend/db/documentos_transaccionales_test.go`.
		- `backend/handlers/eventos_contables_modulos_test.go`.
- Validacion tecnica:
	- `go test ./db -run "TestEmpresaDocumentoFacturacionUpsertAndGet" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaFacturacionTransaccionalEmiteEventosContables|TestEmpresaFacturacionTransaccionalEmitirRechazaSinCumplimientoLegal|TestEmpresaFacturacionTransaccionalRechazaTransicionInvalida" -count=1` (ok).
	- `go test ./db ./handlers -count=1` (ok).

## 2026-04-04
- Punto 7 (gestion de proveedores) - avance funcional de catalogo, precios y condiciones comerciales.
	- `backend/db/productos.go` amplia el modelo `Proveedor` y su migracion segura con campos:
		- `catalogo_referencia`,
		- `precio_base_referencial`,
		- `descuento_porcentaje`,
		- `plazo_pago_dias`,
		- `condicion_entrega`.
	- `backend/handlers/productos.go` agrega validacion HTTP de rango para los nuevos campos en `POST/PUT /api/empresa/proveedores` y enriquece metadata de eventos contables de compras.
	- `web/administrar_empresa/administrar_productos.html` amplia el formulario y la tabla de proveedores para gestionar y visualizar datos comerciales.
	- Cobertura nueva/extendida en:
		- `backend/db/productos_categorias_test.go`.
		- `backend/handlers/eventos_contables_modulos_test.go`.
- Validacion tecnica:
	- `gofmt -w db/productos.go db/productos_categorias_test.go handlers/productos.go handlers/eventos_contables_modulos_test.go` (ok).
	- `go test ./db -run "TestProveedorCRUDIncluyeCatalogoPreciosYCondiciones" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaProveedoresEmiteEventoContableCompras|TestEmpresaProveedoresRechazaCamposComercialesInvalidos" -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 6 (gestion de clientes) - avance funcional de perfil, historial y segmentacion.
	- `backend/db/clientes.go` agrega contratos analiticos (`ClientePerfilComercial`, `ClienteCompraHistorial`, `ClienteSegmentacionResumen`) y funciones de consulta por cliente/empresa.
	- `backend/handlers/clientes.go` amplia `GET /api/empresa/clientes` con `action=perfil`, `action=historial`, `action=segmentacion|segmentos`.
	- `web/administrar_empresa/administrar_clientes.html` agrega paneles de segmentacion y de perfil/historial por cliente con accion `Perfil`.
	- Cobertura nueva en:
		- `backend/db/clientes_test.go`.
		- `backend/handlers/clientes_test.go`.
- Validacion tecnica:
	- `gofmt -w db/clientes.go db/clientes_test.go handlers/clientes.go handlers/clientes_test.go` (ok).
	- `go test ./db -run "TestGetClientePerfilComercialByEmpresaAndHistorial|TestGetClientePerfilComercialByEmpresaSinComprasSegmentoNuevo" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaClientesHandlerPerfilHistorialSegmentacion" -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ continuidad preventiva-compras ciclo documental desde reposicion.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionOrdenEstadoActualizado` y `ActualizarEstadoOrdenCompraDesdeReposicion` para transiciones `recepcionar_compra` y `contabilizar_compra`.
	- `backend/handlers/productos.go` agrega endpoint `POST /api/empresa/compras/plan_reposicion/actualizar_estado`.
	- `backend/main.go` registra `/api/empresa/compras/plan_reposicion/actualizar_estado` bajo permisos de compras.
	- `web/administrar_empresa/administrar_productos.html` amplÃƒÂ­a el flujo a `fases 10-12` con acciones `Recepcionar orden` y `Contabilizar orden` y contexto de estado de OC.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./db -run "TestEmitirOrdenCompraDesdePlanReposicionBorradorPersistDoc|TestActualizarEstadoOrdenCompraDesdeReposicionCiclo"` (ok).
	- `go test ./handlers -run "TestEmpresaComprasPlanReposicionEmitirOrdenHandlerEmiteDocumento|TestEmpresaComprasPlanReposicionActualizarEstadoHandlerGestionaCiclo"` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ continuidad preventiva-compras emitible desde borrador.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionOrdenEmitida` y `EmitirOrdenCompraDesdePlanReposicionBorrador` para emitir OC desde el borrador y persistirla en documentos de compras.
	- `backend/handlers/productos.go` agrega endpoint `POST /api/empresa/compras/plan_reposicion/emitir_orden`.
	- `backend/main.go` registra `/api/empresa/compras/plan_reposicion/emitir_orden` bajo permisos de compras.
	- `web/administrar_empresa/administrar_productos.html` agrega accion `Emitir orden` en el bloque de borrador (fase 10).
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ continuidad preventiva-compras ordenable por proveedor.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionBorradorItem`, `InventarioPlanReposicionBorradorCompra` y `GetInventarioPlanReposicionBorradorByEmpresa` para generar borradores de orden por proveedor con detalle y totales.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/plan_reposicion_borrador`.
	- `backend/main.go` registra `/api/empresa/inventario/plan_reposicion_borrador` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega bloque `Borrador de orden de compra por proveedor (fase 10)` y accion `Borrador OC` desde consolidado fase 9.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ continuidad preventiva-compras consolidada por proveedor.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionProveedorResumen` y `GetInventarioPlanReposicionResumenByEmpresa` para consolidar compra preventiva por proveedor.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/plan_reposicion_resumen`.
	- `backend/main.go` registra `/api/empresa/inventario/plan_reposicion_resumen` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Consolidado de compra por proveedor (fase 9)` y filtro de items del plan por proveedor.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ continuidad preventiva-compras con plan de reposicion por proveedor.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionItem` y `GetInventarioPlanReposicionByEmpresa` para consolidar sugerencias por proveedor con costo estimado.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/plan_reposicion` con validaciones operativas.
	- `backend/main.go` registra `/api/empresa/inventario/plan_reposicion` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Plan de reposicion por proveedor (fase 8)` con resumen de costo estimado y accion `Preparar`.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ continuidad preventiva con proyeccion de quiebre.
	- `backend/db/productos.go` agrega `InventarioProyeccionQuiebre` y `GetInventarioProyeccionQuiebreByEmpresa` para estimar consumo diario, cobertura y sugerido de reposicion por producto/bodega.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/proyeccion_quiebre` con validacion de `dias_ventana`, `bodega_id`, `limit` y `offset`.
	- `backend/main.go` registra `/api/empresa/inventario/proyeccion_quiebre` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Proyeccion de quiebre (preventiva)` y accion `Preparar` para reposicion preventiva guiada.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ continuidad operativa-analitica con balance por bodega.
	- `backend/db/productos.go` agrega `InventarioBalanceBodega` y `GetInventarioBalanceBodegasByEmpresa` para consolidar entradas/salidas/traslados/neto por bodega en rango.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/balance_bodegas` con validacion de fechas y filtros por bodega/rango.
	- `backend/main.go` registra `/api/empresa/inventario/balance_bodegas` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Balance por bodega` y contexto de neto acumulado sincronizado con filtros del kardex.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ continuidad analitica con tendencia diaria.
	- `backend/db/productos.go` agrega `InventarioTendenciaDia` y `GetInventarioTendenciaByEmpresa` para serie diaria por empresa con filtros por bodega/rango.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/tendencia` con validacion de fechas y ventana por `dias`.
	- `backend/main.go` registra `/api/empresa/inventario/tendencia` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla `Tendencia diaria inventario` y contexto de neto acumulado/eventos sincronizado con filtros del kardex.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ continuidad operacional en panel de productos.
	- `web/administrar_empresa/administrar_productos.html` agrega:
		- bloque `Top productos crÃƒÂ­ticos (dÃƒÂ©ficit)` alimentado desde alertas de inventario,
		- priorizaciÃƒÂ³n de crÃƒÂ­ticos por `sin_stock` y mayor dÃƒÂ©ficit,
		- acciÃƒÂ³n `Preparar reposiciÃƒÂ³n` para precargar ajuste de inventario con producto, bodega y cantidad sugerida.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ continuidad KPI operativo en panel de productos.
	- `backend/db/productos.go` agrega `InventarioResumen` y `GetInventarioResumenByEmpresa` para consolidar existencias, alertas y movimientos por rango.
	- `backend/handlers/productos.go` agrega endpoint `GET /api/empresa/inventario/resumen` con validacion de fechas `YYYY-MM-DD`.
	- `backend/main.go` registra `/api/empresa/inventario/resumen` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega KPI visibles de inventario e integra consumo del resumen segun rango del kardex.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`,
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./handlers ./db -count=1` (ok).
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ continuidad UI operativa en panel de productos.
	- `web/administrar_empresa/administrar_productos.html` agrega:
		- filtro por bodega para alertas de quiebre,
		- filtros de kardex por bodega, tipo y rango de fechas,
		- acciones `Filtrar` y `Limpiar` en ambos bloques de consulta.
	- Se actualiza documentacion asociada en plan maestro y estructura tecnica.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) Ã¢â‚¬â€ inicio tecnico: kardex operativo + reglas de stock + alertas de quiebre por bodega.
	- `backend/db/productos.go`:
		- valida `stock_minimo/stock_maximo` en creacion y edicion de productos,
		- agrega `GetAlertasQuiebreByEmpresa`,
		- amplÃƒÂ­a `GetMovimientosByEmpresa` con filtros `bodega_id`, `tipo`, `desde`, `hasta`.
	- `backend/handlers/productos.go`:
		- nuevo endpoint `GET /api/empresa/inventario/alertas`,
		- compatibilidad `action=alertas|alertas_quiebre|quiebre` en existencias,
		- filtros de kardex + validacion de fechas `YYYY-MM-DD` en movimientos.
	- `backend/main.go` registra `/api/empresa/inventario/alertas` bajo permisos de inventario.
	- `web/administrar_empresa/administrar_productos.html` agrega tabla de alertas de quiebre por bodega.
	- `documentos/descripcion_del_proyecto` actualiza la descripcion de inventario con alertas de quiebre y kardex filtrable.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`,
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `runTests` en archivos de prueba modificados (ok).
	- `go test ./handlers ./db -count=1` en `backend` (ok).

## 2026-04-04
- Punto 3 (permisos y seguridad) Ã¢â‚¬â€ continuidad operativa: catalogo frontend por rol + regresion endpoints sin wrapper.
	- `web/js/administrar_empresa.js` agrega catalogo de permisos por enlace y aplica ocultamiento de opciones no autorizadas segun rol autenticado (`GET /me`).
	- Se agrega fallback de navegacion en iframe cuando la ultima pagina guardada no es visible para el rol actual.
	- `backend/handlers/auth_users_carritos_test.go` agrega regresiones de alcance por `empresa_id` para:
		- `POST /api/empresa/usuarios/login`.
		- `POST /api/empresa/usuarios/establecer_password`.
	- `backend/handlers/chat_con_inteligencia_artificial_controller_test.go` agrega regresion de alcance por cuenta Google en `ModelosHandler`.
	- Se actualiza documentacion tecnica en:
		- `documentos/diagramas/diagrama_roles_permisos.md`.
		- `documentos/diagramas/estructura_del_codigo.md`.
- Validacion tecnica:
	- `runTests` sobre `backend/handlers/auth_users_carritos_test.go` y `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`.
	- resultado: 14 pruebas aprobadas, 0 fallidas.
	- `get_errors` sobre `web/js/administrar_empresa.js`: sin errores.

## 2026-04-04
- Punto 3 (permisos y seguridad) Ã¢â‚¬â€ consolidacion documental endpoint/rol y checklist UAT:
	- `documentos/matriz_roles_permisos_pos_multiempresa.md` agrega matriz final endpoint/rol alineada con wrappers reales y reglas por accion.
	- Se documentan endpoints fuera de wrapper con control alterno por handler/cuenta Google.
	- Se agrega checklist UAT de punto 3 con evidencia automatizada.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` agrega seccion de consolidacion con estado operativo y pendientes de cierre total.
- Validacion tecnica:
	- `runTests` sobre `backend/handlers/empresa_permisos_test.go` y `backend/handlers/auditoria_empresa_test.go`.
	- resultado: 25 pruebas aprobadas, 0 fallidas.

## 2026-04-04
- Ajuste editorial de consistencia documental (plan maestro):
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` corrige `Backlog inmediato` para reflejar cierre real de Punto 1 y Punto 2.
	- El backlog siguiente queda enfocado en Punto 3 (permisos y seguridad) y Punto 5 (control de inventarios).
- Validacion tecnica:
	- cambio documental (sin ejecucion de pruebas automatizadas).

## 2026-04-04
- Punto 1 + Punto 2 (plan maestro) Ã¢â‚¬â€ cierre de backlog inmediato con formalizacion tecnica documental.
	- `documentos/matriz_kpi_pos_multiempresa.md` se actualiza a formato formal con:
		- formula implementada por KPI,
		- endpoint canonico de lectura/exportacion,
		- tablas fuente reales por metrica.
	- Se crea `documentos/matriz_entidades_multiempresa_aislamiento.md` con matriz de aislamiento por endpoint:
		- llave primaria `empresa_id`,
		- llaves secundarias por recurso,
		- mecanismo de control de alcance (middleware o validacion interna).
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` marca Punto 1 y Punto 2 como `completado`.
- Validacion tecnica:
	- cambio documental (sin ejecucion de pruebas automatizadas).

## 2026-04-04
- Punto 11 (reportes financieros) Ã¢â‚¬â€ continuidad de backlog inmediato: exportacion unificada del tablero por rango.
	- `backend/handlers/finanzas.go` agrega `action=tablero_export` en `GET /api/empresa/finanzas/movimientos` con:
		- `format=json` para payload unificado del tablero,
		- `format=csv` para matriz unificada por bloque/metrica/valor.
	- La exportacion integra bloques `estado_resultados` y `balance_general` junto con KPI operativos/financieros/contables.
	- `web/administrar_empresa/reportes.html` incorpora botones:
		- `Exportar tablero CSV`,
		- `Exportar tablero JSON`.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega `TestEmpresaFinanzasTableroResumenExportHandler`.
- Validacion tecnica:
	- `go test ./handlers -run "TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasTableroResumenExportHandler|TestEmpresaFinanzasAsientosContablesHandlerConciliacionPeriodo" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 10 (modulo contable integrado) Ã¢â‚¬â€ continuidad de backlog inmediato: vista de conciliacion por periodo (eventos vs asientos).
	- `backend/db/eventos_contables.go` agrega modelos y funcion `GetEmpresaConciliacionContablePorPeriodo` para consolidar por periodo:
		- eventos totales/procesados/pendientes/con error,
		- asientos generados,
		- desfase de conteo y desfase de monto,
		- estado de conciliacion por periodo.
	- `backend/handlers/finanzas.go` agrega `GET /api/empresa/finanzas/asientos_contables?action=conciliacion_periodo|conciliacion`.
	- `web/administrar_empresa/finanzas.html` incorpora vista de conciliacion con filtros, KPIs y tabla comparativa por periodo.
	- `backend/db/eventos_contables_test.go` agrega prueba de conciliacion por periodo.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega prueba del endpoint de conciliacion.
- Validacion tecnica:
	- `go test ./db -run "EventosContables|ConPolitica|Conciliacion" -count=1` (ok).
	- `go test ./handlers -run "AsientosContablesHandler|ConciliacionPeriodo" -count=1` (ok).
	- `go test ./db -count=1` (ok).
	- `go test ./handlers -count=1` (ok).

## 2026-04-04
- Punto 10 (modulo contable integrado) Ã¢â‚¬â€ continuidad de backlog inmediato: ejecucion automatica por lotes de asientos.
	- `backend/db/eventos_contables.go` agrega:
		- `ProcessEmpresaEventosContablesPendientesConPolitica` con soporte de `max_reintentos`,
		- `RunEmpresaAsientosContablesWorkerCycle`,
		- `StartEmpresaAsientosContablesWorker`.
	- `backend/main.go` integra worker automatico de asientos con politica configurable por entorno:
		- `ASIENTOS_WORKER_INTERVAL_MINUTES`,
		- `ASIENTOS_WORKER_BATCH_SIZE`,
		- `ASIENTOS_WORKER_MAX_RETRIES`.
	- `backend/handlers/finanzas.go` permite `max_reintentos` opcional en proceso manual de `/api/empresa/finanzas/asientos_contables?action=procesar_asientos`.
	- `backend/db/eventos_contables_test.go` agrega prueba de politica de reintentos.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega validacion `400` para `max_reintentos` invalido y cobertura del parametro.
- Validacion tecnica:
	- `go test ./db -run "EventosContables|ConPolitica|Asientos" -count=1` (ok).
	- `go test ./handlers -run "AsientosContablesHandler|FinanzasAsientos" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 15 (auditoria por empresa) Ã¢â‚¬â€ continuacion de backlog inmediato 1 y 2:
	- `backend/db/auditoria_empresa.go` agrega filtros avanzados de consulta por `recurso_id` y `codigo_http` en `ListEmpresaAuditoriaEventos`.
	- `backend/handlers/auditoria_empresa.go` valida y expone nuevos filtros en `GET /api/empresa/auditoria/eventos`:
		- `recurso_id`.
		- `codigo_http`.
	- `web/administrar_empresa/auditoria.html` incorpora:
		- filtros avanzados por `codigo_http` y `recurso_id`,
		- exportacion de resultados filtrados a `CSV` y `JSON`.
	- `backend/db/auditoria_empresa_test.go` fortalece cobertura de listado con filtros avanzados.
	- `backend/handlers/auditoria_empresa_test.go` agrega `TestEmpresaAuditoriaEventosHandlerFiltrosAvanzados` para contrato HTTP y validacion de parametros invalidos.
- Validacion tecnica:
	- `go test ./db -run "Auditoria" -count=1` (ok).
	- `go test ./handlers -run "Auditoria" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 15 (auditoria por empresa) Ã¢â‚¬â€ continuacion de backlog 1, 2 y 3:
	- `backend/handlers/empresa_permisos.go` refuerza clasificacion de acciones criticas en `ventas`, `compras` y `facturacion` (alias operativos de aprobacion/eliminacion).
	- `backend/handlers/auditoria_empresa.go` amplia metadata de trazabilidad para recursos de ventas/compras/facturacion (`carrito_id`, `proveedor_id`, `entidad_id`, `documento_codigo`).
	- `backend/handlers/auditoria_empresa_test.go` agrega pruebas de registro automatico de auditoria en acciones criticas de:
		- ventas (`action=cerrar`),
		- compras (`action=emitir_orden`),
		- facturacion (`action=emitir`).
	- `web/administrar_empresa/auditoria.html` agrega vista de consulta filtrable y retencion manual para auditoria por empresa.
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` agregan acceso del menu lateral a la nueva vista `Auditoria`.
	- `backend/db/auditoria_empresa.go` agrega:
		- purga automatica por expiracion (`PurgeExpiredEmpresaAuditoriaEventos`),
		- worker programado (`StartEmpresaAuditoriaRetentionWorker`),
		- calculo de `fecha_expiracion` alineado a `fecha_evento` cuando se provee.
	- `backend/main.go` arranca worker de retencion automatica de auditoria (intervalo 12h).
	- `backend/db/auditoria_empresa_test.go` agrega prueba de purga automatica por expiracion.
- Validacion tecnica:
	- `go test ./handlers -run "Auditoria|WithEmpresa(Ventas|Compras|Facturacion|Finanzas)Permissions" -count=1` (ok).
	- `go test ./db -run "Auditoria" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 15 (auditoria por empresa) Ã¢â‚¬â€ implementacion base minima:
	- `backend/db/auditoria_empresa.go` agrega tabla `empresa_auditoria_eventos`, filtros de consulta y purga por retencion.
	- `backend/handlers/auditoria_empresa.go` agrega endpoint protegido:
		- `GET /api/empresa/auditoria/eventos`.
		- `PUT/POST /api/empresa/auditoria/eventos?action=retener|purgar`.
	- `backend/handlers/empresa_permisos.go` integra registro automatico no bloqueante para acciones criticas (`C/U/D/A`).
	- `backend/main.go` integra `EnsureEmpresaAuditoriaSchema`, migracion `2026-04-04-011-auditoria-empresa` y ruta de auditoria.
	- Pruebas nuevas: `backend/db/auditoria_empresa_test.go` y `backend/handlers/auditoria_empresa_test.go`.
- Validacion tecnica:
	- `go test ./db -run "Auditoria|EventosContables|ReportesTableroResumen" -count=1` (ok).
	- `go test ./handlers -run "Auditoria|AsientosContables|WithEmpresaFinanzasPermissions" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Plan maestro POS multiempresa:
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` se actualiza de 14 a 15 puntos.
	- Se incorpora el nuevo `Punto 15: Modulo de auditoria por empresa` con alcance, entregables iniciales, backlog y criterio de avance.
	- `documentos/descripcion_del_proyecto` se alinea para referenciar el plan de 15 puntos.
- Validacion tecnica:
	- cambio documental (sin cambios de codigo ni ejecucion de pruebas adicionales).

## 2026-04-04
- Punto 10 + Punto 11 (continuacion de backlog 1 y 2):
	- `backend/db/eventos_contables.go` amplÃƒÂ­a `empresa_eventos_contables` con metadatos de procesamiento (`intentos_procesamiento`, `fecha_ultimo_intento`, `error_procesamiento`, `asiento_contable_id`) y crea tabla canonica `empresa_asientos_contables` con hash de idempotencia.
	- `backend/handlers/finanzas.go` agrega `EmpresaFinanzasAsientosContablesHandler`:
		- `GET /api/empresa/finanzas/asientos_contables` para consulta,
		- `POST/PUT action=procesar_asientos|procesar` para procesamiento manual por lote.
	- `backend/handlers/empresa_permisos.go` clasifica `action=procesar_asientos` como accion de aprobacion en finanzas.
	- `backend/main.go` publica `/api/empresa/finanzas/asientos_contables` y registra migracion `2026-04-04-010-asientos-canonicos`.
	- `backend/db/finanzas.go` integra en el tablero los bloques `estado_resultados` y `balance_general`, junto con KPI contables de asientos (`asientos_generados`, `asientos_monto_total`).
	- `web/administrar_empresa/reportes.html` incorpora visualizacion de utilidad operacional, activos/pasivos/patrimonio, resultado del ejercicio y cuadre.
	- `web/administrar_empresa/finanzas.html` aÃƒÂ±ade accion manual `Procesar eventos contables`.
	- Cobertura de pruebas nueva/extendida en `backend/db/eventos_contables_test.go`, `backend/db/finanzas_test.go`, `backend/handlers/eventos_contables_modulos_test.go` y `backend/handlers/empresa_permisos_test.go`.
- Validacion tecnica:
	- `go test ./db -run "EventosContables|ReportesTableroResumen" -count=1` (ok).
	- `go test ./handlers -run "AsientosContables|TableroResumen|WithEmpresaFinanzasPermissions" -count=1` (ok).
	- `go test ./handlers -count=1` (ok).
	- `go test ./db -count=1` (ok).

## 2026-04-04
- Punto 12 + Punto 10 (continuacion de backlog 1 y 2):
	- `backend/handlers/empresa_permisos_test.go` agrega pruebas UAT por rol para `PUT action=aprobar` en `cierres_caja`:
		- rechazo para `cajero`,
		- rechazo para `supervisor_sucursal`,
		- aprobacion permitida para `admin_empresa`.
	- `documentos/matriz_roles_permisos_pos_multiempresa.md` agrega matriz UAT de cierres con casos por rol y transiciones de estado.
	- `documentos/plan_maestro_pos_multiempresa_14_puntos.md` define estrategia de procesamiento de asientos sobre `empresa_eventos_contables` y referencias canonicas documentales (`entidad_id`).
- Validacion tecnica:
	- `go test ./handlers -run "TestWithEmpresaFinanzasPermissions(DeniesCajeroAprobarCierreCaja|DeniesSupervisorAprobarCierreCaja|AllowsAdminAprobarCierreCaja)" -count=1` (ok).

## 2026-04-04
- Punto 12 (cierres de caja) Ã¢â‚¬â€ continuacion con UI operativa en panel empresa:
	- `web/administrar_empresa/finanzas.html` integra modulo visual de cierres de caja por sucursal con:
		- formulario de apertura/actualizacion,
		- calculo de `caja_teorica` y `diferencia_caja`,
		- filtros por sucursal/caja/estado/fecha,
		- tabla de acciones (`cerrar`, `reabrir`, `aprobar`, `anular`, `activar/desactivar`, `eliminar`).
	- La vista queda conectada al endpoint existente `GET/POST/PUT/DELETE /api/empresa/finanzas/cierres_caja`.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/finanzas.html` (ok).

## 2026-04-04
- Punto 12 (cierres de caja) Ã¢â‚¬â€ inicio de flujo operativo por sucursal:
	- `backend/db/finanzas.go` agrega `empresa_cierres_caja` con soporte de apertura, arqueo, cierre, reapertura, aprobacion y anulacion.
	- `backend/handlers/finanzas.go` incorpora `GET/POST/PUT/DELETE /api/empresa/finanzas/cierres_caja`.
	- `backend/main.go` publica la ruta de cierres de caja y registra migracion `2026-04-04-009-cierres-caja`.
	- `backend/handlers/empresa_permisos.go` trata `action=aprobar` en finanzas como accion `A`.
	- Pruebas nuevas:
		- `backend/db/finanzas_test.go`: `TestEmpresaCierresCajaFlow`.
		- `backend/handlers/eventos_contables_modulos_test.go`: `TestEmpresaFinanzasCierresCajaHandler`.
- Validacion tecnica:
	- `go test ./db -run "TestEmpresaCierresCajaFlow|TestGetEmpresaReportesTableroResumen|TestEmpresaFinanzas" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaFinanzasCierresCajaHandler|TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./auth ./db ./handlers ./metrics ./utils -count=1` (ok).

## 2026-04-04
- Punto 11 (reportes financieros) Ã¢â‚¬â€ inicio de tablero minimo financiero-operativo:
	- `backend/db/finanzas.go` agrega `GetEmpresaReportesTableroResumen` con KPI consolidados:
		- operativos (ventas/ticket/clientes/productos/compras),
		- financieros (ingresos/egresos/balance/periodos),
		- contables (eventos y documentos activos).
	- `backend/handlers/finanzas.go` extiende `GET /api/empresa/finanzas/movimientos` con `action=tablero|dashboard|resumen_kpi`.
	- `web/administrar_empresa/reportes.html` incorpora KPI financieros y contables en la misma vista de reportes.
	- Pruebas nuevas:
		- `backend/db/finanzas_test.go`: `TestGetEmpresaReportesTableroResumen`.
		- `backend/handlers/eventos_contables_modulos_test.go`: `TestEmpresaFinanzasTableroResumenHandler`.
- Validacion tecnica:
	- `go test ./db -run "TestGetEmpresaReportesTableroResumen|TestEmpresaFinanzas" -count=1` (ok).
	- `go test ./handlers -run "TestEmpresaFinanzasTableroResumenHandler|TestEmpresaFinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./... -count=1` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) Ã¢â‚¬â€ persistencia canonica de documentos transaccionales para `entidad_id`:
	- Se agrega `backend/db/documentos_transaccionales.go` con tablas y APIs de upsert/lectura para:
		- `empresa_facturacion_documentos`.
		- `empresa_compras_documentos`.
	- `backend/main.go` integra:
		- `EnsureEmpresaDocumentosTransaccionalesSchema`.
		- migracion `2026-04-04-008-documentos-transaccionales`.
	- `backend/handlers/facturacion_electronica.go` y `backend/handlers/productos.go` ahora:
		- consultan estado documental persistido por `documento_codigo`,
		- aplican transicion de ciclo sobre estado canonico,
		- persisten el nuevo estado en tabla de negocio,
		- emiten evento contable usando `entidad_id` canonico (ID persistido en tabla documental).
	- Se agrega `backend/db/documentos_transaccionales_test.go` y se amplian aserciones en `backend/handlers/eventos_contables_modulos_test.go` para verificar estabilidad de `entidad_id` en el ciclo documental.
- Validacion tecnica:
	- `go test ./handlers -run "FacturacionTransaccionalEmiteEventosContables|ComprasTransaccionalEmiteEventosContables|FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./auth ./db ./handlers ./metrics ./utils` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) Ã¢â‚¬â€ estandarizacion de estados en ciclo documental transaccional:
	- Se agrega `backend/handlers/documentos_lifecycle.go` con reglas de transicion por accion y estado previo para facturacion/compras.
	- `backend/handlers/facturacion_electronica.go` ahora valida `estado_actual` en `emitir/anular/nota_credito`, devuelve `409` en conflictos y responde `estado_anterior`/`estado_nuevo` cuando la transicion es valida.
	- `backend/handlers/productos.go` (`EmpresaProveedoresHandler`) aplica validacion equivalente para `emitir_orden/recepcionar_compra/contabilizar_compra`.
	- `backend/handlers/eventos_contables_modulos_test.go` amplÃƒÂ­a cobertura con pruebas de transiciones invalidas para facturacion y compras.
- Validacion tecnica:
	- `runTests` sobre `backend/handlers/eventos_contables_modulos_test.go` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) Ã¢â‚¬â€ eventos transaccionales de factura y orden:
	- `backend/handlers/facturacion_electronica.go` agrega acciones transaccionales:
		- `action=emitir` -> `factura_emitida`.
		- `action=anular` -> `factura_anulada`.
		- `action=nota_credito|emitir_nota_credito` -> `nota_credito_emitida`.
	- `backend/handlers/productos.go` (`EmpresaProveedoresHandler`) agrega acciones transaccionales:
		- `action=emitir|emitir_orden` -> `orden_compra_emitida`.
		- `action=recepcionar|recepcionar_compra` -> `compra_recepcionada`.
		- `action=contabilizar|contabilizar_compra` -> `compra_contabilizada`.
	- `backend/handlers/empresa_permisos.go` amplÃƒÂ­a mapeo de acciones de permisos para compras/facturacion.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega pruebas de emisiones transaccionales de factura/orden.
- Validacion tecnica:
	- `go test ./handlers -run "FacturacionTransaccionalEmiteEventosContables|ComprasTransaccionalEmiteEventosContables|FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras/finanzas) Ã¢â‚¬â€ extension de emision de eventos contables por modulo:
	- Se agrega `backend/handlers/eventos_contables.go` para registro no bloqueante y reutilizable de eventos contables en handlers.
	- Se amplia `backend/db/eventos_contables.go` con eventos operativos de:
		- `facturacion`: `configuracion_facturacion_actualizada`.
		- `compras`: `proveedor_registrado`, `proveedor_actualizado`, `proveedor_activado`, `proveedor_desactivado`, `proveedor_eliminado`.
	- Se integra emision en:
		- `backend/handlers/facturacion_electronica.go`.
		- `backend/handlers/productos.go` (proveedores).
		- `backend/handlers/finanzas.go` (movimientos y periodos).
	- `backend/handlers/carritos_compras.go` migra a helper comun para consistencia del registro contable.
	- Se agregan pruebas en `backend/handlers/eventos_contables_modulos_test.go` para validar emision en facturaciÃƒÂ³n, compras y finanzas.
- Validacion tecnica:
	- `go test ./db -run "EventosContables" -count=1` (ok).
	- `go test ./handlers -run "FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables|CarritosCompraAndItemsFlow" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 4 + Punto 10 (gestion de ventas + modulo contable integrado) Ã¢â‚¬â€ contrato de eventos contables por modulo:
	- Se agrega `backend/db/eventos_contables.go` con contrato base de eventos para `ventas`, `facturacion`, `compras` y `finanzas`.
	- Se crea tabla `empresa_eventos_contables` en `pcs_empresas` para registrar trazabilidad contable por empresa (`modulo`, `evento`, `entidad`, `documento`, `periodo_contable`, `monto`, `payload_json`, `procesado`).
	- Se integra bootstrap en `backend/main.go`:
		- `EnsureEmpresaEventosContablesSchema`.
		- migracion `2026-04-04-007-eventos-contables`.
	- Se actualiza `backend/handlers/carritos_compras.go` para emitir eventos contables en transiciones de venta de carritos (`venta_sesion_activada`, `venta_activada`, `venta_suspendida`, `venta_cerrada`, `venta_reabierta`, `venta_pagada`).
	- Se agregan pruebas:
		- `backend/db/eventos_contables_test.go`.
		- `backend/handlers/auth_users_carritos_test.go` (validacion de emision de `venta_pagada`).
- Validacion tecnica:
	- `go test ./db -run "EventosContables|CarritoEstadoVentaLifecycle|Finanzas" -count=1` (ok).
	- `go test ./handlers -run "EmpresaCarritosCompra|CarritosCompraAndItemsFlow" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 4 (gestion de ventas) Ã¢â‚¬â€ formalizacion de transiciones del ciclo de venta en carritos:
	- `backend/handlers/carritos_compras.go` ahora valida transiciones por accion y estado actual del carrito.
	- Se agregan respuestas de control para integridad de flujo:
		- `404` para carrito inexistente,
		- `409` para transiciones no permitidas (doble pago, reabrir pagada, activar estacion pagada sin `reset_items=1`, etc.).
	- Se agregan pruebas en `backend/handlers/auth_users_carritos_test.go`:
		- `TestEmpresaCarritosCompraRejectsDoublePago`.
		- `TestEmpresaCarritosCompraRejectsReabrirVentaPagada`.
		- `TestEmpresaCarritosCompraRejectsActivarEstacionPagadaSinReset`.
- Validacion tecnica:
	- `go test ./handlers -run "Carritos|EmpresaCarritosCompra" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Cierre validado del punto 3 (permisos y seguridad) con pruebas de endpoints protegidos recien incorporados:
	- `backend/handlers/empresa_permisos_test.go` agrega:
		- `TestWithEmpresaInventarioPermissionsDeniesCajeroWriteGPS`.
		- `TestWithEmpresaVentasPermissionsAllowsCajeroChatAdjuntoMultipart`.
		- `TestWithEmpresaVentasPermissionsRejectsChatAdjuntoWithoutAuth`.
	- Se valida control por rol en GPS, extraccion de `empresa_id` en `multipart/form-data` para adjuntos de chat y rechazo `401` sin autenticacion.
- Inicio del punto 4 (gestion de ventas):
	- `backend/db/carritos_compras.go` incorpora `estado_venta` derivado en el modelo `CarritoCompra` para estandarizar ciclo de vida de venta:
		- `venta_abierta`,
		- `venta_cerrada`,
		- `venta_pagada`,
		- `venta_suspendida`.
	- `backend/handlers/carritos_compras.go` expone `estado_venta` en acciones operativas (`activar_estacion`, `pagar_estacion`, `activar/desactivar`, `cerrar/reabrir`).
	- Se amplian pruebas en:
		- `backend/handlers/auth_users_carritos_test.go`.
		- `backend/db/carritos_inventario_test.go`.
- Validacion tecnica de esta iteracion:
	- `runTests` sobre archivos de pruebas modificados (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Continuacion del punto 3 del plan maestro (permisos y seguridad) con cierre de rutas operativas pendientes:
	- `backend/handlers/empresa_permisos.go` agrega modulo `seguridad` y wrapper `WithEmpresaSeguridadPermissions`.
	- `backend/main.go` amplÃƒÂ­a middleware en rutas:
		- seguridad: `/api/empresa/usuarios`, `/api/empresa/configuracion_avanzada`, `/api/empresa/roles_de_usuario`.
		- inventario: `/api/empresa/productos/imagen`, `/api/empresa/ubicacion_gps/dispositivos`, `/api/empresa/ubicacion_gps/recorridos`.
		- colaboracion operativa (politica ventas): `/api/empresa/chat_tareas/conversaciones`, `/api/empresa/chat_tareas/participantes`, `/api/empresa/chat_tareas/mensajes`, `/api/empresa/chat_tareas/mensajes/adjunto`, `/api/empresa/chat_tareas/tareas`.
	- `backend/handlers/empresa_permisos_test.go` agrega cobertura para modulo seguridad:
		- `TestWithEmpresaSeguridadPermissionsDeniesSupervisorWrite`.
		- `TestWithEmpresaSeguridadPermissionsAllowsSupervisorRead`.
		- `TestWithEmpresaSeguridadPermissionsAllowsAdminApprove`.
	- Validacion tecnica: `go test ./handlers -run "WithEmpresa|ConsultarHandlerRejectsEmpresaFueraDeAlcance" -count=1` (ok) y `go test ./...` (ok).

## 2026-04-04
- Continuacion del punto 3 del plan maestro (permisos y seguridad):
	- `backend/handlers/empresa_permisos.go` amplÃƒÂ­a modulos de autorizacion para `clientes`, `compras` y `facturacion`.
	- Se agregan wrappers: `WithEmpresaClientesPermissions`, `WithEmpresaComprasPermissions`, `WithEmpresaFacturacionPermissions`.
	- `backend/main.go` aplica middleware en rutas: `/api/empresa/clientes`, `/api/empresa/proveedores`, `/api/empresa/facturacion_electronica`, `/api/empresa/facturacion_electronica/pais_detectado`, y `/api/empresa/servicios` (politica inventario).
	- Se amplian pruebas en `backend/handlers/empresa_permisos_test.go` para cobertura de los modulos nuevos.
	- Validacion tecnica: `go test ./handlers -run "WithEmpresa|ConsultarHandlerRejectsEmpresaFueraDeAlcance" -count=1` (ok) y `go test ./...` (ok).

## 2026-04-04
- Se registra nueva credencial Gemini cifrada en configuraciÃƒÂ³n avanzada (`ai.model.google.gemini_2_0_flash.api_key` en `pcs_superadministrador`).
- Se valida consumo de Gemini con la nueva credencial: respuesta del proveedor `429` por cuota excedida (sin error de credencial/servicio bloqueado).
- Se verifica la presencia de la tarjeta de Gemini en `web/super/configuracion_avanzada.html` y se corrige un bloque JavaScript en la carga de estado para mantener consistencia de la vista.
- Se agrega prueba de seguridad de alcance por empresa para chat IA en `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`:
	- `TestConsultarHandlerRejectsEmpresaFueraDeAlcance`.
	- ValidaciÃƒÂ³n: `go test ./handlers -run "TestConsultarHandlerRejectsEmpresaFueraDeAlcance|TestModelosHandlerRequiresGoogleAccount|TestModelosHandlerReturnsPreferredModelForGoogleAccount" -count=1` (ok).

## 2026-04-04
- Chat IA empresarial migrado a Gemini-only:
	- `backend/handlers/chat_con_inteligencia_artificial_controller.go` ahora integra Google Gemini (`generateContent`) y elimina dependencias de OpenAI/DeepSeek/Hugging Face para este mÃƒÂ³dulo.
	- El catÃƒÂ¡logo y la configuraciÃƒÂ³n de credenciales IA quedan en un ÃƒÂºnico modelo soportado: `google:gemini-2.0-flash` (`GEMINI_API_KEY`).
	- `web/super/configuracion_avanzada.html` simplifica la tarjeta IA a una sola credencial Gemini con trazabilidad por cuenta Google.
	- `web/administrar_empresa/chat_con_inteligencia_artificial.html` se rediseÃƒÂ±a con experiencia visual tipo Gemini, chips de contexto y flujo explÃƒÂ­cito de autenticaciÃƒÂ³n Google.
	- Pruebas ajustadas y validadas: `go test ./auth ./db ./handlers ./metrics ./utils` (ok) en `backend`.
- Se agrega gestiÃƒÂ³n de credenciales IA en `super/configuracion_avanzada.html` para 5 modelos populares con plan gratuito limitado:
	- OpenAI GPT-4o mini,
	- OpenAI GPT-4.1 mini,
	- DeepSeek Chat,
	- DeepSeek Reasoner,
	- Meta Llama 3.1 8B Instruct (Hugging Face).
- Se crea endpoint `GET/PUT /super/api/config/ai` en backend para guardar/consultar credenciales con registro de la cuenta Google logueada que realiza cambios.
- El mÃƒÂ³dulo `chat_con_inteligencia_artificial` ahora resuelve credenciales en este orden:
	- configuraciÃƒÂ³n guardada por modelo,
	- configuraciÃƒÂ³n por proveedor,
	- variable de entorno.
- ValidaciÃƒÂ³n tÃƒÂ©cnica ejecutada:
	- `go test ./handlers -run "AIModelsConfigHandler|Chat|ModelosHandler" -count=1` (ok).
	- `go test ./...` en `backend` (ok).
- Se implementa la primera fase tecnica del punto 3 (permisos y seguridad) con middleware de autorizacion por rol + alcance de empresa:
	- nuevo `backend/handlers/empresa_permisos.go`,
	- aplicacion en rutas criticas de ventas, inventario y finanzas desde `backend/main.go`,
	- pruebas nuevas en `backend/handlers/empresa_permisos_test.go` para denegacion/aprobacion por rol y empresa.
- Validacion tecnica de la fase:
	- `go test ./handlers -run WithEmpresa -count=1` (ok).
	- `go test ./...` en `backend` (ok).
- Se actualiza la documentacion del proyecto para continuar el plan maestro de 14 puntos:
	- nuevo `documentos/plan_maestro_pos_multiempresa_14_puntos.md` con estado, entregables y backlog de ejecucion,
	- nueva `documentos/matriz_kpi_pos_multiempresa.md` con formulas/frecuencia/fuentes de KPI,
	- nueva `documentos/matriz_roles_permisos_pos_multiempresa.md` para iniciar el punto 3 de permisos y seguridad,
	- actualizacion de `documentos/descripcion_del_proyecto` para referenciar estos documentos como base de seguimiento.
- ContinuaciÃƒÂ³n de implementaciÃƒÂ³n en `chat_con_inteligencia_artificial`:
	- Se corrige el orden de validaciÃƒÂ³n de autenticaciÃƒÂ³n para cuenta Google en `backend/handlers/chat_con_inteligencia_artificial_controller.go`.
	- Cuando no hay cuenta Google autenticada, los endpoints del mÃƒÂ³dulo IA ahora responden `401` de forma consistente (en lugar de caer en validaciÃƒÂ³n de alcance con `403`).
	- Se centraliza validaciÃƒÂ³n de alcance con `ensureEmpresaAccessByAccount` para reutilizar la cuenta ya validada.
- Se agregan pruebas automÃƒÂ¡ticas del mÃƒÂ³dulo IA:
	- `backend/db/chat_inteligencia_artificial_test.go` (upsert/get de modelo preferido y acumulaciÃƒÂ³n de uso diario).
	- `backend/handlers/chat_con_inteligencia_artificial_controller_test.go` (autorizaciÃƒÂ³n por cuenta Google y respuesta con modelo preferido).
- ValidaciÃƒÂ³n tÃƒÂ©cnica ejecutada en esta continuaciÃƒÂ³n:
	- `go test ./db -run EmpresaAI -count=1` (ok).
	- `go test ./handlers -run ModelosHandler -count=1` (ok).
	- `go test ./...` en `backend` (ok).
- Se amplÃƒÂ­a el mÃƒÂ³dulo `chat_con_inteligencia_artificial` para registrar el modelo preferido por cuenta Google autenticada (por empresa):
	- Nueva tabla `empresa_ai_modelo_preferido` en `pcs_empresas` (UNIQUE por `empresa_id + admin_email`).
	- Nuevas funciones en `backend/db/chat_inteligencia_artificial.go`: `GetEmpresaAIModeloPreferido` y `UpsertEmpresaAIModeloPreferido`.
	- Nuevo endpoint `GET/PUT /api/empresa/chat_con_inteligencia_artificial/modelo_preferido`.
	- `GET /modelos` ahora devuelve `google_account` y `modelo_preferido`.
	- `POST /consultar` ahora persiste el `model_id` usado como preferencia de la cuenta Google y devuelve confirmaciÃƒÂ³n en respuesta.
- Se actualiza `web/administrar_empresa/chat_con_inteligencia_artificial.html` para:
	- cargar automÃƒÂ¡ticamente el modelo preferido de la cuenta Google,
	- guardar el modelo preferido al cambiar selecciÃƒÂ³n,
	- mostrar la cuenta Google vinculada en el bloque de uso diario.
- ValidaciÃƒÂ³n tÃƒÂ©cnica ejecutada para esta ampliaciÃƒÂ³n:
	- `gofmt -w backend/db/chat_inteligencia_artificial.go backend/handlers/chat_con_inteligencia_artificial_controller.go backend/handlers/chat_con_inteligencia_artificial_router.go` (ok).
	- `go test ./...` en `backend` (ok).
- Se fortalece `backend/utils/utils.go` para observabilidad profesional:
	- `LoggingMiddleware` ahora genera `request_id` por solicitud, calcula `empresa_id` (query/header/JSON body) y registra inicio/fin con latencia.
	- Se agregan logs separados por empresa en `backend/logs/empresa_<id>.log` y un fallback global en `backend/logs/empresa_global.log`.
	- `JSONErrorMiddleware` ahora normaliza errores no-JSON incluyendo `request_id` y `empresa_id` cuando aplica, y registra errores API por empresa.
- Se ajustan endpoints multipart para reforzar separaciÃƒÂ³n de logs por empresa:
	- `backend/handlers/chat_tareas.go` y `backend/handlers/productos.go` ahora establecen `X-Empresa-ID` tras parsear `empresa_id` del formulario.
- Se endurece `backend/handlers/usuarios_empresa.go` en autenticaciÃƒÂ³n/primer ingreso:
	- se reemplazan respuestas `500` que exponÃƒÂ­an detalles internos por mensajes profesionales y seguros,
	- se agrega logging servidor con contexto (`empresa_id`, `email`, `id`) para trazabilidad sin filtrar errores sensibles al cliente.
- Se endurece `scripts/iniciar_servidor.ps1` para detectar caÃƒÂ­da temprana de `server.exe`: ahora conserva el `PID`, valida salida prematura y muestra las ÃƒÂºltimas lÃƒÂ­neas de `backend/server.err` para diagnÃƒÂ³stico inmediato.
- ValidaciÃƒÂ³n de correcciÃƒÂ³n ejecutada:
	- `gofmt -w backend/utils/utils.go` (ok).
	- `go test ./...` en `backend` (ok).
- Se corrige `scripts/iniciar_servidor.ps1` en `Resolve-GoogleOAuthCredentials`: la construccion de `envCandidates` ahora usa `Join-Path -Path/-ChildPath` por elemento, evitando el error `CannotConvertArgument` de `Join-Path`.
- Se corrige `backend/db/finanzas.go` en `EnsureEmpresaFinanzasSchema`: los indices que dependen de columnas migradas (`periodo_contable` y `estado` de periodos) se crean al final de la migracion para compatibilidad con bases antiguas.
- Validacion de correccion ejecutada:
	- `go test ./...` en `backend` (ok).
	- `go run .` en `backend` (arranque correcto en `:8080`).
- Se incorpora el modulo `chat_con_inteligencia_artificial` en el panel empresarial con interfaz tipo chat en `web/administrar_empresa/chat_con_inteligencia_artificial.html`.
- Se crean `backend/db/chat_inteligencia_artificial.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go` y `backend/handlers/chat_con_inteligencia_artificial_router.go` para arquitectura modular (DB + controller + router).
- Se publican rutas del modulo IA:
	- `GET /api/empresa/chat_con_inteligencia_artificial/modelos`
	- `POST /api/empresa/chat_con_inteligencia_artificial/consultar`
	- `GET /api/empresa/chat_con_inteligencia_artificial/historial`
- Se agregan tablas en `pcs_empresas` para auditoria y limites diarios:
	- `empresa_ai_consultas`
	- `empresa_ai_uso_diario`
- Se integra `EnsureEmpresaAIChatSchema` y la migracion `2026-04-03-005-chat-ia-empresa` en `backend/main.go`.
- Se implementa aislamiento estricto por `empresa_id`, validacion de alcance de usuario y control de limite free-tier por empresa/proveedor/modelo/dia con opcion de upgrade.
- Se habilitan modelos famosos de OpenAI, DeepSeek y Hugging Face usando credenciales solo en backend mediante variables de entorno (`OPENAI_API_KEY`, `DEEPSEEK_API_KEY`, `HUGGINGFACE_API_KEY`).
- Se amplÃƒÂ­a el mÃƒÂ³dulo financiero con control de periodos contables por empresa:
	- tabla `empresa_finanzas_periodos`.
	- endpoint `GET/POST/PUT /api/empresa/finanzas/periodos`.
	- acciones de cierre y reapertura de periodo.
- Se aplican bloqueos de integridad contable: no se permite crear/editar/eliminar/activar/desactivar movimientos cuando su periodo estÃƒÂ¡ cerrado.
- Se amplÃƒÂ­a `empresa_finanzas_movimientos` con:
	- `periodo_contable`,
	- retenciones (`retencion_fuente`, `retencion_ica`, `retencion_iva`, `total_retenciones`),
	- `total_neto`.
- Se amplÃƒÂ­a `empresa_finanzas_configuracion` con `cuenta_retenciones_cobrar` y `cuenta_retenciones_pagar`.
- Se completa la UI de finanzas para:
	- gestionar periodos (cerrar/reabrir/actualizar),
	- calcular total bruto, retenciones y neto,
	- filtrar por periodo,
	- exportar `balance general`, `libro diario` y `libro mayor` en CSV.
- Se corrige el escaneo de puertos de seguridad para compatibilidad IPv6 usando `net.JoinHostPort` en `backend/handlers/system_empresas_handlers.go`.
- Se ajusta `scripts/iniciar_servidor.ps1` para usar nombre de funciÃƒÂ³n con verbo aprobado de PowerShell en la carga de `.env`.
- ValidaciÃƒÂ³n tÃƒÂ©cnica ejecutada: `go test ./...` en `backend` (ok).
- Se implementa el mÃƒÂ³dulo financiero multiempresa con enfoque unificado de ingresos y egresos en `web/administrar_empresa/finanzas.html`.
- Se crea `backend/db/finanzas.go` con esquema, validaciones y CRUD de:
	- `empresa_finanzas_movimientos`
	- `empresa_finanzas_configuracion`
- Se crea `backend/handlers/finanzas.go` y se publican rutas:
	- `GET/POST/PUT/DELETE /api/empresa/finanzas/movimientos`
	- `GET/POST/PUT /api/empresa/finanzas/configuracion`
- Se actualiza `backend/main.go` para asegurar el esquema financiero y registrar la migraciÃƒÂ³n `2026-04-03-003-finanzas`.
- Se integra el acceso al mÃƒÂ³dulo en `web/administrar_empresa.html` y `web/js/administrar_empresa.js`.
- Se agrega `backend/db/finanzas_test.go` con pruebas de configuraciÃƒÂ³n y flujo CRUD de movimientos financieros.
- Se amplÃƒÂ­a `backend/tools/seed_motel_malibu/main.go` para sembrar configuraciÃƒÂ³n financiera y movimientos demo de ingreso/egreso.
- Se separa visualmente el libro financiero en dos pestaÃƒÂ±as operativas dentro del mÃƒÂ³dulo: `Ingresos` y `Egresos`.
- Se agrega la pestaÃƒÂ±a `Todos` para consolidar ingresos y egresos en una sola vista del libro financiero.
- Se agrega exportaciÃƒÂ³n del libro financiero filtrado por fechas a:
	- Excel (CSV compatible con Excel).
	- PDF (vista de impresiÃƒÂ³n).
	- JSON contable para integraciÃƒÂ³n externa (incluye resumen, detalle y asientos recomendados).
- Se amplÃƒÂ­a la configuraciÃƒÂ³n financiera por empresa para contabilidad externa con parametrizaciÃƒÂ³n de:
	- destino de integraciÃƒÂ³n (`generico`, `siigo`, `world_office`, `alegra`),
	- cuentas base (caja/bancos, ingresos, IVA generado, gastos, IVA descontable),
	- cuentas por categorÃƒÂ­a para ingresos y egresos.
- La exportaciÃƒÂ³n `JSON contable` deja de usar cuentas fijas y ahora construye asientos con la parametrizaciÃƒÂ³n real guardada por empresa.
- El JSON exportado incorpora `accounting_profile` y `erp_projection` por movimiento para facilitar mapeo hacia software contable externo.
- Se actualiza `backend/db/finanzas_test.go` para validar persistencia de la nueva parametrizaciÃƒÂ³n contable.
- Se amplÃƒÂ­a `web/administrar_empresa/finanzas.html` con salidas contables adicionales:
	- Plantilla dedicada SIIGO (CSV) para importaciÃƒÂ³n de asientos.
	- Balance de prueba (CSV).
	- Estado de resultados (CSV).
- Se crea `documentos/plantillas/siigo_plantilla_importacion_asientos.csv` como plantilla de referencia ERP.
- Se crea `documentos/informe_contable_directivo_2026-04-03.md` con revisiÃƒÂ³n de cumplimiento contable/directivo, brechas y plan recomendado.
- ValidaciÃƒÂ³n tÃƒÂ©cnica ejecutada:
	- `go test ./... -count=1` (ok).
	- `go run ./tools/seed_motel_malibu` (ok, incluye creaciÃƒÂ³n de 4 movimientos financieros demo).
	- `runTests` global (ok: 3/3).

## 2026-04-03
- Se implementa control de inventario en carrito: al agregar items de producto se descuenta stock y al desactivar/eliminar items abiertos se revierte automÃƒÂ¡ticamente.
- Se asegura que, al cerrar una venta, el descuento de inventario permanezca aplicado y no se revierta en el pago.
- Se mejoran respuestas de API para stock insuficiente en operaciones de items de carrito.
- Se agrega `backend/db/carritos_inventario_test.go` con pruebas de descuento de inventario y caso de stock insuficiente.
- Se amplÃƒÂ­a `backend/tools/seed_motel_malibu/main.go` para registrar 10 clientes y 10 usuarios de empresa.
- La semilla valida automÃƒÂ¡ticamente el flujo comercial completo: venta cerrada, descuento de inventario al agregar y persistencia tras pagar.
- Se confirma en seed la validaciÃƒÂ³n de impresiÃƒÂ³n con vista previa POS y Carta.
- Se amplÃƒÂ­a `web/administrar_empresa/reportes.html` con reporte de ventas, reporte de productos y reporte de compra de productos, todos con bÃƒÂºsqueda por rango de fechas.
- ValidaciÃƒÂ³n tÃƒÂ©cnica ejecutada: `go test ./auth ./db ./handlers ./metrics ./utils` (ok) y `go run ./tools/seed_motel_malibu` (ok).
- Se agrega el vÃƒÂ­nculo `Ayuda` en el menÃƒÂº flotante global (`web/menu.js`) y se reestructura `web/ayuda/ayuda.html` como centro de ayuda con menÃƒÂº interno y secciÃƒÂ³n de APIs.
- Se adapta `web/administrar_empresa/carrito_de_compras.html` para operaciÃƒÂ³n con lector de cÃƒÂ³digo de barras (escaneo por cÃƒÂ³digo/SKU, Enter para agregar y acumulaciÃƒÂ³n opcional de cantidad).
- Se extiende `web/administrar_empresa/configuracion.html` con configuraciÃƒÂ³n por empresa para el lector: habilitar, autofoco y acumulaciÃƒÂ³n.
- Se amplÃƒÂ­a `web/administrar_empresa/reportes.html` con KPI de productos bajo mÃƒÂ­nimo y reporte de inventario actual por bodega.
- ValidaciÃƒÂ³n tÃƒÂ©cnica ejecutada para flujo carrito/inventario multiempresa: `go test ./db -run Carrito -count=1` (ok) y `go test ./handlers -run Carritos -count=1` (ok).

## 2026-04-02
- Se crea la herramienta `backend/tools/seed_motel_malibu/main.go` para cargar datos demo comerciales en la empresa Motel Malibu.
- La semilla inserta 10 productos con precios COP, 5 clientes y crea una venta de prueba cerrada para validar el flujo comercial.
- Se valida la configuracion de impresion con vista previa de formatos POS y Carta desde la herramienta de seed.
- Se implementa la seccion `web/administrar_empresa/reportes.html` con KPIs, ventas cerradas, top productos, top clientes y resumen de impresion.
- Se reestructura `backend/tools` en subcarpetas por herramienta para eliminar conflictos de compilaciÃƒÂ³n por mÃƒÂºltiples `main`.
- Se valida backend completo con `go test ./...` (ok).
- Se valida el mÃƒÂ³dulo GPS con pruebas especÃƒÂ­ficas:
	- `go test ./db -run TestEmpresaGPSDispositivosYRecorridosCRUD -count=1` (ok).
	- `go test ./handlers -run TestEmpresaUbicacionGPSHandlersCRUDFlow -count=1` (ok).
- Se implementa el modulo de ubicacion GPS por empresa con soporte de multiples dispositivos.
- Se agregan tablas `empresa_gps_dispositivos` y `empresa_gps_recorridos` en `pcs_empresas`.
- Se crean endpoints CRUD para dispositivos y recorridos GPS en `/api/empresa/ubicacion_gps/*`.
- Se agrega la pagina `web/administrar_empresa/ubicacion_gps.html` con mapa OpenStreetMap (Leaflet).
- Se habilita tracking automatico de recorridos cada 10 segundos por dispositivo.
- Se agregan pruebas en `backend/db/ubicacion_gps_test.go` y `backend/handlers/ubicacion_gps_test.go`.



### Added
- ImplementaciÃ¯Â¿Â½n de controladores HTTP (Handlers) para el CRUD de Proveedores integrado en el nuevo mÃ¯Â¿Â½dulo de compras y logÃ¯Â¿Â½stica ERP.



## [2026-04-20] Limpieza Total Themes
- [UI/Temas] Auditoría y barrido de más de 50 páginas y scripts en web/administrar_empresa, web/super y páginas públicas para limpiar colores fijos, migrando lógicas JS a .classList.add('text-danger') y respetando las 6 paletas dinámicas. Completado barrido masivo de vistas.
- **Auditoria + IA contextual**: los wrappers empresariales auditan ahora acciones `R/C/U/D/A` de forma no bloqueante y la IA empresarial/global recibe una ventana reciente de `empresa_auditoria_eventos` como contexto operativo. La integracion vive en auditoria + constructor de prompt, sin insertar IA en cada modulo; si auditoria o IA fallan/deshabilitan, el servidor continua con estado controlado. Archivos: `backend/handlers/auditoria_empresa.go`, `backend/db/auditoria_empresa.go`, `backend/db/chat_inteligencia_artificial.go`, `backend/handlers/chat_con_inteligencia_artificial_controller.go`, `backend/handlers/chat_con_ia_global_super.go`, documentacion relacionada.

## [2026-05-03] Documentacion, ayuda y estado operativo de modulos
- [Docs] Se crea `documentos/reporte_estado_modulos_2026-05-03.md` con estado compacto por modulo, observaciones de calidad y dependencias pendientes de certificacion.
- [Ayuda] Se actualiza `web/ayuda/ayuda.html` con una seccion de estado operativo, estaciones/carrito, tarjetas adaptables, indicadores del panel y limites honestos de validacion.
- [Operacion] Se documentan los cambios recientes: carrito desde estacion, pago con retorno a estaciones, `USD / COP` primero y despliegue VPS correcto.
