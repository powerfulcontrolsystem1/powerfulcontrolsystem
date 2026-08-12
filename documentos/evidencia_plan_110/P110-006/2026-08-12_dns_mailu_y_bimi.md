# P110-006 — DNS corporativo, Mailu y BIMI

Fecha: 2026-08-12  
Ámbito: DNS público, VPS y prueba autenticada en PCS. No se alteraron registros
DNS.

## Resultado

- SPF publicado para el dominio corporativo, con remitente MX y host de correo.
- DMARC publicado con política `quarantine`, porcentaje 100 % y alineación
  estricta de SPF/DKIM.
- DKIM publicado bajo el selector operativo de Mailu.
- Los ocho contenedores Mailu activos estaban en estado `Up` y su almacenamiento
  administrativo era legible.
- BIMI publicado con el logo oficial, pero sin referencia a VMC/CMC en el campo
  `a`. Por ello el logo en el cuerpo del correo puede validarse, mientras que el
  avatar mostrado por proveedores externos no queda garantizado.
- El buzón de Powerful Control System fue reconciliado con Mailu sin registrar
  ni mostrar la contraseña en la evidencia. La autenticación automática abrió
  visualmente la bandeja real dentro del panel de empresa.
- Se envió un único correo real de prueba desde el buzón de soporte al buzón
  corporativo de PCS. SnappyMail mostró el mensaje en la bandeja de entrada, el
  remitente esperado, el asunto de prueba y el logo oficial del computador en
  el cuerpo del correo.
- El contador de no leídos del panel continuó rechazado porque consultaba
  `mailu-imap:143`, mientras que el canal autorizado de webmail
  `mailu-front:10143` sí autenticó. Se preparó la corrección del valor por
  defecto y una prueba Go focalizada.
- La pantalla superadministrador ocultaba el modo vigente `mailu_api` y lo
  rotulaba como manual. Se añadió la opción y el diagnóstico visual de API
  interna.
- Se enviaron y recibieron seis invitaciones reales para los roles QA del Plan
  110. Cada enlace abrió el registro oficial, exigió documento, contraseña,
  confirmación y aceptación del contrato, y terminó en una sesión válida de la
  empresa PCS.

## Límite

Esta evidencia prueba registro e invitación internos, pero no reset, rebote,
recepción externa ni firma DKIM observada por un proveedor externo. El correo interno
mostró `DKIM: none`, por lo que no se usa para certificar DKIM saliente. Tampoco
resuelve el certificado BIMI comercial ni la prueba DIAN oficial; P110-006
permanece parcial hasta publicar el ajuste y completar la matriz externa.
