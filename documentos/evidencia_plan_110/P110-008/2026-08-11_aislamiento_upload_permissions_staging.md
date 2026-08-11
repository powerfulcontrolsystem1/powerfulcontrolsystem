# P110-008 - Aislamiento de upload-permissions en staging

Fecha: 2026-08-11

Ambiente: staging aislado

Producción: no modificada

## Hallazgo

La promoción del candidato inmutable P110 se interrumpió antes de crear el
backend cuando Docker detectó que el nombre `pcs-upload-permissions` ya estaba
ocupado. La inspección de solo lectura confirmó que pertenecía al proyecto de
plataforma, estaba terminado y montaba el volumen de uploads de plataforma.

Eliminarlo habría mezclado el ensayo con recursos de plataforma, por lo que no
se ejecutó esa acción.

## Corrección

El override `deploy/docker-compose.staging.yml` declara el servicio heredado
con:

- `container_name: pcs-staging-upload-permissions`;
- `pcs_staging_web_uploads:/app/web/uploads`.

Así el init-container del candidato utiliza solamente los recursos de staging.
La prueba `TestStagingUploadPermissionsUsesOnlyStagingNameAndVolume` exige los
dos valores y la regresión de promoción exige también ClamAV en el conjunto de
servicios recreados.

La primera inspección posterior confirmó el nombre nuevo del init-container,
pero mostró que `worker` y `migrate` seguían heredando los nombres de volumen
del Compose base dentro del proyecto de staging. Se agregó el override completo
de logs, uploads, almacenamiento privado y backups para worker, y de
almacenamiento privado para migrate. Este segundo hallazgo obliga a regenerar el
candidato: no se certifica una topología con adjuntos visibles en backend y no
en worker.

El ejecutor de restauración acepta ahora la ruta privada explícita de staging;
no intenta inferirla dentro del checkout desechable. También resuelve
`pcs-staging-backend` como la imagen API cuando consulta los digests activos.
El inventario remoto confirmó que los snapshots viven fuera del checkout de
candidato y contienen las dos piezas mínimas (base PostgreSQL y almacenamiento
privado). Por ello el runner acepta también `BackupDir`: el operador pasa la
ruta existente de staging sin copiar, listar contenidos sensibles ni usar una
copia de plataforma.

## Validación local

```text
go test . -run TestStagingDigestPromotionRequiresAllExactImagesBeforeRecreate|TestStagingUploadPermissionsUsesOnlyStagingNameAndVolume -count=1
```

Resultado: PASS.

## Pendiente

Crear un nuevo candidato inmutable desde el commit que contiene el override
completo, promoverlo a staging y verificar que backend, worker, frontend,
migrate e init-container monten exclusivamente los mismos recursos de staging;
después ejecutar health, readiness, ClamAV y restore efímero.

Estado: **parcial**. El hallazgo bloqueó una promoción insegura y no afectó
producción.
