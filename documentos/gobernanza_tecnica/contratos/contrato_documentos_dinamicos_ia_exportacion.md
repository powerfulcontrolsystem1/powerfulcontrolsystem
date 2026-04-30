# Contrato tecnico: documentos dinamicos con IA y exportacion multiformato

Fecha: 2026-04-30
Estado: vigente

## Alcance

Este contrato cubre la generacion de documentos dinamicos a partir de contenido o prompt IA, renderizado con templates Go/HTML y descarga en formatos PDF, DOCX, XLSX, HTML, TXT y JSON.

## Rutas implicadas

- `POST /generate`
- `GET /download?id=<document_id>&type=pdf|docx|xlsx|html|txt|json`
- `POST /api/empresa/chat_documentos/exportar`

## Entradas principales

`POST /generate` acepta:

- `empresa_id`: contexto empresarial cuando aplica.
- `title`: titulo del documento.
- `prompt`: instruccion para generar contenido con IA cuando no se envia `content`.
- `content`: contenido base cuando no se requiere generacion IA.
- `input_format`: `html`, `markdown` o texto plano.
- `template_name`: plantilla logica.
- `model_id`: modelo IA solicitado; por defecto `openai:gpt-5.4-mini`.
- `variables`: mapa de variables para render con `html/template`.
- `formats`: formatos de descarga a preparar o exponer.
- `metadata`: datos auxiliares no sensibles.

`POST /api/empresa/chat_documentos/exportar` acepta:

- `empresa_id`: contexto empresarial obligatorio.
- `content`: respuesta o conversacion generada desde el chat IA.
- `format`: `pdf`, `docx`, `xlsx`, `txt` o `json`.
- `document_type`: `contrato`, `factura`, `reporte`, `acta`, `cotizacion`, `tabla` o `documento`.
- `source_module`: modulo de origen (`chat_ia`, `reportes`, `tareas`, `agenda`, etc.).
- `metadata`: datos auxiliares no sensibles del origen chat.

## Salidas

`POST /generate` devuelve:

- `document_id`
- `title`
- `formats`
- `download_urls`
- estado de generacion

`POST /api/empresa/chat_documentos/exportar` devuelve:

- `document_id`
- `title`
- `format`
- `filename`
- `download_url`
- advertencia y `fallback_format` cuando un conversor falla y se entrega respaldo TXT.

`GET /download` devuelve el archivo con:

- `Content-Type` correcto para el formato.
- `Content-Disposition: attachment` con nombre de archivo seguro.

## Invariantes

1. La IA puede generar contenido, pero el backend controla templates, formatos y rutas de descarga.
2. No se entrega SQL libre ni credenciales al modelo.
3. Las variables deben renderizarse con APIs de template, no con concatenacion insegura.
4. Los documentos temporales deben tener identificador no predecible y ruta de almacenamiento controlada.
5. Si falla IA, wkhtmltopdf o una libreria de exportacion, el servidor debe responder error controlado y no romper el runtime principal.
6. La exportacion desde chat debe redactar patrones de secretos antes de guardar contenido temporal.
7. La exportacion desde chat debe registrar auditoria con usuario, empresa, formato, nombre de archivo, fecha, origen `chat_ia`, exito o error.
8. Para Excel, si el contenido incluye tabla Markdown, debe preservar encabezados, filas y columnas numericas como celdas.
9. Si el flujo se vuelve historico o regulatorio, debe agregarse tabla persistente con `empresa_id`, usuario, formato, estado, hash/ruta y auditoria.

## Side effects

- posible llamada al proveedor IA configurado.
- escritura temporal de HTML y archivos exportables.
- descarga directa del archivo solicitado.

## Evidencia tecnica minima

- prueba de generacion con `content` sin IA.
- prueba de generacion con `prompt` usando modelo configurado o mock.
- prueba de descarga por cada formato soportado.
- validacion de headers `Content-Type` y `Content-Disposition`.
- prueba de error controlado cuando se solicita formato no soportado.

## Documentos relacionados

- `documentos/descripcion_del_proyecto`
- `documentos/diagramas/estructura_del_codigo.md`
- `documentos/estructura_bd.md`
- `documentos/gobernanza_tecnica/contratos/contrato_reportes_contables_financieros_y_exportacion_multiformato.md`
