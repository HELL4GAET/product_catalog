# === BUILD STAGE ===
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git

WORKDIR /app

# 1) Копируем конфиг и .env
COPY internal/config/config.yaml internal/config/config.yaml
COPY .env .env

# 2) Копируем документацию
COPY docs ./docs

# 3) Скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# 4) Копируем остальное
COPY . .

# 5) Сборка
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" \
    -o subscription-service \
    ./cmd/main.go

# === RUN STAGE ===
FROM alpine:3.18
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Копируем бинарник и конфиги
COPY --from=builder /app/subscription-service .
COPY --from=builder /app/internal/config/config.yaml internal/config/config.yaml
COPY --from=builder /app/.env .env

# Копируем документацию
COPY --from=builder /app/docs ./docs

EXPOSE 1488

CMD ["./subscription-service"]