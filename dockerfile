# Используйте официальный образ Golang как базовый
FROM golang:latest as builder

# Установите аргументы для переменных окружения из файла .env
ARG TELEGRAM_TOKEN
ARG POSTGRES_USER
ARG POSTGRES_PASSWORD
ARG POSTGRES_DB

# Устанавливаем рабочую директорию в контейнере
WORKDIR /app

# Скопируйте исходный код в контейнер
COPY . .

# Скачиваем зависимости
RUN go mod download

# Собираем бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Используем образ alpine для финального контейнера из-за его малого размера
FROM alpine:latest 

# Устанавливаем рабочую директорию в контейнере
WORKDIR /root/

# Копируем бинарный файл из предыдущего шага
COPY --from=builder /app/main .

# Скомпилируйте приложение для продакшена
RUN apk add --no-cache ca-certificates &&\
    chmod +x main

EXPOSE 80/tcp

# Запустите скомпилированный бинарный файл
CMD [ "./main" ]
