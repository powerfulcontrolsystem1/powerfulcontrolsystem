# P110-009 — carga autenticada de solo lectura

Fecha: 2026-08-12  
Ámbito: PCS/staging, identidad autorizada, sin mutaciones de negocio.

## Ejecución válida

Se ejecutaron 240 solicitudes GET autenticadas, con concurrencia 12, sobre
CxP, conciliación contable, bandeja de facturación y contexto de permisos de
la empresa PCS.

| Métrica | Resultado |
| --- | ---: |
| p50 | 111 ms |
| p95 | 306 ms |
| p99 | 674 ms |
| Errores HTTP | 0 |
| Tasa de error | 0 % |

La primera corrida se invalidó porque una ruta de prueba codificó los
separadores de query y obtuvo 400; se corrigió la configuración del ensayo y
no se usa como evidencia de rendimiento del producto.

## Límite

P110-009 queda **parcial**. Faltan cuatro cajas mutantes, carga sostenida,
recursos de VPS/pool/colas, alerta externa y simulacro de incidente.
