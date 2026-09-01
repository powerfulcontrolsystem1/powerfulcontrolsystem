# P110-007 — Fail-closed de ClamAV en staging

Fecha: 2026-08-12 (America/Bogota)  
Entorno: staging de PCS, empresa 12. Producción no intervenida.

## Hallazgo inicial

La primera simulación de caída aceptó un PDF limpio de QA. No fue un fallo del
handler: el backend desplegado no tenía `PCS_SUPPORTS_CLAMAV_REQUIRED` ni la
dirección del escáner, porque el Compose remoto anterior no incluía el overlay
de antivirus. Los dos soportes de QA creados durante esa comprobación no se
contabilizaron ni generaron pagos, asientos o documentos DIAN; permanecen
identificados para su depuración auditada en staging.

## Corrección aplicada

Se promovió el candidato inmutable `e308ca4b` con los cuatro archivos Compose
requeridos, incluido `docker-compose.staging-antivirus.yml`. Tras la promoción,
la configuración efectiva del backend mostró antivirus obligatorio y la
dirección interna de ClamAV. El escáner volvió a healthy con política de
reinicio normal y staging recuperó `/ready`.

## Prueba visual y resultado

1. Se desactivó temporalmente el reinicio de ClamAV y se detuvo el contenedor.
2. Desde la pantalla autenticada **Captura inteligente de compras y gastos** se
   intentó radicar un PDF limpio de QA con identificador único.
3. La interfaz mostró: `El servicio no esta disponible en este momento. Intenta
   de nuevo en unos segundos.`
4. Los indicadores y la bandeja no incrementaron después del intento: el
   archivo no llegó a persistirse mientras el antivirus estaba indisponible.
5. ClamAV se restauró, volvió a responder al ping interno y staging quedó listo.

## Límites de certificación

La sonda EICAR se ejecutó desde el VPS de staging en un archivo XML de prueba,
por el endpoint HTTP autenticado y con CSRF oficial. El resultado fue `422` con
el mensaje de rechazo de antivirus; no se creó soporte ni archivo. La misma
firma fue confirmada por el socket de ClamAV. El uso de XML evita que el propio
antivirus local del equipo elimine la sonda antes de la solicitud y demuestra
que el escaneo ocurre antes de la validación de formato o persistencia.

Como limpieza, los siete soportes exclusivamente QA creados durante los intentos
anteriores fueron enviados a papelera mediante la API autenticada; ninguno fue
contabilizado ni afectó pagos, DIAN o producción.

Faltan recuperación bajo carga, concurrencia y alertas P0 con el digest actual.
P110-007 permanece **parcial** y el veredicto global sigue siendo **NO-GO**.
