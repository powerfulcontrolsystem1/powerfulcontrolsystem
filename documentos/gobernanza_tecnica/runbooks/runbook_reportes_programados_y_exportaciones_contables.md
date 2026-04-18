# Runbook: reportes programados y exportaciones contables

Fecha: 2026-04-18
Estado: vigente

## Sintomas cubiertos

- una programacion de reportes no vuelve a ejecutarse o queda con `proximo_ejecutado_en` incoherente.
- la exportacion `csv`, `txt`, `xls` o `pdf` falla aunque el dataset exista.
- la consistencia multiformato devuelve alertas o `inconsistente`.
- una plantilla asociada deja de aplicar o rompe una exportacion.
- la suite completa funciona en `json` pero falla cuando se intenta exportar sin dataset a formato tabular.

## Alcance

Aplica al modulo empresarial de reportes sobre plantillas, programaciones, ejecuciones, exportacion multiformato y validacion de consistencia.

## Fuentes de evidencia

- `backend/handlers/reportes.go`
- `backend/handlers/reportes_programacion.go`
- `backend/db/reportes_programacion.go`
- tablas `empresa_reportes_programaciones`, `empresa_reportes_plantillas`, `empresa_reportes_ejecuciones`
- salida de `action=validar_consistencia`
- salida de `action=ejecuciones`

## Verificaciones iniciales

1. Confirmar `empresa_id`, `dataset_key` y si existe `programacion_id` afectado.
2. Verificar que el dataset exista en el catálogo real del backend.
3. Si la exportación es tabular, confirmar que se esté enviando `dataset`; sin dataset solo se permite `json` de suite.
4. Revisar si la programación tiene `activa=1`, `frecuencia`, `hora_envio`, `timezone` y `proximo_ejecutado_en` válidos.
5. Si hay plantilla, confirmar que `template_codigo` y `template_version` sigan resolviendo una plantilla vigente o existente.
6. Consultar la última fila en `empresa_reportes_ejecuciones` y revisar `estado_ejecucion`, `consistencia_estado`, `error_detalle` y `salida_resumen_json`.

## Causas probables

- `dataset_key` inexistente o mal escrito.
- exportación tabular sin `dataset`.
- `template_codigo` huérfano o `template_version` inexistente.
- `hora_envio` inválida o parámetros de programación mal normalizados.
- diferencias de conteo entre dataset base y formatos tabulares.
- intento de usar un formato no soportado.

## Acciones de recuperacion

1. Repetir primero `action=dataset` para confirmar que el dataset base se construye sin error antes de culpar al exportador.
2. Si la falla ocurre solo al exportar, revisar `format` y confirmar que esté dentro de `json`, `csv`, `txt`, `xls` o `pdf`.
3. Si la exportación es de suite, mantener `format=json`; cualquier formato tabular exige `dataset` explícito.
4. Si hay plantilla asociada, probar el mismo dataset sin `template_codigo` para aislar si el problema está en columnas o configuración de plantilla.
5. Ejecutar `action=validar_consistencia` con los mismos formatos de la programación para identificar si el problema es de generación, de estructura o de conteo de filas.
6. Si una programación dejó de correr bien, revalidar `frecuencia`, `hora_envio`, `timezone`, `destinatarios` y `parametros_json`, y luego lanzar `action=ejecutar_programacion` manualmente.
7. Si el resultado queda `completado_con_alertas`, revisar `consistencia_detalle_json` y `hash_ultima_ejecucion` antes de aceptar la salida como correcta.

## Validacion posterior

- `action=dataset` responde con el dataset esperado.
- `action=export` genera el archivo correcto para el formato pedido.
- `action=validar_consistencia` devuelve `consistente=true` o alerta entendible y trazable.
- `action=ejecutar_programacion` inserta fila en `empresa_reportes_ejecuciones` y actualiza `ultimo_ejecutado_en`, `proximo_ejecutado_en` y `hash_ultima_ejecucion`.
- el equipo puede distinguir entre error de dataset, error de plantilla, error de formato y alerta de consistencia.

## Notas operativas

1. PDF es una salida resumida; puede truncar detalle y recomendar `csv` o `json` para volumen alto.
2. `xls` es realmente TSV compatible con Excel; si el problema es visual, validar también `csv`.
3. La consistencia de `csv` y `xls` compara conteo de filas detectado contra `RowCount`; un desvío ahí no significa necesariamente que el dataset base esté roto, pero sí que la exportación tabular requiere revisión.

## Contratos relacionados

- `documentos/gobernanza_tecnica/contratos/contrato_reportes_contables_financieros_y_exportacion_multiformato.md`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_soporte_remoto_sesiones_y_dispositivos.md`