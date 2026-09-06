# Manual de instalación y arranque controlado

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Se separan instrucciones actuales y migraciones históricas; cada intervención remota requiere entorno identificado y evidencia propia.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Preparar desarrollo

Leer [incorporación](desarrollo/incorporacion.md) y
[configuración de entornos](desarrollo/configuracion_y_entornos.md). Usar Git,
PowerShell, Node según CI, Go según `backend/go.mod` y PostgreSQL aislado.
Resolver dependencias existentes desde los manifiestos; este manual no autoriza
agregar nuevas dependencias. El frontend es estático y el backend Go atiende APIs.

1. Identificar checkout y cambios locales con `git status --short` y `git rev-parse HEAD`.
2. Preparar configuración privada y dos bases del entorno aislado; nunca publicar DSN o claves.
3. Comprobar aislamiento de almacenamiento y destinos externos. El worker puede despachar al arrancar.
4. Validar Compose mediante `scripts/staging_up.ps1 -ConfigOnly` cuando se usa staging.
5. En el entorno autorizado ejecutar migrador y arranque conforme al
   [runbook de staging](gobernanza_tecnica/runbooks/runbook_staging_ci_e2e.md).
6. Comprobar health/ready, después autenticación, permisos y operación del módulo.

## OAuth y sesión

El callback debe coincidir con el origen configurado para ese entorno y el
cliente OAuth autorizado. No mezclar localhost/127.0.0.1 ni dominio/www. Ante
redirect_uri_mismatch comparar la URL emitida con la configuración privada del
cliente, sin copiar secretos. La creación de sesión y aceptación contractual
se rigen por el [contrato de autenticación](gobernanza_tecnica/contratos/contrato_autenticacion_administrativa_y_usuarios_empresa.md).

## Entornos remotos y servicios opcionales

El puerto SSH y fingerprint proceden del inventario privado validado, no de
un ejemplo histórico. Un túnel no convierte producción en entorno de prueba.
[Docker/VPS](docker_vps_operacion.md) define operación y release; Nextcloud,
Mailu, voz y soporte remoto tienen guías propias en el [mapa](mapa_modulos.md).
No arrancar todos los perfiles para completar una instalación básica.

Los procedimientos anteriores de migración de host/servicios están conservados
en la [referencia histórica](historico/2026-09-05/manual_de_instalacion_referencia_acumulada.md).

## Fuentes y aceptación de la revisión

[staging_up.ps1](../scripts/staging_up.ps1), [go.mod](../backend/go.mod).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
