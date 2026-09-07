(function (root) {
  var order = [
    { id: "modulos", title: "Modulos del sistema", path: "/super/diagramas/modulos.html" },
    { id: "erd", title: "Base de datos / ERD", path: "/super/diagramas/base_de_datos_erd.html" },
    { id: "multiempresa", title: "Multiempresa / tenant", path: "/super/diagramas/multiempresa.html" },
    { id: "arquitectura", title: "Arquitectura general", path: "/super/diagramas/arquitectura.html" },
    { id: "ventas-pos", title: "Flujo ventas POS", path: "/super/diagramas/ventas_pos.html" },
    { id: "dian", title: "Facturacion electronica DIAN", path: "/super/diagramas/facturacion_dian.html" },
    { id: "inventario", title: "Flujo de inventario", path: "/super/diagramas/inventario.html" },
    { id: "roles-permisos", title: "Usuarios, roles y permisos", path: "/super/diagramas/roles_permisos.html" },
    { id: "api-endpoints", title: "API / endpoints", path: "/super/diagramas/api_endpoints.html" },
    { id: "despliegue", title: "Despliegue e infraestructura", path: "/super/diagramas/despliegue.html" },
    { id: "seguridad", title: "Seguridad", path: "/super/diagramas/seguridad.html" },
    { id: "auditoria-logs", title: "Auditoria y logs", path: "/super/diagramas/auditoria_logs.html" },
    { id: "reportes", title: "Reportes", path: "/super/diagramas/reportes.html" },
    { id: "integraciones", title: "Integraciones externas", path: "/super/diagramas/integraciones.html" },
    { id: "agentes", title: "Agentes automaticos", path: "/super/diagramas/agentes_automaticos.html" }
  ];

  var diagrams = {
    "modulos": {
      title: "Diagrama de modulos del sistema",
      badge: "Vista funcional",
      summary: "Agrupa las areas principales del POS multiempresa para ubicar cambios, permisos y documentacion por modulo.",
      nodes: [
        { id: "pcs", label: "PCS\nPOS multiempresa", x: 650, y: 40, kind: "core" },
        { id: "empresas", label: "Empresas y usuarios", x: 80, y: 180 },
        { id: "inventario", label: "Inventario\nproductos, bodegas, kardex", x: 360, y: 180 },
        { id: "ventas", label: "Ventas POS\ncarritos, caja, pagos", x: 650, y: 180 },
        { id: "dian", label: "Facturacion electronica\nDIAN / pais fiscal", x: 940, y: 180 },
        { id: "finanzas", label: "Finanzas y contabilidad\nNIIF, impuestos, nomina", x: 1220, y: 180 },
        { id: "operacion", label: "Restaurante, bar, motel\nestaciones, mesas, habitaciones", x: 220, y: 360 },
        { id: "reportes", label: "Reportes y auditoria", x: 520, y: 360 },
        { id: "config", label: "Configuracion\npermisos, impresoras, menu", x: 820, y: 360 },
        { id: "canales", label: "Canales digitales\nportal, WhatsApp, correo", x: 1120, y: 360 },
        { id: "agentes", label: "Agentes automaticos\nmantenimiento y soporte", x: 650, y: 520 }
      ],
      edges: [
        ["pcs", "empresas"], ["pcs", "inventario"], ["pcs", "ventas"], ["pcs", "dian"], ["pcs", "finanzas"],
        ["ventas", "operacion"], ["ventas", "reportes"], ["empresas", "config"], ["dian", "canales"], ["finanzas", "agentes"]
      ],
      source: `flowchart TB
  PCS["PCS POS multiempresa"]
  PCS --> Empresas["Empresas y usuarios"]
  PCS --> Inventario["Inventario: productos, bodegas, kardex"]
  PCS --> Ventas["Ventas POS: carritos, caja, pagos"]
  PCS --> DIAN["Facturacion electronica DIAN / pais fiscal"]
  PCS --> Finanzas["Finanzas, contabilidad, nomina e impuestos"]
  Ventas --> Operacion["Restaurante, bar, motel, estaciones"]
  Ventas --> Reportes["Reportes y auditoria"]
  Empresas --> Config["Configuracion, permisos, impresoras, menu"]
  DIAN --> Canales["Correo, portal, WhatsApp y notificaciones"]
  Finanzas --> Agentes["Agentes automaticos y mantenimiento"]`
    },
    "erd": {
      title: "Diagrama de base de datos / ERD",
      badge: "Obligatorio",
      summary: "Resume las tablas principales y relaciones por dominio. El detalle fisico completo sigue en documentos/estructura_bd.md.",
      nodes: [
        { id: "empresas", label: "empresas\nPK id", x: 650, y: 30, kind: "db" },
        { id: "users", label: "users\nFK empresa_id\nFK rol_usuario_id", x: 80, y: 170, kind: "db" },
        { id: "roles", label: "roles_de_usuario\nroles_de_usuario_permisos", x: 80, y: 330, kind: "db" },
        { id: "productos", label: "productos\ncategorias_productos\nbodegas", x: 360, y: 170, kind: "db" },
        { id: "inventario", label: "inventario_movimientos\nexistencias\ntraslados", x: 360, y: 330, kind: "db" },
        { id: "carritos", label: "carritos_compras\ncarrito_compra_items", x: 650, y: 170, kind: "db" },
        { id: "clientes", label: "clientes\ncrm_leads\ncrm_interacciones", x: 940, y: 170, kind: "db" },
        { id: "facturacion", label: "empresa_facturacion_documentos\nempresa_dian_configuracion\nfacturacion_electronica_reintentos", x: 940, y: 330, kind: "db" },
        { id: "finanzas", label: "empresa_finanzas_movimientos\nempresa_cierres_caja\nempresa_asientos_contables", x: 650, y: 330, kind: "db" },
        { id: "compras", label: "proveedores\nempresa_compras_documentos\ndetalle_compra", x: 1220, y: 170, kind: "db" },
        { id: "super", label: "pcs_superadministrador\nlicencias, administradores,\nconfiguraciones", x: 1220, y: 330, kind: "db" }
      ],
      edges: [
        ["empresas", "users"], ["users", "roles"], ["empresas", "productos"], ["productos", "inventario"],
        ["empresas", "carritos"], ["carritos", "finanzas"], ["carritos", "facturacion"], ["clientes", "carritos"],
        ["clientes", "facturacion"], ["compras", "inventario"], ["compras", "finanzas"], ["empresas", "super"]
      ],
      source: `erDiagram
  EMPRESAS ||--o{ USERS : empresa_id
  ROLES_DE_USUARIO ||--o{ USERS : rol_usuario_id
  EMPRESAS ||--o{ PRODUCTOS : empresa_id
  PRODUCTOS ||--o{ INVENTARIO_MOVIMIENTOS : producto_id
  EMPRESAS ||--o{ BODEGAS : empresa_id
  BODEGAS ||--o{ INVENTARIO_MOVIMIENTOS : bodega_id
  EMPRESAS ||--o{ CLIENTES : empresa_id
  CLIENTES ||--o{ CARRITOS_COMPRAS : cliente_id
  CARRITOS_COMPRAS ||--o{ CARRITO_COMPRA_ITEMS : carrito_id
  CARRITOS_COMPRAS ||--o{ EMPRESA_FINANZAS_MOVIMIENTOS : documento_ref
  CARRITOS_COMPRAS ||--o{ EMPRESA_FACTURACION_DOCUMENTOS : documento_origen
  EMPRESAS ||--o{ PROVEEDORES : empresa_id
  PROVEEDORES ||--o{ EMPRESA_COMPRAS_DOCUMENTOS : proveedor_id
  EMPRESA_FACTURACION_DOCUMENTOS ||--o{ FACTURACION_ELECTRONICA_REINTENTOS : documento_codigo
  EMPRESAS ||--o{ LICENCIAS : empresa_id`
    },
    "multiempresa": {
      title: "Diagrama multiempresa / multi-tenant",
      badge: "Aislamiento",
      summary: "Muestra como cada tabla operativa debe filtrar por empresa_id y como el super administrador gobierna configuraciones globales.",
      nodes: [
        { id: "db", label: "PostgreSQL\npcs_empresas", x: 650, y: 40, kind: "db" },
        { id: "empresa1", label: "empresa_id = 1", x: 220, y: 180, kind: "core" },
        { id: "empresa2", label: "empresa_id = 2", x: 650, y: 180, kind: "core" },
        { id: "empresaN", label: "empresa_id = N", x: 1080, y: 180, kind: "core" },
        { id: "datos1", label: "productos\nventas\nclientes\nfacturas\nusuarios", x: 220, y: 340 },
        { id: "datos2", label: "productos\nventas\nclientes\nfacturas\nusuarios", x: 650, y: 340 },
        { id: "datosN", label: "productos\nventas\nclientes\nfacturas\nusuarios", x: 1080, y: 340 },
        { id: "super", label: "pcs_superadministrador\nlicencias, tipos, roles,\nconfiguracion global", x: 650, y: 520, kind: "db" }
      ],
      edges: [["db", "empresa1"], ["db", "empresa2"], ["db", "empresaN"], ["empresa1", "datos1"], ["empresa2", "datos2"], ["empresaN", "datosN"], ["super", "db"]],
      source: `flowchart TB
  DB[(PostgreSQL / pcs_empresas)]
  DB --> E1["empresa_id = 1"]
  DB --> E2["empresa_id = 2"]
  DB --> EN["empresa_id = N"]
  E1 --> D1["productos, ventas, clientes, facturas, usuarios"]
  E2 --> D2["productos, ventas, clientes, facturas, usuarios"]
  EN --> DN["productos, ventas, clientes, facturas, usuarios"]
  Super[(pcs_superadministrador)]
  Super --> DB
  classDef rule fill:#fff7ed,stroke:#f97316,color:#1f2937
  class D1,D2,DN rule`
    },
    "arquitectura": {
      title: "Diagrama general de arquitectura",
      badge: "Capas",
      summary: "Conecta navegador, frontend estatico, API Go, PostgreSQL, servicios externos y despliegue Docker/VPS.",
      nodes: [
        { id: "usuario", label: "Usuario\nadmin, cajero, cliente", x: 80, y: 250 },
        { id: "web", label: "Frontend web\nHTML, CSS, JS", x: 330, y: 250 },
        { id: "api", label: "API Go\nhandlers + middleware", x: 600, y: 250, kind: "core" },
        { id: "db", label: "PostgreSQL\npcs_empresas\npcs_superadministrador", x: 890, y: 250, kind: "db" },
        { id: "externos", label: "Servicios externos\nDIAN, pagos, SMTP,\nWhatsApp, impresoras", x: 1180, y: 250, kind: "external" },
        { id: "vps", label: "VPS Docker\nNginx / App Go / Backups", x: 600, y: 470, kind: "infra" }
      ],
      edges: [["usuario", "web"], ["web", "api"], ["api", "db"], ["api", "externos"], ["vps", "web"], ["vps", "api"], ["vps", "db"]],
      source: `flowchart LR
  Usuario["Cliente web / POS"] --> Frontend["Frontend HTML/CSS/JS"]
  Frontend --> API["API en Go"]
  API --> DB[(PostgreSQL)]
  API --> DIAN["DIAN"]
  API --> Pagos["Pasarelas de pago"]
  API --> Correo["SMTP / Mailu"]
  API --> WhatsApp["WhatsApp"]
  API --> Impresoras["Impresoras POS"]
  VPS["VPS Docker + Nginx"] --> Frontend
  VPS --> API
  VPS --> DB`
    },
    "ventas-pos": {
      title: "Diagrama de flujo de ventas POS",
      badge: "Operacion",
      summary: "Describe el flujo operativo desde seleccion de empresa hasta inventario, caja, impresion y factura electronica si aplica.",
      nodes: [
        { id: "empresa", label: "Seleccionar empresa", x: 80, y: 260 },
        { id: "productos", label: "Seleccionar productos", x: 300, y: 260 },
        { id: "carrito", label: "Agregar al carrito", x: 520, y: 260, kind: "core" },
        { id: "descuento", label: "Descuento o promocion", x: 740, y: 260 },
        { id: "pago", label: "Medio de pago\nsimple o mixto", x: 960, y: 260 },
        { id: "venta", label: "Registrar venta", x: 1180, y: 260, kind: "core" },
        { id: "inventario", label: "Actualizar inventario", x: 520, y: 440, kind: "db" },
        { id: "caja", label: "Caja / imprimir recibo", x: 740, y: 440 },
        { id: "fe", label: "Factura electronica\nsi aplica", x: 960, y: 440 },
        { id: "dian", label: "Enviar a DIAN", x: 1180, y: 440, kind: "external" }
      ],
      edges: [["empresa", "productos"], ["productos", "carrito"], ["carrito", "descuento"], ["descuento", "pago"], ["pago", "venta"], ["venta", "inventario"], ["venta", "caja"], ["venta", "fe"], ["fe", "dian"]],
      source: `flowchart LR
  Empresa["Seleccionar empresa"] --> Productos["Seleccionar productos"]
  Productos --> Carrito["Agregar al carrito"]
  Carrito --> Descuento["Aplicar descuento o promocion"]
  Descuento --> Pago["Seleccionar medio de pago"]
  Pago --> Venta["Registrar venta"]
  Venta --> Inventario["Actualizar inventario"]
  Venta --> Caja["Caja / imprimir recibo"]
  Venta --> FE["Generar factura electronica si aplica"]
  FE --> DIAN["Enviar a DIAN"]`
    },
    "dian": {
      title: "Diagrama de flujo de facturacion electronica DIAN",
      badge: "Fiscal",
      summary: "Ordena la emision DIAN Colombia: validacion, UBL, firma, envio, acuse, CUFE/PDF y correo al cliente.",
      nodes: [
        { id: "factura", label: "Crear factura", x: 80, y: 250 },
        { id: "cliente", label: "Validar cliente\nNIT/CC, municipio, regimen", x: 300, y: 250 },
        { id: "ubl", label: "Generar XML UBL 2.1\nCUFE/CUDE + QR", x: 540, y: 250, kind: "core" },
        { id: "firma", label: "Firmar digitalmente\nXAdES", x: 780, y: 250 },
        { id: "envio", label: "Enviar SOAP/WCF a DIAN", x: 1020, y: 250, kind: "external" },
        { id: "acuse", label: "Recibir respuesta\naceptado / rechazado", x: 1260, y: 250 },
        { id: "guardar", label: "Guardar CUFE,\nTrackId y estado", x: 540, y: 430, kind: "db" },
        { id: "pdf", label: "Generar PDF / HTML", x: 780, y: 430 },
        { id: "correo", label: "Enviar al cliente", x: 1020, y: 430, kind: "external" },
        { id: "reintento", label: "Cola de reintentos\nsi falla", x: 1260, y: 430, kind: "db" }
      ],
      edges: [["factura", "cliente"], ["cliente", "ubl"], ["ubl", "firma"], ["firma", "envio"], ["envio", "acuse"], ["acuse", "guardar"], ["guardar", "pdf"], ["pdf", "correo"], ["acuse", "reintento"]],
      source: `flowchart LR
  Crear["Crear factura"] --> Cliente["Validar datos del cliente"]
  Cliente --> UBL["Generar XML UBL 2.1 con CUFE/CUDE"]
  UBL --> Firma["Firmar digitalmente XAdES"]
  Firma --> Enviar["Enviar a DIAN"]
  Enviar --> Respuesta["Recibir respuesta"]
  Respuesta --> Guardar["Guardar CUFE / TrackId / estado"]
  Guardar --> PDF["Generar PDF o representacion"]
  PDF --> Correo["Enviar al cliente"]
  Respuesta --> Reintentos["Cola de reintentos si falla"]`
    },
    "inventario": {
      title: "Diagrama de flujo de inventario",
      badge: "Stock",
      summary: "Relaciona entradas, salidas, traslados, kardex y reportes de existencias por empresa y bodega.",
      nodes: [
        { id: "compra", label: "Compra / Entrada", x: 120, y: 260 },
        { id: "aumenta", label: "Aumenta inventario", x: 360, y: 260, kind: "db" },
        { id: "venta", label: "Venta / Salida", x: 600, y: 260 },
        { id: "disminuye", label: "Disminuye inventario", x: 840, y: 260, kind: "db" },
        { id: "traslado", label: "Traslado entre bodegas", x: 1080, y: 260 },
        { id: "kardex", label: "Movimiento en Kardex", x: 600, y: 430, kind: "db" },
        { id: "reporte", label: "Reporte de existencias\nalertas y reposicion", x: 840, y: 430 }
      ],
      edges: [["compra", "aumenta"], ["aumenta", "venta"], ["venta", "disminuye"], ["disminuye", "traslado"], ["aumenta", "kardex"], ["disminuye", "kardex"], ["traslado", "kardex"], ["kardex", "reporte"]],
      source: `flowchart LR
  Compra["Compra / Entrada"] --> Aumenta["Aumenta inventario"]
  Aumenta --> Venta["Venta / Salida"]
  Venta --> Disminuye["Disminuye inventario"]
  Disminuye --> Traslado["Traslado entre bodegas"]
  Aumenta --> Kardex["Movimiento en Kardex"]
  Disminuye --> Kardex
  Traslado --> Kardex
  Kardex --> Reporte["Reporte de existencias"]`
    },
    "roles-permisos": {
      title: "Diagrama de usuarios, roles y permisos",
      badge: "Acceso",
      summary: "Ubica roles super y empresariales. La autorizacion final vive en backend, wrappers y matriz por modulo.",
      nodes: [
        { id: "super", label: "Superadmin del sistema", x: 650, y: 40, kind: "core" },
        { id: "empresa1", label: "Empresa 1", x: 360, y: 190 },
        { id: "empresa2", label: "Empresa 2", x: 940, y: 190 },
        { id: "admin1", label: "Administrador", x: 160, y: 350 },
        { id: "cajero1", label: "Cajero", x: 360, y: 350 },
        { id: "contador1", label: "Contador", x: 560, y: 350 },
        { id: "admin2", label: "Administrador", x: 760, y: 350 },
        { id: "bodega2", label: "Bodeguero", x: 960, y: 350 },
        { id: "vendedor2", label: "Vendedor", x: 1160, y: 350 },
        { id: "permisos", label: "Permisos efectivos\nC/R/U/D/A por modulo", x: 650, y: 520, kind: "db" }
      ],
      edges: [["super", "empresa1"], ["super", "empresa2"], ["empresa1", "admin1"], ["empresa1", "cajero1"], ["empresa1", "contador1"], ["empresa2", "admin2"], ["empresa2", "bodega2"], ["empresa2", "vendedor2"], ["admin1", "permisos"], ["cajero1", "permisos"], ["contador1", "permisos"], ["bodega2", "permisos"]],
      source: `flowchart TB
  Super["Superadmin del sistema"] --> E1["Empresa 1"]
  Super --> E2["Empresa 2"]
  E1 --> Admin1["Administrador"]
  E1 --> Cajero1["Cajero"]
  E1 --> Contador1["Contador"]
  E2 --> Admin2["Administrador"]
  E2 --> Bodeguero2["Bodeguero"]
  E2 --> Vendedor2["Vendedor"]
  Admin1 --> Permisos["Permisos C/R/U/D/A por modulo"]
  Cajero1 --> Permisos
  Contador1 --> Permisos
  Bodeguero2 --> Permisos`
    },
    "api-endpoints": {
      title: "Diagrama de API / endpoints",
      badge: "Backend Go",
      summary: "Organiza las rutas por familias para ubicar handlers, wrappers de permisos y validaciones de empresa_id.",
      nodes: [
        { id: "api", label: "API Go", x: 650, y: 40, kind: "core" },
        { id: "auth", label: "Auth API\n/api/login\n/api/account", x: 80, y: 190 },
        { id: "empresa", label: "Empresa API\n/super/api/empresas\n/api/empresa/*", x: 360, y: 190 },
        { id: "inventario", label: "Inventario API\nproductos, bodegas,\nmovimientos", x: 650, y: 190 },
        { id: "ventas", label: "Ventas API\ncarritos, pagos,\nclientes", x: 940, y: 190 },
        { id: "dian", label: "DIAN API\nfacturacion, pais,\nGetStatusZip", x: 1220, y: 190 },
        { id: "reportes", label: "Reportes API\nventas, caja,\nauditoria", x: 360, y: 390 },
        { id: "config", label: "Configuracion API\nroles, permisos,\nimpresoras", x: 650, y: 390 },
        { id: "super", label: "Super API\nlicencias, VPS,\nalertas, portal", x: 940, y: 390 }
      ],
      edges: [["api", "auth"], ["api", "empresa"], ["api", "inventario"], ["api", "ventas"], ["api", "dian"], ["api", "reportes"], ["api", "config"], ["api", "super"]],
      source: `flowchart TB
  API["API Go"]
  API --> Auth["Auth API: login, account, sesiones"]
  API --> Empresa["Empresa API: /api/empresa/*"]
  API --> Inventario["Inventario API: productos, bodegas, kardex"]
  API --> Ventas["Ventas API: carritos, pagos, clientes"]
  API --> DIAN["DIAN API: facturacion electronica"]
  API --> Reportes["Reportes API: ventas, caja, auditoria"]
  API --> Config["Configuracion API: roles, permisos, impresoras"]
  API --> Super["Super API: licencias, VPS, portal, alertas"]`
    },
    "despliegue": {
      title: "Diagrama de despliegue / infraestructura",
      badge: "VPS",
      summary: "Muestra dominio, proxy, Docker, base de datos, correo, backups y flujo GitHub hacia VPS.",
      nodes: [
        { id: "usuario", label: "Usuario", x: 80, y: 260 },
        { id: "dominio", label: "Dominio web\nHTTPS", x: 300, y: 260 },
        { id: "nginx", label: "Nginx / Proxy", x: 520, y: 260, kind: "infra" },
        { id: "vps", label: "Servidor VPS", x: 740, y: 260, kind: "infra" },
        { id: "docker", label: "Docker compose", x: 960, y: 260, kind: "infra" },
        { id: "app", label: "App Go", x: 1180, y: 160, kind: "core" },
        { id: "pg", label: "PostgreSQL", x: 1180, y: 300, kind: "db" },
        { id: "mailu", label: "Mailu / correo", x: 1180, y: 440, kind: "external" },
        { id: "github", label: "GitHub / rs.ps1", x: 520, y: 460 },
        { id: "backup", label: "Backups\nlocal + nube", x: 740, y: 460, kind: "db" }
      ],
      edges: [["usuario", "dominio"], ["dominio", "nginx"], ["nginx", "vps"], ["vps", "docker"], ["docker", "app"], ["docker", "pg"], ["docker", "mailu"], ["github", "vps"], ["vps", "backup"]],
      source: `flowchart LR
  Usuario --> Dominio["Dominio web HTTPS"]
  Dominio --> Nginx["Nginx / proxy"]
  Nginx --> VPS["Servidor VPS"]
  VPS --> Docker["Docker compose"]
  Docker --> App["App Go"]
  Docker --> PostgreSQL[(PostgreSQL)]
  Docker --> Mailu["Mailu / correo"]
  GitHub["GitHub / rs.ps1"] --> VPS
  VPS --> Backups["Backups locales y nube"]`
    },
    "seguridad": {
      title: "Diagrama de seguridad",
      badge: "Defensa",
      summary: "Resume autenticacion, sesiones, roles, aislamiento por empresa, auditoria, HTTPS y proteccion de secretos.",
      nodes: [
        { id: "login", label: "Login\nadmin / usuario", x: 120, y: 260 },
        { id: "sesion", label: "JWT / sesiones\ncookies seguras", x: 350, y: 260 },
        { id: "roles", label: "Roles y permisos\nwrappers backend", x: 590, y: 260, kind: "core" },
        { id: "empresa", label: "Validacion empresa_id\nIDs secundarios", x: 840, y: 260, kind: "core" },
        { id: "datos", label: "PostgreSQL\nconsultas aisladas", x: 1090, y: 260, kind: "db" },
        { id: "https", label: "HTTPS / firewall\nSSH restringido", x: 350, y: 430, kind: "infra" },
        { id: "secretos", label: "Secretos cifrados\nDIAN, SMTP, tokens", x: 590, y: 430 },
        { id: "auditoria", label: "Auditoria y logs\nacciones criticas", x: 840, y: 430, kind: "db" },
        { id: "backups", label: "Backups sin secretos", x: 1090, y: 430, kind: "db" }
      ],
      edges: [["login", "sesion"], ["sesion", "roles"], ["roles", "empresa"], ["empresa", "datos"], ["https", "sesion"], ["secretos", "roles"], ["empresa", "auditoria"], ["datos", "backups"]],
      source: `flowchart LR
  Login --> Sesion["JWT / sesiones"]
  Sesion --> Roles["Roles y permisos"]
  Roles --> Empresa["Validacion por empresa_id"]
  Empresa --> Datos[(PostgreSQL)]
  HTTPS["HTTPS / firewall / SSH"] --> Sesion
  Secretos["Secretos DIAN, SMTP, tokens"] --> Roles
  Empresa --> Auditoria["Auditoria y logs"]
  Datos --> Backups["Backups sin secretos"]`
    },
    "auditoria-logs": {
      title: "Diagrama de auditoria y logs",
      badge: "Trazabilidad",
      summary: "Estandariza quien hizo que, en que empresa, modulo, fecha, IP y con que resultado.",
      nodes: [
        { id: "accion", label: "Usuario realiza accion", x: 100, y: 260 },
        { id: "permiso", label: "Sistema valida permiso", x: 350, y: 260, kind: "core" },
        { id: "operacion", label: "Ejecuta operacion", x: 600, y: 260 },
        { id: "log", label: "Guarda log\naudit/eventos", x: 850, y: 260, kind: "db" },
        { id: "campos", label: "usuario, empresa,\naccion, fecha, IP,\nmodulo, anterior, nuevo", x: 1100, y: 260, kind: "db" },
        { id: "reporta", label: "Consulta / exporta\npara soporte y control", x: 850, y: 440 }
      ],
      edges: [["accion", "permiso"], ["permiso", "operacion"], ["operacion", "log"], ["log", "campos"], ["log", "reporta"]],
      source: `flowchart LR
  Accion["Usuario realiza accion"] --> Permiso["Sistema valida permiso"]
  Permiso --> Operacion["Ejecuta operacion"]
  Operacion --> Log["Guarda log"]
  Log --> Campos["usuario, empresa, accion, fecha, IP, modulo, dato anterior, dato nuevo"]
  Log --> Reporte["Consulta o exportacion de auditoria"]`
    },
    "reportes": {
      title: "Diagrama de reportes",
      badge: "Analitica",
      summary: "Mapa los reportes principales que consumen ventas, compras, inventario, caja, fiscalidad y utilidad.",
      nodes: [
        { id: "reportes", label: "Reportes", x: 650, y: 40, kind: "core" },
        { id: "ventas", label: "Ventas", x: 80, y: 190 },
        { id: "compras", label: "Compras", x: 300, y: 190 },
        { id: "inventario", label: "Inventario", x: 520, y: 190 },
        { id: "caja", label: "Caja / turnos", x: 740, y: 190 },
        { id: "clientes", label: "Clientes", x: 960, y: 190 },
        { id: "impuestos", label: "Impuestos", x: 1180, y: 190 },
        { id: "fe", label: "Facturacion electronica", x: 300, y: 370 },
        { id: "contabilidad", label: "Contabilidad", x: 520, y: 370 },
        { id: "utilidad", label: "Utilidad / ganancias", x: 740, y: 370 },
        { id: "productos", label: "Productos mas vendidos", x: 960, y: 370 }
      ],
      edges: [["reportes", "ventas"], ["reportes", "compras"], ["reportes", "inventario"], ["reportes", "caja"], ["reportes", "clientes"], ["reportes", "impuestos"], ["reportes", "fe"], ["reportes", "contabilidad"], ["reportes", "utilidad"], ["reportes", "productos"]],
      source: `flowchart TB
  Reportes --> Ventas
  Reportes --> Compras
  Reportes --> Inventario
  Reportes --> Caja
  Reportes --> Clientes
  Reportes --> Impuestos
  Reportes --> FE["Facturacion electronica"]
  Reportes --> Contabilidad
  Reportes --> Utilidad["Utilidad / ganancias"]
  Reportes --> Productos["Productos mas vendidos"]`
    },
    "integraciones": {
      title: "Diagrama de integraciones externas",
      badge: "Conectores",
      summary: "Lista dependencias externas y equipos que el sistema puede usar sin mezclar secretos con documentacion.",
      nodes: [
        { id: "pcs", label: "Sistema POS", x: 650, y: 50, kind: "core" },
        { id: "dian", label: "DIAN", x: 120, y: 210, kind: "external" },
        { id: "pagos", label: "Pasarela de pagos\nWompi / Epayco", x: 360, y: 210, kind: "external" },
        { id: "whatsapp", label: "WhatsApp Business", x: 600, y: 210, kind: "external" },
        { id: "smtp", label: "Correo SMTP / Mailu", x: 840, y: 210, kind: "external" },
        { id: "impresoras", label: "Impresoras POS\ncajon, datafono, QR", x: 1080, y: 210, kind: "external" },
        { id: "github", label: "GitHub", x: 480, y: 390, kind: "external" },
        { id: "backups", label: "Servidor de backups\nrclone / nube", x: 720, y: 390, kind: "external" },
        { id: "ia", label: "Proveedor IA\nuso controlado", x: 960, y: 390, kind: "external" }
      ],
      edges: [["pcs", "dian"], ["pcs", "pagos"], ["pcs", "whatsapp"], ["pcs", "smtp"], ["pcs", "impresoras"], ["pcs", "github"], ["pcs", "backups"], ["pcs", "ia"]],
      source: `flowchart TB
  PCS["Sistema POS"] --> DIAN
  PCS --> Pagos["Pasarela de pagos"]
  PCS --> WhatsApp["WhatsApp Business"]
  PCS --> SMTP["Correo SMTP / Mailu"]
  PCS --> Impresoras["Impresoras POS"]
  PCS --> Cajon["Cajon monedero"]
  PCS --> Datafono
  PCS --> QR["Codigo QR"]
  PCS --> GitHub
  PCS --> Backups["Servidor de backups"]
  PCS --> IA["Proveedor IA controlado"]`
    },
    "agentes": {
      title: "Diagrama de agentes automaticos",
      badge: "Mantenimiento",
      summary: "Ubica agentes internos de soporte, monitoreo y programacion asistida. Codex usa estos diagramas como referencia del repo, no como permiso para ejecutar acciones externas.",
      nodes: [
        { id: "agentes", label: "Agentes del sistema", x: 650, y: 40, kind: "core" },
        { id: "dian", label: "Agente DIAN\ncambios normativos", x: 120, y: 210 },
        { id: "servidor", label: "Agente servidor\nVPS, disco, servicios", x: 360, y: 210 },
        { id: "soporte", label: "Agente soporte\nrespuestas a clientes", x: 600, y: 210 },
        { id: "ventas", label: "Agente ventas\ninteresados y licencias", x: 840, y: 210 },
        { id: "programador", label: "Agente programador\nsugerencias tecnicas para Codex", x: 1080, y: 210, kind: "core" },
        { id: "secretaria", label: "Agente secretaria\nnotificaciones al administrador", x: 650, y: 390 }
      ],
      edges: [["agentes", "dian"], ["agentes", "servidor"], ["agentes", "soporte"], ["agentes", "ventas"], ["agentes", "programador"], ["agentes", "secretaria"]],
      source: `flowchart TB
  Agentes["Agentes del sistema"]
  Agentes --> DIAN["Agente DIAN: cambios normativos"]
  Agentes --> Servidor["Agente servidor: VPS, disco, servicios"]
  Agentes --> Soporte["Agente soporte: clientes"]
  Agentes --> Ventas["Agente ventas: interesados"]
  Agentes --> Programador["Agente programador: sugerencias para Codex"]
  Agentes --> Secretaria["Agente secretaria: notificaciones"]`
    }
  };

  var data = { order: order, diagrams: diagrams };
  root.PCSSuperDiagramData = data;
  if (typeof module !== "undefined" && module.exports) {
    module.exports = data;
  }
})(typeof window !== "undefined" ? window : globalThis);
