# Contratos de API movil

Actualizacion: 2026-07-14.

Contrato fuente: `documentos/api/openapi.mobile.v1.yaml`.

Todas las respuestas v1 contienen `ok`, `data` o `error`, y `request_id`. Los
errores son estables (`invalid_request`, `unauthenticated`, `forbidden`,
`not_found`, `rate_limited`, `method_not_allowed`) y no exponen SQL, secretos ni
detalles de proveedores. Colecciones usan `limit`, `offset` y `meta`; `fields`
solo permite una lista cerrada.

Autenticacion nativa:

- `POST /auth/login`: devuelve una sesion Bearer de dispositivo solo una vez.
- `POST /auth/refresh`: rota el Bearer y revoca el anterior.
- `POST /auth/logout`: revoca la sesion actual.
- `GET /me` y `GET /empresas`: restauran contexto y selector despues de abrir
  la aplicacion.

Las mutaciones POS usan `Idempotency-Key`; el servidor conserva hashes y
respuestas para rechazar reutilizacion con contenido diferente. La API no acepta
`empresa_id` del cuerpo como fuente de autoridad y cada endpoint reutiliza los
wrappers de permisos multiempresa.
