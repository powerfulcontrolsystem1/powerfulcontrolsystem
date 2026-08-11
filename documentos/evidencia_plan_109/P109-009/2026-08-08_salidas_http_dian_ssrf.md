# P109-009 - salidas HTTP DIAN e integraciones protegidas contra SSRF

Fecha: 2026-08-08
Ambiente: candidato local `codex/p109-batch-no-pr`
Datos operativos: ninguno modificado
Efectos externos: ninguno; no se emitieron documentos ni se consultó DIAN real

## Hallazgo

La auditoría posterior al cierre de redirecciones OnlyOffice encontró dos
clientes HTTP adicionales con entradas configurables:

- monitor y prueba de salud de integraciones empresariales;
- envío, acuse, `GetStatusZip`, `GetNumberingRange` y reconexión DIAN;
- despacho y comprobación de conectividad de proveedores fiscales configurados
  mediante `api_base_url`.

La normalización anterior aceptaba cualquier esquema parseable, destinos
loopback/privados y redirecciones por defecto. Además, la identificación de un
endpoint oficial DIAN usaba coincidencia parcial de texto, por lo que un dominio
que solo incluyera `dian.gov.co` podía clasificarse incorrectamente.

## Corrección

El transporte saliente ahora:

- acepta solo HTTP y HTTPS sin credenciales embebidas;
- rechaza `localhost`, `.local`, loopback, RFC1918, link-local, metadata cloud,
  CGNAT, documentación, benchmark, multicast y rangos reservados IPv4/IPv6;
- resuelve DNS antes de conectar, rechaza el conjunto completo si contiene una
  dirección no pública y conecta directamente a una IP ya validada;
- desactiva proxies de entorno para que no eludan la validación del destino;
- permite hasta diez redirecciones y solo dentro del mismo esquema, host y
  puerto;
- exige que overrides DIAN pertenezcan al origen guardado en la configuración
  de la misma empresa;
- reconoce como DIAN oficial únicamente HTTPS sobre `dian.gov.co` o un
  subdominio exacto.

Los servidores locales usados por los tests DIAN conservan un factory HTTP
inyectado únicamente desde `_test.go`; el binario runtime usa siempre el
transporte protegido.

## Pruebas

- 16 direcciones no públicas rechazadas y tres públicas aceptadas;
- protocolos, credenciales URL, loopback y metadata rechazados;
- acuse/override DIAN del mismo origen aceptado y cruce de esquema, host o
  puerto rechazado;
- dominio `dian.gov.co.evil.example` rechazado como no oficial;
- redirecciones del mismo origen aceptadas y redirecciones cruzadas rechazadas;
- un servidor loopback sintético recibió cero solicitudes del probe bloqueado;
- despacho y health de proveedor fiscal contra loopback: bloqueados, cero
  solicitudes recibidas;
- set DIAN 2+2+2 y set configurado: PASS con transporte de test explícito;
- inventario multiempresa regenerado: 204 rutas protegidas y cero revisiones;
- inventario runtime `Ensure` regenerado: 106 llamadas catalogadas;
- `go test ./... -count=1`: PASS;
- `go vet ./...`: PASS.
- preflight profesional `-Full -Strict`: PASS en todas las compuertas.

No se cambió `empresa_id`, permisos, tablas, secretos, XML, firma, CUFE ni reglas
fiscales. Falta publicar un candidato inmutable y ejecutar negativos autenticados
sin destinos sensibles. P109-009 permanece **parcial** y el porcentaje formal
no aumenta con una subcompuerta todavía no desplegada.
