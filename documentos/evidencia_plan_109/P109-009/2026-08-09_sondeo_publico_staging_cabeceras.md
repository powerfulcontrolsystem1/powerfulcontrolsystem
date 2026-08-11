# P109-009 - Sondeo público de salud y cabeceras en staging

Fecha: 2026-08-09
Ambiente: staging público, sin sesión y sin mutaciones

## Resultado

El sondeo HTTP no autenticado obtuvo:

- `/health`: HTTP 200 con estado `ok`.
- `/ready`: HTTP 200 con estado `ready`.
- portada pública: HTTP 200.
- cabeceras presentes: HSTS, Content-Security-Policy, X-Content-Type-Options,
  X-Frame-Options, Referrer-Policy y Permissions-Policy.

La portada no envía `Cache-Control`; se mantiene como contenido público sujeto
a la política normal de caché. Esta observación no aplica a páginas o respuestas
autenticadas, que se validan por separado con `no-store`.

## Límite

El sondeo no sustituye DAST autenticado/hostil, matriz A/B ni revisión de todos
los recursos. P109-009 permanece **parcial**.
