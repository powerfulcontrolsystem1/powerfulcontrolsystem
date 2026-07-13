# Orquestador IA empresarial

## Estado de activacion

La primera entrega esta implementada pero desactivada por defecto. Ninguna
empresa puede recibir cambios desde IA hasta configurar simultaneamente:

```text
AI_ENTERPRISE_ORCHESTRATOR_ENABLED=true
AI_WRITE_TOOLS_ENABLED=true
AI_HOTEL_TOOLS_ENABLED=true
```

`AI_RAG_ENABLED`, `AI_AGENT_MODE_ENABLED` y `AI_INVENTORY_TOOLS_ENABLED` se
mantienen apagados hasta que su catalogo, permisos, evaluaciones y controles de
coste tengan evidencia de aprobacion. No se activa ninguna herramienta por la
sola respuesta textual de un modelo.

## Contrato de seguridad

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
emite documentos fiscales. El siguiente despliegue funcional debe incorporar
la tarjeta del chat ya incorpora el formulario asistido, estado actual,
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
contexto acotado por servidor. La interfaz actual solo usa modo asistido para
la herramienta hotelera.

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
