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
4. Si esta activo el aviso de empresa nueva, se notifica al super administrador.
5. Pruebas: empresa creada, aparece en selector, entra a panel y conserva
   `empresa_id` correcto.

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
7. Eliminar desde el listado revoca la delegacion si la cuenta ya era de otro administrador; no borra su cuenta.
8. Pruebas: principal invita delegado, correo/enlace funciona, delegado ve
   empresas del principal, no ve empresas de otro principal y no puede compartir
   por URL ni boton.

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

1. Desde el checkout de licencia se obtiene resumen publico.
2. Si el total es cero o prueba permitida, `POST /licencias/activar_sin_pago`
   activa la licencia.
3. El backend valida que esa empresa no haya usado antes la licencia gratis.
4. La activacion debe ser idempotente si el primer intento ya dejo la licencia
   vigente.
5. Pruebas: activar una vez, reintentar sin duplicar, bloquear segundo uso real.

## Configurar empresa

1. El menu `Configuracion` abre paginas independientes por seccion.
2. Cada pagina carga datos por `empresa_id`, guarda solo su seccion y conserva el
   resto de configuracion.
3. Las preferencias flexibles de estaciones/carrito usan
   `empresa_estacion_prefs.estaciones_config`.
4. Pruebas: guardar seccion, recargar pagina, validar que otra seccion no cambio.

## Abrir, usar y cerrar caja

1. El usuario entra a caja desde `Corte de Caja` o desde la estacion Caja.
2. La caja puede abrirse manual o automaticamente segun flujo vigente.
3. Cada usuario/caja mantiene turno, pagos, ingresos, egresos y reporte
   independiente.
4. `Corte automatico` calcula desde apertura hasta el momento actual sin pedir
   fechas.
5. `Cerrar turno e imprimir reporte` imprime y luego cierra sesion.
6. Pruebas: abrir caja con dos usuarios, registrar movimientos, cerrar una caja
   sin afectar la otra.

## Venta directa

1. `Venta directa` abre `carrito_de_compras.html` en modo venta directa.
2. Debe usar el mismo carrito unificado de estaciones, con configuracion global
   del carrito.
3. El cajero agrega productos/servicios/recetas, cliente opcional u obligatorio,
   abonos y pagos mixtos.
4. Pruebas: agregar item, buscar producto, pagar, imprimir, validar inventario y
   caja.

## Estaciones y carrito

1. La pagina de estaciones carga configuracion y carritos por `empresa_id`.
2. Al activar una estacion cambia estado y otros usuarios deben ver el cambio.
3. Si el check de primer clic esta activo, el primer clic solo activa; el segundo
   entra al carrito.
4. El carrito de estacion comparte UI y reglas con venta directa.
5. Pruebas: dos sesiones/usuarios, estado compartido, abrir carrito correcto.

## Pagar e imprimir

1. El usuario presiona pagar en el carrito.
2. Backend valida caja, items, totales, abonos, cliente obligatorio si aplica,
   descuentos, inventario y permisos.
3. Se registra venta/pago, se actualiza inventario y se genera documento.
4. La impresion debe salir en blanco y negro como papel real, POS 80mm por
   defecto, sin tema claro/oscuro.
5. Si hay QR DIAN activo y documento con CUFE/CUDE/codigo, se imprime al final.
6. Pruebas: efectivo, debito, credito, otro, pago mixto, vuelto, abono, descuento.

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
2. Si se pierde internet en caja/carrito con offline activo, aparece aviso
   persistente y se permite vender e imprimir provisionalmente.
3. Si se pierde internet en modulo sin soporte offline, el aviso debe pedir
   esperar reconexion.
4. Al volver internet, se muestra aviso, se registra auditoria y se sincroniza la
   cola por `/api/empresa/offline_ventas`.
5. Pruebas: cortar red, vender, imprimir, restaurar red, sincronizar una sola vez.

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

## Alertas super administrador

1. `web/super/alertas_sistema.html` concentra alertas y notificaciones.
2. Las opciones de correo para registros y empresas nuevas se guardan en
   `super_alertas_config`.
3. Los envios quedan auditados en `super_alertas_eventos`.
4. Un fallo SMTP no debe bloquear el flujo de negocio que disparo la alerta.
5. Pruebas: guardar checks, enviar prueba, crear admin/empresa y revisar evento.
