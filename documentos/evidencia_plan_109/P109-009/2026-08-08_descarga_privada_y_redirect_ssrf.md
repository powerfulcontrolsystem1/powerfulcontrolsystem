# P109-009 - descarga privada y redirecciones SSRF

Fecha: 2026-08-08
Empresa: Powerful Control System (`empresa_id=12`)
Entorno dinámico: VPS publicado (`https://powerfulcontrolsystem.com`)
Rama correctiva local: `codex/p109-batch-no-pr`; sin push, PR ni despliegue.

## Descarga autenticada del soporte IA

Se abrió una sesión mediante el login oficial, se consultó exclusivamente el
soporte privado `SCI-0004` ya creado por la prueba hostil anterior y se cerró la
sesión mediante logout. El archivo se verificó en memoria; no se copió a disco
ni se expusieron credenciales, cookies, cuerpo o rutas privadas.

| Caso | HTTP | Resultado |
| --- | ---: | --- |
| Empresa 12 y soporte 8 | 200 | `text/xml`, adjunto, 1.185 bytes, `no-store`, `nosniff`, sin ruta interna |
| Mismo recurso sin sesión | 401 | bloqueado antes de entregar el archivo |
| Soporte 8 solicitado como empresa 53 | 404 | objeto no visible fuera de su `empresa_id` |
| `soporte_id` con expresión de inyección | 400 | identificador rechazado por parseo estricto |
| `soporte_id` ausente | 400 | identificador obligatorio |
| Logout oficial | 200 | sesión de prueba cerrada |

La consulta backend usa conjuntamente `empresa_id` y `soporte_id`. Esta prueba
confirma enlace de objeto para la sesión administrativa usada, pero no reemplaza
la matriz A/B con dos identidades no globales.

## Hallazgo y corrección local SSRF

La revisión de salidas HTTP detectó que el callback firmado de OnlyOffice
validaba esquema, host y puerto de la URL inicial contra el Document Server,
pero usaba la política predeterminada de redirecciones de `http.Client`. Un
servidor permitido comprometido podía redirigir la descarga hacia otro origen.

La corrección agrega un cliente dedicado que vuelve a aplicar la misma allowlist
en cada salto, limita la cadena a diez redirecciones y rechaza antes de conectar
cualquier cambio de esquema, host o puerto. Se mantiene la descarga legítima
con redirecciones relativas del mismo origen.

Pruebas enfocadas:

- redirección del mismo origen: PASS, documento final HTTP 200;
- redirección a un segundo servidor local: PASS, error cerrado y cero
  solicitudes recibidas por el destino bloqueado;
- límite de tamaño y `Cache-Control` temporal existentes: PASS;
- `go vet ./handlers`: PASS.

No se cambió JWT, almacenamiento, permisos ni contratos multiempresa. La
corrección aún no está desplegada y por tanto no recibe crédito de certificación
del candidato. P109-009 permanece **parcial** por A/B no global, matriz de roles,
otros consumidores de URL/exportación y retiro progresivo de contenido inline.
