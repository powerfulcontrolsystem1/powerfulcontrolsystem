# Validacion preproduccion - Powerful Control System

Fecha: 2026-07-11. Alcance: empresa interna Powerful Control System (empresa 12).

## Evidencia confirmada

- Tarifas visuales: Habitacion 1 conserva 120 minutos con COP 2.000 y fraccion
  de 60 minutos COP 1.000 de lunes a viernes; sabado/domingo COP 3.000 y
  COP 1.500. Habitacion 2 aplica COP 10.000 para 1-2 personas y COP 20.000
  para 3-4 personas, con check-in 14:00 y check-out 13:00.
- Estaciones y reportes: las tarifas de motel/hotel se seleccionan por carrito,
  se pueden cambiar manualmente y permanecen aisladas por empresa.
- Impresion: la configuracion de PCS carga cuatro impresoras registradas,
  reglas por funcionalidad y la identidad de computador detectada. La salida
  fisica requiere un agente/driver instalado en el equipo objetivo.
- Pagos de licencias: las pruebas Go cubren firmas, referencias, idempotencia,
  contexto empresa/licencia y contratos de Epayco/Wompi. El endpoint publico
  debe ser la fuente de verdad para la disponibilidad del checkout.
- Offline: las pruebas cubren propietario de sesion, caja obligatoria y claves
  de sincronizacion idempotentes por empresa/cajero.
- Cobro POS: el boton conserva enlace inline, directo y delegado al mismo
  handler, todos protegidos por el cerrojo anti-doble-clic del carrito.

## Condiciones externas antes de produccion

- No confirmar un pago Epayco/Wompi ni una transferencia Bre-B como aprobados
  sin respuesta autenticada de la pasarela, banco o webhook valido.
- Validar en Super administrador que la disponibilidad publica de Epayco y
  Wompi coincida con sus credenciales cifradas y sus switches. Si difieren,
  corregir la configuracion antes de abrir el checkout a clientes.
- Bre-B requiere cuenta/llave receptora real y webhook/API bancaria para
  conciliacion automatica; sin ello el cajero debe exigir comprobante y validar
  la transferencia antes de cerrar.
- La prueba de impresion fisica exige driver, cola y agente local del equipo;
  PCS solo puede validar el enrutamiento y la cola desde el navegador.

## Seguridad de aislamiento

Las rutas empresariales se protegen por sesion, pertenencia, `empresa_id`, rol,
permiso y licencia. Los cambios de esta revision no alteran el aislamiento:
roles personalizados, configuracion de impresoras, estaciones, pagos offline y
carritos mantienen su alcance empresarial en backend.
