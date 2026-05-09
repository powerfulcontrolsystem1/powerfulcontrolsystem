FROM golang:1.24-alpine AS build

RUN apk add --no-cache ca-certificates git
WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pcs-backend .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates curl openssh-client tzdata
WORKDIR /app/backend

COPY --from=build /out/pcs-backend /app/backend/pcs-backend
COPY web /app/web
COPY documentos /app/documentos
COPY descargas /app/descargas

RUN mkdir -p /app/backend/logs /app/web/uploads /app/backup /app/descargas

EXPOSE 8080
CMD ["./pcs-backend"]
