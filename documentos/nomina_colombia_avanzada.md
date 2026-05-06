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

## Flujo operativo

1. Configurar parametros legales base de nomina.
2. Registrar empleados de nomina con salario, auxilio, cargo, contrato y estado.
3. Cargar o ajustar conceptos Colombia: salario basico, auxilio de transporte, salud, pension, aportes y novedades propias.
4. Registrar novedades del periodo y aprobarlas cuando corresponda.
5. Calcular liquidaciones del periodo en el modulo principal.
6. Generar resumen PILA desde las liquidaciones.
7. Consultar dashboard para revisar conceptos activos, novedades pendientes/aprobadas y total de aportes.

## Validacion

- Pruebas unitarias: `go test ./db -run TestNormalizeNominaColombia -count=1`.
- Compilacion de handlers: `go test ./handlers -run '^$' -count=1`.
- QA productivo Calipso: `go run ./tmp_tools/qa_calipso_modulos_nuevos -empresa_id=7`, que registra empleado, concepto, novedad, liquidacion y PILA.
