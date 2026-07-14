# Contrato tecnico: checkout publico de licencias

Fecha: 2026-04-30
Estado: vigente

## Alcance

Este contrato cubre el flujo publico de checkout de licencias administrativas por Epayco y Wompi, incluyendo creacion de transaccion, retorno, polling, webhook, activacion de licencia y correo de activacion.

Para Epayco, el flujo preferido es Smart Checkout v2. Si Smart Checkout no devuelve token y la configuracion tiene `epayco.customer_id` mas `epayco.checkout_key`/`epayco.p_key`, el backend puede entregar un formulario clasico firmado para POST directo a `https://secure.payco.co/checkout.php`.

Las credenciales no son intercambiables: Smart Checkout v2 usa `PUBLIC_KEY` + `PRIVATE_KEY` API de Apify, mientras que el checkout estandar usa `P_CUST_ID_CLIENTE` + `P_KEY`. El backend no debe tratar una `P_KEY` clasica como `private_key` API.

Para Wompi en pago de licencias, el flujo preferido es Web Checkout hospedado mediante formulario GET a `https://checkout.wompi.co/p/`, con referencia unica, monto en centavos, moneda `COP` y `signature:integrity` SHA256 construida como `reference + amount_in_cents + currency + integrity_key`. El flujo directo Nequi se conserva como compatibilidad, pero la pantalla publica de licencia debe presentar Wompi como checkout general para que el cliente elija Nequi, PSE o tarjeta dentro de Wompi.

## Rutas implicadas

- `GET /api/public/licencias/payment_methods`
- `POST /epayco/create_transaction`
- `GET /epayco/transaction_status`
- `POST /epayco/webhook`
- `POST /wompi/create_checkout`
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
- `asesor_id` (codigo de asesor comercial)

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
- `409` si Smart Checkout de Epayco falla y el fallback clasico no puede construirse por falta de `epayco.customer_id` o `epayco.checkout_key`/`epayco.p_key`.

### Respuesta Epayco con fallback clasico

Cuando aplica fallback clasico, `POST /epayco/create_transaction` puede responder:

- `checkout_type: "classic_form"`
- `checkout_url: "https://secure.payco.co/checkout.php"`
- `checkout_form.action`: URL segura de Epayco para POST.
- `checkout_form.method`: `POST`.
- `checkout_form.fields`: campos firmados requeridos por Epayco.
- `mode`: modo efectivo del fallback clasico.
- `mode_source`: origen del modo clasico; para credenciales reales debe ser `classic_credentials`.

El frontend debe crear un formulario temporal y enviarlo por POST. No debe abrir esa URL con GET.

## Invariantes

1. Un pago no puede activar ni confirmar una licencia de otra empresa.
2. El frontend debe reenviar `empresa_id` y `licencia_id` cuando ya conoce el checkout esperado.
3. El backend debe validar ese contexto esperado antes de dar por bueno el polling.
4. El correo de activacion debe resolver la empresa por alcance logico, no solo por el id fisico de la fila.
5. Webhook y polling deben ser idempotentes respecto a activacion de licencia y envio de correo.
6. El fallback clasico de Epayco debe usar POST firmado a `https://secure.payco.co/checkout.php`.
7. El sistema no debe redirigir al usuario por GET a `https://checkout.epayco.co/checkout.php`, porque ese flujo puede devolver XML `AccessDenied`.
8. Los logs o `raw_payload` no deben persistir `p_key` sin enmascarar.
9. El modo del fallback clasico debe resolverse con las credenciales clasicas (`customer_id` + `P_KEY`) y no heredar automaticamente el modo de Smart Checkout; en produccion debe emitir `p_test_request=FALSE` y en pruebas `p_test_request=TRUE`.
10. Smart Checkout solo se considera listo cuando existe `public_key` y una `private_key` API valida. El checkout estandar solo se considera listo cuando existe `customer_id` y `checkout_key`/`p_key`.
11. Wompi Web Checkout solo se considera listo cuando existen `wompi.public_key` y `wompi.integrity_key`; `wompi.private_key` queda reservada para API directa o consultas que la requieran.
12. El frontend no debe pedir celular Nequi en el checkout publico de licencia; Wompi debe abrirse como checkout hospedado para que el cliente seleccione el medio de pago.
13. Una pasarela de licencias con credenciales completas queda habilitada por defecto; `epayco.enabled=0` o `wompi.enabled=0` solo actuan como apagado explicito desde super administrador.

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
- prueba del formulario clasico firmado y saneamiento de `p_key` cuando el cambio afecte Epayco.

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`

## Runbook relacionado

- `documentos/gobernanza_tecnica/runbooks/runbook_checkout_licencias.md`
# Cantidad de períodos

Los endpoints `GET /api/public/licencias/checkout_summary`,
`POST /wompi/create_checkout`, `POST /epayco/create_transaction` y
`POST /licencias/activar_sin_pago` aceptan `cantidad` opcional. Si falta, su
valor es `1`. La respuesta `summary` expone `quantity` y
`duration_total_days`; el backend rechaza valores por encima del límite global,
cantidades para bundles/adicionales y más de una unidad gratuita.
