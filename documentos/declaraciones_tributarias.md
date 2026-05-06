# Declaraciones Tributarias y Motor de Impuestos Colombia

Fecha: 2026-05-06
Estado: implementado

## Alcance

El modulo `declaraciones_tributarias` centraliza por empresa la preparacion, preliquidacion y seguimiento de obligaciones tributarias colombianas. Cubre IVA, retencion en la fuente, ReteIVA, ICA, ReteICA, impuesto al consumo, renta y regimen simple, con saldos a pagar/favor, calendario, soportes y movimientos de conciliacion.

## Componentes

- API privada: `/api/empresa/declaraciones_tributarias`.
- Pantalla administrativa: `web/administrar_empresa/declaraciones_tributarias.html`.
- Menu principal: `linkDeclaracionesTributarias`.
- Menu financiero: `linkDeclaracionesTributariasMenu`.
- Permiso/licencia: `declaraciones_tributarias`.
- Base de datos: `empresa_declaraciones_tributarias`, `empresa_declaraciones_tributarias_movimientos`, `empresa_calendario_tributario`.

## Funcionalidades

- Dashboard con borradores, revisadas, vencidas, pagadas, saldo por pagar y retenciones.
- Preliquidacion automatica por tipo y periodo usando datos tributarios/contables disponibles.
- Edicion manual profesional para ajustar bases, impuestos, retenciones, sanciones, intereses, soportes y estado.
- Calendario tributario configurable por tipo, periodo, periodicidad, rango de NIT y vencimiento.
- Movimientos de conciliacion que conectan ventas, compras, retenciones y contabilidad con cada declaracion.
- Exportacion CSV desde la pantalla.
- Datos demo para validar el modulo por empresa.

## Modelo tributario

El motor calcula saldos de forma parametrizable:

- IVA: IVA generado menos IVA descontable y saldo a favor anterior.
- Retencion en la fuente: retenciones practicadas mas autorretenciones.
- ReteIVA/ReteICA/ICA: retenciones o impuesto territorial del periodo.
- Consumo: impuesto al consumo liquidado.
- Renta y regimen simple: base inicial parametrica sobre ingresos gravados, anticipos y saldos a favor.
- Sanciones e intereses se suman al valor final.

Los vencimientos se guardan en calendario editable porque dependen de NIT, regimen, periodicidad, municipio y normas vigentes.

## Referencias normativas consultadas

- DIAN - Calendario de obligaciones: https://www.dian.gov.co/Contribuyentes-Plus/Paginas/Calendario-de-obligaciones.aspx
- DIAN - Calendario Tributario 2026: https://www.dian.gov.co/Calendarios/Calendario_Tributario_2026.pdf
- DIAN - Formularios e instructivos, formularios 300/350 y relacionados: https://www.dian.gov.co/atencionciudadano/formulariosinstructivos/Paginas/default.aspx

## Gobierno y seguridad

Todas las tablas incluyen `empresa_id`; la API queda protegida por `WithEmpresaDeclaracionesTributariasPermissions` y por el techo de licencia `declaraciones_tributarias`. Las acciones de lectura usan permiso `read`, guardar calendario/declaraciones usa `create/update`, y preliquidar o cargar demo requiere permiso de aprobacion.

## Pruebas

Se agregan pruebas unitarias en `backend/db/declaraciones_tributarias_test.go` para rango de periodo, normalizacion de tipo y calculo de saldos IVA/saldo a favor.
