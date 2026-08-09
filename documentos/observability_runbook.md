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

Los paneles y alertas deben validarse en staging antes de usarse como evidencia
de produccion.
