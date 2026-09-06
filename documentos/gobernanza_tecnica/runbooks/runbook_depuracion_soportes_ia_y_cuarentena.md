# Runbook: depuracion de soportes IA y cuarentena privada

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- La vigencia del procedimiento no certifica la implementación en staging. Conserva los pendientes A/B, caída, dos réplicas y restore.
- Depurar es irreversible para el archivo: exige alcance autorizado, permiso D, motivo y código exacto.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-08-09

## Sintomas cubiertos

- un soporte permanece en `purga_pendiente` despues de reintentar.
- el diagnostico informa diferente cantidad de registros y archivos.
- existe al menos un pendiente que supera el umbral operativo.
- una depuracion responde conflicto porque otra replica la esta ejecutando.
- el archivo ya no esta disponible, pero la tumba `purgado` aun no finaliza.

## Alcance

Aplica exclusivamente a `/api/empresa/soportes_compras_ia` y a los archivos de
`private_storage/soportes_compras_ia/empresa_<id>`. Es un flujo empresarial
autenticado: diagnostico exige Read y depuracion exige Delete.

## Fuentes de evidencia

- filtro `Depuracion pendiente` de la bandeja.
- accion `cuarentena_preview`, que devuelve conteos, bytes, vencidos y bandera
  de revision sin revelar nombres ni rutas.
- eventos `purgar_iniciar` y `purgar` del soporte.
- logs saneados `[soportes_compras_ia]` por `empresa_id` y `soporte_id`.
- backup/restore vigente del volumen privado y PostgreSQL empresarial.

## Verificaciones iniciales

1. Confirmar empresa, usuario, permiso efectivo y codigo visible del soporte.
2. Ejecutar `Diagnostico de cuarentena`; no inspeccionar ni borrar archivos a
   mano como primer paso.
3. Comparar `registros_pendientes`, `archivos_cuarentena`, `pendientes_vencidos`
   y `requiere_revision`.
4. Abrir el soporte pendiente y revisar eventos. Debe existir como maximo un
   `purgar_iniciar` sin su `purgar` final correspondiente.
5. Confirmar que el soporte no esta contabilizado ni convertido a CxP.
6. Antes de cualquier recuperacion extraordinaria, verificar que existe backup
   restaurable de base y almacenamiento del mismo candidato.

## Causas probables

- caida del proceso despues de mover el archivo y antes de iniciar la base.
- caida despues de marcar `purga_pendiente` y antes de eliminar la cuarentena.
- archivo eliminado y fallo temporal antes de confirmar la tumba.
- doble clic, reintento de red o dos replicas compitiendo por la misma empresa.
- mas de una cuarentena heredada para un mismo archivo; el sistema falla cerrado.
- perdida o alteracion manual del volumen privado fuera del flujo oficial.

## Acciones de recuperacion

1. Si el diagnostico coincide y el registro esta pendiente, seleccionar el
   soporte, usar `Depurar archivo`, registrar motivo y escribir su codigo exacto.
   El reintento es idempotente y reutiliza la fase ya completada.
2. Si aparece “otra depuracion en curso”, esperar la respuesta de la primera
   replica y actualizar la bandeja antes de reintentar una sola vez.
3. Si no coinciden registros y archivos, detener la depuracion masiva y conservar
   evidencia. No elegir una cuarentena arbitraria ni usar borrado manual.
4. Si existen cuarentenas ambiguas, escalar con backup, eventos y logs. La
   correccion debe reconciliar el archivo exacto con el hash/metadato del soporte
   y aprobarse como incidente, nunca por nombre parecido.
5. Si la base indica pendiente pero no existe archivo/cuarentena, reintentar el
   soporte: el flujo finaliza solo el metadato y deja el evento auditable.
6. Si el reintento sigue fallando, detenerse. Restaurar o reparar mediante el
   runbook general de desastre en un entorno aislado antes de tocar staging.

## Validacion posterior

- el soporte queda `purgado`, sin descarga ni acciones operativas.
- existen eventos `purgar_iniciar` y `purgar` con empresa, actor y referencia.
- el diagnostico muestra cero vencidos y conteos coherentes.
- un replay de la misma solicitud no crea nuevos efectos ni errores internos.
- otra empresa no ve registros, conteos ni bytes de la empresa intervenida.
- backup y restore posterior conservan la tumba y no reviven el archivo.

## Prohibiciones

- no usar SQL directo para cambiar `estado` o `archivo_url`.
- no borrar archivos `.purge-*` manualmente.
- no copiar cuarentenas entre directorios `empresa_<id>`.
- no depurar soportes contabilizados, convertidos o sin backup verificable.
- no presentar esta operacion como certificada hasta probar PostgreSQL, dos
  replicas, caidas controladas, A/B y restore en staging.

## Contratos y documentos relacionados

- `documentos/checklist_seguridad_endpoint_multiempresa.md`
- `documentos/soportes_compras_ia.md`
- `documentos/plan_109.md` - P109-008 y P109-010
- `documentos/gobernanza_tecnica/runbooks/runbook_recuperacion_desastre_docker_vps.md`
- `documentos/evidencia_plan_109/P109-008/2026-08-09_purga_segura_soportes_ia_local.md`

## Fuentes y aceptación de la revisión

[soportes_compras_ia.go](../../../backend/handlers/soportes_compras_ia.go), [soportes_compras_ia.md](../../soportes_compras_ia.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../requisitos/especificacion_y_trazabilidad.md)).
