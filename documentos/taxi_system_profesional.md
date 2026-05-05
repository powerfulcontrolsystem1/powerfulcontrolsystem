# Taxi System profesional

Fecha: 2026-05-05
Estado: vigente, documentado tambien en el portal publico `web/index.html`

## Alcance

El modulo `Taxi system` opera como central tipo Uber para empresas de transporte, taxis, mototaxis, domicilios o flotas con despacho por proximidad. Incluye portal de cliente, portal de conductor, panel administrativo, mapa GPS y trazabilidad de rutas.

## Presencia en portal publico

El index comercial describe este modulo como `Taxi system tipo Uber` dentro de la seccion ejecutiva de modulos y agrega una tarjeta fallback `Taxi system`. La descripcion publica resume:

- Mapa operativo para solicitudes, conductores, clientes y dispositivos GPS.
- Tipos de GPS configurables para app movil, tracker dedicado, OBD2, celular, tablet, dashcam o webhook externo.
- Estados de solicitud, asignacion, llegada, inicio, finalizacion, tarifa e historial operativo.

## Superficies

- Administracion: `web/administrar_empresa/taxi_system.html`
- Cliente publico: `web/taxi_system.html`
- Conductor movil: `web/taxi_system_conductor.html`
- API privada: `/api/empresa/taxi_system`
- API publica: `/api/public/taxi_system`

## Funciones profesionales agregadas

- Mapa operativo con Leaflet, capas OpenStreetMap, CARTO claro y calles detalladas.
- Filtros de mapa para todo, conductores disponibles, conductores ocupados, solicitudes y GPS externos.
- Panel rapido de operacion con cola, viajes activos, unidades libres y GPS reportando.
- Base de operacion visible en mapa cuando la empresa configura latitud/longitud base.
- Marcadores diferenciados para conductores, solicitudes, destinos y dispositivos GPS externos.
- Trazo visual entre punto de recogida y destino cuando la solicitud tiene coordenadas de destino.
- Boton para centrar el mapa y boton para cargar la ubicacion del navegador como base de operacion.

## GPS y telemetria

Taxi System reutiliza el modulo corporativo de ubicacion GPS (`empresa_gps_dispositivos` y `empresa_gps_recorridos`) para evitar duplicar inventario tecnico. Desde el panel taxi se pueden registrar dispositivos con:

- Tipo: app movil, tracker dedicado, OBD2, celular corporativo, tablet, dashcam o webhook externo.
- Protocolo: app movil, Traccar, Teltonika, GT06/Concox, GPS103, OsmAnd, Webhook HTTP, MQTT o manual.
- Proveedor, IMEI/identificador, placa/activo, intervalo de reporte, marca y modelo.

El conductor ahora puede quedar asociado a un dispositivo GPS mediante los campos:

- `gps_dispositivo_id`
- `gps_codigo`
- `gps_tipo`
- `gps_proveedor`
- `gps_protocolo`

Estas columnas viven en `empresa_taxi_drivers` y se migran automaticamente desde `EnsureEmpresaTaxiSystemSchema`.

## Endpoints privados relevantes

- `GET /api/empresa/taxi_system?action=dashboard&empresa_id=...`
- `GET /api/empresa/taxi_system?action=config&empresa_id=...`
- `GET /api/empresa/taxi_system?action=drivers&empresa_id=...`
- `POST /api/empresa/taxi_system?action=drivers&empresa_id=...`
- `POST /api/empresa/taxi_system?action=dispatch&empresa_id=...&request_id=...`
- `GET /api/empresa/taxi_system?action=route&empresa_id=...&request_id=...`
- `GET /api/empresa/taxi_system?action=gps_devices&empresa_id=...`
- `POST /api/empresa/taxi_system?action=gps_devices&empresa_id=...`
- `PUT /api/empresa/taxi_system?action=gps_devices&empresa_id=...`

## Validacion

- `go test ./...` desde `backend`
- `node --check web/js/taxi_system.js`
