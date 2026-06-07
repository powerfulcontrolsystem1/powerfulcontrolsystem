# Decisiones tecnicas permanentes

Este archivo evita rediscutir reglas ya decididas por el proyecto. Si una tarea
necesita cambiar una de estas reglas, debe pedir autorizacion explicita y dejar
trazabilidad en `documentos/historial_de_cambios`.

## Backend

- Go puro y libreria estandar siempre que aplique.
- No agregar dependencias externas, imports de terceros, binarios ni cambios en
  `go.mod` sin autorizacion explicita.
- PostgreSQL es el unico motor de base de datos permitido.
- No reintroducir SQLite, MySQL, MariaDB u otros motores en runtime, pruebas
  operativas, utilidades vigentes o documentacion actual.
- Mantener compatibilidad con PostgreSQL en SQL, migraciones ligeras y tests.

## Frontend

- Frontend tradicional con HTML, CSS y JavaScript estatico.
- No migrar a frameworks ni bundlers sin aprobacion explicita.
- Mantener responsive real para PC y celular.
- Validar visualmente cambios de pantallas, formularios, botones, carritos,
  reportes e impresion.
- Los botones deben tener iconos relacionados usando la infraestructura global
  existente, sin agregar dependencias.
- El carrito de compras debe mantener apariencia plana: sin sombras, sin efecto
  3D y sin separaciones visuales innecesarias entre tarjetas. Para que no se vea
  todo del mismo bloque, el fondo estructural del carrito debe ser mas oscuro
  que las tarjetas usando variables de tema (`--carrito-page-bg` y
  `--carrito-card-bg`) y debe funcionar en apariencias claras y oscuras.

## Multiempresa y seguridad

- Todo dato operativo debe aislarse por `empresa_id`.
- Todo endpoint o consulta multiempresa debe pasar por
  `documentos/checklist_seguridad_endpoint_multiempresa.md` antes de cerrarse.
- El backend debe validar `empresa_id` y permisos efectivos. No confiar solo en
  parametros de URL, cache, localStorage o controles del frontend.
- Toda consulta o mutacion multiempresa debe filtrar por `empresa_id` cuando la
  tabla pertenezca a una empresa.
- No imprimir secretos, claves, tokens, certificados, contrasenas ni datos
  privados en consola, documentacion o commits.
- Operaciones criticas deben tener auditoria o historial razonable: caja, pagos,
  licencias, facturacion, usuarios, backups, conectividad y configuraciones.
- Las altas y acciones criticas deben ser idempotentes en backend cuando puedan
  duplicarse por doble clic, reintento de red, service worker, modo offline o
  concurrencia. La UI puede bloquear botones, pero la garantia real debe vivir
  en la capa de datos/handler.

## Configuracion empresarial

- Usar `empresa_estacion_prefs.estaciones_config` para configuracion flexible de
  estaciones, carrito y preferencias operativas cuando aplique.
- Usar `carrito_ui_global` como base global del carrito y overrides por estacion
  solo cuando sea necesario.
- POS 80mm es el formato predeterminado para reportes de turno y documentos POS.
- Las empresas nuevas y antiguas de preproduccion pueden recibir defaults
  globales si el usuario lo solicita, porque el sistema aun no esta en produccion.

## Documentos imprimibles

- Facturas, recibos, notas, documentos electronicos y reportes imprimibles deben
  mostrarse como papel real en blanco y negro.
- No deben depender del tema claro/oscuro de la aplicacion.
- La factura o venta debe parecerse al documento fiscal/electronico aplicable
  cuando corresponda.
- El bloque opcional de deducido de impuesto en la impresion de factura o recibo
  es solo presentacional: usa base gravable e impuesto ya calculados y no debe
  modificar XML, CUFE/CUDE, envio, validacion ni reglas legales de la DIAN.
- El reporte de turno debe ser compacto, claro, profesional y adaptable a POS
  80mm y carta, pero no necesita copiar la apariencia de una factura electronica.

## Facturacion electronica

- Colombia mantiene modelo SaaS con software DIAN compartido cuando aplique, pero
  NIT, credenciales, firma, certificados y trazabilidad son por empresa.
- El transporte oficial DIAN SOAP/WCF para habilitacion conserva WS-Security con
  `Timestamp`, `BinarySecurityToken`, firma RSA-SHA256 de `wsa:To`,
  `wsse:Reference URI="#X509-..."` e `InclusiveNamespaces`; cualquier cambio en
  esa forma exige prueba real contra habilitacion y no puede basarse en
  simulacion.
- TrackId/ZipKey y `Batch en proceso de validacion` son recepcion inicial, no
  aceptacion legal. La habilitacion o produccion local solo puede activarse tras
  acuse final aceptado y conteo de minimos por empresa.
- Panama y Ecuador se manejan como configuraciones independientes por pais y
  licencia.
- El submenu de facturacion electronica permanece, pero las paginas internas se
  muestran segun pais detectado, licencia y permisos.
- No guardar certificados, tokens o claves en documentacion ni logs.

## Caja, turnos y carritos

- Caja y turno deben funcionar de manera independiente por usuario/caja dentro de
  una misma empresa.
- Varias cajas pueden operar simultaneamente si la configuracion/licencia lo
  permite.
- Los estados de estaciones deben sincronizarse para que otros usuarios vean
  cambios operativos.
- Venta directa y estaciones deben compartir el mismo carrito unificado y las
  mismas reglas de configuracion.
- Venta directa debe permitir abrir y salir de pantalla completa desde el propio
  carrito. Cuando se abre dentro del panel empresarial, el iframe principal debe
  permitir `fullscreen`.
- Abonos, descuentos, pagos mixtos y cierres deben reflejarse en el pago final y
  en el reporte de turno.
- En Colombia/COP, el carrito no debe mostrar ni aceptar centavos operativos:
  precios, abonos, pagos, devoluciones, QR y totales se normalizan a pesos
  enteros; monedas de otros paises pueden conservar precision decimal.

## Despliegue y portabilidad

- Docker/VPS es el camino operativo principal de despliegue.
- Los paquetes portables no deben incluir `.env`, backups, uploads privados,
  certificados, secretos ni datos runtime.
- `rs` y `sync_to_vps` son scripts operativos reales; revisar
  `documentos/comandos_codex.md` antes de ejecutarlos.

## Compatibilidad historica

- El proyecto aun no esta en produccion. No es obligatorio conservar rutas
  antiguas o paginas duplicadas si el usuario pide unificar o limpiar, salvo que
  exista una razon tecnica concreta para mantenerlas.
