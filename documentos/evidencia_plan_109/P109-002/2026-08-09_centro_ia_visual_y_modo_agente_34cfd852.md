# P109-002 - Centro IA empresarial, visual y modo agente

Fecha: 2026-08-09

Ambiente validado: staging, PCS (`empresa_id=12`), candidato `34cfd852`.

## Resultados visibles

- El Centro IA cargó snapshot real, KPIs, alertas y siete funciones IA.
- El interruptor **Modo agente** inició apagado y comunica que solicita
  propuestas/confirmación sin ampliar permisos.
- Un diagnóstico real completó correctamente y registró uso 1/12 sin crear
  venta, CxP, pago, asiento ni documento.
- Se activó el modo agente para una consulta explícitamente no mutante de
  compras/gastos; completó uso 2/12 sin ejecutar cambios. El interruptor se
  devolvió a apagado.
- La bandeja CxP/IA mostró filas, columnas, importes, estado, confianza y
  auditoría ordenados. `Cargar demo` creó `SCI-0012` como duplicado y lo dejó
  bloqueado/no editable, sin convertirlo a cartera.

## Hallazgo y corrección local

La respuesta segura del proveedor llega como Markdown y se veía literalmente
con asteriscos y encabezados. Se corrigió
`web/administrar_empresa/centro_ia_empresarial.html` con un renderizador mínimo
que primero escapa la salida y solo reconoce títulos, listas y énfasis. No agrega
dependencias ni permite HTML del proveedor. Un contrato Go evita la regresión a
la salida literal.

La corrección permanece local hasta su candidato inmutable y revisión visual en
staging; no se desplegó código nuevo y producción no cambió.

## Resultado

El flujo IA y el modo agente no mutante aprueban en staging. P109-002 permanece
parcial por cobertura completa de botones/roles/empresa B, degradación y
certificación del candidato con la corrección visual.
