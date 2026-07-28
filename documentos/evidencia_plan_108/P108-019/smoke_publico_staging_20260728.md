# P108-019 - Smoke de capacidad pública en staging

Fecha: 2026-07-28
Ambiente: staging autorizado
Alcance: rutas públicas de solo lectura; no incluye sesión, ventas, pagos,
facturas, documentos ni acciones IA.

## Resultado desde el VPS

Se ejecutaron 30 solicitudes secuenciales desde el VPS hacia el dominio público
de staging, rotando inicio, inicio de sesión y pantallas públicas de licencia.

- HTTP 200: 30 de 30
- Errores HTTP/red desde el VPS: 0
- p95 observado: 33 ms

La prueba confirma disponibilidad básica del borde y del frontend para esas
rutas en el momento de la medición.

## Resultado desde el PC de pruebas

La ejecución inicial de `load_smoke_test.mjs` con 30 solicitudes y concurrencia
5 observó p95 de 560 ms, pero 13 errores `fetch failed`. Un contraste con curl
mostró una falla local de resolución DNS para una solicitud, mientras el VPS
completó la misma familia de rutas sin error. Por tanto, la pasada Node no se
atribuye al servicio ni se usa como métrica de capacidad.

## Límite de evidencia

Esta es una señal pública mínima, no una certificación de capacidad. Permanecen
pendientes carga autenticada sostenida, p50/p95/p99, CPU, RAM, conexiones,
locks, disco, colas, degradación de proveedores, backpressure y cuatro cajas
concurrentes sobre el mismo digest inmutable.

Estado de fase: **parcial; no aprobada**.
