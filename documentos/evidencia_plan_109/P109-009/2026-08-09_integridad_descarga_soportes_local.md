# P109-009 - Integridad y descarga segura de soportes

Fecha: 2026-08-09
Entorno: candidato local, sin datos PCS
Rama: `codex/p109-batch-no-pr`

## Controles

- Antes de descargar o enviar el adjunto a IA se verifica SHA-256 contra el hash
  persistido. La comparación usa tiempo constante.
- Un hash malformado, contenido alterado, archivo vacío o mayor de 15 MiB falla
  cerrado y no llega al navegador ni al proveedor.
- Registros históricos sin hash conservan compatibilidad, pero las cargas nuevas
  siempre almacenan hash desde la admisión validada. Los históricos se sirven
  como binario y vuelven a validar firma antes de enviarse a IA.
- La descarga fuerza `attachment`, MIME canónico por extensión o
  `application/octet-stream`, `nosniff`, `sandbox`, `same-origin`, `no-referrer`,
  `DENY` y `no-store`.
- Un contador agregado registra fallos sin empresa, soporte, hash, nombre o ruta.
  Prometheus alerta y Grafana muestra el total por job.
- El incidente se registra transaccionalmente por `empresa_id` y soporte con
  canal/resultado mínimos. Si existía aprobación, se limpia y vuelve a
  `en_revision`; estados terminales contables no se reescriben.

## Pruebas

- Bytes y lector íntegros, hash mayúscula y compatibilidad sin hash: PASS.
- Archivo alterado y hash inválido: rechazo PASS y contador exacto.
- El lector vuelve al inicio después de verificar, por lo que ServeContent no
  entrega un archivo vacío.
- Matriz MIME PDF/XML/HTML/extensión desconocida: PASS.
- Contrato de cabeceras y descarga adjunta: PASS.
- Contrato de `FOR UPDATE`, invalidación de aprobación y evento minimizado: PASS.
- Ensayo PostgreSQL A/B/aprobado/terminal: preparado; SKIP local por falta de
  `PCS_TEST_POSTGRES_DSN` y sin crédito dinámico.
- Go completo, vet y preflight Full/Strict: PASS.

## Límites

No se descargó ni alteró un archivo real de PCS. Falta desplegar el candidato y
probar un fixture reversible PostgreSQL en staging con A/B. P109-009 continúa parcial; el
Plan 109 permanece 53,3 % y NO-GO.
