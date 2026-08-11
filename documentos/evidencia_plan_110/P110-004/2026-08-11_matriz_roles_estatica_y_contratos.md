# P110-004 — matriz estática de roles y contratos de aislamiento

Fecha: 2026-08-11  
Ámbito: código del candidato y pruebas focalizadas locales. Sin mutaciones
operativas ni acceso a datos empresariales.

## Inventario de rutas y roles

`tools/qa_roles_matrix.mjs --strict` aprobó los seis perfiles inventariados:
superadministrador, administrador de empresa, cajero, vendedor, asesor
comercial y soporte. Para cada perfil confirmó que existen las páginas y rutas
API declaradas y que el workflow E2E contiene las variables de autenticación y
viewport requeridas.

## Contratos ejecutados

Las pruebas focalizadas Go aprobaron para:

- rechazar discrepancias de `empresa_id` por query, header, JSON, formularios y
  parámetros repetidos;
- exigir aprobación en cambios de seguridad;
- permitir al contador consultar reportes/IA sin administración;
- exigir habilitación empresarial explícita para páginas IA ocultas;
- limitar la API manual de Finanzas para el cajero;
- impedir elevación de rol basada solo en el correo.

## Límite

P110-004 sigue pendiente: faltan identidades activas por rol, pruebas
autenticadas de acciones permitidas/denegadas, cuatro cajas y decisiones
explícitas de inclusión o exclusión de domótica. El inventario y los contratos
no prueban por sí solos efectos reales de negocio.
