# Plan IA

Estado: **planificacion terminada; NO-GO para activar modo agente o ampliar
escrituras en produccion**.

Fecha de corte: 2026-07-24.

Modelo ejecutor solicitado: **GPT-5.6 Terra con razonamiento medio**.

Este documento complementa el Plan 106. No lo sustituye ni autoriza `rs`,
despliegue, cambios de datos reales, activacion de flags IA, llamadas reales a
proveedores, pagos, documentos fiscales, correos, WhatsApp, borrados o cambios
de infraestructura. La tarea que crea este plan se detiene al registrarlo.

## 1. Objetivo

Convertir la IA dispersa de PCS en un asistente empresarial profesional,
coherente y seguro que:

1. responda con el contexto real permitido al rol y a la empresa activa;
2. use las mismas funciones de dominio que utiliza el sistema, sin SQL ni HTTP
   arbitrario;
3. permita a un administrador crear productos y ejecutar otras modificaciones
   autorizadas mediante propuestas tipadas, confirmacion e idempotencia;
4. no conceda a ningun usuario capacidades superiores a sus permisos;
5. tenga memoria privada por empresa y usuario, con consentimiento, retencion y
   borrado;
6. simplifique el chat: sin selector de agentes y con un unico interruptor
   visible para activar `Modo agente`;
7. deje toda lectura sensible, propuesta, confirmacion, ejecucion, rechazo y
   error trazados en la auditoria correcta;
8. mida calidad, coste, latencia y seguridad antes de habilitar cada funcion.

## 2. Fuentes oficiales y decision de modelo

La guia oficial vigente de OpenAI indica:

- `gpt-5.6-sol` es la opcion de maxima capacidad;
- `gpt-5.6-terra` equilibra capacidad y coste;
- `gpt-5.6-luna` esta orientado a volumen y eficiencia;
- Responses API es la base recomendada para razonamiento, herramientas y
  conversaciones;
- el esfuerzo debe fijarse explicitamente y `medium` es el punto de partida
  equilibrado;
- las herramientas deben usar esquemas estrictos y el servidor debe decidir
  cuales estan permitidas;
- el estado conversacional puede encadenarse, pero su persistencia y retencion
  deben decidirse conscientemente.

Referencias:

- [Guia oficial GPT-5.6](https://developers.openai.com/api/docs/guides/latest-model?model=gpt-5.6)
- [Estado de conversaciones](https://developers.openai.com/api/docs/guides/conversation-state)
- [Function calling](https://developers.openai.com/api/docs/guides/function-calling)

Decision inicial para PCS:

| Carga | Modelo inicial | Esfuerzo | Regla |
|---|---|---|---|
| Chat operativo y planificacion de herramientas | `gpt-5.6-terra` | `medium` | Predeterminado profesional |
| Clasificacion, enrutamiento y tareas simples de volumen | `gpt-5.6-luna` | `low` o `none` | Solo si evals mantienen calidad |
| Casos complejos de calidad prioritaria | `gpt-5.6-sol` | `high` | Escalamiento medido, no predeterminado |
| Rutas GPT-5.4/5.5 existentes | Sin reemplazo ciego | esfuerzo actual | Migrar por carga y con comparacion |

No se habilitan como parte de la migracion base: Pro, multiagente beta,
Programmatic Tool Calling, RAG, cache explicita ni razonamiento persistente.
Cada capacidad opcional requiere una evaluacion separada y una necesidad
medida.

## 3. Diagnostico del estado actual

### 3.1 Base aprovechable

- El catalogo ya contiene GPT-5.4 mini, GPT-5.5 y GPT-5.6 Luna, Terra y Sol.
- GPT-5.5 y GPT-5.6 usan Responses API; GPT-5.4 mini conserva Chat Completions.
- Super Administrador ya gobierna modelos habilitados, modelo operativo,
  modelo de adjuntos y esfuerzo de razonamiento.
- Las consultas y consumos se registran por empresa, proveedor y modelo.
- Existe preferencia de modelo, modo y agente por usuario.
- El orquestador empresarial tiene contexto derivado por servidor, registro
  cerrado de herramientas, propuestas con hash y vencimiento, confirmacion
  separada, idempotencia y auditoria minimizada.
- Ya existen cuatro herramientas cerradas:
  `hotel.inspect_room_station`, `hotel.configure_room_station`,
  `catalog.search_products` y `catalog.create_product`.
- `catalog.create_product` reutiliza la creacion canonica de productos y exige
  `inventario:C`.
- Existen tablas para conversaciones, propuestas, ejecuciones, memoria y
  fuentes de conocimiento.

### 3.2 Inventario estatico revisado

| Superficie | Estado actual |
|---|---|
| Chat empresarial y drawer flotante | Texto, adjuntos, voz, streaming parcial, historial, modelos y agentes |
| Chat global Super Administrador | Contexto global, adjuntos, historial y seleccion de modelo |
| Chat del selector de empresas | Resumen de empresas autorizadas, sin mutaciones |
| Chat publico/tienda | Contexto anonimo limitado al portal o catalogo publico |
| Centro IA empresarial | Diagnostico y borradores por areas; mantiene selector de agente |
| Reportes IA y documentos dinamicos | Interpreta consultas y genera/exporta documentos |
| Soportes de compras/ingresos/egresos | Extrae datos de adjuntos y precarga formularios |
| GRAFOLOGIX | Analisis local y complemento visual con GPT-5.5 |
| Renta IA, impuestos y nomina | Analisis especializado y agentes de internet acotados |
| Pedidos con IA, radio y voz | Pedido asistido por estacion y controles conversacionales |
| Agente de mantenimiento DIAN | Consulta fuentes oficiales y clasifica hallazgos |
| Gobierno IA global y OpenAI propio | Modelos, esfuerzos, limites, flags y credencial cifrada por empresa |
| Orquestador empresarial | Lectura de catalogo/hotel y propuestas de producto/hotel |

Estas superficies no comparten todavia un unico contrato de conversacion,
memoria, permisos, herramientas, auditoria, modelos y evaluaciones. El Plan IA
las trata como una sola plataforma con politicas por contexto; no crea otro
chat paralelo.

### 3.3 Brechas P0 encontradas

1. **Privacidad del historial:** `/historial` valida acceso a la empresa, pero
   lista `empresa_ai_consultas` solo por `empresa_id`. Un usuario autorizado
   puede recibir preguntas y respuestas de otros usuarios de la misma empresa.
2. **Preferencias sin alcance empresarial:** la tabla
   `empresa_ai_usuario_modelo_preferido` usa `usuario_id` como clave primaria y
   no incluye `empresa_id`. Una preferencia personal puede propagarse entre
   empresas del mismo usuario.
3. **Memoria no operativa:** `empresa_ai_memoria` existe, pero no tiene servicio,
   endpoints, politica, recuperacion, interfaz ni pruebas; ningun flujo la usa.
4. **Historial controlado por cliente:** el chat acepta `historial` enviado por
   el navegador. Se limita su forma, pero no es un hilo autoritativo recuperado
   por servidor y ligado a usuario/empresa/conversacion.
5. **Rol incompleto en el chat general:** el orquestador de herramientas deriva
   rol y permisos en backend, pero el prompt general solo contiene reglas
   parciales para administradores y cajeros. No recibe un contrato compacto y
   uniforme de rol, permisos, paginas, sede y acciones permitidas.
6. **Permiso de acceso acoplado a ventas:** las rutas del chat empresarial usan
   `WithEmpresaVentasPermissions`. La IA de contador, auditor, inventario,
   compras o recursos humanos necesita un permiso IA de entrada y permisos de
   dominio por herramienta, no una dependencia transversal de ventas.
7. **Lectura administrativa demasiado amplia:** la lectura total construye
   contexto desde tablas detectadas con `empresa_id` y elimina columnas
   sensibles por nombre. Debe migrar a contratos de datos por dominio, rol y
   proposito, no depender solo de una lista negativa de nombres.
8. **Ejecucion de herramientas separada del modelo:** el frontend reconoce
   intenciones de producto/hotel y construye formularios. Aun no existe un
   bucle profesional de function calling en Responses con herramientas
   estrictas seleccionadas por el servidor.
9. **Cobertura funcional pequena:** solo producto y estacion hotelera tienen
   herramientas empresariales cerradas. No existe cobertura universal de las
   funciones de PCS ni una matriz que demuestre que cada mutacion usa el
   servicio canonico.
10. **Auditoria fragmentada:** existen auditoria general, consultas IA y
    ejecuciones IA, pero no un contrato unico que enlace `request_id`,
    `conversation_id`, `proposal_id`, herramienta, actor, rol efectivo,
    permiso, idempotencia, resultado y recurso modificado. Las rutas del chat
    global super tampoco estan envueltas de forma uniforme por la auditoria
    global.
11. **Documentacion desalineada:** varias pantallas y documentos siguen
    describiendo GPT-5.4 mini/GPT-5.5 y selectores de agentes. La fase 2 del
    orquestador afirma que la unica escritura es hotel, aunque ya existe
    producto.
12. **Linea base no compila completa:** la prueba enfocada del 2026-07-24
    aprobo `backend/ai` y `backend/db`, pero `backend/handlers` no compilo por
    `handlers/reportes_ia_chat.go:357:2: undefined: sort`. Ninguna fase puede
    declararse verde hasta restablecer el build del SHA ejecutado.

### 3.4 UX que debe retirarse

- Selector `Modo` con `Operativo/Ayudante`.
- Selector `Agente` del chat flotante y de las paginas que lo duplican.
- Seleccion manual de `general`, configuracion, ventas, inventario, compras,
  nomina, impuestos o internet.
- Modelo expuesto como decision operativa cotidiana cuando puede elegirse por
  politica.

UX objetivo:

- un solo chat consistente;
- un interruptor accesible `Modo agente`, apagado por defecto;
- modelo mostrado como badge informativo;
- selector de modelo solo en configuracion avanzada autorizada, no en el flujo
  diario;
- el backend elige especialidad y herramientas segun intencion, pagina, rol,
  licencia y permisos;
- al activar agente se muestra alcance, presupuesto y confirmaciones vigentes;
- el interruptor nunca cambia permisos ni activa flags del servidor.

## 4. Arquitectura objetivo

```text
Usuario autenticado
  -> TenantContext y RoleContext derivados en backend
  -> politica IA por empresa, usuario, licencia y pagina
  -> memoria permitida y contexto de dominio minimizado
  -> Responses API con GPT-5.6 y herramientas estrictas permitidas
  -> propuesta del servidor
  -> confirmacion humana independiente
  -> servicio de dominio existente
  -> transaccion/idempotencia/outbox
  -> auditoria unificada
  -> respuesta y estado visible
```

Reglas:

- El modelo nunca elige `empresa_id`, usuario, rol, permiso, endpoint, tabla,
  SQL ni secreto.
- El navegador no declara capacidades; solo solicita activar o desactivar el
  modo agente.
- El registro de herramientas vive en backend y se filtra antes de cada llamada.
- `strict: true`, `additionalProperties: false` y campos requeridos en cada
  herramienta de Responses.
- Para escrituras, `parallel_tool_calls=false`.
- Toda mutacion usa un servicio de dominio canonico compartido por UI/API/IA.
- Toda escritura requiere propuesta server-owned. El nivel de riesgo decide si
  necesita una o dos aprobaciones.
- Modo agente puede encadenar solo lecturas y borradores de bajo riesgo dentro
  de presupuesto. No omite confirmaciones de escritura.
- Pagos, DIAN, roles, secretos, eliminaciones, cierre de caja/periodo y
  comunicaciones externas permanecen bloqueados o con aprobacion reforzada.
- `store:false` sera la postura inicial para OpenAI hasta que una ADR de
  privacidad apruebe otra cosa. PCS conserva solo el estado minimizado que
  necesita.
- Enviar `safety_identifier` estable y no reversible por usuario.

## 5. Memoria profesional por usuario

### 5.1 Alcances

1. **Memoria de sesion:** contexto del chat actual, con expiracion corta.
2. **Preferencias:** idioma, estilo, pagina habitual y formato de respuesta.
3. **Memoria de trabajo:** tarea activa y campos pendientes.
4. **Memoria duradera consentida:** hechos empresariales utiles que el usuario
   aprobo recordar.

### 5.2 Contrato de datos

La clave logica minima sera:

`empresa_id + usuario_id + scope + tipo + clave + version`.

Cada memoria debe incluir origen, consentimiento, nivel de confidencialidad,
fecha de creacion, ultima utilizacion, expiracion, estado y hash de la fuente.
No se guardan contrasenas, tokens, llaves, certificados, datos bancarios
completos, QR, biometria, prompts completos ni respuestas privadas innecesarias.

### 5.3 Reglas visibles

- Memoria duradera apagada por defecto hasta que el usuario acepte.
- Boton `Recordar esto` y explicacion de que se guardara.
- Pantalla `Mi memoria IA` para listar, corregir, olvidar, borrar todo y
  desactivar aprendizaje.
- El usuario ve solo su memoria; administradores ven metricas, no contenido.
- El backend revalida permisos al recuperar cada memoria. Un cambio de rol
  revoca de inmediato memorias que ya no puede usar.
- El historial de chat se filtra por empresa, usuario y conversacion.
- Una conversacion nueva no recibe automaticamente todo el historial: se usa
  resumen compacto y memorias relevantes permitidas.
- Retencion configurable, purga durable por worker y auditoria del borrado sin
  conservar el contenido borrado.

## 6. Modelo de permisos y rol

Crear un modulo de permiso IA independiente para acceder al chat, pero exigir
ademas el permiso real de cada dominio:

| Rol/caso | Respuesta | Lectura | Escritura por IA |
|---|---|---|---|
| `admin_empresa` | Contexto administrativo de su empresa | Dominios permitidos | Crear producto solo con `inventario:C`; demas segun permiso |
| `cajero` | Venta, caja y carrito | Solo su operacion permitida | Agregar al carrito/caja activa; nunca configurar productos/usuarios |
| `vendedor` | Clientes, cotizaciones y ventas | Segun matriz | Acciones comerciales autorizadas |
| `inventario`/bodega | Catalogo y existencias | Inventario permitido | Productos/stock solo con accion exacta |
| `compras` | Proveedores y compras | Compras permitidas | Borradores/recepciones segun permiso |
| `contador`/contabilidad | Finanzas y cumplimiento | Lectura/escritura segun rol efectivo | Nunca mover dinero; solo funciones contables autorizadas |
| `auditor` | Evidencia y trazabilidad | Solo lectura | Ninguna |
| `recursos_humanos` | Personal y nomina | Datos permitidos | Acciones RRHH expresamente autorizadas |
| `empresario` | Resumen ejecutivo | Solo lectura | Ninguna |
| Rol personalizado | Contexto minimo | Permisos efectivos | Interseccion exacta de permisos |

No se autoriza por nombre de rol solamente. La decision final es:

`sesion + empresa + pertenencia + licencia + pagina + permiso de modulo +
accion + estado del recurso + politica de riesgo`.

## 7. Auditoria integral

### 7.1 Contrato comun

Toda accion alcanzable debe clasificarse como `publica`, `lectura`,
`lectura_sensible`, `mutacion`, `financiera`, `fiscal`, `comunicacion`,
`destructiva`, `proveedor`, `worker` o `ia`.

Los eventos IA deben enlazar, sin guardar secretos ni contenido completo:

- empresa, sede y actor;
- rol y permisos efectivos resumidos;
- origen `ui`, `api`, `ai`, `worker` o `provider`;
- `request_id`, `conversation_id`, `proposal_id` e idempotencia hasheada;
- modelo, esfuerzo, herramienta y version de esquema;
- recurso y su identificador;
- decision `permitida`, `rechazada`, `propuesta`, `confirmada`, `cancelada`,
  `ejecutada`, `fallida` o `revertida`;
- codigo, duracion, tokens y categoria de error;
- hash antes/despues cuando sea pertinente, nunca datos privados completos.

### 7.2 Cobertura de todo el sistema

Crear un manifiesto generado de rutas, metodos, acciones, permisos, licencia,
tenant, auditoria, idempotencia y riesgo. El CI falla si:

- una ruta de mutacion no tiene politica de auditoria;
- una funcion IA no se relaciona con un servicio de dominio;
- una herramienta carece de permiso, riesgo, confirmacion o prueba A/B;
- una ruta sensible puede devolver datos sin evento auditable;
- existe una accion de worker/proveedor sin correlacion e idempotencia.

La auditoria debe probarse en UI empresarial, Super Administrador, API movil,
workers, webhooks, exportaciones, archivos y herramientas IA. La presencia de
un wrapper no sustituye verificar que el evento se escribio con el actor,
empresa, resultado y recurso correctos.

## 8. Fases de ejecucion

### PIA-000 - Congelar linea base y corregir build [P0]

1. Confirmar rama, SHA, `git status`, cambios concurrentes y Plan 106.
2. Corregir el build de `reportes_ia_chat.go` sin mezclar cambios ajenos.
3. Ejecutar `go test ./... -count=1`, `go vet ./...`, preflight y pruebas IA.
4. Crear inventario de todas las superficies IA, modelos, prompts, endpoints,
   tablas, botones, flags y proveedores.

Aceptacion: SHA fijo, build verde e inventario sin superficies huerfanas.

### PIA-001 - Cerrar privacidad de historial y preferencias [P0]

1. Cambiar historial a alcance `empresa_id + usuario_id + conversation_id`.
2. Migrar preferencias a `empresa_id + usuario_id`.
3. Evitar que un usuario consulte contenido de otro usuario.
4. Agregar pruebas A/B entre empresas, A/B entre usuarios de la misma empresa,
   roles y manipulacion de query/body/header.
5. Definir migracion y tratamiento de datos historicos sin asignar propietario
   por inferencia insegura.

Aceptacion: cero fuga entre usuarios o empresas y migracion reproducible.

### PIA-002 - Contexto autoritativo de rol y permisos [P0]

1. Crear `AIUserContext` derivado solo del backend.
2. Separar permiso de acceso IA del permiso de ventas.
3. Producir catalogo de capacidades por interseccion de permisos.
4. Sustituir lectura generica de tablas por proveedores de contexto de dominio
   con lista positiva, clasificacion y limites.
5. Añadir pruebas para todos los roles base y roles personalizados.

Aceptacion: la respuesta, contexto y herramientas coinciden con los permisos
efectivos; llamadas forzadas fallan igual que la UI.

### PIA-003 - Simplificar el chat [P1]

1. Retirar selector de agentes de `ai_chat_drawer.js`, chat normal, selector de
   empresas, Centro IA, Pedidos IA y cualquier duplicado.
2. Retirar selector `Operativo/Ayudante`.
3. Añadir un switch accesible `Modo agente`, apagado por defecto.
4. Mostrar badge de modelo, estado, alcance y presupuesto.
5. Ocultar el switch si backend no lo permite; manipular el DOM no lo habilita.
6. Validar escritorio, movil, teclado, lector, zoom, temas y estados.

Aceptacion: una sola decision visible de modo, sin selectores de especialidad y
sin perdida de voz, adjuntos, detener, minimizar, nuevo chat o streaming.

### PIA-004 - Orquestador nativo con Responses tools [P0]

1. Integrar function calling en el backend con esquemas estrictos.
2. El servidor entrega solo herramientas permitidas para ese turno.
3. Desactivar paralelismo en escrituras y limitar llamadas/tiempo/tokens.
4. Mantener propuesta, hash, confirmacion, idempotencia y transaccion.
5. Tratar contenido de usuario, memoria, documentos y resultados como no
   confiable.
6. No convertir salida textual del modelo en acciones.

Aceptacion: ninguna herramienta puede inventarse, alterar su esquema, cambiar
de empresa, saltar confirmacion o ejecutarse dos veces.

### PIA-005 - Memoria por usuario [P0]

1. Evolucionar `empresa_ai_memoria` mediante migracion de `pcs-migrate`.
2. Implementar servicio, endpoints, UI y worker de purga.
3. Añadir consentimiento, tipos, retencion, version, fuentes y revocacion.
4. Crear resumen conversacional server-owned y dejar de confiar en historial
   libre del navegador.
5. Implementar `Mi memoria IA`, olvidar item, borrar todo y exportar.
6. Probar cambio de rol, usuario compartido, empresa compartida, expiracion,
   borrado y concurrencia.

Aceptacion: memoria util, explicable, editable, borrable y sin mezcla de
usuarios/empresas.

### PIA-006 - Herramientas por dominio [P0/P1]

Construir una herramienta a la vez, reutilizando servicios existentes:

1. productos, categorias, precios y stock;
2. clientes, cotizaciones, carrito y ventas;
3. estaciones, reservas y tarifas;
4. proveedores, compras y soportes;
5. finanzas y contabilidad;
6. nomina y recursos humanos;
7. configuracion de empresa;
8. usuarios, roles y permisos;
9. reportes, documentos y exportaciones;
10. integraciones, DIAN y comunicaciones.

Cada herramienta requiere: servicio canonico, esquema estricto, permiso,
licencia, riesgo, confirmacion, idempotencia, transaccion/rollback, auditoria,
limite, errores publicos, prueba A/B y prueba visible.

Pagos, movimientos bancarios, emision DIAN, cierre/reapertura, roles, secretos,
comunicaciones y borrados no se habilitan como agente autonomo. Primero deben
existir politicas y aprobaciones reforzadas especificas.

Aceptacion: matriz de cobertura 100%; una pagina no equivale a una herramienta.

### PIA-007 - Auditoria universal y correlacion [P0]

1. Generar manifiesto de cobertura de rutas y mutaciones.
2. Unificar eventos de empresa, super, IA, worker y proveedor.
3. Correlacionar consulta, propuesta, confirmacion, servicio y recurso.
4. Cubrir rechazos de permiso y ataques A/B, no solo exitos.
5. Aplicar retencion, exportacion, integridad e indices.
6. Actualizar `Auditoria` para filtrar por origen IA, herramienta, propuesta,
   usuario, resultado y recurso, sin mostrar prompts privados.

Aceptacion: toda mutacion y lectura sensible deja evidencia completa y saneada.

### PIA-008 - Migracion por carga a GPT-5.6 [P1]

Para cada uso actual comparar:

1. modelo/prompt/esfuerzo vigente;
2. Terra con el mismo contrato y esfuerzo equivalente;
3. Terra con un nivel menor;
4. Luna para cargas simples;
5. Sol solo si Terra no alcanza calidad.

Medir exito, JSON valido, seleccion de herramienta, argumentos, reintentos,
latencia, tokens, cache, coste por exito, adjuntos y calidad visible. Fijar
detalle de imagen/PDF y `text.verbosity` de forma explicita.

Aceptacion: ningun reemplazo se aprueba solo por cambiar el nombre del modelo.

### PIA-009 - Evals, seguridad y observabilidad [P0]

Crear corpus versionado sin datos privados:

- cada rol base y personalizado;
- cada herramienta y permiso denegado;
- instrucciones maliciosas en usuario, memoria, adjunto y resultado;
- empresa A/B y usuario A/B;
- doble clic, reintento, timeout, replay y concurrencia;
- modelo caido, cuota agotada y respuesta incompleta;
- memoria incorrecta, vencida, revocada o contradictoria;
- acciones financieras, fiscales, destructivas y externas bloqueadas.

Tablero: tasa de exito, rechazo correcto, fuga, confirmacion, cancelacion,
reintentos, latencia, tokens, coste, error de herramienta y satisfaccion.

Aceptacion: cero fuga y cero escalamiento; umbrales de calidad/coste aprobados.

### PIA-010 - Documentacion y despliegue gradual [P0]

1. Actualizar mapas, flujos, roles, BD, arquitectura, ayudas y trazabilidad.
2. Eliminar textos obsoletos GPT-5.4/5.5 solo donde la ruta ya migro.
3. Documentar flags, rollback, runbook, retencion y respuesta a incidentes.
4. Habilitar primero lectura en una empresa de prueba autorizada.
5. Habilitar herramientas una por una y por empresa.
6. Activar modo agente solo despues de PIA-000 a PIA-009 y autorizacion expresa.

Aceptacion: rollout reversible, evidencia visible y ningun flag global activado
por defecto.

## 9. Matriz minima de pruebas

| Prueba | Resultado obligatorio |
|---|---|
| Admin crea producto con `inventario:C` | Propuesta, confirmacion, un producto, auditoria |
| Admin sin `inventario:C` | Sin herramienta y endpoint forzado `403` |
| Cajero pide crear producto | Explicacion de limite; ninguna propuesta |
| Cajero agrega producto al carrito propio | Solo caja/carrito activo y permitido |
| Auditor activa modo agente | Solo lecturas; cero herramientas de escritura |
| Usuario A consulta historial | Nunca ve contenido de usuario B |
| Usuario con dos empresas | Preferencias y memoria separadas |
| Cambio de rol | Memoria/contexto/herramientas se revocan al instante |
| Prompt injection en PDF/memoria | Tratado como dato; no cambia permisos |
| Doble confirmacion | Una sola ejecucion |
| `empresa_id` manipulado | Rechazo y evento auditado |
| Modelo devuelve herramienta inventada | Rechazo sin ejecucion |
| Fallo tras iniciar mutacion | Rollback o compensacion documentada |
| OpenAI no disponible | Error saneado; sistema operativo no bloqueado |

## 10. Instrucciones para GPT-5.6 Terra medio

El ejecutor debe:

1. trabajar una sola tarea `PIA-*` acotada por turno;
2. leer primero los documentos obligatorios de `AGENTS.md` y este plan;
3. confirmar `git status`, SHA y cambios ajenos antes de editar;
4. no usar subagentes salvo peticion expresa del usuario;
5. mantener Go puro, PostgreSQL, HTML/CSS/JS y cero dependencias nuevas;
6. no editar migraciones aplicadas ni ejecutar DDL desde HTTP/worker;
7. no activar flags, proveedores, `rs` ni datos reales sin autorizacion;
8. resolver decisiones con esta prioridad:
   `seguridad > aislamiento > permisos > consistencia > auditoria > UX >
   coste > conveniencia`;
9. conservar funciones existentes y reutilizar servicios de dominio;
10. detenerse si falta una decision que pueda mover dinero, emitir, comunicar,
    borrar, cambiar roles/secretos o ampliar alcance;
11. cerrar cada tarea con codigo, pruebas, evidencia visible, documentos,
    riesgo residual y rollback;
12. no declarar terminado por compilar: debe comprobar el flujo real del rol.

Formato de cierre por tarea:

- objetivo y causa;
- archivos y contratos afectados;
- permisos, tenant y auditoria;
- pruebas positivas, negativas, A/B, concurrencia y visuales;
- resultado observado;
- riesgos, rollback y siguiente `PIA-*`;
- confirmacion de que no hubo deploy ni efecto externo.

## 11. Criterio final de GO

`Modo agente` solo puede habilitarse cuando:

- PIA-000 a PIA-010 estan aprobadas;
- el historial y la memoria estan aislados por empresa y usuario;
- todos los roles tienen evals y pruebas de endpoints forzados;
- la matriz de herramientas esta completa para el alcance aprobado;
- toda mutacion reutiliza funcion canonica, confirmacion e idempotencia;
- auditoria universal demuestra actor, empresa, permiso, recurso y resultado;
- no hay vulnerabilidades criticas/altas ni fugas A/B;
- coste, latencia y calidad cumplen umbrales;
- staging y UAT visible estan aprobados;
- existe rollback;
- el usuario autoriza expresamente activacion y despliegue.

Hasta entonces, el veredicto permanece **NO-GO**.
