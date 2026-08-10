# Informe completo — módulo de Domótica Raspberry Pi

**Fecha:** 2026-08-09  
**Estado:** desplegado y operativo a nivel de plataforma; pendiente validación física final del relay en sitio.

## Resultado entregado

PCS incorpora un módulo multiempresa de domótica para controlar equipos eléctricos desde las estaciones de cada empresa mediante una o varias Raspberry Pi conectadas a módulos de relay. La comunicación es saliente desde la Raspberry hacia PCS: no requiere abrir puertos del router ni publicar la red local de la empresa.

Cada empresa mantiene sus propios controladores, estaciones, equipos, reglas, horarios, lecturas e historial. El servidor valida el alcance empresarial en toda operación; una Raspberry no puede seleccionar ni consultar otra empresa.

## Capacidades para la empresa

### Equipos por estación

En `Administrar empresa > Análisis y control > Domótica` se pueden registrar varios equipos por estación (habitación, mesa u otra ubicación). Cada equipo incluye:

- nombre, descripción, foto o icono;
- consumo nominal en watts;
- Raspberry asignada, GPIO BCM de salida y polaridad del relay;
- estado visible de encendido/apagado;
- programación de encendido y apagado;
- lecturas y bitácora histórica por empresa.

Desde la estación el operador autorizado puede solicitar el encendido o apagado del equipo. La orden queda en una cola durable, por lo que una indisponibilidad temporal de la Raspberry no pierde la solicitud: se entrega cuando el agente se reconecta o expira bajo una política controlada.

Cada tarjeta de estación dispone de un único botón `⚡ Domótica`, activable o
desactivable mediante un check de Configuración de estaciones. Este acceso abre
una página filtrada con equipos, sensores, estados y Raspberry asociadas, sin
activar la estación ni abrir su carrito.

### Entradas GPIO y automatización

Una Raspberry puede configurar un GPIO como entrada para sensor o señal. Una regla asocia dicha entrada con un equipo de salida de la misma Raspberry y de la misma empresa. Incluye resistencia pull, debounce y condición de disparo. El backend evita reutilizar un GPIO activo como salida de dos equipos o como entrada y salida a la vez.

Ejemplo: una señal estable en el GPIO de entrada puede encender o apagar un relay que controla una nevera, luminaria u otro equipo registrado.

### Instalación desde la propia Raspberry

El administrador abre PCS en el navegador de la Raspberry, registra el controlador con su IP local y selecciona **Generar instalador**. Después ejecuta una sola vez el archivo descargado con privilegio administrativo. El instalador crea un identificador único de dispositivo, habilita `pcs-domotica-agent` en systemd y mantiene una conexión HTTPS saliente hacia el VPS.

El paso de ejecución local con `sudo` es intencional: un navegador no debe ejecutar silenciosamente archivos con privilegios del sistema.

Cada Raspberry se registra con identidad e instalador propios. PCS detecta y
muestra varios controladores por empresa y distingue cuáles mantienen actividad
reciente. Ante pérdida de red, Internet, DNS o VPS, el agente reintenta con
backoff; systemd lo reinicia automáticamente incluso después de reiniciar el
dispositivo.

### Supervisión del superadministrador

En `Super administrador > Infraestructura > Transferencia Raspberry` se ve, por empresa y controlador, el estado del túnel, última conexión, versión del agente y consumo acumulado/diario de RX, TX y solicitudes. La vista no expone tokens ni secretos del dispositivo.

## Seguridad y controles operativos

- Los tokens de enrolamiento vencen, son de un solo uso y se almacenan como huella SHA-256, no en texto plano.
- Regenerar el instalador rota la credencial anterior.
- La identidad autenticada del dispositivo determina `empresa_id` y Raspberry; estos valores no son aceptados desde el agente como autoridad.
- El instalador exige HTTPS salvo entorno local explícito y guarda su configuración con permisos restrictivos.
- El servicio del agente aplica endurecimiento de systemd y no abre servicios entrantes en la LAN.
- Los cambios de estado, eventos y lecturas se conservan como historial por empresa.
- Los GPIO usan numeración BCM (0 a 27). Las cargas eléctricas se conectan al módulo de relay, nunca directamente a un GPIO de la Raspberry.

## Persistencia y trazabilidad

El módulo extiende la persistencia empresarial con controladores Raspberry, equipos/relays, reglas GPIO físicas, comandos durables, eventos, lecturas y tráfico diario. Las migraciones se ejecutaron mediante el migrador de PCS durante el despliegue. La bitácora permite reconstruir quién solicitó una acción, qué dispositivo la ejecutó y cuál fue su resultado.

## Validación realizada

- Pruebas enfocadas de base de datos y handlers: aprobadas.
- Validación de sintaxis del agente Python y del JavaScript de interfaz: aprobada.
- Auditorías de rutas, inventario y preflight profesional: aprobadas.
- Despliegue `rs`: completado; migrador ejecutado correctamente.
- Backend, frontend y worker publicados: saludables.
- Endpoints de salud y readiness: respuesta HTTP 200.
- Endpoint de túnel: rechaza enrolamientos inválidos con HTTP 401.
- Conectividad SSH de la Raspberry de prueba en la red local: disponible.

## Pendiente para cierre físico

Falta ejecutar desde la Raspberry el instalador generado por PCS y comprobar con el relay real: enrolamiento, presencia en línea, un encendido/apagado de prueba, la regla de entrada GPIO y el historial resultante. Esta fase debe hacerse verificando antes el canal de relay, polaridad y alimentación para no activar una carga incorrecta.

## Guía rápida de uso

1. Cree o identifique una estación.
2. Registre la Raspberry y descargue/ejecute su instalador desde ella.
3. Confirme que el panel muestre el túnel en línea.
4. Agregue los equipos, asignando GPIO, watts, horario e imagen.
5. Pruebe un equipo sin carga crítica y revise la bitácora.
6. Configure reglas de entrada GPIO solo después de validar el pin y el relay.
7. Revise el consumo de transferencia desde el panel de superadministración.

Para detalles de operación, instalación y diagnóstico consulte `documentos/domotica_raspberry_tunnel.md`.
