# Modulo energia solar

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- El catálogo de marcas/protocolos registra opciones; no demuestra un adaptador interoperable para cada fabricante, batería o BMS.
- Telemetría recibida y reglas de alerta no acreditan exactitud eléctrica ni actuación física; probar_alerta puede enviar correo real.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

Fecha: 2026-06-01

## Objetivo

Permitir que cada empresa registre, monitoree y audite su propio sistema de
energia solar, paneles, inversores, controladoras, baterias y BMS sin mezclar
datos entre empresas.

## Referencias del catálogo base

- Victron Energy: VRM Portal, VictronConnect, Venus OS, Cerbo GX, SmartSolar
  MPPT y MultiPlus-II.
- SMA: Sunny Portal powered by ennexOS, Sunny Boy, Sunny Tripower, Sunny Island
  y Data Manager.
- SolarEdge: Monitoring Platform, Home Hub, HD-Wave, inversores trifasicos y
  Power Optimizers.

Tambien se deja un proveedor `gateway_local` para instalaciones con Modbus,
CAN-bus, RS485, MQTT o API local.

## Baterías registrables en el catálogo base

- Tesla Powerwall.
- BYD Battery-Box Premium.
- Pylontech US5000 / US3000C.
- Enphase IQ Battery.
- Victron Lithium NG / Smart Lithium.

El sistema guarda marca, modelo, serial/banco, protocolo BMS, capacidad kWh y
telemetria de SOC, SOH, voltaje, corriente, carga, descarga, ciclos,
temperatura y diferencia entre celdas.

## Flujo operativo

1. Entrar en `Administrar empresa > Analisis y control > Energia solar`.
2. Registrar el sistema solar con proveedor, equipo, bateria, API o gateway.
3. Configurar correos de alerta por empresa.
4. Ajustar alertas por umbral o estado: SOC bajo, bateria sin carga, SOH bajo,
   paneles sin produccion, temperatura alta, desbalance de celdas, error de
   inversor o error BMS.
5. Registrar lecturas desde API/gateway o manualmente durante pruebas.
6. El backend evalua alertas, registra eventos y envia correo si corresponde.

## Preconfiguracion y licencias

- Las preconfiguraciones de tipos de empresa incluyen `modulos.energia_solar`
  como modulo opcional, apagado por defecto.
- El catalogo base de preconfiguracion registra proveedores Victron, SMA,
  SolarEdge y `gateway_local`, baterias comunes y alertas minimas.
- El rol `tecnico_solar` se crea por defecto y recibe solo
  `energia_solar:R`.
- Los administradores y supervisores pueden configurar sistemas, alertas y
  lecturas segun permisos efectivos de la empresa.
- En licencias nuevas el modulo debe habilitarse como `energia_solar`; para
  licencias antiguas se mantiene compatibilidad por fallback desde
  `control_electrico` o `seguridad`.

## API

Endpoint empresarial protegido:

```http
GET  /api/empresa/energia_solar?empresa_id={id}&action=dashboard
GET  /api/empresa/energia_solar?empresa_id={id}&action=catalogo
GET  /api/empresa/energia_solar?empresa_id={id}&action=sistemas
GET  /api/empresa/energia_solar?empresa_id={id}&action=alertas&sistema_id={id}
GET  /api/empresa/energia_solar?empresa_id={id}&action=lecturas&sistema_id={id}&limit=120
GET  /api/empresa/energia_solar?empresa_id={id}&action=eventos&sistema_id={id}&limit=80
POST /api/empresa/energia_solar?empresa_id={id}&action=sistema
POST /api/empresa/energia_solar?empresa_id={id}&action=alerta
POST /api/empresa/energia_solar?empresa_id={id}&action=lectura
POST /api/empresa/energia_solar?empresa_id={id}&action=probar_alerta&sistema_id={id}
```

Todas las acciones validan `empresa_id`, permisos efectivos, licencia y
pertenencia del `sistema_id` a la empresa.

## Seguridad

- Todas las tablas nuevas tienen `empresa_id` y las consultas filtran por
  empresa.
- El endpoint `/api/empresa/energia_solar` usa
  `WithEmpresaEnergiaSolarPermissions`.
- Las llaves reales no se guardan en texto plano: `api_key_ref` exige formato
  `env:NOMBRE_VARIABLE`.
- Los correos se envian con la configuracion SMTP central; en modo prueba se
  capturan como notificaciones de prueba.

## Archivos principales

- `backend/db/energia_solar.go`
- `backend/handlers/energia_solar.go`
- `web/administrar_empresa/energia_solar.html`
- `web/js/energia_solar.js`
- `web/img/solar-energy.svg`

## Fuentes y aceptación de la revisión

[energia_solar.go](../backend/handlers/energia_solar.go), [energia_solar.go](../backend/db/energia_solar.go), [energia_solar.html](../web/administrar_empresa/energia_solar.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go), [control_electrico_solar_telemetry_test.go](../backend/handlers/control_electrico_solar_telemetry_test.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
