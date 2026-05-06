# Cierre y bloqueo contable/fiscal avanzado

Fecha: 2026-05-06
Estado: implementado

## Objetivo

El modulo `cierre_fiscal` protege la informacion ya cerrada o reportada. Centraliza periodos fiscales, reglas de bloqueo por modulo, excepciones aprobadas, reaperturas con motivo obligatorio y bitacora de intentos permitidos o bloqueados.

## Alcance funcional

- Periodos fiscales por empresa con estados `abierto`, `en_revision`, `cerrado` y `bloqueado`.
- Bloqueo configurable para ventas, compras, caja, inventario, contabilidad y facturacion.
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
