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

## Evidencia local

| Prueba | Resultado |
|---|---|
| `go test ./db ./handlers` | PASS |
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| Cuarentena con rollback y commit | PASS |
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

## Limites

- No se eliminaron archivos ni datos reales de Powerful Control System.
- No hubo PR, push, despliegue, mutacion DIAN ni cambio productivo.
- Falta ejecutar PostgreSQL real, carrera Linux, dos empresas/roles, ClamAV,
  backup/restore posterior y validacion visual autenticada en staging.

P109-008 sigue **parcial**. Plan 109 permanece en **53,3 % de implementación**,
**0 % de certificación de este cambio local** y **NO-GO**.
