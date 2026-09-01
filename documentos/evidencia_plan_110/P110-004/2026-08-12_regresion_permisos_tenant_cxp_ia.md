# P110-004 — Regresión enfocada de permisos, tenant, CxP e IA

Fecha: 2026-08-12  
Ámbito: candidato local P110. Sin mutaciones de datos empresariales.

## Ejecución

La primera batería amplia de paquetes `handlers` y `db` excedió el tiempo de
ejecución disponible sin dar un resultado, por lo que se dividió en conjuntos
deterministas que sí terminaron.

| Conjunto | Resultado |
| --- | --- |
| Manipulación query/header/JSON/form/multipart de `empresa_id` y documento dinámico | PASS |
| Límites contador/cajero, empresa compartida e IA agente cerrada | PASS |
| CxP: esquema atómico, idempotencia de abonos/pagos, ledger y bloqueo legado | PASS |
| IA: migración de aislamiento por usuario y rechazo de marcadores de credenciales | PASS |
| Impresoras/agentes: dispositivo aislado por empresa | PASS |
| Firma DIAN, requerimientos de TestSet, SOAP y correo corporativo sin fuga | PASS |

## Resultado y límite

Estos contratos prueban la frontera de empresa y controles de servidor antes de
acceder a la base en los casos cubiertos. No prueban por sí solos cada pantalla
ni reemplazan la matriz autenticada de mutaciones con identidades reales de
cajero, contador, vendedor y soporte. P110-004 y P110-007 permanecen parciales.
