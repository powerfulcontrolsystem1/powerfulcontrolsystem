# Contrato tecnico: checkout publico de licencias

Fecha: 2026-04-18
Estado: vigente

## Alcance

Este contrato cubre el flujo publico de checkout de licencias administrativas por Epayco y Wompi, incluyendo creacion de transaccion, retorno, polling, webhook, activacion de licencia y correo de activacion.

## Rutas implicadas

- `GET /api/public/licencias/payment_methods`
- `POST /epayco/create_transaction`
- `GET /epayco/transaction_status`
- `POST /epayco/webhook`
- `POST /wompi/create_transaction_nequi`
- `GET /wompi/transaction_status`
- `POST /wompi/webhook`
- frontend de apoyo: `web/pagar_licencia.html`, `web/epayco/respuesta.html`, `web/epayco/pago_exitoso.html`

## Entradas obligatorias

### Creacion de transaccion

- `licencia_id`: obligatorio.
- `empresa_id`: obligatorio cuando el checkout esta asociado a una empresa concreta.

### Polling de estado

- `id` o `reference`: al menos uno para identificar el pago.
- `licencia_id`: obligatorio cuando el frontend conoce el checkout esperado.
- `empresa_id`: obligatorio cuando el frontend conoce la empresa esperada.

## Entradas opcionales

- `customer_email`
- `discount_code`
- `asesor_id` o `vendedor_id`

## Salidas y estados

Estados funcionales esperados:

- `PENDING`
- `APPROVED`
- `DECLINED`
- `ERROR`

Errores de contrato esperados:

- `400` por parametros insuficientes o invalidos.
- `404` si el pago no existe.
- `409` si el pago resuelto no coincide con la `empresa_id` o `licencia_id` esperadas.

## Invariantes

1. Un pago no puede activar ni confirmar una licencia de otra empresa.
2. El frontend debe reenviar `empresa_id` y `licencia_id` cuando ya conoce el checkout esperado.
3. El backend debe validar ese contexto esperado antes de dar por bueno el polling.
4. El correo de activacion debe resolver la empresa por alcance logico, no solo por el id fisico de la fila.
5. Webhook y polling deben ser idempotentes respecto a activacion de licencia y envio de correo.

## Side effects

- persistencia en `pagos_epayco` o `pagos_wompi`
- activacion o actualizacion de licencia
- escritura de marcas de correo en `raw_payload`
- envio de correo de activacion cuando corresponda

## Reglas de compatibilidad

- los helpers de pagos deben tolerar saneamiento de esquema cuando falten tablas o columnas de la capa de pagos.
- el flujo debe mantenerse compatible con PostgreSQL como runtime canonico.

## Evidencia tecnica minima

- pruebas focalizadas del handler de pagos del checkout.
- diagnostico del editor limpio en frontend y backend tocados.
- validacion de contexto esperado en polling cuando el cambio afecte empresa/licencia.

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`

## Runbook relacionado

- `documentos/gobernanza_tecnica/runbooks/runbook_checkout_licencias.md`
