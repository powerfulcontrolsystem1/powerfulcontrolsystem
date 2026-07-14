# Contexto especifico del sistema

Estado: vigente. Ultima actualizacion: 2026-07-13.

Este documento amplia el
[`Contexto general del sistema`](contexto_general_del_sistema.md). No se debe
leer completo por defecto: usar la seccion relacionada con la tarea.

## Inicio de cualquier cambio

- Vision y alcance: `descripcion_del_proyecto`.
- Ubicacion de paginas, APIs, tablas y pruebas: `mapa_modulos.md`.
- Flujos de negocio: `flujos_operativos.md`.
- Comandos, preflight, `rs`, SSH y VPS: `comandos_codex.md`.
- Restricciones permanentes: `decisiones_tecnicas.md`.
- Cambios previos y decisiones vigentes: `historial_de_cambios` y
  `CHANGELOG.md`.

## Seguridad, permisos y multiempresa

Abrir antes de crear o cambiar endpoints, consultas, permisos, exportes,
importaciones, backups, archivos, borrados o datos empresariales:

- `checklist_seguridad_endpoint_multiempresa.md`.
- `matriz_roles_permisos_pos_multiempresa.md`.
- `estructura_bd.md` cuando haya SQL, tablas, migraciones o datos persistentes.
- `diagramas/estructura_del_codigo.md` cuando cambien rutas, arquitectura,
  integraciones o estructura.

Reglas aplicables: resolver `empresa_id` en backend, verificar pertenencia y
permiso, filtrar SQL, sanear errores, no registrar secretos y validar llamadas
forzadas aunque el boton se oculte en frontend.

## Modulos operativos

- Ventas, carrito, caja, inventario, compras, clientes y credito:
  `mapa_modulos.md`, `flujos_operativos.md`, `descripcion_de_modulos`.
- Facturacion electronica, DIAN, XML, numeracion, representacion impresa y
  pruebas: `mapa_modulos.md`, `flujos_operativos.md`,
  `decisiones_tecnicas.md`, `referencias/dian/README.md` y los tutoriales
  empresariales bajo `web/administrar_empresa/`.
- Licencias, checkout, Epayco y Wompi:
  `gobernanza_tecnica/runbooks/runbook_checkout_licencias.md` y
  `gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md`.
- Nomina, impuestos, contabilidad y reportes:
  `descripcion_de_modulos`, `estructura_bd.md` y los documentos especificos del
  modulo en `documentos/`.
- IA, agentes y limites por empresa: `ia_orquestador_empresarial.md`,
  `mapa_modulos.md`, `estructura_bd.md` y `diagramas/diagramas_sistema_pcs.md`.
- Arquitectura modular, migraciones/worker y API movil:
  `plan_101_arquitectura_modular.md`,
  `preparacion_produccion_y_app_movil.md`, `api/mobile_api_v1.md` y
  `api/openapi.mobile.v1.yaml`.
- Correo corporativo y WhatsApp: `mapa_modulos.md`, `flujos_operativos.md`,
  `decisiones_tecnicas.md`, `email_corporativo_mailu.md` y los runbooks de
  correo aplicables.

## Infraestructura, copias y VPS

- Operacion Docker, salud y despliegue: `docker_vps_operacion.md`,
  `comandos_codex.md` y `deploy/`.
- Backups y restauracion: `comandos_codex.md`,
  `gobernanza_tecnica/runbooks/` y `scripts/crear_backup_vps.ps1`.
- Nextcloud empresarial: `nextcloud_empresarial.md`; VPS2: `vps2_operacion.md`
  y `scripts/sync_to_vps2.ps1`. No mezclar ambos servicios.

No exponer host, llaves, contrasenas, DSN ni configuracion privada local en
salidas, documentos o commits.

## Frontend y validacion visual

- Shells y estilos globales: `web/administrar_empresa.html`,
  `web/super_administrador.html`, `web/estilos.css` y `web/js/`.
- Primero buscar un patron existente antes de crear una pagina, componente o
  endpoint duplicado.
- Validar responsive, contraste, texto sin caracteres rotos, estados vacios,
  permisos y el flujo real afectado.
- Documentos impresos se validan como papel real, no como una tarjeta del tema
  actual.

## Documentacion y cierre

Todo archivo nuevo se registra en `descripcion_de_archivos`. Todo cambio
funcional se registra en `descripcion_de_modulos` e `historial_de_cambios`, y
cuando aplica en `CHANGELOG.md`, matriz de roles, estructura de datos y
diagramas. Para una tarea transversal, el cierre incluye evidencia de backend,
frontend y operacion.
