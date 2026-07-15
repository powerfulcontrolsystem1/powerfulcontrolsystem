# Auditoria modular de preproduccion - 2026-07-14

## Alcance y conclusion

Se reviso la arquitectura vigente de Powerful Control System antes de continuar
el desarrollo paralelo web y movil. PCS conserva la decision correcta para su
etapa actual: un monolito modular en Go y PostgreSQL. No se dividio en
microservicios ni se extrajo codigo estable solo para reducir el tamano de los
archivos, porque eso aumentaria el riesgo sobre venta, caja, inventario y
facturacion sin una necesidad operativa demostrada.

La base de crecimiento es la separacion por limites de modulo, contratos API
versionados, empresa validada desde contexto, colas persistentes y pruebas
enfocadas. Cada modulo nuevo debe integrarse por esos limites y no mediante
acceso directo a detalles internos de otro modulo.

## Hallazgos atendidos

| Hallazgo | Decision aplicada | Resultado |
|---|---|---|
| Worker persistente podia dejar tareas en `processing` despues de una caida. | Recuperacion por lease vencido antes de tomar lote. | Las tareas vuelven a `pending` o pasan a `dead` al agotar reintentos. |
| Parada ordenada podia mantener tareas bloqueadas hasta vencer el lease. | Liberacion por `worker_id` al cerrar el runner. | Un relevo puede continuar de inmediato. |
| Reclamo concurrente de una tarea no verificaba filas afectadas. | Verificacion explicita de claim atomico. | Evita que dos workers ejecuten una misma tarea. |
| Error de proveedor podia quedar como texto sin filtrar en `last_error`. | Mensajes operativos genericos y auditables. | No se persisten detalles sensibles de proveedor. |
| Cliente Flutter no tenia plataformas nativas listas para empaquetar. | Scaffold Android/iOS, identificador PCS, iconos y politica de firma. | Base comun lista para Android y preparacion de iPhone. |

## Limites de modulo que se deben conservar

- `autenticacion/permisos`: identidad, sesion, rol y empresa autorizada; no
  delega su autoridad al cliente.
- `ventas/carrito/pagos/facturacion`: usan las mismas reglas fiscales y de
  inventario desde web y API v1; ninguna interfaz duplica calculos.
- `archivos/documentos`: almacenamiento privado segregado por empresa y
  descargas autorizadas, nunca rutas de disco entregadas al cliente.
- `integraciones`: DIAN, pagos, correo, WhatsApp y soporte se invocan por
  adaptadores y trabajos persistentes idempotentes.
- `movil`: consume solo `/api/v1/` por HTTPS; no accede a PostgreSQL ni hereda
  cookies, secretos empresariales o permisos calculados localmente.

## Criterio para futuras extracciones

Los archivos grandes de negocio se separan solo cuando se pueda extraer un caso
de uso con contrato, pruebas y responsabilidad unica. Antes de crear un servicio
independiente debe existir evidencia de carga, aislamiento operativo,
disponibilidad o seguridad que el monolito modular no pueda satisfacer.

## Gates de produccion que requieren infraestructura externa

No son defectos locales ni se pueden declarar aprobados sin evidencia real:

1. Restauracion de una copia anonima en staging desechable.
2. Carga controlada sobre API v1, PostgreSQL y worker.
3. Webhooks reales firmados de pagos, DIAN y canales externos.
4. E2E web y Android contra staging con cuentas de prueba aisladas.
5. Compilacion y firma iOS en Mac con Xcode y certificado de distribucion.

Estos gates quedan en el checklist de liberacion; no deben omitirse al pasar de
desarrollo a trafico publico.
