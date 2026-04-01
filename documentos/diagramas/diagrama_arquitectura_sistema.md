# Diagrama de arquitectura del sistema

Fecha: 2026-04-01

```mermaid
flowchart LR
    U[Browser Web UI]
    JS[Modulos JS externos web/js]
    S[Servidor Go]
    M[Auth Middleware + JSONError + Logging]
    H1[auth_admin_handlers]
    H2[payments_handlers]
    H3[system_empresas_handlers]
    H4[handlers de negocio empresa]
    DB1[(empresas.db)]
    DB2[(superadministrador.db)]
    SMTP[SMTP Gmail]
    MP[Pasarela de pago Mercado Pago / Wompi]

    U -->|HTML/CSS| JS
    U -->|HTTP/HTTPS| S
    S --> M
    M --> H1
    M --> H2
    M --> H3
    M --> H4

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

    DB2 -->|sesiones/roles/config| M
```

Componentes:
- Frontend: paginas HTML y scripts externos en `web/` y `web/js/`.
- Backend: servidor Go con handlers segmentados por dominio en `backend/handlers/`.
- Persistencia: SQLite separada por contexto global y empresarial.
- Integraciones: SMTP para validacion de correo y pasarelas para pagos.
