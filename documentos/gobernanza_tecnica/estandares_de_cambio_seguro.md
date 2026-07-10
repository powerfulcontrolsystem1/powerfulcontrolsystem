# Estandares de cambio seguro

Fecha: 2026-04-18
Estado: obligatorio

## Principios base

1. `empresa_id` es frontera obligatoria de aislamiento en toda operacion empresarial.
2. PostgreSQL en VPS es el runtime productivo canonico; motor legado retirado no define el comportamiento objetivo del sistema.
3. Los flujos publicos con efectos persistentes deben validar contexto esperado y ser idempotentes.
4. Las consultas o escrituras sobre esquemas evolutivos deben tolerar instalaciones legacy cuando el modulo ya lo requiera.
5. La documentacion no debe afirmar la existencia de rutas, archivos o artefactos inexistentes.

## Lectura obligatoria antes de cambiar

### Cambios de backend o arquitectura

Leer:

- `documentos/descripcion_del_proyecto`
- `documentos/diagramas/estructura_del_codigo.md`
- ADR aplicable en `documentos/gobernanza_tecnica/adr/`

### Cambios de base de datos o consultas

Leer:

- `documentos/descripcion_del_proyecto`
- `documentos/estructura_bd.md`
- `documentos/diagramas/estructura_del_codigo.md`

### Cambios de frontend con flujo critico

Leer:

- `documentos/descripcion_del_proyecto`
- contrato tecnico aplicable en `documentos/gobernanza_tecnica/contratos/`
- runbook aplicable si el cambio nace de una falla real

### Cambios en pagos y licencias

Leer ademas:

- `documentos/gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_checkout_licencias.md`

## Reglas por tipo de cambio

### Si cambias rutas o handlers

- documentar request, response, errores y side effects.
- validar wrappers, permisos y alcance de `empresa_id`.
- no ampliar superficie publica sin documentar contrato tecnico.

### Si cambias tablas, migraciones o queries

- verificar `documentos/estructura_bd.md`.
- no asumir columnas nuevas sin saneamiento o migracion clara si el modulo ya convive con legado.
- si el cambio afecta PostgreSQL, validar que no dependa de `LastInsertId` ni de comportamiento exclusivo de motor legado retirado.

### Si cambias flujos publicos de pago

- no confiar solo en `transaction_id` o `reference` cuando exista contexto esperado de empresa/licencia.
- documentar idempotencia y reintentos.
- documentar correo, webhook, polling y persistencia de pagos como side effects separados.

### Si cambias estaciones, sensores o carritos enlazados

- mantener trazabilidad por estacion, carrito y `empresa_id`.
- no romper la identidad canonica `EST-empresa-estacion` / `ESTACION_<id>`.
- documentar claramente que indicador visual representa sensor y cual representa estado de carrito, si ambos existen.

### Si cambias permisos o visibilidad

- actualizar siempre `documentos/descripcion_de_modulos` y `documentos/matriz_roles_permisos_pos_multiempresa.md`.
- verificar wrappers, paneles visibles y rutas publicas/protegidas.

### Si cambias firmas documentales o exportes regulatorios

- no tratar exportes como sustituto automático de la evidencia documental versionada.
- si existe firma, mantener trazabilidad entre documento vigente, `hash_archivo`, `hash_firma`, firmante y fecha.
- si un reporte/exporte respalda un flujo sensible, documentar si es salida informativa o evidencia regulatoria reconciliable.
- enlazar siempre el cambio con el contrato del repositorio documental, el de interoperabilidad documental y el de reportes si el flujo los toca.

## Validacion minima obligatoria

### Cambio de backend

- diagnostico del editor sin errores en archivos tocados.
- pruebas focalizadas del modulo si existen.
- compilacion o smoke test del conjunto afectado.

### Cambio de base de datos

- diagnostico del editor sin errores.
- prueba de lectura/escritura del modulo afectado.
- validacion de compatibilidad con PostgreSQL cuando aplique.

### Cambio de frontend critico

- diagnostico del editor sin errores.
- validacion de navegacion y parametros minimos.
- si el flujo depende de backend, confirmar contrato o pruebas del backend correspondiente.

## Trazabilidad documental obligatoria

Cuando el cambio crea o modifica un flujo tecnico, actualizar en la misma iteracion:

1. `documentos/descripcion_del_proyecto` si cambia alcance o flujo principal.
2. `documentos/descripcion_de_modulos` si cambia un modulo funcional.
3. `documentos/matriz_roles_permisos_pos_multiempresa.md` si cambia permisos o visibilidad.
4. `documentos/descripcion_de_archivos` si se crean, cambian o eliminan archivos.
5. `documentos/historial_de_cambios`.
6. `CHANGELOG.md`.

## Guardrails especificos para Codex

- no inferir nuevas reglas de negocio si no aparecen en las fuentes canonicas.
- no crear archivos de documentacion paralelos si ya existe una fuente canonica del mismo tema.
- no introducir dependencias externas en Go sin autorizacion explicita y justificacion.
- no mover logica critica a `main.go`; mantener el arranque conciso.
- no documentar una ruta como estable si aun no esta registrada o validada.
