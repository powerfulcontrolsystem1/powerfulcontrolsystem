# Nextcloud empresarial

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- La ruta /api/empresa/nextcloud continúa registrada. La existencia de nextcloud_decommission.go no demuestra que el módulo esté retirado en runtime.
- PCS solo admite PostgreSQL; el motor mencionado en el diagnóstico antiguo es un antecedente externo, no una recomendación para nuevos entornos.
- Los nombres de volúmenes son históricos: inventariar los efectivos y obtener backup consistente de BD/archivos; copiar un volumen activo no prueba restauración.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Actualizacion 2026-09-14 - lienzo completo dentro de Administrar empresa

`Administrar empresa > Nextcloud` conserva el menu y las acciones globales del
shell empresarial, pero dedica todo el panel de contenido al iframe de
Nextcloud. La subpagina ya no muestra encabezado, usuario tecnico, cuota,
estado, credencial ni botones de aprovisionamiento, restablecimiento,
activacion o apertura; esas operaciones siguen protegidas en backend y la
preparacion inicial se ejecuta automaticamente antes del autologin.

El iframe usa el alto y ancho completos del panel tanto en escritorio como en
movil. Cerrar sesion, cambiar la clave, configurar la apariencia y demas
acciones de la cuenta permanecen disponibles en el menu propio de Nextcloud.

## Actualizacion 2026-09-13 - inicio automatico empresarial seguro

`Administrar empresa > Nextcloud` emite un token HMAC de 45 segundos ligado
simultaneamente al `empresa_id` validado y al usuario tecnico
`pcs_empresa_<id>`. El componente propio `pcs_sso` de Nextcloud valida firma,
audiencia, vigencia, nonce y coincidencia exacta entre empresa y usuario antes
de iniciar la sesion. El token no contiene contrasenas y PCS no conserva la
clave de la cuenta empresarial.

El mismo componente agrega `https://powerfulcontrolsystem.com` como unico
ancestro externo permitido mediante la API CSP de Nextcloud. Asi se preserva la
politica dinamica y sus nonces. La pagina empresarial usa la URL de autologin al
cargar y tambien inmediatamente despues del aprovisionamiento; ya no marca
como exitosa una navegacion directa que termina en la pagina de login.

`NEXTCLOUD_SSO_SECRET` se genera automaticamente en el VPS y se comparte solo
entre el backend y la configuracion protegida del contenedor Nextcloud. El
script `scripts/configure_nextcloud_pcs_sso.sh` instala o actualiza el
componente, conserva un respaldo de una version previa, valida PHP, configura
el origen y habilita la aplicacion sin imprimir el secreto.

## Actualizacion 2026-07-24 - iframe restringido y version heredada

El shell PCS declara de forma explicita
`https://nextcloud.powerfulcontrolsystem.com` en `frame-src`. El antecedente de
modificar la CSP dinamica desde Nginx queda reemplazado por `pcs_sso`, que usa
la API CSP soportada de Nextcloud y conserva los nonces del proveedor. El
script `deploy/scripts/vps-configure-nextcloud-host-nginx.sh` mantiene su
rechazo seguro cuando detecta que no puede ampliar `frame-ancestors` sin
destruir esa politica.

El script valida que el sitio corresponde al dominio y al upstream
`127.0.0.1:8090`; no crea, recrea, actualiza ni elimina contenedores. Antes de
aplicarlo se debe tener backup verificable y despues comprobar visualmente el
contenido real dentro del iframe, las cabeceras, WebDAV y OCS.

La observación histórica del 2026-07-24 registró Nextcloud 29 con MariaDB, mientras que un
compose heredado no versionado difiere. Nextcloud 29 no es soporte vigente:
queda NO-GO para produccion general hasta reconciliar el compose real, probar
restauracion y actualizar de a una version mayor en staging. Esta restriccion
no autoriza migrar el motor de datos ni ejecutar `docker compose up`.

## Actualizacion 2026-07-21 - cuentas personales y acceso empresarial

La configuracion de Super administrador permite crear cuentas personales de
Nextcloud, con usuario y cuota elegidos por el super administrador. Esas cuotas
no modifican la cuota fija asignada a las empresas. La pagina muestra el enlace
HTTPS publico del VPS y una contrasena temporal una sola vez; no se registra ni
imprime esa contrasena en PCS.

Al abrir la pagina empresarial, PCS aprovisiona automaticamente la cuenta
tecnica si el servicio esta configurado y el espacio esta activo. La
autenticacion sigue siendo propia de Nextcloud: PCS no conserva contrasenas de
empresas. El acceso integrado usa el componente `pcs_sso` y un token firmado
de vida corta para seleccionar exclusivamente la cuenta tecnica ya
aprovisionada de la empresa autorizada.

Cuando la cuenta ya esta aprovisionada, `Administrar empresa > Nextcloud`
permanece dentro del panel derecho del shell y carga la vista de Nextcloud en
el mismo modulo integrado. Para una cuenta nueva, PCS la prepara en segundo
plano y abre el proveedor tan pronto queda lista. El usuario puede cerrar
sesion y cambiar contrasena desde el menu de perfil y Seguridad de Nextcloud.
La expiracion forzada de una contrasena temporal se configura en el servidor
Nextcloud mediante su politica de contrasenas, no en PCS, para que sea aplicada
por el mismo proveedor de identidad.

La incrustacion depende de que `pcs_sso` este habilitado y configurado con
`https://powerfulcontrolsystem.com` como origen. Si no hay secreto SSO, el shell
informa que el inicio automatico esta pendiente y conserva el enlace HTTPS para
abrir Nextcloud en pagina completa.

## Alcance

El servicio Nextcloud empresarial se ejecuta en el VPS principal y es distinto
del Nextcloud auxiliar administrado desde VPS2. PCS crea una cuenta tecnica por
empresa, aplica una cuota y conserva unicamente usuario, cuota y estado. La
contrasena se genera con 32 bytes aleatorios, se entrega una sola vez al
administrador autorizado y no se guarda en PCS.

Al activar el servicio desde Super administrador, PCS asigna automaticamente la
cuenta tecnica con cuota por defecto de 1024 MB a todas las empresas existentes
y a cada empresa nueva. La activacion global no depende del Nextcloud auxiliar
de VPS2. La cuenta remota se aprovisiona de forma idempotente con OCS.

La cuota se lee de la configuracion global de Nextcloud al asignar o actualizar
empresas; el arranque no debe reemplazarla por un valor fijo. La subpagina solo
carga el autologin cuando la cuenta de la empresa esta activa y fue
aprovisionada correctamente. Si el usuario abre la pagina dentro de Administrar
empresa, el identificador se obtiene del contexto protegido del shell y no de
un parametro manipulable como fuente de autoridad.

## Seguridad

- `/api/empresa/nextcloud` usa `WithEmpresaGestionDocumentalPermissions`.
- El handler toma `empresa_id` del contexto autenticado; no confia en JSON,
  cabeceras ni URL como autoridad independiente.
- `/super/api/config/nextcloud` usa `WithSuperAuditoria`.
- El secreto OCS queda cifrado mediante la configuracion segura existente.
- OCS exige HTTPS, TLS 1.2 o superior, timeout, respuesta JSON valida y estado
  interno 100/ok. El cliente no sigue redirecciones para no reenviar Basic Auth.
- Hosts privados requieren `PCS_NEXTCLOUD_ALLOW_PRIVATE_HOSTS=true`; se usa solo
  cuando la topologia privada esta documentada.
- Aprovisionamiento, restablecimiento y eliminacion dejan auditoria sin
  contrasenas.
- El rol empresarial requiere `gestion_documental:R` para consultar la pagina y
  `gestion_documental:C/U` para aprovisionar, activar o desactivar el espacio.
- Antes de eliminar una empresa, PCS elimina el usuario tecnico de Nextcloud
  mediante OCS; si el servicio no responde, la eliminacion se detiene para no
  dejar archivos remotos sin dueño.
- La misma eliminacion limpia tambien Mailu, OnlyOffice, uploads, documentos
  privados, backups y temporales asociados a `empresa_id`.

## Despliegue

1. Mantener el stack empresarial heredado del VPS bajo inventario y backup
   verificable; su compose no se versiona en este repositorio mientras se
   prepara la migracion controlada.
2. Definir dominio, reverse proxy, ruta de datos y proxies fiables en su entorno
   privado. No copiar secretos a este repositorio.
3. No sustituir volumenes, ni cambiar motor de datos, ni actualizar version mayor
   durante una reparacion. Primero se prueba la restauracion en staging.
4. Ejecutar `bash scripts/provision_nextcloud_service_account.sh <checkout-validado>`
   para crear o rotar la cuenta OCS exclusiva `pcs_ocs_service`. El script no
   imprime la contrasena; la registra como variable root-readable y el backend
   la cifra al arrancar.
5. Probar la conexion OCS y verificar el aprovisionamiento de dos empresas de
   ensayo.
6. Ejecutar `bash scripts/configure_nextcloud_pcs_sso.sh <checkout-validado>`,
   recrear backend con el entorno actualizado y comprobar el autologin dentro
   del panel empresarial.

El backend toma `NEXTCLOUD_ENABLED`, `NEXTCLOUD_BASE_URL`,
`NEXTCLOUD_ADMIN_USER`, `NEXTCLOUD_ADMIN_SECRET` y
`NEXTCLOUD_DEFAULT_QUOTA_MB` desde `deploy/.env.platform`. El acceso integrado
lee tambien `NEXTCLOUD_SSO_SECRET`. Nunca se debe usar la cuenta administrativa
inicial como credencial de integracion.

La instalacion heredada conserva su propio motor de datos, separado de PCS. No
reintroduce otro motor dentro del runtime de PCS: PostgreSQL sigue siendo el
unico motor de la aplicacion. Su migracion de infraestructura se planifica y
prueba por separado, con backup/restauracion verificable.

## Backup y restauracion

El backup debe incluir, en la misma ventana consistente:

- volumen `pcs_nextcloud_db`;
- volumen `pcs_nextcloud_html`;
- volumen `pcs_nextcloud_redis` como apoyo operativo;
- directorio configurado en `NEXTCLOUD_DATA_PATH`;
- `.env` y archivos de secretos mediante un canal cifrado separado.

Antes de actualizar una version mayor se prueba restauracion en staging. Las
actualizaciones de Nextcloud avanzan una version mayor por vez. No se eliminan
contenedores, volumenes ni datos sin evidencia de backup y restauracion.

## Validacion minima

- `go test ./handlers -run Nextcloud -count=1`;
- auditoria de permisos y `go test ./...`;
- validar con `docker compose config --quiet` usando el archivo y entorno privados del stack inventariado; `deploy/nextcloud/docker-compose.yml` no existe en este repositorio;
- healthchecks verdes de DB, Redis, Nextcloud y cron;
- prueba OCS desde Super administrador;
- aprovisionamiento, apertura, restablecimiento y rechazo cruzado entre dos
  empresas de staging;
- token SSO vencido o con empresa/usuario alterados rechazado, sesion correcta
  en el iframe y carga de archivo verificada visualmente;
- eliminacion de una empresa con cuenta Nextcloud, verificando que su usuario y
  archivos remotos desaparezcan sin afectar otra empresa.

## Fuentes y aceptación de la revisión

[main.go](../backend/main.go), [nextcloud.go](../backend/handlers/nextcloud.go), [nextcloud.go](../backend/db/nextcloud.go), [nextcloud_decommission.go](../backend/db/nextcloud_decommission.go), [provision_nextcloud_service_account.sh](../scripts/provision_nextcloud_service_account.sh).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
