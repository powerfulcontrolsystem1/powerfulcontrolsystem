# P110-006 — DNS corporativo, Mailu y BIMI

Fecha: 2026-08-12  
Ámbito: DNS público y VPS, solo lectura. No se enviaron correos ni se alteraron
registros DNS.

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

## Límite

Esta evidencia no prueba entrega de registro, invitación, reset, rebote ni
recepción externa. Tampoco resuelve el certificado BIMI comercial ni la prueba
DIAN oficial; P110-006 permanece parcial.
