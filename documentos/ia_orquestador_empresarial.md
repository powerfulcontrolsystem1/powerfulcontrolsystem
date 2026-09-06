# Orquestador IA empresarial

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se reconciliaron modelo global, identidad permanente y flags separados de hotel/catálogo; ninguna respuesta del modelo concede autoridad.
- El registro de fuentes y memoria con consentimiento no implementa RAG; las categorías y la minimización de fase 2 permanecen aplicables.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Estado de activacion

Ampliación local del 2026-09-06: [capacidades y evidencia](chat_ia_capacidades_2026-09-06.md).
El chat integra búsqueda de catálogo, búsqueda de estaciones, propuesta de
consumo y seis familias de reportes. Ventas requiere `AI_SALES_TOOLS_ENABLED`.
La nueva escritura no se considera habilitada ni validada en PostgreSQL real.

El orquestador exige `AI_ENTERPRISE_ORCHESTRATOR_ENABLED=true`. Las escrituras
requieren además `AI_WRITE_TOOLS_ENABLED=true` y el flag específico:
`AI_HOTEL_TOOLS_ENABLED` para hotel o `AI_CATALOG_TOOLS_ENABLED` para productos.
La ejecución en modo agente exige también `AI_AGENT_MODE_ENABLED=true`.
La identidad de interfaz `agente_pcs` no activa estos permisos operativos.
RAG semántico sigue siendo una ampliación pendiente; un catálogo de fuentes no
implementa recuperación por fragmento. Los flags no sustituyen autorización,
confirmación ni evaluación de una herramienta.

## Contrato de seguridad

### Modelos, adjuntos y preferencias

- Super Administrador define un modelo global; la API empresarial solo expone
  el seleccionado. Las preferencias del navegador no conceden otra selección,
  permisos ni acceso a empresas. El catálogo técnico permite cambiar la
  configuración, sujeto a disponibilidad real del proveedor.
- Se aceptan imagenes, PDF, TXT, CSV, DOCX y XLSX de hasta 8 MB. El nombre del
  archivo no se utiliza como ruta ni se publica desde este flujo.

- El contexto se deriva del wrapper de permisos: usuario autenticado, empresa,
  rol efectivo, request y conversacion. El `empresa_id` del cliente solo pasa
  si coincide con el contexto validado.
- El catalogo es cerrado y reside en `backend/ai`. Un modelo no puede elegir
  endpoint, SQL, empresa ni campos fuera del esquema de la herramienta.
- Una propuesta contiene hash del plan, usuario, empresa, conversacion,
  vencimiento de quince minutos, estado y politica de rollback.
- Confirmar es un POST separado con `proposal_id`, `plan_hash` e
  `idempotency_key`. La propuesta se bloquea y se consume de forma atomica.
- Una propuesta ajena, vencida, alterada, cancelada o usada no se ejecuta.
- La auditoria registra identificadores operativos y categorias, no prompts
  completos, secretos, tokens, contrasenas ni valores privados innecesarios.
- El drawer no ejecuta bloques `PCS_ACTION` ni endpoints propuestos por el
  modelo. La unica escritura habilitable usa una tarjeta generada desde el
  contrato del servidor, con `proposal_id`, hash del plan e idempotency key.
- El estado de conversacion se guarda en PostgreSQL por empresa y usuario con
  expiracion; no depende de historial libre del navegador ni acepta una
  conversacion de otro usuario o empresa.
- Las entradas JSON de herramientas son estrictas: se rechazan campos
  desconocidos, cuerpos concatenados y tamanos superiores al contrato.
- Las entradas procedentes de documentos, adjuntos e integraciones se tratan
  como datos no confiables. Las senales de inyeccion no otorgan capacidades y
  las herramientas siguen siendo exclusivamente de propiedad del servidor.
- Solo se permite enviar al proveedor campos incluidos expresamente en una
  lista blanca. Credenciales, datos personales, bancarios, fiscales completos
  y secretos tecnicos se descartan antes de formar cualquier contexto.
- Las ejecuciones guardan metadatos minimizados por empresa y usuario:
  herramienta, riesgo, estado, duracion, categorias de datos y fuentes. No
  conservan prompts completos ni resultados privados.

## Herramientas implementadas

`hotel.inspect_room_station` es una consulta de bajo riesgo que devuelve la
configuracion actual de una estacion hotelera dentro de la empresa validada.
Registra sus fuentes sin incluir tarifas o datos privados en la auditoria.

`hotel.configure_room_station` configura una estacion existente como habitacion
hotelera y registra tarifas diarias por ocupacion. Exige estacion, nombre,
moneda, check-in/check-out y tarifas sin ocupaciones duplicadas. La ejecucion
actualiza configuracion de estaciones y tarifas dentro de una misma transaccion;
si una operacion falla no se confirma ninguna parte.

La herramienta no aplica cambios masivos, no elimina tarifas existentes y no
emite documentos fiscales. La tarjeta del chat incorpora el formulario asistido, estado actual,
cambio propuesto, fuentes y botones Confirmar/Cancelar. Aun asi los flags
de escritura siguen apagados hasta una prueba controlada por empresa.

## Flujo visible

Cuando la consulta identifica una configuracion de habitacion, el chat muestra
un formulario para revisar estacion, tarifas por ocupacion, moneda, horarios,
activacion y conservacion de configuracion. Los valores ausentes se dejan como
campos obligatorios, por lo que el usuario debe completarlos. Al preparar el
plan, el backend consulta la estacion real dentro de la empresa actual y crea
la propuesta temporal. El boton Confirmar realiza otra peticion independiente;
revalida usuario, empresa, licencia, permisos, hash, vencimiento y uso unico.

El modo agente permanece bloqueado salvo `AI_AGENT_MODE_ENABLED=true` y un
contexto acotado por servidor. El chat usa la identidad permanente `agente_pcs`; la ejecución de herramientas
sigue acotada por servidor y por los flags anteriores.

`catalog.search_products` es una consulta de bajo riesgo que devuelve un
catalogo acotado de productos, categorias y bodegas exclusivamente de la
empresa validada. Sirve para desambiguar referencias antes de preparar una
accion y no expone datos de otra empresa.

`catalog.create_product` prepara la creacion de un producto con precio,
impuesto, categoria, bodega y stock inicial. Antes de crear la propuesta valida
el plan, consulta duplicados por nombre/SKU dentro de la empresa y obliga a una
confirmacion separada. Al confirmar reutiliza `CreateProducto`, que valida las
relaciones empresariales y registra producto, inventario inicial e historial de
precio en una transaccion.

Cuando el usuario solicita crear un producto en lenguaje natural, el chat abre
una tarjeta de propuesta con los campos extraidos como borrador. El navegador
no decide herramientas ni ejecuta cambios: el backend valida el plan, revisa
duplicados y solo crea el producto despues de la confirmacion independiente.
Los campos de categoria y bodega siguen siendo opcionales, salvo que se
registre stock inicial; en ese caso la bodega debe pertenecer a la empresa.

Las escrituras se habilitan de forma granular y permanecen apagadas por
defecto: `AI_ENTERPRISE_ORCHESTRATOR_ENABLED=true`,
`AI_WRITE_TOOLS_ENABLED=true` y el flag especifico de la herramienta
(`AI_HOTEL_TOOLS_ENABLED=true` o `AI_CATALOG_TOOLS_ENABLED=true`). Un flag no
omite permisos ni confirmaciones.

## Como agregar una herramienta

1. Definir una entrada de riesgo, permisos, modulo, limite y rollback en
   `backend/ai/enterprise.go`; no aceptar endpoints, SQL ni nombres de tabla
   desde el modelo.
2. Crear un plan JSON tipado y normalizador en `backend/db/ai_enterprise.go`.
   El plan no contiene `empresa_id`, usuario, rol ni valores de autoridad.
3. Reutilizar un servicio de dominio existente que filtre por `empresa_id` y
   use transaccion cuando toque mas de un registro.
4. Registrar propuesta temporal, estado previo/esperado minimizado,
   idempotencia, confirmacion y auditoria antes de exponer el boton.
5. Revalidar herramienta, permisos, licencia, empresa, vencimiento y hash al
   confirmar. Añadir pruebas de tenant, permisos, parametros, duplicado,
   doble confirmacion y fallo parcial.

No se debe registrar una herramienta como disponible solo porque exista una
pagina de interfaz. Los modulos restantes se conectaran uno por uno bajo este
contrato, con pruebas y habilitacion gradual por empresa.

## Plan de ampliacion

1. Lectura con fuentes: servicios de dominio y consultas parametrizadas con
   lista blanca, limites y filtros empresariales.
2. UI de propuestas: plan, estado anterior/esperado, riesgo, fuentes,
   confirmar, cancelar y resultado.
3. Herramientas de bajo riesgo por modulo, una a una, con pruebas de permisos,
   idempotencia y aislamiento.
4. RAG documental con PostgreSQL/pgvector solo despues de validar la extension,
   permisos por fragmento, fuentes y reindexacion. El catalogo de fuentes y la
   memoria con consentimiento ya tienen esquema de aislamiento; aun no existe
   recuperacion semantica ni se enviara contenido documental al proveedor.
5. Modo agente con presupuesto, alcance, duracion, cantidad maxima de
   operaciones, circuit breaker y confirmaciones no omitibles.

## Adjuntos y privacidad del proveedor

Los adjuntos del chat se validan por extension cerrada y por contenido, no por
el `Content-Type` declarado por el navegador. Imagenes requieren firmas
conocidas; PDF debe iniciar con `%PDF-`; TXT/CSV deben ser UTF-8 sin bytes NUL;
DOCX/XLSX deben ser contenedores OpenXML con sus entradas esperadas. Se descartan
HTML, SVG activo, ejecutables, ZIP genericos y archivos con extension fingida.
El MIME que se remite al proveedor se reconstruye desde la validacion local.

Los cuerpos de error de proveedores no se devuelven al navegador ni se
propagan mediante `Error()`. La interfaz recibe un mensaje generico, mientras
la auditoria conserva solo metadatos minimizados sin prompt, token, adjunto ni
respuesta privada.

Cada solicitud de OpenAI Responses incluye `safety_identifier` estable y
pseudonimo, derivado con SHA-256 de una identidad autenticada o del alcance
operativo cuando no existe usuario final. Nunca se transmite correo, IP ni otro
identificador original; las solicitudes conservan `store=false`.
El cliente interno exige este valor en su firma y valida el formato `pcs-`
seguido por 32 caracteres hexadecimales antes de abrir la conexion. Una llamada
sin seudonimo o con identidad cruda falla cerrada y no llega al proveedor.

## Modelos y esfuerzo de razonamiento

Super Administrador selecciona un solo modelo principal para todas las
empresas, usuarios, preguntas y adjuntos. El código configura inicialmente `openai:gpt-5.6-terra` con
esfuerzo `medium` y cliente Responses. Esto describe la configuración local,
no acredita disponibilidad ni capacidades contratadas con el proveedor. La disponibilidad efectiva
depende de la cuenta y permisos vigentes del proveedor.

El catálogo técnico conserva modelos alternativos para que Super Administrador
pueda cambiar la política global, pero la API empresarial expone únicamente el
seleccionado. El navegador no puede escoger otro modelo ni esfuerzo. Voz sigue
siendo una capa de entrada/salida separada; el modelo principal devuelve texto.

## Agente PCS permanente

La experiencia empresarial no muestra selector ni interruptor de agente. El
servidor fuerza `agente_pcs` para chat, adjuntos, Centro IA y Pedidos IA. El
agente identifica el módulo según la intención y el contexto autorizado; no
recibe `empresa_id`, rol, permiso ni confirmación desde el modelo como campos de
autoridad.

La contabilidad local registra tokens y herramientas usadas; la facturación
externa debe reconciliarse con el proveedor y no se deduce de un flag. El historial enviado queda
acotado y el razonamiento inicial es `medium`. El chat general permite hasta
cuatro llamadas secuenciales y una propuesta de escritura por mensaje;
la cuota ligera de agente se reserva al invocar una herramienta, no al
responder una pregunta normal. Toda herramienta conserva flags, permiso
efectivo y confirmación independiente.

## Historial y adjuntos del drawer

Cada chat crea un `conversation_id` y lo persiste junto a `usuario_creador` y
`empresa_id`. El endpoint de historial usa alcance `usuario` por defecto. El
alcance `empresa` se permite solo a roles administrativos autorizados y nunca
omite el tenant. Fotos, PDF, DOCX, XLSX, CSV y texto validado se envían como
datos no confiables al modelo para generar un borrador; toda escritura real
continúa requiriendo una herramienta cerrada, permisos y confirmación humana.

## Fuentes y aceptación de la revisión

[enterprise.go](../backend/ai/enterprise.go), [ai_enterprise.go](../backend/db/ai_enterprise.go), [ai_enterprise_orchestrator.go](../backend/handlers/ai_enterprise_orchestrator.go), [super_chat_ia_logica.go](../backend/handlers/super_chat_ia_logica.go), [chat_con_inteligencia_artificial_controller.go](../backend/handlers/chat_con_inteligencia_artificial_controller.go).

Requisitos aplicables: PCS-REQ-002, PCS-REQ-012, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
