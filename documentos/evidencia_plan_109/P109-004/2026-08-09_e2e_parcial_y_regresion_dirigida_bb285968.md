# P109-004 - E2E parcial y regresión dirigida bb285968

Fecha: 2026-08-09

El barrido de inventario completo ejecutó login oficial contra staging y registró
363 variantes antes de agotar el límite externo de ejecución de 15 minutos.
Ese timeout es inconcluso: no se presenta como éxito ni como defecto del
producto.

Después, el mismo QA se ejecutó con una ruta explícita y ambas vistas:

| Vistas | Botones | Errores de página | Mutaciones bloqueadas | Resultado |
| ---: | ---: | ---: | ---: | --- |
| 2/2 | 108 | 0 | 0 | PASS dirigido |

El cambio de `release_gate.ps1` descubre el ejecutable local Chrome/Edge y la
ruta oficial de Node. De ese modo la compuerta ya no depende de una descarga de
Playwright para ejecutar los barridos posteriores.

P109-004 continúa parcial hasta recorrer el inventario completo por bloques y
registrar todas las acciones/roles de riesgo.
