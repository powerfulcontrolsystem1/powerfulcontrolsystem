> Estado documental: histórico. Línea base y plan de autenticación. Los pendientes siguen como evidencia/aceptación del módulo; no es una instrucción global de ejecución.
> Fuentes actuales: [contrato_autenticacion_administrativa_y_usuarios_empresa.md](gobernanza_tecnica/contratos/contrato_autenticacion_administrativa_y_usuarios_empresa.md).
> Se conserva como antecedente; no autoriza ejecutar acciones ni acredita el candidato actual.

# Plan de revisión y reparación del login para producción

Fecha de creación: 2026-08-21  
Estado inicial: **NO-GO / 0 % certificado**  
Modelo planificador solicitado: **GPT-5.6 Sol, razonamiento alto**  
Modelo ejecutor solicitado: **GPT-5.6 Terra, razonamiento medio**  
Repositorio: `D:\powerfulcontrolsystem`

## 1. Objetivo y límite de este documento

Revisar, reparar y certificar de punta a punta la autenticación administrativa y
la autenticación de usuarios de empresa de PCS antes de entrar en producción.
El alcance incluye registro, invitación, confirmación de correo, login por clave,
Google OAuth, contrato, 2FA, recuperación, cambio de contraseña, sesiones,
logout, roles, aislamiento por empresa, correo Mailu, interfaz, observabilidad,
concurrencia, despliegue y rollback.

Este archivo es **únicamente el plan**. Al crearlo no se autoriza:

- implementar correcciones funcionales;
- transmitir las credenciales suministradas por el propietario;
- crear, confirmar, modificar o desactivar cuentas reales;
- cambiar contraseñas, 2FA, roles, licencias o contratos;
- desplegar, ejecutar `rs`, aplicar migraciones o modificar producción.

La ejecución futura podrá usar las credenciales autorizadas y la empresa
`Powerful Control System` solo desde variables efímeras o el navegador, después
de confirmar el ambiente y el alcance inmediato. La contraseña no debe copiarse
a comandos, documentación, evidencias, capturas, logs ni commits. Nunca se debe
restablecer la contraseña de la cuenta principal durante QA; los flujos de
confirmación y recuperación se completarán con identidades QA dedicadas.

## 2. Fuentes revisadas para crear el plan

Se inventariaron los documentos del repositorio y se rastrearon todas las
referencias a login, registro, autenticación, sesión, recuperación, OAuth,
contratos, roles, correo y multiempresa. Se estudiaron como fuentes rectoras:

- `AGENTS.md`;
- `documentos/contexto_general_del_sistema.md`;
- `documentos/contexto_especifico_del_sistema.md`;
- `documentos/contexto_codex.md`;
- `documentos/mapa_modulos.md`;
- `documentos/flujos_operativos.md`;
- `documentos/comandos_codex.md`;
- `documentos/decisiones_tecnicas.md`;
- `documentos/checklist_seguridad_endpoint_multiempresa.md`;
- `documentos/matriz_roles_permisos_pos_multiempresa.md`;
- `documentos/email_corporativo_mailu.md`;
- `documentos/gobernanza_tecnica/contratos/contrato_autenticacion_administrativa_y_usuarios_empresa.md`.

El contrato vigente separa dos flujos públicos:

1. administrador: registro, confirmación, email/clave, Google, contrato, 2FA y
   selección de empresas;
2. usuario de empresa: invitación, primera clave, email/clave o Google y acceso
   limitado a una empresa y sus permisos efectivos.

Terra debe preservar esta separación y demostrarla en backend, base, sesión,
frontend, correo y pruebas. No basta con que ambas identidades compartan una
pantalla o que la autorización posterior parezca correcta.

## 3. Línea base observada al crear el plan

- Rama observada: `codex/domotica-activation-queue`.
- SHA observado: `76a43b971ca53cb6dfe6ccf642faa6e41781c7ef`.
- `origin/main` observado: `0f00b2ef6e3feb43393754e43bec30fb21a026e9`.
- El árbol principal contiene cambios del propietario en páginas administrativas
  y CSS. No se deben mezclar, descartar ni atribuir a este plan.
- `go test ./handlers ./db ./utils`: PASS.
- `go vet ./handlers ./db ./utils`: PASS.
- `go test -race` no se pudo ejecutar en el Windows actual porque CGO y un
  compilador C no están disponibles. Debe ejecutarse en CI/Linux.
- `/health` y `/ready` públicos respondieron 200.
- API, worker, PostgreSQL y servicios Mailu observados en el VPS estaban activos;
  la cola SMTP estaba vacía. Salud no equivale a entrega ni a login certificado.
- DNS observado: SPF, DKIM, DMARC de enforcement y MX presentes. Falta vincular
  una entrega real de confirmación/recuperación al candidato que se certifique.
- Configuración pública desplegada observada: reCAPTCHA fue solicitado, pero no
  estaba configurado ni habilitado; el bypass de desarrollo estaba apagado. El
  backend actualmente deja continuar cuando esa configuración está incompleta.
- El switch público de 2FA administrativo estaba desactivado.
- Login administrativo inválido fue rechazado con mensaje visible genérico.
- Login empresarial sin `empresa_id` fue aceptado por el contrato HTTP y llegó a
  resolver el correo globalmente; esto contradice el contrato multiempresa.
- La vista administrativa, registro y login de usuario cargaron sin errores de
  consola en el recorrido público observado.
- En viewport móvil real no hubo overflow horizontal en login administrativo.
  El formulario se vio ordenado, pero conserva un correo recordado en
  `localStorage`; privacidad, lector de pantalla, zoom, contraste y teclado aún
  no están certificados.

Estas observaciones no autorizan producción. Las pruebas reales con la cuenta
principal y la cuenta QA quedaron deliberadamente pendientes para la ejecución
del plan.

## 4. Hallazgos técnicos confirmados en el candidato observado

### 4.1 P0 — Protección de contraseñas insuficiente

`backend/handlers/usuarios_empresa.go` genera y verifica las contraseñas de
administradores y usuarios empresariales con un solo SHA-256 de `salt:password`.
Es una función rápida y no una KDF adecuada para contraseñas. La comparación
también usa igualdad normal de cadenas.

Corrección requerida:

- adoptar un formato versionado de hash con algoritmo, parámetros y salt;
- preferir Argon2id mediante `golang.org/x/crypto`, ya fijado en el módulo, solo
  después de la autorización técnica exigida por `AGENTS.md` para un import de
  terceros nuevo; documentar ADR, impacto y alternativa;
- si esa autorización no existe, detener esta subfase y presentar una opción
  compatible con biblioteca estándar y vectores oficiales; no improvisar
  criptografía sin revisión;
- calibrar costo con benchmark de staging y límites contra DoS;
- comparar en tiempo constante;
- verificar hashes antiguos y migrarlos oportunistamente en el login correcto,
  sin invalidar cuentas ni guardar claves en texto;
- permitir elevar parámetros en el futuro y probar rollback de la migración.

### 4.2 P0 — Falta aislamiento obligatorio en login de usuario de empresa

`resolveEmpresaUsuarioForPasswordLogin` busca globalmente por correo cuando no
recibe `empresa_id` y puede escoger una cuenta comparando claves entre empresas.
Recuperación, reenvío de invitación, reset y cambio de clave admiten rutas
equivalentes sin alcance empresarial. `WithEmpresaPublicScope` deja pasar la
solicitud cuando el ID falta.

Corrección requerida:

- exigir `empresa_id > 0` en todos los flujos de usuario de empresa, o derivarlo
  exclusivamente de una invitación firmada o un host empresarial validado;
- no resolver tenant por coincidencia global de correo o contraseña;
- validar igualdad de empresa en URL, query, header, body, token e identidad;
- usar claves/índices compuestos por `empresa_id` e ID secundario;
- devolver error público estable sin informar si el correo existe en otra
  empresa;
- añadir pruebas A/B donde la misma dirección exista en dos empresas con claves,
  estados, roles y tokens diferentes.

### 4.3 P0 — Identidad de sesión administrativa y empresarial mezclada

Al iniciar un usuario de empresa, `createEmpresaUsuarioSession` ejecuta
`UpsertAdministrador`, crea acceso compartido y emite una sesión global por
correo. Si el rol falta o se normaliza como `sin_rol`, lo convierte en
`admin_empresa`. Esto puede fabricar una identidad administrativa o elevar una
cuenta mal configurada.

Corrección requerida:

- modelar un principal tipado: `admin` o `empresa_usuario`;
- persistir en `sesiones` tipo, sujeto, `empresa_id` cuando corresponda, versión
  de seguridad y timestamps, sin depender solo del correo;
- el usuario empresarial debe existir y autorizarse desde su registro canónico
  en `users`, sin crear ni actualizar silenciosamente `administradores`;
- un rol ausente, desconocido, inactivo o sin licencia debe fallar cerrado;
- revalidar estado, rol, permisos, licencia, empresa y versión de sesión en cada
  solicitud sensible;
- migrar con doble lectura acotada, telemetría y revocación segura de sesiones
  legacy. No mantener dos fuentes de privilegios indefinidamente.

### 4.4 P0 — Estados y contrato no se aplican uniformemente

`AdminLoginHandler` confirma correo, clave y 2FA, pero no bloquea explícitamente
`estado=inactivo` ni valida la aceptación de la versión contractual vigente.
Además calcula una redirección propia en vez de reutilizar
`resolveAdminPostLoginRedirect`. El flujo Google sí consulta contrato, por lo
que email/clave y OAuth pueden producir decisiones distintas.

Corrección requerida:

- crear una única política server-side de elegibilidad postautenticación;
- comprobar estado, correo, contraseña configurada, contrato vigente, 2FA y rol;
- producir transiciones explícitas: confirmar correo, aceptar contrato,
  establecer clave, completar 2FA, seleccionar empresa o entrar al panel;
- aplicar la misma política a email/clave, Google, reset y aceptación;
- revocar sesiones al inactivar, cambiar rol/empresa, cambiar contrato de forma
  incompatible o modificar factores de autenticación.

### 4.5 P0 — DDL y reparación de datos desde rutas operativas

`EnsureEmpresaUsuariosAuthSchema` contiene CREATE/ALTER/índices y una eliminación
de usuario reservado. Se invoca desde numerosos métodos de repositorio usados
por login, invitación y recuperación. `GetAdminByEmailFull`, cambios de clave y
2FA también intentan `EnsureAdministradoresAuthSchema` ante columnas faltantes.

Corrección requerida:

- trasladar todo cambio de esquema y reparación de datos a migraciones
  PostgreSQL inmutables, catalogadas y con checksum;
- API y worker solo verifican readiness y fallan cerrado con error saneado;
- la eliminación o conciliación de cuentas reservadas debe ser migración
  explícita, auditable y con reporte previo, nunca efecto colateral del login;
- ejecutar inventario de `Ensure*` y demostrar cero DDL/DML de reparación en
  handlers o repositorios de runtime;
- probar base vacía, upgrade, réplica y drift intencional.

### 4.6 P0 — Recuperación y cambio de clave no revocan igual las sesiones

El reset administrativo revoca sesiones; el reset y cambio empresarial guardan
la clave y crean otra sesión sin revocar primero todas las anteriores. Una sesión
robada podría sobrevivir al cambio de credenciales.

Corrección requerida:

- realizar cambio de hash, consumo de token, incremento de versión de seguridad
  y revocación en una transacción;
- emitir como máximo una sesión de reemplazo después del commit;
- aplicar la regla a primer ingreso, reset, cambio de clave, cambio de correo,
  2FA, inactivación y rol;
- si falla la sesión nueva, la credencial queda segura y el usuario vuelve a
  autenticarse; nunca restaurar sesiones antiguas.

### 4.7 P0 — Protección contra abuso incompleta

Los usuarios empresariales tienen contador/bloqueo, pero el login administrativo
no tiene protección persistente equivalente. reCAPTCHA falla abierto cuando fue
solicitado y quedó sin credenciales. Registro y recuperación necesitan límites
contra spam y enumeración.

Corrección requerida:

- limitar por cuenta normalizada, IP efectiva confiable, prefijo de red, acción
  y ventana usando PostgreSQL/infra compartida entre réplicas;
- aplicar backoff y bloqueo temporal sin permitir bloqueo permanente provocado
  por terceros;
- conservar mensajes y tiempos públicos equivalentes;
- configurar reCAPTCHA completamente o desactivarlo de manera consciente; si
  está solicitado en producción y falta configuración, readiness debe fallar;
- negar bypass en staging equivalente y producción;
- métricas y alertas sin almacenar claves, tokens ni correos completos.

### 4.8 P1 — Enumeración y respuestas inconsistentes

La recuperación administrativa devuelve `email_sent` distinto para una cuenta
inexistente, confirmada o con error SMTP. Recuperación e invitación empresarial
exponen `delivery=masked|manual|email`. El registro administrativo devuelve
conflicto explícito si el correo ya está confirmado.

Corrección requerida:

- cuerpo, estado, tamaño aproximado y latencia pública equivalentes;
- enviar el resultado real solo a auditoría/métricas internas;
- definir taxonomía JSON uniforme con `request_id` y mensajes UTF-8;
- no concatenar errores de proveedor, SQL, rutas o destinatarios;
- revisar si el registro público debe responder siempre de forma no enumerativa
  y notificar de manera segura a la cuenta ya existente.

### 4.9 P1 — Correo de autenticación síncrono y duplicado

Los nombres `sendEmpresaUsuarioGmail*` son engañosos: actualmente terminan en
Mailu. Existe un fallback que vuelve a intentar esencialmente el mismo transporte
y rutas capaces de ejecutar `docker exec/sendmail`, contrario a la separación
operativa documentada.

Corrección requerida:

- consolidar un solo servicio de correo Mailu con contratos tipados;
- registrar confirmación, invitación y recuperación en `pcs_outbox_events` dentro
  de la transacción que crea el token;
- procesar con `pcs-worker`, lease, reintento, idempotencia y dead-letter;
- eliminar Docker CLI/socket del backend;
- no duplicar mensajes por doble clic, timeout o dos réplicas;
- observar accepted, delivered/queued, bounced y dead sin exponer destinatario.

### 4.10 P1 — Tokens y expiraciones requieren cierre transaccional

El reset administrativo acepta una expiración no vacía que no pueda parsearse,
porque solo rechaza si el parseo fue exitoso y la fecha pasó. Los tokens se
guardan como verificadores, lo cual es positivo, pero deben consumirse de forma
atómica ante solicitudes simultáneas.

Corrección requerida:

- fecha ausente, inválida o vencida siempre falla cerrado;
- token de un solo uso, con propósito, sujeto, tenant, versión y TTL;
- consumo mediante UPDATE condicional/RowsAffected dentro de la transacción de
  cambio de clave o confirmación;
- dos solicitudes simultáneas: una sola gana; la otra recibe respuesta estable;
- no incluir tokens en logs, Referer, analytics, capturas ni evidencia.

### 4.11 P1 — UX, privacidad y accesibilidad incompletas

La pantalla guarda el correo recordado en `localStorage`, los mensajes de login,
recuperación y reset no tienen región viva, y existen textos empresariales con
mojibake como `inv?lido` o `confirmaci?n`. El login de empresa comunica que se
puede entrar sin empresa, reproduciendo el contrato inseguro del backend.

Corrección requerida:

- “Recordar usuario” desmarcado por defecto y explicación de que solo guarda el
  correo, nunca la clave; limpiar en logout si el usuario lo solicita;
- `aria-live`, roles, foco al primer error, labels y estados busy accesibles;
- teclado, lector de pantalla, contraste, zoom 200/400 %, movimiento reducido y
  viewports móvil/tableta/escritorio;
- mensajes UTF-8 y contenido consistente con el alcance empresarial obligatorio;
- no mostrar si una cuenta, tenant, 2FA o invitación existe.

## 5. Arquitectura objetivo

### 5.1 Servicio de autenticación único, principales separados

Terra debe extraer una capa interna de autenticación sin crear un segundo módulo
paralelo. Debe haber funciones reutilizables para:

- normalización y validación de identidad;
- verificación/migración de hash;
- política de elegibilidad;
- bloqueo y protección contra abuso;
- emisión, rotación y revocación de sesión;
- token de un solo uso;
- evento de correo outbox;
- auditoría saneada.

Los handlers HTTP solo validan contrato, llaman al servicio y traducen resultados
a respuestas públicas. Las reglas no se duplican entre admin, empresa, Google,
reset o primer ingreso.

### 5.2 Persistencia propuesta

La fase de diseño debe producir una migración revisable que, como mínimo,
contemple:

- versión/algoritmo/parámetros de contraseña;
- principal tipado en sesión, sujeto, empresa opcional y versión de seguridad;
- intentos/bloqueo administrativo y empresarial con timestamps PostgreSQL;
- tokens con hash, propósito, expiración y consumo atómico;
- auditoría de autenticación sin secretos;
- claves idempotentes de correo y outbox.

No se reescriben migraciones aplicadas. No se cambia de PostgreSQL. Todo índice y
FK que involucre usuario empresarial conserva `empresa_id`.

### 5.3 Contrato público

Todos los endpoints deben:

- limitar tamaño y rechazar JSON desconocido o múltiple;
- normalizar correo de forma consistente sin alterar direcciones válidas;
- usar respuestas JSON UTF-8 estables y `Cache-Control: no-store`;
- no revelar cuenta, tenant, estado, factor, correo ni proveedor;
- usar cookies `Secure`, `HttpOnly`, path `/`, SameSite evaluado y expiración
  server-side;
- rotar CSRF tras login/cambio de credenciales y exigirlo en mutaciones con
  cookie;
- rechazar método incorrecto y orígenes inconsistentes.

## 6. Reglas obligatorias para GPT-5.6 Terra medio

1. Usar `gpt-5.6-terra` con razonamiento `medium`.
2. Leer completos `AGENTS.md`, los documentos rectores de la sección 2, este
   plan y la evidencia de la fase antes de editar.
3. Crear un worktree limpio `codex/login-*` desde el candidato aprobado. No
   modificar ni descartar los cambios ajenos del árbol principal.
4. Fijar SHA, digest, ambiente y base antes de cada evidencia. Un timeout es
   inconcluso y una prueba local no recibe crédito de staging.
5. No agregar dependencia/import de terceros ni cambiar `go.mod` sin autorización
   explícita y registro técnico. La selección de KDF es una compuerta P0.
6. Mantener PostgreSQL, migraciones propietarias y cero DDL en API/worker.
7. Mantener `empresa_id` server-side en usuario, sesión, token, correo, cache,
   auditoría y job. Nunca confiar solo en hidden input/localStorage.
8. No usar SQL directo para simular registro, confirmación, login o reset real.
   Los efectos QA se producen por UI/API oficial; SQL de producción solo lectura.
9. No imprimir credenciales ni tokens. Sanear capturas y evidencia antes de
   guardarlas.
10. Corregir causa + prueba de regresión + documentación en el mismo bloque.
11. No hacer PR por cada archivo. Agrupar por fase coherente y no desplegar hasta
    que CI, migración y rollback del candidato estén aprobados.
12. Ante fuga multiempresa, sesión residual, elevación de rol o debilidad de
    credenciales, mantener NO-GO y continuar solo frentes independientes seguros.

## 7. Evidencia y cálculo de avance

Cada fase guarda evidencia en:

`documentos/evidencia_login/<ID>/<fecha>_<descripcion>.md`

Debe registrar: SHA/digest, ambiente, rutas, tablas, rol, empresa, identidad QA
saneada, precondición, datos/efectos, comandos, resultados, capturas saneadas,
logs/metricas, rollback, limpieza y riesgos.

Estados: `pendiente`, `parcial`, `aprobado`, `bloqueado`, `fallido`.

- pendiente/bloqueado/fallido: 0 %;
- parcial: 50 % solo con evidencia reproducible del mismo candidato;
- aprobado: 100 % con todos los criterios de la fase.

Se reportan implementación y certificación por separado. Certificación solo
cuenta fases aprobadas sobre el mismo SHA/digest en staging equivalente.

## 8. Fases de ejecución

### LOGIN-000 — Candidato, contrato e inventario [P0]

Acciones:

1. Crear worktree limpio y fijar SHA, estado CI y versión desplegada.
2. Inventariar ruta/método → handler → servicio → función DB → tabla → página →
   correo → permiso → auditoría para admin y usuario empresarial.
3. Inventariar cookies, local/sessionStorage, caches, goroutines, jobs, OAuth,
   TOTP, contratos, proxies y configuración por ambiente.
4. Crear matriz de estados y transiciones válidas; reconciliarla con el contrato
   documental y marcar código huérfano o duplicado.
5. Diseñar identidades QA, prefijo, responsable, retención y limpieza por flujo
   oficial, sin registrar secretos.

Aceptación: inventario completo, árbol limpio, candidato inmutable y ninguna
ruta/control/tabla fuera de clasificación. Sin LOGIN-000, el avance es 0 %.

### LOGIN-001 — Migraciones y readiness sin DDL runtime [P0]

Acciones:

1. Extraer `EnsureEmpresaUsuariosAuthSchema`,
   `EnsureAdministradoresAuthSchema` y reparaciones asociadas a migraciones
   nuevas, inmutables y catalogadas.
2. Ensayar base vacía y upgrade de snapshot anonimizado en PostgreSQL aislado.
3. Implementar comprobación de esquema de solo lectura para API y worker.
4. Probar drift de columna/índice: API falla cerrado, migrador corrige y rollback
   documentado restaura el candidato anterior.
5. Verificar que login concurrente no ejecuta CREATE, ALTER, DROP ni DELETE de
   conciliación.

Aceptación: migración y rollback PASS; cero DDL/DML reparador en request path;
readiness bloquea esquema incompatible sin fuga de SQL.

### LOGIN-002 — KDF y migración de contraseñas [P0]

Acciones:

1. Aprobar ADR y autorización del algoritmo/import antes de codificar.
2. Crear hash versionado, parámetros calibrados, salt aleatorio y verificación
   en tiempo constante.
3. Mantener lectura del SHA-256 legacy solo para migración; al login correcto,
   rehash y actualización atómica. Nunca rehash en intento fallido.
4. Aplicar a registro admin, invitación empresa, Google con clave, reset y cambio.
5. Probar vectores, Unicode, límites, clave larga, parámetros inválidos, hash
   corrupto, downgrade, replay y benchmark/DoS.
6. Publicar contador de hashes legacy sin correos y plan de retiro.

Aceptación: cuentas nuevas usan KDF aprobada; legacy migra sin perder acceso;
comparación constante; cero texto claro; rendimiento dentro del objetivo medido.

### LOGIN-003 — Tenant e identidad de sesión [P0]

Acciones:

1. Hacer obligatorio el alcance empresarial en login, invitación, confirmación,
   recuperación, reset, cambio, Google y primera clave.
2. Crear principal de sesión tipado y eliminar el `UpsertAdministrador` desde el
   login empresarial.
3. Rechazar rol ausente/desconocido/inactivo; no usar `admin_empresa` como
   fallback de privilegio.
4. Validar sujeto y empresa en middleware, permisos, cache y selector.
5. Migrar/revocar sesiones legacy de forma compatible y auditable.
6. Probar misma dirección en empresas A y B, IDs cruzados, token A en B, sesión
   admin usada como empresa y viceversa.

Aceptación: cero resolución global por correo/clave; cero elevación; A/B devuelve
403/404 estable y ninguna consulta, sesión, job o cache mezcla empresas.

### LOGIN-004 — Política de elegibilidad y contratos [P0]

Acciones:

1. Implementar una política compartida para email/clave y Google.
2. Cubrir estado, correo confirmado, clave, contrato vigente, 2FA, rol, empresa,
   licencia e invitación.
3. Reutilizar una sola decisión de redirección server-side.
4. Definir aceptación de contrato con versión, timestamp y evidencia del actor;
   no confiar en cookie o checkbox sin persistencia.
5. Revocar sesiones ante inactivación, cambio de rol/empresa/contrato/factor.

Aceptación: matriz completa de estados produce la misma decisión por ambos
métodos; una cuenta inactiva o sin contrato vigente nunca recibe sesión útil.

### LOGIN-005 — Tokens, recuperación y revocación [P0]

Acciones:

1. Unificar generación, hash, TTL, propósito y consumo atómico de confirmación,
   invitación y recuperación.
2. Rechazar expiración ausente, inválida o vencida.
3. Cambiar credencial + consumir token + incrementar versión + revocar sesiones
   en una transacción; emitir reemplazo después del commit.
4. Probar doble clic, 20 consumidores simultáneos, token reutilizado, token de
   otro flujo, empresa o usuario, y fallo antes/después del commit.
5. Impedir que URL/token termine en logs, Referer, analytics o captura.

Aceptación: un solo consumidor gana; todas las sesiones anteriores dejan de
autorizar; ningún token válido se filtra o puede cambiar otro sujeto.

### LOGIN-006 — Abuso, enumeración, reCAPTCHA y 2FA [P0]

Acciones:

1. Implementar límites persistentes por acción, cuenta e IP confiable para
   registro, login, confirmación, invitación y recuperación.
2. Igualar cuerpos y tiempos de cuenta existente/inexistente/inactiva/no
   confirmada/SMTP fallido, dentro de tolerancias documentadas.
3. Hacer que readiness falle si reCAPTCHA se solicita pero está incompleto en
   staging equivalente o producción.
4. Probar proveedor reCAPTCHA caído, token inválido/reusado, proxy falso y
   configuración incompleta; no habilitar bypass.
5. Habilitar y probar 2FA para super administrador y roles privilegiados según
   política aprobada: alta, confirmación, login, códigos de recuperación,
   replay, desactivación y sesión revocada.
6. Alertar ataques y bloqueos sin PII completa.

Aceptación: fuerza bruta/spam limitados entre dos réplicas; cero enumeración
observable; reCAPTCHA y 2FA operan fail-closed según política.

### LOGIN-007 — Mailu, outbox e idempotencia [P0]

Acciones:

1. Consolidar confirmación, invitación, recuperación y alerta de seguridad en un
   servicio Mailu único; retirar nombres Gmail y Docker CLI del backend.
2. Insertar evento outbox con el token en la misma transacción de negocio; el
   payload durable debe proteger secretos y cumplir retención.
3. Añadir handlers idempotentes al worker con lease, backoff, máximo, dead-letter
   y recuperación auditada.
4. Probar mensaje limpio, HTML/texto, logo, enlaces canónicos, UTF-8, destinatario,
   SPF/DKIM/DMARC, rebote y no duplicado.
5. Simular Mailu caído, timeout después de aceptar, worker reiniciado y dos
   workers; recuperar sin volver a crear token ni enviar dos veces.

Aceptación: flujo web no depende de SMTP síncrono; un evento produce como máximo
un mensaje lógico; fallos son observables/recuperables y no filtran destinatario.

### LOGIN-008 — Frontend, responsive y accesibilidad [P1]

Acciones:

1. Corregir textos UTF-8 y mensajes consistentes; eliminar la promesa de login
   empresarial sin tenant.
2. Hacer “Recordar usuario” opt-in, guardar solo correo y ofrecer limpieza.
3. Añadir regiones vivas, foco, labels, estados busy, errores asociados y
   navegación de teclado.
4. Probar 320/390/768/1024/1440 px, landscape, zoom 200/400 %, claro/oscuro,
   contraste, lector de pantalla y movimiento reducido.
5. Probar clave visible/oculta, Enter, doble clic, atrás/adelante, recarga,
   offline, expiración y consola/red limpias.
6. Revisar login, registro, recuperación, reset, contrato, Google, 2FA, selector
   de empresa, primer ingreso y logout.

Aceptación: 100 % de controles PASS o exclusión firmada; cero overflow crítico,
error de consola, foco perdido, texto corrupto o estado que revele una cuenta.

### LOGIN-009 — Pruebas automáticas y seguridad dinámica [P0]

Acciones:

1. Añadir unitarias de handlers/servicio y pruebas PostgreSQL reales de
   transacciones, locks, tokens, sesiones y tenant A/B.
2. Ejecutar `go test`, `go vet`, build y `go test -race` en Linux/CI.
3. Probar CSRF, CORS/origin, cookies, headers/no-store, body limit, JSON hostil,
   Unicode, open redirect, session fixation, OAuth state/PKCE y replay TOTP.
4. Ejecutar DAST seguro no autenticado y autenticado contra staging aislado;
   no hacer fuerza bruta contra producción.
5. Cubrir 0/1/muchas empresas, roles, estado, contrato, licencia, concurrencia y
   dos réplicas.

Aceptación: suites estables sin dependencia de orden; race PASS; cero P0/P1 de
seguridad; cada hallazgo confirmado tiene regresión automática.

### LOGIN-010 — Carga, fallos y observabilidad [P0]

Acciones:

1. Medir p50/p95/p99 y error rate de login correcto/incorrecto, sesión y
   recuperación con dataset QA y hashes nuevos/legacy.
2. Ejecutar concurrencia autenticada realista sin convertirla en ataque; separar
   ensayo de rate limit del ensayo de capacidad.
3. Probar dos réplicas sin sticky session, reinicio API/worker, DB lenta/caída,
   Mailu caído, reloj y lease expirado.
4. Añadir métricas de éxito/fallo por causa interna saneada, lockout, migración
   de hash, sesión revocada, outbox y latencia.
5. Configurar alertas y runbook de recuperación sin secretos.

Aceptación: SLO vigente cumplido, sesiones sobreviven únicamente cuando deben,
rate limit es consistente y fallos complejos se detectan/recuperan sin elevar
privilegios ni perder trazabilidad.

### LOGIN-011 — Pruebas reales controladas en PCS [P0]

Precondiciones:

- candidato exacto aprobado hasta LOGIN-010 y desplegado solo en staging
  equivalente;
- producción intacta y ambiente visible en cada captura;
- confirmación inmediata antes de transmitir credenciales o crear cuenta;
- correo principal nunca cambia de clave ni 2FA durante QA;
- identidad QA con dirección alias o buzón corporativo controlado, clave temporal
  generada y plan de desactivación.

Recorrido administrativo real:

1. Login correcto con la cuenta autorizada, redirección, empresa PCS, cookie,
   sesión, logout y rechazo posterior del token.
2. Clave incorrecta, rate limit, cuenta recordada opt-in y mensaje no
   enumerativo; no bloquear la cuenta principal deliberadamente.
3. Crear un administrador QA por formulario oficial, recibir correo real Mailu,
   comprobar remitente/logo/enlace, confirmar y hacer primer login.
4. Recuperar la cuenta QA, comprobar correo, cambiar clave, verificar revocación
   de sesiones anteriores y rechazo del token reutilizado.
5. Aceptar contrato y probar 2FA en la cuenta QA si la política lo exige.
6. Probar Google OAuth hasta donde no aparezca CAPTCHA/consentimiento; si aparece,
   entregar al usuario sin eludirlo.

Recorrido empresarial real:

1. Crear/invitar desde PCS un usuario QA con rol mínimo por UI oficial.
2. Recibir invitación, aceptar contrato, establecer clave e ingresar con
   `empresa_id` explícito.
3. Ver solo PCS y solo módulos del rol; intentar empresa B e IDs cruzados.
4. Cambiar/recuperar clave y comprobar revocación; inactivar y confirmar que toda
   sesión queda inválida.
5. Repetir con la misma dirección QA en empresa B para demostrar aislamiento.

Evidencia y limpieza:

- capturas escritorio/móvil saneadas, consola/red, Set-Cookie sin valor, auditoría,
  métricas, Mailu y outbox;
- lista de objetos QA creados y estado final;
- desactivar identidades QA y revocar sesiones mediante flujos oficiales;
- no borrar evidencia fiscal/operativa ni usar SQL para ocultar efectos.

Aceptación: ambos recorridos PASS en el mismo digest, correo llega y se usa,
A/B PASS, revocación PASS, datos QA conciliados y cuenta principal intacta.

### LOGIN-012 — Release, rollback y GO/NO-GO [P0]

Acciones:

1. Ejecutar preflight completo, CI, SBOM/Trivy, migración/rollback, restore y
   pruebas del digest final.
2. Publicar matriz de rutas, estados, roles, tenants, correos y riesgos; aprobar
   runbooks de cuenta bloqueada, Mailu, OAuth, 2FA, DB y rollback.
3. Hacer backup y restore aislado de tablas/sesiones/outbox; medir RTO/RPO.
4. Promover exactamente el digest probado solo con autorización explícita;
   ejecutar smoke de admin y usuario empresa sin efectos destructivos.
5. Observar ventana acordada y revertir ante enumeración, elevación, fuga A/B,
   sesiones no revocadas, fallos de correo sostenidos, migración o SLO.

Aceptación: LOGIN-000 a LOGIN-012 aprobadas en el mismo candidato, cero P0/P1
abierto, rollback ensayado y decisión humana GO registrada. Una sola compuerta
P0 pendiente mantiene **NO-GO**.

## 9. Matrices obligatorias que Terra debe mantener

1. **Superficie:** ruta, método, público/protegido, handler, servicio, DB, tabla,
   UI, correo y auditoría.
2. **Estado:** tipo de cuenta, confirmación, estado, clave, contrato, 2FA, rol,
   empresa, licencia, resultado y redirección.
3. **Tenant/rol:** principal, empresa A/B, ID secundario, permiso efectivo,
   resultado esperado/observado.
4. **Sesión:** emisión, cookie, CSRF, rotación, expiración, revocación, cambio de
   rol/clave/2FA/estado y dos réplicas.
5. **Tokens:** propósito, TTL, tenant, sujeto, consumo, replay, concurrencia y
   resultado.
6. **Correo:** evento, idempotencia, destinatario saneado, Mailu, SPF/DKIM/DMARC,
   cola, entrega/rebote, retry/dead-letter y duplicados.
7. **UI:** página/control, viewport, teclado, lector, foco, contraste, consola,
   red y resultado.
8. **Release:** SHA/digest, CI, migraciones, restore, staging, PCS real, rollback
   y firma.

## 10. Archivos iniciales que Terra debe revisar

Backend y datos:

- `backend/handlers/auth_admin_handlers.go`;
- `backend/handlers/usuarios_empresa.go`;
- `backend/handlers/accept_handlers.go`;
- `backend/handlers/account_handlers.go`;
- `backend/handlers/admin_totp_handlers.go`;
- `backend/handlers/recaptcha.go`;
- `backend/handlers/mail_utils.go`;
- `backend/handlers/empresa_permisos.go`;
- `backend/db/db.go`;
- `backend/db/usuarios_empresa.go`;
- `backend/db/contrato_super.go`;
- `backend/db/empresa_admin_compartida.go`;
- `backend/db/outbox.go`;
- `backend/internal/platform/outbox/dispatcher.go`;
- `backend/cmd/pcs-worker/main.go`;
- `backend/utils/utils.go`;
- `backend/main.go`.

Frontend:

- `web/login.html`, `web/js/login.js`;
- `web/login_usuario.html`, `web/js/login_usuario.js`;
- `web/registrar_nuevo_usuario_administrador.html`;
- `web/js/registrar_nuevo_usuario_administrador.js`;
- páginas de aceptar contrato, establecer/restablecer/cambiar clave, selector de
  empresa y cuenta/2FA relacionadas por el inventario de LOGIN-000.

Pruebas existentes a ampliar, no reemplazar:

- `backend/handlers/auth_admin_registration_test.go`;
- `backend/handlers/auth_oauth_state_test.go`;
- `backend/handlers/admin_totp_handlers_test.go`;
- `backend/handlers/account_handlers_security_test.go`;
- `backend/handlers/usuarios_empresa_invitation_test.go`;
- `backend/handlers/usuarios_empresa_session_revocation_test.go`;
- `backend/handlers/usuarios_empresa_mailu_test.go`;
- `backend/db/security_tokens_test.go`;
- `backend/db/session_security_test.go`;
- `backend/db/usuarios_empresa_lookup_test.go`;
- `backend/utils/auth_middleware_test.go`;
- `backend/utils/csrf_test.go`;
- `backend/utils/security_headers_test.go`.

## 11. Comandos previstos para la futura ejecución

Estos comandos **no forman parte de la creación del plan**. Terra los adapta al
worktree limpio y al runtime Node documentado:

```powershell
Set-Location D:\powerfulcontrolsystem\backend
go test ./handlers ./db ./utils -count=1
go test ./... -run "Auth|Login|Session|Password|Usuario|OAuth|TOTP|CSRF|Recaptcha" -count=1
go vet ./...

Set-Location D:\powerfulcontrolsystem
& $node tools\runtime_ensure_inventory.mjs --check
& $node tools\migration_audit.mjs --strict
git diff --check
```

La carrera se ejecuta en CI/Linux con CGO disponible. Las pruebas PostgreSQL
usan recursos efímeros o staging aislado. Las credenciales reales solo se
inyectan en memoria/navegador en LOGIN-011 y nunca se imprimen. `rs` queda fuera
hasta LOGIN-012 y requiere autorización explícita.

## 12. Orden de implementación y porcentaje inicial

Orden obligatorio:

1. LOGIN-000 y LOGIN-001.
2. LOGIN-002 y LOGIN-003.
3. LOGIN-004 a LOGIN-007.
4. LOGIN-008 a LOGIN-010.
5. LOGIN-011 y, finalmente, LOGIN-012.

Hay 13 fases de igual peso. La implementación y la certificación iniciales son
**0 %** porque este documento no implementa ni certifica correcciones. El estado
del módulo es **NO-GO** por los P0 confirmados. Terra no debe ejecutar pruebas
reales con credenciales, crear usuarios ni desplegar hasta aprobar las compuertas
anteriores y recibir la confirmación inmediata correspondiente.
