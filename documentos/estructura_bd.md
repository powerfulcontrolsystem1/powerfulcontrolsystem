# Estructura de Base de Datos

Version: 2026-04-05.21
Ultima actualizacion: 2026-04-05

Este documento consolida la estructura activa de SQLite para el proyecto.
Nota de gobernanza documental:
- `documentos/estructura_bd.md` es la fuente canonica del esquema fisico.
- `estructura_bd.md` (raiz) se mantiene como copia de compatibilidad y debe mantenerse sincronizada.
- `documentos/descripcion_de_las_bases_De_datos` es documento complementario funcional (sin duplicar detalle tabla-por-tabla).
Todas las tablas operativas usan como base los campos estandar:
- id INTEGER PRIMARY KEY AUTOINCREMENT
- fecha_creacion TEXT DEFAULT (datetime('now','localtime'))
- fecha_actualizacion TEXT DEFAULT (datetime('now','localtime'))
- usuario_creador TEXT
- estado TEXT DEFAULT 'activo'
- observaciones TEXT

## 1) Base: empresas.db

### Tablas de control y core
- schema_migrations:
  - id, scope, version, description, applied_at
- users:
  - email, name, role, empresa_id, documento_identidad, rol_usuario_id
  - email_confirmado, email_confirm_token, email_confirm_expira, email_confirmado_en
  - password_hash, password_salt, password_set, password_actualizada_en
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
- combos_productos_detalle:
  - empresa_id, combo_id, producto_id
  - cantidad, unidad_medida
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

### Tabla de reservas por estacion/habitacion
- reservas_hotel:
  - empresa_id, carrito_id, estacion_id, codigo_reserva
  - cliente_nombre, cliente_documento, cliente_email, cliente_telefono
  - cantidad_huespedes, fecha_entrada, fecha_salida
  - monto_total, moneda
  - estado_reserva (`pendiente_pago`, `confirmada`, `cancelada`, `expirada`)
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
  - monto_minimo_compra, fecha_vencimiento
  - usos_maximos, usos_actuales

### Tablas de propinas por empresa
- empresa_propinas_configuracion:
  - empresa_id (UNIQUE)
  - habilitar_propina, porcentaje_propina
  - modo_distribucion (`por_usuario` o `universal`)
  - aplicar_automaticamente
- empresa_propinas_movimientos:
  - empresa_id, carrito_id, venta_referencia
  - usuario_origen, usuario_asignado
  - modo_distribucion, moneda
  - base_cobro, porcentaje_propina, monto_propina
  - fecha_movimiento

### Tablas de comisiones por servicio por empresa
- empresa_comisiones_servicio_configuracion:
  - empresa_id (UNIQUE)
  - habilitar_comisiones, porcentaje_comision
  - filtro_servicio (ej. `lavado`)
  - aplicar_automaticamente
- empresa_comisiones_servicio_movimientos:
  - empresa_id, carrito_id, carrito_item_id
  - servicio_id, servicio_codigo, servicio_nombre, servicio_categoria
  - usuario_origen, usuario_lavador
  - venta_referencia, moneda
  - base_servicio, porcentaje_comision, monto_comision
  - fecha_movimiento

### Tablas de configuracion operativa de cobro por empresa y rol
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
  - cerrado_por, aprobado_por, aprobado_en
  - UNIQUE(empresa_id, sucursal_id, caja_codigo, fecha_operacion, turno)
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

### Tablas de documentos transaccionales canonicos
- empresa_facturacion_documentos:
  - empresa_id, tipo_documento, documento_codigo
  - estado_documento, estado_anterior, evento_ultimo
  - periodo_contable, monto_total, moneda
  - numero_legal, codigo_validacion, pais_codigo, ambiente_fe
  - fecha_documento, entidad_relacionada_id
  - UNIQUE(empresa_id, tipo_documento, documento_codigo)
- empresa_compras_documentos:
  - empresa_id, proveedor_id, tipo_documento, documento_codigo
  - estado_documento, estado_anterior, evento_ultimo
  - periodo_contable, monto_total, moneda
  - fecha_documento, entidad_relacionada_id
  - UNIQUE(empresa_id, tipo_documento, documento_codigo)

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
  - tipo_documento_emisor, nit, digito_verificacion
  - razon_social, nombre_comercial, regimen_fiscal, responsabilidad_tributaria
  - email_facturacion, telefono_facturacion, direccion_fiscal, departamento, municipio
  - pais_codigo, codigo_postal
  - ambiente_fe, tipo_operacion, prefijo_factura
  - resolucion_numero, resolucion_fecha_desde, resolucion_fecha_hasta
  - consecutivo_desde, consecutivo_hasta, proximo_consecutivo
  - formato_impresion, imprimir_copia_factura, mostrar_logo, logo_url
  - pie_factura, notas_legales
  - color_carrito_activo, color_carrito_inactivo

### Tabla de facturacion electronica por pais
- facturacion_electronica_pais:
  - empresa_id, pais_codigo, pais_nombre, moneda_codigo
  - proveedor, ambiente, tipo_documento_emisor, identificador_fiscal
  - razon_social, email_facturacion, telefono_facturacion, direccion_fiscal
  - prefijo_factura, resolucion_numero, api_base_url, campos_pais_json
  - UNIQUE(empresa_id, pais_codigo)

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
  - total_recargos_horas_extras, auxilio_transporte, bonificacion, devengado_total
  - ingreso_base_cotizacion, deduccion_salud, deduccion_pension, deduccion_fondo_solidaridad
  - deduccion_fija, otras_deducciones, deduccion_total, neto_pagar
  - origen_calculo, resumen_json, fecha_generacion

### Tabla de registro vehicular por empresa
- empresa_vehiculos_registro:
  - empresa_id
  - patente, tipo_vehiculo
  - conductor_nombre, conductor_documento, conductor_contacto
  - propietario_nombre, motivo_ingreso
  - fecha_ingreso, fecha_salida
  - estado_registro (`en_empresa` o `retirado`)

## 2) Base: superadministrador.db

### Tablas de control y administracion
- schema_migrations:
  - id, scope, version, description, applied_at
- administradores:
  - email, name, role, photo
- sesiones:
  - admin_email, token, ip, user_agent, fecha_inicio, fecha_fin, activo
- configuraciones:
  - config_key (PK), value, encrypted

### Tablas de catalogos globales
- tipos_de_empresas:
  - nombre
- roles_de_usuario:
  - tipo_empresa_id, nombre, descripcion
- tipos_de_usuario:
  - tipo_empresa_id, rol_id, nombre, descripcion
- tipos_de_licencia:
  - nombre
- licencias:
  - empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias
  - fecha_inicio, fecha_fin, activo

### Tablas de pagos y metricas
- pagos_mercadopago:
  - licencia_id, empresa_id, preference_id, payment_id, status, raw_payload
- pagos_wompi:
  - licencia_id, empresa_id, transaction_id, reference, status, raw_payload
- metrics:
  - timestamp, cpu_percent, mem_total, mem_used, mem_percent, net_recv, net_sent
  - fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones

## 3) Relaciones clave
- empresas.id -> users.empresa_id
- empresas.id -> clientes.empresa_id, categorias_productos.empresa_id, productos.empresa_id, carritos_compras.empresa_id, chat_tareas*.empresa_id
- empresas.id -> reservas_hotel.empresa_id
- empresas.id -> combos_productos.empresa_id, combos_productos_detalle.empresa_id
- empresas.id -> codigos_de_descuento.empresa_id
- empresas.id -> empresa_finanzas_movimientos.empresa_id, empresa_finanzas_periodos.empresa_id, empresa_finanzas_configuracion.empresa_id
- empresas.id -> empresa_cierres_caja.empresa_id
- empresas.id -> empresa_facturacion_documentos.empresa_id, empresa_compras_documentos.empresa_id
- empresas.id -> empresa_eventos_contables.empresa_id
- empresas.id -> empresa_asientos_contables.empresa_id
- empresas.id -> empresa_auditoria_eventos.empresa_id
- empresas.id -> empresa_ai_consultas.empresa_id, empresa_ai_uso_diario.empresa_id
- empresas.id -> empresa_ai_modelo_preferido.empresa_id
- empresas.id -> empresa_gps_dispositivos.empresa_id, empresa_gps_recorridos.empresa_id
- empresas.id -> empresa_asistencia_empleados.empresa_id
- empresas.id -> empresa_nomina_configuracion.empresa_id, empresa_nomina_empleados.empresa_id, empresa_nomina_festivos.empresa_id, empresa_nomina_liquidaciones.empresa_id
- empresas.id -> empresa_vehiculos_registro.empresa_id
- empresa_eventos_contables.id -> empresa_asientos_contables.evento_contable_id
- proveedores.id -> empresa_compras_documentos.proveedor_id
- categorias_productos.id -> productos.categoria_id
- combos_productos.id -> combos_productos_detalle.combo_id
- productos.id -> combos_productos_detalle.producto_id
- carritos_compras.id -> carrito_compra_items.carrito_id
- carritos_compras.id -> reservas_hotel.carrito_id
- chat_tareas_conversaciones.id -> chat_tareas_participantes.conversacion_id, chat_tareas_mensajes.conversacion_id, chat_tareas.conversacion_id
- chat_tareas_mensajes.id -> chat_tareas_adjuntos.mensaje_id
- empresa_gps_dispositivos.id -> empresa_gps_recorridos.dispositivo_id
- tipos_de_empresas.id -> roles_de_usuario.tipo_empresa_id / tipos_de_usuario.tipo_empresa_id
- roles_de_usuario.id -> tipos_de_usuario.rol_id

## 4) Historial resumido
- 2026-04-05: se agrega `reservas_hotel` para gestionar reservas por estacion/habitacion con control de disponibilidad por rango, expiracion de pendientes y confirmacion de pago.
- 2026-04-05: se agrega `empresa_vehiculos_registro` para controlar ingreso y salida de vehiculos por empresa con patente, conductor, propietario y motivo operativo.
- 2026-04-05: se agrega `codigos_de_descuento` por empresa para promociones con vigencia, usos y validacion de pago en carrito.
- 2026-04-05: se amplía `carritos_compras` con `metodo_pago` y `referencia_pago` para trazabilidad del cierre de venta por estacion.
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
- 2026-04-02: se agregan tablas del modulo chat_y_tareas en empresas.db y se actualiza este documento.
- 2026-04-02: se agregan `empresa_gps_dispositivos` y `empresa_gps_recorridos` para tracking de ubicacion GPS por empresa, con registro periodico de recorridos.
