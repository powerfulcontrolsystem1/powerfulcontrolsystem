# SLO/SLA operativo de Powerful Control System

Estado: Vigente. Responsable: QA/operación. Revisión documental: 2026-09-05.

## Objetivo

Este documento conserva los objetivos operativos internos del proyecto. No constituye un SLA contractual ni demuestra disponibilidad o recuperabilidad medidas. Su adopción comercial exige responsable, alcance, medición y aprobación específica.

## SLO principales

- Disponibilidad mensual produccion: 99.5% para backend, frontend y base de datos.
- Tiempo de respuesta API critica: p95 menor o igual a 1200 ms en operaciones de login, ventas, pagos, licencias y paneles principales.
- Errores 5xx: menor a 1% por ventana de 15 minutos en rutas criticas.
- Capacidad VPS: disco menor al 80%, memoria menor al 85% y CPU promedio menor al 80% por 10 minutos.
- Backups: respaldo externo diario verificado con retencion minima de 7 dias.

## RTO/RPO

- RTO produccion: restaurar servicio critico en 2 horas ante falla severa.
- RPO produccion: perdida maxima aceptable de datos de 24 horas si se usa respaldo diario; 1 hora cuando se active respaldo incremental.
- RTO staging: 4 horas.
- RPO staging: se puede regenerar desde produccion anonimizada.

## Severidades

- P1: caida de login, ventas, facturacion, pagos o base de datos en produccion. Atencion inmediata.
- P2: degradacion parcial, fallas de modulo critico o disco/memoria por encima de umbral. Atencion el mismo dia.
- P3: errores visuales, inconsistencias de ayuda, mejoras no bloqueantes. Atencion planificada.

## Puertas de salida

Antes de sincronizar o desplegar cambios importantes se debe ejecutar `scripts/release_gate.ps1` o `scripts/profesional_preflight.ps1 -Full`. Los reportes quedan en `documentos/reportes_profesionales`.

## Escalamiento

Las alertas del sistema deben enviar correo a `powerfulcontrolsystem@gmail.com` para capacidad, disponibilidad, trafico, usuarios/conexiones y errores criticos.

## Definición de medición

- Disponibilidad: minutos de servicio utilizable en el mes divididos por minutos
  del período observado; registrar zona horaria, rutas/sondas y cualquier
  exclusión acordada. No deducirla de una sola respuesta de readiness.
- Latencia: p95 por ruta crítica y ventana, con tamaño de muestra y carga
  declarados; no mezclar percentiles de rutas distintas sin metodología.
- Error rate: respuestas 5xx sobre solicitudes observadas de la misma ventana
  y alcance. Registrar fallos de proveedor y timeouts por separado.
- RTO: tiempo desde la interrupción declarada hasta recuperación funcional
  comprobada; RPO: distancia entre incidente y último dato recuperado íntegro.
- La meta RPO de una hora es condicional al respaldo incremental implementado
  y ensayado; no se considera disponible por figurar en este documento.
- Las metas y umbrales se validan en staging/operación. El responsable registra
  evidencia, ventana y decisión ante incumplimiento; no se publica cumplimiento
  mensual sin telemetría de ese período.

Ver [incidentes y continuidad](../operacion/incidentes_y_continuidad.md),
[QA](../calidad/estrategia_verificacion.md) y
[recuperación](runbooks/runbook_recuperacion_desastre_docker_vps.md).
