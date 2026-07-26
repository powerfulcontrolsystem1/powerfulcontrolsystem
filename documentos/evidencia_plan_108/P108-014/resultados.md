# P108-014 - CxP: catálogo operativo y abono controlado en staging

Fecha: 2026-07-26  
Entorno: candidato aislado de staging, SHA `397e42a7fe246ed3e99c1c7fa729f0c6b360bb6e`  
Empresa: Powerful Control System (`empresa_id=12`)  
Alcance: flujo autenticado de Compras a Finanzas/CxP. No se accedió al sitio
público ni se enviaron documentos fiscales, transferencias, pagos bancarios o
archivos reales.

## Hallazgo y corrección validada

La interfaz de Compras guardaba proveedores en `proveedores`, pero el selector
de CxP leía `empresa_proveedores`. Esto dejaba fuera el proveedor creado desde
la ruta operativa. El candidato corrige el catálogo de CxP a `proveedores` y
fuerza el `empresa_id` obtenido del contexto validado al crear o editar un
proveedor.

Prueba visual autenticada:

1. Se creó una única ficha controlada de proveedor desde Compras para staging.
2. En Finanzas se eligió `Cuenta por pagar`.
3. El selector mostró dicha ficha con su identificador de proveedor, confirmando
   que Compras y CxP usan el mismo catálogo operativo.
4. Se creó una CxP de prueba interna por 100, con documento de prueba no fiscal.
5. Se registró un único abono interno de 25 por el control visible de Finanzas.
   La fila quedó en estado parcial con saldo 75 y un movimiento de egreso
   interno asociado; no se ejecutó un pago externo.

## Concurrencia e idempotencia

La implementación del candidato vuelve a consultar el pago por hash de clave
de idempotencia después de bloquear la CxP, por lo que un replay concurrente
de la misma clave devuelve el resultado existente en vez de intentar otro
movimiento. Las pruebas enfocadas de `db` y `handlers` aprobaron antes del
despliegue.

La automatización visual autenticada no expone `fetch`, XHR ni el encabezado de
idempotencia para emitir dos solicitudes con la misma clave desde el navegador.
Por eso esta evidencia no declara aprobada la carrera HTTP exacta: sigue
pendiente ejecutarla con un cliente autenticado de staging que permita fijar la
misma `Idempotency-Key` en dos solicitudes simultáneas. La prueba visual sí
confirma el alta, selección y un abono controlado por el flujo oficial.

## Resultado

- Catálogo operativo CxP: aprobado en staging.
- Aislamiento de escritura de proveedor por cuerpo JSON: corregido en código;
  la validación A/B de otra empresa continúa pendiente.
- Alta CxP y abono visible controlado: aprobados en staging.
- Replay HTTP exactamente simultáneo con igual clave: pendiente.
- Producción: permanece `NO-GO`.
