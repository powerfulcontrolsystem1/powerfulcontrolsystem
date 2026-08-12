# P110-009 — Entrega controlada de Alertmanager por Mailu

Fecha: 2026-08-12 (America/Bogota)  
Entorno: VPS de **staging**; producción no intervenida.  
Alcance: habilitar y comprobar un canal externo de alerta sin versionar
destinatarios, contraseñas ni material de correo.

## Cambio aplicado

- `alertmanager.yml.tmpl` es una plantilla versionada sin secretos.
- `vps-monitoring-up.sh` genera `alertmanager.runtime.yml` desde el archivo
  privado `.env.monitoring`, valida su configuración y espera la disponibilidad
  del servicio antes de declararlo sano.
- El contenedor Alertmanager se une exclusivamente a la red Docker privada de
  Mailu; no se abrió ningún puerto nuevo al exterior.
- El archivo de ejecución se genera con permiso de lectura para el usuario no
  privilegiado de Alertmanager. No contiene credenciales SMTP porque Mailu se
  consume por relay interno.

## Ejecución observada

1. Alertmanager quedó listo por loopback y `amtool check-config` aprobó la
   configuración cargada.
2. La auditoría detectó dos receptores y confirmó al menos una clase de entrega
   externa configurada.
3. Se publicó una alerta sintética única `P110MailDeliveryTest`, con etiqueta
   `p110_test=true`, sin relación con datos de empresas, ventas o DIAN.
4. Alertmanager la registró activa y luego se envió la resolución; la consulta
   posterior mostró cero alertas sintéticas activas.
5. El relay Mailu aceptó y entregó el aviso externo con respuesta SMTP `2.0.0`.
   Los datos de destino y cola se mantuvieron fuera de esta evidencia.

## Resultado y límites

- Entrega de alerta de disparo: comprobada hasta aceptación del servidor remoto.
- Resolución: registrada y la alerta desapareció de Alertmanager; la recepción
  visible de su mensaje queda pendiente de confirmación del destinatario.
- Deduplicación, guardia/on-call, carga mutante de cuatro cajas y decisión de
  piloto aún no se certifican. Por ello P110-009 continúa **parcial**, la
  certificación global permanece en **0 %** y el estado sigue siendo **NO-GO**.
