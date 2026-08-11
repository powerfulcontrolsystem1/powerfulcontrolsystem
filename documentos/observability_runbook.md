# Runbook de observabilidad

1. Confirmar healthchecks de backend, PostgreSQL, Redis, Nginx, correo,
   Nextcloud y servicios opcionales desde una red administrativa.
2. Revisar errores agregados, latencia, saturacion, colas y fallos de webhook.
3. Correlacionar por identificador de solicitud o evento, nunca por secreto,
   token, cookie, cuerpo de webhook ni correo completo.
4. Ante un incidente, contener el servicio afectado, revocar sesiones/tokens
   cuando aplique y conservar auditoria minimizada.
5. Escalar si se pierde aislamiento empresarial, se detecta acceso a archivos
   privados, fallan backups o se supera un limite operacional.
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
