# P108-012 - QA visual y responsive del candidato staging

Fecha: 2026-07-26  
Ambiente: `https://staging.powerfulcontrolsystem.com`  
Empresa autorizada: `12`  
SHA candidato: `74f91956d35e829178050be9127a1fc14fee065c`

## Evidencia aprobada parcial

- Inicio de sesión, selector de empresa y panel de PCS visibles con la sesión
  autorizada.
- En viewport móvil de 390 px, el contenedor publicado no presentó desborde
  horizontal y la navegación se redujo al control `Mostrar menú`.
- Navegación de Finanzas visible y accesible; la cartera presenta encabezados
  de tabla para código, tercero, documento, vencimiento, original, pagado,
  saldo, mora, estado y acciones.
- La carga CxP asistida está visible y aclara que la IA solo propone datos que
  se revisan y editan antes de guardar.
- El asistente publicado muestra el interruptor de modo agente y no expone un
  selector manual de agente. No se probó envío IA para evitar consumo y efectos
  externos sin un fixture de documento autorizado.

## Límites pendientes

- Repetir las mismas pantallas en anchos tableta/móvil y adjuntar capturas
  comparables por rol.
- Cubrir el inventario completo de módulos y botones, manteniendo excluidas las
  acciones con efectos hasta tener fixture y acta de prueba.
- Ejecutar accesibilidad asistida, teclado y lector de pantalla.

Estado: **parcial; no certifica P108-012**.
