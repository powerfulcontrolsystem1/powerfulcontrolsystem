FROM nginxinc/nginx-unprivileged:1.30.3-alpine3.23

USER root
# La imagen base puede conservar paquetes de la capa publicada anteriormente.
# Actualizar explícitamente las librerías usadas por el healthcheck y el parser
# XML evita arrastrar versiones vulnerables desde la caché de la imagen base.
RUN apk add --no-cache --upgrade c-ares=1.34.8-r0 curl libcurl 'libexpat>=2.8.4-r0' 'libuuid>=2.41.6-r0'

COPY web /usr/share/nginx/html
COPY deploy/nginx/pcs.conf /etc/nginx/conf.d/default.conf
COPY deploy/nginx/pcs-static-security-headers.inc /etc/nginx/conf.d/pcs-static-security-headers.inc

EXPOSE 8080
USER 101:101
