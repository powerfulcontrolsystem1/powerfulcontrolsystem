# Módulo Contabilidad Colombia

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- El núcleo contempla PUC, terceros, comprobantes, partidas y cierre. Sus libros deben interpretarse desde comprobantes contabilizados, no como certificación de reportes oficiales.
- El worker ya contiene el evento commerce.sale-paid; la integración contable no debe describirse globalmente como inexistente. Su cobertura y conciliación se verifican por operación.
- Factura, nómina y documento soporte tienen adaptadores distintos; no dependen de habilitar un proveedor genérico futuro.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Objetivo

El módulo `contabilidad_colombia` agrega la capa contable/legal colombiana por empresa. Complementa `finanzas`: finanzas registra operación diaria, mientras contabilidad organiza PUC, terceros, impuestos, comprobantes de doble partida, asientos, libros base y cierres.

## Alcance funcional

- Configuración contable por empresa: moneda, periodo actual, versión PUC, base NIIF y bloqueo de periodos cerrados.
- PUC colombiano base con cuentas de caja, bancos, clientes, proveedores, IVA, retenciones, ingresos, costos y gastos.
- Terceros contables con documento, régimen, responsabilidades y contacto.
- Impuestos y retenciones Colombia: IVA, retefuente, reteICA y extensibles.
- Comprobantes contables: nota contable, ingreso, egreso, causación y ajuste.
- Validación de doble partida: débito y crédito deben cuadrar antes de contabilizar.
- Consulta de comprobantes y detalle de asientos.
- Cierre y reapertura controlada de periodos.
- Dashboard con cuentas, terceros, comprobantes del mes, totales y diferencia.

## Seguridad y aislamiento

- Todas las tablas incluyen `empresa_id`.
- Endpoint protegido: `/api/empresa/contabilidad_colombia`.
- Wrapper de permisos: `WithEmpresaContabilidadColombiaPermissions`.
- Módulo de licencia: `contabilidad_colombia`.
- Página de menú: `linkContabilidadColombia`.

## Archivos principales

- Base de datos: `backend/db/contabilidad_colombia.go`
- Handler: `backend/handlers/contabilidad_colombia.go`
- Ruta: `backend/main.go`
- Permisos: `backend/handlers/empresa_permisos.go`
- Interfaz: `web/administrar_empresa/contabilidad_colombia.html`
- Licencias: `web/super/licencias.html`

## Flujo operativo

1. Cargar el PUC base.
2. Crear o completar terceros.
3. Configurar impuestos y retenciones.
4. Registrar comprobantes con mínimo dos líneas.
5. Verificar que débito y crédito estén balanceados.
6. Cerrar el periodo cuando no existan diferencias.

## Cobertura y aceptación por verificar

- Cobertura de integración automática por origen (ventas, compras, nómina, inventario y bancos); existen mecanismos de worker, pero falta aceptación de todos los flujos.
- Conciliar exportes de diario, mayor, balance y exógena con los comprobantes y el módulo avanzado; su existencia no demuestra cumplimiento de cada formato vigente.
- Verificar cada adaptador fiscal dedicado; RADIAN y las familias no implementadas permanecen fuera del alcance aceptado.

## Fuentes y aceptación de la revisión

[contabilidad_colombia.go](../backend/handlers/contabilidad_colombia.go), [contabilidad_colombia.go](../backend/db/contabilidad_colombia.go), [contabilidad_colombia_test.go](../backend/db/contabilidad_colombia_test.go), [contabilidad_colombia.html](../web/administrar_empresa/contabilidad_colombia.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go), [business_registry.go](../backend/cmd/pcs-worker/business_registry.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
