FROM golang:1.25.12-alpine AS build

RUN apk add --no-cache ca-certificates git
WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-backend .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-migrate ./cmd/pcs-migrate
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-worker ./cmd/pcs-worker

FROM alpine:3.20

RUN apk add --no-cache bash ca-certificates curl openssh-client openssl tzdata \
    && addgroup -S -g 10001 pcs \
    && adduser -S -D -H -u 10001 -G pcs pcs
WORKDIR /app/backend

COPY --from=build /out/pcs-backend /app/backend/pcs-backend
COPY --from=build /out/pcs-migrate /app/backend/pcs-migrate
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
ENV GRAFOLOGIA_TESSERACT_ENABLED=0

RUN mkdir -p /app/backend/logs /app/web/uploads /app/backup /app/descargas \
    && chmod +x /app/project_export/deploy/scripts/vps-provision-mailu-mailbox.sh \
    && chmod +x /app/project_export/deploy/scripts/vps-delete-mailu-mailbox.sh \
    && chown -R pcs:pcs /app

EXPOSE 8080
USER pcs:pcs
CMD ["./pcs-backend"]
