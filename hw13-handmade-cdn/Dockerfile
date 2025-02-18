FROM golang:1.22.5 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY images/ ./images/

EXPOSE 8080

CMD ["./main"]
