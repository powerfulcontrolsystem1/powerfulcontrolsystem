# Registro de riesgos y brechas de documentación

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Cómo usarlo

Este registro separa una base documental organizada de la certificación funcional u organizacional. No sustituye los hallazgos del [informe general](../informe_general_produccion_seguridad_2026-09-05.md) ni la [auditoría fiscal](../auditoria_facturacion_seguridad_20260905.md). No reasigna responsables nominales ni autoriza implementar correcciones ajenas al alcance de la tarea.

| ID | Prioridad / estado | Riesgo o brecha | Responsable por rol | Criterio verificable de cierre |
| --- | --- | --- | --- | --- |
| DOC-01 | Alta / corregida en fuentes de entrada | Contextos mezclaban planes vigentes/retirados y órdenes incompatibles de bootstrap | Coordinación técnica | Contextos breves, estado central y snapshots históricos; revisión del diff |
| DOC-02 | Alta / lote documental reconciliado; brechas específicas abiertas | Se revisó la disposición de 138 referencias y se separó historia de fuentes actuales | Ingeniería del módulo | Mantener fuentes y resolver los hallazgos concretos de la [revisión semántica](revision_semantica_2026-09-06.md); la historia no se certifica línea por línea |
| DOC-03 | Alta / abierta | No existe trazabilidad completa de todos los requisitos a código, prueba y aceptación | Producto, ingeniería y QA | Completar matrices por módulo desde requisitos transversales y contratos; no basta inventario de rutas |
| DOC-04 | Alta / correcciones locales reportadas; producción pendiente | El cierre H01–H12 registra reparaciones locales; QA publicada del 06 confirma falta de paridad de candidato | Seguridad e ingeniería | Repetir aceptación sobre SHA/digests en staging; conservar NO-GO hasta evidencia necesaria |
| DOC-05 | Alta / abierta | No se ha acreditado aquí carga, SLO, restore ni preparación general del candidato | QA/operación | Medir y aprobar objetivos en entorno identificado; release con artefactos reproducibles |
| DOC-06 | Alta / abierta | Ownership nominal, suplencias, guardia y responsabilidades de privacidad no verificados | Dirección técnica y organización | Directorio privado vigente y aprobación real de responsables |
| DOC-07 | Alta / abierta | Retenciones, tratamiento de datos y obligaciones por jurisdicción requieren validación específica | Privacidad/producto y asesoría competente | Registro de tratamientos, plazos, proveedores y requisitos aplicables aprobados |
| DOC-08 | Media / control instalado | Enlaces y catálogo pueden volver a divergir | Ingeniería y revisores | Generación determinista, control CI y revisión humana en cada cambio |
| DOC-09 | Media / abierta | Inventarios generados de rutas/estructura pueden no corresponder al candidato | Ingeniería | Regenerar con herramientas existentes, revisar diff y probar consumidores |
| DOC-10 | Media / abierta | Adopción de normas sin evaluación íntegra de cláusulas | Coordinación técnica | Matriz de aplicabilidad y evidencia revisada con textos/licencias pertinentes; no usar etiquetas de certificación |
| DOC-11 | Media / control instalado | Procedimientos breves de deploy/rollback duplicaban el runbook central | QA/operación | Consolidación, retiro de duplicados y reparación de referencias |

Las prioridades expresan urgencia de decisión, no CVSS ni un incidente confirmado. Para un hallazgo nuevo agregar origen, evidencia segura, impacto, mitigación y próximo criterio; para cerrarlo conservar el ID y enlazar el resultado. No fijar una fecha ficticia de resolución.
