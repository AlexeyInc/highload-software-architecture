FROM golang:1.22.5 as builder

WORKDIR /app
COPY main.go .

RUN go mod init client && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o client .

FROM debian:bullseye-slim

WORKDIR /app
COPY --from=builder /app/client .

RUN apt update && apt install -y curl siege && rm -rf /var/lib/apt/lists/*

EXPOSE 8080
ENTRYPOINT ["/bin/sh", "-c"]
CMD ["./client"]