# Runbook: checkout de licencias no activa o envia correo incorrecto

Fecha: 2026-04-18
Estado: vigente

## Sintomas cubiertos

- el pago aparece aprobado pero la licencia no queda activa.
- el correo de activacion llega con empresa incorrecta o con el texto `tu empresa`.
- el polling de estado devuelve conflicto o no encuentra la referencia esperada.
- el retorno desde la pasarela no reanuda correctamente el checkout.

## Alcance

Aplica al checkout publico de licencias sobre Epayco y Wompi.

## Fuentes de evidencia

- logs de backend
- tablas de pagos (`pagos_epayco`, `pagos_wompi`)
- tabla o estado de licencias
- `raw_payload` del pago
- parametros devueltos a `pagar_licencia.html`

## Verificaciones iniciales

1. confirmar `empresa_id` y `licencia_id` que el frontend tenia abiertos.
2. verificar `transaction_id` o `reference` de la pasarela.
3. revisar si el pago encontrado pertenece a la misma empresa/licencia esperada.
4. comprobar si el correo del comprador vino en payload principal o dentro de `data.customer_email`.
5. revisar si `raw_payload` ya tiene marcas de correo enviado para evitar falsos duplicados.

## Causas probables

- referencia vieja o cruzada entre empresas.
- polling sin `empresa_id` y `licencia_id` esperados.
- empresa resuelta por id fisico y no por alcance logico.
- esquema de pagos incompleto o legacy al primer acceso.
- retorno frontend con contexto incompleto al volver de la pasarela.

## Acciones de recuperacion

1. validar que el frontend este reenviando `empresa_id` y `licencia_id` al consultar `transaction_status`.
2. verificar que el backend compare contexto esperado contra el pago resuelto antes de activarlo.
3. revisar que la empresa se resuelva por alcance logico cuando se construye el correo.
4. si hay errores de tabla o columna faltante en pagos, regularizar el esquema y repetir la consulta de estado.
5. si webhook activó primero y el correo quedo pendiente, confirmar que el polling posterior pueda completar solo la notificacion faltante.

## Validacion posterior

- la licencia queda activa en la empresa correcta.
- el correo muestra la empresa correcta.
- no se duplica el correo en webhooks o polls posteriores.
- el frontend puede cerrar el checkout en la pagina de exito sin perder contexto.

## Contrato relacionado

- `documentos/gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md`

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`
