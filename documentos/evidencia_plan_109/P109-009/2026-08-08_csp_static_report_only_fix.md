# P109-009 - correccion de CSP Report-Only para recursos estaticos

Fecha: 2026-08-08  
Alcance: candidato de staging Plan 109; sin cambio en produccion.

## Hallazgo

La primera promocion del candidato con `PCS_CSP_REPORT_ONLY_STRICT=1` activo
confirmo que el backend recibio la variable, pero `login.html` seguia emitiendo
`unsafe-inline` en `Content-Security-Policy-Report-Only`. La cabecera provenia
del include estatico de Nginx, no del middleware Go que atiende las rutas API.

## Correccion preparada

- La CSP aplicada de Nginx no cambia y conserva la compatibilidad vigente.
- La CSP `Report-Only` estatica elimina `unsafe-inline` de `script-src` y
  `style-src`, y reemplaza los comodines de imagenes/conexiones por origenes
  explicitos revisables.
- La prueba de recursos estaticos ahora extrae la cabecera Report-Only y exige
  que omita compatibilidad inline y destinos amplios, sin confundirla con la
  CSP aplicada.

## Verificacion local

- `go test . -run TestNextcloudFramePolicyUsesExactOrigins -count=1`: PASS.
- `go test ./utils -count=1`: PASS.
- `go vet ./utils`: PASS.
- `git diff --check`: PASS.

## Siguiente compuerta

Se debe construir un nuevo candidato inmutable, promoverlo solo a staging y
comprobar visualmente que las cabeceras aplicada y Report-Only sean distintas:
la primera debe conservar compatibilidad y la segunda no debe contener
`unsafe-inline`. Esta evidencia no autoriza enforcement ni produccion.
