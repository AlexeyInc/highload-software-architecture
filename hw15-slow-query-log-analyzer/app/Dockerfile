FROM golang:1.22.5 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Final image
FROM alpine:latest

RUN apk --no-cache add curl

WORKDIR /app/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
