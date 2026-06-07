# Flujos operativos

Guia corta de los procesos que mas se prueban y modifican. Cada flujo debe
mantener aislamiento por `empresa_id`, permisos por rol y trazabilidad cuando
afecte dinero, documentos, licencias o seguridad.

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
4. La creacion debe ser idempotente: doble clic, reintento o solicitud
   concurrente con el mismo administrador, tipo, nombre y NIT debe devolver la
   empresa ya creada sin insertar otra ni repetir avisos.
5. El backend prepara la carpeta empresarial
   `web/uploads/empresas/empresa_{id}_{slug}/` con subcarpeta `imagenes` y la
   carpeta privada `facturacion_electronica/firma_electronica`.
6. Si esta activo el aviso de empresa nueva, se notifica al super administrador
   solo cuando realmente se inserta una empresa nueva.
7. Pruebas: empresa creada, aparece en selector, entra a panel, conserva
   `empresa_id` correcto.
8. Pruebas negativas: doble submit del mismo formulario y dos POST iguales no
   deben crear empresas duplicadas.

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
6. `Restablecer orden` borra la preferencia y vuelve al orden base.
7. Seguridad: el orden no concede acceso a empresas; solo reordena tarjetas que
   `/super/api/empresas` ya autorizo para la sesion actual.
8. Pruebas: mover tarjetas en activas e inactivas, recargar, confirmar
   persistencia y restablecer.

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
   prueba gratis 15 dias, plan COP 60000, plan COP 100000 y plan COP 150000.
   Las licencias base antiguas por tipo y addons de catalogo se eliminan del
   catalogo sin empresa asignada.
2. Desde el checkout de licencia se obtiene resumen publico.
3. Si el total es cero o prueba permitida, `POST /licencias/activar_sin_pago`
   activa la licencia.
4. El backend valida que esa empresa no haya usado antes la licencia gratis,
   mirando historial completo de activaciones y licencias gratis antiguas,
   aunque la licencia anterior ya este vencida o inactiva.
5. La activacion debe ser idempotente si el primer intento ya dejo la licencia
   vigente.
6. Al quedar activa, el sistema envia un solo correo al administrador de la
   empresa con el mensaje `licencia_activation_payment` y adjunta un PDF de
   licencia de software generado desde la plantilla `licencia_software_pdf`,
   ambas configurables en `web/super/formato_para_emviar_email.html`.
   Si el pago comercial queda aprobado con valor mayor a cero, ademas se emite
   automaticamente una factura electronica desde la empresa interna `Powerful
   Control System` (tambien reconoce el nombre existente `Powerful Control
   Systen`) y el PDF resumen de esa factura se adjunta al mismo correo de
   activacion. El documento se guarda en `empresa_facturacion_documentos` de la
   empresa emisora y el pago queda marcado con
   `licencia_factura_electronica_emitida` en `pagos_epayco` o `pagos_wompi`
   para idempotencia.
   La empresa emisora interna se resuelve por
   `configuraciones.licencias.facturacion_empresa_sistema_id` y mantiene una
   licencia tecnica interna `PCS_SYSTEM_INTERNAL_PERPETUAL` con fecha fin a 100
   anos, limites altos y modulos completos. La empresa interna debe cargar
   carrito, correo corporativo, facturacion y configuracion como cualquier otra
   empresa; la unica diferencia es que esa licencia no pertenece al catalogo
   comercial. Las activaciones con total pagado cero por prueba o descuento
   total solo envian la licencia; no emiten factura electronica en el flujo
   final.
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
11. Pruebas: activar una vez, reintentar sin duplicar mientras sigue vigente,
   bloquear segundo uso real despues del vencimiento y comprobar que el
   historial muestra otras licencias cuando la prueba no es renovable, ademas
   de validar que el correo capturado o enviado incluya el PDF de licencia y,
   cuando el pago sea mayor a cero, el PDF de factura electronica en el mismo
   mensaje; confirmar tambien que descuento total o valor cero no genera factura
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
6. Pruebas: guardar seccion, recargar pagina, validar que otra seccion no cambio
   e imprimir una venta con impuesto para confirmar el bloque.

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
   carga de firma, prueba 2+2+2, set completo, acuse final y activacion local
   de produccion. Debe mantenerse sin secretos reales.
16. Estado operativo actual: la prueba `Enviar prueba 2 + 2 + 2` debe quedar
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
   estado de alerta. En `Pasar test DIAN`, guardar objetivo, validar
   credenciales, ejecutar al menos un envio manual de factura de prueba y correr
   `Enviar prueba 2 + 2 + 2`; despues consultar `GetStatusZip` hasta cierre real
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
2. El usuario abre `login_usuario.html` desde la invitacion para completar
   registro o iniciar con Google. Sin invitacion o usuario empresarial confirmado
   no hay alta publica. En este primer ingreso, un usuario pendiente puede tener
   `estado=inactivo`; si el token es valido, el sistema permite crear la
   contrasena, confirma el correo y cambia el estado a `activo`. Un usuario ya
   confirmado e inactivo sigue bloqueado hasta que el administrador lo active.
3. `Iniciar sesion con Google` usa `/auth/google/usuario/login`, conserva
   `empresa_id` y token de invitacion en cookies tecnicas de corta vida y vuelve
   por `/auth/google/callback`.
4. El callback solo abre sesion si Google confirma el correo y este coincide con
   una invitacion vigente o con un usuario de empresa ya confirmado. Si falta
   contrato vigente, vuelve al formulario para aceptarlo.
5. La sesion redirige siempre a `administrar_empresa.html?id={empresa_id}`; el
   panel carga rol, permisos y estaciones asignadas de esa empresa.
6. Para `cajero`, si la configuracion de estaciones tiene activo
   `solicitar_caja_login_cajero` (activo por defecto), despues de validar
   credenciales se muestra una ventana para elegir la caja fisica de trabajo
   del dia. La lista sale de `estaciones_config.cajas_config`, recuerda la
   ultima caja usada por usuario/empresa en el navegador y propaga
   `caja_codigo`, `caja_nombre` y `caja_descripcion` a estaciones, carrito y
   corte de caja.
7. Para `cajero`, el menu queda limitado a `Venta directa`, `Estaciones` y
   `Corte de Caja`, pero el carrito debe cargar completo: catalogo de productos,
   servicios, recetas, clientes, descuentos, propinas/comisiones y valores por
   medio de pago. Esas APIs auxiliares solo se permiten dentro del alcance del
   carrito, sin mostrar paginas administrativas de Productos o Clientes.
8. Pruebas: Google sin invitacion debe rechazar, Google con invitacion debe
   consumir token y entrar, correo ambiguo exige enlace de empresa, tema claro u
   oscuro se conserva, el boton `Instalar app` permanece visible y el cajero ve
   el selector de caja solo cuando el check de configuracion esta activo.

## Abrir, usar y cerrar caja

1. El usuario entra a caja desde `Corte de Caja` o desde la estacion Caja.
2. La empresa puede configurar varias cajas fisicas en
   `estaciones_config.cajas_config`, cada una con codigo, nombre, descripcion y
   estado activo. La estacion Caja muestra esos nombres, por ejemplo
   `CAJA-1 - FRUTERA`.
3. En la misma seccion de configuracion existe el check
   `solicitar_caja_login_cajero`, activo por defecto, para exigir a los cajeros
   elegir caja al iniciar sesion operativa.
4. Al hacer clic en una caja configurada, `corte_de_caja.html` recibe
   `caja_codigo`, `caja_nombre` y descripcion para abrir el corte de esa caja.
5. La caja puede abrirse manual o automaticamente segun flujo vigente.
6. Cada usuario/caja mantiene turno, pagos, ingresos, egresos y reporte
   independiente.
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
   Las cantidades de items deben ser numeros naturales positivos (`1, 2, 3...`);
   el backend rechaza decimales, cero y negativos aunque se manipule el
   navegador o la API.
   Si la empresa detectada es Colombia o el carrito usa `COP`, precios y pagos
   del carrito se capturan, muestran y sincronizan como pesos enteros positivos,
   sin centavos ni sufijo `.00`.
4. La venta directa usa el carrito canonico `VENTA-DIRECTA-{empresa_id}-0` y
   puede abrirse en pantalla completa desde el boton de la parte superior; el
   mismo boton cambia a `Salir` y vuelve a la vista normal.
5. Visualmente debe conservar el modo plano del carrito: tarjetas sin sombras ni
   apariencia 3D, pero con el fondo estructural mas oscuro que las tarjetas para
   diferenciar zonas en cualquier apariencia.
6. Pruebas: entrar con rol `cajero` por `login_usuario`, agregar item, buscar
   producto, pagar, imprimir, validar inventario y caja, abrir/salir de pantalla
   completa y revisar contraste fondo/tarjetas.

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
6. La impresion debe salir en blanco y negro como papel real, POS 80mm por
   defecto, sin tema claro/oscuro.
7. Si hay QR DIAN activo y documento con CUFE/CUDE/codigo, se imprime al final.
8. Pruebas: efectivo, debito, credito, otro, pago mixto, vuelto, abono,
   descuento, dos cajeros simultaneos y doble solicitud de pago sobre el mismo
   carrito.

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
4. Vista e impresion deben adaptarse a POS 80mm y carta; POS 80mm es default.
5. Pruebas: turno con ventas, descuento, tarjeta, ingreso, egreso, anulacion,
   exportar/imprimir.

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
   `recursos_humanos`, `tecnico_solar`, `cajero`, `portero`,
   `servicio_limpieza`, `contador`, `empresario`, `compras`, `inventario`,
   `contabilidad` y `auditor`.
2. `tecnico_solar` solo consulta el estado de energia solar: dashboard,
   lecturas, eventos y alertas. No puede modificar sistemas ni configuracion.
3. `jefe_bodega` administra inventario y bodegas: existencias, traslados,
   categorias, recetas y codigos; no puede operar ventas, caja ni eliminar
   inventario.
4. `recursos_humanos` gestiona horarios, asistencia y nomina operativa; no abre
   ventas, caja ni configuracion general.
5. Pruebas: crear usuarios con esos roles, iniciar por `login_usuario`, validar
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

1. La ficha de nomina del empleado guarda `sede_codigo`, `sede_nombre` y
   `centro_costo`; estos valores se copian a cada liquidacion para conservar la
   trazabilidad historica aunque luego cambie la ficha.
2. La liquidacion se genera desde asistencia, novedades aprobadas, recargos,
   comisiones, provisiones y deducciones; el dashboard resume empleados,
   liquidaciones, pagos, costo empresa y sedes activas.
3. Para Colombia, la seccion avanzada consulta conceptos, novedades, PILA y el
   resumen de documentos electronicos de nomina.
4. `GET /api/empresa/nomina?action=documentos_electronicos_colombia` valida el
   periodo y muestra liquidaciones listas por empleado para documento soporte de
   pago de nomina electronica.
5. `POST /api/empresa/nomina?action=preparar_nomina_electronica` prepara el lote
   por empleado con devengados, deducciones, neto, IBC, sede y centro de costo.
   El envio real a DIAN sigue dependiendo de firma, CUNE, numeracion,
   credenciales y transporte documental configurados por empresa en facturacion
   electronica.
6. Pruebas: usar `Crear nomina demo Motel Calipso`, verificar empleados en varias
   sedes, liquidaciones por sede, PILA, pagos, desprendible y botones `Ver estado
   DIAN` / `Preparar lote DIAN`.
