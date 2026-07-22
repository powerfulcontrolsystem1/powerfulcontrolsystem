# Nextcloud empresarial

## Actualizacion 2026-07-21 - cuentas personales y acceso empresarial

La configuracion de Super administrador permite crear cuentas personales de
Nextcloud, con usuario y cuota elegidos por el super administrador. Esas cuotas
no modifican la cuota fija asignada a las empresas. La pagina muestra el enlace
HTTPS publico del VPS y una contrasena temporal una sola vez; no se registra ni
imprime esa contrasena en PCS.

Al abrir la pagina empresarial, PCS aprovisiona automaticamente la cuenta
tecnica si el servicio esta configurado y el espacio esta activo. La
autenticacion sigue siendo propia de Nextcloud:
PCS no conserva contrasenas de empresas ni fabrica cookies de sesion. Para
inicio de sesion unico real entre PCS y Nextcloud se requiere configurar un
proveedor SSO compatible (OIDC o SAML) en ambos servicios.

Cuando la cuenta ya esta aprovisionada, volver a abrir `Administrar empresa >
Nextcloud` navega directamente a la pagina principal de Nextcloud fuera del
iframe de PCS. Para una cuenta nueva, PCS muestra primero la contrasena temporal
una unica vez; en la siguiente apertura navega directamente. El usuario puede
cerrar sesion y cambiar contrasena desde el menu de perfil y Seguridad de
Nextcloud. La expiracion forzada de una contrasena temporal se configura en el
servidor Nextcloud mediante su politica de contrasenas, no en PCS, para que sea
aplicada por el mismo proveedor de identidad.

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
empresas; el arranque no debe reemplazarla por un valor fijo. La pagina solo
habilita Abrir y WebDAV cuando la cuenta de la empresa esta activa y fue
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
5. Ejecutar `bash scripts/provision_nextcloud_service_account.sh /root/powerfulcontrolsystem`
   para crear o rotar la cuenta OCS exclusiva `pcs_ocs_service`. El script no
   imprime la contrasena; la registra como variable root-readable y el backend
   la cifra al arrancar.
6. Probar la conexion OCS y verificar el aprovisionamiento de dos empresas de
   ensayo.

El backend toma `NEXTCLOUD_ENABLED`, `NEXTCLOUD_BASE_URL`,
`NEXTCLOUD_ADMIN_USER`, `NEXTCLOUD_ADMIN_SECRET` y
`NEXTCLOUD_DEFAULT_QUOTA_MB` desde `deploy/.env.platform`. Nunca se debe usar
la cuenta administrativa inicial como credencial de integracion.

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
- `docker compose --env-file deploy/nextcloud/.env -f deploy/nextcloud/docker-compose.yml config --quiet`;
- healthchecks verdes de DB, Redis, Nextcloud y cron;
- prueba OCS desde Super administrador;
- aprovisionamiento, apertura, restablecimiento y rechazo cruzado entre dos
  empresas de staging;
- eliminacion de una empresa con cuenta Nextcloud, verificando que su usuario y
  archivos remotos desaparezcan sin afectar otra empresa.
