(function(root){
  root.PCSSuperTechnicalDocs = {
  "updatedAt": "2026-07-09",
  "sourceDocument": "documentos/diagramas/documentacion_tecnica_completa.md",
  "manifest": "documentos/diagramas/documentacion_tecnica_completa_manifest.json",
  "summary": "Documentacion tecnica completa con diagramas Mermaid para super administrador y Codex.",
  "diagrams": [
    {
      "id": "arquitectura_general",
      "title": "Arquitectura general",
      "type": "flowchart",
      "source": "flowchart TB\n  Users[Usuarios web: superadmin, admin empresa, cajero, contador, cliente publico] --> Browser[Navegador HTML CSS JS]\n  Browser --> Shells[Shells web: super_administrador, administrar_empresa, portal publico, POS]\n  Shells --> API[Backend Go net/http]\n  API --> Auth[Autenticacion, sesiones, roles, permisos, licencias]\n  API --> Domain[Servicios de dominio: ventas, inventario, DIAN, finanzas, IA, integraciones]\n  Domain --> DB[(PostgreSQL pcs_empresas)]\n  Domain --> Files[Uploads y archivos empresariales]\n  Domain --> External[DIAN, pasarelas, Mailu SMTP, WhatsApp, Rappi, OpenAI]\n  Ops[rs.ps1, Docker Compose, backups, VPS/VPS2] --> API\n  Ops --> DB"
    },
    {
      "id": "mapa_de_modulos",
      "title": "Mapa de modulos",
      "type": "flowchart",
      "source": "flowchart TB\n  PCS[PCS SaaS POS multiempresa]\n  PCS --> Super[Super administrador]\n  PCS --> Empresa[Administrar empresa]\n  PCS --> Publico[Portal publico y venta publica]\n  Empresa --> Ventas[Ventas POS, carritos, caja, estaciones]\n  Empresa --> Inventario[Inventario, compras, bodegas, produccion MRP, WMS]\n  Empresa --> Fiscal[Facturacion electronica, impuestos, DIAN, pais fiscal]\n  Empresa --> Finanzas[Finanzas, cartera, contabilidad, nomina]\n  Empresa --> Personas[Usuarios, roles, clientes, empleados]\n  Empresa --> Canales[Correo, WhatsApp, Rappi, portal, noticias]\n  Empresa --> IA[Chat IA, agentes, captura inteligente, grafologia]\n  Super --> Licencias[Licencias, pagos, planes, descuentos]\n  Super --> Infra[VPS, Docker, backups, VPS2, alertas]\n  Super --> Gobierno[Configuracion, auditoria, IA global, diagramas]"
    },
    {
      "id": "mapa_de_navegacion",
      "title": "Mapa de navegacion",
      "type": "flowchart",
      "source": "flowchart LR\n  Login[login.html / login_usuario.html] --> Selector[seleccionar_empresa.html]\n  Login --> SuperShell[super_administrador.html]\n  Selector --> EmpresaShell[administrar_empresa.html?id=empresa_id]\n  SuperShell --> SuperPanel[Panel super]\n  SuperShell --> SuperConfig[Configuracion]\n  SuperShell --> Diagramas[Diagramas tecnicos]\n  SuperShell --> LicenciasNav[Licencias y pagos]\n  SuperShell --> InfraNav[Plataforma e infraestructura]\n  EmpresaShell --> PanelEmpresa[Panel empresa]\n  EmpresaShell --> VentaDirecta[Carrito venta directa]\n  EmpresaShell --> ModulosEmpresa[Submenus operativos]\n  PublicIndex[index.html] --> Noticias[noticias.html]\n  PublicIndex --> VentaPublica[venta_publica.html]"
    },
    {
      "id": "erd_global_resumido_por_dominios",
      "title": "ERD global resumido por dominios",
      "type": "erDiagram",
      "source": "erDiagram\n  empresas ||--o{ users : empresa_id_logico\n  empresas ||--o{ productos : empresa_id_logico\n  empresas ||--o{ clientes : empresa_id_logico\n  empresas ||--o{ carritos_compras : empresa_id_logico\n  empresas ||--o{ bodegas : empresa_id_logico\n  empresas ||--o{ empresa_facturacion_documentos : empresa_id_logico\n  empresas ||--o{ empresa_rappi_configuracion : empresa_id_logico\n  roles_de_usuario ||--o{ users : rol_usuario_id_logico\n  clientes ||--o{ carritos_compras : cliente_id_logico\n  carritos_compras ||--o{ carrito_compra_items : carrito_id_logico\n  productos ||--o{ carrito_compra_items : producto_id_logico\n  productos ||--o{ inventario_movimientos : producto_id_logico\n  bodegas ||--o{ inventario_movimientos : bodega_id_logico\n  empresa_facturacion_documentos ||--o{ facturacion_electronica_reintentos : documento_codigo_logico\n  licencias ||--o{ pagos : licencia_id_logico"
    },
    {
      "id": "casos_de_uso",
      "title": "Casos de uso",
      "type": "flowchart",
      "source": "flowchart LR\n  Superactor[Super administrador]\n  Admin[Admin empresa]\n  Cajero[Cajero]\n  Contador[Contador / finanzas]\n  Bodeguero[Bodeguero / produccion]\n  Cliente[Cliente publico]\n  Tecnico[Tecnico soporte / operaciones]\n  Superactor --> UC1[Gestionar empresas, licencias, pagos, planes]\n  Superactor --> UC2[Auditar sistema, alertas, VPS, backups, diagramas]\n  Superactor --> UC3[Configurar IA global, correo, portal, Rappi y mantenimiento]\n  Admin --> UC4[Configurar empresa, usuarios, permisos, estaciones, impresoras]\n  Admin --> UC5[Gestionar catalogo, precios, inventario, clientes y canales]\n  Cajero --> UC6[Vender, cobrar, imprimir, cerrar caja, operar offline]\n  Contador --> UC7[Facturacion electronica, impuestos, nomina, reportes financieros]\n  Bodeguero --> UC8[Compras, bodegas, traslados, MRP, WMS, calidad]\n  Cliente --> UC9[Consultar portal, venta publica, pagar licencia o pedido]\n  Tecnico --> UC10[Monitorear servicios, snapshots, incidencias y logs saneados]"
    },
    {
      "id": "diagrama_de_clases_uml",
      "title": "Diagrama de Clases UML",
      "type": "classDiagram",
      "source": "classDiagram\n  class Empresa {+id +nombre +nit +tipo +configuracion}\n  class Usuario {+id +empresa_id +rol_usuario_id +email +estado}\n  class RolUsuario {+id +nombre +permisos}\n  class Producto {+id +empresa_id +sku +precio +stock}\n  class Cliente {+id +empresa_id +documento +contacto}\n  class Carrito {+id +empresa_id +estado +total +caja}\n  class CarritoItem {+id +carrito_id +producto_id +cantidad +precio}\n  class FacturaElectronica {+id +empresa_id +numero_legal +cufe +estado_dian}\n  class Licencia {+id +empresa_id +plan_id +estado +vigencia}\n  class RappiConfig {+empresa_id +base_url +stores +webhook}\n  class AuthService {+ResolverSesion() +ValidarPermiso()}\n  class EmpresaService {+ResolverEmpresa() +ValidarAislamiento()}\n  class VentaService {+AgregarItem() +PagarEstacion() +CerrarCaja()}\n  class InventarioService {+ReservarStock() +RegistrarKardex()}\n  class FacturacionService {+GenerarUBL() +FirmarXAdES() +EnviarDIAN()}\n  class IntegracionService {+EnviarCorreo() +ProcesarWebhook() +ConsumirProveedor()}\n  Empresa \"1\" --> \"*\" Usuario\n  RolUsuario \"1\" --> \"*\" Usuario\n  Empresa \"1\" --> \"*\" Producto\n  Empresa \"1\" --> \"*\" Cliente\n  Empresa \"1\" --> \"*\" Carrito\n  Carrito \"1\" --> \"*\" CarritoItem\n  Producto \"1\" --> \"*\" CarritoItem\n  Carrito \"1\" --> \"0..1\" FacturaElectronica\n  Empresa \"1\" --> \"*\" Licencia\n  Empresa \"1\" --> \"0..1\" RappiConfig\n  AuthService --> EmpresaService\n  VentaService --> InventarioService\n  VentaService --> FacturacionService\n  IntegracionService --> RappiConfig"
    },
    {
      "id": "login_y_resolucion_de_empresa",
      "title": "Login y resolucion de empresa",
      "type": "sequenceDiagram",
      "source": "sequenceDiagram\n  actor U as Usuario\n  participant W as Frontend\n  participant A as Auth handler Go\n  participant P as Permisos/Licencias\n  participant DB as PostgreSQL\n  U->>W: Ingresa credenciales\n  W->>A: POST login\n  A->>DB: Busca administrador/usuario\n  A->>P: Resuelve rol, empresa, permisos efectivos\n  P->>DB: Consulta acceso y licencia\n  A-->>W: Sesion segura + destino\n  W-->>U: Abre panel correspondiente"
    },
    {
      "id": "venta_pos_con_inventario_y_caja",
      "title": "Venta POS con inventario y caja",
      "type": "sequenceDiagram",
      "source": "sequenceDiagram\n  actor Cajero\n  participant POS as Carrito web\n  participant API as /api/empresa/carritos_compra\n  participant Inv as InventarioService\n  participant Caja as CajaService\n  participant DB as PostgreSQL\n  Cajero->>POS: Busca producto y agrega\n  POS->>API: POST items empresa_id\n  API->>Inv: Valida producto y stock\n  Inv->>DB: Reserva/descuenta segun regla\n  API-->>POS: Item y totales\n  Cajero->>POS: Selecciona pago\n  POS->>API: PUT pagar_estacion\n  API->>Caja: Valida caja/turno/idempotencia\n  Caja->>DB: Venta, abonos, movimientos, auditoria\n  API-->>POS: Documento imprimible"
    },
    {
      "id": "facturacion_electronica_dian",
      "title": "Facturacion electronica DIAN",
      "type": "sequenceDiagram",
      "source": "sequenceDiagram\n  participant Venta\n  participant FE as FacturacionService\n  participant DB as PostgreSQL\n  participant DIAN as DIAN SOAP/WCF\n  participant Mail as Correo/Mailu\n  Venta->>FE: Solicita factura electronica\n  FE->>DB: Lee empresa, cliente, resolucion, items\n  FE->>FE: Genera UBL, CUFE, firma XAdES\n  FE->>DIAN: SendBillSync/GetStatusZip\n  DIAN-->>FE: Acuse o rechazo\n  FE->>DB: Guarda estado, TrackId, CUFE, reintento\n  FE->>Mail: Envia representacion si aceptada/configurada"
    },
    {
      "id": "webhook_rappi_separado_por_empresa",
      "title": "Webhook Rappi separado por empresa",
      "type": "sequenceDiagram",
      "source": "sequenceDiagram\n  participant Rappi\n  participant Hook as /api/public/rappi/webhook\n  participant Svc as RappiService\n  participant DB as PostgreSQL\n  participant UI as Pagina empresa Rappi\n  Rappi->>Hook: Evento con empresa_id y firma\n  Hook->>Svc: Validar HMAC/base_url/config\n  Svc->>DB: Registrar orden/evento por empresa_id\n  UI->>Svc: Listar ordenes y acciones\n  Svc-->>UI: Estado READY/SENT/take/reject/ready"
    },
    {
      "id": "cierre_de_venta",
      "title": "Cierre de venta",
      "type": "flowchart",
      "source": "flowchart TB\n  A[Inicio venta] --> B{Caja abierta?}\n  B -- No --> C[Abrir caja autorizada]\n  B -- Si --> D[Agregar productos]\n  C --> D\n  D --> E[Validar stock/precios/cliente]\n  E --> F{Factura electronica?}\n  F -- No --> G[Registrar venta POS]\n  F -- Si --> H[Generar documento fiscal]\n  H --> I{DIAN acepta?}\n  I -- Si --> J[Guardar CUFE y enviar correo]\n  I -- No --> K[Guardar rechazo y cola/reintento]\n  G --> L[Actualizar caja, auditoria e impresion]\n  J --> L\n  K --> L"
    },
    {
      "id": "despliegue_rs",
      "title": "Despliegue rs",
      "type": "flowchart",
      "source": "flowchart TB\n  A[Ejecutar scripts rs.ps1] --> B[Preflight]\n  B --> C{Preflight OK?}\n  C -- No --> X[Detener y corregir causa]\n  C -- Si --> D[Actualizar repositorio]\n  D --> E[Sincronizar VPS]\n  E --> F[Reconstruir/reiniciar Docker]\n  F --> G[Validar salud local y publica]\n  G --> H[Reportar resultado sin secretos]"
    },
    {
      "id": "diagramas_de_estados",
      "title": "Diagramas de Estados",
      "type": "stateDiagram-v2",
      "source": "stateDiagram-v2\n  [*] --> CarritoAbierto\n  CarritoAbierto --> ConItems: agregar producto\n  ConItems --> EnPago: iniciar cobro\n  EnPago --> Cerrado: pago idempotente OK\n  EnPago --> Pendiente: offline o reintento\n  Pendiente --> Cerrado: sincronizacion OK\n  Cerrado --> [*]"
    },
    {
      "id": "diagramas_de_estados_2",
      "title": "Diagramas de Estados",
      "type": "stateDiagram-v2",
      "source": "stateDiagram-v2\n  [*] --> FacturaCreada\n  FacturaCreada --> EnviadaDIAN\n  EnviadaDIAN --> Aceptada: StatusCode 00 / IsValid true\n  EnviadaDIAN --> Rechazada: reglas DIAN\n  EnviadaDIAN --> PendienteAcuse: batch en proceso\n  Rechazada --> ReintentoProgramado\n  PendienteAcuse --> Aceptada\n  ReintentoProgramado --> EnviadaDIAN"
    },
    {
      "id": "diagramas_de_estados_3",
      "title": "Diagramas de Estados",
      "type": "stateDiagram-v2",
      "source": "stateDiagram-v2\n  [*] --> LicenciaPendiente\n  LicenciaPendiente --> Activa: pago aprobado o trial\n  Activa --> PorVencer\n  PorVencer --> Vencida\n  Vencida --> Activa: renovacion aprobada\n  Activa --> Suspendida: regla administrativa"
    },
    {
      "id": "diagrama_de_componentes",
      "title": "Diagrama de Componentes",
      "type": "flowchart",
      "source": "flowchart TB\n  subgraph Web[Frontend estatico]\n    SuperUI[Super administrador]\n    EmpresaUI[Administrar empresa]\n    POSUI[Carrito POS]\n    PublicUI[Portal publico]\n  end\n  subgraph Backend[Backend Go]\n    Handlers[Handlers HTTP]\n    Middleware[Sesion, permisos, licencia]\n    DBPkg[Paquete db]\n    Services[Servicios dominio]\n    Integrations[Clientes externos]\n  end\n  Web --> Handlers\n  Handlers --> Middleware\n  Handlers --> Services\n  Services --> DBPkg\n  DBPkg --> PG[(PostgreSQL)]\n  Integrations --> DIAN[DIAN]\n  Integrations --> Mailu[Mailu SMTP]\n  Integrations --> Pay[Wompi/Epayco]\n  Integrations --> Rappi[Rappi]\n  Integrations --> OpenAI[OpenAI]"
    },
    {
      "id": "diagrama_de_despliegue",
      "title": "Diagrama de Despliegue",
      "type": "flowchart",
      "source": "flowchart LR\n  Dev[Workspace Windows Codex] --> Git[Git/GitHub]\n  Dev --> RS[scripts/rs.ps1]\n  RS --> VPS[VPS principal]\n  VPS --> Nginx[Nginx/HTTPS]\n  VPS --> Compose[Docker Compose]\n  Compose --> BackendC[pcs-backend]\n  Compose --> FrontC[pcs-frontend/edge]\n  Compose --> PostgresC[pcs-postgres]\n  Compose --> MailuC[Mailu perfil mail]\n  VPS --> Backups[Backups/snapshots]\n  VPS2[VPS2 Nextcloud/monitoreo] -. snapshot .-> SuperUI[Panel VPS2 super]"
    },
    {
      "id": "diagrama_de_paquetes",
      "title": "Diagrama de Paquetes",
      "type": "flowchart",
      "source": "flowchart TB\n  Repo[D:/powerfulcontrolsystem]\n  Repo --> BackendPkg[backend]\n  BackendPkg --> HandlersPkg[handlers]\n  BackendPkg --> DBLayer[db]\n  BackendPkg --> Internal[internal dominio]\n  Repo --> WebPkg[web]\n  WebPkg --> AdminEmpresa[administrar_empresa pages]\n  WebPkg --> SuperPages[super pages]\n  WebPkg --> JSPkg[js compartido]\n  Repo --> ScriptsPkg[scripts despliegue y operacion]\n  Repo --> DeployPkg[deploy Docker/VPS]\n  Repo --> DocsPkg[documentos]\n  DocsPkg --> DiagramasPkg[diagramas y manifiestos]"
    },
    {
      "id": "diagrama_de_flujo_de_datos",
      "title": "Diagrama de Flujo de Datos",
      "type": "flowchart",
      "source": "flowchart LR\n  Actor[Usuario / integracion externa] --> UI[Frontend PCS]\n  UI --> API[API Go]\n  API --> Validacion[Validacion sesion rol permiso licencia empresa_id]\n  Validacion --> Escritura[Mutaciones idempotentes]\n  Validacion --> Lectura[Consultas filtradas]\n  Escritura --> DB[(PostgreSQL)]\n  Lectura --> DB\n  DB --> Reportes[Reportes, auditoria, paneles]\n  API --> Externos[DIAN, Rappi, pasarelas, correo, WhatsApp, IA]\n  Externos --> API\n  API --> Logs[Eventos saneados, alertas, buzon]"
    }
  ]
};
})(typeof window !== "undefined" ? window : globalThis);
