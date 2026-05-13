# Indice documental del proyecto

Fecha: 2026-05-13
Estado: vigente, actualizado con operacion conectada obligatoria, cajas simultaneas por licencia, tickets de ayuda propios, correos masivos, alertas de vencimiento de licencias, mantenimiento programado, login global de usuarios de empresa, OnlyOffice local editable, backups locales, retiro de Nextcloud, VPS portable Docker, seguridad SSH 49222, 30 verticales canonicos, roles y permisos finos, modulos empresariales Colombia, suite contable Colombia avanzada, reportes, auditoria y QA transversal.

Este archivo organiza la lectura tecnica y funcional del repositorio para desarrollo, soporte y trabajo asistido por Copilot.

## Orden de lectura recomendado

1. `documentos/descripcion_del_proyecto`
2. `documentos/estructura_bd.md`
3. `documentos/diagramas/diagrama_entidad_relacion.md`
4. `documentos/diagramas/estructura_del_codigo.md`
5. `documentos/matriz_roles_permisos_pos_multiempresa.md`
6. `documentos/reporte_estado_modulos_2026-05-05.md`
7. `documentos/gobernanza_tecnica/README.md`
8. `documentos/historial_de_cambios`
9. `CHANGELOG.md`

## Estado documental reciente
- 2026-05-13: se agrega `documentos/estado_documentacion_2026-05-13.md` como mapa rapido del estado vigente: reglas de producto, autenticacion, licencias/cajas, operacion conectada, soporte, comunicaciones, portal publico, VPS y validacion recomendada.
- 2026-05-13: la ayuda principal `web/ayuda/ayuda.html` se actualiza con operacion conectada, cajas simultaneas, login por invitacion, soporte por tickets, documentos locales, backups, mantenimiento, correos globales y criterios de validacion.
- 2026-05-13: licencias incorporan `max_cajas_simultaneas`; el default es 2 cajas por empresa y 4 cajas para licencias de 4000 documentos.
- 2026-05-13: se retira la operacion/facturacion offline para clientes; ventas, cobros, documentos y facturacion requieren servidor activo.
- 2026-05-13: `login_usuario.html` queda como acceso global de usuarios operativos y `login.html` mantiene acceso administrativo con presentacion visual profesional.
- 2026-05-10: implementados y documentados los 20 verticales empresariales 2026 sobre el motor comun `empresa_modulos_colombia_*`, con catalogo backend/frontend, licencias por tipo, preconfiguracion, selector de empresas, portada publica, checkout, permisos, ayuda administrativa e IA. Ver `documentos/plan_20_modulos_verticales_2026-05-10.md` y `documentos/arquitectura_modulos_universales.md`.
- 2026-05-10: actualizado el sistema documental de roles/permisos con modulos finos (`crm_unificado`, `reservas_hotel`, `chat_tareas`, `horarios_trabajadores`, `asistencia_empleados`, `vehiculos_registro`, `hoja_vida_operativa`, `ubicacion_gps`, `nomina_sueldos`, `reportes`, `auditoria`, `backups`, `documentos_onlyoffice`, `nextcloud`), wrappers API especificos y compatibilidad de licencias amplias. Ver `documentos/reporte_roles_ayuda_super_2026-05-10.md`.
- 2026-05-10: la ayuda administrativa completa `/ayuda/ayuda.html` se mantiene exclusiva para `super_administrador` y se abre desde el boton `Ayuda super administrador` en `web/super_administrador.html`; el rol `control_super_administrador` no ve ese acceso.
- 2026-05-06: implementados `bancos_pagos`, `gestion_documental`, `cumplimiento_kyc`, `contratos_obligaciones` y `calidad_procesos` sobre un nucleo compartido por `empresa_id`, con APIs privadas, pantallas administrativas, permisos/licencias, datos demo, exportacion CSV y documento `documentos/modulos_empresariales_colombia.md`.
- 2026-05-06: agregado modulo `Logistica avanzada / WMS` con API `/api/empresa/logistica_wms`, ubicaciones internas, ordenes WMS, picking, packing, despachos, rutas, avance por item, bitacora, permiso/licencia `logistica_wms`, pantalla `web/administrar_empresa/logistica_wms.html` y documento `documentos/logistica_wms.md`.
- 2026-05-06: agregado modulo `Declaraciones Tributarias y Motor de Impuestos Colombia` con API `/api/empresa/declaraciones_tributarias`, preliquidacion de IVA, retenciones, ICA, consumo, renta y regimen simple, calendario tributario editable, saldos a pagar/favor, movimientos de conciliacion, permiso/licencia `declaraciones_tributarias`, pantalla `web/administrar_empresa/declaraciones_tributarias.html` y documento `documentos/declaraciones_tributarias.md`.
- 2026-05-06: agregado modulo `Portal de Terceros y Certificados Tributarios` con API `/api/empresa/portal_terceros_certificados`, API publica `/api/public/certificados_tributarios`, maestro de terceros, certificados de retencion/ingresos, enlace publico seguro, bitacora de descargas, permiso/licencia `portal_terceros_certificados`, pantalla `web/administrar_empresa/portal_terceros_certificados.html`, pagina `web/visualizar_certificado_tributario_publico.html` y documento `documentos/portal_terceros_certificados.md`.
- 2026-05-06: agregado modulo formal `Activos Fijos e Intangibles NIIF/Fiscal` con API `/api/empresa/activos_fijos_niif_fiscal`, libro maestro, depreciacion por periodo, vida util NIIF y fiscal, deterioro, valor fiscal, diferencia NIIF/fiscal, eventos, seguros, mantenimientos, permiso/licencia `activos_fijos_niif_fiscal`, pantalla `web/administrar_empresa/activos_fijos_niif_fiscal.html` y documento `documentos/activos_fijos_niif_fiscal.md`.
- 2026-05-06: agregado modulo `Propiedad Horizontal / Administracion de Copropiedades` con API `/api/empresa/propiedad_horizontal`, unidades, propietarios/residentes, cargos, recaudos, PQR, asambleas, dashboard, datos demo, permiso/licencia `propiedad_horizontal` y pantalla `web/administrar_empresa/propiedad_horizontal.html`.
- 2026-05-06: agregada promocion configurable de licencias por codigo de asesor desde `web/super/asesor_comercial.html`; usa `licencias.asesor_promo.enabled` y `licencias.asesor_promo.percent`, aplica descuento adicional en checkout y conserva comisiones.
- 2026-05-06: agregado modulo `Cierre y bloqueo fiscal` con API `/api/empresa/cierre_fiscal`, politicas por modulo, periodos fiscales, bloqueos por fecha, excepciones aprobadas, simulador, bitacora, permiso/licencia `cierre_fiscal`, pantalla `web/administrar_empresa/cierre_fiscal.html`, sincronizacion con cierre/reapertura de Contabilidad Colombia y documento `documentos/cierre_fiscal.md`.
- 2026-05-06: agregado modulo formal `Centros de costo y rentabilidad` con API `/api/empresa/centros_costo`, maestro, reglas de imputacion, presupuesto por periodo, dashboard comparativo, movimientos integrados desde contabilidad/tesoreria/compras/OCR/AIU, permiso/licencia `centros_costo`, pantalla `web/administrar_empresa/centros_costo.html` y documento `documentos/centros_costo.md`.
- 2026-05-06: actualizado `web/index.html` para describir los modulos recientes en la portada publica: Cobranza, Portal contador, Captura IA/OCR de compras y gastos, AIU construccion, Parqueaderos con ticket QR y Apartamentos turisticos, tanto en la seccion fija de modulos como en tarjetas fallback.
- 2026-05-06: agregado `documentos/reporte_qa_modulos_2026-05-06.md` con revision transversal autenticada de Motel Calipso, paginas/API 200, auditoria estatica de enlaces/botones, optimizacion de dashboards de `cobranza`, `portal_contador` y `soportes_compras_ia`, y nota de limitacion del navegador embebido por bloqueo local de Node.
- 2026-05-06: agregado modulo `Captura inteligente de compras y gastos` con API `/api/empresa/soportes_compras_ia`, foto/PDF/XML, extraccion OCR/IA con `openai:gpt-5.5`, aprobacion, contabilizacion CxP, permisos/licencia `soportes_compras_ia`, pantalla `web/administrar_empresa/soportes_compras_ia.html` y documento `documentos/soportes_compras_ia.md`.
- 2026-05-06: ampliado `CRM comercial` con capa `CRM y ventas avanzadas`: metas comerciales, dashboard de pipeline/forecast, scoring de leads, agenda y conversion a cotizacion, API `/api/empresa/crm_avanzado`, pantalla `web/administrar_empresa/crm_ventas_avanzadas.html` y documento `documentos/crm_ventas_avanzadas.md`.
- 2026-05-06: ampliado `Inventario` con inventario avanzado sin duplicar modulo: lotes, seriales, reservas, vencimientos y valorizacion por bodega, API `/api/empresa/inventario_avanzado`, pantalla `web/administrar_empresa/inventario_avanzado.html` y documento `documentos/inventario_avanzado.md`.
- 2026-05-06: ampliado `Compras` con compras avanzadas sin duplicar modulo: requisiciones, cotizaciones, aprobaciones, recepcion parcial/total, API `/api/empresa/compras_avanzadas`, pantalla `web/administrar_empresa/compras_avanzadas.html` y documento `documentos/compras_avanzadas.md`.
- 2026-05-06: agregado modulo `Importaciones y costeo` con API `/api/empresa/importaciones_costeo`, embarques, items, costos de nacionalizacion, distribucion por base y costo aterrizado, permisos/licencia `importaciones_costeo`, pantalla administrativa y documento `documentos/importaciones_costeo.md`.
- 2026-05-06: ampliado `Activos fijos` dentro de `contabilidad_colombia_avanzada` con depreciacion por periodo, mantenimiento, traslados, bajas, eventos, resumen gerencial y documento `documentos/activos_fijos_avanzado.md`.
- 2026-05-06: ampliado `Nomina de sueldos` con capa `Nomina Colombia avanzada` para conceptos legales, novedades aprobables, resumen PILA, dashboard, seccion administrativa y documento `documentos/nomina_colombia_avanzada.md`.
- 2026-05-06: agregado modulo `Produccion / MRP` con API `/api/empresa/produccion_mrp`, recetas/BOM, componentes, ordenes, consumos, calidad, plan MRP, permisos/licencia `produccion_mrp`, pantalla administrativa y documento `documentos/produccion_mrp.md`.
- 2026-05-06: agregado modulo `Tesoreria y presupuesto` con API `/api/empresa/tesoreria_presupuesto`, cuentas banco/caja, presupuestos, partidas, flujo de caja proyectado, permisos/licencia `tesoreria_presupuesto`, pantalla administrativa y documento `documentos/tesoreria_presupuesto.md`.
- 2026-05-05: reorganizado el menu empresarial para fusionar Finanzas, Contabilidad Colombia y Suite contable bajo `Centro financiero y contable`; la ayuda administrativa principal queda restringida a `super_administrador` y se agrega el rol `control_super_administrador` para supervision limitada del panel super.
- 2026-05-05: agregado paquete de reportes avanzados comparables con software colombiano conocido: ventas diarias POS, rentabilidad, Kardex valorizado, compras por proveedor, balance de prueba, libros contables, impuestos/retenciones, exogena base y edades de cartera CxC/CxP.
- 2026-05-05: agregado modulo `Contabilidad Colombia avanzada` con informacion exogena DIAN/medios magneticos, nomina electronica, documento soporte, activos fijos, cartera/CxP, libros oficiales, API `/api/empresa/contabilidad_colombia_avanzada`, pagina administrativa y control por licencia `contabilidad_colombia_avanzada`.
- 2026-05-05: agregado modulo profesional `Carnets empresariales` con API `/api/empresa/carnets`, pagina `administrar_empresa/carnets.html`, plantillas, QR, exportacion, bitacora y control por licencia `carnets`.
- 2026-05-05: rectificado el aislamiento multiempresa de todos los modulos privados `/api/empresa/...`; los wrappers `WithEmpresa*` rechazan inconsistencias de `empresa_id` entre URL, cabecera, formulario/multipart y JSON.
- 2026-05-05: creado `documentos/reporte_estado_modulos_2026-05-05.md` como corte integral del estado actual de modulos, portal publico, carta QR, Motel Calipso, domicilios, Taxi System, roles/licencias, apariencia y base de datos.
- 2026-05-05: actualizado el portal `web/index.html` con descripciones comerciales completas de modulos, incluyendo POS, hotel/motel, gimnasio, odontologia, domicilios tipo Rappi, Taxi System tipo Uber, turnos, control electrico, carta QR, red social, roles/licencias y hoja de vida.
- 2026-05-05: publicada y documentada la operacion real de Motel Calipso: venta publica, carta publica de productos/precios, QR exportable desde administracion y publicaciones en red social comercial.
- 2026-05-05: corregida y documentada la exposicion publica de `visualizar_productos_y_precios_publico.html`; la ruta directa y la ruta `/{empresa_slug}/visualizar_productos_y_precios_publico.html` quedan sin login y validadas en produccion.
- 2026-05-05: documentado `Domicilios` profesional en `documentos/domicilios_profesional.md`, con central, restaurantes, domiciliarios, cliente publico, tracking GPS, codigo de entrega, endpoints, datos demo y control independiente por roles/licencias.
- 2026-05-05: actualizada la matriz de roles/licencias para modulos verticales (`venta_publica`, `gimnasio`, `taxi_system`, `domicilios`, `alquileres`, `odontologia`, `turnos_atencion`, `control_electrico`) con wrappers dedicados y activacion por licencia.
- 2026-05-05: documentado `Taxi System` profesional en `documentos/taxi_system_profesional.md`, con mapa operativo, filtros, GPS por tipo/protocolo, asociacion de dispositivos a conductores y endpoints privados.
- 2026-05-05: documentada la carta publica de productos en `documentos/carta_publica_productos.md`, con modulo administrativo, pagina publica `visualizar_productos_y_precios_publico.html`, rutas por slug/subdominio, permisos y pruebas.
- 2026-05-03: estado de modulos actualizado en `documentos/reporte_estado_modulos_2026-05-03.md`; se documentan reparaciones de estaciones/carrito, retorno a estaciones al pagar, tarjetas adaptables al texto, `USD / COP` como primer indicador y despliegue VPS correcto.
- 2026-05-03: ayuda del sistema actualizada con estado operativo, rutas criticas, configuracion de estaciones, flujo de pago del carrito y advertencia honesta sobre validacion integral por hardware/proveedores.

- 2026-04-30: pagos Epayco documentados con Smart Checkout v2 y fallback clasico firmado por POST a `secure.payco.co`.
- 2026-04-30: pagos Epayco actualizados para separar modo Smart Checkout y modo fallback clasico; el POST clasico usa `Customer ID` + `P_KEY` para decidir produccion/pruebas y evitar "El comercio no fue reconocido".
- 2026-04-30: chat IA actualizado para exportar respuestas y conversaciones como PDF, DOCX, XLSX, TXT o JSON mediante el generador dinamico y auditoria de origen `chat_ia`.
- 2026-04-30: chat flotante documentado con robot IA, secretaria IA estilo caricatura ejecutiva joven y voz femenina automatica para secretaria.
- 2026-04-30: empresas compartidas documentadas con listado/revocacion de administradores compartidos y trazabilidad.
- 2026-04-30: hoja de vida operativa universal documentada para motos de taller, pacientes, vehiculos, equipos, activos o mascotas.
- 2026-04-30: documentos dinamicos asistidos por IA documentados con endpoints `/generate` y `/download`.

## Fuentes canonicas por tema

- Vision funcional y alcance actual: `documentos/descripcion_del_proyecto`
- Esquema fisico de base de datos: `documentos/estructura_bd.md`
- Diagrama entidad relacion vigente: `documentos/diagramas/diagrama_entidad_relacion.md`
- Imagen del DER: `documentos/diagramas/diagrama_entidad_relacion.svg`
- Arquitectura tecnica y mapa de archivos: `documentos/diagramas/estructura_del_codigo.md`
- Evolucion funcional por modulo: `documentos/descripcion_de_modulos`
- Matriz de roles, visibilidad y wrappers: `documentos/matriz_roles_permisos_pos_multiempresa.md`
- Plan de 20 verticales empresariales 2026: `documentos/plan_20_modulos_verticales_2026-05-10.md`
- Arquitectura de modulos universales y verticales: `documentos/arquitectura_modulos_universales.md`
- Roles, permisos finos y ayuda privada super: `documentos/reporte_roles_ayuda_super_2026-05-10.md`
- Carta publica de productos y precios: `documentos/carta_publica_productos.md`
- Suite contable Colombia avanzada: `documentos/contabilidad_colombia_avanzada.md`
- Declaraciones tributarias Colombia: `documentos/declaraciones_tributarias.md`
- Cierre y bloqueo fiscal: `documentos/cierre_fiscal.md`
- Activos fijos NIIF/Fiscal: `documentos/activos_fijos_niif_fiscal.md`
- Portal de terceros y certificados tributarios: `documentos/portal_terceros_certificados.md`
- Propiedad horizontal: `documentos/propiedad_horizontal.md`
- Promocion de licencias por asesor: `documentos/promocion_asesor_licencias.md`
- Centros de costo y rentabilidad: `documentos/centros_costo.md`
- Carnets empresariales: `documentos/carnets_empresariales.md`
- Produccion / MRP: `documentos/produccion_mrp.md`
- Logistica avanzada / WMS: `documentos/logistica_wms.md`
- Tesoreria y presupuesto: `documentos/tesoreria_presupuesto.md`
- Nomina Colombia avanzada: `documentos/nomina_colombia_avanzada.md`
- Activos fijos avanzado: `documentos/activos_fijos_avanzado.md`
- Importaciones y costeo: `documentos/importaciones_costeo.md`
- Compras avanzadas: `documentos/compras_avanzadas.md`
- Captura inteligente de compras/gastos: `documentos/soportes_compras_ia.md`
- QA transversal de modulos 2026-05-06: `documentos/reporte_qa_modulos_2026-05-06.md`
- Inventario avanzado: `documentos/inventario_avanzado.md`
- CRM y ventas avanzadas: `documentos/crm_ventas_avanzadas.md`
- Reporte de estado integral vigente: `documentos/reporte_estado_modulos_2026-05-05.md`
- Taxi System profesional y GPS: `documentos/taxi_system_profesional.md`
- Domicilios profesional: `documentos/domicilios_profesional.md`
- Inventario documental y de archivos: `documentos/descripcion_de_archivos`
- Historial detallado de trabajo: `documentos/historial_de_cambios`
- Resumen ejecutivo de cambios: `CHANGELOG.md`
- Gobernanza tecnica, ADRs, contratos y runbooks: `documentos/gobernanza_tecnica/README.md`
- Estado documental consolidado 2026-05-13: `documentos/estado_documentacion_2026-05-13.md`
- Runbook actualizado de pagos de licencias: `documentos/gobernanza_tecnica/runbooks/runbook_checkout_licencias.md`
- Contrato actualizado de checkout publico: `documentos/gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md`
- Contrato de documentos dinamicos IA: `documentos/gobernanza_tecnica/contratos/contrato_documentos_dinamicos_ia_exportacion.md`

## Paquetes documentales complementarios

- `documentos/erp_multiempresa/`: paquete formal de alcance, diseno tecnico, especificaciones funcionales y guia de implementacion ERP multiempresa.
- `documentos/manual_de_instalacion.md`: referencia de instalacion y arranque.
- `documentos/manual_vps_seguridad.md`: operacion y endurecimiento de VPS.
- `documentos/deploy_nginx_reverse_proxy_vps.md`: publicacion HTTPS y proxy reverso.
- `documentos/actualizaciones_del_repositorio.md`: historial de sincronizacion tras `scripts/actualizar_repositorio.ps1`.

## Regla de uso para cambios tecnicos

Antes de cambiar codigo, infraestructura o flujos criticos:

1. Leer `documentos/descripcion_del_proyecto`.
2. Leer `documentos/estructura_bd.md` si hay impacto de datos, tablas o consultas.
3. Leer `documentos/diagramas/estructura_del_codigo.md` si hay impacto de arquitectura o rutas.
4. Leer `documentos/gobernanza_tecnica/estandares_de_cambio_seguro.md`.
5. Consultar el ADR, contrato tecnico o runbook aplicable cuando exista.
