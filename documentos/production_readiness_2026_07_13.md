# Preparacion para produccion - 2026-07-13

## Alcance y limites

Esta evidencia corresponde a validacion local y CI posterior a la fusion de la
PR #6. No se desplego, no se uso VPS, no se ejecutaron servicios, no se usaron
credenciales ni datos empresariales reales, y no se alteraron protecciones de
GitHub.

## Evidencia aprobada

| Area | Estado | Evidencia |
| --- | --- | --- |
| PR #6 | APROBADO | Revision independiente aprobada, fusion `ecf07e14`; CI de `main` verde. |
| Secretos y PII en logs | APROBADO | Redaccion central para correo, documentos, telefonos, tokens, cookies y Authorization; neutralizacion CR/LF. |
| Utilidades sensibles | APROBADO | Se retiraron inspectores heredados de pagos/login y RustDesk no registrado con secreto plano. |
| Herramientas administrativas | APROBADO | IA y reCAPTCHA simulan por defecto y exigen `-apply` y confirmacion literal. |
| TOTP y tokens | APROBADO | Pruebas de cifrado, codigos de recuperacion y verificadores SHA-256 aprobadas. |
| Aislamiento multiempresa | APROBADO CON RIESGO RESIDUAL | Pruebas de discrepancia de `empresa_id` en query, header y JSON aprobadas; falta corrida contra PostgreSQL efimero. |
| Archivos privados | APROBADO CON RIESGO RESIDUAL | Pruebas de tenant, nombres aleatorios, traversal, MIME, extension, tamano y cabeceras aprobadas. Symlink queda para CI/Linux. |
| WebRTC y soporte remoto | APROBADO CON RIESGO RESIDUAL | Pruebas de Origin, tenant, credenciales y limite de peers aprobadas; falta prueba de socket real efimero. |
| Pagos y webhooks | APROBADO CON RIESGO RESIDUAL | Pruebas de firmas ePayco/Wompi, idempotencia contractual y sanitizacion aprobadas; falta simulador HTTP de proveedor completo. |
| Dependencias | APROBADO CON RIESGO RESIDUAL | `govulncheck` reporta cero vulnerabilidades alcanzables. `GO-2026-5932` es aviso de modulo por `openpgp`, no importado ni alcanzable y sin fix publicado. |
| Rotación CSRF sensible | APROBADO LOCAL | Cambio de contraseña, alta de contraseña local y TOTP rotan el token tras validar la solicitud actual; pruebas de middleware aprobadas. |

## Bloqueantes antes de produccion

1. Ejecutar migraciones y rollback contra PostgreSQL efimero anonimo, incluyendo
   segunda ejecucion idempotente y registros corruptos.
2. Ejecutar migracion de archivos privados solo en modo simulacion sobre un
   inventario de staging, y validar rollback documentado.
3. Restaurar un backup en infraestructura efimera y medir RPO/RTO.
4. Ejecutar CI/Linux para symlinks, `go test -race`, Docker Compose, Trivy, SBOM
   e imagenes endurecidas.
5. Ejecutar pruebas de webhook, WebRTC y carga en staging controlado sin datos
   reales ni proveedores productivos.

## Validación adicional local

- La suite Go y `go vet` se ejecutaron sobre los paquetes del backend; la
  cobertura dirigida incluye CSRF, sesiones, TOTP, tokens temporales,
  callbacks OnlyOffice, WebRTC, ePayco, Wompi y Rappi.
- Esta estación Windows no tiene Docker disponible ni CGO habilitado. La
  validación de Compose, escaneo de imágenes y `go test -race` permanece como
  gate obligatorio de CI/Linux, no como resultado omitido.

## Comandos ejecutados

- `go mod verify`
- `go test ./...`
- `go vet ./...`
- `govulncheck -show verbose ./...`
- pruebas dirigidas de TOTP, tokens, multiempresa, archivos privados, WebRTC,
  RustDesk y firmas de pagos.
- `node tools/security_audit.mjs --out documentos/reportes_profesionales`
- `node tools/openapi_inventory.mjs --out documentos/api/openapi.generated.yaml`

No se marca el producto como listo para produccion. El estado correcto es:
**preparado para iniciar validacion controlada de staging cuando se completen los
bloqueantes anteriores.**
