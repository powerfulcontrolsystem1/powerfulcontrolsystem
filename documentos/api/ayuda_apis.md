# Ayuda de APIs

Estado: Vigente. Responsable: Ingeniería de API. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- La API móvil ya incluye fachadas POS/offline/fiscal/notificaciones; consultar la matriz móvil para el alcance completo.
- modelo_preferido conserva compatibilidad del endpoint, pero el servidor impone el modelo global. Un catálogo de protocolos de cámaras/solar no significa adaptador operativo para cada uno.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Actualizacion: 2026-09-05.

Esta guia define el contrato transversal para consumidores y mantenedores. No
autoriza rutas, permisos ni cambios de datos: la autorizacion efectiva se
resuelve siempre en el backend.

## API movil versionada

Las nuevas aplicaciones Android/iPhone deben consumir `/api/v1/` y el contrato
`documentos/api/openapi.mobile.v1.yaml`. La primera entrega incluye autenticacion
de dispositivo, perfil, productos y clientes con sobre JSON, paginacion y
permisos empresariales. Las rutas `/api/empresa/*` existentes siguen siendo
compatibles con la web, pero no son el contrato recomendado para clientes nuevos.


Esta guia resume como consumir y mantener las APIs de Powerful Control System
sin romper seguridad multiempresa, permisos, licencias, auditoria ni reglas de
negocio.

## Fuente de verdad y alcance documental

| Documento | Finalidad | Uso correcto |
| --- | --- | --- |
| `backend/main.go` y routers internos | Registro efectivo de handlers y wrappers | Fuente de verdad para una ruta desplegada. |
| `documentos/api/openapi.generated.yaml` | Inventario generado desde `backend/main.go` y `/ready` | Descubrimiento; no presupone que GET y POST sean ambos validos para cada ruta. Revisar handler y contrato del modulo. |
| `documentos/api/inventario_api_movil.md` | Inventario completo de registros `http.HandleFunc` en backend | Cobertura y clasificacion; no reemplaza pruebas de autorizacion o negocio. |
| `documentos/api/openapi.mobile.v1.yaml` | Contrato externo estable de la API movil v1 | Fuente de verdad para aplicaciones Android/iPhone. |
| Contratos en `documentos/gobernanza_tecnica/contratos/` | Reglas por dominio sensible | Obligatorios para pagos, licencias, facturacion y permisos. |

En la actualizacion 2026-09-05 el inventario completo registra 374 rutas y el
OpenAPI generado desde `main.go` registra 325. La diferencia es esperada: los
routers internos tambien registran rutas y la especificacion generada no debe
usarse como sustituto de sus contratos concretos.

## Familias de rutas

| Familia | Uso | Reglas principales |
| --- | --- | --- |
| `/api/public/*` | Portal publico, catalogos, visitas, venta publica o documentos publicos controlados | No debe exponer datos privados sin token o criterio publico explicito |
| `/api/empresa/*` | Operacion de una empresa: ventas, carritos, inventario, clientes, caja, facturacion, reportes | Requiere `empresa_id`, sesion/alcance, wrapper de modulo, permisos efectivos y licencia |
| `/super/api/*` | Plataforma global: empresas, licencias, alertas, auditoria, correo, configuracion super | Reservado a super administrador o al alcance principal permitido |

## Regla de seguridad multiempresa

Enviar `empresa_id` no concede acceso. El backend debe validar siempre:

- usuario autenticado real;
- alcance de empresa;
- rol efectivo;
- licencia vigente y modulos habilitados;
- permisos por modulo, pagina y accion;
- que todos los IDs relacionados pertenecen a la misma empresa.

Nunca confiar solamente en URL, localStorage, cache, campos ocultos o controles
del frontend.

## Autenticacion, CSRF e idempotencia

- Las rutas web usan la sesion autenticada de PCS. Una solicitud que modifica
  estado y autentica por cookie debe incluir el mecanismo CSRF vigente del
  frontend.
- La API movil v1 usa `Authorization: Bearer <sesion_movil>` segun
  `openapi.mobile.v1.yaml`; el token se entrega una sola vez y no se documenta,
  registra ni reenvia por canales inseguros.
- `empresa_id` identifica el contexto solicitado, nunca una autorizacion. El
  servidor lo cruza con sesion, rol, permiso, licencia y pertenencia de los IDs
  secundarios.
- Toda mutacion susceptible a reintento debe implementar la idempotencia del
  modulo. En v1 se exige `Idempotency-Key` para las operaciones declaradas en
  su contrato; reutilizar una clave con otro cuerpo es un conflicto, no una
  nueva operacion.

## Contrato de errores y correlacion

Las APIs JSON devuelven errores publicos, sin SQL, trazas, rutas internas,
secretos ni cuerpos de proveedores. Las respuestas 5xx no marcadas se
normalizan con un mensaje generico y `request_id`; los handlers solo pueden
preservar mensajes 5xx que hayan sido marcados expresamente como seguros.

Forma minima esperada para un error JSON normalizado:

```json
{"ok":false,"status":500,"error":"Ocurrio un problema interno. Intenta de nuevo en unos segundos.","request_id":"req_..."}
```

El cliente debe conservar `request_id` para soporte y no intentar interpretar
texto de error como contrato de negocio. Algunas rutas legacy, descargas y
webhooks tienen formatos propios; su handler o contrato de proveedor prevalece.

| Codigo | Significado esperado |
| --- | --- |
| `400` | Faltan datos, payload invalido o aprobacion requerida |
| `401` | No hay sesion o credencial valida |
| `403` | Sin empresa, permiso, licencia o alcance |
| `404` | Recurso inexistente o no pertenece a la empresa |
| `409` | Conflicto de negocio o duplicado no idempotente |
| `422` | Solicitud sintacticamente valida pero no permitida por reglas de negocio o validacion especializada |
| `429` | Limite de uso o de tasa; respetar `Retry-After` cuando exista |
| `5xx` | Fallo temporal o interno; usar el `request_id`, no reintentar mutaciones sin su clave idempotente |

## IA empresarial y documentos

Las rutas de IA empresarial se mantienen como API web interna y exigen el
wrapper de permisos IA de la empresa:

| Ruta | Metodo | Proposito |
| --- | --- | --- |
| `/api/empresa/chat_con_inteligencia_artificial/modelos` | `GET` | Modelo global disponible para la empresa autorizada. |
| `/api/empresa/chat_con_inteligencia_artificial/modelo_preferido` | `GET`/`PUT` | Compatibilidad de preferencia; la política global del servidor limita el modelo efectivo. |
| `/api/empresa/chat_con_inteligencia_artificial/consultar` | `POST` | Consulta de texto con contexto empresarial filtrado. |
| `/api/empresa/chat_con_inteligencia_artificial/consultar_con_adjunto` | `POST` | Consulta con adjunto validado localmente. |
| `/api/empresa/chat_con_inteligencia_artificial/consultar_stream` | `POST` | Flujo SSE; no tratarlo como respuesta JSON unica. |
| `/api/empresa/chat_con_inteligencia_artificial/historial` | `GET` | Historial limitado al usuario y empresa autorizados. |
| `/api/empresa/chat_con_inteligencia_artificial/memoria` | `GET`/`PUT`/`DELETE` | Memoria consentida del propio usuario en la empresa activa. |
| `/api/empresa/chat_documentos/generar` | `POST` | Genera un documento temporal autorizado. |
| `/api/empresa/chat_documentos/exportar` | `POST` | Exporta una respuesta o conversacion autorizada. |
| `/api/empresa/chat_documentos/compartir_email` | `POST` | Comparte un documento autorizado por correo. |

El navegador no envía `safety_identifier`, claves de proveedor, modelo fuera
del catalogo permitido, identidad de otro usuario ni permisos. El servidor
deriva un seudonimo estable para OpenAI Responses, mantiene `store=false`,
valida adjuntos y conserva confirmacion humana para acciones con efecto.
Detalles: `documentos/ia_orquestador_empresarial.md`.

## Carritos, estaciones y venta directa

Endpoints principales:

```http
GET /api/empresa/carritos_compra?empresa_id={id}&include_inactive=1
GET /api/empresa/carritos_compra?empresa_id={id}&modo=venta_directa&perm_page=linkVentaDirecta
POST /api/empresa/carritos_compra
PUT /api/empresa/carritos_compra
DELETE /api/empresa/carritos_compra
GET /api/empresa/carritos_compra/items?empresa_id={id}&carrito_id={id}
POST /api/empresa/carritos_compra/items
PUT /api/empresa/carritos_compra/items
DELETE /api/empresa/carritos_compra/items
```

Venta directa usa el carrito canonico:

```text
VENTA-DIRECTA-{empresa_id}-0
```

Parametros operativos frecuentes:

- `modo=venta_directa`
- `perm_page=linkVentaDirecta`
- `estacion_id={id}` cuando el flujo viene desde estaciones
- `include_inactive=1` cuando se necesita recuperar sesiones o ver historial
- `action=cajas_abiertas`, `action=activar_estacion`, `action=pagar_estacion`

Reglas:

- caja y turno se resuelven por usuario/caja dentro de la empresa;
- abonos, descuentos, pagos mixtos y vuelto deben reflejarse en el cierre;
- no mezclar carritos, items, clientes, cajas ni productos de otra empresa;
- acciones de pago deben ser idempotentes frente a doble clic o reintento;
- modo offline solo aplica si la empresa lo activo y el carrito lo soporta.

## Energia solar

Endpoint empresarial:

```http
GET /api/empresa/energia_solar?empresa_id={id}&action=dashboard
GET /api/empresa/energia_solar?empresa_id={id}&action=catalogo
GET /api/empresa/energia_solar?empresa_id={id}&action=sistemas
GET /api/empresa/energia_solar?empresa_id={id}&action=alertas&sistema_id={id}
GET /api/empresa/energia_solar?empresa_id={id}&action=lecturas&sistema_id={id}&limit=120
GET /api/empresa/energia_solar?empresa_id={id}&action=eventos&sistema_id={id}&limit=80
POST /api/empresa/energia_solar?empresa_id={id}&action=sistema
POST /api/empresa/energia_solar?empresa_id={id}&action=alerta
POST /api/empresa/energia_solar?empresa_id={id}&action=lectura
POST /api/empresa/energia_solar?empresa_id={id}&action=probar_alerta&sistema_id={id}
```

Reglas:

- modulo independiente `energia_solar`, pagina `linkEnergiaSolar`;
- preconfiguracion disponible por tipo de empresa, apagada por defecto;
- rol `tecnico_solar` solo consulta dashboard, lecturas, eventos y alertas;
- proveedores catalogo: Victron VRM, SMA Sunny Portal, SolarEdge Monitoring y
  gateway local;
- las llaves reales deben viajar como referencias `env:*`, no como secretos en
  texto plano.

## Camaras y DVR

Endpoint empresarial:

```http
GET    /api/empresa/camaras?empresa_id={id}&action=dashboard
GET    /api/empresa/camaras?empresa_id={id}&action=catalogo
GET    /api/empresa/camaras?empresa_id={id}&action=camaras
GET    /api/empresa/camaras?empresa_id={id}&action=camara&id={camara_id}
POST   /api/empresa/camaras?empresa_id={id}&action=camara
PUT    /api/empresa/camaras?empresa_id={id}&action=camara
DELETE /api/empresa/camaras?empresa_id={id}&action=camara&id={camara_id}
```

Reglas:

- modulo independiente `camaras`, pagina `linkCamaras`;
- cada camara pertenece a un solo `empresa_id`;
- soporta catalogo operativo para RTSP, ONVIF, HLS, WebRTC, MJPEG e iframe;
- RTSP/ONVIF directo requiere gateway HLS, WebRTC o MJPEG para verse en el
  navegador;
- `url_embed` y `url_snapshot` solo deben usar `http`, `https` o ruta interna;
- `estaciones_config` permite `camaras_enabled`, `camaras_placement` y
  estaciones con `tipo_estacion=camara` mas `camara_id`;
- usuarios y claves reales deben guardarse como referencias seguras
  (`env:CAMARA_EMPRESA_*`), no impresas ni documentadas.

## Checklist para crear o cambiar una API empresarial

1. Ubicar modulo, pagina, handler, tablas y permisos en `documentos/mapa_modulos.md`.
2. Aplicar `documentos/checklist_seguridad_endpoint_multiempresa.md`.
3. Confirmar wrapper correcto en `backend/main.go`.
4. Validar `empresa_id` y todos los IDs relacionados en backend.
5. Filtrar SQL por `empresa_id` cuando la tabla sea empresarial.
6. Manejar idempotencia si la accion puede repetirse por doble clic, red,
   service worker, modo offline o concurrencia.
7. No imprimir secretos, tokens, certificados, contrasenas ni payload sensible.
8. Regenerar `node tools/auditar_api_movil.mjs` y
   `node tools/openapi_inventory.mjs`; actualizar esta ayuda y el contrato del
   modulo si cambia un contrato externo.
9. Agregar pruebas de exito y negativas de cruce entre empresas cuando el cambio
   toque datos.

## Fuentes canonicas

- `documentos/api/openapi.generated.yaml`: inventario automatico de rutas.
- `documentos/gobernanza_tecnica/contratos/contrato_permisos_contexto_y_wrappers_api_empresa.md`: wrappers, permisos y errores.
- `documentos/checklist_seguridad_endpoint_multiempresa.md`: checklist obligatoria.
- `documentos/mapa_modulos.md`: mapa de modulo, pagina, API, tablas, permisos y pruebas.
- `documentos/flujos_operativos.md`: flujos de usuario y QA.
- `backend/main.go`: registro real de rutas.

## Fuentes y aceptación de la revisión

[main.go](../../backend/main.go), [mobile_api_v1.go](../../backend/handlers/mobile_api_v1.go), [empresa_ia_empresarial.go](../../backend/handlers/empresa_ia_empresarial.go), [mobile_api_v1.md](mobile_api_v1.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../requisitos/especificacion_y_trazabilidad.md)).
