# Produccion / MRP

Fecha: 2026-06-11
Estado: modulo operativo reforzado

## Alcance

El modulo `produccion_mrp` agrega una capa profesional de manufactura y planeacion de materiales por empresa. Esta pensado para negocios que producen, ensamblan, preparan kits, alimentos, amenidades hoteleras, productos internos o paquetes comerciales y necesitan controlar recetas, ordenes, consumos y calidad sin duplicar inventario ni compras.

## Superficies

- Administracion: `web/administrar_empresa/produccion_mrp.html`
- Tutorial: `web/administrar_empresa/produccion_mrp_tutorial.html`
- API privada: `/api/empresa/produccion_mrp`
- Permisos: `WithEmpresaProduccionMRPPermissions`
- Licencia: `produccion_mrp`
- Menu: Administrar empresa > Inventario y compras > Produccion / MRP

## Funcionalidad incluida

- Configuracion por empresa: nombre, moneda, modo de costeo, aprobacion de ordenes, consumo de inventario al iniciar y cierre con calidad.
- Recetas/BOM: codigo, version, producto terminado, unidad, cantidad base, costo estandar, merma, tiempo estimado y componentes.
- Ordenes de produccion: codigo automatico, receta, cantidad planificada, prioridad, responsable, estado, costos estimados/reales y fechas operativas.
- Consumos: registro de materiales por orden, cantidad planificada/consumida, lote, costo unitario, costo total y usuario creador.
- Calidad: resultado pendiente/aprobado/rechazado/reproceso, checklist JSON, cantidades aprobadas/rechazadas, responsable y observaciones.
- MRP: generacion base de requerimientos por periodo desde recetas activas con demanda estimada, stock de seguridad y cantidad sugerida a producir.
- Dashboard: KPIs de recetas activas, ordenes abiertas, ordenes en calidad, cerradas, costo abierto y costo real del mes.
- Dashboard profesional: KPIs, flujo por etapa, alertas, siguiente accion recomendada, resumen del plan MRP y actividad reciente.
- Datos demo idempotentes: recetas y ordenes de ejemplo para kit hotelero, paquete de menta PCS y hamburguesa clasica, con componentes, costos, merma, responsables y calidad pendiente.
- Tutorial operativo: pasos para configurar reglas, crear BOM, abrir orden, registrar consumo, controlar calidad y generar MRP.
- Integracion con Productos: lista recetas vendibles activas (`recetas_productos`) y permite importarlas como BOM productiva con codigo `POS-*`.

## Separacion frente a Productos

No se debe duplicar la responsabilidad del modulo `Inventario y productos`.

- Productos mantiene catalogo, precios, impuestos, bodegas, existencias, servicios y recetas vendibles del POS.
- Recetas vendibles (`recetas_productos`) se usan para vender combos/platos/servicios compuestos en el carrito y pueden descontar ingredientes.
- Produccion/MRP se usa cuando una receta o producto requiere fabricacion formal: ordenes, responsable, fecha, consumos reales, calidad, costos y plan de materiales.
- Cuando una receta vendible tambien necesita planificacion productiva, se importa desde MRP con `import_receta_producto` para evitar doble digitacion.
- Los endpoints genericos antiguos de `produccion_bom` se consideran compatibilidad tecnica; el flujo operativo principal es `produccion_mrp`.

## Buenas practicas aplicadas

- Inspirado en patrones de ERP/MRP: lista de materiales BOM, ordenes de produccion, costos estandar/reales, etapas, calidad y requerimientos sugeridos.
- El boton `Cargar demo` actualiza recetas por codigo cuando ya existen, en lugar de duplicarlas o fallar por llave unica.
- Las ordenes demo usan marcador `PCS-DEMO-MRP` para evitar crear duplicados abiertos cada vez que se carga el ejemplo.
- El plan MRP se genera para el periodo actual y tambien conserva un periodo `demo` para pruebas guiadas.

## Aislamiento multiempresa

Todas las tablas del modulo tienen `empresa_id` y las consultas lo filtran de forma obligatoria. La ruta privada usa el wrapper central de permisos empresariales, por lo que valida contexto de empresa, licencia, rol y pagina antes de ejecutar el handler.

## Tablas

- `empresa_produccion_mrp_config`
- `empresa_produccion_recetas`
- `empresa_produccion_receta_componentes`
- `empresa_produccion_ordenes`
- `empresa_produccion_consumos`
- `empresa_produccion_calidad`
- `empresa_produccion_mrp_plan`

## Notas de integracion

Esta primera entrega no descuenta inventario real automaticamente para evitar movimientos contables o de stock sin reglas de costeo confirmadas. El modulo deja listos los campos de producto, lote, costo y cantidades para una siguiente fase de integracion con Kardex, compras y asientos contables.
