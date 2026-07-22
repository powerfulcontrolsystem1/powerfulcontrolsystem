# Releases

Los manifiestos formales de release se generan con:

```powershell
node tools\release_manifest.mjs --version=1.0.0
```

Cada manifiesto deja version, fecha, rama, commit, estado del working tree, commits recientes y checks requeridos antes de publicar.

Antes de generar un manifiesto definitivo puede revisarse sin escribir archivos:

```powershell
node tools\release_manifest.mjs --check --base-ref=origin/main
```

El modo `--strict` bloquea la compuerta si el arbol no esta limpio, el candidato
no desciende de la base indicada, la rama no tiene upstream o faltan referencias
inmutables API/migrador/worker con formato `repositorio@sha256:<digest>`.
