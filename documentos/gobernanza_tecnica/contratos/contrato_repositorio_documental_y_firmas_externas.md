# Contrato tecnico: repositorio documental y firmas externas

Fecha: 2026-04-18
Estado: vigente

## 1. Alcance del flujo

Este contrato formaliza el comportamiento actual del repositorio documental empresarial y del registro de firmas externas sobre documentos versionados.

Aplica a:

- `/api/empresa/documentos/gestion`
- `/api/empresa/documentos/firmas`
- control de acceso por rol y modulo documental
- versionado de documentos con historial visible
- consulta de acceso puntual y por listado

No promete almacenamiento binario, validacion criptografica avanzada de certificados ni firma electronica oficial ante terceros. El alcance real hoy es persistencia de metadatos, control de acceso por rol y trazabilidad documental multiempresa.

## 2. Rutas y acciones implicadas

### 2.1 `/api/empresa/documentos/gestion`

- CRUD generico empresarial sobre `empresa_documentos_gestion`.
- `GET ?action=acceso&empresa_id={id}&id={documento_id}`: valida acceso sobre un documento concreto.
- `GET ?action=acceso&empresa_id={id}&modulo={modulo}`: valida acceso por modulo documental sin id concreto.
- `GET ?action=repositorio&empresa_id={id}`: lista repositorio documental filtrando por acceso.
- `GET ?action=versiones&empresa_id={id}&id={documento_id}`: lista historial de versiones tomando el `documento_codigo` del documento base.
- `GET ?action=versiones&empresa_id={id}&documento_codigo={codigo}`: lista historial por codigo documental.
- `POST|PUT|PATCH ?action=versionar`: crea una nueva version documental y marca la anterior como historica.

Alias aceptados:

- `action=repository` como alias de `repositorio`
- `action=historial_versiones` como alias de `versiones`

### 2.2 `/api/empresa/documentos/firmas`

- CRUD generico empresarial sobre `empresa_documentos_firmas`.
- `GET ?action=acceso&empresa_id={id}&id={firma_id}`: valida acceso a una firma documental, heredando modulo del documento asociado cuando exista `documento_gestion_id`.

## 3. Entradas obligatorias y opcionales

### 3.1 Query params comunes

- `empresa_id`: obligatorio en acciones de consulta especial.
- `id`: obligatorio para acceso puntual a firma y para versionado documental.
- `modulo`: obligatorio en `action=acceso` de documentos cuando no se suministra `id`.
- `documento_codigo`: obligatorio en `action=versiones` si no se suministra `id`.
- `permiso|accion_permiso|permission_action|action_permiso`: opcional; traduce la accion requerida.
- `include_inactive`: opcional; incluye registros con `estado != activo`.
- `include_denegados`: opcional; conserva items denegados marcados con `acceso_permitido=false`.
- `q`, `limit`, `offset`: opcionales para listado del repositorio.

### 3.2 Traduccion de accion de permiso

El backend normaliza estos valores:

- vacio, `r`, `read`, `leer`, `lectura` -> lectura
- `c`, `create`, `crear`, `creacion` -> creacion
- `u`, `update`, `editar`, `actualizar`, `modificar` -> actualizacion
- `d`, `delete`, `eliminar`, `borrar` -> eliminacion
- `a`, `approve`, `aprobar`, `aprobacion` -> aprobacion

Si no se envia accion, el comportamiento por defecto es lectura.

### 3.3 Payload minimo para versionado

`action=versionar` exige:

- `empresa_id`
- `id` del documento base

Puede sobrescribir o completar:

- `documento_codigo`
- `modulo`
- `entidad`
- `entidad_id`
- `nombre_documento`
- `tipo_documento`
- `mime_type`
- `url_archivo`
- `hash_archivo`
- `tamano_bytes`
- `usuario_creador`
- `estado`
- `observaciones`

Si un campo no llega, el backend hereda el valor del documento base.

## 4. Persistencia y columnas relevantes

### 4.1 `empresa_documentos_gestion`

Columnas funcionales relevantes:

- `empresa_id`
- `codigo` unico por empresa
- `modulo`
- `entidad`
- `entidad_id`
- `documento_codigo`
- `nombre_documento`
- `tipo_documento`
- `mime_type`
- `url_archivo`
- `hash_archivo`
- `tamano_bytes`
- `version` default `1`
- `estado_documento` default `vigente`
- `usuario_creador`
- `estado` default `activo`
- `observaciones`

Indice operativo actual:

- `ix_doc_gestion_empresa_modulo (empresa_id, modulo, entidad_id)`

### 4.2 `empresa_documentos_firmas`

Columnas funcionales relevantes:

- `empresa_id`
- `codigo` unico por empresa
- `documento_gestion_id`
- `tipo_firma` default `digital`
- `firmante_nombre`
- `firmante_documento`
- `firmante_email`
- `certificado_serial`
- `algoritmo_firma` default `SHA256`
- `hash_firma`
- `fecha_firma`
- `validez_hasta`
- `estado_firma` default `pendiente`
- `usuario_creador`
- `estado`
- `observaciones`

Indice operativo actual:

- `ix_doc_firmas_empresa_doc (empresa_id, documento_gestion_id, fecha_firma)`

## 5. Reglas de acceso y seguridad

### 5.1 Multiempresa

- Toda lectura y escritura se resuelve por `empresa_id`.
- Un `id` valido en otra empresa no debe exponer ni reutilizar datos fuera de su contexto.

### 5.2 Resolucion del modulo de permiso

El acceso no se decide solo por la ruta documental. El backend traduce `modulo` del documento a un modulo de permisos:

- `ventas` -> ventas
- `inventario`, `produccion`, `logistica`, `bodega` -> inventario
- `finanzas`, `contabilidad`, `nomina`, `rrhh`, `cartera` -> finanzas
- `clientes`, `crm`, `reserva`, `vehiculo` -> clientes
- `compras`, `proveedor` -> compras
- `facturacion`, `factur*`, `dian` -> facturacion
- vacio o no reconocido -> seguridad

### 5.3 Reglas base por rol

- `super_administrador`: acceso total.
- Sin rol en request: el evaluador documental retorna acceso permitido; esto no sustituye wrappers superiores del endpoint.
- `admin_empresa`: puede operar lectura, creacion, actualizacion, eliminacion y aprobacion en seguridad; y segun modulo en el resto.
- `contabilidad`: lectura para todos los modulos base y escritura/aprobacion en finanzas; eliminacion solo en finanzas.
- `inventario`: lectura para todos los modulos base y escritura/aprobacion en inventario.
- `compras`: lectura para todos los modulos base y escritura/aprobacion en compras.
- `cajero` y `supervisor_sucursal`: lectura general y escritura segun modulo permitido por `roleAllowsModuleAction`.

### 5.4 Listados con denegados

- `repositorio` y `versiones` filtran items sin permiso por defecto.
- Si `include_denegados=1`, los items se devuelven con `acceso_permitido=false` en vez de ocultarse.

## 6. Invariantes funcionales

- El historial de versiones se agrupa por `documento_codigo`.
- La version actual se calcula por el mayor valor numerico de `version`.
- `action=versionar` siempre crea un registro nuevo; no muta la fila base como version nueva.
- La nueva version nace con `estado_documento=vigente`.
- La version anterior intenta quedar con `estado_documento=historico`.
- Si falla el cambio de estado de la version anterior, la nueva version sigue creada y la respuesta incluye `warning`.
- La nueva observacion incorpora rastro de auditoria con fecha, id base, version origen, version destino y actor.
- El acceso a firmas hereda el modulo documental del documento asociado cuando `documento_gestion_id > 0`.
- Si una firma no tiene documento asociado, el modulo de permiso cae a `seguridad`.

## 7. Salidas y estados esperados

### 7.1 `action=acceso` en documentos

Respuesta exitosa:

- `ok=true`
- `empresa_id`
- `id` o `modulo_documento`
- `modulo_permiso`
- `accion_requerida`
- `rol`
- `acceso_permitido`
- para acceso por id: `documento_codigo`, `estado_documento`, `estado_registro`

### 7.2 `action=repositorio`

Respuesta exitosa:

- `ok=true`
- `empresa_id`
- `modulo=documentos_gestion`
- `accion_requerida`
- `total_consultados`
- `visibles`
- `denegados`
- `items[]` con columnas del registro y metadatos de acceso

### 7.3 `action=versiones`

Respuesta exitosa:

- `ok=true`
- `empresa_id`
- `documento_codigo`
- `accion_requerida`
- `version_actual`
- `total_versiones`
- `visibles`
- `denegados`
- `items[]`

### 7.4 `action=versionar`

Respuesta exitosa `201`:

- `ok=true`
- `empresa_id`
- `id_anterior`
- `id_nuevo`
- `documento_codigo`
- `version_anterior`
- `version_nueva`
- `rol`
- `modulo_permiso`
- `item_anterior`
- `item_nuevo`
- `acceso_permitido=true`
- `warning` si no fue posible marcar la version anterior como historica

### 7.5 `action=acceso` en firmas

Respuesta exitosa:

- `ok=true`
- `empresa_id`
- `id`
- `documento_id`
- `modulo_documento`
- `modulo_permiso`
- `accion_requerida`
- `rol`
- `acceso_permitido`

## 8. Errores esperados

- `400` si falta `empresa_id`, `id`, `modulo` o `documento_codigo` segun la accion.
- `404` si el documento o la firma no existen en la empresa.
- `405` si la accion especial se invoca con metodo HTTP no permitido.
- `403` en `versionar` cuando el rol no tiene permiso de actualizacion sobre el modulo documental.
- `500` si falla la consulta del repositorio, del historial o la validacion interna de acceso.
- `400` si la nueva version no puede persistirse.

## 9. Side effects y trazabilidad

- `versionar` inserta una nueva fila en `empresa_documentos_gestion`.
- `versionar` intenta actualizar la fila origen a `estado_documento=historico`.
- El historial preserva ambas filas para consultas futuras.
- No existe en este alcance envio automatico de correos, colas, webhooks ni sellado criptografico externo.

## 10. Reconciliacion operativa con contratos vecinos

- Cuando un documento gestionado respalda compras, facturacion o conciliacion fiscal, el `documento_codigo` del repositorio debe poder cruzarse con los contratos de interoperabilidad documental y con el estado transaccional vigente.
- Cuando exista firma asociada, la evidencia documental minima debe poder enlazar `documento_gestion_id`, `hash_archivo`, `hash_firma`, actor y fecha operativa.
- Si un reporte o exporte regulatorio se apoya en este repositorio, la salida debe referenciar la version vigente del documento o declarar explicitamente que el exporte es solo informativo y no evidencia firmada.
- El diagnostico operativo de incidentes documentales debe ejecutarse en conjunto con el runbook de reconciliacion documental fiscal/contable cuando el documento impacta compras, facturacion o exportes regulatorios.

## 11. Evidencia tecnica minima y endurecida

Este contrato se apoya directamente en:

- `backend/handlers/modulos_faltantes.go`
- `backend/db/modulos_faltantes.go`
- `backend/handlers/modulos_faltantes_test.go`

Prueba canonica actual:

- `TestEmpresaDocumentosGestionHandlerVersionadoYControlAcceso`

Comportamiento probado hoy:

- una nueva version incrementa `version_nueva` a `2` desde una base `1`
- el documento anterior pasa a `historico`
- `action=versiones` devuelve al menos dos versiones
- rol `contabilidad` con permiso `U` accede al documento financiero versionado
- rol `inventario` con permiso `U` no obtiene acceso al mismo documento
- `repositorio&include_denegados=1` devuelve items denegados con `acceso_permitido=false`

Evidencia minima endurecida para documentos con valor regulatorio u operativo sensible:

- conservar `documento_codigo`, `version`, `estado_documento`, `url_archivo` o referencia equivalente y `hash_archivo`
- conservar actor y fecha de versionado en `observaciones` o pista de auditoria equivalente
- si hay firma asociada, conservar `documento_gestion_id`, `tipo_firma`, `algoritmo_firma`, `hash_firma`, `firmante_*` y `fecha_firma`
- si el documento participa en exportes regulatorios, registrar dataset/exporte asociado o dejar trazable que el exporte no sustituye el repositorio documental ni la firma

## 12. ADRs y runbooks relacionados

- `documentos/gobernanza_tecnica/adr/ADR-0001-frontera-multiempresa-empresa-id.md`
- `documentos/gobernanza_tecnica/contratos/contrato_permisos_contexto_y_wrappers_api_empresa.md`
- `documentos/gobernanza_tecnica/contratos/contrato_interoperabilidad_documental_contable_y_fiscal_externa.md`
- `documentos/gobernanza_tecnica/contratos/contrato_reportes_contables_financieros_y_exportacion_multiformato.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_versionado_documental_y_firmas_externas.md`
- `documentos/gobernanza_tecnica/runbooks/runbook_reconciliacion_documental_fiscal_y_contable_externa.md`