# Politica de seguridad

## Reporte responsable

No publique vulnerabilidades, secretos ni datos de empresas en issues publicos.
Reporte de forma privada al canal de soporte corporativo indicado en el portal
oficial, incluyendo pasos reproducibles, impacto y evidencia minimizada.

## Manejo inicial

1. Registrar el reporte sin copiar tokens, cookies, documentos ni datos de clientes.
2. Clasificar severidad, alcance multiempresa y posibilidad de explotacion.
3. Contener acceso afectado: revocar sesiones, credenciales o webhook cuando corresponda.
4. Corregir, probar aislado y desplegar mediante revision aprobada.
5. Documentar causa, rollback y comunicacion a clientes si existe impacto.

## Regla de secretos

Los secretos solo viven en el gestor o archivo de entorno no versionado del
servidor. Nunca se incluyen en commits, reportes, capturas, logs ni tickets.
