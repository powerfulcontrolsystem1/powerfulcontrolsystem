# P108-021 - Decisión formal de alcance móvil

Fecha: 2026-07-30

Candidato: `f9396da5e41562968996b05136fffca9991b56f9`

Ambiente: staging

## Decisión

El primer lanzamiento incluye la web responsive y la PWA existente. El cliente
nativo Flutter queda explícitamente fuera del alcance inicial, sin eliminarlo
ni retirar la API móvil preparada.

La exclusión evita bloquear producción con un artefacto que no puede
reproducirse: `mobile/powerful_control_system_app` conserva únicamente
directorios locales de herramientas/compilación (`.dart_tool` y `build`) y no
hay fuente nativa ni manifiesto de paquete versionados para construir, firmar y
auditar una aplicación.

## Comprobaciones

- No existe enlace público de descarga APK/IPA ni instalador nativo PCS.
- La instalación visible al usuario corresponde a la PWA web y se conserva.
- La API `/api/v1/` permanece disponible como base futura, sin convertirse en
  superficie anónima:
  - `GET /api/v1/me`: 401 sin sesión;
  - `GET /api/v1/empresa/productos?empresa_id=12`: 401 sin sesión;
  - `GET /api/v1/auth/mobile-session`: 401 sin sesión.
- Los endpoints v1 mantienen wrappers de autenticación, permiso empresarial,
  idempotencia y aislamiento por `empresa_id`.
- Las opciones `app móvil` de Taxi/Domicilios describen geolocalización desde
  navegador y no publican un binario nativo PCS.

## Condición para una fase posterior

La aplicación nativa solo podrá entrar en un lanzamiento futuro cuando exista
fuente completa versionada, build reproducible, firma y distribución
controladas, almacenamiento seguro de tokens, revocación por dispositivo,
offline/conflictos/reintentos, push/deep links y pruebas multiempresa.

## Resultado

**P108-021 aprobada por exclusión formal.** No se eliminó ningún rastro móvil:
PWA, responsive, API v1, documentación y artefactos locales permanecen
intactos. El cliente nativo no bloquea el GO del lanzamiento web.
