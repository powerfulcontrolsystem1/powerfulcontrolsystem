# Diagrama de flujo de procesos

Fecha: 2026-04-01

```mermaid
flowchart TD
    A[Administrador crea usuario de empresa] --> B[Generar token de confirmacion]
    B --> C[Enviar correo SMTP]
    C -->|Falla| D[Rollback: usuario no se registra]
    C -->|Exito| E[Usuario abre enlace de confirmacion]
    E --> F[Correo confirmado: estado activo]
    F --> G[Usuario abre login_usuario]
    G --> H{Tiene contrasena?}
    H -->|No| I[Crear contrasena primer ingreso]
    I --> J[Guardar hash/salt]
    H -->|Si| K[Autenticacion email+password]
    J --> K
    K --> L[Sesion iniciada]

    L --> L0[Abrir administrar_empresa]
    L0 --> L1[Cargar Inicio por defecto en Panel de la Empresa]

    L --> M[Crear cliente de venta]
    L --> N[Crear bodega y proveedor]
    M --> O[Crear carrito]
    N --> O
    O --> P[Agregar items al carrito]
    P --> Q[Calcular totales]
    Q --> R[Pagar carrito]
    R --> S[Cerrar carrito y guardar resumen de pago]

    L --> T[Administrador abre modulo chat_y_tareas]
    T --> U[Crear conversacion por empresa]
    U --> V[Agregar participantes de la empresa]
    V --> W[Intercambiar mensajes y adjuntos foto/voz]
    W --> X[Crear tareas vinculadas a la conversacion]
    X --> Y[Actualizar avance: pendiente/en_progreso/completada]

    L --> Z[Entrar a modulo facturacion electronica]
    Z --> Z1[Detectar pais automaticamente tz/lang/config empresa]
    Z1 --> Z2[Mostrar bandera del pais detectado en menu flotante]
    Z2 --> Z3[Configurar parametros FE por pais CO/PA/EC]
    Z3 --> Z4[Guardar configuracion por empresa y pais]
```

Resultado esperado:
- Flujo completo desde onboarding de usuario hasta cierre de venta con carrito pagado.
- Flujo colaborativo interno por empresa para comunicacion operativa y seguimiento de tareas.
- Al abrir el Panel de la Empresa, la subpagina inicial predeterminada es Inicio.
- En `super/licencias_resumen`, el conteo refleja solo licencias activas asignadas a empresa.
- En `seleccionar_empresa`, la seccion de licencias se filtra por empresas creadas por el usuario autenticado.
- El sistema detecta país para facturación electrónica y muestra su bandera en el menú flotante.
