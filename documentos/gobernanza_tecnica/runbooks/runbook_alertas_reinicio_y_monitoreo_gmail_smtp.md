# Runbook: alertas de reinicio y monitoreo Gmail SMTP

Fecha: 2026-04-18
Estado: vigente

## Sintomas cubiertos

- el backend parece reiniciarse pero no llega correo de alerta.
- llega correo de reinicio a destinatario incorrecto o no llega a ninguno.
- `Probar Gmail` falla en super o devuelve exito sin trazabilidad.
- en modo de pruebas el correo no aparece capturado.
- el sistema registra arranque, pero no diferencia reinicio inesperado de arranque normal.

## Alcance

Aplica al flujo de configuracion SMTP Gmail del panel super y al registro operativo de eventos de arranque o reinicio del servidor.

## Fuentes de evidencia

- `backend/handlers/usuarios_empresa.go`
- `backend/handlers/server_runtime_notifications.go`
- `backend/handlers/server_runtime_notifications_test.go`
- configuracion super Gmail (`gmail.smtp_email`, `gmail.smtp_app_password`, `gmail.smtp_host`, `gmail.smtp_port`)
- banderas `gmail.restart_alert_to`, `gmail.restart_alert_enabled`, `gmail.smtp_test_mode`
- tabla `super_correo_notificaciones_prueba` cuando el sistema corre en modo test
- tabla `super_servidor_eventos`
- archivo de estado runtime del servidor y logs operativos de reinicio

## Verificaciones iniciales

1. Confirmar que la configuracion Gmail exista y que el secreto SMTP se pueda descifrar si `CONFIG_ENC_KEY` esta activa.
2. Verificar si `gmail.restart_alert_enabled` esta activo.
3. Confirmar el valor real de `gmail.restart_alert_to`; si esta vacio, el sistema puede usar fallback distinto al esperado.
4. Revisar si `gmail.smtp_test_mode` esta encendido para saber si el correo saldra por SMTP real o quedara capturado en pruebas.
5. Consultar el evento mas reciente en `super_servidor_eventos` y revisar la razon inferida del arranque.

## Causas probables

- configuracion SMTP incompleta o credenciales invalidas.
- alerta de reinicio desactivada en configuracion avanzada.
- destinatario de alerta vacio o mal configurado.
- expectativa equivocada: en modo test el correo no sale realmente, se captura en `super_correo_notificaciones_prueba`.
- arranque normal confundido con reinicio inesperado o viceversa por no revisar el estado runtime previo.

## Acciones de recuperacion

1. Ejecutar `POST /super/api/config/gmail?action=test` desde la vista super o por API para validar la configuracion SMTP real.
2. Si `gmail.smtp_test_mode=1`, inspeccionar `super_correo_notificaciones_prueba` y no esperar envio externo.
3. Si el test falla, corregir primero `gmail.smtp_host`, `gmail.smtp_port`, `gmail.smtp_email` y `gmail.smtp_app_password`.
4. Confirmar que `gmail.restart_alert_enabled` este activo y que `gmail.restart_alert_to` apunte al destinatario operativo correcto.
5. Revisar `super_servidor_eventos` para confirmar si el sistema registro `startup`, `unexpected_restart` u otro motivo inferido.
6. Si el backend reinicio pero no se genero alerta, verificar si el flujo se salto el envio porque las alertas estaban deshabilitadas o porque no habia configuracion SMTP valida.
7. Si el test de Gmail funciona pero la alerta de reinicio no llega, comparar el destinatario usado por `action=test` con el destinatario configurado para reinicios.

## Validacion posterior

- `action=test` devuelve resultado consistente con el modo actual del sistema.
- en modo test, el correo queda capturado en `super_correo_notificaciones_prueba`.
- en modo real, el correo llega al destinatario definido por la configuracion.
- `super_servidor_eventos` registra arranque o reinicio con motivo coherente.
- el equipo puede distinguir entre fallo SMTP, alerta desactivada y ausencia real de reinicio inesperado.

## Notas operativas

1. `action=test` valida el canal SMTP, no reemplaza la revision del evento de runtime.
2. `smtp_test_mode` es util para pruebas automatizadas y diagnostico local; no debe confundirse con un fallo de envio real.
3. Las alertas de reinicio dependen de la configuracion super y del estado runtime previo, no solo del hecho de que exista un arranque nuevo.

## Contratos relacionados

- `documentos/gobernanza_tecnica/contratos/contrato_autenticacion_administrativa_y_usuarios_empresa.md`

## Runbooks relacionados

- `documentos/gobernanza_tecnica/runbooks/runbook_arranque_postgresql_tunel_local.md`