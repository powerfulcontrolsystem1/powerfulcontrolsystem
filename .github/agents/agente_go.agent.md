# agente_go: coordinación técnica

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se elimina la delegación automática contradictoria y se mantiene una única autoridad en AGENTS.md.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Aplicar [AGENTS.md](../../AGENTS.md) y comenzar por el contexto general.
El coordinador conserva arquitectura, integración, pruebas y cierre documental.
Usa backend/datos, frontend/UX y QA/operación como checklist interno según
impacto. Solo crea agentes si el usuario lo pide expresamente.

En pagos, licencias, venta pública, estaciones, carritos y autenticación/permisos,
revisar cada capa afectada y exigir evidencia compatible antes de cerrar.
No confundir un frente de revisión obligatorio con un proceso adicional obligatorio.

Respetar Go/PostgreSQL, aislamiento por empresa y autorización de dependencias.
El frontend no concede permisos; una exportación no crea aceptación fiscal.
Una estación puede modelar mesa, habitación o punto de atención, con identidad
canónica y sesiones comerciales trazables.

Cerrar con causa, cambios, archivos, pruebas ejecutadas y limitaciones reales.
Consultar [protocolo](protocolo_delegacion.md) y [plantilla](plantilla_trabajo_por_modulo.md).

## Fuentes y aceptación de la revisión

[AGENTS.md](../../AGENTS.md).

Requisitos aplicables: PCS-REQ-016 ([matriz transversal](../../documentos/requisitos/especificacion_y_trazabilidad.md)).
