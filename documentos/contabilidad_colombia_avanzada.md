# Suite Contable Colombia Avanzada

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- La escritura manual nomina_electronica devuelve 409 nomina_manual_no_permitida. Los datos fiscales se derivan del flujo mensual dedicado, no de campos editables CUNE/estado.
- Las nuevas CxP y abonos a CxP histórica están bloqueados en esta superficie; usar la cuenta canónica de Finanzas y conciliar el legado.
- El documento soporte cuenta con preflight y adaptador dedicado; no se acredita aceptación externa con la presencia del código.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha de actualización: 2026-08-26

El módulo `contabilidad_colombia_avanzada` complementa `contabilidad_colombia` con los submódulos que suelen requerir los sistemas contables colombianos profesionales.

## Submódulos incluidos

- Información exógena DIAN / medios magnéticos: formatos configurables por año gravable, registros por tercero, validaciones básicas y generación desde comprobantes contables.
- Nómina electrónica DIAN: documentos por empleado y periodo con devengados, deducciones, total, estado DIAN, CUNE y payload preparado.
- Documento soporte electrónico: borradores estructurados de compras a
  vendedores no obligados a facturar, preflight no emisor y adaptador dedicado
  del Anexo técnico DIAN 1.1. El CUDS, la firma, el acuse y el estado fiscal no
  pueden escribirse desde el formulario.
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
- Adaptador DIAN: `backend/handlers/dian_documento_soporte.go`.
- Interfaz especializada: `web/js/documento_soporte_dian.js`.

## Acciones API

GET:

- `dashboard`
- `exogena_formatos`
- `exogena_registros`
- `nomina_electronica`
- `documentos_soporte`
- `documento_soporte_preflight` (solo lectura: no reserva número, no genera XML
  y no transmite)
- `activos_fijos`
- `cartera_cxp`
- `libros`
- `libros_resumen`

POST/PUT:

- `seed`
- `exogena_formatos`
- `exogena_registros`
- `generar_exogena`
- `nomina_electronica`: operación manual bloqueada (409); usar el flujo mensual dedicado.
- `documentos_soporte`
- `activos_fijos`
- `cartera_cxp`: nuevas CxP bloqueadas; el legado se consulta/concilia según contrato.

## Separación por empresa

Todas las tablas incluyen `empresa_id` y todas las consultas filtran por empresa. El módulo no duplica la lógica de PUC, terceros ni comprobantes; usa el núcleo `contabilidad_colombia` para generar libros y registros de exógena desde asientos contabilizados.

## Documento soporte DIAN 1.1

- `empresa_contabilidad_documentos_soporte` conserva vendedor, ubicación,
  pago, líneas, importes `NUMERIC(18,2)`, número legal, instantánea de
  numeración y espejo del resultado DIAN por empresa.
- El servidor valida y recalcula todas las líneas. Acepta exactamente los 1.093
  códigos de unidad UN/ECE Revision 4 incorporados en la caja de herramientas
  DIAN 1.1; el navegador solo sugiere las unidades más comunes.
- La configuración de numeración vive separada en
  `empresa_dian_documentos_configuracion` con
  `tipo_documento=documento_soporte`. El prefijo es opcional y, si existe,
  admite hasta cuatro caracteres alfanuméricos; autorización, vigencia, rango y
  consecutivo se validan antes de activar o reservar.
- La reserva usa bloqueo de fila y es idempotente por borrador. Nunca reduce un
  consecutivo consumido ni mezcla rangos de factura de venta.
- La emisión real solo se ofrece tras preflight verde y confirmación escrita
  exacta. Genera `InvoiceTypeCode=05`, CUDS SHA-384, `CustomizationID` 10/11,
  firma XAdES y `SendBillSync`; conserva XML, respuesta y representación en el
  almacenamiento fiscal privado existente.
- Si falla el espejo contable antes del despacho, la operación se detiene sin
  transmitir. Una advertencia de persistencia posterior al despacho se devuelve
  separada en `advertencias_persistencia` y nunca se convierte en un error HTTP
  que invite a reenviar a ciegas el mismo XML.
- Un borrador, configuración incompleta o vendedor/adquirente fiscalmente
  inválido permanece bloqueado sin consumir consecutivo ni transmitir XML.

## Notas de cumplimiento y operación

Los formatos de información exógena siguen siendo configurables por año
gravable. Documento soporte usa el transporte y la firma empresarial del módulo
de facturación electrónica, pero su numeración y su fuente fiscal son propias.
Que el adaptador esté implementado no sustituye la configuración DIAN de cada
empresa ni constituye evidencia de aceptación: producción solo se confirma con
despliegue verificado y acuse real del documento emitido.

## Fuentes y aceptación de la revisión

[contabilidad_colombia_avanzada.go](../backend/handlers/contabilidad_colombia_avanzada.go), [contabilidad_colombia_avanzada.go](../backend/db/contabilidad_colombia_avanzada.go), [contabilidad_colombia_avanzada_test.go](../backend/db/contabilidad_colombia_avanzada_test.go), [contabilidad_colombia_avanzada.html](../web/administrar_empresa/contabilidad_colombia_avanzada.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go), [dian_documento_soporte.go](../backend/handlers/dian_documento_soporte.go), [dian_nomina_electronica.go](../backend/handlers/dian_nomina_electronica.go).

Requisitos aplicables: PCS-REQ-001 a PCS-REQ-009, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
