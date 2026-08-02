# P109-006 - Cuatro cajas reales y carga transaccional

Fecha: 2026-08-02  
Empresa: Powerful Control System (`empresa_id=12`)  
Entorno modificado: staging aislado  
Candidato final: `7c47d4df6ae9bdc1808c1413bc11edd3ef5e3e2e`  
Workflow inmutable: `30735137007`

No se documentan correos completos, contrasenas, cookies, tokens de invitacion
ni secretos. Produccion no fue desplegada y conservo su imagen API local
`sha256:0e1e77ba0dc46f947e3b10cfe124812832c3c5b07d383a4672337289b8a0cf8f`.

## Cadena de suministro y despliegue

El workflow aprobo build, Trivy, SBOM, publicacion y Compose para el SHA exacto.
Se promovieron exclusivamente a staging los siguientes digests:

- API: `sha256:502f21dd304625d788808dc99b363bd39bf6f18c6263f83b04531ebce5d99ad7`;
- migrador: `sha256:21b0d1893681f926e40d86659a0893a2c2fa78d8dbb865559c560c31892a891c`;
- worker: `sha256:32eb94d2716599a4632433f7855a95a2f171153d89544e8ebc58987c53818041`;
- frontend: `sha256:368f73663ca0b6f455e5f5dba445136839d124087366630ecb708bf72b009dd8`.

Backend, worker y frontend quedaron saludables. `/health` devolvio `status=ok`
y `/ready` devolvio `status=ready` al terminar la prueba.

## Cuatro identidades por flujo normal

Las cuatro invitaciones se reenviaron por el flujo oficial de Administrar
usuarios. Mailu registro cuatro entregas SMTP aceptadas y el buzon autorizado
recibio los cuatro mensajes. Cada enlace fresco completo confirmacion, documento,
contrasena y contrato en staging. Los cuatro logins devolvieron HTTP 200 y
crearon cookies de sesion independientes con CSRF.

Al finalizar, las cuatro cuentas quedaron inactivas, confirmadas, con contrasena
y contrato configurados. No quedo ninguna sesion QA activa ni caja QA abierta.

## Ensayo transaccional del digest final

Se abrieron simultaneamente cuatro cajas independientes, con cierres 15, 18,
17 y 16. Se crearon cinco ventas finales, porque la cuarta caja incluyo una
segunda venta para verificar el contrato mixto correcto:

| Control | Resultado |
| --- | --- |
| Cajas abiertas | 4/4 HTTP 200 |
| Ventas pagadas | 5/5 HTTP 200 |
| Medios | efectivo, debito, Nequi y mixto |
| Inventario compartido | 3 unidades disponibles a 0, exactamente 3 vendidas |
| Pago duplicado | 4/4 rechazados con HTTP 409 |
| Cierres | 4/4 cerrados, diferencia 0, incidencia 0 |
| Esperas PostgreSQL por lock | 0 |

La caja mixta final concilio dos ventas por 200: efectivo 150, otros medios 50
y efectivo esperado 150. La vista historica mostro esas cifras, dos renglones
de venta, cero errores visibles y cero desbordamiento de pagina.

En el mismo digest se abrio la caja 19 y se creo la venta 87 con dos renglones.
Un cajero elimino uno antes del pago con HTTP 200, pago el restante con descuento
de 50 sobre total base 100, y cerro con total cobrado y caja fisica de 50. El
reintento de pago y la devolucion posterior al cierre fueron rechazados con HTTP
409. La base confirmo `descuento_tipo=valor`, `descuento_valor=50`,
`total_pagado=50` y diferencia de caja 0.

## Carga, recursos y recuperacion

Una rafaga deliberada de 100 solicitudes simultaneas termino 100/100 sin errores,
pero expuso saturacion de latencia: p50 1.484 ms, p95 3.667 ms, p99 3.692 ms.
Despues de recuperar, el perfil operativo acordado de concurrencia 10 termino
100/100 con p50 139 ms, p95 289 ms, p99 326 ms y maximo 333 ms, dentro del SLO
critico de 1.200 ms.

Al cierre se observaron cero esperas por lock y una conexion activa de la
consulta. Recursos: backend 18,35 MiB, PostgreSQL 264,2 MiB, worker 23 MiB y
frontend 3,61 MiB; los servicios siguieron saludables.

## Defectos encontrados y corregidos

1. `DELETE` de un renglon abierto exigia permiso destructivo. Ahora se clasifica
   como actualizacion para cajero; borrar el carrito completo conserva permiso D.
2. La base permitia intentar borrar renglones de ventas cerradas. La transaccion
   ahora falla con `ErrCarritoYaPagado` y el handler responde HTTP 409.
3. Desactivar un usuario no invalidaba sus sesiones anteriores. El endpoint
   ahora revoca primero todas las sesiones de la identidad y luego persiste el
   estado inactivo. Antes del cambio las cuatro sesiones conservaron HTTP 200;
   en el candidato final las cuatro pasaron a HTTP 401 inmediatamente.
4. El corte de pago mixto separa el efectivo persistido y reclasifica el tramo
   restante, evitando duplicarlo en otros medios.

`go test ./...`, las pruebas enfocadas, `go vet ./handlers ./db` y
`git diff --check` aprobaron.

## Veredicto

P109-006: **aprobada** para `7c47d4df...`. La aprobacion no autoriza produccion:
DIAN, aislamiento A/B, impresoras/dispositivos fisicos, UAT humano y piloto
siguen siendo compuertas independientes del Plan 109.

