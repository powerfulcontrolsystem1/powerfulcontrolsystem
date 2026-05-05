# Carta publica de productos y precios

Fecha: 2026-05-05
Estado: vigente

## Objetivo

El modulo permite publicar una carta externa de solo lectura con productos, precios y fotos. Esta pagina no crea pedidos, no agrega carrito y no inicia pagos; su proposito es mostrar informacion comercial al publico bajo el slug o subdominio de la empresa.

## Rutas

- Administracion: `/administrar_empresa/carta_productos_publica.html`
- Acceso desde menu: `Administrar empresa > Productos > Carta publica`
- Pagina publica: `/visualizar_productos_y_precios_publico.html`
- Pagina publica con slug por ruta: `/{empresa_slug}/visualizar_productos_y_precios_publico.html`
- API publica reutilizada: `/api/public/venta_publica?action=catalogo`
- API administrativa reutilizada: `/api/empresa/venta_publica`

## Flujo administrativo

1. Entrar al modulo `Productos`.
2. Abrir `Carta publica`.
3. Configurar slug/subdominio, nombre visible, descripcion, logo, banner, color, moneda y politica de stock visible.
4. Guardar la seccion principal de la carta.
5. Publicar productos desde el inventario. El modulo sincroniza nombre, descripcion, precio, foto, SKU y stock publicado.
6. Usar el boton `Visualizar carta publica` para abrir la pagina externa.

## Comportamiento publico

La pagina `visualizar_productos_y_precios_publico.html` detecta la empresa por:

- `empresa_slug` en query string.
- `empresa_id` en query string.
- Ruta `/{empresa_slug}/visualizar_productos_y_precios_publico.html`.
- Subdominio compatible con `VENTA_PUBLICA_BASE_DOMAINS`, por ejemplo `motel-calipso.powerfulcontrolsystem.com`.

El publico puede buscar, filtrar por secciones y ordenar por relevancia, menor precio, mayor precio o nombre. No hay botones de compra ni formularios de pedido.

## Permisos y datos

La configuracion y los productos publicados usan las tablas existentes del modulo de venta publica:

- `empresa_venta_publica_configuracion`
- `empresa_venta_publica_paginas`
- `empresa_venta_publica_items`

El menu `linkCartaProductosPublica` queda registrado como pagina de ventas, porque la publicacion externa comparte el contrato administrativo de venta publica. La pantalla tambien consulta inventario para listar productos activos.

## Pruebas

Se agregaron pruebas de backend para validar que el slug se resuelve correctamente desde:

- `/{empresa_slug}/visualizar_productos_y_precios_publico.html`
- Query string `empresa_slug`, que mantiene prioridad sobre la ruta.

Comandos de validacion aplicables:

```powershell
cd backend
go test ./...
```

Validacion frontend de sintaxis:

```powershell
@'
const fs = require('fs');
for (const file of ['web/administrar_empresa/carta_productos_publica.html','web/visualizar_productos_y_precios_publico.html']) {
  const html = fs.readFileSync(file, 'utf8');
  for (const match of html.matchAll(/<script(?:\s[^>]*)?>([\s\S]*?)<\/script>/gi)) new Function(match[1]);
  console.log(`${file}: JS syntax OK`);
}
'@ | & 'C:\Users\ivanm\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\bin\node.exe' -
```
