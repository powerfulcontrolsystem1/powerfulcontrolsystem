# P108-017 - Cabeceras y acceso no autenticado en staging

Fecha: 2026-07-28
Ambiente: staging autorizado
Alcance: comprobaciones HTTP de solo lectura, sin sesion y sin datos de negocio.

## Controles aprobados en esta pasada

- La ruta empresarial de reportes con `empresa_id` manipulado y sin sesion
  responde HTTP 401 con JSON estructurado; no entrega registros ni detalles de
  empresa.
- Un preflight `OPTIONS` originado desde un origen ajeno recibe HTTP 401 y no
  publica cabeceras `Access-Control-Allow-*`; el origen externo no queda
  autorizado por este ensayo.
- Las respuestas publicas y empresariales exponen HSTS, `nosniff`, política de
  referencia, política de permisos, `frame-ancestors 'self'` y
  `X-Frame-Options: SAMEORIGIN` en la capa final observada.

## Hallazgos pendientes

1. El candidato `bc794c89` dejó una única CSP, política de referencia y política
   de permisos en la respuesta final; en la API permanecieron dos valores de
   `X-Content-Type-Options` incluso al consultar directamente el contenedor
   backend. Se preparó una normalización en el proxy interno, pendiente de su
   candidato inmutable y comprobación final en staging.
2. La CSP final conserva `unsafe-inline` para `script-src` y `style-src`. No se
   considera aprobada la eliminacion o excepcion acotada hasta inventariar los
   scripts y estilos embebidos, migrarlos a nonces/hashes o documentar cada
   excepcion con responsable y fecha.
3. Esta pasada no sustituye DAST autenticado, XSS, SSRF, subida hostil, pruebas
   de sesion ni matriz A/B de tenant.

Estado de fase: **parcial; no aprobada**.
