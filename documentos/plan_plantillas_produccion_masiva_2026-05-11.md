# Plan de plantillas para produccion masiva

Fecha: 2026-05-11

## Decision

La version masiva queda cerrada con los 20 plantillas nuevas como plantillas reales. La regla profesional se mantiene: ninguna plantilla puede duplicar clientes, productos/servicios, ventas, pagos, facturacion, reportes ni permisos; cada vertical agrega solo su especialidad operativa sobre `empresa_modulos_colombia_*` y el nucleo central.

La decision usa tres criterios:

- Demanda esperada en Colombia por servicios, salud, comercio, turismo, transporte, entretenimiento y soporte tecnico.
- Facilidad de operar sobre el nucleo unico actual: clientes, servicios/productos, ventas, pagos, facturacion, reportes y permisos.
- Bajo riesgo regulatorio u operativo para escalar sin crear modulos paralelos.

Referencias de mercado usadas como soporte: DANE PIB 2025, MinCIT turismo 2025 y Confecamaras/RUES sobre dinamica de creacion de empresas. Estas fuentes soportan priorizar comercio/servicios, salud, turismo, alojamiento/comida, entretenimiento, transporte y reparacion/servicios tecnicos.

## Integrar en produccion masiva

| Prioridad | Vertical | Motivo |
|---:|---|---|
| 1 | `salon_spa` | Alto volumen PyME, agenda, servicios, productos, paquetes, ventas y pagos simples. |
| 2 | `veterinaria_petshop` | Combina servicios, historia de mascota, productos, inventario y ventas recurrentes. |
| 3 | `clinica_consultorios` | Demanda alta en salud ambulatoria; opera con agenda, pacientes/clientes, servicios y pagos. |
| 4 | `laboratorio_clinico` | Servicios estandarizables, ordenes, muestras, resultados y facturacion central. |
| 5 | `taller_mecanico` | Demanda amplia por reparacion de vehiculos, ordenes de servicio, repuestos y caja. |
| 6 | `servicios_tecnicos` | Aplica a celulares, electrodomesticos, computo y mantenimiento; flujo de tickets y venta central. |
| 7 | `lavanderia_tintoreria` | Operacion simple, repetible, con ordenes, estados, cobros y entregas. |
| 8 | `agencia_viajes` | Turismo activo en Colombia; paquetes, asesorias, reservas y pagos centralizados. |
| 9 | `eventos_boleteria` | Entretenimiento y boleteria conectan bien con clientes, pagos, QR y reportes. |
| 10 | `transporte_carga_tms` | Transporte/logistica tiene demanda B2B y puede iniciar con rutas, guias, clientes y cobros. |
| 11 | `operador_turistico` | Tours, cupos, rutas, guias, check-in y pagos sobre clientes/servicios/ventas centrales. |
| 12 | `colegio_academia` | Cursos, alumnos, cohortes, cartera educativa guia y reportes sin duplicar clientes ni pagos. |
| 13 | `guarderia_infantil` | Acudientes, autorizaciones, asistencia, actividades y evidencias con trazabilidad por empresa. |
| 14 | `inmobiliaria_comercial` | Inmuebles, asesores, leads, contratos y pipeline conectado a CRM y ventas centrales. |
| 15 | `seguridad_privada` | Puestos, guardas, rondas, novedades y SLA sin duplicar facturacion ni cartera. |
| 16 | `club_deportivo` | Disciplinas, clases, torneos, asistencia y pagos sobre agenda, clientes y servicios centrales. |
| 17 | `funeraria_exequial` | Planes, afiliados, servicios, documentos y cierre operativo con gestion documental central. |
| 18 | `parque_recreativo` | Entradas, manillas QR, aforo, atracciones, consumos y reportes conectados al nucleo. |
| 19 | `cooperativa_fondo` | Asociados, aportes, beneficios y cartera guia con controles financieros y permisos estrictos. |
| 20 | `capacitacion_empresarial` | Cursos, empresas cliente, asistencia, evaluaciones, certificados y ventas B2B desde CRM/nucleo. |

## Conexion tecnica

Cada preconfiguracion de tipo de empresa debe incluir `integracion_vertical` dentro del JSON normalizado. Este bloque declara:

- `modulo`
- `estado_integracion`
- `decision`
- `produccion_masiva`
- `prioridad_produccion`
- `motivo_decision`
- `template_activates`
- `tables_touched`
- `required_permissions`
- `sale_flow`
- `reports_produced`

El catalogo de nuevas plantillas expone la misma decision como `integracion_preconfig`, `produccion_masiva`, `prioridad_produccion` y `decision_preconfig`. Asi el panel, super administrador, portal publico y seed de preconfiguraciones leen una sola matriz.

## Plan para version apta para produccion masiva

1. Congelar contrato de plantilla: no publicar vertical sin `integracion_preconfig` completa.
2. Activar comercialmente los 20 plantillas nuevas como plantillas reales de produccion masiva.
3. Validar demo de datos iniciales, roles, permisos, menu, estaciones, servicios guia, tareas guia y reportes para cada vertical.
4. Ejecutar QA por rol: super administrador, administrador empresa, cajero/operador y soporte.
5. Verificar que toda venta/pago llegue al nucleo central y no a tablas paralelas.
6. Crear checklist de onboarding por plantilla: licencia, usuario inicial, productos/servicios, medios de pago, facturacion y reportes.
7. Preparar release masivo con migraciones PostgreSQL revisadas, backup, monitoreo VPS, preflight y rollback documentado.

## Estado actual

- Implementado: conexion de preconfiguracion con matriz extendida.
- Implementado: seleccion de 20 plantillas para produccion masiva.
- Implementado: contrato API para que el catalogo de plantillas nuevas exponga la decision.
- Implementado: pruebas que bloquean catalogos sin metadata extendida y que exigen exactamente 20 plantillas masivos.
- Implementado: pantalla `web/super/plantillas_produccion_masiva.html` en super administrador para KPIs, filtros, ranking, metadata extendida y exportacion CSV.
- Implementado: acciones desde cada vertical hacia `Tipos de empresa`, `Preconfiguraciones` y `Licencias`, con filtros iniciales por `q`, `vertical` o `modulo`.
- Implementado: semaforo `Listo venta` que cruza produccion masiva, metadata completa, preconfiguracion activa y licencia activa por modulo.
- Implementado: accion `Asegurar 20` para crear/actualizar tipos, preconfiguraciones y licencias recomendadas de los 20 plantillas.
- Estado: los 20 plantillas nuevas quedan reales en el catalogo; la validacion automatizada confirma metadata, ranking y exposicion publica de los 20.
