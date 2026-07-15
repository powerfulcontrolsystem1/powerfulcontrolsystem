# Estado de modulos para produccion

Actualizacion: 2026-07-15.

| Area | Estado tecnico | Gate antes de trafico productivo |
|---|---|---|
| Identidad, permisos y multiempresa | Controles de sesion, roles, CSRF y aislamiento empresarial aplicados por backend. | Prueba de regresion de acceso cruzado en staging. |
| Ventas, carrito, caja e inventario | Reglas web y API movil reutilizan el mismo nucleo e idempotencia movil. | Carga concurrente y corte de caja con impresora real. |
| Facturacion DIAN | Flujo y trazabilidad permanecen separados por empresa. | Aceptacion DIAN por cada emisor/rango real. |
| Pagos y licencias | Integraciones y comprobantes con auditoria. | Confirmacion real de Wompi, ePayco y Bre-B con credenciales productivas. |
| Correo, WhatsApp y notificaciones | Configuracion por empresa y registros de entrega. | SPF/DKIM/DMARC, proveedor WhatsApp y rebotes reales. |
| Documentos, Nextcloud y OnlyOffice | Archivos privados y acceso por empresa. | Carga/descarga y limpieza externa en staging. |
| IA empresarial | Limites, roles y configuracion por empresa. | Cuotas y proveedores reales con observabilidad de costo. |
| API movil | `/api/v1/`, OpenAPI, tokens de dispositivo, idempotencia y contrato JSON. | Android/iPhone contra staging, offline y push real. |
| Plataforma | Roles API/migracion/worker, cola/outbox, pools tipados y readiness. | Restauracion de backup, carga, alertas y escalado de replicas. |

Un gate externo pendiente no se resuelve declarandolo aprobado en codigo. Debe
adjuntar evidencia en los reportes de staging antes de abrir trafico productivo.
