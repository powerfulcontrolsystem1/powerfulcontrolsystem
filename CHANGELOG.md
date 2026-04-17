# CHANGELOG

## 2026-04-17
- Exportes operativos: descarga silenciosa sin sacar al usuario del módulo.
	- Archivos modificados: `web/administrar_empresa/administrar_clientes.html`, `web/administrar_empresa/asistencia_empleados.html`, `web/administrar_empresa/backups.html`, `web/administrar_empresa/tarifas_por_dia.html`, `web/administrar_empresa/soporte_remoto.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: los exportes frecuentes de clientes, asistencia, backups, tarifas por día y soporte remoto dejan de reemplazar la vista actual. El archivo se descarga en segundo plano y el usuario permanece en el mismo módulo.
	- Verificacion: diagnostico del editor sin errores en los archivos modificados.

- Navegacion general: misma pestaña por defecto.
	- Archivos modificados: `web/super_administrador.html`, `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/js/seleccionar_empresa.js`, `web/login.html`, `web/registrar_nuevo_usuario_administrador.html`, `web/registrar_contrasena_usuario_de_google.html`, `web/super/venta_digital.html`, `web/super/pagina_principal.html`, `web/super/configuracion_avanzada.html`, `web/administrar_empresa/venta_publica.html`, `web/administrar_empresa/soporte_remoto.html`, `web/super/soporte_remoto.html`, `web/administrar_empresa/administrar_clientes.html`, `web/administrar_empresa/asistencia_empleados.html`, `web/administrar_empresa/backups.html`, `web/administrar_empresa/tarifas_por_dia.html`, `web/administrar_empresa/chat_con_inteligencia_artificial.html`, `web/administrar_empresa/chat_y_tareas.html`, `web/index.html`, `web/Informacion_de_contacto.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: la navegación normal del sistema deja de abrir pestañas nuevas y reutiliza la misma ventana actual. Se mantienen como excepción solo el contrato, los términos legales de pasarela y los popups técnicos de impresión o vista previa documental.
	- Verificacion: búsqueda final de `target="_blank"|window.open(` limitada a excepciones esperadas; diagnóstico del editor sin errores en los archivos modificados.

- Licencias super: valor 0 ya no se oculta en edición ni en listado.
	- Archivos modificados: `web/super/licencias.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el CRUD de licencias en panel super conserva `0` como valor valido visible en la tabla y en el formulario de edición, evitando que una licencia parezca vacía al reabrirla.
	- Verificacion: diagnostico del editor sin errores en `web/super/licencias.html`.

- Licencias del selector: historial con vencimiento y renovacion.
	- Archivos modificados: `backend/db/db.go`, `backend/handlers/payments_handlers_test.go`, `web/super/licencias.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: la ruta `super/licencias.html?scope=mine&con_empresa=1` deja de mostrar el CRUD y pasa a ser un historial de licencias pagadas o vencidas por empresa, con fecha de vencimiento visible, estados operativos y acceso a `Pagar nueva licencia` cuando la licencia esta por vencer o ya vencio. El backend reutiliza el mismo endpoint `/super/api/licencias` exponiendo empresa y fechas para ese flujo.
	- Verificacion: diagnostico del editor sin errores en los archivos modificados; `go test ./handlers -run "TestLicenciasHandlerGetReturnsHistorialFieldsForCreatorScope" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Checkout publico de licencias: Epayco migra a Smart Checkout v2.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `web/super/configuracion_avanzada.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el sistema deja de generar URLs manuales hacia `checkout.php`, porque ese flujo ya responde `AccessDenied`. Ahora el backend crea la sesion oficial Smart Checkout v2 en Apify y el frontend abre `checkout-v2.js` con `sessionId`, manteniendo las mismas rutas publicas de respuesta, verificacion y webhook.
	- Verificacion: `go test ./handlers -run 'TestEpaycoCreateTransactionHandler(UsesConfiguredPublicBaseURLAndKeys|AllowsCheckoutWithoutPrivateKey|AcceptsSamboxAlias)|TestEpaycoTransactionStatusHandler(PreservesPendingOnGenericValidationError|FindsContextUsingInvoiceWhenGatewayIDsDiffer)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Crear clave por correo: ojo para mostrar u ocultar la contrasena.
	- Archivos modificados: `web/registrar_contrasena_usuario_de_google.html`, `web/js/registrar_contrasena_usuario_de_google.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: la pagina `Crear clave para acceso por correo` ahora incluye un icono de ojo en ambos campos de contrasena para poder revisarla visualmente antes de guardarla.
	- Verificacion: diagnostico del editor sin errores en los archivos modificados.

- Elegir licencia: tarjetas con el mismo estilo del home.
	- Archivos modificados: `web/elegir_licencia.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: la pagina `elegir_licencia.html` ahora renderiza las licencias con la misma estructura visual de tarjetas usada en `index.html`, manteniendo sin cambios el flujo de compra hacia `pagar_licencia.html`.
	- Verificacion: diagnostico del editor sin errores en `web/elegir_licencia.html`.

- Reportes globales super: eleccion explicita de una empresa o varias.
	- Archivos modificados: `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js`, `web/estilos.css`, `backend/handlers/reportes_globales_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el modulo `Reportes globales` ahora permite escoger de forma explicita si el analisis se hace sobre una sola empresa o sobre varias. En modo singular la UI cambia a selector puntual y el frontend consulta la API usando `empresa_id`.
	- Verificacion: diagnostico del editor sin errores en los archivos modificados; `go test ./handlers -run "TestSuperReportesGlobalesHandlerFiltraYConsolidaPorAdministrador" -count=1`.

- Login administrativo: Google y correo quedan en una sola tarjeta visual.
	- Archivos modificados: `web/login.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el bloque de acceso por correo deja de renderizarse como un formulario en caja separado dentro de `login.html`. Google, correo, recuperación y reset ahora comparten el mismo contenedor visual principal.
	- Verificacion: diagnostico del editor sin errores en `web/login.html` y `web/estilos.css`.

- Arcade publico: runtime comun de poderes y premios en los nueve juegos activos.
	- Archivos modificados: `web/Juegos/arcade_shared.js`, `web/Juegos/arcade_window.css`, `web/Juegos/patito_volando_plus.html`, `web/Juegos/serpiente_pixel_plus.html`, `web/Juegos/memoria_estelar_plus.html`, `web/Juegos/rebote_bloques_plus.html`, `web/Juegos/pacman_plus.html`, `web/Juegos/tetris_plus.html`, `web/Juegos/carton_fire_plus.html`, `web/Juegos/ajedrez_vs_ia_plus.html`, `web/Juegos/ajedrez_3d_plus.html`, `web/Juegos/menu_juegos.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el arcade publico queda unificado con una misma capa mobile-first de countdown, sonido, records, poderes y premios para los nueve juegos activos del lobby, con economia compartida ajustada para juegos de eventos rapidos y un lobby que muestra mejor el progreso personal y el ranking por titulo.
	- Verificacion: diagnostico del editor sin errores en `web/Juegos/arcade_shared.js`, `web/Juegos/arcade_window.css`, `web/Juegos/menu_juegos.html` y los cuatro juegos integrados en esta fase; busqueda de `createPowerSystem` presente en los 9 juegos activos.

## 2026-04-17
- Arcade publico: nuevo Ajedrez 3D plus con cinco dificultades.
	- Archivos creados: `web/Juegos/ajedrez_3d_plus.html`, `web/img/juegos/ajedrez_3d.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se agrega una nueva variante publica de ajedrez al arcade del portal, con tablero en perspectiva 3D simulada, cronometro arcade, cuenta regresiva de inicio y cinco niveles de dificultad contra la IA.
	- Verificacion: diagnostico del editor sin errores en `web/Juegos/ajedrez_3d_plus.html` y `web/Juegos/menu_juegos.html`.

## 2026-04-17
- Reportes globales super: graficos y lectura ejecutiva.
	- Archivos modificados: `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: la vista global del panel super ahora añade graficos comparativos y una lectura ejecutiva automática del consolidado de empresas seleccionadas, sin cambiar el modelo de permisos ni crear dependencias frontend externas.
	- Verificacion: diagnostico del editor sin errores en HTML/JS modificados.

## 2026-04-16
- Reportes globales super: consolidados por administrador creador.
	- Archivos modificados: `backend/db/db.go`, `backend/main.go`, `backend/handlers/reportes_globales.go`, `backend/handlers/reportes_globales_test.go`, `web/super/reportes_globales.html`, `web/js/super_reportes_globales.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: la vista `Reportes globales` del panel super ahora permite consultar reportes generales, mezclados o individuales de las empresas creadas por el administrador autenticado, reutilizando los datasets empresariales existentes y manteniendo el aislamiento por creador.
	- Verificacion: `go test ./handlers -run "TestSuperReportesGlobalesHandlerFiltraYConsolidaPorAdministrador" -count=1`; diagnostico del editor sin errores en los archivos nuevos y modificados.

## 2026-04-17
- Seleccionar empresa: licencia y descarga quedan en una sola fila.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el render de las tarjetas de seleccion de empresa ahora agrega el boton verde de descarga dentro del mismo bloque `card-actions` que usa el indicador de licencia, evitando que queden en filas separadas.
	- Verificacion: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js`.

## 2026-04-17
- Licencias super: actualizacion compatible con esquemas legacy sin `fecha_actualizacion`.
	- Archivos modificados: `backend/db/db.go`, `backend/db/licencias_schema_test.go`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: la edicion y activacion de licencias en el panel super ya no fallan cuando la tabla `licencias` viene de un esquema antiguo que no incluye `fecha_actualizacion`. El backend intenta regularizar el esquema y, si esa columna sigue ausente, aplica un `UPDATE` de compatibilidad para guardar precio y estado.
	- Verificacion: `go test ./db -run "TestEnsureLicenciasSchemaAddsValorInSQLite|TestCreateAndUpdateLicenciaRepairMissingValorColumn|TestUpdateLicenciaRepairsMissingFechaActualizacionColumn" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

## 2026-04-16
- Checkout publico de licencias: Epayco redirige la misma pestaña al checkout.
	- Archivos modificados: `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el flujo de Epayco deja de depender de una pestaña emergente y ya no arranca el polling antes de que el usuario entre a la pasarela. `pagar_licencia.html` guarda la referencia pendiente, redirige la misma pestaña a Epayco y usa `/epayco/respuesta.html` para retomar la verificacion al volver.
	- Verificacion: `GET /api/public/licencias/payment_methods` con `epayco.available=true`; `POST /epayco/create_transaction` con `checkout_url` publica valida; `GET /epayco/transaction_status?reference=<referencia_recien_creada>` con `PENDING` y `context_found=true`; diagnostico del editor sin errores en `web/pagar_licencia.html`.

- Menu flotante: separación frente a botones superiores cercanos.
	- Archivos modificados: `web/menu.js`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el menu flotante compartido ahora reserva espacio en encabezados y barras de acciones para no montarse sobre botones ubicados en la parte superior derecha de algunas paginas.
	- Verificacion: diagnostico del editor sin errores en los archivos modificados.

## 2026-04-16
- Facturacion electronica: suite `db` estable aun con entorno local en PostgreSQL.
	- Archivos modificados: `backend/db/finanzas_test.go`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: `openFinanzasTestDB` ahora fija el dialecto `sqlite` para evitar que la suite de facturacion electronica y documentos transaccionales herede `DB_DIALECT=postgres` del entorno local y falle con SQL incompatible.
	- Verificacion: `go test ./db -run "Test.*(Facturacion|DIAN|DocumentoFacturacion)" -count=1`; `go test ./handlers -run "Test(VentaCarritoFacturaYResolucionImpresora|EmpresaDIANColombiaHandler.*|EmpresaFacturacionElectronicaReintentosYReconciliacion|EmpresaFacturacionElectronicaEmiteEventoContable|EmpresaFacturacionTransaccional.*)" -count=1`.

- Pagina principal super: el campo de cantidad deja de mostrar un `5` temporal antes de cargar la configuracion real.
	- Archivos modificados: `web/super/pagina_principal.html`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el editor de `pagina_principal` ahora deja `ppCantidad` en estado de carga hasta recibir la configuracion persistida y sincroniza la cantidad con el numero real de tarjetas, evitando la confusion visual entre el panel super, `index.html` y `/descripcion_de_los_sistemas.ht`.
	- Verificacion: consulta local a `/api/public/pagina_principal` con `cantidad=7`; revision directa del flujo de carga del editor super.

## 2026-04-17
- Ventas y facturacion: prueba integrada de carrito pagado con resolucion de impresora.
	- Archivos creados: `backend/handlers/carrito_facturacion_impresion_test.go`.
	- Archivos modificados: `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se agrega una prueba de integracion de handlers que valida una venta pagada en carrito, la emision documental de factura electronica y la resolucion de la impresora `factura_caja` para el flujo de impresion soportado hoy.
	- Verificacion: `go test ./handlers -run TestVentaCarritoFacturaYResolucionImpresora -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Seleccionar empresa: restauracion del formato clasico de tarjetas.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el selector de empresas del panel super vuelve al formato simple de tarjetas `portal-card warm` usado anteriormente, retirando la presentacion enriquecida reciente.
	- Verificacion: revision del render en `web/js/seleccionar_empresa.js`; recomendada validacion visual en `seleccionar_empresa.html`.

- Portal publico: menu flotante navegable en celular.
	- Archivos modificados: `web/menu.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el menu flotante deja de cerrar sus enlaces en `touchstart`, evitando que el toque tactil cancele la navegacion en movil; ademas se mejora la respuesta del toggle y de cada opcion con `touch-action: manipulation`.
	- Verificacion: revision del flujo JS/CSS; recomendada validacion manual en movil o emulacion tactil.

- Usuarios de empresa: portal publico con contrato vigente y subdominio dedicado.
	- Archivos modificados: `backend/db/usuarios_empresa.go`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/main.go`, `web/login_usuario.html`, `web/js/login_usuario.js`, `web/estilos.css`, `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/administrar_empresa/administrar_usuarios.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: `login_usuario.html` pasa a ser el portal publico de usuarios internos creados por administradores, con registro por invitacion, recuperacion, reset, cambio de contrasena y aceptacion obligatoria del contrato vigente. El backend persiste esa aceptacion en `users`, los correos y el panel administrativo apuntan a `usuarios.powerfulcontrolsystem.com`, y el acceso final sigue entrando a `administrar_empresa.html` filtrado por rol.
	- Verificacion: `go test ./handlers -run "TestEmpresaUsuario(LoginHandlerSuccess|LoginHandlerRequiresContractAcceptance|SetPasswordHandlerSuccess|ResolveEmpresaUsuarioLoginURLUsesUsuariosSubdomain)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; diagnostico del editor sin errores en los archivos web modificados.

- Usuarios de empresa: login por subdominio propio de cada empresa.
	- Archivos modificados: `backend/handlers/usuarios_empresa.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/handlers/usuarios_empresa_seguridad_test.go`, `web/administrar_empresa.html`, `web/js/administrar_empresa.js`, `web/administrar_empresa/administrar_usuarios.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el enlace operativo del login de usuarios deja de resolverse a un host global fijo. Ahora se construye con el `empresa_slug` o `dominio_publico` configurado por empresa, tanto en el menu de `administrar_empresa` como en los correos de invitacion y recuperacion; la vista de administrar usuarios elimina el acceso duplicado fuera del menu.
	- Verificacion: `go test ./handlers -run "TestEmpresaUsuario(LoginHandlerSuccess|LoginHandlerRequiresContractAcceptance|SetPasswordHandlerSuccess|ResolveEmpresaUsuarioLoginURLUsesEmpresaSubdomain|PasswordRecoveryFlow|ChangePasswordFlow|ChangePasswordPolicyRejectsWeakPassword|LoginRequiresRotationWhenPolicyEnabled|NotificationsCaptureInMailTestMode)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; diagnostico del editor sin errores en los archivos modificados.

- Soporte remoto: limites por plan y mesa tecnica central multiempresa.
	- Archivos creados: `backend/handlers/super_soporte_remoto.go`, `backend/handlers/super_soporte_remoto_test.go`, `web/super/soporte_remoto.html`.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/db/soporte_remoto_test.go`, `backend/handlers/soporte_remoto.go`, `backend/handlers/soporte_remoto_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/main.go`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el modulo de soporte remoto ahora controla cupos de dispositivos, sesiones y minutos por empresa, persiste consumo mensual con intentos bloqueados y agrega una mesa tecnica central para `super_administrador` en `/super/api/soporte_remoto` y `super/soporte_remoto.html`.
	- Verificacion: `go test ./db ./handlers -run "Test(SoporteRemotoDB|EmpresaSoporteRemotoHandler|PublicSoporteRemotoAgentHeartbeatAndStateUpdate|SuperSoporteRemotoHandlerListsCompaniesAndCreatesSession|SuperEndpointsPermisosPorRol)" -count=1`.

## 2026-04-16
- Arranque local: healthcheck robusto en `scripts/iniciar_servidor.ps1`.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el paso `8/8` deja de reportar timeout falso cuando el backend ya esta arriba. El script ahora usa el `PORT` efectivo, detecta listener TCP con API nativa/fallback y acepta una respuesta HTTP valida para confirmar disponibilidad.
	- Verificacion: `. 'D:\powerfulcontrolsystem\scripts\iniciar_servidor.ps1'`.

- Backend: fix de compilacion en soporte remoto y bootstrap runtime.
	- Archivos modificados: `backend/db/soporte_remoto.go`, `backend/db/productos.go`, `backend/main.go`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se restaura el arranque local del backend corrigiendo la variable temporal usada al leer sesiones de soporte remoto, reescribiendo el cierre del bloque runtime en `main.go` y haciendo idempotente la regularizacion de columnas en PostgreSQL para evitar errores `column already exists` durante `scripts/iniciar_servidor.ps1`.
	- Verificacion: `go build -o server.exe .` en `backend`; `.\scripts\iniciar_servidor.ps1 -Background`.

- Estaciones: sincronizacion backend del carrito base por estacion.
	- Archivos modificados: `backend/db/empresa_estacion_prefs.go`, `backend/db/empresa_estacion_prefs_test.go`, `backend/handlers/empresa_estacion_prefs.go`, `backend/handlers/empresa_estacion_prefs_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: guardar `estaciones_config` ya no depende del frontend para crear los carritos por defecto. El backend sincroniza automaticamente un carrito enlazado por estacion, corrige nombre/codigo/referencia cuando cambia la configuracion y lo deja en estado base `inactivo/cerrado` hasta su activacion operativa.
	- Verificacion: `go test -work ./db -run "Test(EmpresaEstacionPrefs|SyncEmpresaEstacionCarritos)" -count=1`; `go test -work ./handlers -run "TestEmpresaEstacionPrefsHandler_UpsertAndIsolationByEmpresa|TestEmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`.

- Home publico: contacto centrado debajo de las tarjetas y deploy VPS con limpieza de procesos previos.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `scripts/sync_to_vps.ps1`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el home del portal deja `Informacion de contacto` como CTA centrado debajo del grid de tarjetas, manteniendo `Registrarse o iniciar sesión` en la cabecera. En paralelo, el deploy remoto endurece el reinicio del backend: purga procesos viejos de `server_linux_amd64`, corrige la unidad `systemd` para evitar el warning de `StartLimitIntervalSec` mal ubicado y asegura que el binario nuevo quede activo al terminar `sync_to_vps`.
	- Verificacion: `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; validacion de sintaxis PowerShell para `scripts/sync_to_vps.ps1`; diagnostico remoto de `systemctl status powerfulcontrolsystem` y `ss -ltnp` en el VPS.

- Checkout publico de licencias: Epayco acepta alias sambox como sandbox.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: la normalizacion del modo de Epayco ahora tolera `sambox` como alias de `sandbox`, garantizando que el checkout de licencias permanezca en pruebas (`test=true`) aunque la configuracion manual use esa variante.
	- Verificacion: `go test ./handlers -run 'TestEpaycoCreateTransactionHandler(AcceptsSamboxAlias|UsesConfiguredPublicBaseURLAndKeys|AllowsCheckoutWithoutPrivateKey)|TestResolvePaymentBaseURL(FallsBackToCanonicalDomainOnLocalhost|UsesConfiguredCanonicalDomain|IgnoresConfiguredLocalhostAndFallsBackToCanonicalDomain)|TestEpaycoTransactionStatusHandler(PreservesPendingOnGenericValidationError|FindsContextUsingInvoiceWhenGatewayIDsDiffer)' -count=1`.

- Checkout publico de licencias: Epayco legacy + metodo unico sin selector.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el checkout de Epayco vuelve a enviar `p_key` cuando la configuracion dispone de `private_key`, manteniendo compatibilidad con cuentas que exigen parametros legacy en `checkout.php`; ademas, `pagar_licencia.html` ya no muestra el cuadro de seleccion de forma de pago cuando solo una pasarela esta activa.
	- Verificacion: pendiente ejecutar pruebas focalizadas de handlers y confirmar que la URL remota de checkout deje de responder `403 AccessDenied`.

- Arcade publico: set activo de ocho juegos compactos con popup fijo y pausa real.
	- Archivos creados: `web/Juegos/arcade_window.css`, `web/Juegos/patito_volando_plus.html`, `web/Juegos/serpiente_pixel_plus.html`, `web/Juegos/memoria_estelar_plus.html`, `web/Juegos/rebote_bloques_plus.html`, `web/Juegos/pacman_plus.html`, `web/Juegos/tetris_plus.html`, `web/Juegos/carton_fire_plus.html`, `web/Juegos/ajedrez_vs_ia_plus.html`, `web/img/juegos/pacman.svg`, `web/img/juegos/tetris.svg`, `web/img/juegos/carton_fire.svg`, `web/img/juegos/ajedrez_vs_ia.svg`.
	- Archivos eliminados: `web/Juegos/pollitos_cataplum.html`, `web/img/juegos/pollitos_cataplum.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el arcade publico deja de operar con el set anterior y pasa a un lobby de ocho juegos activos con records compartidos por navegador, popup uniforme `700x700` en escritorio y pausa real en todas las experiencias, incluyendo congelacion de IA u oponentes cuando aplica.
	- Verificacion: diagnostico del editor sin errores en `web/Juegos/menu_juegos.html` y en los ocho archivos `*_plus.html` del nuevo arcade.

- Home público: botones superiores más compactos y centrados en móvil.
	- Archivos modificados: `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: los botones `Registrarse o iniciar sesión` e `Informacion de contacto` del `index.html` ahora comparten un ancho más pequeño, menor altura visual y en celular se muestran centrados dentro del header.
	- Verificacion: diagnostico del editor sin errores en `web/estilos.css`.

- Licencias super: autorreparación del esquema y validación real de guardado del valor.
	- Archivos modificados: `backend/db/db.go`, `backend/db/sql_compat.go`, `backend/db/licencias_schema_test.go`, `backend/main.go`, `web/super/licencias.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el backend ahora regulariza la tabla `licencias` tambien en PostgreSQL y reintenta `create/get/update` si faltan columnas como `valor`; la UI de super deja de ocultar errores HTTP al crear/editar licencias, mostrando el mensaje real cuando el backend rechaza la operación.
	- Verificacion: `go test ./db -run "TestEnsureLicenciasSchemaAddsValorInSQLite|TestCreateAndUpdateLicenciaRepairMissingValorColumn" -count=1`.

- Seleccionar empresa: tarjetas adaptables con contenido interno completo y márgenes estrechos.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: la vista `seleccionar_empresa.html` pasa a renderizar tarjetas con estructura interna avanzada (`empresa-card`) y estilos flexibles que permiten envolver títulos, descripciones, estados y metadatos sin cortar contenido. Se mantienen márgenes pequeños y el interior se adapta automáticamente al texto disponible.
	- Verificacion: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js` y `web/estilos.css`.

- Super pagina principal: el editor ya no recorta tarjetas cargadas por usar el valor inicial del input de cantidad.
	- Archivos modificados: `web/super/pagina_principal.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se corrige el render del editor super de `pagina_principal` para que, al recargar configuraciones con mas de 5 tarjetas, use primero `state.config.cantidad` y no el valor HTML inicial del campo `ppCantidad`. Con esto vuelven a mostrarse las 7 tarjetas guardadas y la cantidad visible queda sincronizada con la API.
	- Verificacion: inspeccion de `GET https://powerfulcontrolsystem.com/api/public/pagina_principal` con `cantidad=7`; diagnostico del editor sin errores en `web/super/pagina_principal.html`.

- Infraestructura publica: wildcard HTTPS manual para subdominios y subdominio dedicado de prueba para venta digital.
	- Archivos modificados: `documentos/manual_de_instalacion.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se documenta la emision manual del certificado wildcard `powerfulcontrolsystem.com` + `*.powerfulcontrolsystem.com`, la pauta de renovacion manual por DNS-01 y la publicacion del subdominio de prueba `venta-digital.powerfulcontrolsystem.com` hacia la pagina publica global `venta_digital.html`.
	- Verificacion: HTTPS `200` en `https://powerfulcontrolsystem.com/`, `301` de `https://www.powerfulcontrolsystem.com/` a apex, `302` de `https://venta-digital.powerfulcontrolsystem.com/` a `/venta_digital.html` y `200` final en `https://venta-digital.powerfulcontrolsystem.com/venta_digital.html`.

- Registro administrativo: captura de pais y ciudad con deteccion inicial de pais en frontend.
	- Archivos modificados: `backend/db/db.go`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/account_handlers.go`, `backend/handlers/auth_admin_handlers_test.go`, `backend/db/administradores_auth_schema_test.go`, `web/registrar_nuevo_usuario_administrador.html`, `web/js/registrar_nuevo_usuario_administrador.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el registro de administradores ahora solicita correo, nombre completo, celular, pais y ciudad. El pais se sugiere automaticamente desde el navegador/zona horaria y sigue siendo editable. El backend persiste `pais` y `ciudad` en `administradores`, y se mantiene la exigencia de confirmar el correo antes de continuar al flujo de acceso que luego lleva a `seleccionar_empresa.html`.
	- Verificacion: `go test ./db ./handlers -run 'Test(AdminRegisterHandlerCreatesPendingAdminAndCapturesConfirmationMail|AdminRegisterHandlerRejectsConfirmedExistingAdmin|EnsureAdministradoresAuthSchemaAddsMissingColumnsInSQLite|SetAdministradorPasswordRepairsMissingSecurityColumns)$' -count=1`.

- Autenticacion administrativa: compatibilidad del esquema `administradores` entre SQLite y PostgreSQL.
	- Archivos creados: `backend/db/administradores_auth_schema_test.go`.
	- Archivos modificados: `backend/db/db.go`, `backend/main.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se centraliza la regularizacion de columnas de seguridad de `administradores` con soporte para SQLite y PostgreSQL mediante `EnsureAdministradoresAuthSchema`, y `SetAdministradorPassword` reintenta la operacion cuando encuentra columnas faltantes. Con esto se corrige el flujo donde una cuenta autenticada por Google no podia registrar su primera contrasena local en VPS con PostgreSQL.
	- Verificacion: `go test ./db ./handlers -run 'Test(EnsureAdministradoresAuthSchemaAddsMissingColumnsInSQLite|SetAdministradorPasswordRepairsMissingSecurityColumns|AccountSetGooglePasswordHandlerCreatesInitialPassword)$' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; validacion operativa en VPS de `systemd`, `Nginx`, `UFW`, callback OAuth y dominio publico.

- Super: modulo de seguridad VPS Linux con panel, CLI, cron y exportes multiformato.
	- Archivos creados: `backend/vpssecurity/config/config.go`, `backend/vpssecurity/config/default_vps_security_config.json`, `backend/vpssecurity/parser/lynis.go`, `backend/vpssecurity/parser/nmap.go`, `backend/vpssecurity/parser/trivy.go`, `backend/vpssecurity/scanner/runner.go`, `backend/vpssecurity/scanner/checks.go`, `backend/vpssecurity/reports/report.go`, `backend/vpssecurity/reports/report_test.go`, `backend/vpssecurity/logs/store.go`, `backend/vpssecurity/service.go`, `backend/handlers/security_vps_handlers.go`, `backend/handlers/security_vps_handlers_test.go`, `backend/tools/vps_security_scan/main.go`, `web/js/super_seguridad.js`, `scripts/install_vps_security_tools.sh`, `scripts/run_vps_security_scan.sh`, `scripts/install_vps_security_cron.sh`, `documentos/manual_vps_seguridad.md`.
	- Archivos modificados: `backend/main.go`, `web/super/seguridad.html`, `web/index.html`, `web/estilos.css`, `.gitignore`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se agrega un modulo completo de seguridad VPS Linux para `super_administrador`, con ejecucion de Lynis/Nmap/Trivy y chequeos propios, historial/comparacion en filesystem, exportes `JSON/TXT/HTML/CSV/PDF/XLS`, CLI reutilizable y scripts Ubuntu para instalacion y cron. En el portal publico, `Informacion de contacto` queda anclado al extremo derecho de la misma fila superior del home.
	- Verificacion: `go test ./vpssecurity/... ./handlers ./tools/vps_security_scan -run "TestSecurityVPS|TestGenerateArtifacts|TestCompareDetects" -count=1`; diagnóstico del editor sin errores en los archivos Go/HTML/JS/SH modificados para este cambio.

- Login unificado: eliminado `recordar usuario/cuenta` y retirado `login_hint` en OAuth.
	- Archivos modificados: `web/login.html`, `web/js/login.js`, `web/login_usuario.html`, `web/js/login_usuario.js`, `web/menu.js`, `web/js/super_administrador.js`, `web/js/seleccionar_empresa.js`, `web/super/licencias.html`, `web/super/tipos_empresas.html`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se elimina la persistencia local de cuenta/usuario recordado en ambos logins y se deja `/auth/google/login` sin `login_hint`. Con esto, el acceso queda consistente entre `localhost:8080`, `powerfulcontrolsystem.com` y `www.powerfulcontrolsystem.com`, sin depender de estado guardado por dominio en `localStorage`.
	- Verificacion: `go test -work ./handlers -run "TestHandleGoogleLoginRedirect|TestAccountSetGooglePasswordHandlerCreatesInitialPassword|TestE2E_AcceptContractCreatesSession|TestAdminLoginHandlerCreatesSessionForConfirmedAdmin" -count=1`.

- Login Google: registro obligatorio de contrasena local cuando falta password_set.
	- Archivos creados: `web/registrar_contrasena_usuario_de_google.html`, `web/js/registrar_contrasena_usuario_de_google.js`.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/account_handlers.go`, `backend/main.go`, `backend/handlers/auth_admin_handlers_test.go`, `backend/handlers/e2e_login_acceptance_test.go`, `web/ayuda/login_administradores.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el callback Google y la aceptación de contrato ya no envían al panel final cuando la cuenta administrativa aún no tiene contraseña local; ahora redirigen a `registrar_contrasena_usuario_de_google.html`, que guarda la primera clave mediante `/api/account/set_google_password` para habilitar después el acceso por correo y contraseña.
	- Verificacion: `go test -work ./handlers -run "Test(AccountSetGooglePasswordHandlerCreatesInitialPassword|E2E_AcceptContractCreatesSession|AdminLoginHandlerCreatesSessionForConfirmedAdmin)" -count=1`.

- Super: panel PostgreSQL con carga de tamaño por empresa.
	- Archivos modificados: `backend/handlers/postgres_performance.go`, `backend/handlers/postgres_performance_test.go`, `web/super/administrar_base_de_datos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el panel super de administracion PostgreSQL ahora puede cargar bajo demanda una tabla de consumo estimado por empresa dentro de `pcs_empresas`, ordenada de mayor a menor y mostrando tambien filas estimadas, tablas con datos y la tabla mas pesada por empresa.
	- Verificacion: `go test -work ./handlers -run "TestPostgresPerformanceHandler|TestHumanizeBytesBinary" -count=1`.

- Manual de instalacion: agregado el paso de respuesta, confirmacion y formulario exacto de Epayco.
	- Archivos modificados: `documentos/manual_de_instalacion.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el manual ya incluye las URLs exactas que deben configurarse en Epayco para respuesta y confirmacion, ademas de los valores concretos del formulario de Epayco y una nota operativa sobre el flujo real de validacion del pago.

- Checkout y seleccion de empresa: ajuste visual solicitado.
	- Archivos modificados: `web/pagar_licencia.html`, `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: `pagar_licencia.html` deja mas clara la pasarela activa cuando solo hay un metodo disponible y muestra el logo de Epayco en el selector y en el panel. `seleccionar_empresa.html` vuelve al estilo compacto anterior de tarjetas para empresas.

- Checkout de licencias: Epayco ahora usa una pagina publica fija de respuesta.
	- Archivos creados: `web/epayco/respuesta.html`.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el retorno de Epayco ya no depende de enviar al usuario directamente a `pagar_licencia.html`; ahora existe la landing publica fija `/epayco/respuesta.html`, que puedes registrar en el panel de Epayco y que reenvia al resumen del pago con el contexto necesario para validar y activar la licencia.
	- Verificacion: `go test -work ./handlers -run "TestEpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|TestEpaycoCreateTransactionHandlerAllowsCheckoutWithoutPrivateKey|TestEpaycoTransactionStatusHandlerPreservesPendingOnGenericValidationError|TestResolvePaymentBaseURL" -count=1`.
- Login administrativo: registro separado, confirmación pública corregida y recuperación sin prompts.
	- Archivos creados: `web/registrar_nuevo_usuario_administrador.html`, `web/js/registrar_nuevo_usuario_administrador.js`, `backend/handlers/auth_admin_handlers_test.go`.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/utils/utils.go`, `web/login.html`, `web/js/login.js`, `web/estilos.css`, `web/ayuda/login_administradores.html`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el login administrativo ahora deja el registro en una página pública específica, elimina el campo de nombre incrustado del acceso principal, centra `Iniciar por correo` y agrega debajo `Registrarse` y `¿Olvidó su contraseña?`. El backend valida `nombre`, `telefono` y contraseña segura, evita sobrescribir cuentas confirmadas, corrige el whitelist público para `/auth/confirmar_admin` y sustituye la recuperación por formularios reales dentro de `login.html`.
	- Verificacion: `go test -work ./handlers -run "Test(Admin|AuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI|HandleGoogleLogin|E2E_AcceptContractCreatesSession)" -count=1`; `go test -work ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

- Arcade publico: Patito volando ahora inicia con cuenta regresiva y los cinco juegos refuerzan su modo celular.
	- Archivos modificados: `web/Juegos/arcade_shared.js`, `web/Juegos/patito_volando.html`, `web/Juegos/pollitos_cataplum.html`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html`, `web/Juegos/rebote_bloques.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el arcade publico mantiene sonido compartido en los cinco juegos, `Patito volando` arranca con cuenta regresiva de 5 segundos y el resto del arcade ajusta shells, overlays y acciones para celular. Tambien se agregan sonidos de countdown en `arcade_shared.js` y `Serpiente pixel` suma feedback sonoro al giro durante la partida.
	- Verificacion: validacion sin errores de los seis archivos del arcade modificados y revision de los nuevos breakpoints moviles y del countdown previo al inicio en `Patito volando`.

- Frontend compartido: mejoras base de adaptacion movil y menu flotante.
	- Archivos modificados: `web/menu.js`, `web/estilos.css`, `web/login.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el menu flotante ahora se cierra al seleccionar una opcion, el CTA de WhatsApp de la portada pasa a un icono compacto abajo a la derecha en movil para no tapar otros botones, la capa CSS compartida mejora tablas/sidebar/panel flotante en pantallas pequenas y `login.html` vuelve a cargar la hoja `estilos.css` correcta.
	- Verificacion: validacion sin errores de `web/menu.js`, `web/estilos.css` y `web/login.html`, mas revision de los breakpoints moviles del menu flotante y del CTA de WhatsApp.

- Portal publico: botones superiores de la portada ahora usan el mismo estilo de Explorar oferta.
	- Archivos modificados: `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: los accesos `Registrarse o iniciar sesión` e `Informacion de contacto` del encabezado de `index.html` reutilizan la misma apariencia visual del boton `Explorar oferta` de las tarjetas del home, sin cambiar rutas ni comportamiento responsive.
	- Verificacion: revision del bloque compartido de selectores en `web/estilos.css` y del ajuste pill en `@media (max-width:560px)`.

- Checkout de licencias: Epayco sandbox estable con bootstrap PostgreSQL y polling pendiente consistente.
	- Archivos modificados: `backend/db/db.go`, `backend/main.go`, `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el backend asegura `pagos_epayco` y `pagos_wompi` al arrancar sobre PostgreSQL y deja de degradar a `ERROR` una referencia de Epayco que sigue `PENDING` localmente mientras la validacion externa responde un error transitorio. Ademas se normaliza la configuracion sandbox operativa (`epayco.*` y `gmail.confirm_base_url`) en la base super del VPS para que el checkout genere callbacks publicos validos.
	- Verificacion: `go test ./handlers -run 'TestResolvePaymentBaseURL|TestEpayco(CreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|CreateTransactionHandlerAllowsCheckoutWithoutPrivateKey|TransactionStatusHandlerPreservesPendingOnGenericValidationError)' -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; validacion manual local de `GET /api/public/licencias/payment_methods`, `POST /epayco/create_transaction` y `GET /epayco/transaction_status` tras recompilar con `scripts/iniciar_servidor.ps1 -Background`.

- Portal publico: arcade con cinco juegos, tarjetas cuadradas y perfil compartido.
	- Archivos creados: `web/Juegos/arcade_shared.js`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html`, `web/Juegos/rebote_bloques.html`, `web/img/juegos/patito_volando.svg`, `web/img/juegos/pollitos_cataplum.svg`, `web/img/juegos/serpiente_pixel.svg`, `web/img/juegos/memoria_estelar.svg`, `web/img/juegos/rebote_bloques.svg`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `web/Juegos/patito_volando.html`, `web/Juegos/pollitos_cataplum.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el lobby publico de Juegos se convierte en un arcade visual con portadas cuadradas, panel de jugador y resumen de records. `arcade_shared.js` centraliza nombre, top local y sonido; `Patito volando` y `Pollitos al cataplum` se integran a esa capa y se agregan tres juegos nuevos: `Serpiente pixel`, `Memoria estelar` y `Rebote de bloques`.
	- Verificacion: diagnostico del editor sin errores en `web/Juegos/arcade_shared.js`, `web/Juegos/menu_juegos.html`, `web/Juegos/patito_volando.html`, `web/Juegos/pollitos_cataplum.html`, `web/Juegos/serpiente_pixel.html`, `web/Juegos/memoria_estelar.html` y `web/Juegos/rebote_bloques.html`.

## 2026-04-15
- Portal publico: nuevo juego `Pollitos al cataplum` y menu de Juegos multi-tarjeta.
	- Archivos creados: `web/Juegos/pollitos_cataplum.html`.
	- Archivos modificados: `web/Juegos/menu_juegos.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se agrega un segundo juego publico de resortera con niveles cortos, puntaje y control arrastrar/soltar; ademas, el catalogo de Juegos ahora soporta varias tarjetas con popup propio por juego.

- Licencias: Epayco/Wompi ya no fallan por resolver `localhost` al iniciar checkout.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: `resolvePaymentBaseURL(...)` ahora ignora loopback en configuracion o request, intenta `gmail.confirm_base_url`, `Origin`/`Referer`, host publicado y, si hace falta, cae al dominio canonico `https://powerfulcontrolsystem.com` para construir callbacks publicos validos del checkout.
	- Verificacion: `go test ./handlers -run "Test(ResolvePaymentBaseURLFallsBackToCanonicalDomainOnLocalhost|ResolvePaymentBaseURLUsesConfiguredCanonicalDomain|ResolvePaymentBaseURLIgnoresConfiguredLocalhostAndFallsBackToCanonicalDomain|EpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|EpaycoCreateTransactionHandlerAllowsCheckoutWithoutPrivateKey)" -count=1`.

- Servidor: alerta de inicio/reinicio ahora puede activarse o desactivarse desde configuracion avanzada.
	- Archivos modificados: `backend/handlers/server_runtime_notifications.go`, `backend/handlers/server_runtime_notifications_test.go`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/handlers/super_config_backup_handlers.go`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el backend ya registraba el arranque/reinicio en `super_servidor_eventos`, en `backend/logs/server_reinicio.log` y enviaba correo cuando existia `gmail.restart_alert_to`. Ahora se agrega `gmail.restart_alert_enabled` para activar o desactivar ese correo desde `configuracion_avanzada.html` sin perder el destinatario configurado.
	- Verificacion: `go test ./handlers -run "Test(GmailConfigHandlerSaveRestartAlertTo|GmailConfigHandlerSaveRestartAlertToggle|RegisterServerStartupEventCapturesNotificationAndState|RegisterServerStartupEventDetectsUnexpectedRestart|RegisterServerStartupEventSkipsEmailWhenAlertsDisabled)" -count=1`.

- Seleccion de empresas: tarjetas con iconografia por tipo y rediseño mas profesional.
	- Archivos modificados: `web/js/seleccionar_empresa.js`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: `seleccionar_empresa.html` ahora presenta cada empresa con icono segun `tipo_nombre`, tono visual por categoria, chips de estado y una tarjeta mas colorida/profesional. Se conserva el mismo flujo para abrir la administracion o continuar con la licencia.
	- Verificacion: diagnostico del editor sin errores en `web/js/seleccionar_empresa.js` y `web/estilos.css`.

- Pagina principal: la cantidad de tarjetas ahora se aplica y se guarda en un solo flujo.
	- Archivos modificados: `web/super/pagina_principal.html`, `backend/handlers/pagina_principal_handlers_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se elimina el paso manual `Aplicar cantidad` del editor super de `pagina_principal`. Al cambiar la cantidad, el editor reconstruye las tarjetas visibles y el mismo flujo de `Guardar configuracion` persiste cantidad, contenido y estilos. Ademas se agrega una prueba de persistencia para configuraciones ampliadas.
	- Verificacion: `go test ./handlers -run "TestPaginaPrincipal|TestPublicPaginaPrincipalHandlerExposesLandingFields" -count=1`.

- Portal publico: nuevo menu de juegos y primer juego `Patito volando`.
	- Archivos creados: `web/Juegos/menu_juegos.html`, `web/Juegos/patito_volando.html`.
	- Archivos modificados: `backend/utils/utils.go`, `backend/handlers/auth_users_carritos_test.go`, `web/menu.js`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se agrega la entrada `Juegos` al menu flotante del portal, se crea `menu_juegos.html` con una tarjeta por juego publicado y se implementa `patito_volando.html` como minijuego de ventana pequena con control por barra espaciadora en PC y toque/presion en movil. `AuthMiddleware` deja publico `/Juegos/*` y la prueba del middleware se amplía para cubrir estas rutas.
	- Verificacion: `go test ./handlers -run "TestAuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI" -count=1`.

## 2026-04-15
- Repositorio: restaurado `Pendiente Notas` y auditados los borrados actuales.
	- Archivos modificados: `Pendiente Notas`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se recupera `Pendiente Notas` desde `HEAD` tras detectar que Git lo tenia como borrado local en el arbol de trabajo. La auditoria posterior confirma que no habia otros archivos eliminados en el estado actual del repo. Git no conserva una hora exacta para ese borrado local no confirmado; la ultima hora verificable en historial para el archivo es `2026-04-15 17:37:25 -0500` en el commit `e70884dabea1292d9c0e6d9b1a3f236e94d7c8c4`.
	- Verificacion: `git diff --name-status --diff-filter=D`; `git status --short --untracked-files=no`; `git log -1 --format="%H%n%an%n%ad%n%s" -- "Pendiente Notas"`; `Get-Item -LiteralPath "d:\powerfulcontrolsystem\Pendiente Notas" | Select-Object FullName,Length,CreationTime,LastWriteTime`.

- Errores del sistema: monitor centralizado, recovery global y panel super.
	- Archivos creados: `backend/db/super_errores_sistema.go`, `backend/utils/system_errors.go`, `backend/handlers/super_error_handlers.go`, `backend/handlers/super_error_handlers_test.go`, `web/super/errores.html`.
	- Archivos modificados: `backend/main.go`, `backend/utils/utils.go`, `backend/utils/utils_test.go`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se implementa un sistema robusto de manejo de errores para todo el proyecto. Los errores HTTP y panicos recuperados se registran en `super_errores_sistema` y en `backend/logs/system_errors.log`, el cliente deja de recibir detalles tecnicos en respuestas `5xx` y super obtiene un panel profesional para monitoreo transversal por empresa, fecha, severidad y tipo.
	- Verificacion: `go test ./utils -run "Test(JSONErrorMiddlewareSanitizesInternalServerError|RecoveryMiddlewareRecoversPanicAndLogsIt|JSONErrorMiddlewarePreservesJSONErrorBody|JSONErrorMiddlewareWrapsNonJSONError|JSONErrorMiddlewareWrapsEpaycoNonJSONError)" -count=1`; `go test ./handlers -run "Test(SuperErroresSistemaHandlerFiltersResults|SuperErroresSistemaHandlerMethodNotAllowed)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

- Contrato administrativo: ahora es versionado, editable desde super y exigido por version en el login.
	- Archivos creados: `backend/db/contrato_super.go`, `backend/handlers/super_contrato_handlers.go`, `backend/handlers/super_contrato_handlers_test.go`, `web/super/contrato.html`.
	- Archivos modificados: `backend/main.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/e2e_login_acceptance_test.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/utils/utils.go`, `web/accept.html`, `web/contrato.html`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el contrato que aceptan los administradores deja de ser un HTML estatico y pasa a vivir en la base `superadministrador`, con historial por version y resumen de cambio. Super puede editarlo desde una pagina dedicada, el portal lo publica via `/api/public/contrato` y el login administrativo exige aceptar la ultima version antes de crear sesion.
	- Verificacion: `go test ./handlers -run "Test(PublicContratoHandlerReturnsDefaultVersion|SuperContratoHandlerCreatesNewVersionAndHistory|E2E_AcceptContractCreatesSession|E2E_AcceptContractRequiresNewVersion|AuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en los archivos Go/HTML tocados.

- Deploy VPS: sync_to_vps ahora abre el dominio publico canonico en lugar de la IP.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: `sync_to_vps.ps1` agrega `PublicBaseUrl` con valor por defecto `https://powerfulcontrolsystem.com/` y usa esa URL al finalizar el deploy, manteniendo `RemoteHost` solo para SSH y evitando abrir `http://<ip>:<puerto>/` en el navegador.
	- Verificacion: validacion de sintaxis PowerShell mediante parser (`[System.Management.Automation.Language.Parser]::ParseFile(...)`) sin errores.

- Checkout de licencias: Epayco disponible con Public Key y rutas publicas de pago realmente abiertas.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/utils/utils.go`, `backend/utils/utils_test.go`, `web/pagar_licencia.html`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el backend deja de exigir `epayco.private_key` para mostrar Epayco en el checkout actual y usa `epayco.public_key` como requisito minimo operativo junto al flag `enabled`. Tambien se corrige `AuthMiddleware` para dejar publicas `/api/public/licencias/payment_methods`, `/wompi/*` y `/epayco/*`, y `web/pagar_licencia.html` ahora indica si la pasarela esta desactivada o si falta la `Public Key`.
	- Verificacion: `go test ./handlers -run "Test(EpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|EpaycoCreateTransactionHandlerAllowsCheckoutWithoutPrivateKey|PublicLicenciasPaymentMethodsHandlerOrdersAndAvailability|PublicLicenciasPaymentMethodsHandlerAllowsEpaycoWithPublicKeyOnly)" -count=1`; `go test ./utils -run "Test(AuthMiddlewareAllowsPublicLicenciaPaymentRoutesWithoutSession|JSONErrorMiddlewareWrapsEpaycoNonJSONError)" -count=1`; diagnostico del editor sin errores en los archivos Go/HTML tocados.

- Login admin y configuración Gmail: se simplifica el hint visual y se habilita edición directa.
	- Archivos modificados: `web/login.html`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el login administrativo ya no muestra el bloque `Se recordará ... / Olvidar`, aunque conserva la logica de `Recordar cuenta`, y la seccion Gmail SMTP del panel super deja de bloquear el correo remitente y los demas campos cuando ya existe una configuracion guardada.
	- Verificacion: `go test ./handlers -run "TestGmailConfigHandlerSaveRestartAlertTo" -count=1`; diagnostico del editor sin errores en los archivos HTML tocados.

- Portal publico: pagina_principal ahora define tamanos de tarjetas y texto para home y landing.
	- Archivos modificados: `backend/handlers/pagina_principal_handlers.go`, `backend/handlers/pagina_principal_handlers_test.go`, `web/super/pagina_principal.html`, `web/index.html`, `web/descripcion_de_los_sistemas.ht`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el editor super de pagina_principal agrega ajustes globales de tamano para tarjetas y texto del `index.html` y de `/descripcion_de_los_sistemas.ht`. La API publica mantiene un contrato unico (`tarjetas` + `estilos`) y el frontend aplica esos valores de forma responsive.
	- Verificacion: `go test ./handlers -run "TestPaginaPrincipal|TestPublicPaginaPrincipalHandlerExposesLandingFields" -count=1`; diagnostico del editor sin errores en los archivos Go/HTML/CSS tocados.

- Portal publico: CTA de WhatsApp arriba a la derecha y botones del header con estilo mini-tarjeta.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el home comercial reposiciona el CTA flotante `Contactenos` a la esquina superior derecha y convierte `Registrarse o iniciar sesión` e `Informacion de contacto` en accesos compactos con acabado visual de mini-tarjeta, reutilizando el lenguaje de las tarjetas del portal sin alterar rutas publicas ni comportamiento funcional.
	- Verificacion: diagnostico del editor sin errores en `web/index.html` y `web/estilos.css`.

- Portal publico: la landing descriptiva ahora se configura desde pagina_principal.
	- Archivos creados: `backend/handlers/pagina_principal_handlers_test.go`.
	- Archivos modificados: `backend/handlers/pagina_principal_handlers.go`, `web/super/pagina_principal.html`, `web/descripcion_de_los_sistemas.ht`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: la configuracion super de pagina_principal deja de limitarse al home y ahora tambien guarda la etiqueta, titular ampliado, parrafos y capacidades clave de cada tarjeta para `/descripcion_de_los_sistemas.ht`. La landing descriptiva deja de depender de textos estaticos por nombre de sistema y renderiza el contenido extendido desde la misma API publica usada por `index.html`.
	- Verificacion: `go test ./handlers -run "Test(PaginaPrincipal|AuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en `backend/handlers/pagina_principal_handlers.go`, `backend/handlers/pagina_principal_handlers_test.go`, `web/super/pagina_principal.html` y `web/descripcion_de_los_sistemas.ht`.

- Checkout de licencias: retorno recuperable tras volver de Epayco y Wompi.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/payments_handlers_test.go`, `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el checkout de licencias ya no depende de un `status` estatico en la URL al volver desde la pasarela. Backend y frontend ahora conservan `provider`, `reference`, `transaction_id`, `licencia_id` y `empresa_id`, reanudan la verificacion real del pago desde `web/pagar_licencia.html` y permiten que Wompi consulte estado por `reference` cuando el navegador regresa sin `transaction_id` directo.
	- Verificacion: `go test ./handlers -run "Test(EpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|WompiTransactionStatusHandlerAllowsReferenceLookup|ResolvePaymentBaseURLRejectsLocalhostWithoutPublicConfig|ResolvePaymentBaseURLUsesConfiguredCanonicalDomain|PublicLicenciasPaymentMethodsHandlerOrdersAndAvailability)" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en `web/pagar_licencia.html`, `backend/handlers/payments_handlers.go` y `backend/handlers/payments_handlers_test.go`.

- Checkout de licencias: fix Epayco con `public_key` real y callbacks sobre dominio público.
	- Archivos creados: `backend/handlers/payments_handlers_test.go`.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/system_empresas_handlers_test.go`, `web/super/configuracion_avanzada.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se corrige el flujo de Epayco para separar `public_key`, `private_key` y `customer_id`, mantener compatibilidad con configuraciones legacy y resolver `response`/`confirmation` desde una base pública válida en vez de `localhost`. La pantalla de configuración avanzada deja de confundir la llave pública con el identificador del comercio y Wompi reutiliza la misma base pública para su `redirect_url`.
	- Verificacion: `go test ./handlers -run "TestResolvePaymentBaseURL|TestEpaycoCreateTransactionHandlerUsesConfiguredPublicBaseURLAndKeys|TestPublicLicenciasPaymentMethodsHandlerOrdersAndAvailability" -count=1`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`; diagnostico del editor sin errores en los archivos tocados.

- Login Google: host canónico en dominio raíz y estaciones con carga visible.
	- Archivos modificados: `backend/utils/utils.go`, `backend/utils/utils_test.go`, `backend/main.go`, `backend/.env.example`, `scripts/sync_to_vps.ps1`, `web/administrar_empresa/estaciones.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se corrige la inestabilidad del acceso administrativo tras registrar el dominio público, redirigiendo `www.powerfulcontrolsystem.com` al host canónico `powerfulcontrolsystem.com` antes de procesar OAuth y alineando los defaults de `GOOGLE_REDIRECT_URL` al callback del dominio raíz. Además, la página de estaciones ahora muestra `Cargando estaciones...` mientras consulta configuración, carritos y sensores, con mensaje visible en caso de error.
	- Verificacion: `go test ./utils -run "Test(CanonicalPublicHostMiddleware|LoggingMiddlewareSetsContextAndWritesLogs|JSONErrorMiddlewareWrapsNonJSONError)" -count=1`; `go test ./handlers -run "TestHandleGoogleLogin" -count=1`; diagnóstico del editor sin errores nuevos en los archivos tocados.

- Portal publico: home, landing descriptiva y contacto liberados sin sesion.
	- Archivos modificados: `backend/utils/utils.go`, `backend/handlers/auth_users_carritos_test.go`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: `AuthMiddleware` incorpora `/descripcion_de_los_sistemas.ht` y `/Informacion_de_contacto.html` al whitelist publico y mantiene `index.html` dentro del mismo conjunto, para que las tres paginas comerciales del portal sean accesibles sin login. La prueba de middleware se amplia para cubrir esas rutas junto con `menu.js` y `/api/public/pagina_principal`.
	- Verificacion: `go test ./handlers -run "TestAuthMiddlewareAllowsPublicPortalPagesAssetsAndHomeCardsAPI" -count=1`; diagnostico del editor sin errores en Go y documentos modificados.

- Portal publico: contacto visible por WhatsApp y pagina dedicada de informacion.
	- Archivos creados: `web/Informacion_de_contacto.html`.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se agrega un CTA flotante `Contactenos` en `index.html` que abre WhatsApp con el numero comercial del sistema, un acceso visible a `Informacion_de_contacto.html` desde el encabezado del portal y una nueva pagina publica con descripcion general del sistema, correo `powerfulcontrolsystem@hmail.com` y WhatsApp `3043306506`. Ademas, el acceso principal del header pasa a llamarse `Registrarse o iniciar sesión` y queda junto al boton de contacto.
	- Verificacion: diagnostico del editor sin errores en `web/index.html`, `web/Informacion_de_contacto.html` y `web/estilos.css`.

- Portal publico: landing descriptiva unica para todas las tarjetas del home.
	- Archivos creados: `web/descripcion_de_los_sistemas.ht`.
	- Archivos modificados: `web/index.html`, `web/super/pagina_principal.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el boton `Explorar oferta` del home deja de abrir enlaces directos y pasa a una sola landing publica (`/descripcion_de_los_sistemas.ht`) con anclas por tarjeta, descripciones ampliadas por seccion y un CTA `Probar Gratis` por cada solucion. El enlace configurado desde `super/pagina_principal.html` ahora alimenta ese CTA final.
	- Verificacion: diagnostico del editor sin errores en los archivos HTML/CSS modificados.

- Checkout de licencias: Epayco primero, Wompi debajo y activacion real desde configuracion avanzada.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/handlers/system_empresas_handlers_test.go`, `backend/main.go`, `web/pagar_licencia.html`, `web/super/configuracion_avanzada.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se agrega la ruta publica `GET /api/public/licencias/payment_methods` para publicar la disponibilidad ordenada de pasarelas de licencia, `web/pagar_licencia.html` ya muestra solo Epayco y Wompi con prioridad Epayco -> Wompi, y `web/super/configuracion_avanzada.html` permite activar/desactivar ambas pasarelas manteniendo a Wompi bloqueado en backend cuando esta desactivado o incompleto.
	- Verificacion: `go test ./handlers -run "TestPublicLicenciasPaymentMethodsHandlerOrdersAndAvailability|TestWompiConfigHandlerPersistsEnabledFlag|TestWompiTermsHandlerRejectsWhenDisabled" -count=1`; `go test ./ -run "^$" -count=1`; diagnostico del editor sin errores en HTML/CSS/Go tocados.

- Sync VPS: reparacion del redeploy remoto en fallback PuTTY.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el wrapper PowerShell deja de pasar inline a `plink` los bloques remotos complejos de `bootstrap` y `redeploy`; ahora los escribe en archivos temporales UTF-8 sin BOM y los ejecuta con `plink -m`, estabilizando el `heredoc` de la unidad `systemd` y evitando fallos Bash como `syntax error near unexpected token '('`. Tambien se endurece el quoting del binario remoto y de los directorios de logs.
	- Verificacion: parser PowerShell en verde para `scripts/sync_to_vps.ps1` y diagnostico del editor sin errores nuevos en el archivo.

- Login Google: hardening de `login_hint` y saneamiento de cuenta recordada en escritorio.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`, `web/js/login.js`, `web/menu.js`, `web/js/super_administrador.js`, `web/js/seleccionar_empresa.js`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se evita que el login Google herede un `login_hint` corrupto desde el navegador. El backend solo reenvia hints con formato de correo valido y el frontend limpia/persiste `rememberedEmail` unicamente cuando el dato es plausible, estabilizando el flujo especialmente en escritorio.
	- Verificacion: `go test ./handlers -run "TestHandleGoogleLogin" -count=1` y `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

- Frontend web: refuerzo responsive transversal para portal y paneles administrativos.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se mejora la adaptacion entre escritorio, tablet y movil en la portada publica y en los layouts compartidos. El hero de `index.html` permite salto natural del titulo/subtitulo, el sidebar administrativo colapsa con mejor navegacion horizontal en movil y formularios/tablas/botones se reorganizan para pantallas estrechas.
	- Verificacion: diagnostico del editor sin errores en `web/index.html` y `web/estilos.css`.

- VPS web: restauracion del dominio publico sin puerto con Nginx, UFW y TLS correctos.
	- Archivos modificados: `documentos/deploy_nginx_reverse_proxy_vps.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se corrige la incidencia de publicacion donde `powerfulcontrolsystem.com` dejaba de cargar externamente aunque el backend y Nginx estaban activos en el VPS; la causa fue `443/tcp` ausente en UFW y cobertura incompleta de `www` en TLS. Se abre `443/tcp`, se renueva el certificado LetsEncrypt para `powerfulcontrolsystem.com` y `www.powerfulcontrolsystem.com`, y se documenta la configuracion minima correcta para Nginx/Certbot.
	- Verificacion: `curl -I https://powerfulcontrolsystem.com/` y `curl -I https://www.powerfulcontrolsystem.com/` responden `200 OK`; `certbot certificates` muestra ambos dominios; `ufw status` incluye `443/tcp ALLOW`.

- Sync VPS: limpieza automatica de procesos huerfanos antes del restart remoto.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se endurece el redeploy remoto para detener listeners fuera de `systemd` que sigan ocupando `SERVER_PORT`, registrar el PID/comando conflictivo y abortar con diagnostico si el puerto no se libera; adicionalmente se saneó el VPS donde un binario `server_linux_amd64 (deleted)` mantenía `:8080` ocupado y dejaba `powerfulcontrolsystem.service` en bucle de reinicio.
	- Verificacion: en VPS `powerfulcontrolsystem.service` quedó `active (running)` tras limpiar el listener huérfano; `curl -k -I https://powerfulcontrolsystem.com/auth/google/login` sigue devolviendo `302` con `redirect_uri=https://powerfulcontrolsystem.com/auth/google/callback`.

- Sync VPS: bootstrap endurecido, mensajes accionables y preparación asistida del servidor.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se robustecen ambos scripts de despliegue para validar puertos/timeout antes de conectar, detectar el gestor de paquetes remoto e instalar dependencias base del VPS cuando hay privilegios, actualizar `SERVER_PORT` en cada bootstrap, exigir mensajes etiquetados `BOOTSTRAP_*`/`DEPLOY_*` y devolver hints claros cuando fallan DSN PostgreSQL, `CONFIG_ENC_KEY`, permisos `root/sudo` o el arranque del servicio `systemd`.
	- Verificacion: parser de PowerShell en verde para `scripts/sync_to_vps.ps1`; diagnostico del editor sin errores para `scripts/sync_to_vps.sh`; previsualizacion `./scripts/sync_to_vps.ps1 -PreviewOnly -SkipBuild -OpenPublicUrlAfterDeploy:$false` generando correctamente las etapas remotas; la validacion directa `bash -n` sigue pendiente en este equipo porque `bash.exe` apunta al lanzador de WSL y no hay distribucion instalada.

- Login y menu: correccion de `recordar cuenta` y deteccion visible de sesion.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/usuarios_empresa.go`, `backend/handlers/e2e_login_acceptance_test.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/main.go`, `web/js/login.js`, `web/menu.js`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se deja de depender de la lectura cliente de `session_token` para sincronizar `recordar cuenta`, avatar y enlace de cierre de sesion; el backend emite `browser_session_active` como señal visible no sensible, manteniendo el token real en cookie `HttpOnly` y alineando tambien la limpieza de cookies en logout.
	- Verificacion: `go test ./handlers -run "TestE2E_AcceptContractCreatesSession|TestEmpresaUsuario(LoginHandlerSuccess|LoginHandlerRejectsWrongEmpresaScope|LoginHandlerRejectsWrongEmpresaScopeFromQuery|SetPasswordHandlerSuccess)|TestSuperEndpointsPermisosPorRol" -count=1` y `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1`.

## 2026-04-14
- Sync VPS: backend persistente con `systemd` y autoarranque tras reinicio del servidor.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/deploy_nginx_reverse_proxy_vps.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: el despliegue al VPS deja de usar `nohup` para reiniciar el backend y pasa a instalar/actualizar una unidad `systemd` del proyecto con `Restart=always`, `systemctl enable`, carga de entorno desde `backend/.env.local` y logs persistentes en `backend/server.log` / `backend/server.err`, garantizando que el servicio vuelva solo tras caidas del proceso o reinicios del VPS y que solo se reinicie durante `sync_to_vps`.
	- Verificacion: parser de PowerShell para `scripts/sync_to_vps.ps1`, previsualizacion local del script con `-PreviewOnly -SkipBuild` y diagnostico del editor sin errores para `scripts/sync_to_vps.sh`; la validacion directa con `bash -n` queda pendiente en este equipo porque no hay distro WSL ni Git Bash instalados.

## 2026-04-14
- Manual de instalacion: reposicion del documento y guia Google OAuth para VPS.
	- Archivos creados: `documentos/manual_de_instalacion.md`.
	- Archivos modificados: `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se repone el manual eliminado en `HEAD` y se actualiza con la configuracion exacta de Google Cloud Console para login local y produccion, incluyendo `Authorized redirect URIs` y `Authorized JavaScript origins` para `localhost` y `powerfulcontrolsystem.com`, mas notas de diagnostico para `redirect_uri_mismatch`.
	- Verificacion: revision documental del manual recreado y comprobacion estatica de las URLs de callback/origen documentadas.

- Portal principal: título en una sola línea con subtítulo debajo en la misma columna.
	- Archivos modificados: `web/index.html`, `web/estilos.css`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se agrega el contenedor `portal-intro-copy` para apilar verticalmente el encabezado del home, manteniendo `Sistema de Facturación Electrónica` en una sola fila y moviendo `Toma el control de tu negocio con Powerful Control System` justo debajo, centrado en el mismo bloque visual.
	- Verificacion: revision estatica de estructura HTML/CSS confirmando el nuevo contenedor y la regla `white-space: nowrap` aplicada al título.

- Login administrativo Google: correccion para VPS y local + recordar cuenta estable.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/utils/utils.go`, `backend/handlers/auth_users_carritos_test.go`, `web/login.html`, `web/menu.js`, `web/js/login.js`, `web/index.html`, `web/estilos.css`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se corrige el flujo OAuth para adaptar `redirect_uri` al host real de la solicitud y forzar `https` en dominio publico (`powerfulcontrolsystem.com`), se habilitan rutas publicas que bloqueaban el login (`/js/login.js` y `/api/public/pagina_principal`), se evita consulta a `/me` sin sesion para eliminar ruido `401` en F12 y se completa la experiencia de `recordar cuenta`; adicionalmente se actualiza el encabezado del home a `Sistema de Facturación Electrónica` con subtitulo operativo.
	- Verificacion: `go test ./handlers -run "TestHandleGoogleLogin|TestAuthMiddlewareAllowsPublicLoginAssetsAndHomeCardsAPI" -v -count=1` en verde; en VPS `GET /js/login.js` y `GET /api/public/pagina_principal` responden `200`; `GET /auth/google/login` emite `redirect_uri=https://powerfulcontrolsystem.com/auth/google/callback`; `google.redirect_url` en BD super quedó en HTTPS.

- Inicio local: diagnostico robusto para tunel SSH de PostgreSQL en VPS.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se mejora `Ensure-VpsPostgresTunnel` para esperar el listener con reintentos (hasta ~8s), capturar `stdout/stderr` de `plink` en `backend/tmp/plink_tunnel_<puerto>.*.log` y reportar causa detallada cuando el tunel no abre el puerto local; adicionalmente se corrige el argumento `-i` de `plink` para rutas de llave SSH con espacios (comillas explicitas), evitando el fallo `Host does not exist`.
	- Verificacion: validacion de parseo PowerShell en verde con `[System.Management.Automation.Language.Parser]::ParseFile("scripts/iniciar_servidor.ps1", ...)` y ejecucion real `. "D:\powerfulcontrolsystem\scripts\iniciar_servidor.ps1" -Background` completando arranque con tunel activo y backend en `:8080`.

- Checkout de licencias: cierre operativo de Epayco.
	- Archivos modificados: `backend/handlers/payments_handlers.go`, `backend/main.go`, `web/pagar_licencia.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se completa la implementacion de Epayco para licencias con `POST /epayco/create_transaction`, `GET /epayco/transaction_status` y `POST/GET /epayco/webhook`; se corrige la configuracion super de Epayco para aceptar credenciales reales sin validacion numerica de `cust_id`; y el frontend abre `checkout_url` de Epayco en una nueva pestaña manteniendo polling de estado y activacion automatica de licencia al aprobar.
	- Verificacion: `go test ./ -run "^$" -count=1`, `go test ./handlers -run "^$" -count=1`, `go test ./ ./auth ./db ./handlers ./metrics ./utils -run "^$" -count=1` en verde.

- Chat y tareas: nuevo agente de citas con calendario grande y recordatorios previos.
	- Archivos modificados: `backend/db/chat_tareas.go`, `backend/handlers/chat_tareas.go`, `backend/handlers/chat_tareas_test.go`, `backend/main.go`, `web/administrar_empresa/chat_y_tareas.html`, `web/estilos.css`, `web/super/pagina_principal.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se agrega agenda de citas empresarial en el modulo de chat/tareas (`/api/empresa/chat_tareas/citas`) con calendario mensual de gran formato, programacion/edicion de reuniones, visibilidad compartida por `empresa_id` y banner de recordatorios previos; adicionalmente se incluye un boton inferior de guardado en `web/super/pagina_principal.html`.
	- Verificacion: `$env:DB_DIALECT='sqlite'; go test ./handlers -run ChatTareas -count=1` y `$env:DB_DIALECT='sqlite'; go test ./db -run ChatTareas -count=1`.

- UI administrativa: eliminacion de barra superior de titulo/acciones en todas las paginas de layout.
	- Archivos modificados: `web/super_administrador.html`, `web/administrar_empresa.html`, `web/administrar_empresa/finanzas_menu.html`, `web/administrar_empresa/facturacion_electronica_menu.html`, `web/administrar_empresa/administrar_productos_menu.html`, `web/administrar_empresa/configuracion_menu.html`, `web/administrar_empresa/reportes_menu.html`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se retira por completo el bloque visual `admin-toolbar page-header` del panel super y de los menus administrativos para eliminar la barra superior de la derecha en todas las vistas del layout.
	- Verificacion: busqueda `class="admin-toolbar"` en `web/**/*.html` sin resultados.

- Inicio local: correccion de deteccion de procesos en puerto 8080 bajo StrictMode.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se normaliza la coleccion de PIDs detectados en el paso de liberacion de puerto para evitar `No se encuentra la propiedad 'Count'` cuando solo existe un proceso escuchando.
	- Verificacion: ejecucion real `. 'D:\powerfulcontrolsystem\scripts\iniciar_servidor.ps1'` completando `3/8 Liberando puerto 8080` sin excepcion y arranque exitoso del backend en `:8080`.

- Inicio local: correccion de carga DSN PostgreSQL y tunel DB opcional en script de arranque.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivo local de entorno actualizado (no versionado): `backend/.env.local`.
	- Descripcion: el script ahora carga `DB_DIALECT`, `DB_EMPRESAS_DSN` y `DB_SUPERADMIN_DSN` desde `.env.local/.env` antes de validar prerequisitos; se anade soporte opcional para tunel SSH a PostgreSQL en VPS (`DB_VPS_TUNNEL_*`) con validacion temprana del puerto de tunel y ajuste temporal de DSN al listener local.
	- Verificacion: `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\iniciar_servidor.ps1 -Background` en verde y `curl -I http://127.0.0.1:8080` con `HTTP/1.1 200 OK`.

- Venta publica por subdominio empresarial automatizado.
	- Archivos modificados: `backend/main.go`, `backend/handlers/venta_publica.go`, `backend/handlers/venta_publica_test.go`, `web/venta_publica.html`, `web/administrar_empresa/venta_publica.html`, `documentos/deploy_nginx_reverse_proxy_vps.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se habilita resolucion de `empresa_slug` por subdominio (`{slug}.powerfulcontrolsystem.com`) en backend y frontend de venta publica, con soporte de apertura automatica de tienda desde la raiz del subdominio.
	- Verificacion: `go test ./handlers -run "VentaPublica|ResolveVentaPublicaSlugFromHost" -count=1` en verde.
	- Evidencia VPS: Nginx actualizado con bloque wildcard y captura de slug por host; validado `GET /` en host de subdominio con `302` a `/venta_publica.html?empresa_slug=<slug>` y `GET /venta_publica.html?empresa_slug=<slug>` con `200 OK`; queda pendiente registrar wildcard DNS `*.powerfulcontrolsystem.com` (resolucion publica actual `NXDOMAIN`).

## 2026-04-14
- Guia operativa de dominio con Nginx reverse proxy en VPS.
	- Archivo creado: `documentos/deploy_nginx_reverse_proxy_vps.md`.
	- Archivos modificados: `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `documentos/descripcion_del_proyecto`, `CHANGELOG.md`.
	- Descripcion: se documenta el procedimiento para publicar `powerfulcontrolsystem.com` y `www.powerfulcontrolsystem.com` con Nginx en Ubuntu VPS, manteniendo el backend en `127.0.0.1:8080`, con validaciones de servicio/UFW y opcion de HTTPS con Certbot.
	- Verificacion: guia con comandos en orden, listos para copia/pegado en consola remota.

## 2026-04-14
- Modulo de impresoras operativas por empresa.
	- Archivos creados: `backend/db/empresa_impresoras.go`, `backend/db/empresa_impresoras_test.go`, `backend/handlers/empresa_impresoras.go`.
	- Archivos modificados: `backend/main.go`, `web/administrar_empresa/configuracion.html`, `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/finanzas.html`, `web/administrar_empresa/reportes.html`, `documentos/estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se añade gestion de impresoras por `empresa_id` (predeterminada, estado activo/inactivo, asignacion por funcionalidad y por producto) y resolucion operativa de impresora con prioridad `producto -> funcionalidad -> predeterminada`, integrada en configuracion y en flujos de impresion de carrito/finanzas/reportes.
	- Verificacion: `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1` en verde.

- Super administrador: nuevo panel de administracion de base de datos PostgreSQL.
	- Archivos creados: `backend/handlers/postgres_performance.go`, `backend/handlers/postgres_performance_test.go`, `web/super/administrar_base_de_datos.html`.
	- Archivos modificados: `backend/main.go`, `web/super_administrador.html`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se agrega un tablero profesional para monitoreo de PostgreSQL (salud del cluster, metricas por base, consultas activas prolongadas, `pg_stat_bgwriter` y recomendaciones automaticas), con endpoint protegido `/super/api/postgres/performance`.
	- Verificacion: `go test ./handlers -run "PostgresPerformance" -count=1` y `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1` en verde.

- Migracion cerrada a PostgreSQL-only y retiro de SQLite operativo.
	- Archivos modificados: `backend/main.go`, `backend/db/sql_compat.go`, `scripts/iniciar_servidor.ps1`, `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `scripts/actualizar_repositorio.ps1`, `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/estructura_bd.md`, `documentos/descripcion_de_las_bases_De_datos`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Archivos eliminados: `backend/db/empresas.db`, `backend/db/superadministrador.db`.
	- Descripcion: el backend queda forzado a runtime PostgreSQL-only, sin fallback SQLite en arranque; se limpian los `.db` legados del repositorio y se alinea la operacion local/remota a DSN PostgreSQL obligatorios.

- Estandarizacion documental ERP multiempresa.
	- Archivos creados: `documentos/erp_multiempresa/README.md`, `documentos/erp_multiempresa/01_alcance_erp_multiempresa.md`, `documentos/erp_multiempresa/02_diseno_tecnico_erp_multiempresa.md`, `documentos/erp_multiempresa/03_especificaciones_funcionales_erp_multiempresa.md`, `documentos/erp_multiempresa/04_guia_implementacion_erp_multiempresa.md`.
	- Archivos modificados: `documentos/README.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`, `CHANGELOG.md`.
	- Descripcion: se consolida un paquete ERP estandar listo para revision, con claridad de alcance, arquitectura, requisitos funcionales, reglas de negocio, integraciones y ruta de implementacion por fases.

- Documentacion: reorganizacion profesional, consolidacion de fuentes canonicas y limpieza de artefactos no usados.
	- Archivos modificados: `documentos/README.md`, `documentos/descripcion_del_proyecto`, `documentos/estructura_del_codigo`, `.gitignore`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- Archivos depurados: `documentos/historial_de_cambios_addendum_2026-04-04.md`, `backend/tmp/server.exe`, `backend/server.err`, `backend/server.log`, `backend/server.run.log`, `backend/logs/*.log`, `logs/test_runs/*.log`, `scripts/logs/*.log`, `tmp/doc_audit_report.txt`, `tmp/doc_hash_duplicates.txt`.
	- Descripcion: se centraliza la documentacion en un indice canonico, se evita duplicidad entre documentos estructurales y se eliminan archivos temporales/runtime que no deben versionarse.
	- Verificacion: carpetas de logs temporales quedan limpias y se mantiene solo estado runtime necesario (`backend/logs/server_runtime_state.json`).

- OAuth Google VPS: validacion final de infraestructura HTTPS y diagnostico concluyente de `redirect_uri_mismatch`.
	- Archivos modificados: `CHANGELOG.md`, `documentos/historial_de_cambios`.
	- Descripcion: se verifica en VPS que el backend emite callback seguro `https://2.24.197.58.nip.io/auth/google/callback` y que el proxy TLS (Caddy) esta operativo en `:443`; Google sigue rechazando el flujo por URI no autorizada en el cliente OAuth.
	- Verificacion: prueba E2E desde VPS confirma mismatch para la URI HTTPS publica y matriz de prueba muestra aceptacion solo de `http://localhost:8080/auth/google/callback`.
	- Pendiente externo: agregar la URI exacta del VPS en Google Cloud Console y repetir prueba de login.

- Inicio local: corrección de detección de puerto 8080 para evitar falso bloqueo por PID 0.
	- Archivos modificados: `scripts/iniciar_servidor.ps1`.
	- Descripción: se reemplaza la detección basada en `netstat | findstr ":8080"` por una resolución de listeners locales reales (primero `Get-NetTCPConnection`, con fallback parseado de `netstat` en estado `LISTENING`). Se filtran PID inválidos/no gestionables (`<= 0`) y se evita abortar cuando aparece `System Idle Process` sin listener real del backend.
	- Verificación: ejecución local `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\iniciar_servidor.ps1 -Background` completa en verde; paso 3 muestra `No hay procesos escuchando en el puerto 8080` y el servidor inicia correctamente.

- OAuth Google VPS: prioridad de entorno sobre DB + soporte de `GOOGLE_REDIRECT_URL` en despliegue.
	- Archivos modificados: `backend/main.go`, `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`.
	- Descripción: se ajusta la carga OAuth para que `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` y `GOOGLE_REDIRECT_URL` del entorno tengan prioridad sobre valores almacenados en tabla `configuraciones` (la DB solo completa faltantes). Se añade además propagación de `GOOGLE_REDIRECT_URL` en bootstrap remoto de scripts de sincronización.
	- Verificación: `go test ./handlers -run "TestHandleGoogleLoginRedirect" -count=1` y `go test ./ -count=1` en verde. Diagnóstico en VPS confirma que el bloqueo actual es de política OAuth en Google (`secure-response-handling` / `redirect_uri_mismatch`) y no de base de datos.

- OAuth Google: corrección de callback para evitar `localhost` en entorno VPS.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`, `backend/main.go`.
	- Descripción: se implementa resolución dinámica del `redirect_uri` por host/protocolo de la solicitud y una regla de reescritura segura cuando la configuración existente apunta a loopback (`localhost/127.0.0.1`) pero el acceso real es público (VPS). El callback reutiliza la URL efectiva mediante cookie técnica de corta duración para mantener consistencia en intercambio de token.
	- Verificación: despliegue real a VPS con `DEPLOY_OK:pid=53618 port=8080`; validación HTTP de `/auth/google/login` devuelve `redirect_uri=http://2.24.197.58:8080/auth/google/callback`.

- Sync VPS: guard estricto de DSN para PostgreSQL y recuperación de despliegue estable.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/diagramas/estructura_del_codigo.md`, `CHANGELOG.md`, `documentos/historial_de_cambios`, `documentos/descripcion_de_archivos`.
	- Descripción: el bootstrap remoto ahora conserva valores DB existentes, valida el modo efectivo y bloquea el despliegue con `BOOTSTRAP_ERROR:POSTGRES_MISSING_DSN` cuando `postgres` no tiene ambos DSN; además usa el último valor de cada clave (`tail -n1`) y evita llegar a `DEPLOY_ERROR:process_not_running` por arranque inválido. En paralelo se restableció configuración DSN operativa en VPS para retomar despliegues en modo PostgreSQL.
	- Verificación: ejecución real `./scripts/sync_to_vps.ps1 -SkipBuild -RetryCount 1` primero falla en bootstrap con mensaje explícito de DSN faltantes, luego (tras restablecer DSN en VPS) finaliza con `DEPLOY_OK:pid=... port=8080` y `GET /` = `200`.

- VPS web root: corrección de resolución de estáticos para index/login.
	- Archivos modificados: `backend/main.go`, `scripts/sync_to_vps.ps1`, `CHANGELOG.md`, `documentos/historial_de_cambios`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`.
	- Descripción: se ajusta `resolveWebDir()` para priorizar correctamente `.../web` cuando el binario corre desde `backend/bin`, evitando que el servidor sirva `backend/web/uploads/` como raíz. Se redepliega en VPS y se valida apertura automática de la URL pública.
	- Verificación: `GET /` = `200` con HTML de portal, `GET /index.html` = `200`, `GET /login.html` = `200`, proceso remoto activo en `:8080` y runtime PostgreSQL operativo.

- Sync VPS: hardening para preservar DSN remotos y apertura automática de URL pública.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`.
	- Descripción: se excluye `backend/.env.local` de la sincronización para evitar sobrescribir secretos/DSN del VPS, se robustece el healthcheck de redeploy (detecta proceso caído y valida respuesta HTTP distinta de `000`) y se añade apertura automática de `http://<host>:<puerto>/` al finalizar despliegues exitosos.
	- Verificación: ejecución real `./scripts/sync_to_vps.ps1 -RemoteHost 2.24.197.58 -RemoteUser root -RemotePath /root/powerfulcontrolsystem -DbDialect postgres -DbEmpresasDsn ... -DbSuperadminDsn ...` con `DEPLOY_OK:pid=... port=8080`, `GET / => 200` y backend en modo PostgreSQL con DSN activos.

- Migración PostgreSQL (fase 4): estabilización de salida operativa en contabilidad y runtime VPS.
	- Archivos modificados: `backend/db/eventos_contables.go`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `Pendiente Notas`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- Descripción: se corrige el flujo del worker de asientos/eventos para PostgreSQL usando wrappers SQL portables y retorno de `id` compatible, eliminando el error `syntax error at or near "ORDER"` en runtime. Se restaura además el entorno VPS con DSN PostgreSQL válidos en `backend/.env.local` y se valida arranque estable.
	- Verificación: `go test ./ ./auth ./db ./handlers ./metrics ./utils` en verde; validación remota en VPS con proceso activo, sin errores recientes de `asientos_worker` y healthcheck `HTTP=200`.

- Migración PostgreSQL (fase 3): cierre documental del plan y sincronización de gobernanza por módulos.
	- Archivos modificados: `Pendiente Notas`, `documentos/descripcion_de_modulos`, `documentos/matriz_roles_permisos_pos_multiempresa.md`, `documentos/historial_de_cambios`.
	- Descripción: se marca Fase 3 como completada en el plan operativo, se agrega evidencia técnica de conmutación a PostgreSQL y se alinea la documentación de módulos/permisos sin cambios de privilegios en la matriz CRUD/A.
	- Verificación: se mantiene evidencia de pruebas del bloque core en verde (`go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1`).

- Migración PostgreSQL (fase 3): conmutación de runtime backend a motor PostgreSQL en VPS.
	- Archivos modificados: `backend/main.go`, `backend/go.mod`, `backend/go.sum`, `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`, `documentos/diagramas/estructura_del_codigo.md`.
	- Descripción: el backend ahora selecciona motor por entorno (`DB_DIALECT`), abre conexiones con `pgx` usando `DB_EMPRESAS_DSN` y `DB_SUPERADMIN_DSN`, y omite el bootstrap SQLite cuando el runtime es PostgreSQL. Los scripts de sincronización ahora propagan y verifican estas variables en `backend/.env.local` del VPS durante bootstrap remoto.
	- Verificación: `go test ./ ./auth ./db ./handlers ./metrics ./utils -count=1` en verde.

- Migración PostgreSQL (fase 3): compatibilidad ampliada en núcleo `backend/db`.
	- Archivos modificados: `backend/db/sql_compat.go`, `backend/db/empresa_scope.go`, `backend/db/productos.go`, `backend/db/db.go`.
	- Descripción: se amplía la capa de compatibilidad SQLite/PostgreSQL con wrappers `query/exec` portables, inserciones con `RETURNING id` para PostgreSQL, detección de tablas por `information_schema` y ajuste de `ensureColumnIfMissing` por dialecto con normalización de defaults de fecha. Además, se migra el bloque core de `db.go` (licencias, tipos de empresa, empresas, Wompi, asesores, configuraciones y métricas) para usar placeholders/fechas compatibles con ambos motores.
	- Verificación: `go test ./db -run "Session|Admin|User|Licencia|TipoEmpresa|Empresa|Config|Metric|Wompi|Asesor" -count=1` y `go test ./handlers -run "TestHandleGoogleLoginRedirectIncludesLoginHint|TestE2E_AcceptContractCreatesSession" -count=1` en verde.

- Sync VPS: selección automática de clave de identidad al no pasar `-IdentityFile`.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/README_sync.md`.
	- Descripción: cuando no se especifica `-IdentityFile`, el script ahora prioriza la clave del proyecto `clave privada ssh.ppk` y, si no existe, usa `~/.ssh/id_rsa`. Además, mejora el mensaje de error cuando el VPS rechaza autenticación.
	- Verificación: ejecución real `./scripts/sync_to_vps.ps1 -SkipBuild -RemoteHost 2.24.197.58 -RemoteUser root -RemotePath /root/powerfulcontrolsystem -RetryCount 1` completada con `Sincronización completada por fallback sin WSL (PuTTY)`.

- Sync VPS: redeploy remoto automático de backend tras sincronización.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/sync_to_vps.sh`, `scripts/README_sync.md`.
	- Descripción: la sincronización ahora detiene el proceso viejo del backend en VPS, inicia la nueva versión del binario y valida salud HTTP en el puerto configurado (`SERVER_PORT`), evitando que quede corriendo una versión antigua.
	- Verificación: ejecución real `./scripts/sync_to_vps.ps1 -SkipBuild -RemoteHost 2.24.197.58 -RemoteUser root -RemotePath /root/powerfulcontrolsystem -RetryCount 1` con salida `DEPLOY_OK:pid=... port=8080`.

- Migración PostgreSQL (fase 3): avance inicial en autenticación y sesiones.
	- Archivos añadidos/modificados: `backend/db/sql_compat.go`, `backend/db/db.go`, `documentos/diagramas/estructura_del_codigo.md`.
	- Descripción: se incorpora capa de compatibilidad SQL SQLite/PostgreSQL (rebindeo de placeholders y expresiones de fecha) y se aplica a funciones críticas del flujo de autenticación/sesiones (`UpsertUser`, `UpsertAdministrador`, `CreateSession`, `RevokeSessionByToken`, `GetSessionByToken`, `GetAdminByEmail`).
	- Verificación: `go test ./db -run "Session|Admin|User|Licencia" -count=1` y `go test ./handlers -run "TestHandleGoogleLoginRedirectIncludesLoginHint|TestE2E_AcceptContractCreatesSession" -count=1` en verde.

## 2026-04-13
 - Reparación de login de usuario empresarial: permitir entrada manual de `empresa_id` y persistencia de contexto.
	- Archivos modificados:
		- web/login_usuario.html
		- web/js/login_usuario.js
	- Descripción: se agrega un campo `Empresa ID` en la página de login de usuario de empresa para aceptar el parámetro cuando no viene en la URL. La lógica JS persiste `empresa_id` en session/local storage, asegura que `redirect_url` incluya `empresa_id` y mejora la funcionalidad de "recordar usuario" por empresa.
	- Verificación: validación de sintaxis JS sin errores y flujo de login manual verificado localmente.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/README_sync.md`.
	- Descripción: en fallback sin WSL, el script ahora selecciona transporte por tipo de clave: `ssh.exe` + `scp.exe` para claves OpenSSH (ej. `id_rsa`) y `plink.exe` + `pscp.exe` para `.ppk`. Con esto se evita el error `Unable to use key file ... OpenSSH SSH-2 private key (new format)` al usar la identidad por defecto.
	- Verificación: `.\scripts\sync_to_vps.ps1 -SkipBuild -PreviewOnly -IdentityFile "$env:USERPROFILE\.ssh\id_rsa"` muestra `Fallback sin WSL (OpenSSH)` y comandos con `ssh.exe`/`scp.exe`.

- Migración de datos a PostgreSQL en VPS: instalación, ejecución por etapas y validación inicial.
	- Archivos modificados: `Pendiente Notas`, `documentos/regla_agente_go.md`, `copilot-instructions.md`, `documentos/descripcion_de_las_bases_De_datos`, `documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- Descripción: se instala PostgreSQL en VPS por SSH, se crean las bases `pcs_superadministrador` y `pcs_empresas`, y se inicia la migración desde SQLite con `pgloader` en dos etapas (superadministrador y empresas), validando consistencia por conteo de tablas en cada base. Se formaliza además la regla operativa: base productiva en VPS con PostgreSQL y SQLite local como legado de migración/contingencia.
	- Verificación: `VALIDACION_SUPER_OK` y `VALIDACION_EMPRESAS_OK` tras comparación SQLite vs PostgreSQL por tabla.

- Login administrativo: eliminación del mensaje visual de cuenta recordada y ajuste de OAuth.
	- Archivos modificados: `web/login.html`, `backend/handlers/auth_admin_handlers.go`, `backend/handlers/auth_users_carritos_test.go`.
	- Descripción: se elimina el texto en pantalla `Cuenta recordada ...` del login admin y se ajusta el parámetro OAuth `prompt` a `select_account` para evitar re-consentimiento de Google en cada inicio.
	- Verificación: `go test ./handlers -run TestHandleGoogleLoginRedirectIncludesLoginHint -count=1` en verde.

- Login administrativo: corrección de "Recordar cuenta" para evitar sesión parcial.
	- Archivos modificados: `web/js/login.js`, `web/menu.js`.
	- Descripción: se corrige el flujo para que cerrar sesión no elimine la preferencia cuando `rememberAccount=1`, se mantiene el correo recordado hasta que el usuario pulse "Olvidar" y se agrega sincronización de `rememberedEmail` desde `/me` cuando existe sesión activa.
	- Verificación: revisión de errores en frontend sin incidencias (`get_errors` en ambos archivos).

- Inicio local: hardening de scripts/iniciar_servidor para evitar caídas del host de PowerShell/VS Code.
	- Archivo modificado: `scripts/iniciar_servidor.ps1`.
	- Descripción: se refuerza la liberación de puerto 8080 para terminar únicamente procesos del backend (`server.exe`, `pos-backend`, `go run` del proyecto) y no procesos ajenos. Cuando el puerto está ocupado por un proceso no gestionado, el script ahora informa el PID/nombre y aborta con mensaje claro en lugar de forzar `taskkill` indiscriminado. También se elimina el `Clear-Host` inicial para evitar efectos colaterales en consolas integradas.
	- Verificación: ejecución local `./scripts/iniciar_servidor.ps1 -Background` con `SCRIPT_EXIT=0` y comprobación HTTP local `HTTP_STATUS=200`.

- Unificación de bases SQLite: solo dos archivos canónicos del sistema.
	- Archivos modificados: `backend/main.go`, `documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`, `documentos/descripcion_de_archivos`, `documentos/historial_de_cambios`.
	- Descripción: se normaliza la resolución de rutas en runtime para que el backend use por defecto `backend/db/empresas.db` y `backend/db/superadministrador.db` aunque se ejecute desde otro directorio. Se depuran copias operativas duplicadas en raíz y en `backend/`, dejando únicamente dos archivos `.db` activos.
	- Verificación: inventario local posterior muestra exactamente dos DB (`backend/db/empresas.db` y `backend/db/superadministrador.db`) y pruebas backend en verde con `go test ./ ./auth ./db ./handlers ./metrics ./utils`.

- Sync VPS: bootstrap automático para servidor nuevo y diagnóstico de OAuth.
	- Archivos modificados: `scripts/sync_to_vps.ps1`, `scripts/README_sync.md`.
	- Descripción: se añade bootstrap post-sync en modo sin WSL para instalar dependencias base (`ca-certificates`, `curl`, `sqlite3`), asegurar `backend/.env.local` y reportar estado de variables críticas (`GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `SERVER_PORT`, `CONFIG_ENC_KEY`) con salida `BOOTSTRAP_WARN/BOOTSTRAP_OK`. Se incorporan parámetros opcionales `-GoogleClientId` y `-GoogleClientSecret`.
	- Verificación: ejecución real con `SYNC_EXIT=0` y diagnóstico remoto mostrando faltantes OAuth (`GOOGLE_CLIENT_ID/SECRET` vacíos).

- Instalador de clave pública en Windows: corrección de errores de ejecución.
	- Archivo modificado: `scripts/instalar_clave_publica_vps.ps1`.
	- Descripción: se corrige el flujo para evitar errores remotos tipo `invalid option namepefail` y se adapta a PowerShell 5.1 eliminando sintaxis no soportada (`??`). Ahora usa comando remoto en una sola línea, validación de formato de clave OpenSSH y reintentos por timeout.
	- Verificación: `scripts/instalar_clave_publica_vps.ps1 -PreviewOnly -RemoteHost 2.24.197.58 -User root -Port 22` en verde (exit 0).

- Despliegue VPS: instalación automatizada de clave pública PuTTYgen + robustecimiento de scripts de sincronización
	- Archivos añadidos/modificados:
		- scripts/instalar_clave_publica_vps.ps1 (nuevo: instala clave pública RFC4716 en `~/.ssh/authorized_keys` de VPS Linux)
		- scripts/sync_to_vps.sh (hardening para Linux: validaciones, eliminación de `eval`, chequeo remoto de SO)
		- scripts/sync_to_vps.ps1 (manejo de errores sin cerrar terminal de VS Code, build Linux local previo y fallback PuTTY sin WSL con empaquetado tar)
		- scripts/README_sync.md (guía de ejecución en un comando)
		- web/login.html y web/js/login.js (completa UX de "Recordar cuenta" para login admin)
	- Descripción: se habilita un flujo operativo de un solo comando para preparar acceso por clave pública al VPS y se corrige la causa de cierres de terminal por `exit` en script PowerShell. `sync_to_vps.ps1` ahora compila en local un binario Linux (`backend/bin/server_linux_amd64`) antes de sincronizar y, sin Ubuntu/WSL, opera empaquetando el proyecto en `.tar`, subiéndolo por `pscp.exe` y extrayéndolo en VPS por `plink.exe`, con trazas detalladas y exclusión de archivos sensibles/locales (`*.ppk`, `*.pem`, `*.key`, DB, logs, temporales); además aplica `chmod +x` al binario remoto configurado. Se añadió manejo de `Connection timed out` con prechequeo TCP y reintentos automáticos configurables (`-RetryCount`) por etapa de conexión/subida/extracción.
	- Verificación: `scripts/instalar_clave_publica_vps.ps1 -PreviewOnly -RemoteHost 2.24.197.58` ejecuta correctamente; `scripts/sync_to_vps.ps1 -BuildOnly`, `-DryRun` y ejecución real con `-IdentityFile "D:\powerfulcontrolsystem\clave privada ssh.ppk"` completan con `exit code 0`; en VPS el artefacto quedó como ELF Linux en `/root/powerfulcontrolsystem/backend/bin/server_linux_amd64`.

- Módulo Vendedores / Asesores comerciales: integración de código de descuento y registro de asesor/vendedor en pagos
	- Archivos añadidos/modificados:
		- backend/handlers/payments_handlers.go (extiende payload y persistencia de `pagos_wompi` con `discount_code` y `asesor_id`/`vendedor_id`)
		- backend/db/db.go (helpers para `asesores`, `asesor_comercial` y `asesor_comisiones`, y claves de configuración `vendedor.*`)
		- backend/handlers/vendedores_handlers.go (nuevo: CRUD de asesores / vendedores)
		- backend/handlers/vendedor_config_handlers.go (nuevo: GET/PUT /super/api/vendedor_config)
		- backend/main.go (migraciones: tablas `asesores`, `asesor_comercial`, `asesor_comisiones`; registro de rutas `/super/api/vendedores`, `/super/api/asesor_comercial`, `/super/api/vendedor_config`)
		- backend/tools/insert_asesor.go, backend/tools/insert_plan.go, backend/tools/insert_licencia.go, backend/tools/create_session.go, backend/tools/query_pagos_comisiones.go (herramientas para pruebas locales)
		- web/pagar_licencia.html (nuevo campo `discount_code` y `asesor_id`/`vendedor_id` en el formulario de pago)
		- web/super/activar_asesor.html, web/super/asesor_comercial.html, web/super/vendedor_config_avanzado.html (UI super-administrador para activar vendedores, configurar planes y ajustes globales)
		- documentos/estructura_bd.md (documenta las nuevas tablas y columnas de pagos/comisiones)
		- documentos/descripcion_de_archivos (registro de los nuevos archivos del módulo)
	- Descripción: Se añade soporte opcional para incluir un código de descuento y una referencia al asesor/vendedor en el pago de licencias. Se introduce la entidad de `asesores` (vendedores), planes comerciales (`asesor_comercial`) y el registro de comisiones (`asesor_comisiones`) que crea una comisión inmediata y entradas programadas por meses de renovación según el plan.
	- Verificación: Prueba manual de activación sin pago (`/licencias/activar_sin_pago`) usando sesión administrativa de prueba; se confirmó la creación de una fila en `pagos_wompi` con `discount_code` y `asesor_id` y la creación de registros en `asesor_comisiones` (comisión inmediata + programadas). Tests automatizados pendientes.

- Estaciones: fix de persistencia de `estaciones_config` cuando el frontend no envía `estado`.
	- Archivos modificados: `backend/db/empresa_estacion_prefs.go`, `backend/db/empresa_estacion_prefs_test.go`, `backend/handlers/empresa_estacion_prefs_test.go`.
	- Descripcion: se normaliza `estado` vacio como `activo` en upsert/list/get de preferencias por estacion, evitando que las estaciones desaparezcan despues de guardarse.
	- Verificacion: pruebas en verde de estaciones, sensores, ventas y facturacion documental.

- Estaciones: correccion de flujo 10+, colores movidos a configuracion de estaciones y hardening sensor/carrito.
	- Archivos modificados: `web/administrar_empresa/configuracion_de_estaciones.html`, `web/administrar_empresa/configuracion.html`, `web/administrar_empresa/estaciones.html`, `backend/handlers/empresa_estacion_prefs_test.go`.
	- Descripcion: se consolida la gestion de colores de estado de carrito en la configuracion de estaciones, se fortalece el parseo de `estaciones_config` para tolerar payloads legacy anidados, se mejora la sincronizacion de carritos por estacion ante colisiones idempotentes y se valida el rango de estacion en configuracion de sensores.
	- Verificacion: pruebas dirigidas en verde para handlers y DB en estaciones/sensores/carritos/facturacion, incluyendo `TestEmpresaEstacionPrefsHandler_UpsertAndIsolationByEmpresa`.

- Reparación integral de acceso empresarial y estaciones.
	- Archivos modificados: `web/login_usuario.html`, `web/js/login_usuario.js`, `web/js/seleccionar_empresa.js`, `web/administrar_empresa/configuracion_de_estaciones.html`.
	- Descripción: se corrige la continuidad del flujo `login usuario empresa -> seleccionar empresa -> administrar empresa` con persistencia de `empresa_id` y opción de recordar correo. La página de configuración de estaciones se reconstruye y soporta generación/sincronización masiva de estaciones (incluyendo 10+) con manejo tolerante de conflictos idempotentes al cerrar/inactivar carritos.
	- Verificación: pruebas backend de paquetes principales en verde (`go test ./ ./auth ./db ./handlers ./metrics ./utils`).

## 2026-04-12
- Flujo final de login administrativo: cuenta Google correcta + aceptación única de contrato + reCAPTCHA real.
	- Archivos modificados: `backend/handlers/auth_admin_handlers.go`, `backend/handlers/accept_handlers.go`, `backend/handlers/e2e_login_acceptance_test.go`, `backend/handlers/auth_users_carritos_test.go`, `web/login.html`, `web/js/login.js`, `web/accept.html`, `web/menu.js`, `web/estilos.css`.
	- Descripción: se unificó el flujo en `login.html -> OAuth -> /accept.html -> /accept/complete -> panel`, usando `administradores.acepta_contrato` como fuente canonica de aceptación (sin depender de cookie global), validación server-side de reCAPTCHA y prompt OAuth `select_account consent` para evitar reutilización silenciosa de cuenta incorrecta.
	- Verificación: pruebas dirigidas en verde (`TestE2E_AcceptContractCreatesSession` y `TestHandleGoogleLoginRedirectIncludesLoginHint`).

- Módulo sensor de puertas (Raspberry Pi): backend, handlers, UI y tests.
	- Archivos agregados/modificados:
		- backend/db/sensor_puertas.go (nuevo módulo DB: dispositivos y heartbeats)
		- backend/handlers/sensor_puertas.go (handlers: endpoint público `action=heartbeat` y configuración protegida)
		- backend/db/sensor_puertas_test.go (pruebas unitarias DB)
		- backend/handlers/sensor_puertas_test.go (pruebas handlers: heartbeat y configuración)
		- web/administrar_empresa/configuracion_de_estaciones.html (UI: registrar device → estación)
		- web/administrar_empresa/estaciones.html (indicador visual sensor añadido)
		- web/estilos.css (estilos del indicador)
	- Descripción: Se implementó un módulo ligero para registrar dispositivos Raspberry Pi por empresa y estación, recibir heartbeats públicos y reflejar el estado (negro/verde) en las tarjetas de estaciones. Incluye pruebas unitarias para DB y handlers.
	- Verificación: `go test ./...` ejecutado y tests verdes.

## 2026-04-11
- Generador automático de códigos de descuento: formato moderno `PREFIJO-XXXX-XXXX` (`DSCT-AB12-CD34`).
	- Archivos modificados: `backend/db/codigos_descuento.go`, `web/administrar_empresa/codigos_de_descuento.html`.
	- Se mantiene índice único por `(empresa_id, codigo)` y se implementó reintentos en inserción para manejar colisiones raras.
	- Se actualizaron `documentos/estructura_bd.md`, `documentos/descripcion_de_archivos` y `documentos/historial_de_cambios`.
	- Pruebas unitarias de DB asociadas: todas en verde.

## 2026-04-09
- Gobernanza de agente: se oficializa flujo DIAN SaaS multiempresa en instrucciones del repositorio.
	- `copilot-instructions.md` incorpora regla oficial: software DIAN compartido (un `Software ID`/`Software PIN` para la plataforma) con credenciales tributarias obligatorias por empresa (`nit`, `token_emisor_ref`, `certificado_clave_ref`).
	- Se explicita trazabilidad por `empresa_id` en cada envio real y prohibicion de reutilizar token/firma entre empresas.
	- Trazabilidad sincronizada en `documentos/historial_de_cambios`.
- Facturacion electronica DIAN (Colombia): modo SaaS multiempresa con software compartido y credenciales por empresa.
	- `backend/db/modulos_faltantes.go` amplia `empresa_dian_configuracion` con `usar_software_compartido`, `software_id_compartido_ref`, `software_pin_compartido_ref` e indice `ix_dian_empresa_shared_mode`.
	- `backend/handlers/modulos_faltantes.go` agrega resolucion de software efectivo (`resolveDIANSoftwareCredentials`) con fallback global `DIAN_SHARED_SOFTWARE_ID/DIAN_SHARED_SOFTWARE_PIN`.
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
	- Se registra configuracion DIAN en `backend/db/empresas.db` para `empresa_id=6` usando los datos de `documentos/DATOS DIAN.mhtml` (Software ID/PIN, TestSetId, prefijo, resolucion y rango).
	- Se valida tecnicamente el modulo DIAN con `handlers.EmpresaDIANColombiaHandler` ejecutando:
		- `checklist` y `validar` sin faltantes (`ok=true`).
		- `generar_cufe_demo` y `generar_xml_demo` con resultados correctos para documento de prueba.
	- En esta iteracion no fue necesario crear credenciales API adicionales en configuracion avanzada de super administrador para las pruebas internas de habilitacion.

## 2026-04-08
- Panel `administrar_empresa`: menu izquierdo desacoplado del contenido derecho.
	- `web/administrar_empresa.html` activa shell dedicado (`admin-empresa-shell`).
	- `web/estilos.css` separa scroll de sidebar e iframe para que navegar/desplazar el menu no afecte movimiento ni visibilidad de la subpagina cargada.
	- Incluye ajuste responsive para mantener usabilidad en pantallas moviles.

- Configuracion avanzada (DeepSeek/Gmail/Wompi): cifrado obligatorio robustecido y corrección de guardado cuando faltaba `CONFIG_ENC_KEY`.
	- `backend/main.go` ahora carga `.env.local/.env`, asegura `CONFIG_ENC_KEY` (autogenera y persiste en desarrollo cuando no existe) y normaliza secretos legacy para dejarlos cifrados.
	- `backend/handlers/super_config_backup_handlers.go` fuerza cifrado en restore de secretos y rechaza restore plano sin clave de cifrado.
	- `scripts/iniciar_servidor.ps1` valida/carga/autogenera `CONFIG_ENC_KEY` antes del arranque para evitar errores `400` en `/super/api/config/ai`.
	- `web/super/configuracion_avanzada.html` muestra en DeepSeek solo fecha/hora de ultima actualización, sin exponer fragmentos de credencial.
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

- Arranque local y estáticos web: mejoras en `scripts/iniciar_servidor.ps1` y corrección de raíz `/`.
	- `scripts/iniciar_servidor.ps1` ahora muestra progreso por etapas (`1/8` a `8/8`), mensajes `[INFO]/[OK]/[AVISO]/[ERROR]` y salida explícita para `-Background` sin abrir navegador.
	- `backend/main.go` corrige la resolución de carpeta web para priorizar candidatos con `index.html`, evitando servir accidentalmente `backend/web` (solo `uploads/`).
	- `backend/main.go` agrega manejo de `/favicon.ico` con fallback a `web/img/punto_venta.png` para evitar 404 en consola.
	- `web/index.html` declara favicon explícito con `link rel="icon"`.
	- Validaciones: compilación de `backend/main.go` (`go test . -run "^$"`) y parseo de PowerShell de `scripts/iniciar_servidor.ps1` OK.

- Backups empresariales: nueva opción para eliminar información por fecha de corte.
	- `backend/handlers/backups_empresariales.go` agrega `action=depurar_fecha` en `/api/empresa/backups`, con validación de `fecha_corte` y filtros opcionales `include_tables`/`exclude_tables`.
	- `backend/db/backups_empresariales.go` incorpora `PurgeEmpresaDataByDateCorte` para eliminar registros por `empresa_id` con fecha <= corte (inclusive), con detalle de eliminaciones por tabla.
	- La depuración permite generar backup previo automático antes de ejecutar borrado para trazabilidad y recuperación.
	- `backend/handlers/empresa_permisos.go` clasifica esta acción como `permActionApprove` en módulo seguridad.
	- `web/administrar_empresa/backups.html` incorpora UI de depuración por fecha con confirmación explícita y resumen de resultados.
	- Se agregan pruebas: `TestEmpresaBackupsPurgeByDateCorte` (DB) y `TestEmpresaBackupsHandlerPurgeByDate` (handler).
	- Validaciones: pruebas de backups en verde y compilación dirigida de paquetes backend críticos OK.

- Chat y tareas: documentos/fotos entre usuarios de empresa y administrador.
	- `backend/handlers/chat_tareas.go` deriva autor desde sesion autenticada (usuario/admin), evita suplantacion de `autor_*` y auto-registra participantes emisores en conversaciones.
	- Al crear conversacion desde usuario, se agrega automaticamente el admin propietario de la empresa como participante para habilitar intercambio usuario-admin.
	- Se amplian extensiones permitidas de adjuntos en backend y UI: `doc/docx/xls/xlsx/ppt/pptx/rtf/odt/ods/odp` (ademas de imagen/audio/pdf/txt/csv/json).
	- `web/administrar_empresa/chat_y_tareas.html` ahora envía metadata de actor efectiva (`autor_tipo`, `autor_ref_id`, `autor_nombre`, `autor_email`) segun sesion.
	- Se agrega `backend/handlers/chat_tareas_test.go` con pruebas para actor usuario derivado, upload `.docx` y auto-participacion usuario/admin.
	- Validaciones: pruebas dirigidas de handlers chat/tareas y compilacion de paquetes backend criticos en verde.

- Chat y tareas: higiene de pruebas y limpieza de artefactos temporales.
	- `backend/handlers/chat_tareas_test.go` incorpora limpieza automatica (`t.Cleanup`) de uploads temporales por empresa para mantener el workspace limpio tras las pruebas.
	- Se retiran artefactos locales residuales de validacion (`.docx` y binarios `.test.exe`) para evitar ruido en el arbol de cambios.
	- Validaciones: `go test ./handlers -run "TestEmpresaChatTareas" -count=1` y compilacion dirigida de paquetes backend (`./auth ./db ./handlers ./metrics ./utils`) en verde.

- Configuración monetaria y numérica por empresa en panel de configuración.
	- `backend/db/empresa_configuracion_avanzada.go` amplía `empresa_configuracion_avanzada` con `moneda_codigo`, `sistema_numerico`, `usar_decimales` y `cantidad_decimales`.
	- `web/administrar_empresa/configuracion.html` agrega tarjeta para configurar moneda operativa, sistema numérico y precisión decimal por empresa.
	- `backend/db/carritos_compras.go` aplica la moneda configurada por empresa como fallback al crear carritos sin moneda explícita.
	- `backend/main.go` registra la migración `2026-04-08-030-configuracion-monetaria-numerica`.
	- Validaciones: compilación de `db`, `handlers` y `main` en backend OK.

- Configuración IA migrada de Gemini a DeepSeek en super administrador y chat empresarial corregido.
	- `web/super/configuracion_avanzada.html` ahora gestiona credencial `deepseek:deepseek-chat` y corrige flujo de guardado de credenciales IA.
	- `backend/handlers/ai_credentials_catalog.go` registra `DEEPSEEK_API_KEY` como credencial IA activa en panel super.
	- `backend/handlers/chat_con_inteligencia_artificial_controller.go` usa DeepSeek como proveedor del chat IA por empresa.
	- `web/administrar_empresa/chat_con_inteligencia_artificial.html` actualiza etiquetas/mensajes para modelo IA genérico (sin acoplamiento a Gemini).
	- Validaciones: compilación de `handlers` y `main` en backend OK.

- Gobernanza documental reforzada para Agente Go y limpieza de documentos obsoletos.
	- Se actualiza `copilot-instructions.md` con regla obligatoria: si un modulo se crea o modifica, deben actualizarse `documentos/descripcion_de_modulos` y `documentos/matriz_roles_permisos_pos_multiempresa.md` en la misma iteracion.
	- Se refuerza la regla de sincronizacion de documentacion tecnica relacionada y trazabilidad en `documentos/descripcion_de_archivos` y `documentos/historial_de_cambios`.
	- Se elimina `documentos/modulos del proyecto.md` por duplicidad frente al documento canonico `documentos/descripcion_de_modulos`.
	- Se actualizan `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos` y `documentos/matriz_roles_permisos_pos_multiempresa.md` con la politica nueva.

- Cierre de pasos operativos 1, 2 y 3 solicitados en pendientes.
	- Paso 1: revisión/ajuste de accesos directos de módulos.
		- validación de consistencia de enlaces del panel empresa.
		- se agrega panel de accesos directos dinámico en `web/administrar_empresa/inicio.html` con visibilidad por permisos/licencia.
	- Paso 2: notas de voz en chat y tareas.
		- backend: `chat_tareas` incorpora campos `nota_voz_*` y endpoint `POST /api/empresa/chat_tareas/tareas/nota_voz`.
		- frontend: `chat_y_tareas.html` incorpora grabación con MediaRecorder para mensajes/tareas, envío y reproducción de audio.
	- Paso 3: super rol/permisos por licencia.
		- `licencias` incorpora `modulos_habilitados` y `super_rol_habilitado`.
		- middleware de permisos aplica restricciones por licencia y rol efectivo por empresa.
		- UI super de licencias permite configurar módulos habilitados y super rol por plan.
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
	- Se actualiza `Pendiente Notas` para añadir ejecucion dirigida de pruebas de `venta_publica` en handlers.
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
		- se incorpora resolucion efectiva por contexto (`rol`, `canal_venta`, `sucursal_id`, `turno`) y funciones de snapshot/listado/aplicacion de rollback.
		- `backend/db/configuracion_operativa_test.go` agrega `TestEmpresaConfiguracionOperativaPoliticaContextoYRollback`.
	- Backend `handlers`:
		- `backend/handlers/configuracion_operativa.go` amplía acciones HTTP con `action=politica`, `action=simular`, `action=historial` y `action=rollback`.
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
		- Se corrige robustez en listado de ejecuciones para manejar campos `NULL` de SQLite (`error_detalle`, `programacion_id`, metadatos opcionales) sin error `500`.
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
	- Se amplía `documentos/release_checklist.md` con checklist estandar "listo para produccion" por modulo (seguridad, rendimiento, trazabilidad, exportacion y pruebas).
	- Se amplía `documentos/punto_13_validacion_integral_resultado.md` con evidencia complementaria de cobertura y UAT por rol.
	- Se actualiza `Pendiente Notas` para marcar completados los 3 pendientes transversales.
	- Validaciones ejecutadas:
		- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\validar_punto_13.ps1` (OK).
		- `go test ./auth ./db ./handlers ./metrics ./utils -cover -count=1` (db `51.4%`, handlers `50.4%`).
		- `go test ./handlers -run "Test(SuperEndpointsPermisosPorRol|EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaPermisosContextoHandlerIncluyeMatrizRoles|EmpresaCarritosCompraBloqueaMetodoPagoSegunRol|EmpresaCarritosCompraRespetaBloqueoPropinaYComisionPorRol|EmpresaConfiguracionOperativaHandlerConfigAndRole|EmpresaDocumentosGestionHandlerVersionadoYControlAcceso)$" -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 27 (Ventas simples por estacion):
	- Se amplía `backend/db/carritos_compras.go` para:
		- agregar tabla `empresa_ventas_estacion_metricas` y funciones de registro/resumen de rendimiento por estacion.
		- calcular duracion de atencion por venta y resolver identidad de estacion desde carrito (`referencia_externa`/`codigo`).
	- Se amplía `backend/handlers/carritos_compras.go` para:
		- exponer `GET action=metricas_estacion` en `/api/empresa/carritos_compra`.
		- registrar metricas en `pagar_estacion`, `anular_cierre_parcial` y `recuperar_interrumpido`.
	- Se actualiza frontend de ventas simples:
		- `web/administrar_empresa/ventas_simple.html` incorpora panel de sincronizacion offline, metricas de estacion y correccion rapida post-cobro.
		- `web/js/ventas_simple.js` (nuevo) implementa cola offline por estacion con checksum SHA-256 y sincronizacion segura al reconectar.
		- `web/estilos.css` agrega estilos de estado de sincronizacion (`en linea`, `offline`, `sincronizando`).
	- Se amplía `backend/handlers/auth_users_carritos_test.go` con `TestEmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones|TestEmpresaCarritosCompraAndItemsFlow" -count=1` (OK).
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).

## 2026-04-07
- Cierre tecnico del modulo 26 (Carritos de compra e items):
	- Se amplía `backend/db/carritos_compras.go` para:
		- agregar reintentos transaccionales en operaciones de items frente a bloqueos SQLite (`database is locked/busy`) para fortalecer concurrencia multiestacion.
		- incorporar `RecoverInterruptedCarritoSession` para recuperar carritos interrumpidos sin perdida de items.
		- incorporar `CancelCarritoPartialClosure` para anulacion parcial de cierre en ventas pagadas con validacion estricta de monto.
	- Se amplía `backend/handlers/carritos_compras.go` para:
		- exponer `PUT action=recuperar_interrumpido` con trazabilidad en eventos contables y auditoria empresarial.
		- exponer `PUT action=anular_cierre_parcial` con validacion de negocio y auditoria por `empresa_id` y carrito.
	- Se ajusta `web/administrar_empresa/carrito_de_compras.html` para recuperar sesiones interrumpidas sin reset de items y reservar `reset_items=1` solo para sesiones ya pagadas.
	- Se amplía cobertura en:
		- `backend/db/carritos_inventario_test.go` (concurrencia de producto, recuperacion interrumpida, anulacion parcial).
		- `backend/handlers/auth_users_carritos_test.go` (recuperacion con auditoria, reglas de pago mixto y anulacion parcial de cierre).
	- Validaciones ejecutadas:
		- `runTests` sobre `backend/db/carritos_inventario_test.go` y `backend/handlers/auth_users_carritos_test.go` -> 36 pruebas OK, 0 fallidas.
		- `go test ./... -run "^$" -count=1` (compilacion global backend OK).

## 2026-04-07
- Cierre tecnico del modulo 25 (Panel ERP extendido):
	- Se amplía `web/js/modulos_erp_extendido.js` para incorporar:
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
	- Se amplía `backend/handlers/modulos_faltantes.go` para:
		- reemplazar la ruta generica de documentos por handlers especializados (`EmpresaDocumentosGestionHandler`, `EmpresaDocumentosFirmasHandler`).
		- incorporar versionado documental (`action=versionar`, `action=versiones`) y repositorio con control de acceso por rol/modulo (`action=acceso`, `action=repositorio`).
		- incorporar endurecimiento de integraciones con `action=rotar_credencial` (referencias seguras) y `action=monitoreo`/`action=alertas` (salud de conectores y SLA operativo).
	- Se amplía `backend/handlers/empresa_permisos.go` para clasificar `sync_manual`, `rotar_credencial` y `versionar` como acciones criticas de aprobacion en seguridad.
	- Se amplía cobertura de pruebas en `backend/handlers/modulos_faltantes_test.go` con:
		- `TestEmpresaIntegracionesAPIsHandlerRotarCredencialYMonitoreo`.
		- `TestEmpresaIntegracionesBancosHandlerRotarCredencial`.
		- `TestEmpresaDocumentosGestionHandlerVersionadoYControlAcceso`.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresa(IntegracionesAPIsHandlerRotarCredencialYMonitoreo|IntegracionesBancosHandlerRotarCredencial|DocumentosGestionHandlerVersionadoYControlAcceso)" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 23 (CRM/Produccion/Logistica):
	- Se amplía `backend/handlers/modulos_faltantes.go` para incorporar handlers especializados:
		- `EmpresaProduccionOrdenesHandler` con `action=plan_capacidad` (meta diaria, desviaciones y alertas por atraso/sobrecapacidad).
		- `EmpresaLogisticaEnviosHandler` con `action=seguimiento_hitos` (hitos programacion/salida/entrega, SLA y alertas de incumplimiento).
	- Se extiende `backend/handlers/reportes.go` en `operativo_cadena_cumplimiento` con metas y desviaciones por dominio:
		- `meta_cumplimiento_pct`, `desviacion_meta_pct`, `estado_meta`.
		- resumen global `meta_global_pct` y `desviacion_meta_global_pct`.
	- Se amplía cobertura de pruebas en:
		- `backend/handlers/modulos_faltantes_test.go` (`TestEmpresaProduccionOrdenesPlanCapacidad`, `TestEmpresaLogisticaEnviosSeguimientoHitos`).
		- `backend/handlers/reportes_test.go` (validaciones de metas/desviaciones en cadena).
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaProduccionOrdenesPlanCapacidad|TestEmpresaLogisticaEnviosSeguimientoHitos|TestEmpresaReportesHandlerDatasetOperativoCadenaCumplimiento" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 22 (RRHH extendido: vacaciones/licencias):
	- Se amplía `backend/db/modulos_faltantes.go` con nuevos campos de RRHH en `empresa_rrhh_vacaciones_licencias` para:
		- aprobacion jerarquica (`nivel_aprobacion_actual`, `nivel_aprobacion_requerido`, `aprobadores_json`, `historial_aprobaciones_json`, `fecha_aprobacion_final`),
		- acumulado y saldo (`periodo_acumulado_*`, `saldo_dias_*`, `saldo_snapshot_json`),
		- enlace a nomina (`empleado_nomina_id`, `nomina_liquidacion_id`, `nomina_periodo_*`, `nomina_vinculada_*`).
	- Se amplía `backend/handlers/modulos_faltantes.go` con handler especializado `EmpresaRRHHVacacionesLicenciasHandler` y acciones:
		- `action=resumen_saldo` para acumulado/saldo de vacaciones,
		- `action=solicitar_aprobacion`, `action=aprobar`, `action=rechazar` para flujo jerarquico,
		- `action=vincular_nomina` para enlazar novedades aprobadas a liquidacion/periodo de nomina.
	- Se actualiza `backend/handlers/empresa_permisos.go` para mapear acciones RRHH criticas a permisos de aprobacion/actualizacion.
	- Se amplía `backend/handlers/modulos_faltantes_test.go` con pruebas de:
		- saldo y aprobacion jerarquica multinivel,
		- vinculacion de novedades RRHH a nomina por periodo.
	- Validaciones ejecutadas:
		- `go test ./handlers -run RRHH -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-07
- Cierre tecnico del modulo 21 (Inventario extendido: lotes/series y devolucion proveedor):
	- Se amplía `backend/db/modulos_faltantes.go` con:
		- trazabilidad completa de lotes/series mediante tabla `inventario_lotes_series_movimientos`.
		- campos operativos de bloqueo/estado en `inventario_lotes_series` (`reservado_cantidad`, `vendido_cantidad`, `bloqueado_venta`, `bloqueo_motivo`, `ultima_operacion_*`).
		- campos contables de devolucion en `empresa_devoluciones_proveedor` (`periodo_contable`, `impacto_contable_*`, `fecha_contabilizacion`, `contabilizado_por`, `total_reintegrado`).
	- Se amplía `backend/handlers/modulos_faltantes.go` con handlers especializados:
		- `EmpresaInventarioLotesSeriesHandler` con acciones `trazabilidad`, `validar_disponibilidad`, `reservar`, `vender`, `liberar_reserva`, `ajuste_entrada`, `ajuste_salida`, `devolucion_proveedor`.
		- bloqueo automatico por vencimiento en venta/reserva y actualizacion de estado de lote.
		- `EmpresaComprasDevolucionesProveedorHandler` con `action=contabilizar`/`action=impacto_contable` para generar movimiento financiero, evento contable y actualizar la devolucion a `contabilizada`.
	- Se amplía `backend/db/eventos_contables.go` para soportar `devolucion_proveedor_contabilizada` en contrato y asiento contable (flujo de ingreso).
	- Se amplía `backend/handlers/modulos_faltantes_test.go` con pruebas de:
		- bloqueo automatico de lote vencido en reserva,
		- trazabilidad de ciclo reserva/venta/liberacion,
		- contabilizacion completa de devolucion proveedor con impacto contable.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaInventarioLotesSeriesBloqueoAutomaticoVencido|TestEmpresaInventarioLotesSeriesTrazabilidadCicloVenta|TestEmpresaComprasDevolucionesProveedorContabilizarImpactoCompleto" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).

## 2026-04-06
- Cierre tecnico del modulo 20 (Contabilidad operativa extendida: plan de cuentas, CxC y CxP):
	- Se amplía `backend/handlers/modulos_faltantes.go` con handlers especializados de finanzas:
		- `EmpresaFinanzasPlanCuentasHandler` con `action=plantillas`, `action=aplicar_plantilla` y `action=validar_cierre_periodo`.
		- `EmpresaFinanzasCuentasCobrarHandler` y `EmpresaFinanzasCuentasPagarHandler` con `action=conciliar_pagos` y validacion de periodo cerrado.
	- Se amplía `backend/db/modulos_faltantes.go` con:
		- nuevos metadatos de plantilla en `empresa_plan_cuentas`.
		- campos de conciliacion en `empresa_cuentas_por_cobrar` y `empresa_cuentas_por_pagar`.
		- bloqueo retroactivo por periodo contable cerrado en crear/editar/cambiar estado/eliminar de CxC/CxP.
	- Se amplía `backend/handlers/modulos_faltantes_test.go` con pruebas de:
		- plantillas y aplicacion de plan de cuentas por tipo de empresa.
		- conciliacion automatica CxC contra pagos reales.
		- bloqueo de operaciones CxP cuando el periodo esta cerrado.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaFinanzasPlanCuentasPlantillasYAplicacion|TestEmpresaFinanzasCuentasCobrarConciliacionPagosReales|TestEmpresaFinanzasCarteraBloqueoPeriodoCerrado" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).
- Cierre tecnico del modulo 19 (Gestion comercial extendida: cotizaciones/pedidos/devoluciones):
	- Se amplía `backend/handlers/modulos_faltantes.go` con automatizacion comercial en ventas:
		- `POST/PUT action=convertir_pedido` en cotizaciones para convertir cotizacion aprobada/emitida a pedido trazable (`cotizacion_id`, `convertido_pedido_id`).
		- `POST/PUT action=convertir_documento_final` en cotizaciones y pedidos para generar documento final en `empresa_facturacion_documentos`.
		- `GET action=embudo` en cotizaciones para monitoreo operativo con SLA y alertas de vencimiento.
	- Se incorpora snapshot de embudo comercial cotizacion→pedido→documento final con trazabilidad por `empresa_id`.
	- Se agrega dataset exportable `operativo_ventas_embudo_conversion` en `backend/handlers/reportes.go` con formatos `json/csv/txt/xls/pdf`.
	- Se actualiza `backend/handlers/empresa_permisos.go` para clasificar `convertir_pedido` y `convertir_documento_final` como acciones de aprobacion en ventas.
	- Se agregan pruebas en `backend/handlers/modulos_faltantes_test.go` y `backend/handlers/reportes_test.go` para:
		- conversion cotizacion→pedido→documento final,
		- alertas SLA del embudo,
		- dataset/export CSV del nuevo reporte de conversion.
	- Validaciones ejecutadas:
		- `go test ./handlers -run "TestEmpresaVentasCotizacionesConversionPedidoYDocumentoFinal|TestEmpresaVentasCotizacionesEmbudoYAlertasSLA|TestEmpresaReportesHandlerDatasetOperativoVentasEmbudoConversion" -count=1` (OK).
		- `go test ./handlers -run "DIAN|ModulosFaltantes|OperativoCadenaCumplimiento|OperativoVentasEmbudoConversion" -count=1` (OK).
		- `go test ./handlers -count=1` (OK).
		- `go test ./... -count=1` (OK).
- Cierre tecnico del modulo 18 (Facturacion electronica DIAN Colombia):
	- Se amplía `backend/handlers/modulos_faltantes.go` en `EmpresaDIANColombiaHandler` con acciones operativas reales:
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
	- Se amplía el ciclo documental de compras con aprobacion multinivel:
		- `requiere_aprobacion`, `niveles_aprobacion_requeridos`, `nivel_aprobacion_actual`, `aprobadores_json`.
		- Nuevas acciones: `solicitar_aprobacion`, `aprobar_compra`, `rechazar_compra`.
	- Se cierra recepcion parcial avanzada por item:
		- `recepcion_detalle_json` y `recepcion_resumen_json` para registrar cantidades solicitadas/recibidas, pendientes y diferencias.
		- Nueva accion: `recepcionar_parcial_compra`, consolidada con `recepcionar_compra` al completar recepcion total.
	- Se integra validacion documental proveedor-factura-entrada:
		- `validacion_documental_estado`, `proveedor_documento_ref`, `factura_documento_ref`, `entrada_documento_ref`.
		- Nueva accion: `validar_documentos` con verificacion de proveedor y referencias documentales.
	- Se amplía UI en `web/administrar_empresa/compras.html` con campos, filtros/KPI y acciones operativas del nuevo flujo.
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
	- Se amplía el modelo de comisiones con escalas por rol/servicio y tope por item:
		- nueva tabla `empresa_comisiones_servicio_escalas` (`rol_operacion`, `servicio_filtro`, `porcentaje_comision`, `tope_comision`, `prioridad`).
	- Se amplía `empresa_comisiones_servicio_movimientos` con trazabilidad operativa:
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
	- Se amplía la configuracion empresarial de propinas con reglas fiscales:
		- `pais_fiscal`, `regimen_fiscal`, `tratamiento_fiscal` (`no_gravada`/`gravada`) y `porcentaje_impuesto_propina`.
	- Se amplía el libro de movimientos de propinas con:
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
	- Se amplía `codigos_de_descuento` con reglas contextuales:
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
	- Se agrega cambio autogestionado de contraseña para usuario empresa:
		- `POST /api/empresa/usuarios/cambiar_password`.
	- Se implementan politicas de contraseña configurables desde `configuraciones`:
		- `usuarios.password_min_length`
		- `usuarios.password_require_uppercase`
		- `usuarios.password_require_lowercase`
		- `usuarios.password_require_digit`
		- `usuarios.password_require_symbol`
		- `usuarios.password_rotation_days`.
	- El login de usuario empresa ahora devuelve `password_rotation_required` cuando aplica rotacion obligatoria.
	- Se incorpora captura de notificaciones de confirmacion/restablecimiento en entorno de pruebas de correo:
		- tabla `super_correo_notificaciones_prueba` en `superadministrador.db`.
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
	- Se retira la configuración de Mercado Pago de `web/super/configuracion_avanzada.html` y se deja únicamente la sección de credenciales de Wompi en configuración avanzada del panel super administrador.
	- Se simplifica `web/pagar_licencia.html` eliminando selector/panel/flujo de Mercado Pago para operar solo con Nequi (Wompi) y activación manual interna.
	- Se desregistran rutas de Mercado Pago en `backend/main.go` (`/super/api/config/mercadopago`, `/mercadopago/create_preference`, `/mercadopago/webhook`, `/mercadopago/reconcile`, `/mercadopago/test_preference`).
	- Validación técnica ejecutada: `go test ./... -run "^$" -count=1` (compilación global OK).
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
	- Se rediseña el dataset `operativo_compras_movimientos` en `backend/handlers/reportes.go` para consolidar compras por proveedor, dejando de depender solo de movimientos de inventario.
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
	- `backend/handlers/modulos_faltantes_test.go` amplía pruebas para:
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
		- transiciones validas/inválidas de cotizaciones y leads.
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
		- resolucion de contexto licencia/empresa para Wompi.
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
- Se amplía `web/ayuda/ayuda.html` con seccion detallada para configurar facturacion DIAN desde cero.
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
- Se agrega `ventas_simple.html` como carrito alterno por estación (modo supermercado) con activación/desactivación por estación.
	- `web/administrar_empresa/ventas_simple.html` (nuevo) incorpora flujo rápido para buscar productos, agregarlos al carrito, ajustar cantidades y visualizar total consolidado por estación.
	- Se corrige la visibilidad del campo de referencia de pago para métodos que la requieren (`tarjeta_credito`, `tarjeta_debito`, `transferencia_bancaria`).
	- El cobro se ejecuta con flujo simplificado usando `action=pagar_estacion` y permite iniciar nueva venta con `action=activar_estacion`.
	- `web/administrar_empresa/configuracion_de_estaciones.html` agrega la bandera local `venta_simple_habilitada` por estación.
	- `web/administrar_empresa/estaciones.html` enruta automáticamente cada estación al carrito completo (`carrito_de_compras.html`) o al carrito simple (`ventas_simple.html`) según su configuración.
	- `web/estilos.css` integra estilos responsive para el nuevo módulo y etiqueta visual del modo por estación.

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
- Se crea el modulo de tarifas por dia por estacion con recálculo automático de deuda en carritos hotel activos.
	- `backend/db/tarifas_por_dia.go` (nuevo) agrega esquema `empresa_tarifas_por_dia`, CRUD, horarios `hora_check_in`/`hora_check_out` y calculo de dias/monto.
	- `backend/db/carritos_tarifa_dia.go` (nuevo) integra calculo automático de deuda diaria en carritos de estación y refresco masivo para listados.
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
	- `web/administrar_empresa.html` y `web/js/administrar_empresa.js` integran el acceso lateral `Facturas electrónicas` con permisos de lectura del modulo `facturacion`.
	- `backend/handlers/facturacion_electronica.go` incorpora:
		- `GET action=documentos` para consulta de documentos de facturacion por filtros (`cliente`, `documento`, `fecha_desde`, `fecha_hasta`, `estado_documento`, `tipo_documento`, `q`).
		- `PUT/POST action=reenviar_correo` para reintento manual de envio al correo del cliente.
	- `backend/db/documentos_transaccionales.go` agrega consulta enriquecida con cliente (`nombre`, `email`, `documento`) para listado filtrado.
	- Cobertura agregada en `backend/db/documentos_transaccionales_test.go` para filtros de cliente/fecha/documento.

## 2026-04-05
- Se normaliza la documentacion de base de datos para eliminar duplicidad entre documentos.
	- `estructura_bd.md` queda como fuente canonica del esquema fisico SQLite.
	- `documentos/descripcion_de_las_bases_De_datos` se redefine como guia complementaria funcional y reglas operativas de mantenimiento.
	- Se evita repetir listados tabla-por-tabla en dos archivos distintos.

## 2026-04-05
- Se consolida Configuración avanzada dentro de Facturación electrónica en el panel de empresa.
	- `web/administrar_empresa/facturacion_electronica.html` integra el formulario completo de configuración avanzada fiscal/impresión y su persistencia mediante `/api/empresa/configuracion_avanzada`.
	- `web/administrar_empresa.html` elimina el enlace lateral independiente `Configuración avanzada` para dejar una única entrada funcional en `Facturación electrónica`.
	- `web/js/administrar_empresa.js` retira `linkConfigAvanzada` del catálogo de enlaces/permisos del menú.
	- `web/ayuda/ayuda.html` actualiza el tutorial para indicar que la configuración avanzada ahora se gestiona desde `facturacion_electronica.html`.
	- `web/administrar_empresa/configuracion_avanzada.html` se elimina del repositorio por consolidación funcional.

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
	- `web/administrar_empresa/carrito_de_compras.html` incorpora transferencia bancaria en selectores de pago, habilita validacion de pago mixto con transferencia y envía `pagos_mixtos` al backend.
	- `web/administrar_empresa/finanzas.html` estandariza opcion de `transferencia_bancaria` y mantiene compatibilidad con registros legacy `transferencia`.
	- `web/ayuda/ayuda.html` actualiza descripcion de metodos soportados en cierre de carrito.
	- `documentos/descripcion_del_proyecto`, `documentos/diagramas/estructura_del_codigo.md` y `documentos/diagramas/diagrama_flujo_procesos.md` reflejan el nuevo flujo de pago.

## 2026-04-05
- Robustecimiento del modulo de auditoria empresarial con foco en trazabilidad, seguridad operativa y analisis forense.
	- `backend/db/auditoria_empresa.go` amplía filtros (`metodo_http`, `recurso`, `endpoint`, `search`), agrega `offset`, agrega conteo filtrado (`CountEmpresaAuditoriaEventos`) y refuerza indices de rendimiento.
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
- Facturación electrónica: envío automático del resumen de factura al correo del cliente al emitir.
	- `backend/handlers/facturacion_electronica.go` ahora intenta enviar correo en `action=emitir` de `factura_electronica`.
	- Soporta destinatario por `cliente_email` o por `cliente_id`/`entidad_id` consultando clientes.
	- La respuesta incluye bloque `factura_email` con estado de intento/envío/error sin bloquear la emisión legal.
	- `backend/db/clientes.go` agrega `GetClienteByID` para resolver destinatario desde la base de datos.
	- `backend/main.go` actualiza la inyección de `dbSuper` al handler de facturación para lectura de SMTP.
	- `web/administrar_empresa/facturacion_electronica.html` agrega campos de cliente y muestra el resultado de envío en pantalla.
	- Cobertura añadida en `backend/db/clientes_test.go` y `backend/handlers/eventos_contables_modulos_test.go`.

## 2026-04-05
- Se crea el modulo de codigos de descuento por empresa y validacion de metodos de pago en carrito de compras.
	- `backend/db/codigos_descuento.go` (nuevo) agrega la tabla `codigos_de_descuento`, generacion automatica de codigos, CRUD, validacion por vencimiento/usos y resolucion de descuento aplicable por monto.
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
- Centro de ayuda actualizado con tutorial por cada módulo del sistema.
	- `web/ayuda/ayuda.html` amplía el contenido con una sección de tutoriales por módulos de administración global y módulos del panel de empresa.
	- Se agregan pasos operativos por módulo y enlaces directos a cada pantalla para facilitar onboarding y uso diario.

## 2026-04-04
- Verificacion integral real de modulos + limpieza de artefactos temporales.
	- Validacion real ejecutada (sin simulaciones/mocks) sobre SQLite y capa HTTP:
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
		- `backend/db/empresas.db.20260326-174525.bak`.
		- `backend/db/superadministrador.db.20260326-174324.bak`.
		- `backend/db/superadministrador.db.20260326-174525.bak`.

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
- Punto 5 (control de inventarios) — continuidad preventiva-compras ciclo documental desde reposicion.
	- `backend/db/productos.go` agrega `InventarioPlanReposicionOrdenEstadoActualizado` y `ActualizarEstadoOrdenCompraDesdeReposicion` para transiciones `recepcionar_compra` y `contabilizar_compra`.
	- `backend/handlers/productos.go` agrega endpoint `POST /api/empresa/compras/plan_reposicion/actualizar_estado`.
	- `backend/main.go` registra `/api/empresa/compras/plan_reposicion/actualizar_estado` bajo permisos de compras.
	- `web/administrar_empresa/administrar_productos.html` amplía el flujo a `fases 10-12` con acciones `Recepcionar orden` y `Contabilizar orden` y contexto de estado de OC.
	- Cobertura nueva en:
		- `backend/handlers/productos_categorias_test.go`.
		- `backend/db/productos_categorias_test.go`.
- Validacion tecnica:
	- `go test ./db -run "TestEmitirOrdenCompraDesdePlanReposicionBorradorPersistDoc|TestActualizarEstadoOrdenCompraDesdeReposicionCiclo"` (ok).
	- `go test ./handlers -run "TestEmpresaComprasPlanReposicionEmitirOrdenHandlerEmiteDocumento|TestEmpresaComprasPlanReposicionActualizarEstadoHandlerGestionaCiclo"` (ok).
	- `get_errors` sobre backend/frontend modificado (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad preventiva-compras emitible desde borrador.
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
- Punto 5 (control de inventarios) — continuidad preventiva-compras ordenable por proveedor.
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
- Punto 5 (control de inventarios) — continuidad preventiva-compras consolidada por proveedor.
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
- Punto 5 (control de inventarios) — continuidad preventiva-compras con plan de reposicion por proveedor.
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
- Punto 5 (control de inventarios) — continuidad preventiva con proyeccion de quiebre.
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
- Punto 5 (control de inventarios) — continuidad operativa-analitica con balance por bodega.
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
- Punto 5 (control de inventarios) — continuidad analitica con tendencia diaria.
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
- Punto 5 (control de inventarios) — continuidad operacional en panel de productos.
	- `web/administrar_empresa/administrar_productos.html` agrega:
		- bloque `Top productos críticos (déficit)` alimentado desde alertas de inventario,
		- priorización de críticos por `sin_stock` y mayor déficit,
		- acción `Preparar reposición` para precargar ajuste de inventario con producto, bodega y cantidad sugerida.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — continuidad KPI operativo en panel de productos.
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
- Punto 5 (control de inventarios) — continuidad UI operativa en panel de productos.
	- `web/administrar_empresa/administrar_productos.html` agrega:
		- filtro por bodega para alertas de quiebre,
		- filtros de kardex por bodega, tipo y rango de fechas,
		- acciones `Filtrar` y `Limpiar` en ambos bloques de consulta.
	- Se actualiza documentacion asociada en plan maestro y estructura tecnica.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/administrar_productos.html` (ok).

## 2026-04-04
- Punto 5 (control de inventarios) — inicio tecnico: kardex operativo + reglas de stock + alertas de quiebre por bodega.
	- `backend/db/productos.go`:
		- valida `stock_minimo/stock_maximo` en creacion y edicion de productos,
		- agrega `GetAlertasQuiebreByEmpresa`,
		- amplía `GetMovimientosByEmpresa` con filtros `bodega_id`, `tipo`, `desde`, `hasta`.
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
- Punto 3 (permisos y seguridad) — continuidad operativa: catalogo frontend por rol + regresion endpoints sin wrapper.
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
- Punto 3 (permisos y seguridad) — consolidacion documental endpoint/rol y checklist UAT:
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
- Punto 1 + Punto 2 (plan maestro) — cierre de backlog inmediato con formalizacion tecnica documental.
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
- Punto 11 (reportes financieros) — continuidad de backlog inmediato: exportacion unificada del tablero por rango.
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
- Punto 10 (modulo contable integrado) — continuidad de backlog inmediato: vista de conciliacion por periodo (eventos vs asientos).
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
- Punto 10 (modulo contable integrado) — continuidad de backlog inmediato: ejecucion automatica por lotes de asientos.
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
- Punto 15 (auditoria por empresa) — continuacion de backlog inmediato 1 y 2:
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
- Punto 15 (auditoria por empresa) — continuacion de backlog 1, 2 y 3:
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
- Punto 15 (auditoria por empresa) — implementacion base minima:
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
	- `backend/db/eventos_contables.go` amplía `empresa_eventos_contables` con metadatos de procesamiento (`intentos_procesamiento`, `fecha_ultimo_intento`, `error_procesamiento`, `asiento_contable_id`) y crea tabla canonica `empresa_asientos_contables` con hash de idempotencia.
	- `backend/handlers/finanzas.go` agrega `EmpresaFinanzasAsientosContablesHandler`:
		- `GET /api/empresa/finanzas/asientos_contables` para consulta,
		- `POST/PUT action=procesar_asientos|procesar` para procesamiento manual por lote.
	- `backend/handlers/empresa_permisos.go` clasifica `action=procesar_asientos` como accion de aprobacion en finanzas.
	- `backend/main.go` publica `/api/empresa/finanzas/asientos_contables` y registra migracion `2026-04-04-010-asientos-canonicos`.
	- `backend/db/finanzas.go` integra en el tablero los bloques `estado_resultados` y `balance_general`, junto con KPI contables de asientos (`asientos_generados`, `asientos_monto_total`).
	- `web/administrar_empresa/reportes.html` incorpora visualizacion de utilidad operacional, activos/pasivos/patrimonio, resultado del ejercicio y cuadre.
	- `web/administrar_empresa/finanzas.html` añade accion manual `Procesar eventos contables`.
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
- Punto 12 (cierres de caja) — continuacion con UI operativa en panel empresa:
	- `web/administrar_empresa/finanzas.html` integra modulo visual de cierres de caja por sucursal con:
		- formulario de apertura/actualizacion,
		- calculo de `caja_teorica` y `diferencia_caja`,
		- filtros por sucursal/caja/estado/fecha,
		- tabla de acciones (`cerrar`, `reabrir`, `aprobar`, `anular`, `activar/desactivar`, `eliminar`).
	- La vista queda conectada al endpoint existente `GET/POST/PUT/DELETE /api/empresa/finanzas/cierres_caja`.
- Validacion tecnica:
	- `get_errors` sobre `web/administrar_empresa/finanzas.html` (ok).

## 2026-04-04
- Punto 12 (cierres de caja) — inicio de flujo operativo por sucursal:
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
- Punto 11 (reportes financieros) — inicio de tablero minimo financiero-operativo:
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
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) — persistencia canonica de documentos transaccionales para `entidad_id`:
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
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) — estandarizacion de estados en ciclo documental transaccional:
	- Se agrega `backend/handlers/documentos_lifecycle.go` con reglas de transicion por accion y estado previo para facturacion/compras.
	- `backend/handlers/facturacion_electronica.go` ahora valida `estado_actual` en `emitir/anular/nota_credito`, devuelve `409` en conflictos y responde `estado_anterior`/`estado_nuevo` cuando la transicion es valida.
	- `backend/handlers/productos.go` (`EmpresaProveedoresHandler`) aplica validacion equivalente para `emitir_orden/recepcionar_compra/contabilizar_compra`.
	- `backend/handlers/eventos_contables_modulos_test.go` amplía cobertura con pruebas de transiciones invalidas para facturacion y compras.
- Validacion tecnica:
	- `runTests` sobre `backend/handlers/eventos_contables_modulos_test.go` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras) — eventos transaccionales de factura y orden:
	- `backend/handlers/facturacion_electronica.go` agrega acciones transaccionales:
		- `action=emitir` -> `factura_emitida`.
		- `action=anular` -> `factura_anulada`.
		- `action=nota_credito|emitir_nota_credito` -> `nota_credito_emitida`.
	- `backend/handlers/productos.go` (`EmpresaProveedoresHandler`) agrega acciones transaccionales:
		- `action=emitir|emitir_orden` -> `orden_compra_emitida`.
		- `action=recepcionar|recepcionar_compra` -> `compra_recepcionada`.
		- `action=contabilizar|contabilizar_compra` -> `compra_contabilizada`.
	- `backend/handlers/empresa_permisos.go` amplía mapeo de acciones de permisos para compras/facturacion.
	- `backend/handlers/eventos_contables_modulos_test.go` agrega pruebas de emisiones transaccionales de factura/orden.
- Validacion tecnica:
	- `go test ./handlers -run "FacturacionTransaccionalEmiteEventosContables|ComprasTransaccionalEmiteEventosContables|FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 8 + Punto 9 + Punto 10 (facturacion/compras/finanzas) — extension de emision de eventos contables por modulo:
	- Se agrega `backend/handlers/eventos_contables.go` para registro no bloqueante y reutilizable de eventos contables en handlers.
	- Se amplia `backend/db/eventos_contables.go` con eventos operativos de:
		- `facturacion`: `configuracion_facturacion_actualizada`.
		- `compras`: `proveedor_registrado`, `proveedor_actualizado`, `proveedor_activado`, `proveedor_desactivado`, `proveedor_eliminado`.
	- Se integra emision en:
		- `backend/handlers/facturacion_electronica.go`.
		- `backend/handlers/productos.go` (proveedores).
		- `backend/handlers/finanzas.go` (movimientos y periodos).
	- `backend/handlers/carritos_compras.go` migra a helper comun para consistencia del registro contable.
	- Se agregan pruebas en `backend/handlers/eventos_contables_modulos_test.go` para validar emision en facturacion, compras y finanzas.
- Validacion tecnica:
	- `go test ./db -run "EventosContables" -count=1` (ok).
	- `go test ./handlers -run "FacturacionElectronicaEmiteEventoContable|ProveedoresEmiteEventoContableCompras|FinanzasEmiteEventosContables|CarritosCompraAndItemsFlow" -count=1` (ok).
	- `go test ./...` en `backend` (ok).

## 2026-04-04
- Punto 4 + Punto 10 (gestion de ventas + modulo contable integrado) — contrato de eventos contables por modulo:
	- Se agrega `backend/db/eventos_contables.go` con contrato base de eventos para `ventas`, `facturacion`, `compras` y `finanzas`.
	- Se crea tabla `empresa_eventos_contables` en `empresas.db` para registrar trazabilidad contable por empresa (`modulo`, `evento`, `entidad`, `documento`, `periodo_contable`, `monto`, `payload_json`, `procesado`).
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
- Punto 4 (gestion de ventas) — formalizacion de transiciones del ciclo de venta en carritos:
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
	- `backend/main.go` amplía middleware en rutas:
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
	- `backend/handlers/empresa_permisos.go` amplía modulos de autorizacion para `clientes`, `compras` y `facturacion`.
	- Se agregan wrappers: `WithEmpresaClientesPermissions`, `WithEmpresaComprasPermissions`, `WithEmpresaFacturacionPermissions`.
	- `backend/main.go` aplica middleware en rutas: `/api/empresa/clientes`, `/api/empresa/proveedores`, `/api/empresa/facturacion_electronica`, `/api/empresa/facturacion_electronica/pais_detectado`, y `/api/empresa/servicios` (politica inventario).
	- Se amplian pruebas en `backend/handlers/empresa_permisos_test.go` para cobertura de los modulos nuevos.
	- Validacion tecnica: `go test ./handlers -run "WithEmpresa|ConsultarHandlerRejectsEmpresaFueraDeAlcance" -count=1` (ok) y `go test ./...` (ok).

## 2026-04-04
- Se registra nueva credencial Gemini cifrada en configuración avanzada (`ai.model.google.gemini_2_0_flash.api_key` en `superadministrador.db`).
- Se valida consumo de Gemini con la nueva credencial: respuesta del proveedor `429` por cuota excedida (sin error de credencial/servicio bloqueado).
- Se verifica la presencia de la tarjeta de Gemini en `web/super/configuracion_avanzada.html` y se corrige un bloque JavaScript en la carga de estado para mantener consistencia de la vista.
- Se agrega prueba de seguridad de alcance por empresa para chat IA en `backend/handlers/chat_con_inteligencia_artificial_controller_test.go`:
	- `TestConsultarHandlerRejectsEmpresaFueraDeAlcance`.
	- Validación: `go test ./handlers -run "TestConsultarHandlerRejectsEmpresaFueraDeAlcance|TestModelosHandlerRequiresGoogleAccount|TestModelosHandlerReturnsPreferredModelForGoogleAccount" -count=1` (ok).

## 2026-04-04
- Chat IA empresarial migrado a Gemini-only:
	- `backend/handlers/chat_con_inteligencia_artificial_controller.go` ahora integra Google Gemini (`generateContent`) y elimina dependencias de OpenAI/DeepSeek/Hugging Face para este módulo.
	- El catálogo y la configuración de credenciales IA quedan en un único modelo soportado: `google:gemini-2.0-flash` (`GEMINI_API_KEY`).
	- `web/super/configuracion_avanzada.html` simplifica la tarjeta IA a una sola credencial Gemini con trazabilidad por cuenta Google.
	- `web/administrar_empresa/chat_con_inteligencia_artificial.html` se rediseña con experiencia visual tipo Gemini, chips de contexto y flujo explícito de autenticación Google.
	- Pruebas ajustadas y validadas: `go test ./auth ./db ./handlers ./metrics ./utils` (ok) en `backend`.
- Se agrega gestión de credenciales IA en `super/configuracion_avanzada.html` para 5 modelos populares con plan gratuito limitado:
	- OpenAI GPT-4o mini,
	- OpenAI GPT-4.1 mini,
	- DeepSeek Chat,
	- DeepSeek Reasoner,
	- Meta Llama 3.1 8B Instruct (Hugging Face).
- Se crea endpoint `GET/PUT /super/api/config/ai` en backend para guardar/consultar credenciales con registro de la cuenta Google logueada que realiza cambios.
- El módulo `chat_con_inteligencia_artificial` ahora resuelve credenciales en este orden:
	- configuración guardada por modelo,
	- configuración por proveedor,
	- variable de entorno.
- Validación técnica ejecutada:
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
- Continuación de implementación en `chat_con_inteligencia_artificial`:
	- Se corrige el orden de validación de autenticación para cuenta Google en `backend/handlers/chat_con_inteligencia_artificial_controller.go`.
	- Cuando no hay cuenta Google autenticada, los endpoints del módulo IA ahora responden `401` de forma consistente (en lugar de caer en validación de alcance con `403`).
	- Se centraliza validación de alcance con `ensureEmpresaAccessByAccount` para reutilizar la cuenta ya validada.
- Se agregan pruebas automáticas del módulo IA:
	- `backend/db/chat_inteligencia_artificial_test.go` (upsert/get de modelo preferido y acumulación de uso diario).
	- `backend/handlers/chat_con_inteligencia_artificial_controller_test.go` (autorización por cuenta Google y respuesta con modelo preferido).
- Validación técnica ejecutada en esta continuación:
	- `go test ./db -run EmpresaAI -count=1` (ok).
	- `go test ./handlers -run ModelosHandler -count=1` (ok).
	- `go test ./...` en `backend` (ok).
- Se amplía el módulo `chat_con_inteligencia_artificial` para registrar el modelo preferido por cuenta Google autenticada (por empresa):
	- Nueva tabla `empresa_ai_modelo_preferido` en `empresas.db` (UNIQUE por `empresa_id + admin_email`).
	- Nuevas funciones en `backend/db/chat_inteligencia_artificial.go`: `GetEmpresaAIModeloPreferido` y `UpsertEmpresaAIModeloPreferido`.
	- Nuevo endpoint `GET/PUT /api/empresa/chat_con_inteligencia_artificial/modelo_preferido`.
	- `GET /modelos` ahora devuelve `google_account` y `modelo_preferido`.
	- `POST /consultar` ahora persiste el `model_id` usado como preferencia de la cuenta Google y devuelve confirmación en respuesta.
- Se actualiza `web/administrar_empresa/chat_con_inteligencia_artificial.html` para:
	- cargar automáticamente el modelo preferido de la cuenta Google,
	- guardar el modelo preferido al cambiar selección,
	- mostrar la cuenta Google vinculada en el bloque de uso diario.
- Validación técnica ejecutada para esta ampliación:
	- `gofmt -w backend/db/chat_inteligencia_artificial.go backend/handlers/chat_con_inteligencia_artificial_controller.go backend/handlers/chat_con_inteligencia_artificial_router.go` (ok).
	- `go test ./...` en `backend` (ok).
- Se fortalece `backend/utils/utils.go` para observabilidad profesional:
	- `LoggingMiddleware` ahora genera `request_id` por solicitud, calcula `empresa_id` (query/header/JSON body) y registra inicio/fin con latencia.
	- Se agregan logs separados por empresa en `backend/logs/empresa_<id>.log` y un fallback global en `backend/logs/empresa_global.log`.
	- `JSONErrorMiddleware` ahora normaliza errores no-JSON incluyendo `request_id` y `empresa_id` cuando aplica, y registra errores API por empresa.
- Se ajustan endpoints multipart para reforzar separación de logs por empresa:
	- `backend/handlers/chat_tareas.go` y `backend/handlers/productos.go` ahora establecen `X-Empresa-ID` tras parsear `empresa_id` del formulario.
- Se endurece `backend/handlers/usuarios_empresa.go` en autenticación/primer ingreso:
	- se reemplazan respuestas `500` que exponían detalles internos por mensajes profesionales y seguros,
	- se agrega logging servidor con contexto (`empresa_id`, `email`, `id`) para trazabilidad sin filtrar errores sensibles al cliente.
- Se endurece `scripts/iniciar_servidor.ps1` para detectar caída temprana de `server.exe`: ahora conserva el `PID`, valida salida prematura y muestra las últimas líneas de `backend/server.err` para diagnóstico inmediato.
- Validación de corrección ejecutada:
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
- Se agregan tablas en `empresas.db` para auditoria y limites diarios:
	- `empresa_ai_consultas`
	- `empresa_ai_uso_diario`
- Se integra `EnsureEmpresaAIChatSchema` y la migracion `2026-04-03-005-chat-ia-empresa` en `backend/main.go`.
- Se implementa aislamiento estricto por `empresa_id`, validacion de alcance de usuario y control de limite free-tier por empresa/proveedor/modelo/dia con opcion de upgrade.
- Se habilitan modelos famosos de OpenAI, DeepSeek y Hugging Face usando credenciales solo en backend mediante variables de entorno (`OPENAI_API_KEY`, `DEEPSEEK_API_KEY`, `HUGGINGFACE_API_KEY`).
- Se amplía el módulo financiero con control de periodos contables por empresa:
	- tabla `empresa_finanzas_periodos`.
	- endpoint `GET/POST/PUT /api/empresa/finanzas/periodos`.
	- acciones de cierre y reapertura de periodo.
- Se aplican bloqueos de integridad contable: no se permite crear/editar/eliminar/activar/desactivar movimientos cuando su periodo está cerrado.
- Se amplía `empresa_finanzas_movimientos` con:
	- `periodo_contable`,
	- retenciones (`retencion_fuente`, `retencion_ica`, `retencion_iva`, `total_retenciones`),
	- `total_neto`.
- Se amplía `empresa_finanzas_configuracion` con `cuenta_retenciones_cobrar` y `cuenta_retenciones_pagar`.
- Se completa la UI de finanzas para:
	- gestionar periodos (cerrar/reabrir/actualizar),
	- calcular total bruto, retenciones y neto,
	- filtrar por periodo,
	- exportar `balance general`, `libro diario` y `libro mayor` en CSV.
- Se corrige el escaneo de puertos de seguridad para compatibilidad IPv6 usando `net.JoinHostPort` en `backend/handlers/system_empresas_handlers.go`.
- Se ajusta `scripts/iniciar_servidor.ps1` para usar nombre de función con verbo aprobado de PowerShell en la carga de `.env`.
- Validación técnica ejecutada: `go test ./...` en `backend` (ok).
- Se implementa el módulo financiero multiempresa con enfoque unificado de ingresos y egresos en `web/administrar_empresa/finanzas.html`.
- Se crea `backend/db/finanzas.go` con esquema, validaciones y CRUD de:
	- `empresa_finanzas_movimientos`
	- `empresa_finanzas_configuracion`
- Se crea `backend/handlers/finanzas.go` y se publican rutas:
	- `GET/POST/PUT/DELETE /api/empresa/finanzas/movimientos`
	- `GET/POST/PUT /api/empresa/finanzas/configuracion`
- Se actualiza `backend/main.go` para asegurar el esquema financiero y registrar la migración `2026-04-03-003-finanzas`.
- Se integra el acceso al módulo en `web/administrar_empresa.html` y `web/js/administrar_empresa.js`.
- Se agrega `backend/db/finanzas_test.go` con pruebas de configuración y flujo CRUD de movimientos financieros.
- Se amplía `backend/tools/seed_motel_malibu/main.go` para sembrar configuración financiera y movimientos demo de ingreso/egreso.
- Se separa visualmente el libro financiero en dos pestañas operativas dentro del módulo: `Ingresos` y `Egresos`.
- Se agrega la pestaña `Todos` para consolidar ingresos y egresos en una sola vista del libro financiero.
- Se agrega exportación del libro financiero filtrado por fechas a:
	- Excel (CSV compatible con Excel).
	- PDF (vista de impresión).
	- JSON contable para integración externa (incluye resumen, detalle y asientos recomendados).
- Se amplía la configuración financiera por empresa para contabilidad externa con parametrización de:
	- destino de integración (`generico`, `siigo`, `world_office`, `alegra`),
	- cuentas base (caja/bancos, ingresos, IVA generado, gastos, IVA descontable),
	- cuentas por categoría para ingresos y egresos.
- La exportación `JSON contable` deja de usar cuentas fijas y ahora construye asientos con la parametrización real guardada por empresa.
- El JSON exportado incorpora `accounting_profile` y `erp_projection` por movimiento para facilitar mapeo hacia software contable externo.
- Se actualiza `backend/db/finanzas_test.go` para validar persistencia de la nueva parametrización contable.
- Se amplía `web/administrar_empresa/finanzas.html` con salidas contables adicionales:
	- Plantilla dedicada SIIGO (CSV) para importación de asientos.
	- Balance de prueba (CSV).
	- Estado de resultados (CSV).
- Se crea `documentos/plantillas/siigo_plantilla_importacion_asientos.csv` como plantilla de referencia ERP.
- Se crea `documentos/informe_contable_directivo_2026-04-03.md` con revisión de cumplimiento contable/directivo, brechas y plan recomendado.
- Validación técnica ejecutada:
	- `go test ./... -count=1` (ok).
	- `go run ./tools/seed_motel_malibu` (ok, incluye creación de 4 movimientos financieros demo).
	- `runTests` global (ok: 3/3).

## 2026-04-03
- Se implementa control de inventario en carrito: al agregar items de producto se descuenta stock y al desactivar/eliminar items abiertos se revierte automáticamente.
- Se asegura que, al cerrar una venta, el descuento de inventario permanezca aplicado y no se revierta en el pago.
- Se mejoran respuestas de API para stock insuficiente en operaciones de items de carrito.
- Se agrega `backend/db/carritos_inventario_test.go` con pruebas de descuento de inventario y caso de stock insuficiente.
- Se amplía `backend/tools/seed_motel_malibu/main.go` para registrar 10 clientes y 10 usuarios de empresa.
- La semilla valida automáticamente el flujo comercial completo: venta cerrada, descuento de inventario al agregar y persistencia tras pagar.
- Se confirma en seed la validación de impresión con vista previa POS y Carta.
- Se amplía `web/administrar_empresa/reportes.html` con reporte de ventas, reporte de productos y reporte de compra de productos, todos con búsqueda por rango de fechas.
- Validación técnica ejecutada: `go test ./auth ./db ./handlers ./metrics ./utils` (ok) y `go run ./tools/seed_motel_malibu` (ok).
- Se agrega el vínculo `Ayuda` en el menú flotante global (`web/menu.js`) y se reestructura `web/ayuda/ayuda.html` como centro de ayuda con menú interno y sección de APIs.
- Se adapta `web/administrar_empresa/carrito_de_compras.html` para operación con lector de código de barras (escaneo por código/SKU, Enter para agregar y acumulación opcional de cantidad).
- Se extiende `web/administrar_empresa/configuracion.html` con configuración por empresa para el lector: habilitar, autofoco y acumulación.
- Se amplía `web/administrar_empresa/reportes.html` con KPI de productos bajo mínimo y reporte de inventario actual por bodega.
- Validación técnica ejecutada para flujo carrito/inventario multiempresa: `go test ./db -run Carrito -count=1` (ok) y `go test ./handlers -run Carritos -count=1` (ok).

## 2026-04-02
- Se crea la herramienta `backend/tools/seed_motel_malibu/main.go` para cargar datos demo comerciales en la empresa Motel Malibu.
- La semilla inserta 10 productos con precios COP, 5 clientes y crea una venta de prueba cerrada para validar el flujo comercial.
- Se valida la configuracion de impresion con vista previa de formatos POS y Carta desde la herramienta de seed.
- Se implementa la seccion `web/administrar_empresa/reportes.html` con KPIs, ventas cerradas, top productos, top clientes y resumen de impresion.
- Se reestructura `backend/tools` en subcarpetas por herramienta para eliminar conflictos de compilación por múltiples `main`.
- Se valida backend completo con `go test ./...` (ok).
- Se valida el módulo GPS con pruebas específicas:
	- `go test ./db -run TestEmpresaGPSDispositivosYRecorridosCRUD -count=1` (ok).
	- `go test ./handlers -run TestEmpresaUbicacionGPSHandlersCRUDFlow -count=1` (ok).
- Se implementa el modulo de ubicacion GPS por empresa con soporte de multiples dispositivos.
- Se agregan tablas `empresa_gps_dispositivos` y `empresa_gps_recorridos` en `empresas.db`.
- Se crean endpoints CRUD para dispositivos y recorridos GPS en `/api/empresa/ubicacion_gps/*`.
- Se agrega la pagina `web/administrar_empresa/ubicacion_gps.html` con mapa OpenStreetMap (Leaflet).
- Se habilita tracking automatico de recorridos cada 10 segundos por dispositivo.
- Se agregan pruebas en `backend/db/ubicacion_gps_test.go` y `backend/handlers/ubicacion_gps_test.go`.
