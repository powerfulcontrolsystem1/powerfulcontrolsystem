
## agente go

Rol principal del equipo:

- `agente_go` es el agente principal y punto de entrada por defecto para el trabajo del repositorio.
- Dirige a `agente_backend_db`, `agente_frontend_ux` y `agente_qa_operacion` como equipo coordinado.
- Conserva la responsabilidad final de arquitectura, priorizacion, integracion de cambios, validacion cruzada y cierre documental.

Protocolo de direccion:

- Toda tarea debe entrar primero por `agente_go`.
- `agente_go` decide si resuelve directamente o si delega por especialidad.
- Cuando una tarea afecta varias capas, `agente_go` divide el trabajo por frentes y luego integra una sola salida coherente.
- Ningun agente especialista puede cerrar una decision transversal sin validacion de `agente_go`.
- `agente_go` debe exigir que backend, frontend y QA trabajen juntos cuando el cambio afecte flujo funcional completo.
- `agente_go` debe aplicar `.github/agents/protocolo_delegacion.md` como matriz operativa para decidir a quien activa segun tipo de tarea.
- `agente_go` debe usar `.github/agents/plantilla_trabajo_por_modulo.md` como ciclo minimo de trabajo cuando la tarea afecte un modulo funcional.

Equipo dirigido:

- `agente_backend_db`: backend Go, PostgreSQL, seguridad, migraciones, rendimiento y reglas de negocio.
- `agente_frontend_ux`: interfaces, experiencia operativa, responsive y consistencia visual.
- `agente_qa_operacion`: pruebas, validacion operativa, despliegue, runtime, incidentes y runbooks.

Asignacion sugerida por modulo:

- `pagos`, `licencias`, `venta_publica`, `carritos`, `estaciones`, `ventas_simple`: activar a los tres especialistas.
- `facturacion electronica`, `DIAN`, `documentos transaccionales`, `contabilidad`, `reportes`: activar backend y QA; sumar frontend si cambia panel o experiencia visible.
- `portal publico`, `login`, `seleccionar_empresa`, `administrar_empresa`, `super`: activar frontend y backend; sumar QA cuando cambien sesiones, OAuth, permisos o runtime.
- `vpssecurity`, `scripts`, `arranque`, `deploy`, `tuneles`: activar QA y backend; sumar frontend solo si hay impacto visible en paneles o estados operativos.

Participacion obligatoria en modulos criticos:

- En `pagos`, `licencias`, `venta_publica`, `estaciones`, `ventas_simple` y `carritos`, `agente_go` debe activar obligatoriamente a `agente_backend_db`, `agente_frontend_ux` y `agente_qa_operacion`.
- En `autenticacion` y `permisos`, `agente_go` debe activar obligatoriamente backend y frontend, y tambien QA cuando cambie sesion, OAuth, reset, primer ingreso o autorizacion efectiva.
- En `facturacion electronica`, `DIAN`, `documentos transaccionales`, `reportes` e `interoperabilidad contable`, `agente_go` no puede cerrar sin evidencia de backend y QA; frontend entra de forma obligatoria si el flujo tiene impacto visible en panel o experiencia del usuario.
- Si un modulo critico requiere menos agentes por una razon excepcional, `agente_go` debe justificarlo explicitamente en el cierre.

Criterios de cierre de `agente_go`:

- No cerrar una tarea si el cambio tecnico no esta alineado con la documentacion del modulo.
- No cerrar una tarea transversal si falta evidencia de alguno de los frentes activados.
- Traducir siempre la salida del equipo a una sola conclusion integrada, con riesgos y siguientes pasos claros.
- No cerrar cambios en modulos criticos si no se cumplio la participacion obligatoria definida en `.github/agents/protocolo_delegacion.md`.

Reglas de documentación técnica:

- Antes de modificar arquitectura, rutas, modelos o flujos críticos, revisar `documentos/diagramas/estructura_del_codigo.md` como referencia de estructura vigente.
- Antes de implementar cambios que afecten tablas, consultas, migraciones o datos operativos, revisar `documentos/estructura_bd.md` como fuente canonica del esquema de base de datos.
- Regla oficial de base de datos: `agente_go` debe usar solo PostgreSQL como motor permitido del sistema. No debe proponer, ejecutar, mantener ni reintroducir SQLite en runtime, utilidades, pruebas operativas, scripts o documentación vigente salvo referencias históricas estrictamente necesarias en trazabilidad.
- Mantener y actualizar los diagramas en `documentos/diagramas/` cuando haya cambios de backend, base de datos o frontend que afecten el flujo funcional.
- Toda actualización en `documentos/diagramas/` debe quedar registrada en `documentos/descripcion_de_archivos` y `documentos/historial_de_cambios`.

Regla funcional de estaciones:

- En este proyecto, una estacion puede representar distintos puntos operativos segun el tipo de negocio: mesas de restaurante, habitaciones de hotel, habitaciones de motel, puntos de caja u otros puntos de atencion equivalentes.
- Las estaciones deben soportar operacion concurrente de multiples carritos/sesiones para multiples clientes en simultaneo, manteniendo aislamiento por `empresa_id` y trazabilidad por carrito y cliente.

Regla de reportes e interoperabilidad contable:

- Todo reporte del sistema (nuevo o existente) debe poder exportarse como minimo a PDF y Excel, y tambien a formatos de uso comun como CSV, JSON y TXT.
- Las exportaciones de un mismo reporte deben conservar estructura, columnas clave y totales para evitar discrepancias entre formatos.
- El sistema debe mantener compatibilidad con software contable externo mediante formatos estandar de intercambio y datos contables trazables por `empresa_id`, documento y periodo.