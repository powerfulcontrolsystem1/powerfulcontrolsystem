# P108-005 - CxP: catálogo operativo y abono controlado en staging

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

## Seguimiento visual adicional

La misma pantalla de staging expone el botón `Cargar factura o recibo con IA` y
habilita su selector de archivo. Se verificó solo la interacción inicial, sin
adjuntar archivo ni invocar un proveedor IA. Una navegación negativa con otro
`empresa_id` no mostró proveedores ni cartera, pero no constituye por sí sola
una aprobación A/B porque no demostró un rechazo explícito por permisos.

## Conciliación de fuentes sin escritura - candidato actual

Fecha: 2026-07-28. Empresa de prueba: Powerful Control System (`empresa_id=12`).

Se ejecutó visualmente el botón **Comparar fuente histórica CxP** en staging.
La interfaz confirmó que la acción es solo lectura y mostró:

- canónica: `1`;
- histórica: `0`;
- solo canónica: `1`;
- solo histórica: `0`;
- diferencias de saldo: `0`.

La única fila canónica corresponde a la obligación de prueba previamente
registrada; no se creó, migró, eliminó ni corrigió ningún dato durante esta
verificación. La consola visual no reportó errores. Este resultado valida el
reconciliador para esa muestra, pero no sustituye la conciliación y aprobación
contable de todas las empresas antes de producción.

## Contratos y límite de prueba A/B 2026-07-29

La prueba dirigida `TestRegistrarPagoCxP*` aprobó los rechazos previos al
acceso a base de datos cuando falta `empresa_id` o la clave de idempotencia.
También se ejecutó una navegación autenticada con un `empresa_id` inexistente
en Finanzas y Soportes IA, sin mutaciones. Ese resultado no cuenta como prueba
A/B: el usuario usado tiene privilegios administrativos y el shell de Finanzas
consulta el catálogo global para mostrar el estado de la empresa inexistente.

P108-005 permanece **parcial** hasta contar con una segunda identidad
operativa limitada de otra empresa para probar rechazo explícito, además de dos
abonos HTTP simultáneos con la misma y con distintas claves de idempotencia.

## Concurrencia HTTP real del candidato `5ec1c48f` - 2026-07-30

Se creó mediante la interfaz oficial una CxP QA de `2` para el proveedor
registrado de PCS, con documento `P108-CONC-5EC1-941964`. No se efectuó ningún
pago bancario o externo.

Primera carrera, dos solicitudes simultáneas por `1` con la misma
`Idempotency-Key`:

- ambas respuestas: HTTP `200`;
- ambas respuestas referenciaron el mismo movimiento financiero `33`;
- valor aplicado final: `1`, no `2`;
- saldo final: `1`, estado `parcial`.

Segunda carrera, dos solicitudes simultáneas por `1` con claves diferentes:

- una respuesta HTTP `200`, movimiento financiero `34`;
- la otra fue rechazada porque ya no quedaba saldo;
- valor pagado final: `2`, saldo `0`, estado `pagada`;
- no hubo sobrepago ni movimiento duplicado.

La prueba también reprodujo un defecto visible: seleccionar el proveedor no
completaba el campo HTML obligatorio “Cliente / proveedor”, por lo que el
navegador impedía silenciosamente enviar el formulario. El candidato siguiente
sincroniza automáticamente el nombre al cambiar el selector. Además, la
condición concurrente “sin saldo pendiente” se tipa como error de dominio y se
mapea a HTTP `409 Conflict` en vez de `400`.

Resultado: concurrencia con clave igual y distinta **PASS** sobre staging. La
corrección UX/semántica requiere publicar y repetir el siguiente digest.
P108-005 continúa parcial únicamente por conciliación integral y la prueba A/B
con una segunda identidad empresarial limitada.
