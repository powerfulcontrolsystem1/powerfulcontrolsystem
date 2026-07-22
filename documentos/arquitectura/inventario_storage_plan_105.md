# Inventario de almacenamiento - Plan 105

Fecha de corte: 2026-07-21. Revision estatica de escrituras, lecturas y
borrados en `backend`; no incluye volumenes ni objetos existentes en VPS y no
autoriza mover, borrar ni publicar archivos.

## Estado comprobado

No existe una interfaz de Object Storage en el backend. Las familias revisadas
escriben en filesystem local; por tanto una replica no comparte su contenido y
la disponibilidad depende del volumen del nodo. La configuracion
`PCS_PRIVATE_STORAGE_DIR` solo cambia la raiz local de los archivos privados;
no implementa almacenamiento remoto.

La capa privada actual es una buena base de seguridad: usa raiz por
`categoria/empresa_id`, nombres aleatorios de 32 bytes, permisos 0700/0600,
limite de bytes, bloqueo de extensiones activas, validacion basica de contenido
y descarga autenticada con `no-store` y `nosniff`. No sustituye cifrado en
reposo administrado, antivirus, lifecycle, replicacion ni URLs firmadas.

## Familias identificadas

| Familia | Ubicacion actual | Exposicion/aislamiento | Brecha antes de replicas |
| --- | --- | --- | --- |
| Adjuntos de chat, buzon, finanzas, grafologia, DIAN y soportes IA | `private_storage/<categoria>/empresa_<id>` mediante `private_files.go` | Privada por empresa y proxy autenticado | Adaptador de objetos, metadatos, cuota, retencion, escaneo AV, restore A/B. |
| Imagenes, logos, productos, domotica, ventas y usuarios | `web/uploads/...` mediante `empresa_upload_paths.go` y handlers puntuales | Varias rutas publicas; algunas derivan de carpeta con nombre empresarial | Clasificar publico/privado, claves no adivinables y migrar lectura/borrado atomico. |
| Backups de empresa y superadministrador | `backup/...` mediante `backup_paths.go` | Local con permiso 0700/0600 | Copia externa cifrada, checksum, retencion inmutable y restauracion medida. |
| Snapshots VPS | `backup/vps_snapshots` y proveedor rclone configurado | Archivo local previo a transferencia | Confirmar cifrado, destino, retencion y restauracion independiente del nodo. |
| OnlyOffice/documentos dinamicos/temporales | filesystem local y temporales del proceso | Depende del flujo | Separar temporales efimeros de documentos durables; limpiar por lifecycle. |

## Secuencia ejecutable para Terra

1. Definir una interfaz interna en Go estandar para `Put`, `Open`, `Delete`,
   `Stat` y copia/stream; conservar un adaptador local solo para desarrollo.
   No agregar SDK externo sin autorizacion.
2. Establecer el contrato de clave: `empresas/<empresa_id>/<familia>/<uuid>`;
   prohibir nombres originales como clave y persistir nombre visible, MIME,
   tamano, checksum, fecha, actor y retencion como metadatos.
3. Migrar primero las seis categorias ya privadas. Mantener la autorizacion
   actual del handler: la clave nunca debe permitir elegir empresa ni objeto.
   Hacer dry-run, conteos y rollback por lote; no eliminar origen hasta
   verificar checksum y lectura autorizada.
4. Clasificar cada ruta bajo `web/uploads` como publica o privada. Las privadas
   pasan por proxy autorizado; las publicas requieren claves aleatorias, MIME
   validado y politica explicita de cache. No migrar certificados DIAN ni
   secretos a rutas publicas.
5. Implementar cuota por empresa, limite por familia, retencion y borrado
   diferido/auditado. Añadir escaneo antivirus o una cola de cuarentena antes de
   servir archivos nuevos; documentar el proveedor autorizado cuando se elija.
6. Configurar readiness para fallar en replicas cuando el adaptador sea local.
   Los backups deben copiarse cifrados al storage aprobado y registrar checksum,
   version y fecha de expiracion.
7. En PostgreSQL real y staging ejecutar A/B: subida de A, lectura/borrado de
   B debe ser 403/404; fallo de un nodo, expiracion, restauracion y checksum
   deben conservar el objeto correcto. Medir RPO/RTO y adjuntar evidencia.

## Criterio de cierre P105-015

No cerrar por este inventario. Se requiere adaptador aprobado, todas las
familias durables migradas o con excepcion aceptada, aislamiento A/B y restore
demostrados, y readiness bloqueando replicas con almacenamiento local.
