FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o game-server main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/game-server .
COPY --from=builder /app/init ./init

EXPOSE 50051

CMD ["./game-server"]
