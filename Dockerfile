# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o api-security-scanner .

# Final stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app

# Create non-root user for security
RUN adduser -D -s /bin/sh scanner

# Copy binary and configs
COPY --from=builder /app/api-security-scanner .
COPY --from=builder /app/config.yaml ./config.yaml
COPY --from=builder /app/config-test.yaml ./config-test.yaml

# Change ownership
RUN chown -R scanner:scanner /app

USER scanner

EXPOSE 8080 8081

CMD ["./api-security-scanner", "-dashboard"]
