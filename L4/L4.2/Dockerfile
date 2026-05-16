FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o distributor ./main.go

RUN mkdir -p /app/bin
RUN find . -name "worker.go" -type f | head -5

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o worker ./worker.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/distributor .
COPY --from=builder /app/worker .

RUN chmod +x distributor worker
