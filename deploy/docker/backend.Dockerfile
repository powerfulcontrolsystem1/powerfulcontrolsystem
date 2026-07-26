FROM golang:1.25.12-alpine AS build

RUN apk add --no-cache ca-certificates git
WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-backend .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-migrate ./cmd/pcs-migrate
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-worker ./cmd/pcs-worker
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-migrate-private-uploads ./tools/migrate_private_uploads

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

COPY --chown=pcs:pcs --from=build /out/pcs-backend /app/backend/pcs-backend
COPY --chown=pcs:pcs --from=build /out/pcs-migrate /app/backend/pcs-migrate
COPY --chown=pcs:pcs web /app/web
RUN mkdir -p /app/backend/logs/vps_security/tmp /app/backend/logs/vps_security/trivy-cache /app/private_storage \
    && chown pcs:pcs /app/backend/logs/vps_security/tmp /app/backend/logs/vps_security/trivy-cache /app/private_storage
USER pcs:pcs
CMD ["/bin/sh", "-ec", "/app/backend/pcs-backend && /app/backend/pcs-migrate"]

FROM runtime-base AS worker

USER root
RUN apk add --no-cache postgresql-client
COPY --chown=pcs:pcs --from=build /out/pcs-worker /app/backend/pcs-worker
COPY --chown=pcs:pcs web /app/web
COPY --chown=pcs:pcs documentos /app/documentos
COPY --chown=pcs:pcs backend /app/project_export/backend
COPY --chown=pcs:pcs web /app/project_export/web
COPY --chown=pcs:pcs deploy /app/project_export/deploy
COPY --chown=pcs:pcs scripts /app/project_export/scripts
COPY --chown=pcs:pcs documentos /app/project_export/documentos
COPY --chown=pcs:pcs .dockerignore AGENTS.md CHANGELOG.md /app/project_export/
ENV PCS_PROJECT_EXPORT_ROOT=/app/project_export
RUN mkdir -p /app/backend/logs/vps_security/tmp /app/backend/logs/vps_security/trivy-cache /app/private_storage /app/backup /app/web/uploads \
    && chown pcs:pcs /app/backend/logs/vps_security/tmp /app/backend/logs/vps_security/trivy-cache /app/private_storage /app/backup /app/web/uploads
USER pcs:pcs
CMD ["/app/backend/pcs-worker"]

FROM runtime-base AS api

COPY --chown=pcs:pcs --from=build /out/pcs-backend /app/backend/pcs-backend
COPY --chown=pcs:pcs --from=build /out/pcs-migrate-private-uploads /app/backend/pcs-migrate-private-uploads
COPY --chown=pcs:pcs web /app/web
COPY --chown=pcs:pcs documentos /app/documentos
COPY --chown=pcs:pcs backend /app/project_export/backend
COPY --chown=pcs:pcs web /app/project_export/web
COPY --chown=pcs:pcs deploy /app/project_export/deploy
COPY --chown=pcs:pcs scripts /app/project_export/scripts
COPY --chown=pcs:pcs documentos /app/project_export/documentos
COPY --chown=pcs:pcs .dockerignore AGENTS.md CHANGELOG.md /app/project_export/

ENV PCS_PROJECT_EXPORT_ROOT=/app/project_export

RUN mkdir -p /app/backend/logs/vps_security/tmp /app/backend/logs/vps_security/trivy-cache /app/web/uploads /app/private_storage /app/backup /app/descargas \
    && chmod +x /app/project_export/deploy/scripts/vps-provision-mailu-mailbox.sh \
    && chmod +x /app/project_export/deploy/scripts/vps-delete-mailu-mailbox.sh \
    && chown pcs:pcs /app/backend/logs/vps_security/tmp /app/backend/logs/vps_security/trivy-cache /app/web/uploads /app/private_storage /app/backup /app/descargas

EXPOSE 8080
USER pcs:pcs
CMD ["./pcs-backend"]
