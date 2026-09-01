# Auditoria de idempotencia punta a punta

Fecha: 2026-08-26. Alcance: codigo local de API, PostgreSQL, workers y frontend
en `D:\powerfulcontrolsystem`. No incluye despliegue ni llamadas reales a
Wompi, Epayco, Rappi, SMTP o DIAN.

## Dictamen

El sistema no puede declararse universalmente idempotente para cada mutacion:
gran parte del CRUD conserva contratos de actualizacion normales y no ofrece
replay exacto mediante `Idempotency-Key`. Los flujos con impacto financiero o
externo revisados quedan endurecidos localmente con una de estas garantias:

1. replay exacto de una respuesta ya completada;
2. conflicto si una clave se reutiliza con otra solicitud;
3. transaccion e indice unico para el efecto de dominio;
4. semantica de maximo una vez y estado `incierto` cuando el resultado remoto
   no puede conocerse con seguridad.

Por tanto, el resultado local es **apto para continuar a migracion y pruebas de
integracion**, pero sigue en **NO-GO de produccion** hasta ejecutar las compuertas
del final de este documento.

## Matriz revisada

| Flujo | Identidad y huella | Efecto durable | Repeticion / concurrencia |
| --- | --- | --- | --- |
| API movil mutante | Metodo, ruta escapada, query ordenada, actor y body | `mobile_api_idempotency` por `empresa_id` | Reproduce 2xx; 4xx libera; 5xx queda `incierto` y no se ejecuta solo |
| Checkout Wompi y Epayco | `Idempotency-Key` hasheada, empresa, licencia y payload canonico | `payment_checkout_idempotencia`; referencia opaca estable | Reproduce la respuesta persistida; distinta huella responde conflicto; timeout remoto queda incierto |
| Activacion de licencia pagada | Proveedor + transaccion/referencia ya persistida | Estado en `pagos_wompi` / `pagos_epayco` | `done` no se repite; `processing` no se reclama por tiempo para no sumar vigencia dos veces |
| Correo de activacion | Proveedor + fila de pago + tipo de efecto | `payment_post_effect_idempotencia` | Solo fallos confirmados pueden reintentarse; resultado SMTP ambiguo queda `incierto` |
| Venta y pago de carrito | Transicion condicional del carrito + codigo de operacion | Estado pagado, historial inmutable, efectivo de caja y outbox en una transaccion | `RowsAffected` bloquea doble pago; el evento contable `venta_pagada` tiene unicidad por empresa y carrito |
| Venta offline | `empresa_id`, `sync_key` y hash inmutable de payload | Lease de procesamiento en `empresa_ventas_offline_sync` | Mismo payload reproduce resultado; payload distinto da conflicto; un worker por clave |
| Pago CxP | Clave, empresa, cuenta, importe y solicitud canonica completa | Asignacion y outbox en una transaccion | Mismo pago se reproduce; cualquier cambio de periodo, metodo, referencia, concepto u observacion da conflicto |
| Reintentos DIAN | Empresa, documento y claim con token | Lease atomico con `FOR UPDATE SKIP LOCKED` | Evita que dos workers procesen la misma fila simultaneamente |
| Rappi manual | Clave por empresa, accion, orden y payload | Ledger durable; bitacora por orden | Reproduce resultado; red incierta no reenvia automaticamente; firma webhook exige timestamp fresco |
| Webhook Rappi | Empresa + `rappi_order_id` o hash del payload | Upsert unico por empresa/orden | Duplicados no crean otra orden; el orden semantico de estados depende de los timestamps/eventos del proveedor |
| Venta publica | Orden empresarial | Actualizacion monotona de pago | Un estado aprobado no regresa por un callback tardio |
| Propuesta IA | Propuesta + clave | Resultado persistido | Una propuesta completada reproduce resultado; una ejecucion incierta no vuelve a aplicar herramientas |
| Outbox y jobs | Empresa, tipo/topic y hash de clave | Indices unicos, lease y reintentos | Una clave con version o payload distinto se rechaza; no se oculta una colision logica |

## Invariantes de seguridad

- Las claves crudas no se guardan: se persisten huellas SHA-256.
- La identidad incluye `empresa_id`; una clave de otra empresa no comparte
  resultado ni claim.
- Los IDs hijos siguen sujetos a validacion de pertenencia, permiso y licencia
  en sus handlers; la idempotencia no reemplaza autorizacion.
- Un resultado remoto ambiguo nunca se transforma en reintento automatico. La
  disponibilidad se recupera mediante conciliacion explicita, no arriesgando un
  segundo cobro, correo, vigencia o comando.

## Limitaciones residuales

- El CRUD ordinario no implementa un contrato global de replay HTTP. Sus PUT y
  DELETE pueden ser idempotentes por estado final, pero no prometen devolver la
  misma respuesta byte por byte.
- Un checkout iniciado desde otro navegador o con otra clave representa otra
  operacion. La UI conserva la clave durante la sesion, pero no puede deducir la
  intencion del usuario entre dispositivos.
- Rappi conserva una fila por orden, pero falta certificar con eventos reales
  del proveedor una maquina de estados completa contra callbacks fuera de orden.
- La factura de licencia usa codigo determinista y upsert local. La garantia
  exactamente-una-vez del envio fiscal depende tambien del contrato del
  conector externo y debe validarse en sandbox.
- Los estados `processing` o `incierto` de pagos y efectos externos requieren
  una consulta/conciliacion operativa antes de liberar un reintento.

## Evidencia local

- Pruebas enfocadas de DB y handlers: pagos, Wompi/Epayco, API movil, DIAN,
  offline, CxP, Rappi, IA, venta publica y carrito.
- Regresion de catalogo de migraciones, outbox, jobs y worker.
- `pcs-migrate` incorpora `20260826-001-payment-idempotency-v1` en la base super
  y `20260826-001-sale-accounting-idempotency-v1` junto con
  `20260826-003-operational-idempotency-v1` en la base empresarial.
- El inventario de `Ensure*` y su manifiesto se regeneran con la herramienta
  oficial del repositorio.

## Compuertas antes de GO

1. Ejecutar `pcs-migrate` primero en staging y revisar duplicados historicos de
   eventos `ventas/venta_pagada` si el indice unico bloquea la migracion.
2. Ejecutar pruebas PostgreSQL reales de concurrencia; esta revision local no
   encontro DSN de pruebas configurado.
3. Simular timeout despues de aceptar la solicitud en Wompi, Epayco, Rappi,
   SMTP y conector fiscal; verificar que quede `incierto` y que la conciliacion
   no duplique el efecto.
4. Verificar replay desde la UI, dos replicas de API y dos workers.
5. Ejecutar CI, desplegar y repetir smoke tests antes de declarar estado de
   produccion.
