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

## Repetición sobre el digest `cf49fc7c` 2026-07-30

La misma batería se repitió después de promover las cuatro imágenes exactas del
commit `cf49fc7cefb083e1ac8df1711f05a0f8a22c8afb`.

| Métrica | Resultado |
| --- | ---: |
| HTTP exitoso | 500 / 500 |
| Fallos / 5xx | 0 / 0 |
| Error rate | 0 % |
| p50 | 100 ms |
| p95 | 149 ms |
| p99 | 316 ms |
| Conexiones PostgreSQL antes/después | 24 / 26 |
| Sesiones esperando lock | 0 |

La memoria pasó de 14,61 a 17,41 MiB en API y de 136,7 a 162,8 MiB en
PostgreSQL; worker y frontend permanecieron estables. Al terminar,
`/health=ok`, `/ready=ready`, PostgreSQL negocio/super valía 1, el latido del
worker tenía 2,231 segundos y no había trabajo listo ni leases vencidos.

P108-019 conserva estado **parcial**: la lectura autenticada del digest actual
aprueba, pero no sustituye cuatro cajas, transacciones sostenidas, proveedores
externos ni backpressure.

## Repetición sobre `f9396da5`

Se ejecutaron 500 solicitudes GET autenticadas con concurrencia 10, alternando
las dos rutas canónicas de lectura `/me` y configuración operativa de la empresa
12. Una tentativa previa con una ruta obsoleta de carritos se descartó porque
respondía el 404 esperado y no representaba el contrato vigente.

| Métrica | Resultado |
| --- | ---: |
| HTTP exitoso | 500 / 500 |
| Fallos / 5xx | 0 / 0 |
| Error rate | 0 % |
| p50 | 100 ms |
| p95 | 134 ms |
| p99 | 419 ms |
| Concurrencia | 10 |
| Sesiones PostgreSQL observadas | 9 |
| Sesiones esperando lock | 0 |

Backend, worker, frontend y PostgreSQL permanecieron saludables; colas y leases
quedaron en cero. P108-019 continúa parcial por carga transaccional sostenida,
cuatro cajas, proveedores y backpressure.

## Repetición sobre `5566a213`

El mismo artefacto promovido a staging atendió 500 GET autenticados, alternando
`/me` y configuración operativa de PCS, sin mutaciones:

| Métrica | Resultado |
| --- | ---: |
| HTTP exitoso | 500 / 500 |
| Fallos / 5xx | 0 / 0 |
| Error rate | 0 % |
| p50 | 99 ms |
| p95 | 116 ms |
| p99 | 338 ms |
| Concurrencia | 10 |
| Sesiones PostgreSQL observadas | 25 |
| Esperas `Lock` | 0 |

Después de la carga, API, worker, PostgreSQL y frontend permanecieron
saludables; `/metrics` continuó privado con HTTP 404 desde el frontend. Esta
evidencia conserva P108-019 **parcial** por las pruebas transaccionales
sostenidas, cuatro cajas y backpressure.
