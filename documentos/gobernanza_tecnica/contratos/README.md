# Catalogo de contratos tecnicos

Fecha: 2026-04-30
Estado: vigente

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

## Catalogo inicial

- `contrato_checkout_licencias_publico.md`: creado y actualizado con fallback Epayco clasico firmado por POST.
- `contrato_estaciones_sensores_ventas_simple.md`: creado.
- `contrato_autenticacion_administrativa_y_usuarios_empresa.md`: creado.
- `contrato_venta_publica_empresarial_por_empresa.md`: creado.
- `contrato_permisos_contexto_y_wrappers_api_empresa.md`: creado.
- `contrato_facturacion_electronica_y_documentos_transaccionales.md`: creado.
- `contrato_reportes_contables_financieros_y_exportacion_multiformato.md`: creado.
- `contrato_soporte_remoto_por_empresa_y_mesa_tecnica_central.md`: creado.
- `contrato_conciliacion_bancaria_y_cierre_periodo_contable.md`: creado.
- `contrato_interoperabilidad_documental_contable_y_fiscal_externa.md`: creado.
- `contrato_integraciones_bancarias_y_conectores_externos.md`: creado.
- `contrato_repositorio_documental_y_firmas_externas.md`: creado.
- `contrato_documentos_dinamicos_ia_exportacion.md`: creado para `/generate` y `/download`.

## Contratos prioritarios siguientes

1. conciliacion operativa entre runbooks y contratos nuevos
2. endurecimiento de evidencia para firmas y exportes regulatorios
3. ampliacion del contrato de documentos dinamicos si el flujo pasa de temporal a historial documental persistente.
