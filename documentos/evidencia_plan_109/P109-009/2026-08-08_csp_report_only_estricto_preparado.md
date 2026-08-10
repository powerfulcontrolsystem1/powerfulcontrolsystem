# P109-009 - CSP estricta en modo Report-Only preparada

Fecha: 2026-08-08
Alcance: candidato local; no se desplegó staging ni producción.

## Cambio

Se separó la política CSP aplicada de la política de observación:

- La cabecera `Content-Security-Policy` conserva compatibilidad con
  `unsafe-inline` para no interrumpir páginas heredadas.
- Cuando `PCS_CSP_REPORT_ONLY_STRICT=1`,
  `Content-Security-Policy-Report-Only` elimina `unsafe-inline` de
  `script-src` y `style-src` sin bloquear la página.
- El valor por defecto es `0` en Compose y en los ejemplos de staging y
  producción. No hay cambio de comportamiento hasta habilitarlo de forma
  explícita en staging.

## Validación

- Prueba unitaria de política normal y política de reporte estricto: PASS.
- `go test ./utils -count=1`: PASS.
- `go vet ./utils`: PASS.
- `go test ./... -count=1`: PASS.
- `git diff --check`: PASS.

El barrido actual identificó 191 páginas HTML con scripts embebidos y 173 con
bloques `<style>`. Por eso el modo estricto se entrega primero como observación,
no como enforcement.

## Activación controlada pendiente

Después de publicar un candidato inmutable, habilitar solo en el entorno
staging `PCS_CSP_REPORT_ONLY_STRICT=1`, revisar las violaciones de login,
panel, carrito, pagos, DIAN, reportes y páginas públicas, migrar los lotes y
solamente entonces retirar la compatibilidad aplicada. No se activa en
producción hasta completar esa matriz visual y funcional.
