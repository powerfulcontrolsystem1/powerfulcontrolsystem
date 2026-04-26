# Resumen del proyecto Powerful Control System

Actualizacion: 2026-04-26

Powerful Control System es una plataforma POS/ERP SaaS multiempresa orientada a comercios, restaurantes, hoteles, moteles y operaciones con estaciones o puntos de atencion. El sistema permite administrar empresas desde un panel central, operar ventas por carritos o estaciones, controlar usuarios, inventario, finanzas, facturacion, reportes, soporte remoto y herramientas de inteligencia artificial. La operacion productiva esta pensada para PostgreSQL en VPS, con separacion entre base global de super administrador y base operativa de empresas.

## Tecnologias principales

- **Backend:** Go, con servidor HTTP propio, handlers por modulo, middleware de permisos, utilidades de seguridad y procesos operativos de monitoreo.
- **Frontend:** HTML, CSS y JavaScript nativo. La interfaz se divide entre portal publico, panel super administrador y panel empresarial (`administrar_empresa`).
- **Base de datos:** PostgreSQL como motor canonico. El sistema usa dos contextos principales: `pcs_superadministrador` para configuracion global, administradores, licencias y contrato; y `pcs_empresas` para operacion por empresa.
- **Autenticacion:** Google OAuth para administradores, login por correo/contraseña para administradores confirmados y portal de usuarios internos de empresa con confirmacion por correo, contrato y recuperacion de contraseña.
- **Correo:** envio por SMTP/Gmail configurable desde super administrador para confirmaciones, recuperacion de contraseña, activacion de licencias, alertas y plantillas editables.
- **Pagos:** integraciones con Wompi y Epayco para licencias y venta publica, con configuracion global y por empresa segun el flujo.
- **Facturacion electronica:** modulo por empresa con configuracion por pais. Para Colombia incluye preparacion DIAN, datos tributarios, resoluciones, consecutivos, firma y trazabilidad documental.
- **Inteligencia artificial:** chat IA empresarial, chat IA global para super administrador y chat publico comercial del portal. Soporta adjuntos/fotos en flujos empresariales, acciones confirmables (`PCS_ACTION`) y consumo controlado por proveedor/modelo.
- **Documentos:** OnlyOffice para documentos por empresa y Nextcloud como almacenamiento/colaboracion configurable. Ambos se controlan desde super y empresa con manejo de errores cuando estan desactivados o desconfigurados.
- **Soporte remoto:** RustDesk como flujo recomendado, con configuracion super del servicio y configuracion empresarial para publicar datos de conexion, instrucciones y acceso remoto.
- **Servidor y operacion:** despliegue a VPS con scripts PowerShell/Bash, Nginx como reverse proxy esperado, certificados HTTPS wildcard y herramientas de diagnostico del backend.

## Modulos principales

### Portal publico y comercial

El portal publico presenta la oferta comercial, tarjetas configurables, informacion de contacto, chat IA del index, venta digital y paginas publicas de empresas. Tambien permite consulta publica de catalogos, pagos de productos y pagos de licencias cuando las pasarelas estan configuradas.

### Super administrador

El panel super administra empresas, licencias, tipos de empresa, administradores, contrato, configuracion avanzada, pasarelas de pago, correo SMTP, IA, pagina principal, errores del sistema, metricas del VPS, seguridad, roles, permisos y descarga consolidada de informacion empresarial. Desde este panel tambien se editan plantillas de correo, configuracion de OnlyOffice, Nextcloud, consumos externos y datos oficiales para alimentar la IA del portal.

### Administracion por empresa

Cada empresa opera su propio panel con aislamiento por `empresa_id`. El panel incluye usuarios internos, roles, clientes, productos, bodegas, inventario, compras, ventas, carritos, estaciones, facturacion, finanzas, reportes, auditoria, asistencia de empleados, backups, venta publica, soporte remoto, integraciones y chat/tareas.

### Ventas, carritos y estaciones

El flujo de venta se basa en carritos de compra. Puede operar como mostrador, estacion, habitacion, mesa o punto de caja. Soporta productos, servicios, combos, descuentos, impuestos, propinas, comisiones, pagos mixtos, recuperacion de sesiones y cierre de venta. El carrito puede generar documentos de venta, descontar inventario, registrar metricas y dejar auditoria. El lector de codigo de barras funciona con `codigo_barras` o `sku`; tambien existe generador de etiquetas Code 128 para productos.

### Inventario y compras

Inventario administra productos, categorias, bodegas, proveedores, existencias, movimientos, alertas de quiebre, reposicion preventiva, plan de compra, ordenes, recepciones y contabilizacion de compras. Tambien soporta combos, lotes/series, transferencias, ajustes y reportes multiformato.

### Finanzas y contabilidad

Finanzas registra ingresos y egresos, comprobantes adjuntos, configuracion contable, periodos, cierres de caja, conciliacion bancaria, eventos contables y asientos. Incluye categorias y cuentas base para operacion en Colombia, con destinos configurables como SIIGO, World Office, Alegra, Helisa, Loggro y ContaPyme. Los reportes entregan KPI, flujo de caja, estado de resultados, balance, auditoria y exportes en PDF, XLS, CSV, JSON y TXT.

### Facturacion electronica e impuestos

El sistema permite configurar facturacion electronica por empresa y pais, con soporte inicial para Colombia, Ecuador y Panama. Para Colombia se mantiene trazabilidad por empresa y NIT, sin reutilizar tokens ni firmas entre empresas. El modulo de impuestos permite parametrizar tasas por empresa y generar reportes de deuda estimada.

### Usuarios, permisos y seguridad

Los usuarios internos de empresa se crean desde el panel empresarial, reciben confirmacion por correo, aceptan contrato y operan segun rol. Los permisos se controlan por modulo, accion y visibilidad de paginas, combinando licencia, rol y reglas de menu. Las rutas empresariales usan wrappers de autorizacion por empresa y modulo. La auditoria empresarial registra acciones criticas con usuario, request, modulo, resultado y exportes forenses.

### RRHH, asistencia y operacion interna

Asistencia registra entrada, salida, estado, novedades, horas trabajadas y cierres de periodo. Puede vincular registros con usuarios internos de empresa para mantener coherencia entre usuario, rol y empleado. Tambien existen modulos de nomina, vacaciones/licencias, vehiculos, agenda, chat/tareas y calendario compartido.

### IA, documentos y automatizacion

La IA empresarial puede responder preguntas de operacion, analizar adjuntos, proponer acciones confirmables y registrar productos o egresos desde una foto cuando el administrador confirma. El chat global de super ayuda con diagnostico y contexto del sistema. OnlyOffice permite gestionar documentos por empresa, y Nextcloud puede aprovisionar almacenamiento por empresa cuando esta configurado.

### Soporte remoto, backups y VPS

El sistema incluye soporte remoto empresarial con RustDesk, backups por empresa, exportacion/importacion de configuracion, monitoreo de errores, seguridad VPS, metricas de trafico, procesos y servicios. Los scripts operativos permiten iniciar el servidor local, actualizar repositorio y sincronizar al VPS sin documentar secretos.

## Arquitectura operacional

El backend registra rutas en `backend/main.go`, delega reglas HTTP a `backend/handlers`, persistencia a `backend/db` y seguridad/utilidades a `backend/utils`, `backend/auth`, `backend/secure` y `backend/vpssecurity`. La interfaz vive en `web/`, con paginas publicas, subpaginas super y subpaginas empresariales embebidas en iframes. La documentacion tecnica vigente vive en `documentos/`, incluyendo descripcion del proyecto, modulos, base de datos, permisos, diagramas, runbooks y trazabilidad historica.

## Seguridad y trazabilidad

El proyecto no debe registrar secretos en texto plano. Credenciales de pago, correo, IA, OnlyOffice, Nextcloud y DIAN se manejan por entorno o configuracion cifrada/referenciada. Toda operacion relevante mantiene `empresa_id`, usuario, estado, observaciones y fechas. Los cambios funcionales deben actualizar documentacion y dejar registro en `documentos/historial_de_cambios`.

