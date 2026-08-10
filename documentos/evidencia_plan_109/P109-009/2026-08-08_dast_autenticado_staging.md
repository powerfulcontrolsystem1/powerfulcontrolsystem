# P109-009 - DAST autenticado seguro en staging

Fecha: 2026-08-08 00:29 America/Bogota
Entorno: `https://staging.powerfulcontrolsystem.com`
Empresa operativa: Powerful Control System (`empresa_id=12`)
Candidato validado: staging conservaba el candidato inmutable P109
`7c47d4df`; producción no fue desplegada ni modificada.

Una comprobación final de solo lectura en el VPS confirmó los digests cortos en
ejecución: backend `502f21dd3046`, frontend `368f73663ca0` y worker
`32eb94d27165`.

## Alcance y seguridad de la prueba

Se usó el acceso administrativo autorizado exclusivamente en staging. Las
solicitudes mutantes llevaron un cuerpo JSON vacío y se verificó que el
middleware las rechazara antes de entrar a la lógica de negocio; no se creó ni
editó carrito, venta, usuario, pago, caja, documento fiscal o registro CxP.
Las dos sesiones de prueba (navegador y HTTP) se cerraron mediante el endpoint
oficial de logout al finalizar. Esta evidencia no contiene credenciales, tokens
ni cookies.

## Resultado dinámico

| Control | Resultado observado | Resultado |
| --- | --- | --- |
| Login administrativo normal | HTTP 200 y panel Super Administrador visible | PASS |
| Lectura empresarial sin sesión | HTTP 401 | PASS |
| POST autenticado sin `X-CSRF-Token` | HTTP 403 | PASS |
| POST autenticado con `Origin` externo y token CSRF correcto | HTTP 403 | PASS |
| POST same-origin con token válido y `{}` | HTTP 400 de validación, sin escritura | PASS |
| Lectura empresarial autenticada | HTTP 200 | PASS |
| Límite por empresa | Cabeceras `X-Empresa-RateLimit-Limit: 600` y restante 598 | PASS |
| Preflight desde origen externo | HTTP 401, sin `Access-Control-Allow-Origin` | PASS |
| Logout oficial | HTTP 200; la misma sesión recibió HTTP 401 después | PASS |

La comprobación visual del panel autenticado no detectó errores de consola y el
área principal permaneció visible. El navegador se cerró tras el logout.

## Cabeceras verificadas

La respuesta de `login.html` devolvió HTTP 200 con HSTS (`max-age=15552000` e
`includeSubDomains`), CSP, `X-Content-Type-Options: nosniff`,
`X-Frame-Options: SAMEORIGIN`, `Referrer-Policy: strict-origin-when-cross-origin`,
`Permissions-Policy` restrictiva y `Cache-Control: no-store, must-revalidate,
no-cache, max-age=0`.

El inventario estático volvió a detectar 204 rutas empresariales y 204 wrappers
autoritativos, sin rutas que requieran revisión manual. La auditoría estática
`tools/security_audit.mjs` terminó en estado `ok` y confirmó cookies HttpOnly,
SameSite, Secure, revocación de sesión, allowlist pública, contrato de
reCAPTCHA/CORS y ausencia de archivos de secretos versionados.

## Riesgo abierto y siguiente paso

La CSP aún contiene `unsafe-inline` en `script-src` y `style-src`; permanece
abierto su inventario, reducción por nonce/hash o excepción explícita. Esta
pasada tampoco sustituye el ensayo A/B con dos identidades no globales y dos
empresas para SQL, archivos, caché, jobs, reportes, IA, exportaciones y
descargas, ni el DAST de subidas hostiles/SSRF/XSS completo. Por ello P109-009
continúa **parcial** y el veredicto de producción sigue siendo **NO-GO**.
