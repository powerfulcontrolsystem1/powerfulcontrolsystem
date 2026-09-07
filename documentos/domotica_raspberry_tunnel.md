# Domotica Raspberry Pi por tunel HTTPS saliente

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Identidad del dispositivo autenticada por túnel: la empresa no se toma de un mensaje libre del hardware.
- El instalador versionado es la fuente de capacidades del agente; las pruebas GPIO/relé, reinicio y telemetría requieren hardware identificado y autorización operativa.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

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
2. Entrar en `Administrar empresa > Domotica y Energia Solar > Configuracion de domotica > Controladores`.
3. Registrar la Raspberry; la IP local es opcional para el túnel saliente. Presionar `Generar instalador`.
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
- Regenerar el instalador crea un enrolamiento nuevo sin desconectar el agente
  vigente. El token operativo anterior solo se sustituye cuando el nuevo agente
  completa correctamente el enrolamiento.
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
- `action=solar_telemetry`: bloques VE.Direct validados y asociados a la empresa
  y Raspberry derivadas del token del túnel.

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
- `empresa_control_electrico_escenas` y
  `empresa_control_electrico_escena_items`: estados agrupados de varios aparatos.
- Las columnas `ssh_*_enc` de la Raspberry conservan opcionalmente password y
  sudo cifrados; no forman parte de las respuestas JSON.

## Variables y operacion

`PCS_DOMOTICA_PUBLIC_BASE_URL` puede fijar el origen publico incluido en el
instalador. Debe ser HTTPS, excepto loopback para pruebas locales. Si no existe,
se usa `https://powerfulcontrolsystem.com`.

`PCS_DOMOTICA_SSH_ALLOWED_CIDRS` contiene una lista separada por comas de redes
privadas que el VPS puede alcanzar realmente por VPN o ruta dedicada. Si una IP
privada no pertenece a esa allowlist, la instalación SSH se rechaza para evitar
acceso lateral a la infraestructura del VPS.

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

En cada arranque el agente lee `/proc/sys/kernel/random/boot_id`. El VPS
solo una vez por ese identificador reconstruye las salidas que quedaron
confirmadas en estado `on`, ordenadas por estación/GPIO. La cola usa el retardo
configurado por empresa, un segundo de forma predeterminada, evitando energizar
todos los relés a la vez.

## Instalación alternativa por SSH

- `Instalar por SSH` está reservado a usuarios con aprobación efectiva de
  Domótica y siempre filtra `empresa_id + raspberry_id`.
- El primer contacto devuelve la huella SHA-256 del host; no envía contraseña ni
  instalador hasta que el usuario la confirma.
- Password y sudo pueden guardarse con AES-GCM usando `CONFIG_ENC_KEY`. El
  propósito criptográfico contiene empresa, Raspberry y tipo de secreto, por lo
  que un ciphertext de otro tenant o dispositivo no puede reutilizarse.
- El instalador viaja a un nombre aleatorio en `/tmp`, modo `umask 077`, se
  ejecuta mediante sudo por stdin y se elimina al finalizar. Los secretos no se
  incluyen en argumentos, logs, auditoría o respuestas.

## Victron VE.Direct

El agente incluye autodetección; la versión instalada debe cotejarse con el instalador. Autodetecta adaptadores VE.Direct entre rutas estables
`/dev/serial/by-id` y puertos `ttyUSB`/`ttyACM`. Configura 19200 baudios, 8 bits,
sin paridad, un stop bit y sin control de flujo. Solo acepta un bloque con PID,
voltaje de batería, voltaje de panel y checksum módulo 256 válido.

Cada 15 segundos publica potencia/voltaje del panel, corriente/voltaje de
batería, producción diaria, etapa del cargador, error, firmware, serial y puerto.
El dashboard considera desconectado un sistema sin lecturas recientes. Una
BlueSolar MPPT no mide SOC: PCS conserva `No disponible` hasta recibir esa
métrica desde un BMV, SmartShunt o BMS compatible.

## Gobierno de transferencia y alertas (2026-08-13)

- `empresa_control_electrico_limites_tunel` conserva por `empresa_id` la cuota mensual en MB, el porcentaje de advertencia y si el túnel debe bloquearse al alcanzar el límite.
- Super Administrador muestra RX/TX diario, mensual y acumulado por Raspberry y consolidado por empresa. La cuota predeterminada es 2048 MB, la advertencia 80% y el bloqueo está activo.
- El túnel comprueba la cuota después de autenticar el dispositivo; el enrolamiento inicial queda disponible para recuperar una instalación. El exceso devuelve HTTP 429 sin ejecutar comandos ni aceptar entradas.
- La empresa puede activar `disconnect_alert_enabled`, registrar correo y definir `disconnect_grace_minutes`. El worker espera ese período y publica una alerta de buzón/campanita y correo una sola vez por valor de `last_seen`.
- La identidad `RPI-` contiene 128 bits aleatorios, tiene índice único global y el secreto plano solo aparece en el instalador de un uso. El agente no envía ni elige `empresa_id`; PostgreSQL lo deriva del `device_uid` y token autenticados.
- Para la primera instalación no es necesaria una IP local: se crea el controlador, se genera el instalador desde la página abierta en la Raspberry y el agente inicia el túnel HTTPS saliente.

## Fuentes y aceptación de la revisión

[control_electrico.go](../backend/handlers/control_electrico.go), [instalar_domotica_raspberry.sh.tmpl](../backend/handlers/templates/instalar_domotica_raspberry.sh.tmpl), [control_electrico.go](../backend/db/control_electrico.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
