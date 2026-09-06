# Catalogo de contratos tecnicos

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

Los contratos tecnicos definen el comportamiento esperado de flujos criticos para evitar regresiones silenciosas en rutas, payloads, estados, side effects y validaciones.

## Estructura obligatoria de un contrato

1. alcance del flujo
2. rutas y acciones implicadas
3. entradas obligatorias y opcionales
4. salidas y estados posibles
5. invariantes funcionales y de seguridad
6. side effects en DB, correo, colas o integraciones
7. errores esperados y tratamiento
8. pruebas minimas o evidencia tecnica
9. ADRs y runbooks relacionados

## Catálogo por archivo

- [contrato_autenticacion_administrativa_y_usuarios_empresa](contrato_autenticacion_administrativa_y_usuarios_empresa.md).
- [contrato_centro_soporte](contrato_centro_soporte.md).
- [contrato_checkout_licencias_publico](contrato_checkout_licencias_publico.md).
- [contrato_conciliacion_bancaria_y_cierre_periodo_contable](contrato_conciliacion_bancaria_y_cierre_periodo_contable.md).
- [contrato_documentos_dinamicos_ia_exportacion](contrato_documentos_dinamicos_ia_exportacion.md).
- [contrato_estaciones_sensores_ventas_simple](contrato_estaciones_sensores_ventas_simple.md).
- [contrato_facturacion_electronica_y_documentos_transaccionales](contrato_facturacion_electronica_y_documentos_transaccionales.md).
- [contrato_integraciones_bancarias_y_conectores_externos](contrato_integraciones_bancarias_y_conectores_externos.md).
- [contrato_interoperabilidad_documental_contable_y_fiscal_externa](contrato_interoperabilidad_documental_contable_y_fiscal_externa.md).
- [contrato_matriz_pagos_reales](contrato_matriz_pagos_reales.md).
- [contrato_permisos_contexto_y_wrappers_api_empresa](contrato_permisos_contexto_y_wrappers_api_empresa.md).
- [contrato_reportes_contables_financieros_y_exportacion_multiformato](contrato_reportes_contables_financieros_y_exportacion_multiformato.md).
- [contrato_repositorio_documental_y_firmas_externas](contrato_repositorio_documental_y_firmas_externas.md).
- [contrato_soporte_remoto_por_empresa_y_mesa_tecnica_central](contrato_soporte_remoto_por_empresa_y_mesa_tecnica_central.md).
- [contrato_venta_publica_empresarial_por_empresa](contrato_venta_publica_empresarial_por_empresa.md).

## Contratos prioritarios siguientes

1. conciliacion operativa entre runbooks y contratos nuevos
2. endurecimiento de evidencia para firmas y exportes regulatorios
3. ampliacion del contrato de documentos dinamicos si el flujo pasa de temporal a historial documental persistente.

Estado por fuente y reglas de mantenimiento en el [catálogo](../../catalogo_documental.md)
y [marco documental](../marco_documental.md). Un procedimiento listado no acredita
que se haya ensayado ni que esté autorizado para ejecutarse.
