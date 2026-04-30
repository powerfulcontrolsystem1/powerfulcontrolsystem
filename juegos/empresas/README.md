# Partidas guardadas por empresa

Esta carpeta almacena las partidas del emulador por empresa cuando el servidor se ejecuta localmente.

Estructura esperada:

```text
empresas/
  empresa_7/
    emulador/
      nombre-del-juego/
        latest.state
        latest.save
        latest.png
        meta.json
```

En el VPS se recomienda usar `/opt/juegos/empresas` con propietario `www-data:www-data`.
Los archivos reales de partidas no se versionan en git.
