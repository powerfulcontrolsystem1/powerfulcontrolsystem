# Estructura de Base de Datos

Version: 2026-04-04.3
Ultima actualizacion: 2026-04-04

Este documento consolida la estructura activa de SQLite para el proyecto.
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
- proveedores:
  - empresa_id, codigo, nombre, documento, contacto, telefono, email, direccion
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
  - devolucion_total, total_pagado
- carrito_compra_items:
  - empresa_id, carrito_id, tipo_item, referencia_id, codigo_item, descripcion
  - unidad_medida, cantidad, precio_unitario
  - descuento_porcentaje, impuesto_porcentaje, impuesto_codigo
  - base_gravable, valor_descuento, valor_impuesto, subtotal_linea, total_linea

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
- empresas.id -> empresa_finanzas_movimientos.empresa_id, empresa_finanzas_periodos.empresa_id, empresa_finanzas_configuracion.empresa_id
- empresas.id -> empresa_ai_consultas.empresa_id, empresa_ai_uso_diario.empresa_id
- empresas.id -> empresa_ai_modelo_preferido.empresa_id
- empresas.id -> empresa_gps_dispositivos.empresa_id, empresa_gps_recorridos.empresa_id
- categorias_productos.id -> productos.categoria_id
- carritos_compras.id -> carrito_compra_items.carrito_id
- chat_tareas_conversaciones.id -> chat_tareas_participantes.conversacion_id, chat_tareas_mensajes.conversacion_id, chat_tareas.conversacion_id
- chat_tareas_mensajes.id -> chat_tareas_adjuntos.mensaje_id
- empresa_gps_dispositivos.id -> empresa_gps_recorridos.dispositivo_id
- tipos_de_empresas.id -> roles_de_usuario.tipo_empresa_id / tipos_de_usuario.tipo_empresa_id
- roles_de_usuario.id -> tipos_de_usuario.rol_id

## 4) Historial resumido
- 2026-04-04: se agrega `empresa_ai_modelo_preferido` para persistir el `model_id` preferido por `empresa_id + admin_email` (cuenta Google autenticada).
- 2026-04-04: se agregan `empresa_ai_consultas` y `empresa_ai_uso_diario` para el modulo `chat_con_inteligencia_artificial`, con auditoria y limites diarios por empresa/proveedor/modelo.
- 2026-04-04: se amplía finanzas con `empresa_finanzas_periodos`, control de cierre/reapertura de periodos, retenciones (`fuente/ica/iva`) y `total_neto` en `empresa_finanzas_movimientos`.
- 2026-04-04: se amplía `empresa_finanzas_configuracion` con cuentas de retenciones por cobrar y por pagar para asiento contable.
- 2026-04-04: se amplía `empresa_finanzas_configuracion` con parametrización contable externa por empresa (destino ERP, cuentas base y mapeo por categoría) para exportación JSON contable avanzada.
- 2026-04-04: se agregan `empresa_finanzas_movimientos` y `empresa_finanzas_configuracion` para el módulo financiero por empresa (ingresos/egresos con comprobantes e impresión).
- 2026-04-02: se agrega `categorias_productos`, se incorpora `productos.categoria_id` y se documentan relaciones del catálogo de categorías por empresa.
- 2026-04-02: se agregan tablas del modulo chat_y_tareas en empresas.db y se actualiza este documento.
- 2026-04-02: se agregan `empresa_gps_dispositivos` y `empresa_gps_recorridos` para tracking de ubicacion GPS por empresa, con registro periodico de recorridos.
