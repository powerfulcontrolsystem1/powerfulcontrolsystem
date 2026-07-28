# P108-011 - regresión autenticada de módulos críticos

Fecha: 2026-07-28  
Ambiente: staging aislado  
Empresa autorizada: Powerful Control System (`empresa_id=12`)  
SHA desplegado final por digest:
`fce1655cedff6d3e9235424bfaf0029e80b2ff0c`

## Alcance

Se recorrieron ocho rutas críticas en escritorio 1366 x 900 y móvil
390 x 844:

- panel empresarial;
- Finanzas/CxP;
- reportes ejecutivos;
- facturación electrónica;
- carrito de compras;
- ruta candidata de Corte de Caja;
- correo corporativo;
- resumen super administrador.

El auditor encontró 534 controles, pulsó 30 controles clasificados como
seguros y omitió 101 acciones con posibilidad de mutación. No ejecutó ventas,
pagos, envíos, guardados, anulaciones, impresiones ni cierres.

## Resultado

- 10 de 16 combinaciones ruta/viewport: `ok`;
- 6 de 16: `review`;
- Finanzas, reportes, facturación, correo y resumen super no presentaron 5xx
  inesperados en ambos viewports;
- el inventario usado apuntó a
  `/administrar_empresa/corte_caja.html`, archivo que no existe. Debe
  reconciliarse con la ruta canónica del menú antes de repetir y no se registra
  como fallo del módulo;
- el panel generó advertencias no bloqueantes por atributos redundantes
  `allow`/`allowfullscreen`; la rama candidata retira el atributo duplicado y
  conserva `fullscreen` en la política `allow`;
- el carrito reprodujo cuatro veces HTTP 500 en
  `/api/empresa/configuracion_operativa`.

## Causa y corrección local del 500

Los logs de staging informaron:

`sql: expected 16 destination arguments in Scan, not 18`

`ListEmpresaConfiguracionOperativaRoles` leía en `Scan` los campos
`permitir_ingresos_manuales` y `permitir_egresos_manuales`, pero el `SELECT`
no los incluía. La corrección agrega ambos `COALESCE(..., 0)` manteniendo el
filtro por `empresa_id` y un contrato que verifica 18 columnas/destinos
alineados.

## Estado

La corrección se desplegó en staging. La repetición autenticada de
`carrito_de_compras.html` y de la ruta canónica
`corte_de_caja.html` terminó **4/4 PASS** en 390 x 844 y 1366 x 900, sin
respuestas HTTP 4xx/5xx. `/health` informó `ok`, `/ready` informó `ready` y
backend, worker y PostgreSQL quedaron saludables.

**Parcial**. El P0 reproducido quedó corregido y verificado sobre el SHA final,
pero el barrido no sustituye las 101 acciones mutables ni el inventario completo
exigido por P108-011.

La página de carrito se volvió a abrir autenticada después de promover el
digest exacto; no produjo respuestas HTTP 4xx/5xx.
