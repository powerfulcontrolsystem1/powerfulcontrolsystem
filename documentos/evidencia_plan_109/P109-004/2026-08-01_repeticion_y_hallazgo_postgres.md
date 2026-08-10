# P109-004 - Repetición del barrido y hallazgo PostgreSQL

Fecha: 2026-08-01
Entorno: staging
Candidato: `6da6c13453a40b2d84e23285fa83f255f34788da`

La primera repetición limpia (`30688221005`) no superó la navegación inicial a
login por timeout de 45 segundos. Se clasificó como fallo transitorio y no como
evidencia funcional.

La segunda repetición (`30688330490`) superó login y comenzó el inventario
completo. Durante el recorrido del panel PostgreSQL, los logs del mismo digest
mostraron respuestas 500 repetidas en `/super/api/postgres/performance`. Una
consulta autenticada estable devolvía 200, por lo que se correlacionó el fallo
con la cancelación de requests al navegar entre controles.

El handler concatenaba cualquier error de contexto/DB a una respuesta 500. La
rama de cierre ahora:

- no escribe un error sintético cuando el cliente ya canceló;
- devuelve 503 ante un deadline real;
- conserva el detalle interno solo en log y envía un mensaje público genérico;
- añade pruebas unitarias para las tres ramas.

Se canceló el barrido del digest defectuoso en lugar de presentarlo como PASS.
La suite completa del backend y el preflight profesional aprobaron la corrección
local. Falta fusionar, publicar un digest nuevo y repetir 309 rutas en dos
viewports antes de aceptar este subcriterio.

Estado: **P109-004 parcial**.
