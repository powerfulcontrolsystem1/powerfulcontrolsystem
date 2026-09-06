# Roles, permisos y licencias

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Alcance revisado y límites

Chat IA: búsqueda de estaciones exige Ventas R; propuesta de consumo Ventas
R+C, licencia y confirmación. Reportes IA exige Reportes R y lectura del
dominio correspondiente. Inventario R/C conserva búsqueda/creación de
productos. La ayuda usa nombres de páginas autorizadas. Véase la
[matriz de capacidades](chat_ia_capacidades_2026-09-06.md).

- Se sustituye el documento acumulado por una entrada temática actual; el detalle previo se conserva como antecedente con enlace explícito.
- El recorrido lleva a fuentes de implementación y contratos; la clasificación documental no certifica pruebas ni revisa cada afirmación histórica como vigente.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Resolución efectiva

La autoridad está en [empresa_permisos.go](../backend/handlers/empresa_permisos.go)
y el [contrato de wrappers](gobernanza_tecnica/contratos/contrato_permisos_contexto_y_wrappers_api_empresa.md).
El rol base se combina con overrides, relación usuario/empresa, licencia y
acción del módulo. No copiar una matriz estática antigua como autorización.

| Concepto | Regla |
| --- | --- |
| R/C/U/D/A | Lectura, creación, actualización, borrado y aprobación; ciertas actions cambian el permiso esperado |
| Contexto empresarial | Validar empresa contra sesión y todos los IDs secundarios |
| Super | Privilegios de plataforma autorizados y TOTP confirmado; no confiar en rol enviado |
| admin_empresa y roles operativos | Permisos efectivos publicados por el servidor; la etiqueta del rol no basta |
| Página visible | Ayuda de navegación; no concede acceso al endpoint |
| Licencia | Restringe módulos y límites; no sustituye autorización de usuario |
| Cambio de roles/matriz fina | Aprobación trazable en rutas definidas; usuarios no exige ese código extra |
| Nómina fiscal | Lectura/emisión cruzan los permisos de Nómina y Facturación |
| Vida | empresa_id + usuario_id; un administrador no obtiene datos ajenos por el rol |

Consultar `GET /api/empresa/permisos_contexto` dentro de la empresa autorizada;
`include_matrix=1` expone el catálogo base. Verificar por separado los overrides
y el permiso de cada action. Pruebas negativas deben cubrir rol sin acción,
licencia sin módulo, ID secundario ajeno y usuario B del mismo tenant cuando
los datos sean personales.

El [mapa](mapa_modulos.md) enlaza los contratos de módulos. Las matrices previas
se conservan en la [referencia histórica](historico/2026-09-05/matriz_roles_permisos_pos_multiempresa_referencia_acumulada.md)
y no describen necesariamente los permisos efectivos de una empresa actual.

## Fuentes y aceptación de la revisión

[AGENTS.md](../AGENTS.md), [main.go](../backend/main.go), [descripcion_arquitectura.md](arquitectura/descripcion_arquitectura.md).

Requisitos aplicables: PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
