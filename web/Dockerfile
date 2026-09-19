# ---------- стадия 1: сборка ----------
FROM golang:1.25-alpine AS builder
WORKDIR /app
# Сначала только go.mod/go.sum — слой с зависимостями кэшируется
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/web-server ./cmd/server

# ---------- стадия 2: минимальный образ ----------
FROM alpine:3.22
RUN apk --no-cache add ca-certificates \
    && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder /out/web-server .
USER app
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=3 CMD wget -qO- http://127.0.0.1:8080/health || exit 1
CMD ["./web-server"]
