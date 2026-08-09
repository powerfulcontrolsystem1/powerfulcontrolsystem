# P109-008 - papelera y recuperación visual en 34cfd852

Fecha: 2026-08-09
Ambiente: staging aislado, PCS `empresa_id=12`
Producción: sin cambios.

## Candidato y respaldo

- SHA: `34cfd8526905f2a495b6794787411c66957f99da`.
- Workflow inmutable: `31303668393`, resultado `success`.
- Build, Trivy, SBOM, publicación y Compose por cuatro digests aprobaron.
- Preflight Full/Strict local: 22/22 compuertas aprobadas.
- Respaldo previo: `/root/pcs-staging-backups/p109-pre-dialog-34cfd852-20260809T083357Z`, con ambas bases PostgreSQL verificadas.
- Promoción: cuatro digests exactos, `--no-build`; migrador código 0 y salud/listo aprobados.

## Prueba visual autenticada

1. `SCI-0009` abrió **Enviar a papelera** mediante el nuevo diálogo, sin
   `prompt()` ni error de navegador.
2. Enviar sin motivo mantuvo el diálogo abierto y mostró la validación visible.
3. Con motivo, el soporte pasó a Papelera y registró evento `eliminar`.
4. Recuperar falló cerrado mientras `SCI-0010` seguía activo con el mismo
   archivo; el mensaje explicó el duplicado sin filtrar datos privados.
5. Se envió el duplicado QA `SCI-0010` a Papelera y se repitió la recuperación.
6. `SCI-0009` volvió a Activos y registró evento `restaurar`; `SCI-0010`
   permanece en Papelera como evidencia duplicada reversible.

El diálogo se revisó visualmente: título, explicación, campo de motivo, foco,
botones y fondo modal quedaron legibles, centrados y sin recortes.

## Conciliación

- `SCI-0009`: `rechazado/activo`, `convertido_id=0`.
- `SCI-0010`: `duplicado/eliminado`, `convertido_id=0`.
- CxP: 3; pagos: 5; movimientos: 5, sin cambios.

P109-008 sigue parcial por retención vencida, recuperación ante caída,
antivirus y A/B con segunda identidad. El ensayo no autoriza producción.
