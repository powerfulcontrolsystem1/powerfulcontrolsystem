# P109-008 - depuracion segura de soportes CxP/IA

Fecha: 2026-08-09
Ambiente: candidato local `codex/p109-batch-no-pr`
Resultado: **PASS local / pendiente de staging, restore y PCS**

## Controles implementados

- `purgar` requiere permiso Delete, `empresa_id`, registro `eliminado`, soporte
  no contabilizado/convertido, retencion de 1 a 3650 dias, motivo y codigo exacto.
- La transaccion bloquea la fila, revalida condiciones y conserva fila/eventos
  como `purgado`; borra la URL y hace imposible recuperar u operar el registro.
- El archivo pasa primero a cuarentena en su directorio privado. Un fallo de base
  restaura el original y el borrado ocurre solo despues del commit.
- Descarga, IA, retencion y depuracion exigen el directorio de la empresa
  efectiva. La radicacion JSON/manual descarta URL, nombre, MIME y hash enviados
  por el cliente.
- La UI separa Activos, Papelera y Depurados; solicita motivo, advertencia y
  escritura exacta del codigo antes de enviar la accion irreversible.
- La saga usa `purga_pendiente` y fases transaccionales idempotentes. Un reintento
  recupera archivo ya movido, cuarentena pendiente o archivo ya eliminado; dos
  cuarentenas candidatas se rechazan sin selección arbitraria.
- El diagnóstico Read compara registros pendientes, cantidad de archivos y
  bytes exclusivamente dentro de `empresa_<id>`, sin revelar nombres o rutas.
- La mutacion toma advisory lock PostgreSQL por empresa durante archivo y base;
  un replay ya finalizado y la ausencia concurrente de una cuarentena son
  idempotentes. El diagnostico identifica pendientes que superan el umbral.
- La ruta conserva `WithEmpresaSoportesComprasIAPermissions`: diagnóstico usa
  Read y purga Delete. Consultas, bloqueos, updates y eventos combinan siempre
  `empresa_id` con `soporte_id`; errores internos se registran y la respuesta
  pública queda genérica.

## Evidencia local

| Prueba | Resultado |
|---|---|
| `go test ./db ./handlers` | PASS |
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| Cuarentena con rollback y commit | PASS |
| Reanudacion en tres fronteras de caída y ambigüedad | PASS |
| Diagnostico de cuarentena aislado empresa 12/53 | PASS |
| Errores de negocio permitidos y errores internos saneados | PASS |
| Contrato de lock entre replicas, replay y borrado concurrente | PASS local/contrato; PostgreSQL dinámico pendiente |
| Umbral acotado y conteo de pendientes vencidos | PASS |
| Rechazo de referencia privada empresa 53 desde empresa 12 | PASS |
| Saneamiento de metadatos manuales falsificados | PASS |
| Fechas antiguas, recientes e invalidas | PASS |
| Contrato PostgreSQL temporal de purga y no recuperacion | CREADO / SKIP sin `PCS_TEST_POSTGRES_DSN` |
| `git diff --check` | PASS |
| Sintaxis JavaScript | PASS |
| Preflight profesional `-Full -Strict` | 22/22 PASS |
| UI local escritorio y ancho movil, sin overflow | PASS visual no autenticado |

Reporte de cierre: `documentos/reportes_profesionales/preflight_20260809_000638.md`.
La prueba visual confirmó `Depurados`, `Depurar archivo`, vista previa y controles
en escritorio/móvil. El servidor estático no ofrecía API autenticada, por lo que
los datos dinámicos no se contabilizan como UAT PCS.

Revalidacion de la saga: `documentos/reportes_profesionales/preflight_20260809_001839.md`
con 22/22 compuertas. La UI local mostró `Depuracion pendiente`, `Depurados`,
`Depurar archivo` y `Diagnostico de cuarentena` en escritorio y ancho movil,
sin overflow horizontal. Sigue siendo evidencia visual no autenticada.

Candidato exacto posterior al saneamiento de errores:
`documentos/reportes_profesionales/preflight_20260809_002303.md`, 22/22 PASS.

Candidato con lock entre replicas, replay, antiguedad y runbook:
`documentos/reportes_profesionales/preflight_20260809_003237.md`, 22/22 PASS.
La revision visual local confirmó controles, mensaje diagnóstico y filtro
pendiente en escritorio/móvil sin overflow; no ejecutó la mutación.

## Limites

- No se eliminaron archivos ni datos reales de Powerful Control System.
- No hubo PR, push, despliegue, mutacion DIAN ni cambio productivo.
- Falta ejecutar PostgreSQL real, carrera Linux, dos empresas/roles, ClamAV,
  backup/restore posterior y validacion visual autenticada en staging.

P109-008 sigue **parcial**. Plan 109 permanece en **53,3 % de implementación**,
**0 % de certificación de este cambio local** y **NO-GO**.
