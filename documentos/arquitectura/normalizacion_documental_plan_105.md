# Normalizacion documental por lotes - P105-019

Fecha de linea base: 2026-07-22.

## Hallazgo

La auditoria UTF-8 refinada detecta 218 secuencias sospechosas. El detector
descarta parámetros URL legítimos como `?action`. Distribucion vigente:

| Archivo | Secuencias | Tratamiento |
|---|---:|---|
| `CHANGELOG.md` | 218 | Historico; corregir por secciones fechadas, no por reemplazo global. |

## Secuencia ejecutable

1. Tomar una sección fechada de `CHANGELOG.md` por lote.
2. Guardar antes/despues con `git diff --word-diff` y verificar que no cambien
   comandos, rutas, identificadores, hashes ni fechas historicas.
3. Ejecutar `node tools/docs_normalization_audit.mjs` despues de cada lote.
4. Registrar el conteo resultante en `historial_de_cambios` y no declarar
   cerrado P105-019 hasta que la auditoria quede sin hallazgos.

## Prohibiciones

- No usar reemplazo global, scripts de recodificacion ciega ni normalizadores
  que modifiquen binarios o documentos no incluidos en la linea base.
- No editar trazabilidad historica para ocultar errores o reducir el conteo.
- No mezclar esta deuda P2 con la preparacion de un SHA de produccion sin un
  commit revisable separado.
