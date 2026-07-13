# Checklist de validacion controlada de staging - 2026-07-13

## Precondiciones

- [ ] Base PostgreSQL anonima y efimera; sin copia productiva conectada.
- [ ] Secretos de staging inyectados fuera del repositorio.
- [ ] `PCS_ENV=staging` y almacenamiento privado aislado.
- [ ] Imagenes y Compose validados por CI antes de iniciar contenedores.
- [ ] Responsable de rollback y ventana de prueba definidos.

## Migraciones y datos

- [ ] Ejecutar simulacion TOTP/tokens y registrar solo conteos.
- [ ] Validar datos antiguos, nulos y corruptos en fixtures anonimizados.
- [ ] Ejecutar migracion, rollback y segunda corrida idempotente.
- [ ] Confirmar que no quedan secretos/TOTP/tokens en texto plano.

## Multiempresa y archivos

- [ ] Probar lectura, escritura, exportacion, descarga y borrado cruzado A/B.
- [ ] Probar `empresa_id` manipulado en URL, query, JSON, formulario y header.
- [ ] Ejecutar `migrate_private_uploads` sin `--apply` y conservar solo reporte
  agregado de archivos detectados, elegibles, rechazados y faltantes.
- [ ] Probar symlink, traversal, MIME falso, HTML/SVG activo, exceso de tamano,
  duplicados y descarga sin permiso.

## Canales externos simulados

- [ ] ePayco: firma valida/invalida, replay, cuerpo excesivo, timeout y reintento.
- [ ] Wompi: integridad valida/invalida, estado repetido y pago ya procesado.
- [ ] WebRTC: sesion, permiso, tenant, Origin, expiracion, replay e inactividad.
- [ ] Correo/WhatsApp: errores genericos al navegador y ausencia de PII en logs.

## Resiliencia y cierre

- [ ] Restaurar backup en entorno efimero y verificar conteos y relaciones.
- [ ] Medir login, productos, ventas, reportes, descargas y memoria/conexiones.
- [ ] Revisar request ID, metricas HTTP/DB, rate limiting y alertas propuestas.
- [ ] Ejecutar CI completo y obtener aprobacion independiente de la PR resultante.
