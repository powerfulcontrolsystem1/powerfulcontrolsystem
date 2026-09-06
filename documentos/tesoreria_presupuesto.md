# Tesoreria y presupuesto

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Tesorería conserva cuentas, presupuestos, partidas y proyección de flujo; no equivale a saldo bancario confirmado ni orden de pago ejecutada.
- Las relaciones cuenta/presupuesto y mutaciones recibieron endurecimiento local; validar concurrencia y aislamiento PostgreSQL antes de aceptar el módulo.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-05-06

## Alcance

El modulo `tesoreria_presupuesto` agrega gestion profesional de tesoreria, bancos/caja, presupuesto y flujo de caja por empresa. Complementa `finanzas`, `corte de caja`, `creditos/cartera` y `conciliacion bancaria` sin duplicar esos flujos: este modulo se enfoca en planeacion, disponibilidad y control presupuestal.

## Superficies

- Administracion: `web/administrar_empresa/tesoreria_presupuesto.html`
- API privada: `/api/empresa/tesoreria_presupuesto`
- Menu: Centro financiero y contable > Tesoreria y presupuesto
- Permisos: `WithEmpresaTesoreriaPresupuestoPermissions`
- Licencia: `tesoreria_presupuesto`

## Funcionalidad incluida

- Configuracion por empresa: moneda, periodo de trabajo, metodo de proyeccion, alertas de saldo minimo y aprobacion de pagos.
- Cuentas de tesoreria: banco, caja, pasarela, fiducia u otro, con entidad, numero, saldo inicial, saldo actual, saldo minimo y responsable.
- Presupuestos: periodo, escenario, ingresos meta, egresos meta, saldo inicial, estado y responsable.
- Partidas presupuestales: ingresos/egresos, categoria, concepto, valor presupuestado, valor ejecutado, periodicidad y centro de costo.
- Flujo de caja: proyecciones por fecha y periodo, origen, cuenta, presupuesto, valor proyectado, ejecutado y estado.
- Generacion de flujo desde presupuesto aprobado.
- Dashboard: cuentas activas, presupuestos activos, saldo disponible, ingresos/egresos proyectados, flujo neto y ejecucion.
- Datos demo únicamente para entorno aislado y autorizado, sin sustituir movimientos comerciales reales.

## Tablas

- `empresa_tesoreria_config`
- `empresa_tesoreria_cuentas`
- `empresa_tesoreria_presupuestos`
- `empresa_tesoreria_partidas`
- `empresa_tesoreria_flujo_caja`

## Separacion multiempresa

Todas las tablas incluyen `empresa_id` y la API privada pasa por el wrapper de empresa. La licencia `tesoreria_presupuesto` permite activar o desactivar el modulo sin afectar finanzas operativas ni contabilidad.

## Siguiente integracion recomendada

En una fase posterior, los movimientos financieros reales, documentos por pagar/cobrar, extractos bancarios y pagos aprobados pueden alimentar automaticamente la ejecucion de partidas y el flujo real, manteniendo este modulo como capa de planeacion y control.

## Fuentes y aceptación de la revisión

[tesoreria_presupuesto.go](../backend/handlers/tesoreria_presupuesto.go), [tesoreria_presupuesto.go](../backend/db/tesoreria_presupuesto.go), [tesoreria_presupuesto_test.go](../backend/db/tesoreria_presupuesto_test.go), [tesoreria_presupuesto.html](../web/administrar_empresa/tesoreria_presupuesto.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go), [cierre_reparaciones_produccion_seguridad_2026-09-05.md](cierre_reparaciones_produccion_seguridad_2026-09-05.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
