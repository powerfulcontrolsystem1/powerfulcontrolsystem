# P108-019 - Smoke autenticado de capacidad en staging

Fecha: 2026-07-29  
Ambiente: staging aislado  
Empresa autorizada: Powerful Control System (`empresa_id=12`)  
Modo: solo lectura autenticada.

## Escenario

Se validó la sesión autorizada y se ejecutaron 30 solicitudes GET con
concurrencia 5, alternando `/me` y
`/api/empresa/configuracion_operativa?empresa_id=12`. Ambas rutas son de
lectura; no se enviaron ventas, pagos, facturas, cierres, IA, documentos ni
operaciones de caja.

## Resultado

| Métrica | Resultado |
| --- | ---: |
| HTTP exitoso | 30 / 30 |
| Fallos | 0 |
| Error rate | 0 % |
| p50 | 101 ms |
| p95 | 517 ms |
| p99 | 523 ms |

La pasada quedó dentro del umbral operativo de p95 de 2.5 s y de 5 % de
errores definido para el smoke.

## Límite

El resultado confirma capacidad autenticada mínima sobre dos rutas GET. No
certifica aún sesiones de cajero simultáneas, transacciones, locks, colas,
proveedores externos, backpressure, CPU/RAM/disco ni carga sostenida.

Estado de fase: **parcial; no aprobada**.

## Carga ampliada del candidato 2026-07-30

Se ejecutaron 500 solicitudes GET autenticadas con concurrencia 10, alternando
`/me`, configuración operativa y listado de carritos de `empresa_id=12`.

| Métrica | Resultado |
| --- | ---: |
| HTTP exitoso | 500 / 500 |
| Fallos / 5xx | 0 / 0 |
| Error rate | 0 % |
| p50 | 110 ms |
| p95 | 152 ms |
| p99 | 323 ms |

API, worker, PostgreSQL y frontend siguieron saludables. PostgreSQL pasó de 10
a 27 conexiones; 21 quedaron `idle`, dentro de los pools acotados del runtime,
sin sesiones bloqueadas ni errores 5xx observados. Esta evidencia aprueba la
carga autenticada ampliada de lectura, pero P108-019 sigue parcial hasta cubrir
transacciones sostenidas, cuatro cajas, locks, colas y proveedores externos.
