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
RUN go build -o main .
# Запустите скомпилированный бинарный файл
CMD [ "/app/main" ]
