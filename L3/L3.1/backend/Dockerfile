FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

WORKDIR /app/L3/L3.1/backend

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/delayed-notifier ./cmd/main.go

FROM alpine:latest

COPY --from=builder /app/bin/delayed-notifier /usr/local/bin/delayed-notifier

COPY --from=builder /app/L3/L3.1/frontend /web

EXPOSE 8080

CMD ["delayed-notifier"]