# P108-012 - QA visual y responsive del candidato staging

Fecha: 2026-07-26  
Ambiente: `https://staging.powerfulcontrolsystem.com`  
Empresa autorizada: `12`  
SHA candidato validado por digest:
`fce1655cedff6d3e9235424bfaf0029e80b2ff0c`

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
  `codex/p108-staging-qa` las devuelve al flujo normal del documento mediante
  `position: static` y agrega un contrato de regresión. La primera variante
  `sticky` todavía podía superponerse al botón de prueba y fue descartada en la
  comprobación CDP real.
- El SHA `16c8fbd5` se desplegó en staging y se midió con CDP en 390 x 844:
  `clientWidth=390`, `scrollWidth=390`, `position=static` y
  `overlap=false` entre las acciones globales y `Probar envío`.
- El panel empresarial cargó autenticado sin eventos de consola nuevos tras
  retirar el atributo redundante `allowfullscreen`.

La captura final del ajuste de tabla vive en
`../P108-015/capturas/email_mailu_mobile_viewport_390x844_20260728.png`.
La captura de la corrección global vive en
`capturas/email_mailu_acciones_static_390x844_20260728.png`.

Estado de este hallazgo: **PASS**. P108-012 permanece parcial hasta completar
la matriz responsive, accesibilidad, teclado y roles de todo el inventario.

Después de promover el digest exacto se repitió la medición autenticada en
390 x 844: `clientWidth=390`, `scrollWidth=390`, `position=static`,
`overlap=false` y cero respuestas HTTP 4xx/5xx.

## Revalidación del carrito por digest 2026-07-30

- El candidato inmutable `d262874e` fue promovido exclusivamente a staging con
  los cuatro servicios fijados por digest. API, worker y base respondieron
  saludables antes de abrir la pantalla autenticada de la empresa de prueba.
- En escritorio, cada una de las tres filas observadas mostró sus seis acciones
  en controles separados de 38 px de alto, sin superposición; el contenedor de
  acciones conservó `clientWidth=scrollWidth=358`.
- En móvil 390 x 844 el documento no tuvo desbordamiento horizontal
  (`scrollWidth=390`), y las mismas acciones conservaron controles separados,
  con contenedor interno estable de 358 px. No hubo errores de consola.
- No se pulsaron controles que cambian carritos, inventario, pagos o estados.

Resultado de este hallazgo: **PASS**. P108-012 sigue parcial porque la matriz
visual de todos los módulos, roles y documentos imprimibles aún no está cerrada.
