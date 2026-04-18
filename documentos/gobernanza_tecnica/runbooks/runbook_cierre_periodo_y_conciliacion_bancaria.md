# Runbook: cierre de periodo y conciliacion bancaria

Fecha: 2026-04-18
Estado: vigente

## Sintomas cubiertos

- no se puede cerrar o reabrir un periodo contable.
- el backend rechaza movimientos financieros porque el periodo esta cerrado.
- una importacion de extractos se procesa, pero la conciliacion automatica no encuentra match.
- el resumen por periodo devuelve pendientes o descuadres que no se esperaban.
- una reimportacion de extractos parece duplicar datos o no actualiza el estado esperado.

## Alcance

Aplica al modulo financiero empresarial para periodos contables, movimientos internos, extractos bancarios y conciliacion automatica por periodo.

## Fuentes de evidencia

- `backend/handlers/finanzas.go`
- `backend/db/finanzas.go`
- `backend/db/finanzas_conciliacion_bancaria.go`
- tablas `empresa_finanzas_periodos`, `empresa_finanzas_movimientos`, `empresa_finanzas_bancos_movimientos`
- eventos contables del modulo `finanzas`

## Verificaciones iniciales

1. Confirmar `empresa_id` y `periodo` exacto afectado.
2. Revisar si el periodo esta `abierto` o `cerrado` en `empresa_finanzas_periodos`.
3. Si el problema es de cierre o reapertura, validar que se este enviando `autorizado_por`, `motivo_autorizacion` y `evidencia_autorizacion`.
4. Si el problema es de conciliacion, verificar que los extractos existan y que tengan `estado_conciliacion` correcto.
5. Confirmar `tolerancia_dias`, `tolerancia_monto`, referencias bancarias y rango de fechas usados en la conciliacion.

## Causas probables

- falta de evidencia de autorizacion al cerrar o reabrir.
- periodo ya cerrado al intentar crear, editar o eliminar movimientos.
- extractos importados con referencia, fecha o monto fuera de tolerancia para el matching.
- extracto ya conciliado previamente y reimportado de forma idempotente.
- resumen agregado mostrando `con_pendientes` o `con_descuadre` por extractos sin match.

## Acciones de recuperacion

1. Si el cierre falla, repetir la operacion incluyendo la evidencia completa de autorizacion.
2. Si un movimiento choca contra un periodo cerrado, decidir primero si corresponde reabrir formalmente el periodo con evidencia o registrar el ajuste en un periodo distinto.
3. Si la importacion falla, revisar cada item de `movimientos[]` y confirmar `empresa_id`, `tipo_movimiento`, monto, moneda y fecha.
4. Si la conciliacion automatica no encuentra match, inspeccionar la referencia bancaria y la diferencia real de fecha/monto frente al movimiento interno.
5. Si el extracto queda en `con_desviacion`, revisar primero tolerancias y luego la existencia real del movimiento interno antes de volver a importar o conciliar.
6. Si una reimportacion parece duplicar, validar el `hash_movimiento`; el flujo correcto debe actualizar por hash, no crear una segunda fila equivalente.

## Validacion posterior

- el periodo cambia de estado y queda trazado con evidencia en observaciones y eventos.
- los movimientos del periodo respetan el bloqueo cuando corresponde.
- la conciliacion automatica devuelve cifras coherentes de revisados, conciliados y desviados.
- el resumen por periodo refleja el estado esperado sin falsos pendientes.
- una reimportacion del mismo extracto actualiza o conserva la fila correcta en vez de duplicarla.

## Notas operativas

1. La conciliacion bancaria automatica no garantiza conciliacion total; su resultado correcto puede ser `con_desviacion` o `con_pendientes` si no existe match valido.
2. El bloqueo por periodo cerrado es una proteccion deliberada; no debe bypassarse desde frontend.
3. El tablero financiero exportado no cubre todos los detalles del resumen bancario; para diagnostico fino usar el resumen de conciliacion por periodo o el listado de extractos.

## Contratos relacionados

- `documentos/gobernanza_tecnica/contratos/contrato_conciliacion_bancaria_y_cierre_periodo_contable.md`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_reportes_programados_y_exportaciones_contables.md`