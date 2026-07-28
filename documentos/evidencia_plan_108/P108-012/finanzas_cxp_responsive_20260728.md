# P108-012 - Finanzas y CxP responsive en staging

Fecha: 2026-07-28  
Ambiente: `https://staging.powerfulcontrolsystem.com`  
Empresa: Powerful Control System (`empresa_id=12`)  
Resultado: **PARCIAL APROBADO PARA ESTA PANTALLA**

## Cobertura visual

- Escritorio con el tamaño normal del navegador interno.
- Móvil con viewport temporal de `390 x 844`.
- Tema oscuro vigente.
- Formularios, tarjetas resumen, botones, tablas y desplazamiento horizontal.

## Resultados

- La sección `Cartera CxC / CxP` presenta los controles IA, búsqueda,
  conciliación y comparación histórica sin superposición.
- En móvil los botones se apilan, mantienen texto legible y son accesibles.
- La tabla CxP conserva columnas y desplazamiento horizontal en lugar de
  comprimir o cortar los valores.
- La fila canónica QA mostró de forma consistente original `$100`, pagado
  `$25`, saldo `$75`, estado `parcial`.
- La tabla de movimientos conservó una sola fila antes y después de abrir la
  vista previa; el botón `Imprimir` no duplicó el movimiento.
- La conciliación contable mostró tres eventos, uno procesado, dos pendientes,
  dos con error, un asiento y diferencia monetaria `$0`.
- La bandeja de soportes IA mostró el soporte `SCI-0001` en una fila legible,
  con total `$0`, confianza `0%` y revisión humana.

## Pendientes

- Repetir la vista previa financiera después de desplegar el PR que abre la
  ventana antes de resolver la impresora.
- Completar carta, A4 y POS térmico con PDF/captura y datos de varias líneas.
- La verificación cubre Finanzas/CxP; no certifica todavía P108-012 ni el
  inventario completo de impresiones P108-013.
