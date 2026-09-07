# Contrato del centro de soporte

Estado: Vigente. Responsable: Ingeniería backend y QA. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Define un procedimiento organizativo y sus indicadores; las integraciones esperadas no se presentan como automatizaciones implementadas.
- Guardias, personas responsables y tiempos comprometidos requieren aprobación organizativa; los SLO del repositorio son objetivos.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Alcance

El centro de soporte debe concentrar incidentes, solicitudes operativas, problemas de acceso, pagos, licencias, errores de modulos y evidencias de clientes.

## Flujo

- Crear caso con empresa, usuario, modulo, severidad y descripcion.
- Adjuntar evidencia visual o tecnica cuando exista.
- Clasificar P1, P2 o P3 segun el SLO/SLA operativo.
- Asignar responsable y fecha objetivo.
- Cerrar con causa, solucion aplicada y validacion del usuario.

## Integraciones esperadas

- Alertas del sistema para capacidad y disponibilidad.
- Errores del panel super administrador.
- Matriz de pagos y comprobantes.
- Auditorias profesionales generadas por preflight.

## Indicadores

- Casos abiertos por severidad.
- Tiempo medio de primera respuesta.
- Tiempo medio de cierre.
- Reincidencias por modulo.

## Fuentes y aceptación de la revisión

[incidentes_y_continuidad.md](../../operacion/incidentes_y_continuidad.md), [slo_sla_operativo.md](../slo_sla_operativo.md), [main.go](../../../backend/main.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../../requisitos/especificacion_y_trazabilidad.md)).
