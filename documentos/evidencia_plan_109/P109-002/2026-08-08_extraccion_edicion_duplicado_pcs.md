# P109-002 - extracción, edición y duplicado reales en PCS

Fecha: 2026-08-08
Empresa: Powerful Control System (`empresa_id=12`)
Ambiente: VPS publicado; usuario administrador autorizado.
Rama de corrección local: `codex/p109-batch-no-pr`; sin PR ni despliegue.

## Flujo ejecutado

1. Se radicó un XML sintético controlado como `SCI-0002`, sin proveedor ni
   valores manuales previos.
2. El botón **Extraer IA** obtuvo proveedor, NIT, número, fecha, subtotal
   COP 1.000, IVA COP 190 y total COP 1.190, con confianza 98 %.
3. La revisión humana editó proveedor, categoría, centro de costo y
   observaciones. La vista y la auditoría mostraron los valores guardados.
4. El soporte se rechazó por el flujo oficial. No se aprobó ni contabilizó.
5. Se radicó otra vez el mismo XML como `SCI-0003`. La extracción lo dejó en
   estado terminal `duplicado`, confianza 95 % y referencia `Duplicado #6`.

La bandeja se revisó visualmente: conserva filas, columnas, importes, barras de
confianza, estados y enlaces sin recortes en escritorio. La consola terminó sin
errores ni advertencias.

## Conciliación de solo lectura

- `SCI-0002`: `rechazado`, `convertido_id=0`.
- Eventos de `SCI-0002`: cuatro; eventos `contabilizar`: cero.
- `CXP-SCI-0002`: cero filas.
- Outbox `cuentas_por_pagar.soporte_ia_contabilizado` para el soporte: cero.
- `SCI-0003`: duplicado terminal; no aumenta pendiente ni total pendiente.

Los archivos QA quedan vinculados a soportes terminales dentro de PCS porque el
flujo oficial no ofrece borrado de evidencia. No se usó SQL de escritura. El
archivo sintético local fue eliminado al terminar.

## Hallazgo y corrección local

La UI permitía pulsar **Rechazar** sobre un duplicado y dependía del rechazo del
backend. La rama local sincroniza los controles con la matriz de estados del
servidor y conserva la validación previa dentro del manejador de cada botón:

- extraer: `radicado`, `extraido`, `en_revision`;
- aprobar: `radicado`, `extraido`, `en_revision`;
- rechazar: los anteriores y `aprobado`;
- contabilizar: únicamente `aprobado`.

La sintaxis JavaScript, pruebas enfocadas de soportes, `go test ./...`,
`go vet ./...` y `git diff --check` aprobaron. La corrección no está desplegada
porque el usuario pidió no crear PR y el repositorio prohíbe publicar un árbol
sin revisión. P109-002 permanece **parcial** por A/B no global y la matriz
completa por roles; esta evidencia no aumenta el porcentaje formal.
