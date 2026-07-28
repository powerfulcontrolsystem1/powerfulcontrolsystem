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
- El candidato `9586f80f` fue desplegado en staging y la prueba real
  `Probar envio` terminó con confirmación visible de entrega al buzón
  autorizado.
- El usuario confirmó la recepción en Gmail y que el logo oficial del
  computador PCS se muestra correctamente dentro del cuerpo del mensaje.
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
- desplegar y repetir la comprobación responsive del ajuste final
  `c06de9b1`;
- capturar el resultado recibido en Gmail y en otro cliente compatible;
- adquirir y publicar un certificado VMC o CMC. El registro BIMI actual tiene
  `a=` vacío y la prueba real conservó la inicial `P`;
- verificar el SHA/digest exacto servido por el VPS.

El logo dentro del correo está implementado. La sustitución garantizada de la
inicial en Gmail permanece bloqueada por el certificado externo. El
2026-07-28 se preparó, sin efectuar pago, un CMC de DigiCert para logo sin marca
registrada: suscripción de 12 meses por USD 1.416 antes de impuestos. La compra,
validación legal y emisión requieren intervención del titular de la empresa.
