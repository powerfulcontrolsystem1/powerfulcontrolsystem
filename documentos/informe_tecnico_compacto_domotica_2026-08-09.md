# Informe técnico compacto — Domótica Raspberry Pi

**Estado de release:** desplegado el 2026-08-09; servicios saludables.  
**Alcance:** `control_electrico`, estaciones, túnel Raspberry y monitoreo superadministrador.

## Arquitectura

```text
Panel PCS -> API empresarial -> cola PostgreSQL <- HTTPS saliente <- agente systemd -> GPIO BCM -> relay -> equipo
                                          ^
GPIO de entrada/sensor -> agente -> HTTPS -> reglas PostgreSQL -> comando durable
```

La Raspberry inicia long polling HTTPS. No hay conexiones entrantes desde el VPS hacia la LAN empresarial ni puertos de domótica publicados.

## Componentes principales

| Capa | Responsabilidad |
| --- | --- |
| UI empresarial | Raspberry, equipos, estaciones, GPIO, horarios, fotos, estado e historial. |
| API de túnel | Enrolamiento, poll, ACK, eventos de entrada y telemetría. |
| PostgreSQL | Identidad, reglas, comandos, auditoría, lecturas y tráfico diario. |
| Agente Python stdlib | GPIO, debounce, ejecución de comandos, ACK y systemd. |
| Superadministrador | Transferencia RX/TX y salud de cada Raspberry por empresa. |
| Estaciones | Botón `⚡ Domótica` configurable y vista filtrada de equipos/sensores por estación. |

## Controles relevantes

- Aislamiento obligatorio por `empresa_id`; se deriva de la identidad del dispositivo, no del payload.
- `device_uid` opaco; token de instalación de un uso/vencimiento y hash SHA-256 persistido.
- Cola durable con reintento, expiración y ACK idempotente.
- Validación GPIO BCM `0..27`; bloqueo de colisiones entrada/salida y salidas duplicadas activas.
- Instalador protegido, `Cache-Control: no-store`, HTTPS y configuración local restrictiva; servicio endurecido mediante systemd.
- Identidad e instalador exclusivos por Raspberry, detección de múltiples
  controladores y reconexión automática con backoff más reinicio permanente.

## Modelo de datos

`empresa_control_electrico_raspberry_pis`, `empresa_control_electrico_reles`, `empresa_control_electrico_reglas`, `empresa_control_electrico_comandos`, `empresa_control_electrico_eventos`, `empresa_control_electrico_lecturas` y `empresa_control_electrico_trafico_diario`.

## Operación y verificación

El release ejecutó el migrador y dejó backend, frontend y worker saludables; `/health` y `/ready` respondieron 200. El endpoint de túnel rechaza peticiones de enrolamiento no autorizadas con 401. Pruebas enfocadas de Go, parseo del agente y preflight fueron satisfactorios.

**Pendiente:** prueba de campo con relay real (instalación descargada desde PCS, enrolamiento, comando, entrada GPIO y auditoría). No se certifica esa parte hasta contar con dicha evidencia física.
