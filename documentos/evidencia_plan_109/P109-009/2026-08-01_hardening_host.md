# P109-009 - Escaneo y hardening del host VPS

Fecha: 2026-08-01  
Entorno: VPS principal autorizado

El escaneo rapido reproducible se limito a loopback y a los chequeos locales.
La linea base termino con ocho hallazgos: dos altos, tres medios, uno bajo y dos
informativos. Los dos altos eran UFW inactivo y SSH con acceso root/password
efectivamente habilitado.

Se aplicaron cambios reversibles y se comprobaron inmediatamente:

- SSH conserva root solo mediante llave (`without-password`), deshabilita
  password y autenticacion interactiva, valida `sshd -t`, recarga y acepta una
  conexion nueva por llave;
- UFW queda activo con entrada denegada por defecto y aperturas explicitas para
  SSH PCS, HTTP/HTTPS, Mailu SMTP y RustDesk; VNC queda restringido al origen
  operativo observado;
- Avahi y CUPS, sin uso en la operacion web del VPS, quedan inactivos y
  deshabilitados;
- el scanner consulta UFW en modo verbose e incluye sitios Nginx habilitados
  sin extension para no reportar falsos positivos.

La tercera ejecucion (`scan-20260801T074326Z-8544407e`) termino con cobertura
completa, cero criticos, cero altos, un medio y dos informativos. El unico medio
restante corresponde a 30 paquetes actualizables. El host ya exige reinicio;
la actualizacion y el reboot requieren ventana de mantenimiento para no cortar
los servicios activos.

Produccion y staging respondieron 200 despues del hardening. Este escaneo de
host no sustituye el DAST web autenticado ni cierra CSP/A-B.

Estado: **P109-009 parcial**.
