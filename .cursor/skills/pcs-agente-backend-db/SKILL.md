---
name: pcs-agente-backend-db
description: Especialista en backend Go y PostgreSQL del sistema multiempresa. Use when changing handlers, db access, seguridad, autenticacion, permisos, migraciones, consultas, rendimiento, sesiones, o reglas de negocio por empresa_id.
---

# PCS Agente Backend DB

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Lectura inicial: contexto_general_del_sistema.md, contexto_codex.md y fuente del módulo según AGENTS.md.
- Los frentes son responsabilidades de análisis. Crear subagentes solo cuando el usuario lo solicita; el perfil no altera autorización, dependencias ni evidencia exigida.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Enfoque

- Revisar `documentos/diagramas/estructura_del_codigo.md` antes de cambios de flujo backend.
- Revisar `documentos/estructura_bd.md` antes de cambios en tablas, consultas o persistencia.
- Mantener PostgreSQL como motor unico en runtime.
- No agregar dependencias externas sin autorizacion del usuario.

## Cobertura prioritaria

- autenticacion, usuarios, permisos y sesiones
- pagos, licencias y venta publica
- facturacion electronica, DIAN y documentos transaccionales
- estaciones, ventas_simple y carritos
- reportes, finanzas e interoperabilidad contable

## Salida esperada

- causa tecnica
- decision implementada
- archivos/rutas/tablas afectadas
- riesgo residual
- pruebas que QA debe ejecutar

## Referencia

- `.github/agents/agente_backend_db.agent.md`

## Fuentes y aceptación de la revisión

[AGENTS.md](../../../AGENTS.md), [contexto_general_del_sistema.md](../../../documentos/contexto_general_del_sistema.md), [contexto_codex.md](../../../documentos/contexto_codex.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../../documentos/requisitos/especificacion_y_trazabilidad.md)).
