FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY ./go.* .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /auth-service .
COPY .env .env

EXPOSE 50060

CMD [ "./auth-service", "--config=./.env" ]
