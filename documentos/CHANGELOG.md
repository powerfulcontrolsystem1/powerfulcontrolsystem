## [2026-06-06] Renta IA en finanzas
- [Frontend] `administrar_empresa/renta_ia.html` agrega una pantalla financiera para estimar renta empresarial con periodo, tarifa, sobretasa, ajustes fiscales, retenciones y fuentes reales.
- [Backend] `/api/empresa/finanzas/renta_ia` calcula ingresos, deducciones, renta liquida gravable, impuesto estimado y saldo usando datos de ventas, movimientos financieros, inventario y nomina por `empresa_id`.
- [IA] La explicacion con GPT-5.4 mini usa solo el JSON calculado por backend, respeta limite diario por empresa y conserva fallback numerico cuando la IA esta desactivada.
- [Permisos] Se agrega `linkRentaIA` bajo `finanzas:R`, visible para contador y roles con lectura financiera.
- [QA] Produccion empresa 12: la pagina calcula con fuentes reales sin errores; el analisis generativo queda bloqueado por credencial IA cifrada no descifrable y requiere regrabar la API key en Super administrador > IA.

## [2026-06-06] Modulo Bolsa empresarial
- [Frontend] `administrar_empresa/bolsa.html` agrega un tablero profesional de indicadores internacionales y del pais detectado por empresa.
- [Backend] `/api/empresa/bolsa` consulta datos de mercado en vivo desde servidor, aplica cache corta, detecta pais con la configuracion empresarial y devuelve errores por indicador sin romper la pantalla.
- [Permisos] Se agrega modulo `bolsa`, pagina `linkBolsa` y wrapper `WithEmpresaBolsaPermissions` de lectura, con fallback de licencia a reportes/finanzas.

## [2026-06-06] OnlyOffice JWT y PPTX real
- [Operacion] Se alineo el secreto JWT del Document Server OnlyOffice con el backend en VPS sin exponer el valor, corrigiendo el error visual `The document security token is not correctly formed`.
- [Backend] `onlyOfficeBuildBlankPPTX` genera presentaciones PresentationML con master, layout, theme, propiedades y relaciones minimas para que OnlyOffice abra archivos `.pptx` nuevos.
- [QA] `go test ./handlers -run OnlyOffice -count=1`; verificacion visual autenticada de Word y Excel correctos, y reproduccion visual del fallo de PPTX antes del parche.

## [2026-06-06] Barra de avance DIAN
- [Frontend] `facturacion_electronica_pruebas_dian.html` agrega una barra 0-100% para que el administrador vea el avance del proceso DIAN por hitos reales.
- [Operacion] El porcentaje considera configuracion base, firma, TestSetId, objetivo del set, credenciales validadas, envio real, acuse final y produccion local.
- [Seguridad] La barra no muestra secretos y no reemplaza el acuse final aceptado de DIAN.

## [2026-06-06] Tutorial DIAN en submenu de facturacion electronica
- [Frontend] `facturacion_electronica_tutorial_dian.html` agrega una guia operativa para configurar conexion DIAN, firma digital, prueba 2+2+2, set completo, acuse final y activacion de produccion local.
- [Navegacion] `facturacion_electronica_menu.html` incorpora `Tutorial DIAN` para Colombia con permiso de lectura de facturacion y conserva `empresa_id` al abrir configuracion o pruebas.
- [Seguridad] El tutorial evita publicar secretos reales y orienta a usar la consola DIAN saneada como evidencia operativa.

## [2026-06-06] Set DIAN completo por tandas reales
- [Frontend] `facturacion_electronica_pruebas_dian.html` ejecuta el set automatico completo por tandas reales de 3 documentos cuando no se define limite manual, evitando el `504 Gateway Time-out` del proxy.
- [DIAN] Cada tanda sigue usando `action=pruebas_dian`, TestSetId, firma y envio real a habilitacion; no se reintroduce simulacion.
- [UX] La pantalla acumula procesados, resumen, detalles y lotes en la consola DIAN para cerrar 30 facturas, 10 notas debito y 10 notas credito sin perder evidencia.

## [2026-06-06] Resumen portal DIAN en pruebas
- [Frontend] `facturacion_electronica_pruebas_dian.html` muestra una ficha espejo del portal DIAN con contribuyente, TestSetId, modo de operacion, rango, software, URL, documentos requeridos y aceptados requeridos.
- [Seguridad] PIN y clave tecnica se muestran enmascarados por defecto y solo se revelan con accion explicita en pantalla.
- [Operacion] La pagina permite comparar rapidamente los datos guardados contra el set de pruebas asignado por DIAN antes de ejecutar 2+2+2 o el set completo.

## [2026-06-06] Consola visible de respuestas DIAN
- [Frontend] `facturacion_electronica_pruebas_dian.html` agrega una consola DIAN con historial visible de configuracion, validacion, diagnostico, envios reales, conexion, cola y operaciones documentales.
- [Seguridad] La consola sanea claves, PIN, tokens, certificados y llave tecnica antes de mostrar o copiar respuestas.
- [UX] La consola incluye estado HTTP, duracion, hora, auto-scroll, limpiar y copiar para soporte operativo sin depender de F12.

## [2026-06-06] Documentacion DIAN habilitacion real
- [Documentacion] Se actualiza el contexto operativo, mapa de modulos, flujos, decisiones tecnicas y estructura BD para reflejar que el transporte SOAP/WCF real ya responde con TrackId/ZipKey.
- [DIAN] Se documenta la forma WS-Security vigente: `BinarySecurityToken`, firma de `wsa:To`, `wsse:Reference URI="#X509-..."` e `InclusiveNamespaces`.
- [Operacion] Se aclara que `Batch en proceso de validacion` es estado pendiente y que falta reconciliar `GetStatusZip` hasta acuse final aceptado/rechazado antes de habilitar produccion local.
- [Seguridad] La documentacion evita copiar PIN, claves tecnicas, certificados, contrasenas o tokens.

## [2026-06-06] DIAN WS-Security segun politica WSDL
- [Backend] `SendTestSetAsync` y `GetStatusZip` ajustan la firma WS-Security a una forma compatible con clientes DIAN validados: firma solo `wsa:To`, referencia directa al `BinarySecurityToken` y `InclusiveNamespaces` en `SignedInfo`.
- [Backend] El digest de `wsa:To` incluye los namespaces `soap`, `wcf`, `wsa` y `wsu` de la canonicalizacion esperada por WCF/DIAN.
- [QA] Prueba visual real contra empresa 12 confirmo que la variante anterior alcanzaba DIAN pero recibia SOAP Fault `InvalidSecurity`; esta variante se basa en ejemplos GitHub de transporte SOAP DIAN aceptado en sandbox.
- [QA] `go test ./handlers -run "DIAN|Dian|FacturacionColombia|FacturaElectronicaVenta|NormalizeFacturacionDocumento|ResolveFacturacionTransition|FacturacionPermissions" -count=1`; `go test ./db -run "Dian|DIAN|Facturacion" -count=1`; `go test ./... -run "^$" -count=1`.

## [2026-06-06] Botones DIAN y conexion oficial
- [Frontend] `facturacion_electronica_pruebas_dian.html` autogenera codigos para los botones de emision manual cuando el campo esta vacio y muestra errores JSON reales del backend.
- [Backend] `estado_conexion_dian` usa la configuracion DIAN Colombia oficial de la empresa cuando existe, probando el endpoint SOAP guardado en lugar de depender solo del proveedor FE generico.
- [QA] `go test ./handlers -run "DIAN|Dian|FacturacionColombia|FacturaElectronicaVenta|NormalizeFacturacionDocumento|ResolveFacturacionTransition|FacturacionPermissions" -count=1`; `go test ./db -run "Dian|DIAN|Facturacion" -count=1`; `go test ./... -run "^$" -count=1`.

## [2026-06-06] DIAN WS-Security BinarySecurityToken
- [Backend] `SendTestSetAsync` y `GetStatusZip` oficiales referencian el `BinarySecurityToken` en `KeyInfo`, como pide la guia DIAN SoapUI, en lugar de `ThumbprintSHA1`.
- [DIAN] El timestamp WS-Security usa vigencia de 60 segundos y precision en milisegundos; se mantiene firma RSA-SHA256, digest SHA-256 y firma del header `wsa:To`.
- [QA] `go test ./handlers -run "DIAN|Dian|FacturacionColombia|FacturaElectronicaVenta|NormalizeFacturacionDocumento|ResolveFacturacionTransition|FacturacionPermissions" -count=1`; `go test ./... -run "^$" -count=1`.

## [2026-06-06] DIAN set real sin simulacion
- [Backend] `action=pruebas_dian` y `action=enviar_set_pruebas` rechazan `simular=true`; el flujo automatico ahora exige envio real.
- [DIAN] El set real consulta `GetStatusZip` despues de recibir `ZipKey` y solo cuenta documentos aceptados con acuse final.
- [Seguridad] `pruebas_dian`, `enviar_set_pruebas`, `firmar_xml_xades_base`, `validar_documento_dian` y `activar_produccion_local` pasan a accion de aprobacion en permisos de facturacion.
- [Frontend] `facturacion_electronica_pruebas_dian.html` elimina el checkbox de simulacion y fuerza envio real al ambiente de habilitacion.
- [QA] Se reemplaza la prueba 2+2+2 simulada por un servidor SOAP local estricto que valida `SendTestSetAsync`, ZIP firmado y `GetStatusZip` aceptado.

## [2026-06-05] DIAN oficial sin token proveedor obligatorio
- [Backend] La validacion de credenciales DIAN distingue `*.dian.gov.co` de endpoints de proveedor/API: `token_emisor_ref` es opcional para SOAP/WCF oficial y obligatorio para proveedor bearer.
- [Operacion] `test_set_id` sigue siendo obligatorio para envios reales de habilitacion y se muestra como el faltante real cuando no esta configurado.
- [Frontend] `facturacion_electronica_pruebas_dian.html` agrega el boton `Enviar prueba 2 + 2 + 2`.
- [QA] Pruebas unitarias de credenciales DIAN y sintaxis del script embebido de la pantalla de pruebas.

## [2026-06-05] Selector de empresas ordenable
- [UX] `seleccionar_empresa.html` permite mover tarjetas de empresas con clic sostenido en PC o usando el asa en celular.
- [Alcance] Funciona por separado en empresas con licencia activa y empresas sin licencia activa, conservando esos grupos.
- [Persistencia] `/api/user/configuracion` guarda `selector_empresas_orden` por administrador en `usuario_configuracion.selector_empresas_orden_json`, con respaldo local si la red falla.
- [Seguridad] El orden solo aplica a empresas que `/super/api/empresas` ya autorizo para la sesion; no concede acceso ni mezcla empresas de otros administradores.
- [QA] Prueba visual local con empresas simuladas confirma mover activas, mover inactivas, recargar con persistencia y `Restablecer orden`.

## [2026-06-05] Centro de habilitacion DIAN
- [Facturacion electronica] `facturacion_electronica_pruebas_dian.html` queda como pantalla de `Pasar test DIAN`, con estado de alistamiento, validacion de credenciales, diagnostico, objetivo del set y botones para envio automatico o por tipo documental.
- [BD] `empresa_dian_configuracion` guarda modo de operacion, fechas del set y totales requeridos/aceptados por tipo documental para ajustar la habilitacion a lo que muestra el portal DIAN de cada empresa.
- [Operacion] La empresa interna `Powerful Control System` queda configurada con el set DIAN asignado en el portal; la documentacion solo confirma que los secretos estan registrados, no los imprime.

## [2026-06-05] Llave tecnica DIAN por empresa
- [Facturacion electronica] `empresa_dian_configuracion` agrega `llave_tecnica` para guardar la clave tecnica del rango de numeracion DIAN sin mezclarla con `token_emisor_ref`.
- [Frontend] La configuracion DIAN Colombia muestra y guarda `Llave tecnica DIAN` junto al rango autorizado.
- [Operacion] Se registro en produccion el set de habilitacion DIAN de la empresa interna `Powerful Control System` con software propio, prefijo SETP, resolucion y rango de prueba.

## [2026-06-05] OnlyOffice corrige token JWT del editor
- [Backend] `editor_config` firma la configuracion completa de ONLYOFFICE y deja el JWT solo en `config.token`, evitando enviar tokens duplicados dentro de `document` o `editorConfig`.
- [Compatibilidad] Corrige el error visual `The document security token is not correctly formed` al crear y abrir documentos Word/Excel/PowerPoint.
- [QA] `go test ./handlers -run OnlyOffice -count=1`; comprobacion visual autenticada contra Document Server antes del despliegue reprodujo el fallo de token.

## [2026-06-05] Licencias acumuladas por pago repetido
- [Licencias] La activacion de Epayco, Wompi y activacion manual con descuento total devuelve y registra el ID real de la licencia empresarial creada o extendida.
- [Pagos] Las renovaciones comerciales simples conservan la regla acumulativa: la nueva vigencia inicia en la fecha fin mas lejana ya activa/futura de la empresa y suma la duracion del plan pagado.
- [QA] Se agrega prueba para dos pagos del mismo plan mensual de 30 dias, confirmando 60 dias acumulados.

## [2026-06-05] Ultima carga de firma DIAN segura
- [Facturacion electronica] La tarjeta `Cargar firma electronica (Colombia / DIAN)` muestra fecha de ultima carga, archivo, formato, titular, serial y estado de clave sin exponer la contrasena.
- [Compatibilidad] La carga de P12/PFX agrega fallback interno `ToPEM` y respaldo OpenSSL en el contenedor para certificados modernos con multiples bolsas, cadenas o cifrado no soportado por el lector Go simple.
- [Seguridad] La clave del P12/PFX se usa solo para decodificar la firma; no se guarda ni se muestra en claro.
- [BD] `empresa_dian_configuracion` agrega metadatos seguros de ultima firma cargada por `empresa_id`.
- [QA] Validacion de script embebido de `facturacion_electronica.html`; pruebas Go enfocadas de DIAN/facturacion en handlers y db.

## [2026-06-05] Apariencias de Camaras, Energia solar y Empresas compartidas
- [UX] `web/administrar_empresa/camaras.html` deja de usar estilos inline con colores fijos y hereda la apariencia global del sistema.
- [Temas] `web/estilos.css` normaliza fondos, tarjetas, formularios, estados, bordes y textos para Camaras, Energia solar y Empresas compartidas en temas claros y oscuros.
- [QA] Validacion sintactica de `web/js/camaras.js` y `web/js/energia_solar.js`; comprobacion visual local en claro y oscuro absoluto de las tres paginas.

## [2026-06-05] Pasarelas de licencia Epayco/Wompi
- [Pagos] La disponibilidad publica de Epayco vuelve a resolver credenciales legacy y opcionales sin apagar toda la pasarela si una clave opcional cifrada no se puede descifrar; el fallback `checkout.js` se habilita con Public Key valida sin exigir `P_KEY`.
- [Pagos] Wompi solo se publica como disponible cuando la llave publica y la llave de integridad son realmente legibles para Web Checkout.
- [QA] `go test ./handlers -run "EpaycoCheckoutCredential|DefaultLicenciaPayment|Wompi|PaymentCredential" -count=1`.

## [2026-06-05] Tutorial de nomina con narracion
- [Ayuda] Se agrega `web/ayuda/tutorial_nomina.html` como presentacion guiada de nomina con pasos operativos, tarjetas visuales y narraciones.
- [UX] Cada bloque `Narracion` tiene boton con icono de play para reproducir la guia por voz cuando el navegador soporte sintesis de voz.
- [Nomina] `web/administrar_empresa/nomina_sueldos.html` agrega un boton `Ayuda` con icono que abre el tutorial conservando `empresa_id`.
- [QA] Validacion sintactica de scripts embebidos y comprobacion visual local de los botones de play y del enlace desde nomina.

## [2026-06-05] Carpetas empresariales para firma electronica
- [Archivos] Cada empresa asegura carpeta base `web/uploads/empresas/empresa_{id}_{slug}/` al crearse o resolverse por idempotencia.
- [Facturacion electronica] La carga de firma DIAN extrae y guarda el vencimiento X.509, agrega verificacion visible en la pantalla y envia aviso al administrador cuando el certificado esta vencido o proximo a vencer.
- [Facturacion electronica] La carga de firma DIAN guarda llave privada y certificado en `facturacion_electronica/firma_electronica/` dentro de la carpeta empresarial, con permisos privados y referencias internas `file:`.
- [Seguridad] La eliminacion segura de empresa limpia tambien `web/uploads/empresas/empresa_{id}_*` para no dejar firmas o imagenes huerfanas.
- [QA] Pruebas Go enfocadas para decodificacion DIAN y convencion de ruta de firma electronica.

## [2026-06-05] Empresa interna Powerful Control System operativa
- [Licencias] La empresa interna del SaaS conserva el codigo tecnico `PCS_SYSTEM_INTERNAL_PERPETUAL`, pero ahora se asegura como licencia fechada a 100 anos, con valor cero, limites altos y modulos completos.
- [Licencias] Las consultas PostgreSQL de licencias vigentes/vencidas toleran fechas heredadas vacias o no fechables para que una licencia antigua no bloquee permisos, carrito, correo ni reportes.
- [Permisos] El rol `super_administrador` validado en backend puede acceder globalmente a empresas para soporte y operacion interna, sin quitar los filtros por `empresa_id` de cada endpoint.
- [Operacion] `Powerful Control System` debe cargar carrito, correo corporativo, facturacion, configuracion y reportes como cualquier empresa; la unica diferencia es que la licencia no pertenece al catalogo comercial.
- [Portal] El respaldo editable de `Modulos y caracteristicas principales` agrega GRAFOLOGIX y Camaras/DVR, y completa esos modulos cuando la configuracion guardada es antigua.
- [QA] Pruebas enfocadas en `licencias_empresa_sistema.go`, permisos de licencia y validacion visual del carrito/correo de la empresa interna.

## [2026-06-04] Instalar app en login y login usuario
- [PWA] `web/js/pwa_install.js` prepara el service worker, espera el evento nativo de instalacion, consume el prompt una sola vez como exige Chrome/Edge y evita mostrar de inmediato el mensaje de instalacion manual cuando el navegador aun no entrega el prompt.
- [UX] El boton `Instalar app` conserva el mismo flujo en `login.html` y `login_usuario.html`; si el usuario ya escribio correo o contrasena, no se fuerza recarga para no perder datos.
- [QA] Validado con `node --check web/js/pwa_install.js`, `git diff --check` y prueba visual local de ambos botones usando `?qa_pwa=1`.

## [2026-06-04] Modulo Camaras y DVR
- [Backend] Se agrega `/api/empresa/camaras` con tabla `empresa_camaras`, CRUD por `empresa_id`, catalogo RTSP/ONVIF/HLS/WebRTC/MJPEG/iframe y baja logica.
- [Permisos] Nuevo modulo `camaras`, pagina `linkCamaras`, wrapper `WithEmpresaCamarasPermissions`, fallback de licencia a `control_electrico`/`seguridad` y cobertura en planes enterprise.
- [Frontend] Nueva pagina `Administrar empresa > Analisis y control > Camaras`; Configuracion de estaciones permite mostrar camaras antes/despues y marcar estaciones como tipo camara.
- [Estaciones] El tablero renderiza visores de camara y enlaza al modulo sin pasar por carrito.
- [QA] `go test ./... -run "EmpresaRoutesUsePermissionWrappers|Camaras" -count=1`, `node --check web/js/camaras.js`, validacion de scripts inline y prueba visual local con datos simulados.

## [2026-06-04] Fallback Mailu para correos de licencia
- [Licencias] El correo unificado de activacion de licencia intenta primero Gmail SMTP y, si la credencial no descifra o el envio falla, usa Mailu corporativo como canal de respaldo.
- [Correo] El fallback conserva el mismo asunto, cuerpo, PDF de licencia y adjuntos de factura cuando corresponda, sin exponer secretos en errores o logs.
- [QA] Pruebas Go enfocadas de licencias/pagos y prueba visual real del checkout con descuento total y activacion directa.

## [2026-06-04] Checkout de licencia con correo cliente
- [Checkout] `pagar_licencia.html` agrega `Correo del cliente` y lo envia a activacion directa, Wompi y Epayco.
- [Licencias] `activar_sin_pago` valida `customer_email` antes de activar y lo usa para enviar el correo con PDF de licencia cuando el total queda en cero.
- [Codigos] Un codigo de descuento ya usado por la empresa devuelve mensaje bloqueado sin provocar 500 en el resumen publico.
- [QA] Pruebas Go enfocadas de licencias/pagos y validacion sintactica del script embebido de checkout.

## [2026-06-04] Correo unificado de licencia y factura electronica
- [Licencias] La activacion de licencia envia un solo correo al cliente con el PDF de licencia; si la compra comercial aprobada tiene valor mayor que cero, adjunta en ese mismo mensaje el PDF resumen de la factura electronica.
- [Pagos] Epayco y Wompi ya no disparan un correo separado de factura en el flujo normal; conservan la marca `licencia_factura_electronica_emitida` para idempotencia.
- [Regla fiscal] Las activaciones con total pagado cero por prueba o descuento total no emiten factura electronica en el flujo final.
- [QA] `go test ./handlers -run "Licencia|Factura|Payment|Epayco|Wompi" -count=1`; `go test ./db -run "Licencia|Payment|Factura" -count=1`.

## [2026-06-04] Historial de licencias separado
- [Licencias] `Seleccionar empresa > Licencias` muestra solo licencias activas o por vencer, ocultando las vencidas que ya fueron reemplazadas por una nueva licencia vigente.
- [UX] Se agrega `Historial de licencias` como pagina independiente para consultar todas las licencias visibles del administrador, activas y vencidas desde el inicio.
- [Seguridad] No se crean endpoints nuevos; la vista reutiliza `/super/api/licencias?scope=mine&con_empresa=1` con el alcance multiempresa existente.
- [QA] Scripts embebidos validados con Node y prueba visual local con datos simulados de licencia activa mas vencida.

## [2026-06-04] Factura electronica automatica por compra de licencia
- [Licencias] Las compras comerciales aprobadas generan factura electronica desde la empresa interna `Powerful Control System` y envian el documento al correo del cliente.
- [Backend] Se agrega `backend/db/licencias_empresa_sistema.go` para resolver la empresa existente, incluyendo `Powerful Control Systen`, y asegurar licencia tecnica interna. Desde 2026-06-05 esa licencia se normaliza con vigencia fechada de 100 anos y modulos completos.
- [Pagos] Epayco/Wompi marcan `licencia_factura_electronica_emitida` en el payload del pago para evitar facturas duplicadas por reintentos o webhooks repetidos.
- [QA] `go test -p 1 ./db ./handlers -run "Powerful|Licencia|Facturacion" -count=1`.

## [2026-06-02] Medidas tecnicas completas en reportes GRAFOLOGIX
- [Grafologia] Cada metrica del motor Go ahora guarda `details` con angulo de inclinacion, pendiente, altura de renglones, separacion entre letras/palabras/lineas, continuidad, direccion de linea base, margenes, densidad, regularidad y forma.
- [Reportes] HTML, Word, TXT, CSV y PDF muestran esas medidas; el PDF de GRAFOLOGIX ahora pagina el contenido para no recortar el detalle tecnico.
- [UX] La tabla de metricas de `grafologia.html` muestra las medidas debajo del resultado de cada metrica.
- [QA] `go test ./internal/grafologia -count=1`, `node --check web/js/grafologia.js` con Node local y captura visual Chrome del reporte generado en `tmp/grafologia-visual/reporte_metricas.png`.

## [2026-06-01] Edicion rapida de roles de usuarios empresariales
- [Usuarios] `web/administrar_empresa/administrar_usuarios.html` agrega la tarjeta `Editar rol de usuario` para que administradores de empresa cambien el rol de un usuario sin abrir todo el formulario.
- [UX] Cada fila de usuarios agrega el boton `Cambiar rol`, que selecciona el usuario en la tarjeta, muestra su rol actual y permite elegir un rol activo del catalogo global.
- [Seguridad] Se reutiliza `/api/empresa/usuarios` con `empresa_id` en payload y `id` por query; el backend mantiene validacion de alcance empresarial y permisos `seguridad:U`.
- [QA] Validacion sintactica del script embebido, `git diff --check` y prueba visual simulada con Chrome/Playwright: cambiar rol de usuario actualizo la tabla y mostro mensaje de exito.

## [2026-06-01] Ayuda operativa de GRAFOLOGIX
- [Ayuda] `web/ayuda/ayuda.html` agrega la seccion `GRAFOLOGIX - analisis grafológico` con flujo, motor Go, OCR libre, GPT-5.5, exportaciones, buenas practicas y permisos.
- [UX] `web/administrar_empresa/grafologia.html` agrega el enlace `Ayuda GRAFOLOGIX` apuntando directamente a la seccion de ayuda del modulo.
- [QA] Validacion sintactica del script embebido de ayuda, parseo HTML basico y verificacion visual local del ancla y del enlace desde GRAFOLOGIX.

## [2026-06-01] Diagnostico de invitaciones de usuarios por correo
- [Usuarios] Al crear usuarios empresariales, el sistema conserva el enlace de confirmacion y reporta claramente cuando el correo no puede salir por una credencial SMTP guardada que no se puede descifrar.
- [Super] La configuracion Gmail SMTP muestra si la contrasena SMTP almacenada no descifra con la clave actual del servidor, y la prueba de correo devuelve un error accionable en vez de un 500 generico.
- [QA] Prueba real de creacion de usuario en empresa 7: el usuario se creo, `email_sent=false` y la causa fue `cipher: message authentication failed`; el usuario de prueba fue eliminado. `go test ./handlers -run Gmail -count=1` y validacion sintactica de scripts embebidos.

## [2026-06-01] Campos obligatorios de productos movidos a configuracion
- [UX] La tarjeta `Campos obligatorios para productos` sale de `Administrar Productos` y queda en `Configuracion > Productos y pedidos`.
- [Frontend] `administrar_productos.html` conserva la lectura de la configuracion para marcar campos obligatorios en el formulario, pero ya no muestra ni guarda la tarjeta desde el modulo de productos.
- [QA] Validacion sintactica de scripts embebidos y verificacion visual estatica: configuracion muestra 20 checks y productos ya no muestra la tarjeta.

## [2026-06-01] Propietario conserva empresas compartidas en selector
- [Selector] `/super/api/empresas` ahora incluye tambien las empresas que el administrador autenticado compartio con otro administrador, aunque el `usuario_creador` historico no coincida.
- [Seguridad] `CanAdminAccessEmpresaIA` reconoce esa misma relacion activa como alcance valido del propietario que compartio, sin ampliar el acceso a empresas sin invitacion o sin relacion.
- [QA] `go test ./handlers -run "DecorateEmpresasByEffectiveAccess|AdminEmpresaCompartida|SelectorAdminScope" -count=1` y `go test ./db -run "AdminEmpresaCompartida|CanAdminAccessEmpresaIA" -count=1`.

## [2026-06-01] Codigos de descuento para licencias en super administrador
- [Super] Nueva pagina `web/super/licencias_codigos_descuento.html` en Comercial y licencias para crear, editar, activar/desactivar y eliminar codigos globales de descuento de licencias.
- [Backend] Nuevo endpoint auditado `/super/api/licencias/codigos_descuento` que administra `configuraciones.licencias.discount_codes` con formatos `CODIGO=10%`, `CODIGO=50000` y `CODIGO=gratis`.
- [Checkout] `pagar_licencia.html` ahora precarga `discount_code` o `codigo_descuento` recibido por URL y mantiene la validacion existente de un uso por empresa.
- [QA] `go test ./handlers -run "LicenciaDiscountCode|LicenciasCodigos" -count=1`, `go test ./db -run "LicenciaDiscount|LicenciasGratis" -count=1` y validacion sintactica del JS embebido de checkout y la nueva pagina.

## [2026-06-01] Super administrador puede compartir empresas
- [Backend] `/super/api/empresas/compartidos` permite crear, reenviar y revocar invitaciones/accesos cuando el actor autenticado es super administrador, aunque no sea el propietario `usuario_creador` de la empresa.
- [Seguridad] La excepcion queda limitada al rol super en backend; administradores compartidos o delegados normales siguen sin poder reencadenar empresas ajenas.
- [UX] El selector, `editar_empresa.js` y `empresas_compartidas.html` habilitan las acciones de compartir/gestionar para super administradores.
- [QA] `go test ./handlers -run "AdminEmpresaCompartida|EmpresaCompartida|SelectorAdminScope" -count=1`, validacion sintactica de JS y prueba visual local con empresa ajena.

## [2026-06-01] Nombre real de caja en estaciones
- [UX] La tarjeta especial de Caja en `estaciones.html` ya no muestra el titulo generico `Caja`; usa el codigo y nombre configurado de la caja activa, por ejemplo `CAJA-1 - Caja principal`.
- [UX] Si solo hay una caja activa, se oculta la lista redundante para no repetir `Caja principal`; si hay varias, se conserva la lista de seleccion.
- [QA] Validacion sintactica del script embebido y prueba visual local con caja configurada `CAJA-1 / Caja principal`.

## [2026-06-01] Renovaciones anticipadas de licencias
- [Licencias] Una licencia comercial pagada antes del vencimiento se agenda desde el vencimiento acumulado de la empresa y suma su duracion sin reemplazar la licencia actual.
- [Pagos] Epayco y Wompi guardan estado de activacion por referencia para que webhooks o consultas repetidas no dupliquen dias.
- [QA] `go test ./db -run "Licencia|Renovacion|Payment" -count=1` y `go test ./handlers -run "Licencia|Payment|Wompi|Epayco" -count=1`.

## [2026-06-01] Checkout de licencias con Epayco y Wompi
- [Pagos] `/api/public/licencias/payment_methods` publica Epayco y Wompi por defecto cuando cada proveedor tiene credenciales completas; los `*.enabled=0` siguen apagando explicitamente una pasarela.
- [UX] `web/pagar_licencia.html` conserva ambos metodos visibles y `web/estilos.css` centra/agranda los logos Davivienda y Bancolombia en las tarjetas de pago.
- [QA] Prueba visual del flujo seleccionando Epayco y Wompi, aceptando terminos y llegando a verificacion `PENDING`; prueba Go enfocada para disponibilidad de pasarelas.

## [2026-05-31] Descarga PDF de licencia desde empresa
- [Menu] `web/administrar_empresa.html` agrega el grupo principal `Licencia` y deja alli `Licencia del sistema`.
- [Licencias] `web/administrar_empresa/licencia_sistema.html` reemplaza descargas TXT/HTML por un enlace directo para descargar el PDF oficial de la licencia de la empresa.
- [Backend] `/api/empresa/licencia_sistema/pdf` genera el PDF con el formato `licencia_software_pdf` editable desde Super administrador y queda protegido por permisos `linkLicenciaSistema`.
- [QA] `go test ./handlers -run "LicenciaSoftwarePDF|LicenciaActivationEmailMessage|LicenciaSistemaPDF|Licencia" -count=1`, `go test . -run TestNoExiste -count=1` y validacion sintactica de scripts inline.

## [2026-05-31] PDF de licencia de software por correo
- [Licencias] El correo de licencia activada ahora adjunta automaticamente un PDF de licencia de software cuando el pago queda aprobado o la activacion sin pago permitida queda vigente.
- [Super] `web/super/formato_para_emviar_email.html` agrega la plantilla `licencia_software_pdf` para editar el texto del PDF desde Super administrador.
- [Backend] `backend/handlers/licencias_pdf.go` genera el documento en Go puro, sin dependencias nuevas, y `payments_handlers.go` lo adjunta en MIME multipart.
- [QA] `go test ./handlers -run "LicenciaSoftwarePDF|LicenciaActivationEmailMessage|Licencia" -count=1`.

## [2026-05-31] Catalogo base de roles empresariales
- [Roles] Se agregan roles universales para usuarios de empresa: `supervisor_sucursal`, `vendedor`, `recepcion`, `jefe_bodega`, `recursos_humanos` y `tecnico_solar`, ademas de los roles ya existentes.
- [Solar] `tecnico_solar` solo puede consultar estado, lecturas, eventos y alertas de energia solar.
- [Bodega] `jefe_bodega` administra inventario, bodegas, existencias y traslados sin eliminar inventario ni operar ventas/caja.
- [RRHH] `recursos_humanos` gestiona horarios, asistencia y nomina operativa sin entrar a ventas, caja ni configuracion general.
- [UX/Seguridad] Los roles especializados reciben menus reducidos y las restricciones se aplican tambien en backend despues de overrides.

## [2026-05-31] Rol Servicio de limpieza para estaciones sucias
- [Roles] Las preconfiguraciones de tipos de empresa agregan `servicio_limpieza` con descripcion operativa clara.
- [Permisos] `servicio_limpieza` queda limitado a `ventas:R` para ver estaciones; no puede activar estaciones, abrir carrito, ver items, caja, ventas directas, inventario, reportes ni configuracion.
- [Aseo] El rol puede finalizar el aseo de estaciones marcadas como sucias mediante `/api/empresa/estacion_aseo?action=finalizar`, registrando usuario, duracion y cambio a limpia.
- [UX] En Administrar empresa solo ve `Estaciones`; al hacer clic en una estacion sucia la reporta limpia, y en estaciones limpias solo muestra aviso sin operar.
- [QA] Pruebas enfocadas de permisos, restricciones de carrito y flujo visual de estaciones.

## [2026-05-31] Rol empresario para resultados y reportes
- [Roles] Las preconfiguraciones de tipos de empresa agregan `empresario` como rol base.
- [Permisos] `empresario` queda limitado a `reportes:R`; no tiene ventas, caja, inventario, finanzas, facturacion, usuarios ni configuracion.
- [UX] El menu empresarial muestra solo `Reportes ejecutivos` y el submenu de reportes deja visible el centro de reportes, ocultando reportes de turnos/caja.
- [QA] `go test ./handlers -run 'Portero|Contador|Empresario' -count=1`, `go test ./db -run 'Portero|Contador|Empresario' -count=1`, sintaxis JS, parseo de `reportes_menu.html`, `git diff --check` y validacion visual local.

## [2026-05-31] Rol operativo contador
- [Roles] Las preconfiguraciones de tipos de empresa agregan `contador` como rol base.
- [Permisos] `contador` queda limitado a lectura de `finanzas` y `facturacion` para consultar finanzas e impuestos, sin acciones de creacion, edicion, eliminacion o aprobacion.
- [UX] El menu empresarial muestra solo `Centro financiero y contable` e `Impuestos`; el submenu financiero oculta accesos rapidos fuera de alcance.
- [QA] `go test ./handlers -run 'Portero|Contador' -count=1`, `go test ./db -run 'Portero|Contador' -count=1`, sintaxis JS del menu y parseo de `finanzas_menu.html`.

## [2026-05-31] Rol operativo portero
- [Roles] Las preconfiguraciones de tipos de empresa agregan `portero` como rol base.
- [Permisos] `portero` queda limitado a ver estaciones y ejecutar `activar_estacion`; no puede abrir carrito, items, pagos, caja, corte, venta directa ni configuraciones.
- [UX] En Administrar empresa el menu de `portero` muestra solo `Estaciones`, y la pantalla de estaciones activa sin abrir carrito.
- [QA] `go test ./handlers -run Portero -count=1`, `go test ./db -run Portero -count=1`, validacion de sintaxis JS del menu y parseo del script inline de estaciones.

## [2026-05-31] Feedback sonoro y tactil del carrito
- [Configuracion] `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html` agrega checks independientes para pitido y vibracion de botones en PC y celular.
- [Carrito] `web/administrar_empresa/carrito_de_compras.html` aplica pitido con Web Audio y vibracion visual/fisica segun la configuracion `carrito_ui_global`.
- [QA] Validado con parseo de scripts, `git diff --check` y prueba visual local guardando los checks y simulando clics en PC/celular.

## [2026-05-31] Instalacion PWA del login administrativo
- [PWA] `web/login.html` actualiza la version del instalador para evitar cache antiguo en el boton `Instalar app`.
- [UX] `web/js/pwa_install.js` muestra estado inmediato al abrir la instalacion y conserva un mensaje de respaldo si el navegador no entrega el dialogo nativo.
- [QA] Validado con sintaxis JS y prueba visual local simulando el evento `beforeinstallprompt`.

## [2026-05-31] Login usuario con Google por invitacion
- [Auth] `login_usuario.html` agrega `Iniciar sesión con Google` para usuarios operativos, usando `/auth/google/usuario/login` y el callback canonico `/auth/google/callback`.
- [Seguridad] El callback de usuario no crea cuentas publicas: solo abre sesion si el correo verificado por Google coincide con una invitacion vigente o un usuario empresarial ya confirmado, y redirige a `administrar_empresa.html?id={empresa_id}`.
- [UX/PWA] La pagina conserva apariencia clara/oscura del login de usuario, centra `Acceso de usuarios operativos`, elimina el texto contextual de empresa y agrega boton `Instalar app` con icono.
- [QA] `node --check web/js/login_usuario.js`, `go test ./handlers ./db ./utils -run "Google|Usuario|Auth|Middleware|EmpresaUsuario|LicenciasGratis" -count=1`, `go test ./... -run "^$" -count=1` y validacion visual local con CSS real en movil.

## [2026-05-31] Login con titulo texto e iconos PWA
- [UX] `web/login.html` reemplaza el logo imagen del encabezado por el texto `Powerful Control System`.
- [PWA] Los botones `Instalar app` e `Ir al inicio` muestran iconos propios y el instalador conserva el icono cuando cambia el texto del estado.
- [QA] Validado con Chrome/Playwright en escritorio y movil: iconos visibles, sin desborde movil y click de instalacion simulada con mensaje `Instalacion iniciada`.

## [2026-05-31] Licencia de prueba historica bloqueada
- [Backend] `backend/db/licencias_gratis.go` valida la prueba gratis por historial de empresa, no por vigencia actual, para impedir que una empresa vuelva a activar la prueba de 15 dias despues de vencida o inactiva.
- [Compatibilidad] El fallback detecta licencias gratis antiguas por valor cero, duracion 15 dias o textos `prueba`/`gratis`/`trial`, aunque no exista una marca nueva en `licencias_activaciones_gratis`.
- [QA] `go test ./db ./handlers -run "LicenciasGratis|LicenciaGratis|LicenciaPrueba|CheckoutSummary|ActivateLicenciaSinPago|IsLicenciaPrueba15DiasCatalogo" -count=1`.

## [2026-05-31] Super administrador simplificado
- [UX] `web/super_administrador.html` elimina del grupo Gobierno los accesos `Reportes globales` y `Metricas de trafico`.
- [Navegacion] `web/js/super_administrador.js` deja de permitir esas rutas como paginas favoritas o restaurables del panel super.
- [Limpieza] Se elimina `web/super/metricas_de_trafico_general.html`; las metricas siguen consolidadas en `Centro de mando`.
- [Alcance] `Reportes globales` se conserva para abrirse desde `seleccionar_empresa.html`, con el mismo alcance de empresas visibles.

## [2026-05-31] Licencias fijas globales
- [Backend] `backend/db/licencias_globales.go` asegura solo cuatro planes globales: prueba gratis 15 dias, COP 60000, COP 100000 y COP 150000.
- [Datos] La limpieza de catalogo elimina licencias sobrantes sin empresa asignada, incluidos addons antiguos y duplicados de catalogo; las licencias asignadas a empresas se conservan para historial y pagos.
- [Super] `web/super/licencias.html` muestra el catalogo como fijo y retira acciones de agregar, eliminar u ocultar licencias.
- [QA] `go test ./db -run "GlobalLicencia|LicenciaCatalog" -count=1` y `go test ./handlers -run TestLicencia -count=1`.

## [2026-05-31] Centro de mando super simplificado
- [UX] `web/super/licencias_resumen.html` quita la tarjeta superior `Super administrador / Panel ejecutivo` y elimina la tarjeta de clima del panel super.
- [Interaccion] Los botones `Actualizar` y `Evaluar alertas` se conservan dentro de `Estado general`, evitando que la pagina pierda controles operativos.
- [Orden visual] La primera tarjeta visible ahora es `Favoritos`, seguida por `Estado general`, `Incidentes recientes` y `Prioridades`.

## [2026-05-31] Clima en centro de mando super
- [UX] `web/super/licencias_resumen.html` agrega una tarjeta de clima actual en la parte alta del panel super, antes de favoritos y Estado general.
- [Interaccion] La tarjeta permite usar GPS o escribir ciudad manualmente; guarda la ubicacion en `localStorage:super_admin:weather_location` y consulta Open-Meteo sin agregar dependencias.
- [Orden visual] La pagina queda en secuencia: encabezado, clima, favoritos, Estado general, Incidentes recientes y Prioridades.
- [QA] Parseo de scripts embebidos y prueba visual local con Playwright validando clima, favoritos y orden vertical.

## [2026-05-31] Favoritos de super administrador
- [UX] `web/super_administrador.html` agrega un boton estrella global para marcar o quitar de favoritos cualquier pagina valida del modulo super que se abra en el iframe.
- [Panel] `web/super/licencias_resumen.html` muestra arriba una seccion de favoritos con el icono original del menu, compacta Estado general e Incidentes recientes y elimina la tarjeta Accesos clave.
- [Integracion] `web/js/super_administrador.js` guarda los favoritos en `localStorage` bajo `super_admin:favorites`, refresca el centro de mando por `postMessage` y registra eventos UI en auditoria super.
- [QA] Validacion de sintaxis JS, `git diff --check` y prueba visual local con Playwright: marcar Centro de mando y Alertas sistema, ver dos favoritos, confirmar que Accesos clave no aparece.

## [2026-05-30] Configuracion general PostgreSQL reforzada
- [Backend] `EnsureEmpresaConfiguracionGeneralSchema` asegura la columna `fecha_creacion` en tablas existentes y repara la secuencia/default de `id` con `ensurePostgresTableIDSequence`.
- [Compatibilidad] Las fechas de auditoria de `empresa_configuracion_general` se guardan como texto con `CAST(CURRENT_TIMESTAMP AS TEXT)` para no depender de conversiones implicitas de PostgreSQL.
- [Bugfix] El alta automatica de configuracion por defecto incluye el valor de `clima_fuente`, evitando que empresas nuevas fallen por diferencia entre columnas y valores del `INSERT`.
- [Pagos] El checkout de licencias solo publica Wompi cuando `wompi.public_key` tiene formato `pub_test_` o `pub_prod_`; una llave placeholder queda marcada como configuracion incompleta y no dispara `/wompi/terms`.
- [Pagos] Si `wompi.mode` manual contradice el prefijo real de la llave, el runtime usa el modo inferido por la llave para evitar consultar sandbox con llave `pub_prod_` o produccion con llave `pub_test_`.
- [QA] Pruebas enfocadas en `db` y `handlers` para configuracion general y checkout/licencias.

## [2026-05-30] Informacion de contacto corporativa
- [Portal] `web/Informacion_de_contacto.html` elimina el logo del hero y cambia el contenido hacia una presentacion corporativa de Powerful Control System.
- [Contenido] La pagina agrega vision: dar a pequenas empresas acceso a herramientas tecnologicas para desarrollarse, progresar y competir con mas orden y control.
- [UX] `web/estilos.css` agrega layout de proposito, areas del sistema, vision, que hacemos, canales, asesor, acompanamiento y cierre comercial, validado en escritorio y movil.

## [2026-05-30] Domotica en paginas separadas
- [Navegacion] `web/administrar_empresa/modulo_menu.html` abre Domotica por vistas `pagina=resumen|conexion|raspberry|reles|automatizaciones|reportes|bitacora`, manteniendo el contexto `empresa_id` y la clave tecnica `control_electrico`.
- [UX] `web/administrar_empresa/control_electrico.html` ya no muestra el submenu interno duplicado dentro de `Resumen operativo`; cada vista renderiza solo su seccion funcional.
- [QA] Parseo de scripts embebidos con Node y prueba visual local con Playwright: cada boton del submenu carga la vista esperada, sin enlaces duplicados internos.

## [2026-05-31] Modo offline multi-caja
- [Carritos] La cola local de ventas offline ahora se separa por empresa y cajero.
- [Caja] La venta offline exige una caja abierta cargada para el cajero; ya no cae en una caja por defecto.
- [Backend] `/api/empresa/offline_ventas` valida propietario, `caja_codigo` e idempotencia de carritos ya pagados para evitar duplicados en cajas simultaneas.

## [2026-05-30] Ayuda y APIs actualizadas
- [Ayuda] `web/ayuda/ayuda.html` agrega seccion `Ayuda de APIs`, enlace directo a la nueva pagina de APIs y actualiza la ayuda de carrito/venta directa/offline.
- [Ayuda contextual] `web/ayuda/ayuda_contextual.html` separa la ayuda de carrito de la ayuda de Domotica y agrega respuesta contextual para APIs/OpenAPI/endpoints.
- [APIs] Nueva pagina `web/ayuda/ayuda_apis.html` y guia `documentos/api/ayuda_apis.md` con familias de endpoints, seguridad multiempresa, errores, carritos y checklist para integradores.
- [Docs] `documentos/api/openapi.generated.yaml`, `README.md` y el contrato de wrappers enlazan la nueva ayuda tecnica.

## [2026-05-30] Documentacion del carrito unificado
- [Docs] Se actualizan `contexto_codex.md`, `mapa_modulos.md`, `flujos_operativos.md`, `decisiones_tecnicas.md`, `descripcion_del_proyecto`, `diagramas/estructura_del_codigo.md` y `README.md` para dejar canonico el flujo del carrito: venta directa en pantalla completa, iframe con permiso `fullscreen`, fondo estructural mas oscuro y tarjetas planas sin sombras.
- [Alcance] No cambia codigo runtime, endpoints, tablas ni permisos; documenta el comportamiento ya desplegado por `rs`.

## [2026-05-29] Fondo diferenciado en carrito
- [UX] `web/estilos.css` separa visualmente el fondo estructural del carrito y las tarjetas reales usando variables de tema: el fondo queda mas oscuro que `surface` sin reintroducir sombras ni relieve.
- [Venta directa] `web/administrar_empresa/carrito_de_compras.html` usa el mismo fondo del carrito al entrar en pantalla completa.
- [QA] Validacion visual local con Playwright en escritorio y comprobacion de temas claro/oscuro: tarjetas con fondo `surface`, estructura con `--carrito-page-bg` y sombras desactivadas.

## [2026-05-29] Venta directa en pantalla completa
- [UX] `web/administrar_empresa/carrito_de_compras.html` agrega un boton con icono para abrir la venta directa en pantalla completa y regresar a la vista normal.
- [Integracion] `web/administrar_empresa.html` permite `fullscreen` en el iframe del panel para que la accion funcione tambien cuando venta directa se abre dentro de Administrar empresa.
- [QA] Validacion de sintaxis JS embebida y prueba visual local con Playwright: boton visible, entrada a fullscreen, texto `Salir` y retorno a vista normal.

## [2026-05-29] Panel empresarial sin titulo repetido
- [UX] `web/administrar_empresa/panel.html` cambia el encabezado superior del hero a `Centro operativo` para evitar repetir `Panel de {empresa}` encima de la descripcion.
- [Datos] No cambia endpoints, permisos, tablas ni configuracion.

## [2026-05-29] Auditoria profesional de Plantillas
- [QA] `tools/professional_audit.mjs` deja de leer la ruta antigua `web/js/nuevos_verticales_catalogo.js` y valida el catalogo activo `web/js/plantillas_nuevas_catalogo.js`.
- [Nomenclatura] El check cambia a `plantillas_nuevas_catalogo` y el resumen reporta `plantillas_nuevas`.
- [Verificacion] La auditoria profesional vuelve a terminar con `status: ok`.

## [2026-05-29] Seleccionar empresa con Plantillas
- [Frontend] `web/js/seleccionar_empresa.js` corrige referencias heredadas a la variable `vertical` y usa la variable `plantilla` al renderizar preview, tarjetas y opciones de tipo de empresa.
- [Super] `web/super/preconfiguracion_tipos_empresa.html` corrige el mismo caso en busquedas, secciones y filtros iniciales de plantillas.
- [QA] Validacion de sintaxis JS externa y scripts embebidos; busqueda de referencias `vertical.` sin declarar.

## [2026-05-29] Centro de mando super apilado
- [UX] `web/super/licencias_resumen.html` cambia el panel ejecutivo a una sola columna: layout principal, resumen, KPIs y accesos clave quedan apilados.
- [Responsive] Se elimina la redistribucion lateral en escritorio/tablet para mantener una lectura vertical consistente.
- [Datos] No cambia endpoints, permisos, tablas ni dependencias.

## [2026-05-29] Super administrador en una columna
- [Menu] `web/super_administrador.html` agrega `Asesor en ventas` dentro del grupo Licencias para abrir la configuracion comercial existente.
- [Navegacion] `web/js/super_administrador.js` permite restaurar `/super/asesor_comercial.html` como pagina valida del frame.
- [UX] `web/estilos.css` fuerza los grupos del panel super a una sola columna, incluyendo responsive, para evitar secciones una al lado de la otra.

## [2026-05-29] Idempotencia al crear empresas
- [Backend] `POST /super/api/empresas` usa creacion idempotente por administrador, tipo, nombre y NIT; doble clic o POST concurrente devuelve la empresa existente sin insertar otra.
- [Frontend] El formulario de `seleccionar_empresa.html` bloquea el boton `Guardar` mientras se procesa la creacion.
- [Seguridad] La checklist multiempresa y decisiones tecnicas dejan como norma que las altas/acciones criticas deben ser idempotentes en backend, no solo bloquear botones.
- [QA] Pruebas enfocadas Go para normalizacion de clave idempotente y parseo JS del selector.

## [2026-05-29] Plantillas empresariales
- [Nomenclatura] El catalogo de soluciones empresariales queda nombrado como `Plantillas` en textos visibles, rutas nuevas, scripts, handlers y pruebas enfocadas.
- [Rutas] Se crean/actualizan paginas y endpoints activos bajo `plantillas_nuevas` y `plantillas_integracion`, incluyendo menu empresarial, catalogo publico, super administrador y matriz de integracion.
- [Compatibilidad] Se mantienen aliases o claves tecnicas heredadas solamente donde forman parte de contratos internos ya existentes, como alcance de licencia/configuracion.
- [QA] Validacion Go enfocada en `db`, `handlers` y `utils`; parseo JS/HTML con Node empaquetado.

## [2026-05-29] Tarjeta Domotica configurable en carrito
- [Configuracion] `Configuracion > Carrito unificado` y `Configuracion > Estaciones` agregan el check `Mostrar tarjeta Domotica automaticamente`, guardado en `estaciones_config.carrito_ui_global.mostrar_tarjeta_domotica_carrito` o en la configuracion propia de la estacion.
- [Carrito] La tarjeta `Domotica` aparece automaticamente al volver al carrito cuando la estacion tiene aparatos configurados y la vista de tarjeta esta activa; si no hay aparatos o el check esta apagado, permanece oculta.
- [Datos] No cambia endpoints ni tablas; se reutiliza `/api/empresa/estacion_prefs` y el endpoint existente `/api/empresa/control_electrico` con aislamiento por `empresa_id`.

## [2026-05-29] Email corporativo adaptado a apariencias del sistema
- [Panel empresa] `web/administrar_empresa/panel.html` detecta tema claro/oscuro, adapta la tarjeta de correo y pasa la preferencia al autologin del webmail.
- [UX] La tarjeta ya no muestra el correo de la empresa ni el texto de estado `Buzon activo...`; conserva solo la bandeja integrada y alertas cuando exista un problema.
- [Configuracion] Nueva pagina `Configuracion > Email corporativo` para activar/desactivar apertura automatica del buzon y cambiar la contrasena interna del correo.
- [Backend] `/api/empresa/email_corporativo` acepta `POST` seguro por `empresa_id`; guarda preferencia en `empresa_estacion_prefs` y cifra la nueva clave antes de reprovisionar Mailu.
- [Mailu/SnappyMail] Se agregan temas `PCSLight` y `PCSDark` montados en Docker; SnappyMail queda con `PCSLight@custom` como base y el provisionamiento puede escribir preferencia por usuario.
- [Deploy] El perfil `mail` copia los temas personalizados al contenedor `pcs-mailu-webmail` y conserva el iframe `same-site`.

## [2026-05-29] Favoritos del panel empresarial como botones de menu
- [Panel empresa] En `web/administrar_empresa/panel.html`, el encabezado de Favoritos muestra `Accesos rapidos / Favoritos` en una sola fila.
- [UX] Los favoritos dejan la apariencia de tarjetas pequeñas y adoptan botones compactos alineados al estilo del menu principal de Administrar empresa.

## [2026-05-29] Rediseño empresarial de información de contacto
- [Portal] `/Informacion_de_contacto.html` queda con hero comercial, canales directos, asesor de ventas y áreas de atención.
- [Frontend] `web/estilos.css` moderniza la página con composición responsive, tarjetas compactas, botones claros y ajuste móvil para correos largos.
- [QA] Validación visual local en escritorio y móvil mediante Playwright con rutas estáticas interceptadas.

## [2026-05-29] Fotos de usuarios en carpeta de imagenes empresariales
- [Usuarios] `Administrar usuarios` permite cargar foto para usuarios creados por el administrador de empresa.
- [Storage] Las fotos se guardan en `/uploads/empresas/empresa_{id}_{slug}/imagenes/usuarios/`, compartiendo la carpeta `imagenes` empresarial con Domotica.
- [Backend] `users` agrega `foto_url`; `/api/empresa/usuarios?action=foto` valida `empresa_id`, pertenencia del usuario, extension y tamano antes de guardar.
- [Super] `Storage Imagenes` cuenta el uso de toda la carpeta `/imagenes/` de cada empresa, no solo Domotica.

## [2026-05-29] Domotica empresarial con fotos y storage
- [Backend] `/api/empresa/control_electrico` agrega integraciones Home Assistant/Siri bridge, Matter, Shelly, Hue, Tuya, Zigbee2MQTT y Z-Wave JS; tambien lecturas de consumo, reglas por sensor, alarmas y reportes.
- [Storage] Las fotos de aparatos se guardan por empresa en la subcarpeta `/uploads/empresas/empresa_{id}_{slug}/imagenes/domotica/` con limite configurable por super administrador.
- [Super] Nueva pagina `web/super/domotica_storage.html` y API `/super/api/domotica_storage` para ver carpetas, uso, numero de imagenes y tamano maximo por empresa.
- [QA] `go test ./db ./handlers ./tools/seed_domotica_motel_calipso -run "ControlElectrico|Domotica|Storage|^$" -count=1`; parseo de scripts inline con Node empaquetado.

## [2026-05-29] Domotica como modulo principal
- [UX] `Domotica` sale del submenu de Configuracion y queda como boton principal de Administrar empresa.
- [UX] El submenu de Domotica agrupa resumen, conexion GPIO, controladores, sensores Raspberry, estaciones/aparatos y bitacora.
- [Permisos] `linkControlElectrico` y `linkConfiguracionSensoresRaspberry` quedan agrupados bajo Domotica conservando la clave tecnica `control_electrico`.

## [2026-05-29] Domotica por estacion en carrito
- [UX] El modulo visible `Control electrico` pasa a llamarse `Domotica` en configuracion, licencias, ayuda, carrito e index.
- [Carrito] Si una estacion tiene carrito activo, Domotica habilitada y aparatos configurados, aparece automaticamente la tarjeta `Domotica`; cada aparato abre una ventana con estado, encendido, apagado y programacion horaria.
- [Datos] Se conserva la clave tecnica `control_electrico` y se reutilizan las tablas existentes de configuracion, aparatos y eventos por `empresa_id`.
- [QA] `go test ./handlers -run ControlElectrico -count=1`, parseo de scripts inline con Node y `git diff --check`.

## [2026-05-28] Auditoria especial super administrador
- [Backend] `/super/api/auditoria` acepta `scope=super_panel` solo para roles super y las APIs sensibles de configuracion super quedan envueltas con `WithSuperAuditoria`.
- [Frontend] `web/super_administrador.html` incorpora `Auditoria super`, con `web/super/auditoria_super_admin.html` para filtros, KPIs, detalle y exportacion CSV/JSON.
- [Seguridad] Los eventos visuales y automaticos registran navegacion, edicion y guardado sin persistir tokens, claves, passwords ni secretos.

## [2026-05-28] Auditoria global del selector
- [Backend] Se agrega `/super/api/auditoria`, `super_auditoria_eventos` y el middleware `WithSuperAuditoria` para trazabilidad de empresas, administradores, empresas compartidas, tipos de empresa, licencias y reportes globales.
- [Frontend] `seleccionar_empresa.html` incorpora el boton `Auditoria` y `web/super/auditoria_global.html` muestra KPIs, filtros, detalle y exportacion CSV/JSON.
- [Seguridad] La consulta queda limitada por `principal_email` para administradores normales; el super administrador conserva alcance global y la metadata se sanea sin secretos.
- [QA] `go test ./... -run "^$" -count=1`, `node --check` y captura visual local de la pagina.

## [2026-05-28] Energia solar multiempresa
- [Backend] Se agrega `/api/empresa/energia_solar` con tablas `empresa_energia_solar_sistemas`, `empresa_energia_solar_alertas`, `empresa_energia_solar_lecturas` y `empresa_energia_solar_eventos`.
- [Integraciones] Catalogo base para Victron VRM/VictronConnect, SMA Sunny Portal, SolarEdge Monitoring y gateway local/BMS.
- [Baterias] Soporte operativo para Tesla Powerwall, BYD Battery-Box, Pylontech US5000, Enphase IQ Battery y Victron Lithium con metricas SOC/SOH/voltaje/ciclos/temperatura/celdas.
- [Frontend] Nueva pagina `Administrar empresa > Analisis y control > Energia solar` con dashboard, configuracion, alertas, lecturas y eventos.
- [Permisos] Nuevo modulo `energia_solar`, pagina `linkEnergiaSolar`, wrapper `WithEmpresaEnergiaSolarPermissions` y fallback de licencia a `control_electrico`/`seguridad`.
- [QA] `go test ./db ./handlers -run TestDoesNotExist -count=1`, `node --check` y prueba visual con Chrome headless usando servidor mock.

## [2026-05-28] Descarga de informacion integrada al selector
- [UX] El boton de descarga de cada tarjeta de empresa abre la pagina dentro del panel derecho de `seleccionar_empresa.html`, junto al menu global.
- [Apariencia] `descargar_informacion_de_la_empresa.html` usa variables de tema para respetar apariencias claras y oscuras.
- [Backup] El formato `Backup completo (.json)` descarga un JSON integral versionado con el snapshot de la empresa visible.
- [Navegacion] Se agrega `Regresar a seleccionar empresas` para volver al listado cuando la pagina esta embebida.

## [2026-05-27] Reportes globales con alcance del selector
- [Seguridad] `/super/api/reportes_globales` deja de usar una excepcion global por rol super y ahora reutiliza el alcance efectivo del selector de empresas.
- [Alcance] El catalogo, tablero, datasets y exportaciones solo pueden usar empresas propias, delegadas o compartidas que el administrador autenticado ve en `seleccionar_empresa.html`.
- [QA] `go test ./handlers -run "TestSuperReportesGlobales" -count=1`.

## [2026-05-27] Reportes globales profesionales en selector de empresas
- [UX] `web/super/reportes_globales.html` queda enfocado en fecha desde/hasta, reporte disponible, formato y acciones directas: Ver, Exportar, Imprimir y Enviar por email.
- [Analitica] La vista conserva seleccion multiple de empresas y muestra KPIs, graficos por empresa, lectura ejecutiva, resumen por empresa y detalle del reporte seleccionado.
- [Formatos] Se mantienen los formatos existentes del sistema: PDF, XLS/Excel, CSV, TXT y JSON, usando `/super/api/reportes_globales`.
- [QA] Validacion JS, `git diff --check` y prueba visual Playwright con datos simulados: graficos visibles, exportacion PDF, impresion disparada y movil sin desbordamiento horizontal.

## [2026-05-27] Vista previa de modulos con vinetas por tema
- [UX] La vista previa de `web/super/informacion_de_modulos.html` deja de usar marcadores con color fijo.
- [Apariencias] Las vinetas usan variables de tema y conservan contraste en `Blanco Corporativo`, temas oscuros y `Corporativo Oscuro`.
- [Alcance] Cambio visual sin backend, tablas, endpoints ni permisos.

## [2026-05-27] Tema corporativo oscuro y blanco corporativo movil
- [UX] Se agrega `Corporativo Oscuro` como tema oscuro seleccionable desde el menu flotante.
- [Movil] `Blanco Corporativo` conserva fondo, borde y texto claros en el panel de apariencias y submenus del menu flotante.
- [Compatibilidad] Las paginas embebidas, super administrador, juegos y utilidades reconocen `dark-corporate` para no degradar el tema al navegar.

## [2026-05-27] Empresas compartidas en administrar empresa
- [UX] Se agrega `web/administrar_empresa/empresas_compartidas.html` al grupo `Administracion` del menu empresarial.
- [Funcion] La pagina muestra accesos compartidos activos e invitaciones pendientes de la empresa actual, con opcion para desactivar acceso o cancelar invitacion usando `/super/api/empresas/compartidos`.
- [Permisos] El menu usa permiso `seguridad:U`; el backend existente conserva la validacion de propietario, quien compartio o administrador receptor.
- [QA] Validacion sintactica del script embebido en la pagina; `git diff --check`.

## [2026-05-27] Ayuda en administradores
- [UX] `web/super/administradores.html` agrega una descripcion bajo el titulo para aclarar que las invitaciones son para administrar empresas y que compartir una sola empresa se hace desde el icono Compartir de la tarjeta.
- [QA] Validacion sintactica del script embebido en `web/super/administradores.html`; `git diff --check`.

## [2026-05-27] Estado de invitacion en administradores
- [UI] `web/super/administradores.html` separa `Invitacion` y `Estado cuenta`.
- [Backend] `GetAdministradores` expone `email_confirmado` y el filtro de alcance agrega `invitation_status` para distinguir invitaciones aceptadas, pendientes o cerradas.
- [QA] `go test ./utils ./handlers ./db -run "FilterAdministradoresForPrincipalScope|AdministradoresEffectivePrincipalScope|ValidatePendingAdminInvitationToken" -count=1`; validacion sintactica del script embebido en `web/super/administradores.html`; `git diff --check`.

## [2026-05-27] Administradores del selector filtrados y super administradores por invitacion
- [Selector] El enlace `Administradores` de `seleccionar_empresa.html` abre `/super/administradores.html?scope=principal`.
- [Backend] `/super/api/administradores?scope=principal` filtra por el administrador autenticado aunque la cuenta tenga rol super; sin ese parametro el panel super conserva la lista global.
- [Super administradores] Crear un `super_administrador` tambien genera invitacion y exige token al registrarse; al aceptar conserva rol super y entra al modulo de super administrador.
- [Datos] `GetAdministradores` ahora devuelve `email_confirmado` para distinguir cuentas activas de invitaciones pendientes.
- [QA] `go test ./utils ./handlers ./db -run "AdminLimitedRoute|DecorateEmpresasByEffectiveAccess|FilterAdministradoresForPrincipalScope|AdministradoresEffectivePrincipalScope|ValidatePendingAdminInvitationToken|CanAdminAccessEmpresa|AdminPrincipalDeleg" -count=1`.

## [2026-05-27] Delegacion de portafolio entre administradores
- [Flujo] Si el administrador invitado no existe, se conserva el flujo de invitacion por correo y registro con token.
- [Cuenta existente] Si el correo ya pertenece a un administrador confirmado, no se crea otra cuenta ni se cambia `usuario_creador`; se activa una delegacion de portafolio y el selector muestra sus empresas propias mas las compartidas.
- [Seguridad] Revocar desde Administradores quita solo la delegacion cuando la cuenta no fue creada por el principal; no borra la cuenta del otro administrador.
- [BD] Se agrega `admin_principal_delegaciones` en `pcs_superadministrador` para el acceso muchos-a-muchos entre administradores principales y administradores registrados.
- [QA] `go test ./utils ./handlers ./db -run "AdminLimitedRoute|DecorateEmpresasByEffectiveAccess|FilterAdministradoresForPrincipalScope|ValidatePendingAdminInvitationToken|CanAdminAccessEmpresa|AdminPrincipalDeleg" -count=1`.

## [2026-05-27] Invitacion de administradores delegados
- [Flujo] Agregar administrador desde `Administradores` ya no deja una cuenta lista para login: genera invitacion por correo con token.
- [Registro] El enlace abre `registrar_nuevo_usuario_administrador.html` con correo y token; el invitado acepta la invitacion, completa datos y crea contrasena.
- [Seguridad] Si la cuenta fue invitada por un principal, el registro exige token valido, vigente y no usado; sin enlace no se puede completar el alta.
- [Correo] Se agrega la plantilla `admin_scoped_invitation` para personalizar el mensaje de invitacion.
- [QA] `go test ./utils ./handlers ./db -run "AdminLimitedRoute|DecorateEmpresasByEffectiveAccess|FilterAdministradoresForPrincipalScope|ValidatePendingAdminInvitationToken|CanAdminAccessEmpresa" -count=1`; `node --check web/js/registrar_nuevo_usuario_administrador.js`.

## [2026-05-27] Administradores por administrador principal
- [Selector empresas] `seleccionar_empresa.html` muestra el acceso `Administradores` tambien al administrador principal normal, no solo al super administrador.
- [Backend] `/super/api/administradores` filtra por `administradores.usuario_creador` y por el principal resuelto: el principal no se lista a si mismo y no ve administradores de otros principales.
- [Acceso empresas] Los administradores delegados heredan acceso a las empresas creadas por el principal como `delegated`, pero no quedan como propietarios para compartirlas.
- [Seguridad] El handler mantiene validacion backend; cambiar URL, cache o frontend no concede acceso a administradores ni empresas ajenas.
- [QA] `go test ./utils ./handlers -run "AdminLimitedRoute|DecorateEmpresasByEffectiveAccess|FilterAdministradoresForPrincipalScope" -count=1`; `node --check web/js/seleccionar_empresa.js`; `git diff --check`.

## [2026-05-27] Checklist seguridad endpoint multiempresa
- [Documentacion] Se agrega `documentos/checklist_seguridad_endpoint_multiempresa.md` como requisito operativo para endpoints, consultas y acciones empresariales.
- [Seguridad] La checklist exige validar sesion, `empresa_id`, IDs secundarios, permisos, licencias, SQL aislado, entrada, auditoria, errores saneados y operaciones destructivas.
- [QA] Incluye pruebas minimas negativas: alterar `empresa_id`, usar IDs de otra empresa, rol insuficiente, empresa sin licencia y doble clic/concurrencia.
- [Integracion] `AGENTS.md`, `contexto_codex.md`, `decisiones_tecnicas.md` y `matriz_roles_permisos_pos_multiempresa.md` quedan enlazados a la checklist.

## [2026-05-27] Contexto operativo Codex
- [Documentacion] Se agregan `documentos/contexto_codex.md`, `documentos/mapa_modulos.md`, `documentos/flujos_operativos.md`, `documentos/comandos_codex.md` y `documentos/decisiones_tecnicas.md`.
- [AGENTS] La guia principal ahora exige revisar primero estos archivos para tener contexto de arranque, ubicacion de modulos, flujos, comandos y decisiones tecnicas permanentes.
- [Alcance] Cambio documental sin backend, frontend runtime, tablas, endpoints ni permisos.
- [QA] `git diff --check`.

## [2026-05-27] Alertas sistema para registros y empresas nuevas
- [Super administrador] `web/super/alertas_sistema.html` agrega dos checks: avisar registro de administrador y avisar creacion de empresa nueva.
- [Backend] `super_alertas_config` agrega `admin_register_enabled` y `empresa_nueva_enabled`, con destino existente `recipient_email` y defaults activos.
- [Eventos] `AdminRegisterHandler` y `EmpresasHandler` disparan notificaciones no bloqueantes despues de crear la cuenta administrativa o la empresa.
- [Historial] Cada aviso queda registrado en `super_alertas_eventos` con tipo `admin_registrado_login` o `empresa_nueva_admin`, estado de correo y metadata saneada sin claves ni tokens.
- [QA] `go test ./db ./handlers -run "SuperAlertas|AdminRegister|Empresas" -count=1`; validacion sintactica de `alertas_sistema.html`; verificacion visual local con Chrome headless.

## [2026-05-27] QR DIAN en factura o recibo
- [Configuracion] `Configuracion > Carrito unificado` agrega el check `Mostrar QR DIAN al final de la factura o recibo`, guardado en `estaciones_config.carrito_ui_global.mostrar_qr_factura_electronica`.
- [Carritos] Al cerrar una venta con documento electronico, `carrito_de_compras.html` arma la URL publica DIAN desde CUFE/CUDE/codigo de validacion y la imprime al final del recibo y de la factura electronica si la opcion esta activa.
- [Impresion] `web/js/print_documents.js` soporta un bloque QR comun en documentos POS/carta, en blanco y negro y sin depender de tema claro/oscuro.
- [QA] `go test ./handlers -run CarritoUI -count=1`; validacion sintactica JS/HTML; verificacion visual con Chrome headless de check visible y QR POS renderizado.
- [Alcance] No agrega tablas, endpoints, permisos ni dependencias externas; reutiliza `/vendor/qrcode.min.js` y mantiene aislamiento por `empresa_id`.

## [2026-05-27] Analitica publica solo en super administrador
- [Portal] `web/index.html` elimina la tarjeta visible `Visitas al portal` y sus estilos embebidos.
- [Analitica] El portal conserva un tracker oculto para seguir registrando visitas agregadas por pais sin mostrar la tarjeta al publico.
- [Super administrador] `web/super_administrador.html` mantiene la visualizacion completa con mapa real, ranking y total.
- [Alcance] No cambia backend, endpoints, tablas, permisos, privacidad ni dependencias.

## [2026-05-27] Mapa real en analitica publica
- [Portal] `web/js/portal_visits.js` usa `web/img/world-map-natural-earth-public-domain.svg` como mapa base real del contador `Visitas al portal por pais`.
- [Super administrador] El mismo widget compartido muestra el mapa real al final del panel sin duplicar conteos.
- [Asset] El SVG fue generado desde Natural Earth 1:110m admin 0 countries, de dominio publico.
- [Alcance] No cambia backend, endpoints, tablas, permisos, privacidad ni dependencias.

## [2026-05-27] Vinetas visibles en modulos del index
- [Portal] `web/index.html` reemplaza el marcador nativo por una viñeta visual propia en cada caracteristica de `Modulos y caracteristicas principales`.
- [Tema] Las viñetas usan variables de apariencia para conservar contraste en modo claro y oscuro.
- [Alcance] No cambia backend, endpoints, tablas, permisos ni dependencias.

## [2026-05-27] Informacion editable de modulos del index
- [Super administrador] Se agrega `web/super/informacion_de_modulos.html` para editar titulo, icono y vinetas de cada modulo principal.
- [Portal] `web/index.html` carga `/api/public/informacion_de_modulos` y conserva el HTML actual como respaldo si falla la conexion.
- [Backend] `/super/api/informacion_de_modulos` guarda en `pcs_superadministrador.configuraciones`; `/api/public/informacion_de_modulos` expone solo contenido editorial normalizado.
- [QA] `go test ./handlers -run InformacionModulos -count=1`; validacion sintactica de HTML/JS; verificacion visual local con Chrome headless.

## [2026-05-27] Login administrador con logo imagen
- [Login] `web/login.html` reemplaza el titulo textual `Powerful Control System` por la imagen `web/img/titulo-powerful-control-system-login.png`.
- [UX] `web/estilos.css` define un tamano pequeno y responsive para que el logo no empuje el formulario.
- [Alcance] No cambia backend, endpoints, permisos, reCAPTCHA ni Google OAuth.

## [2026-05-27] Index subtitulo POS multiempresa con domotica
- [Portal] `web/index.html` actualiza el subtitulo del encabezado publico a `Sistema de Facturacion Electronica con domotica integrada`.
- [Alcance] Cambio de texto visible sin backend, endpoints, tablas, permisos ni dependencias.

## [2026-05-27] Index con vinetas reales por caracteristica
- [Portal] `web/index.html` convierte las caracteristicas de cada tarjeta de `Modulos y caracteristicas principales` en listas HTML reales.
- [UX] Cada caracteristica queda con su propia vineta visible y una distribucion compacta dentro de la tarjeta.
- [Alcance] No cambia backend, endpoints, tablas, permisos ni dependencias.

## [2026-05-26] Cliente general configurable en carrito
- [Configuracion] `Configuracion > Carrito unificado` y `Configuracion > Estaciones` agregan el campo `Nombre para ventas sin cliente`.
- [Carritos] El carrito usa el nombre configurado por empresa cuando no hay `cliente_id`, incluyendo cliente actual, listados y documentos imprimibles del carrito.
- [Backend] `defaultEmpresaPreconfigCarritoUI()` define `cliente_general_nombre` como `Cliente General` por defecto.
- [Alcance] Se guarda en `estaciones_config.carrito_ui_global`; no agrega tablas, endpoints, permisos ni dependencias.

## [2026-05-25] Creditos y cartera con submenu
- [UX] `web/administrar_empresa/creditos_menu.html` separa el modulo en botones de subpagina: panel, nuevo credito, cartera, morosidad, riesgo y limites, abonos/operaciones, aprobaciones y estado de cuenta.
- [Frontend] `web/administrar_empresa/creditos.html` soporta vistas por `view=` y conserva el credito seleccionado al saltar desde cartera o morosidad hacia estado de cuenta u operaciones.
- [Permisos] Los nuevos links internos quedan registrados bajo `finanzas:C`; no cambian endpoints, tablas ni dependencias.

## [2026-05-25] Licencia del sistema descargable por empresa
- [Administrar empresa] `web/administrar_empresa.html` agrega `Licencia del sistema` al final del grupo `Administracion`.
- [Frontend] `web/administrar_empresa/licencia_sistema.html` muestra el alcance de licencia por empresa y permite descargar TXT, descargar HTML o imprimir/guardar PDF.
- [Permisos] `linkLicenciaSistema` queda bajo `seguridad:R` en frontend y backend; no se crean endpoints, tablas ni dependencias.

## [2026-05-25] Menu flotante sin ayuda administrador
- [Frontend] `web/menu.js` elimina el enlace `Ayuda administrador` del menu flotante centralizado.
- [Soporte] Se conserva el boton `Crear ticket de ayuda` para solicitudes operativas.
- [Alcance] No cambia backend, permisos, rutas, endpoints, tablas ni dependencias.

## [2026-05-25] Finanzas debajo de inventario
- [UX] `web/administrar_empresa.html` mueve el grupo `Finanzas y cumplimiento` para quedar inmediatamente debajo de `Inventario y compras`.
- [Alcance] Cambio de orden de menu sin modificar permisos, rutas, endpoints, tablas ni dependencias.

## [2026-05-25] Navegacion financiera y paginas huerfanas
- [Finanzas] `web/administrar_empresa.html` muestra `Creditos y cartera` y `Gestion de cobranza` directamente dentro de `Finanzas y cumplimiento`.
- [Centro financiero] `web/administrar_empresa/finanzas_menu.html` agrega accesos rapidos a creditos y cobranza, manteniendo la barra lateral existente.
- [Configuracion] `web/administrar_empresa/configuracion_menu.html` conecta `Configuracion guiada` e `Integraciones`, paginas que ya existian sin link directo.
- [Canales] `Chat IA centralizado` queda disponible desde Canales digitales y colaboracion.
- [Permisos] `web/js/administrar_empresa.js` registra los links visibles con permisos existentes; no se agregan endpoints, tablas ni dependencias.

## [2026-05-25] Index modulos mas compactos
- [Portal] `web/index.html` compacta la seccion `Modulos y caracteristicas principales` para que cada tarjeta tenga menos margen interno y mas ancho util para el texto.
- [UX] Las caracteristicas quedan como texto fluido con puntos negros por elemento, evitando los huecos grandes que generaba justificar cada item corto por separado.
- [Responsive] Se valida escritorio y movil sin desbordamiento horizontal.

## [2026-05-25] Licencia gratis valor cero sin rollback
- [Backend] `backend/db/licencias_gratis.go` registra una sola activacion gratis por empresa usando la licencia activa realmente asignada por `activateLicenciaForEmpresaTx`.
- [PostgreSQL] Evita ignorar un segundo `INSERT` que chocaba con `ux_licencias_gratis_empresa_unica`; aunque el error se ignoraba en Go, PostgreSQL abortaba la transaccion y el `Commit` devolvia `commit unexpectedly resulted in rollback`.
- [Checkout] `web/pagar_licencia.html` no carga tarjetas ni terminos Wompi cuando el resumen ya esta en total cero, porque ese flujo se activa sin pasarela y no debe producir 502 residual de `/wompi/terms`.
- [QA] `go test ./db -run "Licencia|PostgresPrimaryKey|PaymentGateway" -count=1`; `go test ./handlers -run "Licencia|Epayco|Wompi|Checkout|Payment" -count=1`.

## [2026-05-25] Licencia gratis de 15 dias reparada
- [Backend] `backend/db/licencias_gratis.go` crea `licencias_activaciones_gratis.id` como `BIGSERIAL PRIMARY KEY` y repara secuencias/defaults existentes en PostgreSQL antes de insertar marcas de prueba/gratis.
- [Middleware] `backend/utils/utils.go` permite sin sesion `/api/public/licencias/checkout_summary` y `/licencias/activar_sin_pago`, manteniendo la validacion real en los handlers de licencia, empresa y total cero.
- [Licencias] `POST /licencias/activar_sin_pago` queda idempotente para reintentos: si una activacion valor cero ya quedo vigente para la empresa, responde exito y redirige en lugar de devolver error.
- [Operacion] La preconfiguracion del tipo de empresa se intenta aplicar, pero si falla despues de activar la licencia se registra advertencia y no bloquea la licencia de prueba.
- [PWA] `web/sw.js` captura fallos de red en recursos GET cacheables para evitar promesas rechazadas visibles en consola.
- [QA] `go test ./utils -run TestAuthMiddlewarePublicAndProtectedSuperRoutes -count=1`; `go test ./db -run "Licencia|PostgresPrimaryKey|PaymentGateway" -count=1`; `go test ./handlers -run "Licencia|Epayco|Wompi|Checkout|Payment" -count=1`; `node --check web/sw.js`.

## [2026-05-25] Super administrador con analitica al final
- [UX] `web/estilos.css` evita que la tarjeta `Analitica publica / Visitas al portal por pais` reste alto al panel principal de super administrador.
- [Layout] El `iframe` conserva una vista completa y la analitica publica queda debajo, visible al bajar al final del panel.
- [Alcance] No cambia backend, endpoint de visitas, tablas, permisos, privacidad ni dependencias.

## [2026-05-25] Login y registro administrador verificados
- [QA] Se valida en produccion el login administrativo por correo en escritorio y movil, con reCAPTCHA v3 activo y redireccion correcta al panel de super administrador.
- [Registro] `registrar_nuevo_usuario_administrador.html` permite crear una cuenta administrativa nueva desde el enlace de `login.html` y redirige al login con el correo precargado.
- [Google] El boton `Continuar con Google` abre correctamente el flujo OAuth hacia `accounts.google.com` en escritorio y movil, con `redirect_uri` productivo hacia `/auth/google/callback`.
- [UX] `web/estilos.css` permite que el titulo del registro administrador se parta en varias lineas en celular, evitando recortes horizontales.
- [Alcance] No cambia backend, endpoints, tablas, permisos, reCAPTCHA, Google OAuth ni dependencias externas.

## [2026-05-25] Portal y super admin con analitica compartida
- [Portal] `web/index.html` muestra cada modulo principal con un icono mediano relacionado con inventario, POS, pagos, documentos electronicos, finanzas, estaciones, IA, control fisico, gestion y plantillas.
- [Super administrador] `web/super_administrador.html` agrega en la parte baja el mismo contador de visitas por pais en modo lectura, usando el componente comun sin incrementar visitas desde el panel interno.
- [Frontend] `web/js/portal_visits.js` soporta multiples widgets con atributos `data-portal-visits-*`, centraliza la consulta/registro, evita POST duplicados por pagina y agrega halos visuales a los marcadores del mapa.
- [Alcance] No cambia backend, tablas, endpoint `/api/public/portal_visitas`, privacidad, permisos ni dependencias externas.

## [2026-05-25] Index con caracteristicas punteadas por modulo
- [Portal] `web/index.html` convierte las descripciones de `Modulos y caracteristicas principales` en listas compactas.
- [UX] Cada caracteristica se muestra con punto negro grande y texto justificado dentro del ancho de la tarjeta, manteniendo los iconos medianos existentes.
- [Alcance] No cambia backend, endpoints, tablas, permisos ni dependencias.

## [2026-05-25] Selector de empresa con tipos en orden inverso
- [UX] `web/js/seleccionar_empresa.js` invierte el orden visible del selector `Tipo de empresa` al abrir el formulario `Agregar Empresa`.
- [Alcance] No cambia `/super/api/tipos_empresas`, backend, tablas, permisos ni preconfiguraciones; solo cambia la presentacion del listado en esa pantalla.

## [2026-05-25] Contador de visitas compacto con mapa realista
- [Portal] `web/index.html` centra y compacta la tarjeta del contador de visitas por pais para que total, mapa y ranking queden dentro de un solo bloque mas pequeno.
- [Mapa] `web/js/portal_visits.js` reemplaza el mapa esquematico de blobs por una silueta mundial SVG mas realista con graticula, continentes y marcadores de color por pais.
- [Alcance] No cambia backend, endpoint `/api/public/portal_visitas`, tablas ni datos almacenados.

## [2026-05-25] Index con modulos y documentos electronicos en lista
- [Portal] `web/index.html` cambia la descripcion publica de modulos desde un parrafo unico hacia una lista de caracteristicas principales.
- [Documentos electronicos] La lista incluye factura electronica, notas credito/debito, documento soporte, notas de ajuste, nomina electronica, documentos equivalentes electronicos/POS electronico, contingencia y eventos RADIAN para Colombia; tambien menciona factura, nota credito y nota debito para Panama y Ecuador segun pais configurado.
- [Alcance] No cambia backend, tablas, endpoints, permisos ni licencias; es un ajuste de contenido y presentacion del portal publico.

## [2026-05-22] Index con modulos principales en parrafo unico
- [Portal] `web/index.html` simplifica la seccion publica de modulos a un solo parrafo con funciones principales del sistema.
- [Contenido] El resumen incluye inventario, compras, bodegas, datafonos, cajon monedero, cajas simultaneas, caja por usuario, pagos QR, factura electronica, impuestos, modulo del contador, finanzas, IA, control electrico, sensores, reportes y plantillas operativas.
- [Alcance] No cambia backend, endpoints, tablas, permisos ni dependencias externas.

## [2026-05-21] Contador de visitas por pais en index
- [Portal] `index.html` agrega al final un contador total de visitas, mapa mundial con marcadores de color por pais y ranking con barras proporcionales.
- [Backend] `/api/public/portal_visitas` registra y consulta visitas agregadas por pais/fecha en `portal_visitas_paises`.
- [Privacidad] No se guardan IP, user-agent, correos ni identificadores personales; solo pais, fecha y conteo.
- [Validacion] `go test . ./handlers -count=1`, `node --check web/js/portal_visits.js`, parseo inline de `index.html` y prueba visual local con 6 paises/6 marcadores.

## [2026-05-21] Iconos globales de botones
- [Frontend] Se agrega `web/js/button_icons.js`, un decorador comun que detecta botones estaticos o dinamicos y les asigna un icono de color segun funcion: guardar, buscar, pagar, exportar, imprimir, correo, WhatsApp, cliente, inventario, reportes, caja, QR, soporte o configuracion.
- [Servidor] `backend/main.go` inyecta el script en respuestas HTML estaticas mediante `buttonIconsStaticHandler`, sin tocar CSS/JS/imagenes ni duplicar la etiqueta cuando ya existe.
- [UX] Los botones con iconos nativos, botones de cierre, calculadora y controles de juegos se respetan para evitar duplicados o ruido visual.
- [Validacion] `go test . -count=1`, `node --check web/js/button_icons.js` y prueba visual local en `administrar_empresa.html` confirmando 19 botones relevantes con icono y cero faltantes.

## [2026-05-21] Iconos de color en botones del carrito
- [Frontend] `carrito_de_compras.html` decora automaticamente botones fijos y dinamicos con un icono de color segun accion: pago, agregar, buscar, cliente, descuento, QR, abono, cancelar, editar o regresar.
- [UX] Los iconos se renderizan como insignias pequeñas dentro del boton, manteniendo botones planos y legibles en escritorio y movil.
- [Alcance] No cambia backend, endpoints, tablas, permisos ni dependencias externas.

## [2026-05-21] Impuestos en Finanzas y cumplimiento
- [Navegacion] `Impuestos` sale del menu interno de Configuracion y queda como boton directo en `Finanzas y cumplimiento`.
- [Centro financiero] `finanzas_menu.html` agrega `Impuestos` en la barra lateral y en los accesos rapidos superiores.
- [Permisos] `linkImpuestos` conserva el mismo modulo/accion efectivo y queda agrupado en el catalogo financiero-contable.

## [2026-05-21] Emisoras online por pais
- [Backend] `/api/chat_flotante/preferencias` agrega `radio_country` y `radio_custom_stations`, persistidos por `empresa_id` en `empresa_estacion_prefs` sin tablas nuevas.
- [Frontend] El reproductor flotante y `administrar_empresa/radio_online.html` detectan pais por configuracion/facturacion o IP publica, muestran catalogo solo para Panama y Ecuador, y permiten agregar/eliminar emisoras personalizadas de la empresa.
- [Catalogo] `web/js/radio_catalog.js` publica las 10 emisoras base de Panama y las 10 de Ecuador; paises no soportados muestran mensaje operativo y opcion de emisoras propias.
- [Validacion] `go test ./db ./handlers -count=1`, `node --check` de los JS de radio con runtime empaquetado y prueba visual local de Panama, Ecuador, emisora personalizada y drawer flotante.

## [2026-05-21] Avisos globales de conexion
- [Frontend] `empresa_submenu_context.js` escucha eventos `offline` y `online` del navegador para avisar perdida y regreso de internet.
- [UX] Cuando se pierde internet el aviso queda persistente; los modulos normales piden esperar a que vuelva la conexion, y el carrito/caja con facturacion offline activa indica que puede seguir vendiendo e imprimir provisionalmente.
- [Shell] `administrar_empresa.html` carga el monitor comun para que el panel principal tambien emita el aviso, no solo los modulos internos.
- [Auditoria] `/api/empresa/auditoria/eventos?action=conexion` registra `internet_perdido` e `internet_restaurado` en `empresa_auditoria_eventos`; si el navegador queda sin red, el evento se conserva en cola local y se envia al volver.
- [Alcance] No agrega tablas ni dependencias; usa el modulo de auditoria existente con aislamiento por `empresa_id`.

## [2026-05-21] Impresoras por producto, categoria o todos
- [Backend] `/api/empresa/impresoras` agrega `producto_regla` y `catalogo_categorias` para guardar reglas masivas por todos los productos o por categoria, siempre aisladas por `empresa_id`.
- [Resolucion] La prioridad operativa queda `receta -> producto especifico -> categoria de producto -> todos los productos -> funcionalidad -> predeterminada`.
- [Frontend] `Configuracion > Impresora` permite elegir el alcance `Producto especifico`, `Categoria de producto` o `Todos los productos`, y muestra las reglas en una tabla unificada.
- [Validacion] `go test ./db ./handlers -count=1`, parseo de scripts inline y prueba visual local con API simulada de cocina/bar.

## [2026-05-21] Recetas de productos
- [Backend] Se consolida el contrato compuesto como recetas: `/api/empresa/recetas_productos`, estructuras `RecetaProducto`, tablas `recetas_productos`, `recetas_productos_detalle` y `recetas_productos_versiones`.
- [Carrito] Los items compuestos usan `tipo_item=receta` y descuentan inventario por ingredientes con el mismo control transaccional por `empresa_id`.
- [Frontend] El modulo queda como `administrar_empresa/recetas_productos.html`, el menu de productos muestra `Recetas` y el carrito busca recetas desde el catalogo inteligente.
- [Operacion] Las asignaciones de impresora por receta usan `empresa_impresoras_recetas` y el reporte de corte clasifica recetas junto a productos.

## [2026-05-21] Carrito compacto
- [Frontend] La tarjeta de codigo/SKU y cantidad usa grilla compacta para mostrar `Codigo de barras o SKU`, `Cantidad`, `Agregar` y `Buscar Productos` en una sola fila cuando hay ancho.
- [UX] El panel de cliente muestra `Cliente actual` en el encabezado, compacta busqueda por nombre/identificacion y deja el formulario rapido de nuevo cliente en mas columnas.
- [Responsive] En celular las grillas se reorganizan sin perder campos completos ni botones tactiles.
- [Alcance] No cambia backend, endpoints, tablas, permisos ni dependencias.

## [2026-05-21] Busqueda de cliente del carrito
- [Frontend] El formulario de nuevo cliente en el carrito inicia oculto y solo se abre con `Nuevo cliente`.
- [UX] El campo visible `Cliente del carrito` busca clientes existentes por `Nombre` o por `NIT / cedula / identificacion`.
- [Operacion] Al elegir un cliente de las sugerencias se asigna al carrito; si no existe, el cajero abre el formulario de creacion.
- [Alcance] No cambia endpoints, tablas ni permisos; reutiliza clientes y actualizacion de `cliente_id` del carrito.

## [2026-05-21] Botones configurables en acciones del carrito
- [Configuracion] `Configuracion > Carrito unificado` y `Configuracion > Estaciones` permiten activar u ocultar el panel de cliente y cada boton de acciones del carrito.
- [Frontend] El carrito aplica visibilidad individual para `Descuentos`, `Cambiar tarifa`, `Control electrico`, `Cancelar carrito`, `Taxi`, `Clientes`, `Abonos` y `Vehiculo`.
- [Operacion] Si todos los controles de acciones quedan ocultos, la tarjeta de acciones no se muestra vacia.
- [Alcance] Se guarda en el JSON de `estaciones_config`; no agrega tablas, rutas ni permisos.

## [2026-05-21] Alerta visual configurable en carrito
- [Configuracion] `Configuracion > Carrito unificado` y `Configuracion > Estaciones` agregan visibilidad del check, minutos de alerta y activacion por defecto.
- [Frontend] El carrito oculta la alerta si la configuracion no la permite; si se muestra, queda desactivada hasta que el cajero la marque o hasta que la empresa active el default.
- [Operacion] El temporizador usa el tiempo configurado y ya no se reinicia en cada refresco automatico del carrito.
- [Alcance] Se guarda en `estaciones_config.carrito_ui_global` y en overrides por estacion; no agrega tablas, rutas ni permisos.

## [2026-05-21] Clientes en carrito por busqueda de nombre
- [Frontend] `carrito_de_compras.html` retira el texto auxiliar y el selector visible `Cliente registrado` del panel de cliente.
- [UX] El campo `Nombre / razon social` ahora lista clientes activos por nombre/documento; al elegir uno se asigna al carrito actual.
- [Operacion] Si se crea un cliente nuevo desde el mismo panel, queda asociado al carrito; si el carrito no tiene cliente, el campo de nombre inicia en blanco.
- [Alcance] Reutiliza `/api/empresa/clientes` y `/api/empresa/carritos_compra`; no cambia tablas, permisos ni dependencias.

## [2026-05-21] Carrito movil con busqueda primero
- [Frontend] `estilos.css` ajusta el orden responsive del carrito para que en celular la tarjeta de codigo/SKU, `Agregar` y `Buscar Productos` quede primero.
- [UX] La columna de totales deja de subirse por encima de la busqueda en pantallas menores a 900px.
- [Alcance] No cambia HTML, JavaScript, endpoints, tablas, permisos ni el orden de escritorio.

## [2026-05-21] Panel movil sin indicadores economicos
- [Frontend] `administrar_empresa.html` carga `service_worker_update.js` con version nueva para celulares/PWA.
- [PWA] `sw.js` sube el cache a `pcs-shell-v4` y pide navegaciones, JS y CSS con `cache: no-store` antes de guardar la copia nueva.
- [Ayuda] Se retira la referencia antigua que indicaba mostrar USD/COP e indicadores de mercado en el panel.
- [Validacion] Panel y shell administrativo servidos localmente no contienen `Mercado en contexto`, `Indicadores economicos`, `USD / COP`, `Bitcoin`, `Ethereum`, `S&P 500` ni `Nasdaq 100`.

## [2026-05-20] Catalogo DIAN Colombia
- [Backend] Se agrega catalogo de documentos electronicos DIAN Colombia, documentos equivalentes, notas de ajuste, contingencia y eventos RADIAN.
- [Frontend] `facturacion_electronica.html` muestra una tarjeta Colombia para activar documentos del SFE y guardar la seleccion en `campos_pais_json`.
- [Contabilidad] Se separan obligaciones que suelen preparar contadores, como declaraciones, informacion exogena, certificados de retencion y conciliacion fiscal.
- [Alcance] No agrega rutas base, tablas, permisos, licencias ni dependencias nuevas.

## [2026-05-20] Centro de reportes por selector
- [Frontend] `reportes_ejecutivos.html` elimina las tarjetas del catalogo y muestra todos los reportes en un selector y una lista compacta.
- [Vista previa] El reporte seleccionado se renderiza como documento imprimible en blanco y negro, con modo POS 80mm o carta tomado de la configuracion previa de corte/reporte.
- [Exportacion] Los botones PDF, Excel, CSV, JSON y TXT descargan el dataset seleccionado mediante `/api/empresa/reportes?action=export`.
- [Alcance] No agrega rutas, permisos, tablas ni dependencias nuevas.

## [2026-05-20] Traslado entre bodegas
- [Inventario] La vista `Bodegas` vuelve a mostrar la tarjeta `Traslado entre bodegas`, existencias, alertas y movimientos recientes.
- [Backend] `TransferirProductoEntreBodegas` valida producto/bodegas por `empresa_id`, descuenta origen, suma destino y registra kardex con wrappers SQL compatibles con PostgreSQL.
- [Pruebas] Se agrega regresion para impedir `tx.QueryRow`/`tx.Exec` directos en el traslado transaccional.

## [2026-05-20] Nomina profesional
- [Backend] Nomina Colombia avanzada amplia el catalogo de conceptos y agrega `aprobar_novedad_colombia` y `seed_motel_calipso`.
- [Liquidacion] Las novedades aprobadas impactan devengado, deducciones, IBC, salud, pension y neto a pagar.
- [Demo] El seed profesional crea empleados simulados de Motel Calipso, asistencia, novedades, liquidaciones, PILA y pagos.
- [Frontend] La pantalla de nomina agrega tablero de cobertura profesional, boton demo y acciones para aprobar/rechazar novedades.

## [2026-05-20] Nombres configurables de estaciones
- [Configuracion] `Configuracion > Estaciones` permite definir el nombre singular y plural del recurso operativo de la empresa.
- [Frontend] Administrar empresa, estaciones, carrito unificado y configuracion del carrito reemplazan `Estacion/Estaciones` por el nombre configurado.
- [Preconfiguracion] Las plantillas por tipo de empresa registran valores iniciales como `Mesa/Mesas`, `Habitacion/Habitaciones`, `Bahia/Bahias`, `Zona/Zonas` y `Consultorio/Consultorios`.
- [Alcance] No agrega tablas, rutas, permisos ni dependencias; reutiliza `empresa_estacion_prefs.estaciones_config`.

## [2026-05-20] Datáfonos POS multiempresa
- [Backend] Nuevo `/api/empresa/datafonos` permite configurar terminales Redeban, CredibanCo, Bold y BBVA por empresa, iniciar pagos y consultar confirmaciones.
- [Pagos] La respuesta del proveedor se normaliza a `pendiente`, `aprobado`, `rechazado` o `error`, y se valida contra monto/referencia antes de cerrar el carrito.
- [Seguridad] Las credenciales se referencian con `env:*`; no se guardan claves reales ni se imprimen secretos.
- [Alcance] Agrega tablas `empresa_datafonos_config` y `empresa_datafonos_transacciones`, sin dependencias nuevas.

## [2026-05-20] Pagos QR en carrito
- [Configuracion] `Configuracion > Carrito unificado` permite activar pago QR y registrar varias cuentas receptoras con proveedor, llave/cuenta, comercio, payload oficial/plantilla e instrucciones.
- [Carritos] El carrito muestra `QR de pago`, permite elegir la cuenta receptora y genera el codigo localmente con `/vendor/qrcode.min.js`.
- [Pagos] Al usar el QR, el cobro se aplica como `transferencia_bancaria` con referencia automatica para conservar caja, turnos, reportes, facturacion y pagos mixtos.
- [Alcance] Sin tablas, endpoints ni dependencias nuevas; el payload oficial/API de Bre-B o Nequi se configura cuando el comercio tenga la plantilla o credenciales del proveedor.

## [2026-05-20] Facturacion offline para carritos
- [Carritos] El carrito unificado puede operar en modo sin internet cuando la empresa lo active desde Configuracion > Carrito unificado.
- [Frontend] Al perder conexion se muestra aviso flotante persistente, se imprime comprobante/factura provisional y la venta queda en cola local por empresa.
- [Seguridad operativa] Si el mismo carrito ya tiene una venta offline pendiente, el boton de pago queda bloqueado hasta sincronizar para evitar duplicados por doble clic o reintento del cajero.
- [Backend] Nuevo `/api/empresa/offline_ventas` sincroniza ventas con idempotencia por `sync_key`, actualizando carrito, inventario, caja del usuario, metricas y documento de venta.
- [Configuracion] La marca `OFFLINE/Pendiente de sincronizar` del impreso queda activa por defecto y puede desactivarse con un check.
- [Alcance] Sin dependencias externas; agrega tabla `empresa_ventas_offline_sync` y mantiene aislamiento por `empresa_id`.

## [2026-05-19] Envio y WhatsApp para codigos de descuento
- [Empresa] `Codigos de descuento` agrega acciones por cupon para enviar por correo y compartir por WhatsApp.
- [Backend] Nuevo `POST /api/empresa/codigos_de_descuento?action=enviar_correo`, usando SMTP global configurado y modo de prueba de correos.
- [Frontend] Cada fila muestra `Enviar correo` y `WhatsApp` con mensaje comercial listo: codigo, descuento, vigencia, minimo, usos y enlace de consulta.
- [Alcance] No agrega tablas, permisos ni dependencias; mantiene aislamiento por `empresa_id`.

## [2026-05-19] Reinicio operativo desde backups empresariales
- [Empresa] `Backups empresariales` agrega una seccion para reiniciar datos operativos por fecha o desde todos los tiempos.
- [Seguridad] La ejecucion real exige escribir `REINICIAR EMPRESA {empresa_id}` y puede crear backup previo automatico; `dry_run` previsualiza sin borrar.
- [Backend] Nuevo `reset_operativo` en `/api/empresa/backups`, con catalogo protegido para no borrar configuracion, usuarios, permisos, impresoras, integraciones, tarifas ni preferencias.
- [Alcance] No agrega tablas, permisos ni dependencias; opera solo sobre tablas con `empresa_id`.

## [2026-05-19] Codigos de descuento y asesor de un solo uso
- [Carritos] Los codigos de descuento quedan consumidos una sola vez por empresa; anular, reabrir o revertir un carrito conserva la auditoria y no devuelve el cupo.
- [Licencias] La promocion por codigo de asesor solo descuenta si esta activa y con porcentaje configurado; si la empresa ya uso asesor en pagos/activaciones/comisiones, no vuelve a aplicar descuento.
- [Backend] El resumen de checkout y las pasarelas recalculan la regla en servidor antes de cobrar o activar sin pago.
- [Alcance] No agrega tablas, dependencias ni permisos; mantiene aislamiento por `empresa_id`.

## [2026-05-19] Compactacion POS del reporte de turno
- [UX] El reporte de turno en POS 80mm usa dos columnas para datos del turno, resumen financiero y detalle de ventas.
- [Impresion] La vista actual de corte y los reportes historicos imprimibles aprovechan mejor el ancho del ticket y reducen el largo del papel.
- [Alcance] No cambia backend, endpoints, permisos ni tablas.

## [2026-05-19] Docker VPS portable desde Super Administrador
- [Super] Se agrega `Docker VPS` en Plataforma para revisar estado del paquete Docker y descargar un `.tar.gz` portable.
- [Backend] Nuevo endpoint `/super/api/docker_portabilidad?action=status|download`, exclusivo de `super_administrador`.
- [Seguridad] La descarga excluye `.env`, secretos, llaves, uploads, descargas, backups, logs, caches, evidencias y datos runtime.
- [Deploy] La imagen backend contiene `/app/project_export` como snapshot limpio para exportar desde el contenedor.

## [2026-05-19] Facturacion electronica Ecuador
- [Backend] Nuevo endpoint independiente `/api/empresa/facturacion_electronica/ecuador` para configurar y validar checklist Ecuador/SRI sin usar DIAN Colombia ni DGI Panama.
- [Permisos] Ecuador/SRI usa modulo independiente `facturacion_ecuador`, pagina `linkFacturacionEcuador` y wrapper `WithEmpresaFacturacionEcuadorPermissions`.
- [Frontend] Nueva pagina `facturacion_electronica_ecuador.html` enlazada como `Ecuador / SRI` en el submenu de facturacion electronica cuando el pais detectado es EC y la licencia lo permite.
- [Normativa] El perfil EC queda basado en SRI: comprobantes de venta, retencion y documentos complementarios, firma electronica, Facturador SRI o proveedor/sistema propio, RIDE, factura, nota credito, nota debito, retencion y guia de remision.
- [Alcance] No agrega tablas ni dependencias; el transporte real SRI/proveedor se parametriza por proveedor y `api_base_url`.

## [2026-05-19] Facturacion electronica por pais y licencia
- [Permisos] Ecuador/SRI y Panama/DGI usan modulos independientes `facturacion_ecuador` y `facturacion_panama`, paginas `linkFacturacionEcuador` y `linkFacturacionPanama`, y wrappers propios.
- [Licencias] La licencia puede activar Ecuador o Panama sin activar DIAN Colombia, y DIAN Colombia no habilita esos paises automaticamente.
- [Frontend] El submenu `Facturacion electronica` permanece como contenedor; sus paginas internas se muestran segun pais detectado automaticamente y permisos efectivos.
- [Operacion] Colombia muestra configuracion DIAN, pruebas DIAN y proveedores de firma; Panama muestra `Panama / DGI` cuando el pais detectado es PA y la licencia lo permite.

## [2026-05-19] Facturacion electronica Panama
- [Backend] Nuevo endpoint independiente `/api/empresa/facturacion_electronica/panama` para configurar y validar checklist Panama/DGI sin usar DIAN Colombia.
- [Frontend] Nueva pagina `facturacion_electronica_panama.html` enlazada como `Panama / DGI` en el submenu de facturacion electronica.
- [Normativa] El perfil PA queda basado en SFEP/DGI: Facturador Gratuito o PAC, declaracion jurada en e-Tax2.0, firma electronica, RUC/DV, CAFE/CUFE/QR y documentos factura/nota credito/nota debito.
- [Alcance] No agrega tablas ni dependencias; el transporte real PAC/DGI se parametriza por proveedor y `api_base_url`.

## [2026-05-19] Caja y turno por usuario
- [Backend] Cierres de caja, aperturas automaticas, pagos, abonos e ingresos/egresos se resuelven por usuario autenticado dentro de la empresa.
- [Base de datos] `empresa_cierres_caja` usa unicidad por `empresa_id`, sucursal, caja, fecha, turno y `usuario_creador`; se eliminan indices legacy sin usuario y se agrega indice para localizar cajas abiertas por usuario.
- [Operacion] Dos usuarios de la misma empresa pueden trabajar al mismo tiempo sin mezclar caja, turno ni reporte; el mismo usuario reutiliza su caja abierta si intenta abrirla de nuevo.
- [Flujo] El cajero puede abrir/reutilizar caja desde `Corte de Caja`; si cobra desde carrito sin caja abierta, el sistema intenta abrir una automaticamente para ese usuario. El cierre se hace al generar/revisar el reporte de turno, imprimir si aplica y cerrar caja/turno.
- [Estaciones] El estado de estaciones queda comun por empresa y el tablero se refresca cada 3 segundos para que varias cajas vean cambios de ocupacion/disponibilidad.
- [Alcance] No cambia permisos, rutas ni dependencias.

## [2026-05-19] Documentos imprimibles blanco y negro
- [UX] Facturas, ventas, notas, documentos imprimibles y reportes de turno/corte ya no heredan la apariencia clara u oscura del panel.
- [Impresion] Las vistas previas y ventanas de impresion usan hoja blanca, texto negro/gris, bordes simples y sin sombras para verse como saldran en papel o POS.
- [Logos] Se respeta la configuracion existente de logo empresarial y logo del sistema; ventas y facturas esperan la configuracion avanzada antes de pintar la primera vista previa y en documentos imprimibles se renderizan en escala de grises.
- [Alcance] No cambia backend, endpoints, permisos ni tablas.

## [2026-05-19] Reporte de turno POS legible
- [UX] El resumen financiero en POS 80mm muestra importes monetarios en una sola linea cuando caben en el rollo.
- [Impresion] Aplica al reporte actual de corte y al historico de reportes de turnos.
- [Alcance] No cambia backend, endpoints, permisos ni tablas.

## [2026-05-19] Prueba operativa Motel Calipso
- [Backend] Pago y anulacion de carritos usan helpers transaccionales compatibles con PostgreSQL en las rutas tocadas por cierre, anulacion, restauracion de inventario y descuentos.
- [Reportes] El pago persiste `descuento_total`, dejando correcto el total de descuentos en el reporte de turno.
- [Alcance] No cambia tablas, rutas, permisos ni dependencias.

## [2026-05-19] Login de usuarios operativos
- [Backend] Se corrige `UpsertAdminEmpresaCompartidaAcceso` para insertar `estado` y `observaciones` junto con el acceso empresarial creado al iniciar sesion.
- [Operacion] El login de cajeros y otros usuarios de empresa ya no falla por SQL de columnas/valores al asegurar el acceso compartido.
- [Alcance] No cambia tablas, rutas, permisos efectivos ni dependencias.

## [2026-05-19] Reporte de turno profesional compacto
- [UX] El reporte de turno actual e historico queda como informe operativo profesional: compacto, completo, con datos ordenados fila a fila y sin cuadriculas.
- [Impresion] Se mantiene compatibilidad con POS 80mm y papel grande.
- [Ventas] Las ventas y facturas electronicas conservan el estilo visual de factura electronica en POS y carta.
- [Alcance] No cambia backend, permisos ni endpoints.

## [2026-05-19] Titulo compacto del login
- [UX] `Powerful Control System` queda mas compacto, en una sola fila y con fuente mas cuadrada.
- [Alcance] No cambia autenticacion, permisos ni backend.

## [2026-05-19] Instalar app desde login
- [UX] El login agrega el boton `Instalar app` junto al acceso de inicio.
- [PWA] Se agregan manifest, service worker e iconos para que Chrome/Edge/Android puedan instalar la app y crear acceso.
- [Compatibilidad] En iOS o navegadores sin prompt, el boton muestra la indicacion para usar el menu del navegador.
- [Alcance] No cambia autenticacion, permisos ni backend.

## [2026-05-19] Titulo del login
- [UX] `Powerful Control System` queda mas grande en el login de administradores, con ajuste responsive para movil.
- [Alcance] No cambia autenticacion, permisos ni endpoints.

## [2026-05-19] Botones de corte de caja
- [UX] Los botones de `Lectura rapida` en `corte_de_caja.html` quedan con color uniforme y mas claro que el panel.
- [Alcance] No cambia backend, permisos, endpoints ni tablas.

## [2026-05-19] Cajas multiples por empresa
- [Configuracion] `empresa_configuracion_general` permite activar/desactivar cajas simultaneas y definir un limite interno por empresa.
- [Backend] Abrir, reabrir y abrir caja automatica para cobro validan caja activa, cupo por empresa, cupo de licencia y cajas abiertas.
- [Reportes] `Ver reporte de mi turno` mantiene caja, turno, sucursal y `cierre_caja_id` para no mezclar reportes entre cajas.

## [2026-05-19] Impresora POS default global
- [Backend] `empresa_impresoras` centraliza helpers para asegurar `POS_80MM` como impresora activa y predeterminada por empresa.
- [Operacion] `set_pos80_config -all` aplica el default a todas las empresas activas, asignando `general`, `corte_caja`, `turno_reporte` y `cajon_monedero`.
- [Alta de empresas] Las empresas nuevas intentan quedar preparadas con POS 80mm al crearse.

## [2026-05-18] POS 80mm para reporte de turno
- [Configuracion] El formato predeterminado del reporte de corte pasa a `pos` para empresas sin configuracion previa o restauradas.
- [Frontend] Corte de caja, configuracion y reportes historicos priorizan `Ticket POS 80mm`; historicos consultan la configuracion por `empresa_id` y bloquean carta/grande cuando la empresa esta en POS.
- [Operacion] Se agrega `backend/tools/set_pos80_config` para activar una impresora POS 80mm predeterminada y asignarla a funcionalidades de caja por empresa.

## [2026-05-18] Contraste de efectivo esperado en caja
- [UX] `corte_de_caja.html` corrige el bloque final `Efectivo esperado en caja` para evitar fondo blanco con texto blanco en modo oscuro.
- [Alcance] No cambia backend, permisos ni tablas.

## [2026-05-18] Contraste del reporte de turno en modo oscuro
- [UX] `corte_de_caja.html` y `reportes_turnos.html` corrigen textos oscuros del reporte cuando el panel esta en modo oscuro.
- [Impresion] La salida impresa conserva fondo blanco y texto oscuro.
- [Alcance] No cambia backend, permisos ni tablas.

## [2026-05-18] Cerrar turno, imprimir y cerrar sesion
- [Frontend] `corte_de_caja.html` agrega el boton `Cerrar turno e imprimir reporte`.
- [Operacion] La accion guarda el cierre, imprime el reporte y cierra sesion con `/auth/logout` al finalizar la impresion.
- [Alcance] No cambia backend, permisos ni tablas.

## [2026-05-18] Descuentos en reporte de turno
- [Backend] El resumen de corte acumula `descuentos_total` y `descuentos_cantidad` desde ventas cerradas.
- [Configuracion] `empresa_corte_caja_configuracion` agrega `mostrar_total_descuentos` activo por defecto.
- [Frontend] El reporte actual y los historicos muestran `Total descuentos`; las exportaciones incluyen la fila de resumen.

## [2026-05-18] Cantidad de ventas en reporte de turno
- [Backend] `empresa_corte_caja_configuracion` agrega `mostrar_cantidad_ventas` activo por defecto.
- [Frontend] El reporte de turno actual y los reportes historicos muestran `Cantidad de ventas` en el resumen.
- [Reportes] Las exportaciones historicas incluyen una fila de resumen con el conteo de ventas del turno.

## [2026-05-18] Tema del reporte de mi turno
- [UX] `corte_de_caja.html` corrige la vista previa de `Ver reporte de mi turno` para respetar el modo claro/oscuro del panel.
- [Impresion] La salida impresa conserva fondo blanco, texto oscuro y bordes legibles para carta, ejecutivo y POS.
- [Alcance] No cambia backend, endpoints, permisos ni tablas.

## [2026-05-18] Reporte de turno en papel grande y POS
- [Frontend] `corte_de_caja.html` compacta el modo POS 80mm para imprimir detalle de ventas como bloques plantillas con etiquetas.
- [Frontend] `reportes_turnos.html` agrega selector de papel grande/POS para la vista imprimible historica.
- [QA] Se valida que papel grande conserve ancho carta y que POS no genere desborde horizontal.

## [2026-05-18] Reportes de turnos historicos
- [Frontend] Nueva pagina `Reportes de turnos` en el submenu de reportes, con listado de turnos antiguos, filtros, vista imprimible, imprimir, compartir, exportar y enviar por email.
- [Backend] `/api/empresa/corte_caja` agrega acciones historicas para listar cierres, reconstruir un reporte por `cierre_caja_id`, exportarlo en `json/csv/txt/xls/pdf` y enviarlo como adjunto.
- [Permisos] Se agrega `linkReportesTurnos` bajo modulo `reportes` con accion lectura, separando consulta historica del cierre operativo de caja.

## [2026-05-18] Reporte de turno configurable
- [Backend] `/api/empresa/corte_caja` incluye NIT de empresa y fechas de entrada/salida por venta para el detalle del turno.
- [Base de datos] `empresa_corte_caja_configuracion` agrega checks de encabezado, datos de empresa, usuario, consecutivo, columnas del detalle y totales de productos/servicios por `empresa_id`.
- [Frontend] `corte_de_caja.html` renderiza el reporte de turno con encabezado, detalle ordenado por fecha/hora de venta y resumenes de caja; `configuracion.html` permite activar/desactivar los bloques.
- [QA] Se agregan pruebas de default profesional y se validan scripts inline de las pantallas afectadas.

## [2026-05-18] Panel ejecutivo super administrador
- [Frontend] `web/super/licencias_resumen.html` reemplaza el centro de mando saturado por un tablero ejecutivo compacto.
- [UX] Se dejan 6 KPIs unicos, prioridades, accesos clave e incidentes recientes, retirando bloques repetidos de negocio, servicios y costos.
- [Alcance] No cambia backend, endpoints, permisos, tablas ni dependencias.

## [2026-05-18] Login administrador
- [UX] `web/login.html` cambia el titulo superior a `Powerful Control System`.
- [UX] `Acceso de administradores` queda superpuesto sobre la imagen lateral y `Ir al inicio` queda debajo de la tarjeta del login.
- [Alcance] No cambia autenticacion, reCAPTCHA, permisos, tablas ni endpoints.

## [2026-05-18] Fix formato monetario por empresa
- [Backend] `empresa_configuracion_avanzada` normaliza columnas legacy booleanas a enteros `0/1` antes de guardar configuracion avanzada.
- [Backend] El `UPSERT` de configuracion avanzada usa `sqlNowExpr()` para fechas runtime PostgreSQL.
- [Backend] `configuracion_operativa` guarda configuracion, roles, politicas e historial con `RETURNING id`, reparando snapshots y rollback en PostgreSQL.
- [QA] Se agrega regresion para el guardado de configuracion monetaria/numerica y la migracion ligera de flags.
- [QA] Se agrega regresion para impedir `LastInsertId()` en los guardados operativos de configuracion.

## [2026-05-18] Corte de caja: acciones y texto de estaciones
- [Frontend] `corte_de_caja.html` mueve `Generar corte`, `Ver reporte de mi turno`, `Corte automatico`, `Cerrar turno` e `Imprimir seleccion` dentro de `Lectura rapida` como botones plantillas.
- [UX] `Ver reporte de mi turno` centra el reporte como hoja imprimible en pantalla y respeta el formato `Carta`, `Ejecutivo` o `POS`.
- [Frontend/Backend] Los textos visibles del control por sensores usan `Estaciones ocupadas sin factura`; se actualizan etiquetas visibles de reportes, tarifas y preconfiguraciones.
- [Fix] `Ver reporte de mi turno` limpia el cierre/caja automaticos previos antes de consultar el reporte del usuario actual.

## [2026-05-18] Panel sin tarjeta de mercado
- [Frontend] `web/administrar_empresa/panel.html` elimina la tarjeta `Mercado en contexto`.
- [Operacion] El panel ya no consulta indicadores externos de divisas, criptoactivos ni ETFs al cargar.
- [Alcance] No cambia backend, permisos, tablas ni endpoints.

## [2026-05-18] Configuracion del reporte de corte en Configuracion
- [Frontend] `corte_de_caja.html` ya no muestra las tarjetas de configuracion del reporte ni de reportes a imprimir.
- [Frontend] `configuracion.html` agrega la seccion `Reporte de corte` con formato, reportes a imprimir, metricas y botones de guardar/restaurar.
- [Alcance] Reutiliza `/api/empresa/corte_caja/configuracion`; no cambia tablas ni endpoints.

## [2026-05-18] Corte de Caja en operacion y ventas
- [Frontend] `web/administrar_empresa.html` muestra el boton `Corte de Caja` debajo de `Estaciones`.
- [Navegacion] El acceso abre `corte_de_caja.html` con el mismo contexto usado al hacer clic en la tarjeta Caja de estaciones.
- [Permisos] `linkCorteCaja` queda visible como acceso operativo del menu principal; no crea endpoint ni ruta nueva.

## [2026-05-18] Reporte de mi turno desde estacion Caja
- [Frontend] `web/administrar_empresa/corte_de_caja.html` agrega el boton `Ver reporte de mi turno`.
- [Backend] `/api/empresa/corte_caja?action=reporte_mi_turno` reutiliza el filtro seguro de caja abierta del usuario autenticado.
- [UX] El boton muestra e imprime el reporte del turno actual sin habilitar el guardado/cierre accidental del turno.

## [2026-05-17] Venta directa con acciones del carrito de estaciones
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` trata venta directa y estaciones como la misma vista enfocada tambien para la tarjeta de acciones.
- [UX] Los botones operativos del carrito, incluido `Abonos`, cargan en venta directa con la misma configuracion global del carrito.
- [Compatibilidad] No cambia endpoints, tablas, permisos ni el carrito canonico `VENTA-DIRECTA-{empresa_id}-0`.

## [2026-05-17] Menu flotante sin compartir WhatsApp
- [Frontend] `web/menu.js` retira `Compartir por WhatsApp` del submenu `Utilidades`.
- [UX] `Cambiar apariencia` queda como penultimo elemento del panel flotante y `Cerrar sesion`/`Iniciar sesion` queda de ultimo.
- [Alcance] No cambia backend, permisos, endpoints ni configuracion de WhatsApp del portal publico.

## [2026-05-17] Inventario debajo de operacion y ventas
- [Frontend] `web/administrar_empresa.html` mueve el grupo `Inventario y compras` justo debajo de `Operacion y ventas`.
- [Alcance] Solo cambia el orden visual del submenu empresarial; no altera rutas, permisos, endpoints ni tablas.

## [2026-05-17] Abonos operativos en carrito de estacion
- [Backend] Se agrega `carrito_compra_abonos` con aislamiento por `empresa_id` y endpoints `action=abonos|abono` en `/api/empresa/carritos_compra`.
- [Pago] `pagar_estacion` valida abonos registrados, descuenta el saldo final a cobrar y conserva el total de la cuenta para documento/venta.
- [Frontend] El boton `Abonos` del carrito de estacion abre una tarjeta para registrar/listar abonos y el desglose muestra cuenta, abonos y saldo final.
- [QA] Validado con pruebas Go enfocadas, parseo JS y prueba visual Playwright mock: abono COP 30000 sobre cuenta COP 100000 deja `total_pagado=70000`.

## [2026-05-17] Venta directa con carrito igual a estaciones
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` usa la misma vista enfocada de estaciones para venta directa.
- [UX] Items, totales, lector y controles de pago cargan en el panel operativo compartido, evitando que venta directa caiga al grid general antiguo.
- [Compatibilidad] Se conserva el carrito canonico `VENTA-DIRECTA-{empresa_id}-0`; no cambian endpoints, tablas ni permisos.

## [2026-05-17] Favoritos del panel como botones
- [Frontend] `web/administrar_empresa/panel.html` ajusta los favoritos de `Accesos rapidos` para verse y comportarse como botones de accion.
- [UX] Se agregan borde destacado, fondo accionable, cursor, estado activo y foco visible sin cambiar rutas, permisos ni almacenamiento local.

## [2026-05-17] Comunicaciones super unificadas
- [Frontend] `web/super_administrador.html` crea el modulo principal `Comunicaciones`.
- [Navegacion] El modulo agrupa mantenimiento, mensajes masivos, alertas sistema, Gmail SMTP, alertas de licencia y WhatsApp portal.
- [Compatibilidad] No cambia endpoints, permisos ni formularios; solo consolida la ubicacion visual del submenu super.

## [2026-05-17] Mantenimiento super como modulo principal
- [Frontend] `web/super_administrador.html` mueve `Mantenimiento sistema` a un grupo propio del menu principal super.
- [Frontend] `web/super/mantenimiento_sistema.html` agrega `Eliminar alertas viejas` para limpiar avisos desactivados o con fecha anterior a hoy.
- [Backend] `/super/api/config/mantenimiento?action=limpiar_viejos` elimina alertas viejas y resincroniza el aviso visible para empresas sin tocar el bloqueo real.
- [QA] Se agrega prueba de filtrado de avisos viejos e inactivos.

## [2026-05-17] Configuracion super por paginas
- [Frontend] `web/super_administrador.html` agrega el grupo `Configuracion` con paginas independientes para consumos, RustDesk, limitaciones, OnlyOffice, Voz IA, Epayco, Wompi/Nequi, Gmail, alertas de licencia, WhatsApp, reCAPTCHA, 2FA, IA global y respaldo.
- [Frontend] `web/super/configuracion/*.html` son paginas contenedoras por seccion y `web/super/configuracion_avanzada.html` soporta modo aislado `?single=1&section=...` para cargar solo la tarjeta correspondiente.
- [Compatibilidad] Los formularios, endpoints y botones de guardar existentes se conservan; no cambia base de datos ni backend.

## [2026-05-17] Mantenimiento del sistema como pagina super
- [Backend] `/super/api/config/mantenimiento` agrega gestion de avisos individuales con `action=desactivar|eliminar`, persistidos en `mantenimiento_programado.avisos_json`.
- [Frontend] `web/super/mantenimiento_sistema.html` queda como pagina independiente en el submenu principal de Super Administrador y muestra tabla de avisos programados con acciones para cargar, desactivar o eliminar.
- [Navegacion] `web/super_administrador.html` agrega el acceso `Mantenimiento`; `web/super/configuracion_avanzada.html` deja de contener esa tarjeta.
- [Seguridad] Desactivar/eliminar avisos no cambia el bloqueo real `mantenimiento_activo`.
- [QA] Se agregan pruebas enfocadas en `backend/handlers/super_mantenimiento_handlers_test.go`.

## [2026-05-17] Panel primero en administrar empresa
- [Frontend] `web/administrar_empresa.html` elimina el grupo `Inicio` y deja `Panel` como primer boton directo del menu lateral.
- [Backend] `backend/handlers/empresa_permisos.go` retira `linkInicio` del catalogo de paginas; `linkPanelEmpresa` queda como acceso general.
- [Alcance] No cambia rutas, endpoints, tablas ni dependencias.

## [2026-05-17] Carrito simplificado como default por tipo de empresa
- [Backend] `defaultEmpresaPreconfigCarritoUI()` usa el preset operativo limpio para empresas nuevas.
- [Backend] `ApplyDefaultCarritoUIToExistingEmpresaPrefs` aplica el mismo preset a empresas antiguas al arrancar; tambien crea una configuracion base para empresas sin `estaciones_config`.
- [Frontend] Los defaults de `configuracion_carrito_de_compra_empresa.html`, `configuracion_de_estaciones.html`, `carrito_de_compras.html` y `estaciones.html` quedan alineados.
- [UX] Activos por defecto: buscar productos, catalogo, lector, items, totales, acciones, valores por medio de pago, pago mixto y pagar.
- [UX] Apagados por defecto: cobro avanzado, descuentos, propina, comision, desglose de cobro y lavador/comision.
- [QA] Se agregan pruebas enfocadas del preset y de la migracion de empresas antiguas.
- [Alcance] Sin cambios de APIs, tablas, permisos ni dependencias.

## [2026-05-17] Carrito mueve Buscar Productos junto a Agregar
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` reubica el boton de productos por botones junto al boton `Agregar` y lo rotula `Buscar Productos`.
- [Frontend] `web/administrar_empresa/buscar_producto_botones.html` deja la barra `Buscar producto` con placeholder para escribir el nombre.
- [Alcance] Sin cambios de backend, APIs, tablas, permisos ni dependencias.

## [2026-05-17] Carrito sin encabezado descriptivo operativo
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` oculta el encabezado superior en venta directa o estacion.
- [UX] Se elimina el texto `Venta directa usa una sola caja abierta...` para que la pantalla empiece directamente desde los controles del carrito.
- [Alcance] Sin cambios de backend, APIs, tablas, permisos ni dependencias.

## [2026-05-17] Carrito calcula cambio por efectivo recibido
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` agrega el campo `Efectivo recibido` en la seccion de valores por medio de pago.
- [UX] El carrito calcula `Cambio a devolver` cuando el recibido cubre el efectivo esperado, o `Falta recibir` cuando no alcanza.
- [Alcance] Sin cambios de backend, APIs, tablas, permisos ni dependencias.

## [2026-05-17] Carrito oculta secciones avanzadas por check
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` agrega los checks `Mostrar opciones de cobro` y `Mostrar lavador`.
- [UX] `Cobro y estados del carrito` y `Lavador` quedan apagados por defecto y se abren solo cuando el usuario los necesita.
- [Operativo] Al ocultar opciones de cobro se limpian descuento, devolucion, referencia y propina para no aplicar valores invisibles.
- [Alcance] Sin cambios de backend, APIs, tablas, permisos ni dependencias.

## [2026-05-17] Administrar empresa abre Panel por defecto
- [Frontend] `web/js/administrar_empresa.js` usa `linkPanelEmpresa` como pagina preferida de arranque del iframe.
- [UX] El usuario entra primero al tablero de administracion, aunque el grupo `Operacion y ventas` siga arriba en el menu.
- [Alcance] Sin cambios de permisos, APIs, tablas ni dependencias.

## [2026-05-17] Administrar empresa simplifica Operacion y ventas
- [Frontend] `web/administrar_empresa.html` deja el grupo `Operacion y ventas` solo con `Venta directa` y `Estaciones`.
- [Frontend] `Venta publica`, `Red social empresarial`, `Codigos de descuento` y `Chat y tareas` pasan al grupo `Canales digitales y colaboracion`.
- [Frontend] `Reservas` se mueve a `Soluciones por negocio`.
- [Permisos] `backend/handlers/empresa_permisos.go` actualiza los grupos administrativos de esas paginas y mueve `Punto de venta / TPV` a `Permisos base de ventas`, sin cambiar acciones, rutas ni endpoints.
- [Alcance] Sin tablas ni dependencias nuevas; mantiene aislamiento por `empresa_id`.

## [2026-05-17] Facturacion electronica DIAN revisada
- [Backend] `FacturacionPaisVistaFor("CO")` deja de referenciar el modulo ERP/documental viejo y apunta a la subpagina canonica de pruebas DIAN.
- [DIAN] El set base para software propio/proveedor tecnologico queda en 60 facturas, 20 notas debito y 20 notas credito, editable segun el objetivo del portal DIAN.
- [Frontend] `facturacion_electronica_pruebas_dian.html` muestra esos valores base y advierte que deben ajustarse si DIAN asigna otro objetivo.
- [Documentacion] Se registra que la configuracion esta separada por empresa/pais, que Colombia conserva pruebas DIAN aparte y que el correo automatico actual no adjunta XML/PDF fiscal certificado.

## [2026-05-17] Estaciones con primer clic solo activa
- [Frontend] `configuracion_de_estaciones.html` agrega el check `Primer clic solo activa`.
- [Frontend] `estaciones.html` usa `solo_activar_primer_clic` para que el primer clic active la estacion sin abrir carrito y el segundo clic abra el carrito ya activo.
- [Compatibilidad] `abrir_carrito_al_activar=false` sigue funcionando como alias historico.
- [QA] Validado con parseo JS, pruebas Go enfocadas y capturas visuales del check, primer clic y segundo clic.
- [Alcance] Sin tablas, endpoints ni dependencias nuevas.

## [2026-05-17] Reporte de turno y caja verificado
- [Frontend] `web/administrar_empresa/reportes_ejecutivos.html` agrega filtros de usuario/cajero, caja, turno y cierre ID para consultar/exportar `reporte_de_turno`.
- [Backend] `backend/handlers/reportes_catalogo_test.go` protege que `reporte_de_turno` siga en el catalogo profesional de reportes.
- [QA] Se valido visualmente `Corte automatico` -> `Turno cerrado` -> impresion, el reporte automatico de ultimos movimientos por caja/usuario actual y la vista previa filtrada del reporte de turno.
- [Alcance] Sin tablas, endpoints ni dependencias nuevas; mantiene aislamiento por `empresa_id`.

## [2026-05-15] Venta directa abre Buscar por botones
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` prepara o recupera el carrito enfocado antes de abrir `Buscar (botones)`, especialmente en `modo=venta_directa`.
- [Frontend] `web/administrar_empresa/buscar_producto_botones.html` conserva `modo=venta_directa` en el retorno, lo envia en las APIs de carrito/items y acepta respuesta puntual o lista filtrada al resolver el carrito.
- [Backend] `GET /api/empresa/carritos_compra?empresa_id=...&id=...` devuelve el carrito solicitado en vez de ignorar `id` y responder siempre el listado completo.
- [Permisos] Las subrutas `/api/empresa/carritos_compra/*` tambien se resuelven contra `linkVentaDirecta` cuando traen contexto de venta directa.
- [Alcance] Sin tablas ni dependencias nuevas; conserva permisos e aislamiento por `empresa_id`.

## [2026-05-15] Administrar empresa prioriza Venta directa
- [Frontend] `web/administrar_empresa.html` mueve `Operacion y ventas` al inicio del menu; `Venta directa` queda como primer acceso y `Estaciones` como segundo.
- [Frontend] `web/js/administrar_empresa.js` prioriza `linkVentaDirecta` en el orden de enlaces y en el iframe inicial cuando el permiso esta visible.
- [Permisos] `linkCarritoCompras` queda agrupado como `Configuracion - Ventas y cobro`; no cambia ruta, accion ni aislamiento por `empresa_id`.
- [Alcance] Sin tablas ni dependencias nuevas.

## [2026-05-15] Asistencia de empleados reparada y profesionalizada
- [Backend] `backend/db/asistencia_empleados.go` deja de usar SQL runtime incompatible con PostgreSQL en configuracion, cierres, listados, CRUD y marcacion de entrada/salida.
- [Backend] Las operaciones usan `queryRowSQLCompat`, `querySQLCompat`, `execSQLCompat`, `insertSQLCompat` y `sqlNowExpr()`.
- [UX] `Agregar registro` y `Editar` cambian automaticamente a la seccion `Registro`, aunque el submenu interno este en Resumen, Consulta u otra pestaña.
- [UX] Los botones de tabla quedan como acciones compactas visibles; `Entrada ahora` y `Salida ahora` registran la hora actual sin prompts nativos.
- [QA] Validado con `go test ./db ./handlers -count=1` y prueba visual Playwright con API mock: crear, entrada, salida, editar, guardar configuracion, filtrar, limpiar, cerrar periodo y descargar reporte.
- [Alcance] Sin tablas, permisos ni dependencias nuevas; mantiene aislamiento por `empresa_id`.

## [2026-05-15] Centro de reportes sin pagina IA dedicada
- [Frontend] `web/administrar_empresa/reportes_menu.html` retira el boton `Asistente IA`.
- [Frontend] `web/administrar_empresa/reportes_ejecutivos.html` elimina el boton superior y la tarjeta `Asistente IA para reportes`.
- [Limpieza] Se elimina `web/administrar_empresa/reportes_ia_chat.html` y el permiso de pagina `linkReportesIAChat`.
- [Backend] `/api/empresa/reportes_ia_chat` se conserva como soporte tecnico del asistente global en modo reportes, autorizado por el permiso general `linkReportes`.
- [Alcance] Sin tablas ni dependencias nuevas; el centro de reportes queda como pantalla unica de consulta/exportacion.

## [2026-05-15] Submenu de facturacion electronica sin regreso a empresas
- [Frontend] `web/administrar_empresa/facturacion_electronica_menu.html` retira el enlace `Volver a empresas`.
- [UX] El submenu queda enfocado en configuracion, pruebas DIAN, proveedores de firma, facturas electronicas y AIU construccion.
- [Alcance] Sin permisos, endpoints, tablas ni dependencias nuevas.

## [2026-05-15] Nuevo credito compatible con PostgreSQL
- [Backend] `CreateEmpresaCredito` usa expresiones de fecha PostgreSQL via `sqlNowExpr()` en el insert runtime.
- [Backend] La asignacion automatica de codigo y la generacion de cuotas usan `execTxSQLCompat`, evitando `tx.Exec` directo con placeholders SQLite.
- [Causa] El boton `Nuevo credito` podia responder 500 al crear creditos con cuotas porque la transaccion mezclaba `?` y `datetime('now','localtime')`.
- [QA] Validado con prueba Go enfocada y `go test ./db ./handlers -count=1`.
- [Alcance] Sin tablas, permisos, endpoints ni dependencias nuevas; mantiene aislamiento por `empresa_id`.

## [2026-05-15] Productos avisa cuando falta bodega
- [Frontend] `Nuevo producto` valida bodegas activas antes de abrir el formulario.
- [UX] Si no existe bodega activa, se muestra un aviso claro y un boton `Crear bodega`.
- [Resguardo] El guardado de productos nuevos tambien bloquea el caso sin bodega activa.
- [QA] Validado con parseo del script inline y prueba visual Playwright en empresa sin bodegas y con una bodega activa.
- [Alcance] Sin backend, tablas, permisos ni dependencias nuevas.

## [2026-05-15] Compras con proveedores creados previamente
- [Frontend] `compras_avanzadas.html` usa selectores de proveedores activos para proveedor sugerido, cotizacion y recepcion.
- [Frontend] `compras.html` agrega acceso a gestionar proveedores y autocompleta el documento del proveedor elegido.
- [Backend] `compras_avanzadas` valida que `proveedor_id` exista activo en la misma empresa antes de guardar cotizaciones o recepciones.
- [BD] `empresa_compras_recepciones_avanzadas` agrega `proveedor_id` para trazabilidad por proveedor.
- [QA] Validado con `node --check`, `go test ./db ./handlers`, `git diff --check` y prueba visual Playwright con API mock.
- [Alcance] Sin dependencias nuevas; se mantiene PostgreSQL y aislamiento por `empresa_id`.

## [2026-05-15] Menu empresarial con Venta directa primero
- [Frontend] `web/administrar_empresa.html` mueve `Venta directa` al primer lugar de `Operacion y ventas`.
- [Frontend] `Carritos` se retira del grupo operativo principal y queda disponible en `Configuracion > Ventas y cobro`.
- [QA] Validado con Playwright revisando el orden del menu principal y la permanencia de `Carritos` en el submenu de configuracion.
- [Alcance] Sin permisos, rutas, endpoints, tablas ni dependencias nuevas.

## [2026-05-15] Alta de bodegas compatible con PostgreSQL
- [Backend] `backend/db/productos.go` usa `sqlNowExpr()` en `CreateBodega` para `fecha_creacion` y `fecha_actualizacion`, eliminando `datetime('now','localtime')` del insert runtime.
- [QA] `backend/db/productos_bodegas_test.go` blinda que el alta de bodega no vuelva a usar funciones SQLite.
- [Visual] Validado con Playwright en escritorio y movil: abrir `bodega.html`, crear bodega, ver mensaje de exito, KPI actualizado y fila en tabla.
- [Alcance] Sin tablas, permisos, rutas ni dependencias nuevas.

## [2026-05-15] Login y registro movil con controles tactiles
- [Frontend] `web/estilos.css` aumenta a minimo 44 px las areas tactiles de botones, enlaces, inputs y mostrar contrasena dentro de `login-page`.
- [UX] `registrar_nuevo_usuario_administrador.html` marca el body de registro para compactar el chat flotante en celular y evitar que tape el formulario.
- [QA] Validado con Playwright movil en login administrador, login operativo, recuperacion de invitacion, registro operativo por invitacion y registro administrador con mock de endpoints.
- [Alcance] Sin backend, tablas, permisos ni dependencias nuevas.

## [2026-05-15] Venta directa con carrito 0 automatico
- [Frontend] `carrito_de_compras.html?modo=venta_directa` crea o reutiliza `VENTA-DIRECTA-{empresa_id}-0` si no llega `carrito_codigo`.
- [Compatibilidad] Se reutilizan carritos antiguos `VENTA-DIRECTA-{empresa_id}` o `CAJA_DIRECTA` antes de crear uno nuevo.
- [UX] Venta directa deja de mostrarse como estacion y oculta el boton `Regresar a estaciones`.
- [Alcance] Sin tablas, endpoints ni dependencias nuevas.

## [2026-05-15] Empresas compartidas con alcance por rol y modulos
- [Backend] `admin_empresa_compartida` y `admin_empresa_compartida_invitaciones` agregan `nivel_acceso` y `modulos_permitidos`.
- [Permisos] El alcance compartido se aplica sobre modulos, acciones y paginas efectivas despues de licencia, vertical, rol y politicas finas de empresa.
- [UI] `editar_empresa` y el selector de empresas permiten elegir `Solo ver`, `Acceso total` o `Solo ciertos modulos`, muestran el alcance y ofrecen `Dejar de compartir`.
- [Seguridad] Revocar acceso invalida caches de empresa/permisos para retirar el acceso inmediatamente.
- [Alcance] Sin dependencias nuevas; PostgreSQL y aislamiento por `empresa_id`.

## [2026-05-15] Licencia de prueba unica por empresa y asesor comercial
- [Backend] `licencias_activaciones_gratis` agrega `asesor_id` y un indice unico activo por `empresa_id` para impedir mas de una prueba/gratis por empresa.
- [Migracion] El arranque normaliza duplicados activos historicos como `historico_duplicado` antes de crear el indice unico.
- [Checkout] La prueba de 15 dias y la activacion sin pago aceptan `asesor_id`/`codigo_asesor`, validan que el asesor exista, este aceptado y no este inactivo, y dejan trazabilidad en pagos/comisiones.
- [Comercial] La prueba de 15 dias queda con 250 documentos/ventas mensuales y la UI avisa que solo se puede activar una vez por empresa.
- [Alcance] Sin dependencias nuevas; se mantiene PostgreSQL y aislamiento por `empresa_id`.

## [2026-05-14] Nucleo de pantallas empresariales unificado
- [Frontend] `Venta directa` entra directo al carrito unico con `modo=venta_directa` y permiso `linkVentaDirecta`.
- [Frontend] `Productos` usa `administrar_productos.html?view=...` para categorias/proveedores/precios/compras sin wrappers intermedios.
- [Limpieza] Se eliminan los HTML que solo redirigian: `venta_directa.html`, `documentos_onlyoffice_menu.html`, `crm_ventas_avanzadas.html`, `categorias.html` y wrappers bajo `web/administrar_empresa/productos/`.
- [Backend] La preconfiguracion y configuracion guiada guardan la URL canonica de venta directa.
- [Alcance] Sin tablas, permisos ni dependencias nuevas; se conserva aislamiento por `empresa_id`.

## [2026-05-14] Reportes unificados y limpieza preproduccion
- [Frontend] `web/administrar_empresa/reportes_menu.html` queda con Centro de reportes y Asistente IA.
- [Frontend] `web/administrar_empresa/reportes_ejecutivos.html` agrega catalogo, vista previa y exportacion directa sin saltar a paginas antiguas.
- [Limpieza] Se eliminan `reportes.html`, `reportes_inventario.html`, `reportes_finanzas.html`, `graficos_estadisticas.html` y `backend/handlers/graficos_estadisticas.go`.
- [Datos] Se limpian usuarios/clientes/empresa smoke de la base, se normalizan productos `DEMO-*` a `BASE-*`, se retiran evidencias QA locales antiguas y se elimina el runner que generaba finanzas QA en Motel Calipso.
- [Seguridad] Sin tablas, permisos ni dependencias nuevas; los datasets siguen bajo `/api/empresa/reportes`.

## [2026-05-14] Lobby de juegos uniforme con Doon FPS
- [Frontend] `web/Juegos/menu_juegos.html` deja todas las tarjetas del catalogo con tamano uniforme, portada 16:9 y descripcion.
- [Juegos] Se retiran del lobby los juegos 3D y el emulador generico; queda visible solo `N64` como emulador.
- [Nuevo] `web/Juegos/doon_fps/index.html` agrega un shooter retro original en canvas/raycasting con dos sectores y records.
- [Assets] Portadas raster originales en `web/img/juegos`, sin capturas comerciales ni ROMs.

## [2026-05-14] Carrito con valores editables por medio de pago
- [Frontend] Los campos `Efectivo`, `Debito`, `Credito` y `Otro` del carrito ahora aceptan edicion directa.
- [Pago] Cuando hay varios medios con valor, el carrito envia `metodo_pago=mixto` con `pagos_mixtos`; `Otro` se registra como transferencia bancaria.
- [Validacion] La suma debe coincidir con el total final y las tarjetas/transferencias mantienen referencia obligatoria.
- [Alcance] Sin cambios de backend, tablas, permisos ni dependencias.

## [2026-05-14] Carrito limpia titulos y destaca acciones
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` oculta los textos repetidos de items, lector, valores de pago y acciones.
- [UX] `Carrito de compras` y `Estacion: <nombre>` quedan en una misma fila cuando el ancho lo permite.
- [Estilos] `web/estilos.css` conserva tarjetas planas, pero devuelve apariencia de boton al toolbar de acciones.
- [Alcance] Cambio solo visual; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] Boton pagar y cerrar carrito con reintento visual
- [Backend] `pagar_estacion` y `pagar` pasan a permiso operativo `Crear / registrar` en ventas, evitando bloquear cajeros o roles personalizados que venden pero no tienen acciones gerenciales de aprobacion.
- [Frontend] Si la carga inicial de cajas abiertas falla, el boton `Pagar y cerrar carrito` queda habilitado, informa que puede reintentar y al clic vuelve a cargar cajas antes de enviar el pago.
- [QA] Validado con `go test ./handlers ./db` y prueba visual Playwright simulando 403 inicial de cajas: el boton queda clickeable, recarga `CAJA-1` y envia `PUT action=pagar_estacion` con `cierre_caja_id=11`.
- [Alcance] Sin tablas ni dependencias; mantiene validacion final de caja abierta por `empresa_id` en backend.

## [2026-05-13] Carrito cobra con caja abierta desde permisos de ventas
- [Backend] `backend/handlers/carritos_compras.go` agrega `GET /api/empresa/carritos_compra?action=cajas_abiertas` para listar cajas abiertas necesarias al cobrar sin exigir acceso al modulo financiero.
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` carga cajas abiertas desde el flujo de carritos/ventas y deja de depender de `/api/empresa/finanzas/cierres_caja`.
- [QA] Validado con `go test ./handlers ./db` desde `backend` y prueba visual Playwright mock: boton `Pagar y cerrar carrito` habilitado, caja seleccionada y `PUT action=pagar_estacion` con `cierre_caja_id`.
- [Alcance] Sin tablas, dependencias ni motores nuevos; conserva aislamiento por `empresa_id` y validacion final de caja abierta en backend.

## [2026-05-13] Carrito plano reforzado contra temas posteriores
- [Frontend] `web/estilos.css` agrega un override absoluto al final del archivo para `carrito-flat-page`.
- [UX] Ningun tema, modo tactil o estilo legacy posterior puede volver a mostrar sombras, radios o margenes en tarjetas del carrito.
- [QA] Validado con parseo JS y Playwright revisando 15 tarjetas/formularios con `box-shadow: none`, `border-radius: 0px` y margenes `0px`.
- [Alcance] Cambio solo visual; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] Panel empresarial prioriza ciudad por IP
- [Frontend] `web/administrar_empresa/panel.html` detecta primero la ciudad por IP para clima e indicadores cuando no existe ciudad manual guardada.
- [UX] GPS queda como respaldo si falla la IP o como accion manual desde el boton `Usar GPS`; la ciudad manual guardada conserva prioridad.
- [QA] Validado con parseo JS y prueba visual Playwright mock: IP devuelve `Medellin`, el panel muestra esa ciudad y no consulta GPS.
- [Alcance] Cambio solo frontend; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] Carrito de compras compacto tipo tabla plana
- [Frontend] `web/estilos.css` compacta el layout bajo `carrito-flat-page` quitando gaps, margenes, radios y sombras de tarjetas/contenedores del carrito.
- [UX] Las tarjetas quedan pegadas visualmente entre si, con lectura plana similar a la tabla de items.
- [QA] Validado con Playwright comprobando `gap: 0`, `margin: 0`, `border-radius: 0` y `box-shadow: none` en contenedores clave.
- [Alcance] Cambio solo visual; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] Carrito de compras sin sombras en tarjetas
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` agrega clase scoped `carrito-flat-page` para controlar el estilo visual de esta pagina.
- [UX] `web/estilos.css` deja las tarjetas y formularios del carrito sin `box-shadow`, evitando apariencia 3D o relieve.
- [QA] Validado con parseo JS y prueba visual Playwright comprobando tarjetas del carrito con `box-shadow: none`.
- [Alcance] Cambio solo visual; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] Carrito de estacion mueve contexto al encabezado
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` deja de mostrar `Items de estacion: <nombre>` dentro de la tarjeta del carrito.
- [UX] En modo estacion, el encabezado superior mantiene `Carrito de compras` y muestra debajo `Estacion: <nombre>`, dejando la tarjeta interna como `Items del carrito`.
- [Alcance] Cambio solo visual; no modifica APIs, backend, permisos, tablas ni dependencias.

## [2026-05-13] OnlyOffice abre editor con URL publica del Document Server
- [Backend] `backend/handlers/onlyoffice.go` corrige la carpeta temporal de documentos para no crear `/data/empresas/empresas/{empresa_id}` cuando `PCS_DATA_ROOT` ya apunta a `/data/empresas`.
- [Backend] `editor_config` entrega al navegador una URL publica de OnlyOffice cuando la configuracion global apunta a un hostname interno Docker como `onlyoffice-documentserver`; tambien acepta `ONLYOFFICE_PUBLIC_DOCUMENT_SERVER_URL` o `ONLYOFFICE_BROWSER_DOCUMENT_SERVER_URL` como override operativo.
- [Frontend] `web/administrar_empresa/documentos_onlyoffice.html` valida que `api.js` cargue `DocsAPI.DocEditor` y muestra un mensaje claro dentro del marco si el Document Server no es accesible desde el navegador.
- [QA] Se agrega `backend/handlers/onlyoffice_test.go`; validado con `go test ./handlers -run OnlyOffice -count=1` y prueba visual Playwright del flujo crear y abrir editor.
- [Alcance] Sin tablas, permisos ni dependencias nuevas; conserva aislamiento por `empresa_id` y el wrapper de permisos `documentos_onlyoffice`.

## [2026-05-13] Reportes ejecutivos profesionales
- [Frontend] `web/administrar_empresa/reportes_menu.html` reduce el submenu a entradas de alto nivel: centro ejecutivo, analitica, suite, inventario, finanzas e IA.
- [Frontend] `web/administrar_empresa/reportes_ejecutivos.html` agrega una portada de mando con accesos agrupados por direccion, ventas, inventario, finanzas, contabilidad/fiscal e IA.
- [Frontend] `web/administrar_empresa/reportes_finanzas.html` reemplaza botones placeholder por carga real del dataset `contable_movimientos_financieros` y exportacion por formato usando `/api/empresa/reportes`.
- [Alcance] Sin cambios de backend, base de datos, permisos ni dependencias.

## [2026-05-13] Facturacion electronica separa configuracion y pruebas DIAN
- [Frontend] `web/administrar_empresa/facturacion_electronica.html` elimina el boton `Abrir modulo DIAN / documental` y queda enfocada en configurar pais, firma, DIAN Colombia y parametros avanzados de la empresa.
- [Frontend] `web/administrar_empresa/facturacion_electronica_pruebas_dian.html` concentra diagnostico DIAN, set de habilitacion, conexion/cola y emision documental manual.
- [Navegacion] `web/administrar_empresa/facturacion_electronica_menu.html` agrega `Pruebas DIAN y documentos`.
- [Alcance] Sin cambios de backend, base de datos, permisos ni dependencias.

## [2026-05-13] Control de aseo por estacion
- [Backend] Se agrega `users.control_aseo_estaciones`, `empresa_estacion_aseo_eventos` y `/api/empresa/estacion_aseo` para medir el tiempo desde `estacion_estado_sucia=1` hasta el reporte de aseo terminado.
- [Frontend] `administrar_usuarios.html` permite activar/desactivar el control por usuario; `estaciones.html` permite reportar el aseo con un clic sobre la estacion sucia; `reporte_aseo_estaciones.html` muestra tiempos por estacion.
- [Seguridad] La finalizacion se atribuye al usuario autenticado y conserva aislamiento por `empresa_id`; el reporte queda para roles administrativos/supervision.
- [QA] Validado con pruebas Go dirigidas de `db`/`handlers` y parseo JS inline de las paginas tocadas.

## [2026-05-13] Factura electronica de venta no queda emitida sin acuse fiscal
- [Backend] `backend/handlers/carritos_compras.go` mantiene cerrada la venta/comprobante, pero degrada la factura electronica asociada a `pendiente_emision` cuando Colombia produccion requiere acuse fiscal y la integracion DIAN/proveedor no queda `enviado`.
- [Backend] `backend/handlers/facturacion_electronica.go` bloquea Colombia produccion con proveedor `manual`, `local` o `interno`, para no simular como exitoso un envio que debe pasar por DIAN/proveedor real.
- [Operacion] El resultado conserva `integracion_fiscal`, `cola_reintentos` y una observacion con la causa para que el usuario reprocesar o corregir configuracion sin confundir una factura pendiente con una aceptada.
- [QA] Se agrega cobertura en `backend/handlers/facturacion_documentos_electronicos_test.go` para exigir acuse solo en Colombia produccion y confirmar que `fallido` no sea tratado como enviado.
- [Alcance] No cambia tablas, permisos, endpoints publicos ni dependencias; conserva aislamiento por `empresa_id`.

## [2026-05-13] Logos empresariales activos para Calipso y Bollon
- [Activos] Se agregan `web/uploads/empresa_logos/empresa_7/motel-calipso-logo.svg` y `web/uploads/empresa_logos/empresa_32/gimnasio-el-bollon-logo.svg`.
- [Datos] En produccion se actualiza `empresa_configuracion_avanzada.logo_url` para `empresa_id=7` y `empresa_id=32`, dejando `mostrar_logo` y `mostrar_logo_empresa` activos.
- [UX] Panel empresarial, carrito, facturas, recibos y reportes usan esos logos a traves de la configuracion avanzada existente.
- [Alcance] No se crean tablas, endpoints, permisos ni dependencias nuevas; se reutiliza el almacenamiento `/uploads/empresa_logos`.

## [2026-05-13] Indicadores economicos legibles en movil
- [Frontend] `web/administrar_empresa/panel.html` elimina recortes de texto en la seccion `Indicadores economicos importantes` para pantallas moviles.
- [UX] En celulares las tarjetas usan 2 columnas legibles y pasan a 1 columna en pantallas muy estrechas o bajas; nombre, valor, referencia y variacion quedan completos.
- [Alcance] Cambio solo responsive del panel empresarial; no modifica datos, endpoints, permisos, tablas ni dependencias.
- [QA] Validado con parseo de scripts inline y prueba visual Playwright en 390x844 y 360x640 sin elementos truncados.

## [2026-05-13] Facturacion electronica enlaza proveedores de firma digital
- [Frontend] `web/administrar_empresa/facturacion_electronica.html` agrega el boton `Adquirir Firma Electronica` junto a `Cargar firma` en la tarjeta de carga DIAN Colombia.
- [Frontend] `web/administrar_empresa/proveedores_firma_digital.html` publica la pagina `Proveedores de Firma Digital` dentro del submenu de facturacion electronica.
- [Proveedor externo] La pagina agrega a Sensiyo como opcion para adquirir certificado digital/firma electronica DIAN y abre la compra en `https://sensiyo.co/certificados-digitales/`.
- [Alcance] Cambio solo de navegacion/frontend; no modifica endpoints, tablas, permisos ni dependencias.

## [2026-05-13] Corte automatico de turno en Caja
- [Frontend] `web/administrar_empresa/corte_de_caja.html` agrega el boton `Corte automatico`, que calcula el periodo desde la apertura de la caja abierta del usuario actual hasta la hora del corte.
- [Operacion] Al entrar desde `Estaciones -> Caja` con `auto_generar=1`, el corte usa `mi_caja_actual` y autocompleta fecha inicial, fecha final, usuario, caja, turno y efectivo de apertura.
- [Backend] `cerrarCorteCaja` cierra la caja abierta existente cuando recibe `cierre_caja_id`, evitando crear un cierre duplicado para el mismo turno.
- [QA] Parseo JS inline, `go test ./handlers -count=1`, `go test ./db -run "CierreCaja|Finanzas|Carritos" -count=1`, `git diff --check` y prueba visual Playwright con servidor mock.

## [2026-05-13] Ultimos movimientos de caja por usuario actual
- [Backend] `backend/handlers/corte_caja.go` agrega el alcance `mi_caja_actual` para resolver solo la caja abierta del usuario autenticado y filtrar ventas/productos/movimientos por `cierre_caja_id`, `caja_codigo` y usuario.
- [Frontend] `web/administrar_empresa/estaciones.html` envia `solo_usuario_actual=1` desde el boton `Ver ultimos movimientos`; `web/administrar_empresa/ultimos_movimientos_de_caja.html` consume el corte de caja acotado en vez de listados globales.
- [Seguridad] La URL ya no permite mezclar movimientos de cajas de otros usuarios; se conserva aislamiento por `empresa_id` sin agregar tablas, permisos ni dependencias.
- [QA] `go test ./handlers -count=1`, `go test ./db -run "CierreCaja|Finanzas|Carritos" -count=1`, parseo JS inline con Node y `git diff --check`.

## [2026-05-13] Caja recupera tarjeta compacta con boton de movimientos
- [Frontend] `web/administrar_empresa/estaciones.html` vuelve a mostrar la tarjeta `Caja` como bloque compacto con titulo, totales y boton `Ver ultimos movimientos`, sin el texto descriptivo largo.
- [UX] La tarjeta `Caja` toma el color configurado de estaciones mediante `--station-state-bg`, conservando el boton independiente para ultimos movimientos.
- [Operacion] El clic o teclado sobre la tarjeta completa sigue abriendo `corte_de_caja.html`; solo el boton interno navega a `ultimos_movimientos_de_caja.html`.

## [2026-05-13] Carritos recupera esquema incompleto antes de responder 500
- [Backend] `backend/db/carritos_compras.go` valida y completa todas las columnas requeridas por el listado de carritos, items y metricas de estacion antes de marcar el esquema como listo, con cache por base/esquema PostgreSQL.
- [Backend] El reintento reducido del listado ya no selecciona `ic.item_count` cuando se desactiva el join de items, evitando otro 500 en bases con tablas auxiliares rezagadas.
- [Operacion] Si `/api/empresa/carritos_compra` encuentra columnas o relaciones faltantes en una base empresarial rezagada, refresca migraciones ligeras y reintenta la consulta antes de mostrar error en estaciones.
- [QA] Se agrega `backend/db/carritos_compras_schema_test.go`, pasan pruebas enfocadas de `db` y `handlers`, y se valida visualmente el flujo Estaciones -> Zona 1 -> carrito sin mostrar el error inicial.

## [2026-05-13] Juegos moviles, sonido y records globales
- [Backend] `backend/main.go` registra `EnsureSuperJuegosSchema` y `/api/public/juegos/records` para guardar/consultar records en `super_juegos_records`.
- [Frontend] `web/Juegos/juegos_records.js`, `arcade_embed.js` y `open_game_embed.js` agregan ranking, reporte automatico de puntajes, panel de records, controles tactiles y sonido WebAudio.
- [UX] `web/Juegos/menu_juegos.html` muestra tarjetas uniformes con capturas PNG reales de cada juego en `web/img/juegos/`.
- [Emulador] `/emulador/` suma D-pad tactil y `/Juegos/n64/index.html` embebe el emulador real en una pantalla responsive.
- [QA] Validado con `go test ./...`, `node --check` y Playwright movil/escritorio sobre menu, juegos, N64 y emulador.

## [2026-05-13] Caja en estaciones deja de caer al flujo de carrito
- [Frontend] `web/administrar_empresa/estaciones.html` agrega una guarda explicita para la tarjeta `Caja`, de modo que nunca use la activacion generica de estaciones ni abra `carrito_de_compras.html`.
- [Operacion] Clic o teclado sobre `Caja` siempre abren `corte_de_caja.html` para cierre de turno, corte e impresion del reporte del cajero actual.

## [2026-05-13] Estaciones vuelven a abrir carritos legado por nombre
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` ya no restringe la carga inicial del modo estación al filtro `estacion_id`, para poder encontrar carritos previos de la misma empresa que todavía solo coincidían por nombre.
- [Frontend] Si el carrito rescatado pertenece a la estación pero quedó con identidad vieja, la misma apertura lo actualiza a `EST-{empresa}-{estacion}` y `ESTACION_{id}` antes de activar o recuperar la sesión.
- [Operación] Se elimina el escenario donde una estación como `Zona 1` intentaba recrear un carrito con nombre duplicado y terminaba mostrando `No se pudo abrir esta estación`.

## [2026-05-13] Administrar empresa exige sesion valida antes de abrir modulos
- [Frontend] `web/js/administrar_empresa.js` valida `/me` al arrancar el shell empresarial y redirige a `login.html` cuando la sesion ya expiro o no existe, en vez de dejar cargar menus e iframes protegidos como si hubiera acceso.
- [Frontend] `web/administrar_empresa/carrito_de_compras.html`, `web/administrar_empresa/administrar_clientes.html` y `web/administrar_empresa/bodega.html` convierten `401` en redireccion al login administrativo y `403` en mensaje claro de permiso insuficiente.
- [QA] La verificacion visual sobre `administrar_empresa.html` y modulos de `Carritos`, `Venta directa`, `Clientes` y `Bodegas` confirmo que el problema observable provenia de respuestas `401` del API, no de una caida general del frontend.

## [2026-05-13] Login con imagenes reales y glow adaptable
- [Frontend] `web/login.html` y `web/login_usuario.html` usan imagenes reales PNG en lugar de las ilustraciones previas.
- [Assets] Se agregan `web/img/login-admin-real.png` y `web/img/login-usuario-real.png` como copias estables para despliegue web sin espacios en el nombre.
- [UX] `web/estilos.css` incorpora un iluminado exterior suave que responde bien a apariencia clara y oscura en ambos logins.

## [2026-05-13] Caja en Estaciones abre cierre de turno
- [Estaciones] La tarjeta especial `Caja` en `web/administrar_empresa/estaciones.html` deja de lanzar `carrito_de_compras.html` en modo venta directa.
- [Corte de caja] Al entrar desde estaciones, `web/administrar_empresa/corte_de_caja.html` muestra contexto de caja, boton `Cerrar turno`, enlace `Regresar a estaciones`, auto-generacion del reporte y filtro por `caja_codigo` de la empresa.
- [QA] Validacion visual con Playwright en servidor mock local, confirmando navegacion `Caja -> corte_de_caja.html` y carga del reporte del cajero.

## [2026-05-13] Carritos y venta directa vuelven a cargar en Administrar empresa
- [Backend] `backend/handlers/carritos_compras.go` vuelve a asegurar el esquema antes del listado y `backend/db/carritos_compras.go` reintenta la consulta sin joins opcionales cuando una base vieja todavia no tiene `clientes` o `carrito_compra_items` alineados.
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` protege toda la secuencia de carga inicial con el mismo fallback visual, evitando iframes en blanco si el fallo ocurre antes de `loadCarritos()`.
- [QA] Verificacion visual con Playwright en `carrito_de_compras.html` y `venta_directa.html` sobre servidor mock local, mas `go test ./handlers` y `go test ./db` enfocados.

## [2026-05-13] Diagrama entidad relacion canonico
- [Diagramas] Se agrega `documentos/diagramas/diagrama_entidad_relacion.md` como DER resumido y vigente del proyecto.
- [Imagen] Se agrega `documentos/diagramas/diagrama_entidad_relacion.svg` para visualizacion rapida del modelo relacional.
- [Limpieza] `documentos/diagramas/estructura_del_codigo.md` deja de listar como vigentes diagramas historicos que ya no existen fisicamente.

## [2026-05-13] Documentacion y ayuda actualizadas
- [Docs] Se agrega `documentos/estado_documentacion_2026-05-13.md` como mapa consolidado del estado vigente del proyecto.
- [Indice] `documentos/README.md` y `documentos/descripcion_del_proyecto` quedan alineados con operacion conectada, cajas simultaneas, login global, comunicaciones, soporte, Docker/VPS y validacion.
- [Ayuda] `web/ayuda/ayuda.html` suma secciones de acceso de usuarios, operacion conectada, cajas simultaneas, soporte/comunicaciones, documentos locales y backups.

## [2026-05-13] Cajas simultaneas por licencia
- [Licencias] Se agrega `max_cajas_simultaneas` con default 2 y cupo mayor para el plan global superior.
- [Finanzas] Apertura y reapertura de cajas validan el cupo contra la licencia activa de la empresa.
- [Ventas] Cada pago de carrito se enlaza a una caja abierta para mantener cierres y arqueos separados.
- [Super] La pantalla de licencias permite configurar el maximo de cajas simultaneas.

## [2026-05-13] Conexion obligatoria para operacion y facturacion
- [Facturacion] `web/administrar_empresa/facturacion_electronica.html` retira controles de modo offline/contingencia DIAN y muestra conexion obligatoria.
- [Backend] `backend/handlers/facturacion_electronica.go` ignora banderas offline antiguas y bloquea la emision cuando DIAN/proveedor no esta disponible.
- [Operacion] Las ventas, cobros y facturacion quedan documentadas como flujos que requieren conexion activa con el servidor.

## [2026-05-13] Correos masivos globales
- [Super] Se agrega `web/super/correos_masivos.html` al menu del super administrador para enviar comunicados de politicas, actualizaciones, mantenimiento, seguridad o informacion general.
- [Backend] Nuevo endpoint `/super/api/correos_masivos` con validacion de rol super, vista previa con emails enmascarados, deduplicacion de destinatarios y confirmacion obligatoria.
- [DB] Nuevas tablas `super_correos_masivos` y `super_correos_masivos_destinatarios` en `pcs_superadministrador` para trazabilidad de campanas y resultado por destinatario.

## [2026-05-13] Proporcion de tarjetas en carrusel del index
- [Portal] `web/index.html` corrige la barra horizontal para que sus tarjetas no se estiren verticalmente por la tarjeta mas larga.
- [UX] Se controla el alto en escritorio y se limita la descripcion dentro del carrusel, manteniendo las tarjetas superiores sin cambios.
- [Alcance] Solo frontend; sin cambios de backend, tablas, permisos ni dependencias.

## [2026-05-13] Propinas y comisiones enlazadas a usuarios
- [Datos] `empresa_propinas_movimientos` y `empresa_comisiones_servicio_movimientos` guardan ids de `users` para origen, asignado y lavador.
- [Backend] Los movimientos automaticos/manuales resuelven usuarios por id, correo, documento, nombre o etiqueta `Nombre (correo)` dentro del mismo `empresa_id`.
- [UX] Propinas, comisiones y carrito sugieren usuarios creados y muestran el id vinculado en reportes.
- [Validacion] `go test ./...` en `backend/` y parseo de scripts inline HTML con Node.

## [2026-05-13] Backup profesional junto a Configuracion
- [Empresa] `web/administrar_empresa.html` mueve el acceso de backups al grupo `Administracion`, inmediatamente junto a `Configuracion`.
- [Permisos] Se conserva `linkBackups`, `module=backups` y la misma regla existente de permisos; no hay cambios de backend, tablas ni dependencias.

## [2026-05-13] Imagen profesional en login de usuarios
- [UX] `login_usuario.html` agrega una ilustracion de usuario iniciando sesion frente a un computador.
- [Frontend] La imagen queda en la esquina superior derecha en escritorio y se adapta como visual compacto sobre el encabezado en movil.
- [Activo] Nuevo `web/img/login-usuario-computador.svg`; no hay cambios de backend, tablas, permisos ni dependencias.

## [2026-05-13] Login unico para usuarios de empresa
- [Auth] `login_usuario.html` opera como portal global para usuarios de todas las empresas, sin subdominio empresarial obligatorio.
- [Backend] El login resuelve la empresa por email y clave con consulta global por correo; si hay ambiguedad no abre una empresa arbitraria.
- [Invitaciones] El primer password se valida por `token_invitacion`, email y documento, y los enlaces ya no agregan `empresa_id`.
- [UX] La pantalla mantiene compatibilidad con enlaces antiguos, persiste `empresa_id` al autenticar y redirige a `administrar_empresa.html?id=...`.
- [Validacion] `go test ./...` en `backend/`, `node --check web/js/login_usuario.js` y `git diff --check`.

## [2026-05-13] Aviso de mantenimiento programado
- [Super] `web/super/configuracion_avanzada.html` agrega check para publicar aviso, fecha, hora inicio, hora fin, zona horaria y mensaje publico.
- [Backend] `/super/api/config/mantenimiento` conserva `mantenimiento_activo` y persiste `mantenimiento_programado.*`; `/api/empresa/mantenimiento_programado` expone el aviso activo bajo alcance de empresa.
- [Empresa] `web/administrar_empresa/panel.html` muestra una franja de aviso cuando el check esta activo.
- [Seguridad] El aviso no activa el bloqueo global de mantenimiento; no hay tablas ni dependencias nuevas.

## [2026-05-13] OnlyOffice local editable en una sola pantalla
- [Empresa] `web/administrar_empresa/documentos_onlyoffice.html` permite elegir tipo, crear documento, abrir OnlyOffice embebido en la misma vista y guardar el resultado en el PC/celular.
- [Backend] `/api/empresa/documentos?action=create_edit_local` crea una sesion temporal editable y `/api/empresa/documentos?action=download&delete=1` descarga el archivo final eliminando la copia temporal del VPS.
- [Compatibilidad] `documentos_onlyoffice_menu.html` redirige a la pantalla unica y el menu de Administrar empresa apunta directo al nuevo flujo.
- [Seguridad] Sin tablas ni permisos nuevos; se conserva `empresa_id`, saneamiento de nombres y el wrapper de permisos de `documentos_onlyoffice`.

## [2026-05-13] Alertas de vencimiento de licencias
- [Super] `/super/configuracion_avanzada.html` agrega configuracion para activar alertas, definir dias de aviso, revisar pendientes y ejecutar envio manual.
- [Backend] Nuevo `/super/api/licencias/vencimiento_alertas`, worker periodico cada 12 horas y plantilla `licencia_expiry_warning`.
- [Datos] `licencia_vencimiento_notificaciones` registra envios/capturas y evita duplicados por licencia, empresa, correo, fecha y umbral.
- [Validacion] `go test ./...` en `backend/` OK.

## [2026-05-12] SSH VPS cambiado a 49222
- [Operacion VPS] SSH queda escuchando solo en `49222`; el puerto `22` fue cerrado despues de probar conexion externa real por el puerto nuevo.
- [Seguridad] `ssh.socket` queda desactivado y `ssh.service` gestiona el listener para evitar el bloqueo ocurrido anteriormente con socket activation.
- [Deploy] `scripts/pcs_deployment.local.ps1`, `scripts/pcs_deployment.local.ps1.example`, tuneles locales, RustDesk remoto y escaner VPS usan `49222`.

## [2026-05-13] Submenu de Configuracion super alineado
- [UX] `/super/configuracion_avanzada.html` adopta el patron visual del sidebar de `seleccionar_empresa.html` para su submenu interno.
- [Frontend] El submenu usa titulo simple, ancho estable, botones compactos, estados activo/hover coherentes y colapso movil.
- [Alcance] No cambia backend, endpoints, permisos, configuraciones guardadas, base de datos ni dependencias.

## [2026-05-13] Apariencia claro/oscuro del Explorador de Archivos super
- [Frontend] `/super/explorador_archivos.html` inicializa tema desde `localStorage`/cookie `pcs_theme`.
- [UX] Fondos, tarjetas, tabla, botones, input de ruta, estados y hover ahora usan variables globales compatibles con claro/oscuro.
- [Alcance] No cambia backend, endpoint, permisos, filesystem, base de datos ni dependencias.

## [2026-05-12] Tickets de ayuda empresariales profesionalizados
- [Empresa] El menu flotante crea tickets con contacto preferido, telefono opcional, contexto tecnico seguro de la pantalla activa y tickets recientes de la empresa.
- [Backend] `/api/empresa/tickets_ayuda` permite detalle y comentarios propios con validacion por `empresa_id`; `super_tickets_ayuda` guarda contacto y contexto tecnico.
- [Super] `/super/tickets_ayuda.html` muestra categoria, modulo, ruta, contacto y diagnostico para triage profesional.
- [Seguridad] Las notas internas no se devuelven a empresas y el contexto no incluye cookies, localStorage, tokens ni secretos.

## [2026-05-12] Retiro de mesa de ayuda legado
- [Backend] Se elimina la ruta empresarial heredada y su wrapper de permisos; el soporte oficial queda en `/api/empresa/tickets_ayuda`.
- [Frontend] Se retiran la pagina, el enlace de menu y la opcion de licencias del modulo legado.
- [Producto] Los tickets propios quedan centralizados en `super_tickets_ayuda` y `super_ticket_ayuda_mensajes`, con bandeja exclusiva del super administrador.

## [2026-05-12] Mesa central de tickets de ayuda
- [Backend] Se agregan `/api/empresa/tickets_ayuda` y `/super/api/tickets_ayuda` con almacenamiento PostgreSQL en `super_tickets_ayuda` y `super_ticket_ayuda_mensajes`.
- [Empresa] El menu flotante global suma `Crear ticket de ayuda`, detecta empresa activa, modulo y ruta para enviar la solicitud al soporte del SaaS.
- [Super] `web/super/tickets_ayuda.html` permite filtrar, responder, cambiar estado/prioridad, asignar responsable y cerrar tickets.
- [Seguridad] La creacion valida alcance por `empresa_id`; la bandeja central queda exclusiva de `super_administrador`.

## [2026-05-12] Carrusel index con tarjetas iguales
- [Frontend] `web/index.html` alinea el ancho de las tarjetas del carrusel con la grilla superior.
- [Responsive] El carrusel usa 3 tarjetas por vista en escritorio, 2 en tablet y 1 en movil, igualando la lectura visual de la grilla.
- [UX] Las flechas avanzan una tarjeta usando el gap real calculado por CSS.

## [2026-05-12] Paquete vpssecurity/logstore en Docker
- [Docker] El paquete Go de almacenamiento del escaner VPS se mueve a `backend/vpssecurity/logstore`.
- [Correccion] El backend Docker deja de confundir codigo Go con carpetas runtime llamadas `logs` ignoradas por `.dockerignore`.

## [2026-05-12] Limpieza de codigo backend remoto
- [Deploy] `scripts/sync_to_vps.ps1` borra codigo fuente backend obsoleto antes de extraer el paquete en el VPS.
- [Proteccion] Conserva `.env`, `.env.local`, `logs`, `bin`, `tmp` y `secure`.
- [Correccion] Evita builds Docker rotos por archivos eliminados localmente que seguian presentes en el VPS.

## [2026-05-12] SSH VPS restablecido temporalmente a 22
- [Operacion VPS] Hostinger quedo restablecido temporalmente para SSH en el puerto `22` antes del cambio seguro posterior a `49222`.

## [2026-05-12] Index con carrusel de sistemas
- [Frontend] `web/index.html` renderiza las primeras 6 tarjetas en la grilla principal y las demas en una hilera horizontal navegable.
- [UX] La barra de tarjetas restantes queda antes de `Modulos del sistema` y se controla con flechas izquierda/derecha.
- [Alcance] Sin cambios de backend, permisos, tablas ni dependencias.

## [2026-05-12] Fotos del index en landing descriptiva
- [Frontend] `web/descripcion_de_los_sistemas.html` ahora usa `imagen_secundaria_url` como imagen principal de cada tarjeta descriptiva, igual que las tarjetas del index.
- [UX] La landing conserva el logo pequeno de cada sistema como apoyo, pero la foto/ilustracion visible queda sincronizada con el portal principal.
- [Alcance] Sin cambios de backend, permisos, tablas ni dependencias.

## [2026-05-12] VPS portable 100% Docker
- [Deploy] `deploy/docker-compose.platform.yml` agrega perfil `edge` con Nginx publico para `80/443` y perfil `certbot` para certificados Let's Encrypt.
- [Operacion] Nuevos scripts `deploy/scripts/vps-docker-edge-up.sh` y `deploy/scripts/vps-docker-edge-renew.sh` permiten mover TLS/Nginx publico al stack Docker y renovar certificados sin depender de Nginx del host.
- [Portabilidad] Se agregan volumenes `pcs_letsencrypt` y `pcs_certbot_www`; migrar un VPS futuro requiere proyecto/imagenes, `deploy/.env.platform` y volumenes Docker.
- [Limpieza] Las plantillas `.env` y scripts de staging dejan de generar variables Nextcloud legacy.

## [2026-05-12] Explorador de Archivos super
- [Super] `web/super/explorador_archivos.html` agrega una vista tipo explorador de Windows para navegar carpetas del filesystem visible para el backend en el VPS.
- [Backend] Nuevo `GET /super/api/explorador_archivos?action=list&path=...`, protegido con `paginaPrincipalRequireSuperAdmin`, devuelve raices, ruta actual, ruta padre y metadata de archivos/carpetas sin leer contenido.
- [Seguridad] La operacion queda en modo solo lectura y no expone acciones de borrar, editar, subir, descargar ni abrir archivos; sin tablas ni dependencias nuevas.

## [2026-05-12] Retiro de Nextcloud y cuota DB por empresa
- [Backend] Se retiran rutas y handlers Nextcloud; el arranque elimina la tabla legacy `empresa_nextcloud_accounts` y las claves super `nextcloud.*`.
- [Super] `Limitaciones por empresa` cambia la cuota de almacenamiento por `db_max_gb`, usada como tamano maximo de base de datos por empresa y visible en el panel PostgreSQL por empresa.
- [Deploy] Nextcloud sale del Compose oficial; `deploy/scripts/vps-remove-nextcloud.sh` apaga contenedores legacy y puede purgar volumenes con confirmacion operativa.

## [2026-05-12] Documentos y backups locales por dispositivo
- [OnlyOffice] `web/administrar_empresa/documentos_onlyoffice.html` crea documentos en modo local por defecto; `backend/handlers/onlyoffice.go` devuelve el archivo como descarga con `action=create_local` sin guardarlo en el VPS.
- [Backups] `web/administrar_empresa/backups.html` agrega backup automatico local por navegador y descarga directa al equipo/celular.
- [Backend] `/api/empresa/backups` suma `exportar_local` y `exportar_configuracion_local`, construyendo JSON en memoria sin crear historial ni archivos en disco.

## [2026-05-12] Logos configurables en documentos empresariales
- [Backend] `empresa_configuracion_avanzada` agrega `mostrar_logo_empresa` y `mostrar_logo_sistema`, manteniendo `mostrar_logo` como compatibilidad general.
- [Frontend] Configuracion de empresa muestra checks separados para activar el logo empresarial o el logo del sistema en facturas, recibos, reportes y documentos imprimibles.
- [Impresion] `web/js/print_documents.js` soporta varios logos por documento; facturas, ventas y carrito pasan la configuracion avanzada al motor comun.

## [2026-05-12] Regreso a estaciones en ultimos movimientos de caja
- [Frontend] `web/administrar_empresa/ultimos_movimientos_de_caja.html` agrega `Regresar a estaciones` en el encabezado.
- [UX] El enlace conserva `empresa_id` para volver a `estaciones.html` dentro de la misma empresa.
- [Alcance] Sin backend, permisos, tablas ni dependencias nuevas.

## [2026-05-12] Login usuario simplificado y reenvio de invitacion
- [Frontend] `web/login_usuario.html` retira controles y enlaces no necesarios del login operativo: apariencia manual, inicio, acceso admin, completar/cambiar/volver redundantes.
- [UX] El formulario principal conserva solo `¿Olvidó su contraseña?` y `Recuperar email de invitación`.
- [Backend] Nuevo `POST /api/empresa/usuarios/recuperar_invitacion`, con reCAPTCHA y respuesta enmascarada, reenvia invitacion si el usuario ya fue creado por el administrador y no completo password.
- [Seguridad] El reenvio rota el token de invitacion por 48 horas y no revela existencia de correos.

## [2026-05-12] Registro operativo solo por invitacion
- [Backend] `EmpresaUsuarioSetPasswordHandler` ahora exige `token_invitacion` para el primer registro, valida expiracion y consume el token al guardar la contrasena.
- [Correo] Las invitaciones apuntan a `login_usuario.html?token_invitacion=...&modo=registro`; la ruta legacy `/auth/confirmar_correo` redirige al mismo flujo sin confirmar por separado.
- [Sesion] El primer ingreso confirma correo, abre sesion, devuelve `empresa_id`, `rol` y `redirect_url` hacia `administrar_empresa.html?id=...`.
- [Permisos] `administrar_empresa` conserva `/api/empresa/permisos_contexto` como fuente efectiva para ocultar/ver acciones por rol, licencia y reglas finas.
- [Alcance] Sin dependencias nuevas ni columnas nuevas; se reutilizan `email_confirm_token`, `email_confirm_expira` y `email_confirmado`.

## [2026-05-12] Tema del login usuario desde cookies
- [Frontend] `web/login_usuario.html` lee primero la cookie `pcs_theme` antes de caer a `localStorage`, aplicando el tema antes de cargar estilos.
- [Tema global] `web/menu.js` tambien prioriza `pcs_theme` para que el gestor compartido no sobrescriba el login con un valor local antiguo.
- [UX] El login de usuarios operativos conserva la apariencia clara/oscura del ultimo usuario que inicio sesion o cambio tema en el navegador.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Ayuda interna en CRM unificado
- [Menu] `web/administrar_empresa/modulo_menu.html` agrega `Ayuda` como ultimo boton del menu de `crm_unificado`.
- [Frontend] `web/administrar_empresa/crm_comercial.html` agrega una pestana `Ayuda` que primero define que es CRM y despues explica el uso del modulo.
- [UX] La ayuda cubre tablero ejecutivo, leads, seguimientos, cotizaciones, forecast, metas, embudo documental y la regla de no duplicar clientes/ventas/facturacion.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Totales del carrito por tipo de empresa
- [Frontend] `web/administrar_empresa/carrito_de_compras.html` adapta la tarjeta `Totales y detalles` con perfiles de negocio: estadia, gimnasio, clinico, pedido, transporte, parqueadero, alquiler, copropiedad, farmacia, belleza, orden de servicio, academico, obra y plantillas nuevas.
- [Contexto] La pagina consulta `/api/empresa/permisos_contexto` para leer `vertical_scope` y usa configuracion general como respaldo; si el contexto no llega, conserva el perfil POS universal.
- [UX] `web/estilos.css` agrega un indicador compacto de perfil sin cambiar el ancho de la tarjeta ni el flujo de pago.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Gimnasio empresarial
- [Frontend] `web/administrar_empresa/gimnasio.html` suma un mando ejecutivo con salud operativa, alertas, acciones rapidas y logo del vertical.
- [UX] `web/js/gimnasio.js` calcula riesgos desde socios, planes, clases, pagos, dispositivos y bitacora de acceso; agrega busqueda/filtro global y badges de estado en tablas.
- [Menu] `web/administrar_empresa/gimnasio_menu.html` organiza el modulo por Direccion, Comercial, Operacion deportiva y Acceso, usando el patron agrupado del panel empresarial.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Menu de configuracion super alineado
- [Frontend] `web/super/configuracion_avanzada.html` usa el mismo patron visual de menu lateral agrupado que Administrar empresa: `admin-sidebar`, `admin-nav-grouped`, grupos plegables, iconos y boton de colapso movil.
- [UX] El buscador de secciones se conserva y ahora abre/oculta grupos completos segun coincidencias, manteniendo scroll interno a cada tarjeta.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Scroll superior en Tipos de empresa
- [Frontend] `web/super/tipos_empresas.html` agrega un scroll horizontal arriba de la tabla y lo sincroniza con el scroll inferior.
- [Alcance] Sin cambios de API, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Plantillas 20 actualizado
- [Frontend] `web/super/plantillas_produccion_masiva.html` muestra semaforo ejecutivo, brechas principales, tarjetas de foco comercial y KPIs de licencias/base/readiness.
- [Operacion] Se agregan filtros `Sin licencia` y `Sin preconfig` para priorizar tareas antes de publicar comercialmente un vertical.
- [Contrato] El cruce de readiness acepta payloads con `items`, `licencias` o arreglos directos y valida catalogo, metadata, preconfiguracion activa y licencia base.
- [Alcance] Sin endpoints, permisos, tablas ni dependencias nuevas.

## [2026-05-12] Limpieza visual del super administrador
- [Frontend] `web/super_administrador.html` retira los indicadores compactos `PostgreSQL`, `VPS`, `Licencias` y `Seguridad` del encabezado lateral.
- [Alcance] Sin cambios de rutas, iframes, permisos, backend, base de datos ni dependencias.

## [2026-05-12] Identidad visual empresarial
- [Frontend] `web/administrar_empresa/configuracion.html` agrega una seccion visible para cargar el logo de la empresa y sincronizarlo con la configuracion de factura/documentos.
- [Panel] `web/administrar_empresa/panel.html` muestra el logo encima de `Panel de <empresa>` con tamano fijo, sin modificar la columna ni dimensiones de la tarjeta de clima.
- [Backend] `/api/empresa/configuracion_avanzada/logo` acepta PNG/JPG/WEBP/GIF hasta 5 MB, guarda en `/uploads/empresa_logos/empresa_<id>/` y persiste `logo_url` + `mostrar_logo`.
- [Datos] No hay tablas nuevas; se reutiliza `empresa_configuracion_avanzada` como fuente unica para panel, factura y documentos.

## [2026-05-12] CRM empresarial profesional
- [Backend] `GET /api/empresa/crm_avanzado` con `action=dashboard` suma salud comercial, valor en riesgo, leads sin contacto, oportunidades estancadas, acciones priorizadas, responsables y canales.
- [Frontend] `web/administrar_empresa/crm_comercial.html` y `web/js/crm_comercial.js` muestran cockpit ejecutivo, plan de accion, responsables y canales dentro de CRM unificado.
- [Permisos] Los CRUD `/api/empresa/crm/*` quedan bajo `WithEmpresaCRMUnificadoPermissions`.
- [Alcance] Sin tablas, rutas publicas ni dependencias nuevas.

## [2026-05-12] Nucleo configurable por plantilla
- [Backend] `tipo_empresa_preconfiguraciones.config_json` normaliza `adaptacion_nucleo` para usuarios operativos, productos/servicios y estaciones como recursos del negocio.
- [Aplicacion] Al aplicar una plantilla se guarda `preconfiguracion_tipo_empresa_adaptacion_nucleo` y `estaciones_config` incluye `tipo_recurso`, `tipo_recurso_plural` y `representa_recurso_negocio`.
- [Super] `web/super/preconfiguracion_tipos_empresa.html` muestra la seccion `Nucleo configurable` y envia la metadata al guardar.
- [Alcance] Sin tablas, columnas, permisos ni dependencias nuevas.

## [2026-05-12] Matriz profesional de 30 plantillas
- [Backend] `/api/*/plantillas_integracion/catalogo` publica exactamente 30 plantillas canonicos: 10 clasicos reales y 20 nuevos. `consultorio_odontologico` se fusiona en `odontologia`, `taxi` en `taxi_system` y `turnos_atencion`/`turnos` quedan como soporte transversal.
- [Contrato] Cada vertical calcula `professional_ready`, `readiness_score`, checks, alcance de configuracion, ingresos, egresos, tablas financieras y metadata de `fused_modules`, `support_modules` o `similar_templates` cuando aplica.
- [Frontend] `web/administrar_empresa/plantillas_integracion.html` muestra perfil activo por empresa, contrato, ventas/finanzas, reportes y configuracion; cada vertical queda marcado como `Profesional` o `Brecha`.
- [Finanzas] Los 30 plantillas quedan atados a `empresa_finanzas_movimientos` y a los modulos centrales de ventas, pagos, bancos/tesoreria y reportes para ingresos y egresos.
- [Configuracion] El acceso conserva `linkPlantillasIntegracion` y permiso `seguridad:R`, pero se presenta como `Adaptacion por tipo`.
- [Alcance] No hay tablas, endpoints de escritura, permisos ni dependencias nuevas.

## [2026-05-12] Licencias ocultables para clientes
- [Super] `web/super/licencias.html` expone la bandera como `Visibilidad comercial`: visible u oculta para clientes.
- [Backend] Los checkouts públicos de licencia rechazan licencias con `activo=0`, incluyendo resumen público, Wompi, Nequi, Epayco, activación sin pago y addons seleccionados manualmente.
- [Datos] Reutiliza `licencias.activo`; no agrega tablas, columnas ni dependencias.

## [2026-05-12] Indicadores economicos compactos en panel empresarial
- [Frontend] `web/administrar_empresa/panel.html` presenta los indicadores de mercado en una tabla compacta de dos indicadores por fila en escritorio.
- [Responsive] En movil conserva las tarjetas reducidas existentes para evitar desbordes horizontales.
- [Alcance] Sin cambios de API, permisos, base de datos ni dependencias.

## [2026-05-19] Administrar empresa movil
- [Frontend/PWA] El panel `Administrar empresa` actualiza service worker a `pcs-shell-v3`, limpia caches antiguas y usa network-first para CSS/JS/manifest para que celulares y PWA instaladas no muestren estilos viejos.
- [Responsive] Menu, submenus, botones e iframe del shell empresarial se ajustan al ancho movil sin desborde horizontal ni doble scroll innecesario; el panel inicial deja de cortar titulo, ciudad, clima y pie en pantallas pequenas.
- [QA] Validacion visual en viewport movil 390x844 y validacion de sintaxis de scripts/service worker.

## [2026-05-19] Clientes desde carrito
- [Carritos] El boton `Clientes` abre un panel interno para crear/asignar cliente al carrito activo sin salir de venta directa o estaciones.
- [Configuracion] Nuevo check `Exigir cliente registrado para pagar` dentro de la configuracion del carrito.
- [Backend] `pagar_estacion` bloquea el cierre cuando `cliente_obligatorio_pago` esta activo y el carrito no tiene cliente.

## [2026-05-12] Enlace Probar Gratis del index
- [Frontend] `web/index.html` cambia el destino de ficha comercial a `/descripcion_de_los_sistemas.html` conservando `accion=probar_gratis`, `tipo_empresa`, modulo, secciones y ancla de la tarjeta elegida.
- [Backend] `backend/main.go` atiende la ruta legacy `/descripcion_de_los_sistemas.ht` sirviendo la version `.html` con `Content-Type: text/html`, evitando que navegadores la descarguen.
- [Seguridad] `AuthMiddleware` mantiene publicas ambas rutas descriptivas; no se abren rutas privadas ni permisos nuevos.
- [Operacion] `web/super/pagina_principal.html` y la auditoria profesional usan la ruta oficial `.html`.

## [2026-05-12] Apariencia claro/oscuro del Centro de mando
- [Frontend] `web/super/licencias_resumen.html` usa las variables globales `--bg`, `--surface`, `--surface-soft`, `--text`, `--muted`, `--border`, `--accent` y `--accent-2` para integrarse con temas claros y oscuros.
- [UX] Se corrigen fondos, bordes, botones, pills de estado, tablas, graficas SVG y aro de score para evitar contraste roto al cambiar de tema.
- [Alcance] Sin cambios de API, permisos, base de datos ni dependencias.

## [2026-05-12] Centro de mando super reconstruido
- [Frontend] `web/super/licencias_resumen.html` se reemplaza completo por una consola ejecutiva responsive con score operativo, KPIs de plataforma, PostgreSQL, seguridad, negocio SaaS, costos, SLO, riesgos, servicios e incidentes.
- [Operacion] La vista agrega controles directos para actualizar, evaluar alertas y abrir gobierno de alertas, PostgreSQL, seguridad VPS, licencias, empresas, tipos de empresa, roles, plantillas, IA, configuracion y reportes.
- [Alcance] Reutiliza APIs existentes del panel super; no agrega endpoints, permisos, tablas, dependencias ni cambios en `go.mod`.
- [QA] Parseo de script inline con Node OK; `git diff --check -- web/super/licencias_resumen.html` sin errores.

## [2026-05-11] Matriz de integracion en configuracion empresarial
- [Menu] `Matriz de integracion` sale de Soluciones por negocio y queda en Administrar empresa > Configuracion > Base empresarial.
- [Permisos] `linkPlantillasIntegracion` conserva `seguridad:R`, ahora agrupado como Administracion y configuracion.

## [2026-05-11] Emisora online por empresa
- [Backend] `/api/chat_flotante/preferencias` acepta `empresa_id` y persiste `chat_flotante.*`, incluida `radio_online_enabled`, en `empresa_estacion_prefs`.
- [Frontend] `Configurar chat y robot` agrega el check `Activar emisora online`; el panel compacto del chat y `radio_player.js` sincronizan el reproductor flotante con esa preferencia.
- [QA] `node --check web/js/ai_chat_drawer.js`; `node --check web/js/radio_player.js`; `go test ./...` en `backend/`.

## [2026-05-11] Alcance vertical por licencia
- [Backend] `/api/empresa/permisos_contexto` calcula `vertical_scope` desde tipo/preconfiguracion/licencia y desactiva acciones de plantillas ajenos sin tocar el nucleo universal.
- [Licencias] El checkout, activacion manual/gratuita y confirmaciones de pago validan que la licencia base corresponda al tipo de empresa elegido.
- [Frontend] `elegir_licencia.html` consulta licencias filtradas por `tipo_id` y `editar_empresa.js` conserva `tipo_id/tipo_nombre` al renovar.
- [QA] `go test ./handlers`; `go test ./db`.

## [2026-05-11] 2FA del login desde configuracion avanzada
- [Seguridad] El login de administradores oculta el campo de codigo 2FA salvo que `security.admin_2fa.enabled` este activo.
- [Backend] `/config.js` publica `ADMIN_2FA_LOGIN_ENABLED` y `AdminLoginHandler` solo exige OTP cuando el switch global y el TOTP de la cuenta estan activos.
- [Frontend] `web/super/configuracion_avanzada.html` agrega la tarjeta `2FA login` para activar/desactivar la exigencia global sin tocar secretos por cuenta.
- [QA] `go test ./handlers -run "TestAdminTOTPLoginRequiredForAdmin" -count=1`; `go test ./... -count=1`; validacion JS de `login.js` y scripts inline.

## [2026-05-11] Catalogos publicos de plantillas sin sesion
- [Seguridad] `backend/utils/utils.go` agrega a la lista publica `/api/public/plantillas_nuevas/catalogo` y `/api/public/plantillas_integracion/catalogo`.
- [Producto] La portada publica y las fichas comerciales pueden consultar el catalogo real de plantillas sin depender de una sesion administrativa.
- [QA] `backend/utils/auth_middleware_test.go` valida que ambas rutas pasen sin cookie y que las rutas privadas sigan protegidas.

## [2026-05-11] Sincronizacion idempotente de pagos plantillas
- [Backend] `backend/db/odontologia.go` y `backend/db/gimnasio.go` reutilizan `carritos_compras.referencia_externa` antes de crear carritos desde pagos historicos.
- [QA] Se agregan pruebas para fijar la llave historica de pagos en odontologia y gimnasio.
- [Alcance] No hay tablas, endpoints, permisos ni dependencias nuevas.

## [2026-05-11] Correccion de cargas parciales en plantillas integrados
- [Backend] `backend/db/odontologia.go` y `backend/db/gimnasio.go` crean indices de integracion solo despues de asegurar columnas nuevas en bases PostgreSQL existentes.
- [Frontend] `web/js/consultorio_odontologico.js`, `web/js/gimnasio.js` y `web/js/alquileres.js` limpian el aviso de carga parcial cuando la recarga completa no devuelve errores.
- [QA] Se agregan pruebas para impedir que los indices de `cliente_id`, `servicio_id` y `carrito_id` vuelvan a ejecutarse antes de las columnas.
- [Alcance] No hay tablas, endpoints, permisos ni dependencias nuevas.

## [2026-05-11] Fix arranque PostgreSQL parqueadero
- [Backend] `backend/db/parqueadero.go` ahora asegura columnas de integracion al nucleo antes de crear el indice por `carrito_id`.
- [Operacion] Corrige el fallo de despliegue en VPS con bases existentes que todavia no tenian `empresa_parqueadero_tickets.carrito_id`.
- [Alcance] No hay nuevas tablas, endpoints, permisos ni dependencias.

## [2026-05-11] Consistencia del panel super
- [Frontend] `web/js/super_administrador.js` permite restaurar la ayuda privada del super administrador mediante validacion explicita de ruta.
- [Seguridad] La ayuda privada no queda dentro de la lista limitada del rol `control_super_administrador`.
- [QA] Alinea el contrato esperado por `TestSuperAdminPanelExposesPrivateHelpButton`.

## [2026-05-11] 20 plantillas nuevas reales
- [Backend] `backend/db/plantillas_nuevas_bootstrap.go` promueve los 20 plantillas nuevas a produccion masiva con ranking 1-20.
- [API] `/super/api/plantillas_nuevas/catalogo` acepta `asegurar_20_licencias` y conserva `asegurar_v1_licencias` como alias compatible.
- [Frontend] `web/js/plantillas_nuevas_catalogo.js`, `web/index.html` y `web/super/plantillas_produccion_masiva.html` publican y gobiernan las 20 plantillas reales.
- [QA] Las pruebas actualizadas exigen 20 plantillas masivos, metadata extendida y decision de produccion masiva en nuevas plantillas.
- [Alcance] No hay tablas, dependencias ni circuitos paralelos de clientes, productos, ventas o pagos.

## [2026-05-11] Portada index alineada a modulos reales
- [Frontend] `web/index.html` y los defaults de `/api/public/pagina_principal` actualizan el texto de cobertura y las tarjetas publicas con nucleo unico, modulos reales y plantillas clasificados.
- [Producto] Los 20 plantillas nuevas siguen en catalogo y quedan publicables como tarjetas operativas de `Probar gratis`.
- [Catalogo] `web/js/plantillas_nuevas_catalogo.js` agrega decision, ranking, metadata de plantilla, permisos, flujo de venta y reportes para sincronizar la portada con la matriz extendida.
- [Alcance] No hay endpoints, tablas, permisos, dependencias ni cambios en `go.mod`.

## [2026-05-11] Aseguramiento comercial de plantillas
- [Backend] `POST /super/api/plantillas_nuevas/catalogoaction=asegurar_20_licencias` llama `EnsureNuevasPlantillasProduccionMasivaLicencias`; `asegurar_v1_licencias` queda como alias compatible.
- [Producto] La accion asegura tipos de empresa, preconfiguraciones y cuatro planes recomendados para los 20 plantillas.
- [Frontend] `web/super/plantillas_produccion_masiva.html` agrega `Asegurar 20` y refresca el semaforo despues de ejecutar.
- [Alcance] No hay tablas, rutas nuevas, permisos nuevos ni dependencias.

## [2026-05-11] Semaforo listo para venta en plantillas
- [Frontend] `web/super/plantillas_produccion_masiva.html` cruza plantillas, preconfiguraciones y licencias activas para marcar `Listo venta`.
- [Regla] Un vertical queda listo solo si tiene metadata completa, preconfiguracion activa con `integracion_vertical` y licencia activa que incluye el modulo.
- [Alcance] No hay cambios de esquema, endpoints, permisos ni dependencias.

## [2026-05-11] Acciones de gobierno para plantillas 20
- [Frontend] Cada fila de `web/super/plantillas_produccion_masiva.html` enlaza a tipos, preconfiguraciones y licencias del vertical.
- [UX] `web/super/tipos_empresas.html`, `web/super/preconfiguracion_tipos_empresa.html` y `web/super/licencias.html` aplican filtros iniciales desde `q`, `vertical` o `modulo`.
- [Alcance] No se agregan endpoints, tablas, permisos ni dependencias.

## [2026-05-11] Gobierno super de plantillas de produccion masiva
- [Frontend] Se agrega `web/super/plantillas_produccion_masiva.html` con KPIs, filtros, ranking, decision, metadata extendida y exportacion CSV.
- [Menu] `web/super_administrador.html` incorpora `Plantillas 20` dentro de Licencias y `web/js/super_administrador.js` permite restaurar la pagina.
- [Seguridad] Se reutiliza `/super/api/plantillas_nuevas/catalogo`; no hay endpoints, permisos, esquemas ni dependencias nuevas.

## [2026-05-11] Preconfiguraciones y plantillas de produccion masiva
- [Backend] `config_json` de tipos de empresa puede incluir `integracion_vertical` con decision, prioridad, permisos, flujo de venta, tablas y reportes.
- [Catalogos] Los endpoints de plantillas nuevas publican `integracion_preconfig`, `produccion_masiva`, `prioridad_produccion` y `decision_preconfig`.
- [Producto] Se priorizan los 20 plantillas nuevas para produccion masiva en `documentos/plan_plantillas_produccion_masiva_2026-05-11.md`.
- [QA] Las pruebas exigen metadata extendida y exactamente 20 plantillas marcados como produccion masiva; no hay cambios de esquema ni dependencias.

## [2026-05-11] Matriz extendida de plantillas plantillas
- [Backend] El catalogo de integracion agrega `template_activates`, `tables_touched`, `required_permissions`, `sale_flow` y `reports_produced`.
- [Frontend] La matriz empresarial muestra modulos, plantilla, tablas, permisos, flujo de venta y reportes por vertical.
- [QA] La prueba de contrato impide publicar plantillas visibles sin metadata completa; no hay cambios de esquema ni dependencias.

## [2026-05-11] Sincronizacion segura de matriz vertical
- [Frontend] La matriz consulta `/api/empresa/permisos_contexto`, calcula sincronizaciones permitidas, deshabilita botones sin permiso efectivo y confirma antes de ejecutar POST.
- [Seguridad] El endpoint vertical conserva la autorizacion final por rol, licencia y `empresa_id`; no hay nuevas dependencias ni cambios de esquema.

## [2026-05-11] Sincronizacion desde matriz vertical
- [Frontend] `web/administrar_empresa/plantillas_integracion.html` agrega botones `Sincronizar` por vertical y muestra resultado/resumen de la accion.
- [Seguridad] La vista conserva permiso `seguridad:R`; cada POST mantiene la autorizacion real del endpoint vertical correspondiente.

## [2026-05-11] Pantalla de matriz vertical en empresa
- [Frontend] Se agrega `web/administrar_empresa/plantillas_integracion.html` para consultar KPIs, estado, nucleo, especialidad y sincronizacion por vertical.
- [Menu] `web/administrar_empresa/configuracion_menu.html` incorpora `Matriz de integración` dentro de Configuracion > Base empresarial.
- [Permisos] `linkPlantillasIntegracion` queda registrado con `seguridad:R` en backend y frontend.

## [2026-05-11] Indicador de matriz vertical en panel empresa
- [Frontend] `web/administrar_empresa.html` agrega un indicador compacto en el sidebar empresarial.
- [JS] `web/js/administrar_empresa.js` lo alimenta con el resumen de `web/js/plantillas_integracion_catalogo.js`.
- [UX] El panel muestra fuente API/local y conteo de plantillas visibles/ocultos sin cambiar permisos, licencias ni rutas.

## [2026-05-11] Frontend consume matriz API de plantillas
- [Frontend] `web/js/administrar_empresa.js` carga `/api/empresa/plantillas_integracion/catalogo` antes de aplicar permisos/licencias del menu empresarial.
- [Fallback] `web/js/plantillas_integracion_catalogo.js` conserva el catalogo local y ahora permite fusionar items recibidos desde backend.
- [Gobernanza] El menu deja de depender solo de un archivo JS estatico para decidir si una vertical clasica puede mostrarse como operativa.

## [2026-05-11] Catalogo API de integracion vertical
- [Backend] Se agrega `backend/handlers/empresa_plantillas_integracion.go` para exponer la matriz de plantillas clasicos.
- [API] Nuevas rutas de solo lectura: `/api/public/plantillas_integracion/catalogo`, `/api/empresa/plantillas_integracion/catalogo` y `/super/api/plantillas_integracion/catalogo`.
- [QA] `backend/handlers/empresa_plantillas_integracion_test.go` bloquea plantillas visibles con duplicados del nucleo.

## [2026-05-11] AIU construccion integrado al nucleo
- [Backend] `aiu_construccion` enlaza clientes de obra con clientes centrales, contratos/conceptos con servicios y facturas AIU con ventas centrales en carritos.
- [Frontend] El panel AIU incluye accion de sincronizacion y resumen de clientes, servicios y facturas conectadas.
- [Gobernanza] AIU queda visible como plantilla integrada; sus tablas propias se conservan para capitulos, calculo AIU, retenciones, anticipo, garantia, avance, riesgo y auditoria tecnica.

## [2026-05-11] Drogueria/farmacia validada al nucleo
- [Backend] `drogueria_farmacia` se mantiene sobre `empresa_modulos_colombia_*` como expediente sanitario, sin tablas paralelas de productos, inventario, ventas ni pagos.
- [Frontend] La pagina de drogueria/farmacia declara que opera sobre productos, inventario, ventas y facturacion centrales.
- [Catalogo] La vertical queda visible como `plantilla_integrada_nucleo` y sin duplicados del nucleo.
- [Gobernanza] Lotes, INVIMA, formulas, controlados, dispensacion, devoluciones y farmacovigilancia quedan como especialidad sanitaria.

## [2026-05-11] Alquileres integrado al nucleo
- [Backend] `alquileres` enlaza clientes de contratos a clientes centrales, activos/tarifas a servicios y contratos con valor a ventas centrales en carritos.
- [Frontend] El panel de alquileres incluye accion de sincronizacion y resumen de clientes, servicios y contratos conectados.
- [Gobernanza] Alquileres queda visible como plantilla integrada; sus tablas propias se conservan para activos, garantias, kilometraje, GPS, mantenimiento, entrega y devolucion.

## [2026-05-11] Propiedad horizontal integrada al nucleo
- [Backend] `propiedad_horizontal` enlaza propietarios/residentes a clientes centrales, unidades/cargos a servicios y recaudos a ventas centrales en carritos.
- [Frontend] El panel de propiedad horizontal incluye accion de sincronizacion y resumen de clientes, servicios y recaudos conectados.
- [Gobernanza] Propiedad horizontal queda visible como plantilla integrada; sus tablas propias se conservan para unidades, coeficientes, cartera, PQR y asambleas.

## [2026-05-11] Apartamentos turisticos integrado al nucleo
- [Backend] `apartamentos_turisticos` enlaza huespedes a clientes centrales, unidades a servicios y reservas cerradas a ventas centrales en carritos.
- [Frontend] El panel de apartamentos incluye accion de sincronizacion y resumen de reservas, servicios, clientes y observaciones.
- [Gobernanza] Apartamentos turisticos queda visible como plantilla integrada; sus tablas propias se conservan para unidades, tarifas, disponibilidad, codigos de acceso, limpieza y mantenimiento.

## [2026-05-11] Domicilios integrado al nucleo
- [Backend] `domicilios` enlaza clientes de pedidos a clientes centrales, productos de menu a servicios y pedidos entregados a ventas centrales en carritos.
- [Frontend] El panel de domicilios incluye accion de sincronizacion y resumen de pedidos, servicios de menu, clientes y observaciones.
- [Gobernanza] Domicilios queda visible como plantilla integrada; sus tablas propias se conservan para restaurantes, domiciliarios, ofertas, GPS, tracking y estados logisticos.

## [2026-05-11] Fases de integracion profesional de plantillas
- [Gobernanza] Se agrega `documentos/matriz_integracion_plantillas.md` como contrato para mantener clientes, productos/servicios, ventas, pagos, facturacion, reportes y permisos en el nucleo.
- [Frontend] `web/js/plantillas_integracion_catalogo.js` clasifica plantillas clasicos y oculta del menu operativo los que siguen duplicando funciones centrales.
- [Catalogo] `web/js/plantillas_nuevas_catalogo.js` y los endpoints de plantillas nuevas publican estado de integracion, visibilidad operativa, modulos base y duplicados detectados.

## [2026-05-11] Gimnasio integrado al nucleo
- [Backend] `gimnasio` enlaza socios a clientes, planes a servicios y pagos a ventas centrales en carritos.
- [Frontend] El dashboard de gimnasio incluye accion de sincronizacion y resumen de clientes/servicios/ventas sincronizados.
- [Gobernanza] Gimnasio queda visible como plantilla integrada; sus tablas propias se conservan para acceso, clases y asistencia.

## [2026-05-11] Odontologia integrada al nucleo
- [Backend] `odontologia` enlaza pacientes a clientes, tratamientos a servicios y pagos a ventas centrales en carritos.
- [Frontend] El panel de consultorio incluye accion de sincronizacion y resumen de pacientes/tratamientos/pagos sincronizados.
- [Gobernanza] Odontologia queda visible como plantilla integrada; sus tablas propias se conservan para historia clinica, odontograma, agenda y presupuesto clinico.

## [2026-05-11] Parqueadero integrado al nucleo
- [Backend] `parqueadero` enlaza tickets cobrados a clientes opcionales, servicios y ventas centrales en carritos.
- [Frontend] El panel de parqueadero incluye accion de sincronizacion y resumen de tickets sincronizados.
- [Gobernanza] Parqueadero queda visible como plantilla integrada; su tabla propia se conserva para placas, QR, entrada/salida, tarifas y anulaciones.

## [2026-05-11] Taxi system integrado al nucleo
- [Backend] `taxi_system` enlaza clientes registrados/invitados a clientes centrales, servicios de viaje a servicios y viajes completados a ventas centrales en carritos.
- [Frontend] El panel de taxi incluye accion de sincronizacion y resumen de viajes, clientes y pendientes.
- [Gobernanza] Taxi system queda visible como plantilla integrada; sus tablas propias se conservan para conductores, GPS, despacho, ofertas y rutas.

## [2026-05-11] Panel super profesional
- [Shell] `web/super_administrador.html` queda con un menu ejecutivo de 16 accesos necesarios.
- [UX] Se agrega cabecera compacta PCS, alcance operativo y estilo visual mas denso para trabajo diario.
- [Navegacion] `web/js/super_administrador.js` solo restaura paginas visibles del panel principal.
- [Alcance] No se eliminan modulos ni endpoints; los accesos secundarios dejan de ocupar el panel inicial.

## [2026-05-11] Limpieza PostgreSQL-only
- [Backend] Las verificaciones residuales de indices en finanzas y propinas consultan `pg_indexes` y ya no conservan ramas de motor legado.
- [Frontend] Los helpers de fecha en carrito y codigos de descuento usan nombres neutrales de backend.
- [Operacion] Scripts de sincronizacion y actualizacion dejan de contemplar extensiones de motores retirados; se eliminaron artefactos locales generados en perfiles temporales.
- [Gobernanza] Documentacion e instrucciones quedan alineadas a PostgreSQL como unico motor permitido, sin dependencias nuevas ni cambios de esquema.

## [2026-05-11] Centro de mando profesional super
- [Super] `web/super/licencias_resumen.html` ahora abre con una lectura ejecutiva del VPS y del proyecto.
- [Metricas] El panel consolida CPU, memoria, disco, trafico, historico, PostgreSQL, alertas, errores, servicios, procesos, licencias, empresas y consumo OpenAI estimado.
- [Gobernanza] Se reutilizan endpoints existentes y no se agregan dependencias, tablas ni permisos nuevos.

## [2026-05-11] Cierre implementable de pendientes 1 a 8
- [Pagos] `web/pagar_licencia.html` permite elegir manualmente el pais de pago, guarda la preferencia local y recarga disponibilidad de Wompi/Epayco por `pais_codigo`.
- [Docs] Se corrigen referencias activas a documentos historicos inexistentes y se apunta a fuentes vigentes.
- [Alcance] DIAN oficial SOAP/WSDL, proveedores/hardware reales, E2E con credenciales y normalizacion masiva de mojibake quedan documentados como pendientes externos/controlados, no como cierres locales.

## [2026-05-11] Madurez empresarial de 12 pasos
- [Staging] `deploy/scripts/vps-refresh-staging-from-production.sh` anonimiza datos por defecto con `deploy/scripts/vps-anonymize-staging.sh`.
- [Monitoreo] Se agrega stack Prometheus/Grafana/node-exporter/cAdvisor en `deploy/monitoring/` y script `deploy/scripts/vps-monitoring-up.sh`.
- [Backups] Se agrega backup externo opcional por `rclone` o `s3` con cron instalable.
- [QA] Se agregan auditorias `qa_roles_matrix`, `payment_matrix_audit`, `support_center_audit`, `docs_normalization_audit` y `load_smoke_test`.
- [Release] `scripts/release_gate.ps1` genera manifiesto con `tools/release_manifest.mjs` e integra prueba de carga smoke.
- [CI] `professional-ci.yml` incorpora auditorias nuevas y despliegue opcional a staging con `PCS_ENABLE_STAGING_DEPLOY=true`.
- [Verificacion] `.\scripts\profesional_preflight.ps1 -Full` OK y `node tools\load_smoke_test.mjs` contra staging OK.

## [2026-05-11] Operacion profesional diaria de los 12 frentes
- [VPS] Se agregan scripts para instalar Docker Buildx, activar staging Nginx, tomar snapshots de observabilidad y programar backups por cron.
- [Staging] Se agrega refresco controlado de staging desde produccion con `deploy/scripts/vps-refresh-staging-from-production.sh`.
- [Release] Se agrega `scripts/release_gate.ps1` y runbook de release profesional.
- [CI/E2E] El workflow E2E queda programado contra staging y se documentan secretos requeridos de GitHub.
- [Auditorias] Se agregan auditorias de migraciones, QA funcional de modulos criticos y consistencia UX global al preflight.

## [2026-05-11] Implementacion profesional de 12 frentes operativos
- [Staging] Se agrega `deploy/docker-compose.staging.yml`, `deploy/.env.staging.example`, `scripts/staging_up.ps1` y `deploy/scripts/vps-staging-up.sh`.
- [CI/CD] Se agregan workflows `.github/workflows/professional-ci.yml` y `.github/workflows/e2e-visual.yml`.
- [Auditorias] Se agregan `tools/security_audit.mjs`, `tools/permissions_license_audit.mjs`, `tools/observability_report.mjs` y `tools/openapi_inventory.mjs`.
- [Backups] Se agrega `scripts/vps_restore_validation.ps1` para validar snapshots y ejecutar restauracion temporal PostgreSQL.
- [API/DRP] Se genera `documentos/api/openapi.generated.yaml` y runbooks de staging, CI, E2E y recuperacion ante desastre.
- [Verificacion] Se creo snapshot real en VPS `backups/vps-snapshots/20260511_055520`, se validaron tarballs y se ejecuto restauracion temporal PostgreSQL exitosa.

## [2026-05-11] Base profesional de QA, respaldo y auditoria
- [QA] Se agrega `scripts/profesional_preflight.ps1` para validar sintaxis, Docker, auditoria de modulos/permisos y `git diff --check` antes de despliegues.
- [Auditoria] Se agrega `tools/professional_audit.mjs` para revisar catalogo de 20 plantillas, permisos backend, wrappers, portal publico y documentacion obligatoria.
- [RS] `rs.ps1` ejecuta preflight por defecto antes de actualizar repositorio y sincronizar VPS; se puede omitir con `-SkipPreflight`.
- [Backups] Se agrega `scripts/vps_backup_operacion.ps1` para crear dump PostgreSQL y tarballs de volumenes persistentes en la VPS con retencion.
- [Docs] Se agrega `documentos/plan_profesionalizacion_plataforma.md` como guia de los siete frentes profesionales.

## [2026-05-11] Limpieza segura en sync_to_vps
- [Deploy] `scripts/sync_to_vps.ps1` agrega `-CleanupRemoteUnusedFiles` para limpiar temporales antiguos de sync, caches locales no persistentes, contenedores detenidos, imagenes Docker dangling y cache BuildKit no usado.
- [Seguridad] La limpieza evita `docker volume prune` y no toca PostgreSQL, uploads, descargas ni backups persistentes.
- [RS] `rs.ps1` y `scripts/rs.ps1` heredan la limpieza por defecto y permiten ajustar antiguedad de temporales/cache.

## [2026-05-11] Script rapido rs
- [Deploy] Se agrega `scripts/rs.ps1` para ejecutar en secuencia `actualizar_repositorio.ps1` y `sync_to_vps.ps1`.
- [Raiz] Se agrega `rs.ps1` como acceso corto desde la raiz del proyecto.
- [Seguridad operativa] Si falla la actualizacion del repositorio, no se ejecuta la sincronizacion al VPS.

## [2026-05-11] Reubicacion de ayuda super administrador
- [Super] El boton `Ayuda super administrador` se mueve al grupo `Infraestructura y comunicaciones`, justo al lado de `Configuracion avanzada`.
- [Seguridad] Se conserva la misma ruta privada `/ayuda/ayuda.html` y el mismo filtro de rol existente.

## [2026-05-11] Portada publica con plantillas completos
- [Index] Las 20 nuevas empresas del catalogo publico usan descripciones largas, similares a las tarjetas principales de la portada.
- [Probar gratis] El enlace de cada tarjeta conserva contexto de titulo, descripcion, modulo/tipo de empresa y secciones para llegar a una ficha de detalle mas completa.
- [Detalle publico] `descripcion_de_los_sistemas.ht` reutiliza el catalogo ampliado para mostrar informacion especifica de cada vertical antes del registro de prueba.

## [2026-05-10] Preconfiguraciones inteligentes y robot no automatico
- [Preconfiguracion] La siembra por tipo de empresa completa faltantes reales por `tipo_empresa_id`, aunque existan plantillas antiguas o sobrantes.
- [Plantillas] Los 20 tipos nuevos usan su plantilla inteligente como default si aun no tienen preconfiguracion guardada.
- [UX] Se retira el acceso visible a `Configuracion guiada` del submenu de configuracion empresarial.
- [Robot] Al crear una empresa ya no se pregunta ni se agenda el inicio automatico del robot de configuracion.

## [2026-05-10] Carritos en modo tactil
- [Carritos] Se agrega `modo_pantalla_tactil` a la configuracion unificada de carrito por empresa/estacion.
- [UI] `carrito_de_compras.html` adapta botones, campos, toolbar, lector, cobro y resumen para uso profesional en pantallas tactiles.
- [Productos] `buscar_producto_botones.html` agranda grilla, tarjetas y buscador cuando el carrito abre el catalogo en modo tactil.
- [Persistencia] La opcion se guarda dentro de `estaciones_config`, sin tablas nuevas ni dependencias externas.

## [2026-05-10] Alertas automaticas del sistema super
- [Super] Se agrega `web/super/alertas_sistema.html` como modulo privado con configuracion, estado actual, historial, prueba de correo y evaluacion manual.
- [Backend] Se agrega `/super/api/alertas_sistema` y el worker `super.alertas_worker` para evaluar cada minuto disco VPS, trafico, sesiones administrativas y conexiones PostgreSQL.
- [Base de datos] Se crean `super_alertas_config` y `super_alertas_eventos`; `metrics` ahora persiste `disk_total`, `disk_used` y `disk_percent`.
- [Correo] Las alertas reutilizan Gmail SMTP global y por defecto notifican a `powerfulcontrolsystem@gmail.com`, con enfriamiento configurable para no repetir correos.
- [Gobernanza] Se aplican las reglas de `copilot-instructions.md`: Go puro, sin dependencias externas, documentacion de archivos, modulos, base de datos y matriz de permisos.
- [Verificacion] `go test ./...` en `backend/`.

## [2026-05-10] Roles finos y ayuda privada super
- [Permisos] Se documentan modulos finos para CRM unificado, reservas, chat/tareas, horarios, asistencia, vehiculos, hoja de vida operativa, GPS, nomina, reportes, auditoria, backups, OnlyOffice y Nextcloud.
- [Backend] Las rutas empresariales recientes quedan asociadas a wrappers especificos y la prueba de seguridad de rutas reconoce esos wrappers.
- [Licencias] Se mantiene compatibilidad para licencias antiguas con modulos amplios (`ventas`, `seguridad`, `finanzas`, `inventario`, `clientes`) para evitar perdida de acceso tras separar modulos.
- [Super] `web/super_administrador.html` agrega el boton `Ayuda super administrador`, que abre `/ayuda/ayuda.html` dentro del panel.
- [Seguridad] `/ayuda/ayuda.html` permanece exclusiva de `super_administrador`; el rol `control_super_administrador` no recibe el boton ni la ruta en su lista limitada.
- [Docs] Se agrega `documentos/reporte_roles_ayuda_super_2026-05-10.md` y se actualizan README documental, resumen, descripcion del proyecto, descripcion de modulos, matriz de roles y ayuda web.

## [2026-05-07] QA E2E Motel Calipso y ajustes flotantes
- [QA Motel Calipso] Se ejecuta regresion autenticada sobre `empresa_id=7` con 60 modulos de Administrar empresa: escritorio con 60/60 modulos cargados y validacion dirigida posterior sin errores para activos fijos, auditoria, red social, control electrico, venta publica, radio y configuracion.
- [QA profunda] Se agrega runner `backend/tmp_tools/qa_calipso_operativo/deep_flows_calipso.mjs` para crear datos QA reales en parqueadero, WMS, centros de costo, activos fijos, red social con imagen, carta publica, venta publica y validacion QR publica. Resultado final: 6/6 pasos OK, sin errores de consola, red ni pagina.
- [QA movil] Se habilita viewport configurable en `frontend_buttons_calipso.mjs` y se recorren los 60 modulos en 390x844. Los dos hallazgos moviles se corrigen y se validan de forma dirigida en venta publica y asistencia de empleados.
- [UX flotante] El robot/secretaria separa los globos del avatar en movil, se evita que el drawer del asistente quede bajo el boton de favoritos y la radio flotante se compacta como boton circular inferior izquierdo para reducir bloqueos de botones.
- [Robustez frontend] Activos fijos exporta aunque el dashboard aun no haya cargado, auditoria tolera ausencia del rotulo de empresa, red social muestra errores controlados sin romper consola, graficos/estadisticas degradan con aviso visual.
- [Riesgos externos] Queda documentado que impresora fisica, sensores electricos reales, GPS fisico, DIAN y pasarelas en produccion requieren credenciales/dispositivos externos para prueba final fuera del entorno local.

## [2026-05-06] Modulos empresariales Colombia - fases compartidas
- [ERP Colombia] Se implementan `bancos_pagos`, `gestion_documental`, `cumplimiento_kyc`, `contratos_obligaciones` y `calidad_procesos` sobre nucleo compartido por `empresa_id`.
- [Backend] APIs privadas por modulo con acciones `dashboard`, `plantilla`, `reporte`, `registros`, `eventos`, `evidencias`, `registro`, `estado`, `evento`, `evidencia`, `importar_registros` y `seed_demo`.
- [UI] Pantallas administrativas compartidas con KPIs, reporte ejecutivo, CSV, seguimiento, cambio de estado y evidencias/soportes.
- [Workflow] Se agrega flujo de aprobaciones por nivel, destinatario, vencimiento, decision y bitacora.
- [Workflow] Se agrega gestion de tareas por registro con responsable, prioridad, vencimiento y estados operativos.
- [Operacion] Se agrega expediente 360 por registro para consolidar eventos, evidencias, aprobaciones, tareas, resumen y recomendacion.
- [Operacion] Se agrega agenda de alertas por modulo con vencidos, proximos vencimientos, tareas, aprobaciones pendientes y acceso al expediente.
- [Gobierno] Se agrega cierre controlado para impedir cierre sin evidencia o con aprobaciones/tareas abiertas.
- [Operacion] Se agrega generador de plan de accion para crear tareas desde alertas de agenda sin duplicar tareas abiertas.
- [Operacion] Se agrega tablero de responsables con carga pendiente, vencidos y recomendaciones por responsable.
- [Operacion] Se agrega tablero SLA con cumplimiento, semaforo, buckets de vencimiento y recomendaciones.
- [Operacion] Se agrega matriz de riesgo operativo con score, nivel, factores ponderados y recomendaciones.
- [Auditoria] Se agrega exportacion CSV multi-seccion con resumen, registros, agenda, SLA, riesgo, responsables, tareas, aprobaciones, evidencias y bitacora.
- [Operacion] Se agrega busqueda avanzada de backend por texto, estado, tipo, categoria, prioridad, responsable, vencidos y proximos vencimientos.
- [Operacion] Se agregan acciones masivas controladas para cambiar estado, prioridad y responsable con bitacora por registro.
- [BD] Tablas compartidas `empresa_modulos_colombia_registros`, `empresa_modulos_colombia_eventos`, `empresa_modulos_colombia_evidencias`, `empresa_modulos_colombia_aprobaciones` y `empresa_modulos_colombia_tareas`.
- [Docs/QA] Documentacion actualizada y validacion con `go test ./... -count=1`.

## [2026-05-06] Portal de Terceros y Certificados Tributarios
- [Finanzas/tributario] Nuevo modulo `portal_terceros_certificados` para proveedores, clientes, empleados, contratistas y terceros externos.
- [Backend] Nueva API administrativa `/api/empresa/portal_terceros_certificados` y API publica `/api/public/certificados_tributarios`.
- [UI] Nueva pantalla administrativa y nueva pagina publica de visualizacion/impresion de certificados por token.
- [Permisos/licencias] Nuevo modulo activable por licencia y controlado por roles financieros/contables.
- [Auditoria] Bitacora de descargas con IP, navegador, canal y fecha.

## [2026-05-06] Activos Fijos e Intangibles NIIF/Fiscal
- [Finanzas/contabilidad] Nuevo modulo formal `activos_fijos_niif_fiscal` para PPE e intangibles por empresa.
- [Backend] Nueva API `/api/empresa/activos_fijos_niif_fiscal` con dashboard, libro, depreciaciones, eventos, registro de activos y datos demo.
- [BD] Se amplio `empresa_contabilidad_activos_fijos` con campos NIIF/fiscales, deterioro, valor razonable, valor fiscal y diferencia NIIF/fiscal.
- [UI] Nueva pantalla `web/administrar_empresa/activos_fijos_niif_fiscal.html` enlazada desde el menu principal y el centro financiero.
- [Permisos/licencias] Nuevo modulo `activos_fijos_niif_fiscal` activable por licencia y roles.

## [2026-05-06] Propiedad horizontal y promocion por asesor
- [Plantillas] Nuevo modulo `propiedad_horizontal` para copropiedades, conjuntos, edificios y condominios.
- [Backend] Nueva API `/api/empresa/propiedad_horizontal` con configuracion, unidades, personas, cargos, recaudos, PQR, asambleas, dashboard y datos demo.
- [Permisos] Nueva clave de modulo/licencia `propiedad_horizontal`, pagina `linkPropiedadHorizontal` y wrapper `WithEmpresaPropiedadHorizontalPermissions`.
- [Super] En `Asesor comercial` se agrega promocion activable/desactivable para descuento adicional por codigo de asesor, con porcentaje configurable.
- [Checkout] `pagar_licencia.html`, Wompi, Epayco y activacion sin pago consideran `asesor_id` para aplicar descuento si la promocion esta activa y conservar comisiones.
- [Docs/QA] Se agregan `documentos/propiedad_horizontal.md` y `documentos/promocion_asesor_licencias.md`; pruebas `go test ./... -count=1` en `backend/`.

## [2026-05-06] Cierre y bloqueo fiscal
- [Finanzas/contabilidad] Nuevo modulo `cierre_fiscal` para proteger periodos cerrados, documentos reportados y operaciones post-cierre.
- [Backend] Nueva API `/api/empresa/cierre_fiscal` con dashboard, politicas, periodos, excepciones, validacion, bitacora y datos demo.
- [Base de datos] Nuevas tablas `empresa_cierre_fiscal_politicas`, `empresa_cierre_fiscal_periodos`, `empresa_cierre_fiscal_excepciones` y `empresa_cierre_fiscal_eventos`, todas por `empresa_id`.
- [Integracion] El cierre/reapertura de `contabilidad_colombia` sincroniza el periodo fiscal para evitar doble control de cierres.
- [Permisos] Nueva clave de modulo/licencia `cierre_fiscal`, paginas `linkCierreFiscal`/`linkCierreFiscalMenu` y wrapper `WithEmpresaCierreFiscalPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/cierre_fiscal.html` dentro del Centro financiero y contable.
- [Docs/QA] Se crea `documentos/cierre_fiscal.md`; pruebas `go test ./... -count=1` en `backend/`.

## [2026-05-06] Centros de costo y rentabilidad
- [Finanzas/contabilidad] Nuevo modulo formal `centros_costo` para medir rentabilidad por sucursal, area, unidad de negocio o proyecto.
- [Backend] Nueva API `/api/empresa/centros_costo` con dashboard, centros, reglas, presupuestos, movimientos integrados y datos demo.
- [Base de datos] Nuevas tablas `empresa_centros_costo`, `empresa_centros_costo_reglas` y `empresa_centros_costo_presupuestos`, aisladas por `empresa_id`.
- [Integracion] El dashboard consolida movimientos existentes con `centro_costo` desde contabilidad Colombia, tesoreria, compras avanzadas, soportes OCR/IA y AIU construccion, sin duplicar modulos financieros.
- [Permisos] Nueva clave de modulo/licencia `centros_costo`, paginas `linkCentrosCosto`/`linkCentrosCostoMenu` y wrapper `WithEmpresaCentrosCostoPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/centros_costo.html` dentro del Centro financiero y contable, con modo claro/oscuro, exportacion CSV y formularios CRUD.
- [Docs/QA] Se crea `documentos/centros_costo.md`; pruebas `go test ./... -count=1` en `backend/`.

## [2026-05-06] QA transversal y profesionalizacion de modulos
- [Portal publico] `web/index.html` actualiza la descripcion de modulos para incluir Cobranza, Portal contador, Captura IA/OCR de compras y gastos, AIU construccion, Parqueaderos con ticket QR y Apartamentos turisticos en la seccion publica y tarjetas fallback.
- [Permisos] `web/js/administrar_empresa.js` reconoce `administrador_total` como rol de acceso total en la evaluacion local del menu, alineado con backend.
- [Rendimiento] Los dashboards de `cobranza`, `portal_contador` y `soportes_compras_ia` evitan validaciones repetidas de esquema dentro de la misma peticion.
- [Frontend] La pagina `soportes_compras_ia.html` valida enlaces dinamicos de archivos antes de renderizarlos.
- [QA Motel Calipso] Login super administrador, paginas principales y APIs de modulos recientes verificadas con HTTP 200 para `empresa_id=7`.
- [Auditoria] Revision estatica de enlaces, botones `onclick` e IDs de paginas empresariales; sin botones muertos relevantes.
- [Docs] Se agrega `documentos/reporte_qa_modulos_2026-05-06.md` con alcance, pruebas, observaciones y estado final.

## [2026-05-06] Portal contador
- [Finanzas/contabilidad] Nuevo modulo `portal_contador` como oficina virtual para contadores y firmas contables.
- [Backend] Nueva API `/api/empresa/portal_contador` con dashboard, clientes, obligaciones, solicitudes, comunicaciones y datos demo.
- [Base de datos] Nuevas tablas `empresa_portal_contador_clientes`, `empresa_portal_contador_obligaciones`, `empresa_portal_contador_solicitudes` y `empresa_portal_contador_comunicaciones`, todas aisladas por `empresa_id`.
- [Permisos] Nueva clave de modulo/licencia `portal_contador`, paginas `linkPortalContador`/`linkPortalContadorMenu` y wrapper `WithEmpresaPortalContadorPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/portal_contador.html` dentro del Centro financiero y contable.
- [Docs/QA] Se crea `documentos/portal_contador.md`; pruebas `go test ./db -run TestPortalContador -count=1` y `go test ./... -count=1`.

## [2026-05-06] Gestion de cobranza
- [Finanzas] Nuevo modulo profesional `cobranza` para recuperar cartera sin duplicar cuentas por cobrar.
- [Backend] Nueva API `/api/empresa/cobranza` con dashboard, cuentas, plantillas, campanas, gestiones, promesas, simulacion de envio y datos demo.
- [Base de datos] Nuevas tablas `empresa_cobranza_plantillas`, `empresa_cobranza_campanas`, `empresa_cobranza_gestiones` y `empresa_cobranza_promesas`, enlazadas por `empresa_id` y `cuenta_id`.
- [Permisos] Nueva clave de modulo/licencia `cobranza`, paginas `linkCobranza`/`linkCobranzaMenu` y wrapper `WithEmpresaCobranzaPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/cobranza.html` dentro del Centro financiero y contable.
- [Docs/QA] Se crea `documentos/cobranza.md`; pruebas `go test ./db -run TestCobranza -count=1` y `go test ./... -count=1`.

## [2026-05-06] AIU construccion
- [Mejora profesional] El modulo AIU ahora incluye responsable, centro de costo, modalidad contractual, riesgo, avance, retenciones, anticipo, garantia, neto a cobrar y flujo validado de estados.
- [Backend] Nuevas acciones `facturas`, `reporte` y `estado`; el dashboard suma contratos/facturas, neto, retenciones, pendiente por facturar y alertas.
- [Frontend] Pantalla reorganizada con KPIs, filtros, acciones de estado, resumen financiero, facturas recientes y exportacion CSV.
- [Facturacion/obra] Nuevo modulo `aiu_construccion` para arquitectos, constructoras, contratistas y pequenas empresas de obra.
- [Backend] Nueva API `/api/empresa/aiu_construccion` con dashboard, contratos, conceptos, calculadora AIU, datos demo y generacion de factura electronica AIU.
- [Base de datos] Nuevas tablas `empresa_aiu_contratos`, `empresa_aiu_items` y `empresa_aiu_facturas`, todas aisladas por `empresa_id`.
- [Permisos] Nueva clave de licencia/rol `aiu_construccion`, pagina `linkAIUConstruccion` y wrapper `WithEmpresaAIUConstruccionPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/aiu_construccion.html` enlazada desde el submenu de facturacion electronica.
- [Docs/QA] Se crea `documentos/aiu_construccion.md`; pruebas `go test ./db -run Test.*AIU -count=1` y `go test ./...`.

## [2026-05-06] Documentos electronicos DIAN/Siigo
- [Facturacion electronica] Se amplia el ciclo documental existente para cubrir factura electronica, nota credito, nota debito, documento soporte, nomina electronica y documento equivalente POS electronico por empresa.
- [Backend] `/api/empresa/facturacion_electronica` normaliza aliases Siigo/DIAN, valida tipos soportados, conserva auditoria/eventos contables y usa la misma cola DIAN/proveedor para los nuevos documentos.
- [Frontend] `web/administrar_empresa/facturacion_electronica.html` agrega selector de tipo documental y botones rapidos para emitir factura, notas, soporte, nomina y POS electronico.
- [Docs/QA] Verificado con pruebas unitarias enfocadas en normalizacion y transiciones documentales, mas prueba de defaults de facturacion por pais.

## [2026-05-06] CRM y ventas avanzadas
- [CRM] Se amplia `clientes`/CRM comercial sin duplicar modulo con metas, forecast, scoring, agenda y conversion de lead a cotizacion.
- [Backend] Nueva API `/api/empresa/crm_avanzado` con dashboard, metas, scores, demo y cotizacion desde lead.
- [Base de datos] Nueva tabla `empresa_crm_metas_comerciales`; el dashboard reutiliza `crm_leads`, `crm_interacciones`, `empresa_cotizaciones_venta` y `empresa_pedidos_venta`.
- [Permisos] Se reutiliza el modulo/licencia `clientes`, con pagina `linkCRMAvanzado`.
- [Frontend] Nueva pantalla `web/administrar_empresa/crm_ventas_avanzadas.html` en Ventas, clientes y caja.
- [Docs/QA] Se crea `documentos/crm_ventas_avanzadas.md` y el QA Calipso valida lead, meta, scoring, cotizacion y forecast.

## [2026-05-06] Inventario avanzado
- [Inventario] Se amplia el modulo existente `inventario` sin duplicarlo con lotes, seriales, reservas, vencimientos y valorizacion por bodega.
- [Backend] Nueva API `/api/empresa/inventario_avanzado` con dashboard, lotes, seriales, reservas, valorizacion, demo y confirmacion de salida.
- [Base de datos] Nuevas tablas `empresa_inventario_lotes_avanzados`, `empresa_inventario_seriales_avanzados` y `empresa_inventario_reservas_avanzadas`, todas aisladas por `empresa_id`.
- [Integracion] La entrada de lote actualiza `inventario_existencias` y `inventario_movimientos`, manteniendo el kardex existente como eje operativo.
- [Permisos] Se reutiliza el modulo/licencia `inventario`, con pagina `linkInventarioAvanzado`.
- [Frontend] Nueva pantalla `web/administrar_empresa/inventario_avanzado.html` enlazada desde el menu de Productos.
- [Docs/QA] Se crea `documentos/inventario_avanzado.md` y el QA Calipso valida lote, serial, reserva, salida y valorizacion.

## [2026-05-06] Compras avanzadas
- [Compras] Se amplia el modulo existente `compras` sin duplicarlo con requisiciones internas, cotizaciones, aprobaciones y recepcion parcial/total por empresa.
- [Backend] Nueva API `/api/empresa/compras_avanzadas` con dashboard, requisiciones, detalle, cotizaciones, aprobaciones, recepciones y datos demo.
- [Base de datos] Nuevas tablas `empresa_compras_requisiciones`, `empresa_compras_requisicion_items`, `empresa_compras_cotizaciones`, `empresa_compras_aprobaciones`, `empresa_compras_recepciones_avanzadas` y `empresa_compras_recepcion_items_avanzadas`.
- [Permisos] Se reutiliza el modulo/licencia `compras`, con pagina `linkComprasAvanzadas` dentro del submenu de Compras.
- [Frontend] Nueva pantalla `web/administrar_empresa/compras_avanzadas.html` enlazada desde `compras_menu.html`.
- [Docs/QA] Se crea `documentos/compras_avanzadas.md` y el QA Calipso valida requisicion, cotizacion, aprobacion, recepcion y dashboard.

## [2026-05-06] Importaciones y costeo de nacionalizacion
- [Compras] Se agrega el modulo `importaciones_costeo` para compras internacionales, incoterms, TRM, items importados, fletes, aranceles, aduana y costo aterrizado.
- [Backend] Nueva API `/api/empresa/importaciones_costeo` con dashboard, importaciones, detalle, items, costos, distribucion y datos demo por empresa.
- [Base de datos] Nuevas tablas `empresa_importaciones_costeo`, `empresa_importaciones_costeo_items` y `empresa_importaciones_costeo_costos`, todas aisladas por `empresa_id`.
- [Permisos] Nueva clave de licencia `importaciones_costeo`, pagina `linkImportacionesCosteo` y wrapper `WithEmpresaImportacionesCosteoPermissions`.
- [Frontend] Nueva pantalla `web/administrar_empresa/importaciones_costeo.html` dentro de Inventario y compras.
- [Docs/QA] Se crea `documentos/importaciones_costeo.md` y el QA Calipso valida embarque, items, costos, distribucion y dashboard.

## [2026-05-06] Activos fijos avanzado
- [Contabilidad] Se amplia `contabilidad_colombia_avanzada` con activos fijos avanzados sin duplicar modulo: depreciacion por periodo, eventos, mantenimiento, traslados y bajas.
- [Backend] La API `/api/empresa/contabilidad_colombia_avanzada` agrega `activos_resumen`, `activos_depreciaciones`, `activos_eventos`, `generar_depreciacion_activos` y `activo_evento`.
- [Base de datos] Se enriquecen activos fijos con serial, placa, metodo de depreciacion, centro de costo, proveedor, poliza y mantenimiento programado; se agregan `empresa_contabilidad_activos_depreciacion` y `empresa_contabilidad_activos_eventos`.
- [Frontend] La suite contable agrega pestaña `Activos avanzado` para generar depreciacion, registrar eventos y consultar inventario gerencial.
- [Docs/QA] Se crea `documentos/activos_fijos_avanzado.md` y el QA Calipso valida activo, depreciacion, mantenimiento y resumen avanzado.

## [2026-05-06] Nomina Colombia avanzada
- [Nomina] Se amplia el modulo existente `nomina_sueldos` sin duplicarlo: conceptos legales Colombia, novedades aprobables y resumen PILA por empresa.
- [Backend] La API `/api/empresa/nomina` agrega acciones `conceptos_colombia`, `novedades_colombia`, `pila_colombia`, `dashboard_colombia`, `concepto_colombia`, `novedad_colombia`, `generar_pila` y `seed_colombia`.
- [Base de datos] Nuevas tablas aisladas por `empresa_id`: `empresa_nomina_colombia_conceptos`, `empresa_nomina_colombia_novedades` y `empresa_nomina_colombia_pila_resumen`.
- [Frontend] La pantalla de nomina incluye una seccion `Nomina Colombia avanzada` con KPIs, conceptos, novedades y PILA.
- [Docs/QA] Se crea `documentos/nomina_colombia_avanzada.md` y el QA Calipso registra empleado, concepto, novedad, liquidacion y PILA.

## [2026-05-06] Tesoreria y presupuesto
- [Tesoreria] Se agrega el modulo `tesoreria_presupuesto` para cuentas banco/caja, presupuestos, partidas, ejecucion y flujo de caja proyectado por empresa.
- [Backend] Nueva API `/api/empresa/tesoreria_presupuesto`, tablas aisladas por `empresa_id` y dashboard con saldo disponible, ingresos/egresos proyectados y flujo neto.
- [Permisos] Nueva clave de licencia `tesoreria_presupuesto`, pagina `linkTesoreriaPresupuesto` y wrapper `WithEmpresaTesoreriaPresupuestoPermissions`, alineado con roles financieros/contables.
- [Frontend] Nueva pantalla `web/administrar_empresa/tesoreria_presupuesto.html` dentro del Centro financiero y contable.
- [Docs/QA] Se crea `documentos/tesoreria_presupuesto.md` y se amplian pruebas de normalizacion y QA Calipso.

## [2026-05-06] Produccion / MRP empresarial
- [Produccion] Se agrega el modulo `produccion_mrp` para recetas/BOM, componentes, ordenes de produccion, consumos, control de calidad, plan MRP y datos demo por empresa.
- [Backend] Nueva API `/api/empresa/produccion_mrp`, tablas aisladas por `empresa_id` y dashboard operativo con KPIs de ordenes, calidad y costos.
- [Permisos] Nueva clave de licencia `produccion_mrp`, pagina `linkProduccionMRP` y wrapper `WithEmpresaProduccionMRPPermissions`; roles de inventario/compras pueden operar segun matriz.
- [Frontend] Nueva pantalla `web/administrar_empresa/produccion_mrp.html` bajo Inventario y compras, con tabs de recetas, ordenes, consumos/calidad, MRP y configuracion.
- [Docs/QA] Se crea `documentos/produccion_mrp.md` y pruebas unitarias de normalizacion del modulo.

## [2026-05-05] Reportes colombianos avanzados
- [Reportes] Se agregan datasets profesionales que suelen exigir los sistemas contables/POS usados en Colombia: ventas diarias por medio de pago, rentabilidad por producto, Kardex valorizado, compras detalladas por proveedor, balance de prueba, libro auxiliar, libro mayor, impuestos/retenciones, informacion exogena base, edades de cartera CxC y edades CxP.
- [Backend] Los reportes quedan en el endpoint existente `/api/empresa/reportes`, reutilizan filtros por rango, exportacion `JSON/CSV/TXT/XLS/PDF`, programacion/plantillas y separacion estricta por `empresa_id`.
- [Frontend] El menu de reportes agrega accesos directos a los nuevos datasets sin duplicar modulos ni romper los enlaces anteriores.
- [QA] Se agrega prueba de catalogo para evitar datasets duplicados y asegurar metadatos/formats completos.
- [Administrar empresa] El menu principal queda reorganizado en categorias plegables; se separan `Operacion por tipo` y `Ventas, clientes y caja` para reducir botones visibles sin cambiar rutas ni permisos.

## [2026-05-05] Organizacion de modulos y control super
- [Administrar empresa] Se fusiona la navegacion repetida de Finanzas, Contabilidad Colombia y Suite contable Colombia bajo `Centro financiero y contable`, manteniendo rutas internas compatibles.
- [Ayuda] La ayuda administrativa principal `/ayuda/ayuda.html` queda protegida para `super_administrador`; las ayudas publicas especificas se conservan.
- [Super] Se agrega el rol `control_super_administrador` para supervision limitada de administradores, seguridad, errores, metricas y reportes globales.
- [Seguridad] El contralor super no puede eliminar ni desactivar el super administrador principal ni administrar otros contralores super.

## [2026-05-06] Captura inteligente de compras y gastos con OCR/IA
- [Compras] Se agrega el modulo `soportes_compras_ia` para radicar soportes con foto, PDF o XML, detectar duplicados y gestionar estados de revision, aprobacion y contabilizacion.
- [IA] La extraccion usa la capa existente de IA con modelo recomendado `openai:gpt-5.5`, prompt contable colombiano y registro en historial de consultas IA.
- [Backend] Nueva API protegida `/api/empresa/soportes_compras_ia`, tablas `empresa_soportes_compras_ia` y `empresa_soportes_compras_ia_eventos`, guardado de archivos por empresa y conversion a `empresa_cuentas_por_pagar`.
- [Frontend] Nueva pantalla `web/administrar_empresa/soportes_compras_ia.html` enlazada desde el menu de Compras con dashboard, carga de archivos, acciones GPT-5.5, aprobacion, rechazo, contabilizacion y exportacion CSV.
- [Permisos] Nuevo modulo de rol/licencia `soportes_compras_ia` con acceso operativo para compras, contabilidad, supervisor y administrador de empresa.
- [Docs/QA] Se crea `documentos/soportes_compras_ia.md`; validado con `go test ./db -run Test.*Soporte.*IA -count=1`, `go test ./... -count=1` y `git diff --check`.

## [2026-05-05] Suite contable Colombia avanzada
- [Contabilidad] Se agrega `contabilidad_colombia_avanzada` con informacion exogena DIAN/medios magneticos, nomina electronica, documento soporte, activos fijos, cartera/CxP y libros oficiales por empresa.
- [Backend] Nueva API `/api/empresa/contabilidad_colombia_avanzada`, tablas empresariales aisladas por `empresa_id` y generacion de exogena/libros desde comprobantes contabilizados del nucleo `contabilidad_colombia`.
- [Permisos] Nuevo modulo de licencia `contabilidad_colombia_avanzada`, pagina `linkContabilidadColombiaAvanzada` y wrapper `WithEmpresaContabilidadColombiaAvanzadaPermissions`.
- [Frontend] Nueva vista `web/administrar_empresa/contabilidad_colombia_avanzada.html` con dashboard y pestañas profesionales para cada submodulo.
- [Docs/QA] Se crea `documentos/contabilidad_colombia_avanzada.md`; pruebas Go y auditoria de rutas/permisos actualizadas.

## [2026-05-05] Portal publico, carta QR y Motel Calipso publicado
## [2026-05-18] Configuracion empresarial por paginas
- [Frontend] El submenu `Configuracion empresarial` se divide en paginas enfocadas para Productos y pedidos, Identidad visual, Formato monetario, Cobro operativo, Reporte de corte, Respaldo y Pasarelas de pago.
- [Permisos] Se registran claves de pagina para las nuevas secciones bajo el modulo `seguridad` con accion de actualizacion, conservando los endpoints empresariales existentes.
- [QA] Verificacion visual desktop y movil con emulacion CDP de 390 px sin desborde horizontal; pruebas Go enfocadas de catalogo de permisos.

- [Permisos] Se audita el enlace por empresa de todos los modulos visibles en Administrar empresa: menu frontend, catalogo de paginas backend y licencias quedan alineados, sin claves duplicadas ni rutas `/api/empresa` duplicadas.
- [Carnets] Se agrega modulo empresarial profesional `/api/empresa/carnets` y `web/administrar_empresa/carnets.html` para emitir carnets modernos de empleados/usuarios con plantillas, QR, foto, exportacion PNG/SVG, impresion y bitacora.
- [Licencias] Se agrega modulo `carnets`, pagina `linkCarnets` y soporte en `licencias.modulos_habilitados` para activarlo/desactivarlo por empresa.
- [Aislamiento] Se rectifica el control multiempresa en wrappers `WithEmpresa*`: query, cabecera `X-Empresa-ID`, formulario/multipart y JSON no pueden declarar empresas distintas para una misma peticion.
- [Seguridad] Las rutas privadas `/api/empresa/...` registradas en `main.go`, chat IA empresarial y modulos ERP faltantes quedan revisadas bajo wrappers de empresa.
- [Portal] Se actualizan las descripciones de modulos en `web/index.html` para reflejar el alcance real del sistema: POS, estaciones, hotel/motel, gimnasio, odontologia, domicilios tipo Rappi, Taxi System tipo Uber, turnos, control electrico, inventario, finanzas, facturacion, usuarios/permisos/licencias, IA, reportes, carta publica QR, red social y hoja de vida.
- [Carta publica] `visualizar_productos_y_precios_publico.html` queda documentada y habilitada como pagina publica de solo lectura, directa o bajo `/{empresa_slug}/visualizar_productos_y_precios_publico.html`, con QR exportable desde Administrar empresa.
- [Motel Calipso] Se documenta la publicacion real del slug `motel-calipso` con venta publica, carta publica, paginas, items de ejemplo y publicaciones en red social comercial.
- [Seguridad] `AuthMiddleware` permite la carta publica sin sesion; la administracion conserva control por modulo `venta_publica`, rol, licencia y pagina `linkCartaProductosPublica`.
- [Verificacion] `go test ./utils`; `go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1`; validacion productiva HTTP 200 de venta publica, carta publica y red social.

## [2026-05-05] Roles y licencias para modulos plantillas
- [Permisos] Se agregan modulos independientes para venta publica/carta, gimnasio, taxi system, domicilios, alquileres, odontologia, turnos de atencion y control electrico.
- [Licencias] La pantalla de licencias permite activar/desactivar estos modulos desde `modulos_habilitados`, con presets actualizados.
- [Backend] Los endpoints administrativos plantillas usan wrappers dedicados para que licencia, rol y pagina del menu bloqueen con `403` cuando corresponda.
- [Docs] Se actualiza la matriz de roles/permisos y la documentacion de domicilios.

## [2026-05-05] Modulo profesional de domicilios
- [Domicilios] Se agrega modulo tipo marketplace para restaurantes, domiciliarios, cliente publico y central administrativa con pedidos, menu, tarifas, comisiones, autoasignacion y estados operativos.
- [Tracking] Los domiciliarios reportan ubicacion GPS desde el navegador movil; la central visualiza restaurantes, clientes y domiciliarios en mapa Leaflet.
- [Seguridad] Cada pedido genera codigo de entrega y token de seguimiento para el cliente; restaurantes y domiciliarios usan PIN operativo.
- [Docs] Se crea `documentos/domicilios_profesional.md` con superficies, flujo, endpoints, datos demo y recomendaciones de produccion.

## [2026-05-05] Taxi System profesional con mapa y GPS
- [Taxi System] El panel administrativo incorpora operacion tipo Uber con mapa, filtros por conductores disponibles/ocupados, solicitudes y GPS externos.
- [GPS] Se integra el inventario corporativo de dispositivos GPS dentro de Taxi System para registrar apps moviles, trackers, OBD2, celulares, tablets, dashcams y webhooks con protocolo/proveedor.
- [Conductores] Se agregan campos de asociacion GPS en `empresa_taxi_drivers`: `gps_dispositivo_id`, `gps_codigo`, `gps_tipo`, `gps_proveedor` y `gps_protocolo`.
- [Docs] Se crea `documentos/taxi_system_profesional.md` con alcance, superficies, endpoints y validacion.

## [2026-05-04] Control electrico Raspberry Pi por estacion
- [Control electrico] Nuevo modulo en Administrar empresa para configurar Raspberry Pi, IP/puerto/ruta API, token opcional, timeout y sincronizacion automatica.
- [Estaciones] Cada estacion puede mapearse a multiples relés GPIO con salida/carga (luces, jacuzzi, aire, puerta u otro), nombre, pin, logica activo alto, pulso opcional y prueba manual ON/OFF.
- [Carrito] El carrito de estacion incorpora boton `Control electrico` para abrir un panel operativo y controlar manualmente salidas de la estacion sin salir de la venta.
- [Automatizacion] Al activar/recuperar/reabrir una estacion se envia `on`; al pagar/cerrar/desactivar se envia `off`. Tambien se engancha con autoactivacion por sensor de puertas.
- [Auditoria] Se agrega bitacora electrica por empresa con comando, estado objetivo, GPIO, HTTP status, respuesta/error, actor, origen y fecha.
- [Backend] Se registran tablas `empresa_control_electrico_config`, `empresa_control_electrico_reles` y `empresa_control_electrico_eventos`, mas endpoint protegido `/api/empresa/control_electrico`.

## [2026-05-02] Gimnasio, impresoras y documentacion del proyecto
- [Gimnasio] Se robustece el esquema del modulo con migraciones defensivas para tablas antiguas de empresas, evitando errores internos al abrir dashboard, acceso, credenciales o dispositivos cuando faltaban columnas historicas.
- [Gimnasio] Se agrega preconfiguracion operativa propia del modulo: sede principal, RFID/NFC/QR, planes base, clases iniciales y dispositivos de acceso, todo aplicable desde el dashboard del gimnasio.
- [Impresoras] Se corrige el guardado de configuracion avanzada para que `modo_documento_venta` gobierne correctamente la activacion o desactivacion de facturacion electronica automatica.
- [Impresoras] Se incorpora `cajon_monedero` como funcionalidad asignable de impresora dentro de `Configuracion > Impresora`, alineando la UI con la operacion real de caja.
- [Docs] Se actualiza `RESUMEN_DEL_PROYECTO.md` para reflejar configuracion guiada por IA, impresion empresarial, horarios laborales y modulos plantillas ya integrados como gimnasio, odontologia, taxi system, turnos de atencion y alquileres.

## [2026-04-30] Pagos, chat IA, empresas compartidas, hoja de vida operativa y documentos dinamicos
- [Pagos/Epayco] Smart Checkout v2 conserva fallback clasico firmado por POST a `https://secure.payco.co/checkout.php`; se elimina la redireccion GET que producia XML `AccessDenied` y se documenta el requisito de `epayco.customer_id` para fallback.
- [Pagos/Epayco] El fallback clasico resuelve su modo con `epayco.customer_id` + `epayco.checkout_key`/`epayco.p_key`, separado de las llaves Smart Checkout, para no enviar cuentas reales como pruebas y evitar el error "El comercio no fue reconocido".
- [Chat IA] La secretaria IA 3D se rediseña como avatar estilo caricatura ejecutiva joven y habla siempre con voz femenina (`es-CO-female`), manteniendo el robot con voz configurable.
- [Empresas compartidas] El editor de empresa permite consultar y retirar administradores compartidos desde ambos lados del acceso, con trazabilidad del actor.
- [Administrar empresa] Se implementa la hoja de vida operativa universal para motos de taller, pacientes, vehiculos, equipos, activos o mascotas, con ficha, eventos, servicios, alertas y resumen operativo.
- [Documentos IA] Se documenta el flujo `/generate` + `/download` para generar documentos dinamicos con IA/templates y exportar PDF, DOCX, XLSX, HTML, TXT o JSON.
- Nueva funcionalidad: MÒ³dulo Red Social Comercial con portal pÒºblico y administraciÒ³n por empresa. EliminaciÒ³n de modulo juegos y venta de licencias desde cliente.

## [2026-04-23] Retiro Tipos de usuario (panel super)
- [Super/DB] Eliminación del módulo Tipos de usuario: sin API ni UI; tabla `tipos_de_usuario` removida al arranque; documentación alineada.

## [2026-04-23] reCAPTCHA, backup y manual de instalación
- [Docs/Operación] Se actualizó el manual de instalación con reCAPTCHA v2/v3/Enterprise, variables, panel super y fallos frecuentes (dominios, tipo de clave). Se documentan las claves y copias best-effort en `backup/super_administrador` y `backup/empresas/<empresa_id>`. Ajustes en `descripcion_de_archivos` e `historial_de_cambios` y alineación con `CHANGELOG.md` raíz.

## [2026-04-20] Limpieza Total Themes
- [UI/Temas] Auditoría y barrido de más de 50 páginas y scripts en web/administrar_empresa, web/super y páginas públicas para limpiar colores fijos, migrando lógicas JS a .classList.add('text-danger') y respetando las 6 paletas dinámicas. Completado barrido masivo de vistas.
- **2026-04-30 - Pagos ePayco de licencias**: el fallback estandar ahora usa `checkout.js` con `external: "true"` y `PUBLIC_KEY`; se evita el POST legacy a `secure.payco.co/checkout.php`, `P_KEY` queda solo en backend para validacion de webhooks con SHA256 y el frontend `pagar_licencia.html` soporta `checkout_type=classic_js`. Verificacion: `go test ./handlers -run Test.*Epayco -count=1` y `go test ./... -count=1`.

## [2026-05-03] Documentacion, ayuda y estado operativo de modulos
- [Docs] Se crea `documentos/reporte_estado_modulos_2026-05-03.md` con estado compacto por modulo, observaciones de calidad y dependencias pendientes de certificacion.
- [Ayuda] Se actualiza `web/ayuda/ayuda.html` con una seccion de estado operativo, estaciones/carrito, tarjetas adaptables, indicadores del panel y limites honestos de validacion.
- [Operacion] Se documentan los cambios recientes: carrito desde estacion, pago con retorno a estaciones, `USD / COP` primero y despliegue VPS correcto.
- **Logistica avanzada / WMS**: nuevo modulo `logistica_wms` con ubicaciones internas, ordenes WMS, picking, packing, despachos, rutas, bitacora, permisos/licencia, pantalla administrativa y documentacion. Verificacion prevista: pruebas unitarias del motor WMS, `go test ./... -count=1` y `git diff --check`.

- **Declaraciones Tributarias y Motor de Impuestos Colombia**: nuevo modulo `declaraciones_tributarias` con API privada, dashboard, preliquidacion, calendario editable, saldos a pagar/favor, movimientos de conciliacion, permisos/licencia, pantalla administrativa y documentacion. Verificacion prevista: pruebas unitarias del motor, `go test ./... -count=1` y `git diff --check`.
- 2026-05-06: implementados modulos empresariales Colombia `bancos_pagos`, `gestion_documental`, `cumplimiento_kyc`, `contratos_obligaciones` y `calidad_procesos` con nucleo compartido, APIs privadas por empresa, paginas administrativas, permisos/licencias y documentacion.
## [2026-05-11] Contrato universal de 30 plantillas
- [Backend] `backend/handlers/empresa_plantillas_integracion.go` deja de publicar acciones de migracion manual antigua y declara las plantillas clasicas como plantillas sobre nucleo comun.
- [Licencias] La activacion de licencia aplica la preconfiguracion idempotente del tipo de empresa sin ejecutar migraciones automaticas.
- [Frontend] La matriz empresarial queda como auditoria de plantilla, nucleo, permisos, flujo de venta y reportes; los dashboards clasicos ya no muestran botones de migracion manual.
- [QA] `go test ./...`; validacion JS de catalogos y pantallas empresariales tocadas.
## [2026-05-12] Menu empresarial ajustado
- [Frontend] `web/administrar_empresa.html` elimina el cuadro de evidencia `Plantillas · conteo · API/local` del encabezado del menu lateral.
- [Navegacion] `Soluciones por negocio` queda reubicado en la parte baja del menu, inmediatamente encima de `Administracion`.
- [Alcance] Sin cambios de API, permisos, base de datos ni dependencias.
- Creditos diarios para ventas financiadas de motos.
	- Archivos modificados: `backend/db/creditos.go`, `backend/handlers/creditos.go`, `backend/db/creditos_postgres_test.go`, `backend/main.go`, `web/administrar_empresa/creditos.html`, `documentos/estructura_bd.md`, `documentos/descripcion_del_proyecto`, `documentos/descripcion_de_modulos`, `documentos/historial_de_cambios`.
	- Descripcion: el modulo de creditos registra periodicidad de cuota, valor pactado y omision opcional de domingos; permite cuotas diarias largas y expone cuotas/dias vencidos para ver rapidamente cuanto debe cada cliente.
	- Verificacion: `go test ./db -run "TestCredito" -count=1`, `go test ./handlers -run '^$' -count=1` y sintaxis JS del modulo.

## [2026-05-28] Descarga informacion empresa
- [Selector] El boton de descarga de cada empresa visible abre la pagina dedicada sin error y permite exportar JSON, PDF, XLS, CSV o TXT.
- [Backend] `system_empresas_export.go` deja de abortar todo el snapshot cuando una tabla falla; compara identificadores como texto y agrega advertencias por tabla.
- [QA] Se comprobo visualmente el flujo selector -> pagina de descarga -> descarga JSON.

## [2026-05-28] Instalacion PWA desde login
- [PWA] `manifest.webmanifest` y `sw.js` quedan disponibles sin sesion para habilitar el flujo nativo de instalacion desde el login.
- [Portal] `index.html` usa el mismo icono PWA que `login.html` como favicon, apple touch icon y logo visible del encabezado.
- [QA] Se comprobo visualmente el boton `Instalar app` y la consistencia de iconos del index.

## [2026-05-28] Rediseño descarga informacion empresa
- [UX] La pagina de descarga empresarial queda en una sola accion: elegir formato y descargar.
- [Frontend] Se retiran tarjetas tecnicas y vista previa de tablas de `descargar_informacion_de_la_empresa.html`.
- [QA] Se comprobo visualmente en escritorio y movil, incluyendo descarga CSV.

## [2026-05-28] Apariencias oscuras profesionales
- [Temas] El selector del menu flotante incorpora `Negro Absoluto` y `Obsidiana Profesional`.
- [Frontend] Los iframes y subpaginas internas reconocen `dark-absolute` y `dark-obsidian` al heredar la apariencia.
- [QA] Se comprobo visualmente el cambio de tema y guardado de preferencia.

## [2026-05-29] Facturacion offline en carrito
- [Carrito] El aviso flotante de conectividad ahora tambien confirma cuando internet vuelve y cuando las ventas offline quedan sincronizadas.
- [UX] El mensaje online permanece visible aunque el pago deje el carrito inactivo y el panel interno cambie de estado.
- [QA] Se comprobo visualmente con Motel Calipso: venta directa, corte de internet simulado, venta provisional offline, cola local pendiente y sincronizacion al restaurar conexion.

## [2026-05-29] Panel empresarial
- [UX] La tarjeta de clima del panel de administrar empresa queda aproximadamente 20% mas compacta verticalmente en escritorio y movil.

## [2026-05-29] Cajeros simultaneos por estaciones
- [Backend] Los endpoints de carritos e items validan `estaciones_config.acceso_estaciones_cajeros` para impedir operar estaciones no asignadas por URL/API.
- [Frontend] `Administrar usuarios` permite activar el control, marcar estaciones por cajero y definir si ve la estacion Caja.
- [Caja] Los totales de la tarjeta Caja se consultan por usuario autenticado, manteniendo cuentas y corte de turno independientes para varias cajas simultaneas.
- [Docs] Se documenta el flujo operativo y el mapa de modulos para estaciones, usuarios y caja.

## [2026-05-29] Favoritos con icono original
- [Panel] Los favoritos del panel empresarial muestran el icono real del boton original del menu, no la estrella de favorito.
- [Navegacion] Al marcar un favorito nuevo se guarda tambien su icono de origen para conservarlo en la tarjeta.

## [2026-05-30] Bandeja de correo corporativo
- [Email corporativo] El autologin de SnappyMail preserva la redireccion `index.php?sso&hash=...` al agregar parametros de tema.
- [Panel] La bandeja empresarial puede abrir embebida sin caer en 403 sobre `/webmail/sso.php`.
- [QA] Validado con `go test ./handlers -run CorporateEmailAppendTheme -count=1` y prueba visual del panel empresarial.

## [2026-05-30] Menu super Portal publico e index
- [Super administrador] Nuevo grupo lateral `Portal publico e index` para concentrar edicion de tarjetas del index, modulos, descripcion de sistemas y WhatsApp del portal.
- [Navegacion] Se retiran esos accesos de `Gobierno` y `Comunicaciones` para que la administracion del portal publico quede en un solo lugar.
- [QA] Validacion de sintaxis de `web/js/super_administrador.js` y revision visual del menu.

## [2026-05-30] Menu super Mensajeria y alertas
- [Super administrador] El grupo lateral `Mensajeria y alertas` concentra alertas por correo, alertas de licencia, formatos de email, mensajes masivos, avisos de mantenimiento, Gmail SMTP y email corporativo.
- [Licencias] Los formatos usados en compra/pago de licencias quedan accesibles desde el mismo menu de mensajeria.
- [QA] Validacion de sintaxis de `web/js/super_administrador.js`, existencia de paginas e iconos.

## [2026-05-30] Menu super sin paginas huerfanas
- [Super administrador] `Asesores de ventas` pasa a ser el primer boton del grupo `Comercial y licencias`.
- [Navegacion] Se enlazan las paginas super que no tenian boton directo: reportes globales, auditoria global, metricas de trafico, preconfiguracion, contrato, administradores Frecuencia FE, chat IA global, voz IA, servidores, soporte remoto y configuracion avanzada.
- [QA] Validado cruce menu/paginas: 52 paginas HTML en `web/super`, 52 enlaces super en el menu y ningun enlace fuera de paginas permitidas.
- 2026-05-30: `Licencias globales compartidas` reduce el catalogo base a cuatro planes globales para todos los tipos de empresa; desde 2026-05-31 elimina planes heredados repetidos sin empresa asignada, mantiene la prueba gratis de 15 dias una sola vez por empresa y agrega una tarjeta visible de reglas en Super administrador > Licencias.
- 2026-05-30: `Configuracion general PostgreSQL` corrige el esquema y consultas de `empresa_configuracion_general` para usar `BIGSERIAL`, fecha compatible y placeholders traducidos, evitando 500 al entrar al panel de empresas nuevas.
- 2026-05-30: `Checkout Wompi de licencias` ajusta la consulta publica de terminos del comercio para no enviar cabecera `Authorization` al endpoint merchants de Wompi y desbloquear la prueba visual de planes comerciales.
## [2026-05-31] Encabezado visual compacto en logins
- [UX] `login.html` y `login_usuario.html` muestran su foto propia en la parte superior como encabezado compacto, tanto en escritorio como en celular.
- [Login] Se elimina el titulo textual `Powerful Control System` del login administrador y se conserva `Acceso de administradores` bajo la foto.
- [QA] Validacion visual local con navegador en escritorio y movil: imagen superior pequena, sin solapes y sin desborde horizontal.

## [2026-05-31] Pago idempotente y stock en cajas simultaneas
- [Backend] `PayCarritoStationSession` ahora cierra el carrito con una transicion atomica que exige carrito abierto, activo y sin `pagado_en`; los reintentos concurrentes no duplican documento, caja, metricas ni venta.
- [Inventario] Se documenta que productos y recetas reservan/descuentan stock al agregarse al carrito, para que varias cajas vean el inventario disponible en tiempo real antes del pago.
- [Seguridad] `pagar_estacion` responde de forma idempotente si el carrito ya quedo pagado por una solicitud anterior.

## [2026-05-31] Cajas fisicas con nombre y descripcion
- [Configuracion] `Configuracion de estaciones` permite administrar varias cajas fisicas por empresa con codigo, nombre, descripcion y estado activo.
- [Estaciones] La tarjeta Caja muestra las cajas configuradas, por ejemplo `CAJA-1 - FRUTERA`, y abre el corte con el codigo/nombre elegido.
- [Carrito/Corte] Los selectores y reportes muestran el nombre operativo de la caja cuando existe en `estaciones_config.cajas_config`.

## [2026-05-31] Email corporativo visible en celular
- [Panel] La bandeja de correo corporativo del panel empresarial conserva un alto util en movil y ya no se colapsa por `height:auto`.
- [UX] El iframe del webmail usa ancho completo, scroll tactil interno y evita desbordes horizontales en pantallas pequenas.

## [2026-05-31] Pagina Noticias
- [Portal] Se agrega `/noticias.html`, una pagina publica tipo red social con portada, foto de perfil, biografia y feed de publicaciones.
- [Super administrador] Se agrega `Noticias` en `Portal publico e index` para editar portada, perfil y publicaciones.
- [DIAN] El contenido predeterminado incluye una noticia sobre doctrina DIAN 2026 de facturacion electronica y otra sobre controles de facturacion, ambas con enlace a fuente oficial.

## [2026-05-31] Eliminacion de empresa con descarga previa
- [Selector] Antes de eliminar definitivamente una empresa se pregunta si desea descargar toda la informacion; si acepta, se abre la descarga y luego continua el borrado.
- [Editar empresa] El mismo paso previo se aplica desde `editar_empresa.html`, conservando la confirmacion por nombre, frase `ELIMINAR` y aceptacion de riesgo.
- [Seguridad] Se mantiene el endpoint destructivo actual y se registra en el payload si la descarga fue ofrecida.

## [2026-05-31] Cupo de cuentas de email por empresa
- [Super administrador] `Email corporativo Mailu` agrega el campo `Maximo cuentas por empresa`.
- [Backend] La configuracion global guarda `email_corporativo.max_accounts_per_empresa`, con default 5 por empresa, y normaliza el valor en servidor.
- [Seguridad] La creacion/sincronizacion de buzones valida el cupo por `empresa_id` antes de crear nuevas cuentas.
## [2026-06-01] Documentacion, APIs y preconfiguracion solar
- [Preconfiguracion] `tipo_empresa_preconfiguraciones.config_json` incluye `modulos.energia_solar` apagado por defecto con catalogo de proveedores, baterias y alertas.
- [Roles/Licencias] `tecnico_solar` queda como solo lectura; las licencias nuevas deben habilitar `energia_solar` como clave independiente y el fallback legacy queda documentado.
- [Ayuda/API] Se actualizan ayuda general, ayuda de APIs, OpenAPI generado, mapa de modulos, flujos, matriz de roles y estructura BD.

## [2026-06-01] GRAFOLOGIX visor de zoom
- [UX] El visor del manuscrito usa una tarjeta de altura estable y el canvas rellena todo el espacio disponible.
- [Grafologia] Al aplicar zoom, la imagen se recorta proporcionalmente desde el centro y ocupa toda la tarjeta sin franjas vacias.
- [QA] Validado con imagen manuscrita local: en 100% y 220% el canvas conserva el mismo tamano visual que la tarjeta.

## [2026-06-03] Nomina multi-sede y DIAN
- [Nomina] Empleados, liquidaciones y desprendibles conservan sede y centro de costo por empresa.
- [Motel Calipso] La demo profesional distribuye empleados entre sede principal, Rodadero y administracion para validar empresas con varias sedes.
- [DIAN] Se agregan consulta y preparacion de lote de documento soporte de pago de nomina electronica por empleado, listo para el flujo documental con firma/CUNE/numeracion.
- [QA] Pruebas Go de nomina/facturacion y validacion visual con datos simulados de Motel Calipso.

## [2026-06-05] Index - Documentos electronicos DIAN Colombia
- [Portal publico] La tarjeta `Documentos electronicos` del index muestra el bloque solicitado de documentos y eventos SFE para Colombia.
- [DIAN] Se listan factura electronica de venta, nota credito, nota debito, reporte de contingencia, documento soporte y nota de ajuste del documento soporte.
- [Super administrador] El contenido predeterminado del editor de `Informacion de modulos` se alinea con la nueva tarjeta y actualiza configuraciones antiguas que aun conservaban la lista generica.
