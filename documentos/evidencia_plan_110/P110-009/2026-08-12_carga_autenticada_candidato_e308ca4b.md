# P110-009 — Carga autenticada del candidato e308ca4b

Fecha: 2026-08-12 (America/Bogota)  
Entorno: PCS/staging, lectura autenticada; no hubo mutaciones de negocio.

## Ejecución

Se repitió la carga contra las superficies CxP/IA, permisos efectivos, cuentas
por pagar, contabilidad y facturación de la empresa 12. La primera corrida de
500 detectó una ruta de permisos mal escrita (`404`), por lo que se descartó:
no fue una prueba de capacidad válida. Se corrigió al endpoint público real
`permisos_contexto` y se ejecutó nuevamente sin mutaciones de negocio.

| Métrica | Resultado |
| --- | ---: |
| Solicitudes | 500 |
| Concurrencia | 20 |
| Errores HTTP | 0 |
| Códigos | 500 × `200` |
| p50 | 180 ms |
| p95 | 424 ms |
| p99 | 882 ms |

## Resultado

La carga de lectura aprobó el umbral técnico de 2,5 s y no generó ventas,
pagos, cuentas, facturas, documentos ni cambios de sesión permanentes.
El smoke ahora agrega el conteo completo por código HTTP para que futuros
ensayos distingan un límite de tasa de un error funcional, sin registrar
cookies ni credenciales.

## Límite

No es una prueba mutante de cuatro cajas ni sustituye medición de pool, locks,
colas, recursos, deduplicación de Alertmanager o simulacro de incidente.
P110-009 continúa **parcial** y el estado global se mantiene **NO-GO**.
