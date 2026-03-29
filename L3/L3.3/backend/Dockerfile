FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Собираем
RUN CGO_ENABLED=0 GOOS=linux go build -o comment ./cmd/main.go

FROM alpine:latest

COPY --from=builder /app/comment /usr/local/bin/comment

EXPOSE 8080

CMD ["comment"]