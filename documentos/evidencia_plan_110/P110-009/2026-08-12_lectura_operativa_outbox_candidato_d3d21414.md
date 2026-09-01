# P110-009 — lectura operativa de worker y outbox del candidato actual

Fecha: 2026-08-12  
Ambiente: staging. Alcance estrictamente de lectura de salud, estado de
contenedores y métricas agregadas. No se reintentaron eventos, no se ejecutaron
pagos, no se modificó PCS y producción no fue tocada.

## Candidato y disponibilidad

Las imágenes activas de backend, worker, frontend y ClamAV coincidieron con
los digests inmutables del candidato `d3d21414`. Los cinco contenedores de
staging estaban saludables; `/health` y `/ready` aprobaron. Worker y backend
no tenían reinicios y sus últimos 24 h no mostraron líneas de error, fallo de
worker ni pánico en el filtro agregado.

La consulta interna de métricas confirmó heartbeat reciente, cero elementos
`ready`, `processing` o con lease vencido en outbox y jobs asíncronos, y ClamAV
configurado como obligatorio sin bypass.

## Hallazgo que bloquea cierre

La métrica agregada de outbox empresarial informa **dos** elementos en estado
`dead`. El API público de staging no expone `/metrics`; la métrica se consultó
solo desde el loopback del contenedor. Por tanto, el valor no se confunde con
una lectura pública ni se atribuye a una empresa, tópico o pago concreto.

No se realizó una reactivación: el flujo protegido de recuperación exige
previsualizar eventos exactos, confirmar empresa y tópico, motivo auditado y
una acción explícita de superadministración. Reintentar sin esa conciliación
podría duplicar o alterar un efecto contable.

## Recuperación controlada de CxP en PCS

La vista previa autenticada de recuperación mostró un único evento `dead` de
CxP perteneciente a PCS, del tópico habilitado, sin recuperación previa y con
la causa histórica de handler no habilitado. La primera mutación sin token CSRF
recibió `403`; esa negativa confirmó la protección y no produjo cambios.

Con el token CSRF legítimo de la misma sesión se ejecutó una sola reactivación
por el endpoint auditado, con empresa, tópico, identificador y razón explícitos.
Respondió `200`, reactivó una unidad y la vista previa posterior quedó en cero.
Tras esperar al worker, `/health` y `/ready` aprobaron y no quedaron jobs
asíncronos dead. La métrica global conserva un outbox dead de otro alcance; no
se consultó ni modificó porque no pertenece a la recuperación PCS autorizada.

## Siguiente paso obligatorio

La conciliación CxP de PCS debe incorporar la recuperación auditada y confirmar
que el asiento/efecto correspondiente aparece una sola vez. El outbox global
restante requiere responsable y alcance propios; no debe borrarse ni reintentarse
desde una empresa ajena. Después se repite alerta y simulacro P110-009.

**Resultado:** P110-009 sigue **parcial**; el estado global continúa **NO-GO**.
