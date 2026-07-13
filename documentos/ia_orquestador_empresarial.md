# Orquestador IA empresarial

## Estado de activacion

La primera entrega esta implementada pero desactivada por defecto. Ninguna
empresa puede recibir cambios desde IA hasta configurar simultaneamente:

```text
AI_ENTERPRISE_ORCHESTRATOR_ENABLED=true
AI_WRITE_TOOLS_ENABLED=true
AI_HOTEL_TOOLS_ENABLED=true
```

`AI_RAG_ENABLED` y `AI_AGENT_MODE_ENABLED` se reservan para fases posteriores
y deben permanecer apagados. No se activa ninguna herramienta por la sola
respuesta textual de un modelo.

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

## Herramienta inicial

`hotel.configure_room_station` configura una estacion existente como habitacion
hotelera y registra tarifas diarias por ocupacion. Exige estacion, nombre,
moneda, check-in/check-out y tarifas sin ocupaciones duplicadas. La ejecucion
actualiza configuracion de estaciones y tarifas dentro de una misma transaccion;
si una operacion falla no se confirma ninguna parte.

La herramienta no aplica cambios masivos, no elimina tarifas existentes y no
emite documentos fiscales. El siguiente despliegue funcional debe incorporar
la tarjeta del chat que muestre la propuesta y sus botones Confirmar/Cancelar
antes de activar los flags para cualquier empresa.

## Plan de ampliacion

1. Lectura con fuentes: servicios de dominio y consultas parametrizadas con
   lista blanca, limites y filtros empresariales.
2. UI de propuestas: plan, estado anterior/esperado, riesgo, fuentes,
   confirmar, cancelar y resultado.
3. Herramientas de bajo riesgo por modulo, una a una, con pruebas de permisos,
   idempotencia y aislamiento.
4. RAG documental con PostgreSQL/pgvector solo despues de validar la extension,
   permisos por fragmento, fuentes y reindexacion. No se agregara otro motor.
5. Modo agente con presupuesto, alcance, duracion, cantidad maxima de
   operaciones, circuit breaker y confirmaciones no omitibles.
