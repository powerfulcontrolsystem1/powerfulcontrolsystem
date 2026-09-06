# Revisión de seguridad del VPS y de PCS

Estado: Evidencia. Responsable: Seguridad y QA/operación. Revisión documental: 2026-09-06.

## Dictamen y alcance

Se corrigieron controles del VPS en producción y defectos del código local.
**La revisión permanece abierta: hay vulnerabilidades altas/críticas en imágenes
auxiliares, un certificado wildcard vencido, MFA privilegiado pendiente y un
reinicio pendiente para activar actualizaciones del host. No es una certificación
de seguridad ni demuestra resistencia a DDoS volumétrico.**

Autorización: revisión y reparación general solicitada por el usuario, con
validación en su empresa PCS. Se conservaron cambios concurrentes. El árbol pasó
por `05d23561` mediante otro proceso `rs` durante el trabajo; los resultados locales
corresponden al árbol observado, no a un release inmutable aprobado en esta tarea.
Posteriormente otro `rs` integró también cambios de seguridad en `13995ec0`, seguido
por `fc3dd9b4`. Esa actividad ajena no acredita despliegue; en la última comprobación
los contenedores productivos seguían con más de nueve horas de ejecución.
Los cambios de Nginx/SSH se aplicaron directamente al host con respaldo y validación;
los cambios Go requieren verificar su publicación mediante el flujo de release.

Cobertura: SSH efectivo, UFW, bindings y privilegios Docker, TLS, políticas Nginx,
origen de IP, abuso del chat público, autenticación observada, disponibilidad,
roles PostgreSQL agregados, actualización automática y análisis de imágenes.
No se modificaron ventas, cobros, documentos fiscales, credenciales ni datos de
empresas. No se ejecutó una restauración, migración de servicios auxiliares, reinicio
del VPS ni prueba de denegación de servicio.

## Reparaciones y pruebas

| Hallazgo | Corrección | Evidencia y estado |
| --- | --- | --- |
| El borde HTTP carecía de límites de frecuencia/conexiones | Nginx: 60 solicitudes/s por IP, ráfaga 240; acceso/recuperación 1/s con ráfaga 20; 100 conexiones activas por IP; rechazo 429 | **Aplicado VPS**. 36 GET sin credenciales al login, con cabecera IP variable: 22 respuestas 405 y 14 respuestas 429. No se intentaron contraseñas ni se llamó a proveedores |
| Cabeceras IP del cliente podían falsear auditoría y contadores | El borde reemplaza X-Forwarded-For por la IP del socket y X-Forwarded-Host por el host aceptado | **Aplicado VPS** en sitios habilitados; fuentes edge y generadores Mailu/staging corregidos localmente |
| Helpers Go confiaban en el primer elemento X-Forwarded-For o sin validar proxy | `utils.ClientIP`: recorre de derecha a izquierda hasta el primer salto no confiable; valida IP, rechaza cadena inválida, solo confía en CIDRs explícitos | **Validado localmente** en logging, chat, mensajes públicos, certificados, auditoría y túnel Domótica. Sin cambios de tenant, permisos ni tablas |
| Una cookie nueva reiniciaba el límite del chat público | Presupuesto adicional de 30 solicitudes/5 minutos por IP, compartido entre tiendas y variantes del chat | **Validado localmente**: rotación de cookie y tienda no reinicia el límite. El límite por conversación sigue siendo 10/5 minutos |
| El mapa de contadores públicos crecía sin límite | Máximo 16.384 entradas, recuperación de expiradas una vez por minuto y rechazo conservador al saturarse | **Validado localmente**. No se expulsan límites activos para admitir un atacante nuevo. El presupuesto es por proceso; para varias réplicas requiere contador compartido o control equivalente en el borde |
| El archivo SSH de endurecimiento existía pero no se aplicaba | Política efectiva al principio de sshd_config: X11 y agente desactivados; MaxAuthTries 3, LoginGraceTime 30, MaxStartups 10:30:60 | **Aplicado VPS**; `sshd -t`, `sshd -T`, recarga y segunda conexión SSH correctos. Puerto y acceso por llave preservados |
| Faltaba HSTS en el dominio principal y se publicaba versión Nginx | HSTS de un año para HTTPS, versión oculta, tiempos de cabecera/cuerpo/envío y política CSP básica adicional | **Aplicado VPS**. Login 200 con HSTS. Se conserva la CSP completa del backend; la política adicional restringe base, objetos y ancestros |
| El proxy Webmail eliminaba la CSP del proveedor | Se conserva su CSP y se añade por separado la restricción de iframe PCS | **Aplicado VPS** y generador corregido. Respuesta Webmail 302 vuelve a conservar default-src/script-src/connect-src del proveedor. Redirección anónima posterior devuelve 403; no acredita login al buzón |
| Venta digital usaba el wildcard vencido | Certificado propio `pcs-venta-digital`, emitido mediante Nginx/ACME y asociado solo a ese virtual host | **Aplicado VPS**. HTTPS válido y respuesta 302; vence 2026-12-05. Renovación configurada con plugin nginx; no se ha simulado todavía un ciclo de renovación completo |
| Auditoría host podía aprobar UFW inactivo y confundir un archivo con configuración SSH efectiva | Comprueba contenido de `ufw status`, `sshd -T`, Fail2ban activo, vencimiento TLS y reinicio; modo `--strict` | **Corregido localmente**. El verificador Node declara expresamente alcance documental y `runtime_verified=false` |
| Un test de métricas rechazaba la palabra tenant en HELP | Se inspeccionan muestras, no comentarios, para detectar etiquetas de tenant/empresa | **Validado localmente**; no se cambió el dato publicado ni la regla de confidencialidad |

Implementación: [hardening HTTP](../../deploy/scripts/vps-http-hardening.py),
[hardening SSH](../../deploy/scripts/vps-ssh-hardening.sh),
[auditor host](../../deploy/scripts/vps-hardening-audit.sh),
[resolución IP](../../backend/utils/utils.go),
[chat público](../../backend/handlers/public_chat_portal.go) y
[regresiones](../../backend/handlers/client_ip_security_test.go).

## Estado observado del VPS

- Ubuntu 24.04.4; kernel activo `6.8.0-111-generic`, más de 116 días sin reinicio.
  El indicador de reinicio incluye libc y kernels posteriores instalados.
- UFW activo: deniega entradas y tráfico enrutado por defecto; SSH por llave y
  Fail2ban SSH activo. VNC tiene una excepción de IP operativa, no acceso general.
- Prueba TCP desde este equipo: accesibles 25, 80, 443, 49222 y 21117;
  inaccesibles 22, 5432, 6379, 8081, 8082, 9090, 9093, 3001 y 5901.
  Se comprobaron esos puertos TCP; no es un barrido completo TCP/UDP/IPv6.
- Backend, bases, Mailu HTTP/IMAP/submission, Nextcloud, OnlyOffice, Grafana y
  Prometheus publicados al host usan loopback o red Docker. RustDesk publica
  sus puertos operativos. DOCKER-USER no tiene reglas adicionales.
- API y worker se ejecutan como usuario PCS, filesystem de solo lectura y
  no-new-privileges. cAdvisor permanece privilegiado; el gestor de disco tiene
  autoridad sobre Docker. Su revisión y reducción de privilegios siguen pendientes.
- Conexiones PostgreSQL observadas incluyen siete conexiones con rol runtime sin
  superusuario/CREATEDB/CREATEROLE/BYPASSRLS y la conexión de auditoría privilegiada.
  Esto no prueba todas las ACL de tablas ni todos los pools de cada servicio.
- `unattended-upgrade --dry-run --debug` no encontró paquetes elegibles pendientes
  en su catálogo actual. No se hizo `apt update` ni se certifica que el índice esté
  actualizado. `reboot-required` sigue presente.
- `/health`, `/ready` y login: 200 después de las recargas. `/metrics`: 404;
  endpoint empresarial y API de seguridad super sin sesión: 401. `.env` y
  `.git/config`: 401 y no exposición de contenido.
- Login real autorizado llegó al panel Super con menús e iframe funcionando.
  **La sesión entró sin TOTP**: las correcciones locales de MFA no acreditan que
  esa protección esté activa en el backend publicado.

## Imágenes: riesgo pendiente

Trivy analizó vulnerabilidades HIGH/CRITICAL, sin ejecutar un escáner de secretos.
Los resultados se cruzaron con el **image ID exacto de los contenedores activos**;
un tag mutable no se aceptó como identidad por sí solo. Los números cuentan
hallazgos de paquetes, pueden repetir CVE entre paquetes e imágenes y no demuestran
explotabilidad de cada servicio. No son una suma de ataques confirmados.

| Servicio / imagen activa | Críticas | Altas | Acción pendiente |
| --- | ---: | ---: | --- |
| Backend, worker, frontend y gestor de disco productivos | 0 | 0 | Verificar próximo artefacto; cero hallazgos no prueba ausencia de defectos de lógica |
| Nextcloud | 93 | 1296 | Upgrade por versiones mayores consecutivas, apps compatibles y restauración aislada |
| OnlyOffice | 25 | 424 | Imagen corregida, compatibilidad JWT/editor/guardado y rollback |
| Grafana | 10 | 121 | Backup de configuración/datos, actualización y comprobación de paneles/alertas |
| cAdvisor | 6 | 60 | Imagen corregida y revisión especial de privilegios host |
| Prometheus | 6 | 94 | Upgrade validado con reglas, retención y recuperación |
| Alertmanager | 6 | 90 | Upgrade validado; comprobar rutas sin enviar alertas de prueba no autorizadas |
| Node exporter | 2 | 43 | Actualizar imagen y validar métricas |
| PostgreSQL base de contenedor | 1 | 30 | Diferenciar paquetes OS del motor; imagen corregida y restauración antes de recrear |
| Mailu: admin/front/IMAP/SMTP/antispam/resolver, cada imagen | 0 | 16 | Actualización coordinada compatible de la familia Mailu |
| Webmail | 0 | 18 | Actualización compatible y prueba real de buzón |
| Redis de Nextcloud | 4 | 52 | Actualización compatible; Redis de Mailu no tuvo HIGH/CRITICAL |
| Motor heredado de Nextcloud | 1 | 25 | Dependencia preexistente fuera de PostgreSQL; planificar migración conforme a reglas PCS, sin convertir datos en esta revisión |
| Backend staging | 2 | 0 | Reemplazar por candidato corregido y validado |
| Frontend staging | 0 | 22 | Reconstruir/actualizar candidato |
| ClamAV staging | 4 | 86 | Actualizar y repetir prueba de escaneo obligatorio |
| Worker staging | — | — | Trivy falló tanto con referencia corta como con image ID completo; cobertura incompleta |
| RustDesk y Redis Mailu | 0 | 0 | Conservar control de acceso/red; no acredita prueba funcional del proveedor |

Detalle minimizado: [inventario de imágenes](revision_vps_2026-09-06_imagenes.json).
Los resultados completos y logs permanecen en almacenamiento privado del host,
sin credenciales ni datos empresariales copiados al informe.

## Condiciones necesarias para cerrar

1. **DNS wildcard:** el certificado `*.powerfulcontrolsystem.com` venció el
   2026-07-15. Su método manual carece de hook DNS y falla en Certbot. Venta digital
   ya tiene certificado independiente; los slugs empresariales siguen requiriendo
   validación DNS en Hostinger. Se solicitó acceso a una sesión del panel, sin pedir
   secretos por chat. No se eliminó el certificado ni se amplió otro certificado
   suponiendo que cubría el wildcard.
2. **MFA:** enrolamiento y recuperación del segundo factor por el titular,
   publicación de la política obligatoria y prueba negativa sin OTP. Activarlo
   cambiando una bandera sin enrolamiento puede bloquear al operador.
3. **Servicios vulnerables:** preparar imágenes por digest, backups consistentes,
   ensayo de restore y actualización compatible en aislamiento. Nextcloud requiere
   pasos entre versiones mayores; sustituirlo directamente por latest puede romper
   esquema, apps y documentos. No se ejecutaron esas migraciones en esta tarea.
4. **Host:** acordar ventana de interrupción, verificar backup y consola alternativa,
   reiniciar para activar kernel/libc actualizados y comprobar todos los servicios.
5. **Aplicación:** publicar el candidato revisado por el flujo Git/CI/staging/rs,
   probar autorización A/B y roles de todos los módulos críticos. Los nuevos cambios
   Go no se inyectaron en contenedores productivos.
6. **DDoS y continuidad:** revisar protección del proveedor/CDN, origen protegido y
   alertas; límites de Nginx no absorben saturación del enlace. Faltan ensayo de
   restauración, carga autorizada y verificación IPv6/UDP completa.

## Validación y recuperación

Pruebas locales aprobadas en el árbol observado: `go test -p 1 ./...`,
`go vet -p 1 ./...`, `go mod verify`, pruebas específicas de IP/cookies/capacidad
del limitador y métricas. `govulncheck ./...`: cero vulnerabilidades alcanzables;
una vulnerabilidad de módulo requerido no utilizada según el análisis.
El primer intento de govulncheck agotó memoria; se repitió con GOMAXPROCS=2 y
GOMEMLIMIT. Hubo compilaciones intermedias fallidas mientras otro proceso cambiaba
el árbol; no se usaron como evidencia final de aprobación.

El auditor host corregido se ejecutó por SSH: comprobó los controles efectivos y
devolvió `AUDIT_EXIT=2` por dos advertencias reales, wildcard vencido y reinicio
pendiente. Se validó sintaxis Bash de ambos scripts. Las seis pruebas del validador
documental pasaron; el catálogo se regeneró después de registrar esta evidencia.

Nginx y SSH tuvieron validación antes y después y recarga sin reinicio del host.
Respaldos privados bajo `/var/backups/pcs-security/`: cambios HTTP inicial/final,
configuración SSH y virtual host del certificado de venta digital. Para revertir,
identificar el respaldo correspondiente, restaurar únicamente los archivos del
cambio, validar con `nginx -t` o `sshd -t` y recargar. Mantener una conexión SSH
abierta hasta comprobar otra conexión. No mezclar respaldos de diferentes cambios.

## Fuentes técnicas

- [Nginx limit_req](https://nginx.org/en/docs/http/ngx_http_limit_req_module.html):
  límites, ráfagas y código de rechazo.
- [Docker y firewall](https://docs.docker.com/engine/network/packet-filtering-firewalls/):
  publicar puertos Docker puede eludir reglas UFW; se verificaron bindings reales.
- [Actualización Nextcloud](https://docs.nextcloud.com/server/stable/admin_manual/maintenance/upgrade.html):
  progresión por versiones y preparación de mantenimiento.
- [Seguridad Docker](https://docs.docker.com/engine/security/): límites de privilegios
  y autoridad del daemon/socket.

Requisitos aplicables: PCS-REQ-001, PCS-REQ-002, PCS-REQ-016.
