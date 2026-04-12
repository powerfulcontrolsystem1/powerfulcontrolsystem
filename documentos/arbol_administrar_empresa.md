# Árbol jerárquico — Página `Administrar Empresa`

Este documento muestra la estructura actual de la página `Administrar Empresa` y qué páginas son subpáginas de otras (menús-iframe). Úsalo para reorganizar el árbol según tus necesidades.

**Archivo shell principal:** [Administrar Empresa](web/administrar_empresa.html)

## Estructura (nivel 1 → subniveles)

- [Administrar Empresa (shell)](web/administrar_empresa.html)
  - [Inicio](web/administrar_empresa/inicio.html)
  - [Ventas](web/administrar_empresa/ventas.html)
  - [Venta pública y pagos](web/administrar_empresa/venta_publica.html)

  - [Productos (MENÚ)](web/administrar_empresa/administrar_productos_menu.html)
    - [Administrar productos](web/administrar_empresa/productos/administrar_productos.html)
    - [Bodegas](web/administrar_empresa/productos/bodegas.html)
    - [Categorías](web/administrar_empresa/productos/categorias.html)
    - [Precios](web/administrar_empresa/productos/precios.html)
    - [Combos de productos](web/administrar_empresa/productos/combos_productos.html)
    - [Compras](web/administrar_empresa/productos/compras.html)
  
  - [Configuración (MENÚ)](web/administrar_empresa/configuracion_menu.html)
    - [General](web/administrar_empresa/configuracion.html)
    - [Permisos](web/administrar_empresa/configuracion_permisos.html)
    - [Avanzada](web/super/configuracion_avanzada.html)
    - [Integraciones](web/administrar_empresa/configuracion_integraciones.html)
    - [Carrito de compras](web/administrar_empresa/carrito_de_compras.html)
    - [Propinas](web/administrar_empresa/propinas.html)
    - [Comisiones](web/administrar_empresa/comisiones.html)
    - [Configuración de estaciones](web/administrar_empresa/configuracion_de_estaciones.html)
    - [Tarifas por minutos](web/administrar_empresa/tarifas_por_minutos.html)
    - [Tarifas por día](web/administrar_empresa/tarifas_por_dia.html)
  
  - [Facturación electrónica (MENÚ)](web/administrar_empresa/facturacion_electronica_menu.html)
    - [Facturación electrónica](web/administrar_empresa/facturacion_electronica.html)
    - [Facturas electrónicas](web/administrar_empresa/facturas_electronicas.html)
  
  - [Chat con IA](web/administrar_empresa/chat_con_inteligencia_artificial.html)

  - [Finanzas (MENÚ)](web/administrar_empresa/finanzas_menu.html)
    - [Finanzas](web/administrar_empresa/finanzas.html)
    - [Créditos y cartera](web/administrar_empresa/creditos.html)
    - [Nómina de sueldos](web/administrar_empresa/nomina_sueldos.html)
    - [ERP extendido](web/administrar_empresa/modulos_erp_extendido.html)

  - [Backups empresariales](web/administrar_empresa/backups.html)
  - [Soporte remoto](web/administrar_empresa/soporte_remoto.html)
  - [Ubicación GPS](web/administrar_empresa/ubicacion_gps.html)
  - [Reservas](web/administrar_empresa/reservas_hotel.html)

  - [Reportes (MENÚ)](web/administrar_empresa/reportes_menu.html)
    - [Reportes generales](web/administrar_empresa/reportes.html)
    - [Inventario](web/administrar_empresa/reportes_inventario.html)
    - [Finanzas](web/administrar_empresa/reportes_finanzas.html)
    - [Gráficos y estadísticas](web/administrar_empresa/graficos_estadisticas.html) 

  - [Usuarios](web/administrar_empresa/administrar_usuarios.html)
  - [Códigos de descuento](web/administrar_empresa/codigos_de_descuento.html)
  - [Asistencia de empleados](web/administrar_empresa/asistencia_empleados.html)
  - [Registro de vehículos](web/administrar_empresa/vehiculos_registro.html)
  - [Auditoría](web/administrar_empresa/auditoria.html)
  - [Chat y tareas](web/administrar_empresa/chat_y_tareas.html)
  - [Clientes](web/administrar_empresa/administrar_clientes.html)
  - [Calculadora](web/administrar_empresa/calculadora.html)

## Notas y criterios actuales

- Las páginas marcadas como **(MENÚ)** tienen un `*_menu.html` con menú izquierdo + iframe derecho. Las subpáginas se cargan dentro del `iframe` del menú.
- Para convertir una página en subpágina independiente se usa el patrón: crear `web/administrar_empresa/<modulo>_menu.html` y cambiar en `web/administrar_empresa.html` el `href` del enlace principal para que apunte al nuevo `*_menu.html` (ej.: `linkReportes` → `reportes_menu.html`).
- Algunas subpáginas son cargadas con parámetros (ej.: `reportes_menu.html?subpage=graficos_estadisticas.html`).
- Las rutas referenciadas arriba corresponden a archivos del workspace; si quieres reorganizar, indícame el nuevo padre/orden y yo aplicaré los cambios (creación/actualización de `*_menu.html` y ajustes de `iframe` automáticamente).

---

Archivo generado: `documentos/arbol_administrar_empresa.md` — modifica libremente y dime cómo quieres reorganizar el árbol.
