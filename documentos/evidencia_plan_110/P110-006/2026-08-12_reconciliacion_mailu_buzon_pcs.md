# P110-006 - reconciliacion verificable del buzon Mailu

Fecha: 2026-08-12
Empresa: PCS de pruebas autorizadas

## Hallazgo real en staging

La pagina de correo corporativo mostro el buzon en
`pendiente_provision`, sin conteo de no leidos ni apertura automatica. Por tanto,
la entrega SMTP aceptada previamente por Mailu no prueba por si sola que el
INBOX empresarial sea autenticable y tampoco permite confirmar visualmente la
alerta desde esa pagina.

## Correccion local, aun no desplegada

- La direccion heredada exacta `mailu-imap:143` se reconduce al front interno de
  Mailu. Con autenticacion por tokens entre servicios, una credencial normal de
  buzon debe entrar por el front y no directamente al servicio IMAP.
- El boton `Actualizar` ejecuta ahora un POST protegido por CSRF para conciliar
  el estado.
- El backend solo cambia a `provisionado` cuando una autenticacion real contra
  INBOX fue comprobada. Si Mailu no autentica, responde conflicto y conserva el
  estado anterior.
- Las pruebas cubren remapeo, reconciliacion positiva y negativa, y contrato
  frontend de POST/CSRF.

## Resultado y limite

La correccion evita marcar un buzon sano por inferencia. Falta desplegarla y
comprobar visualmente el INBOX, la alerta recibida, no leidos, reset y rebote.
P110-006 continua parcial; no se ejecuta `rs` en este bloque.
