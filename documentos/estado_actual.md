# Estado de entrega y evidencia

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-06.


## Corte documental

Chat IA (2026-09-06): [capacidades y verificación local](chat_ia_capacidades_2026-09-06.md).
Interfaz clara/oscura, ayuda, historial y dictado simulado verificados. Consumos
y reportes integrados en el candidato; sin activación/despliegue ni aceptación
de la nueva escritura en PostgreSQL real. No implica paridad con todo el ERP.

Este documento concilia las fuentes disponibles al 2026-09-06. Se revisa un árbol local con cambios previos, sobre HEAD `73cc7dfda1cbb2aaa374e99dc889c4b18dda1077`; no es un candidato congelado. Esta tarea cambia documentación y controles documentales, sin repetir auditorías funcionales ni verificar producción.

## Lectura por alcance

| Área | Evidencia disponible | Límite de la afirmación |
| --- | --- | --- |
| Sistema general | [Informe general](informe_general_produccion_seguridad_2026-09-05.md) y [cierre local de reparaciones](cierre_reparaciones_produccion_seguridad_2026-09-05.md) | El cierre registra correcciones H01–H12 y comprobaciones locales; conserva NO-GO productivo por evidencia externa pendiente. Esta revisión documental no repite esas pruebas |
| Producción autenticada PCS | [QA autenticada del 2026-09-06](evidencia_qa_autenticada_pcs_2026-09-06.md) | Login, tenant 12 y lecturas de Productos, Compras, Finanzas, MRP, WMS, Tesorería, CRM, Auditoría, Usuarios, Facturación y API Domótica funcionan; la UI Domótica falla por código publicado atrasado y producción conserva MFA/HSTS/CSP/seguridad VPS pendientes |
| Fiscal por país | [Auditoría fiscal](auditoria_facturacion_seguridad_20260905.md) | Candidato local endurece fronteras; EC/PA son configuración, sin adaptador productivo acreditado |
| Familias Colombia | [Contrato fiscal](gobernanza_tecnica/contratos/contrato_facturacion_electronica_y_documentos_transaccionales.md) | Factura, nota total, soporte y nómina tienen alcances distintos; fuente/código local no acredita habilitación ni aceptación oficial de cada familia |
| Pagos y licencias | [Contrato checkout](gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md) | Requiere evidencia específica de firma, importe, idempotencia, datos y proveedor |
| Esquema y runtime | [Gobierno de datos](arquitectura/gobierno_datos.md) | Migrador y roles separados en código; no acredita ledger ni permisos reales de cada entorno |
| Capacidad y recuperación | [Objetivos operativos](gobernanza_tecnica/slo_sla_operativo.md) | Son objetivos; no hay nueva medición de carga, restore o disponibilidad en esta tarea |
| Documentación | [Reorganización](gobernanza_tecnica/revision_documental_2026-09-05.md) y [revisión semántica](gobernanza_tecnica/revision_semantica_2026-09-06.md) | Se resolvió la disposición de las 138 referencias del lote: fuentes revisadas, historia y contratos diferenciados; no auditoría funcional de cada afirmación histórica |

## Cómo registrar una entrega

Cada módulo/candidato debe indicar requisito, SHA/digests, cambios locales incluidos, entorno, responsable, fecha, comandos, resultado, omisiones y enlace a evidencia minimizada. Separar: implementado; validado localmente; validado con PostgreSQL; UI autenticada; staging; proveedor/hardware; producción.

Un registro anterior de DIAN aceptada solo acredita ese documento y ambiente. `/health` o `/ready`, un botón visible y un inventario de rutas no acreditan aislamiento, corrección comercial ni preparación general.

No hay una hoja de ruta global vigente inferida de los planes 101–110. Los pendientes se priorizan por [riesgo](gobernanza_tecnica/riesgos_y_brechas.md) y alcance acordado. Al cerrar un hallazgo, enlazar la prueba que lo refuta o resuelve; no borrar la evidencia anterior.

## Evidencia nueva del 2026-09-06

La [revisión VPS](seguridad/revision_vps_2026-09-06.md) confirma reparaciones reales
de Nginx/SSH, HSTS, CSP Webmail y TLS de venta digital. Login autenticado funciona;
MFA obligatorio no está acreditado en el backend publicado. Persisten vulnerabilidades
altas/críticas de imágenes auxiliares, wildcard vencido y reinicio pendiente.
Las nuevas correcciones Go están probadas localmente; verificar su release aparte.
