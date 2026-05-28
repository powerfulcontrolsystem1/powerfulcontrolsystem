# Modulo energia solar

Fecha: 2026-05-28

## Objetivo

Permitir que cada empresa registre, monitoree y audite su propio sistema de
energia solar, paneles, inversores, controladoras, baterias y BMS sin mezclar
datos entre empresas.

## Proveedores base investigados

- Victron Energy: VRM Portal, VictronConnect, Venus OS, Cerbo GX, SmartSolar
  MPPT y MultiPlus-II.
- SMA: Sunny Portal powered by ennexOS, Sunny Boy, Sunny Tripower, Sunny Island
  y Data Manager.
- SolarEdge: Monitoring Platform, Home Hub, HD-Wave, inversores trifasicos y
  Power Optimizers.

Tambien se deja un proveedor `gateway_local` para instalaciones con Modbus,
CAN-bus, RS485, MQTT o API local.

## Baterias soportadas en catalogo base

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
