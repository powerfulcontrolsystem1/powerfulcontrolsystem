# Nomina Colombia avanzada

Fecha: 2026-05-06

## Alcance

El modulo amplifica la nomina existente sin duplicarla. Mantiene la liquidacion, asistencia, provisiones, pagos y desprendibles en `nomina_sueldos`, y agrega una capa Colombia para conceptos legales, novedades aprobables y resumen PILA por empresa.

## Superficies

- Backend: `/api/empresa/nomina`.
- Pantalla: `web/administrar_empresa/nomina_sueldos.html`.
- Tablas nuevas: `empresa_nomina_colombia_conceptos`, `empresa_nomina_colombia_novedades`, `empresa_nomina_colombia_pila_resumen`.
- Aislamiento: todas las tablas nuevas usan `empresa_id`; la ruta sigue usando los wrappers de nomina/finanzas existentes.

## Acciones API

- `GET action=conceptos_colombia`: lista conceptos legales por tipo.
- `POST action=concepto_colombia`: crea o actualiza un concepto.
- `GET action=novedades_colombia`: lista novedades por periodo y estado.
- `POST action=novedad_colombia`: registra una novedad de nomina.
- `POST action=generar_pila`: genera resumen PILA desde liquidaciones del periodo.
- `GET action=pila_colombia`: consulta resumen PILA.
- `GET action=dashboard_colombia`: consolida KPIs, conceptos, novedades y PILA.
- `POST action=seed_colombia`: carga parametros demo base para Colombia.
- `POST action=aprobar_novedad_colombia`: cambia el estado de aprobacion de una novedad activa por empresa.
- `POST action=seed_motel_calipso` / `seed_profesional`: crea una nomina demo profesional con empleados, asistencia, novedades aprobadas, liquidaciones, PILA y pagos simulados.

## Flujo operativo

1. Configurar parametros legales base de nomina.
2. Registrar empleados de nomina con salario, auxilio, cargo, contrato y estado.
3. Cargar o ajustar conceptos Colombia: salario basico, auxilio de transporte, salud, pension, aportes y novedades propias.
4. Registrar novedades del periodo y aprobarlas cuando corresponda.
5. Calcular liquidaciones del periodo en el modulo principal. Las novedades aprobadas de tipo `devengado` y `deduccion` se incorporan al devengado, IBC, deducciones y neto de la liquidacion.
6. Generar resumen PILA desde las liquidaciones.
7. Consultar dashboard para revisar conceptos activos, novedades pendientes/aprobadas y total de aportes.

## Refuerzo profesional 2026-05-20

- El seed Colombia ahora carga un catalogo amplio de conceptos: salario basico, auxilio, horas extra, recargos, bonificaciones, vacaciones, incapacidades, salud, pension, solidaridad, prestamos, embargos, retencion, aportes patronales, parafiscales y provisiones.
- La liquidacion aplica novedades aprobadas del periodo por empleado, recalcula salud, pension, fondo de solidaridad, deducciones totales y neto a pagar.
- La pantalla de nomina agrega un tablero de cobertura profesional, accion para crear demo Motel Calipso y aprobacion/rechazo de novedades pendientes.
- El flujo demo crea empleados simulados `CAL-NOM-*`, asistencia del periodo, novedades aprobadas, liquidaciones, resumen PILA y pagos de prueba sin agregar tablas nuevas.

## Validacion

- Pruebas unitarias: `go test ./db -run TestNormalizeNominaColombia -count=1`.
- Compilacion de handlers: `go test ./handlers -run '^$' -count=1`.
- QA productivo Calipso: `go run ./tmp_tools/qa_calipso_modulos_nuevos -empresa_id=7`, que registra empleado, concepto, novedad, liquidacion y PILA.
- 2026-05-20: `go test -count=1 ./db ./handlers`; validacion visual con servidor mock en `nomina_sueldos.html?empresa_id=33`, boton `Crear nomina demo Motel Calipso` y aprobacion de novedad.
