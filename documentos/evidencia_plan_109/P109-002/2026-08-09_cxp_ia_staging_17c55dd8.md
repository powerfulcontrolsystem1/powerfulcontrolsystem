# P109-002 - flujo CxP/IA autenticado en el candidato 17c55dd8

Fecha: 2026-08-09
Empresa: Powerful Control System (`empresa_id=12`)
Ambiente: staging, candidato `17c55dd86030c47dc9e40d9abc99d2447b9091d9`
Producción: sin cambios.

## Flujo real ejecutado

1. Se radicó por la interfaz oficial un XML sintético controlado como
   `SCI-0009`.
2. **Extraer IA** leyó proveedor, NIT, número, fechas, subtotal COP 500, IVA
   COP 95 y total COP 595, con confianza 98 %.
3. La revisión humana editó proveedor, número, IVA y total; la vista guardó
   COP 96 de IVA y COP 596 de total y agregó dos eventos de edición.
4. Se vinculó un proveedor registrado, se aprobó desde la bandeja y se comprobó
   el tablero con COP 596 aprobado. No se pulsó **Contabilizar**.
5. Se rechazó desde el flujo oficial; la auditoría conserva la transición
   `extraido -> aprobado -> rechazado`.
6. El mismo archivo se radicó de nuevo como `SCI-0010` y quedó bloqueado en
   estado terminal `duplicado`, con referencia al soporte original.
7. Un segundo XML único se radicó como `SCI-0011`. **Cancelar IA** estuvo
   habilitado durante la petición, la cancelación mostró confirmación visible y
   el soporte conservó `radicado`; después se rechazó para cerrar el ensayo.

## Conciliación PostgreSQL de solo lectura

Antes del bloque existían 8 soportes, 3 CxP, 5 pagos y 5 movimientos. Al cierre:

- soportes: 11;
- cuentas por pagar: 3;
- pagos CxP: 5;
- movimientos financieros: 5;
- `SCI-0009`: rechazado, `convertido_id=0`, total COP 596;
- `SCI-0010`: duplicado de `SCI-0009`, `convertido_id=0`;
- `SCI-0011`: rechazado, `convertido_id=0`.

El aumento corresponde únicamente a las tres filas auditables de soporte. No se
creó deuda, pago ni asiento por aprobación, duplicado o cancelación.

## Revisión visual

La bandeja del candidato se inspeccionó en navegador autenticado. Las filas y
columnas de código, estado, proveedor, documento, fecha, total, confianza,
control y archivo permanecieron alineadas; importes, chips y barras de confianza
no presentaron recortes ni solapamientos en escritorio. La revisión editable y
la auditoría mostraron los valores guardados.

## Límites y compuertas pendientes

- El navegador interno no implementa `window.prompt()`. Por ello la prueba de
  papelera/recuperación no pudo completarse visualmente; la sesión administrativa
  independiente rechazó el endpoint con 403, que es el comportamiento seguro.
- La consulta autenticada de una empresa inexistente devolvió un conjunto vacío
  y no expuso datos de PCS, pero una sesión global no reemplaza la prueba A/B con
  segunda identidad y empresa autorizadas.
- Falta repetir papelera/recuperación, matriz completa por roles, degradación del
  proveedor y A/B. P109-002 y P109-008 permanecen parciales.

Resultado: **PASS parcial**, sin efecto financiero y sin cambio del veredicto
general **NO-GO**.
