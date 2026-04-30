# Runbook: checkout de licencias no activa o envia correo incorrecto

Fecha: 2026-04-30
Estado: vigente

## Sintomas cubiertos

- el pago aparece aprobado pero la licencia no queda activa.
- el correo de activacion llega con empresa incorrecta o con el texto `tu empresa`.
- el polling de estado devuelve conflicto o no encuentra la referencia esperada.
- el retorno desde la pasarela no reanuda correctamente el checkout.
- al presionar "Pagar con Epayco" aparece `failed to authenticate with Epayco Smart Checkout: epayco login did not return token`.
- despues de redireccionar a Epayco aparece una pagina XML con `AccessDenied`.
- Epayco abre su pantalla pero muestra `El comercio no fue reconocido`.

## Alcance

Aplica al checkout publico de licencias sobre Epayco y Wompi.

## Fuentes de evidencia

- logs de backend
- tablas de pagos (`pagos_epayco`, `pagos_wompi`)
- tabla o estado de licencias
- `raw_payload` del pago
- parametros devueltos a `pagar_licencia.html`
- configuracion super de Epayco: `epayco.public_key`, `epayco.private_key`, `epayco.customer_id`, `epayco.checkout_key`/`epayco.p_key`, modo pruebas/produccion y URLs de respuesta/confirmacion.

## Verificaciones iniciales

1. confirmar `empresa_id` y `licencia_id` que el frontend tenia abiertos.
2. verificar `transaction_id` o `reference` de la pasarela.
3. revisar si el pago encontrado pertenece a la misma empresa/licencia esperada.
4. comprobar si el correo del comprador vino en payload principal o dentro de `data.customer_email`.
5. revisar si `raw_payload` ya tiene marcas de correo enviado para evitar falsos duplicados.
6. si Smart Checkout fallo, verificar que el backend haya respondido `checkout_type=classic_form` y `checkout_form.action=https://secure.payco.co/checkout.php`.
7. confirmar que el frontend hizo POST al formulario clasico y no GET a una URL legacy de `checkout.epayco.co/checkout.php`.
8. si el comercio no es reconocido, revisar en la respuesta de `/epayco/create_transaction` que `mode=production`, `mode_source=classic_credentials` y `checkout_form.fields.p_test_request=FALSE` cuando se usan credenciales reales.
9. confirmar que `epayco.private_key` sea la llave API de Smart Checkout que inicia por `prv_`. La `P_KEY` del checkout estandar debe guardarse en `epayco.checkout_key` o `epayco.p_key`, no en `epayco.private_key`.
10. confirmar que `epayco.checkout_key`/`epayco.p_key` sea la P_KEY real del dashboard de Epayco, no la contrasena de acceso a la cuenta. Si el backend reporta `checkout_key_format_valid=false`, el checkout clasico no debe habilitarse.

## Causas probables

- referencia vieja o cruzada entre empresas.
- polling sin `empresa_id` y `licencia_id` esperados.
- empresa resuelta por id fisico y no por alcance logico.
- esquema de pagos incompleto o legacy al primer acceso.
- retorno frontend con contexto incompleto al volver de la pasarela.
- credenciales de Smart Checkout incompletas o no aceptadas por Epayco.
- falta `epayco.customer_id` o `epayco.checkout_key`/`epayco.p_key`, lo que impide construir el fallback clasico firmado.
- frontend antiguo intentando abrir por GET la URL de checkout clasico, lo que produce XML `AccessDenied`.
- modo de pruebas aplicado por error al fallback clasico con credenciales reales, lo que puede hacer que Epayco no reconozca el comercio.
- mezcla de llaves: usar `P_KEY` del checkout estandar como `private_key` API de Smart Checkout causa fallo de autenticacion contra Apify; usar `PUBLIC_KEY`/`PRIVATE_KEY` API como credenciales clasicas causa comercio no reconocido.
- P_KEY incorrecta o con forma de contrasena: Epayco abre su pantalla pero no reconoce el comercio. Desde 2026-04-30 el backend bloquea localmente estas llaves para evitar enviar usuarios a una pasarela que ya va a fallar.

## Acciones de recuperacion

1. validar que el frontend este reenviando `empresa_id` y `licencia_id` al consultar `transaction_status`.
2. verificar que el backend compare contexto esperado contra el pago resuelto antes de activarlo.
3. revisar que la empresa se resuelva por alcance logico cuando se construye el correo.
4. si hay errores de tabla o columna faltante en pagos, regularizar el esquema y repetir la consulta de estado.
5. si webhook activó primero y el correo quedo pendiente, confirmar que el polling posterior pueda completar solo la notificacion faltante.
6. si aparece XML `AccessDenied`, actualizar `web/pagar_licencia.html` y backend para usar `checkout_form` por POST a `https://secure.payco.co/checkout.php`.
7. si `/epayco/create_transaction` devuelve `409` por autenticacion de Smart Checkout y falta de fallback, registrar `epayco.customer_id` y `epayco.checkout_key`/`epayco.p_key` en la configuracion super.
8. si Epayco muestra `El comercio no fue reconocido`, confirmar que el fallback clasico no herede el modo Smart Checkout y que el campo `p_test_request` salga en `FALSE` para produccion.
9. separar credenciales por tipo de integracion:
   - Smart Checkout v2: `epayco.public_key` + `epayco.private_key` API.
   - Checkout estandar: `epayco.customer_id` + `epayco.checkout_key`/`epayco.p_key`.
10. si el backend marca `checkout_key_format_valid=false`, reemplazar la llave guardada por la P_KEY real del dashboard de Epayco en `Configuracion > Personalizaciones > Llaves secretas`.
11. revisar `raw_payload` solo como evidencia saneada; nunca copiar `p_key` real a tickets o documentos.

## Validacion posterior

- la licencia queda activa en la empresa correcta.
- el correo muestra la empresa correcta.
- no se duplica el correo en webhooks o polls posteriores.
- el frontend puede cerrar el checkout en la pagina de exito sin perder contexto.
- Epayco abre desde formulario POST seguro cuando se usa fallback clasico.
- no vuelve a mostrarse XML `AccessDenied` por apertura GET de checkout clasico.
- en credenciales reales, el fallback clasico queda en modo produccion y no como solicitud de pruebas.
- con credenciales incompletas o P_KEY invalida, el sistema falla localmente como configuracion incompleta y no redirige a Epayco.
- los webhooks con `x_signature` se validan con SHA256 usando `p_cust_id_cliente^p_key^x_ref_payco^x_transaction_id^x_amount^x_currency_code`; si la firma es invalida, se rechaza la confirmacion.

## Contrato relacionado

- `documentos/gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md`

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`
