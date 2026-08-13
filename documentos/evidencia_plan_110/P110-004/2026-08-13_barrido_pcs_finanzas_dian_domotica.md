# P110-004 - Barrido autenticado PCS de superficies P0

Fecha: 2026-08-13

Empresa: Powerful Control System (`empresa_id=12`)

Entorno: dominio principal actualmente publicado

Modo: sesión administrativa autorizada, navegación de solo lectura

## Cobertura

| Superficie | Escritorio | Móvil | Resultado visible |
| --- | --- | --- | --- |
| Finanzas | aprobado | aprobado | configuración, movimientos, comprobantes y cierres cargaron |
| Facturación electrónica | aprobado | aprobado | filtros, totales, documentos y acciones cargaron en filas/columnas |
| Domótica | aprobado | aprobado | 2 Raspberry, 16 aparatos, consumo actual 129 W |

Las tres páginas conservaron `scrollWidth == clientWidth` en el viewport móvil
del navegador interno, sin desbordamiento horizontal. Facturación mostró la
factura y nota crédito de la prueba controlada anterior junto con las
identidades QA por rol ya existentes. No se emitieron, reenviaron, anularon ni
exportaron documentos en este barrido.

## Seguridad y efectos

- El alcance empresarial fue `empresa_id=12` resuelto por el flujo oficial de
  selección de empresa.
- No se pulsaron acciones mutantes de Finanzas, `Anular`/`Compartir` en
  Facturación, ni `Ejecutar agenda`, `Sincronizar` o controles de aparatos en
  Domótica.
- El buzón mantiene seis avisos DIAN históricos; no se reintentaron porque el
  Plan 110 ya registra la aceptación posterior de `1PCS7` y su nota crédito.
- El error aislado `MutationObserver.observe` volvió a aparecer con una marca
  temporal anterior a la navegación. La investigación histórica lo atribuye al
  entorno inyectado del navegador y los observadores PCS ya validan nodos; no se
  aplicó una corrección especulativa.

## Límite

Esta evidencia cubre visualización administrativa real en dos tamaños. No
sustituye las acciones por cuatro roles, cajas simultáneas, hardware Raspberry,
impresión física, UAT contable ni piloto firmado. No se ejecutó `rs`.

## Consulta IA e inventario

El asistente central respondió una consulta QA neutra con modo agente apagado,
sin propuesta ni acción. Informó stock negativo de `menta` y de un producto QA.
El endpoint oficial de resumen confirmó 2 alertas sin stock, déficit total 19 y
93 movimientos, pero los KPI visibles seguían en cero: la vista predeterminada
`productos` no invocaba la carga del resumen. El candidato añade esa llamada y
una regresión estática. La corrección visual queda pendiente de despliegue; no
se ajustaron existencias ni se creó movimiento de inventario.
