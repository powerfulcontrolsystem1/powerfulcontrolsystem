# Promocion de licencias por codigo de asesor

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- El descuento de asesor es parte del cálculo de checkout; importe y moneda deben validarse con los valores esperados del servidor antes de activar una licencia.
- La guía de promoción no sustituye el contrato checkout ni autoriza una activación gratuita, comisión o cobro reales.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-05-06

## Objetivo

Permitir que el super administrador active una promocion global para compradores de licencias: cuando el cliente ingresa un codigo de asesor comercial aceptado, el checkout aplica un descuento adicional configurable antes de abrir Wompi, Epayco o la activacion directa sin pago.

## Configuracion

Ubicacion: `web/super/asesor_comercial.html`.

Campos:

- `Activar descuento por codigo de asesor`: habilita o deshabilita la promocion.
- `Descuento adicional`: porcentaje aplicado al subtotal despues de otros codigos promocionales.

Claves guardadas en `configuraciones`:

- `licencias.asesor_promo.enabled`
- `licencias.asesor_promo.percent`
- `licencias.asesor_promo.updated_by`

## Comportamiento

- El comprador escribe el codigo en `pagar_licencia.html`.
- En la licencia gratis de 15 dias, tambien puede escribir el codigo desde
  `elegir_licencia.html`; ese valor viaja al checkout como `asesor_id`.
- El resumen publico `/api/public/licencias/checkout_summary` recibe `asesor_id`.
- Si la promocion esta activa y el asesor existe, esta activo y acepto la invitacion, se aplica el porcentaje configurado.
- El descuento por asesor se suma al descuento total mostrado, pero se conserva como `advisor_discount_value` y `advisor_discount_percent` para trazabilidad.
- Wompi, Epayco y activacion sin pago reciben el mismo `asesor_id`; el sistema conserva la liquidacion de comisiones del asesor.
- Si la promocion esta desactivada, el codigo de asesor sigue sirviendo para comision, pero no modifica el precio.

## Fuentes y aceptación de la revisión

[payments_handlers.go](../backend/handlers/payments_handlers.go), [licencias_globales.go](../backend/db/licencias_globales.go), [asesor_comercial.html](../web/super/asesor_comercial.html), [contrato_checkout_licencias_publico.md](gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md).

Requisitos aplicables: PCS-REQ-002, PCS-REQ-003, PCS-REQ-005, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
