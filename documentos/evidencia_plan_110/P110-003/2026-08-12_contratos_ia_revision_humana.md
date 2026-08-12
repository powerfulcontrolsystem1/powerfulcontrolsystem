# P110-003 — Contratos de IA y revisión humana

Fecha: 2026-08-12 (America/Bogota)  
Alcance: pruebas deterministas del candidato; no se generaron CxP, pagos ni
asientos durante esta comprobación.

## Controles cubiertos

- Los controles visibles de Captura IA separan `Extraer IA`, `Cancelar IA`,
  `Aprobar`, `Rechazar`, `Contabilizar`, papelera y depuración; la interfaz
  exige selección/estado válido y confirmación para acciones mutantes.
- La extracción fuerza revisión humana; los datos extraídos se editan en una
  pantalla separada y no contabilizan por sí solos.
- Las pruebas aprobaron respuesta IA estructurada/inválida/adversarial,
  proveedor caído, doble clic, proveedor canónico, papelera recuperable,
  estados cerrados e historial aislado por empresa y usuario.
- El modo agente se mantiene cerrado por defecto y el servidor conserva la
  propiedad del cambio de modo.

## Verificación

`go test ./handlers` y `go test ./db` focalizados: PASS.  
`go vet ./handlers ./db`: PASS.

## Límites

Faltan ejecutar todos los botones IA visibles bajo cada rol en navegador,
pruebas de proveedor real/timeout/cancelación controlada y evidencia visual de
edición completa. P110-003 permanece **parcial**; no autoriza contabilizar ni
cerrar la compuerta de IA.
