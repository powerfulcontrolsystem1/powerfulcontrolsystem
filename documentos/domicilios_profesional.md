# Modulo profesional de domicilios

Fecha: 2026-05-11
Estado: vigente, documentado tambien en el portal publico `web/index.html`

El modulo de domicilios agrega una operacion tipo marketplace local: central administrativa, restaurantes aliados, domiciliarios moviles, cliente publico, asignacion por cercania, tracking GPS por navegador y codigo de entrega.

## Superficies

- Central administrativa: `/administrar_empresa/domicilios.html`
- Cliente publico: `/domicilios.html?empresa_id=ID`
- Restaurante: `/domicilios_restaurante.html?empresa_id=ID`
- Domiciliario: `/domicilios_domiciliario.html?empresa_id=ID`

## Presencia en portal publico

El index comercial describe este modulo como `Domicilios tipo Rappi` dentro de la seccion ejecutiva de modulos y agrega una tarjeta fallback `Domicilios y entregas`. La descripcion publica resume tres frentes:

- Restaurante: menu, productos, disponibilidad y estados de preparacion.
- Domiciliario: PIN, presencia, ofertas, ubicacion GPS y estados de entrega.
- Central administrativa: configuracion, metricas, pedidos, tracking y trazabilidad.

## Flujo operativo

1. La central configura radios, tarifa base, tarifa por kilometro, comision, autoasignacion y codigo de entrega.
2. Se registran restaurantes con codigo y PIN.
3. Se registran domiciliarios con documento y PIN.
4. El restaurante publica productos del menu.
5. El cliente crea el pedido desde el portal publico.
6. El restaurante confirma, prepara y marca `Pedido listo`.
7. El sistema genera ofertas a domiciliarios online y disponibles dentro del radio.
8. El domiciliario acepta, comparte ubicacion y actualiza estados.
9. La entrega se cierra con el codigo de seguridad visible para el cliente.
10. Al quedar `entregado`, el pedido genera cliente, servicios de menu, carrito, items y pago central en el nucleo comercial.

## Integracion con nucleo

- Los productos del menu se enlazan con `servicios` mediante `servicio_id`.
- Los pedidos se enlazan con `clientes` mediante `cliente_id`, reutilizando telefono cuando existe.
- Los pedidos entregados crean `carritos_compras` con canal `domicilios`, referencia externa del pedido y metodo de pago normalizado.
- Cada linea del pedido queda vinculada a `carrito_compra_items` mediante `carrito_item_id`; tarifa de domicilio y propina se agregan como servicios centrales.
- La accion protegida `POST /api/empresa/domicilios?action=sincronizar_nucleo` migra pedidos entregados historicos sin borrar restaurantes, domiciliarios, ofertas, tracking ni estados logisticos.

## Datos demo

Desde la central se puede usar `Cargar demo productiva`.

- Restaurante demo: `REST-DEMO`, PIN `1234`
- Domiciliario demo 1: documento `1001`, PIN `1234`
- Domiciliario demo 2: documento `1002`, PIN `1234`

## Endpoints principales

- Protegido: `/api/empresa/domicilios`
- Publico: `/api/public/domicilios`

Acciones protegidas: `dashboard`, `config`, `restaurants`, `couriers`, `menu`, `orders`, `dispatch`, `order_state`, `sincronizar_nucleo`, `seed_demo`.

Acciones publicas: `catalog`, `order`, `tracking`, `courier_login`, `courier_presence`, `courier_location`, `courier_offers`, `courier_orders`, `respond_offer`, `restaurant_login`, `restaurant_orders`, `order_state`.

## Roles, usuarios y licencias

- El modulo queda registrado como `domicilios` en la matriz central de permisos.
- En licencias, `domicilios` puede activarse o desactivarse de forma independiente a `ventas`.
- En usuarios/roles, la pantalla `Permisos por rol` permite controlar acciones R/C/U/D/A del modulo y el enlace `linkDomicilios` del panel empresa.
- El endpoint protegido usa `WithEmpresaDomiciliosPermissions`; si la licencia no incluye `domicilios` o el rol no tiene permiso, el backend responde `403`.
- Las paginas publicas siguen siendo solo de consulta/operacion externa controlada por token/PIN, sin acceso al panel administrativo.

## Puesta en produccion

- Configurar correctamente `empresa_id` en los enlaces publicos.
- Crear restaurantes reales con ubicacion lat/lng para asignacion por cercania.
- Exigir que domiciliarios entren desde celular y acepten permisos de ubicacion.
- Mantener activo el codigo de entrega para reducir entregas incorrectas.
- Usar HTTPS en produccion para que la geolocalizacion del navegador funcione de forma confiable.
