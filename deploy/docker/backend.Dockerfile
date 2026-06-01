FROM golang:1.24-alpine AS build

RUN apk add --no-cache ca-certificates git
WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-backend .

FROM alpine:3.20

RUN apk add --no-cache bash ca-certificates curl docker-cli openssh-client tesseract-ocr tesseract-ocr-data-eng tesseract-ocr-data-spa tzdata
WORKDIR /app/backend

COPY --from=build /out/pcs-backend /app/backend/pcs-backend
COPY web /app/web
COPY documentos /app/documentos
COPY descargas /app/descargas
COPY backend /app/project_export/backend
COPY web /app/project_export/web
COPY deploy /app/project_export/deploy
COPY scripts /app/project_export/scripts
COPY documentos /app/project_export/documentos
COPY .dockerignore AGENTS.md CHANGELOG.md /app/project_export/

ENV PCS_PROJECT_EXPORT_ROOT=/app/project_export
ENV GRAFOLOGIA_TESSERACT_ENABLED=1
ENV GRAFOLOGIA_TESSERACT_BIN=tesseract
ENV GRAFOLOGIA_TESSERACT_LANG=spa+eng

RUN mkdir -p /app/backend/logs /app/web/uploads /app/backup /app/descargas
RUN chmod +x /app/project_export/deploy/scripts/vps-provision-mailu-mailbox.sh

EXPOSE 8080
CMD ["./pcs-backend"]
