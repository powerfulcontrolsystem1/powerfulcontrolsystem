# P110-002 - CxP concurrente en el candidato 6bed2315

Fecha: 2026-08-11  
Ambiente: staging aislado, Powerful Control System (#12)  
Producción: no modificada.

## Corrección incluida

El flujo oficial de alta recibía un `proveedor_id` válido, pero rechazaba la
solicitud antes de que el servidor derivara `proveedor_nombre` desde ese
proveedor aislado por empresa. El candidato
`6bed231500524bf57a283ec0e1003b287a9ec084` corrige el contrato: el nombre se
deriva exclusivamente en servidor y nunca se acepta como identidad suministrada
por el cliente.

## Ensayo autenticado y controlado

Se creó por `POST /api/empresa/finanzas/cuentas_pagar` una obligación trazable
de COP 2 usando un proveedor activo de PCS. No se usó SQL directo. Los
identificadores, referencia, cookies, CSRF, proveedor y respuestas completas
no se conservan en esta evidencia.

| Control | Resultado | Estado |
| --- | --- | --- |
| Alta sin `proveedor_nombre` del cliente | HTTP 201 | PASS |
| Dos abonos simultáneos de COP 1, misma clave | HTTP 200 / 200, mismo pago | PASS |
| Dos abonos simultáneos de COP 1, claves distintas | HTTP 200 / 409 | PASS |
| Detalle posterior | HTTP 200, saldo COP 0, pagado COP 2, `pagada` | PASS |
| Cierre de sesión oficial | ejecutado | PASS |

La obligación de ensayo queda pagada y auditada en staging, junto con sus
movimientos y evento transaccional; no se revierte para no romper la
trazabilidad financiera del propio ensayo.

## Lecturas de conciliación y outbox

Después del ensayo, `reconciliacion_fuentes` respondió HTTP 200 y confirmó
alcance de empresa #12. La previsualización autenticada de recuperación del
topic CxP también respondió HTTP 200 para la misma empresa; mostró un evento
elegible sanitizado, sin `payload_json` ni material privado. No se reactivó
ningún evento: esa acción se reserva para una incidencia real con responsable
contable.

## Estado

P110-002 queda **parcial**. Aún exige conciliación contable independiente,
recuperación operativa de outbox elegible, diferencia cero en reportes y UAT
firmado por contador sobre el candidato congelado. El veredicto global sigue
siendo **NO-GO**.
