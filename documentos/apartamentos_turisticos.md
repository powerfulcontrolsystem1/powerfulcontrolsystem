# Módulo Apartamentos Turísticos

## Objetivo

El módulo `apartamentos_turisticos` administra alquiler vacacional por empresa: unidades, tarifas, reservas, huéspedes, disponibilidad, check-in, checkout, códigos de acceso y tareas de limpieza o mantenimiento.

## Alcance funcional

- Inventario de apartamentos, suites, casas, estudios o cabañas con capacidad, habitaciones, camas, baños, foto, amenidades y reglas de casa.
- Tarifas globales o por apartamento, canal y temporada.
- Reservas con fechas de entrada/salida, noches, huésped, contacto, canal, estado de pago, impuestos, limpieza, depósito y saldo.
- Validación de conflicto por fechas para impedir doble reserva del mismo apartamento.
- Dashboard con apartamentos disponibles, ocupados, reservas activas, check-ins/checkouts del día e ingresos del mes.
- Operación de check-in, checkout y cancelación.
- Programación de tareas de limpieza, inspección, lavandería, inventario o mantenimiento.
- Cambio manual de estado operativo y ocupación de cada apartamento.

## Integracion con nucleo

- Cada unidad se enlaza con `servicios` mediante `servicio_id` para vender noches de alojamiento sin crear catalogo paralelo.
- Cada reserva se enlaza con `clientes` mediante `cliente_id`, reutilizando documento, telefono o email cuando ya existe.
- El checkout genera `carritos_compras` con canal `apartamentos_turisticos`, referencia externa de reserva, item de alojamiento, item de limpieza cuando aplica e impuesto calculado por carrito.
- `carrito_id` y `carrito_item_id` quedan guardados en la reserva para trazabilidad y para evitar ventas duplicadas.

## Seguridad y aislamiento

- Todas las tablas usan `empresa_id`.
- El endpoint protegido es `/api/empresa/apartamentos_turisticos`.
- El acceso pasa por `WithEmpresaApartamentosTuristicosPermissions`.
- La licencia controla el módulo `apartamentos_turisticos`; los roles pueden activar o desactivar la página `linkApartamentosTuristicos`.

## Archivos principales

- Backend DB: `backend/db/apartamentos_turisticos.go`
- Backend handler: `backend/handlers/apartamentos_turisticos.go`
- Ruta: `backend/main.go`
- Permisos: `backend/handlers/empresa_permisos.go`
- Interfaz: `web/administrar_empresa/apartamentos_turisticos.html`
- Menú empresa: `web/administrar_empresa.html` y `web/js/administrar_empresa.js`
- Licencias: `web/super/licencias.html`

## Endpoints

- `GET action=dashboard`
- `GET action=config`
- `GET action=apartamentos`
- `GET action=tarifas`
- `GET action=reservas`
- `GET action=tareas`
- `POST action=config`
- `POST action=apartamentos`
- `POST action=tarifas`
- `POST action=reservas`
- `POST action=checkin`
- `POST action=checkout`
- `POST action=cancelar`
- `POST action=estado_apartamento`
- `POST action=tareas`
- `POST action=estado_tarea`

## Flujo operativo

1. Crear apartamentos con capacidad, precio base, tarifa de limpieza, depósito sugerido, reglas y amenidades.
2. Crear tarifas por canal o temporada.
3. Registrar reservas con fechas, huésped y canal.
4. Realizar check-in al ingreso del huésped.
5. Realizar checkout al cierre de la estadía.
6. Completar la tarea de limpieza o mantenimiento y devolver la unidad a disponible.

## Validación realizada

Se validan compilación Go, presencia de ruta protegida, permisos, menú, licencia y sintaxis JavaScript de la pantalla. La interfaz usa variables centralizadas de apariencia para modo claro/oscuro.
