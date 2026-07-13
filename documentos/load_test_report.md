# Reporte de carga y escalado

Estado: **PENDIENTE DE STAGING**.

No se generó trafico contra servicios reales. La prueba requerida incluye
login, carrito, cierre de venta, consulta de facturas, archivos privados,
webhooks firmados y consultas por empresa. Debe registrar concurrencia,
latencia p50/p95/p99, errores, uso de PostgreSQL/Redis, CPU, memoria y limites
de rate limiting.

No se habilita escalado horizontal ni se fijan objetivos de capacidad hasta
contar con esa evidencia.
