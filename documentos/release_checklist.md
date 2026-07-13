# Checklist de release

- [ ] Commit e imagen inmutables identificados.
- [ ] Revision independiente y CI verde: pruebas, race, vet, vulnerabilidades,
  secretos, dependencias, Compose, imagenes, SBOM e IaC.
- [ ] Staging efimero aprobado con datos anonimos.
- [ ] Migraciones idempotentes y rollback demostrados.
- [ ] Backup/restauracion con RPO/RTO registrados.
- [ ] Archivos privados, multiempresa, roles, CSRF y sesiones verificados.
- [ ] Proveedores externos validados en sandbox autorizado.
- [ ] Carga, limites, alertas y observabilidad aprobados.
- [ ] Plan de despliegue, rollback y responsables aprobados.
- [ ] Cambio de produccion autorizado por responsable designado.

Si cualquier item permanece sin evidencia, el release queda bloqueado.
