# Checklist de release

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

- [ ] Commit e imagen inmutables identificados.
- [ ] Revision independiente y CI verde: pruebas, race, vet, vulnerabilidades,
  secretos, dependencias, Compose, imagenes, SBOM e IaC.
- [ ] Staging efimero aprobado con datos anonimos.
- [ ] Migraciones verificadas en esquema vacío y clon anonimizado; recuperación
  y compatibilidad del binario anterior demostradas. No exigir un down
  destructivo cuando corresponde una corrección hacia adelante.
- [ ] Backup/restauracion con RPO/RTO registrados.
- [ ] Archivos privados, multiempresa, roles, CSRF y sesiones verificados.
- [ ] Proveedores externos validados en sandbox autorizado.
- [ ] Carga, limites, alertas y observabilidad aprobados.
- [ ] Plan de despliegue, rollback y responsables aprobados.
- [ ] Cambio de produccion autorizado por responsable designado.

Si cualquier item permanece sin evidencia, el release queda bloqueado.

Procedimiento principal: [runbook de release](gobernanza_tecnica/runbooks/runbook_release_profesional.md).
La documentación y evidencia deben corresponder al mismo candidato; consultar
[estrategia QA](calidad/estrategia_verificacion.md) y [estado actual](estado_actual.md).
