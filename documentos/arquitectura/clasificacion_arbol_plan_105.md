# Clasificacion del arbol de trabajo - P105-001

Fecha: 2026-07-22.

## Proposito

La compuerta de release exige un arbol limpio y un SHA unico. Esta clasificacion
evita incluir en un commit del Plan 105 cambios concurrentes cuyo alcance no se
ha revisado en este bloque de trabajo. Es una guia de separacion; no autoriza
`git add`, commit, reset, limpieza ni despliegue.

## Estado observado

`git status --short` reporta 69 entradas: 57 modificadas y 12 nuevas. El
candidato `72fa7007` tiene `origin/main` como ancestro y upstream configurado,
pero no puede convertirse aun en release por este arbol pendiente y porque no
hay digests inmutables para API, migrador y worker.

## Grupo A: trabajo confirmado del Plan 105

Estos cambios pertenecen a los lotes ya validados localmente y deben revisarse
como una serie coherente antes de crear el SHA candidato:

- Retiro de DDL HTTP/readiness: `backend/db/chat_tareas.go`,
  `empresa_estacion_prefs.go`, `energia_solar.go`, `grafologia.go`,
  `hoja_vida_operativa.go`, `reportes_programacion.go`, `reservas_hotel.go`,
  `ubicacion_gps.go`, `schema_readiness_test.go` y el manifiesto generado.
- Handlers asociados de estacion, GPS/taxi, grafologia, energia, hoja de vida,
  reservas, reportes, Chat/Tareas y mensajes privados.
- Redaccion de errores: agente fiscal, orquestador IA, chat publico,
  seguridad VPS, alertas, correos masivos y mantenimiento, junto con sus siete
  pruebas de regresion `*_public_error_test.go`.
- Gates/documentacion P105: `plan_105.md`, matriz de piloto, inventarios de
  bootstrap/runtime/errores/goroutines, matriz A/B, historial, migracion,
  contexto general, descripcion de archivos, `release_gate.ps1`,
  `vps_restore_validation.ps1` y `release_manifest.mjs`.

Antes de comprometer este grupo, Terra debe regenerar los inventarios, ejecutar
`go test ./... -count=1`, `go vet ./...`, `git diff --check` y revisar cada diff
contra el inventario generado. Debe crear un commit o PR separado por frente si
el review muestra que el retiro de DDL y la redaccion no forman una unidad
revisable.

## Grupo B: cambios concurrentes que se deben excluir del commit P105

No incluir sin un review independiente: worker de registro de negocio,
empresa compartida, finanzas Bre-B/QR, Nextcloud (backend, frontend y
documentacion), snapshots VPS, y los cambios de modulos/estructura/flujo
asociados. Pueden ser validos, pero no fueron implementados ni validados como
parte de los gates P105.

## Grupo C: archivos con alcance que requiere atribucion antes de decidir

`chat_flotante_config.go`, `empresa_preconfiguracion.go`,
`panel_empresa_config.go`, `contexto_codex.md`, `docker_vps_operacion.md` y
`documentos/releases/README.md` deben revisarse por diff y por autor/objetivo
antes de asignarlos a un lote. No se presume propiedad por coincidencia de
fecha ni se mezclan para limpiar el arbol.

## Secuencia segura de consolidacion

1. Revisar y aprobar el Grupo A por diff.
2. Aislar Grupo B/C en commits existentes, otra rama o una lista de trabajo
   aprobada; nunca descartar cambios para obtener un arbol limpio.
3. Ejecutar CI requerido sobre el SHA resultante.
4. Construir una vez API, migrador y worker; registrar sus digests.
5. Repetir `release_manifest.mjs --strict --check`; solo entonces crear
   tag/PR candidato y continuar a staging.
