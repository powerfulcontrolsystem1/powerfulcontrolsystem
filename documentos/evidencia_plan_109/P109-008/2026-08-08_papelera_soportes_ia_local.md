# P109-008 - papelera recuperable de soportes CxP/IA

Fecha: 2026-08-08
Ambiente: candidato local `codex/p109-batch-no-pr`
Alcance: P109-002 y P109-008
Resultado: **PASS local / pendiente de staging y PCS**

## Cambio verificado

- La bandeja consulta exclusivamente registros `activo` o `eliminado` y todas
  las consultas conservan `empresa_id`.
- El wrapper efectivo exige Delete para enviar a papelera y Update para
  recuperar; las acciones no heredan Create por usar POST.
- Enviar a papelera y recuperar usan una transaccion con bloqueo de la fila,
  motivo obligatorio, actor y evento auditable. No se borra el archivo ni el
  historial del soporte.
- Los soportes contabilizados o con una CxP convertida no pueden eliminarse.
- Recuperar falla si otro soporte activo de la empresa tiene el mismo hash o
  numero de documento.
- Un registro eliminado queda cerrado para descarga, extraccion IA, revision,
  aprobacion, rechazo y contabilizacion hasta su recuperacion.
- La interfaz separa Activos/Papelera, muestra `Eliminado`, solicita motivo y
  habilita solo la accion valida. Los enlaces de archivo aceptan el mismo origen.

## Pruebas ejecutadas

| Prueba | Resultado |
|---|---|
| `go test ./db ./handlers -count=1` | PASS |
| Transiciones activo/eliminado, idempotencia y bloqueo contable | PASS |
| Filtro de papelera cerrado a activo ante valores no permitidos | PASS |
| Contrato UI de filtro, botones, motivo y bloqueo de edicion | PASS |
| Contrato backend de `empresa_id`, auditoria y descarga activa | PASS |
| Contrato de permisos Delete/Update para eliminar/restaurar | PASS |
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| Sintaxis JavaScript del modulo | PASS |
| `profesional_preflight.ps1 -Full -Strict` (22 compuertas) | PASS |
| Revision visual local escritorio/movil, filtros y seis botones | PASS |
| `go test -race ./db ./handlers` | NO EJECUTABLE: runtime Go sin CGO |
| Integracion `TestSoporteComprasIAPapeleraPostgres` | CREADA / SKIP: DSN no configurado |

La revisión visual local usó una página servida sin backend autenticado; confirmó
composición, responsive, estados deshabilitados y ausencia de errores de consola,
pero no se contabiliza como UAT dinámica de datos.

El cierre automatizado definitivo quedó en
`documentos/reportes_profesionales/preflight_20260808_193344.md`. La ausencia de
CGO mantiene pendiente la carrera real en Linux/staging y no recibe crédito.
Se agregó una prueba PostgreSQL con tablas temporales para ejecutar al disponer
de `PCS_TEST_POSTGRES_DSN`; el equipo local no tiene DSN ni Docker y por eso el
resultado omitido no se presenta como PASS.

## Seguridad y datos

- No se usaron credenciales ni se imprimieron secretos.
- No se modificaron datos de Powerful Control System.
- No hubo borrado fisico, PR, push, despliegue ni emision DIAN.
- La prueba local no sustituye aislamiento dinámico entre dos empresas ni una
  validacion de permisos por roles efectivos.

## Pendientes para cerrar las fases

1. Publicar el candidato aislado en staging por el flujo oficial.
2. Ejecutar eliminación, descarga denegada, auditoría y recuperación visual con
   un soporte de prueba no contabilizado en PCS.
3. Repetir cruce A/B, roles permitidos/denegados y carrera de acciones.
4. Validar réplica A/B, retención, antivirus y restore del soporte eliminado.
5. Probar todos los botones IA y degradación real del proveedor sin crear CxP
   antes de revisión y aprobación humanas.

## Veredicto

El ciclo recuperable queda implementado y cubierto localmente, pero P109-002 y
P109-008 siguen **parciales**. El Plan 109 permanece en **53,3 % de
implementación**, **0 % de certificación de este arreglo local** y **NO-GO**.
