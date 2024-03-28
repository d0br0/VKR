# Используйте официальный образ Golang как базовый
FROM alpine:latest

# Установите аргументы для переменных окружения из файла .env
ARG TELEGRAM_TOKEN
ARG POSTGRES_USER
ARG POSTGRES_PASSWORD
ARG POSTGRES_DB

# Скопируйте исходный код в контейнер
COPY . .

# Скомпилируйте приложение для продакшена
RUN apk add --no-cache ca-certificates &&\
    chmod +x code

EXPOSE 80/tcp

# Запустите скомпилированный бинарный файл
CMD [ "./code" ]
