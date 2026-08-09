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
