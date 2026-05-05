# Carta publica de productos y precios

Fecha: 2026-05-05
Estado: vigente

## Actualizacion 2026-05-05

- La carta publica queda incluida en la descripcion comercial del portal `web/index.html` como `Venta publica, carta QR y red social`.
- La ruta `/visualizar_productos_y_precios_publico.html` y la ruta `/{empresa_slug}/visualizar_productos_y_precios_publico.html` quedan consideradas publicas por el middleware de autenticacion; no deben pedir sesion porque son vistas externas de solo lectura.
- Motel Calipso queda publicado con slug `motel-calipso`, carta publica y venta publica en produccion.
- Validacion productiva realizada contra:
  - `https://powerfulcontrolsystem.com/motel-calipso/venta_publica.html`
  - `https://powerfulcontrolsystem.com/motel-calipso/visualizar_productos_y_precios_publico.html`
  - `https://powerfulcontrolsystem.com/red_social_comercial.html`

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
7. Generar el QR del enlace publico y exportarlo para impresion.

## QR imprimible

La pantalla administrativa genera localmente un codigo QR a partir de la URL publica actual de la carta. No depende de servicios externos. Cada cambio de slug, dominio o nombre actualiza el QR visible.

Formatos disponibles:

- PNG: recomendado para avisos rapidos, imagenes y documentos sencillos.
- SVG: recomendado para imprenta o material que necesite escala sin perdida.
- PDF: hoja lista para imprimir con QR, nombre de la carta y direccion publica.
- Imprimir QR: abre una vista imprimible en el navegador.

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

El menu `linkCartaProductosPublica` queda registrado dentro del modulo independiente `venta_publica`, porque la publicacion externa comparte el contrato administrativo de venta publica. La pantalla envia `perm_page=linkCartaProductosPublica` en sus llamadas administrativas para que la anulacion por pagina del rol aplique a la carta y no a toda la venta publica. Tambien consulta inventario para listar productos activos, asi que el rol debe conservar lectura/consulta de inventario cuando administre productos de la carta.

## Publicacion operativa Motel Calipso

Datos semilla aplicados en empresa `7`:

- Configuracion `empresa_venta_publica_configuracion`: slug `motel-calipso`, tienda `Motel Calipso`, moneda `COP`, stock visible y pedidos/tracking habilitados para la operacion publica.
- Paginas publicas activas: `experiencias-calipso`, `carta-productos-precios` y `pos-motel-calipso`.
- Items publicados de ejemplo: decoracion de habitacion, noche romantica, combo bebidas y snacks, kit de aseo premium, desayuno en habitacion y POS Motel Calipso.
- Publicaciones de red social: `POS y carta publica de Motel Calipso` y `Experiencias Calipso disponibles en linea`.

La operacion se sembro de forma idempotente con `backend/tmp_tools/seed_motel_calipso_publicacion`, para poder repetir la publicacion sin duplicar registros.

## Pruebas

Se agregaron pruebas de backend para validar que el slug se resuelve correctamente desde:

- `/{empresa_slug}/visualizar_productos_y_precios_publico.html`
- Query string `empresa_slug`, que mantiene prioridad sobre la ruta.

Comandos de validacion aplicables:

```powershell
cd backend
go test ./...
```

Pruebas puntuales aplicadas tras liberar la ruta publica:

```powershell
cd backend
go test ./utils
go test ./ ./auth ./db ./handlers ./metrics ./utils -run '^$' -count=1
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
