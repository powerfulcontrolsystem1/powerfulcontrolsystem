# P110-007 - linea base CSP y Grafologia sin inline

Fecha: 2026-08-12
Candidato base: `70bed763`

## Inventario reproducible

El nuevo auditor recorre todas las paginas HTML y separa scripts inline sin
proteccion, bloques `style`, atributos `style`, eventos `on*` y URLs
`javascript:`. La primera medicion, antes de corregir Grafologia, observo 1.347
bloqueos en 243 paginas; 975 pertenecian al area empresarial.

La linea base comprometida se toma despues de la correccion de Grafologia:

- 1.335 bloqueos en 242 paginas.
- 223 scripts inline sin proteccion.
- 187 bloques `style` sin proteccion.
- 906 atributos `style`.
- 19 eventos inline.
- cero URLs `javascript:`.
- 963 bloqueos en `administrar_empresa`.

La compuerta de CI falla si cualquier total o area aumenta y tambien si se
elimina la linea base. Los valores son maximos temporales: una reduccion pasa
sin editar el baseline; solo debe bajarse el baseline al cerrar un lote.

## Reduccion material

Grafologia movio su bloque de estilos y once atributos inline a
`web/grafologia.css`. La pagina queda con cero bloqueos inventariados y conserva
sus temas, grillas, responsive y estados dinamicos por clases. Esto reduce doce
bloqueos y una pagina completa de la deuda empresarial.

La revision visual local cargo la hoja externa sin errores de consola. En vista
de escritorio conservo dos columnas; en el viewport movil la grilla paso a una
columna y los KPI a dos columnas. En ambos casos `scrollWidth` fue menor que
`innerWidth`, sin desborde horizontal. No se invocaron API ni camara.

## Segundo lote: eventos inline en cero

Se externalizaron los 19 atributos `on*` restantes de nueve superficies:
Productos, carrito, ventas, facturas electronicas, Domotica, carta publica,
publicaciones sociales, asesores comerciales y certificado tributario publico.
El helper comun `web/js/ui_event_controls.js` atiende impresion/cierre de
ventanas, autoimpresion controlada y respaldo seguro de imagenes; las acciones
de negocio conservan listeners explicitos o delegados propios de cada modulo.

La linea base decreciente queda en:

- 1.316 bloqueos en 242 paginas.
- 223 scripts inline sin proteccion.
- 187 bloques `style` sin proteccion.
- 906 atributos `style`.
- cero eventos inline y cero URLs `javascript:`.
- 948 bloqueos en `administrar_empresa`, 216 en `super` y 138 en `publico`.

La sintaxis de los diez scripts HTML afectados y del helper externo aprobo. La
suite completa `go test ./...` aprobo; tambien aprobaron preflight estricto,
auditorias profesionales, permisos, seguridad, UX, migraciones, OpenAPI y
Docker Compose. Se corrigieron dos contratos de prueba: el inventario estatico
ya reconoce solo atributos HTML completos y el carrito exige listener explicito
sin restaurar `onclick`.

## Tercer lote: secciones de configuracion compartidas

Las seis paginas ligeras de Cobro operativo, Formato monetario, Pasarelas de
pago, Perfil tributario Colombia, Productos y pedidos y Respaldo dejaron de
duplicar el mismo bloque `style`. Todas cargan ahora
`web/administrar_empresa/configuracion/seccion_configuracion.css`.

El inventario baja a 1.310 bloqueos en 236 paginas: 181 bloques `style` y 942
bloqueos empresariales. La revision visual local comprobo las seis rutas, la
hoja cargada, encabezado responsive, iframe presente y cero desborde horizontal.
La vista de escritorio conservo el encabezado flexible; el breakpoint movil
paso a grilla y boton ancho. La consola de la pagina directa quedo limpia. Un
error observado solo al instrumentar iframes fue aislado al script interno del
navegador (`url` vacia/Electron), no a un recurso PCS.
Despues del lote aprobaron `go test ./...`, `go vet ./...`, preflight estricto,
auditorias profesionales, permisos, seguridad, UX, migraciones y Docker Compose.

## Guardia contra CI congelado

La primera ejecucion remota de la PR completo preflight, pruebas normales y
escaneo de secretos, pero permanecio sin finalizar durante horas en el detector
de carreras y la construccion de imagenes. Se cancelo esa corrida inconclusa y
se relanzo. Para evitar que un paso congelado consuma toda la ventana del job,
el detector de carreras queda limitado a doce minutos con timeout interno de
diez, y la construccion de imagenes a veinte minutos. Windows local no puede
ejecutar `-race` porque el runtime disponible no tiene CGO; la evidencia valida
de ese control debe provenir del runner Linux obligatorio.

## Limite y veredicto

La CSP aplicada todavia necesita `unsafe-inline` debido a los 1.310 bloqueos
restantes. La politica Report-Only estricta continua siendo la observacion
segura. P110-007 permanece parcial y el veredicto es NO-GO. No se ejecuto `rs`.
