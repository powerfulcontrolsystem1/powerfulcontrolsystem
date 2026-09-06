# Gestion de cobranza

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Además de simular_envio, el handler vigente permite configurar y ejecutar recordatorios por email/WhatsApp. Los envíos pueden tener efectos externos y se deduplican con registros de envío.
- dry_run cuenta candidatos sin enviar; el contador Enviadas de esa simulación no es entrega al destinatario. Una respuesta sent del adaptador tampoco demuestra lectura del mensaje.
- La cartera canónica sigue en empresa_cuentas_por_cobrar; gestiones y promesas no reemplazan un abono financiero.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Actualizacion: 2026-05-06

El modulo de Gestion de cobranza centraliza la recuperacion de cartera por empresa sin duplicar cuentas por cobrar. La fuente financiera sigue siendo `empresa_cuentas_por_cobrar`; cobranza agrega la capa operativa de priorizacion, campanas, plantillas multicanal, gestiones y promesas de pago.

## Alcance funcional

- Dashboard con saldo total, saldo vencido, mora critica, cuentas vencidas, promesas pendientes y gestiones del dia.
- Bandeja de cuentas por cobrar filtrable por cliente, documento, estado y mora minima.
- Registro de gestion por llamada, WhatsApp, email o SMS, con resultado, contacto, mensaje, proximo contacto y observaciones.
- Promesas de pago con valor, fecha, estado pendiente/cumplida/incumplida/cancelada y trazabilidad de cumplimiento.
- Campanas de cobranza preventiva, recuperacion, juridica, masiva o VIP.
- Plantillas de mensaje por canal y rango de mora.
- Simulación de gestión y recordatorios configurables por email/WhatsApp mediante adaptadores; distinguir simulación, despacho y entrega.
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

- `GET/POST/PUT action=configuracion` (alias `config`): consultar/guardar configuración de recordatorios.
- `POST/PUT action=ejecutar_recordatorios` (alias `enviar_recordatorios`): procesar recordatorios; `dry_run=true` evita envíos.


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

## Fuentes y aceptación de la revisión

[cobranza.go](../backend/handlers/cobranza.go), [cobranza.go](../backend/db/cobranza.go), [cobranza_test.go](../backend/db/cobranza_test.go), [cobranza.html](../web/administrar_empresa/cobranza.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go).

Requisitos aplicables: PCS-REQ-001 a PCS-REQ-003, PCS-REQ-009, PCS-REQ-010, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
