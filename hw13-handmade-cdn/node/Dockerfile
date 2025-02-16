FROM golang:1.22 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go mod tidy && go build -o image_server

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/image_server .
COPY images/ ./images/

EXPOSE 8080

CMD ["./image_server"]
