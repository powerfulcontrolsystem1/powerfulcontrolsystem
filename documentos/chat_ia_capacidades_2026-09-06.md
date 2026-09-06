# Chat IA: capacidades y verificación local

Estado: Evidencia local. Responsable: Ingeniería y QA. Revisión documental: 2026-09-06.

## Alcance

Revisión del chat compartido, herramientas por rol, consumos en habitaciones,
reportes, historial, dictado y explicación de funciones. El árbol compartido
contiene otros cambios: esta evidencia no acredita el despliegue del conjunto.

## Cobertura del candidato

| Función | Implementación | Condiciones y límites |
| --- | --- | --- |
| Explicar una función | Instrucciones de ayuda, nombres de páginas autorizadas y acceso «Explícame paso a paso» | Una petición explicativa excluye herramientas de escritura y formularios automáticos. Guía inicial de productos, estaciones y reportes; no inventar controles no documentados. |
| Buscar productos | `catalog.search_products` | Inventario R y licencia; máximo 20 productos, categorías/bodegas acotadas. El proveedor recibe una proyección de campos. |
| Crear producto | `catalog.create_product` y tarjeta de servidor | Inventario C, flags, validación y confirmación separada; conserva el servicio canónico existente. |
| Consultar estaciones | `sales.inspect_station` | Ventas R; nombres y disponibilidad de cuentas, sin datos del huésped. |
| Agregar consumo | `sales.add_station_product` | Ventas R+C; `AI_SALES_TOOLS_ENABLED`, flags generales y confirmación. Una cuenta existente y abierta; 1–99 unidades de un producto por propuesta. |
| Reportes | `reports.generate` | Reportes R más permiso R del dominio: ventas, productos más vendidos, inventario, compras, resultados y flujo de caja. Periodo máximo de 366 días, muestra de hasta 50 filas y enlace PDF. |
| Historial | Usuario y empresa en PostgreSQL | Vista propia por defecto. Un administrador puede consultar el historial empresarial; continuar desde esa vista crea una conversación nueva a su nombre. |
| Voz a texto | Reconocimiento de voz del navegador, es-CO | El dictado normal deja texto editable. El modo conversación opcional conserva comandos de voz. Depende de soporte y permiso de micrófono. |
| Temas/responsive | Mensajes con etiquetas Tú/Agente PCS y colores distintos | Claro/oscuro, escritorio 1440×1000 y móvil 390×844. |
| Fotos/documentos | Validación de adjuntos existente | Su análisis genera texto/borradores; no equivale a importar o guardar automáticamente todas las entidades. |

## Seguridad y operación

El acceso del agente representa las capacidades del usuario, no un usuario SQL
sin restricciones. El catálogo no permite SQL, HTTP, empresa, rol ni confirmación
elegidos por el modelo. Los permisos expuestos se filtran también por licencia.
El contexto agregado legado solo se conserva para administradores con lectura
de los dominios que incluye; sus muestras genéricas se limitan a tablas
empresariales conocidas, excluyendo configuraciones, memoria privada y tablas
arbitrarias. Otros roles obtienen datos mediante herramientas autorizadas.

Se permiten hasta cuatro llamadas secuenciales para buscar referencias y
preparar una operación, con una sola propuesta de escritura por mensaje. La
cuota ligera se reserva una vez al usar herramientas; los tokens de todas las
rondas se suman. Los reportes separados usan ahora el modelo global.

La confirmación de consumo revalida propiedad de cuenta/producto, cuenta abierta,
sesión de habitación y precio/impuesto bajo bloqueo transaccional. Reutiliza
inserción de ítems, stock e importes del carrito; rechaza stock insuficiente.
Nunca abre/reinicia una habitación, cobra o cierra una venta automáticamente.
La propuesta conserva hash, vencimiento y uso único. Ante cierre de propuesta
fallido después del consumo, no se repite automáticamente la operación.

La propiedad de una conversación se comprueba por filas afectadas al registrar
o refrescar. Una colisión de ID con otro usuario/empresa falla; el mismo usuario
puede volver a abrir su conversación vencida sin recuperar autoridad ajena.

## Evidencia ejecutada

- `go test ./...`: aprobado en la revisión; incluye suites existentes. Las
  pruebas que exigen PostgreSQL externo pueden omitir ejecución si falta DSN.
- Regresiones nuevas: catálogo por permiso/flag, explicaciones sin escritura,
  parámetros de autoridad rechazados, cantidades, datasets cerrados, llamadas
  no anunciadas al modelo y suma de tokens de varias rondas.
- `tools/qa_ai_chat.cjs`: navegador Chromium con CSS/JS reales y API local
  controlada. Primer envío, abrir/cerrar/reabrir, ayuda, historial, dictado
  editable sin autoenvío y tarjeta de confirmación. Dos temas y dos tamaños,
  sin desbordamiento horizontal ni excepciones JavaScript.
- Capturas: `test_runs/ai_chat_20260906/`. El audio, respuestas IA y confirmación
  se simulan; no acreditan micrófono, OpenAI ni escritura real en PostgreSQL.
- Defectos reproducidos y corregidos: historial en memoria no inicializado
  bloqueaba el primer envío; botón flotante superpuesto impedía enviar en móvil;
  la lista tenía un alto mínimo heredado que desperdiciaba espacio.

## Pendientes explícitos

No existe paridad total con todas las operaciones del POS/ERP. Nómina,
facturación, cobros, cierres, usuarios/permisos, borrados, compras completas y
otros módulos requieren herramientas específicas y aceptación individual.
Tener una página o tabla no habilita esas acciones en el agente.

No se activaron flags ni se desplegó. Falta ejecutar la nueva escritura de
consumos con PostgreSQL aislado, roles/empresas A/B, stock, concurrencia,
rollback y doble confirmación; y completar UAT autenticada con OpenAI y
micrófono reales. El módulo crítico de ventas no se declara cerrado ni
habilitado para producción con esta evidencia.

## Fuentes

[Contrato IA](ia_orquestador_empresarial.md),
[operaciones](../backend/handlers/ai_enterprise_operations.go),
[transacción de consumo](../backend/db/ai_station_product.go),
[prueba visual](../tools/qa_ai_chat.cjs),
[function calling de OpenAI](https://developers.openai.com/api/docs/guides/function-calling).
