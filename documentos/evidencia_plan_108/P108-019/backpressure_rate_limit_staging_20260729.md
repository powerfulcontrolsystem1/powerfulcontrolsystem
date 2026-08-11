# P108-019 - Backpressure autenticado en staging

Fecha: 2026-07-29  
Ambiente: staging aislado  
Empresa autorizada: Powerful Control System (`empresa_id=12`)  
Modo: solo lectura autenticada.

## Escenario

Se ejecutaron cinco rondas consecutivas de 500 GET autenticados, concurrencia
10, alternando `/me` y configuración operativa. No se enviaron ventas, pagos,
facturas, IA, caja, documentos ni escrituras.

## Resultado

| Resultado | Valor |
| --- | ---: |
| Solicitudes realizadas | 2.500 |
| Rondas iniciales correctas | 4 / 5 |
| Solicitudes 200 | 2.390 |
| Respuestas 429 controladas | 110 |
| Respuestas 5xx | 0 |
| Peor p95 observado | 138 ms |
| Peor p99 observado | 304 ms |

Los 110 rechazos ocurrieron al alcanzar el límite configurado de 600 solicitudes
por empresa para la ruta API. El backend registró el evento como `rate_limit` y
respondió 429 en milisegundos; no se observó degradación a 5xx ni incremento de
latencia. Esto demuestra backpressure básico, no una caída del servicio.

## Límite

P108-019 permanece **parcial**. Faltan duración sostenida acordada, métricas de
CPU/RAM/disco/conexiones/locks/colas, tráfico transaccional seguro, proveedores
lentos y cuatro cajas simultáneas.
