# P108-017 - Seguridad dinámica no autenticada en staging

Fecha: 2026-07-29  
Ambiente: staging con TLS válido  
Modo: solicitudes externas de solo lectura, sin sesión.

## Casos ejecutados

| Caso | Resultado |
| --- | --- |
| Reportes empresariales con `empresa_id` inexistente/manipulado | HTTP 401 JSON; no se entregaron datos ni detalles de empresa. |
| `OPTIONS` con origen externo no autorizado | HTTP 401; no se emitieron cabeceras `Access-Control-Allow-*`. |
| `GET /metrics` | HTTP 200, texto Prometheus mínimo y `Cache-Control: no-store`. |
| TLS, `/health` y `/ready` | Certificado válido de staging y respuestas 200. |

Las cabeceras finales observadas incluyen CSP, HSTS, `nosniff`, política de
referencia, política de permisos y `X-Frame-Options`. La prueba no utilizó
cookies, credenciales ni datos empresariales.

## Límite

P108-017 permanece **parcial**: falta DAST autenticado, XSS/SSRF, archivos
hostiles, sesiones, matriz A/B de tenant y eliminación o justificación acotada
de `unsafe-inline` en CSP.
