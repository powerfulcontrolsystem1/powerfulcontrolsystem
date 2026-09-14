(function () {
  "use strict";

  var items = [
    { id: "preconfigHotel", module: "hotel", title: "Hotel", fullTitle: "Hotel", lead: "Habitaciones, reservas, tarifas por día, consumos y recepción.", description: "Preconfiguración inicial para habitaciones, reservas, recepción, consumos, caja y reportes.", icon: "/img/hotel-logo.svg", secondaryIcon: "/img/sistema motel.png", sections: ["Habitaciones", "Reservas", "Recepción", "Consumos", "Caja", "Reportes"] },
    { id: "preconfigMotel", module: "motel", title: "Motel", fullTitle: "Motel", lead: "Habitaciones por turnos, minibar, tarifas y recepción.", description: "Preconfiguración inicial para habitaciones, tarifas por tiempo, minibar, recepción, caja y aseo.", icon: "/img/motel.png", secondaryIcon: "/img/sistema motel.png", sections: ["Habitaciones", "Tarifas", "Minibar", "Recepción", "Caja", "Aseo"] },
    { id: "preconfigRestaurante", module: "restaurante", title: "Restaurante", fullTitle: "Restaurante", lead: "Mesas, cocina, pedidos, productos y venta directa.", description: "Preconfiguración inicial para mesas, pedidos, cocina, productos, caja y facturación.", icon: "/img/restaurante.png", secondaryIcon: "/img/sistema restaurante.png", sections: ["Mesas", "Pedidos", "Cocina", "Productos", "Caja", "Facturación"] },
    { id: "preconfigBar", module: "bar", title: "Bar", fullTitle: "Bar", lead: "Mesas, barra, bebidas, eventos y caja.", description: "Preconfiguración inicial para mesas, barra, bebidas, comandas, caja y reportes.", icon: "/img/bar-logo.svg", secondaryIcon: "/img/sistema bar.png", sections: ["Mesas", "Barra", "Bebidas", "Comandas", "Caja", "Reportes"] },
    { id: "preconfigPymes", module: "pymes", title: "Pymes", fullTitle: "Pymes", lead: "Venta directa, productos, servicios, clientes y caja.", description: "Preconfiguración inicial para una pyme con punto de venta, productos, servicios, clientes, caja y reportes.", icon: "/img/pymes-logo.svg", secondaryIcon: "/img/sistema punto de venta.png", sections: ["Venta directa", "Productos", "Servicios", "Clientes", "Caja", "Reportes"] },
    { id: "preconfigSalonBelleza", module: "salon_belleza", title: "Salón de belleza", fullTitle: "Salón de belleza", lead: "Sillas, estilistas, agenda, servicios y comisiones.", description: "Preconfiguración inicial para agenda, sillas, profesionales, servicios, insumos, comisiones y caja.", icon: "/img/salon-belleza-logo.svg", secondaryIcon: "/img/sistema salon de belleza.png", sections: ["Agenda", "Sillas", "Profesionales", "Servicios", "Comisiones", "Caja"] },
    { id: "preconfigLavaderoAutos", module: "lavadero_autos", title: "Lavadero de autos", fullTitle: "Lavadero de autos", lead: "Bahías, vehículos, servicios, tiempos y comisiones.", description: "Preconfiguración inicial para bahías, vehículos, servicios de lavado, tiempos, comisiones y caja.", icon: "/img/lavadero-autos-logo.svg", secondaryIcon: "/img/sistema lavadero de automovil.png", sections: ["Bahías", "Vehículos", "Servicios", "Tiempos", "Comisiones", "Caja"] }
  ];

  var featured = [
    {
      id: "featuredTallerMotos",
      module: "taller_mecanico",
      title: "Taller de motos",
      fullTitle: "Taller de motos",
      lead: "Órdenes de trabajo, diagnóstico, repuestos, mano de obra y entrega.",
      description: "Gestión de motocicletas, órdenes de trabajo, diagnóstico, repuestos, mano de obra, garantías, caja y entrega al cliente.",
      icon: "/img/settings-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/taller-mecanico.jpg",
      sections: ["Motocicletas", "Órdenes de trabajo", "Diagnóstico", "Repuestos", "Mano de obra", "Entrega"]
    }
  ];

  var retired = [
    "parqueadero", "parqueaderos_con_ticket_qr", "domicilios", "domicilios_y_entregas",
    "alquileres", "alquileres_de_activos", "aiu_construccion", "construccion_aiu",
    "eventos_boleteria", "eventos_y_boleteria", "salon_spa", "salon_barberia_y_spa",
    "veterinaria_petshop", "veterinaria_y_pet_shop", "lavanderia_tintoreria",
    "lavanderia_y_tintoreria", "transporte_carga_tms", "servicios_tecnicos",
    "funeraria_exequial", "funeraria_y_servicios_exequiales", "parque_recreativo",
    "parque_recreativo_y_atracciones", "turnos_atencion", "turnos_de_atencion"
  ];

  window.PCS_PRECONFIGURACIONES_BASICAS = items;
  window.PCS_PRECONFIGURACIONES_BASICAS_KEYS = items.map(function (item) { return item.module; });
  window.PCS_SISTEMAS_DESTACADOS = featured;
  window.PCS_PORTAL_RETIRED_CARD_KEYS = retired;
})();
