# P110-006 — bloqueo de configuración DIAN en PCS/staging

Fecha: 2026-08-11  
Ámbito: `staging`, empresa efectiva PCS (`empresa_id=12`). Producción no fue
consultada ni modificada.

## Comprobación visual autenticada

Se abrió el Centro de habilitación DIAN mediante el flujo web oficial. El
resumen visible mostró:

- Ambiente: **No configurado**.
- Estado DIAN: **No configurado**.
- TestSetId: **No configurado**.
- Rango: **No configurado**.
- Avance de validación: 10 %, con siguiente paso «Configuración base».
- Historial TrackId / ZipKey: cero registros.

La consola de la misma pantalla confirmó que la lectura de configuración
respondió correctamente pero no devolvió una configuración registrada para la
empresa. Esto es consistente con que la bandeja documental no muestre facturas
electrónicas en el periodo probado.

## Contraste de paridad, solo lectura

Se contrastaron metadatos no sensibles de la misma empresa en ambos entornos,
sin imprimir certificados, claves, NIT ni valores DIAN:

| Control | Principal PCS | Staging aislado |
| --- | ---: | ---: |
| Fila de configuración DIAN | 1 | 0 |
| Metadato/referencia de certificado | presente | ausente |
| Referencia de clave de certificado | presente | ausente |
| Ambiente, rango, TestSetId e identidad fiscal | presentes | ausentes |
| Emisión local habilitada | sí | no aplicable (sin fila) |

Por tanto, PCS sí está configurada en el entorno principal. La brecha es una
desalineación de la réplica de staging, no una ausencia de configuración de
la empresa. Copiar una clave privada fiscal al entorno de pruebas amplía el
alcance de un secreto; requiere un mecanismo de restauración seguro o una
autorización explícita y acotada antes de ejecutarse.

## Validación criptográfica del principal, sin revelar secretos

Desde la identidad de ejecución del backend principal se comprobó que las dos
referencias de archivo se resuelven a archivos no vacíos y legibles. Se analizó
la clave como RSA y el certificado como X.509; ambos análisis aprobaron. La
clave pública derivada de la privada coincide con la clave pública del
certificado. No se imprimieron rutas, nombres de archivo, NIT, huellas, fechas
ni contenido criptográfico.

Conclusión: PCS tiene el par de firma válido y accesible al backend principal.
No hay base para reemplazar, regenerar ni modificar su firma.

## Decisión segura

No se emitió una venta/factura ni se solicitó una nota crédito: en staging,
sin firma, ambiente, rango y TestSetId, no habría una aceptación oficial DIAN
válida y se crearía evidencia fiscal engañosa. La autorización de prueba queda
vigente para reanudar el flujo mínimo (factura, consulta de acuse, nota crédito
y reimpresión) únicamente después de restaurar y validar la configuración
empresarial de staging mediante un mecanismo seguro.

## Efecto en la compuerta

P110-006 continúa pendiente. No hay evidencia de aceptación fiscal, anulación
oficial, aislamiento con empresa B ni entrega de correo de factura.
