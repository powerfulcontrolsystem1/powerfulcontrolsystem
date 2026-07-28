# P108-006 - Carga IA de factura de proveedor en staging

Fecha: 2026-07-28  
Ambiente: `https://staging.powerfulcontrolsystem.com`  
Empresa: Powerful Control System (`empresa_id=12`)  
Resultado: **BLOQUEADO EN EXTRACCIÓN IA**

## Flujo comprobado

1. Inicio de sesión y selección efectiva de la empresa 12.
2. Apertura de Finanzas con `empresa_id=12`.
3. Selección del formulario `Cuenta por pagar`.
4. El botón `Cargar factura o recibo con IA` abrió un selector de un solo
   archivo.
5. Se cargó un XML sintético marcado para QA, sin datos personales ni fiscales
   reales.
6. El backend radicó el archivo como soporte privado `SCI-0001` y lo dejó en
   estado `Radicado`, con revisión humana obligatoria.
7. La extracción respondió de forma segura que la credencial IA guardada no
   puede descifrarse con la llave actual y solicita registrarla nuevamente en
   Super administrador > IA.

## Controles verificados

- La carga no creó una cuenta por pagar.
- El formulario editable permaneció sin documento, valor ni soporte confirmado.
- La cartera canónica conservó un solo registro QA previo: original `$100`,
  pagado `$25`, saldo `$75`.
- No se guardó cartera, no se ejecutó pago y no se contabilizó el soporte.
- La bandeja muestra el soporte con total `$0`, confianza `0%` y revisión
  humana.

## Bloqueo y siguiente prueba

El responsable de configuración debe registrar de nuevo la credencial OpenAI
de staging usando la llave de cifrado vigente. Después se debe repetir
`Extraer IA` sobre el soporte radicado y comprobar:

- proveedor, identificación, documento, fechas, moneda, subtotal, impuestos y
  total;
- edición humana antes de guardar;
- detección de duplicado por empresa y hash;
- creación única de la CxP solo tras confirmación;
- aislamiento negativo frente a otra empresa.

Esta evidencia no certifica P108-006 ni producción.
