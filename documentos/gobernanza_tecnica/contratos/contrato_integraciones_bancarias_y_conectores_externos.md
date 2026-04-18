# Contrato tecnico: integraciones bancarias y conectores externos

Fecha: 2026-04-18
Estado: vigente

## Alcance

Este contrato cubre el comportamiento operativo de las integraciones empresariales de APIs externas y conectores bancarios, incluyendo consulta de estado, verificacion de conectividad, sincronizacion manual, rotacion de referencias de credencial y monitoreo estructurado.

## Endpoints cubiertos

### APIs externas

- `GET /api/empresa/integraciones/apis?empresa_id={id}`
- `POST /api/empresa/integraciones/apis`
- `PUT /api/empresa/integraciones/apis`
- `DELETE /api/empresa/integraciones/apis?empresa_id={id}&id={registro_id}`
- `GET /api/empresa/integraciones/apis?empresa_id={id}&action=estado`
- `GET /api/empresa/integraciones/apis?empresa_id={id}&action=health_check`
- `POST /api/empresa/integraciones/apis?empresa_id={id}&action=sync_manual`
- `PUT /api/empresa/integraciones/apis?action=rotar_credencial`
- `GET /api/empresa/integraciones/apis?empresa_id={id}&action=monitoreo`

### Bancos y conectores bancarios

- `GET /api/empresa/integraciones/bancos?empresa_id={id}`
- `POST /api/empresa/integraciones/bancos`
- `PUT /api/empresa/integraciones/bancos`
- `DELETE /api/empresa/integraciones/bancos?empresa_id={id}&id={registro_id}`
- `GET /api/empresa/integraciones/bancos?empresa_id={id}&action=estado`
- `GET /api/empresa/integraciones/bancos?empresa_id={id}&action=health_check`
- `POST /api/empresa/integraciones/bancos?empresa_id={id}&action=sync_manual`
- `PUT /api/empresa/integraciones/bancos?action=rotar_credencial`
- `GET /api/empresa/integraciones/bancos?empresa_id={id}&action=monitoreo`

## Persistencia canonica

- `empresa_integraciones_apis`
- `empresa_integraciones_bancos`

Campos funcionales relevantes:

- `codigo`
- `estado_integracion`
- `estado`
- `base_url` o `api_endpoint`
- `api_key_ref` o `credencial_ref`
- `ultima_sincronizacion` o `ultima_conciliacion`
- `respuesta_ultimo_sync`
- `observaciones`

## Entradas obligatorias

### Alta o actualizacion basica

- `empresa_id`
- `codigo`
- endpoint del conector, usando `base_url` en APIs o `api_endpoint` en bancos

### Consulta de estado, health_check, sync_manual y monitoreo

- `empresa_id`
- `id` cuando se consulta un registro puntual

### Rotacion de credencial

- `empresa_id`
- `id`
- `nueva_credencial_ref`

## Entradas opcionales relevantes

- `limit`
- `offset`
- `q`
- `include_inactive`
- `validar`
- `latencia_alerta_ms`
- `stale_hours`
- `persistir`

## Salidas funcionales

### Action `estado`

- `200` con:
  - `ok`
  - `empresa_id`
  - `modulo`
  - `items[]`

Cada item expone como minimo:

- `id`
- `codigo`
- `nombre`
- `endpoint`
- `estado_integracion`
- `estado_registro`
- `ultima_ejecucion`
- `respuesta_ultimo_sync` si el modulo la soporta

### Action `health_check` y `sync_manual`

- `200` con:
  - `ok`
  - `empresa_id`
  - `modulo`
  - `accion`
  - `ejecutado_en`
  - `resultados[]`
  - `errores[]`

Cada resultado expone al menos:

- `id`
- `endpoint`
- `http_status`
- `reachable`
- `latency_ms`
- `estado_integracion`
- `message`
- `updated`

### Action `rotar_credencial`

- `200` con:
  - `ok`
  - `empresa_id`
  - `id`
  - `modulo`
  - `accion`
  - `campo_credencial`
  - `rotada`
  - `item`
  - `validacion` cuando se solicita validar

### Action `monitoreo`

- `200` con:
  - `ok`
  - `empresa_id`
  - `modulo`
  - `ejecutado_en`
  - `items[]`
  - `alertas[]`
  - `errores_persistencia[]`
  - agregados de salud del lote

## Estados y alertas canonicas

### Estado de integracion

- `activa`
- `inactiva`
- `error`

### Alertas de monitoreo observables

- `endpoint_invalido`
- `sin_conectividad`
- `latencia_alta`
- `sin_sync_reciente`
- `sync_atrasada`
- `estado_error`

## Invariantes

1. Toda operación queda aislada por `empresa_id`.
2. `action=estado`, `health_check`, `sync_manual`, `rotar_credencial` y `monitoreo` trabajan sobre el mismo registro empresarial y no cruzan conectores entre empresas.
3. Si el endpoint normalizado queda vacío, `estado_integracion` debe considerarse `inactiva`.
4. `health_check` y `sync_manual` actualizan `estado_integracion` según el probe real y registran un snapshot en `respuesta_ultimo_sync` cuando el módulo lo soporta.
5. `sync_manual` además actualiza el campo temporal del módulo: `ultima_sincronizacion` para APIs o `ultima_conciliacion` para bancos.
6. La rotación de credencial nunca debe aceptar secreto plano; la referencia debe validarse antes de persistirse.
7. Formatos como `env:` y `vault:` son válidos para referencias de credencial; una cadena libre como token directo debe rechazarse.
8. Cuando la referencia nueva es igual a la actual, la operación responde `rotada=false` y no modifica el registro.
9. Una rotación válida reinicia el `estado_integracion` a `inactiva` y fuerza revalidación posterior.
10. Si se solicita `validar=true` en la rotación, el backend ejecuta un probe inmediato y refleja el resultado en `validacion` y en el registro persistido.
11. El monitoreo debe poder correr sin persistir cambios; solo si `persistir=1` aplica actualización del snapshot estructurado.
12. Un `401` o `403` remoto no implica necesariamente `sin_conectividad`; el conector puede seguir siendo alcanzable a nivel de red.
13. El monitoreo usa umbrales configurables de latencia y antigüedad de sync; si el caller no los envía, el backend usa defaults operativos.
14. Los errores de persistencia parcial de un registro no deben romper el lote completo; se reportan en `errores_persistencia` o `errores`.

## Errores esperados y tratamiento

- `400` si faltan `empresa_id` o `id`, o si `limit`, `offset`, `latencia_alerta_ms` o `stale_hours` son inválidos.
- `400` si la rotación recibe referencia de credencial inválida.
- `404` si el registro no existe para la empresa.
- `500` si falla la consulta base de los registros o la preparación del monitoreo.

Mensajes observables relevantes:

- `id required`
- `id invalido`
- `registro no encontrado`
- `No se pudo consultar estado de integracion`
- `No se pudo preparar verificacion`
- `No se pudo preparar monitoreo de integraciones`
- `el modulo no soporta rotacion de credenciales`

## Side effects obligatorios

- actualización del estado de integración tras probes y validaciones.
- almacenamiento de snapshot estructurado en `respuesta_ultimo_sync`.
- reseteo de último sync al rotar credenciales.
- auditoría en `observaciones` cuando el módulo soporta esa columna.

## Evidencia tecnica minima

- `backend/handlers/modulos_faltantes.go`
- `backend/handlers/modulos_faltantes_test.go`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_contingencias_integraciones_bancarias_y_conectores.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_reconciliacion_documental_fiscal_y_contable_externa.md`

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`