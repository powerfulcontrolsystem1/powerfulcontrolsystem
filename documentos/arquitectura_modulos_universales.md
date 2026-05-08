# Arquitectura de modulos universales

## Regla principal

Todo modulo del sistema debe nacer como un nucleo universal reutilizable. Los tipos de empresa no deben duplicar logica; solo activan permisos, licencias, plantillas, nombres visibles, datos iniciales y configuraciones recomendadas.

## Capas

- Nucleo universal: rutas, tablas, permisos, validaciones, auditoria, reportes y reglas de negocio compartidas.
- Plantilla por tipo de empresa: licencias disponibles, permisos iniciales, datos semilla, etiquetas visibles y flujos sugeridos.
- Experiencia de usuario: menus, textos y agrupaciones que explican el modulo segun el negocio sin cambiar la clave interna.

## Criterios para crear o ampliar modulos

- Si el flujo aplica a varios negocios, se amplia el modulo universal existente.
- Si el flujo necesita campos especiales, se agregan configuraciones o subtipos dentro del modulo universal.
- Si el negocio requiere datos iniciales, se agregan semillas por tipo de empresa.
- Si la restriccion afecta facturacion, documentos, usuarios, activos o ventas, se implementa como regla configurable por licencia o por empresa.
- Si se necesita un nombre comercial diferente, se cambia la etiqueta visible, no la ruta ni la clave estable.

## Ejemplos aplicados

- Alquiler universal cubre herramientas, motos, equipos, espacios, vehiculos y cualquier objeto alquilable.
- Inventario universal cubre productos, servicios, insumos, lotes, bodegas, combos y costeo.
- Operacion universal cubre estaciones, carritos, venta directa, turnos, reservas y venta publica.
- Finanzas universales cubren caja, bancos, cartera, egresos, ingresos, contabilidad, impuestos y reportes.
- CRM universal cubre clientes, embudos, seguimiento, cartera comercial y comunicaciones.
- Personas y activos cubre usuarios, empleados, carnets, asistencia, vehiculos, equipos e historial operativo.

## Bloques canonicos del sistema

- Acceso general: inicio y panel principal.
- Soluciones universales por negocio: plantillas y capacidades especializadas activadas por licencia.
- Operacion universal y ventas: puntos de venta, carritos, estaciones, reservas, turnos y canales publicos.
- CRM universal y clientes: clientes, embudos, comunicaciones y cartera comercial.
- Inventario y compras universales: productos, servicios, compras, bodegas, logistica, produccion y costeo.
- Finanzas universales y cumplimiento: caja, bancos, cartera, contabilidad, impuestos, facturacion y reportes.
- Personas y activos universales: usuarios, empleados, asistencia, carnets, vehiculos, equipos e historial.
- Analisis universal y control: auditoria, calidad, procesos, indicadores, backups y control ejecutivo.
- Documentos universales, nube y soporte: documentos, contratos, aprobaciones, nube, soporte remoto y helpdesk.
- Administracion universal: configuracion, seguridad, integraciones, sensores, tarifas y reglas operativas.

## Regla de integridad tecnica

La capa interna puede conservar claves historicas para no romper rutas, permisos ni licencias. La capa visible y las respuestas de API deben exponer los bloques canonicos universales. Las pruebas de backend deben fallar si un bloque legacy vuelve a salir como grupo visible de permisos.

## Lo que no se debe hacer

- No crear modulos duplicados por cada tipo de empresa si el flujo puede vivir en un nucleo universal.
- No cambiar claves internas estables solo para mejorar un nombre visible.
- No acoplar licencias a una sola industria cuando la capacidad puede parametrizarse.
- No repetir permisos, endpoints o tablas si basta con un subtipo o configuracion.

## Checklist antes de agregar un modulo

- Existe un modulo universal que ya cubra el 70% del flujo.
- La licencia puede activar la capacidad sin crear una rama especial.
- Los permisos usan grupos comunes y acciones comunes.
- La interfaz explica el contexto del negocio sin duplicar pantallas.
- Los reportes pueden filtrar por empresa, tipo de activo, tipo de documento o subtipo operativo.
