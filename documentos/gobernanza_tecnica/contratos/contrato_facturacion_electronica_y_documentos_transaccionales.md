# Contrato tecnico: facturacion electronica y documentos transaccionales

Fecha: 2026-04-18
Estado: vigente

Actualizacion 2026-08-21:

- El adaptador DIAN implementado cubre exclusivamente la familia del Anexo
  Tecnico de Factura Electronica de Venta: `factura_electronica`,
  `nota_credito` y `nota_debito`.
- `documento_soporte`, nomina electronica, documentos equivalentes y RADIAN
  permanecen catalogados, pero su emision esta bloqueada con HTTP 422 hasta
  contar con esquema, datos, numeracion, identificador y transporte propios.
- Ningun tipo desconocido o de otra familia puede convertirse por defecto en
  `Invoice`. El bloqueo ocurre antes de generar XML, reservar consecutivo,
  persistir una emision o transmitir informacion fiscal.
- La consulta visual distingue el comprobante comercial `CP-*` de la factura
  fiscal `FV-*`; una venta con factura asociada no se cuenta dos veces como
  factura electronica.

Actualizacion 2026-06-18:

- PCS quedo validado en DIAN produccion para `empresa_id=12` con prefijo `1PCS`, resolucion `18764111318575`, rango `1-100000` y Software ID asociado.
- Evidencia fiscal aceptada: portal DIAN produccion muestra `1PCS2` y `1PCS3` como `Aprobado con notificacion`; `1PCS3` tambien obtuvo acuse SOAP/WCF `estado_dian=aceptado` y `acuse_estado=aceptado`.
- `Regla 90, Documento procesado anteriormente` no es aceptacion automatica. Solo se considera cierre aceptado si existe acuse oficial aceptado, documento visible en portal DIAN o evidencia equivalente.
- Despues de una prueba directa fuera del flujo documental, los contadores deben quedar adelantados para no reutilizar el folio ya aprobado.

## Alcance

Este contrato cubre el ciclo documental empresarial de facturacion y documentos de venta, la persistencia comun en `empresa_facturacion_documentos`, la configuracion por pais de facturacion electronica, el envio automatico de resumen por correo al cliente, la cola de reintentos y reconciliacion fiscal, y la base operativa actual del endpoint DIAN Colombia.

## Endpoints cubiertos

### Facturacion electronica general

- `GET /api/empresa/facturacion_electronica?empresa_id={id}`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&pais_codigo=CO`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&action=documentos`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&action=reintentos`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&action=reconciliacion`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&action=artefactos&tipo_documento={tipo}&documento_codigo={codigo}`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&action=descargar_artefacto&id={id}`
- `POST|PUT /api/empresa/facturacion_electronica`
- `POST|PUT /api/empresa/facturacion_electronica?action=emitir`
- `POST|PUT /api/empresa/facturacion_electronica?action=anular`
- `POST|PUT /api/empresa/facturacion_electronica?action=nota_credito`
- `POST|PUT /api/empresa/facturacion_electronica?action=emitir_nota_credito`
- `POST|PUT /api/empresa/facturacion_electronica?action=reenviar_correo`
- `POST /api/empresa/facturacion_electronica?action=procesar_reintentos`
- `POST /api/empresa/facturacion_electronica?action=reconciliar_estados`

### Deteccion y catalogo de paises FE

- `GET /api/empresa/facturacion_electronica/pais_detectado?empresa_id={id}`
- `GET /api/empresa/facturacion_electronica/paises_disponibles`

### Base operativa DIAN Colombia

- `GET|POST|PUT /api/empresa/facturacion_electronica/dian`
- `GET /api/empresa/facturacion_electronica/dian?action=guia_onboarding`
- `GET /api/empresa/facturacion_electronica/dian?action=checklist`
- `GET /api/empresa/facturacion_electronica/dian?action=validar`
- `POST /api/empresa/facturacion_electronica/dian?action=validar_credenciales`
- `POST /api/empresa/facturacion_electronica/dian?action=subir_firma`
- `POST /api/empresa/facturacion_electronica/dian?action=generar_cufe_demo`
- `POST /api/empresa/facturacion_electronica/dian?action=generar_xml_demo`
- `POST /api/empresa/facturacion_electronica/dian?action=generar_xml_ubl_base`
- `POST /api/empresa/facturacion_electronica/dian?action=firmar_xml_real`
- `POST /api/empresa/facturacion_electronica/dian?action=firmar_xml_xades_base`
- `GET /api/empresa/facturacion_electronica/dian?action=diagnostico_oficial`
- `POST /api/empresa/facturacion_electronica/dian?action=enviar_documento_real`
- `GET /api/empresa/facturacion_electronica/dian?action=consultar_acuse_real`
- `POST /api/empresa/facturacion_electronica/dian?action=reconexion_dian`
- `POST /api/empresa/facturacion_electronica/dian?action=enviar_set_pruebas`

## Persistencia canonica

### Tabla documental comun

- `empresa_facturacion_documentos`

Campos clave:

- `empresa_id`
- `tipo_documento`
- `documento_codigo`
- `numero_legal`
- `codigo_validacion`
- `pais_codigo`
- `ambiente_fe`
- `estado_documento`
- `estado_anterior`
- `evento_ultimo`
- `periodo_contable`
- `monto_total`
- `moneda`
- `fecha_documento`
- `entidad_relacionada_id`
- `estado`
- `observaciones`

Restriccion clave:

- unicidad por `empresa_id + tipo_documento + documento_codigo`

### Artefactos fiscales privados

- `empresa_facturacion_artefactos` conserva metadatos, SHA-256, MIME, tamano y referencia privada del XML firmado, respuesta del proveedor y representacion PDF.
- Los archivos viven fuera de `web/`, separados por `empresa_id`; ninguna API devuelve la ruta o el nombre interno de almacenamiento.
- Un reintento DIAN reutiliza el XML firmado persistido. No regenera fecha/CUFE para el mismo intento documental.

## Tipos documentales cubiertos

- `factura_electronica`
- `nota_credito`
- `nota_debito`
- `comprobante_pago`

Catalogados sin emision DIAN disponible:

- `documento_soporte` y su nota de ajuste
- `nomina_electronica` y su nota de ajuste
- `documento_equivalente_pos`, los demas documentos equivalentes y sus notas
  de ajuste
- `eventos_radian_recepcion`

## Fuentes de generacion documental

### Facturacion electronica explicita

- operaciones directas sobre `/api/empresa/facturacion_electronica`

### Venta por carrito o estacion

- `PUT /api/empresa/carritos_compra?action=pagar_estacion`
- el backend consulta `modo_documento_venta` de la empresa y genera automaticamente:
  - `factura_electronica`, o
  - `comprobante_pago`

## Maquina de estados canonica de facturacion

### Accion `emitir`

- estado anterior permitido: `borrador`, `pendiente_emision`
- estado nuevo: `emitida`
- evento contable: `factura_emitida`

### Accion `anular`

- estado anterior permitido: `emitida`
- estado nuevo: `anulada`
- evento contable: `factura_anulada`

### Accion `nota_credito` o `emitir_nota_credito`

- estado anterior permitido: `emitida`
- estado nuevo: `ajustada`
- evento contable: `nota_credito_emitida`

## Entradas obligatorias por operacion

### Configuracion de pais FE

- `empresa_id`
- `pais_codigo`

### Operaciones documentales

- `empresa_id`
- `documento_codigo`

### Emision normativa de factura electronica

- `empresa_id`
- `documento_codigo`
- `monto_total`
- `moneda`
- `pais_codigo` directo o resoluble por configuracion

### Reenvio de correo

- `empresa_id`
- `documento_codigo`

### DIAN Colombia base

- `empresa_id`
- segun `action`, el payload operativo correspondiente

## Entradas opcionales relevantes

- `tipo_documento`
- `cliente_id`
- `cliente_email`
- `cliente_nombre`
- `entidad_id`
- `periodo_contable`
- `observaciones`
- `estado_actual`
- filtros de consulta documental: `tipo_documento`, `estado_documento`, `cliente`, `documento`, `q`, `fecha_desde`, `fecha_hasta`, `limit`, `offset`

## Invariantes

1. Todo documento transaccional de facturacion queda aislado por `empresa_id`.
2. La tabla `empresa_facturacion_documentos` es la fuente canonica del estado documental operativo de facturas, notas credito y comprobantes de pago.
3. Ninguna operacion documental puede ejecutarse sin `empresa_id > 0`.
4. Ninguna operacion documental puede ejecutarse sin `documento_codigo` no vacio.
5. La transicion documental debe respetar la maquina de estados definida; una accion fuera del estado permitido devuelve conflicto.
6. La emision de `factura_electronica` exige validacion previa de cumplimiento normativo mediante `PrepareFacturacionDocumentoLegal`.
7. Cuando una emision FE supera la validacion normativa, el documento debe persistir `numero_legal`, `pais_codigo`, `ambiente_fe` y `fecha_documento`. En Colombia `codigo_validacion` permanece vacio hasta obtener un CUFE/CUDE SHA-384 del XML o acuse DIAN; nunca se sustituye por un hash local.
8. Si el documento ya existia, la persistencia conserva los campos legales previos cuando el payload nuevo no los sobreescribe.
9. Toda operacion documental exitosa registra evento contable no bloqueante.
10. La integracion fiscal posterior a `emitir`, `anular` o `nota_credito` no bloquea la persistencia del documento; su resultado se refleja aparte como `integracion_fiscal` y opcionalmente en cola de reintentos.
11. El correo al cliente es un side effect no bloqueante del flujo de emision FE y del action `reenviar_correo`; su ausencia o fallo no revierte el documento emitido.
12. El destinatario del correo se resuelve primero por payload y luego por `cliente_id` en la empresa; si no existe correo valido, se reporta el motivo sin romper la operacion principal.
13. La configuracion por pais FE es por `empresa_id + pais_codigo` y debe mantenerse separada del estado documental de cada factura.
14. `modo_documento_venta` decide el tipo documental generado al cerrar una venta; el flujo de cobro es comun y la diferencia ocurre en el documento persistido.
15. La base DIAN Colombia actual es operativa para onboarding, validacion, firma base y pruebas reales de habilitacion; `pruebas_dian` y `enviar_set_pruebas` no aceptan simulacion.
16. El software DIAN puede operar en modo `compartido` o `empresa`; el software compartido no elimina la obligacion de token y firma por empresa.
17. Las referencias sensibles DIAN (`token_emisor_ref`, `certificado_clave_ref`, software compartido) deben resolverse por `env:`, `file:` o `base64:`; no deben quedar como secretos en codigo fuente.
18. El set de pruebas DIAN respeta consecutivos y rango configurado; si el rango no alcanza, la operacion debe fallar con conflicto.
19. La documentacion debe distinguir explicitamente entre `firma base` y `firma oficial`, y entre `envio real base` y `transporte oficial DIAN`.
20. Para Colombia no existe un objetivo numerico local por defecto. La empresa debe copiar y guardar el objetivo exacto asignado por el portal DIAN; sin ese dato el envio automatico del set queda bloqueado.
21. En produccion Colombia, un documento solo queda fiscalmente aceptado cuando DIAN devuelve acuse aceptado, el portal DIAN lo muestra aprobado o existe evidencia oficial equivalente; `Regla 90` por si sola deja el caso pendiente de consulta.
22. `Aprobado con notificacion` en portal DIAN cuenta como documento aprobado; la notificacion debe conservarse como observacion operativa y corregirse si apunta a datos maestros.
23. Antes de reenviar un documento con el mismo prefijo/folio, el sistema u operador debe consultar historial DIAN, cola de reintentos, CUFE/TrackId o portal para evitar duplicados y consumo de consecutivos.
24. Si se ejecuta una prueba directa contra DIAN por fuera de `empresa_facturacion_documentos`, debe registrarse en historial de cambios y adelantar `empresa_dian_configuracion.consecutivo_actual` y `empresa_configuracion_avanzada.proximo_consecutivo` al siguiente folio disponible.
25. Una factura comercial Colombia solo puede generarse desde un
    `fuente_fiscal_json` cargado por el servidor y perteneciente a la misma
    empresa/documento; cuerpo HTTP, valores por defecto y líneas de habilitación
    no son fuentes fiscales válidas.
26. La fuente conserva las líneas y partes existentes al pagar. Su SHA-256 es
    inmutable; cualquier diferencia, dato DANE ausente, total no conciliado,
    unidad o impuesto desconocido bloquea antes de firma y transporte.
27. Nota crédito y nota débito no reutilizan la factura ni un fixture sintético:
    requieren una fuente de ajuste propia y referencia verificable a una factura
    aceptada. Mientras falte, catálogo y endpoint deben responder bloqueados.
25. Un `GET` de conectividad, listado o consulta nunca transmite documentos. El procesamiento manual de cola exige `POST`; el automatico pertenece a `pcs-worker`.
26. Una respuesta asincrona con TrackId queda `pendiente`; el worker consulta `GetStatusZip` con el mismo XML/TrackId hasta aceptacion, rechazo o agotamiento controlado. Un rechazo final no se reenvia automaticamente.
27. No se fabrican referencias externas. Si DIAN/proveedor no entrega TrackId, ZipKey o referencia equivalente, el campo permanece vacio.
28. Toda descarga y adjunto fiscal vuelve a calcular SHA-256 contra los metadatos tenant-scoped; una diferencia de integridad bloquea la entrega.
29. Al reutilizar un XML firmado, la fecha fiscal se toma de `IssueDate`/`IssueTime` del mismo XML y no se reemplaza por la hora del reintento.
30. El envío o consulta de acuse de un mismo `empresa_id + tipo_documento + documento_codigo` usa bloqueo asesor documental; una segunda ejecución concurrente no transmite el documento.

## Salidas y estados funcionales

### Configuracion FE por pais

- `200` con `configuracion` guardada o consultada
- `400` si faltan `empresa_id` o `pais_codigo`
- `500` si falla lectura no controlada del backend

### Operaciones documentales generales

- `200` con:
  - `accion`
  - `evento`
  - `estado_anterior`
  - `estado_nuevo`
  - `entidad_id`
  - `documento_codigo`
  - `numero_legal`
  - `codigo_validacion`
  - `pais_codigo`
  - `ambiente_fe`
  - `integracion_fiscal`
  - `cola_reintentos` si aplica
  - `cumplimiento_normativo` cuando hubo preparacion legal
  - `factura_email` cuando se emite una factura electronica
- `400` por payload faltante o invalido
- `404` si el documento a reenviar no existe
- `409` si la transicion documental es invalida
- `422` si falla el cumplimiento normativo previo a emitir
- `500` si falla persistencia o integracion no controlada

### Listado de documentos

- `200` con `items[]` filtrados por empresa

### Cola de reintentos y reconciliacion

- `200` con resumen operativo o items de cola
- `400` si `empresa_id` o parametros numericos son invalidos

### DIAN Colombia base

- `200` con `ok=true` o reporte funcional de base
- `400` por payload faltante, firma invalida, referencias secretas vacias o configuracion inexistente
- `405` si el metodo HTTP no corresponde al action
- `409` por rango insuficiente o conflicto operativo del set

## Integracion fiscal generica

El flujo normal de `backend/handlers/facturacion_electronica.go` procesa una integracion fiscal generica posterior a la persistencia documental.

Estados observables:

- `no_aplica`
- `pendiente`
- `enviado`
- `fallido`
- `contingencia`
- `reconciliado`

Metadatos observables:

- `accion`
- `pais_codigo`
- `proveedor`
- `ambiente`
- `estado_envio`
- `intentos`
- `max_intentos`
- `proximo_intento`
- `contingencia_activa`
- `referencia_externa`
- `error`

## Correo al cliente

### Resolucion del destinatario

Orden de prioridad:

1. `cliente_email` del payload
2. `cliente_id` o `entidad_id` resuelto contra clientes de la misma empresa

### Comportamiento

- para `factura_electronica`: asunto y cuerpo fiscal con `numero_legal`, `codigo_validacion`, `monto_total`, `moneda`, `pais_codigo` y `ambiente_fe`
- en Colombia produccion, el correo fiscal espera `estado_envio=aceptado`; XML firmado y representacion PDF son obligatorios y el envio falla cerrado si falta alguno o no supera la validacion de integridad
- para `comprobante_pago`: asunto y cuerpo comercial equivalentes, sin forzar detalle FE
- si SMTP global no esta disponible, el handler responde el error dentro de `factura_email` sin deshacer el documento

## Contrato DIAN Colombia: alcance real actual

### Capacidades implementadas hoy

- CRUD de configuracion DIAN por empresa
- guia de onboarding
- checklist funcional y validacion
- validacion de credenciales y referencias secretas
- carga segura de firma PEM por empresa
- generacion demo de CUFE y XML
- generacion de XML UBL base interna
- firma RSA-SHA256 real sobre XML base
- firma XAdES base interna no oficial
- diagnostico de brechas frente al objetivo oficial
- envio base de documento real usando la capa actual
- consulta base de acuse
- reconexion operativa
- ejecucion de set de pruebas con envio real, `ZipKey` y consulta de acuse `GetStatusZip`

### Limites explicitamente vigentes

- el cliente SOAP/WSDL base ya construye sobres para `SendBillAsync`, `SendBillSync`, `SendTestSetAsync` y `GetStatusZip`, pero el cierre certificable depende de ejecutar acuses reales con credenciales/firma de la empresa.
- `GetNumberingRange` permanece como objetivo declarado para automatizar consulta de rangos; los rangos de produccion deben estar configurados y asociados en DIAN antes de emitir.
- la firma XMLDSig/XAdES base existe para validacion tecnica, pero la aceptacion fiscal final depende del acuse DIAN/proveedor y de las politicas vigentes de DIAN para la empresa.
- el XML firmado, el acuse del proveedor y la representacion PDF se persisten en almacenamiento privado por empresa y pueden descargarse desde la bandeja de facturas; el XML y PDF se adjuntan al correo posterior a la aceptacion.

## Errores de contrato esperados

- `empresa_id` faltante o no resoluble: `400`
- `documento_codigo` faltante en operacion documental: `400`
- `fecha_desde` o `fecha_hasta` fuera de formato `YYYY-MM-DD`: `400`
- `limit` o `offset` invalidos: `400`
- `cliente_email` mal formado: error encapsulado en `factura_email`
- transicion fuera del estado permitido: `409`
- intento de emitir FE sin resolucion/prefijo/consecutivo legal valido: `422`
- reenvio de correo sobre documento inexistente: `404`
- firma DIAN sin `certificado_clave_ref`: `400`
- set de pruebas con rango insuficiente: `409`
- configuracion DIAN inexistente para empresa: `400`

## Side effects obligatorios

- upsert en `empresa_facturacion_documentos`
- registro de evento contable no bloqueante
- posible registro de integracion fiscal y cola de reintentos
- persistencia privada de XML firmado antes de transmitir, acuse del proveedor y representacion PDF con SHA-256
- consulta automatica de acuses pendientes mediante `pcs-worker`
- posible envio de correo al cliente
- actualizacion de consecutivo o metadata DIAN cuando aplica set de pruebas o carga de firma

## Evidencia tecnica minima

- `backend/handlers/eventos_contables_modulos_test.go`
- `backend/handlers/facturacion_electronica_reintentos_test.go`
- `backend/handlers/carrito_facturacion_impresion_test.go`
- `backend/handlers/modulos_faltantes_test.go`
- `backend/db/facturacion_electronica_test.go`

## Contratos relacionados

- `documentos/gobernanza_tecnica/contratos/contrato_permisos_contexto_y_wrappers_api_empresa.md`
- `documentos/gobernanza_tecnica/contratos/contrato_estaciones_sensores_ventas_simple.md`
- `documentos/gobernanza_tecnica/contratos/contrato_venta_publica_empresarial_por_empresa.md`
