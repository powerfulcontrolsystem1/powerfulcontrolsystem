# Contrato tecnico: venta publica empresarial por empresa

Fecha: 2026-04-18
Estado: vigente

## Alcance

Este contrato cubre la configuracion administrativa de la tienda publica por `empresa_id`, la publicacion del catalogo visible para clientes finales, la creacion de ordenes publicas, la apertura de pago por Wompi o Epayco, la consulta de estado posterior y la conciliacion de webhooks sobre `empresa_venta_publica_ordenes`.

## Rutas implicadas

### Rutas empresariales autenticadas

- `GET /api/empresa/venta_publica?action=config`
- `POST /api/empresa/venta_publica?action=config`
- `PUT /api/empresa/venta_publica?action=config`
- `GET /api/empresa/venta_publica?action=catalogo`
- `GET /api/empresa/venta_publica?action=detalle&id={item_id}`
- `POST /api/empresa/venta_publica?action=crear`
- `PUT /api/empresa/venta_publica?action=actualizar&id={item_id}`
- `PUT /api/empresa/venta_publica?action=activar&id={item_id}`
- `PUT /api/empresa/venta_publica?action=desactivar&id={item_id}`
- `DELETE /api/empresa/venta_publica?id={item_id}`
- `POST /api/empresa/venta_publica?action=subir_imagen`
- `GET /api/empresa/venta_publica?action=ordenes`

### Rutas publicas

- `GET /api/public/venta_publica?action=catalogo`
- `POST /api/public/venta_publica?action=crear_pago`
- `GET /api/public/venta_publica?action=estado_pago`
- `POST /wompi/webhook`
- `POST /epayco/webhook`

### Frontend asociado

- `web/administrar_empresa/venta_publica.html`
- `web/administrar_empresa/configuracion.html`
- `web/administrar_empresa/configuracion_integraciones.html`
- `web/venta_publica.html`

## Persistencia implicada

- `empresa_venta_publica_configuracion`
- `empresa_venta_publica_items`
- `empresa_venta_publica_ordenes`
- `productos` e `inventario_existencias` como fuente de hidratacion opcional de catalogo

## Entradas obligatorias

### Configuracion de tienda autenticada

- `empresa_id` por query y wrapper de permisos
- si `wompi_activo=1`: `wompi_public_key`, `wompi_private_key_ref`, `wompi_integrity_key_ref`
- si `epayco_activo=1`: `epayco_public_key`, `epayco_private_key_ref`

### Item publico autenticado

- `empresa_id` por query y wrapper de permisos
- `nombre`
- `precio`

### Pago publico

- `empresa_id` o `empresa_slug`
- `metodo_pago`
- `comprador_nombre`
- `comprador_email`
- `accept_terms=true`
- `items[]` con al menos un `item_id` valido y `cantidad > 0`

### Consulta publica de estado

- `empresa_id` o `empresa_slug`
- `order_code`

## Entradas opcionales

- configuracion: `empresa_slug`, `nombre_tienda`, `descripcion_tienda`, `logo_url`, `banner_url`, `color_primario`, `moneda`, `dominio_publico`, `mostrar_stock`, `wompi_mode`, `wompi_event_key_ref`, `epayco_mode`, `epayco_customer_id`, `observaciones`
- item: `producto_id`, `codigo_publico`, `descripcion`, `moneda`, `imagen_url`, `stock_publicado`, `orden_visual`, `destacado`, `observaciones`
- subida de imagen: `item_id` y archivo `imagen`
- pago publico: `comprador_telefono`
- estado publico: `transaction_id`, `metodo_pago`, `reference`

## Salidas y estados funcionales

### Configuracion autenticada

- `200` con `ok=true`, `config` sanitizada para el panel y `public_path`
- slug normalizado en `config.empresa_slug`

### Catalogo autenticado

- `200` con `ok=true`, `empresa_id`, `total` y `rows`
- `detalle` responde `item`

### Catalogo publico

- `200` con `ok=true`, `empresa_id`, `empresa_slug`, `tienda`, `items` y `paths`
- `tienda.payment_methods` expone solo metodos activos, nunca referencias secretas

### Pago publico

- `200` con orden creada y metadata del proveedor cuando la pasarela esta operativa
- `412` con `order_id` y `order_code` si la orden se crea pero la pasarela no esta activa o su configuracion minima no permite abrir el cobro
- `4xx` cuando faltan datos del comprador, items validos o telefono Nequi en Wompi

### Estado de orden publica

- `200` con `order`, `status`, `status_local`, `provider` y `data`
- `status_local` canonico: `pendiente`, `aprobado`, `rechazado`, `error`

## Codigos de error esperados

- `400` por `action` invalida, `empresa_id` invalido, `id` faltante, JSON invalido, `order_code` faltante, `metodo_pago` invalido o payload incompleto
- `404` cuando la empresa o la orden no existen en el alcance resuelto
- `405` por metodo HTTP no permitido
- `412` cuando la pasarela no esta activa o la configuracion minima requerida no existe para la tienda
- `502` cuando la API externa de Wompi o Epayco falla durante la apertura del cobro

## Invariantes

1. Toda operacion autenticada del modulo debe ejecutarse bajo `WithEmpresaVentasPermissions` y nunca fuera del `empresa_id` del contexto.
2. Toda operacion publica debe resolver empresa por `empresa_id` o `empresa_slug`, pero la frontera final sigue siendo la empresa resuelta en backend.
3. `empresa_slug` debe quedar normalizado con formato URL-safe y ser reutilizable en `/venta_publica.html?empresa_slug=...`, `/{slug}/venta_publica.html` y subdominios compatibles.
4. El catalogo publico solo puede listar items con estado distinto de `inactivo`.
5. El payload publico de `tienda` nunca debe exponer `wompi_private_key_ref`, `wompi_integrity_key_ref`, `wompi_event_key_ref`, `epayco_private_key_ref` ni otros secretos resueltos.
6. Las referencias seguras de credenciales empresariales deben validarse como referencias permitidas (`env:`, `file:` o equivalente validado) y no como secretos en texto plano arbitrario.
7. Si `producto_id` existe, el backend puede hidratar nombre, descripcion, imagen, precio y codigo desde `productos`, pero solo dentro del mismo `empresa_id`.
8. Toda orden publica debe persistirse primero en `empresa_venta_publica_ordenes` con `estado_pago=pendiente` antes de depender de la respuesta de la pasarela.
9. Una orden publica pertenece a una sola empresa y su conciliacion por webhook o consulta de estado no debe poder reasignarse a otra empresa.
10. La referencia externa canonica de venta publica debe mantener el formato `VP|empresa_id|codigo_orden`.
11. La consulta publica de estado solo puede responder una orden del `empresa_id` o `empresa_slug` solicitado; no debe servir como lookup transversal de ordenes ajenas.
12. `wompi_nequi` exige telefono colombiano de 10 digitos iniciado en `3`; Epayco no exige ese formato duro.
13. El frontend publico puede abrir la tienda desde query string, slug en path o slug inferido desde subdominio, pero el backend mantiene la resolucion final.
14. `payment_methods` visibles para cliente deben derivarse solo de la configuracion activa y nunca de defaults del navegador.

## Side effects obligatorios

- upsert de `empresa_venta_publica_configuracion`
- CRUD y activacion/desactivacion de `empresa_venta_publica_items`
- subida de imagen a `web/uploads/venta_publica/empresa_{empresa_id}/`
- creacion de `empresa_venta_publica_ordenes` con snapshot de items y totales
- actualizacion posterior de `estado_pago`, `transaction_id`, `referencia_externa`, `pasarela_payload_json` y `pagado_en`
- conciliacion desde `GET estado_pago` y desde webhooks globales `/wompi/webhook` y `/epayco/webhook`

## Reglas por proveedor

### Wompi

- requiere `wompi_public_key`, `wompi_private_key_ref` y `wompi_integrity_key_ref` si `wompi_activo=1`
- crea transaccion tipo `NEQUI`
- firma la referencia `VP|empresa_id|codigo_orden` con `integrity_key`
- usa `redirect_url` hacia `venta_publica.html` con `empresa_slug`, `order_code`, `provider=wompi` y `status=pending`
- consulta estado por `transaction_id`

### Epayco

- requiere `epayco_public_key` y `epayco_private_key_ref` si `epayco_activo=1`
- crea sesion Smart Checkout v2
- usa `invoice=reference` y `confirmation=/epayco/webhook`
- conserva `order_code`, `empresa_id` y `reference` en `extras`
- consulta estado posterior por `reference`

## Errores de contrato esperados

- el backend debe rechazar configuracion activa de Wompi o Epayco si faltan credenciales minimas.
- el backend debe rechazar item con `precio < 0` o `stock_publicado < 0`.
- el backend debe rechazar `empresa_id` del body si no coincide con el `empresa_id` del query autenticado.
- el backend debe rechazar pago cuando `items` esta vacio o todos los items solicitados quedan invalidos.
- el backend debe conservar la orden creada aunque la pasarela no este activa, devolviendo un error operacional trazable con `order_id` y `order_code`.

## Evidencia tecnica minima

- pruebas de `backend/handlers/venta_publica_test.go` para configuracion, catalogo, activacion/desactivacion, catalogo publico, orden pendiente y consulta de estado
- validacion de rutas publicas en `backend/utils/utils_test.go`
- diagnostico limpio en la documentacion de gobernanza que describa este modulo

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`
- `ADR-0002-postgresql-runtime-canonico-vps.md`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_checkout_licencias.md`