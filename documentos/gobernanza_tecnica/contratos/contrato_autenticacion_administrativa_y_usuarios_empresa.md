# Contrato tecnico: autenticacion administrativa y usuarios de empresa

Fecha: 2026-04-18
Estado: vigente

## Alcance

Este contrato cubre el acceso administrativo por Google o correo, el registro y confirmacion de administradores, la aceptacion obligatoria del contrato, la definicion de contraseña local para cuentas Google, y el acceso de usuarios de empresa con primer password, login, recuperacion, restablecimiento y cambio de contraseña por `empresa_id`.

## Rutas implicadas

### Administradores

- `POST /super/api/administradores/register`
- `POST /super/api/administradores/login`
- `GET /auth/confirmar_admin`
- `POST /super/api/administradores/solicitar_recuperacion`
- `POST /super/api/administradores/restablecer_password`
- `GET /auth/google/login`
- `GET /auth/google/callback`
- `GET /accept.html`
- `POST /accept/complete`
- `POST /api/account/change_password`
- `POST /api/account/set_google_password`
- `GET /api/account`

### Usuarios de empresa

- `POST /api/empresa/usuarios/login`
- `POST /api/empresa/usuarios/establecer_password`
- `POST /api/empresa/usuarios/solicitar_recuperacion_password`
- `POST /api/empresa/usuarios/restablecer_password`
- `POST /api/empresa/usuarios/cambiar_password`
- `GET /auth/confirmar_correo`

### Frontend de apoyo

- `web/login.html`
- `web/accept.html`
- `web/registrar_nuevo_usuario_administrador.html`
- `web/registrar_contrasena_usuario_de_google.html`
- `web/login_usuario.html`

## Entradas obligatorias

### Registro administrativo

- `email`
- `name`
- `telefono`
- `pais`
- `ciudad`
- `password`

### Login administrativo por correo

- `email`
- `password`

### Recuperacion administrativa

- solicitud: `email`
- restablecimiento: `email`, `token`, `password`

### OAuth y aceptacion administrativa

- `code` en `GET /auth/google/callback`
- `payload` cifrado en `POST /accept/complete`

### Usuarios de empresa

- login: `empresa_id`, `email`, `password`
- primer password: `empresa_id`, `email`, `documento_identidad`, `password`
- recuperacion: `empresa_id`, `email`
- restablecimiento: `empresa_id`, `email`, `token`, `password`
- cambio de password: `empresa_id`, `email`, `current_password`, `new_password`

## Entradas opcionales

- `accept_contract` en login, primer password, restablecimiento o cambio de password de usuario de empresa.
- `password_confirm` o campos equivalentes de confirmacion cuando la UI los utilice.
- `login_hint` en la URL publica de OAuth no forma parte del contrato funcional actual y no debe considerarse requisito del flujo.
- `next` dentro del payload cifrado de aceptacion administrativa para orientar el redirect final permitido.

## Salidas y estados funcionales

### Administradores

- `200` con `ok=true` y `email_sent=true|false` en registro y solicitud de recuperacion.
- `200` con `ok=true` y `redirect_url` en login administrativo exitoso.
- `200` con `ok=false` y `password_setup_required=true` cuando la cuenta administrativa aun no tiene contraseña local activa.
- `302` desde OAuth hacia Google, `accept.html`, `super_administrador.html` o `seleccionar_empresa.html` segun el punto del flujo.

### Usuarios de empresa

- `200` con sesion creada y redirect operativo implícito al panel de empresa cuando el login es exitoso.
- `200` con `password_setup_required=true` cuando el usuario aun no definio su primera contraseña.
- `200` con `password_rotation_required=true` cuando la politica de seguridad exige cambio de contraseña.
- respuesta estructurada de requerimiento de contrato cuando falta aceptar la version vigente.

### Codigos de error esperados

- `400` por payload invalido, correo invalido, token faltante o contraseña insuficiente.
- `401` por credenciales invalidas o token de recuperacion invalido.
- `403` por correo no confirmado, usuario inactivo o acceso super no permitido.
- `404` cuando la cuenta no existe en el alcance esperado.
- `409` cuando se intenta definir contraseña inicial sobre una cuenta que ya la tiene configurada.
- `429` para usuario de empresa bloqueado temporalmente por intentos fallidos.

## Invariantes

1. El flujo administrativo publico solo puede promover a `super_administrador` al correo reservado `powerfulcontrolsystem@gmail.com`; toda cuenta nueva o reingresada por flujo publico queda como `administrador`.
2. Un administrador no puede iniciar sesion por correo hasta confirmar el correo y disponer de contraseña local activa.
3. El callback OAuth debe crear o actualizar la cuenta administrativa, pero solo puede crear sesion inmediata si la version vigente del contrato ya fue aceptada para esa cuenta.
4. Si la cuenta autenticada por Google no ha aceptado la version vigente del contrato, el flujo debe redirigir a `accept.html` usando un payload cifrado con expiracion corta.
5. `POST /accept/complete` debe persistir aceptacion de contrato y crear la sesion administrativa sin depender de estado client-side previo.
6. Si una cuenta administrativa autenticada por Google no tiene `password_set`, debe redirigirse a `registrar_contrasena_usuario_de_google.html` y no debe asumir login por correo como ya habilitado.
7. Las rutas publicas de usuarios de empresa deben mantenerse acotadas por `empresa_id` usando `WithEmpresaPublicScope` o validacion equivalente del handler.
8. Un usuario de empresa no puede autenticarse, definir contraseña ni recuperar acceso fuera del alcance de su `empresa_id`.
9. Los usuarios de empresa deben tener correo confirmado y estado activo para login, primer password, reset o cambio de contraseña.
10. La aceptacion del contrato vigente es obligatoria para login, primer password, restablecimiento y cambio de contraseña de usuario de empresa.
11. El login de usuario de empresa debe registrar intentos fallidos y puede bloquear temporalmente el acceso; un login exitoso debe limpiar el contador de fallos.
12. La cookie real de autenticacion debe seguir siendo `session_token` con `HttpOnly`; la UI solo puede apoyarse en cookies auxiliares visibles para estado de navegador.

## Side effects obligatorios

- alta o upsert de administradores en `pcs_superadministrador`
- generacion y persistencia de tokens de confirmacion o recuperacion
- envio de correos de confirmacion y recuperacion, o respuesta degradada `email_sent=false` cuando la cuenta ya fue creada pero el SMTP falla
- creacion de sesiones administrativas y de usuario de empresa en la base super
- persistencia de aceptacion contractual por version vigente
- actualizacion de password hash y salt para administradores o usuarios de empresa
- registro y limpieza de fallos de login para usuarios de empresa

## Errores de contrato esperados

- registro administrativo no debe sobrescribir una cuenta ya confirmada con el mismo correo.
- una cuenta administrativa no confirmada debe recibir rechazo explicito al login por correo.
- la recuperacion administrativa debe responder de manera no enumerativa cuando la cuenta no existe o no esta confirmada.
- el login de usuario de empresa debe rechazar mismatch de `empresa_id` aunque el correo exista en otra empresa.
- el reset de usuario de empresa debe invalidar tokens expirados y limpiar el token cuando corresponda.

## Reglas de compatibilidad

- el flujo administrativo debe funcionar igual en local y VPS sin depender de `rememberedEmail`, `login_hint` persistido o diferencias entre `localhost`, dominio raiz y `www`.
- las columnas de seguridad de `administradores` y `users` deben autorregularizarse cuando falten en esquemas legacy compatibles.
- las URIs publicas de OAuth deben respetar el host canónico y no mezclar `www.powerfulcontrolsystem.com` con `powerfulcontrolsystem.com`.

## Evidencia tecnica minima

- pruebas de `auth_admin_handlers` para registro, login, recuperación, callback OAuth y preservación del rol administrado.
- pruebas E2E de aceptación contractual administrativa.
- pruebas de `usuarios_empresa` para login, primer password, recuperación, reset, cambio de contraseña y rechazo por `empresa_id` incorrecto.
- diagnostico del editor limpio en documentos o vistas modificadas cuando se cambie el flujo.

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_arranque_postgresql_tunel_local.md`