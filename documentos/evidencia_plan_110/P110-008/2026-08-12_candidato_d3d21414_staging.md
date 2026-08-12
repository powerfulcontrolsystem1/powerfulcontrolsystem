# P110-008 — Candidato inmutable `d3d21414` en staging

Fecha: 2026-08-12 (America/Bogota)  
Alcance: actualización controlada de **staging**, sin promoción a producción.

## Cadena verificable

- Commit: `d3d2141498a540a3d5db8a06ab37855a6ee70757`.
- Workflow `Immutable release candidate`: ejecución 31571742059, aprobada.
- El workflow construyó una vez API, migrador, worker y frontend; escaneó los
  archivos de imagen, generó SBOM y publicó referencias por digest.
- La comprobación local de release inmutable aceptó las cuatro referencias
  `repositorio@sha256`, sin credenciales ni variables privadas en la evidencia.

## Promoción de staging

La primera invocación se detuvo antes de tocar servicios porque faltaba el
digest obligatorio de ClamAV. El segundo intento incluyó el digest ClamAV que
ya estaba certificado y recreó solamente los servicios de staging. Backend,
worker, frontend, PostgreSQL y ClamAV quedaron saludables; `/health` y
`/ready` aprobaron.

El aviso de Compose sobre una red Mailu preexistente fue informativo: no impidió
la recreación ni alteró la red. Esta evidencia no congela el candidato como
final: cualquier cambio posterior exige repetir las certificaciones afectadas.
