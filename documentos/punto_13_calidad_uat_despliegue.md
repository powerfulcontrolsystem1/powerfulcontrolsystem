# Punto 13: Calidad, UAT y Despliegue Controlado

Fecha: 2026-04-07
Estado: completado (baseline transversal)

## 1. Objetivo

Establecer un flujo repetible para validar calidad tecnica, ejecutar UAT operativa y habilitar salida controlada a produccion sin romper modulos activos.

## 2. Entregables de este punto

- Script de validacion tecnica integral: `scripts/validar_punto_13.ps1`.
- Reporte tecnico generado por ejecucion: `documentos/punto_13_validacion_integral_resultado.md`.
- Checklist de release actualizado con gate de calidad y UAT: `documentos/release_checklist.md`.
- Acta UAT formal por rol y cierre transversal: este mismo documento (seccion 6).

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

## 5. Objetivo minimo de cobertura por modulo/capa (corte 2026-04-07)

Regla transversal vigente:

- Por modulo funcional, mantener al menos 1 prueba de flujo en `db` y 1 prueba de flujo en `handlers` cuando aplique, ademas de compilacion global backend en verde.
- Gate de cobertura por capa:

| Capa | Meta minima | Evidencia de corrida | Estado |
|---|---|---|---|
| db | >= 50% | `go test ./auth ./db ./handlers ./metrics ./utils -cover -count=1` -> `51.4%` | aprobado |
| handlers | >= 50% | `go test ./auth ./db ./handlers ./metrics ./utils -cover -count=1` -> `50.4%` | aprobado |
| auth/metrics/utils | Suite productiva en verde + mejora continua de cobertura dedicada | mismo comando: cobertura `auth 85.3%`, `metrics 78.0%`, `utils 71.1%` | aprobado |

## 6. Acta UAT formal por rol (2026-04-07)

Ambiente: local (desarrollo), backend Go + SQLite.

| Rol | Alcance UAT validado | Evidencia | Resultado |
|---|---|---|---|
| super_admin | Gate de acceso super y bloqueo a roles no super en endpoints criticos | `TestSuperEndpointsPermisosPorRol` | aprobado |
| admin_empresa (contabilidad) | Contexto efectivo de permisos, matriz por rol y control de acceso documental | `TestEmpresaPermisosContextoHandlerRetornaPermisosPorRol`, `TestEmpresaPermisosContextoHandlerIncluyeMatrizRoles`, `TestEmpresaDocumentosGestionHandlerVersionadoYControlAcceso` | aprobado |
| usuario_empresa (cajero) | Restricciones operativas por rol en metodos de pago, propina y comision | `TestEmpresaCarritosCompraBloqueaMetodoPagoSegunRol`, `TestEmpresaCarritosCompraRespetaBloqueoPropinaYComisionPorRol`, `TestEmpresaConfiguracionOperativaHandlerConfigAndRole` | aprobado |

Ejecucion consolidada del acta UAT por rol:

- `go test ./handlers -run "Test(SuperEndpointsPermisosPorRol|EmpresaPermisosContextoHandlerRetornaPermisosPorRol|EmpresaPermisosContextoHandlerIncluyeMatrizRoles|EmpresaCarritosCompraBloqueaMetodoPagoSegunRol|EmpresaCarritosCompraRespetaBloqueoPropinaYComisionPorRol|EmpresaConfiguracionOperativaHandlerConfigAndRole|EmpresaDocumentosGestionHandlerVersionadoYControlAcceso)$" -count=1` -> OK.

## 7. Matriz UAT por modulo (resumen)

| Modulo | Caso UAT minimo | Resultado esperado | Estado |
|---|---|---|---|
| Autenticacion | Login admin y login usuario empresa | Acceso permitido y sesion activa | Aprobado |
| Clientes | Alta + consulta perfil/historial | Persistencia y respuesta consistente | Aprobado |
| Inventario | Movimiento de entrada/salida y consulta kardex | Balance y trazabilidad correctos | Aprobado |
| Compras | Crear documento y transicionar ciclo documental | Flujo `borrador -> emitida -> recepcionada -> contabilizada` | Aprobado |
| Facturacion | Emitir y anular documento | Cumplimiento normativo y eventos contables | Aprobado |
| Finanzas | Registrar movimiento y consultar resumen | Indicadores y eventos alineados | Aprobado |
| Auditoria | Ejecutar accion critica y consultar registro | Evento visible con metadatos | Aprobado |

## 8. Regla de salida controlada

No publicar a entorno productivo si falla cualquiera de los siguientes gates:

- Gate tecnico (tests/compilacion).
- Gate funcional (UAT formal por rol + matriz UAT por modulo).
- Gate de seguridad/logs y rollback.

## 9. Evidencia

Cada iteracion del punto 13 debe dejar evidencia en:

- `documentos/punto_13_validacion_integral_resultado.md`
- `documentos/release_checklist.md`
- este documento (secciones 5, 6 y 7)
- `documentos/historial_de_cambios`
- `CHANGELOG.md`
