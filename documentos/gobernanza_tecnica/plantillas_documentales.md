# Plantillas de información técnica

Estado: Vigente. Responsable: Coordinación técnica. Revisión documental: 2026-09-05.

## Uso

Adaptar estas estructuras dentro de la fuente principal del tema. Crear archivo independiente solo cuando su propósito, audiencia o ciclo de vida lo justifiquen. Sustituir todos los campos antes de marcar un documento como vigente; no registrar una plantilla vacía como evidencia.

## Documento técnico

```text
# Título específico
Estado: Propuesto / Vigente / Sustituido.
Responsable: rol; persona en directorio privado si corresponde.
Revisión documental: AAAA-MM-DD.
Propósito, audiencia, alcance y exclusiones.
Fuentes y candidato/entorno cuando se describe una implementación.
Contenido por tema, reglas y ejemplos seguros.
Aceptación, evidencia y limitaciones.
Referencias relacionadas; sustitución/retirada si aplica.
```

## Requisito y contrato

```text
ID estable; origen; prioridad; responsable.
Actor y precondiciones.
Obligación observable; entradas/salidas; errores y límites.
Tenant, IDs secundarios, rol, permiso y licencia.
Datos, transacción, idempotencia, concurrencia y efectos externos.
Compatibilidad y consumidores.
Casos positivos/negativos y criterio de aceptación.
Vínculos a diseño, código, pruebas y evidencia por candidato.
Estado de implementación distinto del estado de aceptación.
```

## ADR

```text
# ADR-NNNN — Decisión
Estado: Propuesta / Aceptada / Sustituida por ADR-NNNN.
Fecha; responsables; alcance autorizado.
Contexto, restricciones e interesados.
Alternativas consideradas y criterios de elección.
Decisión y justificación.
Consecuencias, riesgos, costos y compatibilidad.
Plan de adopción/reversión y evidencia.
Referencias a requisitos, implementación y ADR relacionados.
```

## Runbook

```text
Síntoma/objetivo; entorno; responsable y escalamiento.
Precondiciones y acceso necesarios, sin secretos.
Cómo identificar entorno/candidato y descartar condiciones peligrosas.
Pasos numerados con resultado esperado y criterio de detenerse.
Efectos persistentes/externos y alcance requerido.
Recuperación/rollback y compatibilidad de datos.
Verificación final, registro de evidencia y siguiente escalamiento.
Último ensayo: fecha, candidato, entorno, resultado o pendiente.
```

## Informe de prueba, release o postincidente

```text
ID; fecha/zona; autor/revisor; objetivo y requisitos.
SHA/digests y estado del árbol; configuración/entorno identificados sin secretos.
Datos y alcance autorizado; herramientas/versiones.
Comandos/casos; resultado esperado y observado.
Aprobadas, fallidas y omitidas con motivo.
Evidencia minimizada y ubicación privada cuando corresponda.
Impacto, decisiones y recuperación (si incidente/release).
Pendientes con responsable y criterio de cierre.
Veredicto limitado al alcance; aprobación real si aplica.
```
