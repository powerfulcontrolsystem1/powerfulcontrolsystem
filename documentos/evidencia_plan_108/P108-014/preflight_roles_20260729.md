# P108-014 - Preflight de permisos para cajas concurrentes

Fecha: 2026-07-29  
Ambiente evaluado: código integrado y staging autorizado  
Empresa objetivo: Powerful Control System (`empresa_id=12`)

## Resultado estático

`tools/qa_roles_matrix.mjs` aprobó la matriz de Super administrador,
Administrador de empresa, Cajero, Vendedor, Asesor comercial y Soporte. Para
el rol Cajero no faltan páginas ni APIs operativas requeridas. Las pruebas Go
de permisos de cajero también aprobaron, incluyendo el acceso restringido a
carrito, cobros operativos e inventario de consulta.

## Límite

Este preflight no simula cajas: no hubo cuatro sesiones, productos, pagos,
devoluciones, cierres ni cambios de inventario. La aceptación de P108-014 exige
cuatro credenciales temporales de cajero, apertura de cajas separadas,
transacciones concurrentes y conciliación/limpieza posterior sobre staging.

Estado de fase: **pendiente de corrida operativa concurrente**.

## Disponibilidad de identidades 2026-07-30

Una consulta agregada y sin correos ni datos personales confirmó que la empresa
12 no tiene cuatro cajeros utilizables en staging. Existen tres registros
inactivos (administrador, caja y vendedor), todos sin contraseña configurada y
sin correo confirmado.

Se repitieron los contratos Go de permisos operativos, APIs auxiliares, cobro y
código inmutable del documento: **PASS**. No se crearon usuarios por SQL ni se
simularon credenciales. P108-014 permanece pendiente hasta aprovisionar cuatro
cuentas temporales mediante el flujo oficial y ejecutar la corrida concurrente.
