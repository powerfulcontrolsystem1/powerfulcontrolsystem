# P110-003 - Adjunto IA real y revisión humana en `945d1751`

Fecha: 2026-08-12 (America/Bogota)
Empresa: Powerful Control System (`empresa_id=12`)
Ambiente: staging con el candidato inmutable `945d1751`.

## Flujo ejecutado

1. Se radicó por la interfaz oficial un XML sintético, sin valor fiscal ni
   datos personales, con proveedor, NIT, fechas, subtotal, IVA y total
   controlados.
2. El antivirus aceptó el archivo limpio y creó el soporte `SCI-0028`.
3. El botón **Extraer IA** procesó el archivo con el proveedor configurado y
   dejó el soporte en `En revision`, sin aprobarlo ni contabilizarlo.
4. La lectura recuperó proveedor, NIT, número, fechas, subtotal `10.000`, IVA
   `1.900` y total `11.900`, con confianza visible de `78 %`.
5. La revisión humana permitió modificar el proveedor leído, centro de costo e
   impacto de inventario. La auditoría registró `editar revision` y conservó el
   estado `En revision`.
6. El soporte se rechazó por el diálogo oficial; quedó `Rechazado`, sin crear
   cuenta por pagar, asiento, pago ni movimiento de inventario.
7. Una búsqueda posterior del documento en la fuente canónica CxP devolvió
   `No hay registros de cartera para el filtro actual`.

## Controles observados

- La extracción nunca contabilizó automáticamente.
- Los datos leídos fueron editables antes de cualquier aprobación.
- El soporte y sus eventos quedaron aislados en la empresa 12.
- No se repitió el envío ante resultados inciertos.
- La ventana para cancelar fue visible durante el procesamiento, pero el
  proveedor terminó antes de que el control pudiera accionarse; la cancelación
  real continúa pendiente de una respuesta deliberadamente lenta.

## Estado

P110-003 continúa **parcial**. Esta pasada cierra el adjunto limpio, la
extracción real, la edición humana, el rechazo y la ausencia de CxP, pero aún
faltan la matriz completa de botones/roles, timeout controlado, proveedor caído,
doble envío, evals y aislamiento A/B integral.

No se ejecutó `rs` en este bloque.
