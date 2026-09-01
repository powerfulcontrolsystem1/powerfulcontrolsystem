# P110-007 - Lote de scripts externos en 14 páginas

Fecha: 2026-08-13

## Alcance

Se externalizó el único script inline de catorce páginas que ya estaban libres
de estilos, eventos y URL JavaScript inline. El lote cubre menús empresariales,
historial, bodegas, frecuencia de facturación, chat IA, soporte remoto,
respuesta ePayco y dos superficies del superadministrador.

Cada etiqueta externa permanece en el mismo punto del documento y conserva la
ejecución síncrona original. Ningún script usa import dinámico, import.meta,
document.write o currentScript.

## Pruebas

- Comparación exacta entre el cuerpo inline de HEAD y cada archivo JS: PASS
  para 14/14 páginas.
- node --check sobre los 14 archivos: PASS.
- Inventario CSP: 209 scripts inline, 1.269 bloqueos y 195 páginas.
- Regresión frente a la línea base: cero.

## Límite

Esta comprobación contractual no reemplaza la repetición visual/autenticada del
candidato desplegado. P110-007 continúa parcial; no se retiró unsafe-inline,
no se ejecutó rs y producción no fue modificada.
