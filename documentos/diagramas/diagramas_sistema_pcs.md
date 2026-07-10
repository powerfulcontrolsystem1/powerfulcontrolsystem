# Diagramas del sistema PCS para super administrador y Codex

Actualizacion: 2026-07-07

Este documento concentra los 15 diagramas solicitados para que el super administrador los vea en paginas propias y para que Codex pueda leer la fuente tecnica en texto. La fuente visual del panel vive en `web/js/super_diagramas_data.js` y cada pagina estatica esta en `web/super/diagramas/`.

Para documentacion tecnica ampliada y escalable, usar tambien
`documentos/diagramas/documentacion_tecnica_completa.md`. Ese paquete agrega
ERD PostgreSQL extraido del backend con catalogo completo de tablas y atributos,
casos de uso, clases UML, secuencias, actividades, estados, componentes,
despliegue, paquetes, mapa de navegacion y flujo de datos. Su manifiesto para
Codex vive en `documentos/diagramas/documentacion_tecnica_completa_manifest.json`.

## Indice

- [Diagrama de modulos del sistema](#modulos) -> `/super/diagramas/modulos.html`
- [Diagrama de base de datos / ERD](#erd) -> `/super/diagramas/base_de_datos_erd.html`
- [Diagrama multiempresa / multi-tenant](#multiempresa) -> `/super/diagramas/multiempresa.html`
- [Diagrama general de arquitectura](#arquitectura) -> `/super/diagramas/arquitectura.html`
- [Diagrama de flujo de ventas POS](#ventas-pos) -> `/super/diagramas/ventas_pos.html`
- [Diagrama de flujo de facturacion electronica DIAN](#dian) -> `/super/diagramas/facturacion_dian.html`
- [Diagrama de flujo de inventario](#inventario) -> `/super/diagramas/inventario.html`
- [Diagrama de usuarios, roles y permisos](#roles-permisos) -> `/super/diagramas/roles_permisos.html`
- [Diagrama de API / endpoints](#api-endpoints) -> `/super/diagramas/api_endpoints.html`
- [Diagrama de despliegue / infraestructura](#despliegue) -> `/super/diagramas/despliegue.html`
- [Diagrama de seguridad](#seguridad) -> `/super/diagramas/seguridad.html`
- [Diagrama de auditoria y logs](#auditoria-logs) -> `/super/diagramas/auditoria_logs.html`
- [Diagrama de reportes](#reportes) -> `/super/diagramas/reportes.html`
- [Diagrama de integraciones externas](#integraciones) -> `/super/diagramas/integraciones.html`
- [Diagrama de agentes automaticos](#agentes) -> `/super/diagramas/agentes_automaticos.html`

## Diagrama de modulos del sistema {#modulos}

Agrupa las areas principales del POS multiempresa para ubicar cambios, permisos y documentacion por modulo.

```mermaid
flowchart TB
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
  Finanzas --> Agentes["Agentes automaticos y mantenimiento"]
```

## Diagrama de base de datos / ERD {#erd}

Resume las tablas principales y relaciones por dominio. El detalle fisico completo sigue en documentos/estructura_bd.md.

```mermaid
erDiagram
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
  EMPRESAS ||--o{ LICENCIAS : empresa_id
```

## Diagrama multiempresa / multi-tenant {#multiempresa}

Muestra como cada tabla operativa debe filtrar por empresa_id y como el super administrador gobierna configuraciones globales.

```mermaid
flowchart TB
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
  class D1,D2,DN rule
```

## Diagrama general de arquitectura {#arquitectura}

Conecta navegador, frontend estatico, API Go, PostgreSQL, servicios externos y despliegue Docker/VPS.

```mermaid
flowchart LR
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
  VPS --> DB
```

## Diagrama de flujo de ventas POS {#ventas-pos}

Describe el flujo operativo desde seleccion de empresa hasta inventario, caja, impresion y factura electronica si aplica.

```mermaid
flowchart LR
  Empresa["Seleccionar empresa"] --> Productos["Seleccionar productos"]
  Productos --> Carrito["Agregar al carrito"]
  Carrito --> Descuento["Aplicar descuento o promocion"]
  Descuento --> Pago["Seleccionar medio de pago"]
  Pago --> Venta["Registrar venta"]
  Venta --> Inventario["Actualizar inventario"]
  Venta --> Caja["Caja / imprimir recibo"]
  Venta --> FE["Generar factura electronica si aplica"]
  FE --> DIAN["Enviar a DIAN"]
```

## Diagrama de flujo de facturacion electronica DIAN {#dian}

Ordena la emision DIAN Colombia: validacion, UBL, firma, envio, acuse, CUFE/PDF y correo al cliente.

```mermaid
flowchart LR
  Crear["Crear factura"] --> Cliente["Validar datos del cliente"]
  Cliente --> UBL["Generar XML UBL 2.1 con CUFE/CUDE"]
  UBL --> Firma["Firmar digitalmente XAdES"]
  Firma --> Enviar["Enviar a DIAN"]
  Enviar --> Respuesta["Recibir respuesta"]
  Respuesta --> Guardar["Guardar CUFE / TrackId / estado"]
  Guardar --> PDF["Generar PDF o representacion"]
  PDF --> Correo["Enviar al cliente"]
  Respuesta --> Reintentos["Cola de reintentos si falla"]
```

## Diagrama de flujo de inventario {#inventario}

Relaciona entradas, salidas, traslados, kardex y reportes de existencias por empresa y bodega.

```mermaid
flowchart LR
  Compra["Compra / Entrada"] --> Aumenta["Aumenta inventario"]
  Aumenta --> Venta["Venta / Salida"]
  Venta --> Disminuye["Disminuye inventario"]
  Disminuye --> Traslado["Traslado entre bodegas"]
  Aumenta --> Kardex["Movimiento en Kardex"]
  Disminuye --> Kardex
  Traslado --> Kardex
  Kardex --> Reporte["Reporte de existencias"]
```

## Diagrama de usuarios, roles y permisos {#roles-permisos}

Ubica roles super y empresariales. La autorizacion final vive en backend, wrappers y matriz por modulo.

```mermaid
flowchart TB
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
  Bodeguero2 --> Permisos
```

## Diagrama de API / endpoints {#api-endpoints}

Organiza las rutas por familias para ubicar handlers, wrappers de permisos y validaciones de empresa_id.

```mermaid
flowchart TB
  API["API Go"]
  API --> Auth["Auth API: login, account, sesiones"]
  API --> Empresa["Empresa API: /api/empresa/*"]
  API --> Inventario["Inventario API: productos, bodegas, kardex"]
  API --> Ventas["Ventas API: carritos, pagos, clientes"]
  API --> DIAN["DIAN API: facturacion electronica"]
  API --> Reportes["Reportes API: ventas, caja, auditoria"]
  API --> Config["Configuracion API: roles, permisos, impresoras"]
  API --> Super["Super API: licencias, VPS, portal, alertas"]
```

## Diagrama de despliegue / infraestructura {#despliegue}

Muestra dominio, proxy, Docker, base de datos, correo, backups y flujo GitHub hacia VPS.

```mermaid
flowchart LR
  Usuario --> Dominio["Dominio web HTTPS"]
  Dominio --> Nginx["Nginx / proxy"]
  Nginx --> VPS["Servidor VPS"]
  VPS --> Docker["Docker compose"]
  Docker --> App["App Go"]
  Docker --> PostgreSQL[(PostgreSQL)]
  Docker --> Mailu["Mailu / correo"]
  GitHub["GitHub / rs.ps1"] --> VPS
  VPS --> Backups["Backups locales y nube"]
```

## Diagrama de seguridad {#seguridad}

Resume autenticacion, sesiones, roles, aislamiento por empresa, auditoria, HTTPS y proteccion de secretos.

```mermaid
flowchart LR
  Login --> Sesion["JWT / sesiones"]
  Sesion --> Roles["Roles y permisos"]
  Roles --> Empresa["Validacion por empresa_id"]
  Empresa --> Datos[(PostgreSQL)]
  HTTPS["HTTPS / firewall / SSH"] --> Sesion
  Secretos["Secretos DIAN, SMTP, tokens"] --> Roles
  Empresa --> Auditoria["Auditoria y logs"]
  Datos --> Backups["Backups sin secretos"]
```

## Diagrama de auditoria y logs {#auditoria-logs}

Estandariza quien hizo que, en que empresa, modulo, fecha, IP y con que resultado.

```mermaid
flowchart LR
  Accion["Usuario realiza accion"] --> Permiso["Sistema valida permiso"]
  Permiso --> Operacion["Ejecuta operacion"]
  Operacion --> Log["Guarda log"]
  Log --> Campos["usuario, empresa, accion, fecha, IP, modulo, dato anterior, dato nuevo"]
  Log --> Reporte["Consulta o exportacion de auditoria"]
```

## Diagrama de reportes {#reportes}

Mapa los reportes principales que consumen ventas, compras, inventario, caja, fiscalidad y utilidad.

```mermaid
flowchart TB
  Reportes --> Ventas
  Reportes --> Compras
  Reportes --> Inventario
  Reportes --> Caja
  Reportes --> Clientes
  Reportes --> Impuestos
  Reportes --> FE["Facturacion electronica"]
  Reportes --> Contabilidad
  Reportes --> Utilidad["Utilidad / ganancias"]
  Reportes --> Productos["Productos mas vendidos"]
```

## Diagrama de integraciones externas {#integraciones}

Lista dependencias externas y equipos que el sistema puede usar sin mezclar secretos con documentacion.

```mermaid
flowchart TB
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
  PCS --> IA["Proveedor IA controlado"]
```

## Diagrama de agentes automaticos {#agentes}

Ubica agentes internos de soporte, monitoreo y programacion asistida. Codex usa estos diagramas como referencia del repo, no como permiso para ejecutar acciones externas.

```mermaid
flowchart TB
  Agentes["Agentes del sistema"]
  Agentes --> DIAN["Agente DIAN: cambios normativos"]
  Agentes --> Servidor["Agente servidor: VPS, disco, servicios"]
  Agentes --> Soporte["Agente soporte: clientes"]
  Agentes --> Ventas["Agente ventas: interesados"]
  Agentes --> Programador["Agente programador: sugerencias para Codex"]
  Agentes --> Secretaria["Agente secretaria: notificaciones"]
```
