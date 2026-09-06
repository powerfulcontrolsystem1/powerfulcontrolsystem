# Runbook: diagnóstico de checkout de licencias

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- El diagnóstico sigue la operación original y el fallback comprobado, con separación entre estado de pago, licencia y correo.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Síntomas y alcance

Pago pendiente, licencia sin activar, retorno fallido o notificación incorrecta.
Aplica a licencias, no al checkout de venta pública por empresa.

## Evidencia y diagnóstico

1. Identificar candidato/entorno, empresa, licencia y referencia de la operación original.
2. Consultar estado interno y proveedor con permisos autorizados; comparar referencia,
   monto, COP y ambiente. Guardar solo evidencia minimizada.
3. Si Epayco falla al abrir, comprobar el tipo devuelto: el fallback actual es
   `classic_js`, con `checkout_script_url`, `checkout_config` y `checkout_data`.
   No crear un formulario legacy ni exponer P_KEY al navegador.
4. En Wompi verificar que el formulario hospedado conserva firma de integridad,
   referencia y cantidad calculadas en servidor, con claves del mismo ambiente.
5. Comprobar si ya ocurrió activación y si solo falta el correo. Revisar marcas
   de notificación sin copiar raw_payload ni secretos al incidente.
6. Un esquema ausente requiere el migrador versionado dentro de una intervención
   autorizada; no ejecutar DDL desde un request ni cambiar el pago con SQL.

## Recuperación

Corregir primero la causa configurada. Ante timeout, consultar la referencia
original y el estado de idempotencia; no regenerar otro cobro a ciegas. Una nueva
transacción, cobro o reembolso requiere alcance autorizado. Reconciliar la
operación existente a través del flujo canónico; un correo pendiente no exige
nuevo pago. Un webhook inválido o un contexto cruzado debe continuar rechazado.

## Validación y escalamiento

Confirmar una activación por los periodos cobrados, misma empresa/licencia,
importe conciliado, estado monótono y notificación sin duplicado. Si hay desacuerdo
proveedor/registro, escalar a ingeniería de pagos con candidato, referencia y
secuencia de eventos; mantener pendiente hasta evidencia suficiente.

Contrato: [checkout de licencias](../contratos/contrato_checkout_licencias_publico.md).

## Fuentes y aceptación de la revisión

[payments_handlers.go](../../../backend/handlers/payments_handlers.go), [contrato_checkout_licencias_publico.md](../contratos/contrato_checkout_licencias_publico.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../requisitos/especificacion_y_trazabilidad.md)).
