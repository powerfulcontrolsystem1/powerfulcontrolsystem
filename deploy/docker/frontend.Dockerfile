FROM nginxinc/nginx-unprivileged:1.30.3-alpine3.23

USER root
# La imagen base puede conservar paquetes de la capa publicada anteriormente.
# Actualizar explícitamente las librerías usadas por el healthcheck evita que
# una compilación reproducible arrastre un curl vulnerable desde la caché base.
RUN apk add --no-cache --upgrade c-ares=1.34.8-r0 curl libcurl

COPY web /usr/share/nginx/html
COPY deploy/nginx/pcs.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080
USER 101:101
