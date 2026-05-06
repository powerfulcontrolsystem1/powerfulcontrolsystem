# Indice documental del proyecto

Fecha: 2026-05-05
Estado: vigente, actualizado con reportes colombianos avanzados, suite contable Colombia avanzada, domicilios profesional, Taxi System profesional y carta publica de productos 2026-05-05

Este archivo organiza la lectura tecnica y funcional del repositorio para desarrollo, soporte y trabajo asistido por Copilot.

## Orden de lectura recomendado

1. `documentos/descripcion_del_proyecto`
2. `documentos/estructura_bd.md`
3. `documentos/diagramas/estructura_del_codigo.md`
4. `documentos/descripcion_de_modulos`
5. `documentos/matriz_roles_permisos_pos_multiempresa.md`
6. `documentos/reporte_estado_modulos_2026-05-05.md`
7. `documentos/gobernanza_tecnica/README.md`
8. `documentos/historial_de_cambios`
9. `CHANGELOG.md`

## Estado documental reciente
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
- Arquitectura tecnica y mapa de archivos: `documentos/diagramas/estructura_del_codigo.md`
- Evolucion funcional por modulo: `documentos/descripcion_de_modulos`
- Matriz de roles, visibilidad y wrappers: `documentos/matriz_roles_permisos_pos_multiempresa.md`
- Carta publica de productos y precios: `documentos/carta_publica_productos.md`
- Suite contable Colombia avanzada: `documentos/contabilidad_colombia_avanzada.md`
- Carnets empresariales: `documentos/carnets_empresariales.md`
- Reporte de estado integral vigente: `documentos/reporte_estado_modulos_2026-05-05.md`
- Taxi System profesional y GPS: `documentos/taxi_system_profesional.md`
- Domicilios profesional: `documentos/domicilios_profesional.md`
- Inventario documental y de archivos: `documentos/descripcion_de_archivos`
- Historial detallado de trabajo: `documentos/historial_de_cambios`
- Resumen ejecutivo de cambios: `CHANGELOG.md`
- Gobernanza tecnica, ADRs, contratos y runbooks: `documentos/gobernanza_tecnica/README.md`
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
