# Matriz de preconfiguraciones basicas

Estado: Vigente. Responsable: Ingeniería de módulos. Revisión documental: 2026-09-14.

## Alcance

Estas son las sugerencias iniciales disponibles al crear una empresa. No son
sistemas separados: configuran el núcleo común de inventario, ventas, pagos,
finanzas, usuarios y reportes sin duplicar datos ni relajar el aislamiento por
`empresa_id`.

| Preconfiguracion | Clave publica | Base inicial |
| --- | --- | --- |
| Hotel | `hotel` | Habitaciones, reservas, tarifas por dia, recepcion, consumos y caja |
| Motel | `motel` | Habitaciones por turnos, tarifas por tiempo, minibar, recepcion y aseo |
| Restaurante | `restaurante` | Mesas, pedidos, cocina, productos, caja y facturacion |
| Bar | `bar` | Mesas, barra, bebidas, comandas, caja y reportes |
| Pymes | `pymes` | Venta directa, productos, servicios, clientes, caja y reportes |
| Salon de belleza | `salon_belleza` | Agenda, sillas, profesionales, servicios, comisiones y caja |
| Lavadero de autos | `lavadero_autos` | Bahias, vehiculos, servicios, tiempos, comisiones y caja |

`Taller de motos` se conserva como sistema destacado del portal con clave
`taller_mecanico`. No aumenta a ocho el catálogo de preconfiguraciones básicas.

## Contrato publico

1. El index y la landing descriptiva agregan exactamente estas siete sugerencias
   y el sistema destacado Taller de motos.
2. Configuraciones antiguas no pueden volver a publicar Parqueadero, Domicilios,
   Alquileres, Construccion/AIU, Eventos, Veterinaria, Lavanderia/tintoreria,
   Transporte, Servicios tecnicos, Funeraria, Parque recreativo ni Turnos.
3. Cada página cubierta declara nodos de apariencia y responde al modo claro u
   oscuro.
4. Los módulos generales del núcleo pueden seguir apareciendo como capacidades
   del sistema; no se presentan como plantillas empresariales adicionales.

## Fuentes y verificacion

- Catálogo: `web/js/preconfiguraciones_basicas_catalogo.js`.
- Portada: `web/index.html` y `web/descripcion_de_los_sistemas.html`.
- Normalización persistida: `backend/handlers/pagina_principal_handlers.go`.
- Contratos: `backend/main_portal_navigation_appearance_contract_test.go` y
  `backend/handlers/pagina_principal_catalogo_test.go`.
