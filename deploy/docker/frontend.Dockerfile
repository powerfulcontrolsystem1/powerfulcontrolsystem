FROM nginxinc/nginx-unprivileged:1.30.3-alpine3.23

USER root
RUN apk add --no-cache --upgrade c-ares=1.34.8-r0

COPY web /usr/share/nginx/html
COPY deploy/nginx/pcs.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080
USER 101:101
