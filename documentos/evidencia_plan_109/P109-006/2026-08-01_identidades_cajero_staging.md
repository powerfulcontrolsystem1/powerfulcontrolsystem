# P109-006 - Estado de identidades para cuatro cajas

Fecha: 2026-08-01
Entorno: staging, PCS (`empresa_id=12`)

Una consulta agregada, sin correos, tokens ni contraseñas, confirmó los cuatro
usuarios temporales creados por el flujo oficial:

| Control | Resultado |
| --- | ---: |
| Usuarios `P108-C1` a `P108-C4` | 4 |
| Rol Cajero | 4 |
| Activos | 0 |
| Correo confirmado | 0 |
| Contraseña configurada | 0 |

No se activaron por SQL ni se falsificó confirmación. La corrida de cuatro cajas
continúa bloqueada hasta que los cuatro enlaces se abran por el flujo normal,
se configuren contraseñas y cada identidad pueda iniciar su sesión independiente.

Estado: **P109-006 pendiente**; este preflight no aumenta el porcentaje.

## Verificación del buzón y enlace (2026-08-01)

El buzón autorizado contiene las cuatro invitaciones independientes y confirma
la entrega real desde Mailu. Los mensajes históricos del 30 de julio tenían el
host público de producción aunque las cuentas pertenecen a staging; no se abrió
ninguno contra producción. El runtime actual de staging ya declara
`PUBLIC_BASE_URL=https://staging.powerfulcontrolsystem.com`.

El usuario PCS no tiene actualmente el permiso efectivo `seguridad:A` para
reenviar las invitaciones: el endpoint oficial respondió 403 antes de mutar el
token o enviar correo. No se elevó el permiso ni se usó SQL. Al conservar el
token emitido originalmente y dirigirlo al host correcto de staging, la primera
invitación abrió correctamente `Completar invitación` con correo precargado,
documento, nueva contraseña, confirmación, contrato y botón final.

La primera pantalla quedó entregada al usuario para completar. Crear la
contraseña y aceptar el contrato es una acción personal que no se envió de forma
automatizada. Después deben completarse las otras tres invitaciones y recién
entonces ejecutar cuatro sesiones/cajas independientes. P109-006 continúa
**pendiente** y no aumenta el porcentaje.
