# P109-005/P109-006 - PDF real y prerrequisito de cuatro usuarios PCS

Fecha: 2026-08-09. Empresa: Powerful Control System (`empresa_id=12`).

## Impresión virtual autenticada

Se inició sesión por el flujo oficial, se abrió la factura electrónica real
`1PCS6` y Chrome generó una impresora virtual PDF A4. La primera captura encontró
logo roto, estilos dependientes de red, tabla recortada y metadatos DIAN
comprimidos. El candidato vuelve autosuficiente la hoja imprimible, usa como
respaldo el logo oficial y limita tablas, CUFE, URL y observaciones al ancho de
la página.

La repetición contra los datos reales de PCS produjo dos páginas A4, cero
imágenes rotas y revisión visual PASS: logo, encabezado, cuatro metadatos,
cliente, resumen, cinco columnas, totales, control documental, observaciones y
QR DIAN quedaron legibles y sin recortes. El PDF está en
`output/pdf/P109_factura_real_1PCS6.pdf`. No se creó, reenvió ni anuló otro
documento fiscal.

El barrido sintético del mismo candidato aprobó además 20/20 formatos Carta/POS,
cero casos en revisión y cero fallos de autoimpresión.

Límite: la hoja candidata se sustituyó localmente en Chrome mientras consumía
las APIs reales; todavía debe desplegarse y repetirse desde el artefacto servido
por staging antes de acreditar cierre productivo.

## Cuatro identidades independientes

La inspección visual del flujo oficial de usuarios encontró seis registros en
PCS: uno confirmado y activo, dos activos pendientes de confirmación y tres
inactivos pendientes. No existen hoy cuatro identidades confirmadas que permitan
cuatro sesiones independientes reales.

Antes de emitir nuevas invitaciones se comprobó el buzón propio de PCS. La
cuenta aparece provisionada, pero el contador IMAP informa rechazo. El acceso
automático reproducible termina en `302 -> 302 -> 403` al llegar a
`/webmail/sso.php`; por eso no se enviaron invitaciones adicionales que el
usuario no pudiera confirmar.

El candidato alinea el `Host`, `X-Forwarded-Host` y `X-Forwarded-Proto` de la
llamada SSO interna con el host público de Mailu. La prueba enfocada de backend
aprueba y la sonda autenticada conserva evidencia saneada sin cookies, tokens ni
claves. Falta desplegar este arreglo, confirmar que webmail/IMAP responden y
entonces crear y confirmar tres cuentas por el flujo de invitación antes de
repetir cuatro sesiones y cuatro cajas.

## Estado

P109-005 sigue parcial por despliegue del candidato, accesibilidad asistida y
prueba física. La aprobación histórica P109-006 en staging no se invalida, pero
la repetición solicitada sobre el PCS servido queda bloqueada por el canal de
correo. El Plan 109 conserva **56,7 % de implementación** y **NO-GO**.
