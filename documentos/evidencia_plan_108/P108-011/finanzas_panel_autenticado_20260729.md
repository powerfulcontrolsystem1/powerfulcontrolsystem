# P108-011 - Panel y Finanzas autenticados en staging

Fecha: 2026-07-29  
Ambiente: staging aislado  
Empresa autorizada: Powerful Control System (`empresa_id=12`)  
Alcance: navegación autenticada, visual y clics no mutantes.

## Recorrido

Se ejecutó `tools/qa_e2e_buttons.cjs` con la sesión autorizada, en escritorio
1366 x 900 y móvil 390 x 844, sobre el panel empresarial y Finanzas. Las
credenciales se entregaron únicamente como variables efímeras al proceso de
prueba y no quedaron registradas en los artefactos.

- 4 de 4 combinaciones ruta/viewport: `ok`;
- 116 controles detectados;
- 8 interacciones clasificadas como seguras, sin guardado ni emisión;
- 46 acciones mutables o de salida (guardar, pago, anulación, exportación,
  impresión, eliminación y cierre) omitidas;
- 0 errores visuales, de consola, de red o de autorización observados.

Una confirmación de caja surgió durante el recorrido y fue descartada antes de
enviar cualquier operación. No se crearon ventas, pagos, movimientos, facturas,
documentos, impresiones ni cambios de configuración.

## Comprobación visual

Las capturas de Finanzas confirmaron una disposición legible y sin desbordamiento
horizontal en ambos viewports. En móvil, el encabezado pasa a dos líneas y los
campos se apilan conservando etiqueta, valor y controles accesibles.

## Estado

**Parcial aprobado para navegación y responsive no mutante.** No sustituye la
prueba de impresiones, las transiciones financieras, cuatro cajas concurrentes
ni la carga autenticada sostenida requerida por otras compuertas del Plan 108.
