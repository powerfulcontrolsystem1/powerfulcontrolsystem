# P109-000 - Candidato staging bb285968 y compuerta reproducible

Fecha: 2026-08-09

## Despliegue aislado

El SHA `bb285968256c08e879becbdff7a7a9164378f349` se desplegó solo en staging
desde `/root/powerfulcontrolsystem-staging-p109-bb285968`, separado del checkout
que sirve producción. El script comprobó la coincidencia entre rama y SHA,
construyó el candidato y dejó `/health` y `/ready` en `ok`/`ready`.

Frontend, backend, worker y PostgreSQL quedaron saludables. Producción no fue
reconstruida ni reiniciada.

## Validación visible

Centro IA empresarial de PCS mostró la respuesta del proveedor con títulos,
énfasis y listas legibles. No quedaron marcadores Markdown literales ni HTML
interpretado. Las siete funciones IA completaron sin error dentro del límite
diario; el modo agente inició y terminó apagado.

Las invariantes posteriores conservaron 3 CxP, 5 pagos y 5 movimientos. La
prueba agregó solo un soporte demo duplicado bloqueado; el total de soportes QA
de PCS quedó en 12.

## Compuerta automatizada

- Preflight completo: PASS.
- E2E íntegro: 363 variantes registradas antes de que el límite externo de 15
  minutos interrumpiera el proceso. No se atribuye PASS completo a ese barrido.
- E2E dirigido: 2/2 vistas escritorio/móvil, 108 botones detectados, cero
  errores de página, mutaciones bloqueadas por la guardia.
- Impresiones: 20/20 PASS, cero revisión y cero fallo de autoimpresión.
- Carga staging: p95 411 ms, error 0 %.

`release_gate.ps1` ahora resuelve Node y Chrome/Edge instalados sin instalar
dependencias. `release_manifest.mjs` exige además el digest inmutable de
frontend, que antes no figuraba entre los bloqueos.

## Resultado

El candidato técnico y la corrección visual aprueban staging. La compuerta
estricta permanece **NO-GO** porque el build de staging usa tags locales y aún
no existen cuatro digests inmutables publicados; la evidencia no lo sustituye.
