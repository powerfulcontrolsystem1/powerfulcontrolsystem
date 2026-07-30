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

## Digest inmutable y scrape real 2026-07-30

GitHub Actions construyó, escaneó, generó SBOM y publicó el commit
`cf49fc7cefb083e1ac8df1711f05a0f8a22c8afb` en cuatro imágenes inmutables:

- API: `sha256:db738706efeaa50aaa3b53646458cb70a540ddac0a10f00f38d01c156743cc0e`;
- migrador: `sha256:443b573e51d68d713102b6fcf2a851b112eca48cd586ef43bad5d8b45f6bf3ee`;
- worker: `sha256:111879c2307fcc6122e36d10bba733356b0e1883025cea656d942229ee61f00d`;
- frontend: `sha256:10841ec48035b8550cea690b4a378d0246098069b29c9304e639266dad6b4ad0`.

La promoción inicial detectó que el VPS conservaba archivos Compose anteriores
que no incluían el frontend por digest ni las credenciales del rol runtime para
el migrador. La operación se detuvo; producción no se tocó. Staging se recuperó
usando los tres archivos Compose exactos del candidato, el migrador terminó con
código cero y los cinco servicios quedaron saludables. El script versionado
ahora valida los cuatro digests renderizados antes de recrear servicios y limita
el `up` a PostgreSQL, migrador, API, worker y frontend.
El preflight corregido se ejecutó contra los Compose desactualizados del VPS y
los rechazó por faltar el digest exacto del frontend antes de ejecutar `pull` o
`up`; staging permaneció con salud y readiness 200.

### Resultado del candidato activo

| Señal | Resultado |
| --- | --- |
| `/health` y `/ready` externos | 200 / 200 |
| `/metrics` por frontend público | 404 |
| target privado de Prometheus | `up=1` |
| PostgreSQL negocio/super | `1 / 1` |
| edad observada del worker | 7,228 segundos |
| outbox listo / leases vencidos | `0 / 0` |
| trabajos listos / leases vencidos | `0 / 0` |
| consultas operativas | `1` en las tres fuentes |
| expresiones sanas de las cinco alertas | conjunto vacío, sin falsos positivos |

La autenticación autorizada, selección de Powerful Control System y panel de la
empresa cargaron visualmente con este digest. En ese corte P108-018 seguía
parcial por worker, PostgreSQL, leases, configuración permanente y
escalamiento; la sección siguiente registra el cierre posterior de worker y
configuración.

## Configuración permanente y simulacro del worker 2026-07-30

Las ocho reglas y la versión 2 del tablero se instalaron en el stack compartido
de monitoreo. Antes de reemplazar los archivos se conservaron copias con sufijo
`bak.p108.20260730_203424`. Prometheus se recreó sin borrar su volumen histórico
para remontar el archivo actualizado; Grafana observó el tablero con SHA-256
`0b6ea8689d299020bb4b6cdfece07e52a8eff838edc9dc9a580000547f77e0bb`.

Se detuvo exclusivamente `pcs-staging-worker` durante 270 segundos mediante un
proceso con reinicio automático. No se detuvo API, frontend, PostgreSQL ni
producción.

| Verificación | Resultado |
| --- | --- |
| Estado inicial | worker saludable, colas listas 0, leases vencidos 0 |
| Edad sin latido antes del umbral | 69,406 s, sin alerta |
| Estado al superar 120 s | `pending`, una instancia |
| Edad antes de disparar | 204,407 s, aún `pending` |
| Estado tras ventana de dos minutos | `firing`, una instancia |
| Recepción en Alertmanager | 1 `PCSWorkerSinLatido` de staging |
| Reinicio automático | Correcto |
| Edad posterior del latido | 9,683 s |
| Estado posterior de la regla | `inactive`, cero instancias |
| Alerta activa posterior en Alertmanager | 0 |
| Colas y leases posteriores | todos en 0 |
| Aplicación posterior | `/health=ok`, `/ready=ready` |

El subcriterio de caída, detección, recepción y resolución del worker queda
**PASS**. En ese corte P108-018 seguía parcial por PostgreSQL, leases,
almacenamiento y escalamiento; la sección siguiente registra el simulacro
posterior de PostgreSQL.

## Simulacro de PostgreSQL staging 2026-07-30

Antes del ensayo se comprobó PostgreSQL y worker saludables, cero sesiones
esperando lock y el snapshot restaurable
`/root/powerfulcontrolsystem/backups/vps-snapshots/20260730_194951`.

Se ejecutó `docker stop` limpio únicamente sobre
`pcs-staging-postgres`, con reinicio automático después de 165 segundos. La
aplicación de producción y su base no participaron.

| Verificación | Resultado |
| --- | --- |
| PostgreSQL staging durante el ensayo | detenido |
| `/health` durante la caída | 200 |
| `/ready` durante la caída | 503 |
| Estado inicial de la regla | `pending`, dos pools |
| Estado tras ventana | `firing`, dos pools |
| Alertas recibidas por Alertmanager | 2, negocio y super |
| Reinicio automático | correcto, contenedor saludable |
| Estado posterior de la regla | `inactive`, cero instancias |
| Alertas activas posteriores | 0 |
| PostgreSQL negocio/super posterior | `1 / 1` |
| Migraciones aplicadas negocio/super | `41 / 13` |
| Worker posterior | latido 4,767 s |
| Colas listas y leases vencidos | todos en 0 |
| Aplicación posterior | `/health=ok`, `/ready=ready` |

El subcriterio de caída, readiness, alerta, recepción y recuperación de
PostgreSQL queda **PASS**. P108-018 permanece **parcial** por simulacro de lease,
señal de almacenamiento privado y canal externo/responsables de escalamiento.
