FROM golang:1.25.1-alpine

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3

COPY . .

RUN swag init -g main.go -o docs

RUN go build -o app ./cmd/main.go

COPY ca.pem ./

COPY docs ./docs/

EXPOSE 8080

CMD ["./app"]