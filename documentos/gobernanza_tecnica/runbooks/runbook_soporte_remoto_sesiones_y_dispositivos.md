# Runbook: soporte remoto, sesiones y dispositivos

Fecha: 2026-04-18
Estado: vigente

## Sintomas cubiertos

- un dispositivo no aparece en línea aunque el agente esté activo.
- el heartbeat responde `dispositivo no autorizado`.
- una sesión no puede crearse o queda rechazada por límite.
- `resolver_visualizacion` o `resolver_acceso_publico` devuelven token inválido, sesión no activa o portal deshabilitado.
- la mesa técnica super ve la empresa pero no logra abrir sesión o finalizarla.

## Alcance

Aplica al módulo empresarial de soporte remoto, al agente público de heartbeat y al flujo super de mesa técnica central.

## Fuentes de evidencia

- `backend/handlers/soporte_remoto.go`
- `backend/handlers/super_soporte_remoto.go`
- `backend/db/soporte_remoto.go`
- `backend/db/soporte_remoto_test.go`
- tablas `empresa_soporte_remoto_configuracion`, `empresa_soporte_remoto_dispositivos`, `empresa_soporte_remoto_sesiones`

## Verificaciones iniciales

1. Confirmar `empresa_id`, `codigo_dispositivo` o `dispositivo_id` y, si aplica, `codigo_sesion`.
2. Revisar la configuración efectiva de soporte remoto: `habilitado`, `proveedor_preferido`, `modo_operacion`, `portal_publico_habilitado`, `auto_cerrar_minutos` y topes del plan.
3. Consultar el uso mensual para validar si la empresa ya agotó dispositivos, conexiones o minutos.
4. Verificar el estado del dispositivo: `estado`, `estado_conexion`, `ultimo_heartbeat` y si `acceso_publico_habilitado` está activo.
5. Si hay problema de visualización, revisar el estado de la sesión y si el token usado es el emitido en la creación original.

## Causas probables

- soporte remoto deshabilitado en la empresa.
- PIN del dispositivo incorrecto o código de dispositivo inválido.
- sesión creada sobre un dispositivo de otra empresa.
- agotamiento de `max_dispositivos`, `max_conexiones_mes` o `max_minutos_mes`.
- sesión en estado `pendiente`, `rechazada`, `finalizada` o `expirada` al momento de intentar visualizar.
- portal público deshabilitado a nivel empresa o dispositivo.

## Acciones de recuperacion

1. Si el dispositivo no reporta online, reenviar `heartbeat_dispositivo` con `empresa_id`, `codigo_dispositivo` y `acceso_pin` correctos.
2. Si el heartbeat falla con autorización, validar el PIN y confirmar que el dispositivo pertenezca a la misma empresa.
3. Si no se puede crear sesión, revisar primero el consumo mensual y los límites activos del plan antes de volver a intentar.
4. Si la sesión fue bloqueada por plan, localizar la fila rechazada con `bloqueada_por_limite=1` y corregir capacidad o política antes de reabrir.
5. Si `resolver_visualizacion` falla, comprobar que la sesión esté `aprobada` o `activa` y que el token corresponda a esa misma sesión.
6. Si el acceso público falla, validar además `portal_publico_habilitado` y `acceso_publico_habilitado` del dispositivo.
7. Si la mesa técnica super no puede operar, repetir la prueba desde la API empresarial para distinguir si el problema está en la capa central o en la empresa objetivo.

## Validacion posterior

- el dispositivo queda `online` y actualiza `ultimo_heartbeat`.
- la sesión se crea o cambia de estado correctamente.
- `resolver_visualizacion` devuelve acceso permitido cuando la sesión está lista.
- el portal público solo entrega acceso cuando empresa, dispositivo y sesión cumplen las condiciones.
- la mesa técnica super refleja el mismo estado que la operación empresarial.

## Notas operativas

1. Una sesión puede existir aunque esté bloqueada por límite; eso no significa que el flujo haya quedado habilitado.
2. `stream_url` se enmascara en datasets exportados; para diagnóstico fino, revisar el dispositivo real además del reporte.
3. Si el proveedor operativo es RustDesk, verificar también `rustdesk_device_id`, `rustdesk_server_host` y la resolución del secreto cifrado.

## Contratos relacionados

- `documentos/gobernanza_tecnica/contratos/contrato_soporte_remoto_por_empresa_y_mesa_tecnica_central.md`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_reportes_programados_y_exportaciones_contables.md`