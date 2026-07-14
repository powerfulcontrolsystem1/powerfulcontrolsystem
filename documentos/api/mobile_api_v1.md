# API movil v1

Actualizacion: 2026-07-13.

## Alcance inicial estable

La API movil es una capa aditiva: no reemplaza las rutas actuales de la web ni
rompe sus contratos. Sus rutas empiezan por `/api/v1/` y responden siempre:

```json
{"ok":true,"data":{},"meta":{},"request_id":"m_..."}
```

Un error nunca devuelve SQL, trazas, secretos ni mensajes internos:

```json
{"ok":false,"error":{"code":"forbidden","message":"No fue posible completar la solicitud."},"request_id":"m_..."}
```

Contrato OpenAPI: `documentos/api/openapi.mobile.v1.yaml`.

| Ruta | Estado | Notas |
|---|---|---|
| `POST /api/v1/auth/mobile-session` | Disponible | Crea una sesion dedicada de dispositivo desde una sesion web autenticada. El valor se entrega una vez y PostgreSQL conserva solo su hash de sesion. |
| `GET /api/v1/me` | Disponible | Admite cookie de sesion o `Authorization: Bearer`. |
| `GET /api/v1/empresa/productos` | Disponible | Aislamiento por empresa, permisos de inventario, `limit` maximo 100, `offset`, filtros y `fields` permitido. |
| `GET /api/v1/empresa/clientes` | Disponible | Aislamiento por empresa, permisos de clientes, paginacion, filtros y `fields` permitido. |

## Autenticacion y permisos

- La aplicacion movil puede usar `Authorization: Bearer <sesion_movl>` en las
  rutas v1. El middleware de autenticacion valida la misma sesion hash que usa
  PCS y las revocaciones por contrasena, rol, segundo factor o desactivacion se
  aplican de inmediato.
- La creacion del token de dispositivo exige una sesion web autenticada y el
  token se genera con 32 bytes aleatorios. El cliente debe protegerlo en el
  almacenamiento seguro nativo, nunca en preferencias sin cifrar.
- El `empresa_id` es solo una solicitud de contexto: los wrappers de permisos
  lo verifican contra la relacion usuario-empresa, rol, licencia y reglas finas
  antes de consultar datos.
- Las mutaciones con cookie siguen cubiertas por CSRF; solicitudes Bearer no
  comparten ese vector y siguen usando TLS, autenticacion y control de permiso.

## Auditoria global de APIs

Se inventariaron 361 rutas registradas desde `backend/main.go` y los routers
internos mediante `tools/auditar_api_movil.mjs`; el resultado versionado vive en
`documentos/api/inventario_api_movil.md`. Las familias cubiertas incluyen
autenticacion, empresas/roles, ventas/carritos/caja, productos e
inventario, clientes, compras, finanzas, cartera, nomina, impuestos,
facturacion, reportes, documentos/OnlyOffice/Nextcloud, GPS, estaciones,
domotica, pagos, integraciones, IA, soporte remoto, super administrador y
webhooks publicos.

Hallazgos y criterio de migracion:

1. Las APIs web existentes se mantienen por compatibilidad. Muchas devuelven
   arreglos directos o usan mensajes historicos; se catalogan como **legacy web
   interno**, no como contrato movil nuevo.
2. El contrato movil nuevo normaliza JSON, codigos, `request_id`, paginacion,
   limite de pagina y seleccion cerrada de campos. No usa `empresa_id` del
   cliente como autoridad.
3. Las rutas publicas y webhooks no se reversionan sin una prueba firmada del
   proveedor. Conservan validaciones de firma, idempotencia y limites ya
   implementados; su migracion se hace proveedor por proveedor.
4. Los siguientes candidatos quedan obsoletos para consumidores nuevos, pero
   no se eliminan todavia: respuestas sin sobre de `/api/empresa/productos` y
   `/api/empresa/clientes`, y cualquier ruta `/api/empresa/*` que el inventario
   marque como consumo de UI historica. La retirada requiere telemetria de uso,
   version v1 equivalente y una ventana de deprecacion.

## Siguientes lotes v1

Ventas/facturacion, inventario de escritura, pagos y sincronizacion offline se
migran por dominio con los mismos criterios: contrato OpenAPI, prueba de
aislamiento negativo, idempotency key para mutaciones, paginacion y errores
uniformes. No se debe exponer una tabla completa ni un endpoint generico SQL a
la aplicacion movil.
