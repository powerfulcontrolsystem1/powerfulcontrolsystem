# P109-009 - Cadena de suministro y cabeceras del candidato

Fecha: 2026-08-01
Entorno: staging
Candidato: `6da6c13453a40b2d84e23285fa83f255f34788da`

- Cuatro imágenes: cero vulnerabilidades Trivy y cuatro SBOM CycloneDX.
- Portal: 200, HSTS, CSP, `X-Content-Type-Options`, `X-Frame-Options`,
  `Referrer-Policy` y `Permissions-Policy`.
- Endpoint empresarial sin sesión: 401.
- Preflight con origen malicioso: 401 y sin cabeceras CORS permisivas.
- El preflight profesional completo aprobó auditoría de seguridad, permisos,
  licencias, hardening, migraciones, UX y todo el backend Go.

Un escaneo rápido compilado desde el código versionado se limitó a
`127.0.0.1:8082`, pero superó la ventana de tres minutos y terminó sin artefacto;
se registra como inconcluso, no como PASS. El binario y la ruta temporal se
limpiaron automáticamente.

Durante el E2E se observaron 500 repetidos al abandonar el panel PostgreSQL
mientras sus consultas seguían activas. El diagnóstico demostró que el contexto
cancelado del cliente se convertía en error interno. La rama de cierre corrige
la clasificación: cancelación sin respuesta sintética, deadline como 503 y
detalle interno solo en log. Debe promoverse un nuevo digest y repetirse.

Estado: **P109-009 parcial** hasta DAST autenticado, sesión/CSRF/SSRF/subidas
hostiles, matriz A/B y cierre de `unsafe-inline`.
