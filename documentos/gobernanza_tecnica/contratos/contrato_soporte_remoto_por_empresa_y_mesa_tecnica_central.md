# Contrato tecnico: soporte remoto por empresa y mesa tecnica central

Fecha: 2026-04-18
Estado: vigente

## Alcance

Este contrato cubre el modulo de soporte remoto empresarial, el portal publico del agente o visor remoto y la mesa tecnica central super. Incluye configuracion por empresa, limites de plan, dispositivos, sesiones, tokens de visualizacion, heartbeat del agente, exportacion de sesiones y operacion consolidada multiempresa.

## Endpoints cubiertos

### Panel empresa

- `GET /api/empresa/soporte_remoto?empresa_id={id}&action=config`
- `POST /api/empresa/soporte_remoto?empresa_id={id}&action=config`
- `GET /api/empresa/soporte_remoto?empresa_id={id}&action=dispositivos`
- `GET /api/empresa/soporte_remoto?empresa_id={id}&action=detalle_dispositivo&id={dispositivo_id}`
- `POST /api/empresa/soporte_remoto?empresa_id={id}&action=crear_dispositivo`
- `PUT|PATCH /api/empresa/soporte_remoto?empresa_id={id}&action=actualizar_dispositivo&id={dispositivo_id}`
- `PUT|PATCH /api/empresa/soporte_remoto?empresa_id={id}&action=activar_dispositivo&id={dispositivo_id}`
- `PUT|PATCH|DELETE /api/empresa/soporte_remoto?empresa_id={id}&action=desactivar_dispositivo&id={dispositivo_id}`
- `GET /api/empresa/soporte_remoto?empresa_id={id}&action=sesiones`
- `POST /api/empresa/soporte_remoto?empresa_id={id}&action=solicitar_sesion`
- `POST|PUT|PATCH /api/empresa/soporte_remoto?empresa_id={id}&action=aprobar_sesion`
- `POST|PUT|PATCH /api/empresa/soporte_remoto?empresa_id={id}&action=finalizar_sesion`
- `GET /api/empresa/soporte_remoto?empresa_id={id}&action=resolver_visualizacion&codigo_sesion={codigo}&token={token}`
- `GET /api/empresa/soporte_remoto?empresa_id={id}&action=export_sesiones&format={json|csv|txt|xls|pdf}`

### Portal publico / agente remoto

- `GET /api/public/soporte_remoto?empresa_id={id}&action=resolver_acceso_publico&codigo_sesion={codigo}&token={token}`
- `POST /api/public/soporte_remoto?action=heartbeat_dispositivo`
- `POST /api/public/soporte_remoto?action=aprobar_sesion&codigo_sesion={codigo}`
- `POST /api/public/soporte_remoto?action=finalizar_sesion&codigo_sesion={codigo}`

### Mesa tecnica central super

- `GET /super/api/soporte_remoto?action=empresas`
- `GET /super/api/soporte_remoto?action=dispositivos&empresa_id={id}`
- `GET /super/api/soporte_remoto?action=sesiones&empresa_id={id}`
- `GET /super/api/soporte_remoto?action=reporte&format={json|csv|txt|xls|pdf}`
- `POST|PUT|PATCH /super/api/soporte_remoto?action=solicitar_sesion`
- `POST|PUT|PATCH /super/api/soporte_remoto?action=aprobar_sesion`
- `POST|PUT|PATCH /super/api/soporte_remoto?action=finalizar_sesion`

## Entradas obligatorias

### Configuracion por empresa

- `empresa_id`

### Alta de dispositivo

- `empresa_id`
- `nombre_equipo`
- al menos uno de:
  - `stream_url`
  - `rustdesk_device_id`

### Solicitud de sesion

- `empresa_id`
- `dispositivo_id`

### Visualizacion y portal publico

- `empresa_id`
- `codigo_sesion`
- `token`

### Heartbeat del agente

- `empresa_id`
- `codigo_dispositivo`

## Entradas opcionales relevantes

- `proveedor_preferido`
- `modo_operacion`
- `requiere_aprobacion_operador`
- `auto_cerrar_minutos`
- `max_conexiones_mes`
- `max_minutos_mes`
- `max_dispositivos`
- `portal_publico_habilitado`
- `rustdesk_server_host`
- `rustdesk_server_key`
- `cliente_windows_url`
- `cliente_linux_url`
- `carpeta_transferencia`
- `alias_operativo`
- `ubicacion`
- `sistema_operativo`
- `agente_version`
- `acceso_pin`
- `acceso_publico_habilitado`
- `operador_nombre`
- `operador_email`
- `motivo`
- `duracion_min`
- `observaciones`

## Normalizaciones canonicas

### Proveedor

- `novnc`
- `rustdesk_web`
- `rustdesk_oss`
- `guacamole`
- `custom_url`

### Modo de operacion

- `agente_web`
- `agente_local`
- `cliente_local`
- `hibrido`

### Estado de conexion de dispositivo

- `online`
- `offline`
- `intermitente`

### Estado de sesion

- `pendiente`
- `aprobada`
- `activa`
- `finalizada`
- `rechazada`
- `expirada`

## Persistencia canonica

- `empresa_soporte_remoto_configuracion`
- `empresa_soporte_remoto_dispositivos`
- `empresa_soporte_remoto_sesiones`

Campos estructurales clave:

- aislamiento por `empresa_id`
- trazabilidad por `codigo_dispositivo`
- trazabilidad por `codigo_sesion`
- `token_visualizacion_hash`
- `acceso_pin_hash`
- `bloqueada_por_limite`
- `duracion_minutos_solicitada`
- `duracion_minutos_consumida`
- `ultimo_heartbeat`

## Invariantes

1. Todo dispositivo, configuracion y sesion queda aislado por `empresa_id`.
2. Un dispositivo no puede registrarse sin `nombre_equipo` ni sin al menos `stream_url` o `rustdesk_device_id`.
3. El PIN del dispositivo no se persiste en texto plano; se almacena como hash.
4. La clave RustDesk del dispositivo no se persiste en texto plano; se cifra antes de guardar.
5. Una sesion solo puede crearse sobre un dispositivo de la misma empresa.
6. Si el soporte remoto de la empresa esta deshabilitado, no se puede crear sesion nueva.
7. Los limites de plan por mes y por dispositivos prevalecen sobre cualquier solicitud de alta o sesion.
8. Cuando una solicitud excede el plan, el sistema registra el intento bloqueado como sesion rechazada con `bloqueada_por_limite=1`.
9. El uso mensual se calcula sobre el mes corriente y distingue sesiones validas de intentos bloqueados.
10. El token de visualizacion se entrega solo al crear la sesion; luego se valida contra `token_visualizacion_hash`.
11. `resolver_visualizacion` solo concede acceso cuando la sesion esta `activa` o `aprobada`.
12. `resolver_acceso_publico` exige ademas que `portal_publico_habilitado=1` en la configuracion y `acceso_publico_habilitado=1` en el dispositivo.
13. El agente remoto solo puede aprobar o finalizar una sesion que pertenezca a su propio dispositivo autenticado por PIN.
14. El heartbeat actualiza estado del dispositivo y puede fallar con `dispositivo no autorizado` si el PIN o el codigo no coinciden.
15. La mesa tecnica super opera sobre empresas existentes y consolida uso multiempresa, pero no rompe el aislamiento interno de cada empresa.
16. La exportacion de sesiones usa el mismo writer multiformato canonico de reportes.

## Salidas y errores esperados

### Empresa

- `200` para consultas, cambios de estado, resolucion de visualizacion y exportacion valida
- `201` para alta de dispositivo o sesion
- `400` por `empresa_id`, `id`, `codigo_sesion`, `codigo_dispositivo`, correo o JSON invalidos
- `401` o `403` cuando el acceso/token/dispositivo no corresponde al contexto permitido
- `404` si el dispositivo o la sesion no existen en la empresa
- `412` cuando el soporte esta deshabilitado o el plan fue excedido
- `500` ante error interno no controlado

Mensajes observables relevantes:

- `El soporte remoto esta deshabilitado para esta empresa`
- `dispositivo no autorizado`
- `sesion/token invalido`
- `portal publico deshabilitado`
- `sesion no corresponde al dispositivo`

### Super

- `200` para resúmenes, listados, reporte consolidado y cambios de estado validos
- `201` para solicitud centralizada de sesion
- `400` por parametros invalidos
- `412` cuando el plan de la empresa objetivo no permite abrir la sesion

## Side effects obligatorios

- calculo de uso mensual (`sesiones_mes`, `minutos_consumidos_mes`, `intentos_bloqueados_mes`)
- generacion de `codigo_dispositivo` automatica cuando no se provee
- generacion de `codigo_sesion` y token de visualizacion al crear sesion
- expiracion operativa basada en `auto_cerrar_minutos` y duracion solicitada
- enmascaramiento de `stream_url` en datasets exportados

## Evidencia tecnica minima

- `backend/handlers/soporte_remoto.go`
- `backend/handlers/super_soporte_remoto.go`
- `backend/db/soporte_remoto.go`
- `backend/db/soporte_remoto_test.go`
- `backend/handlers/super_soporte_remoto_test.go`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_soporte_remoto_sesiones_y_dispositivos.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_reportes_programados_y_exportaciones_contables.md`

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`