# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy all source files
COPY . .

# Download dependencies and build
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s" -o slackliner .

# Runtime stage
FROM scratch

# Copy the binary from builder
COPY --from=builder /app/slackliner /slackliner

# Copy SSL certificates for HTTPS connections
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Run the application
ENTRYPOINT ["/slackliner"]
