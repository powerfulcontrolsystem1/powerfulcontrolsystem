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

Este bloque evita nueva deuda y corrige cancelación IA, pero no reduce aún las
104 llamadas `Ensure*`, 52 grupos duplicados ni los grandes handlers. Dos
mutaciones GET de Domótica fueron retiradas: el nodo Raspberry principal ahora
se sincroniza explícitamente solo al guardar configuración, con error observable
en vez de descartado. P110-001A
continúa parcial. El avance formal del Plan 110 se mantiene en 38,5 %, la
certificación en 0 % y el veredicto en NO-GO.
