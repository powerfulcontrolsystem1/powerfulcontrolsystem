FROM nginx:1.27-alpine

COPY web /usr/share/nginx/html
COPY deploy/nginx/pcs.conf /etc/nginx/conf.d/default.conf
