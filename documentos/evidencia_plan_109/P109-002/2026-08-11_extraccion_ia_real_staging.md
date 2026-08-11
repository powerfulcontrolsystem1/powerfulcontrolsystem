# P109-002 - extracción IA real y revisión humana en PCS staging

Fecha: 2026-08-11

Estado: **evidencia parcial; no aprobada para producción**.

## Alcance controlado

Se usó la sesión autorizada de PCS en staging, empresa `empresa_id=12`, sobre
un soporte de prueba ya existente. No se aprobó, contabilizó ni creó una cuenta
por pagar.

## Resultado observable

1. La pantalla `soportes_compras_ia` cargó con el contexto empresarial correcto
   y mostró la acción **Extraer IA** únicamente para el soporte radicado.
2. La extracción real con `openai:gpt-5.5` terminó de forma controlada: el
   contenido de prueba, sin datos comerciales suficientes, quedó marcado para
   revisión humana/duplicado y explicó esa condición sin crear una CxP.
3. La métrica agregada del backend registró `human_review=1`; los resultados de
   proveedor, respuesta inválida, cancelación y persistencia permanecieron en
   cero para esta ejecución.
4. En un soporte de prueba radicado se guardó una revisión humana visible y
   auditable. La interfaz confirmó que la revisión no contabiliza por sí sola.

## Límites que permanecen

- Esta evidencia no cubre factura/recibo reales con extracción de campos útiles,
  cancelación en curso, doble clic, error/reintento del proveedor, ReportSpec,
  memoria A/B ni evaluación completa de IA.
- No sustituye aprobación humana contable, acciones de contabilización ni UAT.
- P109-002 sigue parcial y el Plan 109 conserva NO-GO.
