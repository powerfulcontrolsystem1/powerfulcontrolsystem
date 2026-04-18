# Runbook: contingencias de integraciones bancarias y conectores externos

Fecha: 2026-04-18
Estado: vigente

## Sintomas cubiertos

- `health_check` o `sync_manual` dejan el conector sin actualización visible.
- el monitoreo reporta `endpoint_invalido`, `sin_conectividad`, `latencia_alta`, `sin_sync_reciente`, `sync_atrasada` o `estado_error`.
- una rotación de credencial deja la integración en `inactiva`.
- la empresa usa una referencia de credencial inválida y la rotación retorna error.
- el equipo interpreta un `401` o `403` del endpoint externo como caída total del conector.

## Alcance

Aplica a los endpoints empresariales de integraciones API y bancarias:

- `/api/empresa/integraciones/apis`
- `/api/empresa/integraciones/bancos`

Acciones cubiertas:

- `estado`
- `health_check`
- `sync_manual`
- `rotar_credencial`
- `monitoreo`

## Fuentes de evidencia

- `backend/handlers/modulos_faltantes.go`
- `backend/handlers/modulos_faltantes_test.go`
- tablas `empresa_integraciones_apis` y `empresa_integraciones_bancos`

## Verificaciones iniciales

1. Confirmar `empresa_id`, `id` del conector y si se trata de `apis` o `bancos`.
2. Revisar el endpoint configurado: `base_url` para APIs o `api_endpoint` para bancos.
3. Revisar la referencia de credencial activa: `api_key_ref` o `credencial_ref`.
4. Verificar el último timestamp operativo: `ultima_sincronizacion` o `ultima_conciliacion`.
5. Consultar `action=estado` antes de correr un `health_check` manual para no confundir estado previo con estado recién calculado.

## Causas probables

- endpoint vacío o inválido.
- endpoint alcanzable pero con autenticación externa rechazada.
- credencial rotada a un formato no permitido.
- falta de sincronización reciente, aunque el endpoint todavía responda.
- latencia alta del proveedor externo.
- estado residual del registro en `error` o `inactiva` tras una rotación.

## Acciones de recuperacion

1. Si el endpoint está vacío, corregir primero la URL base; `health_check` no resolverá una configuración nula.
2. Si se sospecha caída, ejecutar `health_check`; este flujo valida alcanzabilidad y latencia, no garantía funcional completa del negocio remoto.
3. Si el proveedor responde `401` o `403`, tratarlo como conectividad alcanzable pero autenticación externa pendiente; no clasificarlo automáticamente como caída de red.
4. Si se necesita reflejar operación manual, usar `sync_manual`; además de sondear el endpoint, actualiza `ultima_sincronizacion` o `ultima_conciliacion`.
5. Si la credencial cambió, usar `rotar_credencial` con una referencia válida y no con secreto plano.
6. Tras una rotación, si se requiere confirmación inmediata, repetir con `validar=true` o correr luego `health_check`.
7. Si el monitoreo marca `sync_atrasada`, revisar primero si el conector realmente depende de una corrida manual periódica antes de abrir incidente de red.

## Restricciones operativas relevantes

1. Las referencias de credencial deben usar formatos compatibles como `env:` o `vault:`; una cadena plana como token directo debe rechazarse.
2. La rotación de credencial reinicia el estado de integración a `inactiva` y puede vaciar el último sync para forzar validación posterior.
3. `monitoreo` puede persistir resultados si se solicita `persistir=1`; si no, solo entrega diagnóstico del momento.
4. Un conector bancario puede quedar `activa` aun si el endpoint respondió `401`, porque el chequeo actual prioriza alcanzabilidad del endpoint sobre éxito de autenticación de negocio.

## Validacion posterior

- `action=estado` devuelve endpoint normalizado y estado coherente con la prueba más reciente.
- `health_check` o `sync_manual` registran respuesta y timestamp cuando corresponda.
- `rotar_credencial` deja la nueva referencia persistida y sin secreto plano en la base.
- `monitoreo` reduce o elimina alertas estructurales una vez corregidos endpoint, sincronización y credencial.

## Alertas esperadas de monitoreo

- `endpoint_invalido`
- `sin_conectividad`
- `latencia_alta`
- `sin_sync_reciente`
- `sync_atrasada`
- `estado_error`

## Contratos relacionados

- `documentos/gobernanza_tecnica/contratos/contrato_interoperabilidad_documental_contable_y_fiscal_externa.md`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_cierre_periodo_y_conciliacion_bancaria.md`