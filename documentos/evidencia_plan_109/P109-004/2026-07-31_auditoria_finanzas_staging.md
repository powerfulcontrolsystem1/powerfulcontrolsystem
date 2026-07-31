# P109-004 - Auditoría autenticada limitada de Finanzas en staging

Fecha: 2026-07-31
Entorno: staging aislado, candidato `331e7a222ce806a4151b605b9553744fe4a9bd50`
Empresa: Powerful Control System (`empresa_id=12`)
Alcance: `/administrar_empresa/finanzas.html`, escritorio 1366x900 y móvil 390x844.

## Método

Se ejecutó `tools/qa_e2e_buttons.cjs` con destino, empresa y sesión explícitos.
El auditor limitó los clics a controles de consulta o navegación. Las acciones
con efecto, las exportaciones, los envíos y los controles ambiguos quedaron
fuera del clic automático.

Antes de la repetición se corrigió el clasificador: los textos genéricos
`Cerrar` y `Cancelar`, incluidos los aportados por atributos accesibles, se
marcan para revisión manual. No se usaron para accionar caja, anulación,
descarte ni confirmación.

## Resultado observado

| Vista | Controles | Clics seguros | Diálogos | Errores de página/consola/HTTP >=400 |
| --- | ---: | ---: | ---: | --- |
| Escritorio | 59 | 3 (`Nuevo`, `Editar`, `Buscar`) | 0 | 0 / 0 / 0 |
| Móvil | 59 | 3 (`Nuevo`, `Editar`, `Buscar`) | 0 | 0 / 0 / 0 |

- Total: 118 controles revisados y 6 clics limitados.
- Ningún control cuyo texto contuviera `Cerrar` o `Cancelar` fue pulsado.
- Se observaron dos `net::ERR_ABORTED` por vista al reconstruir secciones tras
  navegación de la auditoría; no hubo respuesta 5xx, error de JavaScript ni
  diálogo de confirmación. Se mantienen como diagnóstico del arnés, no como
  aprobación global del módulo.
- No se confirmó ni envió ningún formulario, por lo que la auditoría no creó,
  modificó ni anuló datos empresariales.

## Límite de esta evidencia

Este resultado es un PASS parcial de P109-004 para una ruta y dos viewports.
Permanecen pendientes el inventario total, los roles limitados, los botones IA,
las mutaciones por flujo oficial y la matriz A/B exigida por el Plan 109.

## Repetición de seguridad del auditor

Después de retirar las altas y acciones IA del clic automático, se repitió el
recorrido sobre los mismos 118 controles. Solo se pulsaron `Editar` y `Buscar`
en cada viewport; no hubo creación, ejecución IA, diálogo, error de página,
consola ni HTTP 4xx/5xx.
