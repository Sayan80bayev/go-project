# 1. Build stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 🔧 Статическая сборка для Alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth_service ./cmd/server/main.go

# ------------------------------------

# 2. Final stage (minimal Alpine)
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/auth_service .

COPY ./certs/localhost.pem /etc/keycloak/certs/localhost.pem
COPY ./certs/localhost-key.pem /etc/keycloak/certs/localhost-key.pem
COPY --from=builder /app/config ./config

ENV KEYCLOAK_CERT_PATH=/etc/keycloak/certs/localhost.pem

EXPOSE 8082

CMD ["./auth_service"]