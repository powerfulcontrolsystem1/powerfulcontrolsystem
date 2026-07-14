# Plan 101 - Arquitectura modular y evolucion sostenible

Estado: implementado de forma incremental el 2026-07-13.

## Decision arquitectonica

PCS continua como un monolito modular en Go y PostgreSQL. No se introducen
microservicios ni dependencias nuevas: los limites de modulo se conservan en
handlers con permisos, casos de uso ya consolidados y acceso a datos por
paquetes `db`. Este enfoque evita duplicar reglas de caja, inventario, impuestos
y facturacion mientras la carga aun no justifica servicios independientes.

## Hallazgos y controles vigentes

- Las rutas empresariales se registran bajo wrappers `WithEmpresa...Permissions`.
  El wrapper valida empresa, acceso, licencia, rol, pagina y consistencia entre
  query, cabecera y JSON antes de insertar el `empresaID` validado en contexto.
- Los modulos de mayor complejidad siguen siendo carrito, pagos, reportes,
  permisos y modulos verticales. Su separacion es gradual: no se movio codigo
  estable solo por reducir lineas, para no introducir regresiones operativas.
- La API movil v1 es una fachada aditiva: reutiliza las reglas existentes de
  venta, caja y facturacion, y no crea una segunda implementacion fiscal.
- Las mutaciones moviles de carrito, items, cobro, sincronizacion offline y
  emision fiscal requieren `Idempotency-Key`, se almacenan por empresa con hash
  y solo repiten una respuesta exitosa.
- El contrato movil preserva su propio sobre JSON incluso cuando un handler v1
  genera la respuesta; evita respuestas anidadas y entrega `request_id`.

## Reglas de evolucion

1. Todo endpoint empresarial debe estar detras de un wrapper de empresa y usar
   el `empresaID` del contexto como autoridad.
2. Un flujo nuevo debe reutilizar una regla de negocio existente o extraer una
   funcion pura/probable; no copiar SQL ni calculos entre handlers.
3. Operaciones repetibles o de efecto economico/fiscal deben exigir una clave de
   idempotencia y conservarla por empresa.
4. Exportaciones, adjuntos, tareas externas y reportes masivos permanecen
   aislados por empresa, con limites de pagina/tamano y sin exponer rutas de
   almacenamiento privado.
5. Antes de separar un servicio se debe demostrar una necesidad de carga,
   disponibilidad, seguridad o despliegue independiente; el limite de modulo
   debe existir primero dentro del monolito.

## Riesgos que requieren evidencia de preproduccion

- Pruebas de carga con pool PostgreSQL y datos anonimizados.
- Restauracion completa en entorno desechable.
- Confirmacion real de webhooks de pagos, DIAN y proveedores externos.
- Pruebas E2E de navegadores y aplicaciones moviles contra staging.

Estos son gates operativos, no defectos que puedan resolverse de forma segura
con cambios locales sin infraestructura de preproduccion.
