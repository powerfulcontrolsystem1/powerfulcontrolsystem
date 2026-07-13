# Nextcloud empresarial

## Alcance

El servicio Nextcloud empresarial se ejecuta en el VPS principal y es distinto
del Nextcloud auxiliar administrado desde VPS2. PCS crea una cuenta tecnica por
empresa, aplica una cuota y conserva unicamente usuario, cuota y estado. La
contrasena se genera con 32 bytes aleatorios, se entrega una sola vez al
administrador autorizado y no se guarda en PCS.

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
- Aprovisionamiento y restablecimiento dejan auditoria sin contrasenas.

## Despliegue

1. Copiar `deploy/nextcloud/.env.example` a `deploy/nextcloud/.env` en el VPS.
2. Crear tres archivos de secreto fuera del repositorio con `chmod 600`.
3. Definir dominio, reverse proxy, ruta `NEXTCLOUD_DATA_PATH` y proxies fiables.
4. Ejecutar `bash deploy/scripts/vps-nextcloud-up.sh`.
5. Configurar en Super administrador una contrasena de aplicacion OCS distinta
   de la contrasena inicial del contenedor.
6. Probar la conexion OCS y aprovisionar una empresa de ensayo.

Las imagenes se fijan en Nextcloud 34.0.1 Apache, PostgreSQL 16.14 Alpine y Redis
7.4.9 Alpine. La imagen Apache requiere iniciar como root para preparar volumenes
y luego ejecuta el servicio con `www-data`; se mitiga con red interna,
`no-new-privileges`, secretos montados, puertos internos no publicados, tmpfs,
healthchecks y limites de CPU/RAM.

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
  empresas de staging.
