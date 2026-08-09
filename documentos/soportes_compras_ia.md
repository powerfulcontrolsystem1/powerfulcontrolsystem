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
2. Guardar el archivo en almacenamiento privado por empresa y exponerlo solo
   mediante la descarga autenticada del modulo.
3. Calcular hash SHA-256 y detectar duplicados por archivo o documento.
4. Ejecutar extraccion con IA GPT-5.5 usando las limitaciones configuradas en Super Administrador.
5. Normalizar proveedor, NIT, tipo/numero de documento, fechas, subtotal, IVA, retenciones, total, categoria, centro de costo e impacto en inventario.
6. Marcar revision humana cuando la confianza sea baja o el modelo lo indique.
7. Aprobar o rechazar.
8. Contabilizar soporte aprobado como cuenta por pagar en `empresa_cuentas_por_pagar`.
9. Enviar a papelera un soporte no contabilizado con motivo obligatorio o
   recuperarlo conservando archivo, estado del flujo y auditoria.
10. Consultar una vista previa de retencion por dias, cantidad y bytes antes de
    cualquier politica irreversible.
11. Depurar el archivo de un soporte vencido en papelera con permiso Delete,
    motivo, retencion y confirmacion exacta del codigo. Una interrupcion queda
    como `purga_pendiente` reanudable; al terminar, la fila y los eventos
    permanecen como tumba `purgado` no recuperable.
12. Consultar el diagnostico de cuarentena por empresa: compara registros
    pendientes, archivos, bytes y antiguedad sin revelar nombres ni modificar datos.

## Estados

- `radicado`: soporte recibido.
- `extraido`: datos extraidos por IA.
- `en_revision`: requiere validacion humana.
- `aprobado`: listo para contabilizar.
- `rechazado`: no procede.
- `duplicado`: detectado por hash o documento.
- `contabilizado`: convertido en cuenta por pagar.

El estado del registro es independiente del estado funcional: `activo` permite
operar, `eliminado` lo mantiene en papelera, `purga_pendiente` permite reanudar
una depuracion interrumpida y `purgado` conserva metadatos sin archivo ni
posibilidad de recuperacion. Un registro eliminado no se puede
descargar, editar, extraer con IA, aprobar, rechazar ni contabilizar hasta ser
recuperado. La recuperacion se bloquea si ya existe otro soporte activo con el
mismo hash o numero de documento.

## Permisos por rol

- Lectura: `admin_empresa`, `supervisor_sucursal`, `cajero`, `inventario`, `compras`, `contabilidad`, `auditor`.
- Crear, extraer, aprobar y contabilizar: `admin_empresa`, `supervisor_sucursal`, `compras`, `contabilidad`.
- Papelera recuperable: usa el mismo permiso mutante del modulo, exige motivo,
  actor y `empresa_id`; no borra fisicamente el archivo ni los eventos.
- Depuracion: exige permiso `D`, motivo, antiguedad de 1 a 3650 dias y escribir
  el codigo del soporte. Nunca aplica a soportes contabilizados o convertidos.
- Diagnostico de cuarentena: usa `R` y entrega solo conteos, bytes y desalineacion
  de la empresa efectiva; no expone nombres ni rutas. El umbral operativo se
  limita a 5..1440 minutos y usa 15 minutos por defecto.

## Consideraciones de produccion

- Requiere IA activada en configuracion avanzada super y modelo `openai:gpt-5.5` disponible.
- Las credenciales del proveedor IA deben venir de configuracion segura o entorno.
- La carga valida firma real contra extension, normaliza el MIME y rechaza XML
  con DTD, entidades, instrucciones o elementos activos; esto endurece la
  admision. Si `PCS_SUPPORTS_CLAMAV_ADDR` está configurado, además envía el
  contenido a `clamd` mediante INSTREAM; `PCS_SUPPORTS_CLAMAV_REQUIRED=1`
  bloquea la carga cuando el servicio no responde.
- La red privada de monitoreo recibe contadores agregados por resultado mediante
  `pcs_support_antivirus_scans_total` y los gauges
  `pcs_support_antivirus_required`/`pcs_support_antivirus_configured`. No se usan
  etiquetas de empresa, usuario, soporte, archivo o ruta; las alertas conservan
  el modo obligatorio fail-closed y conducen al runbook de observabilidad.
- La extracción IA admite como máximo 128 KiB y únicamente el esquema
  documentado. Rechaza tipos compuestos en escalares, claves inesperadas,
  valores no finitos/negativos, confianza fuera de 0..1, textos excesivos y
  fechas inválidas. Los resultados operativos se exponen agregados en
  `pcs_support_ai_extractions_total` sin datos del tenant o documento.
- Antes de descargar o adjuntar a IA se verifica el SHA-256 almacenado y el
  límite privado de 15 MiB. La descarga usa MIME canónico o binario seguro,
  attachment y cabeceras sandbox/nosniff/same-origin/no-referrer/DENY/no-store.
  Un legado sin hash se descarga como binario y revalida firma antes de IA. Los
  fallos incrementan únicamente una métrica agregada de integridad.
- Cada incidente se bloquea con `FOR UPDATE` por empresa. Una aprobación abierta
  vuelve a `en_revision`, se limpian actor/fecha de aprobación y el evento mínimo
  se confirma en la misma transacción. Un estado contable terminal se preserva.
- Si la fila no puede persistirse despues de escribir el archivo privado, el
  adjunto recien creado se elimina dentro de la raiz segura para evitar huerfanos.
- `Cancelar IA` aborta la solicitud HTTP y propaga el contexto al proveedor; el
  soporte conserva su estado anterior si no alcanzó a guardar una extracción y
  devuelve la reserva diaria avanzada con contadores acotados a cero.
- Los soportes con baja confianza, valores inconsistentes o datos tributarios incompletos deben revisarse antes de aprobar.
- Para documentos oficiales DIAN, la extraccion IA no reemplaza validacion tributaria ni aceptacion/rechazo legal del documento.
- La pantalla valida el enlace de archivo antes de renderizarlo y solo permite direcciones navegables seguras, evitando protocolos no esperados en soportes cargados.
- Un soporte convertido a CxP no puede enviarse a papelera, porque su origen
  documental debe permanecer visible para conciliacion y auditoria contable.
- La vista previa de retencion es solo lectura. La accion separada de depuracion
  mueve primero el archivo a cuarentena, inicia `purga_pendiente` con `FOR UPDATE`,
  elimina el archivo y finaliza `purgado` en una segunda transaccion. Un reintento
  recupera las caidas antes de iniciar, antes de borrar o antes de finalizar; una
  cuarentena ambigua falla cerrada. Un advisory lock PostgreSQL por empresa
  serializa replicas y los replays completos no duplican efectos. Su ejecucion real sigue prohibida hasta
  certificar backup/restore y el flujo en staging.
- Las cargas JSON/manuales no aceptan nombre, MIME, hash ni `private://` enviados
  por el cliente. Descarga, IA, retencion y depuracion exigen que la ruta privada
  pertenezca al `empresa_id` efectivo.

## Pruebas

- `go test ./db -run Test.*Soporte.*IA -count=1`.
- `go test ./... -count=1`.
- `git diff --check`.
- QA 2026-05-06: pagina y dashboard validados con HTTP 200 en Motel Calipso (`empresa_id=7`); ver `documentos/reporte_qa_modulos_2026-05-06.md`.
