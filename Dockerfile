# Multi-stage production Dockerfile for PrefixOS

# Stage 1: Build binary
FROM golang:1.22-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git make protoc

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/prefixos cmd/server/main.go

# Stage 2: Runtime image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bin/prefixos /app/prefixos
COPY --from=builder /app/configs/config.yaml /app/configs/config.yaml

EXPOSE 50051 8080 9090 7000

ENTRYPOINT ["/app/prefixos", "--config=/app/configs/config.yaml"]
