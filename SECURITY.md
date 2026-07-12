# Politica de seguridad

## Versiones soportadas

La rama `main` y la ultima version publicada son las unicas lineas con
correcciones de seguridad. Las ramas de trabajo, pruebas y presentacion no son
canales de despliegue ni de soporte.

## Reporte privado

No publique vulnerabilidades, secretos ni datos de empresas en issues,
discusiones, capturas o Pull Requests publicos. Reporte de forma privada por el
canal corporativo de seguridad indicado en el portal oficial. Incluya una
descripcion minimizada, versiones afectadas, pasos reproducibles y evidencia
sin tokens, cookies, documentos ni informacion de clientes.

## Respuesta y alcance

Se acusa recibo inicial dentro de cinco dias habiles. Se priorizan fallas de
autenticacion, autorizacion, aislamiento multiempresa, pagos, facturacion,
archivos privados, integraciones, infraestructura y exposicion de secretos.
No estan permitidas pruebas sobre datos reales, denegacion de servicio,
ingenieria social, acceso a cuentas ajenas ni exploracion fuera del alcance
autorizado.

## Divulgacion coordinada

La vulnerabilidad debe mantenerse privada hasta que exista correccion,
validacion y una fecha de divulgacion coordinada. PCS no solicita publicar
detalles tecnicos en issues publicos antes de corregirlos.

## Manejo de incidentes

1. Registrar el evento sin copiar datos sensibles.
2. Clasificar severidad, alcance empresarial y posibilidad de explotacion.
3. Contener acceso: revocar sesiones, credenciales, tokens o webhooks cuando
   corresponda.
4. Corregir mediante revision, pruebas aisladas y cambios trazables.
5. Comunicar impacto, rollback y medidas preventivas a las partes afectadas
   cuando aplique.

## Politica de actualizacion

Las dependencias, imagenes y toolchains se revisan en CI. Las vulnerabilidades
alcanzables se corrigen con una version parcheada o una mitigacion documentada;
nunca se silencian con exclusiones genericas. Las credenciales se almacenan
solo fuera del repositorio, en configuracion segura del entorno.
