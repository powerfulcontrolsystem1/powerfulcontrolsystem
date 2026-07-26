# P108-011 - línea base visual autenticada de staging

Fecha local: 2026-07-25
Ambiente: `https://staging.powerfulcontrolsystem.com`
Empresa autorizada: `12`
Modo: solo lectura; cero clics, formularios, pagos, envíos, IA o impresiones.
Estado: **línea base, no certifica el candidato local**

## Cobertura

- Finanzas empresariales, escritorio y móvil: aprobado en el barrido pasivo.
- Reportes ejecutivos, escritorio y móvil: aprobado en el barrido pasivo.
- Centro IA empresarial, escritorio y móvil: revisar.
- Total: 6 pantallas, 208 botones detectados, 46 acciones riesgosas omitidas.

## Hallazgos

1. Centro IA solicitó `/api/empresa/ia_empresarial` y recibió HTTP 403 para la
   empresa autorizada. El mensaje visible fue una denegación de rol.
2. La página publicada conserva el selector antiguo de agente; no aparece el
   interruptor único `Modo agente` del candidato local.
3. La página publicada de Finanzas no contiene el botón de carga de factura o
   recibo con IA del candidato local.

## Conclusión

El staging actual está sano en `/health` y `/ready`, pero no contiene el commit
candidato. Después de desplegar el digest de P108 se debe repetir esta prueba y
resolver el 403 con una prueba de permisos efectiva, no relajando controles.
