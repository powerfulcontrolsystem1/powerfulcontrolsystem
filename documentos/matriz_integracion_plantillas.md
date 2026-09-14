# Matriz de integracion de plantillas

Estado: Sustituido. Responsable: Ingeniería de módulos. Revisión documental: 2026-09-14.

La matriz comercial anterior de trece plantillas dejó de ser la fuente vigente.
Se conserva esta ruta únicamente para evitar enlaces documentales rotos.

La definición actual está en la
[matriz de preconfiguraciones básicas](matriz_preconfiguraciones_basicas.md):
Hotel, Motel, Restaurante, Bar, Pymes, Salón de belleza y Lavadero de autos.
Taller de motos permanece como sistema destacado del portal, no como una octava
preconfiguración básica.

El núcleo operativo continúa siendo único: clientes, inventario, ventas, pagos,
finanzas, facturación, reportes y seguridad. Toda consulta o mutación empresarial
mantiene `empresa_id`, permisos efectivos y límites de licencia.

Los catálogos históricos de Parqueadero, Domicilios, Alquileres, Construcción/AIU,
Eventos, Veterinaria, Lavandería/tintorería, Transporte, Servicios técnicos,
Funeraria y Parque recreativo no se publican en el index ni en su landing. La
normalización de `pagina_principal_handlers.go` los descarta incluso si una
configuración anterior del super administrador aún los contiene.
