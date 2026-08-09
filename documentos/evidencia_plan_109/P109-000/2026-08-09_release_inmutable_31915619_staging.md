# P109-000 - Release inmutable `31915619` promovido a staging

Fecha: 2026-08-09  
Ambiente: staging aislado (`empresa_id=12` disponible para las validaciones)  
Producción: no modificada

## Objetivo y alcance

Promover a staging exactamente las cuatro imágenes construidas una vez por CI
desde `31915619a74227216b9590b5268e036b3e6a51b4`, sin recompilar ni tocar los
contenedores productivos.

## Evidencia verificable

- Workflow GitHub `31308770525`: construcción, Trivy, SBOM, publicación de
  cuatro imágenes y validación de Compose: aprobados.
- Artefacto `immutable-release-candidate-31915619...`: cuatro referencias
  `ghcr.io/...@sha256` y commit exacto, verificadas antes de promover.
- `deploy/scripts/vps-staging-digest-up.sh`: promoción por digest sin build.
  Staging respondió `health=ok` y `ready=ready`.
- API, worker y frontend quedaron saludables; el migrador terminó con código
  `0`.
- PostgreSQL de staging registra la migración
  `platform:20260809-001-dian-local-production-flag-v1` como `applied`; la
  columna `empresa_dian_configuracion.produccion_local_activa` existe.
- Las huellas de imagen y `/health` de producción fueron leídas antes/después:
  no cambió ninguna imagen productiva y su salud continuó `ok`.

## Resultado

La parte técnica de la promoción por digest aprobó para el SHA indicado. La
fase P109-000 sigue **parcial**: por decisión vigente no se creó PR, por lo que
el candidato no cuenta con revisión ni fusión a `main`. No es una certificación
de producción.

## Riesgos abiertos

- Completar los lotes E2E autenticados y las acciones mutantes/roles/IA.
- UAT contable, DIAN real controlada, cajas concurrentes, antivirus, impresión
  física/tableta, piloto y decisión GO/NO-GO.
