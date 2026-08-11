# P109-002 - cancelacion IA y validacion de adjuntos

Fecha: 2026-08-08
Ambiente: candidato local `codex/p109-batch-no-pr`
Resultado: **PASS local / pendiente de staging y PCS**

## Controles incorporados

- Extensión y firma real deben corresponder para PNG, JPEG, WebP y PDF.
- XML se analiza en modo estricto y rechaza DTD, entidades, instrucciones de
  procesamiento no XML, profundidad excesiva y elementos activos.
- El MIME persistido se deriva del tipo validado, no de la cabecera del cliente.
- Si la base rechaza la fila tras escribir, se elimina solo el archivo nuevo
  resuelto dentro del almacenamiento privado empresarial.
- El botón `Cancelar IA` usa `AbortController`; el backend pasa `r.Context()` a
  Responses API con `NewRequestWithContext`, conserva liberación del lock y
  devuelve la reserva avanzada sin permitir contadores negativos.
- Retención ofrece una vista previa no destructiva de eliminados antiguos,
  excluye contabilizados/convertidos y calcula bytes con rutas privadas seguras.
- ClamAV usa el protocolo INSTREAM con límite de lectura y deadlines. Malware
  bloquea antes de escribir y el modo obligatorio falla cerrado si clamd cae.

## Evidencia automatizada

- Firmas válidas y MIME canónico: PASS.
- PDF/imagen suplantados: rechazados.
- XML XXE/DTD, script, instrucción y XML roto: rechazados.
- Limpieza confinada del adjunto sin fila: PASS.
- Proveedor local lento y cancelación de contexto: PASS.
- Parseo de retención 1-3650 y bytes privados seguros: PASS.
- Integración PostgreSQL de retención: preparada, pendiente de DSN.
- Clamd simulado: limpio PASS, firma malware bloqueada, modo opcional PASS y
  modo obligatorio sin endpoint bloqueado.
- Revisión visual local: siete acciones y vista previa visibles; `Cancelar IA`
  inicia deshabilitado, retención muestra 90 días y mensaje no destructivo.
- Ancho móvil efectivo 480 px: todos los botones/filtros quedaron dentro del
  viewport, sin scroll horizontal y sin errores de consola.
- `go test ./... -count=1`: PASS.
- `go vet ./...`: PASS.
- Sintaxis JavaScript y `git diff --check`: PASS.
- Preflight profesional Full/Strict: 22/22 compuertas PASS; reporte
  `documentos/reportes_profesionales/preflight_20260808_195938.md`.

## Límites

- No se usó proveedor IA real, credenciales PCS ni datos empresariales.
- No existe daemon ClamAV real certificado en este equipo; el PASS corresponde
  al protocolo simulado, no al servicio del VPS.
- La vista previa no ejecuta purga física.
- No hubo PR, push, despliegue ni cambio en producción/staging.

## Veredicto

Los controles locales avanzan P109-002, P109-008 y P109-009, pero las tres fases
siguen parciales. El Plan 109 conserva **53,3 % de implementación**, **0 % de
certificación del arreglo local** y **NO-GO**.
