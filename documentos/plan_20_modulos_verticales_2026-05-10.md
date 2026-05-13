# Plan e implementacion de 20 modulos verticales

Fecha: 2026-05-10

## Enfoque profesional

Los 20 modulos se implementan sobre el motor comun de `EmpresaModuloColombiaHandler` y las tablas compartidas `empresa_modulos_colombia_*`. Esto evita duplicar 20 CRUD aislados y deja todos los verticales con la misma base operativa:

- dashboard, KPI y busqueda avanzada;
- registros operativos con estados, tipos y categorias por negocio;
- agenda, SLA, riesgo, responsables y reporte ejecutivo;
- evidencias, aprobaciones, tareas, bitacora, importacion CSV y exportacion;
- permisos por rol, restricciones por licencia y auditoria por empresa.

## Modulos agregados

1. Agencia de viajes y planes turisticos (`agencia_viajes`)
2. Operador turistico local (`operador_turistico`)
3. Eventos y boleteria (`eventos_boleteria`)
4. Salon de belleza, barberia y spa (`salon_spa`)
5. Veterinaria y pet shop (`veterinaria_petshop`)
6. Clinica medica y consultorios multiples (`clinica_consultorios`)
7. Laboratorio clinico (`laboratorio_clinico`)
8. Colegio, academia o instituto (`colegio_academia`)
9. Guarderia y jardin infantil (`guarderia_infantil`)
10. Lavanderia y tintoreria (`lavanderia_tintoreria`)
11. Taller mecanico, motos y autos (`taller_mecanico`)
12. Transporte de carga / TMS (`transporte_carga_tms`)
13. Servicios tecnicos a domicilio (`servicios_tecnicos`)
14. Inmobiliaria comercial (`inmobiliaria_comercial`)
15. Seguridad privada y vigilancia (`seguridad_privada`)
16. Club deportivo y escuela deportiva (`club_deportivo`)
17. Funeraria y servicios exequiales (`funeraria_exequial`)
18. Parque recreativo y atracciones (`parque_recreativo`)
19. Cooperativa y fondo de empleados (`cooperativa_fondo`)
20. Centro de capacitacion empresarial (`capacitacion_empresarial`)

## Roles y permisos

Los modulos quedan en el catalogo de permisos de empresa y en la matriz del super administrador. Regla base:

- `super_administrador` y `administrador_total`: acceso total.
- `admin_empresa`, `supervisor_sucursal`, `cajero`: lectura, creacion, actualizacion y aprobacion.
- Eliminacion: `admin_empresa` y `supervisor_sucursal`.
- Roles operativos restantes: lectura si la licencia y las restricciones finas lo permiten.

## Frontend

Los 20 modulos por tipo de negocio se muestran como botones propios dentro de `Administrar Empresa > Soluciones por negocio`, al mismo nivel visual de Gimnasio, Odontologia, Taxi system, Domicilios y otros negocios existentes. El antiguo boton agrupador `20 verticales nuevos` ya no queda como paso intermedio en el menu principal.

Cada boton vertical abre `modulo_menu.html?module=<modulo>`, con botones internos para areas como dashboard, configuracion, registros, aprobaciones, evidencias, seguimiento, SLA y reportes. El contenido operativo carga en `modulo_colombia.html`, que toma `module`, `title` y `lead` por query string.

El submenu de cada vertical ahora interpreta la intencion de cada seccion (`dashboard`, `registros`, `seguimiento`, `responsables`, `aprobacion`, `evidencia` o `control`) y abre la zona correcta del modulo con `section` e `intent`. La pagina operativa muestra una banda fija de contexto con la seccion activa y accesos rapidos, para que en desktop y celular el usuario sepa si esta trabajando en dashboard, registros, seguimiento, aprobaciones, evidencias o reportes.

La pagina operativa carga `web/js/nuevos_verticales_catalogo.js` para resolver iconos, resumenes y textos de portada, pero ya no muestra el bloque superior de seccion activa ni la ruta numerada dentro del formulario operativo. La navegacion principal queda en el submenu de botones de cada modulo, para que la pagina de trabajo arranque directamente con indicadores, configuracion, diagnostico y registros.

Cada vertical incorpora una seccion de configuracion operativa (`mcConfig`) alimentada por su plantilla backend: tipos de registro, categorias, estados de flujo, acciones sugeridas, etiquetas de tercero/referencia y metadata JSON base. Desde esa misma tarjeta se puede descargar una plantilla CSV de carga y copiar la metadata de ejemplo, manteniendo la configuracion profesional sin duplicar interfaces por industria.

La configuracion incluye diagnostico de preparacion con checklist exportable. La pantalla consulta `action=diagnostico` en el backend para validar empresa detectada, modulo soportado, base de datos operativa, tipos/categorias, estados/acciones, etiquetas, metadata JSON y registros operativos; si el endpoint no responde, conserva un fallback local para no bloquear la operacion.

El catalogo visual de los 20 verticales queda centralizado en `web/js/nuevos_verticales_catalogo.js` para no repetir la matriz en cada pantalla. Adicionalmente, el lanzador empresarial consulta `/api/empresa/verticales_nuevos/catalogo`, la pantalla de licencias del super administrador consulta `/super/api/verticales_nuevos/catalogo` y la portada publica puede consultar `/api/public/verticales_nuevos/catalogo` para alinear page key, titulo, resumen, secciones y plantilla desde backend, conservando el JS visual como respaldo e iconografia local. Lo consumen:

- `web/administrar_empresa/verticales_nuevos_menu.html`
- `web/administrar_empresa/modulo_menu.html`
- `web/js/administrar_empresa.js`
- `web/super/licencias.html`
- `web/index.html`

La cobertura de modulos en licencias ya no incluye 20 checkboxes copiados a mano: se inyectan desde el catalogo central. La pagina publica `index.html` tambien concatena las tarjetas de estos verticales desde ese mismo origen.

La pagina publica de descripcion `descripcion_de_los_sistemas.ht` consume el mismo catalogo y genera fichas profesionales para cada vertical con anclas estables por modulo (`vertical-<modulo>`). Asi, una tarjeta del portal puede abrir directamente su detalle aunque `/api/public/pagina_principal` entregue tarjetas personalizadas desde backend.

El flujo comercial de licencias tambien queda conectado al catalogo: `elegir_licencia.html` identifica licencias verticales por `modulos_habilitados`, tipo o nombre de plan, muestra icono/tono de industria y etiqueta `Vertical`; `pagar_licencia.html` recibe `modulos_habilitados` y `max_documentos_mensuales` desde el resumen publico para mostrar vertical, tipo y cupo de documentos en el checkout.

El selector de empresas tambien usa el catalogo compartido. Las tarjetas de `seleccionar_empresa.html` reconocen los nuevos tipos de empresa y muestran icono, tono, etiqueta y texto operativo del vertical; el formulario de nueva empresa incluye una vista previa del tipo seleccionado con secciones de trabajo antes de guardar.

El panel de super administrador tambien queda alineado: `super/tipos_empresas.html` muestra resumen de tipos, activos, inactivos y verticales 2026, con icono y etiqueta para cada vertical; `super/preconfiguracion_tipos_empresa.html` marca las plantillas verticales y muestra chips de las secciones principales para auditar rapidamente que cada tipo tiene flujo operativo.

La ayuda administrativa y el contexto canonico de IA tambien quedan alineados. `/ayuda/ayuda.html` incluye una seccion `20 verticales nuevos` con catalogo, activacion, operacion, pantallas conectadas y regla de soporte; `defaultContextoIALogicaNegocioText()` explica a la IA que estos verticales usan licencias, roles, `modulos_habilitados`, diagnostico y el motor compartido `empresa_modulos_colombia_*`.

Los botones directos de los verticales en el menu empresarial se filtran por licencia/rol usando el mismo contexto de permisos. El lanzador `verticales_nuevos_menu.html` se conserva como pagina de catalogo y tambien consulta `/api/empresa/permisos_contexto` para filtrar sus tarjetas a las permitidas para la empresa y el rol actuales.

Las pantallas de permisos (`super/permisos_rol.html` y `administrar_empresa/configuracion_permisos.html`) muestran la regla agrupada como `Cualquiera de 20 verticales`, con detalle de los modulos incluidos y una regla legible (`Crear en al menos un vertical permitido`) para evitar que el super administrador vea un modulo vacio o ambiguo.

## Backend

Cada modulo tiene ruta interna:

`/api/empresa/<modulo>?empresa_id=<id>&action=<accion>`

Todas las rutas usan `WithEmpresaModuloVerticalPermissions`, que valida que el modulo pertenezca al catalogo nuevo y aplica `resolveVerticalPermissionAction`.

Los permisos, page keys y rutas de backend se derivan del catalogo `NuevosVerticalesTipoEmpresaCatalog`, evitando listas separadas para los mismos 20 modulos.

El catalogo `NuevosVerticalesTipoEmpresaCatalog` ya no repite nombres ni observaciones de negocio: se construye desde `GetEmpresaModuloColombiaPlantilla`. La metadata separada queda limitada a lo que si cambia para el bootstrap de empresa: prefijo de estaciones, cantidad inicial y roles sugeridos.

Los endpoints `/api/empresa/verticales_nuevos/catalogo`, `/super/api/verticales_nuevos/catalogo` y `/api/public/verticales_nuevos/catalogo` devuelven los 20 verticales con `id`, `page`, `module`, `title`, `summary`, `sections` y `plantilla`. Esto permite que el lanzador visual, licencias, portada publica y las pruebas comparen el contrato real del backend contra la UI sin duplicar matrices.

La accion `diagnostico` queda centralizada en `BuildEmpresaModuloColombiaDiagnostico`. Devuelve puntuacion, estado (`listo` o `revisar`), checks obligatorios/informativos y recomendaciones, incluyendo la ruta de trabajo `secciones_flujo`, de modo que todos los verticales nuevos comparten el mismo control previo antes de importar o operar.

## Activacion por licencia

Si una licencia tiene `modulos_habilitados` vacio, no restringe modulos. Si una licencia declara modulos especificos, se debe agregar la clave correspondiente para habilitar cada nuevo vertical.

## Tipos de empresa y licencias

El arranque del backend ejecuta `EnsureNuevosVerticalesTipoEmpresaYLicencias`, que verifica los 20 tipos de empresa en `tipos_de_empresas`, crea/actualiza sus preconfiguraciones iniciales y asegura 4 licencias base por tipo:

- prueba 15 dias: 250 documentos;
- 30 dias: 1000 documentos;
- 30 dias: 2000 documentos;
- 30 dias: 4000 documentos.

Cada licencia incluye el modulo vertical correspondiente y los modulos base de ventas, inventario, compras, clientes/CRM, finanzas, bancos, tesoreria, cobranza, gestion documental, contratos, calidad, facturacion y seguridad. Los tickets de ayuda se gestionan aparte en el sistema propio centralizado.

## Evidencia visual

- `documentos/evidencias_qa/20_verticales_2026-05-10/verticales_nuevos_menu.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/administrar_empresa_soluciones_20_verticales_directos.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/administrar_empresa_soluciones_20_verticales_mobile.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/administrar_empresa_soluciones_20_botones_integrados.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_sin_cuadros_superiores.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/agencia_viajes_modulo.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/verticales_catalogo_centralizado.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/submenu_agencia_viajes_centralizado.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/licencias_catalogo_verticales.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/index_verticales_dinamicos.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/verticales_filtrado_permisos.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/permisos_rol_fila_verticales.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/configuracion_permisos_fila_verticales.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/verticales_menu_filtrado_mock.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_contexto_seccion.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_contexto_registros.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_contexto_mobile.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_ruta_trabajo.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_ruta_trabajo_mobile.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_configuracion_operativa.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_configuracion_mobile.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_diagnostico_configuracion.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_diagnostico_mobile.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_diagnostico_backend.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/modulo_vertical_secciones_backend_final.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/verticales_catalogo_backend_endpoint.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/super_licencias_catalogo_backend_verticales.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/index_verticales_catalogo_publico_backend.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/index_verticales_catalogo_publico_backend_final.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/index_verticales_merge_backend_custom_cards.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/descripcion_vertical_agencia_viajes_backend.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/elegir_licencia_vertical_agencia_viajes.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/pagar_licencia_vertical_agencia_viajes.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/seleccionar_empresa_tarjeta_vertical_agencia.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/seleccionar_empresa_preview_tipo_vertical.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/super_tipos_empresas_verticales.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/super_preconfiguracion_verticales.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/ayuda_verticales_2026.png`
- `documentos/evidencias_qa/20_verticales_2026-05-10/ayuda_verticales_2026_mobile.png`

## Verificacion tecnica

- `go test ./...` en `backend`.
- `node --check web/js/administrar_empresa.js`
- `node --check web/js/modulo_colombia_admin.js`
- `node --check web/js/nuevos_verticales_catalogo.js`
- Validacion de scripts inline en `verticales_nuevos_menu.html`, `modulo_menu.html`, `super/licencias.html` e `index.html`.
- Validacion cruzada de catalogos backend/frontend: 20 modulos exactos y sin faltantes.
- Validacion estatica de no duplicacion: `index.html` y `super/licencias.html` no conservan tarjetas/checks estaticos de `agencia_viajes` a `capacitacion_empresarial`.
- Validacion backend de no duplicacion: las pruebas comprueban que cada tipo/licencia use el titulo derivado de su plantilla operativa y que el numero de plantillas coincida con el catalogo de tipos.
- Validacion de permisos: los 20 botones directos del menu empresarial usan las mismas claves `link<Vertical>` del catalogo y se filtran por modulo/rol/licencia.
- Validacion visual controlada de permisos: capturas con contexto mock local para comprobar que `linkNuevosVerticales` aparece como regla agrupada y que el lanzador solo muestra los verticales permitidos.
- Validacion visual del flujo por seccion: capturas desktop y celular del modulo `agencia_viajes`, confirmando banda fija de seccion activa y salto directo a `Registros`.
- Validacion visual de ruta de trabajo: capturas desktop y celular con pasos numerados de `agencia_viajes`, resaltando `Reservas y vouchers`.
- Validacion visual de configuracion: capturas desktop y celular con `mcConfig`, plantilla activa, chips de estados/categorias/acciones y metadata de ejemplo.
- Validacion visual de diagnostico: capturas desktop y celular con checklist de preparacion y exportacion `Checklist CSV`.
- Validacion backend de diagnostico: pruebas unitarias para estado listo, fallos de configuracion, checks informativos y recomendaciones.
- Validacion backend de secciones: cada vertical debe exponer al menos 4 secciones de flujo desde su plantilla.
- Validacion backend de catalogo: el endpoint de catalogo debe exponer exactamente 20 verticales con page key, modulo, titulo, secciones y plantilla.
- Validacion publica de catalogo: la portada puede consumir `/api/public/verticales_nuevos/catalogo` y renderizar las tarjetas desde contrato backend con fallback local.
- Validacion publica con tarjetas personalizadas: `index.html` conserva las tarjetas configuradas en `/api/public/pagina_principal` y agrega los 20 verticales; `descripcion_de_los_sistemas.ht#vertical-agencia_viajes` abre una ficha generada desde backend.
- Validacion comercial de licencias: `elegir_licencia.html` muestra planes verticales con icono/tono del catalogo y `pagar_licencia.html` muestra vertical y cupo de documentos desde el contrato publico de checkout.
- Validacion de selector de empresas: tarjetas y formulario reconocen los nuevos tipos verticales desde `web/js/nuevos_verticales_catalogo.js`, incluyendo vista previa de secciones antes de crear empresa.
- Validacion de super administrador: tipos de empresa y preconfiguraciones identifican visualmente los verticales, muestran conteos y exponen secciones principales del flujo.
- Validacion de ayuda e IA: prueba automatica del contexto canonico de IA y captura visual de la seccion `20 verticales nuevos` en el centro de ayuda.
- Validacion de simplificacion UI: `Administrar empresa > Soluciones por negocio` muestra los 20 tipos como botones directos, sin encabezado separado; la pagina operativa del modulo ya no renderiza el bloque superior de seccion activa, accesos rapidos ni ruta numerada.
