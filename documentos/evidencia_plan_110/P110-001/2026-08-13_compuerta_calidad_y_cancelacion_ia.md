# P110-001A - Compuerta de calidad y cancelación IA

Fecha: 2026-08-13

Ambiente: código y pruebas locales. Sin despliegue, sin mutaciones PCS y sin
ejecutar `rs`.

## Resultado

Se convirtió la auditoría estructural en una compuerta decreciente escrita en
Go estándar. El control mide código de producción y falla cuando aumenta
cualquiera de estas deudas:

- funciones de más de 100 y 200 líneas;
- tamaño de la función más grande;
- grupos de cuerpos de función idénticos entre archivos;
- llamadas DB sin variante con contexto;
- resultados descartados explícitamente mediante `_`.

La línea base no declara los valores aceptables para producción: son máximos
temporales que solo pueden mantenerse o disminuir. La herramienta no se cuenta
a sí misma y genera un JSON de evidencia en los reportes profesionales. Quedó
integrada tanto al preflight local como al CI obligatorio.

## Cancelación de proveedores IA

El contexto del request HTTP ahora llega a la capa común de proveedores y a las
solicitudes de OpenAI Responses, OpenAI Chat Completions y Gemini. Se actualizaron
los puntos HTTP de chat empresarial/super, adjuntos, reportes IA, Grafología,
pedidos, renta, selector, Centro IA, documentos dinámicos, portal público,
prueba global y lectura DIAN por IA.

Los jobs programados pueden conservar un contexto propio; no dependen de un
request de usuario. Los wrappers históricos sin contexto permanecen solo para
compatibilidad interna y deben desaparecer cuando todos los callers durables
declaren explícitamente su contexto.

## Pruebas

- pruebas del auditor: comparación de baseline y clasificación DB;
- cancelación real de servidor HTTP de prueba para Responses, Chat Completions
  y Gemini;
- pruebas enfocadas de handlers IA y herramienta: PASS;
- compuerta sobre el árbol completo: PASS, sin regresiones frente a baseline.

## Límite

Este bloque evita nueva deuda y corrige cancelación IA. La preconfiguración por
tipo de empresa dejó de ejecutar diez llamadas DDL desde tráfico HTTP: productos,
usuarios, configuración operativa, comisiones, tres clases de tarifas y
Domótica verifican ahora el contrato migrado y fallan cerrados. La limpieza de
esa preconfiguración tampoco descarta errores de sus dos escrituras. El
inventario queda en 72 llamadas `Ensure*`, ninguna en tráfico HTTP, frente a
104/29 al
inicio de este bloque. La segunda extracción cubre además eventos contables de
Créditos/Devoluciones, permisos finos, Nextcloud y destinatarios empresariales
de correo masivo; Rappi, permisos por rol, alertas y agentes de mantenimiento
completan este lote. El correo empresarial ya no se aprovisiona durante GET:
el alta de empresa y la sincronización administrativa son sus únicas acciones
explícitas. Nextcloud, plantillas/licencias y el token de voz se clasifican como
provisionamiento o sincronización, no como migraciones. Persisten 52
grupos duplicados y grandes handlers. Dos
mutaciones GET de Domótica fueron retiradas: el nodo Raspberry principal ahora
se sincroniza explícitamente solo al guardar configuración, con error observable
en vez de descartado. P110-001A
continúa parcial. El avance formal del Plan 110 se mantiene en 38,5 %, la
certificación en 0 % y el veredicto en NO-GO.

El primer CI de esta compuerta reveló tres alertas `gosec` dentro del propio
auditor: lectura de baseline suministrada por CLI y permisos demasiado amplios
del directorio/reporte. La ruta se normalizó y se documentó como entrada local
de operador confiable; los permisos bajaron a `0750` y `0600`. Esta corrección
forma parte del candidato y debe pasar nuevamente el CI obligatorio.

Las 72 llamadas restantes corresponden al arranque protegido de API. Su retiro
requiere completar la autoridad del migrador y ensayar base vacía/upgrade; no se
deben borrar mecánicamente ni confundir con el cierre ya conseguido de HTTP.

El CI posterior detectó seis vulnerabilidades alcanzables de la biblioteca
estándar corregidas en Go 1.25.13. Se elevó consistentemente el toolchain del
módulo, las dos matrices de GitHub Actions y la imagen de compilación Docker de
1.25.12 a 1.25.13. La versión está publicada en el catálogo oficial de Go; el
nuevo `govulncheck` remoto debe confirmar el cierre.

La validación local con Go 1.25.13 se ejecutó con `GOTMPDIR` y `GOCACHE` en D:
porque C: tenía menos de 5 MB libres. Aprobaron DB, handlers, aplicación, auditor,
`go vet` y `govulncheck ./...`; este último informó cero vulnerabilidades
alcanzables. No se borraron temporales ni datos del usuario porque la operación
de limpieza fue bloqueada por la política del entorno.
