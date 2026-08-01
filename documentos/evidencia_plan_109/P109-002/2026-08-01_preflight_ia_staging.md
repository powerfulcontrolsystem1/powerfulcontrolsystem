# P109-002 - Preflight autenticado de IA y CxP/IA

Fecha: 2026-08-01
Entorno: staging, PCS (`empresa_id=12`)
Candidato: `6da6c13453a40b2d84e23285fa83f255f34788da`

Sin revelar valores privados, el runtime confirmó presencia de
`CONFIG_ENC_KEY`, `CONFIG_ENC_KEY_ID` y `OPENAI_API_KEY`. La lectura autenticada
aprobó:

- catálogo IA: HTTP 200, cinco modelos y preferencia con alcance por usuario;
- preferencia e historial: HTTP 200;
- configuración OpenAI propia: HTTP 200, desactivada y sin devolver `api_key`;
- modo inicial `operativo`, agente `general` y streaming habilitado;
- dashboard CxP/IA: HTTP 200.

El upload CxP/IA rechazó una extensión no permitida con HTTP 400. La cantidad
de soportes permaneció `1 -> 1`, demostrando fallo antes de guardar archivo o
fila.

No se llamó al proveedor externo ni se generó costo en este ciclo. Continúan
pendientes la extracción real autorizada, edición/confirmación/cancelación,
doble clic, degradación del proveedor, evals y aislamiento A/B.

Estado: **P109-002 parcial**.
