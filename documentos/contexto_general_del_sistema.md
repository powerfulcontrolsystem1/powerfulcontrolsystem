# Contexto general del sistema

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Qué es PCS

Powerful Control System (PCS) es una plataforma SaaS POS/ERP multiempresa. El
backend está escrito en Go, persiste únicamente en PostgreSQL y sirve un frontend
estático HTML/CSS/JavaScript. Incluye administración de empresas, usuarios,
licencias, ventas, caja, inventario, compras, finanzas, documentos, reportes,
integraciones de pago y fiscales, IA, impresión y Domótica Raspberry.

El [mapa de módulos](mapa_modulos.md) ubica páginas, handlers, tablas y pruebas.
La [arquitectura](arquitectura/descripcion_arquitectura.md), la [estructura del
código](diagramas/estructura_del_codigo.md) y la [estructura BD](estructura_bd.md)
amplían este resumen. El [estado actual](estado_actual.md) registra límites que
siguen vigentes sin convertir evidencia anterior en garantía del candidato.

## Arquitectura operativa

- `backend/`: API y procesos Go; `main.go` ensambla rutas y servicios.
- `backend/db/`: acceso PostgreSQL y reglas de persistencia.
- `backend/handlers/`: autorización, validación y contratos HTTP.
- `backend/migrations/`: migraciones ordenadas ejecutadas por `pcs-migrate`.
- `web/`: aplicación estática y recursos consumidos por navegadores.
- `deploy/`: Compose, Nginx, monitoreo y scripts operativos no secretos.
- `scripts/`: desarrollo, QA, release y administración de worktrees.
- `documentos/`: fuentes vigentes, contratos, decisiones, mapas y runbooks.

## Reglas no negociables

- PostgreSQL es el único motor de runtime; API y worker no crean esquema de
  producción. Las migraciones históricas son inmutables y los cambios usan una
  migración nueva y ordenada.
- Toda operación empresarial valida sesión, licencia, permiso y ownership por
  `empresa_id`. Un ID secundario se valida como `(empresa_id, id)`; el frontend no
  es una frontera de seguridad.
- Credenciales, tokens, dumps, datos fiscales/financieros, inventarios privados y
  evidencias con datos reales no se versionan ni se imprimen en logs.
- Pagos, documentos DIAN, correo, hardware y despliegues requieren aceptación del
  proveedor o entorno correspondiente; una prueba local no acredita producción.
- Se prefiere Go estándar. Dependencias externas, cambios en `go.mod`, datos reales
  y efectos externos requieren autorización explícita.

## Trabajo concurrente

`D:\powerfulcontrolsystem` es el checkout de integración y debe permanecer en
`main` limpio. Cada agente trabaja en un worktree exclusivo bajo
`D:\powerfulcontrolsystem.worktrees\`, creado desde `origin/main` mediante
`scripts\agent_worktree.ps1`. Las rutas compartidas y hotspots se asignan antes
de editar; solo el coordinador integra, despliega y retira worktrees.

Cada entrega informa rama, base/final, rutas, pruebas, estado del árbol, efectos
externos y riesgos. Los cambios no terminados se preservan en su worktree o en un
respaldo nombrado; no se usan stashes anónimos como coordinación ordinaria.

## Ciclo de cambio

1. Leer este contexto, `contexto_codex.md` y la fuente específica del módulo.
2. Revisar contrato, datos, permisos, consumidores y riesgos del cambio.
3. Implementar de forma acotada y mantener aislamiento multiempresa.
4. Ejecutar pruebas enfocadas; sumar PostgreSQL, UI, proveedor o hardware cuando
   el resultado dependa de esas capas.
5. Actualizar las fuentes vigentes afectadas y regenerar el catálogo documental.
6. Integrar por PR con evidencia minimizada en CI o almacenamiento externo.

La historia cronológica se consulta en commits, PR y el respaldo externo de la
reescritura. No se conserva dentro del snapshot operativo ni actúa como orden.

## Índice de ampliación

Usar [contexto específico](contexto_especifico_del_sistema.md) para elegir la
fuente por módulo; [comandos Codex](comandos_codex.md) antes de pruebas o release;
[decisiones técnicas](decisiones_tecnicas.md) para restricciones permanentes; y
[checklist multiempresa](checklist_seguridad_endpoint_multiempresa.md) ante
cualquier endpoint, consulta, permiso, importación, exportación o borrado.
