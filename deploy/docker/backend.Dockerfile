FROM golang:1.25.12-alpine AS build

RUN apk add --no-cache ca-certificates git
WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-backend .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-migrate ./cmd/pcs-migrate
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-worker ./cmd/pcs-worker

FROM alpine:3.20 AS lynis-source

RUN apk add --no-cache git \
    && git clone --depth 1 https://github.com/CISOfy/lynis.git /src/lynis \
    && cd /src/lynis \
    && git fetch --depth 1 origin 06153321ea50d53a27446084e646d9f43fe46e0e \
    && git checkout --detach 06153321ea50d53a27446084e646d9f43fe46e0e

FROM alpine:3.20 AS runtime-base

RUN apk add --no-cache bash ca-certificates curl nmap nmap-scripts openssh-client openssl tzdata \
    && addgroup -S -g 10001 pcs \
    && adduser -S -D -H -u 10001 -G pcs pcs
COPY --from=lynis-source /src/lynis /opt/lynis
RUN printf '%s\n' '#!/bin/sh' 'cd /opt/lynis' 'exec /opt/lynis/lynis "$@"' > /usr/local/bin/lynis \
    && chmod 0755 /usr/local/bin/lynis
WORKDIR /app/backend
ENV GRAFOLOGIA_TESSERACT_ENABLED=0

FROM runtime-base AS migrate

COPY --from=build /out/pcs-backend /app/backend/pcs-backend
COPY --from=build /out/pcs-migrate /app/backend/pcs-migrate
COPY web /app/web
RUN mkdir -p /app/backend/logs/vps_security/tmp /app/backend/logs/vps_security/trivy-cache /app/private_storage \
    && chown -R pcs:pcs /app
USER pcs:pcs
CMD ["/bin/sh", "-ec", "/app/backend/pcs-backend && /app/backend/pcs-migrate"]

FROM runtime-base AS worker

USER root
RUN apk add --no-cache postgresql-client
COPY --from=build /out/pcs-worker /app/backend/pcs-worker
COPY web /app/web
COPY documentos /app/documentos
COPY backend /app/project_export/backend
COPY web /app/project_export/web
COPY deploy /app/project_export/deploy
COPY scripts /app/project_export/scripts
COPY documentos /app/project_export/documentos
COPY .dockerignore AGENTS.md CHANGELOG.md /app/project_export/
ENV PCS_PROJECT_EXPORT_ROOT=/app/project_export
RUN mkdir -p /app/backend/logs/vps_security/tmp /app/backend/logs/vps_security/trivy-cache /app/private_storage /app/backup /app/web/uploads \
    && chown -R pcs:pcs /app
USER pcs:pcs
CMD ["/app/backend/pcs-worker"]

FROM runtime-base AS api

COPY --from=build /out/pcs-backend /app/backend/pcs-backend
COPY web /app/web
COPY documentos /app/documentos
COPY backend /app/project_export/backend
COPY web /app/project_export/web
COPY deploy /app/project_export/deploy
COPY scripts /app/project_export/scripts
COPY documentos /app/project_export/documentos
COPY .dockerignore AGENTS.md CHANGELOG.md /app/project_export/

ENV PCS_PROJECT_EXPORT_ROOT=/app/project_export

RUN mkdir -p /app/backend/logs/vps_security/tmp /app/backend/logs/vps_security/trivy-cache /app/web/uploads /app/private_storage /app/backup /app/descargas \
    && chmod +x /app/project_export/deploy/scripts/vps-provision-mailu-mailbox.sh \
    && chmod +x /app/project_export/deploy/scripts/vps-delete-mailu-mailbox.sh \
    && chown -R pcs:pcs /app

EXPOSE 8080
USER pcs:pcs
CMD ["./pcs-backend"]
