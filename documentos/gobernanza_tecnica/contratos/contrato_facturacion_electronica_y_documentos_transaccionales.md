# Contrato tecnico: facturacion electronica y documentos transaccionales

## Numeracion multicaja y offline (candidato 2026-09-06)

La reserva de factura CO persiste por empresa/documento en la misma transaccion
que incrementa el contador. Repetir la operacion conserva numero y fecha;
cambiar importe, moneda o ambiente falla. No se renumeran documentos historicos.
Un formulario viejo no reduce el contador de la misma serie fiscal.

Offline exige empresa, operador original autenticado, caja y sync_key estable.
El navegador requiere Web Locks y almacenamiento legible/disponible; solo
imprime el comprobante provisional despues de guardar. El servidor limita
lotes y reclama cada key una vez por empresa. Sincronizar una venta no valida
la factura ni implementa contingencia legal DIAN.

El pago persiste en `commerce.sale-paid` la decision documental resuelta bajo el
mismo `FOR UPDATE` que avanza la frecuencia por empresa. El worker reconstruye
el documento faltante idempotentemente, despues de recuperar contabilidad, sin
repetir el cobro ni volver a decidir por configuracion mutable.

La contingencia Colombia es un dominio separado por empresa: historial de
autorizaciones de papel, incidente, documento y plazo. Las relaciones usan
claves compuestas con `empresa_id`, cada incidente congela la autorizacion que
le corresponde y una renovacion no borra ni mezcla la serie anterior. Solo se
marca estado `contingencia` ante
falla de conectividad si existe un incidente DIAN activo. Para falla del
facturador, registrar papel exige venta pagada, comprobante/fuente inmutables,
rango vigente y numero siguiente exacto. El documento electronico tipo 03 no se
delega al generador tipo 01: sigue bloqueado hasta implementar CUDE,
`AdditionalDocumentReference`, validacion XSD y transporte/acuses especificos.

Evidencia y limites: `documentos/auditoria_facturacion_multicaja_offline_20260906.md`.

## Frontera fiscal por pais y empresa (2026-09-05)

La configuracion heredada exige igualdad con el pais de
`empresa_configuracion_avanzada`. Un pais desconocido no se convierte en CO.
La reserva legal comun solo opera para CO; otros paises necesitan numeracion
y adaptador propios. El transporte requiere configuracion activa y coincidencia
de empresa/pais/ambiente; solo CO con proveedor DIAN tiene adaptador productivo.
Manual/local/mock y HTTP generico no acreditan aceptacion fiscal.
Los checklists EC/PA exponen `emision_habilitada=false`, incluso completos.
No se alteran configuraciones persistidas ni documentos historicos.
Ver `documentos/auditoria_facturacion_seguridad_20260905.md`; candidato, no
certificacion integral de produccion.

Fecha: 2026-04-18
Estado: vigente

Actualizacion 2026-08-26 - Nomina electronica mensual (candidato local):

- `nomina_electronica` ordinaria deja de reutilizar el adaptador comercial. PCS
  consolida en servidor todas las liquidaciones activas y pagadas de un
  trabajador dentro de un mes calendario cerrado y genera un solo
  `NominaIndividual` por empresa/trabajador/mes.
- La reserva atomica sella fuente/configuracion, conserva consecutivo y fecha,
  calcula CUNE SHA-384, firma XAdES, valida el XML y transmite en produccion con
  `SendNominaSync`. Un reintento reutiliza el XML ya firmado.
- Configuracion de familia, perfil fiscal del trabajador e identidad del
  proveedor del software son explicitos. No se reutiliza la resolucion/rango de
  factura y los secretos no entran en la fuente fiscal.
- Emision y reenvio requieren aprobacion efectiva en Facturacion y Nomina. La
  lectura de documentos, cola, configuracion y artefactos de nomina exige
  lectura en ambos dominios; el listado general la omite cuando falta Nomina.
- El procesamiento manual general de cola omite nomina. El reenvio individual
  exige `REENVIAR NOMINA ELECTRONICA DIAN`; el worker conserva el reintento
  automatico del documento ya autorizado.
- Permanecen cerrados `NominaIndividualDeAjuste`, habilitacion automatica por
  `SendTestSetAsync` y la representacion/entrega dedicada. El correo generico de
  factura rechaza nomina antes de generar o adjuntar un PDF incorrecto.
- El XML firmado de fixture paso el XSD oficial 1.0.6. No hubo emision real de
  nomina en este candidato; PR, CI, despliegue, migracion y prueba fiscal son
  compuertas separadas.

Actualizacion 2026-08-25:

- La factura comercial no admite emision libre: solo
  `action=facturar_desde_venta` sobre una venta pagada puede crearla, y la fuente
  fiscal inmutable debe validar antes de reservar numeracion.
- La anulacion total toma un bloqueo transaccional sobre la factura origen,
  reutiliza cualquier nota ya relacionada y crea una nueva como
  `pendiente_emision`. La factura solo cambia a `anulada` si la nota esta
  `emitida`, su CUDE SHA-384 coincide con el de la cola y el acuse esta
  `aceptado` o `reconciliado`.
- El correo fiscal colombiano de produccion espera documento emitido y
  aceptacion DIAN. Una respuesta pendiente, fallida o rechazada nunca envia al
  cliente una factura como definitiva.
- Las acciones DIAN con efectos externos o fiscales exigen permiso `A`; la
  anulacion total conserva permiso `D`. Los chequeos de vencimiento silenciosos
  son lectura. Toda accion GET DIAN desconocida responde 400 antes del CRUD y
  antes de cualquier posible exposicion de secretos descifrados.
- El listado publica solo el booleano `fuente_fiscal_disponible`; nunca expone
  ruta, hash ni contenido de la fuente. Una factura historica sin esa fuente no
  ofrece anulacion aunque conserve un estado o codigo fiscal legado.
- `reconciliar_aceptados_local` completa documento y cola exclusivamente desde
  un acuse ya aceptado/reconciliado y omite el resto antes de cualquier
  despacho. Esta reparacion local no incrementa intentos, no cambia fechas de
  envio y no retransmite XML. `reconciliar_estados` conserva el procesamiento
  operativo general y no sustituye este modo restringido.
- `nota_credito` solo se emite mediante `anular_factura_nota_credito` sobre una
  factura aceptada con CUFE oficial y fuente fiscal inmutable. El servidor
  deriva y sella una fuente separada para la nota; no acepta lineas, partes,
  impuestos ni referencia fiscal libres del cliente.
- La fuente de la nota conserva el numero legal, CUFE y fecha fiscal de la
  factura. El UBL `CreditNote` usa CUDE, `DiscrepancyResponse`,
  `BillingReference` y todas las lineas reales. La factura original permanece
  emitida hasta que DIAN acepte la nota.
- `nota_credito` generica/parcial y `nota_debito` siguen en HTTP 422. El catalogo
  publica la anulacion total como capacidad parcial, no como emision generica.

Actualizacion 2026-08-24:

- El candidato actual habilita para emision comercial DIAN exclusivamente
  `factura_electronica`. Su XML UBL se construye desde `fuente_fiscal_json`
  inmutable, capturada del carrito pagado y aislada por `empresa_id`.
- `nota_credito` y `nota_debito` se conservan como registros historicos y de
  consulta, pero toda emision/anulacion fiscal de esas familias responde HTTP
  422 hasta disponer de una fuente inmutable de ajuste, referencia DIAN y
  adaptador UBL/CUDE propios. No consumen consecutivo ni transmiten datos.
- Los actions directos `firmar_xml_real`, `firmar_xml_xades_base` y
  `enviar_documento_real` estan cerrados para emisiones comerciales. La firma
  y el transporte solo los invoca el despachador interno luego de validar la
  fuente, la resolucion, el consecutivo y el preflight.
- Un XML firmado persistido solo puede reintentarse cuando existe y vuelve a
  validar la fuente fiscal del mismo documento. Los artefactos heredados sin
  esa trazabilidad quedan bloqueados de forma segura.
- La evidencia de portal indicada en actualizaciones anteriores es historica;
  no sustituye las pruebas de habilitacion, XSD, Schematron, PostgreSQL ni la
  aprobacion vigente requeridas antes de declarar produccion.
- Verificacion real y visual de PCS (`empresa_id=12`) en produccion: las
  credenciales/firma superaron el preflight, el diagnostico quedo en
  `pre_envio_validable` y `GetNumberingRange` devolvio la resolucion
  `18764111318575`, prefijo `1PCS`, rango `1-100000` y vigencia
  `2026-06-17` a `2028-06-17`. La prueba de conectividad solo demostro
  alcance del endpoint; no equivale a aceptacion SOAP.
- Una reconsulta real de un TrackId historico devolvio codigo DIAN `66`
  (identificador inexistente). El binario desplegado degrado indebidamente el
  estado global a `rechazado`; la prueba se revirtio de inmediato a `enviado`
  conservando el rechazo individual. El candidato registra `GetStatusZip`
  exclusivamente en `empresa_dian_track_historial` y tiene una prueba de
  contrato que impide reintroducir esa mutacion global.
- La barra visual mide solo verificaciones ejecutadas durante la sesion actual.
  No hereda un `aceptado` global como acuse vigente ni presenta el 100 % como
  certificacion de todas las familias documentales.

Actualizacion historica 2026-08-21 (supeditada al alcance 2026-08-24):

- En esa revision se considero que el adaptador cubria la familia del Anexo
  Tecnico de Factura Electronica de Venta. La auditoria posterior limito la
  emision comprobada a `factura_electronica`; `nota_credito` y `nota_debito`
  permanecen bloqueadas hasta contar con fuente de ajuste y adaptador propios.
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
- `POST|PUT /api/empresa/facturacion_electronica?action=nota_credito` (bloqueado
  para emision fiscal DIAN actual)
- `POST|PUT /api/empresa/facturacion_electronica?action=emitir_nota_credito`
  (bloqueado para emision fiscal DIAN actual)
- `POST|PUT /api/empresa/facturacion_electronica?action=reenviar_correo`
- `POST /api/empresa/facturacion_electronica?action=emitir_nomina_electronica`
- `POST /api/empresa/facturacion_electronica?action=procesar_reintentos`
- `POST /api/empresa/facturacion_electronica?action=reconciliar_estados`
- `POST /api/empresa/facturacion_electronica?action=reconciliar_aceptados_local`

### Fuente y preflight de nomina electronica

- `GET /api/empresa/nomina?action=perfil_dian&empleado_nomina_id={id}`
- `POST /api/empresa/nomina?action=perfil_dian`
- `GET /api/empresa/nomina?action=documentos_electronicos_colombia`
- `GET /api/empresa/nomina?action=nomina_electronica_preflight&liquidacion_id={id}`
- `POST /api/empresa/nomina?action=preparar_nomina_electronica` (revision sin
  reserva, firma ni transporte)

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
  (solo fixture explicito de habilitacion, no venta comercial)
- `POST /api/empresa/facturacion_electronica/dian?action=firmar_xml_real`
  (cerrado para uso directo)
- `POST /api/empresa/facturacion_electronica/dian?action=firmar_xml_xades_base`
  (cerrado para uso directo)
- `GET /api/empresa/facturacion_electronica/dian?action=diagnostico_oficial`
- `POST /api/empresa/facturacion_electronica/dian?action=enviar_documento_real`
  (cerrado para uso directo)
- `GET /api/empresa/facturacion_electronica/dian?action=consultar_acuse_real`
- `POST /api/empresa/facturacion_electronica/dian?action=reconexion_dian`
- `POST /api/empresa/facturacion_electronica/dian?action=enviar_set_pruebas`
- `POST /api/empresa/facturacion_electronica/dian?action=get_numbering_range`
- `POST /api/empresa/facturacion_electronica/dian?action=consultar_acuse_real`

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
- Un reintento DIAN reutiliza el XML firmado persistido solo si existe una
  `fuente_fiscal_json` inmutable del mismo documento y esta vuelve a superar
  los bloqueos fiscales. No regenera fecha/CUFE para el mismo intento documental.

### Fuente canonica de nomina electronica

- `empresa_nomina_liquidaciones` y `empresa_nomina_pagos` aportan la fuente
  operativa real; el cliente HTTP no puede reemplazar sus totales o partes.
- `empresa_nomina_dian_perfiles` aporta los atributos fiscales explicitos del
  trabajador, siempre por `empresa_id + empleado_nomina_id`.
- `empresa_contabilidad_nomina_electronica` conserva un espejo por
  `empresa_id + empleado_nomina_id + periodo_reporte`, numero/fecha legal,
  CUNE, respuesta, intentos y las instantaneas selladas.
- El documento comun usa `tipo_documento=nomina_electronica`; sus artefactos y
  cola permanecen privados para quienes no tengan permiso de Nomina.

## Tipos documentales cubiertos

- `factura_electronica` desde venta pagada.
- `nota_credito` solo para anulacion total de una factura aceptada;
  `nota_credito` libre/parcial y `nota_debito` siguen bloqueadas.
- `documento_soporte` ordinario desde borrador de compra estructurado; su nota
  de ajuste sigue bloqueada.
- `nomina_electronica` ordinaria desde liquidaciones/pagos mensuales y perfil
  fiscal; su nota de ajuste sigue bloqueada.
- `comprobante_pago` comercial.

Catalogados sin emision DIAN disponible:

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

- Actualmente bloqueada con HTTP 422 para DIAN. No cambia el estado fiscal de
  la factura ni crea una nota hasta contar con referencia, fuente de ajuste y
  adaptador DIAN propios.

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
10. La integracion fiscal posterior a `emitir` de factura puede quedar en cola
    solo despues de capturar la fuente fiscal y reservar el consecutivo. Las
    notas y anulaciones DIAN no se simulan ni se persisten como emitidas.
11. El correo al cliente es un side effect no bloqueante del flujo de emision FE y del action `reenviar_correo`; su ausencia o fallo no revierte el documento emitido.
12. El destinatario del correo se resuelve primero por payload y luego por `cliente_id` en la empresa; si no existe correo valido, se reporta el motivo sin romper la operacion principal.
13. La configuracion por pais FE es por `empresa_id + pais_codigo` y debe mantenerse separada del estado documental de cada factura.
14. `modo_documento_venta` decide el tipo documental generado al cerrar una venta; el flujo de cobro es comun y la diferencia ocurre en el documento persistido.
15. La base DIAN Colombia permite onboarding, validacion y pruebas reales de
    habilitacion por el flujo interno; las firmas y envios directos estan
    bloqueados y `pruebas_dian`/`enviar_set_pruebas` no aceptan simulacion.
16. El software DIAN puede operar en modo `compartido` o `empresa`; el software compartido no elimina la obligacion de token y firma por empresa.
17. Las credenciales DIAN se cifran por empresa y campo al persistirse; no
    deben salir en respuestas, logs, codigo fuente ni historial.
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
28. Un `GET` de conectividad, listado o consulta nunca transmite documentos. El procesamiento manual de cola exige `POST`; el automatico pertenece a `pcs-worker`.
29. Una respuesta asincrona con TrackId queda `pendiente`; el worker consulta `GetStatusZip` con el mismo XML/TrackId hasta aceptacion, rechazo o agotamiento controlado. Un rechazo final no se reenvia automaticamente.
30. La reconsulta de un TrackId historico actualiza exclusivamente `empresa_dian_track_historial`; nunca cambia el estado global de `empresa_dian_configuracion`.
31. No se fabrican referencias externas. Si DIAN/proveedor no entrega TrackId, ZipKey o referencia equivalente, el campo permanece vacio.
32. Toda descarga y adjunto fiscal vuelve a calcular SHA-256 contra los metadatos tenant-scoped; una diferencia de integridad bloquea la entrega.
33. Al reutilizar un XML firmado, la fecha fiscal se toma de `IssueDate`/`IssueTime` del mismo XML y no se reemplaza por la hora del reintento.
34. El envío o consulta de acuse de un mismo `empresa_id + tipo_documento + documento_codigo` usa bloqueo asesor documental; una segunda ejecución concurrente no transmite el documento.
35. Nomina ordinaria se consolida por mes calendario cerrado. Dos
    liquidaciones del mismo trabajador/mes producen una sola fuente; periodos
    cruzados o solapados bloquean antes de numerar.
36. Cada liquidacion incluida exige un unico pago activo real y se conservan
    todas las fechas de pago. PCS no fabrica pagos, intervalos horarios ni
    atributos fiscales del trabajador.
37. La reserva de nomina es idempotente por empresa/trabajador/mes. Fuente,
    configuracion, numero y fecha fiscal son inmutables para reintentos.
38. `SendNominaSync` recibe solo `contentFile`; `TestSetId` pertenece al flujo
    separado de habilitacion y no puede agregarse al sobre de produccion.
39. Emision y reenvio de nomina exigen permiso de aprobacion de Facturacion y
    Nomina. Lectura/descarga explicita exige lectura en ambos dominios.
40. Un listado general sin permiso de Nomina excluye esa familia antes de
    aplicar paginacion. No entrega documento, CUNE, total, cola, configuracion
    ni hash de artefactos de nomina.
41. El procesamiento manual general de cola no transmite nomina. Solo el
    reenvio individual con la frase fuerte puede solicitarlo; el worker puede
    continuar el documento que ya fue autorizado y sellado.
42. El correo/PDF de factura nunca se reutiliza para nomina; hasta existir una
    representacion y entrega dedicadas, esa accion responde bloqueada.

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
- si falta lectura de Nomina, el listado general excluye esa familia; pedirla
  expresamente responde `403`

### Cola de reintentos y reconciliacion

- `200` con resumen operativo o items de cola
- `400` si `empresa_id` o parametros numericos son invalidos
- el procesamiento manual general informa y omite nomina; el worker conserva
  la cola automatica del documento ya autorizado

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
- para `nomina_electronica`: bloqueado; no usa asunto, destinatario,
  representacion PDF ni adjuntos de factura. La entrega requiere su contrato
  dedicado de nomina.
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
- firma XMLDSig/XAdES verificada localmente contra su certificado y validacion
  UBL XSD; la validacion Schematron oficial y el acuse DIAN final siguen siendo
  puertas obligatorias antes de produccion
- diagnostico de brechas frente al objetivo oficial
- envio del adaptador interno de factura comercial desde fuente fiscal
  inmutable; no existe emision generica libre
- emision dedicada de documento soporte ordinario desde su borrador de compra
  y fuente fiscal propia
- generacion/preflight y transporte productivo dedicado de
  `NominaIndividual` mensual desde liquidaciones/pagos reales, con CUNE,
  XAdES, XSD oficial y `SendNominaSync`
- consulta base de acuse
- reconexion operativa
- ejecucion de set de pruebas con envio real, `ZipKey` y consulta de acuse `GetStatusZip`

### Limites explicitamente vigentes

- el cliente SOAP/WSDL base ya construye sobres para `SendBillAsync`, `SendBillSync`, `SendTestSetAsync` y `GetStatusZip`, pero el cierre certificable depende de ejecutar acuses reales con credenciales/firma de la empresa.
- `GetNumberingRange` esta implementado y fue comprobado contra DIAN para PCS;
  la respuesta debe coincidir exactamente con resolucion, prefijo, rango y
  vigencia configurados antes de actualizar la clave tecnica.
- la firma XMLDSig/XAdES supera la validacion criptografica y XSD local. La
  corrida Schematron oficial permanece pendiente por falta de un procesador
  XSLT 3 compatible en el entorno actual, y la aceptacion fiscal final depende
  siempre del acuse DIAN/proveedor.
- `factura_electronica` tiene emision comercial; `nota_credito` solo la ruta
  especializada de anulacion total; documento soporte ordinario y nomina
  ordinaria tienen adaptadores propios. Nota debito, notas de ajuste,
  equivalentes POS y RADIAN siguen bloqueados de forma segura.
- la habilitacion automatica de nomina mediante `SendTestSetAsync` y
  `TestSetId` no esta implementada. Produccion solo puede usarse despues de una
  habilitacion oficial verificada por fuera de este action.
- la nomina ordinaria aun requiere prueba real posterior al despliegue con una
  fuente genuina; el fixture XSD no es un acuse DIAN.
- XML firmado y acuse se persisten en almacenamiento privado por empresa. La
  factura y documento soporte conservan su PDF propio; nomina no genera ni
  envia el PDF generico de facturacion.

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
- consulta o artefacto explicito de nomina sin permiso adicional: `403`
- mes abierto, fuente mensual ambigua o preflight de nomina incompleto: `422`
- emision de nomina sin frase exacta o reenvio sin su frase fuerte: `409`
- correo generico solicitado para nomina: `409`

## Side effects obligatorios

- upsert en `empresa_facturacion_documentos`
- registro de evento contable no bloqueante
- posible registro de integracion fiscal y cola de reintentos
- persistencia privada de XML firmado antes de transmitir, acuse del proveedor y representacion PDF con SHA-256
- consulta automatica de acuses pendientes mediante `pcs-worker`
- posible envio de correo al cliente
- actualizacion de consecutivo o metadata DIAN cuando aplica set de pruebas o carga de firma
- para nomina, reserva mensual atomica, sellado de fuente/configuracion y espejo
  en documento/cola; ninguno de estos efectos ocurre durante el preflight

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
