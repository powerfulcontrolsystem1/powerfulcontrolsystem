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
