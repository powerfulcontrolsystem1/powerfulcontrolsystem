# P109-008 - retención real de papelera en 34cfd852

Fecha: 2026-08-09

Ambiente: staging aislado, PCS `empresa_id=12`.

Producción: sin cambios.

## Flujo autenticado y visual

Se inició sesión por el formulario oficial, se abrió **Captura inteligente de
compras y gastos**, se cambió el filtro de registro a **Papelera** y se
seleccionó el soporte QA reversible `SCI-0010`.

1. La retención se configuró en un día.
2. **Vista previa de retención** informó cero candidatos y 0,00 KB, aclarando
   visualmente que no eliminaba archivos.
3. **Depurar archivo** abrió el diálogo propio con motivo y código fuerte
   `SCI-0010` obligatorios.
4. Al confirmar, el backend rechazó la operación con el mensaje visible
   `el soporte aun no cumple la retencion configurada`.
5. La tabla continuó organizada en filas/columnas y mostró el soporte como
   `Duplicado / Eliminado`; no cambió a depuración pendiente ni depurado.

La pantalla se inspeccionó visualmente en escritorio: filtros, botones,
retención, mensaje de error y fila seleccionada quedaron legibles, sin
solapamientos ni desbordamiento horizontal de la página.

## Conciliación

- `SCI-0010`: `eliminado`, `convertido_id=0`.
- Eventos del soporte: solo `radicar` y `eliminar`; cero inicio/final de purga.
- PCS: CxP 3, pagos 5, movimientos 5.
- La sesión de prueba terminó en la pantalla de login.

La guardia previa a retención queda aprobada. P109-008 sigue parcial: falta un
soporte QA realmente vencido para probar la depuración completa y reanudación
ante caída, antivirus efectivo y aislamiento con una segunda identidad A/B.
