# Estructura del Base de Datos

Version: 2026-05-15.1.0
Ultima actualizacion: 2026-05-15

Este documento consolida la estructura relacional activa del proyecto.
Nota de gobernanza documental:
- `documentos/estructura_bd.md` es la fuente canonica del esquema fisico.
- `documentos/diagramas/diagrama_entidad_relacion.md` y `documentos/diagramas/diagrama_entidad_relacion.svg` resumen visualmente el DER vigente, pero no sustituyen el detalle completo de este documento.
Bases operativas PostgreSQL en VPS:
- `pcs_empresas`
- `pcs_superadministrador`
Regla de motor:
- PostgreSQL es el unico motor de base de datos permitido para runtime, pruebas operativas, scripts y documentacion vigente.
- Las consultas de introspeccion deben usar catalogos PostgreSQL como `pg_indexes` e `information_schema`.
- No hay artefactos locales de base de datos versionados como fuente operativa vigente.
Todas las tablas operativas usan como base los campos estandar:
- id (clave primaria)
- fecha_creacion
- fecha_actualizacion
- usuario_creador TEXT
- estado TEXT DEFAULT 'activo'
- observaciones TEXT

Actualizacion 2026-05-18 (configuracion operativa en PostgreSQL)
- No se agregan tablas ni columnas fisicas.
- Los guardados de `empresa_configuracion_operativa`, `empresa_configuracion_operativa_roles`, `empresa_configuracion_operativa_politicas` y `empresa_configuracion_operativa_historial` retornan el `id` con `RETURNING id`.
- Se elimina la dependencia runtime de `LastInsertId()` para snapshots de historial y rollback operativo.

Actualizacion 2026-05-17 (configuracion super por paginas)
- No se agregan tablas ni columnas fisicas.
- Las paginas `web/super/configuracion/*.html` solo separan visualmente las secciones de configuracion super.
- Los formularios siguen guardando mediante las claves y endpoints existentes de configuracion global.

Actualizacion 2026-05-17 (abonos de carrito por estacion)
- Nueva tabla `carrito_compra_abonos` para registrar abonos activos por `empresa_id` y `carrito_id`.
- Campos principales: `monto`, `metodo_pago`, `referencia_pago`, `cierre_caja_id`, `caja_codigo`, `caja_turno`, `caja_sucursal_id`, `fecha_abono`, `usuario_creador`, `estado`, `observaciones`.
- Indices: `ix_carrito_abonos_empresa_carrito` para consulta por cuenta/estacion y `ix_carrito_abonos_empresa_caja` para trazabilidad de caja.
- `pagar_estacion` descuenta la suma activa de abonos del saldo a cobrar; los abonos no modifican `devolucion_total`.

Actualizacion 2026-05-17 (venta directa con vista de carrito de estaciones)
- No se agregan tablas ni columnas fisicas.
- La venta directa sigue usando `carritos_compras` con codigo canonico `VENTA-DIRECTA-{empresa_id}-0` y referencia `CAJA_DIRECTA`.
- El cambio solo alinea la vista frontend con el mismo panel operativo que usan las estaciones.

Actualizacion 2026-05-17 (mantenimiento super principal)
- No se agregan tablas ni columnas fisicas.
- La accion `limpiar_viejos` elimina entradas del JSON `mantenimiento_programado.avisos_json` cuando estan desactivadas o vencidas.
- La limpieza resincroniza las claves legacy del aviso visible sin cambiar `mantenimiento_activo`.

Actualizacion 2026-05-17 (avisos de mantenimiento programados)
- No se agregan tablas ni columnas fisicas.
- El super administrador guarda la lista de avisos en configuracion global con la clave `mantenimiento_programado.avisos_json`.
- La administracion visual de la lista vive en `web/super/mantenimiento_sistema.html`; no modifica el esquema fisico.
- Las claves legacy `mantenimiento_programado.aviso_activo`, `fecha`, `hora_inicio`, `hora_fin`, `zona_horaria` y `mensaje_publico` se sincronizan desde el primer aviso activo para conservar el contrato de `/api/empresa/mantenimiento_programado`.
- El bloqueo real sigue separado en `mantenimiento_activo`.

Actualizacion 2026-05-17 (estaciones: primer clic solo activa)
- No se agregan tablas ni columnas fisicas.
- Se reutiliza `empresa_estacion_prefs` con `estacion_id=0`, `clave='estaciones_config'`.
- El JSON `estaciones_config.station_card_ui` puede incluir `solo_activar_primer_clic`; cuando vale `true`, el primer clic sobre una estacion disponible solo activa su carrito base y el segundo clic abre el carrito.
- Compatibilidad: `abrir_carrito_al_activar=false` se conserva como alias historico de `solo_activar_primer_clic=true`.

Actualizacion 2026-05-17 (default carrito simplificado por tipo de empresa)
- No se agregan tablas ni columnas fisicas.
- El JSON `estaciones_config.carrito_ui_global` conserva la misma estructura, pero el default para empresas nuevas deja activas las tarjetas operativas basicas y apaga por defecto las tarjetas avanzadas de cobro/descuentos/propina/comision/lavador.
- En arranque, `ApplyDefaultCarritoUIToExistingEmpresaPrefs` normaliza las filas activas antiguas de `empresa_estacion_prefs` (`estacion_id=0`, `clave='estaciones_config'`) para aplicar el mismo preset y fuerza `usar_configuracion_global=true` por estacion. Las empresas activas sin esa preferencia reciben una configuracion base con `Estacion 1`.

Actualizacion 2026-05-15 (empresas compartidas con alcance)
- `pcs_superadministrador.admin_empresa_compartida` agrega `nivel_acceso` (`solo_ver`, `acceso_total`, `modulos`) y `modulos_permitidos` para guardar el alcance efectivo del acceso aceptado.
- `pcs_superadministrador.admin_empresa_compartida_invitaciones` agrega las mismas columnas para que el alcance elegido por el propietario viaje con la invitacion y se materialice al aceptarla.
- El backend asegura ambas columnas en arranque y mantiene el aislamiento por `empresa_id`; no se agregan motores ni dependencias.

Actualizacion 2026-05-15 (licencia de prueba unica por empresa)
- `pcs_superadministrador.licencias_activaciones_gratis` agrega `asesor_id` para trazabilidad comercial de pruebas/activaciones sin pago.
- Se crea `ux_licencias_gratis_empresa_unica` como indice unico parcial sobre `empresa_id` cuando `estado` esta activo, impidiendo que una misma empresa use mas de una prueba/gratis activa.
- Al asegurar esquema, duplicados activos historicos se marcan como `historico_duplicado` antes de crear el indice unico.
- La prueba automatica de 15 dias se registra con 250 documentos/ventas mensuales y conserva aislamiento por `empresa_id`.

Actualizacion 2026-05-12 (tickets de ayuda empresariales profesionalizados)
- `pcs_superadministrador.super_tickets_ayuda` agrega `contacto_telefono`, `contacto_preferido` y `contexto_json` para soporte profesional. El contexto se limita a claves tecnicas seguras de pantalla/modulo/navegador y no guarda cookies, localStorage, claves ni secretos.
- `pcs_superadministrador.super_ticket_ayuda_mensajes` conserva `interno` para notas privadas del super administrador; las consultas empresariales no devuelven mensajes internos.
- Aislamiento: todo detalle o comentario empresarial se valida por `empresa_id`; la vista global queda exclusiva del super administrador.

Actualizacion 2026-05-12 (tickets de ayuda SaaS)
- Se agregan tablas en `pcs_superadministrador` para la mesa central de soporte:
  - `super_tickets_ayuda`: ticket maestro con `codigo`, `empresa_id`, `empresa_nombre`, solicitante, origen, modulo/ruta, asunto, categoria, prioridad, estado, ultimo mensaje, asignacion y cierre.
  - `super_ticket_ayuda_mensajes`: mensajes del ticket con `ticket_id`, tipo de autor (`usuario`, `super`, `sistema`), autor, mensaje, bandera `interno` y trazabilidad de creacion.
- El endpoint empresarial valida alcance por `empresa_id`; el panel super consulta la bandeja central sin escribir en bases empresariales.
- No se agregan motores, dependencias ni almacenamiento de archivos adjuntos.

Actualizacion 2026-05-12 (retiro de Nextcloud y cuota DB)
- Se elimina el uso runtime de la tabla legacy `empresa_nextcloud_accounts`; el arranque ejecuta `DROP TABLE IF EXISTS empresa_nextcloud_accounts` para retirar credenciales antiguas de empresas.
- Se eliminan las claves super `nextcloud.enabled`, `nextcloud.base_url`, `nextcloud.admin_user` y `nextcloud.admin_secret`.
- No se crea una tabla nueva para cuotas. La configuracion `pcs_superadministrador.configuraciones.config_key='empresa.limitaciones.db.max_gb'` define el tamano maximo de base de datos asignado por empresa.
- Para conservar el valor ya configurado en instalaciones existentes, el backend lee el valor legacy `empresa.limitaciones.nextcloud.max_gb` cuando la nueva clave aun no existe y luego guarda en `empresa.limitaciones.db.max_gb`.

Actualizacion 2026-05-12 (registro operativo por invitacion)
- No se agregan tablas ni columnas fisicas.
- El flujo de primer ingreso reutiliza `pcs_empresas.users.email_confirm_token`, `email_confirm_expira` y `email_confirmado` como invitacion de un solo uso.
- Al completar el primer password, el backend confirma el correo, consume el token, limpia expiracion y mantiene el usuario aislado por `empresa_id`.
- La sesion se materializa en `pcs_superadministrador.administradores`, `sessions` y acceso compartido existente para que `/api/empresa/permisos_contexto` cargue rol efectivo y paginas permitidas.

Actualizacion 2026-05-12 (adaptacion del nucleo por plantilla)
- No se agregan tablas ni columnas fisicas.
- El JSON `tipo_empresa_preconfiguraciones.config_json` puede incluir `adaptacion_nucleo` con `fuente_unica`, `usuarios_desde_nucleo`, `productos_servicios_desde_nucleo`, `estaciones_como_recursos_configurados`, nombres de entidad de estacion, roles, productos/servicios guia, estaciones guia y reglas.
- Al aplicar una plantilla se guarda la preferencia `preconfiguracion_tipo_empresa_adaptacion_nucleo` en `empresa_estacion_prefs`.
- `estaciones_config` agrega metadata JSON de recurso (`tipo_recurso`, `tipo_recurso_plural`, `representa_recurso_negocio`) para que la misma tabla de estaciones represente estaciones, apartamentos, puestos, vehiculos, bahias, aulas, consultorios u otros recursos.

Actualizacion 2026-05-12 (matriz profesional de 30 verticales)
- No se agregan tablas ni columnas fisicas.
- La matriz `/api/*/verticales_integracion/catalogo` consume metadatos ya existentes de `tipo_empresa_preconfiguraciones.config_json`, `empresa_modulos_colombia_*` y catalogos de licencias/tipos para publicar preparacion profesional de 30 verticales canonicos exactos.
- `professional_ready`, `readiness_score`, `readiness_checks`, `configuration_scope`, `fused_modules`, `support_modules`, `similar_templates`, `financial_core_modules`, `income_flow`, `expense_flow`, `financial_tables` y `financial_reports` son campos de contrato API calculados; no se almacenan como columnas.
- El amarre de ingresos/egresos usa las tablas existentes `carritos_compras`, `carrito_compra_items`, `empresa_finanzas_movimientos`, `empresa_finanzas_configuracion` y `empresa_finanzas_periodos`; no se crea una tabla financiera por vertical.
- Los alias `consultorio_odontologico` y `taxi`, y los soportes `turnos_atencion`/`turnos`, no crean tablas nuevas ni filas de producto vertical; se resuelven por configuracion, permisos y plantilla canonica.
- Los 20 verticales nuevos conservan sus datos operativos transversales en `empresa_modulos_colombia_*` por `empresa_id` y `modulo`; ventas, pagos, productos, clientes, facturacion y reportes siguen en las tablas centrales.

Actualizacion 2026-05-12 (visibilidad comercial de licencias)
- No se agregan tablas ni columnas fisicas.
- Se reutiliza `pcs_superadministrador.licencias.activo`: `1` significa visible para clientes y habilitada para checkout; `0` significa oculta para clientes y bloqueada en compras/activaciones publicas nuevas.
- Los filtros publicos de licencias usan `COALESCE(activo, 1) = 1` para compatibilidad con registros antiguos.

Actualizacion 2026-05-11 (2FA login global)
- No se agregan tablas ni columnas fisicas.
- Se reutiliza `pcs_superadministrador.configuraciones` con la clave `security.admin_2fa.enabled` para activar/desactivar globalmente la exigencia de OTP en login de administradores.
- Las credenciales TOTP por cuenta permanecen en `administradores.totp_enabled`, `administradores.totp_secret` y `administradores.totp_confirmado_en`.

Actualizacion 2026-05-11 (integracion profesional de verticales)
- La primera tanda de matriz/visibilidad no agrego tablas ni columnas.
- Los 20 verticales nuevos siguen usando las tablas compartidas `empresa_modulos_colombia_*` por `empresa_id` y `modulo`.
- Los verticales clasicos con tablas propias solo pueden quedar visibles cuando sus datos cobrables migran o referencian el nucleo de clientes, productos/servicios, ventas y pagos.

Actualizacion 2026-05-11 (preconfiguraciones verticales produccion masiva)
- No se agregan tablas ni columnas fisicas.
- El JSON de `tipo_empresa_preconfiguraciones.config_json` puede incluir `integracion_vertical` para conectar la plantilla con la matriz extendida.
- `integracion_vertical` registra `modulo`, `estado_integracion`, `decision`, `produccion_masiva`, `prioridad_produccion`, `motivo_decision`, `template_activates`, `tables_touched`, `required_permissions`, `sale_flow` y `reports_produced`.
- Los 20 verticales nuevos quedan marcados como produccion masiva en el JSON, con prioridad 1-20.
- La accion super `asegurar_20_licencias` reutiliza las tablas existentes `tipos_empresas`, `tipo_empresa_preconfiguraciones` y `licencias`; no introduce esquema nuevo. Se conserva `asegurar_v1_licencias` como alias compatible.

Actualizacion 2026-05-11 (gimnasio integrado al nucleo)
- `empresa_gimnasio_socios.cliente_id`: referencia al cliente central creado o reutilizado para el socio.
- `empresa_gimnasio_planes.servicio_id`: referencia al servicio vendible central creado desde el plan.
- `empresa_gimnasio_pagos.cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id`: referencias a cliente, servicio e item/venta central generados por el recaudo de gimnasio.
- Las tablas de gimnasio conservan solo la especialidad operativa: acceso, clases, asistencia, credenciales y lectura fitness; el cobro queda reconciliable con `carritos_compras` y `carrito_compra_items`.
- Compatibilidad PostgreSQL: los indices de integracion (`servicio_id`, `cliente_id`, `carrito_id`) se crean despues de asegurar columnas para que bases existentes no fallen durante la carga del modulo.

Actualizacion 2026-05-11 (odontologia integrada al nucleo)
- `empresa_odontologia_pacientes.cliente_id`: referencia al cliente central creado o reutilizado para el paciente facturable.
- `empresa_odontologia_tratamientos.servicio_id`: referencia al servicio vendible central creado desde el tratamiento.
- `empresa_odontologia_pagos.cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id`: referencias a cliente, servicio e item/venta central generados por el recaudo odontologico.
- Las tablas clinicas de odontologia conservan la especialidad: historia clinica, odontograma, profesionales, consultorios, citas, presupuestos y trazabilidad clinica; el circuito comercial queda reconciliable con `clientes`, `servicios`, `carritos_compras` y `carrito_compra_items`.
- Compatibilidad PostgreSQL: los indices de integracion (`cliente_id`, `servicio_id`, `carrito_id`) se crean despues de asegurar columnas para evitar cargas parciales en instalaciones actualizadas por fases.

Actualizacion 2026-05-11 (parqueadero integrado al nucleo)
- `empresa_parqueadero_tickets.cliente_id`: referencia opcional al cliente central creado o reutilizado cuando el ticket trae cliente/documento.
- `empresa_parqueadero_tickets.servicio_id`: referencia al servicio vendible central por tipo de vehiculo.
- `empresa_parqueadero_tickets.carrito_id` y `carrito_item_id`: referencias al carrito e item central generados al cobrar salida.
- La tabla de parqueadero conserva la especialidad operativa: placa, QR, entrada, salida, minutos, tarifas y anulaciones; el cobro queda reconciliable con `servicios`, `carritos_compras` y `carrito_compra_items`.

Actualizacion 2026-05-11 (taxi system integrado al nucleo)
- `empresa_taxi_customers.cliente_id`: referencia al cliente central creado o reutilizado para el cliente registrado del portal.
- `empresa_taxi_requests.cliente_id`: referencia al cliente central usado por el viaje, incluso si la solicitud fue como invitado.
- `empresa_taxi_requests.servicio_id`: referencia al servicio vendible central para viajes de taxi.
- `empresa_taxi_requests.carrito_id` y `carrito_item_id`: referencias al carrito e item central generados al completar el viaje.
- `empresa_taxi_requests.metodo_pago`: metodo normalizado contra el flujo de carrito; por defecto `efectivo` mientras no exista pasarela propia del viaje.
- Las tablas de taxi conservan la especialidad operativa: conductores, ofertas, GPS, rutas, estados de viaje y trazabilidad de despacho; el cobro queda reconciliable con `clientes`, `servicios`, `carritos_compras` y `carrito_compra_items`.

Actualizacion 2026-05-11 (domicilios integrado al nucleo)
- `empresa_domicilios_menu_items.servicio_id`: referencia al servicio vendible central creado o reutilizado para cada producto de menu.
- `empresa_domicilios_orders.cliente_id`: referencia al cliente central creado o reutilizado desde nombre, telefono y direccion del pedido.
- `empresa_domicilios_orders.carrito_id`: referencia al carrito central generado cuando el pedido pasa a `entregado`.
- `empresa_domicilios_order_items.servicio_id` y `carrito_item_id`: referencias al servicio vendido y al item central de carrito por linea del pedido.
- La tabla de domicilios conserva la especialidad operativa: restaurantes aliados, domiciliarios, ofertas, tracking GPS, estados logisticos, codigo de entrega y calculo de tarifa; el cobro queda reconciliable con `clientes`, `servicios`, `carritos_compras` y `carrito_compra_items`.

Actualizacion 2026-05-11 (apartamentos turisticos integrado al nucleo)
- `empresa_apartamentos_turisticos_unidades.servicio_id`: referencia al servicio vendible central creado o reutilizado para cada apartamento/unidad.
- `empresa_apartamentos_turisticos_reservas.cliente_id`: referencia al cliente central creado o reutilizado desde huesped, documento, telefono o email.
- `empresa_apartamentos_turisticos_reservas.servicio_id`: referencia al servicio central de alojamiento usado por la reserva.
- `empresa_apartamentos_turisticos_reservas.metodo_pago`: metodo normalizado contra el flujo de carrito; por defecto `efectivo` cuando no existe pasarela/canal externo conciliado.
- Las tablas de apartamentos conservan la especialidad operativa: unidades, disponibilidad, tarifas, codigos de acceso, check-in/check-out, limpieza y mantenimiento; el cobro queda reconciliable con `clientes`, `servicios`, `carritos_compras` y `carrito_compra_items`.

Actualizacion 2026-05-11 (propiedad horizontal integrada al nucleo)
- `empresa_propiedad_horizontal_unidades.servicio_id`: referencia al servicio vendible central creado o reutilizado para la cuota base de la unidad.
- `empresa_propiedad_horizontal_personas.cliente_id`: referencia al cliente central creado o reutilizado para propietarios, residentes, arrendatarios y apoderados.
- `empresa_propiedad_horizontal_cargos.servicio_id`: referencia al servicio vendible central creado o reutilizado para cada concepto cobrable.
- `empresa_propiedad_horizontal_recaudos.cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id`: referencias a cliente, servicio e item/venta central generados por el recaudo de copropiedad.
- Las tablas de propiedad horizontal conservan la especialidad operativa: unidades, coeficientes, cartera, PQR, asambleas y trazabilidad de copropiedad; el cobro queda reconciliable con `clientes`, `servicios`, `carritos_compras` y `carrito_compra_items`.

Actualizacion 2026-05-11 (alquileres integrado al nucleo)
- `empresa_alquileres_activos.servicio_id`: referencia al servicio vendible central creado o reutilizado para cada activo alquilable.
- `empresa_alquileres_tarifas.servicio_id`: referencia al servicio vendible central creado o reutilizado para cada tarifa/modalidad cobrable.
- `empresa_alquileres_contratos.cliente_id`, `servicio_id`, `carrito_id` y `carrito_item_id`: referencias a cliente, servicio e item/venta central generados por el contrato de alquiler.
- Las tablas de alquileres conservan la especialidad operativa: contratos, garantias, kilometraje, GPS, mantenimiento, entrega y devolucion; el flujo cobrable queda reconciliable con `clientes`, `servicios`, `carritos_compras` y `carrito_compra_items`.

Actualizacion 2026-05-11 (drogueria/farmacia validada al nucleo)
- `drogueria_farmacia` no agrega tablas fisicas propias de productos, inventario, ventas ni pagos.
- Usa `empresa_modulos_colombia_registros`, `empresa_modulos_colombia_eventos`, `empresa_modulos_colombia_evidencias`, `empresa_modulos_colombia_aprobaciones` y `empresa_modulos_colombia_tareas` con `modulo='drogueria_farmacia'` para expediente sanitario.
- Productos, lotes operativos de inventario, compras, clientes, ventas, pagos y facturacion se mantienen en los modulos centrales existentes.

Actualizacion 2026-05-11 (AIU construccion integrado al nucleo)
- `empresa_aiu_contratos.cliente_id`: referencia al cliente central creado o reutilizado para el cliente de obra.
- `empresa_aiu_contratos.servicio_id`: referencia al servicio vendible central creado o reutilizado para el contrato AIU.
- `empresa_aiu_items.servicio_id`: referencia al servicio vendible central creado o reutilizado para cada concepto/capitulo cobrable.
- `empresa_aiu_facturas.carrito_id` y `carrito_item_id`: referencias al carrito e item central generados por la factura AIU.
- Las tablas AIU conservan la especialidad de construccion: capitulos, porcentajes AIU, base IVA, retenciones, anticipo, garantia, avance, riesgo y auditoria tecnica; el circuito cobrable queda reconciliable con `clientes`, `servicios`, `carritos_compras`, `carrito_compra_items` y `empresa_facturacion_documentos`.

Actualizacion 2026-05-11 (catalogo API de integracion vertical)
- Se agregan endpoints de solo lectura para exponer la matriz operativa de verticales clasicos: `/api/public/verticales_integracion/catalogo`, `/api/empresa/verticales_integracion/catalogo` y `/super/api/verticales_integracion/catalogo`.
- Las tablas listadas en `tables_touched` son declarativas para auditoria de integracion: documentan que tablas del vertical y del nucleo participan en el flujo, sin crear relaciones nuevas en esta fase.

Actualizacion 2026-05-10 (alertas automaticas super)
- `metrics`: se amplia con `disk_total`, `disk_used` y `disk_percent` para que el panel super y las alertas midan la capacidad real del filesystem del VPS/contenedor.
- `super_alertas_config`: configuracion global del modulo de alertas super. Incluye activacion, correo destino, umbrales de disco, trafico, sesiones, conexiones PostgreSQL y ventana de enfriamiento entre correos repetidos.
- `super_alertas_eventos`: historial de alertas evaluadas y enviadas, con tipo, severidad, valor, umbral, destinatario, asunto, cuerpo, resultado SMTP, metadata JSON, fecha y auditoria.
- El trafico por porcentaje usa `hostinger.bandwidth.limit_gb` de `configuraciones` cuando existe; si no, el modulo permite umbral absoluto en GB.

Actualizacion 2026-05-06 (modulos empresariales Colombia)
- `empresa_modulos_colombia_registros`: registro transversal por `empresa_id` y `modulo` para bancos/pagos, gestion documental, KYC/KYB, contratos y calidad. Incluye tipo, codigo, nombre, tercero, responsable, categoria, referencia, prioridad, estado, fechas, valor, metadata JSON y auditoria. Los tickets de ayuda ya no usan esta tabla; se almacenan en `super_tickets_ayuda` y `super_ticket_ayuda_mensajes`.
- `empresa_modulos_colombia_eventos`: bitacora por `empresa_id`, modulo y registro, con evento, cambio de estado, detalle, usuario y fecha.
- `empresa_modulos_colombia_evidencias`: soportes por `empresa_id`, modulo y registro. Incluye tipo, nombre, URL/ruta, descripcion, usuario y fecha.
- `empresa_modulos_colombia_aprobaciones`: flujo de aprobaciones por `empresa_id`, modulo y registro. Incluye nivel, destinatario, solicitante, estado, comentario, decisor, vencimiento y fechas.
- `empresa_modulos_colombia_tareas`: compromisos por `empresa_id`, modulo y registro. Incluye titulo, responsable, prioridad, estado, vencimiento, comentario, creador y fechas.
- La restriccion unica `empresa_id + modulo + codigo` evita duplicados dentro de cada empresa sin mezclar modulos ni companias.

Actualizacion 2026-05-06 (portal de terceros y certificados tributarios)
- `empresa_portal_terceros`: maestro de terceros por `empresa_id`, tipo, documento, DV, razon social, contacto, regimen, estado, token y auditoria.
- `empresa_certificados_tributarios`: certificados por tercero y empresa, tipo de certificado, periodo, base, retenciones, total, estado, firma, token publico y fechas de emision/envio/anulacion.
- `empresa_certificados_tributarios_descargas`: bitacora de descargas publicas con certificado, tercero, canal, IP, navegador y fecha.

Actualizacion 2026-05-06 (activos fijos NIIF/Fiscal)
- `empresa_contabilidad_activos_fijos`: se amplia como libro maestro de PPE e intangibles por `empresa_id`, con campos NIIF (`vida_util_meses`, `metodo_depreciacion`, `depreciacion_acumulada`, `deterioro_acumulado`, `valor_razonable`, `valor_libros`) y fiscales (`base_fiscal`, `vida_util_fiscal_meses`, `metodo_depreciacion_fiscal`, `depreciacion_acumulada_fiscal`, `valor_fiscal`, `diferencia_niif_fiscal`).
- `empresa_contabilidad_activos_depreciacion`: conserva la depreciacion mensual por activo y periodo, usada por la nueva API `/api/empresa/activos_fijos_niif_fiscal`.
- `empresa_contabilidad_activos_eventos`: registra traslados, mantenimientos, ajustes, bajas, ventas y retiros con ubicacion, responsable, valor, detalle y auditoria.

Actualizacion 2026-05-06 (centros de costo y rentabilidad)
- `empresa_centros_costo`: maestro de centros por `empresa_id`, codigo, nombre, tipo, nivel, padre, responsable, sucursal, area, unidad de negocio, meta de margen, estado y auditoria.
- `empresa_centros_costo_reglas`: reglas de imputacion por `empresa_id`, centro, origen de modulo, categoria, tercero/cuenta patron, porcentaje, prioridad, activa, estado y auditoria.
- `empresa_centros_costo_presupuestos`: presupuesto por `empresa_id`, centro, periodo, escenario, ingresos, egresos, meta de margen, responsable, estado y auditoria.
- `empresa_declaraciones_tributarias`: declaraciones por `empresa_id`, tipo, periodo, anio, periodicidad, rango, vencimiento, NIT/municipio, formulario, bases, IVA, retenciones, consumo, anticipos, sanciones, intereses, saldos, estado, soportes y auditoria.
- `empresa_declaraciones_tributarias_movimientos`: detalle de conciliacion por `empresa_id`, declaracion, tipo, periodo, modulo origen, referencia, tercero, concepto, base, impuesto, retencion, naturaleza y estado.
- `empresa_calendario_tributario`: calendario editable por `empresa_id`, tipo de declaracion, anio, periodo, periodicidad, rango, vencimiento, digitos de NIT, estado y observaciones.
- `empresa_wms_ubicaciones`: ubicaciones internas por `empresa_id`, codigo, bodega, zona, pasillo, rack, nivel, posicion, tipo, capacidad, ocupacion y estado.
- `empresa_wms_ordenes`: ordenes WMS por `empresa_id`, codigo, tipo, documento origen, tercero/cliente, fecha compromiso, prioridad, responsable, estado y auditoria.
- `empresa_wms_items`: items de picking/packing por `empresa_id`, orden, producto/SKU, ubicacion origen/destino, lote, serial, cantidades solicitadas/pickeadas/empacadas y estado.
- `empresa_wms_despachos`: despachos por `empresa_id`, orden, codigo, transportadora, guia, conductor, vehiculo, ruta, fechas, flete y estado.
- `empresa_wms_eventos`: bitacora de operaciones WMS por `empresa_id`, referencia, evento, cambio de estado, detalle y usuario.
- El dashboard del modulo no crea movimientos duplicados: consolida datos existentes con `centro_costo` desde contabilidad Colombia, tesoreria, compras avanzadas, captura OCR/IA y AIU construccion.

Actualizacion 2026-05-06 (propiedad horizontal)
- `empresa_propiedad_horizontal_config`: configuracion de copropiedad por `empresa_id`, con datos fiscales, contacto, mora, facturacion electronica y portal de residentes.
- `empresa_propiedad_horizontal_unidades`: unidades privadas/comunes, torres, coeficiente, cuota base, parqueadero, deposito y estado.
- `empresa_propiedad_horizontal_personas`: propietarios, residentes, arrendatarios y apoderados vinculados a unidades.
- `empresa_propiedad_horizontal_cargos`: cuotas ordinarias/extraordinarias, multas, mora, descuentos, total, saldo y vencimiento.
- `empresa_propiedad_horizontal_recaudos`: pagos aplicados a unidad/cargo con metodo, referencia y valor.
- `empresa_propiedad_horizontal_pqrs`: peticiones, quejas, reclamos, mantenimiento y seguridad con prioridad/responsable.
- `empresa_propiedad_horizontal_asambleas`: convocatorias, quorum, acta y estado.
- `configuraciones`: claves `licencias.asesor_promo.enabled`, `licencias.asesor_promo.percent` y `licencias.asesor_promo.updated_by` para promocion de licencias por asesor.

Actualizacion 2026-05-06 (cierre y bloqueo fiscal)
- `empresa_cierre_fiscal_politicas`: politica por `empresa_id` y modulo, con bloqueo automatico, dias de edicion retroactiva, reapertura aprobada, excepciones, notificacion post-cierre, estado y auditoria.
- `empresa_cierre_fiscal_periodos`: periodos por empresa con rango de fechas, estado (`abierto`, `en_revision`, `cerrado`, `bloqueado`) y banderas para bloquear ventas, compras, caja, inventario, contabilidad y facturacion.
- `empresa_cierre_fiscal_excepciones`: autorizaciones temporales por periodo, modulo, accion y documento para operar sobre periodos cerrados.
- `empresa_cierre_fiscal_eventos`: bitacora de validaciones, bloqueos, cierres, reaperturas y acciones post-cierre.
- `empresa_contabilidad_colombia_periodos` queda sincronizado con `empresa_cierre_fiscal_periodos` al cerrar o reabrir desde Contabilidad Colombia.

Actualizacion 2026-05-05 (carnets empresariales por empresa)
- `empresa_carnets_plantillas`: plantillas visuales por `empresa_id` para carnets modernos. Incluye tipo, orientacion, ancho/alto, colores, visibilidad de logo/foto/QR/codigo de barras, campos visibles, diseno JSON, plantilla predeterminada, estado y auditoria basica.
- `empresa_carnets`: carnets emitidos por `empresa_id`, vinculables a `users.id`. Incluye codigo unico por empresa, tipo de persona, datos de identidad, cargo, area, foto, nivel de acceso, grupo sanguineo, contacto de emergencia, emision, vencimiento, payload QR, estado del carnet y ultima impresion.
- `empresa_carnets_eventos`: bitacora por `empresa_id` y `carnet_id` para emision, actualizacion, suspension, revocacion, vencimiento e impresion/exportacion.
- El modulo se expone por `/api/empresa/carnets` con wrapper `WithEmpresaCarnetsPermissions`; no comparte carnets entre empresas.

Actualizacion 2026-05-05 (suite contable Colombia avanzada)
- `empresa_contabilidad_exogena_formatos`: formatos DIAN/medios magneticos configurables por `empresa_id`, formato, version, año gravable, concepto, periodicidad, estado y ultima generacion.
- `empresa_contabilidad_exogena_registros`: registros por tercero/formato para exogena, con documento, razon social, cuenta, base, IVA, retencion, total, validaciones y estado.
- `empresa_contabilidad_nomina_electronica`: documentos de nomina electronica por empleado/documento y periodo, con salario, devengados, deducciones, total, CUNE, estado DIAN, respuesta y payload.
- `empresa_contabilidad_documentos_soporte`: documentos soporte electronicos para compras a no obligados a facturar, con proveedor, periodo, subtotal, IVA, retenciones, total, CUDS, estado DIAN y payload.
- `empresa_contabilidad_activos_fijos`: activos fijos por empresa con costo, valor residual, vida util, depreciacion mensual/acumulada, valor en libros y cuentas contables.
- `empresa_contabilidad_cartera_cxp`: cuentas por cobrar y por pagar con tercero, documento, vencimiento, valor original, valor pagado, saldo, estado y referencia externa.
- El modulo reutiliza `empresa_contabilidad_colombia_comprobantes` y `empresa_contabilidad_colombia_lineas` para generar libros oficiales y registros de exogena sin duplicar el PUC ni los asientos.

Actualizacion 2026-05-05 (venta publica, carta QR y red social Motel Calipso)
- No introduce tablas nuevas; consolida el uso productivo de las tablas existentes de publicacion externa.
- `empresa_venta_publica_configuracion` mantiene la configuracion publica por empresa, incluyendo slug, estado visible y datos comerciales.
- `empresa_venta_publica_paginas` registra paginas publicas por empresa. Para `empresa_id=7` (`Motel Calipso`) quedan activas `experiencias-calipso`, `carta-productos-precios` y `pos-motel-calipso`.
- `empresa_venta_publica_items` almacena productos, servicios, experiencias y paquetes visibles en la pagina publica. Para Motel Calipso quedan ejemplos activos de decoracion romantica, noche romantica, bebidas/snacks, kit de aseo, desayuno y acceso POS.
- `empresa_publicaciones_red_social` registra publicaciones comerciales visibles en la red social del sistema; Motel Calipso queda con publicaciones activas para carta QR y experiencias.
- La exposicion publica de `visualizar_productos_y_precios_publico.html` se resolvio en middleware/rutas; no requiere migracion ni cambio fisico de esquema.

Actualizacion 2026-04-30 (pagos, empresas compartidas y documentos IA)
- admin_empresa_compartida:
  - tabla super para registrar administradores que tienen acceso compartido a una empresa.
  - campos clave: `empresa_id`, `admin_email`, `compartido_por_email`, `invitacion_id`, `fecha_aceptada`, `fecha_revocada`, `revocado_por_email`, `estado`.
  - permite que el administrador que compartio y el administrador receptor vean el acceso vigente y puedan retirarlo, dejando trazabilidad del actor que revoca.
- admin_empresa_compartida_invitaciones:
  - tabla super para invitaciones pendientes o historicas de acceso compartido.
  - campos clave: `empresa_id`, `admin_email`, `invitado_por_email`, `token_hash`, `estado`, `expira_en`, `fecha_aceptada`, `fecha_revocada`, `revocada_por_email`.
  - los tokens se guardan como hash; no se documentan ni almacenan tokens planos.
- pagos_epayco:
  - `raw_payload` puede incluir el diagnostico del checkout, `checkout_type=classic_form` y una copia saneada del formulario clasico cuando Smart Checkout v2 no entrega token.
  - el formulario clasico firmado se envia al navegador solo como respuesta transitoria para hacer POST a `https://secure.payco.co/checkout.php`; en auditoria/persistencia se enmascara `p_key`.
  - si falta `epayco.customer_id` o `epayco.checkout_key`/`epayco.p_key`, el backend devuelve error controlado `409` y no intenta redireccionar a URLs legacy que producen XML `AccessDenied`.
  - el diagnostico del fallback puede incluir `mode`, `mode_source`, `smart_mode` y `smart_mode_source`; para credenciales reales el formulario clasico debe persistir evidencia saneada de `p_test_request=false`.
- Documentos dinamicos IA:
  - el flujo `/generate` + `/download` genera registros temporales de documento y archivos exportables en almacenamiento temporal.
  - no introduce una tabla permanente nueva en esta version; si se requiere historial documental duradero, debe crearse una tabla dedicada con `empresa_id`, usuario, formato, estado y ruta segura.

Actualizacion 2026-04-21 (compras y finanzas: comprobantes adjuntos por empresa)
- empresa_compras_documentos:
  - agrega `comprobante_url` y `comprobante_nombre_archivo` para enlazar la evidencia física cargada por la empresa en cada documento de compra.
- empresa_finanzas_movimientos:
  - mantiene `comprobante_url` como referencia del soporte físico asociado a ingresos y egresos; desde esta fecha el valor puede provenir de subida local al repositorio web y no solo de URL externa manual.
- Regla operativa de almacenamiento:
  - los comprobantes físicos de compras y finanzas se guardan en el filesystem bajo `web/uploads/comprobantes/empresa_<empresa_id>/compras/` y `web/uploads/comprobantes/empresa_<empresa_id>/finanzas/`, manteniendo aislamiento por `empresa_id`.

Actualizacion 2026-04-26 (finanzas/inventario/asistencia: integracion operativa)
- empresa_finanzas_configuracion:
  - `integracion_contable_destino` acepta valores operativos `generico`, `siigo`, `world_office`, `alegra`, `helisa`, `loggro`, `contapyme`.
  - Las categorias/cuentas por defecto se amplian para operacion Colombia (ventas, estaciones, restaurante, bar, lavanderia, propinas, compras, nomina, servicios publicos, arriendo, mantenimiento, impuestos, bancos).
- productos:
  - `codigo_barras` se usa como destino persistente del generador de etiquetas Code 128 por empresa y como fuente prioritaria del lector en carritos.
- empresa_asistencia_empleados:
  - `empleado_id` puede vincular el registro de asistencia al usuario interno de la empresa (`users.id`) para mantener trazabilidad operativa de llegada/salida por cuenta creada.

Actualizacion 2026-04-29 (auditoria como fuente de contexto IA)
- empresa_auditoria_eventos:
  - la tabla existente pasa a ser la fuente central de contexto operativo reciente para IA empresarial y global.
  - los wrappers protegidos registran tambien acciones de lectura (`R`), ademas de crear/actualizar/eliminar/aprobar (`C/U/D/A`), para que la IA pueda ver actividad real de usuarios sin incrustarse en cada modulo.
  - el contexto IA usa consultas agregadas y eventos recientes acotados por ventana temporal; no inyecta `metadata_json` completo ni secretos.
  - si la tabla no existe, falla la consulta o la IA esta desactivada, el servidor mantiene operacion normal y la IA recibe una nota de contexto degradado cuando corresponda.
- empresa_auditoria_ia_consultas:
  - nueva bitacora de uso de auditoria por IA para consultas empresariales y globales.
  - campos clave: `empresa_id`, `alcance`, `modelo`, `usuario_consulta`, `pregunta_hash`, `pregunta_resumen`, `filtros_json`, `resultados_json`, `eventos_consultados`, `contexto_caracteres`, `resultado`, `fecha_consulta`.
  - registra que contexto de auditoria se preparo para GPT-5.4 mini o el modelo activo, sin guardar la pregunta completa cuando puede bastar con hash/resumen.
  - permite trazabilidad posterior de busquedas profundas y consultas DB seguras entregadas a la IA.
- Regla de consultas IA sobre base de datos:
  - la IA no ejecuta SQL libre ni recibe credenciales.
  - el backend deriva intencion desde la pregunta, aplica `empresa_id` cuando corresponde y solo ejecuta consultas parametrizadas/whitelist para usuarios activos, ventas, finanzas, inventario y clientes.
  - los resultados se devuelven como contexto compacto y se registran en `empresa_auditoria_ia_consultas`.
- Configuracion super de lectura DB para chat empresarial:
  - `ai.chat.empresa.db_query_enabled`: activa/desactiva el acceso de lectura total controlada de GPT-5.4 mini/modelo activo a tablas con `empresa_id`; default `true`.
  - `ai.chat.empresa.db_query_max_tables`: maximo de tablas incluidas por pregunta; default `25`, tope tecnico `100`.
  - `ai.chat.empresa.db_query_rows`: filas por tabla entregadas a la IA; default `8`, tope tecnico `30`.
  - `ai.chat.super.empresa_solo_lectura`: habilita contexto de lectura de la base de empresas para el chat global super; default `true`.
  - la consulta es solo lectura (`SELECT`), parametrizada por `empresa_id`, sin credenciales y con omision de columnas sensibles como password, token, secret, hash, salt, keys, JWT, PIN o certificados.

## 1) Base: pcs_empresas


### Compras Operativas y Proveedores
- empresa_proveedores:
  - id, empresa_id, nit, nombre_comercial, razon_social, direccion, telefono, email, cuenta_bancaria, plazo_dias_pago
- empresa_ordenes_compra:
  - id, empresa_id, proveedor_id, bodega_destino_id, numero_orden, referencia_externa, moneda, total, total_impuestos, estado_orden, estado_pago, fecha_emision, fecha_esperada
- empresa_ordenes_compra_items:
  - id, empresa_id, orden_compra_id, producto_id, descripcion, unidad_medida, cantidad, cantidad_recibida, precio_unitario, impuesto_porcentaje, subtotal, total
- empresa_compras_recepciones:
  - id, empresa_id, orden_compra_id, proveedor_id, bodega_id, numero_factura, total_recepcion, fecha_recepcion

### Tablas de control y core
- schema_migrations:
  - id, scope, version, description, applied_at
- users:
  - email, name, role, empresa_id, documento_identidad, rol_usuario_id, control_aseo_estaciones
  - email_confirmado, email_confirm_token, email_confirm_expira, email_confirmado_en
  - password_hash, password_salt, password_set, password_actualizada_en
  - login_failed_attempts, login_failed_last_at, login_locked_until
  - password_reset_token, password_reset_expira, password_reset_requested_en
  - El login operativo global consulta por `lower(email)` sin `empresa_id` visible y luego resuelve la cuenta concreta con clave o token; las sesiones siguen quedando aisladas por `empresa_id`.
- empresa_estacion_aseo_eventos:
  - empresa_id, estacion_id, estacion_nombre
  - sucia_desde, aseo_fin, duracion_segundos
  - usuario_id, usuario_email, usuario_nombre, rol_nombre
  - origen, estado (`pendiente`/`finalizado`), observaciones
  - mide el tiempo de aseo desde que una estacion queda sucia hasta que un usuario con `control_aseo_estaciones=1` reporta el aseo terminado.
- empresas:
  - nombre, nit, tipo_id, tipo_nombre

### Tablas de clientes e inventario
- clientes:
  - empresa_id, tipo_documento, numero_documento, digito_verificacion
  - tipo_persona, nombre_razon_social, nombre_comercial
  - regimen_fiscal, responsabilidad_tributaria
  - email, telefono, direccion, pais, departamento, municipio, codigo_postal
- bodegas:
  - empresa_id, codigo, nombre, ubicacion, responsable
- categorias_productos:
  - empresa_id, codigo, nombre, descripcion, color_hex, orden
- productos:
  - empresa_id, bodega_principal_id, proveedor_principal_id, categoria_id, sku, codigo_barras
  - nombre, descripcion, categoria, marca, unidad_medida
  - costo, precio, impuesto_porcentaje, stock_minimo, stock_maximo, imagen_url
- combos_productos:
  - empresa_id, codigo, nombre, descripcion, unidad_medida
  - precio, impuesto_porcentaje
  - receta_version, costo_teorico, costo_real, variacion_costo, variacion_costo_porcentaje
- combos_productos_detalle:
  - empresa_id, combo_id, producto_id
  - cantidad, unidad_medida
- combos_productos_versiones:
  - empresa_id, combo_id, receta_version
  - ingredientes_json
  - costo_teorico, costo_real, variacion_costo, variacion_costo_porcentaje
  - motivo
- proveedores:
  - empresa_id, codigo, nombre, documento, contacto, telefono, email, direccion
  - catalogo_referencia, precio_base_referencial, descuento_porcentaje, plazo_pago_dias, condicion_entrega
- servicios:
  - empresa_id, codigo, nombre, descripcion, categoria, duracion_minutos
  - costo_referencial, precio, impuesto_porcentaje, imagen_url
- producto_precios_historial:
  - empresa_id, producto_id
  - costo_anterior, costo_nuevo, precio_anterior, precio_nuevo
  - impuesto_anterior, impuesto_nuevo, motivo, referencia, fecha_cambio
- inventario_existencias:
  - empresa_id, producto_id, bodega_id, cantidad
- inventario_movimientos:
  - empresa_id, producto_id, bodega_origen_id, bodega_destino_id
  - tipo, cantidad, costo_unitario, referencia, fecha_movimiento
- empresa_inventario_configuracion:
  - empresa_id (UNIQUE)
  - politica_costo (`promedio` o `peps`)
- inventario_costos_lotes:
  - empresa_id, producto_id, bodega_id
  - referencia, cantidad, costo_unitario, fecha_movimiento
- inventario_conteos_ciclicos:
  - empresa_id, producto_id, bodega_id
  - cantidad_sistema, cantidad_contada, variacion
  - tipo_ajuste, movimiento_id, referencia, fecha_conteo, usuario_revisor

### Tablas de ventas
- carritos_compras:
  - empresa_id, codigo, nombre, canal_venta, cliente_id
  - estado_carrito, moneda, referencia_externa
  - subtotal, descuento_total, impuesto_total, total
  - activado_en, pagado_en, descuento_tipo, descuento_codigo, descuento_valor
  - devolucion_total, total_pagado, metodo_pago, referencia_pago
- carrito_compra_items:
  - empresa_id, carrito_id, tipo_item, referencia_id, codigo_item, descripcion
  - unidad_medida, cantidad, precio_unitario
  - descuento_porcentaje, impuesto_porcentaje, impuesto_codigo
  - base_gravable, valor_descuento, valor_impuesto, subtotal_linea, total_linea
- empresa_configuracion_general:
  - empresa_id (UNIQUE)
  - imprimir_orden_servicio, area_despacho, copias_orden_servicio, nota_orden_servicio
  - descuentos_habilitados, permitir_descuento_porcentaje, permitir_descuento_codigo, permitir_descuento_valor
  - codigos_descuento
  - lector_codigo_barras_habilitado, lector_codigo_barras_autofoco, lector_codigo_barras_acumular

### Preferencias empresariales por clave
- empresa_estacion_prefs:
  - empresa_id, estacion_id, clave
  - valor, estado, observaciones, usuario_creador, fecha_creacion, fecha_actualizacion
  - clave `chat_flotante.radio_online_enabled`: activa/desactiva la emisora online por empresa cuando `estacion_id = 0`
  - claves `chat_flotante.*`: permiten que chat, robot, secretaria, voz y emisora se persistan por `empresa_id` sin mezclar empresas.

### Tablas de venta publica por empresa
- empresa_venta_publica_configuracion:
  - empresa_id (UNIQUE), empresa_slug (UNIQUE)
  - nombre_tienda, descripcion_tienda, logo_url, banner_url, color_primario
  - moneda, dominio_publico, mostrar_stock
  - wompi_activo, wompi_mode, wompi_public_key
  - wompi_private_key_ref, wompi_integrity_key_ref, wompi_event_key_ref
  - epayco_activo, epayco_mode, epayco_public_key
  - epayco_private_key_ref, epayco_customer_id
- empresa_venta_publica_paginas:
  - empresa_id, slug (UNIQUE por empresa), nombre, descripcion, banner_url
  - orden_visual, estado, usuario_creador, observaciones
- empresa_venta_publica_items:
  - empresa_id, pagina_id, producto_id, codigo_publico
  - nombre, descripcion, precio, moneda, imagen_url
  - stock_publicado, orden_visual, destacado
- empresa_venta_publica_ordenes:
  - empresa_id, codigo_orden (UNIQUE por empresa)
  - comprador_nombre, comprador_email, comprador_telefono
  - moneda, subtotal, descuento_total, impuesto_total, total
  - metodo_pago, estado_pago, transaction_id, referencia_externa
  - pasarela_payload_json, items_json, pagado_en, observaciones

### Tablas de soporte remoto por empresa
- empresa_soporte_remoto_configuracion:
  - empresa_id (UNIQUE)
  - habilitado
  - proveedor_preferido (`novnc`, `guacamole`, `rustdesk_web`, `rustdesk_oss`, `custom_url`)
  - modo_operacion (`agente_web`, `agente_local`, `cliente_local`, `hibrido`)
  - requiere_aprobacion_operador
  - auto_cerrar_minutos
  - max_conexiones_mes
  - max_minutos_mes
  - max_minutos_dia_rustdesk
  - max_dispositivos
  - portal_publico_habilitado
  - rustdesk_server_host, rustdesk_server_key
  - cliente_windows_url, cliente_linux_url
  - servidor_windows_url, servidor_linux_url
  - carpeta_transferencia
  - instrucciones_publicas
- empresa_soporte_remoto_dispositivos:
  - empresa_id, codigo_dispositivo (UNIQUE por empresa)
  - nombre_equipo, alias_operativo, ubicacion
  - sistema_operativo, agente_version
  - stream_url
  - rustdesk_device_id, rustdesk_password_enc
  - carpeta_transferencia
  - acceso_publico_habilitado
  - estado_conexion (`online`, `offline`, `intermitente`)
  - ultimo_heartbeat
  - acceso_pin_hash
- empresa_soporte_remoto_sesiones:
  - empresa_id, dispositivo_id, codigo_sesion (UNIQUE por empresa)
  - solicitada_por, operador_nombre, operador_email
  - motivo, estado_sesion (`pendiente`, `aprobada`, `activa`, `finalizada`, `rechazada`, `expirada`)
  - duracion_minutos_solicitada
  - duracion_minutos_consumida
  - bloqueada_por_limite
  - token_visualizacion_hash
  - url_visualizacion
  - iniciada_en, expira_en, finalizada_en
  - Regla operativa: cuando la empresa usa RustDesk y configura `max_minutos_dia_rustdesk > 0`, el sistema bloquea nuevas activaciones o aprobaciones que excedan el cupo diario consumido para ese `empresa_id`.

### Tabla de metricas de ventas simples por estacion
- empresa_ventas_estacion_metricas:
  - empresa_id, carrito_id, estacion_id, estacion_codigo, estacion_nombre
  - evento_operacion (`venta_pagada`, `cierre_parcial_anulado`, `sesion_recuperada`, `operacion`)
  - metodo_pago, moneda
  - monto_total, monto_pagado, monto_anulado, devolucion_total
  - duracion_segundos
  - activado_en, pagado_en, referencia_operacion
  - fecha_evento

### Tabla de reservas por estacion
- reservas_hotel:
  - empresa_id, carrito_id, estacion_id, codigo_reserva
  - cliente_nombre, cliente_documento, cliente_email, cliente_telefono
  - cantidad_huespedes, fecha_entrada, fecha_salida
  - monto_total, moneda
  - estado_reserva (`pendiente_pago`, `confirmada`, `en_curso`, `cancelada`, `expirada`, `no_show`)
  - estado_pago (`pendiente`, `confirmado`, `cancelado`, `expirado`)
  - referencia_pago, pago_confirmado_en, fecha_expiracion
  - confirmado_por, canal_origen, request_id

### Tabla de tarifas por minutos por estacion
- empresa_tarifas_por_minutos:
  - empresa_id, estacion_id, estacion_codigo, estacion_nombre
  - dia_semana_desde, dia_semana_hasta
  - minutos_base, valor_base
  - minutos_extra, valor_extra
  - moneda, prioridad
- empresa_tarifas_por_minutos_configuracion:
  - empresa_id (UNIQUE)
  - redondeo_modo (`ninguno`, `arriba`, `abajo`, `matematico`)
  - redondeo_unidad
  - monto_minimo_diario
  - monto_maximo_diario

### Tabla de tarifas por dia por estacion
- empresa_tarifas_por_dia:
  - empresa_id, estacion_id, estacion_codigo, estacion_nombre
  - servicio_nombre, valor_dia
  - hora_check_in, hora_check_out
  - moneda, prioridad
  - aplicar_automaticamente (0/1)

### Tabla de codigos de descuento por empresa
- codigos_de_descuento:
  - empresa_id, codigo, tipo_descuento, valor, moneda
  - codigo: formato moderno `PREFIJO-XXXX-XXXX` (ej. DSCT-AB12-CD34). Unicidad garantizada por el índice `ux_codigos_descuento_empresa_codigo` (empresa_id, codigo).
  - monto_minimo_compra, fecha_vencimiento
  - usos_maximos, usos_actuales
  - segmento_cliente, canal_venta
  - horario_desde, horario_hasta, dias_semana
  - max_usos_por_cliente, ventana_horas_fraude
- codigos_descuento_redenciones:
  - empresa_id, codigo_descuento_id, carrito_id, cliente_id
  - codigo, canal_venta, segmento_cliente
  - monto_base, valor_descuento
  - estado_redencion (`aplicada`, `revertida`, `anulada`)
  - motivo, referencia_operacion, fecha_redencion

### Tablas de propinas por empresa
- empresa_propinas_configuracion:
  - empresa_id (UNIQUE)
  - habilitar_propina, porcentaje_propina
  - modo_distribucion (`por_usuario` o `universal`)
  - aplicar_automaticamente
  - pais_fiscal, regimen_fiscal
  - tratamiento_fiscal (`no_gravada` o `gravada`)
  - porcentaje_impuesto_propina
- empresa_propinas_movimientos:
  - empresa_id, carrito_id, cierre_caja_id, venta_referencia
  - usuario_origen, usuario_origen_id, usuario_asignado, usuario_asignado_id
  - `usuario_origen_id` y `usuario_asignado_id` enlazan con `users.id` dentro de la misma empresa; el texto se conserva como etiqueta historica.
  - modo_distribucion, origen_movimiento (`venta`/`ajuste_manual`)
  - ajuste_manual (0/1), referencia_ajuste
  - moneda
  - base_cobro, porcentaje_propina, monto_propina
  - fiscal_pais, fiscal_regimen, fiscal_tratamiento
  - fiscal_porcentaje_impuesto, fiscal_impuesto_monto, fiscal_total
  - conciliado_en
  - fecha_movimiento

### Tablas de comisiones por servicio por empresa
- empresa_comisiones_servicio_configuracion:
  - empresa_id (UNIQUE)
  - habilitar_comisiones, porcentaje_comision
  - filtro_servicio (ej. `lavado`)
  - aplicar_automaticamente
- empresa_comisiones_servicio_escalas:
  - empresa_id
  - rol_operacion, servicio_filtro
  - porcentaje_comision, tope_comision, prioridad
- empresa_comisiones_servicio_movimientos:
  - empresa_id, carrito_id, carrito_item_id
  - servicio_id, servicio_codigo, servicio_nombre, servicio_categoria
  - usuario_origen, usuario_origen_id, usuario_lavador, usuario_lavador_id, rol_operacion, escala_id
  - `usuario_origen_id` y `usuario_lavador_id` enlazan con `users.id` dentro de la misma empresa; el texto se conserva como etiqueta historica.
  - venta_referencia, moneda
  - base_servicio, porcentaje_comision, monto_comision_bruto, tope_comision_aplicado, monto_comision
  - origen_movimiento (`venta`/`ajuste_manual`)
  - ajuste_manual (0/1), referencia_ajuste, ajuste_estado (`pendiente`/`aprobado`/`rechazado`)
  - aprobado_por, aprobado_en
  - liquidacion_nomina_id, periodo_liquidacion_desde, periodo_liquidacion_hasta
  - liquidado_en, liquidado_por
  - fecha_movimiento

### Tablas de configuracion operativa de cobro por empresa, rol y contexto operativo
- empresa_configuracion_operativa:
  - empresa_id (UNIQUE)
  - metodo_pago_efectivo
  - metodo_pago_tarjeta_credito
  - metodo_pago_tarjeta_debito
  - metodo_pago_transferencia_bancaria
  - metodo_pago_mixto
  - metodo_pago_codigo_descuento
  - habilitar_propinas
  - habilitar_comisiones
- empresa_configuracion_operativa_roles:
  - empresa_id
  - rol
  - metodo_pago_efectivo
  - metodo_pago_tarjeta_credito
  - metodo_pago_tarjeta_debito
  - metodo_pago_transferencia_bancaria
  - metodo_pago_mixto
  - metodo_pago_codigo_descuento
  - habilitar_propinas
  - habilitar_comisiones
  - indice unico: (empresa_id, rol)
- empresa_configuracion_operativa_politicas:
  - empresa_id
  - canal_venta (`''`/`mostrador`/`app`/`estacion`/`reserva`/`online`/`delivery`/`kiosko`)
  - sucursal_id (0 = aplica a todas)
  - turno (`''` = aplica a todos)
  - prioridad (entero menor = mayor prioridad)
  - metodo_pago_efectivo
  - metodo_pago_tarjeta_credito
  - metodo_pago_tarjeta_debito
  - metodo_pago_transferencia_bancaria
  - metodo_pago_mixto
  - metodo_pago_codigo_descuento
  - habilitar_propinas
  - habilitar_comisiones
  - indice unico: (empresa_id, canal_venta, sucursal_id, turno)
- empresa_configuracion_operativa_historial:
  - empresa_id
  - evento (`publicar`/`simular`/`rollback`)
  - rollback_de_historial_id
  - snapshot_json (estado completo de configuracion para recuperacion)
  - simulacion_json (resultado contextual opcional)
  - usuario_creador
  - estado
  - observaciones
  - indice: (empresa_id, fecha_creacion DESC, id DESC)

### Tablas de impresoras operativas por empresa
- empresa_impresoras:
  - empresa_id, codigo, nombre
  - tipo_conexion (`red`/`usb`/`windows`/`bluetooth`)
  - direccion, area_operativa
  - formato_impresion (`pos`/`carta`)
  - es_predeterminada (0/1)
  - indices: unico `(empresa_id, codigo)`, por estado y predeterminada
- empresa_impresoras_funcionalidades:
  - empresa_id, funcionalidad, impresora_id
  - prioridad
  - indice unico: `(empresa_id, funcionalidad)`
- empresa_impresoras_productos:
  - empresa_id, producto_id, impresora_id
  - indice unico: `(empresa_id, producto_id)`
- Regla de resolucion operativa:
  - prioridad de asignacion: `producto` -> `funcionalidad` -> `predeterminada`.

### Tablas de calculadora operativa por empresa
- empresa_calculadora_configuracion:
  - empresa_id (UNIQUE)
  - integrar_carritos (0/1)
  - integrar_cotizaciones (0/1)
  - usuario_creador
  - estado
  - observaciones
- empresa_calculadora_operaciones:
  - empresa_id
  - expresion
  - resultado
  - etiquetas_json
  - cliente_id, cliente_nombre
  - documento_tipo, documento_codigo
  - carrito_id, cotizacion_id
  - fecha_operacion
  - metadata_json
  - usuario_creador
  - estado
  - observaciones
  - indices: (empresa_id, fecha_operacion DESC, id DESC), (empresa_id, usuario_creador, fecha_operacion DESC), (empresa_id, carrito_id, cotizacion_id)

### Tablas de finanzas empresariales
- empresa_finanzas_movimientos:
  - empresa_id, tipo_movimiento, codigo, fecha_movimiento
  - periodo_contable
  - categoria, subcategoria, concepto, descripcion, metodo_pago, moneda
  - monto, impuesto
  - retencion_fuente, retencion_ica, retencion_iva, total_retenciones
  - total, total_neto
  - tercero_nombre, tercero_documento
  - tipo_comprobante, numero_comprobante, comprobante_url
  - referencia_externa, aprobado_por
  - UNIQUE(empresa_id, codigo)
- empresa_finanzas_periodos:
  - empresa_id, periodo (UNIQUE por empresa)
  - fecha_inicio, fecha_fin
  - fecha_cierre, cerrado_por
  - estado (abierto/cerrado/inactivo)
- empresa_cierres_caja:
  - empresa_id, sucursal_id, caja_codigo, turno
  - fecha_operacion, fecha_apertura, fecha_cierre
  - estado_cierre (abierto/cerrado/aprobado/anulado)
  - apertura_monto, ingresos_efectivo, egresos_efectivo, retiros_efectivo
  - caja_teorica, caja_fisica, diferencia_caja
  - tiene_incidencia, umbral_incidencia
  - propinas_movimientos, propinas_total, propinas_ajustes
  - propinas_impuesto, propinas_neto
  - propinas_conciliado_en, propinas_conciliado_por
  - cerrado_por, aprobado_por, aprobado_en
  - UNIQUE(empresa_id, sucursal_id, caja_codigo, fecha_operacion, turno)
- empresa_corte_caja_configuracion:
  - empresa_id (UNIQUE)
  - mostrar_encabezado, mostrar_empresa_datos, mostrar_fecha_hora, mostrar_usuario_reporte, mostrar_consecutivo
  - mostrar_resumen, mostrar_numero_facturas, mostrar_total_ventas
  - mostrar_efectivo, mostrar_debito, mostrar_credito, mostrar_transferencias, mostrar_otros_medios
  - mostrar_ingresos, mostrar_egresos, mostrar_anulaciones, mostrar_devoluciones
  - mostrar_caja_esperada, mostrar_diferencia_caja, mostrar_ventas_detalle
  - mostrar_detalle_fecha_entrada, mostrar_detalle_fecha_salida, mostrar_detalle_numero_venta
  - mostrar_detalle_estacion, mostrar_detalle_cajero, mostrar_detalle_medio_pago, mostrar_detalle_total
  - mostrar_movimientos, mostrar_items, mostrar_total_productos, mostrar_total_servicios
  - mostrar_sensores_puertas, mostrar_auditoria
  - formato_impresion, estado, observaciones
- empresa_finanzas_configuracion:
  - empresa_id (UNIQUE)
  - habilitar_ingresos, habilitar_egresos, moneda
  - categorias_ingreso, categorias_egreso
  - prefijo_ingreso, prefijo_egreso
  - formato_impresion, requiere_aprobacion
  - integracion_contable_destino
  - cuenta_caja_bancos, cuenta_ingresos, cuenta_iva_generado
  - cuenta_gastos, cuenta_iva_descontable
  - cuenta_retenciones_cobrar, cuenta_retenciones_pagar
  - cuentas_ingreso_categoria, cuentas_egreso_categoria
- empresa_finanzas_bancos_movimientos:
  - empresa_id, periodo_contable, fecha_movimiento, fecha_valor
  - cuenta_bancaria, banco_nombre
  - tipo_movimiento (`ingreso`/`egreso`), descripcion
  - referencia_bancaria, documento_codigo
  - moneda, monto, total
  - movimiento_finanzas_id
  - estado_conciliacion (`pendiente`/`conciliado`/`con_desviacion`)
  - conciliado_en, conciliado_por
  - origen, hash_movimiento
  - UNIQUE(empresa_id, hash_movimiento)
- empresa_plan_cuentas:
  - se usa como plan de cuentas operativo del modulo financiero; la UI permite plantilla PUC por tipo de empresa, alta/edicion/inactivacion y validacion de `admite_movimiento`.
- empresa_cuentas_por_cobrar / empresa_cuentas_por_pagar:
  - gestionan cartera operativa con valor original, valor pagado, saldo, estado_cartera, dias_mora, referencia_pagos_json y notas.
  - los abonos/pagos se registran como `empresa_finanzas_movimientos` y actualizan saldo/estado para mantener trazabilidad financiera por `empresa_id`.

### Tabla de eventos contables empresariales
- empresa_eventos_contables:
  - empresa_id, modulo, evento
  - entidad, entidad_id
  - documento_tipo, documento_codigo
  - periodo_contable
  - monto_total, moneda
  - payload_json, origen
  - fecha_evento
  - procesado, fecha_procesado
  - intentos_procesamiento, fecha_ultimo_intento
  - error_procesamiento, asiento_contable_id
  - incluye evento de trazabilidad: `finanzas.tarifa_por_minutos_calculada`

### Tabla canonica de asientos contables empresariales
- empresa_asientos_contables:
  - empresa_id, evento_contable_id
  - modulo, evento
  - fecha_asiento, periodo_contable
  - documento_tipo, documento_codigo
  - moneda
  - total_debito, total_credito, diferencia
  - lineas_json
  - hash_idempotencia
  - payload_origen_json
  - fecha_procesado, procesado_por
  - UNIQUE(empresa_id, evento_contable_id)
  - UNIQUE(empresa_id, hash_idempotencia)

### Tabla de auditoria empresarial
- empresa_auditoria_eventos:
  - empresa_id, modulo, accion
  - recurso, recurso_id
  - metodo_http, endpoint
  - resultado, codigo_http
  - request_id, ip_origen, user_agent
  - metadata_json
  - retencion_dias, fecha_evento, fecha_expiracion

### Objetos de busqueda full-text de auditoria (FTS operativa)
- empresa_auditoria_eventos_fts (tabla virtual):
  - indexa contenido textual de `empresa_auditoria_eventos` para `search` full-text.
  - columnas indexadas: `modulo`, `accion`, `recurso`, `endpoint`, `metadata_json`, `observaciones`.
- Triggers de sincronizacion FTS:
  - `empresa_auditoria_eventos_ai`: inserta en FTS cuando se crea evento.
  - `empresa_auditoria_eventos_au`: refresca fila FTS cuando se actualiza evento.
  - `empresa_auditoria_eventos_ad`: elimina fila FTS cuando se elimina evento.
- Backfill inicial FTS:
  - al crear el esquema FTS se repueblan filas existentes para consistencia historica.

### Tablas de documentos transaccionales canonicos
- empresa_facturacion_documentos:
  - empresa_id, tipo_documento, documento_codigo
  - estado_documento, estado_anterior, evento_ultimo
  - periodo_contable, monto_total, moneda
  - numero_legal, codigo_validacion, pais_codigo, ambiente_fe
  - fecha_documento, entidad_relacionada_id
  - uso operativo actual: soporta `factura_electronica`, `nota_credito` y `comprobante_pago`
  - UNIQUE(empresa_id, tipo_documento, documento_codigo)
- empresa_compras_documentos:
  - empresa_id, proveedor_id, tipo_documento, documento_codigo
  - estado_documento, estado_anterior, evento_ultimo
  - periodo_contable, monto_total, moneda
  - fecha_documento, entidad_relacionada_id
  - requiere_aprobacion, niveles_aprobacion_requeridos, nivel_aprobacion_actual, aprobadores_json
  - recepcion_detalle_json, recepcion_resumen_json
  - validacion_documental_estado, proveedor_documento_ref, factura_documento_ref, entrada_documento_ref
  - UNIQUE(empresa_id, tipo_documento, documento_codigo)

### Tablas de compras avanzadas
- empresa_compras_requisiciones:
  - empresa_id, codigo, solicitante, area, centro_costo, prioridad
  - fecha_solicitud, fecha_necesidad, estado_flujo, total_estimado, justificacion
  - UNIQUE(empresa_id, codigo)
- empresa_compras_requisicion_items:
  - empresa_id, requisicion_id, producto_id, producto_nombre, cantidad_solicitada
  - cantidad_recibida, unidad, costo_estimado, proveedor_sugerido, especificacion, estado
- empresa_compras_cotizaciones:
  - empresa_id, requisicion_id, proveedor_id, proveedor_nombre, numero
  - fecha_cotizacion, validez_hasta, tiempo_entrega_dias, subtotal, impuestos, total
  - condiciones_pago, observaciones, estado
  - UNIQUE(empresa_id, requisicion_id, numero)
- empresa_compras_recepciones_avanzadas:
  - empresa_id, requisicion_id, cotizacion_id, proveedor_id, proveedor_nombre
  - documento, fecha_recepcion, estado_recepcion, responsable, observaciones
- empresa_compras_recepcion_items_avanzadas:
  - empresa_id, recepcion_id, requisicion_item_id, producto_nombre
  - cantidad_ordenada, cantidad_recibida, cantidad_pendiente, costo_unitario, lote, estado_calidad

### Tablas de IA empresarial
- empresa_ai_consultas:
  - empresa_id, provider, model_id
  - pregunta, respuesta
  - prompt_tokens, completion_tokens, total_tokens
  - fecha_consulta, plan_actual
- empresa_ai_uso_diario:
  - empresa_id, provider, model_id, fecha_uso
  - consultas_total, tokens_total
  - plan_actual
  - UNIQUE(empresa_id, provider, model_id, fecha_uso)
- empresa_ai_modelo_preferido:
  - empresa_id, admin_email
  - provider, model_id
  - UNIQUE(empresa_id, admin_email)

### Tabla de configuracion empresarial
- empresa_configuracion_avanzada:
  - empresa_id (UNIQUE)
  - modo_documento_venta (`factura_electronica` o `comprobante_pago`)
  - tipo_documento_emisor, nit, digito_verificacion
  - razon_social, nombre_comercial, regimen_fiscal, responsabilidad_tributaria
  - email_facturacion, telefono_facturacion, direccion_fiscal, departamento, municipio
  - pais_codigo, codigo_postal
  - ambiente_fe, tipo_operacion, prefijo_factura
  - resolucion_numero, resolucion_fecha_desde, resolucion_fecha_hasta
  - consecutivo_desde, consecutivo_hasta, proximo_consecutivo
  - formato_impresion, imprimir_copia_factura, mostrar_logo, mostrar_logo_empresa, mostrar_logo_sistema, logo_url
  - pie_factura, notas_legales
  - color_carrito_activo, color_carrito_inactivo
  - moneda_codigo, sistema_numerico, usar_decimales, cantidad_decimales
  - Las banderas historicas que pudieran existir como BOOLEAN en PostgreSQL se regularizan a INTEGER `0/1` durante `EnsureEmpresaConfiguracionAvanzadaSchema` para mantener un unico contrato de guardado.

### Tabla de facturacion electronica por pais
- facturacion_electronica_pais:
  - empresa_id, pais_codigo, pais_nombre, moneda_codigo
  - proveedor, ambiente, tipo_documento_emisor, identificador_fiscal
  - razon_social, email_facturacion, telefono_facturacion, direccion_fiscal
  - prefijo_factura, resolucion_numero, api_base_url, campos_pais_json
  - UNIQUE(empresa_id, pais_codigo)

### Tabla de reintentos de integracion fiscal FE
- facturacion_electronica_reintentos:
  - empresa_id, tipo_documento, documento_codigo
  - pais_codigo, proveedor, ambiente
  - estado_envio (`pendiente`, `fallido`, `enviado`, `reconciliado`, `contingencia`, `no_aplica`)
  - intentos, max_intentos, proximo_intento, fecha_ultimo_intento
  - ultimo_error, respuesta_proveedor_json
  - contingencia_activa, fecha_contingencia
  - referencia_externa
  - numero_legal, codigo_validacion, fecha_emision_legal
  - UNIQUE(empresa_id, tipo_documento, documento_codigo)

### Tablas de chat y tareas (nuevo modulo)
- chat_tareas_conversaciones:
  - empresa_id, titulo, descripcion, prioridad, estado_conversacion, ultimo_mensaje_en
- chat_tareas_participantes:
  - empresa_id, conversacion_id
  - participante_tipo, participante_ref_id, nombre, email, activo_chat
- chat_tareas_mensajes:
  - empresa_id, conversacion_id
  - autor_tipo, autor_ref_id, autor_nombre, autor_email
  - contenido, tipo_mensaje, fecha_envio
- chat_tareas_adjuntos:
  - empresa_id, mensaje_id
  - tipo_archivo, nombre_archivo, mime_type, file_url, tamano_bytes, duracion_segundos
- chat_tareas:
  - empresa_id, conversacion_id
  - titulo, descripcion, prioridad, fecha_limite
  - asignado_tipo, asignado_ref_id, asignado_nombre, asignado_email
  - creado_por_tipo, creado_por_email
  - estado_tarea, porcentaje_avance, completada_en
- chat_tareas_citas:
  - empresa_id, conversacion_id
  - titulo, descripcion, tipo_cita
  - fecha_inicio, fecha_fin, ubicacion
  - notificar_minutos_antes
  - creado_por_tipo, creado_por_ref_id, creado_por_nombre, creado_por_email
  - estado_cita (`programada`/`completada`/`cancelada`)
  - recordatorio_enviado, recordatorio_enviado_en
  - visibilidad (`empresa`/`privada`)

### Tablas de ubicacion GPS por empresa
- empresa_gps_dispositivos:
  - empresa_id, codigo, nombre, descripcion
  - ultima_latitud, ultima_longitud, ultima_precision_metros, ultima_velocidad_kmh
  - ultimo_reporte_en
  - UNIQUE(empresa_id, codigo)
- empresa_gps_recorridos:
  - empresa_id, dispositivo_id
  - latitud, longitud, precision_metros, velocidad_kmh
  - rumbo_grados, altitud_metros, fuente, capturado_en

### Tabla de asistencia de empleados por empresa
- empresa_asistencia_empleados:
  - empresa_id, empleado_id, empleado_codigo, empleado_nombre, empleado_documento
  - cargo, turno, fecha_asistencia
  - hora_entrada, hora_salida, minutos_tarde, horas_trabajadas
  - estado_asistencia, novedad
- empresa_asistencia_configuracion:
  - empresa_id (UNIQUE)
  - tolerancia_entrada_minutos, tolerancia_salida_minutos
  - hora_inicio_turno_manana, hora_inicio_turno_tarde, hora_inicio_turno_noche
  - permitir_turno_nocturno, permitir_turno_cruzado
- empresa_asistencia_periodos_cerrados:
  - empresa_id, periodo_desde, periodo_hasta
  - fecha_cierre, cerrado_por, motivo
  - se usa para bloquear edicion de asistencia en periodos cerrados

### Tablas de nomina de sueldos por empresa
- empresa_nomina_configuracion:
  - empresa_id (UNIQUE)
  - pais_codigo, moneda
  - horas_ordinarias_semana, horas_ordinarias_dia, dias_nomina_mes, divisor_hora_ordinaria
  - hora_nocturna_desde, hora_nocturna_hasta
  - recargo_nocturno_porcentaje
  - hora_extra_diurna_porcentaje, hora_extra_nocturna_porcentaje
  - recargo_dominical_diurno_porcentaje, recargo_dominical_nocturno_porcentaje
  - hora_extra_dominical_diurna_porcentaje, hora_extra_dominical_nocturna_porcentaje
  - deduccion_salud_porcentaje, deduccion_pension_porcentaje, deduccion_fondo_solidaridad_porcentaje
- empresa_nomina_empleados:
  - empresa_id, empleado_id, empleado_codigo, empleado_nombre, empleado_documento
  - cargo, tipo_contrato, fecha_ingreso
  - salario_basico_mensual, auxilio_transporte_mensual, bonificacion_fija_mensual, deduccion_fija_mensual
  - jornada_horas_dia, incluir_auxilio_transporte
- empresa_nomina_festivos:
  - empresa_id, fecha_festivo (UNIQUE por empresa), descripcion
- empresa_nomina_liquidaciones:
  - empresa_id, empleado_nomina_id, empleado_id, empleado_codigo, empleado_nombre, empleado_documento, cargo
  - periodo_desde, periodo_hasta, dias_liquidados
  - horas_asistencia_total, registros_asistencia
  - horas_ordinarias, horas_recargo_nocturno
  - horas_extra_diurnas, horas_extra_nocturnas
  - horas_dominicales_diurnas, horas_dominicales_nocturnas
  - horas_extra_dominicales_diurnas, horas_extra_dominicales_nocturnas
  - valor_hora_ordinaria, base_salario_proporcional
  - valor_recargo_nocturno, valor_dominical_diurno, valor_dominical_nocturno
  - valor_extra_diurna, valor_extra_nocturna
  - valor_extra_dominical_diurna, valor_extra_dominical_nocturna
  - total_recargos_horas_extras, auxilio_transporte, bonificacion
  - comisiones_servicio_total, comisiones_servicio_movimientos, comisiones_servicio_ajustes
  - devengado_total
  - ingreso_base_cotizacion, deduccion_salud, deduccion_pension, deduccion_fondo_solidaridad
  - deduccion_fija, otras_deducciones, deduccion_total, neto_pagar
  - origen_calculo, resumen_json, fecha_generacion

### Tabla de registro vehicular por empresa
- empresa_vehiculos_registro:
  - empresa_id
  - patente, tipo_vehiculo
  - marca, modelo, color
  - conductor_nombre, conductor_documento
  - propietario_nombre, propietario_documento
  - motivo_ingreso, referencia_externa
  - fecha_ingreso, fecha_salida
  - estado_registro (`en_empresa` o `retirado`)
  - usuario_salida
- empresa_vehiculos_configuracion:
  - empresa_id (UNIQUE)
  - pais_codigo
  - patente_regex, patente_descripcion
  - evitar_duplicado_activo
  - estado

### Tablas ERP extendidas por empresa (2026-04-07)
- empresa_cotizaciones_venta:
  - empresa_id, codigo, cliente_id, cliente_nombre
  - fecha_documento, vigencia_hasta, estado_documento
  - subtotal, descuento_total, impuesto_total, total, moneda
  - origen, convertido_pedido_id
- empresa_pedidos_venta:
  - empresa_id, codigo, cliente_id, cliente_nombre, cotizacion_id
  - fecha_pedido, fecha_entrega_estimada, estado_pedido
  - subtotal, descuento_total, impuesto_total, total, moneda
- empresa_devoluciones_venta:
  - empresa_id, codigo, carrito_id, documento_referencia, motivo
  - fecha_devolucion, estado_devolucion, subtotal, impuesto_total, total, moneda
- empresa_plan_cuentas:
  - empresa_id, codigo, nombre, tipo_cuenta, naturaleza, nivel
  - cuenta_padre_codigo, admite_movimiento, aplica_impuesto
  - plantilla_tipo_empresa, plantilla_codigo, plantilla_version
  - cuenta_clave, requerida, orden_plantilla
- empresa_cuentas_por_cobrar:
  - empresa_id, codigo, cliente_id, cliente_nombre
  - documento_tipo, documento_codigo, fecha_emision, fecha_vencimiento
  - valor_original, valor_pagado, saldo, estado_cartera, moneda
  - periodo_contable, referencia_pagos_json, fecha_ultimo_pago
  - conciliado_en, conciliado_por
- empresa_cuentas_por_pagar:
  - empresa_id, codigo, proveedor_id, proveedor_nombre
  - documento_tipo, documento_codigo, fecha_emision, fecha_vencimiento
  - valor_original, valor_pagado, saldo, estado_cartera, moneda
  - periodo_contable, referencia_pagos_json, fecha_ultimo_pago
  - conciliado_en, conciliado_por
- empresa_creditos:
  - empresa_id, codigo, cliente_id, cliente_nombre
  - tipo_credito (`rotativo`/`fijo`/`cuotas`)
  - monto_aprobado, cupo_credito, saldo_actual, saldo_disponible
  - tasa_interes, tasa_mora
  - plazo_dias, plazo_cuotas
  - fecha_inicio, fecha_vencimiento, fecha_ultimo_pago
  - dias_mora, clasificacion_cartera (`al_dia`/`vencido`/`castigado`)
  - bloqueo_automatico_mora, venta_origen_id, documento_origen, estado_credito
- empresa_creditos_cuotas:
  - empresa_id, credito_id, numero_cuota
  - fecha_vencimiento
  - valor_cuota, capital_cuota, interes_cuota, interes_mora
  - valor_pagado, saldo_cuota, estado_cuota (`pendiente`/`parcial`/`pagada`/`vencida`/`anulada`)
  - fecha_ultimo_pago
- empresa_creditos_movimientos:
  - empresa_id, credito_id, cuota_id
  - tipo_movimiento (`abono`/`cargo_interes`/`interes`/`mora`/`reverso`/`ajuste`/`refinanciacion`)
  - monto, capital_aplicado, interes_aplicado, mora_aplicada
  - metodo_pago, referencia_pago, comprobante
  - aplicado_automatico, fecha_movimiento
- empresa_creditos_workflow:
  - empresa_id, credito_id, movimiento_origen_id
  - tipo_solicitud (`reverso_abono`/`anulacion_abono`/`refinanciacion`)
  - estado_solicitud (`pendiente_aprobacion`/`aprobada`/`rechazada`/`ejecutada`/`error_ejecucion`)
  - nivel_aprobacion_actual, nivel_aprobacion_requerido
  - motivo_solicitud, motivo_decision
  - payload_json, historial_aprobaciones_json, resultado_json
  - aprobado_por, codigo_aprobacion, rechazado_por
  - fecha_solicitud, fecha_decision, fecha_ejecucion
  - movimiento_resultado_id
- empresa_creditos_clientes_limites:
  - empresa_id, cliente_id
  - limite_saldo_total
  - max_creditos_activos
  - requiere_aprobacion_exceso
  - fecha_actualizacion, usuario_creador, estado, observaciones
  - indice unico: (empresa_id, cliente_id)
- empresa_backups:
  - empresa_id, codigo, nombre, descripcion
  - version_schema, alcance, tipo_backup
  - include_tables_json, exclude_tables_json
  - total_tablas, total_registros, tamano_bytes
  - hash_contenido, snapshot_json, metadata_json
  - restaurado_en, restaurado_por
- empresa_backups_restauraciones:
  - empresa_id, backup_id, codigo_backup
  - tablas_restauradas, registros_restaurados
  - tablas_omitidas_json, resultado, detalle_json
  - fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
- inventario_lotes_series:
  - empresa_id, producto_id, bodega_id, tipo_control, codigo_lote_serie
  - fecha_fabricacion, fecha_vencimiento, cantidad_inicial, cantidad_disponible
  - reservado_cantidad, vendido_cantidad, costo_unitario
  - estado_lote, bloqueado_venta, bloqueo_motivo
  - ultima_operacion_tipo, ultima_operacion_ref, ultima_operacion_en
- inventario_lotes_series_movimientos:
  - empresa_id, lote_serie_id, producto_id, bodega_id, codigo_lote_serie
  - tipo_operacion, cantidad, saldo_lote
  - referencia_tipo, referencia_codigo
  - cliente_id, cliente_nombre, detalle_json, fecha_operacion
- empresa_devoluciones_proveedor:
  - empresa_id, codigo, proveedor_id, proveedor_nombre, documento_compra_codigo
  - fecha_devolucion, motivo, estado_devolucion, subtotal, impuesto_total, total, moneda
  - periodo_contable, impacto_contable_movimiento_id, impacto_contable_evento_id
  - fecha_contabilizacion, contabilizado_por, total_reintegrado
- empresa_rrhh_vacaciones_licencias:
  - empresa_id, codigo, empleado_id, empleado_nomina_id, empleado_nombre, tipo_novedad
  - fecha_inicio, fecha_fin, dias, remunerada, estado_novedad, soporte_url, aprobado_por
  - nivel_aprobacion_actual, nivel_aprobacion_requerido, aprobadores_json, historial_aprobaciones_json, fecha_aprobacion_final
  - periodo_acumulado_desde, periodo_acumulado_hasta, saldo_dias_antes, saldo_dias_despues, saldo_snapshot_json
  - nomina_liquidacion_id, nomina_periodo_desde, nomina_periodo_hasta, nomina_vinculada_en, nomina_vinculada_por
- crm_leads:
  - empresa_id, codigo, nombre, empresa_origen, email, telefono, canal_origen
  - estado_lead, valor_potencial, probabilidad, propietario, proximo_contacto
- crm_interacciones:
  - empresa_id, codigo, lead_id, cliente_id, tipo_interaccion, fecha_interaccion
  - resumen, resultado, usuario_responsable, proxima_accion, estado_interaccion
- crm_campanas:
  - empresa_id, codigo, nombre, canal, objetivo, presupuesto
  - fecha_inicio, fecha_fin, estado_campana, audiencia, kpi_objetivo, resultado_json
- empresa_crm_metas_comerciales:
  - empresa_id, periodo, propietario, canal
  - meta_valor, meta_leads, meta_conversion_pct, estado
  - usuario_creador, fecha_creacion, fecha_actualizacion
- produccion_bom:
  - empresa_id, codigo, producto_id, producto_nombre, version
  - rendimiento, unidad_medida, costo_estimado_total, estado_bom
- produccion_bom_detalle:
  - empresa_id, bom_id, insumo_producto_id, insumo_nombre, cantidad, unidad_medida
  - costo_unitario, costo_total, merma_porcentaje
- produccion_ordenes:
  - empresa_id, codigo, bom_id, producto_id, producto_nombre
  - cantidad_programada, cantidad_producida, fecha_programada, fecha_inicio, fecha_fin
  - estado_orden, costo_estimado, costo_real, responsable
- logistica_transportistas:
  - empresa_id, codigo, nombre, documento, telefono, email, placa, vehiculo_tipo
  - capacidad_carga, estado_transportista
- logistica_rutas:
  - empresa_id, codigo, nombre, origen, destino, distancia_km, tiempo_estimado_min, estado_ruta
- logistica_envios:
  - empresa_id, codigo, cliente_id, cliente_nombre, documento_referencia
  - direccion_entrega, ruta_id, transportista_id
  - fecha_programada, fecha_salida, fecha_entrega, estado_envio, costo_envio
  - latitud, longitud, observaciones_seguimiento
- empresa_documentos_gestion:
  - empresa_id, codigo, modulo, entidad, entidad_id, documento_codigo
  - nombre_documento, tipo_documento, mime_type, url_archivo, hash_archivo, tamano_bytes
  - version, estado_documento
- empresa_documentos_firmas:
  - empresa_id, codigo, documento_gestion_id, tipo_firma
  - firmante_nombre, firmante_documento, firmante_email
  - certificado_serial, algoritmo_firma, hash_firma, fecha_firma, validez_hasta, estado_firma
- empresa_integraciones_apis:
  - empresa_id, codigo, nombre_integracion, tipo_integracion, base_url, auth_tipo, api_key_ref
  - estado_integracion, ultima_sincronizacion, respuesta_ultimo_sync
- empresa_integraciones_bancos:
  - empresa_id, codigo, banco_nombre, tipo_conexion, numero_cuenta, titular, moneda
  - api_endpoint, credencial_ref, estado_integracion, ultima_conciliacion
- empresa_dian_configuracion:
  - empresa_id (UNIQUE), codigo
  - nit, digito_verificacion, razon_social, tipo_ambiente
  - software_id, software_pin
  - usar_software_compartido, software_id_compartido_ref, software_pin_compartido_ref
  - test_set_id
  - certificado_url, certificado_clave_ref
  - prefijo, resolucion_numero, resolucion_fecha_desde, resolucion_fecha_hasta
  - rango_desde, rango_hasta, consecutivo_actual
  - url_dian, token_emisor_ref, ultimo_envio, estado_dian

## 2) Base: pcs_superadministrador

### Tablas de control y administracion
- schema_migrations:
  - id, scope, version, description, applied_at
- administradores:
  - email, name, role, photo
  - acepta_contrato, contrato_version_aceptada, fecha_acepta_contrato
- sesiones:
  - admin_email, token, ip, user_agent, fecha_inicio, fecha_fin, activo
  - fecha_fin se usa para expiracion y revocacion de sesion
- configuraciones:
  - config_key (PK), value, encrypted
  - claves relevantes de seguridad de usuarios empresa:
    - usuarios.password_min_length
    - usuarios.password_require_uppercase
    - usuarios.password_require_lowercase
    - usuarios.password_require_digit
    - usuarios.password_require_symbol
    - usuarios.password_rotation_days
    - gmail.smtp_test_mode
- super_contrato_versiones:
  - version
  - titulo, resumen, contenido
  - nota_aceptacion, resumen_cambio
  - fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
- super_errores_sistema:
  - nivel, tipo_error, mensaje, mensaje_publico
  - detalle, stack_trace
  - empresa_id, usuario_email
  - endpoint, modulo, metodo_http, codigo_http, request_id
  - origen, ip, user_agent, metadata_json
  - fecha_error, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
- super_ai_consultas:
  - admin_email, provider, model_id, pregunta, respuesta
  - prompt_tokens, completion_tokens, total_tokens, fecha_consulta, plan_actual
  - fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
- super_ai_uso_diario:
  - admin_email, provider, model_id, fecha_uso
  - consultas_total, tokens_total, plan_actual
  - fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
  - indice unico: (admin_email, provider, model_id, fecha_uso)
- super_ai_modelo_preferido:
  - admin_email, provider, model_id
  - fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
  - indice unico: (admin_email)

### Tablas de catalogos globales
- tipos_de_empresas:
  - nombre
- roles_de_usuario:
  - tipo_empresa_id, nombre, descripcion
- roles_de_usuario_permisos:
  - rol_id, modulo, accion, permitido
  - indice unico: (rol_id, modulo, accion)
- roles_de_usuario_paginas_permisos:
  - rol_id, pagina_clave, permitido
  - indice unico: (rol_id, pagina_clave)
- tipos_de_licencia:
  - nombre
- licencias:
  - empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias
  - modulos_habilitados, super_rol_habilitado
  - fecha_inicio, fecha_fin, activo
  - fecha_actualizacion, usuario_creador, estado, observaciones
- licencias_activaciones_gratis:
  - licencia_id, empresa_id, discount_code, asesor_id, motivo
  - fecha_creacion, fecha_actualizacion, estado, observaciones
  - indice unico: `(licencia_id, empresa_id)`
  - indice unico parcial: `empresa_id` cuando `estado` activo, para garantizar una sola prueba/gratis por empresa
  - Descripcion: marca operativa para impedir reutilizar pruebas/licencias sin pago por empresa y conservar trazabilidad de descuento/cortesia/asesor.
- tipo_empresa_preconfiguraciones:
  - tipo_empresa_id, enabled, nombre, descripcion, config_json
  - usuario_creador, fecha_creacion, fecha_actualizacion, estado
  - la siembra completa faltantes por `tipo_empresa_id` activo y conserva personalizaciones existentes
- super_venta_digital_configuracion:
  - nombre_tienda, descripcion_tienda, logo_url, banner_url, color_primario
  - moneda, wompi_activo
- super_venta_digital_items:
  - codigo_publico, nombre, descripcion, precio, moneda, imagen_url
  - licencia_codigo, instrucciones_archivo_url
  - orden_visual, destacado
- super_venta_digital_ordenes:
  - codigo_orden
  - item_id, item_nombre, item_precio, item_moneda
  - comprador_nombre, comprador_email, comprador_telefono
  - metodo_pago, estado_pago
  - transaction_id, referencia_externa
  - provider_payload_json, pagado_en, error_pago
  - correo_entregado, correo_entregado_en, correo_error
  - licencia_codigo_enviado, instrucciones_archivo_url

### Tablas de pagos y metricas
- pagos_wompi:
  - licencia_id, empresa_id, transaction_id, reference, status, raw_payload
  - discount_code (TEXT) : código de descuento aplicado por cliente (opcional)
  - asesor_id (TEXT) : codigo de asesor comercial que originó la venta (opcional)
  - payment_method, provider_payload_json (opcional): metadatos del proveedor/método de pago

  - Descripción: tabla canonica para registrar pagos/operaciones de Wompi/Nequi y activaciones manuales. Se registran metadatos de descuento y referencia de asesor para habilitar cálculo y trazabilidad de comisiones.
- pagos_epayco:
  - licencia_id, empresa_id, transaction_id, reference, status, raw_payload
  - discount_code (TEXT) : codigo de descuento aplicado por cliente (opcional)
  - asesor_id (TEXT) : codigo de asesor comercial que genero la venta (opcional)
  - payment_method, provider_payload_json (opcional): metadatos del proveedor/metodo de pago

  - Descripcion: tabla canonica para registrar operaciones de checkout/confirmacion de Epayco, incluyendo estado por referencia y soporte de activacion automatica de licencia tras aprobacion.
- super_correo_notificaciones_prueba:
  - tipo, empresa_id, destinatario, asunto, cuerpo, token_ref, metadata_json, fecha_evento
  - se usa para validar notificaciones de confirmacion/restablecimiento en entorno de pruebas de correo
- super_servidor_eventos:
  - tipo_evento, motivo, motivo_detalle, origen_arranque
  - hostname, process_id, listen_addr
  - reinicio_inesperado, previo_estado, previo_process_id, previo_inicio_en, previo_fin_en
  - correo_destino, correo_enviado, correo_error
  - metadata_json, fecha_evento
  - fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
- metrics:
  - timestamp, cpu_percent, mem_total, mem_used, mem_percent, net_recv, net_sent
  - fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones

### Tablas de asesor comercial (superadministrador)
- asesores_comerciales:
  - id, admin_email, admin_nombre, codigo (UNIQUE)
  - porcentaje_comision, meses_asociacion
  - metodo_pago_comision, entidad_financiera, tipo_cuenta, numero_cuenta
  - titular_cuenta, documento_titular, email_pagos, telefono_pagos
  - periodicidad_pago, dia_pago, pago_minimo, requiere_soporte_pago
  - estado_invitacion (`pendiente`, `aceptada`, `expirada`), invitacion_token_hash, invitacion_expira_en, invitado_por_email, aceptado_en
  - fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
  - Descripción: administradores invitados por super para operar como asesores comerciales. Al aceptar la invitación reciben un codigo comercial que puede incluirse en el checkout publico de licencias.

- asesor_comercial_comisiones:
  - id, asesor_id, asesor_codigo, asesor_email
  - empresa_id, empresa_nombre, licencia_id
  - pago_provider, pago_id, transaction_id, referencia
  - valor_pagado, porcentaje_comision, monto_comision, fecha_pago
  - asociado_desde, asociado_hasta
  - pagado, fecha_pago_comision, pagado_por
  - estado_pago_comision, metodo_pago_comision, referencia_pago_comision
  - fecha_programada_pago, soporte_pago_url
  - fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
  - Descripción: historial de ventas/renovaciones de licencia asociadas a asesores comerciales. Si una empresa pagó con codigo de asesor, las renovaciones dentro de `meses_asociacion` siguen generando comisión hasta `asociado_hasta`; vencido ese plazo ya no se muestra en `mis_clientes`.

## Relaciones adicionales
- asesores_comerciales.id -> asesor_comercial_comisiones.asesor_id
- empresas.id -> asesor_comercial_comisiones.empresa_id
- licencias.id -> asesor_comercial_comisiones.licencia_id
- administradores.contrato_version_aceptada -> super_contrato_versiones.version (referencia logica de aceptacion)
- super_contrato_versiones.usuario_creador -> administradores.email (referencia logica de publicacion)
- super_errores_sistema.empresa_id -> pcs_empresas.empresas.id (referencia logica para trazabilidad de incidencias por empresa)
- super_errores_sistema.usuario_email -> administradores.email o cuentas operativas asociadas (referencia logica de actor)


## 3) Relaciones clave
- empresas.id -> users.empresa_id
- empresas.id -> clientes.empresa_id, categorias_productos.empresa_id, productos.empresa_id, carritos_compras.empresa_id, chat_tareas*.empresa_id
- empresas.id -> empresa_inventario_configuracion.empresa_id, inventario_costos_lotes.empresa_id, inventario_conteos_ciclicos.empresa_id
- empresas.id -> reservas_hotel.empresa_id
- empresas.id -> combos_productos.empresa_id, combos_productos_detalle.empresa_id
- empresas.id -> codigos_de_descuento.empresa_id
- empresas.id -> codigos_descuento_redenciones.empresa_id
- empresas.id -> empresa_comisiones_servicio_configuracion.empresa_id, empresa_comisiones_servicio_escalas.empresa_id, empresa_comisiones_servicio_movimientos.empresa_id
- empresas.id -> empresa_finanzas_movimientos.empresa_id, empresa_finanzas_periodos.empresa_id, empresa_finanzas_configuracion.empresa_id
- empresas.id -> empresa_cierres_caja.empresa_id
- empresas.id -> empresa_facturacion_documentos.empresa_id, empresa_compras_documentos.empresa_id, facturacion_electronica_pais.empresa_id, facturacion_electronica_reintentos.empresa_id
- empresas.id -> empresa_eventos_contables.empresa_id
- empresas.id -> empresa_asientos_contables.empresa_id
- empresas.id -> empresa_auditoria_eventos.empresa_id
- empresas.id -> empresa_ai_consultas.empresa_id, empresa_ai_uso_diario.empresa_id
- empresas.id -> empresa_ai_modelo_preferido.empresa_id
- empresas.id -> empresa_gps_dispositivos.empresa_id, empresa_gps_recorridos.empresa_id
- empresas.id -> empresa_asistencia_empleados.empresa_id
- empresas.id -> empresa_asistencia_configuracion.empresa_id, empresa_asistencia_periodos_cerrados.empresa_id
- empresas.id -> empresa_nomina_configuracion.empresa_id, empresa_nomina_empleados.empresa_id, empresa_nomina_festivos.empresa_id, empresa_nomina_liquidaciones.empresa_id
- empresas.id -> empresa_vehiculos_registro.empresa_id
- empresas.id -> empresa_vehiculos_configuracion.empresa_id
- empresas.id -> empresa_cotizaciones_venta.empresa_id, empresa_pedidos_venta.empresa_id, empresa_devoluciones_venta.empresa_id
- empresas.id -> empresa_plan_cuentas.empresa_id, empresa_cuentas_por_cobrar.empresa_id, empresa_cuentas_por_pagar.empresa_id
- empresas.id -> empresa_creditos_clientes_limites.empresa_id
- empresas.id -> empresa_backups.empresa_id, empresa_backups_restauraciones.empresa_id
- empresas.id -> inventario_lotes_series.empresa_id, inventario_lotes_series_movimientos.empresa_id, empresa_devoluciones_proveedor.empresa_id, empresa_rrhh_vacaciones_licencias.empresa_id
- empresas.id -> crm_leads.empresa_id, crm_interacciones.empresa_id, crm_campanas.empresa_id, empresa_crm_metas_comerciales.empresa_id
- empresas.id -> produccion_bom.empresa_id, produccion_bom_detalle.empresa_id, produccion_ordenes.empresa_id
- empresas.id -> logistica_transportistas.empresa_id, logistica_rutas.empresa_id, logistica_envios.empresa_id
- empresas.id -> empresa_documentos_gestion.empresa_id, empresa_documentos_firmas.empresa_id
- empresas.id -> empresa_integraciones_apis.empresa_id, empresa_integraciones_bancos.empresa_id, empresa_dian_configuracion.empresa_id
- empresas.id -> empresa_ventas_estacion_metricas.empresa_id
- empresa_eventos_contables.id -> empresa_asientos_contables.evento_contable_id
- empresa_facturacion_documentos.(empresa_id,tipo_documento,documento_codigo) -> facturacion_electronica_reintentos.(empresa_id,tipo_documento,documento_codigo) [relacion logica de reconciliacion FE]
- proveedores.id -> empresa_compras_documentos.proveedor_id
- categorias_productos.id -> productos.categoria_id
- combos_productos.id -> combos_productos_detalle.combo_id
- productos.id -> combos_productos_detalle.producto_id
- productos.id -> inventario_costos_lotes.producto_id, inventario_conteos_ciclicos.producto_id
- bodegas.id -> inventario_existencias.bodega_id, inventario_movimientos.bodega_(origen|destino)_id, inventario_costos_lotes.bodega_id, inventario_conteos_ciclicos.bodega_id
- inventario_movimientos.id -> inventario_conteos_ciclicos.movimiento_id
- crm_leads.id -> crm_interacciones.lead_id
- produccion_bom.id -> produccion_bom_detalle.bom_id, produccion_ordenes.bom_id
- logistica_transportistas.id -> logistica_envios.transportista_id
- logistica_rutas.id -> logistica_envios.ruta_id
- empresa_documentos_gestion.id -> empresa_documentos_firmas.documento_gestion_id
## Actualizacion 2026-05-12 - Identidad visual empresarial

- No se agregan tablas nuevas. La configuracion avanzada ahora usa `mostrar_logo_empresa` y `mostrar_logo_sistema` para separar la visibilidad del logo empresarial y del logo del sistema en documentos imprimibles; `mostrar_logo` queda como compatibilidad general.
- Se reutiliza `empresa_configuracion_avanzada.logo_url` como URL canonica del logo empresarial.
- Los archivos cargados por el endpoint empresarial se almacenan bajo `web/uploads/empresa_logos/empresa_<id>/` y se sirven como ruta publica `/uploads/empresa_logos/empresa_<id>/<archivo>`.
- El dato queda aislado por `empresa_id` y es compartido por panel, factura, comprobantes y documentos empresariales que leen la configuracion avanzada.

- carritos_compras.id -> carrito_compra_items.carrito_id
- carritos_compras.id -> reservas_hotel.carrito_id
- carritos_compras.id -> codigos_descuento_redenciones.carrito_id
- carritos_compras.id -> empresa_comisiones_servicio_movimientos.carrito_id
- carritos_compras.id -> empresa_ventas_estacion_metricas.carrito_id
- clientes.id -> empresa_creditos.cliente_id, empresa_creditos_clientes_limites.cliente_id
- empresa_creditos.id -> empresa_creditos_cuotas.credito_id, empresa_creditos_movimientos.credito_id
- empresa_creditos_cuotas.id -> empresa_creditos_movimientos.cuota_id (opcional)
- empresa_backups.id -> empresa_backups_restauraciones.backup_id [relacion logica de restauracion]
- empresa_control_electrico_config.empresa_id -> empresas.id
- empresa_control_electrico_reles.empresa_id -> empresas.id; estacion_id referencia logica a estaciones/carritos `ESTACION_<id>`
- empresa_control_electrico_eventos.empresa_id -> empresas.id; rele_id -> empresa_control_electrico_reles.id; estacion_id referencia logica a estaciones/carritos `ESTACION_<id>`
- carrito_compra_items.id -> empresa_comisiones_servicio_movimientos.carrito_item_id
- servicios.id -> empresa_comisiones_servicio_movimientos.servicio_id
- empresa_comisiones_servicio_escalas.id -> empresa_comisiones_servicio_movimientos.escala_id
- empresa_nomina_liquidaciones.id -> empresa_comisiones_servicio_movimientos.liquidacion_nomina_id
- users.id -> empresa_propinas_movimientos.usuario_origen_id, empresa_propinas_movimientos.usuario_asignado_id
- users.id -> empresa_comisiones_servicio_movimientos.usuario_origen_id, empresa_comisiones_servicio_movimientos.usuario_lavador_id
- codigos_de_descuento.id -> codigos_descuento_redenciones.codigo_descuento_id
- clientes.id -> codigos_descuento_redenciones.cliente_id
- chat_tareas_conversaciones.id -> chat_tareas_participantes.conversacion_id, chat_tareas_mensajes.conversacion_id, chat_tareas.conversacion_id
- chat_tareas_mensajes.id -> chat_tareas_adjuntos.mensaje_id
- empresa_gps_dispositivos.id -> empresa_gps_recorridos.dispositivo_id
- tipos_de_empresas.id -> roles_de_usuario.tipo_empresa_id
- roles_de_usuario.id -> roles_de_usuario_permisos.rol_id
- roles_de_usuario.id -> roles_de_usuario_paginas_permisos.rol_id
- empresas.id -> super_correo_notificaciones_prueba.empresa_id (trazabilidad de notificaciones en modo pruebas)
- empresas.id -> super_correos_masivos_destinatarios.empresa_id (trazabilidad de destinatarios usuario_empresa en comunicados globales)
- super_correos_masivos.id -> super_correos_masivos_destinatarios.correo_masivo_id
- super_venta_digital_items.id -> super_venta_digital_ordenes.item_id

## 4) Historial resumido
- 2026-05-18: `empresa_corte_caja_configuracion` agrega banderas para que cada empresa active/desactive encabezado, datos de empresa, fecha/hora, usuario, consecutivo, columnas del detalle de ventas y metricas de productos/servicios en el reporte de turno/corte de caja.
- 2026-05-13: el aseguramiento ligero de `carritos_compras`, `carrito_compra_items` y `empresa_ventas_estacion_metricas` valida y completa ahora todas las columnas usadas por el listado operativo antes de marcar el esquema como listo, con cache por base/esquema PostgreSQL. Esto evita 500 en `/api/empresa/carritos_compra` cuando una empresa conserva migraciones rezagadas; no crea tablas nuevas ni cambia relaciones.
- 2026-05-13: `licencias` incorpora `max_cajas_simultaneas` para limitar cajas abiertas simultaneas por empresa segun licencia activa. El valor por defecto es 2 cajas; las licencias de 4000 documentos quedan en 4 cajas. `carritos_compras`, `empresa_ventas_estacion_metricas` y `empresa_finanzas_movimientos` enlazan operaciones con `cierre_caja_id`, `caja_codigo`, `caja_turno` y `caja_sucursal_id` para cierres separados por caja.
- 2026-05-13: se agregan `super_correos_masivos` y `super_correos_masivos_destinatarios` en `pcs_superadministrador` para auditar comunicados globales enviados por super administrador. La campana registra codigo, categoria, alcance, asunto, totales, estado, modo prueba, usuario creador y fechas; cada destinatario guarda email, tipo (`administrador` o `usuario_empresa`), empresa asociada cuando aplique, rol, resultado y error resumido.
- 2026-05-13: se agrega `licencia_vencimiento_notificaciones` en `pcs_superadministrador` para registrar avisos de vencimiento enviados/capturados por licencia base o adicional, empresa, correo administrador, fecha de vencimiento y umbral de dias. La configuracion global vive en `configuraciones` con claves `licencias.vencimiento_alertas.*`.
- 2026-05-04: se agregan `empresa_control_electrico_config`, `empresa_control_electrico_reles` y `empresa_control_electrico_eventos` para controlar reles GPIO en Raspberry Pi por estacion. La configuracion guarda conexion HTTP por empresa; los reles asignan estacion + `salida_codigo` + `tipo_carga` a GPIO y estado runtime; los eventos auditan comandos `on/off`, respuesta de la Raspberry, actor y origen.
- 2026-04-08: se agrega `super_servidor_eventos` en `pcs_superadministrador` para auditoria de inicio/reinicio del servidor (incluye estado previo, motivo, resultado de envio de correo y metadata operativa); ademas se incorpora clave de configuracion `gmail.restart_alert_to` para correo destino de alertas.
- 2026-04-08: se amplía `licencias` en `pcs_superadministrador` con `modulos_habilitados` y `super_rol_habilitado` para gobernar permisos efectivos por empresa desde la licencia activa, junto con columnas de trazabilidad (`fecha_actualizacion`, `usuario_creador`, `estado`, `observaciones`).
- 2026-04-08: se agregan `super_venta_digital_configuracion`, `super_venta_digital_items` y `super_venta_digital_ordenes` en `pcs_superadministrador` para venta de licencias/software administrada por super, con pago Wompi y entrega por correo posterior a aprobacion.
- 2026-04-08: se agregan `roles_de_usuario_permisos` y `roles_de_usuario_paginas_permisos` en `pcs_superadministrador` para configuracion dinamica de permisos por rol (modulo/accion y pagina), con indices unicos por rol para garantizar consistencia de matriz.
- 2026-04-07: se agregan `empresa_backups` y `empresa_backups_restauraciones` para el modulo 36 de backups empresariales, incluyendo snapshot JSON por `empresa_id`, trazabilidad de hash de contenido y bitacora de restauraciones.
- 2026-05-12: se agregan exportes locales de backups y configuracion que construyen el snapshot en memoria y lo descargan al dispositivo del usuario sin insertar registros en `empresa_backups` ni escribir copias en disco del VPS.
- 2026-04-07: se agrega `empresa_creditos_clientes_limites` para gobernar limites por cliente (`limite_saldo_total`, `max_creditos_activos`, `requiere_aprobacion_exceso`) y reforzar aislamiento por `empresa_id` en validaciones de alta/edicion de creditos.
- 2026-04-07: se agregan `empresa_creditos`, `empresa_creditos_cuotas` y `empresa_creditos_movimientos` para base del modulo 35 (creditos), con trazabilidad de cupo/saldo, amortizacion por cuotas, abonos y movimientos por `empresa_id`/`credito_id`/`cliente_id`.
- 2026-04-07: se agregan objetos de busqueda full-text para auditoria empresarial (`empresa_auditoria_eventos_fts` + triggers `ai/au/ad`) para soportar `search` con mejor rendimiento y fallback seguro.
- 2026-04-07: se agrega `empresa_ventas_estacion_metricas` para cierre del modulo 27 de ventas simples por estacion, con trazabilidad de `venta_pagada`, `cierre_parcial_anulado` y `sesion_recuperada`, incluyendo `duracion_segundos` y montos operativos por `empresa_id`/`carrito_id`/`estacion_id`.
- 2026-04-07: se amplía `empresa_rrhh_vacaciones_licencias` para cierre del modulo 22 de RRHH con aprobacion jerarquica (`nivel_aprobacion_*`, `aprobadores_json`, `historial_aprobaciones_json`, `fecha_aprobacion_final`), calculo de saldo/acumulado (`periodo_acumulado_*`, `saldo_dias_*`, `saldo_snapshot_json`) y enlace de novedades a nomina (`empleado_nomina_id`, `nomina_liquidacion_id`, `nomina_periodo_*`, `nomina_vinculada_*`).
- 2026-04-06: se agrega `facturacion_electronica_reintentos` para cola de integracion fiscal FE por documento (`estado_envio`, `intentos`, `max_intentos`, `proximo_intento`, `contingencia_activa`, `referencia_externa`) y soporte de reconciliacion de estados.
- 2026-04-06: se amplía `empresa_compras_documentos` para cierre del modulo 16 de compras con aprobacion multinivel (`requiere_aprobacion`, `niveles_aprobacion_requeridos`, `nivel_aprobacion_actual`, `aprobadores_json`), recepcion parcial por item (`recepcion_detalle_json`, `recepcion_resumen_json`) y validacion documental proveedor-factura-entrada (`validacion_documental_estado`, `proveedor_documento_ref`, `factura_documento_ref`, `entrada_documento_ref`).
- 2026-04-06: se amplía comisiones por servicio con tabla de escalas/topes (`empresa_comisiones_servicio_escalas`), flujo de ajustes manuales con aprobacion (`ajuste_estado`, `aprobado_por`, `aprobado_en`) y enlace a nomina (`liquidacion_nomina_id`, `periodo_liquidacion_*`, `liquidado_*`); `empresa_nomina_liquidaciones` incorpora `comisiones_servicio_total`, `comisiones_servicio_movimientos` y `comisiones_servicio_ajustes`.
- 2026-04-06: se amplía el modulo de propinas con reglas fiscales por empresa (`pais_fiscal`, `regimen_fiscal`, `tratamiento_fiscal`, `porcentaje_impuesto_propina`), ajustes manuales auditados (`origen_movimiento`, `ajuste_manual`, `referencia_ajuste`) y conciliacion por `cierre_caja_id`; `empresa_cierres_caja` incorpora resumen persistido de propinas conciliadas (`propinas_*`).
- 2026-05-13: propinas y comisiones por servicio agregan vinculo duro a usuarios creados (`users.id`) mediante `usuario_origen_id`, `usuario_asignado_id` y `usuario_lavador_id`, manteniendo las etiquetas texto para compatibilidad historica.
- 2026-04-06: se amplía `codigos_de_descuento` con reglas avanzadas por contexto (segmento/canal/horario/dias) y controles antifraude por cliente (`max_usos_por_cliente`, `ventana_horas_fraude`); se agrega `codigos_descuento_redenciones` para trazabilidad de estados `aplicada/revertida/anulada` por carrito/cliente.
- 2026-04-06: se agregan `empresa_inventario_configuracion`, `inventario_costos_lotes` e `inventario_conteos_ciclicos` para cierre del modulo 11 de inventario (politica promedio/peps por empresa, trazabilidad por lotes de costo y conteo ciclico con ajuste auditado).
- 2026-04-06: se fortalece `reservas_hotel` con politica automatica avanzada (expiracion + no_show) y reconversion operativa a carrito; el estado de reserva extiende valores operativos con `en_curso` y `no_show`.
- 2026-04-06: se agrega `empresa_vehiculos_configuracion` para parametrizar validacion de placa/patente por pais y regex por `empresa_id`, junto con regla de duplicidad activa; se incorpora reporte operativo `operativo_vehiculos_permanencia` con exportacion PDF/XLS/CSV/JSON/TXT.
- 2026-04-06: se agregan `empresa_asistencia_configuracion` y `empresa_asistencia_periodos_cerrados` para parametrizar tolerancias/turnos y bloquear ediciones por cierre de periodo en asistencia; se publica reporte operativo `operativo_asistencia_nomina_auditoria` para auditoria de nomina.
- 2026-04-06: se agrega `super_correo_notificaciones_prueba` en `pcs_superadministrador` para captura de confirmacion/restablecimiento de usuarios de empresa en entorno de pruebas de correo, junto con politicas configurables `usuarios.password_*` y rotacion opcional de contraseña.
- 2026-04-06: se retira la operacion activa de Mercado Pago en backend y se deja Wompi como pasarela unica; el registro operativo de pagos se concentra en `pagos_wompi`.
- 2026-04-06: se agregan tablas ERP extendidas por `empresa_id` para ventas avanzadas (cotizaciones/pedidos/devoluciones), contabilidad (plan de cuentas y cartera CxC/CxP), inventario por lotes/series, RRHH (vacaciones/licencias), CRM, produccion (BOM y ordenes), logistica, gestion documental, integraciones externas y configuracion DIAN Colombia.
- 2026-04-05: se agrega `reservas_hotel` para gestionar reservas por estacion con control de disponibilidad por rango, expiracion de pendientes y confirmacion de pago.
- 2026-04-05: se agrega `empresa_vehiculos_registro` para controlar ingreso y salida de vehiculos por empresa con patente, conductor, propietario y motivo operativo.
- 2026-04-05: se agrega `codigos_de_descuento` por empresa para promociones con vigencia, usos y validacion de pago en carrito.
- 2026-04-05: se amplía `carritos_compras` con `metodo_pago` y `referencia_pago` para trazabilidad del cierre de venta por estacion.
- 2026-04-17: `empresa_configuracion_avanzada` agrega `modo_documento_venta` para definir por empresa si la venta pagada genera `factura_electronica` o `comprobante_pago`, reutilizando `empresa_facturacion_documentos` como repositorio canonico de ambos documentos.
- 2026-04-05: se agregan `combos_productos` y `combos_productos_detalle` para venta compuesta con precio unico y receta de ingredientes por empresa.
- 2026-04-04: se amplía `proveedores` con campos comerciales (`catalogo_referencia`, `precio_base_referencial`, `descuento_porcentaje`, `plazo_pago_dias`, `condicion_entrega`) para gestionar catálogo, precios y condiciones por empresa.
- 2026-04-04: se agrega `empresa_auditoria_eventos` para trazabilidad de acciones criticas por `empresa_id`, modulo/accion/recurso, resultado HTTP y metadatos (`request_id`, IP, user-agent), con retencion configurable y purga.
- 2026-04-04: se agrega `empresa_asientos_contables` como persistencia canonica de asientos por evento procesado, con idempotencia por `hash_idempotencia` y referencia a `evento_contable_id`.
- 2026-04-04: se amplía `empresa_eventos_contables` con metadatos de procesamiento (`intentos_procesamiento`, `fecha_ultimo_intento`, `error_procesamiento`, `asiento_contable_id`) para trazabilidad de lotes y reintentos.
- 2026-04-04: se agrega `empresa_cierres_caja` para soportar apertura/arqueo/cierre/aprobacion de caja por sucursal y turno, con diferencia e incidencia de arqueo.
- 2026-04-04: se agregan `empresa_facturacion_documentos` y `empresa_compras_documentos` para persistencia canonica del ciclo documental y referencia estable de `entidad_id` en eventos contables.
- 2026-04-04: se agrega `empresa_eventos_contables` para contrato de eventos contables por modulo (`ventas`, `facturacion`, `compras`, `finanzas`) y trazabilidad de integracion contable.
- 2026-04-04: se amplia contrato operativo de `empresa_eventos_contables` con emision activa en `facturacion` (configuracion), `compras` (proveedores) y `finanzas` (movimientos/periodos).
- 2026-04-04: se activa emision transaccional en endpoints existentes para `facturacion` (`factura_emitida`, `factura_anulada`, `nota_credito_emitida`) y `compras` (`orden_compra_emitida`, `compra_recepcionada`, `compra_contabilizada`).
- 2026-04-04: se agrega `empresa_ai_modelo_preferido` para persistir el `model_id` preferido por `empresa_id + admin_email` (cuenta Google autenticada).
- 2026-04-04: se agregan `empresa_ai_consultas` y `empresa_ai_uso_diario` para el modulo `chat_con_inteligencia_artificial`, con auditoria y limites diarios por empresa/proveedor/modelo.
- 2026-04-04: se amplía finanzas con `empresa_finanzas_periodos`, control de cierre/reapertura de periodos, retenciones (`fuente/ica/iva`) y `total_neto` en `empresa_finanzas_movimientos`.
- 2026-04-04: se amplía `empresa_finanzas_configuracion` con cuentas de retenciones por cobrar y por pagar para asiento contable.
- 2026-04-04: se amplía `empresa_finanzas_configuracion` con parametrización contable externa por empresa (destino ERP, cuentas base y mapeo por categoría) para exportación JSON contable avanzada.
- 2026-04-04: se agregan `empresa_finanzas_movimientos` y `empresa_finanzas_configuracion` para el módulo financiero por empresa (ingresos/egresos con comprobantes e impresión).
- 2026-04-02: se agrega `categorias_productos`, se incorpora `productos.categoria_id` y se documentan relaciones del catálogo de categorías por empresa.
- 2026-04-02: se agregan tablas del modulo chat_y_tareas en pcs_empresas y se actualiza este documento.
- 2026-04-02: se agregan `empresa_gps_dispositivos` y `empresa_gps_recorridos` para tracking de ubicacion GPS por empresa, con registro periodico de recorridos.
### Tabla: super_juegos_records (pcs_superadministrador)
Almacena los top scores globales de todos los juegos publicados en `/Juegos/*` y el emulador, para todas las empresas y el público. El frontend envía records mediante `/api/public/juegos/records` con `juego`, `nombre_jugador`, `empresa_id`, `puntaje` y `nivel`.
- **Columnas**: `id` (INTEGER/SERIAL), `juego` (TEXT), `nombre_jugador` (TEXT), `empresa_id` (TEXT, DEFAULT 'Publico'), `puntaje` (INTEGER), `nivel` (INTEGER), `fecha_creacion` (TEXT/TIMESTAMP), `fecha_actualizacion` (TEXT/TIMESTAMP), `usuario_creador` (TEXT), `estado` (TEXT), `observaciones` (TEXT).
- **Ãšnico**: `id` autoincremental.
- **Índice**: `idx_super_juegos_records_top` en (`juego`, `puntaje` DESC, `fecha_creacion` ASC).
