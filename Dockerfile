FROM golang:1.23.0-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main main.go

FROM alpine:3.20 AS final

RUN adduser -D user

WORKDIR /app

COPY --from=builder /app/main .

RUN chown user:user main

USER user

EXPOSE 8080

CMD ["./main"]