# Contrato técnico: checkout público de licencias

Estado: Vigente. Responsable: Ingeniería de pagos y QA. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se corrigió el fallback real classic_js después de comprobar el handler y el consumidor HTML; se retiró la instrucción obsoleta de POST a checkout.php.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Alcance y rutas

Licencias administrativas por Epayco/Wompi. La venta pública empresarial tiene
su [contrato propio](contrato_venta_publica_empresarial_por_empresa.md).

| Entrada | Propósito |
| --- | --- |
| GET /api/public/licencias/payment_methods | Disponibilidad calculada desde configuración |
| GET /api/public/licencias/checkout_summary | Resumen calculado en servidor |
| POST /epayco/create_transaction | Smart Checkout o fallback estándar |
| POST /wompi/create_checkout | Web Checkout hospedado |
| POST /wompi/create_transaction_nequi | Compatibilidad con pago Nequi directo |
| GET /epayco/transaction_status y /wompi/transaction_status | Consulta y reconciliación del pago esperado |
| POST /epayco/webhook y /wompi/webhook | Confirmación validada del proveedor |

## Entrada, autoridad e idempotencia

`licencia_id` identifica el producto. `empresa_id`, cuando aplica, debe coincidir
con el contexto de la orden. Descuento, asesor, bundle y adicionales pasan por
las reglas comerciales del servidor. El navegador no decide importe, vigencia,
moneda ni estado aprobado. El flujo de creación usa el mecanismo durable
`beginPaymentCheckoutIdempotency`; ante resultado incierto se consulta la
operación original antes de generar otro cobro.

`cantidad` vale uno por defecto y respeta el máximo global (cinco inicial,
configurable entre uno y 24). Solo aplica a la licencia individual; bundles,
adicionales y licencias gratuitas/de prueba mantienen una unidad. `quantity`
y `duration_total_days` del resumen describen lo que se cobra y activa.

## Respuestas por proveedor

Epayco intenta Smart Checkout v2. Si no está listo o falla y el estándar está
configurado, devuelve `checkout_type=classic_js`, `checkout_script_url`,
`checkout_config` y `checkout_data`. El navegador carga el script permitido y
abre el checkout con esos objetos. La ruta actual no devuelve `classic_form`:
el helper heredado de formulario no demuestra uso en la ruta activa. `P_KEY`
permanece en backend para validar la confirmación; no se publica en el formulario.
Las credenciales API de Smart Checkout no son la contraseña de la cuenta ni
la clave clásica de confirmación.

Wompi devuelve formulario GET hospedado, referencia única, COP, importe en
centavos y firma de integridad calculada en servidor. El estado del navegador
no activa una licencia. Llaves, modo y consulta deben corresponder al mismo
ambiente. No usar el contrato Nequi de venta pública como especificación de
este checkout general.

## Estados, efectos y fallos

`PENDING`, `APPROVED`, `DECLINED` y `ERROR` describen el pago; aprobación es
monótona frente a callbacks tardíos. La confirmación valida firma, referencia,
importe, moneda y ambiente contra el registro de `pagos_epayco`/`pagos_wompi`.
Webhook y consulta pueden reconciliar licencia/notificación, por lo que la
consulta de estado no es una lectura sin efectos locales.

Un error de configuración puede devolver 409; un pago inexistente 404 y un
contexto empresarial/licencia incompatible 409. No convertir un timeout o
error externo en aprobación. El correo fallido se reconcilia sin cobrar otra vez.
Las URLs de retorno provienen de la configuración canónica, no de Host/Origin
aportados por el cliente. El esquema se prepara mediante `pcs-migrate`.

## Aceptación

Validar creación y replay, callback duplicado y tardío, firmas inválidas,
importes/ambientes cruzados, empresa/licencia ajena, cantidad y correo pendiente.
Guardar candidato, entorno, referencia minimizada y resultado; no payloads
privados. La [matriz de pagos](contrato_matriz_pagos_reales.md) define la
aceptación externa y el [runbook](../runbooks/runbook_checkout_licencias.md)
orienta diagnóstico. Los resultados anteriores permanecen en el
[antecedente](../../historico/2026-09-05/checkout_licencias_antes_revision.md).

## Fuentes y aceptación de la revisión

[payments_handlers.go](../../../backend/handlers/payments_handlers.go), [pagar_licencia.html](../../../web/pagar_licencia.html).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../requisitos/especificacion_y_trazabilidad.md)).
