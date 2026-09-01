# Modulo Vida

Estado: candidato local. Ultima actualizacion: 2026-08-31.

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
| `GET` | `recibo&id=<id>` | Descarga privada del comprobante propio |
| `POST` | `gasto` | Crea gasto; acepta JSON o `multipart/form-data` con `recibo` |
| `POST` | `factura_ia` | Lee imagen/PDF y crea atomicamente gasto, recibo privado y productos |
| `PUT` | `gasto` | Actualiza los datos confirmados del gasto propio |
| `DELETE` | `gasto` | Elimina el gasto propio y su comprobante privado |
| `POST` | `suscripcion` | Crea una suscripcion personal |
| `PUT` | `suscripcion` | Actualiza la suscripcion propia |
| `POST` | `renovar&id=<id>` | Avanza la proxima fecha con cierre correcto de fin de mes |
| `DELETE` | `suscripcion` | Elimina la suscripcion propia |

Las altas requieren `Idempotency-Key` o `client_request_id`, ligada a empresa,
usuario y huella de la solicitud. Repetir el mismo intento devuelve el registro
original; reutilizar la clave con otro contenido produce conflicto.

## Persistencia

La migracion `20260831-001-vida-personal-v1`, ejecutada exclusivamente por
`pcs-migrate`, crea gastos y suscripciones. La migracion aditiva
`20260831-003-vida-price-history-ai-v1` crea:

- `empresa_vida_gastos`, con fecha, categoria, comercio, descripcion, monto,
  moneda, medio de pago y referencia privada del comprobante;
- `empresa_vida_suscripciones`, con periodicidad, intervalo, proxima
  renovacion, anticipacion, tipo de recordatorio, autorrenovacion y estado.
- `empresa_vida_precios`, con gasto de origen, fecha, codigo de barras, nombre,
  comercio, cantidad, precio unitario/total, moneda y origen manual, lector o IA.

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

Estas alarmas del navegador requieren que PCS este abierto. No se implementa en
esta version envio por correo, SMS, WhatsApp ni una renovacion o cancelacion
automatica con el proveedor.

## Activacion y validacion pendiente

El candidato permanece local. Antes de habilitarlo en un entorno se debe:

1. ejecutar `pcs-migrate` con el rol propietario autorizado;
2. desplegar backend y frontend del mismo candidato;
3. probar dos usuarios de una misma empresa y otra empresa distinta;
4. probar foto real, PDF, lectura IA, reintento idempotente, limites, descarga y borrado;
5. probar camara y fallback manual en movil compatible y no compatible;
6. verificar recordatorios con permisos concedidos y denegados.

Esta documentacion no autoriza migracion, despliegue ni cambios de produccion.
