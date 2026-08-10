# P109-009 - Inventario CSP inline

Fecha: 2026-07-31
Alcance: 309 páginas HTML bajo `web`.

## Línea base

| Recurso que requiere CSP permisiva | Conteo |
| --- | ---: |
| Scripts sin `src` | 225 |
| Bloques `<style>` | 186 |
| Manejadores `on*=` | 23 |

La CSP central actual contiene `unsafe-inline` en `script-src` y `style-src`.
Retirarlo en el candidato actual bloquearía funciones existentes y no es una
corrección segura de último momento.

## Prioridad de migración

Las pantallas con mayor concentración son: carrito de compras (13), administrar
productos (7), facturas electrónicas (7), ventas (7), carta pública (5), chat y
tareas (5) y Finanzas (5). Deben migrar primero sus scripts a archivos versionados
y sus estilos a clases/CSS compartido; los manejadores `on*=` deben convertirse a
listeners registrados por JavaScript.

## Criterio para cierre

Implementar por lote, con pruebas visuales y de flujo oficial, una CSP de
reporte sin `unsafe-inline`; luego retirar la directiva del modo enforce cuando
no haya violaciones críticas. Hasta entonces P109-009 permanece parcial y
mantiene NO-GO para la compuerta de seguridad.
