# Suite Contable Colombia Avanzada

Fecha: 2026-05-05

El módulo `contabilidad_colombia_avanzada` complementa `contabilidad_colombia` con los submódulos que suelen requerir los sistemas contables colombianos profesionales.

## Submódulos incluidos

- Información exógena DIAN / medios magnéticos: formatos configurables por año gravable, registros por tercero, validaciones básicas y generación desde comprobantes contables.
- Nómina electrónica DIAN: documentos por empleado y periodo con devengados, deducciones, total, estado DIAN, CUNE y payload preparado.
- Documento soporte electrónico: compras a proveedores no obligados a facturar con subtotal, IVA, retenciones, total, estado DIAN y CUDS.
- Activos fijos: control por código, categoría, fecha de compra, costo, valor residual, vida útil, depreciación mensual, depreciación acumulada y valor en libros.
- Cartera y cuentas por pagar: cuentas por cobrar y por pagar con vencimiento, saldo, estado, tercero y origen.
- Libros oficiales: libro diario, mayor/auxiliar, balance de prueba y resúmenes base desde comprobantes contabilizados.

## Integración técnica

- Endpoint protegido: `/api/empresa/contabilidad_colombia_avanzada`.
- Wrapper de permisos: `WithEmpresaContabilidadColombiaAvanzadaPermissions`.
- Módulo de licencia: `contabilidad_colombia_avanzada`.
- Página administrativa: `web/administrar_empresa/contabilidad_colombia_avanzada.html`.
- Base de datos: `backend/db/contabilidad_colombia_avanzada.go`.
- Handler: `backend/handlers/contabilidad_colombia_avanzada.go`.

## Acciones API

GET:

- `dashboard`
- `exogena_formatos`
- `exogena_registros`
- `nomina_electronica`
- `documentos_soporte`
- `activos_fijos`
- `cartera_cxp`
- `libros`
- `libros_resumen`

POST/PUT:

- `seed`
- `exogena_formatos`
- `exogena_registros`
- `generar_exogena`
- `nomina_electronica`
- `documentos_soporte`
- `activos_fijos`
- `cartera_cxp`

## Separación por empresa

Todas las tablas incluyen `empresa_id` y todas las consultas filtran por empresa. El módulo no duplica la lógica de PUC, terceros ni comprobantes; usa el núcleo `contabilidad_colombia` para generar libros y registros de exógena desde asientos contabilizados.

## Notas de cumplimiento

Los formatos DIAN se manejan como configurables por año gravable para permitir ajustes normativos sin recompilar el sistema. La transmisión real a servicios DIAN debe conectarse con el módulo existente de facturación electrónica y firma cuando se habilite el proveedor tecnológico o el ambiente de pruebas correspondiente.
