FROM golang:1.22.5 as builder

WORKDIR /app
COPY main.go .

RUN go mod init client && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o client .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/client .

# Install curl in the Alpine container
RUN apk add --no-cache curl

EXPOSE 8080
CMD ["./client"]