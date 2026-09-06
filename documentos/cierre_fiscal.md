# Cierre y bloqueo contable/fiscal avanzado

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Existe administración de períodos/políticas/excepciones y un validador explícito. La búsqueda de ValidarEmpresaCierreFiscalOperacion encuentra su uso en el handler de consulta, no una integración general en cada escritura de ventas/compras/caja.
- La sincronización de períodos contables no equivale a bloqueo transversal. Cada operación debe invocar y probar la regla antes de afirmar que el período impide modificaciones.
- El simulador es diagnóstico; no autoriza ni ejecuta una operación contable o fiscal.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-05-06

## Objetivo

El módulo `cierre_fiscal` administra reglas destinadas a proteger información cerrada; la protección efectiva depende de su integración en cada operación. Centraliza periodos fiscales, reglas de bloqueo por modulo, excepciones aprobadas, reaperturas con motivo obligatorio y bitacora de intentos permitidos o bloqueados.

## Alcance funcional

- Periodos fiscales por empresa con estados `abierto`, `en_revision`, `cerrado` y `bloqueado`.
- Políticas configurables para ventas, compras, caja, inventario, contabilidad y facturación; no se acredita su aplicación automática en todos esos módulos.
- Politicas por modulo con dias de edicion retroactiva, bloqueo automatico, excepciones y reapertura aprobada.
- Excepciones aprobadas por periodo, modulo, accion, documento y fecha de expiracion.
- Simulador de validacion para saber si una operacion queda permitida o bloqueada.
- Bitacora de validaciones, cierres, reaperturas, bloqueos y eventos post-cierre.
- Sincronizacion desde el cierre/reapertura de `contabilidad_colombia` para no crear dos verdades contables.
- Datos demo para probar periodo cerrado, periodo abierto y excepcion aprobada.

## Backend

- API privada: `/api/empresa/cierre_fiscal`
- Wrapper: `WithEmpresaCierreFiscalPermissions`
- Modulo/licencia: `cierre_fiscal`
- Paginas de permiso: `linkCierreFiscal` y `linkCierreFiscalMenu`

Acciones:

- `GET dashboard`: resumen gerencial.
- `GET politicas`: reglas por modulo.
- `GET periodos`: periodos fiscales.
- `GET excepciones`: excepciones activas/historicas.
- `GET eventos`: bitacora.
- `GET validar`: validacion de una operacion por fecha, modulo y accion.
- `POST/PUT politica`: crear o actualizar politica.
- `POST/PUT periodo`: crear o actualizar periodo.
- `POST estado_periodo`: cerrar, bloquear, revisar o reabrir con motivo.
- `POST excepcion`: crear excepcion aprobada.
- `POST seed_demo`: datos demo.

## Base de datos

- `empresa_cierre_fiscal_politicas`
- `empresa_cierre_fiscal_periodos`
- `empresa_cierre_fiscal_excepciones`
- `empresa_cierre_fiscal_eventos`

Todas las tablas incluyen `empresa_id`; la API pasa por el middleware empresarial para validar que la peticion pertenece a la empresa correcta.

## Frontend

Pantalla: `web/administrar_empresa/cierre_fiscal.html`

Ubicacion:

- Administrar empresa > Finanzas y cumplimiento > Cierre fiscal.
- Centro financiero y contable > Cierre fiscal.

La pantalla incluye KPIs, periodos, politicas, excepciones, simulador y bitacora, adaptable a modo claro/oscuro.

## Permisos

La matriz base lo trata como control financiero sensible:

- Lectura: roles operativos con lectura financiera.
- Crear/actualizar politicas y periodos: `admin_empresa` y `contabilidad`.
- Cierre, bloqueo, reapertura y excepciones: accion de aprobacion.
- `super_administrador` y `administrador_total` conservan acceso total.

## Integracion actual

`contabilidad_colombia` sincroniza el periodo fiscal al cerrar o reabrir un periodo contable. Los demas modulos pueden usar `ValidarEmpresaCierreFiscalOperacion` para bloquear operaciones por fecha y registrar el intento.

## Verificacion

- `go test ./... -count=1` ejecutado en `backend/`.
- Pruebas unitarias nuevas en `backend/db/cierre_fiscal_test.go`.

## Fuentes y aceptación de la revisión

[cierre_fiscal.go](../backend/handlers/cierre_fiscal.go), [cierre_fiscal.go](../backend/db/cierre_fiscal.go), [cierre_fiscal_test.go](../backend/db/cierre_fiscal_test.go), [cierre_fiscal.html](../web/administrar_empresa/cierre_fiscal.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
