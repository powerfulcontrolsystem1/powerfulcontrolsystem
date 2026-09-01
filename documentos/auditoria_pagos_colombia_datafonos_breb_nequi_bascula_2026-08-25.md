# Auditoria de pagos Colombia, datafonos, Bre-B, Nequi y bascula

Fecha de corte: 2026-08-25.

## Alcance

- Datafonos Redeban, CredibanCo, Bold y BBVA vinculados al POS.
- Pagos manuales con tarjeta, transferencia Bre-B y transferencia Nequi.
- Wompi Web Checkout, consulta de transaccion y webhook para licencias.
- QR Bre-B/Nequi configurado por empresa.
- Bascula serial/USB mediante Web Serial en el carrito.

## Decisiones de seguridad aplicadas

- El monto de un cobro de datafono asociado a carrito se calcula en servidor con total menos abonos y se vuelve a comprobar antes del cierre.
- El cliente HTTP de datafonos bloquea HTTP, credenciales en URL, destinos privados/reservados, DNS hacia redes privadas y redirecciones fuera del origen.
- El adaptador JSON generico queda bloqueado por defecto. Solo se habilita un proveedor incluido expresamente en `PCS_DATAFONO_HOMOLOGATED_PROVIDERS` despues de contrato y homologacion.
- Una aprobacion de datafono exige identificador, referencia, monto y moneda coincidentes. Las respuestas crudas se depuran recursivamente y la solicitud de auditoria no conserva PII del comprador.
- Bre-B no anuncia conciliacion automatica mientras no exista un conector bancario firmado. Todo registro manual nace `pendiente` y se limita a COP.
- Las llaves Bre-B se validan por tipo y los QR configurables son payloads estaticos exactos. PCS no fabrica un QR bancario ni modifica plantillas EMV.
- Wompi usa el secreto de eventos y el orden dinamico de `signature.properties`, seguido de `timestamp`, para comprobar SHA-256. La `integrity key` queda reservada al checkout.
- Una aprobacion Wompi solo activa cuando referencia, monto en centavos, moneda, empresa, licencia y ambiente coinciden con el registro previo.
- La bascula requiere varias muestras coherentes, frescura maxima y ausencia de marca de inestabilidad antes de copiar el peso o agregar un producto.

## Estado de produccion

### Listo en codigo local

- Controles de aislamiento por `empresa_id`, validacion de hijos y cierre fail-closed.
- Idempotencia existente por referencias y actualizaciones condicionadas, reforzada con comprobacion de evidencia.
- Secretos Wompi cifrados y enmascarados, incluido `wompi.events_secret`.
- Mensajes de interfaz que distinguen registro manual, QR estatico, integracion real y control metrologico.

### Bloqueos externos obligatorios

- Redeban/CredibanCo/BBVA: contrato, especificacion privada, terminales y homologacion TEF/adquirente.
- Bold: comercio habilitado para integracion, Smart o Smart Pro compatible, serial/modelo y pruebas certificadas de App Checkout.
- Wompi: llaves reales, secreto de eventos, URL de evento registrada por ambiente y pruebas sandbox/produccion con montos controlados.
- Bre-B/Nequi: QR estatico emitido por participante o API oficial para QR dinamico; webhook/API bancaria autenticada para conciliacion automatica.
- Bascula: instrumento apto para la actividad, verificacion metrologica aplicable, protocolo documentado y prueba con el hardware real.
- PCI DSS: confirmacion de alcance y responsabilidades con adquirentes/proveedores; PCS no debe recibir PAN, CVV ni datos de pista.

Mientras falte cualquiera de estos puntos, el componente afectado es `NO-GO` para cobro automatico real aunque el codigo local este preparado.

## Fuentes oficiales consultadas

- Banco de la Republica, reglamentacion Bre-B: https://www.banrep.gov.co/es/normatividad/sistemas-pago/pagos-inmediatos-bre-b
- Banco de la Republica, compendio DSP-465: https://www.banrep.gov.co/sites/default/files/reglamentacion/archivos/compendio-dsp-465.pdf
- SIC, control metrologico de instrumentos de pesaje no automaticos: https://sedeelectronica.sic.gov.co/transparencia/normativa/control-metrologico-de-instrumentos-de-pesaje-de-funcionamiento-no-automatico-balanzas
- PCI SSC, PCI DSS: https://www.pcisecuritystandards.org/document_library/?class=pcidss&doc=pci_dss
- Wompi Colombia, eventos: https://docs.wompi.co/docs/colombia/eventos/
- Wompi Colombia, transacciones: https://docs.wompi.co/docs/colombia/transacciones/
- Wompi Colombia, seguimiento: https://docs.wompi.co/docs/colombia/seguimiento-de-transacciones/
- Bold, App Checkout: https://developers.bold.co/api-integrations/integration
- Redeban, productos TEF: https://www.redeban.com/nuestros-productos

## Evidencia minima antes de GO

1. Pruebas enfocadas Go y validacion de scripts embebidos sin fallos.
2. Prueba de concurrencia: carrito modificado entre autorizacion y cierre no se cierra.
3. Webhook Wompi valido, firma alterada, monto alterado, referencia ajena y ambiente cruzado.
4. Prueba real por proveedor homologado con aprobado, rechazado, timeout, reintento e idempotencia.
5. Prueba de bascula con peso inestable, estable, lectura vencida, tara local y desconexion.
6. Evidencia visual desktop/movil y comprobacion de consola/red.
7. Preflight, CI, despliegue y smoke en el entorno objetivo; una prueba local no equivale a produccion.
