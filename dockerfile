# Dockerfile для контейнера с Telegram-ботом
# Используем официальный образ Go в качестве базового образа
FROM golang:1.17

# Установите аргументы для переменных окружения из файла .env
ARG TELEGRAM_TOKEN
ARG POSTGRES_USER
ARG POSTGRES_PASSWORD
ARG POSTGRES_DB

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем исходный код на Go в контейнер
COPY . .

# Собираем Go-приложение
RUN go build -o telegram-bot

# Открываем порт, на котором будет работать бот
EXPOSE 8080

# Запускаем бота
CMD ["./telegram-bot"]

# Dockerfile для контейнера с базой данных (пример: PostgreSQL)
# Используем официальный образ PostgreSQL в качестве базового образа
FROM postgres:13

# Устанавливаем переменные окружения для базы данных
ENV POSTGRES_USER myuser
ENV POSTGRES_PASSWORD mypassword
ENV POSTGRES_DB mydb

# Открываем стандартный порт PostgreSQL
EXPOSE 5432

# Запускаем сервер PostgreSQL
CMD ["postgres"]