# Portal contador / Oficina virtual para contadores

Actualizacion: 2026-05-06

El Portal contador agrega una oficina virtual para firmas contables, contadores externos y equipos de outsourcing contable que atienden varias empresas/clientes desde el sistema. No reemplaza contabilidad Colombia ni reportes; los coordina desde una capa operativa de clientes, obligaciones, solicitudes y comunicaciones.

## Alcance funcional

- Portafolio de clientes contables por empresa.
- Dashboard con clientes activos, clientes en riesgo, obligaciones pendientes, vencimientos a 7 dias, solicitudes abiertas y comunicaciones del mes.
- Ficha de cliente con NIT, regimen, periodicidad, contador responsable, contacto, estado, riesgo y dia de cierre mensual.
- Obligaciones DIAN/contables por cliente: IVA, retencion, ICA, renta, exogena, nomina electronica, medios magneticos y cierre contable.
- Solicitudes de documentos al cliente: soportes, extractos, facturas, nomina, inventario, impuestos y contratos.
- Comunicaciones internas por cliente con canal, asunto, mensaje y trazabilidad de usuario.
- Datos demo para validar el flujo de oficina contable.
- Exportacion CSV de clientes.

## Integracion tecnica

- API empresarial: `GET/POST /api/empresa/portal_contador?empresa_id=...&action=...`.
- Pantalla: `web/administrar_empresa/portal_contador.html`.
- Menu: Centro financiero y contable.
- Permisos: modulo independiente `portal_contador`, paginas `linkPortalContador` y `linkPortalContadorMenu`, wrapper `WithEmpresaPortalContadorPermissions`.
- Tablas:
  - `empresa_portal_contador_clientes`
  - `empresa_portal_contador_obligaciones`
  - `empresa_portal_contador_solicitudes`
  - `empresa_portal_contador_comunicaciones`

## Acciones API

- `dashboard`: tablero ejecutivo.
- `clientes`: portafolio de clientes.
- `obligaciones`: vencimientos y obligaciones.
- `solicitudes`: documentos solicitados.
- `comunicaciones`: historial de mensajes.
- `cliente`: crea o actualiza cliente.
- `obligacion`: crea o actualiza obligacion.
- `solicitud`: crea o actualiza solicitud.
- `comunicacion`: registra comunicacion.
- `seed_demo`: crea datos de ejemplo.

## Separacion por empresa

Todas las tablas incluyen `empresa_id`. Cuando un cliente contable corresponde a una empresa real del sistema, puede guardarse `cliente_empresa_id`; aun asi, el acceso operativo sigue pasando por permisos y licencia del modulo `portal_contador`.

## Pruebas

- `go test ./db -run TestPortalContador -count=1`
- `go test ./... -count=1`
- QA 2026-05-06: dashboard optimizado para validar esquema una sola vez por peticion y probado con HTTP 200 en Motel Calipso (`empresa_id=7`); ver `documentos/reporte_qa_modulos_2026-05-06.md`.
