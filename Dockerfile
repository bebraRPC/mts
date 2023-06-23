# Используйте базовый образ с поддержкой Go
FROM golang:1.19 AS builder

# Установка рабочей директории
WORKDIR /app

# Копирование зависимостей и файлов проекта
COPY go.mod go.sum ./
COPY . .

# Сборка исполняемого файла для gate
RUN cd cmd/gate && go build -o gate

# Сборка исполняемого файла для worker
RUN cd cmd/worker && go build -o worker

# Финальный образ
FROM debian:buster-slim

# Установка зависимостей для работы с Kafka и других необходимых пакетов
RUN apt-get update && apt-get install -y ca-certificates curl

# Копирование исполняемых файлов из предыдущего образа
COPY --from=builder /app/cmd/gate/gate /app/gate
COPY --from=builder /app/cmd/worker/worker /app/worker

# Копирование конфигурационных файлов
COPY cmd/gate/env /app/env
COPY cmd/worker/env /app/env

# Установка разрешений на выполнение
RUN chmod +x /app/gate
RUN chmod +x /app/worker

# Установка переменных окружения
ENV MY_ENV_VAR=value
ENV KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181

# Запуск исполняемых файлов при старте контейнера
CMD ["./app"]
