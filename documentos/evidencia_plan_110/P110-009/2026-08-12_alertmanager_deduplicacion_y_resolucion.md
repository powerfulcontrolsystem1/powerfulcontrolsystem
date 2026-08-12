# P110-009 — Deduplicación y resolución de Alertmanager

Fecha: 2026-08-12 (America/Bogota)  
Entorno: VPS de staging; sin datos empresariales, ventas ni producción.

## Simulacro aislado

Se publicó dos veces la misma alerta sintética de nivel informativo con una
identidad de deduplicación exclusiva de P110. Alertmanager respondió `200` en
ambas solicitudes, pero su API mostró **una** alerta activa para esa identidad.
Se publicó después la resolución de la misma alerta: respondió `200` y la
consulta posterior mostró **cero** alertas activas de ese caso.

| Control | Resultado |
|---|---:|
| Publicación 1 | HTTP 200 |
| Publicación 2 idéntica | HTTP 200 |
| Alertas activas con la misma identidad | 1 |
| Resolución | HTTP 200 |
| Alertas restantes del caso | 0 |

## Límites

La prueba demuestra deduplicación y resolución internas de Alertmanager. La
entrega SMTP externa fue acreditada anteriormente hasta la aceptación del
servidor; todavía falta que el responsable confirme visualmente el aviso y su
resolución, además de guardia/on-call, carga mutante de cajas y el simulacro
operativo completo. P110-009 permanece **parcial** y el veredicto global es
**NO-GO**.
