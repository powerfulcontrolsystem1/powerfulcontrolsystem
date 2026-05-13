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

## Verticales empresariales 2026-05-10

Se agregaron 20 verticales nuevos sobre el motor comun de `empresa_modulos_colombia_*`: viajes, turismo, eventos, salon/spa, veterinaria, clinica, laboratorio, colegio, guarderia, lavanderia, taller, transporte TMS, servicios tecnicos, inmobiliaria, seguridad privada, club deportivo, funeraria, parque recreativo, cooperativa y capacitacion empresarial.

Cada vertical usa una plantilla propia de tipos, categorias, estados, acciones sugeridas y metadata, pero comparte dashboard, agenda, SLA, riesgo, evidencias, aprobaciones, tareas, importacion/exportacion y auditoria. La activacion se controla por licencia y por la matriz de roles/paginas del super administrador.

En `Administrar empresa > Soluciones por negocio`, cada modulo por tipo de negocio aparece como boton propio, no como un unico agrupador. El submenu visual de cada negocio no duplica pantallas por industria: interpreta cada seccion como una intencion operativa comun (`dashboard`, `registros`, `seguimiento`, `responsables`, `aprobacion`, `evidencia` o `control`) y abre la zona correspondiente del motor universal. La ruta de trabajo viaja en la plantilla backend como `secciones_flujo`; el catalogo visual conserva iconos y textos de portada, pero la pantalla operativa ya no muestra una ruta numerada superior para no duplicar el submenu de botones.

La configuracion visible del vertical se obtiene de la plantilla backend: tipos, categorias, estados, acciones sugeridas, etiquetas y metadata. La UI permite descargar una plantilla CSV de carga desde el mismo motor comun, evitando formularios especiales por cada industria.

Antes de operar o importar datos, el modulo muestra un diagnostico de preparacion exportable. La validacion principal se sirve desde el backend con `action=diagnostico` sobre la misma ruta `/api/empresa/<modulo>`, revisando contexto de empresa, modulo soportado, base de datos, completitud de plantilla, metadata JSON y existencia de registros. La UI mantiene un fallback local para que la pantalla no se rompa si una version antigua del backend aun no expone el endpoint.

El bootstrap `EnsureNuevosVerticalesTipoEmpresaYLicencias` registra los tipos de empresa, sus licencias comerciales y la preconfiguracion inicial para que una empresa nueva pueda nacer con tipo, roles guia, productos/servicios demo, tareas y modulos recomendados.

El lanzador de verticales puede consultar `/api/empresa/verticales_nuevos/catalogo`, el super administrador puede consultar `/super/api/verticales_nuevos/catalogo` y la portada publica puede consultar `/api/public/verticales_nuevos/catalogo` para obtener el contrato backend completo de los 20 verticales. La respuesta incluye page key, modulo, titulo, resumen, secciones de flujo y plantilla; el archivo visual local queda como respaldo para iconos y experiencia de portada.

Las tarjetas publicas y la pagina de descripcion usan anclas estables por modulo (`vertical-<modulo>`) para que el enlace comercial no dependa del orden de las tarjetas configuradas. Si el administrador cambia las tarjetas de portada, el catalogo publico sigue agregando los 20 verticales y mantiene una ficha descriptiva coherente por negocio.

El flujo de licencias propaga el mismo contrato comercial: las licencias verticales guardan `modulos_habilitados`, el checkout publico expone `modulos_habilitados` y `max_documentos_mensuales`, y las pantallas de elegir/pagar licencia usan esos datos para mostrar industria, icono, tipo de empresa y cupo documental sin duplicar reglas por cada vertical.

El selector de empresas reutiliza el catalogo visual de verticales para evitar reglas aisladas por pantalla: al listar empresas, la tarjeta toma icono, tono y texto operativo del vertical; al crear empresa, el formulario muestra una vista previa con las secciones del negocio antes de guardar y aplicar la preconfiguracion inicial.

Las pantallas de super administrador para tipos y preconfiguraciones tambien leen el mismo catalogo visual. Esto permite auditar los 20 verticales con conteos, etiquetas e indicadores de flujo sin mantener una lista paralela en cada vista administrativa.

La ayuda administrativa y el contexto de IA forman parte de la arquitectura. Cuando se agregan verticales, la ayuda privada del super administrador debe explicar catalogo, activacion, licencias, permisos y operacion; el contexto canonico de IA debe nombrar el motor comun, endpoints de catalogo, cupos documentales y regla de no crear modulos duplicados si basta con plantilla o preconfiguracion.

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
