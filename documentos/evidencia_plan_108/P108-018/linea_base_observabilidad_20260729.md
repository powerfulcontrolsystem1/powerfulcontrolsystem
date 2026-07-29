# P108-018 - Línea base de observabilidad en staging

Fecha: 2026-07-29  
Ambiente: staging autorizado

## Señales disponibles

- PostgreSQL, backend, worker y frontend saludables;
- worker responde `ok` y `ready` en su endpoint de loopback;
- sin locks esperando; 10 conexiones PostgreSQL, una no idle;
- uso de disco del VPS: 47 %;
- existen reglas versionadas para disco, memoria y backend.

## Bloqueo

No hay contenedores activos de Prometheus, Grafana, Alertmanager, node-exporter
ni cAdvisor. Por ello no se puede demostrar captura histórica, tableros, entrega
de alertas ni simulacros medibles. P108-018 permanece **parcial y no aprobada**.
