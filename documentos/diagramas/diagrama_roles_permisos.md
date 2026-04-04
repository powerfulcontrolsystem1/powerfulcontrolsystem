# Diagrama de roles y permisos

Fecha: 2026-04-01

```mermaid
flowchart TD
    R1[super_administrador]
    R2[administrador]
    R3[usuario_empresa]

    P1[Gestion global de empresas]
    P2[Gestion de roles y tipos]
    P3[Configuracion avanzada global]
    P4[CRUD clientes]
    P5[CRUD inventario: bodegas/productos/proveedores]
    P6[CRUD carritos e items]
    P7[Pago y cierre de carrito]
    P8[Login usuario empresa]
    P9[Confirmacion correo]
    P10[Primer ingreso: crear contrasena]

    R1 --> P1
    R1 --> P2
    R1 --> P3

    R2 --> P4
    R2 --> P5
    R2 --> P6
    R2 --> P7

    R3 --> P8
    R3 --> P9
    R3 --> P10
    R3 --> P7
```

Matriz resumida:

| Permiso | super_administrador | administrador | usuario_empresa |
|---|---|---|---|
| Gestion global del sistema | Si | No | No |
| Gestion operativa por empresa | Parcial | Si | Parcial |
| Confirmar correo | No | No | Si |
| Crear contrasena primer ingreso | No | No | Si |
| Login en login_usuario | No | No | Si |

## Actualizacion 2026-04-04 (catalogo de permisos frontend en panel empresa)

Objetivo: ocultar opciones del menu lateral de `administrar_empresa` segun rol autenticado sin reemplazar la validacion backend (el backend sigue siendo la autoridad final).

```mermaid
flowchart LR
    A[Usuario autenticado en panel empresa] --> B[GET /me]
    B --> C[Normalizar rol frontend]
    C --> D[Catalogo link a modulo y accion]
    D --> E{Permiso por rol}
    E -->|Permitido| F[Mostrar enlace]
    E -->|Denegado| G[Ocultar enlace]
    G --> H[Recalcular pagina inicial visible del iframe]
```

Cobertura aplicada:

- Archivo de catalogo: `web/js/administrar_empresa.js`.
- Fuente de rol autenticado: endpoint `GET /me`.
- Matriz de evaluacion frontend alineada a modulos y acciones de permisos (`ventas`, `inventario`, `finanzas`, `clientes`, `facturacion`, `seguridad`).
