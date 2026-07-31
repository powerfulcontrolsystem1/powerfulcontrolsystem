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

## Barrido responsive ampliado 2026-07-30

El barrido autenticado de 48 rutas en escritorio y móvil inspeccionó 96
combinaciones. La repetición dirigida confirmó sin desborde bloqueante Compras,
Clientes, Configuración, Contabilidad Colombia, Facturación electrónica y
Usuarios. En Cobranza móvil detectó un desborde horizontal interno de 6 px en
`Guardar configuración`.

La causa era una grilla de tres columnas a 390 px. En el breakpoint móvil,
Recordatorios automáticos usa ahora una sola columna, los campos `wide` vuelven
al flujo normal y sus tres botones ocupan el ancho disponible. La captura
publicada antes de la corrección conserva el hallazgo; falta validar visualmente
el CSS corregido sobre el nuevo digest.

P108-012 permanece parcial por accesibilidad, teclado, lector de pantalla,
tableta y la matriz completa por rol.

## Verificación visual del digest `f9396da5`

La repetición autenticada en 390 x 844 informó cero problemas visuales para
Cobranza. La inspección de la captura confirmó checks, campos, texto extenso y
botones apilados dentro del panel, sin el desborde horizontal previo:

`capturas/cobranza_mobile_f9396da5_390x844.png`

El hallazgo de Cobranza queda **PASS** sobre el candidato activo; la fase
completa conserva los límites de accesibilidad y roles indicados arriba.

## Impresiones del candidato `5ec1c48f` - 2026-07-30

El workflow inmutable `30591586319` renderizó 18/18 formatos sin casos para
revisión, desbordes, nodos inválidos ni errores de consola:

- factura electrónica, recibo de venta, comprobantes de ingreso y egreso,
  orden de servicio, corte de caja, parqueadero y turno de atención;
- formatos carta y POS donde aplican;
- tres sondas reales de `window.print()` aprobadas: factura POS, ticket de
  turno y recibo de parqueadero.

La inspección visual humana de las capturas de factura electrónica carta y POS
confirmó datos ordenados en filas y columnas, totales legibles, información
legal contenida y firmas alineadas. La evidencia automatizada conserva además
capturas y PDF de los 18 casos.

Estado: **parcial**. La batería de formatos es `PASS`, pero P108-012 todavía
requiere documentos reales extensos, impresión física, teclado, lector de
pantalla, tableta y matriz visual completa por rol.
