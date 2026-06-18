# Captura inteligente de compras y gastos con IA GPT-5.5

## Alcance

Modulo empresarial para radicar soportes de compras y gastos por `empresa_id` desde foto, PDF o XML. Usa la capa IA existente del sistema con el modelo recomendado `openai:gpt-5.5` para extraer datos contables, tributarios y operativos, dejando auditoria de eventos y registro de consumo IA.

## Superficies

- Pagina: `/administrar_empresa/soportes_compras_ia.html`.
- Menu: `Administrar empresa > Compras > Captura IA GPT-5.5`.
- API: `/api/empresa/soportes_compras_ia`.
- Wrapper: `WithEmpresaSoportesComprasIAPermissions`.
- Modulo de permiso/licencia: `soportes_compras_ia`.

## Flujo funcional

1. Radicar soporte con archivo o datos manuales.
2. Guardar archivo bajo `/uploads/soportes_compras_ia/empresa_<id>/`.
3. Calcular hash SHA-256 y detectar duplicados por archivo o documento.
4. Ejecutar extraccion con IA GPT-5.5 usando las limitaciones configuradas en Super Administrador.
5. Normalizar proveedor, NIT, tipo/numero de documento, fechas, subtotal, IVA, retenciones, total, categoria, centro de costo e impacto en inventario.
6. Marcar revision humana cuando la confianza sea baja o el modelo lo indique.
7. Aprobar o rechazar.
8. Contabilizar soporte aprobado como cuenta por pagar en `empresa_cuentas_por_pagar`.

## Estados

- `radicado`: soporte recibido.
- `extraido`: datos extraidos por IA.
- `en_revision`: requiere validacion humana.
- `aprobado`: listo para contabilizar.
- `rechazado`: no procede.
- `duplicado`: detectado por hash o documento.
- `contabilizado`: convertido en cuenta por pagar.

## Permisos por rol

- Lectura: `admin_empresa`, `supervisor_sucursal`, `cajero`, `inventario`, `compras`, `contabilidad`, `auditor`.
- Crear, extraer, aprobar y contabilizar: `admin_empresa`, `supervisor_sucursal`, `compras`, `contabilidad`.
- Eliminacion funcional: no habilitada; se usa rechazo/anulacion trazable.

## Consideraciones de produccion

- Requiere IA activada en configuracion avanzada super y modelo `openai:gpt-5.5` disponible.
- Las credenciales del proveedor IA deben venir de configuracion segura o entorno.
- Los soportes con baja confianza, valores inconsistentes o datos tributarios incompletos deben revisarse antes de aprobar.
- Para documentos oficiales DIAN, la extraccion IA no reemplaza validacion tributaria ni aceptacion/rechazo legal del documento.
- La pantalla valida el enlace de archivo antes de renderizarlo y solo permite direcciones navegables seguras, evitando protocolos no esperados en soportes cargados.

## Pruebas

- `go test ./db -run Test.*Soporte.*IA -count=1`.
- `go test ./... -count=1`.
- `git diff --check`.
- QA 2026-05-06: pagina y dashboard validados con HTTP 200 en Motel Calipso (`empresa_id=7`); ver `documentos/reporte_qa_modulos_2026-05-06.md`.
