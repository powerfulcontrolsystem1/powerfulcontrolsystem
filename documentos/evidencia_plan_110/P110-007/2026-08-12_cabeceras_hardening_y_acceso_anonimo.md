# P110-007 — cabeceras, acceso anónimo y hardening de staging

Fecha: 2026-08-12  
Ambiente: staging y VPS; solo lecturas y solicitudes no autenticadas. No hubo
mutaciones de negocio ni cambios de infraestructura.

## Resultado técnico

La auditoría estática de seguridad aprobó los contratos de cookies seguras,
revocación de sesión, allowlist pública, límite de login, CORS sin wildcard,
scope empresarial y ausencia de archivos runtime versionados. La auditoría de
observabilidad y SLO también aprobó sus controles estáticos.

La comprobación dinámica desde un cliente no autenticado verificó que CxP/IA,
CxP financiera, facturación electrónica y contexto de permisos devuelven
`401`. Un `OPTIONS` desde origen externo también devolvió `401`.

La respuesta HTTPS pública contiene `HSTS`, `nosniff`, política de referente y
`X-Frame-Options: SAMEORIGIN`. El VPS mantiene SSH sin contraseña y login root
restringido, UFW activo, servicios de staging sanos y uso de disco de 69 %.

## Hallazgos abiertos

- La CSP pública conserva `unsafe-inline` para scripts y estilos, además de los
  orígenes externos necesarios de los proveedores. Requiere reducción gradual
  o una excepción formal con responsable y vencimiento antes de certificar.
- El chequeo VPS informó que `fail2ban` no está habilitado. No se instaló ni se
  cambió infraestructura en esta ejecución; su activación exige ventana,
  configuración de exclusiones y verificación operativa.

P110-007 continúa **parcial**: faltan DAST integral autenticado, A/B de todos
los dominios, pruebas de archivo avanzadas, cierre CSP y el control operativo
pendiente de hardening.
