FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Собираем
RUN CGO_ENABLED=0 GOOS=linux go build -o shortener ./cmd/main.go

FROM alpine:latest

COPY --from=builder /app/shortener /usr/local/bin/shortener

EXPOSE 8080

CMD ["shortener"]