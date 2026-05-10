# Resumen del proyecto Powerful Control System

Actualizacion: 2026-05-10

Revision 2026-05-10: se actualiza el sistema de roles y permisos con modulos finos para CRM unificado, reservas hoteleras, chat/tareas, horarios, asistencia, vehiculos, hoja de vida operativa, GPS, nomina, reportes, auditoria, backups, documentos OnlyOffice y Nextcloud. Backend, menu empresarial, paginas `link...`, wrappers API y compatibilidad de licencias quedan documentados en `documentos/reporte_roles_ayuda_super_2026-05-10.md`.

Ayuda super 2026-05-10: la ayuda administrativa completa `/ayuda/ayuda.html` sigue restringida al rol `super_administrador` y ahora se abre desde el boton `Ayuda super administrador` dentro de `web/super_administrador.html`. El rol `control_super_administrador`, administradores de empresa y usuarios operativos no ven esa ayuda privada; conservan ayudas publicas o contextuales.

Powerful Control System es una plataforma POS/ERP SaaS multiempresa orientada a comercios, restaurantes, hoteles, moteles, gimnasios, consultorios odontologicos, copropiedades, domicilios, taxi/flotas y operaciones con estaciones o puntos de atencion. El sistema permite administrar empresas desde un panel central, operar ventas por carritos o estaciones, controlar usuarios, carnets empresariales, inventario, finanzas, bancos y pagos masivos, gestion documental, KYC/KYB y riesgo LAFT, contratos y obligaciones, helpdesk, calidad/procesos, cierre y bloqueo fiscal, centros de costo/rentabilidad, activos fijos e intangibles NIIF/Fiscal, certificados tributarios para terceros, facturacion, propiedad horizontal, reportes, soporte remoto, venta publica, red social comercial, carta QR y herramientas de inteligencia artificial. La operacion productiva esta pensada para PostgreSQL en VPS, con separacion entre base global de super administrador y base operativa de empresas.

Revision 2026-05-06: se ejecuto QA transversal autenticado sobre Motel Calipso (`empresa_id=7`), con paginas y APIs principales de modulos recientes respondiendo HTTP 200. Se profesionalizaron permisos locales de menu para `administrador_total`, rendimiento de dashboards de Cobranza, Portal contador y Captura inteligente de compras/gastos, y validacion segura de enlaces dinamicos de soportes IA. El detalle queda en `documentos/reporte_qa_modulos_2026-05-06.md`.

La portada publica `web/index.html` se actualizo para que la descripcion comercial de modulos incluya tambien Bancos y pagos masivos, Gestion documental, KYC/KYB y LAFT, Contratos, Helpdesk, Calidad/procesos, Cobranza, Portal contador, Captura IA/OCR de compras y gastos, AIU construccion, Parqueaderos con ticket QR y Apartamentos turisticos, manteniendo coherencia con roles/licencias y la documentacion funcional vigente.

## Tecnologias principales

- **Backend:** Go, con servidor HTTP propio, handlers por modulo, middleware de permisos, utilidades de seguridad y procesos operativos de monitoreo.
- **Frontend:** HTML, CSS y JavaScript nativo. La interfaz se divide entre portal publico, panel super administrador y panel empresarial (`administrar_empresa`).
- **Base de datos:** PostgreSQL como motor canonico. El sistema usa dos contextos principales: `pcs_superadministrador` para configuracion global, administradores, licencias y contrato; y `pcs_empresas` para operacion por empresa.
- **Autenticacion:** Google OAuth para administradores, login por correo/contraseña para administradores confirmados y portal de usuarios internos de empresa con confirmacion por correo, contrato y recuperacion de contraseña.
- **Correo:** envio por SMTP/Gmail configurable desde super administrador para confirmaciones, recuperacion de contraseña, activacion de licencias, alertas y plantillas editables.
- **Pagos:** integraciones con Wompi y Epayco para licencias y venta publica, con configuracion global y por empresa segun el flujo. El checkout de licencias usa Smart Checkout v2 cuando hay sesion valida y fallback clasico firmado por POST a `https://secure.payco.co/checkout.php` para evitar redirecciones GET que terminan en `AccessDenied`; el fallback clasico resuelve produccion/pruebas con `Customer ID` + `P_KEY` para no enviar comercios reales como prueba.
- **Facturacion electronica:** modulo por empresa con configuracion por pais. Para Colombia incluye preparacion DIAN, datos tributarios, resoluciones, consecutivos, firma y trazabilidad documental. El ciclo documental soporta factura electronica, nota credito, nota debito, documento soporte en adquisiciones, nomina electronica y documento equivalente POS electronico, usando la misma auditoria, cola fiscal y separacion por `empresa_id`.
- **Inteligencia artificial:** chat IA empresarial, chat IA global para super administrador y chat publico comercial del portal. Soporta adjuntos/fotos en flujos empresariales, acciones confirmables (`PCS_ACTION`), voz natural por servidor abierto configurable, avatar robot, secretaria IA estilo caricatura ejecutiva con voz femenina y consumo controlado por proveedor/modelo.
- **Documentos:** OnlyOffice para documentos por empresa, Nextcloud como almacenamiento/colaboracion configurable y generacion dinamica de documentos asistida por IA (`/generate`, `/download`) con salida HTML/PDF/DOCX/XLSX/TXT/JSON. Las integraciones se controlan desde super y empresa con manejo de errores cuando estan desactivadas o desconfiguradas.
- **Soporte remoto:** RustDesk como flujo recomendado, con configuracion super del servicio y configuracion empresarial para publicar datos de conexion, instrucciones y acceso remoto.
- **Servidor y operacion:** despliegue a VPS con scripts PowerShell/Bash, Nginx como reverse proxy esperado, certificados HTTPS wildcard, servicio opcional de voz IA streaming con Piper TTS y herramientas de diagnostico del backend.

## Modulos principales

### Portal publico y comercial

El portal publico presenta la oferta comercial, tarjetas configurables, informacion de contacto, chat IA del index, venta digital y paginas publicas de empresas. Desde 2026-05-05 el index describe los modulos con enfoque comercial y operativo actualizado: POS, hotel/motel, gimnasio, odontologia, domicilios tipo Rappi, Taxi System tipo Uber, turnos, control electrico, carta publica con QR, red social, permisos/licencias y hoja de vida. Tambien permite consulta publica de catalogos, pagos de productos y pagos de licencias cuando las pasarelas estan configuradas.

### Super administrador

El panel super administra empresas, licencias, tipos de empresa, administradores, contrato, configuracion avanzada, pasarelas de pago, correo SMTP, IA, pagina principal, errores del sistema, metricas del VPS, seguridad, roles, permisos y descarga consolidada de informacion empresarial. Desde este panel tambien se editan plantillas de correo, configuracion de OnlyOffice, Nextcloud, consumos externos y datos oficiales para alimentar la IA del portal.

Desde 2026-04-30, el flujo de pago de licencias por Epayco conserva Smart Checkout v2 como ruta primaria y solo usa checkout clasico mediante formulario POST firmado. Si Smart Checkout falla y no existen `epayco.customer_id` y `epayco.checkout_key`/`epayco.p_key`, el backend devuelve error controlado en vez de redirigir al usuario a una URL remota invalida. El formulario clasico usa modo independiente de Smart Checkout y debe salir como produccion cuando las credenciales son reales.

### Administracion por empresa

Cada empresa opera su propio panel con aislamiento por `empresa_id`. El panel incluye usuarios internos, roles, clientes, productos, bodegas, inventario, compras, ventas, carritos, estaciones, facturacion, finanzas, reportes, auditoria, asistencia de empleados, backups, venta publica, soporte remoto, integraciones y chat/tareas.

La administracion empresarial incorpora una configuracion guiada asistida por IA para la puesta en marcha de nuevas empresas. El robot puede levantar datos del negocio, sugerir parametros por industria y aplicar configuraciones base para estaciones, impresion, cobro, facturacion y operacion inicial.

La administracion empresarial tambien permite ver desde el editor de empresa a que administradores se compartio el acceso, tanto para quien comparte como para quien recibe. Cualquier administrador con acceso vigente puede retirar otro acceso compartido con trazabilidad y registro del actor.

La hoja de vida operativa universal permite llevar historial de motos de taller, pacientes, vehiculos, equipos, activos o mascotas dentro de Administrar empresa. Cada ficha concentra eventos, servicios, alertas, recurrencias y resumen operativo para seguimiento empresarial.

### Ventas, carritos y estaciones

El flujo de venta se basa en carritos de compra. Puede operar como mostrador, estacion, habitacion, mesa o punto de caja. Soporta productos, servicios, combos, descuentos, impuestos, propinas, comisiones, pagos mixtos, recuperacion de sesiones y cierre de venta. El carrito puede generar documentos de venta, descontar inventario, registrar metricas y dejar auditoria. El lector de codigo de barras funciona con `codigo_barras` o `sku`; tambien existe generador de etiquetas Code 128 para productos.

La configuracion de impresoras ya esta separada de la configuracion general y permite definir impresoras por equipo, funcionalidad y producto. La venta siempre genera comprobante y puede, si asi se configura, emitir ademas factura electronica automaticamente o manualmente a partir de una venta ya realizada.

### Inventario y compras

Inventario administra productos, categorias, bodegas, proveedores, existencias, movimientos, alertas de quiebre, reposicion preventiva, plan de compra, ordenes, recepciones y contabilizacion de compras. Tambien soporta combos, lotes/series, transferencias, ajustes y reportes multiformato.

### Finanzas y contabilidad

Finanzas registra ingresos y egresos, comprobantes adjuntos, configuracion contable, periodos, cierres de caja, conciliacion bancaria, eventos contables y asientos. Incluye categorias y cuentas base para operacion en Colombia, con destinos configurables como SIIGO, World Office, Alegra, Helisa, Loggro y ContaPyme. El Centro financiero y contable incorpora Cierre y bloqueo fiscal para proteger periodos cerrados, documentos reportados y operaciones post-cierre con politicas por modulo, excepciones aprobadas, simulador y bitacora; tambien incorpora Centros de costo y rentabilidad para medir utilidad por sucursal, area, unidad de negocio o proyecto con maestro, reglas, presupuesto, dashboard y movimientos integrados desde contabilidad/tesoreria/compras/OCR/AIU. El modulo de Activos Fijos e Intangibles NIIF/Fiscal administra PPE e intangibles con vida util contable y fiscal, depreciacion, deterioro, valor fiscal, diferencia NIIF/fiscal, seguros, ubicaciones, responsables, traslados, bajas y mantenimientos. Tambien incorpora Gestion de cobranza como modulo independiente para recuperar cartera reutilizando `empresa_cuentas_por_cobrar`, con dashboard, campanas, plantillas multicanal, gestiones, promesas de pago, simulacion de envio y exportacion CSV por empresa, Portal contador como oficina virtual para firmas contables, y Portal de terceros y certificados tributarios para proveedores, clientes, empleados y contratistas con enlaces publicos seguros, impresion y bitacora de descargas. Los reportes entregan KPI, flujo de caja, estado de resultados, balance, auditoria y exportes en PDF, XLS, CSV, JSON y TXT.

### Facturacion electronica e impuestos

El sistema permite configurar facturacion electronica por empresa y pais, con soporte inicial para Colombia, Ecuador y Panama. Para Colombia se mantiene trazabilidad por empresa y NIT, sin reutilizar tokens ni firmas entre empresas. El flujo operacional cubre documentos electronicos adicionales a la factura: nota credito, nota debito, documento soporte, nomina electronica y documento equivalente POS electronico, todos registrables desde la pantalla de facturacion electronica y conciliables en la cola DIAN/proveedor. El modulo de impuestos permite parametrizar tasas por empresa y generar reportes de deuda estimada.

El modulo AIU construccion complementa facturacion electronica para arquitectos, constructoras, contratistas y pequenas empresas de obra. Permite crear contratos con responsable, centro de costo, modalidad, riesgo, avance, capitulos y conceptos; calcula Administracion/Imprevistos/Utilidad, define si la base AIU se suma o no al total, permite escoger base de IVA, retenciones, anticipo, garantia y neto a cobrar, controla estados de aprobacion/ejecucion/cierre y genera una factura electronica AIU enlazada al repositorio de documentos de facturacion por `empresa_id`.

### Captura inteligente de compras y gastos

El modulo `soportes_compras_ia` permite radicar soportes de compra o gasto con foto, PDF o XML, extraer datos con OCR/IA usando `openai:gpt-5.5`, detectar duplicados por hash/documento, marcar revision humana por confianza, aprobar o rechazar y convertir soportes aprobados en cuentas por pagar. Queda enlazado por `empresa_id`, controlado por permisos/licencia y ubicado en `Administrar empresa > Compras > Captura IA/OCR`.

### Usuarios, permisos y seguridad

Los usuarios internos de empresa se crean desde el panel empresarial, reciben confirmacion por correo, aceptan contrato y operan segun rol. Los permisos se controlan por modulo, accion y visibilidad de paginas, combinando licencia, rol y reglas de menu. Las rutas empresariales usan wrappers de autorizacion por empresa y modulo. La auditoria empresarial registra acciones criticas con usuario, request, modulo, resultado y exportes forenses.

### RRHH, asistencia y operacion interna

Asistencia registra entrada, salida, estado, novedades, horas trabajadas y cierres de periodo. Puede vincular registros con usuarios internos de empresa para mantener coherencia entre usuario, rol y empleado. Tambien existen modulos de nomina, vacaciones/licencias, vehiculos, agenda, chat/tareas y calendario compartido.

El sistema tambien incluye un modulo profesional de horarios laborales para programar turnos de trabajadores por sede, area, cargo y recurrencia, con reglas de jornada, publicacion, conflictos y cobertura pendiente.

## Modulos verticales especializados

### Gimnasios

El sistema incluye un modulo de gimnasio con dashboard, socios, planes, entrenadores, clases, inscripciones, asistencias, pagos y control de acceso. Soporta RFID, NFC, QR, PIN, biometria y facial, con credenciales, dispositivos, bitacora de eventos y politica de acceso configurable. Desde 2026-05-02 el modulo incorpora preconfiguracion operativa para crear planes base, clases iniciales y dispositivos de ingreso, y el esquema de base de datos migra de forma defensiva para empresas que vengan de versiones anteriores.

### Consultorios odontologicos

Existe un modulo de consultorio odontologico con pacientes, profesionales, consultorios, citas, historias clinicas, odontogramas, tratamientos, presupuestos y pagos. La intencion es cubrir agenda, trazabilidad clinica y operacion financiera del consultorio desde el mismo panel empresarial.

### Domicilios y restaurantes

El modulo de domicilios opera como una central tipo Rappi para restaurantes, clientes y domiciliarios. Incluye productos/menu, pedidos, estados de cocina, ofertas a domiciliarios, presencia movil, ubicacion GPS, tracking en tiempo real y codigo de entrega. La central administrativa administra configuracion, restaurantes, domiciliarios, pedidos y metricas.

### Taxi System y flotas

Taxi System opera como una central tipo Uber para conductores, clientes y administracion. Incluye mapa operativo, solicitudes, despacho, estados del viaje, conductores, vehiculos, GPS configurable, dispositivos externos y trazabilidad de ruta. El modulo puede controlarse por licencia, rol y usuario.

### Venta publica, carta QR y red social

La venta publica permite publicar tiendas, secciones, productos, precios, fotos y checkout bajo slug o subdominio de la empresa. La carta publica `visualizar_productos_y_precios_publico.html` es una vista externa de solo lectura con QR exportable en PNG, SVG y PDF desde Administrar empresa. La red social comercial permite publicar posts empresariales con imagen/video para visibilidad externa.

### Taxi system y movilidad

El proyecto incorpora `Taxi system`, un modulo de despacho tipo ridesharing con clientes web opcionales registrados, conductores, central, ofertas por cercania, GPS en tiempo real, mapas, estados del servicio y uso del celular del conductor para geoposicion. Esta capa tambien sirve como base de tracking para domicilios y otros flujos de movilidad.

### Turnos de atencion

Se dispone de un modulo de turnos tipo banco/EPS con emision de tickets, puestos de atencion, kiosco publico, pantalla TV y estados del turno. Esto permite operar recepciones, ventanillas, cajas o procesos de fila ordenada dentro de la misma plataforma.

### Alquileres

El modulo de alquileres administra activos rentables como herramientas, equipos, vehiculos o maquinaria. Incluye catalogo de activos, tarifas, contratos, reservas, entrega, devolucion, garantias, mantenimientos, GPS y KPI de utilizacion e ingresos.

### IA, documentos y automatizacion

La IA empresarial puede responder preguntas de operacion, analizar adjuntos, proponer acciones confirmables y registrar productos o egresos desde una foto cuando el administrador confirma. El chat global de super ayuda con diagnostico y contexto del sistema. OnlyOffice permite gestionar documentos por empresa, y Nextcloud puede aprovisionar almacenamiento por empresa cuando esta configurado.

#### Funciones con inteligencia artificial

- Chat IA empresarial con contexto de `empresa_id`, restricciones de modelo y uso, y soporte para adjuntos/fotos en consultas avanzadas.
- Chat IA global para super administrador con consolidacion multinivel de datos entre empresas, vision holistica y capacidades de diagnostico de sistema.
- Chat flotante configurable por empresa entre ventana cuadrada, robot IA y secretaria IA. La secretaria usa apariencia de caricatura ejecutiva joven y voz femenina automaticamente; el robot conserva voz configurable.
- Contexto IA por auditoria en tiempo real: GPT-5.4 mini o el modelo asociado recibe actividad reciente, busqueda profunda por intencion y resultados de consultas DB seguras ya resueltas por el backend.
- Voz IA streaming: servicio abierto FastAPI + Piper TTS desplegable en VPS, desactivado por defecto y configurable desde Super Administrador; el chat usa proxy backend para no exponer la URL interna y degrada a voz del navegador/texto si el servicio no responde.
- Generacion dinamica de documentos: el backend recibe contenido o prompt IA, aplica variables con templates Go, renderiza HTML profesional y permite descargar el resultado en PDF, DOCX, XLSX, HTML, TXT o JSON.
- Descubrimiento de musica YouTube asistido por IA en la tarjeta de estaciones: sugiere playlists y videos actuales, permite cargar enlaces validos y ofrece busquedas inteligentes cuando no hay fuente directa.
- Integracion profesional con OpenAI/GPT gestionando modelo preferido, limites diarios, registro de consumo y configuracion centralizada desde super administrador.

### Soporte remoto, backups y VPS

El sistema incluye soporte remoto empresarial con RustDesk, backups por empresa, exportacion/importacion de configuracion, monitoreo de errores, seguridad VPS, metricas de trafico, procesos y servicios. Los scripts operativos permiten iniciar el servidor local, actualizar repositorio y sincronizar al VPS sin documentar secretos.

## Arquitectura operacional

El backend registra rutas en `backend/main.go`, delega reglas HTTP a `backend/handlers`, persistencia a `backend/db` y seguridad/utilidades a `backend/utils`, `backend/auth`, `backend/secure` y `backend/vpssecurity`. La interfaz vive en `web/`, con paginas publicas, subpaginas super y subpaginas empresariales embebidas en iframes. La documentacion tecnica vigente vive en `documentos/`, incluyendo descripcion del proyecto, modulos, base de datos, permisos, diagramas, runbooks y trazabilidad historica.

## Seguridad, trazabilidad e IA contextual

El proyecto no debe registrar secretos en texto plano. Credenciales de pago, correo, IA, OnlyOffice, Nextcloud y DIAN se manejan por entorno o configuracion cifrada/referenciada. Toda operacion relevante mantiene `empresa_id`, usuario, estado, observaciones y fechas. Desde 2026-04-29 la IA empresarial y global obtiene una ventana reciente de `empresa_auditoria_eventos` como contexto operativo en tiempo real: usuarios, modulos, endpoints, resultados y errores. La integracion es centralizada en auditoria y prompt, no en cada modulo; si la auditoria o la IA fallan, el servidor conserva operacion normal con contexto degradado.

La auditoria tambien alimenta una capa de contexto profundo para GPT-5.4 mini/modelo activo: detecta preguntas sobre usuarios, modulos, errores, ventas, finanzas, inventario o clientes; busca eventos auditados relevantes; ejecuta consultas permitidas por whitelist y `empresa_id`; y registra cada consulta IA en `empresa_auditoria_ia_consultas` con filtros, resumen y cantidad de eventos usados. Adicionalmente, el chat empresarial tiene por defecto acceso de lectura total controlada a la base operativa de la empresa: el backend lista tablas con `empresa_id`, consulta filas recientes/relevantes con SELECT parametrizado, omite columnas sensibles y entrega resultados ya resueltos al prompt. Super administrador puede activar/desactivar esta capacidad y ajustar limites de tablas/filas desde la configuracion logica del chat IA. El modelo no recibe credenciales ni permiso para ejecutar SQL libre. Los cambios funcionales deben actualizar documentacion y dejar registro en `documentos/historial_de_cambios`.

## Actualizacion 2026-05-03 - Estado operativo y ayuda

Se actualiza la documentacion del proyecto con el reporte `documentos/reporte_estado_modulos_2026-05-03.md` y la ayuda web `web/ayuda/ayuda.html`. La version documentada incluye reparacion de estaciones/carrito, pago con retorno automatico a estaciones, tarjetas de estaciones con tamano `Se adapta al texto`, indicador `USD / COP` primero en el panel empresarial y despliegue VPS operativo.

Criterio de calidad: los modulos principales estan implementados u operativos, pero no se declara ausencia absoluta de errores sin una matriz integral de pruebas por rol, empresa, hardware real y proveedor externo.
Actualizacion 2026-05-06: se agrega `Logistica avanzada / WMS` como modulo operativo de bodega, con ubicaciones internas, ordenes WMS, picking, packing, despachos, rutas, bitacora, permisos/licencia y pantalla profesional en Administrar empresa > Inventario y compras.

Actualizacion 2026-05-06: se agrega `Declaraciones Tributarias y Motor de Impuestos Colombia` como modulo financiero formal, con preliquidacion de IVA, retenciones, ICA, consumo, renta y regimen simple, calendario tributario editable, movimientos de conciliacion, saldos, soportes, permisos/licencia y pantalla profesional en Administrar empresa.
