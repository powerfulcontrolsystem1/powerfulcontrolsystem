# P109-004 - Barrido global protegido del candidato d4e613e2

Fecha: 2026-08-01  
Entorno: staging, PCS (`empresa_id=12`)  
Candidato: `d4e613e2d2c6b34451838bcc18a5db5f7afcf2d0`

La repeticion dirigida `30691170513` aprobo VPS2 y Configuracion avanzada en
escritorio/movil: cuatro vistas, 290 controles, cinco clics seguros, cero
mutaciones, cero errores y estado VPS2 degradado sin 5xx.

El barrido global `30691247791` termino correctamente:

| Control | Resultado |
| --- | ---: |
| Vistas / rutas | 618 / 309 |
| Vistas OK / review | 600 / 18 |
| Controles inventariados | 11.062 |
| Clics seguros | 103 |
| Acciones riesgosas omitidas | 2.081 |
| Mutaciones bloqueadas | 12 |
| Escrituras HeadlessChrome antes/despues | `101/15519` sin cambio |

Los 12 POST bloqueados corresponden a la lectura de renta IA y al contador de
visitas publicas; no llegaron al servidor. Los cuatro 403 son permisos cerrados
del Centro IA y reporte de aseo, y los cuatro 404 corresponden a dos imagenes
historicas de Red Social. No hubo `pageerror` ni 5xx del backend.

La unica incidencia de layout fue el grupo de seis filtros Kardex de Bodegas:
los anchos minimos globales extendian el documento a 1.929 px en un viewport de
1.366 px. El grupo ahora envuelve sus controles y reduce sus anchos de forma
acotada; el contrato profesional exige conservar esta regla.

La impresion sintetica del mismo job aprobo 20/20 formatos. Factura y recibo
extensos generaron seis paginas cada uno; la inspeccion visual confirmo filas,
columnas, total, pie y QR sin recortes.

Estado: **P109-004 parcial**. El inventario de lectura esta cubierto, pero las
acciones mutantes requieren flujos oficiales por rol; la correccion Kardex debe
integrarse y repetirse sobre su digest. P109-005 permanece parcial por
documentos reales, tableta e impresion fisica.
