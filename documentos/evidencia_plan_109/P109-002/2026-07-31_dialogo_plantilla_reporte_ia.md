# P109-002 - diálogo accesible para guardar plantillas de reportes IA

Fecha: 2026-07-31
Estado: parcial
Ambiente de esta prueba: servidor local aislado con respuestas deterministas; sin mutaciones en PCS, staging ni producción.

## Cambio validado

- `Guardar plantilla` dejó de usar `window.prompt` y abre un diálogo con código,
  nombre, cancelar y confirmación explícita.
- El foco inicia en el código; Escape y Cancelar cierran el diálogo y devuelven
  el foco al botón de origen.
- Código y nombre tienen etiquetas, ayuda, límites y validación. El backend
  vuelve a validar el código y el tamaño del nombre antes de persistir.
- La solicitud conserva el `empresa_id` resuelto por el contexto y el handler
  compara cualquier empresa recibida con el tenant autenticado. Las consultas,
  versiones y lecturas continúan filtradas por `empresa_id`.

## Evidencia ejecutada

- Contrato Go del HTML y validador backend: PASS.
- Flujo de navegador: generar vista determinista, abrir diálogo, rechazar código
  inválido y guardar código válido: PASS.
- Escritorio: ancho del diálogo 520 px dentro de un viewport de 2529 px, foco
  inicial correcto y sin desborde interno: PASS.
- Móvil emulado a 390 x 844: diálogo de 352,4 px, márgenes de 19 px, ancho de
  documento 390 px, sin desborde horizontal: PASS.
- Teclado: Escape cierra y devuelve el foco a `btnSaveAIReportTemplate`: PASS.
- Consola del navegador y diálogo JavaScript inesperado: cero errores y ningún
  `prompt`: PASS.

## Límites y siguiente compuerta

El diálogo todavía debe promoverse como imagen inmutable y repetirse en staging.
El cupo diario de Reportes IA de PCS ya se consumió en la validación real del
candidato anterior; no se alteró cuota ni auditoría por SQL para forzar otra
ejecución. P109-002 continúa parcial hasta completar CxP IA, Centro IA, matriz
A/B, errores/reintentos y evals definidos por el plan.
