# Contrato tecnico: permisos_contexto y wrappers de /api/empresa

Fecha: 2026-04-18
Estado: vigente

## Alcance

Este contrato cubre la capa de autorizacion por `empresa_id` para rutas `/api/empresa/*`, el endpoint `GET /api/empresa/permisos_contexto`, los wrappers por modulo, la extraccion de alcance para rutas publicas empresariales y la interaccion entre rol base, overrides dinamicos y restricciones de licencia.

## Wrappers cubiertos

- `WithEmpresaVentasPermissions`
- `WithEmpresaInventarioPermissions`
- `WithEmpresaFinanzasPermissions`
- `WithEmpresaClientesPermissions`
- `WithEmpresaComprasPermissions`
- `WithEmpresaFacturacionPermissions`
- `WithEmpresaSeguridadPermissions`
- `WithEmpresaPublicScope`

## Rutas implicadas

### Endpoint de contexto

- `GET /api/empresa/permisos_contexto?empresa_id={id}`
- `GET /api/empresa/permisos_contexto?empresa_id={id}&include_matrix=1`

### Rutas publicas empresariales permitidas con scope minimo

- `POST /api/empresa/usuarios/login`
- `POST /api/empresa/usuarios/establecer_password`
- `POST /api/empresa/usuarios/solicitar_recuperacion_password`
- `POST /api/empresa/usuarios/restablecer_password`
- `POST /api/empresa/usuarios/cambiar_password`

### Familias protegidas por wrappers de modulo

- ventas
- inventario
- finanzas
- clientes
- compras
- facturacion
- seguridad

## Entradas obligatorias

### Para wrappers protegidos

- `empresa_id` resuelto desde query string, header `X-Empresa-ID` o payload permitido
- identidad administrativa autenticada (`adminEmail`) para wrappers de modulo

### Para `permisos_contexto`

- `empresa_id`

## Entradas opcionales

- `include_matrix=1` para exponer la matriz catalogada de roles
- evidencia de aprobacion para cambios criticos en seguridad:
  - query/body/header `aprobado_por` o `approved_by`
  - query/body/header `codigo_aprobacion` o `approval_code`
  - query/body/header `motivo_aprobacion` o `approval_reason`

## Fuentes de resolucion de `empresa_id`

Orden de prioridad:

1. query string `empresa_id`
2. header `X-Empresa-ID`
3. body JSON: `empresa_id`, `empresaId` o `empresa.id`
4. body `application/x-www-form-urlencoded`: `empresa_id`
5. body `multipart/form-data`: `empresa_id`

## Salidas y estados funcionales

### Wrappers de modulo

- `200+` cuando el handler downstream completa sin denegacion del wrapper
- `400` si falta `empresa_id` o si una accion de seguridad requiere aprobacion trazable y no se suministra
- `401` si no hay usuario administrativo autenticado valido
- `403` si la empresa esta fuera del alcance, la licencia no habilita el modulo o el rol efectivo no permite la accion
- `500` si falla la validacion de alcance, admin o licencia

### `permisos_contexto`

- `200` con:
  - `empresa_id`
  - `admin_email`
  - `rol`
  - `rol_efectivo`
  - `acciones_catalogo`
  - `modulos[]`
  - `paginas{}`
  - `resumen`
  - `licencia` cuando exista politica vigente
  - `incluye_matriz`
  - `matriz_roles[]` solo con `include_matrix=1`

## Catalogos base

### Modulos canonicos

- `ventas`
- `inventario`
- `finanzas`
- `clientes`
- `compras`
- `facturacion`
- `seguridad`

### Acciones canonicas

- `R`
- `C`
- `U`
- `D`
- `A`

### Roles canonicos

- `super_administrador`
- `admin_empresa`
- `supervisor_sucursal`
- `cajero`
- `inventario`
- `compras`
- `contabilidad`
- `auditor`

## Invariantes

1. Toda ruta registrada bajo `/api/empresa/*` debe quedar envuelta por uno de los wrappers permitidos; no puede exponerse directamente sin wrapper.
2. `WithEmpresaPublicScope` solo puede usarse en las cinco rutas publicas empresariales de autenticacion ya autorizadas.
3. Ningun wrapper protegido puede continuar si `empresa_id <= 0`.
4. Ningun wrapper protegido puede continuar sin `adminEmail` autentico distinto de vacio o `sistema`.
5. El wrapper protegido siempre valida alcance real de empresa con `CanAdminAccessEmpresaIA`; conocer el `empresa_id` no concede acceso.
6. La licencia activa vigente puede restringir modulos por `modulos_habilitados`; si el modulo no esta habilitado, el wrapper debe devolver `403` aunque el rol base normalmente lo permita.
7. `permisos_contexto` es una excepcion controlada: aun estando bajo wrapper de seguridad, la licencia no bloquea la consulta del propio endpoint aunque `seguridad` no este habilitado por licencia.
8. El rol efectivo puede diferir del rol base: si `super_rol_habilitado=1` en la licencia, `supervisor_sucursal` escala a `admin_empresa` para permisos efectivos.
9. Los overrides dinamicos por rol en `roles_de_usuario_permisos` y `roles_de_usuario_paginas_permisos` prevalecen sobre la politica base por rol.
10. La visibilidad de paginas del panel empresa debe derivarse de la matriz efectiva por modulo/accion y luego aplicar overrides por pagina.
11. Para cambios criticos del modulo `seguridad`, el sistema exige aprobacion trazable antes de ejecutar el handler downstream.
12. La evidencia de aprobacion puede venir por query, headers o body JSON, incluyendo el objeto anidado `aprobacion`.
13. Los cambios criticos de seguridad aprobados deben propagar metadata de aprobacion y quedar auditados.
14. Todo wrapper protegido debe fijar `X-Empresa-ID` en la respuesta y `X-Admin-Role` / `X-Admin-Role-Efectivo` en la solicitud reenviada.
15. Los wrappers protegidos deben registrar auditoria no bloqueante con modulo, accion, codigo HTTP y duracion.

## Politica base por rol y modulo

### Regla global

- `super_administrador` permite todo
- lectura `R` en todos los modulos para: `admin_empresa`, `supervisor_sucursal`, `cajero`, `inventario`, `compras`, `contabilidad`, `auditor`

### Politica base adicional

- `ventas`: `C/U/D/A` para `admin_empresa`, `supervisor_sucursal`, `cajero`
- `inventario`: `C/U/D/A` para `admin_empresa`, `supervisor_sucursal`, `inventario`
- `finanzas`: `C/U/A` para `admin_empresa`, `contabilidad`; `D` solo para `contabilidad`
- `clientes`: `C/U/A` para `admin_empresa`, `supervisor_sucursal`, `cajero`; `D` denegado por politica base
- `compras`: `C/U/A` para `admin_empresa`, `supervisor_sucursal`, `compras`; `D` denegado por politica base
- `facturacion`: `C/U/A` para `admin_empresa`, `cajero`; `D` denegado por politica base
- `seguridad`: `C/U/D/A` solo para `admin_empresa`

## Mapeo de acciones por wrapper

### Por metodo HTTP

- `GET|HEAD|OPTIONS -> R`
- `POST -> C`
- `PUT|PATCH -> U`
- `DELETE -> D`

### Overrides por query `action`

- ventas: `cerrar`, `reabrir`, `pagar_estacion`, `activar_estacion`, `pagar`, `suspender`, `reactivar`, `convertir_* -> A`
- finanzas: `cerrar`, `reabrir`, `aprobar`, `procesar_asientos`, `conciliar_*`, `aprobar_*`, `rechazar_* -> A`; `anular -> D`
- compras: `emitir_*`, `recepcionar_*`, `contabilizar_*`, `aprobar_*`, `rechazar_*`, `validar_documentos -> A`; `anular|cancelar -> D`
- facturacion: `emitir*`, `nota_credito`, `procesar_reintentos`, `reconciliar_estados`, `firmar_xml_real`, `enviar_documento_real`, `reconexion_dian`, `consultar_acuse_real -> A`; `anular -> D`
- seguridad: `versionar`, `restaurar|restore`, `depurar_fecha`, `sync_manual`, `rotar_credencial`, `aprobar`, `rechazar`, `vincular_nomina`, `reenviar_confirmacion -> A`

## Cambios criticos que requieren aprobacion trazable

Aplica solo al modulo `seguridad` cuando la accion efectiva es `C`, `U`, `D` o `A` y se toca:

- `/api/empresa/usuarios`, excepto `action=reenviar_confirmacion` o `action=activar`
- `/api/empresa/roles_de_usuario` en cualquier metodo distinto de `GET`

## Side effects obligatorios

- enriquecimiento del contexto con `empresaID`, `adminRole` y `adminRoleEfectivo`
- fijacion de headers operativos `X-Empresa-ID`, `X-Admin-Role`, `X-Admin-Role-Efectivo`
- auditoria no bloqueante del acceso o denegacion
- carga opcional de overrides dinamicos por rol
- carga opcional de politica vigente de licencia por empresa

## Errores de contrato esperados

- ruta `/api/empresa/*` sin wrapper: falla de politica estructural, cubierta por `main_empresa_routes_security_test.go`
- wrapper publico aplicado a una ruta no permitida: falla de politica estructural, cubierta por prueba de rutas
- rol con permiso base pero override denegado: el override debe prevalecer
- rol con permiso base sobre modulo no habilitado por licencia: debe prevalecer la licencia y devolver `403`
- intento de alta o cambio critico de seguridad sin aprobacion trazable: debe devolver `400`

## Evidencia tecnica minima

- pruebas de `backend/handlers/empresa_permisos_test.go` para politica base, overrides, `include_matrix`, restricciones por licencia y aprobacion trazable
- prueba de `backend/main_empresa_routes_security_test.go` para asegurar que todas las rutas `/api/empresa/*` usan wrappers autorizados
- pruebas de auditoria sobre wrappers criticos cuando aplique

## ADRs relacionados

- `ADR-0001-frontera-multiempresa-empresa-id.md`

## Contratos relacionados

- `documentos/gobernanza_tecnica/contratos/contrato_autenticacion_administrativa_y_usuarios_empresa.md`
- `documentos/gobernanza_tecnica/contratos/contrato_venta_publica_empresarial_por_empresa.md`