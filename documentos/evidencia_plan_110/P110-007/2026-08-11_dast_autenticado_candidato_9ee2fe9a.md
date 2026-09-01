# P110-007 - DAST autenticado del candidato 9ee2fe9a

Fecha: 2026-08-11  
Ambiente: staging aislado `https://staging.powerfulcontrolsystem.com`  
Empresa: Powerful Control System (#12)  
Producción: no modificada.

## Candidato y resguardo

La pasada se realizó contra el código `9ee2fe9a96cb2b9f608759814de4eab6a5ff9c2c`
promovido exclusivamente por sus cuatro digests inmutables. Se utilizó una
sesión administrativa autorizada, sin guardar cookies, CSRF, credenciales ni
respuestas con datos empresariales. Las solicitudes mutantes usaron un objeto
vacío y fueron rechazadas antes de crear datos.

## Matriz observada

| Control | Resultado | Estado |
| --- | --- | --- |
| Lectura CxP anónima | HTTP 401 | PASS |
| Lectura financiera autenticada | HTTP 200 y límite por empresa | PASS |
| POST same-origin sin CSRF | HTTP 403 | PASS |
| POST con CSRF y origen externo | HTTP 403 | PASS |
| POST same-origin con cuerpo vacío | HTTP 400, sin escritura | PASS |
| `OPTIONS` desde origen externo | HTTP 401, sin CORS permisivo | PASS |
| Logout oficial | HTTP 200; sesión posterior HTTP 401 | PASS |

El control de sesión terminó por el endpoint oficial de logout. La operación
no ejecutó altas, pagos, ventas, documentos, usuarios, cajas ni facturación.

## Límite y estado

Esta pasada confirma las defensas dinámicas básicas sobre el candidato actual,
pero no sustituye el DAST integral de cargas/SSRF/XSS, el inventario y ajuste
de CSP, la matriz A/B por todos los dominios ni la entrega de alertas a un
canal externo. P110-007 continúa **parcial** y el veredicto global es
**NO-GO**.
