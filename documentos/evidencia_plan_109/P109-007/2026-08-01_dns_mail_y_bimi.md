# P109-007 - DNS de correo y límite BIMI

Fecha: 2026-08-01
Alcance: consulta DNS y HTTP de solo lectura.

- SPF: publicado para MX y `mail.powerfulcontrolsystem.com`.
- DKIM: clave pública RSA disponible en el selector `dkim`.
- DMARC: `p=quarantine`, `pct=100`, alineación SPF/DKIM estricta y reportes al
  buzón corporativo.
- PTR de `2.24.197.58`: `mail.powerfulcontrolsystem.com`.
- BIMI: `default._bimi.powerfulcontrolsystem.com` apunta al SVG oficial.
- El SVG responde 200 con `Content-Type: image/svg+xml` y `nosniff`.
- El campo BIMI `a=` continúa vacío; no existe VMC/CMC publicado.

Esta verificación no envió correo. SPF/DKIM/DMARC publicados no sustituyen una
entrega nueva con cabeceras PASS, rebote y comprobación visual. La sustitución
del avatar con inicial en clientes que exigen certificado queda pausada hasta
la adquisición externa autorizada.

Estado: **P109-007 pendiente** por DIAN, correo real, pagos, proveedores y
certificado BIMI; esta evidencia no aumenta el porcentaje.
