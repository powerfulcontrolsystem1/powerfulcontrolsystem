FROM nginxinc/nginx-unprivileged:1.30.3-alpine3.23

COPY web /usr/share/nginx/html
COPY deploy/nginx/pcs.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080
USER 101:101
