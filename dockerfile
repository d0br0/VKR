# Используйте официальный образ Go как сборщик
FROM golang:latest as builder

# Установите аргументы для переменных окружения из файла .env
ARG TELEGRAM_TOKEN
ARG POSTGRES_USER
ARG POSTGRES_PASSWORD
ARG POSTGRES_DB

# Установите рабочую директорию внутри контейнера
WORKDIR /app

# Скопируйте исходный код в контейнер
COPY . .

# Загрузите зависимости
RUN go mod download

# Соберите бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Используем образ alpine для финального контейнера из-за его малого размера
FROM alpine:latest

# Установите рабочую директорию внутри контейнера
WORKDIR /root/

# Скопируйте бинарный файл из предыдущего этапа
COPY --from=builder /app/main .

# Установите ca-certificates (для поддержки HTTPS) и сделайте бинарный файл исполняемым
RUN apk add --no-cache ca-certificates && chmod +x main

EXPOSE 80/tcp

# Запустите скомпилированный бинарный файл
CMD ["./main"]