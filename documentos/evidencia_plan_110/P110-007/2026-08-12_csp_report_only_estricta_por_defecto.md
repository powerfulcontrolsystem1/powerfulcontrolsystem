# P110-007 - CSP estricta de observacion por defecto

Fecha: 2026-08-12
Candidato observado en staging: `945d1751`

## Observacion real

- Portada y login respondieron HTTP 200 con HSTS, `nosniff`,
  `SAMEORIGIN` y politica de referencia.
- El frontend estatico publico emitio CSP Report-Only estricta sin
  `unsafe-inline`.
- La API anonima rechazo la solicitud empresarial con HTTP 401, pero el backend
  aun emitio su CSP Report-Only con `unsafe-inline` por el valor predeterminado
  de despliegue.

## Correccion local, aun no desplegada

`deploy/docker-compose.platform.yml` activa por defecto
`PCS_CSP_REPORT_ONLY_STRICT=1`. Solo endurece la cabecera de observacion; no
cambia todavia la CSP aplicada ni rompe compatibilidad. Una prueba contractual
impide que el Compose vuelva silenciosamente al valor anterior.

## Resultado y limite

Falta desplegar, observar reportes legítimos, eliminar handlers inline y pasar
la politica a modo aplicado cuando el conteo sea cero. P110-007 permanece
parcial y el veredicto sigue NO-GO.
