# P110-008 — Candidato inmutable `1a6dc4fa` en staging

Fecha: 2026-08-12  
Ambiente: staging aislado. Producción no fue modificada.

## Candidato

El workflow de candidato inmutable aprobó construcción única, escaneo,
generación de SBOM, publicación y validación de Compose para el SHA
`1a6dc4fadc93cb25435a7868ebc0a48004d7767c`.

Se descargaron los artefactos del workflow y se promovieron los cuatro digests
exactos de API, migrador, worker y frontend. La imagen ClamAV ya existente se
mantuvo fijada por digest.

## Verificación posterior

- Backend, worker y frontend fueron recreados explícitamente con las referencias
  inmutables del candidato.
- `/health` y `/ready` respondieron correctamente.
- PostgreSQL y ClamAV permanecieron saludables; el ping de ClamAV aprobó.
- La configuración DIAN aislada de PCS conservó una fila, ambiente de
  habilitación y emisión local desactivada.

## Hallazgo operativo

El checkout operativo remoto usa un Compose anterior que no añade de forma
predeterminada el archivo antivirus. Por ello su primer comando reportó ClamAV
como huérfano, aunque el servicio continuó saludable. El candidato no se
certifica todavía hasta alinear el checkout/Compose remoto y repetir el drill
efímero de restore/rollback con sus digests.
