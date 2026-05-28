# Promocion de licencias por codigo de asesor

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
