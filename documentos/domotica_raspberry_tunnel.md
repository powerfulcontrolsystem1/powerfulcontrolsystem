# Domotica Raspberry Pi por tunel HTTPS saliente

Actualizacion: 2026-08-11

## Objetivo

Cada empresa puede registrar varias Raspberry Pi y asignar varios aparatos a
cada estacion. Un aparato conserva nombre, descripcion, foto, potencia en
watts, GPIO BCM, polaridad, programacion horaria, estado, lecturas y bitacora.
Una regla puede usar un GPIO de la misma Raspberry como entrada y activar o
desactivar otro aparato.

En la pagina de Estaciones, cada habitacion o estacion muestra un unico boton
`⚡ Domotica`. El check `Mostrar el boton Domotica en cada estacion o
habitacion`, ubicado en Configuracion de estaciones, permite mostrarlo u
ocultarlo. El boton abre una vista filtrada con todos los equipos y sensores de
esa estacion, sus estados y la Raspberry asignada, sin activar el carrito.

## Flujo de instalacion desde la Raspberry

1. Abrir PCS en el navegador de la Raspberry e iniciar sesion en la empresa.
2. Entrar en `Administrar empresa > Analisis y control > Domotica > Raspberry`.
3. Registrar la Raspberry con su IP local y presionar `Generar instalador`.
4. Ejecutar una sola vez el archivo descargado:

   ```sh
   sudo sh ~/Downloads/instalar-pcs-domotica-*.sh
   ```

5. El servicio `pcs-domotica-agent` se habilita con systemd, enrola el ID unico
   y mantiene solicitudes HTTPS salientes hacia el VPS. No se abren puertos en
   el router ni el VPS inicia conexiones hacia la red privada.

Cada Raspberry requiere su propio registro e instalador. Si una empresa tiene
varios controladores, PCS los identifica por separado, muestra cuales tienen
actividad reciente y mantiene comandos, GPIO y transferencia asociados al ID
correcto.

El navegador no puede ejecutar un archivo descargado con privilegios de forma
automatica. El paso `sudo` es deliberado y visible para evitar ejecucion remota
silenciosa. El instalador exige Python 3, usa solamente la biblioteca estandar
y detecta `pinctrl`, `raspi-gpio` o GPIO sysfs.

## Identidad y seguridad

- El panel genera un `device_uid` global opaco y un token de enrolamiento de un
  solo uso con vencimiento de 24 horas.
- PostgreSQL conserva unicamente SHA-256 de los tokens. El secreto plano solo
  aparece dentro del instalador descargado, con `Cache-Control: no-store`.
- Despues del primer enrolamiento, la Raspberry guarda el token operativo en
  `/etc/pcs-domotica/agent.json` con modo `0600`.
- El endpoint publico deriva `empresa_id` y `raspberry_id` de la identidad
  autenticada; nunca acepta que el dispositivo elija empresa.
- Regenerar el instalador rota el token operativo anterior.
- El servicio usa `NoNewPrivileges`, `ProtectHome` y `ProtectSystem=strict`.

## Canal y comandos

La Raspberry hace long polling mediante `POST /api/public/domotica/tunnel`:

- `action=enroll`: canje unico del token de instalacion.
- `action=poll`: heartbeat, configuracion de entradas y entrega durable de un
  comando pendiente.
- `action=ack`: confirmacion de salida GPIO y actualizacion de estado/historial.
- `action=input`: cambio estable de GPIO despues del debounce y evaluacion de
  reglas de la misma empresa y Raspberry.
- `action=telemetry`: lecturas asociadas a aparatos que pertenecen al dispositivo.

La cola `empresa_control_electrico_comandos` reintenta entregas no confirmadas,
expira comandos antiguos y procesa ACK repetidos/concurrentes sin duplicar el
historial. Tambien evita mezclar empresa o Raspberry. Los contadores
acumulados y diarios de RX/TX se consultan en `Super administrador >
Infraestructura > Transferencia Raspberry`.

## Modelo GPIO

Los numeros son BCM (`0..27`), no la posicion fisica del conector. No se debe
conectar una carga directamente a un GPIO: la Raspberry controla un modulo de
relay apropiado, con tierra comun y alimentacion dimensionada. Antes de activar
un equipo real se debe confirmar el canal, la polaridad `active_high` y que el
GPIO no este reservado por I2C, SPI, UART u otra funcion del sistema.
El backend rechaza que un GPIO activo se reutilice para dos aparatos o que el
mismo pin quede simultaneamente como entrada y salida.

En `Domotica > Raspberry`, el boton `Probar GPIO` muestra las salidas BCM
admitidas de la placa. Cada prueba usa el tunel autenticado y aplica un pulso
de un segundo que vuelve la salida a apagado; queda registrada en la bitacora.
En el modulo PCS de 16 relés el pulso se emite activo-bajo. No debe usarse con
cargas críticas conectadas.

## Persistencia

- `empresa_control_electrico_raspberry_pis`: identidad, estado del tunel,
  ultima actividad, version de agente, ultimo `boot_id` y transferencia acumulada.
- `empresa_control_electrico_reles`: aparatos, estacion, GPIO de salida,
  descripcion/foto, categoria, watts y agenda.
- `empresa_control_electrico_reglas`: GPIO de entrada, pull, debounce, condicion
  y aparato objetivo.
- `empresa_control_electrico_comandos`: cola durable y resultado.
- `empresa_control_electrico_eventos` y `empresa_control_electrico_lecturas`:
  auditoria e historial por empresa.
- `empresa_control_electrico_trafico_diario`: RX, TX y solicitudes por dia.

## Variables y operacion

`PCS_DOMOTICA_PUBLIC_BASE_URL` puede fijar el origen publico incluido en el
instalador. Debe ser HTTPS, excepto loopback para pruebas locales. Si no existe,
se usa `https://powerfulcontrolsystem.com`.

Comandos utiles en la Raspberry:

```sh
systemctl status pcs-domotica-agent --no-pager
journalctl -u pcs-domotica-agent -n 100 --no-pager
systemctl restart pcs-domotica-agent
```

Los registros del agente no imprimen tokens ni credenciales PCS.

## Reconexion automatica

El agente conserva long polling con backoff exponencial cuando no hay red,
Internet, DNS o respuesta del VPS. Al recuperarse la conectividad vuelve a
conectarse automaticamente y recibe los comandos durables pendientes. systemd
usa reinicio permanente y limite de arranque desactivado para recuperar tambien
un cierre inesperado o el reinicio de la Raspberry. No requiere intervencion
manual ni una IP publica en la empresa.

En cada arranque el agente genera un identificador efimero de inicio. El VPS
solo una vez por ese identificador reconstruye las salidas que quedaron
confirmadas en estado `on`, ordenadas por estación/GPIO. El agente espera un
segundo entre confirmaciones, evitando energizar todos los relés a la vez.

## Gobierno de transferencia y alertas (2026-08-13)

- `empresa_control_electrico_limites_tunel` conserva por `empresa_id` la cuota mensual en MB, el porcentaje de advertencia y si el túnel debe bloquearse al alcanzar el límite.
- Super Administrador muestra RX/TX diario, mensual y acumulado por Raspberry y consolidado por empresa. La cuota predeterminada es 2048 MB, la advertencia 80% y el bloqueo está activo.
- El túnel comprueba la cuota después de autenticar el dispositivo; el enrolamiento inicial queda disponible para recuperar una instalación. El exceso devuelve HTTP 429 sin ejecutar comandos ni aceptar entradas.
- La empresa puede activar `disconnect_alert_enabled`, registrar correo y definir `disconnect_grace_minutes`. El worker espera ese período y publica una alerta de buzón/campanita y correo una sola vez por valor de `last_seen`.
- La identidad `RPI-` contiene 128 bits aleatorios, tiene índice único global y el secreto plano solo aparece en el instalador de un uso. El agente no envía ni elige `empresa_id`; PostgreSQL lo deriva del `device_uid` y token autenticados.
- Para la primera instalación no es necesaria una IP local: se crea el controlador, se genera el instalador desde la página abierta en la Raspberry y el agente inicia el túnel HTTPS saliente.
