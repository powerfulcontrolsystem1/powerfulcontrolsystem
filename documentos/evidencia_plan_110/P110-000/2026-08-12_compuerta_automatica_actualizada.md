# P110-000 — Compuerta automática actualizada

Fecha: 2026-08-12 (America/Bogota)  
Entorno: staging; ejecución de solo lectura.

La compuerta P110 aprobó `/health`, `/ready`, paridad DIAN de PCS y existencia
de entrega externa en Alertmanager. La configuración DIAN de staging conserva
habilitación y emisión local desactivada. No se imprimieron secretos ni se
modificó producción.

Alertmanager mantenía una alerta activa `PCSAntivirusSoportesDetectoMalware`,
esperada por la sonda EICAR controlada; ClamAV y staging ya estaban saludables.
La alerta queda en ventana temporal de Prometheus y no se suprime manualmente.

La aprobación automática no equivale a GO: siguen pendientes roles/cajas,
UAT, impresión física, restore/rollback integral y piloto.
