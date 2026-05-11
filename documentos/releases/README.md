# Releases

Los manifiestos formales de release se generan con:

```powershell
node tools\release_manifest.mjs --version=1.0.0
```

Cada manifiesto deja version, fecha, rama, commit, estado del working tree, commits recientes y checks requeridos antes de publicar.
