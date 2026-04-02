# CHANGELOG

## 2026-04-02
- Se implementa el modulo de ubicacion GPS por empresa con soporte de multiples dispositivos.
- Se agregan tablas `empresa_gps_dispositivos` y `empresa_gps_recorridos` en `empresas.db`.
- Se crean endpoints CRUD para dispositivos y recorridos GPS en `/api/empresa/ubicacion_gps/*`.
- Se agrega la pagina `web/administrar_empresa/ubicacion_gps.html` con mapa OpenStreetMap (Leaflet).
- Se habilita tracking automatico de recorridos cada 10 segundos por dispositivo.
- Se agregan pruebas en `backend/db/ubicacion_gps_test.go` y `backend/handlers/ubicacion_gps_test.go`.
