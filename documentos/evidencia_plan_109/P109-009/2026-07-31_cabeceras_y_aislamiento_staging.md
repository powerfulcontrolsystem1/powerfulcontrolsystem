# P109-009 - Cabeceras y aislamiento básico de staging

Fecha: 2026-07-31
Entorno: staging, solicitudes no autenticadas y sin efectos.

## Resultado

- La portada respondió `200` con HSTS (`max-age=15552000; includeSubDomains`),
  CSP y `X-Content-Type-Options: nosniff`.
- Una consulta empresarial sin sesión respondió `401 Unauthorized`.
- Un preflight desde un origen no autorizado respondió `401` y no publicó
  `Access-Control-Allow-Origin` ni `Access-Control-Allow-Credentials`.

## Límite y riesgo abierto

La cabecera CSP aún contiene `unsafe-inline` en `style-src` y `script-src`.
Esta prueba no es DAST autenticado ni cubre XSS, CSRF, SSRF, IDOR, cargas,
sesiones, rate limits o la matriz A/B completa. P109-009 continúa parcial y el
riesgo CSP no está cerrado para el GO.

## Repetición sobre `eb853788` (2026-08-01)

Después de la promoción inmutable se repitió la prueba externa: portada 200,
HSTS `max-age=15552000; includeSubDomains`, CSP presente, `nosniff`, endpoint
empresarial sin sesión 401 y preflight de origen ajeno 401 sin publicar
`Access-Control-Allow-Origin` ni credenciales. El resultado conserva el cierre
básico, pero no sustituye el DAST autenticado ni elimina el riesgo CSP descrito.
