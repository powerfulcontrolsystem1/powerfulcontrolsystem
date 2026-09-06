# Especificación transversal y trazabilidad de requisitos

Estado: Vigente. Responsable: Producto y coordinación técnica. Revisión documental: 2026-09-05.

## Alcance

Base de requisitos extraída de reglas y contratos existentes de PCS. No sustituye especificaciones detalladas por módulo ni constituye aceptación comercial nueva. Los requisitos se redactan como obligaciones verificables; la columna de evidencia indica dónde comprobarlos, sin afirmar que ya pasaron todas las pruebas.

Usuarios: super administrador, administrador empresarial, roles operativos, clientes públicos y procesos/dispositivos autorizados. El detalle de acciones está en la [matriz de permisos](../matriz_roles_permisos_pos_multiempresa.md). Escenarios de negocio en [flujos operativos](../flujos_operativos.md).

## Requisitos base

Todos son obligatorios según las fuentes del proyecto. Su implementación y aceptación deben verificarse por módulo/candidato; no se asigna porcentaje global a esta tabla.

| ID | Obligación verificable | Fuente/diseño | Aceptación requerida |
| --- | --- | --- | --- |
| PCS-REQ-001 | Toda operación empresarial debe autorizar la empresa y pertenencia de cada ID relacionado en backend | [Checklist multiempresa](../checklist_seguridad_endpoint_multiempresa.md) | Positivo A; sin sesión; rol insuficiente; tenant B; ID hijo de B; ninguna lectura ni escritura cruzada |
| PCS-REQ-002 | Cada acción debe aplicar sesión, rol, permiso y licencia que corresponda a su riesgo | [Contrato de permisos](../gobernanza_tecnica/contratos/contrato_permisos_contexto_y_wrappers_api_empresa.md) | Forzar HTTP aunque el botón esté oculto; verificar denegación sin efectos |
| PCS-REQ-003 | Un reintento de operación crítica no debe duplicar pago, inventario, documento o alta | [Decisiones](../decisiones_tecnicas.md) | Doble petición concurrente y replay con misma clave; payload distinto se rechaza según contrato |
| PCS-REQ-004 | Cobro, caja e inventario deben conservar sus invariantes transaccionales ante fallos | [Flujos POS](../flujos_operativos.md) | Inyectar fallo antes del commit, concurrencia y pérdida de respuesta; comprobar saldos y movimientos |
| PCS-REQ-005 | Un pago de licencia debe validarse contra referencia, importe, moneda y contexto esperados | [Contrato checkout](../gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md) | Firma inválida, monto/moneda distintos, replay y carrera webhook/polling; proveedor aislado |
| PCS-REQ-006 | Cada documento fiscal debe usar fuente y adaptador de su empresa, país y familia | [Contrato fiscal](../gobernanza_tecnica/contratos/contrato_facturacion_electronica_y_documentos_transaccionales.md) | Rechazar cruces de país/tenant/familia; validar fuente y numeración antes del transporte |
| PCS-REQ-007 | Recepción o duplicado fiscal no debe convertirse en aceptación sin acuse oficial verificable | [Contrato fiscal](../gobernanza_tecnica/contratos/contrato_facturacion_electronica_y_documentos_transaccionales.md) | Pendiente/rechazo mantienen estado; aceptación requiere evidencia oficial, no solo HTTP 200 |
| PCS-REQ-008 | DDL productivo debe ejecutarse por migrador y conservar checksums históricos | [Gobierno de datos](../arquitectura/gobierno_datos.md) | Esquema vacío y actualización de clon aislado; drift falla; API/worker sin privilegio DDL |
| PCS-REQ-009 | Secretos y archivos privados no deben exponerse por URL, logs, errores o exportación indebida | [Amenazas](../seguridad/modelo_amenazas_y_privacidad.md) | Traversal, ID ajeno, URL directa, error 5xx y exporte sin permiso no revelan contenido |
| PCS-REQ-010 | Un trabajo durable debe recuperarse de expiración/fallo sin duplicar su efecto lógico | [Arquitectura](../arquitectura/descripcion_arquitectura.md) | Caída tras efecto y antes de ack; lease vencido y reconciliación; consumidor idempotente |
| PCS-REQ-011 | Vida personal debe aislar información por empresa y usuario | [Vida](../vida.md) | Usuario B de misma empresa y usuario de otro tenant no acceden a datos/adjuntos de A |
| PCS-REQ-012 | Herramientas IA deben respetar alcance, permisos y confirmación de escritura | [Orquestador IA](../ia_orquestador_empresarial.md) | Prompt no puede ampliar tenant, SQL o HTTP; sin confirmación no muta; replay no duplica |
| PCS-REQ-013 | Los flujos UI deben ser utilizables con teclado y tamaños móviles; la impresión debe ser legible e independiente del tema | [Decisiones](../decisiones_tecnicas.md) | Sesión real, foco/errores/estados, viewport móvil y escritorio, POS 80 mm/carta según documento |
| PCS-REQ-014 | Un release debe identificar código, imágenes, migraciones, configuración y recuperación aprobados | [Checklist release](../release_checklist.md) | Evidencia ligada al candidato, staging y restore; autorización explícita y rollback compatible |
| PCS-REQ-015 | Logs y alertas deben permitir correlación y diagnóstico sin secretos ni datos de otros tenants | [Observabilidad](../observability_runbook.md) | Solicitud/error correlacionados; agregados seguros; acceso administrativo y redacción |
| PCS-REQ-016 | La documentación afectada debe actualizarse y conservar trazabilidad del cambio | [Marco documental](../gobernanza_tecnica/marco_documental.md) | Contratos/datos/permisos alineados, catálogo y enlaces válidos, historial y revisión humana |

## Atributos de calidad

| Atributo | Escenario y medida | Fuente de objetivo |
| --- | --- | --- |
| Adecuación funcional | Venta/cobro/documento respetan invariantes y permisos | PCS-REQ-001 a 007; aceptación del módulo |
| Eficiencia | Medir p95 de rutas críticas con carga, dataset y recursos declarados | [SLO existentes](../gobernanza_tecnica/slo_sla_operativo.md) |
| Compatibilidad | API y esquema mantienen consumidores identificados durante actualización | Contrato API, migración y prueba del release |
| Capacidad de interacción | Recorrido autenticado por teclado, móvil y errores comprensibles | PCS-REQ-013 y estrategia QA |
| Fiabilidad | Recuperar jobs y restore dentro del objetivo medido | PCS-REQ-010, SLO/RTO/RPO |
| Seguridad | Denegar cruces de tenant, acceso indebido y exposición | PCS-REQ-001, 002, 009, 011, 012 |
| Mantenibilidad | Localizar requisito, operación, esquema, prueba y responsable de cada cambio | PCS-REQ-016, catálogo y ADR |
| Flexibilidad | Incrementar réplicas con datos/storage/sesiones coherentes | Condiciones de escala en arquitectura; requiere prueba |
| Seguridad frente a daños | La domótica no debe interpretar estado visual como confirmación física | [Contrato estaciones](../gobernanza_tecnica/contratos/contrato_estaciones_sensores_ventas_simple.md); aceptación física autorizada |

## Matriz de evidencia por cambio

Cada cambio registra: ID de requisito; caso y resultado esperado; archivo/símbolo y contrato; prueba; entorno/dataset; SHA/digest y condición del árbol; resultado; evidencia minimizada; responsable/revisor; riesgo residual. Los tests omitidos se marcan `pendiente` con causa. Conservar vínculo desde el informe al requisito y desde el cierre del requisito al informe.

Puntos de partida verificables, sin afirmar ejecución nueva:

| Requisitos | Pruebas presentes para ampliar |
| --- | --- |
| 001–002 | [tenant_context_test.go](../../backend/handlers/tenant_context_test.go), [empresa_permisos_tenant_test.go](../../backend/handlers/empresa_permisos_tenant_test.go) |
| 003–005 | [payment_checkout_idempotency_test.go](../../backend/db/payment_checkout_idempotency_test.go), [contrato checkout](../../backend/handlers/payment_checkout_idempotency_contract_test.go) |
| 008 | [migrations_test.go](../../backend/db/migrations_test.go), [readiness](../../backend/handlers/runtime_schema_readiness_test.go) |
| 009 | [private_files_test.go](../../backend/handlers/private_files_test.go) |
| 010 | [outbox_worker_integration_test.go](../../backend/db/outbox_worker_integration_test.go) |
| 016 | Control `tools/docs_catalog.mjs --check` y revisión humana del diff |

## Cambios de alcance

Un nuevo módulo añade requisitos propios con ID estable, origen, prioridad, dependencias, exclusiones y aceptación. No reutilizar IDs retirados. Cambiar una obligación exige registrar el motivo y revisar consumidores, amenazas, datos y pruebas. El detalle completo por todos los módulos es una brecha registrada, no una especificación terminada por esta tabla transversal.
