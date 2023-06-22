# Используйте базовый образ с поддержкой Go
FROM golang:1.19 AS builder

# Установка рабочей директории
WORKDIR /app

# Копирование зависимостей и файлов проекта
COPY go.mod go.sum ./
COPY . .

# Сборка исполняемого файла
RUN go build -o app

# Финальный образ
FROM debian:buster-slim

# Установка зависимостей для работы с Kafka и других необходимых пакетов
RUN apt-get update && apt-get install -y ca-certificates curl

# Копирование исполняемого файла из предыдущего образа
COPY --from=builder /app/app /app/app

# Установка переменных окружения
ENV MY_ENV_VAR=value

# Запуск исполняемого файла при старте контейнера
CMD ["/app/app"]
