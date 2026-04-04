# Diagrama de arquitectura del sistema

Fecha: 2026-04-04

```mermaid
flowchart LR
    U[Browser Web UI]
    JS[Modulos JS externos web/js]
    S[Servidor Go]
    M[Auth Middleware + JSONError + Logging]
    P[EmpresaRolePermissions middleware]
    H1[auth_admin_handlers]
    H2[payments_handlers]
    H3[system_empresas_handlers]
    H4[handlers de negocio empresa]
    H5[chat_tareas handlers]
    H6[ubicacion_gps handlers]
    H7[finanzas handlers]
    H8[chat_con_inteligencia_artificial handlers]
    H9[ai_config handlers super]
    DB1[(empresas.db)]
    DB2[(superadministrador.db)]
    FS[(web/uploads/chat_tareas)]
    OSM[OpenStreetMap Tiles]
    SMTP[SMTP Gmail]
    MP[Pasarela de pago Mercado Pago / Wompi]
    AI[API IA externa Google Gemini]

    U -->|HTML/CSS| JS
    U -->|HTTP/HTTPS| S
    S --> M
    M --> H1
    M --> H2
    M --> H3
    M --> P
    P --> H4
    M --> H5
    M --> H6
    P --> H7
    M --> H8
    M --> H9

    U -->|Leaflet + Tiles| OSM

    H1 --> DB1
    H1 --> DB2
    H2 --> DB1
    H2 --> DB2
    H2 --> MP
    H3 --> DB1
    H3 --> DB2
    H4 --> DB1
    H4 --> DB2
    H4 --> SMTP
    H5 --> DB1
    H5 --> FS
    H6 --> DB1
    H7 --> DB1
    H8 --> DB1
    H8 --> DB2
    H9 --> DB2
    H8 --> AI

    DB2 -->|sesiones/roles/config| M
```

Componentes:
- Frontend: paginas HTML y scripts externos en `web/` y `web/js/`.
- Backend: servidor Go con handlers segmentados por dominio en `backend/handlers/`.
- Persistencia: SQLite separada por contexto global y empresarial.
- Colaboracion interna: modulo `chat_y_tareas` con adjuntos en `web/uploads/chat_tareas/` y metadatos en `empresas.db`.
- Geolocalizacion empresarial: modulo `ubicacion_gps` con mapa OpenStreetMap y almacenamiento de recorridos por `empresa_id`.
- Finanzas empresariales: modulo `finanzas` con configuracion por empresa, gestion de periodos contables (abrir/cerrar), retenciones y registro de ingresos/egresos con comprobantes.
- Chat IA empresarial: modulo `chat_con_inteligencia_artificial` con alcance por `empresa_id`, limites free-tier, auditoria de consultas/respuestas y persistencia de `modelo_preferido` por cuenta Google (`empresa_id + admin_email`), usando Google Gemini.
- Configuracion IA super: endpoint administrativo para credencial Gemini con almacenamiento seguro en `superadministrador.db`.
- Seguridad por rol/empresa: middleware de permisos empresariales para rutas criticas de ventas, inventario y finanzas antes de ejecutar handlers de negocio.
- Integraciones: SMTP para validacion de correo y pasarelas para pagos.
