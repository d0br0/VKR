# Используйте официальный образ Golang как базовый
FROM alpine

# Установите аргументы для переменных окружения из файла .env
ARG TELEGRAM_TOKEN
ARG POSTGRES_USER
ARG POSTGRES_PASSWORD
ARG POSTGRES_DB

ENV LANGUAGE="en"

WORKDIR /app

# Скопируйте исходный код в контейнер
COPY /code/code .

# Скомпилируйте приложение для продакшена
RUN apk add --no-cache ca-certificates &&\
    chmod +x /app/code

EXPOSE 80/tcp

# Запустите скомпилированный бинарный файл
CMD [ "/app/code" ]
