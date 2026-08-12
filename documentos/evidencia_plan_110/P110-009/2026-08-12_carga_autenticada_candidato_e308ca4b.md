# P110-009 — Carga autenticada del candidato e308ca4b

Fecha: 2026-08-12 (America/Bogota)  
Entorno: PCS/staging, lectura autenticada; no hubo mutaciones de negocio.

## Ejecución

Se iniciaron 300 solicitudes GET autenticadas con concurrencia 15 contra las
superficies CxP/IA, permisos, cuentas por pagar y facturación de la empresa 12.

| Métrica | Resultado |
| --- | ---: |
| Solicitudes | 300 |
| Errores HTTP | 0 |
| p50 | 623 ms |
| p95 | 907 ms |
| p99 | 1890 ms |

## Resultado

La carga de lectura aprobó el umbral técnico de 2,5 s y no generó ventas,
pagos, cuentas, facturas, documentos ni cambios de sesión permanentes.

## Límite

No es una prueba mutante de cuatro cajas ni sustituye medición de pool, locks,
colas, recursos, deduplicación de Alertmanager o simulacro de incidente.
P110-009 continúa **parcial** y el estado global se mantiene **NO-GO**.
