# P109-001 - Aislamiento A/B y concurrencia CxP en staging

Fecha: 2026-08-09

Empresa modificada: Powerful Control System (`empresa_id=12`).

Empresa comparada solo en lectura: empresa activa `empresa_id=33`.

Entorno: staging aislado del candidato `349712fb32d029b997109edd4a2b18a6bb299331`.

Producción: no modificada.

## Alcance y resguardo

Se usó una sesión de Super Administrador autorizada. La empresa comparada no
recibió altas, pagos, ediciones ni cambios de estado. Los identificadores de
sesión, CSRF, nombres de terceros, respuestas completas y claves de
idempotencia no se conservan en este documento.

## Matriz A/B de solo lectura

| Control | PCS (12) | Empresa comparada (33) | Resultado |
| --- | ---: | ---: | --- |
| Lista CxP canónica | 200, 3 registros iniciales | 200, 0 registros | Aislado |
| Detalle de una CxP de PCS consultado desde 33 | -- | 404 | Rechazado |
| Reconciliación de fuentes | 200 | 200 | Alcance por empresa |
| Proveedores disponibles para CxP | 200, 1 | 200, 0 | Aislado |
| Vista previa de recuperación outbox CxP | 200, 1 evento de PCS | 200, 0 eventos | Aislado |

Las formas de respuesta de CxP, proveedores y vista previa no expusieron
`payload_json`, claves privadas, tokens, secretos ni contraseñas. El evento de
PCS consultado con el `empresa_id` de la segunda empresa no se resolvió.

## Ensayo controlado de concurrencia

Por el endpoint oficial `POST /api/empresa/finanzas/cuentas_pagar` se creó en
PCS una obligación trazable `P109-CXP-CONC-20260809`, por COP 2. No se usó SQL
directo. El proveedor se validó en servidor como proveedor activo de PCS.

1. Dos `POST ...?action=registrar_pago` simultáneos por COP 1 con la misma
   clave de idempotencia devolvieron HTTP 200. Ambos devolvieron el mismo
   `pago_id` y el mismo `movimiento_finanzas_id`; una respuesta fue el registro
   original y la otra indicó `idempotent_replay=true`. El saldo quedó COP 1.
2. Dos solicitudes simultáneas posteriores por COP 1, con claves distintas,
   compitieron por ese último saldo: una devolvió HTTP 200 y la otra HTTP 409
   con el rechazo de saldo pendiente agotado. No se creó un sobrepago.
3. La consulta oficial final de detalle confirmó saldo COP 0, valor pagado COP
   2 y estado `pagada`.

La conciliación de fuentes posterior devolvió HTTP 200; presentó cuatro
registros canónicos y cero históricos. La vista previa del outbox CxP siguió
sanitizada (HTTP 200, sin payload crudo). La obligación de ensayo queda pagada
y auditada en staging; no se revierte porque sus pagos, movimiento y evento
outbox son trazabilidad financiera deliberada del ensayo.

## Veredicto

La carrera de idempotencia, el bloqueo contra sobrepago y el aislamiento A/B
de CxP/proveedores/outbox quedan **aprobados en staging** para el digest
indicado. P109-001 permanece **parcial**: falta conciliación y aceptación
firmada por contador, recuperación operativa de un evento elegible del mismo
candidato y cobertura A/B de los demás dominios P0/P1 antes de declararla
cerrada. El estado general continúa **NO-GO**.
