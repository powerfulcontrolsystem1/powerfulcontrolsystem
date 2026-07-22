# Matriz de alcance de piloto - Plan 105

Estado: **borrador tecnico pendiente de aprobacion del usuario**.

Fecha: 2026-07-21.

Relacion: ejecutar junto con `documentos/plan_105.md`, P105-001, P105-002 y
P105-017. Esta matriz no activa, desactiva ni modifica licencias, permisos,
menus, endpoints, jobs o datos de ninguna empresa.

## Objetivo

Evitar que un despliegue de piloto se interprete como autorizacion para todos
los modulos del ERP. Cada fila requiere evidencia por SHA, entorno y empresa;
un modulo sin evidencia queda fuera del piloto de forma efectiva en interfaz,
API y worker antes de publicar.

## Empresa y roles propuestos

- Empresa piloto propuesta: `Powerful Control System` (`empresa_id=12`).
- Roles de prueba: superadministrador, administrador de empresa, cajero y rol
  restringido.
- Datos: solo productos, usuarios y operaciones QA expresamente marcados. No
  usar datos de terceros, pagos reales, facturas fiscales ni mensajeria sin
  autorizacion puntual.
- El carrito de QA abierto registrado en Plan 105 debe resolverse con la accion
  de negocio autorizada antes del piloto; no se borra directamente de base de
  datos.

## Matriz de modulos

| Dominio | Estado propuesto | Condiciones minimas antes de habilitar | Evidencia obligatoria | Reversion |
|---|---|---|---|---|
| Autenticacion, seleccion de empresa y sesiones | candidato P0 | pruebas de sesion, CSRF, expiracion, cambio de empresa y rol restringido | pruebas automatizadas y E2E staging | deshabilitar acceso de piloto y revocar sesiones |
| Usuarios, roles, permisos y licencias | candidato P0 | A/B multiempresa en API/UI, licencia y pagina/accion coherentes | matriz de roles y prueba negativa A/B | restaurar configuracion versionada de permisos |
| Productos, inventario y catalogo | candidato P0 | aislamiento A/B, stock y auditoria consistentes | E2E y consulta de DB redactada | desactivar modulo y restaurar desde backup probado |
| Carrito, venta directa y caja | candidato P0 | idempotencia, bloqueo concurrente, stock, totales y cierre controlado | pruebas de concurrencia y E2E sin pago | detener ventas y aplicar runbook de conciliacion |
| Reportes operativos basicos | candidato P1 | autorizacion, exportacion aislada, rendimiento y redaccion | A/B, exportacion y carga staging | deshabilitar exportacion/reporte afectado |
| Facturacion DIAN | fuera hasta P105-011 | DIAN acepta documento real con `GetStatusZip StatusCode=00` | evidencia fiscal redactada y conciliada | apagar emision, conservar cola/auditoria; no borrar documentos |
| Wompi, ePayco y Bre-B | fuera hasta P105-005/P105-011 | webhook firmado, replay, conciliacion e idempotencia | sandbox/autorizada y proveedor real si entra al alcance | apagar pasarela y conciliar pendientes |
| Mailu, SMTP y correo corporativo | fuera hasta P105-011 | provision, entrega, DKIM/SPF/DMARC, rebote y logs redactados | evidencia de entrega autorizada | apagar envio automatico y conservar cola |
| WhatsApp y Rappi | fuera hasta P105-011 | credenciales, webhook/replay, aislamiento y consentimiento | evidencia autorizada por proveedor | apagar integracion y procesar pendientes manualmente |
| Nextcloud, OnlyOffice y RustDesk | fuera hasta P105-011/P105-015 | storage privado A/B, restore, JWT/sesiones y consentimiento | E2E de archivo/sesion en staging | retirar acceso publico y conservar archivos segun retencion |
| IA y voz | fuera por defecto | presupuesto, privacidad, aislamiento, limites y aprobacion de excepcion Python/servicio | pruebas sin PII y evidencia de costo | feature flag por empresa en apagado |
| Taxi/GPS, domotica, motel, hotel y modulos verticales | fuera hasta pruebas por vertical | empresa piloto representativa, hardware/datos simulados y A/B | E2E por vertical y runbook | deshabilitar por licencia/permiso/configuracion |
| App movil | fuera hasta P105-020 | fuente versionada, CI, firma y pruebas | build reproducible y pruebas de dispositivo | no distribuir binario |

## Controles de activacion efectiva

Antes de declarar un dominio habilitado, Terra debe verificar en este orden:

1. La licencia contiene el modulo y el permiso de rol/empresa lo permite.
2. El frontend no expone accesos a un rol o empresa fuera del alcance.
3. El handler exige sesion, empresa validada, permiso y licencia; no confia en
   `empresa_id` suministrado por el cliente.
4. El worker no procesa jobs, cron ni reintentos del modulo fuera de alcance.
5. El proveedor externo queda apagado si no existe evidencia aprobada.
6. La reversa esta documentada y no requiere borrar registros financieros,
   fiscales o de auditoria.

No se acepta como control suficiente ocultar un boton o menu. Los endpoints,
jobs y callbacks deben rechazar el modulo fuera de alcance de forma explicita.

## Aprobaciones pendientes

| Decision | Responsable | Estado |
|---|---|---|
| Empresa piloto y ventana | Usuario responsable | Pendiente |
| Dominios P0 habilitados | Usuario responsable y responsable tecnico | Pendiente |
| Proveedores que entran al piloto | Usuario responsable | Pendiente |
| RPO/RTO y capacidad objetivo | Usuario responsable y operaciones | Pendiente |
| Autorizacion de despliegue | Usuario responsable | Pendiente |

Sin estas aprobaciones, el estado es **NO-GO** y cualquier modulo marcado
"candidato" sigue deshabilitado para trafico comercial.
