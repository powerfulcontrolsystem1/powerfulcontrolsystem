# P108-016 - almacenamiento privado y ciclo de archivos

Fecha: 2026-07-30

Candidato desplegado: `f9396da5e41562968996b05136fffca9991b56f9`

Estado: **parcial / NO-GO**

## Evidencia aprobada en este corte

- El almacenamiento privado usa `PCS_PRIVATE_STORAGE_DIR` y separa cada
  categoría por `empresa_<id>`.
- API y worker comparten el volumen `pcs_private_storage`; staging usa su
  volumen aislado `pcs_staging_private_storage`.
- `/ready` respondió `200 {"status":"ready"}` en staging. Ese probe abre,
  escribe, cierra y elimina un archivo temporal con permisos privados, además
  de comprobar las dos bases y las migraciones.
- Los archivos privados se crean con nombre aleatorio de 256 bits, directorio
  `0700`, archivo `0600`, límite de tamaño y validación de contenido.
- La descarga exige el wrapper empresarial del módulo, fuerza
  `Content-Disposition: attachment`, `no-store` y `nosniff`.
- Las referencias rechazan rutas absolutas, `..` y escapes mediante enlaces
  simbólicos. La migración heredada acepta filtros explícitos por empresa y
  categoría.
- La firma DIAN queda en la categoría privada `dian` por empresa y las rutas
  técnicas privadas se redactan del error visible y del buzón.

## Pruebas ejecutadas

```text
go test ./handlers -run 'Test(SaveEmpresaPrivateUpload|ResolveEmpresaPrivateFile|ServeEmpresaPrivateFile|ResolveLegacyPrivatePath|PrivateMigrationInventoryQuery|RuntimePrivateStorageReady|SafeSoporteComprasIAPath|ResolveExistingPrivateFileUnderRoot|EmpresaFacturacionFirmaElectronicaUsesPrivateCompanyFolder|DIANUserVisibleErrorRedactsPrivateSignaturePath|EmpresaBuzonSanitizaAlertasDIANHistoricasConRutaPrivada)' -count=1
PASS
```

También se verificó el contrato Compose versionado y el `ready` público del
digest desplegado. No se guardaron rutas privadas, credenciales ni contenido de
empresa en esta evidencia.

## Pendiente para aprobar

1. Crear un archivo autenticado en una réplica y descargarlo desde otra.
2. Ejecutar matriz negativa A/B con dos identidades empresariales reales.
3. Restaurar volumen, metadatos y checksums desde el snapshot del candidato.
4. Conciliar migración de todos los archivos heredados y huérfanos.
5. Probar cuotas, retención, borrado y recuperación.
6. Integrar y ensayar el antivirus de archivos cuando corresponda.

P108-016 pasa de pendiente a **parcial**. Esta evidencia no sustituye la prueba
multi-réplica, multiempresa ni el restore completo.
