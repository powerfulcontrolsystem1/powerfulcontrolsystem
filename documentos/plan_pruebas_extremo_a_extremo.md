# Plan de pruebas: Módulo a módulo y extremo a extremo

**Propósito:** validar cada módulo con datos simulados, ejecutar pruebas unitarias y flujos E2E críticos, y documentar/corregir errores detectados.

## Prerrequisitos
- Entorno Windows con `go` en PATH (compilación/ejecución de tests).
- (Opcional) `sqlite3.exe` ubicado en la raíz del repositorio para importar `scripts/seed_data.sql`.
- Ejecutar el servidor local cuando se prueben handlers/E2E: `scripts/iniciar_servidor.ps1` o `cd backend; go run main.go`.
- Herramientas: PowerShell (scripts incluidos), `curl` o `Invoke-RestMethod` para llamadas HTTP.

## Resumen del enfoque
1) Preparar entorno y datos de prueba (semillas).
2) Ejecutar tests unitarios por módulo (`db`, `auth`, `handlers`, `utils`, `metrics`).
3) Ejecutar tests de handlers / endpoints (requiere servidor en ejecución si son de integración).
4) Ejecutar pruebas E2E para flujos críticos (p.ej. venta con código de descuento).
5) Registrar fallos, corregir, re-ejecutar pruebas y documentar cambios.

## Comandos rápidos
- Importar seed + ejecutar tests por módulo (usa `sqlite3.exe` si existe):
  - `.\scripts\run_tests_by_module.ps1 -Seed`
- Ejecutar tests individuales (ejemplos):
  - `cd backend`
  - `go test ./db -v`
  - `go test ./handlers -v`
  - `go test ./... -v`

## Datos de prueba (seed)
- Archivo de ejemplo: `scripts/seed_data.sql`. Modificar según necesites para cubrir más módulos (productos, carritos, usuarios, etc.).
- Si necesitas seeds adicionales, crea `scripts/seed_extra_*.sql` y ejecuta manualmente con `sqlite3.exe testdata/seed.db ".read 'scripts/seed_extra_x.sql'"`.

## Tests de handlers / endpoints
- Inicia servidor local antes de pruebas de integración: `scripts/iniciar_servidor.ps1` o `cd backend; go run main.go`.
- Ejemplo: crear un código de descuento via API (ajusta ruta/payload si tu proyecto difiere):
```
curl -s -X POST "http://localhost:8080/api/empresa/codigos_de_descuento" \
  -H "Content-Type: application/json" \
  -d '{"empresa_id":1,"codigo":"","tipo_descuento":"valor_fijo","valor":5000,"usuario_creador":"test"}'
```
- Listar códigos:
```
curl "http://localhost:8080/api/empresa/codigos_de_descuento?empresa_id=1&include_inactive=1"
```

## Pruebas E2E recomendadas (flujos críticos)
- **Flujo Venta con descuento (pasos):**
  1. Crear empresa (si no existe).
  2. Crear producto (o usar seed).
  3. Crear carrito y agregar item(s).
  4. Crear/Generar código de descuento (UI o API).
  5. Aplicar código al carrito y validar `valor_descuento`.
  6. Pagar carrito (simular método) y verificar que `usos_actuales` del código incrementa y que existe registro en `codigos_descuento_redenciones`.

## Registro y diagnóstico
- Logs de backend: revisa `backend/logs/` y `logs/` del repo para errores y trazas.
- Para capturar fallo en un test especifico: `go test ./db -run TestNombre -v` y copia la salida.
- Si los handlers fallan en integración, inspecciona la salida del servidor (`stdout/stderr`) y los archivos en `backend/logs/`.

## Corrección de errores y ciclo de validación
1. Reproducir localmente el fallo con el seed y el test afectado.
2. Arreglar la función/handler y ejecutar únicamente el test afectado.
3. Ejecutar la batería completa de tests del módulo y los handlers relacionados.
4. Actualizar `documentos/historial_de_cambios` y `documentos/descripcion_de_archivos` con resumen de la corrección.
5. Preparar commit/PR con descripcion clara y pasos para reproducir el error y la solución.

## Plantilla mínima para reportar un bug/fallo
- Descripción corta: 1 línea.
- Pasos para reproducir: comandos y seed exacto.
- Resultado esperado vs. resultado actual.
- Logs relevantes (adjuntar).
- Tests fallidos (nombre y salida completa).

## Notas finales
- Si quieres que ejecute la batería completa de tests aquí, puedo hacerlo (avísame).
- Si prefieres usar tu script de despliegue/commit, modifica `scripts/seed_data.sql` y ejecuta `scripts/run_tests_by_module.ps1 -Seed` localmente.

Fin del plan de pruebas.
