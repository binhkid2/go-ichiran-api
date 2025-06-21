FROM golang:1.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./api

FROM gcr.io/distroless/base-debian11
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
