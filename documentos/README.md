# Indice documental del proyecto

Fecha: 2026-04-30
Estado: vigente

Este archivo organiza la lectura tecnica y funcional del repositorio para desarrollo, soporte y trabajo asistido por Copilot.

## Orden de lectura recomendado

1. `documentos/descripcion_del_proyecto`
2. `documentos/estructura_bd.md`
3. `documentos/diagramas/estructura_del_codigo.md`
4. `documentos/descripcion_de_modulos`
5. `documentos/matriz_roles_permisos_pos_multiempresa.md`
6. `documentos/gobernanza_tecnica/README.md`
7. `documentos/historial_de_cambios`
8. `CHANGELOG.md`

## Estado documental reciente

- 2026-04-30: pagos Epayco documentados con Smart Checkout v2 y fallback clasico firmado por POST a `secure.payco.co`.
- 2026-04-30: pagos Epayco actualizados para separar modo Smart Checkout y modo fallback clasico; el POST clasico usa `Customer ID` + `P_KEY` para decidir produccion/pruebas y evitar "El comercio no fue reconocido".
- 2026-04-30: chat IA actualizado para exportar respuestas y conversaciones como PDF, DOCX, XLSX, TXT o JSON mediante el generador dinamico y auditoria de origen `chat_ia`.
- 2026-04-30: chat flotante documentado con robot IA, secretaria IA estilo caricatura ejecutiva joven y voz femenina automatica para secretaria.
- 2026-04-30: empresas compartidas documentadas con listado/revocacion de administradores compartidos y trazabilidad.
- 2026-04-30: hoja de vida operativa universal documentada para motos de taller, pacientes, vehiculos, equipos, activos o mascotas.
- 2026-04-30: documentos dinamicos asistidos por IA documentados con endpoints `/generate` y `/download`.

## Fuentes canonicas por tema

- Vision funcional y alcance actual: `documentos/descripcion_del_proyecto`
- Esquema fisico de base de datos: `documentos/estructura_bd.md`
- Arquitectura tecnica y mapa de archivos: `documentos/diagramas/estructura_del_codigo.md`
- Evolucion funcional por modulo: `documentos/descripcion_de_modulos`
- Matriz de roles, visibilidad y wrappers: `documentos/matriz_roles_permisos_pos_multiempresa.md`
- Inventario documental y de archivos: `documentos/descripcion_de_archivos`
- Historial detallado de trabajo: `documentos/historial_de_cambios`
- Resumen ejecutivo de cambios: `CHANGELOG.md`
- Gobernanza tecnica, ADRs, contratos y runbooks: `documentos/gobernanza_tecnica/README.md`
- Runbook actualizado de pagos de licencias: `documentos/gobernanza_tecnica/runbooks/runbook_checkout_licencias.md`
- Contrato actualizado de checkout publico: `documentos/gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md`
- Contrato de documentos dinamicos IA: `documentos/gobernanza_tecnica/contratos/contrato_documentos_dinamicos_ia_exportacion.md`

## Paquetes documentales complementarios

- `documentos/erp_multiempresa/`: paquete formal de alcance, diseno tecnico, especificaciones funcionales y guia de implementacion ERP multiempresa.
- `documentos/manual_de_instalacion.md`: referencia de instalacion y arranque.
- `documentos/manual_vps_seguridad.md`: operacion y endurecimiento de VPS.
- `documentos/deploy_nginx_reverse_proxy_vps.md`: publicacion HTTPS y proxy reverso.
- `documentos/actualizaciones_del_repositorio.md`: historial de sincronizacion tras `scripts/actualizar_repositorio.ps1`.

## Regla de uso para cambios tecnicos

Antes de cambiar codigo, infraestructura o flujos criticos:

1. Leer `documentos/descripcion_del_proyecto`.
2. Leer `documentos/estructura_bd.md` si hay impacto de datos, tablas o consultas.
3. Leer `documentos/diagramas/estructura_del_codigo.md` si hay impacto de arquitectura o rutas.
4. Leer `documentos/gobernanza_tecnica/estandares_de_cambio_seguro.md`.
5. Consultar el ADR, contrato tecnico o runbook aplicable cuando exista.
