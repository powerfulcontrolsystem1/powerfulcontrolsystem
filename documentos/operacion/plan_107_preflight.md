# Preflight Plan 107

Antes de aplicar fixtures `P107-QA` en staging ejecutar desde `backend`:

```powershell
go run ./tools/plan107_preflight -endpoint http://127.0.0.1:8082/health
```

La utilidad es de solo lectura. Comprueba Docker, health del endpoint y deja
explícito que el plan es exclusivo de staging. Si `ready_for_fixture_data` es
`false`, no se crean ventas, compras, cartera, pagos, documentos DIAN ni
movimientos externos.
