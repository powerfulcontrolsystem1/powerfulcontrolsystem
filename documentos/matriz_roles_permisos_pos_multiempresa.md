# Roles, permisos y licencias

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Alcance revisado y límites

Chat IA: búsqueda de estaciones exige Ventas R; propuesta de consumo Ventas
R+C, licencia y confirmación. Reportes IA exige Reportes R y lectura del
dominio correspondiente. Inventario R/C conserva búsqueda/creación de
productos. La ayuda usa nombres de páginas autorizadas. Véase la
[matriz de capacidades](chat_ia_capacidades_2026-09-06.md).

Configuracion de tarifas mediante Chat IA: requiere rol efectivo
`administrador`, `admin_empresa`, `administrador_total` o
`super_administrador`, permiso `ventas:U`, herramienta de tarifas habilitada y
confirmacion separada de una propuesta temporal. Un rol operativo no recibe la
herramienta aunque intente enviar sus argumentos manualmente; confirmacion y
ejecucion vuelven a verificar rol, permiso, usuario y `empresa_id`.

Personal variable por estacion: configurar `mesero_asignado`,
`comisionista_asignado`, visibilidad o memoria del ultimo comisionista requiere
permiso de configuracion/seguridad sobre la empresa. El cobro conserva su permiso
efectivo de ventas y valida en backend que mesero y comisionista sean usuarios
activos del mismo `empresa_id`. Los reportes y ajustes de propinas/comisiones
mantienen el wrapper financiero; ocultar el selector no concede ni retira
permisos del endpoint.

Acceso por caja: configurar `cajas_config[].estaciones`, `modo_estaciones` o
asignar `acceso_estaciones_cajeros.usuarios[email].caja_codigo` requiere permiso
de configuracion/seguridad de la empresa. Backend intersecta la lista de caja y
usuario. El modo `solo_activar` permite únicamente leer el tablero y ejecutar
`action=activar_estacion`; no habilita cobros ni sustituye el rol `portero` para
ocultar las demás páginas del menú.

Activacion del carrito de estacion: `action=activar_estacion` acepta
`ventas:A` para el rol Portería y `ventas:C` para quien inicia una venta
operativa. La alternativa se limita a esa ruta/acción; no concede pago, cierre,
configuración, acceso a otra caja ni mutaciones adicionales.

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
| Super | Privilegios de plataforma autorizados por sesión y rol persistido; no confiar en rol enviado |
| admin_empresa y roles operativos | Permisos efectivos publicados por el servidor; la etiqueta del rol no basta |
| Página visible | Ayuda de navegación; no concede acceso al endpoint |
| Domótica desde estación | Las actions operativas usan `ventas:R/U`, respetan rango de caja y excluyen `solo_activar`; administrar Domótica conserva `control_electrico` |
| Auditoría de login | Consulta global exclusiva de Super Administrador; los intentos no dependen de que el usuario logre crear sesión |
| Licencia | Restringe módulos y límites; no sustituye autorización de usuario |
| Cambio de roles/matriz fina | Aprobación trazable en rutas definidas; usuarios no exige ese código extra |
| Nómina fiscal | Lectura/emisión cruzan los permisos de Nómina y Facturación |
| Vida | empresa_id + usuario_id; un administrador no obtiene datos ajenos por el rol |

Consultar `GET /api/empresa/permisos_contexto` dentro de la empresa autorizada;
`include_matrix=1` expone el catálogo base. Verificar por separado los overrides
y el permiso de cada action. Pruebas negativas deben cubrir rol sin acción,
licencia sin módulo, ID secundario ajeno y usuario B del mismo tenant cuando
los datos sean personales.

El [mapa](mapa_modulos.md) enlaza los contratos de módulos. Esta matriz y la
respuesta efectiva del backend describen el snapshot; no inferir permisos desde
capturas o documentos externos anteriores.

## Fuentes y aceptación de la revisión

[AGENTS.md](../AGENTS.md), [main.go](../backend/main.go), [descripcion_arquitectura.md](arquitectura/descripcion_arquitectura.md).

Requisitos aplicables: PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
