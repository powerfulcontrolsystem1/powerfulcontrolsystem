# Runbook de observabilidad

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Revisar solo servicios realmente desplegados; Redis/Nextcloud son perfiles o stacks separados, no dependencias universales.
- Alertas locales configuradas no prueban receptor de guardia ni entrega; ensayar la cadena de notificación con autorización.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

1. Confirmar healthchecks de backend, PostgreSQL, Redis, Nginx, correo,
   Nextcloud y servicios opcionales desde una red administrativa.
2. Revisar errores agregados, latencia, saturacion, colas y fallos de webhook.
3. Correlacionar por identificador de solicitud o evento, nunca por secreto,
   token, cookie, cuerpo de webhook ni correo completo.
4. Ante un incidente, contener el servicio afectado, revocar sesiones/tokens
   cuando aplique y conservar auditoria minimizada.
5. Escalar si se pierde aislamiento empresarial, se detecta acceso a archivos
   privados, fallan backups o se supera un limite operacional.
   Ante `PCSCarrilOperativoSaturado`, abrir `Super administrador > Capacidad de
   colas`, identificar el carril y comparar pendientes, antiguedad, empresa con
   mayor presion, CPU, memoria y conexiones. No elevar replicas, concurrencia ni
   pools simultaneamente: aislar primero el tenant ruidoso, confirmar drenaje y
   ajustar un parametro por vez conforme a
   `arquitectura/capacidad_colas_multiempresa.md`.
6. Ante `PCSSoporteIAPurgaVencida`, abrir primero el diagnostico de cuarentena
   empresarial y aplicar
   `gobernanza_tecnica/runbooks/runbook_depuracion_soportes_ia_y_cuarentena.md`.
   No eliminar filas, archivos `.purge-*` ni referencias privadas por SQL o
   comandos manuales. Las metricas son agregadas y no revelan empresa, usuario,
   nombre ni ruta del soporte.
7. Ante `PCSAntivirusSoportesSinConfigurar`, mantener las cargas cerradas y
   restaurar el endpoint clamd antes de reintentar; no cambiar temporalmente el
   modo obligatorio. Ante `PCSAntivirusSoportesNoDisponible`, verificar salud,
   red privada y firmas. Ante `PCSAntivirusSoportesOmitido`, detener el piloto y
   corregir la política antes de nuevas cargas. Ante
   `PCSAntivirusSoportesDetectoMalware`, conservar la auditoria agregada y no
   abrir, descargar ni extraer el adjunto rechazado.
8. Ante `PCSExtraccionIASoportesProveedorFallando`, conservar el soporte sin
   cambios y revisar proveedor/cuota antes de reintentar. Una
   `PCSExtraccionIASoportesRespuestaInvalida` no autoriza copiar datos manuales
   desde la respuesta rechazada. Ante
   `PCSExtraccionIASoportesPersistenciaFallida`, verificar PostgreSQL y auditoria
   antes de repetir, para evitar consumo o decisiones duplicadas.
9. Ante `PCSSoporteArchivoIntegridadFallida`, bloquear el flujo del soporte,
   conservar fila/eventos y revisar backup, volumen privado y auditoria. No
   recalcular/reemplazar el hash ni sobrescribir el archivo para silenciar la
   alerta; recuperar únicamente mediante el procedimiento autorizado. Una
   aprobación abierta queda invalidada y requiere nueva revisión explícita.

Los paneles y alertas deben validarse en staging antes de usarse como evidencia
de produccion.

## Fuentes y aceptación de la revisión

[prometheus.yml](../deploy/monitoring/prometheus.yml), [incidentes_y_continuidad.md](operacion/incidentes_y_continuidad.md).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
