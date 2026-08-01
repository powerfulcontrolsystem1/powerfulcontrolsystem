# P109-001 - Pago mínimo CxP y precisión monetaria

Fecha: 2026-08-01
Entorno: staging, PCS (`empresa_id=12`)
Producción: sin cambios

El soporte controlado `SCI-0003` se convirtió por la UI oficial en la CxP
canónica ID 16, código `CXP-SCI-0003`, documento `FE-1024` y valor 214200.
Repetir `Contabilizar` conservó exactamente una fila y un solo evento.

Se aplicó un abono mínimo real de 0,01. La transacción produjo exactamente:

- un registro `empresa_cxp_pagos` ID 5;
- un movimiento financiero ID 36 por 0,01;
- un outbox ID 7, publicado en un intento;
- un evento contable ID 100, procesado en un intento;
- un asiento ID 89, débito 0,01, crédito 0,01 y diferencia cero.

La prueba descubrió un defecto P0 de persistencia: la tabla canónica conservaba
`valor_original`, `valor_pagado` y `saldo` como PostgreSQL `REAL`. Aunque el
abono fue 0,01 y `valor_pagado` quedó 0,01, el saldo almacenado fue 214199,98 en
lugar de 214199,99 por pérdida de precisión `float4` a esa magnitud.

La corrección local cambia CxC y CxP a `NUMERIC(18,2)`, recalcula el saldo exacto
y añade una restricción `saldo = valor_original - valor_pagado`. La migración
falla cerrada si encuentra negativos, sobrepagos o diferencias superiores a
0,02 que requieran conciliación manual. El preflight de solo lectura encontró
16 CxP y 15 CxC en staging, cero incompatibilidades; producción tiene cero filas
en ambas tablas.

Una prueba PostgreSQL aislada con tablas temporales confirmó `214199,99` y el
rechazo de drift material. No se aplicó la migración al staging activo porque el
bloque se ejecuta sin PR ni candidato inmutable.

Estado: **P109-001 parcial**. Pago, outbox, evento y asiento quedan conciliados
1:1; faltan publicación, migración por digest y concurrencia con identidades
distintas.
