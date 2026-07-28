# Resultados P108-015

Estado: **parcial / NO-GO**

## Aprobado

- SPF publicado para el dominio propio.
- DKIM publicado con selector `dkim`.
- DMARC en `p=quarantine`, `sp=quarantine`, `pct=100` y alineación estricta.
- BIMI publicado en `default._bimi.powerfulcontrolsystem.com`.
- El SVG público responde HTTP 200 como `image/svg+xml`, es cuadrado, pesa
  menos de 32 KiB y declara SVG Tiny-PS.
- El PNG oficial `web/img/Logo pcs 1.png` responde HTTP 200 como `image/png`.
- El constructor común produce `multipart/related`, referencia `cid:` y adjunta
  el logo local.
- La página super carga autenticada y muestra el logo configurado
  `/img/Logo pcs 1.png`.
- Suite Go completa y preflight profesional local en verde.

## Hallazgo corregido localmente

La prueba visual autenticada reprodujo HTTP 403 al pulsar `Probar envio`. La
página cargaba por GET, pero sus mutaciones no enviaban `X-CSRF-Token`. La rama
candidata agrega el token a POST/PUT/PATCH/DELETE y cubre el contrato con test.

Al desplegar el primer candidato en staging se detectó además que la cuota
predeterminada `1024 MB` no era válida para el `step=50` del formulario. El
navegador bloqueaba silenciosamente el submit y por eso el módulo no quedaba
activo. La cuota ahora acepta pasos de 1 MB y conserva 1024 como valor válido.

La captura inicial a 390 x 844 detectó que la tabla con ancho mínimo ampliaba
el item del grid y generaba scroll horizontal de toda la página. Las tarjetas
ahora declaran `min-width:0`, el contenedor de tabla limita su ancho y la página
reserva espacio inferior para las acciones flotantes.

## Pendiente para aprobar

- integrar y desplegar la rama candidata;
- repetir `Probar envio` al buzón autorizado;
- comprobar en el correo recibido el logo inline, remitente y autenticación;
- capturar el resultado en Gmail y en otro cliente compatible;
- adquirir y publicar un certificado VMC o CMC. El registro BIMI actual tiene
  `a=` vacío, por lo que Gmail puede conservar la inicial `P`;
- verificar el SHA/digest exacto servido por el VPS.

El logo dentro del correo está implementado. La sustitución garantizada de la
inicial en Gmail permanece bloqueada por el certificado externo.
