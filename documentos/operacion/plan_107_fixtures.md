# P107 - contrato de fixtures contables

Estado: **implementado como manifiesto local; ejecución de datos pendiente de staging autorizado**.

El generador `backend/tools/plan107_fixture_manifest` no abre una base de
datos, no lee credenciales y no crea registros. Construye el contrato que debe
aprobarse antes de que un ejecutor futuro cree datos de prueba en staging.

## Uso

Desde `D:\powerfulcontrolsystem\backend`:

```powershell
go run ./tools/plan107_fixture_manifest -empresa-id 12 -run-id P107-QA-CONTADOR
```

El comando solo produce JSON en stdout. `empresa-id` debe ser positivo y el
`run-id` debe comenzar con `P107-QA`; no acepta un prefijo de producción.

## Contenido del manifiesto

- saldo de apertura y comprobante balanceado;
- venta de menta en efectivo, transferencia y crédito;
- abono parcial de CxC;
- compra a crédito y abono CxP;
- impuestos y retenciones;
- reverso auditable;
- pasos obligatorios de conciliación y limpieza.

Las claves de idempotencia se derivan de `run-id` y del caso, y el manifiesto
incluye SHA-256. Dos generaciones con el mismo `empresa_id` y `run-id` deben
producir la misma huella.

## Compuertas antes de ejecutar datos

1. Staging aislado y confirmado, nunca producción.
2. Autorización puntual de empresa, responsable y ventana.
3. Verificación del producto `menta`, precio, impuesto, costo y stock vivos.
4. Copia/restauración disponible y evidencia inicial capturada.
5. Ejecutor con transacciones, reversos y auditoría; no SQL manual.
6. Conciliación de inventario, caja, banco, CxC, CxP, impuestos y asientos
   antes y después de la limpieza.

Este contrato no autoriza DIAN, movimientos bancarios externos, cierre de
períodos, comunicaciones a proveedores ni `rs`.
