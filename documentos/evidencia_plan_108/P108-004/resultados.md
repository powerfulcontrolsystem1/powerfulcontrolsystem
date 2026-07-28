# P108-004 - aislamiento multiempresa y autorización efectiva

Fecha: 2026-07-28  
Ámbito: validación estática enfocada; sin acceder a datos de otra empresa.

## Controles aprobados localmente

Se ejecutó:

```text
go test ./handlers ./db -run 'Test(EmpresaPermisosTenant|TenantContext|SoporteRemotoSignalingRejectsCrossTenantQuery|PrivateMigrationInventoryQueryScopesTenant|SaveEmpresaPrivateUploadIsTenantScopedAndRandom|ReconcileEmpresaCxPRowsIsReadOnlyAndTenantScoped|CxPSupplierCatalogIsTenantFilteredAndActiveOnly|MobileNormalizeEmpresaJSONUsesQueryTenant)' -count=1
```

Resultado: **PASS** en `handlers` y `db`.

La cobertura fija que el contexto tenant no se derive de `empresa_id` aportado
por cliente, rechaza parámetros, cabeceras y JSON de otra empresa, preserva el
scope en archivos privados, señalización remota, CxP, proveedores y API móvil.

## Límite

P108-004 permanece **parcial**. La prueba A/B autenticada con una segunda
empresa real, los roles de contador/cajero, revocación de sesión, licencias,
CSRF, descargas, IA, cachés, exportaciones y jobs aún deben ejecutarse de forma
controlada. No se usó ni inspeccionó información de otra empresa en esta etapa.
