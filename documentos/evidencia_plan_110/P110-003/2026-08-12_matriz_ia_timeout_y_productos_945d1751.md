# P110-003 - matriz IA, cancelacion y boton de Productos

Fecha: 2026-08-12
Candidato revisado en staging: `945d1751`
Empresa: PCS de pruebas autorizadas

## Evidencia real en staging

- El asistente central respondio una consulta neutra con el texto solicitado.
- En modo agente genero una propuesta de alta de producto y mostro confirmacion
  obligatoria; se pulso `Cancelar` y el catalogo posterior no contenia el
  producto propuesto. No hubo mutacion empresarial.
- Los botones de Compras, Egresos e Ingresos rechazaron visualmente la accion
  sin adjunto y solicitaron primero una imagen o PDF.
- El boton `Cargar carta/precios con IA` de Productos anuncio la apertura del
  asistente, pero no traslado el prompt al drawer superior. La causa fue el uso
  de `window.postMessage` dentro del iframe anidado.
- El intento de pulsar Grafologia IA sin imagen no concluyo por timeout de la
  interaccion del navegador. El boton usaba un `alert` bloqueante para su guarda
  sin imagen y no se contabiliza como aprobado en staging.

## Correcciones locales, aun no desplegadas

- Productos dirige el mensaje al `window.top` cuando se ejecuta dentro del
  shell empresarial y conserva `window` cuando la pagina es independiente.
- La extraccion de soportes limita la llamada al proveedor a 90 segundos,
  distingue cancelacion de timeout, devuelve la reserva diaria y conserva el
  estado anterior del soporte.
- Se agrego una metrica separada de timeout y pruebas de propagacion de contexto
  hasta la solicitud HTTP al proveedor.
- `tools/qa_ai_button_inventory.mjs` inventario 20 controles IA empresariales de
  forma reproducible. El inventario no equivale a su aprobacion funcional.
- Grafologia presenta ahora las guardas y errores GPT dentro del panel accesible
  existente, sin dialogs bloqueantes. La regresion contractual conserva los dos
  mensajes previos y exige que no vuelvan a `alert`.
- Finanzas/CxP ya no usa un `confirm` bloqueante al iniciar la carga desde la
  vista CxC. Cambia visiblemente a CxP, explica que los datos siguen editables y
  abre el selector. La confirmacion humana de negocio permanece en Guardar.

## Resultado y limite

Las pruebas focalizadas de contexto, metricas y contratos aprobaron. Tambien
aprobaron `go test ./...`, `go vet ./handlers ./db ./utils`, la auditoria
profesional, el parseo del inventario IA, `git diff --check` y el escaneo de
secretos limitado a los archivos cambiados. Como por
instruccion no se ejecuto `rs`, las correcciones locales no se han repetido en
staging. P110-003 permanece parcial hasta desplegar un candidato limpio, probar
los 20 controles por rol, proveedor caido, timeout real, cancelacion visual y
evals A/B. No se autoriza produccion.
