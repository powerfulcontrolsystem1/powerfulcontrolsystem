# P109-001 - Negativos autenticados del recuperador CxP

Fecha: 2026-08-01
Entorno: staging, empresa PCS (`empresa_id=12`)
Candidato: `6da6c13453a40b2d84e23285fa83f255f34788da`

La sesión de superadministración consultó la vista previa oficial. El resultado
contenía un solo evento histórico excluido del alcance, no expuso
`payload_json` y conservó únicamente campos operativos sanitizados.

La API respondió:

| Caso | Resultado |
| --- | ---: |
| `empresa_id` ausente | 400 |
| topic no permitido | 400 |
| IDs repetidos | 400 |
| ID inexistente | 409 |
| ID ya publicado | 409 |

Después de las pruebas, el evento excluido continuó `dead`, con intentos `5/5`;
no se reactivó ni se creó un pago o asiento. La recuperación histórica previa
conserva una auditoría que agrupó los dos eventos autorizados ya conciliados.

No existe una segunda empresa controlada válida en este staging. La prueba con
un identificador sin empresa devolvió 404 antes de consultar eventos; por ello
la matriz A/B operativa completa permanece pendiente y no se presenta como
aprobada.

Estado: **P109-001 parcial**.
