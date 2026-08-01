# P109-002 - Preflight autenticado de IA y CxP/IA

Fecha: 2026-08-01
Entorno: staging, PCS (`empresa_id=12`)
Candidato: `6da6c13453a40b2d84e23285fa83f255f34788da`

Sin revelar valores privados, el runtime confirmó presencia de
`CONFIG_ENC_KEY`, `CONFIG_ENC_KEY_ID` y `OPENAI_API_KEY`. La lectura autenticada
aprobó:

- catálogo IA: HTTP 200, cinco modelos y preferencia con alcance por usuario;
- preferencia e historial: HTTP 200;
- configuración OpenAI propia: HTTP 200, desactivada y sin devolver `api_key`;
- modo inicial `operativo`, agente `general` y streaming habilitado;
- dashboard CxP/IA: HTTP 200.

El upload CxP/IA rechazó una extensión no permitida con HTTP 400. La cantidad
de soportes permaneció `1 -> 1`, demostrando fallo antes de guardar archivo o
fila.

No se llamó al proveedor externo ni se generó costo en este ciclo. Continúan
pendientes la extracción real autorizada, edición/confirmación/cancelación,
doble clic, degradación del proveedor, evals y aislamiento A/B.

## Segunda pasada sobre el candidato `ea9642dd`

El candidato inmutable fue promovido por digest solo a staging. Centro IA
aprobó carga autenticada, diagnóstico real, consola sin errores y revocación y
restauración inmediata del permiso explícito de PCS. Producción permaneció sin
cambios.

La extracción real del soporte `SCI-0001` mantuvo el documento en `Radicado` y
respondió con degradación pública segura, pero el proveedor devolvió HTTP 400.
La prueba aislada desde el mismo contenedor confirmó HTTP 200 tanto para texto
como para un `input_file` construido como URL de datos. La causa quedó acotada
al contrato local: el cliente enviaba únicamente los bytes Base64, sin el
prefijo `data:<mime>;base64,` requerido para `file_data`.

Se corrigió localmente la construcción de adjuntos, normalizando el tipo MIME y
neutralizando valores inválidos. Las pruebas enfocadas, el paquete completo de
handlers, `db`, `secure` y `go vet ./handlers` aprobaron.

Para validar sin crear PR ni alterar producción se construyó un overlay temporal
que sustituyó únicamente el binario del API sobre el digest inmutable. Trivy
reportó cero HIGH/CRITICAL. Con ese overlay, el botón oficial `Extraer IA`
procesó `SCI-0001`: pasó a `Extraido`, leyó proveedor, NIT, documento, fechas,
subtotal 100, IVA 19, total 119 y confianza 99 %. `convertido_id` permaneció en
cero; no se creó cuenta por pagar, pago ni contabilización automática.

La revisión humana mostró todos los valores editables. Se guardó y comprobó una
marca QA en observaciones y después se restauró el texto original; auditoría
registró dos ediciones y una sola transición `Radicado -> Extraido`.

Las dos credenciales heredadas se volvieron a cifrar desde el secreto de entorno
del propio contenedor con la llave activa, sin imprimir ni transportar su valor.
Al terminar se restauró staging al digest `ea9642dd` y se eliminaron las imágenes
y binarios temporales. Salud y readiness de staging y producción permanecieron
en HTTP 200; producción no se modificó.

Estado: **P109-002 parcial**. Extracción y edición E2E quedan demostradas sobre
el overlay controlado; faltan publicar la corrección en un candidato inmutable y
cerrar confirmación/cancelación, doble clic, duplicados y aislamiento A/B.

## Tercera pasada funcional sobre el digest restaurado

Se recorrieron las confirmaciones de `Aprobar`, `Rechazar` y `Contabilizar`:
cancelarlas conservó `SCI-0001` en `Extraido`. Un doble clic real sobre
`Extraer IA` generó una sola petición mientras todos los botones permanecieron
deshabilitados; la degradación conocida del digest fue segura y no cambió datos.

Después se vinculó el proveedor canónico por el formulario editable. La
aprobación real dejó `SCI-0001` en `aprobado`, con `convertido_id=0`; el rechazo
posterior lo dejó en `rechazado`, también sin CxP, pago o asiento. La auditoría
registró exactamente una transición de aprobación y una de rechazo.

Para probar la conversión se creó el soporte controlado `SCI-0003`, se vinculó
el proveedor, se aprobó y se contabilizó mediante los botones oficiales. Se
creó exactamente una fila activa en la fuente canónica
`empresa_cuentas_por_pagar`, ID 16, código `CXP-SCI-0003`, saldo 214200 y estado
pendiente. Repetir `Contabilizar` conservó una sola fila y un solo evento.

Una segunda carga demo con el mismo documento `FE-1024` produjo `SCI-0004` en
estado `duplicado`, referenció al soporte 3, mantuvo `convertido_id=0` y no creó
otra CxP. Esto demuestra revisión humana, confirmaciones, cancelación, doble
clic, conversión idempotente y duplicado por documento dentro de PCS.

Estado: **P109-002 parcial**. Quedan publicación inmutable del arreglo de
`file_data`, recibo real y aislamiento A/B con una identidad empresarial no
global; esos huecos impiden aprobar la fase completa.

## Cuarta pasada sobre el candidato inmutable `4ab318c`

El workflow `30720230306` construyó y escaneó las cuatro imágenes exactas sin
PR. El candidato se promovió solo a staging y conservó producción intacta. La
corrección de `file_data` ya forma parte del digest desplegado.

Se radicó el soporte controlado `SCI-0005` por el endpoint público oficial y se
reintentó la extracción desde el botón visible. El primer intento agotó el
límite diario avanzado. El límite global de staging se amplió temporalmente de
5 a 10 mediante el endpoint administrativo oficial y se restauró inmediatamente
a 5 después del diagnóstico.

El segundo intento alcanzó OpenAI, descartando configuración, cifrado y cuota,
pero devolvió HTTP 400. La causa fue distinta a la ya corregida: el soporte era
XML y el API de Responses no acepta ese contenido como `input_file`. El contrato
local sí admite XML. La corrección pendiente de publicación convierte MIME
textuales cerrados (`text/plain`, CSV, XML y JSON UTF-8) en `input_text`
delimitado como contenido no confiable; imágenes conservan `input_image` y PDF/
Office conservan `input_file`. La prueba enfocada y `go vet ./handlers`
aprobaron.

Estado: **P109-002 parcial**. El digest desplegado corrige el formato Base64,
pero falta publicar y comprobar la degradación XML a texto, completar recibo y
aislamiento A/B.
