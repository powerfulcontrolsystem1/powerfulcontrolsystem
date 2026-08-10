# P109-002 - contratos IA y soportes de compras

Fecha: 2026-08-08  
Alcance: pruebas locales de seguridad y reglas; sin llamada a proveedor IA ni
mutación de datos empresariales.

## Verificación

`go test ./handlers -run ... -count=1 -v` aprobó los contratos de:

- IA empresarial cerrada por defecto, conversación opaca y redacción de causa
  inesperada.
- Revisión de soporte visible y editable por humano.
- Extracción que fuerza revisión humana y rechaza JSON inválido.
- Degradación del proveedor y protección de doble clic.
- Aprobación humana con proveedor canónico, sin aceptar un proveedor inferido
  no registrado.

## Límite

No se pulsó extracción, aprobación ni pago desde la interfaz, para no crear
CxP ni movimientos. P109-002 permanece parcial hasta ejecutar el flujo
autenticado con un soporte reversible, rol no global, proveedor disponible y
matriz A/B de empresa.
