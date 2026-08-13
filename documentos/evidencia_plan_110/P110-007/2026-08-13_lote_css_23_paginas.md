# P110-007 - Lote CSS externo en 23 páginas

Fecha: 2026-08-13  
Ambiente: candidato local de la PR 162; producción y staging no modificados.

## Alcance

Se externalizó el único bloque style de 23 páginas que no tenían scripts,
atributos de estilo, eventos ni URL JavaScript inline. El lote cubre el shell
empresarial, CRM, WMS, MRP, Ayuda de APIs, la vista de conceptos de marca,
auditorías y seguridad del superadministrador y trece accesos de configuración.
Los trece accesos con CSS idéntico comparten una sola hoja; Nextcloud conserva
su hoja específica.

## Resultado reproducible

- Bloqueos CSP: 1.306 -> 1.283.
- Páginas afectadas: 232 -> 209.
- Bloques style: 177 -> 154.
- Eventos inline y URL JavaScript: continúan en cero.
- Los 23 HTML objetivo quedan en cero bloqueos del inventario por página.
- Todos los enlaces CSS nuevos resuelven a un archivo versionado.

## Validación

- go test ./...: PASS.
- go vet ./...: PASS.
- preflight profesional completo y estricto, sin repetir Go: PASS.
- auditorías de seguridad, permisos, UX, migraciones, contratos y Docker: PASS.
- git diff --check: PASS, únicamente con avisos de normalización LF/CRLF.

La navegación visual local no se contabiliza como PASS: el navegador interno
bloqueó la URL file: y el servidor auxiliar no quedó escuchando. Por ello
la repetición visual en navegador sobre el candidato desplegado sigue
obligatoria antes de aplicar una CSP estricta.

## Estado

P110-007 continúa **parcial**. Restan 1.283 bloqueos, DAST integral y la matriz
A/B completa; no se retiró unsafe-inline, no se ejecutó rs y no se
modificó producción.
