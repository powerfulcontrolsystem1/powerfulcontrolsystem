Run tests por módulo y seed

Uso rápido:

- Importar seed y ejecutar todos los tests por módulo (PowerShell):
  - `.
un_tests_by_module.ps1 -Seed`

- Ejecutar tests de un módulo específico:
  - `cd backend`
  - `go test ./db -v`

Detalles:
- `run_tests_by_module.ps1` ejecuta, por defecto, los módulos: `db`, `auth`, `handlers`, `utils`, `metrics`.
- Para agregar más módulos, edita la variable `$modules` dentro del script.
- El script intenta usar `sqlite3.exe` en la raíz del repo para importar `scripts/seed_data.sql` a `testdata/seed.db`.

Logs: los resultados de cada módulo se guardan en `logs/test_runs/` con timestamps.
