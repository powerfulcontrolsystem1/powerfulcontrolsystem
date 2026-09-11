# Contrato técnico: permisos de roles y empresa

Estado: Vigente. Responsable: Ingeniería backend y QA. Revisión documental: 2026-09-11.

## Autoridad y alcance

El motor de [autorización](../../../backend/handlers/empresa_permisos.go) es la
fuente de permisos efectivos. La matriz publicada describe una petición y su
identidad validada; ocultar enlaces, cambiar un nombre o enviar headers de rol no
concede permisos. Este contrato cubre usuarios operativos, administradores,
catálogo global, perfiles propios y límites de cada empresa.

## Identidad y aislamiento

Una sesión `empresa_usuario` identifica `principal_id` y `empresa_id`. El motor
consulta al usuario por ambos IDs, comprueba correo, confirmación y estado activo,
y resuelve el `rol_usuario_id` actual. Nunca sustituye esa identidad por el
administrador que comparta correo. El acceso administrativo sigue una sesión
administrativa y relación empresarial propias.

Las rutas Super exigen una sesión tipada `admin`, sin empresa y vinculada al ID
administrativo vigente. Una cookie de usuario operativo que comparta correo con
un Super recibe `403`. Un cambio de rol, desactivación o revocación de la sesión
administrativa se observa en la siguiente petición.

Cada rol propio pertenece a `roles_de_usuario.empresa_id` y referencia una base
global activa mediante `rol_base_id`. Un ID de otra empresa, una base ajena,
inactiva o de plataforma se rechaza. El nombre libre no determina privilegios.
Las lecturas y escrituras de permisos verifican el rol y su empresa en backend.

Los cambios de correo o rol, la desactivación y la eliminación de usuarios revocan
sus sesiones operativas por empresa, ID y correo. No revocan sesiones de otras
empresas ni administrativas por compartir correo. La autorización se vuelve a
resolver en cada petición; solo se reutiliza el snapshot dentro de esa petición.
La membresía administrativa, propietario y acceso compartido se consultan sin
caché de autorización; transferencias, desactivaciones y revocaciones se observan
en la siguiente petición. Ser un compartidor histórico no concede acceso actual.
La [plantilla administrativa](../../../backend/db/roles_autorizacion_empresa.go)
se resuelve por el tipo de empresa y luego por la variante universal. No hereda
una matriz de otra vertical; una configuración aplicable ambigua o inactiva
requiere corrección del catálogo y deniega acceso mientras tanto.

## Resolución de permisos

1. Resolver sesión, usuario, empresa activa y rol persistido.
2. Construir la política base del rol operativo; para perfiles propios, usar el
   rol base global exacto, conservando el ID del perfil.
3. Aplicar ajustes del rol base y después ajustes del perfil propio por ID.
4. Intersectar con licencia, módulos de la vertical, techo de la empresa y alcance
   de acceso compartido. Estos límites no se convierten en concesiones.
5. Derivar páginas desde las acciones efectivas y aplicar restricciones de página.
6. Autorizar módulo y acción de la ruta. Los IDs secundarios y reglas de negocio
   siguen siendo responsabilidad del handler y sus consultas por empresa.

Las acciones son `R` lectura, `C` creación, `U` actualización, `D` eliminación y
`A` aprobación. Los métodos aportan el valor inicial; las acciones financieras,
fiscales y administrativas pueden exigir otro permiso. Consultar el mapeo real
por wrapper; no deducir permisos solamente de `GET` o `POST`.

Un fallo al leer la política, el rol o sus ajustes detiene la autorización. No
se recupera acceso mediante la política base. Los perfiles operativos y las
preferencias de menú no reactivan denegaciones. `super_rol_habilitado` en una
licencia no eleva un supervisor a administrador.

## Catálogo y escalabilidad

[Roles](../../../backend/db/roles_tipos_usuario.go) administra catálogo global y
perfiles propios. El catálogo empresarial publica `asignable` y `nombre_visible`,
prioriza variantes del tipo de la empresa y conserva sus IDs. Si solo existen
variantes de otros tipos, muestra cada una con su tipo e ID; no decide sus
permisos por orden de creación. Conserva los IDs globales históricos usados por
sus usuarios. Las asignaciones se consultan mediante
[IDs distintos por empresa](../../../backend/db/empresa_usuario_role_ids.go), sin
recorrer perfiles ni credenciales de cada usuario.

Los perfiles propios tienen matrices independientes en las tablas existentes
`roles_de_usuario_permisos` y `roles_de_usuario_paginas_permisos`. No se necesita
una nueva rama condicional en Go por cada nombre de perfil. Los módulos, acciones
y páginas pertenecen al catálogo central del motor: un módulo nuevo requiere
registrar su wrapper y reglas, y ampliar las pruebas. No se aceptan claves libres
que no correspondan a ese catálogo.

Un rol global activo sin política base conocida parte sin concesiones y obtiene
solo sus permisos persistidos. La herencia de módulos y páginas se carga en dos
consultas agrupadas por empresa y por IDs, conservando el orden base/perfil.
El perfil Mesero puede registrar pedidos, pero parte con tarifas por minuto/día,
tarifas de motel y administración de descuentos ocultas. Un permiso explícito
puede habilitarlas. Cajero conserva la consulta de descuentos necesaria para el
POS; crear, editar o eliminar códigos exige la página y acción correspondientes.

Las altas homónimas se serializan por empresa. El reemplazo de una matriz se
realiza en una transacción que bloquea el rol y su base, valida nuevamente su
alcance y compara una revisión del estado persistido. Un formulario obsoleto se
rechaza con `409`; la ausencia de revisión devuelve `428`. Un error conserva la
matriz anterior completa. Las operaciones runtime verifican
esquema existente; solo `pcs-migrate` aplica DDL.

El editor global también compara una revisión antes de guardar. Cambiar la base
de un perfil propio exige `expected_rol_base_id`, capturado al abrir el formulario;
una descripción editada desde un formulario antiguo no restaura una base revocada.

## APIs

| Ruta | Contrato |
| --- | --- |
| `GET /api/empresa/permisos_contexto?empresa_id=N` | Identidad, rol, acciones, módulos y páginas efectivos para la petición |
| Contexto con `include_matrix=1` | Requiere actualización de seguridad; la matriz global base es informativa, no sustituye el contexto efectivo |
| `GET /api/empresa/roles_de_usuario?empresa_id=N` | Catálogo global asignable y perfiles propios, conservando IDs usados |
| `POST/PUT/DELETE /api/empresa/roles_de_usuario` | Crear, editar o desactivar perfiles propios; nunca modifica roles de otra empresa |
| `GET /api/empresa/roles_de_usuario?empresa_id=N&action=permisos&rol_id=M` | Matriz del perfil propio, base y catálogo de acciones/páginas |
| `PUT` a la misma URL | Reemplazo de los ajustes de módulos y páginas del perfil propio |
| `/api/empresa/permisos_empresa` | Techo empresarial independiente de la matriz de cada perfil |
| `/super/api/roles_de_usuario` y `/super/api/roles_de_usuario/permisos` | Gestión global exclusiva de la sesión Super |

El editor empresarial delega en
[EmpresaRolDeUsuarioPermisosHandler](../../../backend/handlers/roles_tipos_usuario.go)
desde [usuarios_empresa.go](../../../backend/handlers/usuarios_empresa.go). Exige
`TenantContext`, `rol_id` coherente y perfil de la empresa. El JSON contiene
`rol_id`, `revision`, `permisos_modulo[{modulo,accion,permitido}]` y
`permisos_pagina[{pagina_clave,permitido}]`. Las listas contienen únicamente
excepciones explícitas; una clave ausente hereda su base. Rechaza claves desconocidas,
duplicadas, IDs cruzados y cuerpos inválidos o excesivos antes de guardar.

Las mutaciones de roles y del techo empresarial conservan la aprobación trazable
exigida por el wrapper de seguridad: aprobador, código y motivo. El CRUD de usuarios
usa permisos, tenant y auditoría sin exigir ese código adicional. La auditoría
posterior del wrapper no garantiza atomicidad con la transacción de negocio.

## Interfaz

[Administrar usuarios](../../../web/administrar_empresa/administrar_usuarios.html)
usa el [editor por rol](../../../web/js/empresa_role_permissions.js): búsqueda,
agrupar módulos y páginas, cambios pendientes y modo consulta cuando no existe
`seguridad:U`. Cada acción o página permite heredar, permitir o denegar. El guardado
conserva las excepciones aunque haya filtros visibles y no fija la herencia de
las casillas que no se editaron. Un conflicto o resultado de guardado incierto
conserva el borrador y exige recarga explícita antes de otro intento.

[Configuración de permisos](../../../web/administrar_empresa/configuracion_permisos.html)
edita el techo empresarial. [Permisos globales](../../../web/super/permisos_rol.html)
conserva el ID de la matriz cargada; cambiar rápidamente el selector no permite
guardar los datos de un rol contra otro. Durante carga o error no se guarda una
matriz vacía ni anterior.

## Validación requerida

Ejecutar pruebas dirigidas de roles, permisos, sesiones y rutas con PostgreSQL
local aislado. Cubrir empresas A/B, dos perfiles con la misma base, IDs ajenos,
rol/base inactivos, denegaciones por módulo/página/licencia, errores de esquema,
revocación en la siguiente petición, rollback y concurrencia de altas/matrices.

[qa_role_permissions.cjs](../../../tools/qa_role_permissions.cjs) valida el
controlador y puede servir datos sintéticos para interacción de escritorio/móvil.
No acredita autenticación ni API productiva. Las pruebas de sesiones están en
[empresa_usuario_sessions_test.go](../../../backend/db/empresa_usuario_sessions_test.go)
y la implementación en
[empresa_usuario_sessions.go](../../../backend/db/empresa_usuario_sessions.go).

La comprobación visual autenticada, pruebas de cuentas operativas, integración,
CI y despliegue son evidencias distintas. Un catálogo estático o un healthcheck
no certifica permisos de todos los usuarios.

Pruebas de regresión relacionadas:

- [Aislamiento y transacciones de roles](../../../backend/db/roles_empresa_postgres_test.go).
- [Membresía administrativa actual](../../../backend/db/admin_empresa_authorization_test.go).
- [Revocación de licencia](../../../backend/db/licencias_permiso_cache_test.go).
- [Contrato del editor](../../../backend/handlers/roles_empresa_permisos_test.go).
- [Denegaciones e identidad tipada](../../../backend/handlers/empresa_permisos_security_test.go).
- [Permisos efectivos con PostgreSQL](../../../backend/handlers/empresa_permisos_postgres_test.go).
- [Identidad y permisos en IA](../../../backend/handlers/ai_enterprise_request_permissions_test.go).
- [Servidor visual local con PostgreSQL](../../../backend/handlers/roles_visual_postgres_test.go):
  habilitar `PCS_ROLES_VISUAL_QA=1` y `PCS_TEST_POSTGRES_DSN` de una base local
  aislada, ejecutar desde `backend` la prueba `TestRolesVisualPostgresQA` con
  `go test -p 1 -timeout 12m -count=1 -v ./handlers -run ^TestRolesVisualPostgresQA$`
  y abrir `http://127.0.0.1:8875`. Sirve el editor real contra handlers y PostgreSQL;
  reemplaza únicamente el login por identidades sintéticas. Se limita a loopback,
  caduca a los diez minutos y limpia su esquema al terminar.
- [Frontera de sesión Super](../../../backend/handlers/roles_super_session_test.go).

## Operación Domótica desde una estación

La excepción operativa se limita a `GET action=estacion_controls` y a los POST
`action=probar_rele|temporizador_rele` de `/api/empresa/control_electrico`. Se
autorizan mediante `ventas` (`R` para consultar y `U` para operar), aunque la
licencia no incluya el módulo administrativo `control_electrico`.

La excepción no concede acceso a configuración, Raspberry, relés globales,
programaciones, escenas ni SSH. El handler valida `empresa_id`, rango de estaciones
de la caja y rechaza Domótica cuando la caja se configura como `solo_activar`.

## Referencias

[Matriz de roles](../../matriz_roles_permisos_pos_multiempresa.md),
[checklist multiempresa](../../checklist_seguridad_endpoint_multiempresa.md),
[contrato de autenticación](contrato_autenticacion_administrativa_y_usuarios_empresa.md),
[autorización OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html)
y [seguridad multiempresa OWASP](https://cheatsheetseries.owasp.org/cheatsheets/Multi_Tenant_Security_Cheat_Sheet.html).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002 y PCS-REQ-016.
