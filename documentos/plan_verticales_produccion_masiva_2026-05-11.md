# Plan de verticales para produccion masiva

Fecha: 2026-05-11

## Decision

La version masiva debe vender pocas plantillas muy bien integradas, no veinte promesas medianas. Los 20 verticales nuevos se conservan como catalogo tecnico, pero solo 10 quedan priorizados para produccion masiva v1. Los otros 10 quedan diferidos, ocultables o vendibles solo bajo implementacion asistida hasta completar flujos especificos.

La decision usa tres criterios:

- Demanda esperada en Colombia por servicios, salud, comercio, turismo, transporte, entretenimiento y soporte tecnico.
- Facilidad de operar sobre el nucleo unico actual: clientes, servicios/productos, ventas, pagos, facturacion, reportes y permisos.
- Bajo riesgo regulatorio u operativo para escalar sin crear modulos paralelos.

Referencias de mercado usadas como soporte: DANE PIB 2025, MinCIT turismo 2025 y Confecamaras/RUES sobre dinamica de creacion de empresas. Estas fuentes soportan priorizar comercio/servicios, salud, turismo, alojamiento/comida, entretenimiento, transporte y reparacion/servicios tecnicos.

## Integrar en v1 masiva

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

## Diferir de v1 masiva

| Vertical | Decision |
|---|---|
| `operador_turistico` | Diferir: se cubre inicialmente con `agencia_viajes`; necesita cupos, rutas, liquidacion y reservas mas profundas. |
| `colegio_academia` | Diferir: requiere periodos academicos, notas, cartera educativa y permisos mas sensibles. |
| `guarderia_infantil` | Diferir: maneja menores, autorizaciones, acudientes y trazabilidad sensible. |
| `inmobiliaria_comercial` | Diferir: venta consultiva larga; conviene madurar CRM, contratos y pipeline inmobiliario. |
| `seguridad_privada` | Diferir: necesita turnos, puestos, rondas, novedades y cumplimiento laboral mas profundo. |
| `club_deportivo` | Diferir: se solapa con gimnasio; conviene estabilizar agenda/clases primero. |
| `funeraria_exequial` | Diferir: vertical sensible y especializado; requiere soporte documental y atencion diferencial. |
| `parque_recreativo` | Diferir: requiere aforo, manillas, hardware, taquilla y control de acceso mas maduro. |
| `cooperativa_fondo` | Diferir: entra en cartera, creditos, aportes y controles financieros mas regulados. |
| `capacitacion_empresarial` | Diferir: puede reutilizar colegio/CRM; mejor para fase B2B posterior. |

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

El catalogo de nuevos verticales expone la misma decision como `integracion_preconfig`, `produccion_masiva`, `prioridad_produccion` y `decision_preconfig`. Asi el panel, super administrador, portal publico y seed de preconfiguraciones leen una sola matriz.

## Plan para version apta para produccion masiva

1. Congelar contrato de plantilla: no publicar vertical sin `integracion_preconfig` completa.
2. Activar comercialmente solo los 10 priorizados en v1 masiva.
3. Mantener los 10 diferidos en catalogo tecnico, pero sin prometerlos como operativos masivos hasta cerrar flujos propios.
4. Para cada vertical priorizado, validar demo de datos iniciales, roles, permisos, menu, estaciones, servicios guia, tareas guia y reportes.
5. Ejecutar QA por rol: super administrador, administrador empresa, cajero/operador y soporte.
6. Verificar que toda venta/pago llegue al nucleo central y no a tablas paralelas.
7. Crear checklist de onboarding por plantilla: licencia, usuario inicial, productos/servicios, medios de pago, facturacion y reportes.
8. Preparar release masivo con migraciones PostgreSQL revisadas, backup, monitoreo VPS, preflight y rollback documentado.

## Estado actual

- Implementado: conexion de preconfiguracion con matriz extendida.
- Implementado: seleccion de 10 verticales v1 masiva y 10 diferidos.
- Implementado: contrato API para que el catalogo de verticales nuevos exponga la decision.
- Implementado: pruebas que bloquean catalogos sin metadata extendida y que exigen exactamente 10 verticales masivos.
- Implementado: pantalla `web/super/verticales_produccion_masiva.html` en super administrador para KPIs, filtros, ranking, metadata extendida y exportacion CSV.
- Implementado: acciones desde cada vertical hacia `Tipos de empresa`, `Preconfiguraciones` y `Licencias`, con filtros iniciales por `q`, `vertical` o `modulo`.
- Implementado: semaforo `Listo venta` que cruza produccion masiva, metadata completa, preconfiguracion activa y licencia activa por modulo.
- Implementado: accion `Asegurar v1` para crear/actualizar tipos, preconfiguraciones y licencias recomendadas de los 10 verticales priorizados.
- Pendiente: validar visualmente la pantalla con sesion real de super administrador en entorno local o staging.
