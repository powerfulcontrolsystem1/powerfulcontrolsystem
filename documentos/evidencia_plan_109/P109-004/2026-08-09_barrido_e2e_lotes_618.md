# P109-004 - Barrido E2E autenticado por lotes

Fecha: 2026-08-09  
Ambiente: staging, empresa PCS 12  
Protección: el auditor bloqueó toda mutación HTTP

## Cobertura

| Lote | Vistas | Resultado principal |
| --- | ---: | --- |
| 0–79 | 160 | 160 `ok` |
| 80–159 | 160 | 157 `ok`, 3 revisión |
| 160–239 | 160 | 160 `ok` |
| 240–308 | 138 | 126 `ok`, 12 revisión |

Total: 618 vistas, 11.258 botones y 75 clics seguros. La carga GET adicional
de 500 solicitudes, concurrencia 10, aprobó con p95 135 ms y 0 % de errores.

## Renta IA

El digest `e51228dd` quitó el POST automático de inicio. La repetición dirigida
aprobó escritorio y móvil: 2/2 `ok`, cero mutaciones bloqueadas y cero errores.

## Pendiente

Persisten seis escrituras automáticas en páginas públicas y dos 502 móviles de
ayuda. El barrido no sustituye acciones mutantes, pruebas por rol ni botones IA
que requieren aprobación humana.

## Repetición de páginas públicas

El QA `31323978918` repitió las cinco páginas públicas en ambos viewports: 10/10
vistas `ok`, cero mutaciones operativas bloqueadas, 10 eventos de telemetría de
visitas bloqueados y registrados, y cero errores de página. Solo permanecen los
dos 502 de ayuda móvil y las pruebas operativas fuera del alcance no mutante.

## Repetición de ayudas móviles

La ejecución aislada `31325595643` repitió Ayuda APIs y Ayuda contextual en
móvil y aprobó. Los 502 vistos durante la corrida paralela no se reprodujeron,
por lo que se registran como condición transitoria de carga y no como defecto
funcional abierto.
