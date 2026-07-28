# P108-001 - calidad y cadena de suministro local

Fecha local: 2026-07-25
Commit evaluado: `f6f9546aeb926d01f76d67e250f2c78962aa0ced`
Estado: **parcial; NO-GO para promoción**

## Aprobado localmente

- `go test ./... -count=1`.
- `go vet ./...`.
- Compilación de `pcs-migrate` y `pcs-worker`.
- `tools/ensure_bootstrap_inventory.mjs --check` con 154 funciones `Ensure*`
  y 122 pasos de catálogo legado.
- `tools/migration_audit.mjs --strict`.
- `scripts/profesional_preflight.ps1 -Full`, reporte
  `documentos/reportes_profesionales/preflight_20260725_232918.md`.
- `git diff --check` antes del commit de candidato.

## Bloqueos comprobados

1. `go test -race` con CGO desactivado falla de forma esperada porque Go exige
   CGO para el detector de carreras.
2. La repetición con `CGO_ENABLED=1` falla porque no hay compilador `gcc` en
   `PATH`.
3. Docker no está disponible en este equipo; no se puede crear imagen, obtener
   digest, ejecutar Compose ni generar SBOM/Trivy locales.

## Cierre requerido

Ejecutar en CI Linux con compilador C y Docker: `go test -race` sobre al menos
`./db`, `./handlers`, `./ai`, worker y outbox; construir la imagen por digest,
analizarla con Trivy, generar SBOM y adjuntar los resultados al mismo SHA.

## Seguimiento 2026-07-26 - corrección previa a nueva ejecución CI

- La ejecución CI del candidato detectó una intermitencia en la prueba SOAP de
  habilitación: una secuencia Base64 de `ds:Signature` podía coincidir por azar
  con el detector textual de marcadores de ejemplo.
- Se corrigió el detector para evaluar el XML fiscal excluyendo `ds:Signature`,
  sin relajar el bloqueo de marcadores en los campos del documento.
- Validación local posterior: `go test ./handlers` y el caso de regresión junto
  al flujo SOAP se ejecutaron veinte veces correctamente. Falta publicar el SHA
  corregido y obtener una CI profesional verde antes del despliegue aislado.

## Seguimiento 2026-07-28 - CI Linux verde

- El SHA integrado `e65f6dcddca0733f85d95cc0ae07ef33ef35e7c3`
  completó el workflow `Professional CI` en verde.
- La ejecución incluyó suite Go, `go vet`, `go test -race ./...`, auditoría de
  vulnerabilidades, Gosec, Gitleaks, Trivy de filesystem y contenedores, SBOM,
  licencias y contratos Docker.
- El mismo SHA fue desplegado en staging y respondió `/health` y `/ready`.

Estado actualizado: **parcial**. CI y race están aprobados, pero las imágenes
que staging usa todavía fueron reconstruidas en el VPS. La nueva canalización
`release-candidate.yml` debe publicar y promover las cuatro imágenes exactas
por digest antes de aprobar P108-001.
