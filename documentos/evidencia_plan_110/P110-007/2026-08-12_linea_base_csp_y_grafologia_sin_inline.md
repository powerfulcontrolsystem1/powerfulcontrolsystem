# P110-007 - linea base CSP y Grafologia sin inline

Fecha: 2026-08-12
Candidato base: `70bed763`

## Inventario reproducible

El nuevo auditor recorre todas las paginas HTML y separa scripts inline sin
proteccion, bloques `style`, atributos `style`, eventos `on*` y URLs
`javascript:`. La primera medicion, antes de corregir Grafologia, observo 1.347
bloqueos en 243 paginas; 975 pertenecian al area empresarial.

La linea base comprometida se toma despues de la correccion de Grafologia:

- 1.335 bloqueos en 242 paginas.
- 223 scripts inline sin proteccion.
- 187 bloques `style` sin proteccion.
- 906 atributos `style`.
- 19 eventos inline.
- cero URLs `javascript:`.
- 963 bloqueos en `administrar_empresa`.

La compuerta de CI falla si cualquier total o area aumenta y tambien si se
elimina la linea base. Los valores son maximos temporales: una reduccion pasa
sin editar el baseline; solo debe bajarse el baseline al cerrar un lote.

## Reduccion material

Grafologia movio su bloque de estilos y once atributos inline a
`web/grafologia.css`. La pagina queda con cero bloqueos inventariados y conserva
sus temas, grillas, responsive y estados dinamicos por clases. Esto reduce doce
bloqueos y una pagina completa de la deuda empresarial.

La revision visual local cargo la hoja externa sin errores de consola. En vista
de escritorio conservo dos columnas; en el viewport movil la grilla paso a una
columna y los KPI a dos columnas. En ambos casos `scrollWidth` fue menor que
`innerWidth`, sin desborde horizontal. No se invocaron API ni camara.

## Limite y veredicto

La CSP aplicada todavia necesita `unsafe-inline` debido a los 1.335 bloqueos
restantes. La politica Report-Only estricta continua siendo la observacion
segura. P110-007 permanece parcial y el veredicto es NO-GO. No se ejecuto `rs`.
