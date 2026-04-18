# Checklist operativa: evidencia documental para QA y soporte

Fecha: 2026-04-18
Estado: vigente

## Objetivo

Checklist corta para validar incidentes o UAT sobre documentos empresariales, firmas externas y exportes regulatorios sin perder trazabilidad entre estado transaccional, repositorio documental y evidencia presentada.

## Uso recomendado

- QA antes de cerrar un caso documental como aprobado.
- Soporte antes de escalar un incidente a backend.
- Operacion antes de aceptar un PDF, CSV, XLS, TXT o JSON como respaldo suficiente.

## Checklist

1. Confirmar `empresa_id`, `documento_codigo`, rol involucrado y endpoint exacto del incidente.
2. Confirmar si el flujo afectado pertenece a compras, facturacion, repositorio documental, firmas o reportes/exportes.
3. Consultar la version vigente del documento en `/api/empresa/documentos/gestion?action=versiones` por `id` o `documento_codigo`.
4. Verificar que exista una sola version `vigente` y que las anteriores queden como `historico`.
5. Verificar `action=acceso` y, si aplica, `action=repositorio&include_denegados=1` con el mismo rol del incidente.
6. Si existe firma, verificar `documento_gestion_id`, `hash_firma`, `algoritmo_firma`, firmante y `fecha_firma`.
7. Si existe exporte regulatorio, identificar dataset, formato, timestamp y si el exporte es informativo o evidencia reconciliable.
8. No aceptar un exporte aislado como evidencia formal si no se puede enlazar con `documento_codigo`, version vigente y firma asociada cuando aplique.
9. Si el caso toca compras o facturacion, cruzar el documento con su estado transaccional e integracion fiscal/contable real.
10. Registrar en el cierre del caso si la evidencia validada fue documental versionada, firma asociada, exporte informativo o una combinacion de ellas.

## Criterio rapido de decision

- Verde: documento vigente conciliable, acceso correcto, firma coherente si aplica y exporte no contradictorio.
- Amarillo: documento existente pero con historial incompleto, firma huérfana o exporte sin enlace directo a la evidencia documental.
- Rojo: no existe versión vigente clara, el acceso esperado falla sin explicación, o se intenta usar un exporte como único respaldo regulatorio.

## Escalamiento minimo

- Escalar a backend si falla el mapeo de permisos, el versionado o la reconciliacion entre documento y transaccion.
- Escalar a operacion si el problema es de evidencia presentada, exporte regulatorio o uso incorrecto de formatos.
- Escalar a QA si el caso no trae request reproducible, rol efectivo o referencia documental suficiente.

## Artefactos relacionados

- `documentos/gobernanza_tecnica/contratos/contrato_repositorio_documental_y_firmas_externas.md`
- `documentos/gobernanza_tecnica/contratos/contrato_interoperabilidad_documental_contable_y_fiscal_externa.md`
- `documentos/gobernanza_tecnica/contratos/contrato_reportes_contables_financieros_y_exportacion_multiformato.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_reconciliacion_documental_fiscal_y_contable_externa.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_versionado_documental_y_firmas_externas.md`