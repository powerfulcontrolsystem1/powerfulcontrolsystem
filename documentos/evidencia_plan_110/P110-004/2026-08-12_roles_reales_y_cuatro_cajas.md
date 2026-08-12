# P110-004 — roles reales y cuatro cajas PCS

Fecha: 2026-08-12  
Empresa: Powerful Control System (`empresa_id=12`)  
Ámbito: candidato publicado antes del ajuste IMAP; flujos oficiales de interfaz
y API, sin SQL directo.

## Identidades y alta real

- Se crearon identidades QA dedicadas para administrador de empresa, contador y
  cuatro cajeros. No se modificaron las cuentas personales existentes.
- Las seis invitaciones se enviaron por Mailu a alias corporativos aislados,
  llegaron al buzón PCS y cada registro se completó con documento, contraseña
  aleatoria no registrada en evidencia y aceptación del contrato.
- Las seis sesiones independientes iniciaron correctamente en la empresa 12.
  El administrador abrió el panel, el contador obtuvo el rol contable y cada
  cajero fue dirigido al contexto de Caja principal.

## Matriz autenticada

| Rol | Usuarios | Productos | Cartera CxP | Empresa ajena |
|---|---:|---:|---:|---:|
| Administrador empresa | 200 | 200 | 200 | 403 |
| Contador | 200 | 200 | 200 | 403 |
| Cajero 1 | 403 | 200 | 403 | 403 |
| Cajero 2 | 403 | 200 | 403 | 403 |

Los dos primeros cajeros también recibieron 403 al intentar una mutación de
usuarios con origen y CSRF válidos; el administrador alcanzó la validación del
payload y recibió 400 sin escritura. Esto diferencia autorización de rol y
validación de datos.

## Concurrencia

- Ráfaga deliberada: 160 GET simultáneos entre cuatro cajas, 160 respuestas
  200, p95 7.749 ms y p99 7.833 ms. Se conserva como señal de saturación y no
  como aprobación de rendimiento.
- Patrón operativo: cuatro trabajadores concurrentes, 20 lecturas secuenciales
  por caja sobre productos y clientes; 80/80 respuestas 200, p95 309 ms, p99
  514 ms y duración total 2.673 ms.

## Validación visual

La venta directa se abrió como cajero real en 390 x 844. Búsqueda, cantidades,
cliente, detalle de pago y botón de cierre permanecieron visibles, sin overflow
horizontal del documento. No se emitió venta ni documento fiscal en esta fase.

## Estado

P110-004 permanece parcial. Esta ejecución cubre identidades reales, roles,
aislamiento, denegaciones y cuatro cajas concurrentes, pero aún faltan la matriz
mutante completa por todos los módulos, impresión física, hardware GPIO y firma
humana del responsable operativo.
