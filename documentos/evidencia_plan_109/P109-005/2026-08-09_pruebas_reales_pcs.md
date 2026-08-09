# P109-005/P109-007 - pruebas reales visuales y operativas PCS

Fecha: 2026-08-09.

Alcance: producción actual de la empresa PCS, con autorización expresa del
propietario. El candidato local `codex/p109-batch-no-pr` no fue desplegado; por
lo tanto esta evidencia descubre defectos reales y valida flujos existentes,
pero no certifica todavía el arreglo local.

## Barrido autenticado y visual

- Se recorrieron 24 rutas críticas en escritorio y móvil: 48 variantes, 1.260
  botones detectados, 12 clics seguros, 277 mutaciones omitidas por la guardia,
  42 variantes correctas y 6 para revisión.
- La bandeja de facturación electrónica desbordaba horizontalmente en móvil por
  la tabla de países y el texto largo de la firma. El candidato contiene
  scroll local de tabla, contención de tarjetas y corte seguro de palabras.
- Centro IA respondió 403 para la sesión de PCS; queda pendiente confirmar si
  corresponde a licencia/rol o a un wrapper incorrecto antes de cambiar acceso.
- El iframe corporativo fue bloqueado por CSP porque el origen exacto de Mailu
  no figuraba en `frame-src`. El candidato incorpora el origen exacto, sin
  comodines, tanto en Nginx estático como en el middleware dinámico.

Salida del barrido: `test_runs/p109_real_20260809_critical/` (artefacto local
ignorado por Git).

## Impresiones reales

- Se renderizó el documento real `1PCS5` en Carta, compacta y POS.
- Texto fiscal, filas, CUFE y QR cargaron; el QR tuvo dimensiones naturales.
- Carta y compacta perdieron la hoja global y los tres formatos perdieron el
  logo porque la ventana se abre desde una URL `blob:` con recursos relativos.
- El candidato convierte logo y QR de raíz a URL absoluta del mismo origen y
  enlaza `estilos.css` de forma absoluta antes de crear el blob.
- El contrato ejecutable confirmó que `/uploads/...` y `/api/qr/...` quedan
  resueltos contra `https://powerfulcontrolsystem.com`.

Salida visual: `test_runs/p109_real_20260809_prints/` (artefacto local ignorado
por Git).

## Captura CxP con IA

- Se radicó por la interfaz un UBL sintético controlado. La extracción leyó
  proveedor, identificación, número, fechas, subtotal, IVA y total.
- Se editaron campos en revisión humana y se comprobó la presentación ordenada.
- El soporte `SCI-0005` se rechazó mediante el flujo oficial.
- La consulta posterior del número controlado devolvió cero cuentas por pagar;
  no se creó pago, asiento ni cartera.

## DIAN

- El diagnóstico autenticado y la validación de credenciales respondieron
  correctamente; la configuración muestra ambiente de producción, rango y
  documentos reales recientes.
- Se encontró una inconsistencia P0: la activación local de producción se
  representa únicamente en `estado_dian`, pero cada envío posterior sobrescribe
  ese mismo campo. La interfaz puede mostrar producción local inactiva aun
  existiendo envíos productivos.
- Por seguridad fiscal no se emitió otro consecutivo ni se intentó una nueva
  anulación durante el diagnóstico. El candidato agrega el indicador persistente,
  migra solo configuraciones con evidencia histórica de activación, impide que
  el CRUD lo active y lo limpia al regresar a habilitación. Falta desplegar y
  reconciliar antes de otro efecto fiscal; la autorización de prueba no elimina
  el riesgo de duplicarlo.

## Correo corporativo

- `CONFIG_ENC_KEY` está presente. El registro PCS conservaba un cifrado legado
  sin identificador de clave y la clave anterior no estaba configurada.
- Se renovó la contraseña por la página oficial, sin mostrarla ni escribirla en
  archivos. El registro quedó versionado con el identificador activo y dejó de
  fallar el descifrado.
- IMAP continuó rechazando la autenticación. El adaptador actual enviaba POST a
  Mailu primero y solo usaba PATCH si recibía exactamente 409; en esta
  instalación el POST de un usuario existente no renovó la clave aunque reportó
  éxito. El candidato cambia a PATCH primero y POST únicamente ante 404.
- Falta desplegar el candidato y repetir guardado, IMAP, autologin e iframe.

## Validación local del candidato

- `go test ./utils ./handlers`: PASS.
- `go test .`: PASS.
- sintaxis `print_documents.js`: PASS.
- contrato ejecutable de activos absolutos en blob: PASS.
- `profesional_preflight.ps1 -Full`: PASS completo.
- `git diff --check`: PASS.

## Cierre

P109-005 y P109-007 continúan parciales. La evidencia real aumentó cobertura y
produjo correcciones concretas, pero el candidato sigue sin despliegue
inmutable, repetición visual, validación A/B ni cierre de la inconsistencia DIAN.
El Plan 109 permanece en 53,3 % de implementación, 0 % de certificación del
arreglo local y NO-GO.
