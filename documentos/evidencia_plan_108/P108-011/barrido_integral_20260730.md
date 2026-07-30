# P108-011 - barrido integral autenticado

Fecha: 2026-07-30

Candidato auditado: `f9396da5e41562968996b05136fffca9991b56f9`

Workflow: `30583852262`

Estado: **parcial / NO-GO**

## Cobertura

- 309 rutas en escritorio y móvil: 618 combinaciones.
- 10.979 controles inventariados.
- 896 clics seguros ejecutados.
- 1.947 acciones mutantes o riesgosas preservadas.
- 564 vistas `ok`, 54 en revisión y cero fallos fatales del runner.
- Matriz de roles, formatos imprimibles y matriz de pagos terminaron `PASS`.

La repetición dirigida del runner corregido (`30585717175`) auditó Estaciones y
Productos en ambos viewports. Estaciones terminó 2/2 sin errores, ejecutó los
tres controles seguros de cada vista y eliminó los falsos fallos causados por
reconstrucción del DOM.

## Hallazgos reales corregidos en el siguiente candidato

1. Bre-B QR mezclaba columnas de fecha `TEXT` con `CURRENT_TIMESTAMP` dentro de
   `COALESCE` en PostgreSQL.
2. Hoja de vida comparaba el flag entero heredado `recurrente` como booleano.
3. Nextcloud podía encontrar una tabla heredada sin la columna `provisioned`;
   se añadió una migración inmutable v2 y readiness cerrado.
4. Productos contenía el carácter `ú` donde debía construir `?empresa_id` en
   dos APIs de inventario.
5. La entrada genérica Colombia llamaba `/api/empresa/` sin módulo.
6. CSP bloqueaba Leaflet, Chart.js y Google Fonts. Se añadieron orígenes exactos
   configurables, `font-src`, versiones fijas y SRI para Leaflet/Chart.js.
7. El runner ahora restaura la ruta cuando un clic seguro cambia la visibilidad
   de los controles restantes y omite los que siguen ocultos.

## Hallazgos clasificados, no declarados como PASS

- 403 en Centro IA, Renta IA y Aseo: requieren matriz de licencia/rol real.
- 400 en avisos de vencimiento DIAN y Noticias: requieren configuración
  operativa del módulo y repetición del nuevo candidato.
- 404 en imágenes históricas de red social y catálogo público sin slug válido:
  requieren reconciliación de datos/fixture, no se corrigieron ocultándolos.
- Desbordes y controles sin etiqueta de tutoriales/diagramas permanecen en la
  cola visual.

P108-011 sigue parcial hasta desplegar el nuevo digest, repetir las rutas
afectadas y probar las acciones riesgosas mediante sus flujos oficiales.
