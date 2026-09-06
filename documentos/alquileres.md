# Alquileres / Renta universal de activos

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- La ruta empresarial usa el wrapper Alquileres y conserva contratos, activos, tarifas, mantenimientos y ubicaciones por empresa.
- El enlace del contrato al carrito representa integración local. Saldo cero o carrito marcado pagado no demuestra confirmación de una pasarela; verificar origen del pago y duplicación al reintentar.
- Las ubicaciones GPS almacenadas no prueban seguimiento continuo ni funcionamiento de un dispositivo físico.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-05-11

## Objetivo

Modulo empresarial para administrar alquiler de herramientas, motos, maquinaria, mobiliario, tecnologia, objetos, garantias, contratos, devoluciones, mantenimiento y seguimiento GPS bajo alcance estricto por empresa.

## Alcance funcional

- Categorias y activos alquilables con estado, sede, marca, modelo, serie, placa, deposito sugerido y reglas operativas.
- Tarifas por hora, dia, semana, mes, kilometro o evento.
- Contratos, reservas, entregas, devoluciones, garantias, kilometraje y saldo pendiente.
- Mantenimientos preventivos/correctivos y ubicaciones operativas con mapa GPS.
- Dashboard de disponibilidad, contratos vencidos, ingresos, depositos retenidos, utilizacion, activos en riesgo y vencimientos.

## Seguridad y permisos

- Clave de modulo/licencia: `alquileres`.
- Pagina empresarial: `web/administrar_empresa/alquileres.html`.
- Endpoint protegido: `/api/empresa/alquileres`.
- Wrapper: `WithEmpresaAlquileresPermissions`.
- Todas las tablas usan `empresa_id`; no hay rutas publicas ni mezcla de informacion entre empresas.

## Tablas

- `empresa_alquileres_config`
- `empresa_alquileres_categorias`
- `empresa_alquileres_activos`
- `empresa_alquileres_tarifas`
- `empresa_alquileres_contratos`
- `empresa_alquileres_mantenimientos`
- `empresa_alquileres_ubicaciones`

## Integracion con el nucleo

- Los clientes de contratos se sincronizan con `clientes` por documento, telefono o email.
- Activos y tarifas se sincronizan con `servicios` para no duplicar catalogo vendible.
- Cada contrato con valor crea o reutiliza una venta central en `carritos_compras` y un item en `carrito_compra_items`.
- Cuando el contrato queda sin saldo, el carrito se marca pagado con referencia reconciliable al contrato.

## Fuentes y aceptación de la revisión

[alquileres.go](../backend/handlers/alquileres.go), [alquileres.go](../backend/db/alquileres.go), [alquileres_test.go](../backend/db/alquileres_test.go), [alquileres.html](../web/administrar_empresa/alquileres.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go).

Requisitos aplicables: PCS-REQ-001 a PCS-REQ-004, PCS-REQ-009, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
