# P110-006 — auditoría de paridad DIAN en VPS

Fecha: 2026-08-11  
Ámbito: VPS, solo lectura, empresa PCS. Producción y staging no fueron
modificados.

## Control implementado

`deploy/scripts/vps-dian-parity-audit.sh` compara principal y staging para una
empresa, sin imprimir valores fiscales ni secretos. Revisa únicamente señales
booleanas de identidad, ambiente, rango, TestSetId, referencias de certificado
y clave, activación local y capacidad del backend de resolver las referencias.

## Ejecución saneada

| Señal | Principal | Staging |
| --- | --- | --- |
| Fila DIAN | presente | ausente |
| Identidad, ambiente, rango y TestSetId | presentes | ausentes |
| Referencias de certificado y clave | presentes | ausentes |
| Referencias legibles por backend | sí | no aplicable |
| Emisión local | habilitada | ausente |

La ejecución terminó correctamente y no copió, leyó, imprimió ni persistió
material criptográfico.

## Uso operativo

Ejecutar el script antes de una prueba fiscal en staging. Un resultado de
paridad incompleta bloquea la emisión hasta restaurar por una vía segura la
configuración y el material de firma autorizados para pruebas.
