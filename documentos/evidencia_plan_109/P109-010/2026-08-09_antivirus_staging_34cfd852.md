# P109-010 - estado real del antivirus en staging 34cfd852

Fecha: 2026-08-09

La inspección del backend exacto desplegado consultó únicamente variables y
métricas específicas del antivirus, sin exponer secretos ni datos de empresa.

| Señal | Valor |
| --- | ---: |
| `pcs_support_antivirus_required` | 0 |
| `pcs_support_antivirus_configured` | 0 |
| Escaneos clean/malware/unavailable/bypassed | 0/0/0/0 |
| Contenedor clamd/antivirus en el VPS | inexistente |

Resultado: **BLOCKED**, no PASS. Staging acepta que el antivirus sea opcional y
no tiene endpoint clamd. Las pruebas unitarias/simuladas existentes demuestran
el contrato, pero no sustituyen firmas reales actualizadas ni una muestra EICAR
controlada. Antes del piloto se debe incorporar y operar un servicio clamd,
activar `PCS_SUPPORTS_CLAMAV_REQUIRED=1`, comprobar rechazo EICAR, aceptación de
archivo limpio, fallo cerrado si clamd cae y alertas Prometheus.

No se cambió staging ni producción durante esta inspección.
