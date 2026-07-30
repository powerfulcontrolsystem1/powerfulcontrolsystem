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

## Simulacro de caída y recuperación 2026-07-30

Se detuvo únicamente `pcs-staging-backend` durante 145 segundos mediante un
proceso con reinicio automático. El ensayo no tocó producción, bases de datos
ni operaciones de negocio.

| Verificación | Resultado |
| --- | --- |
| Target staging antes del ensayo | `up=1` |
| Target durante la caída | `up=0` |
| Estado de regla | `pending` y luego `firing` |
| Recepción en Alertmanager | `PCSBackendCaido`, job staging |
| Reinicio automático | Correcto |
| Salud posterior | `/health=ok`, `/ready=ready` |
| Target recuperado | `up=1` |
| Alerta activa posterior | Ninguna para staging |
| Entrega externa | No configurada/no enviada |

El subcriterio de caída y resolución del backend de staging queda **PASS**.
P108-018 continúa parcial por worker, PostgreSQL, almacenamiento, colas,
responsables y canal de escalamiento.

## Extensión de observabilidad operativa del candidato 2026-07-30

La revisión del endpoint `/metrics` confirmó que solo publicaba
`pcs_backend_up`. Ese dato no permitía distinguir un proceso HTTP vivo con
PostgreSQL caído, un worker detenido o una cola durable bloqueada.

Se amplió el candidato local con consultas agregadas, acotadas por un contexto
de dos segundos y sin etiquetas de empresa, payloads, errores de proveedor,
credenciales ni datos personales. Las nuevas series cubren:

- disponibilidad separada de los pools PostgreSQL de negocio y super;
- edad del último trabajo `maintenance.system-metrics` completado por el worker;
- elementos listos, en proceso, dead-letter y leases vencidos en outbox;
- trabajos durables listos, en proceso, dead-letter y leases vencidos;
- éxito o fallo de cada consulta operativa agregada.

El frontend bloquea ahora exactamente `/metrics` con HTTP 404. Prometheus
conserva el scrape directo del backend por las redes Docker privadas. El
backend reutiliza el agregado durante diez segundos para impedir que múltiples
solicitudes provoquen consultas repetidas a PostgreSQL.

Se añadieron reglas para PostgreSQL no disponible, worker sin latido, cola
acumulada, lease vencido y consulta operativa fallida. También se agregaron seis
paneles a Grafana para hacer visibles estas señales.

### Evidencia ejecutada

| Verificación | Resultado |
| --- | --- |
| Pruebas enfocadas de `/metrics` | PASS |
| Compilación de paquetes principales Go | PASS |
| Auditoría estática de observabilidad | PASS |
| `promtool check rules` en contenedor Linux del VPS | PASS, 8 reglas |
| Consulta SQL de negocio sobre outbox | `0/0/0/0` |
| Consulta SQL super sobre outbox | `0/0/0/0` |
| Consulta SQL super sobre trabajos durables | `0/0/0/0` |
| Edad observada del latido worker | 7 segundos |
| Datos de empresa o payload publicados | Ninguno |

Las consultas SQL se ejecutaron en modo lectura contra el runtime autorizado
para demostrar que la sintaxis y las tablas existen. El endpoint ampliado y las
reglas nuevas todavía deben publicarse en un digest inmutable de staging para
simular PostgreSQL, worker, acumulación y lease vencido. P108-018 permanece
**parcial** y no se considera certificada por esta evidencia local.
