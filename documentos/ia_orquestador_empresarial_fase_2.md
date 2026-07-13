# IA empresarial - fase 2 de gobierno de datos

Esta fase amplía el contrato definido en `ia_orquestador_empresarial.md` sin
habilitar un agente autonomo ni una recuperacion documental semantica.

## Controles implementados

- Clasificacion de datos: publico, interno, confidencial de negocio, personal,
  financiero, tributario, credencial y altamente sensible.
- Minimizacion previa a proveedor: solo se conservan campos expresamente
  autorizados; los nombres sensibles se descartan incluso si llegan incluidos
  en la lista de campos.
- Deteccion defensiva de instrucciones no confiables en contenido recuperado.
  Es una senal de seguridad, no una autorizacion para ejecutar acciones.
- Registro empresarial de ejecuciones con resultado saneado, riesgo, duracion,
  fuentes y categorias, sin prompt, secreto, token, QR ni respuesta privada.
- Catalogo de fuentes preparado para version, hash, modulo, nivel de
  confidencialidad y permiso requerido. La tabla no habilita RAG por si sola.
- Memoria por empresa y usuario, con consentimiento y vencimiento. Ningun flujo
  la consume hasta que exista una politica de retencion, borrado y evaluacion.

## Limites vigentes

- `AI_RAG_ENABLED` debe permanecer apagado: no hay recuperacion por fragmento,
  embedding, indexacion ni envio de archivos a un proveedor en esta fase.
- `AI_AGENT_MODE_ENABLED` debe permanecer apagado: el modelo no puede navegar,
  pulsar controles, elegir endpoints, ejecutar SQL ni encadenar herramientas.
- La unica escritura habilitable sigue siendo la propuesta hotelera del
  servidor, con confirmacion humana, hash, idempotencia y alcance de empresa.

## Activacion futura

Antes de habilitar RAG o agente se requiere una revision independiente de:

1. permisos por fragmento y revocacion inmediata;
2. reindexacion, eliminacion y retencion por empresa;
3. presupuesto, limite de operaciones y circuit breaker;
4. evaluaciones de aislamiento multiempresa e inyeccion documental;
5. telemetria agregada de coste, latencia, errores y tasa de confirmacion.
