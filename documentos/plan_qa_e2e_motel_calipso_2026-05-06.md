# Plan QA extremo a extremo - Motel Calipso

Fecha: 2026-05-06
Estado: plan de ejecucion

## Seguridad

La prueba se ejecutara con el usuario indicado por el propietario del sistema. La contrasena no debe guardarse en este documento ni en archivos del repositorio; debe ingresarse solo en la sesion de navegador o cargarse desde una variable local temporal.

## Objetivo

Probar directamente en navegador las funciones criticas del sistema usando Motel Calipso, pulsando botones reales y verificando que no queden errores visibles, errores HTTP 4xx/5xx inesperados ni errores de consola JavaScript.

## Preparacion

1. Confirmar que produccion responda en `https://powerfulcontrolsystem.com/`.
2. Iniciar sesion con el usuario autorizado.
3. Seleccionar empresa Motel Calipso.
4. Activar captura de consola/red del navegador.
5. Probar en escritorio y movil.
6. Ejecutar un pase en tema claro y otro en tema oscuro.

## Matriz de prueba por bloque

1. Acceso, seleccion de empresa y menus
- Login.
- Seleccion Motel Calipso.
- Menu principal escritorio.
- Menu movil.
- Permisos del super administrador.
- Cambio de apariencia claro/oscuro.

2. Operacion POS/hotel/motel
- Estaciones.
- Activar estacion.
- Agregar productos/servicios.
- Carrito.
- Control electrico cuando aplique.
- Pago.
- Impresora/configuracion documento.
- Corte de caja.

3. Inventario y compras
- Productos.
- Inventario avanzado.
- Compras.
- Compras avanzadas.
- OCR/IA de compras.
- Importaciones y costeo.
- Produccion/MRP.
- Logistica WMS: ubicacion, orden, item, avance, despacho y bitacora.

4. Finanzas y contabilidad
- Egresos/ingresos.
- Creditos/cartera.
- Cobranza.
- Contabilidad Colombia.
- Contabilidad avanzada.
- Tesoreria/presupuesto.
- Centros de costo.
- Cierre fiscal.
- Declaraciones tributarias.
- Activos fijos NIIF/Fiscal.
- Portal terceros/certificados.
- Portal contador.

5. Verticales
- Domicilios.
- Taxi System.
- Parqueadero.
- Apartamentos turisticos.
- Propiedad horizontal.
- Gimnasio.
- Odontologia.
- Turnos.
- Carnets.
- Hoja de vida operativa.

6. Publico y comercial
- Carta publica de productos/precios.
- QR de carta.
- Venta publica.
- Red social comercial.
- Publicaciones de Motel Calipso.

7. Configuracion y administracion
- Configuracion empresa.
- Configuracion impresora.
- Configuracion estaciones.
- Configuracion sensores.
- Roles/permisos.
- Licencias.
- Chat flotante/radio flotante.
- Usuarios.
- Backups.

## Evidencia esperada por modulo

- URL probada.
- Botones pulsados.
- Resultado visible.
- Respuesta HTTP principal.
- Errores de consola, si existen.
- Captura de pantalla si hay fallo visual.
- Registro del dato demo creado, cuando aplique.

## Criterios de aprobacion

- No hay errores 500 en APIs principales.
- No hay botones visibles sin accion.
- No hay errores JavaScript al guardar, cargar demo, exportar, actualizar o abrir tabs.
- Las pantallas no quedan con texto ilegible en temas claros/oscuros.
- El menu no oculta modulos permitidos para el usuario autorizado.
- Los datos creados quedan aislados por `empresa_id` de Motel Calipso.

## Orden recomendado de ejecucion

1. Smoke test autenticado de menus y permisos.
2. Smoke test de nuevos modulos: WMS, declaraciones, activos, portal terceros, centros de costo, cierre fiscal, propiedad horizontal.
3. Prueba funcional de POS/hotel/motel.
4. Prueba financiera/contable.
5. Prueba de verticales.
6. Prueba publica sin sesion.
7. Reporte de defectos y correcciones.

## Ejecucion 2026-05-07

Ambiente:
- Base URL local: `http://127.0.0.1:8080`.
- Empresa: Motel Calipso, `empresa_id=7`.
- Usuario: super administrador autorizado.

Evidencias generadas:
- `backend/tmp_tools/qa_calipso_operativo/frontend_buttons_calipso_full_final_report.json`: regresion escritorio de 60 modulos, botones seguros y trials controlados.
- `backend/tmp_tools/qa_calipso_operativo/deep_flows_calipso_report.json`: flujos reales con datos QA, 6/6 pasos OK.
- `backend/tmp_tools/qa_calipso_operativo/frontend_buttons_calipso_mobile_modules_final_report.json`: pasada movil completa inicial y hallazgos.
- `backend/tmp_tools/qa_calipso_operativo/frontend_buttons_calipso_mobile_overlay_fix_report.json`: validacion movil dirigida posterior, 2/2 modulos OK.

Flujos reales cubiertos:
- Parqueadero: emision de ticket, token QR publico, calculo automatico, cobro/cierre y anulacion controlada de ticket QA.
- WMS: ubicacion, orden, item, avance de picking/packing y despacho.
- Centros de costo: centro, presupuesto, regla de imputacion y dashboard.
- Activos fijos NIIF/fiscal: alta de activo, depreciacion y evento.
- Red social: carga de imagen local y publicacion empresarial.
- Integraciones verificables: carta publica, venta publica, QR publico, Taxi System/mapa/GPS visible.

Correcciones derivadas:
- Ajustes defensivos en activos fijos, auditoria, red social y graficos/estadisticas.
- Robot/secretaria y radio flotante compactados en movil para no interceptar botones.
- Runner QA con limpieza de overlays y viewport configurable.

Pendiente externo:
- Impresora fisica, sensores electricos reales, GPS real, DIAN y pasarelas en produccion deben validarse con hardware/credenciales reales.
