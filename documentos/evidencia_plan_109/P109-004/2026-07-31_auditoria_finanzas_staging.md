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

## Barrido empresarial ampliado del candidato

El workflow autenticado `30684126476` recorrió el candidato exacto
`2df580b7a0de03d03091e3631061e8073f0c2746` en staging, con 40 rutas
empresariales en escritorio y las mismas 40 en móvil.

| Medida | Resultado |
| --- | ---: |
| Vistas recorridas | 80 |
| Controles detectados | 1.664 |
| Clics seguros ejecutados | 106 |
| Acciones riesgosas omitidas | 213 |
| Vistas correctas | 78 |
| Vistas a revisar | 2 |
| Errores de página | 0 |
| Respuestas HTTP 5xx | 0 |
| Pérdidas de sesión | 0 |

Las dos vistas a revisar corresponden a Centro IA empresarial, en escritorio y
móvil. La interfaz quedó ordenada y responsive, pero el endpoint devolvió 403 y
mostró `rol sin acceso a la funcionalidad solicitada`. La revisión del código
confirmó que `linkCentroIAEmpresarial` está oculto por defecto y requiere una
habilitación explícita por empresa; el wrapper conservó el aislamiento y no se
relajó el permiso para convertir la prueba en PASS.

Las otras seis vistas con hallazgos conservaron estado `ok`: fueron abortos de
requests al navegar entre páginas durante los clics seguros, más un timeout del
botón de notificaciones del panel. No hubo excepción JavaScript, 5xx ni
mutación automática. Este barrido amplía P109-004, pero no sustituye los roles
permitido/denegado, botones IA, exportaciones y acciones reales auditadas que
siguen pendientes.
