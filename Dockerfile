FROM golang:1.22.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-ichiran-api ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates docker-cli docker-cli-compose

WORKDIR /app

COPY --from=builder /go-ichiran-api /app/go-ichiran-api

EXPOSE 8080

CMD ["/app/go-ichiran-api"]