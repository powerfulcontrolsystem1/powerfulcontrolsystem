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

## Primer candidato de reparación `7819f775`

El CI profesional, el escaneo de imágenes, Trivy, SBOM y los cuatro artefactos
inmutables terminaron correctamente. Staging promovió los digests exactos,
ejecutó el migrador con código `0` y conservó intacta la huella de los
contenedores de producción.

La repetición posterior cubrió 22 vistas, 400 controles y 42 clics seguros:

- Bre-B QR, la entrada Colombia y Productos/Bodegas quedaron sin respuestas
  HTTP erróneas en escritorio y móvil.
- La tabla histórica Nextcloud tenía además ausente `quota_mb`; se prepara una
  migración v3 acumulativa porque la v2 publicada no se puede modificar.
- Hoja de vida conservaba otra comparación directa entre la fecha heredada
  `TEXT` y `CURRENT_TIMESTAMP`; debe normalizarla con `pcs_ts`.
- El CSP dinámico del backend quedó configurado, pero las páginas estáticas
  reciben el CSP propio de Nginx. Ese archivo también debe autorizar los
  orígenes exactos y versionados de Leaflet, Chart.js y fuentes.

El candidato `7819f775` no se declara aprobado para esas tres rutas. Las
correcciones complementarias requieren un nuevo digest y repetición autenticada.

El candidato complementario `424b1896` ejecutó correctamente la v3 y verificó
las seis columnas de operación. La primera inserción oficial reveló una última
restricción heredada: `password_encrypted NOT NULL`. El código vigente no usa
ni debe almacenar esa credencial localmente, por lo que una v4 inmutable retira
solo esa columna obsoleta antes de repetir el flujo.

La misma verificación visual confirmó que Leaflet ya carga sin el bloqueo CSP.
Domicilios mostró luego un estado vacío serializado como `null`; backend y
frontend se endurecen para representar colecciones vacías como `[]` y evitar
que `renderMenu` invoque `.map` sobre un valor nulo.

## Candidato consolidado `efc416b3`

- Migrador: código `0`.
- Salud/readiness: `ok` / `ready`.
- Producción: misma huella de contenedores antes y después.
- Nextcloud: respuesta `ok:true`, asignación local `pcs_empresa_12`, cuota
  1024, sin provisionamiento externo y sin columna local obsoleta de contraseña.
- Workflow visual `30588981805`: 22/22 vistas `ok`, 420 controles, 42 clics
  seguros, 56 acciones riesgosas preservadas y cero respuestas HTTP erróneas.
- Bre-B, Hoja de vida, Nextcloud, Colombia, Productos/Bodegas, Domicilios,
  Taxi, GPS y mapas públicos aprobaron en escritorio y móvil.

Venta pública conserva ocho advertencias del runner al intentar controles de
paneles fijos cerrados fuera del viewport; no produjo HTTP, error de página ni
bloqueo de seguridad. P108-011 permanece parcial porque el barrido completo no
se repitió sobre este digest y las acciones mutantes siguen reservadas para sus
flujos oficiales.

## Barrido integral final del candidato `5ec1c48f`

El workflow inmutable `30591586319` terminó correctamente sobre los cuatro
digests publicados por el commit
`5ec1c48f98e224a2c8283324be32f08a82ac737b`:

- 618 vistas ejecutadas: 309 rutas en escritorio y móvil;
- 604 vistas `ok` y 14 vistas `review`;
- 10.998 controles inventariados;
- 1.062 clics seguros ejecutados;
- 1.975 acciones mutantes o riesgosas preservadas;
- cero errores de página, cero HTTP 500 y cero bloqueos CSP;
- matrices de roles y pagos: `PASS`.

Las 14 vistas en revisión corresponden a las mismas siete rutas en ambos
viewports:

1. Centro IA y Renta IA: `403` por rol/licencia, comportamiento de seguridad
   esperado que no debe corregirse ampliando permisos.
2. Reporte de aseo: `403` por permiso efectivo esperado.
3. Facturación electrónica: avisos de vencimiento retornan `400` porque PCS no
   tiene completa la configuración DIAN de certificado/resolución.
4. Noticias: `400` por configuración operativa ausente.
5. Red social comercial: dos imágenes históricas inexistentes de otra empresa;
   no se alteraron datos fuera de la empresa 12 autorizada.
6. Catálogo público: el host técnico `staging` era interpretado como slug y el
   JSON del error producía desborde móvil. La corrección queda implementada y
   requiere el siguiente digest para su prueba visual.

Los 246 avisos `safe-button-click-failed` son timeouts del runner después de
cambios de estado visibles, principalmente en drawers compartidos de IA; no
son respuestas 5xx ni errores de página. También se conservaron 30 controles
sin etiqueta, 37 desbordes internos de contenido y ocho navegaciones externas
para revisión específica.

Estado: **parcial / NO-GO**. El barrido demuestra amplitud y ausencia de fallos
fatales, pero no sustituye los flujos oficiales de las 1.975 acciones mutantes,
la evaluación completa de botones IA ni la repetición del catálogo público
corregido sobre un digest nuevo.
