# P109-004 - Inventario estático de controles

Fecha: 2026-07-31
Alcance: 309 archivos HTML bajo `web`.

## Línea base

| Medida | Resultado |
| --- | ---: |
| Páginas HTML | 309 |
| Páginas con controles estáticos | 242 |
| Marcadores estáticos de control | 2.160 |
| Manejadores `onclick` | 18 |
| Referencias IA encontradas | 657 |

Los 2.160 son marcadores estáticos no deduplicados (`button`, `input`, roles y
enlaces con clase de acción); no equivalen todavía a controles dinámicos ni a
acciones certificadas.

## Prioridad de recorrido

Las pantallas con más controles son Configuración avanzada super (87), Productos
(62), Carrito (60), Finanzas (57), Chat/Tareas (39), Administración de empresa
y Nómina (37). Facturación electrónica/pruebas DIAN, Venta pública, Estaciones
y Contabilidad avanzada forman el siguiente lote crítico por efecto operativo.

## Límite de cierre

La evidencia organiza P109-004 pero no aprueba el inventario completo. Faltan
descubrimiento dinámico, clasificación de riesgo por control, pruebas por rol,
acciones IA, exportaciones, anulaciones y resultados auditables por flujo.
