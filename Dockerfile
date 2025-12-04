FROM golang:1.25.1-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Install swag CLI
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3

COPY . .

# Generate Swagger docs
RUN swag init -g ./cmd/main.go -o ./docs

# Build the application
RUN go build -v -o main ./cmd

# Final stage - minimal image
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary and docs from builder
COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./main"]