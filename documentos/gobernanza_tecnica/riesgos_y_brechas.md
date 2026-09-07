# Registro vigente de riesgos y brechas

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

Este registro resume condiciones aún aplicables al snapshot actual. No reemplaza
pruebas del candidato ni autoriza efectos externos.

| ID | Prioridad | Riesgo o brecha | Criterio verificable de cierre |
| --- | --- | --- | --- |
| PCS-R01 | Alta | Cobertura incompleta de requisitos a código, prueba y aceptación | Matriz del módulo enlaza requisito, implementación y prueba ejecutada |
| PCS-R02 | Alta | Un resultado local puede confundirse con aceptación de producción | La entrega separa evidencia local, PostgreSQL, UI autenticada, staging, proveedor, hardware y producción |
| PCS-R03 | Alta | Ownership incompleto en IDs secundarios puede cruzar empresas | Pruebas negativas verifican `(empresa_id, id)`, usuario y permiso sin efectos colaterales |
| PCS-R04 | Alta | Migración, restore, carga o rollback no medidos para el candidato | Ensayo sobre artefacto identificado con resultado y recuperación comprobada |
| PCS-R05 | Alta | Integraciones fiscales y pagos dependen de ambiente, firma y proveedor | Aceptación oficial de cada flujo y reconciliación idempotente, sin inferirla de simulaciones |
| PCS-R06 | Alta | Secretos o evidencia sensible pueden volver a Git | Escaneo de secretos en CI, archivos privados ignorados y rotación ante exposición |
| PCS-R07 | Media | Documentación generada puede divergir del código | Generación determinista, catálogo sin bloqueos y revisión humana en cada cambio |
| PCS-R08 | Media | Trabajo paralelo puede mezclar cambios o perder WIP | Worktree y rama exclusivos por agente, rutas asignadas y entrega estructurada al integrador |

El responsable del módulo actualiza la fuente vigente cuando cierre una brecha.
Resultados y adjuntos se conservan fuera de Git y se referencian desde la PR.
