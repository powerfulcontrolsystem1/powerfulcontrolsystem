# P109-004 - Barrido completo del candidato 89d6e042

Fecha: 2026-08-01

Entorno: staging, PCS (`empresa_id=12`)

Candidato: `89d6e042e24d57cba920439521704acaacf7bd00`

El primer intento del workflow `30701434297` tuvo timeout al abrir el login y
se clasifico como inconcluso. El reintento, sin cambiar codigo ni digest,
completo correctamente:

- 618 vistas sobre 309 rutas en escritorio y movil;
- 600 vistas `ok` y 18 `review` conocidas;
- 11.062 controles detectados y 97 clics seguros;
- 2.087 acciones riesgosas omitidas;
- 12 POST automaticos bloqueados por la guardia;
- cero mutaciones de negocio atribuibles al barrido.

Las 18 revisiones corresponden a denegaciones 403 esperadas, POST publicos o
IA bloqueados por la guardia y dos imagenes historicas 404 de red social. Otras
24 vistas `ok` registraron recursos abortados al desmontar iframes o cambiar de
ruta; no se clasificaron como fallo funcional.

La misma ejecucion volvio a aprobar 20/20 formatos imprimibles, incluidas
factura y recibo extensos de seis paginas, sin casos a revisar.

Estado: **P109-004 parcial**. El inventario no mutante del candidato esta
cerrado; faltan acciones riesgosas por flujo oficial, rol, empresa A/B y firma
del alcance incluido/excluido.
