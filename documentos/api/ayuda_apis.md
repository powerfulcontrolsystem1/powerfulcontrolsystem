# Ayuda de APIs

Fecha: 2026-06-01
Estado: vigente

Esta guia resume como consumir y mantener las APIs de Powerful Control System
sin romper seguridad multiempresa, permisos, licencias, auditoria ni reglas de
negocio.

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

## Contrato de errores

| Codigo | Significado esperado |
| --- | --- |
| `400` | Faltan datos, payload invalido o aprobacion requerida |
| `401` | No hay sesion o credencial valida |
| `403` | Sin empresa, permiso, licencia o alcance |
| `404` | Recurso inexistente o no pertenece a la empresa |
| `409` | Conflicto de negocio o duplicado no idempotente |
| `500` | Error interno; debe incluir `request_id` o identificador de error cuando aplique |

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

## Checklist para crear o cambiar una API empresarial

1. Ubicar modulo, pagina, handler, tablas y permisos en `documentos/mapa_modulos.md`.
2. Aplicar `documentos/checklist_seguridad_endpoint_multiempresa.md`.
3. Confirmar wrapper correcto en `backend/main.go`.
4. Validar `empresa_id` y todos los IDs relacionados en backend.
5. Filtrar SQL por `empresa_id` cuando la tabla sea empresarial.
6. Manejar idempotencia si la accion puede repetirse por doble clic, red,
   service worker, modo offline o concurrencia.
7. No imprimir secretos, tokens, certificados, contrasenas ni payload sensible.
8. Actualizar OpenAPI, ayuda y contratos si cambia el contrato externo.
9. Agregar pruebas de exito y negativas de cruce entre empresas cuando el cambio
   toque datos.

## Fuentes canonicas

- `documentos/api/openapi.generated.yaml`: inventario automatico de rutas.
- `documentos/gobernanza_tecnica/contratos/contrato_permisos_contexto_y_wrappers_api_empresa.md`: wrappers, permisos y errores.
- `documentos/checklist_seguridad_endpoint_multiempresa.md`: checklist obligatoria.
- `documentos/mapa_modulos.md`: mapa de modulo, pagina, API, tablas, permisos y pruebas.
- `documentos/flujos_operativos.md`: flujos de usuario y QA.
- `backend/main.go`: registro real de rutas.
