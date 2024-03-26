# Используем официальный образ Golang как базовый
FROM golang:latest as builder

# Устанавливаем рабочую директорию в контейнере
WORKDIR /app

# Копируем исходный код в контейнер
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

# Открываем порт, который будет использоваться ботом
EXPOSE 8080

# Запускаем бинарный файл
CMD ["./code"]

# Используем официальный образ PostgreSQL
FROM postgres:latest

# Устанавливаем переменные окружения для базы данных
ENV POSTGRES_DB=dbname
ENV POSTGRES_USER=user
ENV POSTGRES_PASSWORD=password

# Открываем порт для подключения к базе данных
EXPOSE 5432