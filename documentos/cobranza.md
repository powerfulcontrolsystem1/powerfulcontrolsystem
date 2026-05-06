# Gestion de cobranza

Actualizacion: 2026-05-06

El modulo de Gestion de cobranza centraliza la recuperacion de cartera por empresa sin duplicar cuentas por cobrar. La fuente financiera sigue siendo `empresa_cuentas_por_cobrar`; cobranza agrega la capa operativa de priorizacion, campanas, plantillas multicanal, gestiones y promesas de pago.

## Alcance funcional

- Dashboard con saldo total, saldo vencido, mora critica, cuentas vencidas, promesas pendientes y gestiones del dia.
- Bandeja de cuentas por cobrar filtrable por cliente, documento, estado y mora minima.
- Registro de gestion por llamada, WhatsApp, email o SMS, con resultado, contacto, mensaje, proximo contacto y observaciones.
- Promesas de pago con valor, fecha, estado pendiente/cumplida/incumplida/cancelada y trazabilidad de cumplimiento.
- Campanas de cobranza preventiva, recuperacion, juridica, masiva o VIP.
- Plantillas de mensaje por canal y rango de mora.
- Simulacion de envio para dejar evidencia operativa sin depender todavia de un proveedor externo de SMS/email/WhatsApp.
- Exportacion CSV de cartera priorizada.
- Datos demo por empresa para validar el flujo en ambientes de prueba.

## Integracion tecnica

- API empresarial: `GET/POST /api/empresa/cobranza?empresa_id=...&action=...`.
- Pantalla administrativa: `web/administrar_empresa/cobranza.html`.
- Menu: Centro financiero y contable, debajo de Creditos y cartera.
- Permisos: modulo independiente `cobranza`, pagina `linkCobranza`/`linkCobranzaMenu` y wrapper `WithEmpresaCobranzaPermissions`.
- Tablas nuevas por empresa:
  - `empresa_cobranza_plantillas`
  - `empresa_cobranza_campanas`
  - `empresa_cobranza_gestiones`
  - `empresa_cobranza_promesas`
- Tabla reutilizada: `empresa_cuentas_por_cobrar`.

## Acciones API

- `dashboard`: resumen ejecutivo.
- `cuentas`: cartera abierta filtrable.
- `plantillas`: listado de plantillas.
- `campanas`: listado de campanas.
- `gestiones`: historial de gestiones.
- `promesas`: listado de promesas por estado.
- `plantilla`: crea o actualiza plantilla.
- `campana`: crea o actualiza campana.
- `gestion`: registra gestion y puede crear promesa automaticamente.
- `promesa`: crea o actualiza promesa manual.
- `marcar_promesa`: marca promesa como cumplida o incumplida.
- `simular_envio`: registra evidencia de envio simulado.
- `seed_demo`: crea cartera, plantillas y campana de ejemplo si la empresa no tiene datos base.

## Separacion por empresa

Todas las consultas y escrituras filtran `empresa_id`. Las gestiones no crean una cartera paralela: referencian las cuentas por cobrar existentes mediante `cuenta_id` cuando aplica, y copian cliente/documento solo como snapshot operativo para auditoria.

## Pruebas

- `go test ./db -run TestCobranza -count=1`
- `go test ./... -count=1`
- QA 2026-05-06: dashboard optimizado para validar esquema una sola vez por peticion y probado con HTTP 200 en Motel Calipso (`empresa_id=7`); ver `documentos/reporte_qa_modulos_2026-05-06.md`.
