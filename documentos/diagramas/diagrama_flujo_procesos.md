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
    L1 --> L11[Usuario navega a una subpagina del panel]
    L11 --> L12[Guardar subpagina actual en sessionStorage]
    L12 --> L13[Al presionar F5, restaurar la misma subpagina]

    L --> S0[Configurar colores de estado del carrito en Configuración de empresa]
    S0 --> S1[Guardar color activo/inactivo en configuración avanzada]
    S1 --> S2[Sincronizar estaciones con carritos en estado inactivo/cerrado]
    S2 --> S3[Abrir módulo estaciones]
    S3 --> S4[Tarjetas inician inactivas]
    S4 --> S5[Usuario selecciona estación]
    S5 --> S6[Activar carrito de estación y registrar activado_en]
    S6 --> S7[Tarjeta activa muestra color configurado y fecha/hora de entrada]
    S7 --> S8[Finalizar compra en carrito de estación]
    S8 --> S9[Marcar carrito inactivo/cerrado]
    S9 --> S10[Tarjeta vuelve a estado inactivo y oculta fecha/hora]

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
    Z3 --> Z31[Si no existe configuracion FE, prellenar desde configuracion avanzada]
    Z31 --> Z4[Guardar configuracion por empresa y pais]
```

Resultado esperado:
- Flujo completo desde onboarding de usuario hasta cierre de venta con carrito pagado.
- Flujo colaborativo interno por empresa para comunicacion operativa y seguimiento de tareas.
- Al abrir el Panel de la Empresa, la subpagina inicial predeterminada es Inicio.
- En `super/licencias_resumen`, el conteo refleja solo licencias activas asignadas a empresa.
- En `seleccionar_empresa`, la seccion de licencias se filtra por empresas creadas por el usuario autenticado.
- El sistema detecta país para facturación electrónica y muestra su bandera en el menú flotante.
- En `estaciones`, los carritos de estación inician inactivos, se activan al seleccionar la estación y vuelven a inactivos al finalizar la compra.
- En `estaciones`, la tarjeta activa muestra fecha y hora de entrada (`activado_en`), y las inactivas no muestran esa marca.
- En `administrar_empresa`, `super_administrador` y `seleccionar_empresa`, al recargar con F5 se restaura la subpagina/vista que estaba abierta.
