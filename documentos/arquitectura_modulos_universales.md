# Arquitectura de modulos universales

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se conserva el principio de núcleo reutilizable y se retira el catálogo obsoleto de veinte plantillas.
- Las reglas de integración son requisitos de arquitectura a comprobar en cada flujo, no una garantía universal de interoperabilidad.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

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
- Inventario universal cubre productos, servicios, insumos, lotes, bodegas, recetas y costeo.
- Operacion universal cubre estaciones, carritos, venta directa, turnos, reservas y venta publica.
- Finanzas universales cubren caja, bancos, cartera, egresos, ingresos, contabilidad, impuestos y reportes.
- CRM universal cubre clientes, embudos, seguimiento, cartera comercial y comunicaciones.
- Personas y activos cubre usuarios, empleados, carnets, asistencia, vehiculos, equipos e historial operativo.

## Catálogo y especialización

El catálogo actual contiene trece plantillas: cuatro clásicas y nueve nuevas.
La [matriz de integración](matriz_integracion_plantillas.md) mantiene sus nombres,
estado y retiradas; no reutilizar la lista histórica de veinte como catálogo
comercial. Tipos, estados y metadatos viven en el motor común y sus plantillas.
El diagnóstico de plantilla no demuestra que cada proceso sectorial, integración
o requisito regulatorio esté implementado. Los registros demo son solo para
un entorno aislado autorizado, nunca fuentes de venta, nómina o emisión fiscal.

## Bloques canonicos del sistema

- Acceso general: inicio y panel principal.
- Soluciones universales por negocio: plantillas y capacidades especializadas activadas por licencia.
- Operacion universal y ventas: puntos de venta, carritos, estaciones, reservas, turnos y canales publicos.
- CRM universal y clientes: clientes, embudos, comunicaciones y cartera comercial.
- Inventario y compras universales: productos, servicios, compras, bodegas, logistica, produccion y costeo.
- Finanzas universales y cumplimiento: caja, bancos, cartera, contabilidad, impuestos, facturacion y reportes.
- Personas y activos universales: usuarios, empleados, asistencia, carnets, vehiculos, equipos e historial.
- Analisis universal y control: auditoria, calidad, procesos, indicadores, backups y control ejecutivo.
- Documentos universales, nube y soporte: documentos, contratos, aprobaciones, nube, soporte remoto y tickets de ayuda propios.
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

## Fuentes y aceptación de la revisión

[modulos_plantillas_nuevas.go](../backend/db/modulos_plantillas_nuevas.go), [empresa_plantillas_integracion.go](../backend/handlers/empresa_plantillas_integracion.go), [matriz_integracion_plantillas.md](matriz_integracion_plantillas.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
