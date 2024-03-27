# Используйте официальный образ Golang как базовый
FROM golang:latest

# Установите аргументы для переменных окружения из файла .env
ARG TELEGRAM_TOKEN
ARG POSTGRES_USER
ARG POSTGRES_PASSWORD
ARG POSTGRES_DB

# Установите рабочий каталог в контейнере
WORKDIR /app

# Скопируйте модульные файлы и загрузите зависимости
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Скопируйте исходный код в контейнер
COPY . .

# Скомпилируйте приложение для продакшена
RUN apk add --no-cache ca-certificates &&\
    chmod +x code

EXPOSE 80/tcp
# Запустите скомпилированный бинарный файл
CMD [ "./code" ]
