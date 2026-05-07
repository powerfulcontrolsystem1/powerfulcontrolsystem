# Modulos empresariales Colombia

Fecha: 2026-05-06

## Alcance

Se implementa una familia de seis modulos empresariales faltantes comparables con funcionalidades de suites ERP/contables usadas en Colombia:

- `bancos_pagos`: bancos, conciliacion y pagos masivos.
- `gestion_documental`: expedientes, versiones, aprobaciones y vencimientos.
- `cumplimiento_kyc`: debida diligencia KYC/KYB, riesgo LAFT y alertas.
- `contratos_obligaciones`: contratos, polizas, hitos, renovaciones y firma externa/manual.
- `helpdesk`: mesa de ayuda, tickets, SLA, prioridades y evidencias.
- `calidad_procesos`: procesos, auditorias, no conformidades y acciones correctivas.

## Arquitectura

Los seis modulos comparten el mismo nucleo tecnico para evitar duplicacion:

- Backend comun: `backend/db/modulos_empresariales_colombia.go`.
- Handler comun parametrizado: `backend/handlers/modulos_empresariales_colombia.go`.
- Frontend comun: `web/js/modulo_colombia_admin.js`.
- Pantallas livianas por modulo en `web/administrar_empresa/*.html`.

Cada registro tiene `empresa_id`, `modulo`, `codigo`, `tipo`, tercero/area, responsable, categoria, referencia, prioridad, estado, fechas, valor, metadata JSON, usuario creador y auditoria temporal. La bitacora registra eventos por `empresa_id`, modulo y registro.

## Integracion

- APIs privadas:
  - `/api/empresa/bancos_pagos`
  - `/api/empresa/gestion_documental`
  - `/api/empresa/cumplimiento_kyc`
  - `/api/empresa/contratos_obligaciones`
  - `/api/empresa/helpdesk`
  - `/api/empresa/calidad_procesos`
- Acciones: `dashboard`, `registros`, `eventos`, `registro` y `seed_demo`.
- Acciones profesionales compartidas:
  - `plantilla`: devuelve tipos, categorias, estados, etiquetas y metadata sugerida segun el modulo.
  - `reporte`: entrega analitica por estado, tipo, categoria, prioridad, vencimientos, valor pendiente, criticidad y recomendaciones.
  - `agenda`: concentra vencidos, proximos vencimientos, tareas, aprobaciones y recomendaciones operativas.
  - `responsables`: resume carga por responsable, tareas, registros y aprobaciones pendientes.
  - `sla`: calcula cumplimiento, semaforo, vencidos, proximos vencimientos y buckets de antiguedad.
  - `riesgo`: calcula score operativo, nivel de riesgo, factores y recomendaciones.
  - `exportacion`: genera un paquete CSV de auditoria con resumen, registros, agenda, responsables, SLA, riesgo, tareas, aprobaciones, evidencias y bitacora.
  - `buscar`: filtra registros desde backend por texto, estado, tipo, categoria, prioridad, responsable, vencidos y proximos vencimientos.
  - `accion_masiva`: actualiza estado, prioridad o responsable en varios registros seleccionados y deja bitacora individual.
  - `expediente`: consolida registro, eventos, evidencias, aprobaciones y tareas en una vista 360.
  - `estado`: cambia rapidamente el estado de un registro y deja bitacora.
  - `evento`: registra seguimiento, comentario, evidencia, aprobacion o cierre sin duplicar tablas.
  - `evidencias`: lista soportes, enlaces, actas, fotos, contratos o documentos asociados a un registro.
  - `evidencia`: agrega una evidencia al registro y registra el evento `evidencia_agregada`.
  - `aprobaciones`: lista aprobaciones por modulo, registro y estado.
  - `aprobacion_solicitar`: crea una solicitud de aprobacion con nivel, destinatario, vencimiento y comentario.
  - `aprobacion_decidir`: aprueba o rechaza una solicitud y deja trazabilidad.
  - `tareas`: lista compromisos por modulo, registro y estado.
  - `tarea`: crea una tarea con responsable, prioridad, vencimiento y comentario.
  - `tarea_estado`: actualiza una tarea a pendiente, en proceso, cumplida o cancelada.
  - `cierre_controlado`: cierra un registro solo si tiene evidencias, no tiene aprobaciones pendientes y no tiene tareas abiertas.
  - `generar_plan_accion`: convierte alertas de agenda en tareas accionables, evitando duplicados abiertos.
  - `importar_registros`: importa hasta 1000 registros por lote desde CSV/JSON y registra el evento de importacion.
- Menus:
  - Finanzas y cumplimiento: Bancos y pagos, KYC/KYB y LAFT.
  - Documentos, nube y soporte: Gestion documental, Contratos, Helpdesk.
  - Analisis y control: Calidad y procesos.
- Licencias: cada modulo se activa/desactiva con su propia clave en `licencias.modulos_habilitados`.
- Roles: lectura transversal; creacion/actualizacion/aprobacion para `admin_empresa`, `supervisor_sucursal`, `contabilidad` y `auditor`; eliminacion solo para `admin_empresa`, salvo overrides finos.

## Reglas de gobierno

- No se duplican terceros, productos, usuarios, contabilidad ni inventario.
- La metadata flexible permite particularidades por empresa sin crear tablas repetidas para cada variacion.
- Las pantallas usan variables centrales de tema y se adaptan a modo claro, oscuro y variantes configuradas.
- Los datos demo se cargan por modulo y por empresa, sin mezclar informacion entre companias.

## Fase 2 implementada

Se agrega una capa profesional comun para los seis modulos:

- Plantillas operativas por modulo con tipos, categorias, estados y etiquetas de negocio.
- Formulario dinamico que cambia su lenguaje segun bancos, documentos, KYC, contratos, helpdesk o calidad.
- Seguimiento por registro con bitacora, usuario, detalle y fecha.
- Cambio rapido de estado con auditoria de estado anterior y nuevo.
- Filtro por estado y exportacion CSV desde la misma pantalla.

Esto mantiene una sola implementacion de UI, handler y tablas, evitando seis copias del mismo codigo.

## Fase 3 implementada

Se agrega reporte ejecutivo comun para operar los modulos en produccion:

- Agrupacion por estado, tipo, categoria y prioridad.
- Indicadores de vencidos, proximos vencimientos a 7 y 30 dias, criticos abiertos y registros sin responsable.
- Valor pendiente y valor vencido por empresa y modulo.
- Recomendaciones automaticas para priorizar vencidos, criticidad y asignacion de responsables.
- Visualizacion en la pantalla compartida de cada modulo sin duplicar codigo frontend.

## Fase 4 implementada

Se agrega importacion masiva profesional:

- Boton `Importar CSV` en cada pantalla del modulo.
- Reutiliza el formato exportado por `Exportar CSV` para permitir correcciones y recarga.
- Backend comun `importar_registros` con limite de 1000 filas por lote.
- Guardado por `empresa_id` y `modulo`, con validacion de codigo/nombre y bitacora de importacion.
- En caso de errores parciales, la pantalla informa cuantas filas guardo y cuales fallaron.

## Fase 5 implementada

Se agrega gestion profesional de evidencias y soportes:

- Tabla compartida `empresa_modulos_colombia_evidencias`, aislada por `empresa_id`, `modulo` y `registro_id`.
- Soportes por tipo: soporte, contrato, foto, acta, documento y enlace.
- Pantalla compartida para agregar URL/ruta, descripcion y usuario responsable.
- Bitacora automatica `evidencia_agregada` al adjuntar cada soporte.
- Validacion para evitar evidencias sobre registros inexistentes o de otra empresa/modulo.

## Fase 6 implementada

Se agrega flujo profesional de aprobaciones:

- Tabla compartida `empresa_modulos_colombia_aprobaciones`, aislada por `empresa_id`, `modulo` y `registro_id`.
- Solicitudes por nivel: operativo, supervisor, contable, gerencia, cumplimiento y juridico.
- Campos de destinatario, solicitante, comentario y fecha de vencimiento.
- Decision aprobada/rechazada con usuario decisor y fecha de decision.
- Al aprobar, el registro pasa a `aprobado` cuando no esta cerrado/cancelado/resuelto.
- Bitacora automatica `aprobacion_solicitada` y `aprobacion_decidida`.

## Fase 7 implementada

Se agrega gestion de tareas y compromisos:

- Tabla compartida `empresa_modulos_colombia_tareas`, aislada por `empresa_id`, `modulo` y `registro_id`.
- Tareas con titulo, responsable, prioridad, vencimiento, comentario y estado.
- Estados operativos: pendiente, en proceso, cumplida y cancelada.
- Pantalla compartida para crear tareas y cambiar su estado.
- Bitacora automatica `tarea_creada` y `tarea_actualizada`.

## Fase 8 implementada

Se agrega expediente 360 por registro:

- Accion `expediente` para consultar un registro con toda su trazabilidad.
- Consolida eventos, evidencias, aprobaciones y tareas por `empresa_id`, `modulo` y `registro_id`.
- Resumen ejecutivo con conteo de evidencias, aprobaciones pendientes y tareas abiertas.
- Recomendacion automatica del siguiente paso operativo.
- Boton `Expediente` en la tabla principal de cada modulo.

## Fase 9 implementada

Se agrega agenda y alertas operativas:

- Accion `agenda` para listar registros vencidos, proximos vencimientos, tareas vencidas y aprobaciones pendientes.
- Semaforo de severidad critica, alta y media.
- Recomendaciones automaticas para priorizar vencidos, aprobaciones y vencimientos de los proximos 7 dias.
- Vista `Agenda y alertas` en la pantalla compartida de cada modulo.
- Enlace directo desde cada alerta al expediente 360 del registro.

## Fase 10 implementada

Se agrega cierre controlado:

- Accion `cierre_controlado` para cerrar registros con reglas de gobierno.
- Exige al menos una evidencia o soporte.
- Bloquea el cierre si existen aprobaciones pendientes.
- Bloquea el cierre si existen tareas pendientes o en proceso.
- Registra bitacora `cierre_controlado` con estado anterior, estado nuevo y usuario.
- Boton `Cerrar validado` desde el expediente 360.

## Fase 11 implementada

Se agrega generador de plan de accion:

- Accion `generar_plan_accion` basada en la agenda del modulo.
- Convierte alertas criticas/altas/medias en tareas pendientes por registro.
- Asigna prioridad segun severidad: critica -> urgente, alta -> alta, media -> normal.
- Evita duplicar tareas abiertas con el mismo titulo para el mismo registro.
- Registra bitacora `plan_accion_generado`.
- Boton `Generar plan de accion` en la agenda.

## Fase 12 implementada

Se agrega tablero de responsables y carga:

- Accion `responsables` para calcular carga por responsable o destinatario.
- Cuenta registros abiertos, registros vencidos, tareas abiertas, tareas vencidas y aprobaciones pendientes.
- Calcula total pendiente y recomendacion por responsable.
- Vista `Responsables y carga` en la pantalla compartida.
- Agrupa registros sin asignacion en `Sin responsable`.

## Fase 13 implementada

Se agrega SLA y cumplimiento:

- Accion `sla` con total abierto, vencidos, proximos 7 dias, sin vencimiento, tareas abiertas y tareas vencidas.
- Calcula porcentaje de cumplimiento y semaforo verde/amarillo/rojo.
- Buckets de vencimiento: sin vencimiento, vencido, 0-7, 8-30 y mas de 30 dias.
- Recomendaciones automaticas para vencidos, registros sin SLA y compromisos de la semana.
- Vista `SLA y cumplimiento` en la pantalla compartida.

## Fase 14 implementada

Se agrega matriz de riesgo operativo:

- Accion `riesgo` con score de 0 a 100 y nivel bajo/medio/alto.
- Factores ponderados: vencidos, criticos abiertos, aprobaciones pendientes, tareas vencidas, tareas abiertas, sin responsable y sin evidencia.
- Recomendaciones automaticas segun los factores detectados.
- Vista `Matriz de riesgo operativo` en la pantalla compartida.

## Fase 15 implementada

Se agrega exportacion profesional para auditoria y comites:

- Accion `exportacion` que consolida informacion operativa por `empresa_id` y `modulo`.
- CSV multi-seccion con resumen ejecutivo, registros, agenda, responsables, metricas, SLA, riesgo, tareas, aprobaciones, evidencias y bitacora.
- Boton `Exportar auditoria CSV` desde la pantalla compartida.
- La exportacion sale desde el backend comun para que los seis modulos mantengan el mismo formato y no dupliquen codigo.

## Fase 16 implementada

Se agrega busqueda avanzada compartida:

- Accion `buscar` con filtros reales de backend por texto, estado, tipo, categoria, prioridad, responsable, vencidos y proximos vencimientos.
- Barra de busqueda profesional en la pantalla comun, con limpieza de filtros y ejecucion por boton o tecla Enter.
- La busqueda conserva aislamiento por `empresa_id` y `modulo`, sin mezclar registros de otras empresas ni crear consultas duplicadas por modulo.
- El orden prioriza registros urgentes/criticos y vencimientos mas cercanos para facilitar operacion diaria.

## Fase 17 implementada

Se agregan acciones masivas controladas:

- Accion `accion_masiva` para actualizar estado, prioridad y responsable sobre varios registros seleccionados.
- Limite de 200 registros por operacion para proteger el backend y evitar cambios accidentales demasiado amplios.
- Cada registro actualizado deja bitacora `accion_masiva` con estado anterior, estado nuevo, usuario y detalle operativo.
- La pantalla comun incorpora seleccion por fila, seleccion total visible y contador de registros seleccionados.
