# Evidencia de endurecimiento del login - 2026-08-21

## Alcance

Candidato aislado `codex/login-hardening`, sin despliegue ni ejecucion de `rs`.
Se revisaron registro de administradores, confirmacion de correo, login por
correo y Google, recuperacion/restablecimiento, 2FA, sesiones y login de usuarios
operativos por empresa. Ninguna credencial ni token real se conserva aqui.

## Correcciones del candidato

- Sesiones tipadas para administrador y usuario empresarial con identidad,
  empresa y rol persistidos. Un usuario empresarial no puede abrir rutas
  `/super/*` ni cambiar el perfil o la clave de un administrador.
- `empresa_id` obligatorio y consistente para el acceso empresarial; se elimina
  la busqueda global por correo que podia mezclar usuarios homonimos.
- Hash de contrasena versionado con PBKDF2-HMAC-SHA256 y migracion transparente
  desde el verificador legado al autenticar correctamente.
- Tokens de confirmacion y recuperacion atomicos y de un solo uso; el cambio de
  clave revoca las sesiones existentes.
- Limitacion durable de intentos administrativos, compartida entre replicas,
  con ventana, bloqueo temporal, `Retry-After` y limpieza al autenticar.
- reCAPTCHA falla cerrado si fue activado y su configuracion es ilegible o
  incompleta; el bypass de desarrollo se ignora en produccion.
- Politica de contrasena consistente en registro, primer ingreso, recuperacion,
  cuenta y 2FA; respuestas genericas de recuperacion evitan enumerar cuentas.
- Confirmacion administrativa vuelve al login disenado, sin conservar el token
  en la URL, y muestra un estado accesible.
- Estados con `role=status`, `aria-live=polite` y `aria-atomic=true`.
- El cajon IA cerrado deja de aumentar artificialmente el alto de las paginas de
  autenticacion.
- El selector de empresas deja de usar una envoltura exclusiva de super admin:
  administradores confirmados pueden listar/crear sus empresas y gestionar sus
  accesos compartidos bajo el alcance efectivo; catalogos y mutaciones globales
  siguen reservados al super admin. La licencia de prueba exige propietario.
- El inicio Google reduce sus cabeceras de cookies por flujo. La prueba real
  encontro `upstream sent too big header` y 502: la respuesta tenia 4602 bytes
  y doce `Set-Cookie`. El candidato conserva state/PKCE, limpia solo el flujo
  contrario y verifica por prueba que ambas respuestas quedan bajo 4096 bytes.
- El login por contrasena valida la version vigente del contrato antes de crear
  sesion. Si falta aceptacion, redirige mediante payload cifrado al flujo comun
  de contrato, igualando la garantia que ya existia para Google.

## Pruebas ejecutadas

- `go test ./db ./handlers ./utils -count=1`: PASS.
- `go vet ./db ./handlers ./utils`: PASS.
- Sintaxis de los cuatro JavaScript modificados: PASS.
- Aislamiento empresarial, sesion tipada, KDF/legado, reCAPTCHA fail-closed,
  throttle durable y atomicidad de tokens: PASS.
- Revision visual local en escritorio: registro, login, recuperacion,
  confirmacion y login empresarial sin desbordamiento horizontal.
- Confirmacion visual dentro de la pantalla oficial y anunciada como estado.
- Login empresarial sin empresa o invitacion: Google queda deshabilitado y el
  formulario muestra una instruccion visible sin salir de la pagina.
- Politica visual de clave: una clave debil muestra el requisito incumplido.
- Configuracion publica real del 2026-08-21: reCAPTCHA esta solicitado pero no
  configurado ni habilitado efectivamente. El candidato lo tratara como fallo
  operativo cerrado; debe completarse antes del despliegue para no bloquear el
  login.
- Flujo real autorizado: cuenta QA creada, correo aceptado por Gmail, enlace
  recibido y activado, rechazo correcto antes de confirmar, acceso correcto
  despues de confirmar, cierre de sesion y recuperacion recibida. El enlace real
  abre el formulario de nueva clave. No se ejecuto el envio final de cambio de
  contrasena porque esa accion debe ser entregada al usuario.
- Cuenta administrativa PCS existente: login por correo abre el panel super;
  logout invalida la sesion y la ruta protegida vuelve a mostrar el acceso.
- Hallazgo real reparado en el candidato: la cuenta QA llegaba al selector pero
  `/super/api/empresas` respondia 403 por una envoltura super-only. Se separo la
  auditoria de cuenta de la autorizacion global y se agrego prueba contractual.
- Hallazgo real reparado en el candidato: Google devolvia 502 aunque el backend
  generaba correctamente el 302. Los registros Nginx confirmaron cabecera de
  upstream excesiva; se agrego presupuesto automatico de cabeceras OAuth.
- Hallazgo legal: la cuenta QA tenia contrato no aceptado y el login por correo
  creaba sesion. El candidato ahora obliga a aceptar la version vigente antes de
  emitir la cookie de sesion.
- Login operativo real sin empresa: produccion intento resolver credenciales de
  forma global, dejo Google activo y mostro `credenciales inv?lidas`. El
  candidato exige empresa o invitacion, deshabilita Google sin ese alcance y
  normaliza el mensaje visible a `Credenciales invalidas` con codificacion UTF-8.
- Flujo operativo real PCS de punta a punta: se creo desde la interfaz un usuario
  QA nuevo con rol Cajero y correo verificable; Gmail recibio la invitacion, el
  enlace abrio el formulario oficial, se validaron documento, contrato y politica
  de clave, y el primer ingreso redirigio a la empresa 12. No se documentaron
  correo completo, clave ni token.
- Hallazgo visual reparado en el candidato: el correo recibido decia
  `sistema de motel` aunque la plantilla es global para cualquier tipo de
  empresa. El texto y HTML ahora dicen `plataforma Powerful Control System` y
  normalizan al renderizar la copia legada que pudiera seguir configurada en la
  base de datos.
- El login normal posterior del Cajero con `empresa_id=12` abrio el panel de PCS y
  mostro solo el menu operativo esperado. El acceso directo a administracion de
  usuarios fue rechazado por permiso efectivo, `/super/api/empresas` respondio
  403 y una solicitud con `empresa_id=11` devolvio
  `empresa_id fuera del alcance del usuario autenticado`.
- Las mismas credenciales fueron rechazadas al intentar iniciar sesion con la
  empresa 11 y aceptadas de nuevo con la empresa 12. Produccion mantiene en ese
  rechazo el texto visible corrupto `credenciales inv?lidas`; la correccion UTF-8
  ya forma parte de este candidato.
- El cierre de sesion redirigio al login y una visita posterior a la pagina
  protegida no pudo cargar datos: la API respondio `unauthorized`.
- `go test ./... -count=1`: los paquetes de autenticacion pasan; el barrido global
  conserva un fallo basal ajeno en
  `TestEditarEmpresaDeleteButtonRequiresSafeConfirmation`, relacionado con
  `editar_empresa.js`, archivo no modificado por este candidato.

## Evidencia real pendiente

El primer establecimiento de clave de la invitacion operativa ya fue completado.
La entrega final del flujo independiente de recuperacion de clave queda en manos
del usuario por seguridad del navegador. Tambien falta desplegar el candidato en
staging para repetir alli el selector reparado y la matriz completa.

Antes de desplegar se deben ejecutar las dos migraciones `super`, configurar
completamente reCAPTCHA si esta marcado como activo y repetir la matriz real en
staging equivalente. Hasta entonces el estado es NO-GO para produccion general.

## Incidente y recuperación operativa 2026-08-25

- Se reprodujo públicamente HTTP 503 en
  `POST /super/api/administradores/login` incluso con una cuenta inexistente;
  el fallo ocurría antes de validar credenciales y afectaba también registro y
  recuperación.
- `/health` y `/ready` respondían HTTP 200. `/config.js` confirmó reCAPTCHA
  solicitado pero no configurado. PostgreSQL conservaba la clave pública y una
  clave privada cifrada, pero el backend no podía descifrar esta última; las
  variables de entorno equivalentes estaban vacías.
- La recuperación reversible cambió solo `security.recaptcha.enabled` a `0` y
  registró el actor operativo. No se borró la clave cifrada, no se cambió una
  contraseña y no se desactivó el throttle durable.
- Después del cambio, el mismo probe devolvió HTTP 401 con
  `Credenciales inválidas.`; salud y readiness permanecieron correctos.
- La revisión visual publicada cubrió escritorio y móvil: login, error de
  credenciales, recuperación no enumerativa, formulario de nueva clave y
  registro administrativo sin desbordamiento horizontal, con acciones táctiles
  de 44 px en móvil.
- El candidato `codex/login-recaptcha-runtime-repair` impide activar una clave
  cifrada ilegible, valida las credenciales antes de persistir `enabled=1` y
  aplica la configuración compuesta atómicamente.

La clave privada válida de Google debe volver a ingresarse desde Super
administración y probarse contra Google antes de reactivar reCAPTCHA. No se
considera operativo solo porque exista un valor cifrado en la base de datos.
