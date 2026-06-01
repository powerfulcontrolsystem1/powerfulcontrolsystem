# Flujos operativos

Guia corta de los procesos que mas se prueban y modifican. Cada flujo debe
mantener aislamiento por `empresa_id`, permisos por rol y trazabilidad cuando
afecte dinero, documentos, licencias o seguridad.

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
5. Si esta activo el aviso de empresa nueva, se notifica al super administrador
   solo cuando realmente se inserta una empresa nueva.
6. Pruebas: empresa creada, aparece en selector, entra a panel, conserva
   `empresa_id` correcto.
7. Pruebas negativas: doble submit del mismo formulario y dos POST iguales no
   deben crear empresas duplicadas.

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
5. Pruebas: no permitir borrado sin validaciones, ofrecer descarga previa,
   eliminar solo la empresa indicada y volver al selector sin filtrar datos de
   otra empresa.

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
6. Al quedar activa, el sistema envia correo al administrador de la empresa con
   el mensaje `licencia_activation_payment` y adjunta un PDF de licencia de
   software generado desde la plantilla `licencia_software_pdf`, ambas
   configurables en `web/super/formato_para_emviar_email.html`.
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
   de validar que el correo capturado o enviado incluya el PDF adjunto y que la
   descarga empresarial devuelva `application/pdf`; para pago, seleccionar
   Epayco y Wompi, aceptar terminos y comprobar que cada proveedor pasa a
   verificacion con referencia propia. Para renovaciones comerciales, simular
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
5. Pruebas: guardar seccion, recargar pagina, validar que otra seccion no cambio.

## Login usuarios operativos

1. El administrador de una empresa crea el usuario desde `Administrar usuarios`;
   el sistema envia invitacion por correo y guarda token temporal en `users`.
2. El usuario abre `login_usuario.html` desde la invitacion para completar
   registro o iniciar con Google. Sin invitacion o usuario empresarial confirmado
   no hay alta publica.
3. `Iniciar sesion con Google` usa `/auth/google/usuario/login`, conserva
   `empresa_id` y token de invitacion en cookies tecnicas de corta vida y vuelve
   por `/auth/google/callback`.
4. El callback solo abre sesion si Google confirma el correo y este coincide con
   una invitacion vigente o con un usuario de empresa ya confirmado. Si falta
   contrato vigente, vuelve al formulario para aceptarlo.
5. La sesion redirige siempre a `administrar_empresa.html?id={empresa_id}`; el
   panel carga rol, permisos y estaciones asignadas de esa empresa.
6. Pruebas: Google sin invitacion debe rechazar, Google con invitacion debe
   consumir token y entrar, correo ambiguo exige enlace de empresa, tema claro u
   oscuro se conserva y el boton `Instalar app` permanece visible.

## Abrir, usar y cerrar caja

1. El usuario entra a caja desde `Corte de Caja` o desde la estacion Caja.
2. La empresa puede configurar varias cajas fisicas en
   `estaciones_config.cajas_config`, cada una con codigo, nombre, descripcion y
   estado activo. La estacion Caja muestra esos nombres, por ejemplo
   `CAJA-1 - FRUTERA`.
3. Al hacer clic en una caja configurada, `corte_de_caja.html` recibe
   `caja_codigo`, `caja_nombre` y descripcion para abrir el corte de esa caja.
4. La caja puede abrirse manual o automaticamente segun flujo vigente.
5. Cada usuario/caja mantiene turno, pagos, ingresos, egresos y reporte
   independiente.
6. `Corte automatico` calcula desde apertura hasta el momento actual sin pedir
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
4. La venta directa usa el carrito canonico `VENTA-DIRECTA-{empresa_id}-0` y
   puede abrirse en pantalla completa desde el boton de la parte superior; el
   mismo boton cambia a `Salir` y vuelve a la vista normal.
5. Visualmente debe conservar el modo plano del carrito: tarjetas sin sombras ni
   apariencia 3D, pero con el fondo estructural mas oscuro que las tarjetas para
   diferenciar zonas en cualquier apariencia.
6. Pruebas: agregar item, buscar producto, pagar, imprimir, validar inventario y
   caja, abrir/salir de pantalla completa y revisar contraste fondo/tarjetas.

## Estaciones y carrito

1. La pagina de estaciones carga configuracion y carritos por `empresa_id`.
2. Al activar una estacion cambia estado y otros usuarios deben ver el cambio.
3. Si el check de primer clic esta activo, el primer clic solo activa; el segundo
   entra al carrito.
4. El carrito de estacion comparte UI y reglas con venta directa.
5. La apariencia del carrito de estacion tambien depende de `carrito-flat-page`:
   fondo mas oscuro que tarjetas, sin sombras, con botones de accion visibles
   como botones.
6. Pruebas: dos sesiones/usuarios, estado compartido, abrir carrito correcto,
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
   pruebas y cola documental.
3. Panama y Ecuador tienen paginas propias con configuracion de DGI/SRI.
4. Credenciales, firma, NIT/RUC y trazabilidad son por empresa.
5. Pruebas: guardar configuracion por pais, validar checklist, generar documento,
   enviar correo si aplica, revisar cola/reintentos.

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
