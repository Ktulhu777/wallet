# Используем официальный образ Golang
FROM golang:1.22.2-alpine

# Устанавливаем необходимые зависимости
RUN apk add --no-cache git

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./

# Устанавливаем зависимости
RUN go mod download

# Копируем весь исходный код в контейнер
COPY . .

# # Указываем порт, на котором работает приложение
EXPOSE 7777

# Команда для запуска миграций и сервера
CMD ["sh", "-c", "go run cmd/migrator/main.go -migrations-path=./migrations -db-user=${DB_USER} -db-password=${DB_PASS} -db-host=${DB_HOST} -db-port=${DB_PORT} -db-name=${DB_NAME} -sslmode=${DB_SSL_MODE} && go run cmd/wallet/main.go"]