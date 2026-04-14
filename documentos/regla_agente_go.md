Regla: "Todo se construye como real"
=================================

Fecha: 2026-04-11

Descripción:

Esta regla establece que todas las páginas, menús, subpáginas y componentes nuevos creados por el equipo o por el agente deben construirse como flujos "reales": es decir, con endpoints HTTP backend, persistencia en base de datos, validaciones y pruebas mínimas de integración, en lugar de permanecer como mocks o placeholders indefinidos.

Alcance:

- Páginas administrativas bajo `web/administrar_empresa/` (productos, reportes, configuración, etc.).
- Formularios y vistas que recolectan datos del usuario (categorías, precios, permisos, integraciones).
- Integraciones externas (Wompi, Gmail, DeepSeek, otros): el diseño debe incluir handlers backend y almacenamiento cifrado de credenciales cuando apliquen.

Requisitos de implementación:

1. Endpoint backend: Por cada formulario o recurso, debe existir un endpoint REST (o handler HTTP) en `backend/handlers` que reciba, valide y persista datos.
2. Persistencia: Los datos deben guardarse en el esquema DB correspondiente (empresa_id cuando aplique), siguiendo la regla estándar de tablas (campos id, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones).
3. Cifrado de secretos: Si se almacenan credenciales sensibles, usar cifrado con `CONFIG_ENC_KEY` (AES-GCM recomendado). No guardar secretos en texto plano.
4. Tests: Añadir pruebas unitarias mínimas para el handler y, cuando sea práctico, una prueba de integración que valide la ruta feliz del guardado.
5. Documentación: Actualizar `documentos/descripcion_de_modulos` y `documentos/descripcion_de_archivos` describiendo el nuevo recurso/archivo creado, y añadir entrada en `documentos/historial_de_cambios` con fecha y archivos modificados.
6. UI: Los placeholders pueden permanecer como vistas iniciales, pero deben incluir la implementación del cliente (fetch/XHR) que consuma el endpoint real y manejar estados (cargando, éxito, error, validación).

Excepciones:

- Prototipos tempranos que estén explícitamente marcados como "mock no persistente" en su documentación y aprobados por el equipo podrán permanecer sin persistencia por un plazo limitado; se debe registrar la excepción en `documentos/historial_de_cambios`.

Motivación:

Construir "como real" evita desfase entre frontend y backend, garantiza trazabilidad y facilita pruebas end-to-end y despliegues en ambientes de staging/producción.

Referencias:

- `documentos/estructura_bd.md` — esquema canonico y campos estándar.
- `copilot-instructions.md` — gobernanza documental y reglas del agente.
- `backend/handlers` — ubicación objetivo para nuevos handlers.

Responsable: agente_go / equipo de desarrollo (según contexto de la tarea)


Regla: "Base de datos operativa en VPS con PostgreSQL"
========================================================

Fecha: 2026-04-13

Descripción:

A partir de esta fecha, la base de datos operativa del sistema deja de ejecutarse en local y pasa a operar en el servidor virtual (VPS) usando PostgreSQL como motor relacional principal.

Reglas obligatorias:

1. Fuente de verdad de datos: toda operación productiva debe usar PostgreSQL en VPS.
2. Las bases SQLite locales quedan como legado para migración, respaldo controlado o pruebas puntuales, pero no como runtime productivo.
3. Toda migración se ejecuta por etapas y con validación por base:
	- Etapa 1: `superadministrador`.
	- Etapa 2: `empresas`.
	- Cada etapa debe validar consistencia mínima (conteo por tabla) antes de continuar.
4. Toda intervención en esquema/datos debe dejar trazabilidad en:
	- `CHANGELOG.md`
	- `documentos/historial_de_cambios`
	- documentación técnica asociada (`documentos/estructura_bd.md`, `estructura_bd.md`, `documentos/diagramas/estructura_del_codigo.md`).
5. Si se requiere rollback, debe realizarse desde respaldo previo y quedar documentado con fecha, motivo y alcance.

Responsable: agente_go / equipo de desarrollo (según contexto de la tarea)
