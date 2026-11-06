# builder
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/wallet

# образ
FROM alpine:3.20

WORKDIR /root/

COPY --from=builder /main .

EXPOSE 8080

CMD ["./main"]