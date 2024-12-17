# Stage 1: Build
FROM golang:1.22-alpine AS builder

# Install build tools
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main ./cmd/api

# Stage 2: Run
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
    #&& apk del tzdata

# Create non-root user
#RUN adduser -D -g '' appuser

# Create necessary directories
RUN mkdir -p /app/config /app/logs
    #&& chown -R appuser:appuser /app

WORKDIR /app

# Copy binary and config from builder
COPY --from=builder /app/main .
COPY --from=builder /app/cmd/config/config.docker.yaml /app/config/config.yaml

# Use non-root user
#USER appuser

# Environment variables
ENV APP_ENV=production \
    TZ=Asia/Shanghai \
    CONFIG_PATH=/app/config \
    CONFIG_NAME=config

# Expose port
EXPOSE 3000

# Start application
CMD ["./main"]
