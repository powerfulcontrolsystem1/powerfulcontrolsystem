# P108-019 - Carga autenticada ampliada en staging

Fecha: 2026-07-29  
Ambiente: staging aislado  
Candidato: `41be623ad2ed6c10ff86027063870b0848db2af1`  
Empresa autorizada: Powerful Control System (`empresa_id=12`)  
Modo: solo lectura autenticada.

## Escenario

Se inició sesión con la cuenta autorizada y se ejecutaron 500 solicitudes GET,
con concurrencia 10, alternando estas rutas de lectura:

- `/me`
- `/api/empresa/configuracion_operativa?empresa_id=12`

La ejecución no creó ventas, pagos, documentos, facturas, cierres de caja,
acciones de IA ni mutaciones de datos.

## Resultado

| Métrica | Resultado |
| --- | ---: |
| Solicitudes exitosas | 500 / 500 |
| Concurrencia | 10 |
| Fallos | 0 |
| Error rate | 0 % |
| p50 | 91 ms |
| p95 | 118 ms |
| p99 | 293 ms |
| Umbral p95 | 2.500 ms |
| Umbral error rate | 5 % |

Resultado: **OK** para este escenario autenticado de lectura.

## Límite de evidencia

La evidencia amplía el smoke anterior, pero P108-019 sigue **parcial y no
aprobada**: faltan duración sostenida acordada, carga transaccional controlada,
sesiones de cajero simultáneas, métricas de CPU/RAM/disco/conexiones/locks,
colas, backpressure y degradación de proveedores sobre el mismo candidato.
