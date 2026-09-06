# Modulo Vida

Estado: Vigente. Responsable: Ingeniería del módulo. Revisión documental: 2026-09-05.

## Alcance revisado y límites

- Datos privados por empresa y usuario, sin integración comercial automática.
- La migración inicial crea gastos/suscripciones; la de precios añade empresa_vida_precios. Recordatorios y entregas requieren evidencia separada de la programación local.

Esta revisión contrasta documentación con las fuentes locales citadas; no ejecuta el flujo comercial ni acredita UI, proveedor, hardware o producción. Las pruebas y estados fechados del cuerpo son antecedentes, no resultados nuevos.

## Proposito

Vida es un registro personal de gastos cotidianos y suscripciones dentro de
cada empresa. No crea movimientos contables, cuentas por pagar, inventario,
compras empresariales, caja, impuestos ni documentos fiscales. La empresa solo
aporta el contexto de acceso; la fuente de verdad pertenece a la cuenta
autenticada.

Casos principales:

- registrar supermercado, alimentacion, transporte, salud, hogar, educacion,
  familia, entretenimiento, servicios, ropa, mascotas y otros gastos;
- tomar o seleccionar una foto del recibo desde un telefono, o adjuntar un PDF,
  y registrar manualmente fecha, comercio, categoria, valor y medio de pago;
- leer codigos de barras con la camara del celular usando `BarcodeDetector`,
  con entrada manual cuando el navegador no ofrezca esa API;
- cargar una factura con el boton explicito `IA: leer y registrar factura`,
  extraer sus datos y productos, y crear el gasto personal en una transaccion;
- consultar por producto o codigo el historial de precio unitario pagado;
- consultar resumen mensual, promedio diario y distribucion por categoria;
- consultar reportes filtrables por periodo (hasta 366 dias), categoria,
  comercio y medio de pago, con totales y agrupaciones por dia, categoria,
  comercio y medio de pago;
- registrar suscripciones semanales, mensuales, trimestrales, semestrales,
  anuales o personalizadas;
- recordar renovacion, cancelacion o ambas antes de la fecha configurada.

La captura manual conserva el archivo como evidencia privada. La captura IA es
un flujo separado: el texto del boton informa que el clic autoriza la lectura y
el alta automatica. La respuesta se valida en servidor; si la confianza es baja
se registra con aviso de revision y el gasto permanece editable. Nunca crea una
compra empresarial, proveedor, inventario, asiento contable ni documento fiscal.

## Aislamiento y autorizacion

Toda consulta, alta, actualizacion, renovacion, descarga y borrado filtra por la
combinacion `empresa_id + usuario_id + id`. `usuario_id` se deriva en backend de
la cuenta autenticada normalizada; nunca se acepta desde el navegador. Un
usuario de la misma empresa no puede listar ni abrir los registros de otro.

El permiso `vida` esta disponible para todos los roles empresariales
autenticados y no depende del catalogo comercial de la licencia. No concede
lectura administrativa de registros ajenos. El frontend oculta o muestra el
acceso, pero la autorizacion efectiva vive en
`WithEmpresaVidaPermissions` y en los filtros de persistencia.

Los comprobantes se guardan bajo `PCS_PRIVATE_STORAGE_DIR/vida/empresa_<id>` o
su raiz privada predeterminada, fuera de `web/`, con nombre aleatorio, permisos
restrictivos, maximo 10 MiB y validacion de contenido para JPEG, PNG, WebP o
PDF. La descarga primero comprueba que el gasto pertenece a la cuenta y empresa
solicitantes.

## Superficies y contrato

- Menu: `Vida` en `web/administrar_empresa.html`.
- Pagina: `web/administrar_empresa/vida.html`.
- Logica: `web/js/vida.js`.
- Estilos: `web/administrar_empresa/vida.css`.
- API: `/api/empresa/vida?empresa_id=<id>&action=<accion>`.

Acciones:

| Metodo | Accion | Resultado |
|---|---|---|
| `GET` | `dashboard` | Resumen del mes, gastos recientes, suscripciones y alertas del usuario |
| `GET` | `gastos` | Lista personal filtrable por fecha y categoria |
| `GET` | `suscripciones` | Lista personal y alertas vigentes |
| `GET` | `precios` | Historial personal filtrable por codigo de barras o producto |
| `GET` | `reporte` | Totales y agrupaciones personales filtrables por fecha, categoria, comercio y medio de pago |
| `GET` | `notificaciones` | Preferencias privadas de aviso, con telefono enmascarado |
| `GET` | `recibo&id=<id>` | Descarga privada del comprobante propio |
| `POST` | `gasto` | Crea gasto; acepta JSON o `multipart/form-data` con `recibo` |
| `POST` | `factura_ia` | Lee imagen/PDF y crea atomicamente gasto, recibo privado y productos |
| `PUT` | `gasto` | Actualiza los datos confirmados del gasto propio |
| `DELETE` | `gasto` | Elimina el gasto propio y su comprobante privado |
| `POST` | `suscripcion` | Crea una suscripcion personal |
| `PUT` | `suscripcion` | Actualiza la suscripcion propia |
| `PUT` | `notificaciones` | Activa/desactiva email y WhatsApp propios, numero privado y hora local |
| `POST` | `renovar&id=<id>` | Avanza la proxima fecha con cierre correcto de fin de mes |
| `DELETE` | `suscripcion` | Elimina la suscripcion propia |

Las altas requieren `Idempotency-Key` o `client_request_id`, ligada a empresa,
usuario y huella de la solicitud. Repetir el mismo intento devuelve el registro
original; reutilizar la clave con otro contenido produce conflicto.

## Persistencia

La migracion `20260831-001-vida-personal-v1`, ejecutada exclusivamente por
`pcs-migrate`, crea gastos y suscripciones. La migracion aditiva
`20260831-003-vida-price-history-ai-v1` añade precios. Distribución de tablas:

- `empresa_vida_gastos`, con fecha, categoria, comercio, descripcion, monto,
  moneda, medio de pago y referencia privada del comprobante;
- `empresa_vida_suscripciones`, con periodicidad, intervalo, proxima
  renovacion, anticipacion, tipo de recordatorio, autorrenovacion y estado.
- `empresa_vida_precios`, con gasto de origen, fecha, codigo de barras, nombre,
  comercio, cantidad, precio unitario/total, moneda y origen manual, lector o IA.

La migracion aditiva `20260905-001-vida-reports-reminders-v1` crea
`empresa_vida_notificacion_configuracion` y `empresa_vida_notificaciones`.
La primera mantiene opt-in por `empresa_id + usuario_id`, canal, hora local y
telefono WhatsApp privado; la segunda registra un unico intento por
suscripcion, fecha de renovacion y canal, y permite hasta cinco reintentos
espaciados para errores de entrega. No se guardan destinatarios duplicados en
el historial de envios.

Los indices comienzan por `empresa_id, usuario_id`; el producto personal no
referencia `productos` ni inventario de la empresa. Gasto y lineas de precio se
crean juntos y `ON DELETE CASCADE` elimina las lineas si se elimina el gasto. La
repeticion de una factura IA con la misma clave devuelve el resultado anterior
sin consumir de nuevo al proveedor. API y worker no ejecutan DDL.

## Privacidad de IA

Vida reutiliza la integracion OpenAI Responses configurada para PCS: foto o PDF
como entrada, `store=false`, `safety_identifier` seudonimo y clave resuelta en
backend. El archivo se trata como contenido no confiable; no puede dar
instrucciones al sistema. La salida JSON tiene limites de tamano, fecha, moneda,
montos, confianza y maximo 200 productos antes de persistirse.

## Recordatorios

La pagina muestra alertas cuando una suscripcion activa entra en su ventana de
anticipacion o esta vencida. El menu de Administrar empresa consulta el resumen
al iniciar y cada 15 minutos, muestra un contador y, si el usuario ya concedio
permiso de notificaciones al navegador, emite una notificacion deduplicada por
empresa, suscripcion y fecha de renovacion. El boton `Activar alarmas` solicita
ese permiso de forma explicita.

Estas alarmas del navegador requieren que PCS este abierto. Adicionalmente,
cada usuario puede activar email a su propia cuenta autenticada y/o WhatsApp a
un numero que registra expresamente, y escoger la hora local de revision. El
`pcs-worker` ejecuta el envio durable; respeta el interruptor global del evento
`vida_suscripcion`, el modo de pruebas de correo/WhatsApp y la preferencia
individual. El API nunca retorna el telefono completo. Desactivar un canal
detiene futuros avisos, sin afectar suscripciones ni gastos.

WhatsApp depende de que Super Administrador haya configurado el proveedor Meta
Cloud y habilitado el evento `vida_suscripcion`; si no lo esta, el intento queda
registrado como omitido y el usuario no recibe un envio ficticio. Vida no
ejecuta renovaciones, cancelaciones ni cobros con proveedores externos.

## Activacion y validacion pendiente

El candidato permanece local. Antes de habilitarlo en un entorno se debe:

1. ejecutar `pcs-migrate` con el rol propietario autorizado;
2. desplegar backend y frontend del mismo candidato;
3. probar dos usuarios de una misma empresa y otra empresa distinta;
4. probar foto real, PDF, lectura IA, reintento idempotente, limites, descarga y borrado;
5. probar camara y fallback manual en movil compatible y no compatible;
6. verificar reportes filtrados con dos usuarios de la misma empresa y otra
   empresa, sin datos cruzados;
7. verificar email y WhatsApp en modo de pruebas, opt-in/opt-out, numero
   enmascarado, deduplicacion y reintento; usar un dispositivo/servicio real
   antes de afirmar entrega productiva.

Esta documentacion no autoriza migracion, despliegue ni cambios de produccion.

## Fuentes y aceptación de la revisión

[vida.go](../backend/handlers/vida.go), [vida.go](../backend/db/vida.go), [vida_test.go](../backend/db/vida_test.go), [vida.html](../web/administrar_empresa/vida.html), [main.go](../backend/main.go), [empresa_permisos.go](../backend/handlers/empresa_permisos.go).

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016 ([matriz transversal](requisitos/especificacion_y_trazabilidad.md)).
