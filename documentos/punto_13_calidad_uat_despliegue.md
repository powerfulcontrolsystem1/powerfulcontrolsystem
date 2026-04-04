# Punto 13: Calidad, UAT y Despliegue Controlado

Fecha: 2026-04-04
Estado: en curso

## 1. Objetivo

Establecer un flujo repetible para validar calidad tecnica, ejecutar UAT operativa y habilitar salida controlada a produccion sin romper modulos activos.

## 2. Entregables de este punto

- Script de validacion tecnica integral: `scripts/validar_punto_13.ps1`.
- Reporte tecnico generado por ejecucion: `documentos/punto_13_validacion_integral_resultado.md`.
- Checklist de release actualizado con gate de calidad y UAT: `documentos/release_checklist.md`.

## 3. Flujo operativo

1. Ejecutar validacion tecnica integral:
   - `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\validar_punto_13.ps1`
2. Revisar resultado del reporte tecnico generado.
3. Ejecutar smoke/UAT manual por modulo critico.
4. Autorizar salida controlada solo si se cumplen criterios de aceptacion.

## 4. Criterios de aceptacion minima

- Suite productiva backend en verde:
  - `go test ./auth ./db ./handlers ./metrics ./utils -count=1`
- Suite completa backend en verde:
  - `go test ./... -count=1`
- Sin errores de compilacion en el backend.
- UAT manual validada en modulos criticos:
  - autenticacion y sesiones,
  - clientes,
  - inventario,
  - compras,
  - facturacion,
  - finanzas/eventos contables,
  - auditoria.
- Plan de rollback disponible (backup DB + referencia de commit).

## 5. Matriz UAT manual

| Modulo | Caso UAT minimo | Resultado esperado | Estado |
|---|---|---|---|
| Autenticacion | Login admin y login usuario empresa | Acceso permitido y sesion activa | Pendiente |
| Clientes | Alta + consulta perfil/historial | Persistencia y respuesta consistente | Pendiente |
| Inventario | Movimiento de entrada/salida y consulta kardex | Balance y trazabilidad correctos | Pendiente |
| Compras | Crear documento y transicionar ciclo documental | Flujo `borrador -> emitida -> recepcionada -> contabilizada` | Pendiente |
| Facturacion | Emitir y anular documento | Cumplimiento normativo y eventos contables | Pendiente |
| Finanzas | Registrar movimiento y consultar resumen | Indicadores y eventos alineados | Pendiente |
| Auditoria | Ejecutar accion critica y consultar registro | Evento visible con metadatos | Pendiente |

## 6. Regla de salida controlada

No publicar a entorno productivo si falla cualquiera de los siguientes gates:

- Gate tecnico (tests/compilacion).
- Gate funcional (UAT manual minima).
- Gate de seguridad/logs y rollback.

## 7. Evidencia

Cada iteracion del punto 13 debe dejar evidencia en:

- `documentos/punto_13_validacion_integral_resultado.md`
- `documentos/historial_de_cambios`
- `CHANGELOG.md`
