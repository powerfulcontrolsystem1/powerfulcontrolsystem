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

## Seguimiento 2026-07-28

- El SHA integrado `e65f6dcd` se comprobó autenticado en 390 x 844.
- La página de correo corporativo terminó con `scrollWidth=375` y
  `clientWidth=375`; su tabla ancha queda contenida por scroll interno.
- Se detectó que las acciones globales `Favorito` y `Panel super` permanecían
  fijas sobre botones operativos en móvil. La rama
  `codex/p108-staging-qa` las devuelve al flujo del documento mediante
  `position: sticky` y agrega un contrato de regresión.

La captura final del ajuste de tabla vive en
`../P108-015/capturas/email_mailu_mobile_viewport_390x844_20260728.png`.
La corrección global de acciones aún debe desplegarse y recapturarse.
