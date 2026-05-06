# Produccion / MRP

Fecha: 2026-05-06
Estado: primera entrega funcional

## Alcance

El modulo `produccion_mrp` agrega una capa profesional de manufactura y planeacion de materiales por empresa. Esta pensado para negocios que producen, ensamblan, preparan kits, alimentos, amenidades hoteleras, productos internos o paquetes comerciales y necesitan controlar recetas, ordenes, consumos y calidad sin duplicar inventario ni compras.

## Superficies

- Administracion: `web/administrar_empresa/produccion_mrp.html`
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
- Datos demo: receta de kit hotelero, orden inicial y plan MRP demo.

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
