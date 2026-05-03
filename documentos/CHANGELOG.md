## [2026-05-02] Gimnasio, impresoras y documentacion del proyecto
- [Gimnasio] Se robustece el esquema del modulo con migraciones defensivas para tablas antiguas de empresas, evitando errores internos al abrir dashboard, acceso, credenciales o dispositivos cuando faltaban columnas historicas.
- [Gimnasio] Se agrega preconfiguracion operativa propia del modulo: sede principal, RFID/NFC/QR, planes base, clases iniciales y dispositivos de acceso, todo aplicable desde el dashboard del gimnasio.
- [Impresoras] Se corrige el guardado de configuracion avanzada para que `modo_documento_venta` gobierne correctamente la activacion o desactivacion de facturacion electronica automatica.
- [Impresoras] Se incorpora `cajon_monedero` como funcionalidad asignable de impresora dentro de `Configuracion > Impresora`, alineando la UI con la operacion real de caja.
- [Docs] Se actualiza `RESUMEN_DEL_PROYECTO.md` para reflejar configuracion guiada por IA, impresion empresarial, horarios laborales y modulos verticales ya integrados como gimnasio, odontologia, taxi system, turnos de atencion y alquileres.

## [2026-04-30] Pagos, chat IA, empresas compartidas, hoja de vida operativa y documentos dinamicos
- [Pagos/Epayco] Smart Checkout v2 conserva fallback clasico firmado por POST a `https://secure.payco.co/checkout.php`; se elimina la redireccion GET que producia XML `AccessDenied` y se documenta el requisito de `epayco.customer_id` para fallback.
- [Pagos/Epayco] El fallback clasico resuelve su modo con `epayco.customer_id` + `epayco.checkout_key`/`epayco.p_key`, separado de las llaves Smart Checkout, para no enviar cuentas reales como pruebas y evitar el error "El comercio no fue reconocido".
- [Chat IA] La secretaria IA 3D se rediseña como avatar estilo caricatura ejecutiva joven y habla siempre con voz femenina (`es-CO-female`), manteniendo el robot con voz configurable.
- [Empresas compartidas] El editor de empresa permite consultar y retirar administradores compartidos desde ambos lados del acceso, con trazabilidad del actor.
- [Administrar empresa] Se implementa la hoja de vida operativa universal para motos de taller, pacientes, vehiculos, equipos, activos o mascotas, con ficha, eventos, servicios, alertas y resumen operativo.
- [Documentos IA] Se documenta el flujo `/generate` + `/download` para generar documentos dinamicos con IA/templates y exportar PDF, DOCX, XLSX, HTML, TXT o JSON.
- Nueva funcionalidad: MÃ³dulo Red Social Comercial con portal pÃºblico y administraciÃ³n por empresa. EliminaciÃ³n de modulo juegos y venta de licencias desde cliente.

## [2026-04-23] Retiro Tipos de usuario (panel super)
- [Super/DB] Eliminación del módulo Tipos de usuario: sin API ni UI; tabla `tipos_de_usuario` removida al arranque; documentación alineada.

## [2026-04-23] reCAPTCHA, backup y manual de instalación
- [Docs/Operación] Se actualizó el manual de instalación con reCAPTCHA v2/v3/Enterprise, variables, panel super y fallos frecuentes (dominios, tipo de clave). Se documentan las claves y copias best-effort en `backup/super_administrador` y `backup/empresas/<empresa_id>`. Ajustes en `descripcion_de_archivos` e `historial_de_cambios` y alineación con `CHANGELOG.md` raíz.

## [2026-04-20] Limpieza Total Themes
- [UI/Temas] Auditoría y barrido de más de 50 páginas y scripts en web/administrar_empresa, web/super y páginas públicas para limpiar colores fijos, migrando lógicas JS a .classList.add('text-danger') y respetando las 6 paletas dinámicas. Completado barrido masivo de vistas.
- **2026-04-30 - Pagos ePayco de licencias**: el fallback estandar ahora usa `checkout.js` con `external: "true"` y `PUBLIC_KEY`; se evita el POST legacy a `secure.payco.co/checkout.php`, `P_KEY` queda solo en backend para validacion de webhooks con SHA256 y el frontend `pagar_licencia.html` soporta `checkout_type=classic_js`. Verificacion: `go test ./handlers -run Test.*Epayco -count=1` y `go test ./... -count=1`.

## [2026-05-03] Documentacion, ayuda y estado operativo de modulos
- [Docs] Se crea `documentos/reporte_estado_modulos_2026-05-03.md` con estado compacto por modulo, observaciones de calidad y dependencias pendientes de certificacion.
- [Ayuda] Se actualiza `web/ayuda/ayuda.html` con una seccion de estado operativo, estaciones/carrito, tarjetas adaptables, indicadores del panel y limites honestos de validacion.
- [Operacion] Se documentan los cambios recientes: carrito desde estacion, pago con retorno a estaciones, `USD / COP` primero y despliegue VPS correcto.
