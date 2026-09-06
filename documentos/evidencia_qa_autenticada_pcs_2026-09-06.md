# Evidencia de QA autenticada en la empresa PCS

Estado: Evidencia. Responsable: Coordinación técnica. Fecha de ejecución: 2026-09-06.

## Alcance

Se validó el dominio publicado `https://powerfulcontrolsystem.com` con la cuenta administrativa autorizada y la empresa Powerful Control System, resuelta por el sistema como `empresa_id=12`. La prueba fue deliberadamente no destructiva: navegación, lecturas HTTP, controles negativos y revisión responsive. No se crearon ventas, cobros, facturas, movimientos de inventario, registros de nómina, mensajes, datos demo, escaneos completos ni acciones sobre hardware.

## Resultados por módulo

| Módulo o superficie | Evidencia observada | Resultado |
| --- | --- | --- |
| Autenticación administrativa | El acceso por correo creó una sesión Super válida y redirigió al panel global. | Funciona, pero el entorno publicado permitió acceso privilegiado sin solicitar TOTP. El endurecimiento MFA local aún no está desplegado. |
| Selector y empresa PCS | La lista mostró la empresa PCS con licencia activa y permitió abrir su panel. | La empresa correcta fue resuelta como `empresa_id=12`. Desde el panel Super, la primera navegación se mantuvo anidada en su iframe; al abrir el panel empresarial directamente, el contexto cargó correctamente. |
| Productos e inventario | Página y `GET /api/empresa/productos?empresa_id=12` cargaron con 200: 18 productos, 2 bodegas, 1 servicio, 5 categorías, 2 alertas y 106 movimientos del periodo. | Lectura autenticada aprobada; no se creó, ajustó, transfirió, importó ni eliminó inventario. |
| Compras | Menú, formulario y `GET /api/empresa/compras/documentos?empresa_id=12&limit=10` cargaron con 200 y 2 documentos históricos. | Lectura aprobada; no se adjuntó soporte ni se cambió el ciclo de aprobación. |
| Finanzas y contabilidad | Menú y configuración cargaron; configuración y movimientos respondieron 200. La colección de movimientos consultada estaba vacía. | Lectura aprobada; no se guardó configuración, movimiento, cierre, cartera, conciliación ni plantilla. |
| Producción/MRP | Dashboard API 200 y página funcional con indicadores, tabs, alertas y acciones. | Lectura aprobada; no se cargó demo ni se creó orden. Vista móvil 390x844 sin desbordamiento horizontal visible en el tramo revisado. |
| Logística/WMS | Dashboard API 200 y página funcional con indicadores, flujo, alertas y órdenes. | Lectura aprobada; no se cargó demo, ubicación, orden ni despacho. Vista móvil 390x844 sin desbordamiento horizontal visible en el tramo revisado. |
| Tesorería/presupuesto | Dashboard API 200 con configuración, cuentas y presupuestos en su contrato JSON. | Lectura autenticada aprobada; no se guardaron datos. |
| Auditoría empresarial | `GET /api/empresa/auditoria/eventos?empresa_id=12&limit=5` devolvió 200 y 5 eventos. | Lectura aprobada; no se exportó evidencia ni se aplicó retención. |
| Usuarios, roles y estaciones | Página y `GET /api/empresa/usuarios?empresa_id=12&include_inactive=1` cargaron con 200 y 13 usuarios. La interfaz mostró roles propios de la empresa y asignación de estaciones. | Lectura aprobada; no se creó, editó, activó, eliminó ni reenvió invitación alguna. |
| CRM unificado | API de interacciones 200 con colección vacía; menú y pantalla de seguimientos cargaron. | Lectura aprobada. La UI publicada todavía ofrece estado y cliente como campos de alta; el backend local ya restringe esos campos, por lo que frontend publicado y candidato local no están en paridad. La vista móvil del tablero no mostró desbordamiento en el tramo revisado. |
| Correo y panel empresarial | El panel mostró estado activo y contadores de buzón corporativo. | Lectura aprobada; no se abrió correo ni se envió mensaje. |
| DIAN | El buzón empresarial conserva múltiples alertas históricas de una factura rechazada/fallida. | No se ejecutó emisión ni reenvío. La aceptación fiscal productiva continúa pendiente y debe tratarse con el runbook DIAN. |
| Facturación electrónica | El menú resolvió Colombia y el buzón de documentos cargó; `GET /api/empresa/facturacion_electronica?empresa_id=12&limit=10` respondió 200. | Lectura aprobada; no se emitió, anuló, reenvió, descargó ni modificó documento fiscal. |
| Domótica | La API de resumen respondió 200, pero la página publicada mostró `isDoorSensor is not defined` y dejó los indicadores sin materializar. | Fallo reproducido en producción. El árbol local ya inicializa la variable antes de renderizar y tiene una prueba de contrato; falta desplegar y repetir con el candidato, sin accionar relés ni Raspberry. |
| Seguridad VPS | El panel cargó configuración, historial, puertos y el último reporte. | El reporte más reciente visible es del 2026-07-24: hardening 60/100, 41 hallazgos, 2 críticos, 6 altos, 28 medios y 4 bajos. Es evidencia desactualizada para el candidato actual. El panel informa firewall indeterminado y puerto 8080 abierto solo en loopback; esto no demuestra exposición pública. |
| Auditoría de pagos Super | `GET /super/api/pagos/auditoria?limit=1` devolvió 404. | La nueva ruta local aún no está desplegada. |
| Disponibilidad | `GET /ready` devolvió 200. | Solo acredita disponibilidad básica del proceso. |

## Controles negativos de seguridad

| Control | Resultado |
| --- | --- |
| Lectura empresarial sin cookies | 401 JSON saneado. |
| Query `empresa_id=12` con cabecera empresarial contradictoria | 400 JSON saneado; no se devolvieron datos. |
| Respuestas 4xx | Incluyen identificadores de correlación y no mostraron SQL, DSN ni secretos. |
| HSTS | Ausente tanto en login como en respuestas API del dominio publicado. |
| CSP del login | Continúa permitiendo `unsafe-inline`, `img-src https:` y `connect-src https: wss:` amplios. |
| CSP empresarial/API | Más restringida que la del login, pero conserva `unsafe-inline`. |

## Hallazgos del navegador

- La consola de la pestaña registró un `TypeError` de `MutationObserver` durante la navegación autenticada. No se reprodujo al recargar Seguridad VPS con captura de excepciones activa, por lo que falta aislar la página y el origen exactos antes de atribuirlo a un módulo.
- Domótica reprodujo un error visible y determinista: `isDoorSensor is not defined`. La API subyacente respondió 200, pero la página no terminó de representar el resumen operativo. La fuente local ya contiene la inicialización previa que falta en el runtime publicado.
- El primer intento de selección de empresa desde el shell Super dejó el panel empresarial dentro del iframe de Super. El comportamiento corregido en el árbol local aún no aparece en producción.
- MRP y WMS se presentaron correctamente en ancho móvil observado. Esta inspección no recorrió todos los tabs, diálogos ni tablas con datos voluminosos.

## Dictamen

La autenticación, el tenant PCS y las APIs de lectura examinadas funcionan en Productos/Inventario, Compras, Finanzas, MRP, WMS, Tesorería, CRM, Auditoría, Usuarios, Facturación y Domótica. Los controles negativos básicos de sesión y conflicto de tenant también funcionan. El entorno publicado permanece **NO-GO para acreditar las reparaciones del informe**, porque no contiene todavía el candidato local completo, Domótica falla al renderizar, y conserva MFA privilegiado sin exigencia visible, HSTS ausente, CSP amplia, auditoría de pagos sin ruta y un escaneo VPS antiguo con hallazgos abiertos.

Siguiente evidencia requerida: desplegar en staging el SHA exacto, aplicar las migraciones con `pcs-migrate`, ejecutar las pruebas PostgreSQL/race/Compose, repetir esta matriz autenticada y luego efectuar un escaneo VPS nuevo. Ninguna observación de esta prueba autoriza por sí misma un despliegue productivo.
