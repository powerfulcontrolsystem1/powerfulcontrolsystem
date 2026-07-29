# P108-018 - Simulacro del circuito de alertas en staging

Fecha: 2026-07-29  
Ambiente: VPS de staging/monitoreo aislado  
Alcance: Prometheus y Alertmanager locales; sin entrega externa.

## Configuración validada

- Prometheus y Alertmanager están ligados a `127.0.0.1` en el host.
- Prometheus reconoce la configuración de Alertmanager.
- El receptor `observabilidad-interna` no envía correo, webhook ni datos de
  empresas a un tercero.
- El target `pcs-staging-backend` está `up` y consulta `/metrics` del candidato
  inmutable `41be623ad2ed6c10ff86027063870b0848db2af1`.

## Simulacro

La regla existente `PCSBackendCaido`, con ventana de dos minutos, se evaluó
contra el target productivo que sigue sin publicar el endpoint de métricas.
No se modificó el backend de producción ni se generó una operación de negocio.

| Verificación | Resultado |
| --- | --- |
| Prometheus configuró Alertmanager | Sí |
| Estado inicial de la regla | `pending` |
| Estado tras la ventana | `firing` |
| Alerta recibida por Alertmanager | `PCSBackendCaido` |
| Correo/webhook externo | No enviado |

## Límite y acción pendiente

P108-018 permanece **parcial**. Falta definir responsables y canal de
escalamiento aprobado, probar recuperación/resolve, simular worker, PostgreSQL,
almacenamiento y colas, y corregir el scrape del backend productivo mediante un
candidato de producción aprobado. No se ocultó ni se desactivó la alerta del
backend productivo: su estado refleja que ese artefacto aún no publica
`/metrics`.
