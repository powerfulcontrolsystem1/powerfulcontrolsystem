# Docker Compose de PCS

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se consolida la documentación de despliegue en documentos/ y se elimina la receta de migración antigua duplicada.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

La guía de operación se mantiene en [documentos/docker_vps_operacion.md](../documentos/docker_vps_operacion.md).

Usar [plataforma](docker-compose.platform.yml), [staging](docker-compose.staging.yml) o [release](docker-compose.release.yml) según el entorno. Los overrides se combinan con el archivo base. Configuración privada, migrador, digests y restore se verifican antes de abrir tráfico.

## Fuentes y aceptación de la revisión

[docker_vps_operacion.md](../documentos/docker_vps_operacion.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](../documentos/requisitos/especificacion_y_trazabilidad.md)).
