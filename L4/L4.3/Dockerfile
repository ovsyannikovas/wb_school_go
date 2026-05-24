FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o calendar-service ./cmd/server/

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/calendar-service .

RUN chmod +x calendar-service

EXPOSE 8080

CMD ["./calendar-service"]
