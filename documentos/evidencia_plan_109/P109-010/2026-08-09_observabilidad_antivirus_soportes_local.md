# P109-010 - Observabilidad local del antivirus de soportes

Fecha: 2026-08-09
Entorno: candidato local, sin despliegue ni archivos PCS
Rama: `codex/p109-batch-no-pr`

## Alcance

- El scanner clamd registra contadores de resultados `clean`, `malware`,
  `unavailable` y `bypassed` mediante atomicos seguros entre goroutines.
- `/metrics` expone esos cuatro resultados y los estados `required` y
  `configured`. Solo usa la etiqueta acotada `result`; no publica empresa,
  usuario, archivo, ruta, contenido ni respuesta del proveedor.
- Prometheus alerta cuando el antivirus es obligatorio sin endpoint, cuando
  falla u omite un analisis y cuando bloquea malware.
- Grafana muestra configuracion/obligatoriedad y resultados recientes por job.
- El runbook conserva el adjunto cerrado y prohibe abrirlo, descargarlo o
  desactivar el modo obligatorio como respuesta al incidente.

## Pruebas locales

- Servidor TCP clamd simulado: limpio, malware, obligatorio sin endpoint y
  opcional desactivado.
- Los cuatro contadores avanzaron exactamente una vez en sus casos.
- 64 invocaciones simultaneas incrementaron el resultado `bypassed` sin perder
  eventos.
- Render Prometheus, alertas y dashboard: contrato PASS.
- Go completo, vet y preflight Full/Strict: PASS.

## Limites

Los contadores son por proceso y Prometheus los agrega por job. Falta desplegar
clamd real con firmas actualizadas, usar una muestra EICAR controlada en staging,
observar alerta/recuperacion y verificar receptor externo. P109-008 y P109-010
siguen parciales; no cambia el 53,3 % ni el veredicto NO-GO.
