FROM golang:1.26-alpine AS builder

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# ✅ МИНИМАЛЬНЫЕ переменные - БЕЗ ПРОБЛЕМ
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o server ./cmd/server

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl tzdata

WORKDIR /app

COPY --from=builder /app/server .

# Минимальная копия (добавьте только то, что нужно)
COPY --from=builder /app/migrations ./migrations 2>/dev/null || true
COPY --from=builder /app/api ./api 2>/dev/null || true
COPY --from=builder /app/config ./config 2>/dev/null || true

RUN addgroup -g 1001 appgroup && \
    adduser -S -u 1001 appuser appgroup && \
    chown -R appuser:appgroup /app

USER appuser
EXPOSE 8080
CMD ["./server"]