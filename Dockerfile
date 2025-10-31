FROM golang:1.25.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3

COPY . .

RUN swag init -g ./cmd/main.go -o ./docs

RUN go build -o app ./cmd/main.go

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/app .
COPY --from=builder /app/docs ./docs
COPY ca.pem ./ca.pem

EXPOSE 8080

CMD ["./app"]
