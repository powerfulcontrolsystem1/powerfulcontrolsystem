# Checklist de release (pre-despliegue)

Fecha de referencia: 2026-04-01

## 1) Respaldo y estado del entorno

- [ ] Crear backup de `backend/db/empresas.db`.
- [ ] Crear backup de `backend/db/superadministrador.db`.
- [ ] Verificar que no hay procesos antiguos ocupando el puerto de servicio (`8080` en local).
- [ ] Confirmar variables sensibles por entorno (`GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `CONFIG_ENC_KEY`, SMTP, pasarelas).

## 2) Validacion tecnica minima

- [ ] Ejecutar script operativo del punto 13:
	- `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\validar_punto_13.ps1`
- [ ] Verificar reporte tecnico generado:
	- `documentos/punto_13_validacion_integral_resultado.md`
- [ ] Ejecutar pruebas backend productivas:
	- `go test ./auth ./db ./handlers ./metrics ./utils`
- [ ] Verificar compilacion del backend sin errores.
- [ ] Confirmar que no hay lecturas de secretos desde documentos en el arranque.

## 3) Smoke tests funcionales (manuales)

- [ ] Rutas publicas responden correctamente:
	- `/`
	- `/login.html`
	- `/login_usuario.html`
- [ ] Rutas protegidas bloquean acceso sin sesion (401/403 segun corresponda).
- [ ] Login Google admin redirige correctamente.
- [ ] Flujo usuario empresa:
	- confirmacion de correo,
	- primer ingreso y creacion de contrasena,
	- login posterior con `email + password`.
- [ ] Flujo de carrito:
	- crear carrito,
	- agregar item,
	- validar recálculo de totales.

## 4) Pagos y webhooks (si aplica en la iteracion)

- [ ] Validar configuracion de Mercado Pago (modo sandbox/productivo correcto).
- [ ] Validar configuracion de Wompi/Nequi (modo correcto y llaves vigentes).
- [ ] Validar URL de webhook con plantilla segura (sin almacenar URL publica real en docs).

## 5) Seguridad y logs

- [ ] Revisar logs recientes para evitar exposición de secretos, tokens o URLs sensibles.
- [ ] Confirmar que documentacion no contiene secretos en texto plano.
- [ ] Confirmar que los archivos de referencia de túnel (`last_ngrok_url.md`) usan solo plantillas.

## 6) Documentacion y trazabilidad

- [ ] Adjuntar/actualizar evidencia del punto 13:
	- `documentos/punto_13_validacion_integral_resultado.md`
- [ ] Actualizar guia de calidad/UAT/despliegue si cambia el flujo:
	- `documentos/punto_13_calidad_uat_despliegue.md`
- [ ] Actualizar `documentos/descripcion_de_archivos` con archivos nuevos o modificados.
- [ ] Actualizar `documentos/historial_de_cambios` con resumen y fecha.
- [ ] Si cambia arquitectura/flujo, actualizar `documentos/diagramas/`.

## 7) Rollback rapido

- [ ] Tener backups DB listos para restauracion inmediata.
- [ ] Tener hash/commit de referencia previo al release.
- [ ] Definir comando/paso de reversión para cambios críticos.
