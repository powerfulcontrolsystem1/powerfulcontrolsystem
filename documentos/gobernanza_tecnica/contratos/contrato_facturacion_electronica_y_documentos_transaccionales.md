# Contrato tecnico: facturacion electronica y documentos transaccionales

Fecha: 2026-04-18
Estado: vigente

## Alcance

Este contrato cubre el ciclo documental empresarial de facturacion y documentos de venta, la persistencia comun en `empresa_facturacion_documentos`, la configuracion por pais de facturacion electronica, el envio automatico de resumen por correo al cliente, la cola de reintentos y reconciliacion fiscal, y la base operativa actual del endpoint DIAN Colombia.

## Endpoints cubiertos

### Facturacion electronica general

- `GET /api/empresa/facturacion_electronica?empresa_id={id}`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&pais_codigo=CO`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&action=documentos`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&action=reintentos`
- `GET /api/empresa/facturacion_electronica?empresa_id={id}&action=reconciliacion`
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

## Tipos documentales cubiertos

- `factura_electronica`
- `nota_credito`
- `comprobante_pago`

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
7. Cuando una emision FE supera la validacion normativa, el documento debe persistir `numero_legal`, `codigo_validacion`, `pais_codigo`, `ambiente_fe` y `fecha_documento`.
8. Si el documento ya existia, la persistencia conserva los campos legales previos cuando el payload nuevo no los sobreescribe.
9. Toda operacion documental exitosa registra evento contable no bloqueante.
10. La integracion fiscal posterior a `emitir`, `anular` o `nota_credito` no bloquea la persistencia del documento; su resultado se refleja aparte como `integracion_fiscal` y opcionalmente en cola de reintentos.
11. El correo al cliente es un side effect no bloqueante del flujo de emision FE y del action `reenviar_correo`; su ausencia o fallo no revierte el documento emitido.
12. El destinatario del correo se resuelve primero por payload y luego por `cliente_id` en la empresa; si no existe correo valido, se reporta el motivo sin romper la operacion principal.
13. La configuracion por pais FE es por `empresa_id + pais_codigo` y debe mantenerse separada del estado documental de cada factura.
14. `modo_documento_venta` decide el tipo documental generado al cerrar una venta; el flujo de cobro es comun y la diferencia ocurre en el documento persistido.
15. La base DIAN Colombia actual es operativa para onboarding, validacion, firma base, pruebas y simulaciones, pero no equivale aun a la integracion SOAP/WSDL oficial completa.
16. El software DIAN puede operar en modo `compartido` o `empresa`; el software compartido no elimina la obligacion de token y firma por empresa.
17. Las referencias sensibles DIAN (`token_emisor_ref`, `certificado_clave_ref`, software compartido) deben resolverse por `env:`, `file:` o `base64:`; no deben quedar como secretos en codigo fuente.
18. El set de pruebas DIAN respeta consecutivos y rango configurado; si el rango no alcanza, la operacion debe fallar con conflicto.
19. La documentacion debe distinguir explicitamente entre `firma base` y `firma oficial`, y entre `envio real base` y `transporte oficial DIAN`.

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
- ejecucion de set de pruebas con simulacion o envio

### Limites explicitamente vigentes

- el cliente SOAP/WSDL oficial DIAN aun no esta integrado en el flujo normal de facturacion
- `SendBillAsync`, `SendBillSync`, `SendTestSetAsync`, `GetStatusZip` y `GetNumberingRange` permanecen como objetivo declarado, no como garantia ya cerrada del flujo principal
- el empaquetado ZIP oficial, `TrackId` y la firma XMLDSig/XAdES certificable final siguen pendientes
- el XML UBL generado hoy es base interna y debe tratarse como preparacion tecnica, no como cumplimiento final certificable por si solo

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