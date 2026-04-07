# Diagrama de casos de uso

Fecha: 2026-04-07

```mermaid
flowchart LR
    A[Super Administrador]
    B[Administrador de Empresa]
    C[Usuario de Empresa]

    U1((Gestionar tipos de empresa))
    U2((Gestionar roles y tipos de usuario))
    U3((Crear y administrar empresas))
    U4((Configurar SMTP y pagos))

    U5((Gestionar clientes))
    U6((Gestionar productos y bodegas))
    U7((Gestionar proveedores y servicios))
    U8((Gestionar carritos e items))
    U9((Configurar facturacion e impresion))
    U10((Crear usuarios de empresa))
    U15((Gestionar backups empresariales))
    U16((Exportar snapshots y reportes de backup))
    U17((Restaurar backup con control de aprobacion))

    U11((Confirmar correo))
    U12((Primer ingreso: crear contrasena))
    U13((Iniciar sesion usuario empresa))
    U14((Operar carrito y cerrar pago))

    A --> U1
    A --> U2
    A --> U3
    A --> U4

    B --> U5
    B --> U6
    B --> U7
    B --> U8
    B --> U9
    B --> U10
    B --> U15
    B --> U16
    B --> U17

    C --> U11
    C --> U12
    C --> U13
    C --> U14
```

Notas:
- El usuario de empresa queda habilitado solo despues de confirmar correo y crear contrasena.
- El administrador de empresa gestiona operacion comercial y configuraciones por empresa.
- La restauracion de backup es una operacion sensible y se audita con control de permisos de aprobacion.
