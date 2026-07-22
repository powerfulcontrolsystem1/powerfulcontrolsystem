# Inventario de goroutines - Plan 105

Fecha de corte: 2026-07-21. Fuente: `rg -n 'go\\s+func' backend -g '*.go'`.

Este inventario clasifica los 23 lanzamientos encontrados. Es una revision
estatica: no demuestra cancelacion, idempotencia ni recuperacion en runtime. No
autoriza cambiar efectos externos o transaccionales sin prueba en PostgreSQL y
staging.

## Criterio operativo

- **Supervisada:** vida controlada por `context`, `WaitGroup`, canal de parada o
  proceso worker; conservar y cubrir con prueba de apagado.
- **Sincronica acotada:** paralelismo local que espera `WaitGroup`; conservar
  solo con limites y propagacion de cancelacion/errores.
- **Best effort:** no altera el resultado HTTP; debe documentar la perdida
  tolerada o pasar a outbox si su evidencia es obligatoria.
- **Durable requerida:** un reinicio no puede perder el efecto; usar outbox en
  la misma transaccion del hecho de negocio, worker idempotente y DLQ.

## Matriz de clasificacion

| Ubicacion | Cant. | Tipo actual | Riesgo | Accion P105 | Prioridad |
| --- | ---: | --- | --- | --- | --- |
| `cmd/pcs-worker/main.go:121` | 1 | Supervisada; `ctx` y canal de error | Bajo | Probar apagado y propagacion de error del runner. | P1 |
| `internal/platform/worker/health.go:103,109` | 2 | Servidor de salud y apagado por `ctx.Done` | Bajo | Probar cierre de listener y no dejar goroutine al cancelar. | P1 |
| `internal/platform/worker/worker.go:118` | 1 | Renovacion de lease con `WaitGroup` y `stopLease` | Medio | Probar cancelacion durante handler lento y liberacion de lease. | P0 |
| `handlers/auditoria_modulos_especificos.go:34`, `auditoria_empresa.go:645`, `auditoria_super.go:191` | 3 | Auditoria best effort a BD | Medio | Definir retencion de perdida. Para auditoria exigible, publicar outbox en la misma transaccion. | P1 |
| `handlers/control_electrico.go:1216` | 1 | Despacho de hardware no durable | Alto | Reemplazar por evento outbox idempotente antes de produccion del modulo; no ejecutar dos veces al reintentar. | P0 |
| `handlers/dynamic_documents.go:523` | 1 | Espera respuesta IA en goroutine | Medio | Propagar `ctx` al cliente/proveedor, imponer timeout y comprobar que no continúa tras cancelacion. | P1 |
| `handlers/empresa_permisos.go:2638,2642,2701,2705` | 4 | Fan-out de dos lecturas, espera `WaitGroup` | Medio | Añadir context/timeout a consultas y prueba de error/cancelacion; no persistir efectos. | P1 |
| `handlers/empresa_permisos.go:3989,3996,4003,4010,4017` | 5 | Snapshot de autorizacion, espera `WaitGroup` | Alto | Limitar concurrencia, propagar cancelacion y probar que un fallo no concede acceso ni filtra empresa. | P0 |
| `handlers/server_runtime_notifications.go:180` | 1 | Correo de arranque best effort | Bajo | Mantener solo si la perdida es aceptada; si es alerta operativa obligatoria, outbox/worker. | P2 |
| `handlers/super_alertas.go:408` | 1 | Alerta de negocio best effort | Medio | Persistir evento y entregar mediante worker con idempotencia, reintento y DLQ. | P1 |
| `handlers/super_mantenimiento_agentes.go:195` | 1 | Worker con ticker y canal `stop` | Medio | Convertir a `context.Context`, esperar fin en apagado y probar una unica instancia. | P1 |
| `handlers/usuarios_empresa.go:2869,2882` | 2 | Calentamiento de cache, incluye `Sleep` | Bajo | Hacer cancelable y sin retraso huérfano; mantenerlo prescindible para autorizacion. | P2 |

Total: **23 goroutines**. Nueve estan en `empresa_permisos.go`: todas esperan
un `WaitGroup`, pero las cinco del snapshot son P0 porque determinan el acceso
efectivo por empresa.

Revision 2026-07-21: las cinco operaciones actuales son acceso IA, politica de
licencia, matriz de modulos, overrides empresariales y acceso compartido. No
reciben un `context.Context` cancelable en este punto; una falla o demora de
una dependencia no debe terminar concediendo acceso. El despacho de control
electrico tambien se ejecuta en goroutine best-effort sin outbox durable.

## Orden de ejecucion para Terra

1. Crear pruebas de apagado para worker, lease y health sin usar `Sleep` fijo.
2. Corregir cancelacion y timeout de los nueve flujos de permisos; probar fallo
   cerrado y aislamiento A/B.
3. Diseñar el evento durable de control electrico: payload minimo, clave de
   idempotencia, lease, reintento, DLQ y auditoria por `empresa_id`.
4. Aplicar el mismo patron a alertas y a cualquier auditoria cuya perdida no sea
   admisible. No meter correo de arranque en la transaccion de ventas.
5. Ejecutar pruebas de reinicio, proveedor lento, doble entrega y worker caido
   contra PostgreSQL real y staging antes de cerrar P105-005.

## Criterio de cierre P105-005

No cerrar con esta matriz sola. Se requiere evidencia de que los efectos P0 son
durables e idempotentes, las goroutines supervisadas finalizan al apagar, las
lecturas de autorizacion fallan cerradas y las metricas exponen cola, intentos,
DLQ y duracion.
