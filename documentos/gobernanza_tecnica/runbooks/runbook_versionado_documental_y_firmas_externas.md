# Runbook: versionado documental y firmas externas

Fecha: 2026-04-18
Estado: vigente

## 1. Sintoma

Usar este runbook cuando ocurra alguno de estos sintomas:

- un documento no aparece en el repositorio documental para un rol esperado
- una version nueva se crea, pero la anterior no queda historica
- el historial de versiones devuelve menos registros de los esperados
- una firma documental queda visible o invisible de forma inconsistente
- una consulta `acceso` responde denegado para un rol que el negocio esperaba habilitar

## 2. Alcance del incidente

Aplica al backend empresarial:

- `/api/empresa/documentos/gestion`
- `/api/empresa/documentos/firmas`
- tablas `empresa_documentos_gestion` y `empresa_documentos_firmas`
- mapeo `modulo documental -> modulo de permisos`

No cubre recuperacion de binarios fisicos, revocacion criptografica real de certificados ni integraciones notariales externas.

## 3. Fuentes de evidencia

- request exacta usada por frontend, QA o soporte
- `empresa_id`, `id`, `documento_codigo`, `documento_gestion_id`
- rol enviado en header administrativo
- filas actuales en `empresa_documentos_gestion`
- filas actuales en `empresa_documentos_firmas`
- salida de la prueba `TestEmpresaDocumentosGestionHandlerVersionadoYControlAcceso`

## 4. Verificaciones iniciales

1. Confirmar `empresa_id` correcto del incidente.
2. Confirmar si el caso es de lectura, creacion, actualizacion, eliminacion o aprobacion documental.
3. Confirmar el rol real enviado en request y no solo el rol esperado por negocio.
4. Consultar el documento puntual con `action=acceso` y luego el listado con `action=repositorio`.
5. Si existe firma, consultar `action=acceso` sobre la firma y validar `documento_gestion_id`.
6. Si el documento soporta un exporte regulatorio o reporte formal, identificar dataset, formato y referencia usada por operación o auditoría.

Ejemplos de verificacion manual:

```text
GET /api/empresa/documentos/gestion?action=acceso&empresa_id=522&id=17&permiso=U
GET /api/empresa/documentos/gestion?action=repositorio&empresa_id=522&permiso=U&include_denegados=1
GET /api/empresa/documentos/gestion?action=versiones&empresa_id=522&id=17&permiso=R&include_denegados=1
GET /api/empresa/documentos/firmas?action=acceso&empresa_id=522&id=9&permiso=R
```

## 5. Diagnostico por escenario

### 5.1 Documento no visible en repositorio

Validar:

- si el item existe con `include_inactive=1`
- si el item existe con `include_denegados=1`
- `modulo` del documento
- `modulo_permiso` resuelto por el backend
- accion requerida traducida desde `permiso`

Causa frecuente:

- el documento quedó ligado a un `modulo` que mapea a otro frente de permisos distinto del esperado.

### 5.2 Version nueva creada pero base no historica

Validar:

- respuesta de `versionar` por presencia de `warning`
- `item_nuevo.id`
- `item_anterior.estado_documento`
- auditoria en `observaciones` de ambas filas

Interpretacion:

- si existe `id_nuevo` y `warning`, la insercion principal fue exitosa y falló solo la marca historica del documento base.

### 5.3 Historial incompleto

Validar:

- `documento_codigo` real en todas las filas relacionadas
- `estado` del registro si el listado no usa `include_inactive=1`
- `permiso` solicitado y `include_denegados`
- valor numerico de `version`

Causa frecuente:

- versiones creadas con `documento_codigo` distinto o filtro por acceso que oculta parte del historial.

### 5.4 Firma con acceso inconsistente

Validar:

- `documento_gestion_id`
- existencia real del documento asociado
- `modulo` del documento asociado
- rol enviado por request

Interpretacion:

- si la firma tiene `documento_gestion_id > 0`, su acceso hereda el modulo del documento.
- si no tiene documento asociado, cae a permisos del modulo `seguridad`.

### 5.5 Exporte regulatorio tratado como evidencia única

Validar:

- dataset y formato del exporte utilizado
- si existe `documento_codigo` conciliable con el repositorio
- si existe firma documental o solo exporte tabular/presentacional

Interpretacion:

- un exporte puede resumir evidencia, pero no reemplaza la versión documental vigente ni la firma asociada cuando el flujo exige respaldo reforzado.

## 6. Consultas de apoyo recomendadas

Repositorio documental:

```sql
SELECT id, empresa_id, codigo, modulo, entidad, entidad_id, documento_codigo,
       version, estado_documento, estado, url_archivo, hash_archivo,
       usuario_creador, observaciones
FROM empresa_documentos_gestion
WHERE empresa_id = $1
  AND UPPER(COALESCE(documento_codigo, '')) = UPPER($2)
ORDER BY CAST(COALESCE(NULLIF(version, ''), '0') AS INTEGER) DESC, id DESC;
```

Firmas por documento:

```sql
SELECT id, empresa_id, codigo, documento_gestion_id, tipo_firma, firmante_nombre,
       firmante_email, certificado_serial, algoritmo_firma, hash_firma,
       fecha_firma, validez_hasta, estado_firma, estado
FROM empresa_documentos_firmas
WHERE empresa_id = $1
  AND documento_gestion_id = $2
ORDER BY datetime(COALESCE(NULLIF(fecha_firma, ''), fecha_creacion)) DESC, id DESC;
```

## 7. Acciones de recuperacion

### 7.1 Rol sin acceso por modulo documental mal clasificado

1. Confirmar si el `modulo` persistido representa el frente correcto del negocio.
2. Si el problema es de dato, corregir `modulo` en el documento afectado bajo ventana controlada y con trazabilidad.
3. Si el problema es de mapeo backend, ajustar `mapDocumentoModuloToPermissionModule` y cubrirlo con prueba.

### 7.2 Version anterior no historica

1. Confirmar que la nueva fila quedó creada y es la vigente.
2. Revisar `observaciones` y `estado_documento` de la fila origen.
3. Si no hubo actualización de la base anterior, corregirla manualmente solo sobre el `id` afectado y registrar auditoría operativa.
4. Reconsultar `action=versiones` para confirmar consistencia.

### 7.3 Historial partido por `documento_codigo`

1. Identificar las filas del mismo flujo con codigos divergentes.
2. Definir el `documento_codigo` canonico junto con negocio o QA.
3. Corregir solo las filas afectadas y revalidar el historial completo.

### 7.4 Firma huérfana

1. Confirmar si la firma debía apuntar a un documento gestionado.
2. Si la relacion faltó por error operativo, completar `documento_gestion_id` con respaldo de trazabilidad.
3. Si la firma debe permanecer independiente, aceptar que el permiso cae al modulo `seguridad`.

### 7.5 Exporte sin reconciliacion con documento o firma

1. Confirmar si el exporte es informativo o regulatorio.
2. Si es regulatorio, exigir enlace trazable con `documento_codigo`, versión vigente y firma si aplica.
3. Si no existe ese enlace, reclasificar el exporte como salida informativa hasta completar la evidencia.

## 8. Validacion posterior

Validar como minimo:

1. `action=acceso` sobre el documento afectado con el rol involucrado.
2. `action=repositorio` con y sin `include_denegados=1`.
3. `action=versiones` por `id` o `documento_codigo`.
4. `action=acceso` de la firma si aplica.
5. prueba de handler focalizada.
6. si hubo exporte regulatorio, validar consistencia del dataset y dejar claro que no quedó desacoplado del repositorio documental.

Comando recomendado:

```text
go test ./handlers -run '^TestEmpresaDocumentosGestionHandlerVersionadoYControlAcceso$' -count=1
```

## 9. Escalamiento

- Escalar a backend cuando el problema esté en mapeo de permisos, versionado o consistencia entre filas.
- Escalar a QA cuando el incidente no tenga caso reproducible o falte request exacta.
- Escalar a operacion si el incidente combina repositorio documental con exportaciones, conciliacion fiscal o integraciones externas.

## 10. Contratos y ADRs relacionados

- `documentos/gobernanza_tecnica/contratos/contrato_repositorio_documental_y_firmas_externas.md`
- `documentos/gobernanza_tecnica/contratos/contrato_permisos_contexto_y_wrappers_api_empresa.md`
- `documentos/gobernanza_tecnica/contratos/contrato_interoperabilidad_documental_contable_y_fiscal_externa.md`
- `documentos/gobernanza_tecnica/contratos/contrato_reportes_contables_financieros_y_exportacion_multiformato.md`
- `documentos/gobernanza_tecnica/adr/ADR-0001-frontera-multiempresa-empresa-id.md`