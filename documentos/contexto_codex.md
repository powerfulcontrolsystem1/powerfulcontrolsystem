# Contexto rapido para Codex

Este archivo es la primera lectura operativa antes de tocar el proyecto. Resume
lo que Codex debe tener en memoria para evitar redescubrir rutas, flujos y
decisiones en cada tarea.

## Actualizacion 2026-06-10 - Snapshot completo VPS

- El super administrador tiene una funcion en `Docker VPS` para crear snapshots
  restaurables en `backup/vps_snapshots` mediante `/super/api/vps_snapshots`.
- El snapshot incluye proyecto portable, dump PostgreSQL si esta disponible,
  volumenes Docker PCS, manifiesto y guia `RESTAURAR_VPS.md`; las imagenes
  Docker son opcionales porque pueden pesar mucho.
- La nube se maneja con `rclone` configurado fuera de PCS. El sistema solo guarda
  proveedor y ruta remota (`gdrive:...`, `mega:...`, `onedrive:...`, `s3:...`),
  no tokens OAuth ni claves.
- `.env.platform`, certificados, llaves y secretos no se incluyen por defecto.
  Cualquier restauracion real debe revisar el manifiesto y configurar secretos
  manualmente en la VPS destino.
- Para conectarse por SSH, compilar, publicar, consultar Docker/PostgreSQL y
  probar salud del VPS, leer `documentos/comandos_codex.md` secciones
  `Conexion SSH al VPS`, `Compilar y publicar en VPS` y
  `documentos/docker_vps_operacion.md`. La configuracion privada local vive en
  `scripts/pcs_deployment.local.ps1`; no imprimir host, llaves, DSN ni secretos.

## Actualizacion 2026-06-11 - Empresa de pruebas operativas

- Las pruebas visuales y operativas solicitadas por el usuario deben realizarse
  en la empresa propia `Powerful Control System` / `PCS` cuando el flujo requiera
  una empresa real de trabajo.
- No usar empresas de ejemplo como Alfaix salvo que el usuario lo pida
  explicitamente o que el caso de prueba corresponda a datos ya cargados alli.

## Actualizacion 2026-06-11 - Carrito busqueda por nombre con teclado

- El carrito permite buscar productos por nombre en un campo separado del codigo
  de barras/SKU.
- La lista de coincidencias se puede navegar con flecha arriba/abajo; `Enter`
  agrega el resultado seleccionado y, si solo existe una coincidencia, toma la
  primera sin exigir clic.
- El icono de pantalla completa de venta directa queda visualmente como solo
  icono, con `title` y `aria-label` para accesibilidad.

## Actualizacion 2026-06-11 - Produccion/MRP unificado con Productos

- `Inventario y productos` conserva catalogo, inventario, servicios y recetas
  vendibles del POS.
- `Produccion / MRP` queda como modulo profesional para BOM productiva, ordenes,
  consumos, calidad y plan de requerimientos de materiales.
- Desde MRP se pueden listar recetas vendibles activas e importarlas como BOM
  con codigo `POS-*`; si ya existen, se actualizan por codigo para evitar
  duplicados.
- Los endpoints genericos antiguos `/api/empresa/produccion/bom` se consideran
  legado tecnico; no deben ser el flujo operativo principal cuando exista
  `produccion_mrp`.

## Actualizacion 2026-06-10 - Nomina en Finanzas y cumplimiento

- `Nomina` queda como acceso directo del grupo principal `Finanzas y
  cumplimiento` y como entrada prioritaria del centro financiero.
- El nuevo hub `web/administrar_empresa/nomina_menu.html` contiene submenu
  propio: centro, empleados, configuracion legal, liquidaciones, pagos/PILA,
  nomina electronica DIAN y tutorial.
- Las subpaginas reutilizan `nomina_sueldos.html?seccion=...` para no duplicar
  reglas ni APIs; la vista antigua sin `seccion` conserva pantalla completa.
- El boton `Enviar nomina electronica a DIAN` prepara el lote, bloquea requisitos
  pendientes y llama el endpoint fiscal existente de facturacion electronica con
  `tipo_documento=nomina_electronica` y `empresa_id`.

## Actualizacion 2026-06-10 - IA flotante activa para todas las empresas

- En contexto empresarial, `/api/chat_flotante/preferencias?empresa_id=...`
  responde `chat_enabled=true` aunque existan preferencias antiguas apagadas.
- El arranque reaplica `chat_flotante.chat_enabled=1` a empresas existentes con
  la migracion `20260610_chat_ia_activo_empresas_reaplicar`.
- `web/js/ai_chat_drawer.js` fuerza el icono circular flotante como visible en
  empresas y no permite que `localStorage` viejo lo oculte; robot/secretaria
  siguen retirados y la voz se conserva como opcion.
- Si el backend local no tiene `OPENAI_API_KEY` en `backend/.env.local`, puede
  cargarla como fallback desde `deploy/.env.platform` sin imprimir el secreto.
  La credencial descifrable registrada en Super administrador sigue siendo la
  principal; la variable de entorno valida actua como respaldo cuando existan
  credenciales cifradas antiguas que no puedan leerse con la `CONFIG_ENC_KEY`
  actual.
- El icono flotante del chat usa un simbolo circular tipo OpenAI con las letras
  `IA`; paginas viejas con `gpt.svg` se normalizan desde `ai_chat_drawer.js`.

## Actualizacion 2026-06-09 - IA empresarial sin robot ni secretaria

- El asistente IA empresarial vuelve al recuadro normal: no debe mostrarse robot,
  secretaria ni avatar 3D en nuevas pantallas, configuraciones o ayudas.
- El icono flotante es un circulo compacto; la voz se conserva como opcion del
  chat y se persiste en `chat_flotante.voice_enabled`/`robot_voice` por
  compatibilidad historica del nombre de clave.
- En preproduccion, las empresas existentes y las preconfiguraciones por tipo
  quedan con `chat_flotante.chat_enabled=1`, `chat_flotante.robot_enabled=0` y
  `chat_flotante.personality_mode=normal`; la emisora sigue siendo opt-in.
- La IA no puede ejecutar SQL libre ni endpoints genericos de base de datos. El
  patron profesional es contexto de lectura construido por backend con
  `empresa_id`, tool/function calling con endpoints permitidos, confirmacion
  humana y validacion real de rol/licencia en backend.
- Segun documentacion oficial OpenAI 2026, la arquitectura recomendada para PCS
  es Responses API/function calling para herramientas de negocio; GPT-5.5 para
  razonamiento, vision y tareas complejas, y el modelo mini configurado para
  consultas frecuentes. La voz debe ser una capa reutilizable del mismo agente,
  con Realtime/WebRTC solo cuando se requiera baja latencia.

## Actualizacion 2026-06-08 - Herramientas locales Codex

- Chrome esta instalado y la Codex Chrome Extension esta habilitada en el perfil
  `Default`; Codex pudo listar pestañas reales de Chrome desde este hilo.
- Computer Use / Remote Control respondio correctamente en este hilo con
  `list_apps` y detecto Chrome como ventana controlable.
- Para contexto de archivos del proyecto no se requiere una extension adicional:
  Codex tiene acceso directo a `D:\powerfulcontrolsystem` y debe leer el
  repositorio con `rg`, documentacion y archivos locales.
- Usar Chrome cuando haga falta la sesion real del usuario en portales como PCS
  o DIAN. Usar Computer Use solo para aplicaciones Windows o flujos que no sean
  resolubles con APIs, navegador interno o comandos del proyecto.
- Mantener seguridad: no enviar formularios, instalar extensiones, aceptar
  permisos, ingresar credenciales o ejecutar acciones externas sin aprobacion
  explicita del usuario.

## Actualizacion 2026-06-08 - DIAN XAdES v2 y estado real

- Correccion posterior: el bloqueo `ZE02` se ataco ajustando la canonicalizacion
  XMLDSig/XAdES para calcular `KeyInfo`, `SignedProperties` y `SignedInfo` con
  el contexto completo de namespaces UBL heredado del documento. Un XML de prueba
  PCS paso verificacion externa con `lxml` para digest y `Node crypto` para
  `RSA-SHA256`.
- Verificacion real posterior al despliegue: `SETP990000195` obtuvo
  `TrackId 46208d27-216d-4c86-8a81-93dff2f1ee75`, `StatusCode=00`,
  `IsValid=true` y `Procesado Correctamente`; DIAN dejo de devolver `ZE02`.
- DIAN respondio despues que el `TestSetId db98ef26-0c2a-468f-a3d0-31667aba47e1`
  se encuentra aceptado, por lo que nuevos envios de ese set pueden devolver el
  codigo operacional `2` sin validar XML.
- Para factura/nota FEV, los ejemplos principales oficiales (`Generica.xml`,
  `Consumidor Final.xml`, `CreditNote.xml`, `DebitNote.xml`) usan la politica
  `https://facturaelectronica.dian.gov.co/politicadefirma/v1/politicadefirmav2.pdf`.
- La firma XAdES base debe seguir el patron de los XML oficiales de la Caja de
  herramientas DIAN FE V19 V2026: `xades:Description`, namespace `xades141` y
  `SignedDataObjectProperties/DataObjectFormat` apuntando a la referencia del
  documento XML.
- La reconsulta real de `GetStatusZip` para Powerful Control System deja como
  historico que `SETP990000186` a `SETP990000194` estuvieron rechazados por
  `StatusCode=99` con `ZE02`, pero el diagnostico corregido `SETP990000195`
  cerro la brecha de firma con aceptacion DIAN `00`.

## Actualizacion 2026-06-16 - Pruebas reales DIAN produccion PCS

- Para probar la facturacion electronica de Powerful Control System no iniciar
  por localhost: usar `https://powerfulcontrolsystem.com` y la empresa PCS
  (`empresa_id=12`) con sesion real o login API autorizado por el usuario.
- El flujo rapido reproducible es: autenticar, verificar configuracion DIAN,
  crear venta/factura de una `menta`, emitir o reenviar por
  `/api/empresa/facturacion_electronica?action=emitir|reenviar_dian&empresa_id=12`
  y revisar `integracion_fiscal`, `cola_reintentos` y acuse DIAN saneado.
- No imprimir contrasenas, PIN, certificado, llave tecnica ni token. Se puede
  mostrar numero legal, prefijo, resolucion, ambiente, codigo de regla DIAN y
  estado `aceptado/rechazado/contingencia`.
- La resolucion vigente de PCS registrada el 2026-06-16 usa prefijo `PCS`,
  rango `1-1000000` y ambiente produccion. DIAN rechazo las pruebas `PCS1/PCS2`
  despues de llegar a `SendBillSync` por `FAB05c` (Software ID no corresponde
  al rango de numeracion) y `FAD06` (CUFE no calculado correctamente); ya no
  hay rechazo por guion/rango tras normalizar el numero legal a `PCS1`, `PCS2`.
- Actualizacion 2026-06-17: la nueva Autorizacion de Numeracion DIAN Formulario
  1876 `18764111318575` para PCS usa prefijo `1PCS`, rango `1-100000`, fecha
  inicial `2026-06-17`, vigencia 24 meses y fecha final `2028-06-17`. El PDF de
  prueba local del usuario fue `C:\Users\ivanm\Documents\18764111318575 Autorizacion numercion DIAN 17 JUNIO 2026.pdf`.
- La prueba real autorizada del 2026-06-17 creo venta de una `menta` para el
  cliente IVAN FRANCISCO CAYON GUARNIZO y genero factura `FV-FE-MENTA-20260617151719`
  con numero legal `1PCS1`. La factura llego a DIAN y fue rechazada por
  `FAK61`, `FAB05c` y `FAD06`; ya no se observo rechazo por rango/prefijo
  `FAD05a/FAD05b/FAD05c`.
- Cliente PCS para pruebas de facturacion autorizadas por el usuario: IVAN
  FRANCISCO CAYON GUARNIZO, CC `84456779`, persona natural, Colombia,
  Magdalena, Santa Marta, direccion `calle 28 N5-116`, celular `3043306506`.
  En PCS existe como cliente `id=22`.
- Si aparece `FAB05c`, revisar primero en el portal DIAN que la resolucion/rango
  de numeracion este asociado al mismo Software ID configurado en PCS. No
  sustituirlo por el TestSetId de habilitacion ni por valores copiados en
  documentacion.
- Despues de solicitar la numeracion y obtener el Formulario 1876, es obligatorio
  entrar al portal DIAN de produccion `https://catalogo-vpfe.dian.gov.co/User/Login`
  y asociar el prefijo/rango autorizado con el Software ID. Si este paso falta,
  DIAN rechaza con `FAB05c` aunque el PDF de numeracion este cargado en PCS.
- La aceptacion de habilitacion historica `SETP990000195` demuestra que el
  transporte SOAP y la firma XAdES base pueden llegar a acuse `StatusCode=00`;
  produccion exige ademas que el rango autorizado quede vinculado al software
  correcto y que el CUFE use la hora Colombia `-05:00`.
- Nunca guardar claves de usuario, software PIN, certificados, token, llave
  tecnica ni contrasenas en documentacion. Si un chat nuevo necesita login, debe
  pedir autorizacion al usuario o reutilizar sesion activa/cookie temporal.

## Actualizacion 2026-06-16 - Guia Codex para navegador, VPS y carrito PCS

- En chats nuevos, Codex debe leer primero este archivo y
  `documentos/comandos_codex.md` antes de tocar carrito, ventas, caja,
  facturacion, permisos o despliegue.
- Para pruebas visuales usar el navegador interno de Codex cuando este
  disponible. El flujo es abrir o reutilizar una pestana, navegar a
  `https://powerfulcontrolsystem.com`, conservar la sesion existente y validar
  con interacciones reales de mouse/teclado mas lecturas puntuales del DOM.
- Para responsive movil del carrito usar el control de viewport del navegador
  interno o una ventana equivalente de celular, revisar que no exista scroll
  horizontal y que buscador, cliente, detalle de productos, pagos, acciones y
  totales queden en una sola columna usable.
- Para pruebas de produccion solicitadas por el usuario, usar siempre PCS
  (`empresa_id=12`) salvo instruccion contraria. No probar primero en localhost
  cuando el reporte es de `powerfulcontrolsystem.com`.
- El flujo minimo del carrito PCS es: abrir venta directa, buscar producto por
  nombre con mouse/teclado, seleccionar o usar el primer resultado visible,
  presionar `Agregar`, comprobar que aparece o aumenta la linea del producto,
  cambiar cantidad con los botones `+/-`, revisar totales, probar medios de
  pago visibles sin cerrar ventas reales salvo autorizacion explicita y revisar
  que no haya errores de consola.
- Si el usuario pide `rs`, ejecutar `.\rs.ps1` desde la raiz. Ese wrapper hace
  preflight, actualiza repositorio cuando aplica, sincroniza con VPS, reconstruye
  servicios Docker y verifica salud publica. No imprimir secretos ni rutas
  privadas sensibles.
- Despues de publicar con `rs`, repetir la prueba visual contra
  `https://powerfulcontrolsystem.com/...&empresa_id=12&qa={timestamp}` para
  evitar cache viejo.

## Actualizacion 2026-06-08 - UBL DIAN realista y errores completos

- `generateDIANUBLBase` ya no genera XML UBL minimo/inventado: ahora emite
  factura, nota credito y nota debito con `UBL 2.1`, `ProfileID` especifico
  por tipo, `CustomizationID` oficial por tipo, `CUFE/CUDE-SHA384`,
  `DianExtensions`, `SoftwareSecurityCode`, QR DIAN, parties, `PaymentMeans`,
  totales, `InvoiceLine`, `CreditNoteLine` o `DebitNoteLine` segun corresponda.
- La prueba real de DIAN 2026-06-08 confirmo rechazos por `PrePaidAmount`
  mal capitalizado, literales DIAN sin tildes, falta de `PaymentMeans`, perfil
  incompleto, CUDE/firma y notas referenciando factura no aceptada.
- El set automatico ya no usa una factura apenas generada como referencia de
  notas: primero debe existir factura aceptada por DIAN (`StatusCode=00`);
  sin esa referencia aceptada, las notas quedan bloqueadas antes de enviarse.
- El preflight DIAN bloquea estructuras equivocadas antes de enviar: notas sin
  `DiscrepancyResponse/BillingReference`, UUID sin esquema SHA384, lineas de
  tipo incorrecto, falta de extensiones DIAN o falta de `SoftwareSecurityCode`.
- La firma XAdES base incluye la referencia a `KeyInfo` y firma exactamente el
  `SignedInfo` embebido; si DIAN sigue devolviendo `ZE02`, queda pendiente
  sustituir la firma base por un firmador XMLDSig/XAdES con canonicalizacion
  completa validado por la caja DIAN.
- `GetStatusZip` y respuestas SOAP ahora parsean `ErrorMessageList` completo;
  si DIAN devuelve varias reglas de rechazo, PCS las conserva saneadas como
  lista visible y no las reduce a un estado generico.
- Las referencias oficiales DIAN descargadas localmente viven en
  `documentos/referencias/dian/2026-06-08/` y no se versionan por tamano/binarios.
  El README versionado deja las URLs/artefactos; `scripts/validar_dian_xsd.ps1`
  valida un XML contra los XSD oficiales cuando la caja esta descargada.
- La aceptacion fiscal sigue dependiendo solo del acuse DIAN final
  `IsValid=true`/`StatusCode=00`; `Batch en proceso de validacion` nunca cuenta
  como aceptado.

## Actualizacion 2026-06-08 - Historial TrackId DIAN

- `facturacion_electronica_pruebas_dian.html` muestra una tarjeta `Historial
  TrackId / ZipKey DIAN` para consultar identificadores reales por empresa.
- El backend guarda cada TrackId/ZipKey en `empresa_dian_track_historial` y lo
  actualiza al consultar `GetStatusZip`; no guarda XML completo, claves,
  certificados, PIN ni tokens.
- `GetStatusZip` ahora toma `StatusDescription` del SOAP DIAN para distinguir
  `Batch en proceso de validacion` como acuse pendiente, no aceptacion final.

## Actualizacion 2026-06-08 - DIAN sin preset reducido

- `facturacion_electronica_pruebas_dian.html` no debe mostrar boton ni preset
  rapido 2+2+2/pequeno para habilitacion DIAN.
- El Centro de habilitacion DIAN conserva el objetivo del portal
  `30 + 10 + 10`, el historico excepcional `60 + 20 + 20` y `Personalizado`.
- Si llega un valor viejo de preset reducido, la pantalla lo normaliza al
  objetivo real del portal para evitar envios parciales accidentales.

## Actualizacion 2026-06-08 - Centro de habilitacion DIAN

- El acceso operativo antes llamado `Pasar test DIAN` ahora se presenta como
  `Centro de habilitación DIAN`.
- `facturacion_electronica_pruebas_dian.html` sigue siendo la pagina de
  validacion, objetivo del set, envios manuales y set automatico real.
- Las respuestas de `diagnostico_oficial` se muestran como resumen operativo
  legible; el JSON tecnico queda plegado en `Ver detalle técnico` y saneado para
  no exponer secretos.

## Actualizacion 2026-06-08 - Otros paises en facturacion electronica

- `web/administrar_empresa/facturacion_electronica_menu.html` mantiene a
  Colombia/DIAN como flujo principal visible del submenu de facturacion.
- El submenu usa un selector `Pais fiscal` alimentado por
  `/api/empresa/facturacion_electronica/paises_disponibles`; Colombia carga la
  pagina DIAN actual y Ecuador/Panama cargan sus paginas fiscales existentes
  desde la misma seleccion, sin enlaces duplicados.
- El selector conserva `empresa_id` en el iframe y no sustituye validaciones de
  backend, permisos ni licencia.

## Actualizacion 2026-06-07 - Bodega base por empresa

- Cada empresa nueva queda preconfigurada con una bodega activa llamada
  `Bodega 1`.
- El backend usa `EnsureEmpresaBodega1` para crearla de forma idempotente por
  `empresa_id`: si ya existe, la reactiva; si falta, la crea sin productos,
  existencias, movimientos ni stock simulado.
- El arranque aplica la migracion `20260607_bodega_1_default` para empresas
  existentes de preproduccion, independiente de las migraciones anteriores de
  impuestos y nomina.

## Actualizacion 2026-06-08 - Ayuda contextual en formularios

- `web/js/form_field_help.js` expone `window.PCSFieldHelp.install(map)` para
  agregar botones `?` junto a labels de formularios sin dependencias externas.
- `administrar_productos.html` usa el helper en el formulario `Nuevo producto`
  para explicar los 25 campos operativos antes de guardar.
- `facturacion_electronica.html` usa el mismo patron en firma, DIAN Colombia,
  configuracion por pais y configuracion avanzada de facturacion/impresion.
- Los textos de DIAN deben seguir sin secretos: no imprimir PIN, claves,
  certificados, tokens ni valores reales sensibles.

## Actualizacion 2026-06-08 - Facturacion electronica inicia en DIAN

- `web/administrar_empresa/facturacion_electronica.html` inicia con un selector
  visible de pais detectado/seleccionado y luego conserva `Configuracion DIAN
  Colombia` y `Cargar firma electronica (Colombia / DIAN)`.
- La deteccion de pais carga Colombia/DIAN sin romper el flujo existente; si el
  usuario elige Ecuador o Panama se redirige a la pagina fiscal especifica de
  ese pais conservando `empresa_id`.
- La tarjeta antigua `Perfil de facturacion` sigue fuera de la vista para no
  duplicar informacion.

## Actualizacion 2026-06-07 - Ayuda integrada con caja IA

- `web/js/help_ai_bridge.js` conecta enlaces de ayuda, tutoriales y botones con
  `data-pcs-help-ai` a la caja IA del panel empresarial.
- `web/js/ai_chat_drawer.js` expone `PCSAIChatHelp` y acepta el mensaje
  `pcs-help-ai-open` para mostrar ayuda contextual estatica sin ejecutar
  acciones ni consumir IA automaticamente.
- La ayuda se abre en el drawer normal en modo `Ayudante por pasos`; si la
  IA/chat estan apagados, se abre un panel de ayuda estatica sin activar IA ni
  cambiar preferencias y con enlace a la guia HTML completa.
- `web/js/empresa_submenu_context.js` carga el puente de ayuda en subpaginas
  empresariales para que los iframes puedan pedir ayuda al panel padre sin
  romper `empresa_id`.
- El panel padre tambien inyecta `help_ai_bridge.js` en iframes empresariales
  same-origin y anidados, evitando que una subpagina con cache viejo pierda el
  enlace con la caja de ayuda.

## Actualizacion 2026-06-07 - Barra de configuracion DIAN

- `web/administrar_empresa/facturacion_electronica.html` muestra una barra
  `Avance de configuracion DIAN` dentro de `Configuracion DIAN Colombia`.
- El avance se calcula en frontend con los campos ya cargados: identidad fiscal,
  ambiente/TestSetId, software, numeracion, llave tecnica y firma/certificado.
- La pantalla lista campos faltantes por nombre, sin mostrar valores sensibles,
  y no crea endpoint, tabla ni persistencia nueva.

## Actualizacion 2026-06-07 - Checks de impresion por empresa

- `web/administrar_empresa/configuracion_impresora.html` agrega dos matrices de
  checks por empresa: campos del recibo de venta y campos de corte/cierre de
  turno.
- La persistencia usa `/api/empresa/configuracion_avanzada` y los campos
  `empresa_configuracion_avanzada.impresion_recibo_items_json` e
  `impresion_corte_items_json`, validados como JSON de booleanos.
- `carrito_de_compras.html` aplica esos checks solo al recibo operativo de
  venta. La factura electronica DIAN no cambia sus campos legales ni su XML.
- `corte_de_caja.html` y `reportes_turnos.html` aplican los checks a encabezado
  y detalle impreso de reportes operativos por empresa.

## Actualizacion 2026-06-07 - Modulo NIIF

- `web/administrar_empresa/niif.html` agrega un centro NIIF por empresa dentro
  de Finanzas y Suite contador.
- La pagina lee, cuando hay permiso, el dashboard real de
  `/api/empresa/contabilidad_colombia?action=dashboard` y no crea endpoint ni
  tablas nuevas.
- Incluye diagnostico de adopcion, politicas contables, calculos de deterioro,
  depreciacion, valor razonable, conciliacion contable-fiscal, checklist de
  cierre y notas exportables a JSON/TXT.
- `linkNIIF` queda bajo `finanzas:R`; el rol `contador` puede verlo sin ganar
  permisos de escritura, aprobacion, caja, ventas ni configuracion.

## Actualizacion 2026-06-07 - IA oculta por defecto por empresa

- Las paginas y accesos IA empresariales (`linkChatIA`,
  `linkCentroIAEmpresarial`, `linkRentaIA`, `linkSoportesComprasIA` y
  `linkSoportesComprasIAMenu`) quedan ocultos por defecto para todas las
  empresas.
- El backend solo los devuelve visibles en `/api/empresa/permisos_contexto`
  cuando la empresa tiene una regla fina explicita de pagina permitida y el rol,
  licencia y wrapper base ya lo permiten.
- El menu principal, el chat flotante, el Centro financiero y la Suite contador
  no muestran IA mientras no exista esa habilitacion explicita por empresa.

## Actualizacion 2026-06-07 - Guardado de configuracion FE por pais

- `web/administrar_empresa/facturacion_electronica.html` guarda el perfil de
  pais con `POST /api/empresa/facturacion_electronica?action=config_pais`.
- `backend/handlers/facturacion_electronica.go` reconoce `config_pais`,
  `guardar_config_pais` y `configuracion_pais` como acciones de configuracion,
  separadas de las acciones fiscales que exigen `documento_codigo`.
- El flujo evita que el boton `Guardar configuracion pais` se confunda con
  emision/anulacion de documentos electronicos y mantiene aislamiento por
  `empresa_id`.

## Actualizacion 2026-06-07 - Boton Enviar nota credito DIAN

- `web/administrar_empresa/facturacion_electronica_pruebas_dian.html` reemplaza
  la confirmacion nativa del navegador por un modal integrado para envios reales
  de pruebas DIAN.
- El boton `Enviar nota credito` conserva el payload real
  `notas_credito=1`, `max_envios=1` y el endpoint
  `/api/empresa/facturacion_electronica/dian?action=pruebas_dian`.
- La reparacion evita que el click quede bloqueado o cancelado por el popup
  nativo antes de llegar al envio real.

## Actualizacion 2026-06-07 - Retiro del boton 2+2+2 DIAN

- `web/administrar_empresa/facturacion_electronica_pruebas_dian.html` ya no
  muestra el boton rapido `Enviar prueba 2 + 2 + 2`.
- La pagina conserva `Ejecutar set automatico` para el objetivo real guardado y
  los envios individuales `Enviar factura`, `Enviar nota debito` y
  `Enviar nota credito`.
- El retiro evita pruebas parciales accidentales y deja la operacion alineada
  con el set asignado por DIAN para cada `empresa_id`.

## Actualizacion 2026-06-07 - Suite contador

- `web/administrar_empresa/suite_contador.html` centraliza el trabajo del
  contador por empresa: portal contador, contabilidad Colombia, suite avanzada,
  impuestos, DIAN, declaraciones, certificados, cierres, activos, reportes,
  nomina, bancos, compras IA y Renta IA.
- La pagina no crea endpoint ni tablas; consulta
  `/api/empresa/permisos_contexto?empresa_id={id}` y muestra cada modulo como
  disponible o bloqueado segun permisos/licencia.
- `linkSuiteContador` queda registrado como `finanzas:R`. El rol `contador`
  puede ver la suite y accesos contables clave, pero cada operacion real sigue
  protegida por su wrapper y permiso propio.
- La portada publica `web/index.html` ahora presenta `Suite contador 360` dentro
  de Finanzas y cumplimiento.

## Actualizacion 2026-06-06 - Renta IA financiera

- `web/administrar_empresa/renta_ia.html` vive dentro del centro financiero y
  usa `/api/empresa/finanzas/renta_ia?action=renta_ia`.
- El backend calcula una estimacion gerencial de renta con datos reales por
  `empresa_id`: ventas cerradas, movimientos financieros, compras de inventario
  y nomina liquidada. No guarda declaracion ni crea tablas.
- La IA empresarial solo interpreta el JSON ya calculado, registra consumo por
  empresa y debe mantener la advertencia de que no reemplaza formulario oficial
  ni revision del contador.
- `linkRentaIA` usa `finanzas:R`, por lo que el rol `contador` puede verlo sin
  permisos de creacion ni aprobacion.

## Actualizacion 2026-06-06 - DIAN SOAP real y acuse pendiente

- El transporte oficial DIAN para habilitacion usa SOAP/WCF con
  `SendTestSetAsync` y `GetStatusZip`. El sobre vigente firma `wsa:To`,
  referencia el `BinarySecurityToken` con `wsse:Reference URI="#X509-..."` e
  incluye `InclusiveNamespaces`; no se debe volver a `ThumbprintSHA1` para este
  flujo.
- Los envios reales de habilitacion para la empresa interna Powerful Control
  System ya reciben HTTP 200, TrackId/ZipKey y respuesta
  `Batch en proceso de validacion`; el bloqueo de transporte `InvalidSecurity`
  queda superado.
- Lo pendiente para cerrar habilitacion no es simular ni rehacer el transporte:
  falta reconciliar cada TrackId con `GetStatusZip` hasta acuse final
  aceptado/rechazado, persistir el estado final por documento/lote, resumirlo en
  la pantalla y habilitar produccion local solo cuando se cumplan los minimos
  aceptados configurados por empresa.
- No documentar PIN, claves tecnicas, certificados, contrasenas ni tokens. El
  TestSetId y demas datos operativos deben leerse de `empresa_dian_configuracion`
  por `empresa_id`, no copiarse como secretos en guias o logs.

## Actualizacion 2026-06-05 - Set automatico DIAN y licencias

- `web/administrar_empresa/facturacion_electronica_pruebas_dian.html` permite
  guardar por empresa el objetivo exacto que muestra el portal DIAN:
  `test_set_id`, facturas, notas debito, notas credito, total requerido y
  minimos aceptados.
- El preset principal de software propio/proveedor queda en 30 facturas, 10
  notas debito y 10 notas credito, con minimo aceptado total 1 y minimo de
  facturas 1, porque ese es el set registrado para la empresa interna Powerful
  Control System. El boton automatico usa siempre lo guardado por empresa, por
  lo que otra empresa puede tener otro objetivo.
- En endpoint oficial SOAP/WCF DIAN no se exige `token_emisor_ref`; ese campo
  solo aplica cuando la empresa usa proveedor/API con bearer token. En
  habilitacion real si es obligatorio `test_set_id`.
- `action=pruebas_dian` y `action=enviar_set_pruebas` no aceptan simulacion:
  deben enviar documentos reales al ambiente de habilitacion, recibir `ZipKey`
  cuando aplique y cerrar el acuse con `GetStatusZip`. Solo una ejecucion real
  con acuse aceptado de DIAN/proveedor puede permitir pasar a produccion local.
- Compra de licencias: el correo de bienvenida puede adjuntar la factura
  electronica en PDF cuando el pago comercial aprobado es mayor que cero. La
  licencia del software ya no se adjunta por correo; se descarga desde
  Administrar empresa > Licencia > Licencia del sistema. Una activacion valor
  cero o con descuento total activa la licencia, pero no emite factura
  electronica.

## Actualizacion 2026-06-05 - Tutorial guiado de nomina

- Ayuda visual: `web/ayuda/tutorial_nomina.html`.
- Acceso desde nomina: `web/administrar_empresa/nomina_sueldos.html` muestra
  un boton `Ayuda` que conserva `empresa_id` y abre la presentacion guiada.
- La presentacion tiene diapositivas operativas y cada bloque `Narracion`
  incluye un boton con icono de play para reproducir la guia por voz cuando el
  navegador soporte sintesis de voz.

## Actualizacion 2026-06-03 - Nomina multi-sede y DIAN

- Modulo: `web/administrar_empresa/nomina_sueldos.html`.
- API: `/api/empresa/nomina`, protegida por permisos empresariales de nomina.
- Empleados y liquidaciones guardan `sede_codigo`, `sede_nombre` y
  `centro_costo`; las liquidaciones conservan esos datos historicos.
- `seed_motel_calipso` crea demo profesional con empleados distribuidos en sede
  principal, Rodadero y administracion.
- Acciones DIAN de nomina: `documentos_electronicos_colombia` consulta estado y
  `preparar_nomina_electronica` arma lote por empleado para documento soporte de
  pago de nomina electronica. El envio real requiere credenciales, firma, CUNE,
  numeracion y transporte documental en facturacion electronica por empresa.

## Actualizacion 2026-06-04 - Camaras y DVR

- Nuevo modulo empresarial `camaras` visible como `Administrar empresa >
  Analisis y control > Camaras`.
- API: `/api/empresa/camaras`, protegida por
  `WithEmpresaCamarasPermissions`.
- Backend: `backend/db/camaras.go` y `backend/handlers/camaras.go`.
- BD: `empresa_camaras` guarda camaras por `empresa_id`, DVR/NVR, canal,
  protocolo origen, visor web, estacion asociada, orden, estado y referencias
  seguras de usuario/clave.
- UI: `web/administrar_empresa/camaras.html` y `web/js/camaras.js`.
- Estaciones: `estaciones_config` acepta `camaras_enabled`,
  `camaras_placement`, `tipo_estacion=camara` y `camara_id`; una estacion puede
  mostrarse como visor de camara y las camaras globales pueden cargar antes o
  despues de las estaciones.
- Nota tecnica: RTSP/ONVIF directo no se reproduce en navegador; debe pasar por
  gateway HLS, WebRTC, MJPEG o iframe confiable.

## Actualizacion 2026-06-05 - Empresa interna Powerful Control System

- La empresa interna `Powerful Control System` opera como cualquier empresa del
  SaaS: panel, carrito, correo corporativo, facturacion, reportes y modulos
  empresariales pasan por los mismos endpoints y permisos multiempresa.
- La licencia tecnica interna `PCS_SYSTEM_INTERNAL_PERPETUAL` queda retirada:
  no se debe crear, reactivar ni usar para otorgar acceso. Si existe en datos
  heredados, el arranque la desactiva y la empresa debe comprar, renovar o
  adquirir licencias comerciales como las demas empresas.
- No usar excepciones sin fecha para esta empresa; las consultas de licencia
  deben resolver siempre una licencia comercial vigente normal.
- El rol `super_administrador` validado en backend puede acceder globalmente a
  empresas para soporte, auditoria, comparticion y operacion interna, manteniendo
  filtros por `empresa_id` en las consultas.

## Actualizacion 2026-06-05 - Carpetas empresariales y firma electronica

- Cada empresa debe tener carpeta propia bajo
  `web/uploads/empresas/empresa_{id}_{slug}/`.
- Al crear una empresa se preparan carpetas base de uploads e imagenes; si la
  creacion fue idempotente, se asegura la misma carpeta sin duplicar empresa.
- La firma electronica para DIAN se guarda en la subcarpeta privada
  `facturacion_electronica/firma_electronica/` dentro de la carpeta de la
  empresa. Los archivos se guardan con permisos `0600`, la carpeta con `0700` y
  se referencian internamente con `file:`; no deben exponerse como URL publica.
- Al eliminar una empresa, el limpiador de archivos incluye tambien
  `web/uploads/empresas/empresa_{id}_*` para no dejar firmas, imagenes o
  adjuntos huérfanos.

- Al cargar una firma DIAN, el backend lee el certificado X.509 y guarda
  `certificado_vencimiento`, `certificado_vencimiento_en`,
  `certificado_alerta_dias`, `certificado_alerta_ultimo_envio` y
  `certificado_alerta_email` en `empresa_dian_configuracion`.
- La ultima carga de firma conserva metadatos seguros
  (`certificado_ultima_carga_en`, archivo original, formato, titular, emisor,
  serial y estado de clave). La clave del P12/PFX nunca debe guardarse ni
  mostrarse en claro; solo se usa para decodificar el archivo.
- El endpoint
  `/api/empresa/facturacion_electronica/dian?action=vencimiento_certificado`
  calcula dias restantes, estado vigente/proximo/vencido y envia aviso por
  correo al administrador de la empresa cuando faltan 30 dias o menos, con
  control de no repetir alertas dentro de 24 horas.

## Actualizacion 2026-06-01 - GRAFOLOGIX

- Nuevo modulo empresarial `grafologia` visible como `Administrar empresa >
  Analisis y control > GRAFOLOGIX`.
- API: `/api/empresa/grafologia`, protegida por
  `WithEmpresaGrafologiaPermissions`.
- Backend: `backend/internal/grafologia` contiene el motor Go puro; no usa
  dependencias externas.
- Docker/VPS: desde 2026-06-18 no se instala Tesseract/pdftoppm para OCR; el
  analisis complementario debe usar IA GPT-5.5 y transcripcion manual opcional.
- BD: `empresa_grafologia_analisis` en `pcs_empresas`.
- UI: `web/administrar_empresa/grafologia.html` y `web/js/grafologia.js`.
- La pantalla asocia cada manuscrito a un cliente central de la empresa:
  busca/crea desde `/api/empresa/clientes`, guarda `cliente_id` validado por
  `empresa_id` y conserva descripcion/caracteristicas de la persona en el
  informe.
- La pantalla tambien ofrece `Analizar con GPT-5.5`, que reutiliza el catalogo
  de Chat IA empresarial (`openai:gpt-5.5`), valida limites diarios por empresa,
  envia la imagen como adjunto de vision y registra la consulta en la auditoria
  de uso IA sin crear dependencias nuevas.
- Documento completo: `documentos/grafologix_arquitectura.md`.
- Advertencia permanente: las interpretaciones grafológicas son heuristicas y
  orientativas; no son diagnostico psicologico ni criterio automatico de
  contratacion.

## Resumen del sistema

Powerful Control System es un POS/ERP SaaS multiempresa. El backend esta escrito
en Go, la persistencia oficial es PostgreSQL y el frontend usa HTML, CSS y
JavaScript tradicional servido como archivos estaticos. El sistema cubre login
administrativo, creacion de empresas, licencias, carritos, estaciones, caja,
inventario, compras, creditos, facturacion electronica por pais, reportes,
offline, super administrador y despliegue Docker/VPS.

No se deben agregar dependencias externas ni cambiar `go.mod` sin autorizacion
explicita. No se deben documentar secretos, claves, tokens ni contrasenas.

## Arranque y estructura base

- Raiz del repo: `D:\powerfulcontrolsystem`.
- Backend Go: `backend`.
- Frontend estatico: `web`.
- Paginas empresariales: `web/administrar_empresa.html` y
  `web/administrar_empresa/`.
- Paginas super administrador: `web/super_administrador.html` y `web/super/`.
- Documentacion operativa: `documentos/`.
- Scripts de operacion: `rs.ps1`, `scripts/rs.ps1`,
  `scripts/sync_to_vps.ps1`, `scripts/profesional_preflight.ps1`.
- Docker/VPS: `deploy/`, `docker-compose*.yml`, `documentos/docker_vps_operacion.md`.

El servidor registra rutas principales en `backend/main.go`. El frontend se
sirve desde `web` y algunas paginas reciben inyecciones globales, como iconos de
botones, desde handlers estaticos del backend.

## Bases de datos

- `pcs_empresas`: datos operativos por empresa.
- `pcs_superadministrador`: configuracion global, licencias, portal publico,
  alertas super administrador y catalogos globales.
- Todo cambio multiempresa debe filtrar y validar `empresa_id` en backend. Nunca
  se debe confiar solamente en URL, localStorage, cache o datos enviados por el
  navegador.
- La configuracion empresarial flexible se guarda con frecuencia en
  `empresa_estacion_prefs`, especialmente `estaciones_config` y
  `carrito_ui_global`.

## Rutas y endpoints de referencia

- Login y registro admin: `web/login.html`,
  `web/registrar_nuevo_usuario_administrador.html`,
  `backend/handlers/auth_admin_handlers.go`.
- Seleccion y creacion de empresas: `web/seleccionar_empresa.html`,
  `web/js/seleccionar_empresa.js`, `/super/api/empresas`.
- Administradores delegados: `web/super/administradores.html`,
  `/super/api/administradores`, `backend/handlers/auth_admin_handlers.go`.
  Si el correo no existe se registra por invitacion con token; si ya existe y
  esta confirmado se usa `admin_principal_delegaciones` para que vea sus empresas
  propias mas las empresas compartidas, sin cambiar `usuario_creador`.
  Desde el panel global de super administrador, intentar agregar un correo ya
  confirmado no debe fallar con 409: responde OK, evita duplicar la cuenta y solo
  actualiza nombre/rol cuando el actor tiene permisos super suficientes.
  El selector de empresas debe resolver alcance efectivo por cuatro caminos:
  propietario por `usuario_creador`, delegado del principal, empresa compartida
  con el administrador y empresa que el administrador compartio con otro usuario;
  esta ultima conserva `access_source=owner` para que el propietario no pierda la
  empresa despues de compartirla.
  Desde `seleccionar_empresa.html` siempre se abre con `scope=principal` y solo
  debe mostrar invitados del administrador autenticado; sin ese parametro el
  panel super mantiene la vista global. Los nuevos `super_administrador` tambien
  se crean por invitacion con token.
- Auditoria global del selector: `web/super/auditoria_global.html`,
  `/super/api/auditoria`, `backend/handlers/auditoria_super.go`,
  `pcs_superadministrador.super_auditoria_eventos`. Desde el selector se abre
  con `scope=principal`: un administrador normal ve solo su alcance y el super
  administrador puede ver global desde el panel super.
- Auditoria especial super administrador: `web/super/auditoria_super_admin.html`
  en `web/super_administrador.html > Acceso`. Usa
  `/super/api/auditoria?scope=super_panel`, reservado a roles super, para revisar
  navegacion, botones `Editar`, guardados/pruebas y endpoints sensibles del panel
  super. Nunca se deben guardar secretos en metadata.
- Licencias: `web/elegir_licencia.html`, `web/pagar_licencia.html`,
  `web/super/licencias.html`, `web/super/formato_para_emviar_email.html`,
  `web/super/licencias_codigos_descuento.html`,
  `web/administrar_empresa/licencia_sistema.html`, `/super/api/licencias`,
  `/super/api/licencias/codigos_descuento`, `/licencias/activar_sin_pago` y
  `/api/empresa/licencia_sistema/pdf`.
  El catalogo base vigente es global para todos
  los tipos de empresa (`tipo_id=0`, `pais_codigo=GLOBAL`) con ocho planes:
  prueba gratis 15 dias, prueba pagada de 1 dia por COP 1000, mensuales
  COP 60000, COP 110000 y COP 200000, y anuales COP 600000, COP 1100000 y
  COP 2200000. La
  prueba gratis solo se puede activar una vez por empresa, incluso cuando la
  prueba anterior ya vencio, quedo inactiva o viene de datos antiguos; cada
  plan global puede ocultarse o mostrarse desde Super administrador > Licencias
  sin que el bootstrap vuelva a activarlo automaticamente; las
  licencias base antiguas por tipo y addons de catalogo sin empresa asignada se
  eliminan del catalogo comercial. La renovacion adelantada de la misma licencia
  se limita con `licencias.max_compras_adelantadas_misma_licencia`, por defecto
  2 compras adicionales. Al activarse una licencia por pago o por flujo de valor
  cero permitido, `backend/handlers/payments_handlers.go` puede enviar un correo
  de bienvenida configurable desde Super administrador > Licencias y Formatos de
  email. El PDF de licencia de software no se adjunta al correo; se descarga
  desde Administrar empresa > Licencia > Licencia del sistema y su texto se
  edita con la plantilla `licencia_software_pdf`.
  En compras comerciales con total pagado mayor que cero, el mismo flujo emite
  una factura electronica automatica desde la empresa interna `Powerful Control
  System`/`Powerful Control Systen` ya existente y, si el switch esta activo,
  adjunta el PDF resumen de esa factura al correo de bienvenida. El documento se registra en
  `empresa_facturacion_documentos` de la empresa emisora y marca `pagos_epayco` o `pagos_wompi` con
  `licencia_factura_electronica_emitida` para no duplicar documentos por
  webhooks repetidos. La empresa emisora interna se guarda en
  `configuraciones.licencias.facturacion_empresa_sistema_id`. La licencia
  tecnica interna heredada `PCS_SYSTEM_INTERNAL_PERPETUAL` se desactiva si
  existe; la empresa emisora debe tener una licencia comercial vigente para
  operar sus propios modulos. Las activaciones con
  total pagado cero por prueba o descuento total no emiten factura electronica
  en el flujo final.
  Si una empresa paga una licencia comercial antes de que venza la licencia
  actual, la nueva vigencia no reemplaza ni acorta la anterior: se programa
  desde el vencimiento acumulado mas lejano de esa empresa y queda lista para
  iniciar automaticamente al terminar la licencia vigente. Las tablas
  `pagos_epayco` y `pagos_wompi` guardan `licencia_activation_status`,
  `licencia_activada_id` y `licencia_activada_en` para que una consulta o
  webhook repetido no sume dias dos veces.
  Los codigos de descuento de licencias se administran desde Super
  administrador > Comercial y licencias > Codigos descuento; se guardan en
  `configuraciones.licencias.discount_codes` con formato `CODIGO=10%`,
  `CODIGO=50000` o `CODIGO=gratis`, y el checkout registra su uso en pagos o
  activaciones sin pago para bloquear reutilizacion por la misma empresa.
  El checkout publico de licencia debe mostrar Epayco y Wompi cuando sus
  credenciales reales estan configuradas; `*.enabled` solo se usa como override
  explicito para apagar una pasarela lista. La disponibilidad debe calcularse
  con credenciales realmente legibles: Epayco puede usar credenciales Smart o
  fallback `checkout.js` con Public Key valida aunque no exista `P_KEY`; Wompi
  requiere `public_key` con prefijo valido mas `integrity_key` descifrable para
  Web Checkout.
- Menu super administrador: `web/super_administrador.html` debe enlazar solo
  las paginas activas del panel super; `web/js/super_administrador.js` debe
  permitir restaurar cada enlace con `target="contentFrame"`. `Reportes globales`
  no va en el menu super y se conserva como vista del selector de empresas.
  `Metricas de trafico` no existe como pagina independiente; sus datos viven en
  `Centro de mando`. El acceso `Asesores de ventas` vive al inicio de
  `Comercial y licencias`.
- Panel empresarial: `web/administrar_empresa.html`,
  `web/administrar_empresa/panel.html`.
- Domotica: boton principal en `web/administrar_empresa.html`, submenu
  `web/administrar_empresa/modulo_menu.html?module=control_electrico`,
  consola `web/administrar_empresa/control_electrico.html` con vistas por
  `pagina=resumen|conexion|raspberry|reles|automatizaciones|reportes|bitacora`
  para que cada boton del submenu abra una pagina/vista independiente,
  endpoint `/api/empresa/control_electrico` y storage super en
  `web/super/domotica_storage.html` con `/super/api/domotica_storage`.
  Conserva la clave tecnica `control_electrico`; la carpeta empresarial de
  imagenes es `web/uploads/empresas/empresa_{id}_{slug}/imagenes/`, con
  subcarpetas como `domotica` y `usuarios`.
- Configuracion empresa: `web/administrar_empresa/configuracion_menu.html` y
  paginas bajo `web/administrar_empresa/configuracion/`. La configuracion de
  `Campos obligatorios para productos` vive en Configuracion > Productos y
  pedidos y guarda en `/api/empresa/inventario/configuracion`; el modulo de
  productos solo la consume para validar y marcar campos del formulario. El
  impuesto del formulario de productos se elige desde un selector construido con
  `/api/empresa/impuestos?action=context`; solo muestra impuestos habilitados
  para ventas/ambos de la empresa y guarda el porcentaje en
  `productos.impuesto_porcentaje`. En Configuracion > Impresora > Documento de
  venta, `empresa_configuracion_avanzada.mostrar_deducido_impuesto_factura`
  controla si recibos/facturas impresas muestran base gravable e impuesto
  deducido; es una opcion visual de impresion y no altera el XML ni la
  validacion legal de factura electronica DIAN.
- Usuarios empresa: `web/administrar_empresa/administrar_usuarios.html` consume
  `/api/empresa/roles_de_usuario`, que entrega un catalogo global deduplicado de
  `roles_de_usuario` para todos los tipos de empresa. La asignacion guarda
  `rol_usuario_id` y `rol_nombre`; los permisos efectivos se calculan por nombre
  normalizado y matriz de rol, no por tipo de empresa. Si al crear un usuario el
  correo ya existe dentro de la misma empresa, `/api/empresa/usuarios` devuelve
  `409` con `usuario_existente`; la pantalla recarga usuarios inactivos o
  pendientes, resalta el registro y permite usar `Reenviar confirmacion`.
- Configuracion carrito: `web/administrar_empresa/configuracion_carrito_de_compra_empresa.html`,
  `/api/empresa/estacion_prefs`; la visibilidad automatica de la tarjeta
  Domotica se guarda como `mostrar_tarjeta_domotica_carrito` dentro de
  `carrito_ui_global` o del override por estacion.
- Estaciones: `web/administrar_empresa/estaciones.html`,
  `/api/empresa/carritos_compra`.
- Carrito y venta directa: `web/administrar_empresa/carrito_de_compras.html`,
  `/api/empresa/carritos_compra`, `/api/empresa/carritos_compra/items`.
  Venta directa usa el carrito canonico `VENTA-DIRECTA-{empresa_id}-0`,
  comparte la UI unificada de estaciones y tiene boton de pantalla completa.
  Las cantidades de items en carrito son numeros naturales positivos: el
  frontend usa `min=1`/`step=1` y el backend rechaza cero, negativos, decimales o
  valores no finitos antes de tocar inventario.
  En Colombia/COP, la parte monetaria del carrito trabaja sin centavos: la UI
  muestra y edita pesos enteros y el backend normaliza precios unitarios, abonos,
  pagos simples/mixtos y ventas offline a enteros cuando `moneda=COP`.
  Los productos y recetas descuentan/reservan inventario en tiempo real al
  agregarse al carrito mediante `carrito_compra_items`; el pago no debe volver a
  descontar stock. El cierre `action=pagar_estacion` debe ser idempotente en
  backend: solo una solicitud puede pasar el carrito de abierto a cerrado, y los
  reintentos concurrentes no duplican documento, caja, metricas ni kardex.
  La apariencia plana del carrito se controla desde `web/estilos.css` con
  `body.carrito-flat-page`: no debe recuperar sombras, relieves ni tarjetas con
  apariencia 3D. El fondo estructural usa `--carrito-page-bg`, mas oscuro que
  las tarjetas `--carrito-card-bg`, para diferenciar zonas en todas las
  apariencias. Si se abre dentro de `web/administrar_empresa.html`, el iframe
  debe conservar `allow="geolocation; fullscreen"` y `allowfullscreen`.
- Caja y corte: `web/administrar_empresa/corte_de_caja.html`,
  `/api/empresa/corte_caja`, `/api/empresa/corte_caja/configuracion`.
- Reportes de turnos: `web/administrar_empresa/reportes_turnos.html`,
  `/api/empresa/corte_caja?action=turnos|turno_reporte|turno_export|turno_email`.
- Facturacion electronica: `web/administrar_empresa/facturacion_electronica_menu.html`,
  `/api/empresa/facturacion_electronica`,
  `/api/empresa/facturacion_electronica/panama`,
  `/api/empresa/facturacion_electronica/ecuador`.
- Facturacion offline: `/api/empresa/offline_ventas`,
  `backend/db/offline_ventas.go`. El carrito guarda la cola local por
  `empresa_id + usuario/cajero`, exige caja abierta cargada antes de vender sin
  internet y sincroniza con `sync_key` idempotente que incluye empresa, cajero,
  caja y carrito.
- Alertas sistema super administrador: `web/super/alertas_sistema.html`,
  `/super/api/alertas_sistema`.
- Mensajeria y alertas en super administrador: el menu lateral agrupa
  `web/super/alertas_sistema.html`, `web/super/configuracion/alertas_licencia.html`,
  `web/super/formato_para_emviar_email.html`, `web/super/correos_masivos.html`,
  `web/super/mantenimiento_sistema.html`, `web/super/configuracion/gmail_smtp.html`
  y `web/super/email_corporativo.html`. Los mensajes de compra/pago de licencia
  se editan desde `Formatos de email`; en esa misma pagina tambien se configura
  el texto del PDF `licencia_software_pdf`, que cada empresa descarga desde
  Administrar empresa > Licencia > Licencia del sistema y ya no se adjunta al
  correo de bienvenida.
- Email corporativo Mailu: `web/super/email_corporativo.html`,
  `/super/api/email_corporativo`, `/api/empresa/email_corporativo`,
  `backend/handlers/email_corporativo_handlers.go`,
  `backend/db/email_corporativo.go`,
  `deploy/scripts/vps-provision-mailu-mailbox.sh` y
  `deploy/scripts/vps-configure-mailu-host-nginx.sh`.
  El modulo genera email unico por empresa al crearla. El proveedor activo es
  Mailu con webmail SnappyMail en el perfil Docker `mail`. En la VPS se usa
  `mailu_direct`, que ejecuta el script directo contra `pcs-mailu-admin` y crea
  o actualiza buzones con `flask mailu user` y `flask mailu password`. El mismo
  script crea la identidad principal en SnappyMail para evitar el modal inicial.
  Los servicios Mailu tienen IPs fijas en `pcs_mailu_internal` para que IMAP,
  SMTP y webmail se hablen por la red confiable de Mailu. El archivo
  `deploy/mailu/snappymail-application.ini` conserva `secfetch_allow` para que
  el webmail pueda abrir en iframe `same-site` dentro del panel empresarial y
  usa `PCSLight@custom` como tema base. Los temas `PCSLight@custom` y
  `PCSDark@custom` se montan desde `deploy/mailu/themes`; el panel envia
  `theme=light|dark` al endpoint empresarial y al autologin. No imprimir claves: la clave
  inicial del buzon se guarda cifrada con `CONFIG_ENC_KEY` cuando aplica. En
  Docker portable, `deploy/.env.platform` provee `EMAIL_CORPORATIVO_*` /
  `MAILU_*`; `EMAIL_CORPORATIVO_AUTOLOGIN_SECRET` firma tokens HMAC de 2
  minutos para entrar al webmail sin exponer contrasenas. El proxy del host
  limpia cabeceras publicas y solo el backend inyecta cabeceras SSO hacia
  SnappyMail. La pagina de super administrador incluye diagnostico operativo y boton
  `Probar Mailu`; el panel empresarial abre la bandeja automaticamente cuando el
  buzon esta asignado, salvo que `empresa_estacion_prefs` tenga
  `email_corporativo_config.auto_open=false`. La pagina
  `web/administrar_empresa/configuracion/email_corporativo.html` permite cambiar
  esa preferencia y actualizar la contrasena interna del buzon; la clave siempre
  se guarda cifrada y no se muestra al usuario. La configuracion global del
  servidor de email define tambien `max_accounts_per_empresa`, con default 5,
  para limitar desde backend cuantas cuentas corporativas puede tener una misma
  empresa.
- Informacion de modulos del index: `web/super/informacion_de_modulos.html`,
  `/super/api/informacion_de_modulos`,
  `/api/public/informacion_de_modulos`.
- Noticias del portal: `web/noticias.html`, editor
  `web/super/noticias.html`, `/super/api/noticias` y
  `/api/public/noticias`. Se guarda en
  `pcs_superadministrador.configuraciones` con la clave
  `super.noticias_portal.v1`; la pagina publica se abre desde el menu flotante
  y tiene portada, foto de perfil y publicaciones tipo red social.
- Portal publico e index en super administrador: el menu lateral de
  `web/super_administrador.html` agrupa tarjetas del index
  (`web/super/pagina_principal.html`), modulos del index
  (`web/super/informacion_de_modulos.html`), noticias
  (`web/super/noticias.html`), descripcion de sistemas para IA y portal
  (`web/super/informacion_de_la_empresa_y_de_los_sistemas_para_ia.html`),
  WhatsApp del portal (`web/super/configuracion/whatsapp_portal.html`) y accesos
  de lectura a `web/index.html` y `web/descripcion_de_los_sistemas.html`.
- Energia solar: `web/administrar_empresa/energia_solar.html`,
  `web/js/energia_solar.js`, `/api/empresa/energia_solar`,
  tablas `empresa_energia_solar_*`. El modulo es por empresa, usa permiso
  `energia_solar`, soporta Victron/SMA/SolarEdge/gateway local y alerta por
  correo usando SMTP configurado. Las preconfiguraciones por tipo incluyen
  `modulos.energia_solar` apagado por defecto, con catalogo de proveedores,
  baterias y alertas; el rol `tecnico_solar` solo recibe lectura.
- Analitica publica por pais: `/api/public/portal_visitas`,
  `web/js/portal_visits.js`.
- Chat IA/emisora flotante: `web/js/ai_chat_drawer.js`,
  `web/js/radio_player.js`, `web/js/radio_online.js`,
  `/api/chat_flotante/preferencias`. En contexto empresarial, el chat IA debe
  iniciar activo por defecto en preproduccion, siempre en modo `normal`; robot y
  secretaria no se muestran ni deben prenderse por configuracion global o
  `localStorage` viejo. La emisora online sigue apagada salvo preferencia
  explicita por `empresa_id`.

## Flujo de login

El usuario entra por `login.html`. El backend valida credenciales u OAuth Google
en handlers de autenticacion administrativa. El registro de administradores usa
la pagina de registro y crea cuentas administrativas con confirmacion segun la
configuracion vigente. Las alertas super administrador pueden enviar correo
cuando se registra un administrador, sin incluir contrasenas ni tokens.

Los usuarios operativos entran por `login_usuario.html`. El acceso por correo,
contrasena o Google siempre debe resolver un usuario ya creado/invitado por una
empresa; no existe alta publica operativa. Para Google se usa
`/auth/google/usuario/login`, que marca el flujo como usuario y vuelve por el
callback canonico `/auth/google/callback`; el callback solo abre sesion si el
correo verificado por Google coincide con una invitacion vigente o con un
usuario empresarial ya confirmado. La sesion redirige a
`administrar_empresa.html?id={empresa_id}` para que el panel cargue roles y
permisos efectivos de esa empresa.

## Flujo de creacion de empresa

Desde `seleccionar_empresa.html`, el administrador crea una empresa eligiendo
tipo. El backend crea la empresa, aplica preconfiguracion por tipo, prepara
permisos/modulos y puede activar una licencia gratis de 15 dias si corresponde.
La licencia gratis solo puede usarse una vez por empresa y pertenece al catalogo
global compartido por todos los tipos de empresa. El bloqueo usa historial de
activaciones y licencias gratis antiguas, no solo licencias vigentes. La
creacion de empresa puede disparar una alerta por correo al super administrador
si el check esta activo.

## Flujo de administradores delegados

El administrador principal normal puede abrir `Administradores` desde
`seleccionar_empresa.html`. Esa pagina no es global para el: `/super/api/administradores`
filtra por el principal resuelto, excluye al propio principal y permite gestionar
solo cuentas con `administradores.usuario_creador` dentro de su alcance. El alta
se hace por invitacion: se crea cuenta pendiente, se envia correo con
`invitation_token`, el invitado completa `registrar_nuevo_usuario_administrador.html`
y solo despues de validar el token queda confirmado para login. Los delegados
heredan acceso a las empresas creadas por el principal como `access_source=delegated`,
pero no pueden compartirlas ni administrar otros administradores. La validacion
real vive en backend y en `CanAdminAccessEmpresaIA`.

## Flujo de carrito, venta, caja y facturacion

- Venta directa usa el carrito canonico de venta directa de la empresa.
- Estaciones usan carritos asociados a cada estacion.
- El carrito agrega productos, servicios o recetas, calcula totales, abonos,
  pagos mixtos y medio de pago.
- Cada usuario/caja debe operar de forma independiente dentro de la misma
  empresa.
- Las cajas fisicas configurables de la empresa se guardan en
  `empresa_estacion_prefs.estaciones_config.cajas_config` con `codigo`,
  `nombre`, `descripcion` y `activa`; la estacion Caja y el carrito muestran el
  nombre operativo, por ejemplo `CAJA-1 - FRUTERA`.
- El pago cierra el carrito, actualiza inventario/caja, genera documento
  imprimible y, si aplica, documento electronico.
- Caja y corte usan `corte_de_caja.html`; el reporte de turno se calcula por
  usuario/caja/turno y se imprime por defecto en POS 80mm.
- Los documentos imprimibles, facturas, recibos, notas y reportes fiscales deben
  verse como papel real en blanco y negro, sin depender de tema claro u oscuro.
- Si esta activo, el QR DIAN al final de factura/recibo se genera desde CUFE,
  CUDE o codigo de validacion.

## Donde se guardan configuraciones frecuentes

- Configuracion visual/operativa del carrito:
  `empresa_estacion_prefs.estaciones_config.carrito_ui_global`.
- Configuracion chat IA/emisora:
  claves `chat_flotante.*` en `empresa_estacion_prefs` con `estacion_id=0`.
  `chat_enabled` queda activo por defecto en preproduccion; `robot_enabled`
  permanece en `0`, `personality_mode=normal` y la emisora es opt-in por empresa.
- Overrides por estacion:
  `empresa_estacion_prefs.estaciones_config.estaciones[].carrito.configuracion`.
- Configuracion de estaciones y nombres singular/plural:
  `empresa_estacion_prefs.estaciones_config`.
- Catalogo de cajas fisicas simultaneas:
  `empresa_estacion_prefs.estaciones_config.cajas_config`.
- Configuracion de corte/reporte de caja:
  `empresa_corte_caja_configuracion`.
- Impresoras por empresa y POS 80mm:
  `empresa_impresoras*`.
- Facturacion electronica por pais:
  `facturacion_electronica_pais`.
- Reintentos/cola documental electronica:
  `facturacion_electronica_reintentos`.
- Configuracion super global:
  `pcs_superadministrador.configuraciones` o tablas super dedicadas.
- Alertas sistema:
  `super_alertas_config` y `super_alertas_eventos`.

## Scripts reales

- `.\rs.ps1`: wrapper operativo principal solicitado por el usuario para
  preflight, actualizacion/sincronizacion y tareas de runtime segun el script.
- `.\scripts\rs.ps1`: script base relacionado con el flujo `rs`.
- `.\scripts\sync_to_vps.ps1`: sincroniza hacia VPS.
- `.\scripts\sync_to_vps.sh`: alternativa shell para sincronizacion.
- `.\scripts\profesional_preflight.ps1`: validaciones previas.
- `.\scripts\actualizar_repositorio.ps1`: actualizacion de repositorio.
- `.\scripts\publicar_git_y_vps.ps1`: publicacion coordinada Git/VPS.

Antes de ejecutar scripts operativos revisar `documentos/comandos_codex.md`.

## Datos de prueba permitidos

- Empresa de prueba para motel/POS/estaciones/caja: `Motel Calipso`.
- Empresa de prueba para creditos de motos: `Venta Moto`.
- Usuario administrativo de prueba autorizado por el usuario:
  `powerfulcontrolsystem@gmail.com`. No repetir ni guardar claves en
  documentacion, consola o commits.
- Cuando se creen datos de prueba, dejar claro si son demo/preproduccion y no
  mezclar empresas sin validar `empresa_id`.

## Seguridad que siempre debe conservarse

- Validar `empresa_id` en backend y en consultas SQL.
- No permitir que una empresa lea, edite o borre datos de otra.
- Antes de crear, modificar o revisar endpoints empresariales, aplicar
  `documentos/checklist_seguridad_endpoint_multiempresa.md`.
- No imprimir secretos ni credenciales.
- Mantener auditoria en operaciones criticas: caja, pagos, facturacion,
  licencias, usuarios, backups, conectividad y cambios de configuracion.
- En tareas de limpieza, backup o reinicio de datos, conservar configuracion,
  usuarios, permisos e integraciones salvo instruccion explicita.

## IA super administrador 2026-06-09

- La configuracion IA del super administrador queda remodelada como
  `Gobierno de inteligencia artificial` en
  `web/super/configuracion/ia_global.html` y en el card `aiConfigCard` de
  `web/super/configuracion_avanzada.html`.
- La pantalla concentra estado global, proveedor OpenAI, credencial cifrada,
  consumo, prueba real y accesos a limites, contexto de negocio y voz.
- La IA visible ya no usa robot ni secretaria: se presenta como recuadro normal
  y boton flotante circular, conservando voz cuando este habilitada.
- El icono flotante de IA se unifico al mismo simbolo limpio usado en el resto
  del proyecto, sin contenedor blanco ni variantes visuales por pagina.
- La credencial OpenAI se conserva en `pcs_superadministrador.configuraciones`
  mediante `/super/api/config/ai`; no se imprime completa ni se duplica en
  documentacion.
- La politica operativa mantiene lectura preparada por servidor, roles,
  aislamiento por `empresa_id` y acciones estructuradas solo contra endpoints
  autorizados; no hay SQL libre ni secretos expuestos a la IA.

## Facturacion electronica Colombia DIAN 2026-06-17

- PCS (`empresa_id=12`) ya usa la autorizacion DIAN Formulario 1876 con
  resolucion `18764111318575`, prefijo `1PCS`, rango `1-100000` y vigencia
  `2026-06-17` a `2028-06-17`.
- La numeracion nueva debe estar asociada en el portal DIAN de produccion
  `https://catalogo-vpfe.dian.gov.co/User/Login` al software
  `Powerful Control System` antes de emitir. Si falta ese paso, DIAN responde
  `FAB05c`.
- El sistema consulta `GetNumberingRange` para traer/actualizar la clave tecnica
  del rango; esa clave es necesaria para calcular CUFE correctamente y no debe
  imprimirse.
- Revalidacion 2026-06-18: la venta de producto `menta` en PCS con cliente IVAN
  FRANCISCO CAYON GUARNIZO genero factura `1PCS2`; el usuario confirmo luego en
  portal DIAN produccion que aparece como `Aprobado con notificacion`. Un
  reenvio puede devolver `Regla: 90, Documento procesado anteriormente`; PCS no
  debe marcar esa regla como aceptacion automatica sin acuse original, portal
  DIAN, CUFE/TrackId o evidencia oficial equivalente.
- Prueba real cerrada 2026-06-18: se emitio `1PCS3` en produccion mediante
  SOAP/WCF `SendBillSync`; DIAN respondio HTTP 200 con `estado_dian=aceptado`,
  `acuse_estado=aceptado`, CUFE registrado y notificacion `RUT01` informativa.
- Confirmacion de portal 2026-06-18: el usuario valido en DIAN produccion,
  documentos recibidos, que `1PCS3` (18-06-2026) y `1PCS2` (17-06-2026) aparecen
  como `Aprobado con notificacion`, factura electronica, valor `$ 100`.
- El siguiente consecutivo operativo despues de la prueba directa quedo en
  `1PCS4`; no reutilizar `1PCS2` ni `1PCS3`.
- Cliente de prueba autorizado: IVAN FRANCISCO CAYON GUARNIZO, CC `84456779`,
  Colombia, Santa Marta, direccion `calle 28 N5-116`, celular `3043306506`,
  persona natural. No guardar contrasenas en documentacion.
- Los errores DIAN visibles se explican en
  `web/administrar_empresa/facturacion_electronica_tutorial_dian.html` y la
  consola de `facturacion_electronica_pruebas_dian.html` los muestra en rojo con
  solucion operativa.
- Si un envio DIAN queda fallido, `processFacturacionIntegracionForDocumento`
  crea una alerta en `empresa_buzon_mensajes` para el administrador/creador de la
  empresa, con prioridad alta, modulo `facturacion_electronica` y enlace directo
  al centro DIAN.

## IA super administrador 2026-06-17

- El menu lateral del super administrador debe mostrar solo una opcion bajo `IA`:
  `Configuracion IA` -> `web/super/configuracion/ia_global.html`.
- Las paginas tecnicas antiguas de IA pueden existir como soporte interno, pero
  no deben repetirse como cinco entradas del submenu. La pagina unica concentra
  proveedor, credencial, reglas, contexto, chat global y voz.
